// kol 是《武士傳說》繁體中文重製版的可執行入口。
//
// 目前是 M4 的垂直切片：載入已抽取的 PNG 與 typed data，以原版解析度繪製書本畫面
// 與角色建立流程，並提供截圖模式讓對拍工具取得 remake 側的同狀態畫面。
package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"image"
	"image/draw"
	"image/png"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/kol-remake-cht/internal/assets"
	"github.com/wicanr2/kol-remake-cht/internal/book"
	"github.com/wicanr2/kol-remake-cht/internal/charcreate"
	"github.com/wicanr2/kol-remake-cht/internal/chardata"
	"github.com/wicanr2/kol-remake-cht/internal/font"
	"github.com/wicanr2/kol-remake-cht/internal/i18n"
	"github.com/wicanr2/kol-remake-cht/internal/roster"
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
	sceneBookOpen   scene = "book-open"
	sceneMenu       scene = "menu"
	sceneRoster     scene = "roster"
)

var scenes = []scene{sceneCover, sceneContents, sceneOpenSpread, sceneNewChar, sceneScript, sceneTitle, sceneScriptCJK, sceneBookOpen, sceneMenu, sceneRoster}

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

// rgbaFrom 把抽出的 PNG 解成 RGBA，供 CPU 端的動畫重播使用。
func rgbaFrom(lib *assets.Library, archive string, index int) (*image.RGBA, error) {
	src, _, err := lib.Decode(archive, index)
	if err != nil {
		return nil, err
	}
	out := image.NewRGBA(src.Bounds())
	draw.Draw(out, out.Bounds(), src, src.Bounds().Min, draw.Src)
	return out, nil
}

// writeRGBA 把畫面寫成 PNG；對拍工具讀的就是這個檔案。
func writeRGBA(path string, img *image.RGBA) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// runCharacterWrite 用 remake 的 writer 產出一筆槽位記錄，供與原版實際寫入逐 byte 對拍。
// 樣板是本機的原版 CHARDATA，不入庫；輸出也只作本機比對。
func runCharacterWrite(outPath, templatePath string, slot int, contractPath,
	itemPath, chardataOut string, c chardata.NewCharacter, stats string) error {
	if templatePath == "" {
		return fmt.Errorf("需要 -chardata-template")
	}
	raw, err := os.ReadFile(templatePath)
	if err != nil {
		return err
	}
	if len(raw) != chardata.RecordBytes*chardata.SlotCount {
		return fmt.Errorf("樣板長度 %d 不是 %d", len(raw),
			chardata.RecordBytes*chardata.SlotCount)
	}
	if slot < 0 || slot >= chardata.SlotCount {
		return fmt.Errorf("槽位 %d 超出範圍", slot)
	}
	parts := strings.Split(stats, ",")
	if len(parts) != 7 {
		return fmt.Errorf("-chardata-stats 需要七個值")
	}
	for i, p := range parts {
		v, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return fmt.Errorf("屬性 %d 不是整數：%w", i, err)
		}
		c.Primary[i] = v
	}
	contracts, err := chardata.LoadContracts(contractPath)
	if err != nil {
		return err
	}
	contract, ok := contracts.Get(c.ProfessionIndex)
	if !ok {
		return fmt.Errorf("沒有職業記錄 %d 的起始契約", c.ProfessionIndex)
	}
	items, err := chardata.LoadWeaponTemplates(itemPath)
	if err != nil {
		return err
	}
	base := slot * chardata.RecordBytes
	record, err := chardata.WriteRecord(raw[base:base+chardata.RecordBytes], c,
		contract, items)
	if err != nil {
		return err
	}
	doc := map[string]any{
		"schema": "kol-chardata-write-v1", "slot": slot,
		"character": map[string]any{"name": c.Name, "race": c.Race,
			"secondary": c.Secondary, "profession_index": c.ProfessionIndex,
			"primary_internal": c.Primary},
		"contract": contract,
		"written_offsets": chardata.WrittenOffsets(),
		"record_hex": hex.EncodeToString(record),
	}
	blob, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(outPath, append(blob, '\n'), 0o644); err != nil {
		return err
	}
	if chardataOut != "" {
		updated := make([]byte, len(raw))
		copy(updated, raw)
		copy(updated[base:base+chardata.RecordBytes], record)
		if err := os.WriteFile(chardataOut, updated, 0o644); err != nil {
			return err
		}
	}
	return nil
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
	roster  *roster.Roster
	canvas  *screen.Canvas
	state   *charcreate.State
	sequence *title.Sequence
	player   *title.Player
	book       *book.OpenPlayer
	bookAnim   *book.OpenAnimation
	bookWait   float64
	bookFrom   string // 動畫的起始畫面
	bookTarget string // 動畫的目標畫面
	bookLib    *assets.Library

	scene    scene
	portrait int
	selected int
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
	case sceneRoster:
		if g.strings != nil {
			return ui.RosterCJK(g.canvas.Image(), g.sprites, g.cjk, g.strings,
				g.roster.Occupied())
		}
		return ui.Roster(g.canvas.Image(), g.sprites, g.small, g.roster.Occupied())
	case sceneMenu:
		if g.strings != nil {
			return ui.CharacterOptionsCJK(g.canvas.Image(), g.sprites, g.cjk,
				g.strings, g.selected)
		}
		return ui.CharacterOptions(g.canvas.Image(), g.sprites, g.small, g.selected)
	case sceneBookOpen:
		if err := g.ensureBookPlayer(); err != nil {
			return err
		}
		g.canvas.Image().WritePixels(g.book.Image().Pix)
		return nil
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

// ensureBookPlayer 在遊戲迴圈內建立翻書動畫的播放器。
//
// 動畫的來源是「畫好的目標頁」，因此要先把目標畫面畫進離屏畫布再讀回；這一步只能在
// 繪圖環境就緒之後做，所以延後到第一次繪製時。
func (g *game) ensureBookPlayer() error {
	if g.book != nil {
		return nil
	}
	start, err := g.renderPage(g.bookFrom)
	if err != nil {
		return err
	}
	source, err := g.renderPage(g.bookTarget)
	if err != nil {
		return err
	}
	player, err := book.NewOpenPlayer(g.bookAnim, start, source)
	if err != nil {
		return err
	}
	g.book = player
	return nil
}

// renderPage 把一頁畫進離屏畫布再讀回。翻書動畫的起點與來源都是「畫好的那一頁」，
// 而這一步只能在繪圖環境就緒之後做，所以延後到第一次繪製時。
func (g *game) renderPage(name string) (*image.RGBA, error) {
	aux := screen.NewCanvas()
	switch name {
	case "cover", "contents":
		index := book.Cover
		if name == "contents" {
			index = book.Contents
		}
		img, err := g.sprites.Get("BOOKPICS", index)
		if err != nil {
			return nil, err
		}
		aux.DrawAt(img, 0, 0)
	case "menu":
		// 動畫期間沒有任何項目反白——反白是在動畫結束之後才畫上去的。
		// 這一點由原版 accept-and-save-run1 的 frame-036、039–043 逐像素確認。
		if err := ui.CharacterOptions(aux.Image(), g.sprites, g.small, -1); err != nil {
			return nil, err
		}
	case "menu-selected":
		if err := ui.CharacterOptions(aux.Image(), g.sprites, g.small, g.selected); err != nil {
			return nil, err
		}
	case "roster":
		if err := ui.Roster(aux.Image(), g.sprites, g.small, g.roster.Occupied()); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("未知的書頁 %q", name)
	}
	return aux.Snapshot(), nil
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
	if g.scene == sceneBookOpen && g.book != nil {
		// 每一步固定等 delay_ticks_per_step 個 BIOS tick；Ebitengine 固定 60 TPS。
		g.bookWait += title.TickHz / 60
		for g.bookWait >= float64(g.bookAnim.DelayTicks) && !g.book.Done() {
			g.bookWait -= float64(g.bookAnim.DelayTicks)
			g.book.Step()
		}
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
	menuSelected := flag.Int("menu-selected", 0, "menu 場景反白的項目索引")
	rosterPath := flag.String("roster",
		"assets/original-derived-data/character-roster.json", "角色槽位 typed data")
	shot := flag.String("shot", "", "單張截圖輸出路徑")
	shotDir := flag.String("shot-dir", "", "批次截圖輸出目錄")
	script := flag.String("script", "", "以空白分隔的操作腳本；每一步後截一張圖")
	seed := flag.Int("seed", 1, "亂數種子，供可重播的屬性擲骰與標題序列使用")
	titlePath := flag.String("title-sequence", "docs/re/title-sequence.json",
		"標題序列腳本（由 RE-047 從 TITLE.EXE 的 _main 轉寫）")
	titleShots := flag.String("title-shot-dir", "", "把標題序列一輪的每個畫面輸出到這個目錄")
	bookAnimPath := flag.String("book-animation",
		"assets/original-derived-data/book-open-animation.json", "翻書動畫參數表")
	bookShots := flag.String("book-shot-dir", "",
		"把翻書動畫的每一步輸出到這個目錄")
	bookTransition := flag.String("book-transition", "open",
		"翻書動畫的段落：open 是封面翻到目錄頁，menu 是目錄頁翻到角色選項頁")
	titleTimeline := flag.String("title-timeline", "",
		"把一輪標題序列的時間軸寫成 JSON 後結束")
	titleChoice := flag.String("title-choice", "rng",
		"標題序列兩處亂數的取法：rng 依 seed 重播，first／second 固定取第一或第二個選項")
	scaleFlag := flag.Int("scale", 3, "視窗放大倍率")
	listAssets := flag.Bool("list-assets", false, "只列出 manifest 摘要後結束")
	writeRecord := flag.String("chardata-write", "",
		"把一筆角色記錄寫成 JSON 後結束，供 tools/verify_chardata_write.py 對拍")
	writeTemplate := flag.String("chardata-template", "",
		"槽位樣板檔（本機的原版 CHARDATA），與 -chardata-slot 搭配")
	writeSlot := flag.Int("chardata-slot", 14, "要寫入的槽位編號")
	writeName := flag.String("chardata-name", "zed", "要寫入的角色名字")
	writeRace := flag.Int("chardata-race", 0, "種族 enum")
	writeSecondary := flag.Int("chardata-secondary", 0, "secondary selector")
	writeProfession := flag.Int("chardata-profession", 0, "職業記錄索引")
	writeStats := flag.String("chardata-stats", "", "七個逗號分隔的內部屬性值")
	writeScript := flag.String("chardata-script", "",
		"改用操作腳本跑完整個建立流程，把結果寫進槽位；與 -chardata-stats 二選一")
	writeOut := flag.String("chardata-out", "",
		"把寫入後的完整 CHARDATA 存到這個檔案（含原版樣板 bytes，只作本機比對）")
	itemTemplates := flag.String("chardata-weapons",
		"assets/original-derived-data/weapon-templates.json", "起始武器樣板表")
	writerSummary := flag.String("chardata-contract",
		"docs/re/ida-fingerprints/KOLBOOK-character-writer-summary.json",
		"職業起始契約（IDA 安全摘要）")
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
	if *writeRecord != "" {
		spec := chardata.NewCharacter{
			Name: *writeName, Race: *writeRace, Secondary: *writeSecondary,
			ProfessionIndex: *writeProfession,
		}
		stats := *writeStats
		if *writeScript != "" {
			createData, err := chardata.Load(*dataPath)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			st := charcreate.New(createData, &lcg{state: uint32(*seed)})
			for _, token := range strings.Fields(*writeScript) {
				if err := applyToken(st, token); err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(2)
				}
			}
			if st.Stage != charcreate.StageDone {
				fmt.Fprintf(os.Stderr, "腳本沒有走到完成，停在第 %d 層\n", st.Stage)
				os.Exit(2)
			}
			spec = chardata.NewCharacter{Name: st.Name, Race: st.Race,
				Secondary: st.Secondary,
				ProfessionIndex: st.ProfessionRecordIndex()}
			values := make([]string, 0, 7)
			for _, v := range st.Primary {
				values = append(values, strconv.Itoa(v))
			}
			stats = strings.Join(values, ",")
		}
		if err := runCharacterWrite(*writeRecord, *writeTemplate, *writeSlot,
			*writerSummary, *itemTemplates, *writeOut, spec, stats); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
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

	var slots *roster.Roster
	if scene(*sceneName) == sceneRoster || (*bookShots != "" && *bookTransition == "roster") {
		if slots, err = roster.Load(*rosterPath); err != nil {
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
		selected: *menuSelected,
		roster:   slots,
	}

	if scene(*sceneName) == sceneBookOpen || *bookShots != "" {
		anim, err := book.LoadOpenAnimation(*bookAnimPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		g.bookAnim = anim
		g.bookLib = lib
		switch *bookTransition {
		case "open":
			g.bookFrom, g.bookTarget = "cover", "contents"
		case "menu":
			g.bookFrom, g.bookTarget = "contents", "menu"
		case "roster":
			// 進入清單時，原版先把反白拿掉再開始捲動：frame-045 還有反白，
			// frame-046 起就沒有了。
			g.bookFrom, g.bookTarget = "menu", "roster"
		default:
			fmt.Fprintf(os.Stderr, "未知的 -book-transition %q\n", *bookTransition)
			os.Exit(2)
		}
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
	case *bookShots != "":
		g.scene = sceneBookOpen
		for i := 0; i <= len(g.bookAnim.Steps); i++ {
			job := shotJob{name: fmt.Sprintf("open-%03d", i),
				path: filepath.Join(*bookShots, fmt.Sprintf("open-%03d.png", i))}
			if i > 0 {
				job.apply = func() { g.book.Step() }
			}
			g.jobs = append(g.jobs, job)
		}
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
