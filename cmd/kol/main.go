// kol 是《武士傳說》繁體中文重製版的可執行入口。
//
// 目前是 M4 的垂直切片：載入已抽取的 PNG 與 typed data，以原版解析度繪製書本畫面
// 與角色建立流程，並提供截圖模式讓對拍工具取得 remake 側的同狀態畫面。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/kol-remake-cht/internal/assets"
	"github.com/wicanr2/kol-remake-cht/internal/book"
	"github.com/wicanr2/kol-remake-cht/internal/charcreate"
	"github.com/wicanr2/kol-remake-cht/internal/chardata"
	"github.com/wicanr2/kol-remake-cht/internal/font"
	"github.com/wicanr2/kol-remake-cht/internal/i18n"
	"github.com/wicanr2/kol-remake-cht/internal/screen"
	"github.com/wicanr2/kol-remake-cht/internal/title"
	"github.com/wicanr2/kol-remake-cht/internal/ui"
)

type scene string

const (
	sceneCover      scene = "cover"
	sceneContents   scene = "contents"
	sceneOpenSpread scene = "spread"
	sceneNewChar    scene = "newchar"
	sceneScript     scene = "script"
	sceneTitle      scene = "title"
	sceneScriptCJK  scene = "script-cjk"
)

var scenes = []scene{sceneCover, sceneContents, sceneOpenSpread, sceneNewChar, sceneScript, sceneTitle, sceneScriptCJK}

// lcg 是可重播的亂數來源。原版以 BIOS tick 播種，remake 用固定 seed 才能重播。
type lcg struct{ state uint32 }

func (r *lcg) next() uint32 {
	r.state = r.state*1664525 + 1013904223
	return r.state
}

// Choose 供標題序列的二選一使用；原版是 rand()%n，remake 用同一個可重播來源。
func (r *lcg) Choose(n int) int {
	if n <= 0 {
		return 0
	}
	return int(r.next()>>16) % n
}

func (r *lcg) Roll(n int) int {
	if n <= 0 {
		return 0
	}
	return int(r.next()>>16)%n + 1
}

// fixedChoice 固定取同一個選項，用來窮舉標題序列兩處亂數的每一條分支。
// 對拍的判準是集合包含，因此兩次固定取法的聯集就涵蓋所有會出現的畫面。
type fixedChoice int

func (c fixedChoice) Choose(n int) int {
	if int(c) >= n {
		return n - 1
	}
	return int(c)
}

// canvasFor 依語言決定畫布大小：中文模式在放大後的畫布上作畫。
func canvasFor(lang string) *screen.Canvas {
	if lang == "en" {
		return screen.NewCanvas()
	}
	return screen.NewCanvasScaled(book.CJKScale)
}

type spriteCache struct {
	lib *assets.Library
	m   map[string]*ebiten.Image
}

func (c *spriteCache) Get(archive string, index int) (*ebiten.Image, error) {
	key := fmt.Sprintf("%s/%d", archive, index)
	if img, ok := c.m[key]; ok {
		return img, nil
	}
	src, _, err := c.lib.Decode(archive, index)
	if err != nil {
		return nil, err
	}
	img := ebiten.NewImageFromImage(src)
	c.m[key] = img
	return img, nil
}

type shotJob struct {
	name  string
	path  string
	apply func()
}

type game struct {
	sprites *spriteCache
	small   *font.Font
	large   *font.Font
	cjk     *font.Font
	strings *i18n.Table
	canvas  *screen.Canvas
	state   *charcreate.State
	sequence *title.Sequence
	player   *title.Player

	scene    scene
	portrait int
	frame    title.Frame

	jobs  []shotJob
	jobAt int
	drawn bool
}

func (g *game) render() error {
	g.canvas.Clear()
	switch g.scene {
	case sceneCover, sceneContents, sceneOpenSpread:
		index := map[scene]int{
			sceneCover: book.Cover, sceneContents: book.Contents,
			sceneOpenSpread: book.OpenSpread,
		}[g.scene]
		img, err := g.sprites.Get("BOOKPICS", index)
		if err != nil {
			return err
		}
		g.canvas.DrawAt(img, 0, 0)
	case sceneNewChar:
		bg, err := g.sprites.Get("BOOKPICS", book.OpenSpread)
		if err != nil {
			return err
		}
		g.canvas.DrawAt(bg, 0, 0)
		portrait, err := g.sprites.Get("CHARPICS", g.portrait)
		if err != nil {
			return err
		}
		g.canvas.DrawAt(portrait, book.PortraitOrigin.X, book.PortraitOrigin.Y)
	case sceneScript:
		return ui.NewCharacter(g.canvas.Image(), g.sprites, g.small, g.large, g.state)
	case sceneScriptCJK:
		return ui.NewCharacterCJK(g.canvas.Image(), g.sprites, g.cjk, g.strings, g.state)
	case sceneTitle:
		bg, err := g.sprites.Get("BACKPICS", g.frame.Back)
		if err != nil {
			return err
		}
		g.canvas.DrawAt(bg, 0, 0)
		if g.frame.Archive == "" || g.frame.Overlay < 0 {
			return nil
		}
		rect, ok := g.sequence.Rect(g.frame.Archive)
		if !ok {
			return fmt.Errorf("沒有 %s 的疊圖矩形", g.frame.Archive)
		}
		overlay, err := g.sprites.Get(g.frame.Archive, g.frame.Overlay)
		if err != nil {
			return err
		}
		g.canvas.DrawAt(overlay, rect.X, rect.Y)
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
		if apply := g.jobs[g.jobAt].apply; apply != nil {
			apply()
		}
		g.drawn = false
		return nil
	}
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	if g.scene == sceneTitle && g.player != nil {
		// 原版的等待單位是 BIOS tick；Ebitengine 固定 60 TPS，換算後推進播放器。
		g.player.Advance(title.TickHz / 60)
		g.frame = g.player.Current()
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

func (g *game) Layout(int, int) (int, int) {
	b := g.canvas.Image().Bounds()
	return b.Dx(), b.Dy()
}

// applyToken 把腳本 token 套到狀態機；token 對應原版的鍵盤操作。
func applyToken(st *charcreate.State, token string) error {
	switch token {
	case "RET":
		st.Accept()
	case "ESC":
		st.Cancel()
	case "UP":
		st.MoveCursor(-1)
		st.MovePortrait(-1)
	case "DOWN":
		st.MoveCursor(1)
		st.MovePortrait(1)
	case "LEFT":
		st.MovePortrait(-1)
	case "RIGHT":
		st.MovePortrait(1)
	case "BS":
		st.Backspace()
	case "SPACE":
		st.TypeRune(' ')
	default:
		r := []rune(token)
		if len(r) != 1 {
			return fmt.Errorf("未知腳本 token %q", token)
		}
		st.TypeRune(r[0])
	}
	return nil
}

func main() {
	assetRoot := flag.String("assets", "assets/original-derived-png/ega", "抽取資產根目錄")
	fontRoot := flag.String("fonts", "assets/original-derived-png/fonts", "字型資產目錄")
	dataPath := flag.String("data", "assets/original-derived-data/character-creation.json",
		"角色建立 typed data")
	sceneName := flag.String("scene", string(sceneCover), "要繪製的場景")
	portrait := flag.Int("portrait", 0, "newchar 場景使用的角色圖像圖號")
	shot := flag.String("shot", "", "單張截圖輸出路徑")
	shotDir := flag.String("shot-dir", "", "批次截圖輸出目錄")
	script := flag.String("script", "", "以空白分隔的操作腳本；每一步後截一張圖")
	seed := flag.Int("seed", 1, "亂數種子，供可重播的屬性擲骰與標題序列使用")
	titlePath := flag.String("title-sequence", "docs/re/title-sequence.json",
		"標題序列腳本（由 RE-047 從 TITLE.EXE 的 _main 轉寫）")
	titleShots := flag.String("title-shot-dir", "", "把標題序列一輪的每個畫面輸出到這個目錄")
	titleTimeline := flag.String("title-timeline", "",
		"把一輪標題序列的時間軸寫成 JSON 後結束")
	titleChoice := flag.String("title-choice", "rng",
		"標題序列兩處亂數的取法：rng 依 seed 重播，first／second 固定取第一或第二個選項")
	scaleFlag := flag.Int("scale", 3, "視窗放大倍率")
	listAssets := flag.Bool("list-assets", false, "只列出 manifest 摘要後結束")
	dumpLayout := flag.String("dump-cjk-layout", "",
		"把中文模式的版面座標寫成 JSON 後結束，供文字安全矩形檢查使用")
	lang := flag.String("lang", "en", "介面語言：en 走原版英文字模的嚴格對拍模式，"+
		"zh-Hant 走中文模式（畫布放大兩倍、16 px 中文字格，不對拍）")
	uiFontRoot := flag.String("ui-fonts", "assets/ui-font", "中文 UI 字模表目錄")
	stringsPath := flag.String("strings", "docs/i18n/newchar-zh-Hant.json",
		"中文 UI 字串表")
	flag.Parse()

	lib, err := assets.Load(*assetRoot)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *dumpLayout != "" {
		raw, err := json.MarshalIndent(map[string]any{
			"schema": "kol-cjk-layout-v1", "scale": book.CJKScale,
			"cell": book.CJKCell, "layout": book.CJKLayout, "text_box": book.CJKTextBox,
		}, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := os.WriteFile(*dumpLayout, append(raw, '\n'), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
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
	small, err := font.Load(*fontRoot, "book-small")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	large, err := font.Load(*fontRoot, "book-large")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var cjk *font.Font
	var table *i18n.Table
	if *lang != "en" {
		cjk, err = font.Load(*uiFontRoot, "zh-hant-16")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		table, err = i18n.Load(*stringsPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	data, err := chardata.Load(*dataPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	sequence, err := title.Load(*titlePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	g := &game{
		sprites:  &spriteCache{lib: lib, m: map[string]*ebiten.Image{}},
		sequence: sequence,
		small:    small,
		large:    large,
		cjk:      cjk,
		strings:  table,
		canvas:   canvasFor(*lang),
		state:    charcreate.New(data, &lcg{state: uint32(*seed)}),
		scene:    scene(*sceneName),
		portrait: *portrait,
	}

	if scene(*sceneName) == sceneTitle {
		// 實際播放模式：依 BIOS tick 推進，按鍵中止，與原版同一條腳本。
		frames, err := sequence.Expand(&lcg{state: uint32(*seed)})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		g.player = title.NewPlayer(frames, title.ThunderSeconds*title.TickHz)
		g.frame = g.player.Current()
	}

	switch {
	case *titleTimeline != "":
		frames, err := sequence.Expand(&lcg{state: uint32(*seed)})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		thunder := title.ThunderSeconds * title.TickHz
		times := sequence.Timeline(frames, thunder)
		raw, err := json.MarshalIndent(map[string]any{
			"schema": "kol-title-timeline-v1", "seed": *seed,
			"thunder_seconds": title.ThunderSeconds, "tick_hz": title.TickHz,
			"loop_seconds": times[len(times)-1],
			"frames": frames, "start_seconds": times[:len(times)-1],
		}, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := os.WriteFile(*titleTimeline, append(raw, '\n'), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("frames=%d loop_seconds=%.2f\n", len(frames), times[len(times)-1])
		return
	case *titleShots != "":
		var chooser title.Rand
		switch *titleChoice {
		case "rng":
			chooser = &lcg{state: uint32(*seed)}
		case "first":
			chooser = fixedChoice(0)
		case "second":
			chooser = fixedChoice(1)
		default:
			fmt.Fprintf(os.Stderr, "未知的 -title-choice %q\n", *titleChoice)
			os.Exit(2)
		}
		frames, err := sequence.Expand(chooser)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		g.scene = sceneTitle
		log := make([]title.Frame, 0, len(frames))
		for i, f := range frames {
			frame := f
			log = append(log, frame)
			g.jobs = append(g.jobs, shotJob{
				name: fmt.Sprintf("title-%03d", i),
				path: filepath.Join(*titleShots, fmt.Sprintf("title-%03d.png", i)),
				apply: func() { g.frame = frame },
			})
		}
		g.frame = frames[0]
		raw, err := json.MarshalIndent(map[string]any{
			"schema": "kol-title-frames-v1", "choice": *titleChoice, "seed": *seed,
			"sequence": *titlePath, "frames": log}, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := os.MkdirAll(*titleShots, 0o755); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := os.WriteFile(filepath.Join(*titleShots, "frames.json"),
			append(raw, '\n'), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case *script != "":
		if *shotDir == "" {
			fmt.Fprintln(os.Stderr, "-script 需要搭配 -shot-dir")
			os.Exit(2)
		}
		g.scene = sceneScript
		if *lang != "en" {
			g.scene = sceneScriptCJK
		}
		g.jobs = append(g.jobs, shotJob{
			name: "step-000", path: filepath.Join(*shotDir, "step-000.png")})
		for i, token := range strings.Fields(*script) {
			tok := token
			g.jobs = append(g.jobs, shotJob{
				name: fmt.Sprintf("step-%03d", i+1),
				path: filepath.Join(*shotDir, fmt.Sprintf("step-%03d.png", i+1)),
				apply: func() {
					if err := applyToken(g.state, tok); err != nil {
						fmt.Fprintln(os.Stderr, err)
						os.Exit(2)
					}
				},
			})
		}
	case *shotDir != "":
		for _, s := range []struct {
			name  string
			scene scene
		}{{"cover", sceneCover}, {"contents", sceneContents}, {"spread", sceneOpenSpread}} {
			sc := s.scene
			g.jobs = append(g.jobs, shotJob{
				name: s.name, path: filepath.Join(*shotDir, s.name+".png"),
				apply: func() { g.scene = sc },
			})
		}
		for i := 0; i < lib.Count("CHARPICS"); i++ {
			index := i
			g.jobs = append(g.jobs, shotJob{
				name: fmt.Sprintf("newchar-%03d", i),
				path: filepath.Join(*shotDir, fmt.Sprintf("newchar-%03d.png", i)),
				apply: func() { g.scene, g.portrait = sceneNewChar, index },
			})
		}
		g.scene = sceneCover
	case *shot != "":
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
		g.jobs = []shotJob{{name: *sceneName, path: *shot}}
	}

	bounds := g.canvas.Image().Bounds()
	ebiten.SetWindowSize(bounds.Dx()**scaleFlag, bounds.Dy()**scaleFlag)
	ebiten.SetWindowTitle("武士傳說")
	if err := ebiten.RunGame(g); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
