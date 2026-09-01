// kol 是《武士傳說》繁體中文重製版的可執行入口。
//
// 目前是 M4 的垂直切片：只載入已抽取的 PNG、以原版解析度繪製書本畫面，並提供
// 截圖模式讓對拍工具取得 remake 側的同狀態畫面。玩法規則尚未接上。
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/kol-remake-cht/internal/assets"
	"github.com/wicanr2/kol-remake-cht/internal/book"
	"github.com/wicanr2/kol-remake-cht/internal/screen"
)

type scene string

const (
	sceneCover      scene = "cover"
	sceneContents   scene = "contents"
	sceneNewChar    scene = "newchar"
	sceneOpenSpread scene = "spread"
)

var scenes = []scene{sceneCover, sceneContents, sceneNewChar, sceneOpenSpread}

// shotJob 是截圖模式的一項工作：畫哪個場景、用哪個角色圖像、寫到哪個檔案。
type shotJob struct {
	name     string
	scene    scene
	portrait int
	path     string
}

type game struct {
	lib     *assets.Library
	canvas  *screen.Canvas
	scene   scene
	portrai int

	jobs  []shotJob
	jobAt int
	drawn bool
}

func (g *ebitenImageCache) get(lib *assets.Library, archive string, index int) (*ebiten.Image, error) {
	key := fmt.Sprintf("%s/%d", archive, index)
	if img, ok := g.m[key]; ok {
		return img, nil
	}
	src, _, err := lib.Decode(archive, index)
	if err != nil {
		return nil, err
	}
	img := ebiten.NewImageFromImage(src)
	g.m[key] = img
	return img, nil
}

type ebitenImageCache struct{ m map[string]*ebiten.Image }

var cache = &ebitenImageCache{m: map[string]*ebiten.Image{}}

func (g *game) render() error {
	g.canvas.Clear()
	switch g.scene {
	case sceneCover:
		img, err := cache.get(g.lib, "BOOKPICS", book.Cover)
		if err != nil {
			return err
		}
		g.canvas.DrawAt(img, 0, 0)
	case sceneContents:
		img, err := cache.get(g.lib, "BOOKPICS", book.Contents)
		if err != nil {
			return err
		}
		g.canvas.DrawAt(img, 0, 0)
	case sceneOpenSpread:
		img, err := cache.get(g.lib, "BOOKPICS", book.OpenSpread)
		if err != nil {
			return err
		}
		g.canvas.DrawAt(img, 0, 0)
	case sceneNewChar:
		bg, err := cache.get(g.lib, "BOOKPICS", book.OpenSpread)
		if err != nil {
			return err
		}
		g.canvas.DrawAt(bg, 0, 0)
		portrait, err := cache.get(g.lib, "CHARPICS", g.portrai)
		if err != nil {
			return err
		}
		g.canvas.DrawAt(portrait, book.PortraitOrigin.X, book.PortraitOrigin.Y)
	default:
		return fmt.Errorf("未知場景 %q", g.scene)
	}
	return nil
}

func (g *game) Update() error {
	if len(g.jobs) > 0 {
		if !g.drawn {
			return nil
		}
		job := g.jobs[g.jobAt]
		if err := g.canvas.WritePNG(job.path); err != nil {
			return err
		}
		fmt.Printf("%s %s\n", job.name, job.path)
		g.jobAt++
		if g.jobAt >= len(g.jobs) {
			return ebiten.Termination
		}
		next := g.jobs[g.jobAt]
		g.scene, g.portrai, g.drawn = next.scene, next.portrait, false
		return nil
	}
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	return nil
}

func (g *game) Draw(dst *ebiten.Image) {
	if err := g.render(); err != nil {
		fmt.Fprintln(os.Stderr, "繪製失敗：", err)
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest
	dst.DrawImage(g.canvas.Image(), op)
	g.drawn = true
}

func (g *game) Layout(int, int) (int, int) { return screen.Width, screen.Height }

func main() {
	assetRoot := flag.String("assets", "assets/original-derived-png/ega", "抽取資產根目錄")
	sceneName := flag.String("scene", string(sceneCover), "要繪製的場景")
	portrait := flag.Int("portrait", 0, "New Character 場景使用的角色圖像圖號")
	shot := flag.String("shot", "", "單張截圖輸出路徑；指定後畫一幀就結束")
	shotDir := flag.String("shot-dir", "", "批次截圖輸出目錄：三個書本畫面加全部角色圖像")
	scaleFlag := flag.Int("scale", 3, "視窗放大倍率")
	listAssets := flag.Bool("list-assets", false, "只列出 manifest 摘要後結束")
	flag.Parse()

	lib, err := assets.Load(*assetRoot)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *listAssets {
		m := lib.Manifest()
		fmt.Printf("source_zip=%s sha256=%s archives=%d pngs=%d\n",
			m.SourceZip.Name, m.SourceZip.SHA256, m.ArchiveNum, m.PNGCount)
		for _, a := range m.Archives {
			fmt.Printf("  %-10s %d\n", a.Name, len(a.Images))
		}
		return
	}

	valid := false
	for _, s := range scenes {
		if scene(*sceneName) == s {
			valid = true
		}
	}
	if !valid {
		fmt.Fprintf(os.Stderr, "未知場景 %q\n", *sceneName)
		os.Exit(2)
	}

	g := &game{
		lib:     lib,
		canvas:  screen.NewCanvas(),
		scene:   scene(*sceneName),
		portrai: *portrait,
	}
	switch {
	case *shotDir != "":
		for _, s := range []struct {
			name  string
			scene scene
		}{{"cover", sceneCover}, {"contents", sceneContents}, {"spread", sceneOpenSpread}} {
			g.jobs = append(g.jobs, shotJob{
				name: s.name, scene: s.scene,
				path: filepath.Join(*shotDir, s.name+".png"),
			})
		}
		for i := 0; i < lib.Count("CHARPICS"); i++ {
			g.jobs = append(g.jobs, shotJob{
				name:     fmt.Sprintf("newchar-%03d", i),
				scene:    sceneNewChar,
				portrait: i,
				path:     filepath.Join(*shotDir, fmt.Sprintf("newchar-%03d.png", i)),
			})
		}
		g.scene, g.portrai = g.jobs[0].scene, g.jobs[0].portrait
	case *shot != "":
		g.jobs = []shotJob{{name: string(g.scene), scene: g.scene,
			portrait: g.portrai, path: *shot}}
	}

	ebiten.SetWindowSize(screen.Width**scaleFlag, screen.Height**scaleFlag)
	ebiten.SetWindowTitle("武士傳說")
	if err := ebiten.RunGame(g); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
