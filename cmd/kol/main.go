// kol 是《武士傳說》繁體中文重製版的可執行入口。
//
// 目前是 M4 的垂直切片：載入已抽取的 PNG 與 typed data，以原版解析度繪製書本畫面
// 與角色建立流程，並提供截圖模式讓對拍工具取得 remake 側的同狀態畫面。
package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/wicanr2/kol-remake-cht/internal/assets"
	"github.com/wicanr2/kol-remake-cht/internal/book"
	"github.com/wicanr2/kol-remake-cht/internal/charcreate"
	"github.com/wicanr2/kol-remake-cht/internal/chardata"
	"github.com/wicanr2/kol-remake-cht/internal/flow"
	"github.com/wicanr2/kol-remake-cht/internal/font"
	"github.com/wicanr2/kol-remake-cht/internal/i18n"
	"github.com/wicanr2/kol-remake-cht/internal/roster"
	"github.com/wicanr2/kol-remake-cht/internal/save"
	"github.com/wicanr2/kol-remake-cht/internal/screen"
	"github.com/wicanr2/kol-remake-cht/internal/title"
	"github.com/wicanr2/kol-remake-cht/internal/town"
	"github.com/wicanr2/kol-remake-cht/internal/ui"
	"github.com/wicanr2/kol-remake-cht/internal/world"
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
	sceneDelete     scene = "delete"
	sceneHarbor     scene = "harbor"
	sceneInn        scene = "inn"
	sceneOutdoor    scene = "outdoor"
	sceneBuilding   scene = "building"
	scenePlay       scene = "play"
	sceneFlow       scene = "flow"
	sceneWorld      scene = "world"
)

var scenes = []scene{sceneCover, sceneContents, sceneOpenSpread, sceneNewChar, sceneScript, sceneTitle, sceneScriptCJK, sceneBookOpen, sceneMenu, sceneRoster, sceneDelete, sceneHarbor, sceneInn, sceneOutdoor, sceneBuilding, scenePlay, sceneFlow, sceneWorld}

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
		"contract":        contract,
		"written_offsets": chardata.WrittenOffsets(),
		"record_hex":      hex.EncodeToString(record),
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

// applyPlayToken 把腳本 token 套到城鎮狀態機；token 對應原版的鍵盤操作。
func applyPlayToken(s *town.Session, token string) error {
	defer func() {
		// 在城門選了 DOOR 就換到世界地圖。位置是這座城鎮該入城點在世界地圖上的座標。
		if s.LeftTown && s.World != nil && !s.WorldEntered {
			s.WorldEntered = true
			s.OnLeaveTown(s.LeftGate)
		}
	}()
	// 招募清單開著的時候方向鍵移動反白，不是走路——原版的清單也是吃上下鍵。
	if s.Recruiting() {
		switch token {
		case "U":
			s.MoveList(-1)
			return nil
		case "D":
			s.MoveList(1)
			return nil
		}
	}
	switch token {
	case "L":
		s.Walk(-1, 0)
	case "R":
		s.Walk(1, 0)
	case "U":
		s.Walk(0, -1)
	case "D":
		s.Walk(0, 1)
	case "A":
		s.Walk(-1, -1)
	case "S":
		s.Walk(1, -1)
	case "Z":
		s.Walk(-1, 1)
	case "X":
		s.Walk(1, 1)
	case "<":
		s.MoveCursor(-1)
	case ">":
		s.MoveCursor(1)
	case "RET":
		s.Accept()
	case "SPACE":
		s.Continue()
	case "LISTEN":
		s.Listen()
	default:
		return fmt.Errorf("未知的城鎮操作 token：%q", token)
	}
	return nil
}

// playToken 把腳本 token 交給目前掌管畫面的那個狀態機：人在城裡就是城鎮，
// 走出城門之後就是野外。
func (g *game) playToken(token string) error {
	if g.play != nil && g.play.LeftTown && g.worldSession != nil {
		return applyWorldToken(g.worldSession, token)
	}
	return applyPlayToken(g.play, token)
}

// townSheet 把角色檔抽出來的資料頁欄位換成城鎮這一側的型別。
func townSheet(s *roster.Sheet) *town.Sheet {
	if s == nil || len(s.Stats) != town.StatCount {
		return nil
	}
	out := &town.Sheet{
		Race: s.Race, Profession: s.Profession, Rank: s.Rank,
		AdventurePoints: s.AdventurePoints, MovementRate: s.MovementRate,
		Nutrition: s.Nutrition, HeightFeet: s.HeightFeet,
		HeightInches: s.HeightInches, Weight: s.Weight,
		Armour: s.Armour, Shield: s.Shield,
	}
	copy(out.Stats[:], s.Stats)
	for _, w := range s.Weapons {
		out.Weapons = append(out.Weapons, town.WeaponRow{
			Name: w.Name, Offence: w.Offence, Defence: w.Defence})
	}
	return out
}

// buildSave 把目前的遊玩狀態組成一份存檔（SPEC-006）。人在城裡就記城鎮座標，
// 走到野外就記世界座標；兩邊都只放有 RE 證據的欄位。
func buildSave(s *town.Session, w *world.Screen, faithful bool) *save.Save {
	out := save.New()
	out.SavedAt = time.Now().UTC().Format(time.RFC3339)
	out.Settings.FaithfulInnStay = faithful
	if w != nil {
		t := w.Travel
		out.World = save.World{X: t.X, Y: t.Y, SubX: t.SubX, SubY: t.SubY}
	}
	if s.LeftTown {
		out.Location = save.Location{Kind: save.LocationWorld}
	} else {
		x, y := s.Outdoor.PartyPosition()
		out.Location = save.Location{Kind: save.LocationTown,
			Town: s.Outdoor.Town, EntryPoint: s.Outdoor.EntryPoint, X: x, Y: y}
	}
	for _, m := range s.Party {
		out.Party = append(out.Party, save.Member{
			Name: m.Name, Portrait: m.Portrait, Female: m.Female,
			Wounds: m.Wounds, Gold: m.Gold, Place: m.Place,
			TownCode: m.TownCode, BuildingID: m.BuildingID,
			Renown: m.Renown, Spellcaster: m.Spellcaster, Speaker: m.Speaker,
		})
	}
	// 存檔裡的發言者要明確：畫面上沒有標記時發言者就是隊長，寫進去也照這個規則，
	// 讀回來才不會變成「沒有人在說話」。
	if i := out.Speaker(); i >= 0 && !out.Party[i].Speaker {
		out.Party[i].Speaker = true
	}
	return out
}

// parsePair 解析 `X,Y` 形式的參數。
func parsePair(spec string) (int, int, error) {
	var a, b int
	if _, err := fmt.Sscanf(spec, "%d,%d", &a, &b); err != nil {
		return 0, 0, err
	}
	return a, b, nil
}

// enterWorld 把隊伍放到世界地圖上的某一格。城門選 DOOR 之後就是走這一條。
func enterWorld(s *world.Screen, p town.Point) {
	if s == nil {
		return
	}
	s.Travel = world.New(s.Travel.Data, s.Travel.Grid, p.X, p.Y)
}

// applyWorldToken 把腳本 token 套到野外狀態機。
func applyWorldToken(s *world.Session, token string) error {
	switch token {
	case "L":
		s.Walk(-1, 0)
	case "R":
		s.Walk(1, 0)
	case "U":
		s.Walk(0, -1)
	case "D":
		s.Walk(0, 1)
	case "A":
		s.Walk(-1, -1)
	case "S":
		s.Walk(1, -1)
	case "Z":
		s.Walk(-1, 1)
	case "X":
		s.Walk(1, 1)
	case "<":
		s.MoveCursor(-1)
	case ">":
		s.MoveCursor(1)
	case "RET":
		s.Accept()
	case "TICK":
		s.Screen.Phase++
	default:
		return fmt.Errorf("未知的野外操作 token：%q", token)
	}
	return nil
}

// parseIcons 解析 -outdoor-icons：`none` 是空列，其餘是逗號分隔的 KOLMPICS 圖號。
func parseIcons(spec string) ([]int, error) {
	if spec == "none" {
		return nil, nil
	}
	var out []int
	for _, field := range strings.Split(spec, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(field))
		if err != nil {
			return nil, fmt.Errorf("-outdoor-icons 的 %q 不是圖號：%w", field, err)
		}
		out = append(out, n)
	}
	return out, nil
}

// loadGrid 從資產目錄讀出某城鎮的地圖格子陣列（8-bit 灰階 PNG，像素值即磚塊編號）。
func loadGrid(lib *assets.Library, archive string) (*town.Grid, error) {
	src, _, err := lib.Decode(archive, 0)
	if err != nil {
		return nil, err
	}
	return town.LoadGrid(src)
}

type shotJob struct {
	name  string
	path  string
	apply func()
}

type game struct {
	worldState   *world.Screen
	worldSession *world.Session
	worldParty   []town.Member
	sprites      *spriteCache
	small        *font.Font
	large        *font.Font
	cjk          *font.Font
	strings      *i18n.Table
	roster       *roster.Roster
	remove       *roster.DeleteFlow
	inn          *town.Inn
	outdoor      *town.Outdoor
	play         *town.Session
	flow         *flow.Flow
	walkTick     int
	markerStep   int
	canvas       *screen.Canvas
	state        *charcreate.State
	sequence     *title.Sequence
	player       *title.Player
	book         *book.OpenPlayer
	bookAnim     *book.OpenAnimation
	bookWait     float64
	bookFrom     string // 動畫的起始畫面
	bookTarget   string // 動畫的目標畫面
	bookLib      *assets.Library

	scene    scene
	portrait int
	selected int
	frame    title.Frame

	jobs  []shotJob
	jobAt int
	drawn bool
}

// renderScene 用指定的場景跑一次既有的繪製分派，供 flow 借用。
func (g *game) renderScene(sc scene) error {
	saved := g.scene
	g.scene = sc
	err := g.render()
	g.scene = saved
	return err
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
	case sceneDelete:
		if g.strings != nil {
			return ui.DeleteListCJK(g.canvas.Image(), g.sprites, g.cjk, g.strings,
				g.remove.List(), g.remove.Cursor, g.remove.Confirm)
		}
		return ui.DeleteList(g.canvas.Image(), g.sprites, g.small,
			g.remove.List(), g.remove.Cursor, g.remove.Confirm)
	case sceneBuilding, sceneInn:
		if g.strings != nil {
			return ui.InnCJK(g.canvas.Image(), g.sprites, g.cjk, g.strings, g.inn)
		}
		return ui.Inn(g.canvas.Image(), g.sprites, g.small, g.large, g.inn)
	case sceneFlow:
		return g.renderFlow()
	case scenePlay:
		if g.play.LeftTown && g.worldState != nil {
			if g.strings != nil {
				return ui.WorldCJK(g.canvas.Image(), g.sprites, g.cjk, g.strings,
					g.worldParty, g.worldState, g.worldSession)
			}
			return ui.WorldSession(g.canvas.Image(), g.sprites, g.small,
				g.worldParty, g.worldState, g.worldSession)
		}
		if g.play.Mode == town.ModeInside {
			if g.strings != nil {
				return ui.InnCJK(g.canvas.Image(), g.sprites, g.cjk, g.strings, g.play.Inside)
			}
			return ui.Inn(g.canvas.Image(), g.sprites, g.small, g.large, g.play.Inside)
		}
		if g.strings != nil {
			return ui.OutdoorCJK(g.canvas.Image(), g.sprites, g.cjk, g.strings, g.play.Outdoor)
		}
		return ui.Outdoor(g.canvas.Image(), g.sprites, g.small, g.play.Outdoor)
	case sceneWorld:
		if g.strings != nil {
			return ui.WorldCJK(g.canvas.Image(), g.sprites, g.cjk, g.strings,
				g.worldParty, g.worldState, g.worldSession)
		}
		return ui.WorldSession(g.canvas.Image(), g.sprites, g.small,
			g.worldParty, g.worldState, g.worldSession)
	case sceneOutdoor:
		if g.strings != nil {
			return ui.OutdoorCJK(g.canvas.Image(), g.sprites, g.cjk, g.strings, g.outdoor)
		}
		return ui.Outdoor(g.canvas.Image(), g.sprites, g.small, g.outdoor)
	case sceneHarbor:
		if g.strings != nil {
			return ui.HarborMasterCJK(g.canvas.Image(), g.sprites, g.cjk, g.strings)
		}
		return ui.HarborMaster(g.canvas.Image(), g.sprites, g.small)
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
	if g.scene == scenePlay && g.play != nil {
		g.updatePlay()
	}
	if g.scene == sceneTitle && g.player != nil {
		// 原版的等待單位是 BIOS tick；Ebitengine 固定 60 TPS，換算後推進播放器。
		g.player.Advance(title.TickHz / 60)
		g.frame = g.player.Current()
	}
	return nil
}

// renderFlow 依 flow 目前的畫面分派到既有的繪製函式。每一張都是已經對拍過的畫面，
// flow 只決定現在畫哪一張。
func (g *game) renderFlow() error {
	dst := g.canvas.Image()
	switch g.flow.Scene {
	case flow.SceneTitle:
		return g.renderScene(sceneTitle)
	case flow.SceneCover:
		return g.renderScene(sceneCover)
	case flow.SceneContents:
		return g.renderScene(sceneContents)
	case flow.SceneMenu:
		if g.strings != nil {
			return ui.CharacterOptionsCJK(dst, g.sprites, g.cjk, g.strings, g.flow.MenuCursor)
		}
		return ui.CharacterOptions(dst, g.sprites, g.small, g.flow.MenuCursor)
	case flow.SceneRoster:
		return g.renderScene(sceneRoster)
	case flow.SceneNewChar:
		return g.renderScene(sceneNewChar)
	case flow.SceneDelete:
		return g.renderScene(sceneDelete)
	case flow.SceneHarbor:
		return g.renderScene(sceneHarbor)
	case flow.SceneSelect:
		if g.strings != nil {
			return ui.DeleteListCJK(dst, g.sprites, g.cjk, g.strings,
				g.roster.Occupied(), g.flow.SelectCursor, false)
		}
		return ui.SelectCharacter(dst, g.sprites, g.small,
			g.roster.Occupied(), g.flow.SelectCursor)
	case flow.SceneInn:
		if g.strings != nil {
			return ui.InnCJK(dst, g.sprites, g.cjk, g.strings, g.flow.Inn)
		}
		return ui.Inn(dst, g.sprites, g.small, g.large, g.flow.Inn)
	case flow.SceneTown:
		if g.flow.Town.Mode == town.ModeInside {
			if g.strings != nil {
				return ui.InnCJK(dst, g.sprites, g.cjk, g.strings, g.flow.Town.Inside)
			}
			return ui.Inn(dst, g.sprites, g.small, g.large, g.flow.Town.Inside)
		}
		if g.strings != nil {
			return ui.OutdoorCJK(dst, g.sprites, g.cjk, g.strings, g.flow.Town.Outdoor)
		}
		return ui.Outdoor(dst, g.sprites, g.small, g.flow.Town.Outdoor)
	}
	return fmt.Errorf("未知的 flow 畫面：%s", g.flow.Scene)
}

// walkTicksPerFrame 是隊伍標記的動畫節拍。原版用 `word_2DF28` 當延遲計數，
// 每三次繪製推進一格；Ebitengine 固定 60 TPS，取相近的節拍。
const walkTicksPerFrame = 8

// updatePlay 是城鎮這一段的輸入迴圈。鍵位照參考卡：方向鍵移動、
// `<`／`>` 移動圖示列白框、Return 執行。
func (g *game) updatePlay() {
	g.walkTick++
	if g.walkTick >= walkTicksPerFrame {
		g.walkTick = 0
		g.play.Outdoor.AdvanceMarker(&g.markerStep)
	}
	for _, k := range []struct {
		key    ebiten.Key
		dx, dy int
	}{
		{ebiten.KeyArrowLeft, -1, 0}, {ebiten.KeyArrowRight, 1, 0},
		{ebiten.KeyArrowUp, 0, -1}, {ebiten.KeyArrowDown, 0, 1},
		// 參考卡的斜走鍵。
		{ebiten.KeyA, -1, -1}, {ebiten.KeyS, 1, -1},
		{ebiten.KeyZ, -1, 1}, {ebiten.KeyX, 1, 1},
	} {
		if inpututil.IsKeyJustPressed(k.key) {
			g.play.Walk(k.dx, k.dy)
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyComma) {
		g.play.MoveCursor(-1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyPeriod) {
		g.play.MoveCursor(1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter) ||
		inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.play.Accept()
	}
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

// applyDeleteToken 把腳本 token 套到刪除流程；token 對應原版的鍵盤操作。
func applyDeleteToken(f *roster.DeleteFlow, token string) error {
	switch token {
	case "RET":
		f.Accept()
	case "UP":
		f.MoveCursor(-1)
	case "DOWN":
		f.MoveCursor(1)
	case "y", "Y":
		f.Answer(true)
	case "n", "N":
		f.Answer(false)
	default:
		return fmt.Errorf("未知刪除腳本 token %q", token)
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
	innPath := flag.String("inn", "docs/re/town-inn-sample.json",
		"旅店畫面的對拍樣本")
	innSelected := flag.Int("inn-selected", -1, "覆寫旅店圖示列的白框位置（-1 為沿用樣本）")
	outdoorPath := flag.String("outdoor", "docs/re/town-outdoor-sample.json",
		"戶外畫面樣本狀態 JSON")
	outdoorCamera := flag.String("outdoor-camera", "", "覆寫戶外鏡頭座標，格式 X,Y")
	outdoorFrame := flag.Int("outdoor-frame", -1, "覆寫隊伍標記的動畫畫格（-1 為沿用樣本）")
	outdoorMarker := flag.String("outdoor-marker", "", "覆寫隊伍標記在視窗內的偏移，格式 列,欄")
	townI18nPath := flag.String("town-i18n", "docs/i18n/town-zh-Hant.json",
		"城鎮建築的譯文表")
	townTextPath := flag.String("town-text",
		"assets/original-derived-data/town-text.json", "城鎮文字 typed data")
	buildingTown := flag.Int("building-town", 7, "建築內部畫面的城鎮編號")
	buildingDoor := flag.Int("building-door", 5, "建築內部畫面的門序號")
	worldPath := flag.String("world-map",
		"assets/original-derived-data/world-map.json", "野外地圖 typed data")
	worldPos := flag.String("world-pos", "", "覆寫野外世界座標，格式 X,Y")
	worldSub := flag.String("world-sub", "", "覆寫野外次格座標，格式 次格X,次格Y")
	worldWeather := flag.Int("world-weather", 0, "天候組 0…5")
	worldPhase := flag.Int("world-phase", 0, "雲層旋轉過的像素數")
	worldTime := flag.Int("world-time", 0, "時段 0…4")
	worldScript := flag.String("world-script", "", "野外的操作腳本；每一步後截一張圖")
	topicsPath := flag.String("town-topics",
		"assets/original-derived-data/town-keywords.json", "城鎮話題 typed data")
	uiTextPath := flag.String("ui-text",
		"assets/original-derived-data/ui-text.json", "UI 文字庫 typed data")
	buildingsPath := flag.String("town-buildings",
		"assets/original-derived-data/town-buildings.json", "城鎮建築 typed data")
	flowScript := flag.String("flow-script", "",
		"整條遊玩路徑的操作腳本，空白分隔；從標題序列開始，經書本、角色選項、"+
			"角色選擇清單、旅店走到戶外地圖")
	playScript := flag.String("play-script", "",
		"城鎮遊玩的操作腳本，空白分隔：L R U D 是方向鍵、"+
			"A S Z X 是斜走、< > 移動圖示白框、RET 執行")
	outdoorAtEdge := flag.Bool("outdoor-at-edge", false,
		"戶外畫面畫成走到地圖邊界的城門提示：地圖視窗換成林地圖，圖示列 DOOR 與 U-TURN")
	outdoorSelected := flag.Int("outdoor-selected", 0, "戶外圖示列白框的位置")
	outdoorAtDoor := flag.Bool("outdoor-at-door", false,
		"強制進入門口狀態（原版是移動到門的座標時才觸發）")
	outdoorIcons := flag.String("outdoor-icons", "", "覆寫戶外圖示列，逗號分隔的 KOLMPICS 圖號；`none` 代表整列淨空")
	deleteScript := flag.String("delete-script", "",
		"delete 場景的按鍵腳本，token 為 UP／DOWN／RET／y／n，逐步輸出截圖")
	rosterPath := flag.String("roster",
		"assets/original-derived-data/character-roster.json", "角色槽位 typed data")
	shot := flag.String("shot", "", "單張截圖輸出路徑")
	shotDir := flag.String("shot-dir", "", "批次截圖輸出目錄")
	script := flag.String("script", "", "以空白分隔的操作腳本；每一步後截一張圖")
	seed := flag.Int("seed", 1, "亂數種子，供可重播的屬性擲骰與標題序列使用")
	titlePath := flag.String("title-sequence", "docs/re/title-sequence.json",
		"標題序列腳本（由 RE-047 從 TITLE.EXE 的 _main 轉寫）")
	titleShots := flag.String("title-shot-dir", "", "把標題序列一輪的每個畫面輸出到這個目錄")
	titleFrame := flag.Int("title-frame", -1,
		"只輸出標題序列展開後的第 N 個畫面到 -shot，供展示圖使用")
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
	deleteSlot := flag.Int("chardata-delete-slot", -1,
		"以 remake 的刪除規則清掉這個槽位，把結果寫到 -chardata-out（本機比對用）")
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
	savePath := flag.String("save", "", "存檔路徑；客棧住宿時寫進這個檔（SPEC-006）")
	faithfulInn := flag.Bool("faithful-inn-stay", false,
		"客棧住宿完全照原版：問「存到硬碟嗎」、角色離隊、隊伍空了問是否離開遊戲。"+
			"與原版對拍時要打開")
	stringsPath := flag.String("strings", "docs/i18n/newchar-zh-Hant.json",
		"中文 UI 字串表")
	flag.Parse()

	lib, err := assets.Load(*assetRoot)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *deleteSlot >= 0 {
		if *writeTemplate == "" || *writeOut == "" {
			fmt.Fprintln(os.Stderr, "-chardata-delete-slot 需要 -chardata-template 與 -chardata-out")
			os.Exit(2)
		}
		raw, err := os.ReadFile(*writeTemplate)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := chardata.DeleteSlot(raw, *deleteSlot); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := os.WriteFile(*writeOut, raw, 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
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
				Secondary:       st.Secondary,
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
			"confirm": book.CJKConfirm, "town_inn": town.CJKInn,
			"town_outdoor": town.CJKOutdoor, "world": world.CJK,
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
	if scene(*sceneName) == sceneRoster || scene(*sceneName) == sceneDelete ||
		scene(*sceneName) == sceneFlow || *deleteScript != "" ||
		(*bookShots != "" && *bookTransition == "roster") {
		if slots, err = roster.Load(*rosterPath); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	var innState *town.Inn
	if scene(*sceneName) == sceneInn {
		if innState, err = town.LoadInn(*innPath); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if *innSelected >= 0 {
			if *innSelected >= len(innState.Icons) {
				fmt.Fprintf(os.Stderr, "-inn-selected %d 超出範圍\n", *innSelected)
				os.Exit(2)
			}
			innState.Selected = *innSelected
		}
	}

	var worldScreenState *world.Screen
	var worldSessionState *world.Session
	var worldPartyState []town.Member
	var outdoorState *town.Outdoor
	if scene(*sceneName) == sceneOutdoor {
		if outdoorState, err = town.LoadOutdoor(*outdoorPath); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if *outdoorCamera != "" {
			var cx, cy int
			if _, err := fmt.Sscanf(*outdoorCamera, "%d,%d", &cx, &cy); err != nil {
				fmt.Fprintf(os.Stderr, "-outdoor-camera 格式應為 X,Y：%v\n", err)
				os.Exit(2)
			}
			outdoorState.Camera.X, outdoorState.Camera.Y = cx, cy
		}
		if *outdoorFrame >= 0 {
			if *outdoorFrame >= town.MarkerFrames {
				fmt.Fprintf(os.Stderr, "-outdoor-frame %d 超出範圍\n", *outdoorFrame)
				os.Exit(2)
			}
			outdoorState.Marker.Frame = *outdoorFrame
		}
		if *outdoorMarker != "" {
			var mr, mc int
			if _, err := fmt.Sscanf(*outdoorMarker, "%d,%d", &mr, &mc); err != nil {
				fmt.Fprintf(os.Stderr, "-outdoor-marker 格式應為 列,欄：%v\n", err)
				os.Exit(2)
			}
			outdoorState.Marker.Row, outdoorState.Marker.Col = mr, mc
		}
		if *outdoorIcons != "" {
			icons, err := parseIcons(*outdoorIcons)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(2)
			}
			outdoorState.Icons = icons
			outdoorState.Selected = 0
		}
		if *outdoorSelected != 0 {
			outdoorState.Selected = *outdoorSelected
			outdoorState.DoorCursor = *outdoorSelected
		}
		grid, err := loadGrid(lib, outdoorState.Archive())
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		outdoorState.SetGrid(grid)
		markerSrc, _, err := lib.Decode(town.MarkerArchive, 0)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		markers, err := town.LoadMarkers(markerSrc)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		outdoorState.SetMarkers(markers)
		buildings, err := town.LoadBuildings(*buildingsPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		info := buildings.Town(outdoorState.Town)
		if info == nil {
			fmt.Fprintf(os.Stderr, "城鎮建築資料沒有城鎮 %d\n", outdoorState.Town)
			os.Exit(1)
		}
		outdoorState.SetBuildings(info)
		if *outdoorAtDoor {
			if outdoorState.DoorUnderParty() == nil {
				fmt.Fprintln(os.Stderr, "-outdoor-at-door：隊伍座標上沒有門")
				os.Exit(2)
			}
			outdoorState.AtDoorCell = true
		}
		if *outdoorAtEdge {
			outdoorState.AtEdge = true
		}
	}

	if scene(*sceneName) == sceneWorld || scene(*sceneName) == scenePlay {
		data, err := world.LoadData(*worldPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		src, _, err := lib.Decode("TOWN0", 0)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		grid, err := world.LoadGrid(src)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		sample, err := town.LoadOutdoor(*outdoorPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		px, py := 40, 23
		if *worldPos != "" {
			if px, py, err = parsePair(*worldPos); err != nil {
				fmt.Fprintf(os.Stderr, "-world-pos 格式應為 X,Y：%v\n", err)
				os.Exit(2)
			}
		}
		travel := world.New(data, grid, px, py)
		if *worldSub != "" {
			sx, sy, err := parsePair(*worldSub)
			if err != nil {
				fmt.Fprintf(os.Stderr, "-world-sub 格式應為 次格X,次格Y：%v\n", err)
				os.Exit(2)
			}
			travel.SubX, travel.SubY = sx, sy
		}
		base, overlay, mask := world.WeatherLayers(*worldWeather)
		layers := make([]image.Image, 0, 3)
		for _, index := range []int{base, overlay, mask} {
			img, _, err := lib.Decode(world.WeatherArchive, index)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			layers = append(layers, img)
		}
		weather, err := world.NewWeather(layers[0], layers[1], layers[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		bank, err := town.LoadUIText(*uiTextPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		trans, err := town.LoadTranslations(*townI18nPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		names := world.BankNamer{
			Padded: func(index int) string { return bank.RecordPadded(town.BankUI, index) }}
		if *lang != "en" {
			names.Zh = trans.UIRecord
		}
		worldPartyState = sample.Party
		worldScreenState = &world.Screen{Travel: travel, Weather: weather,
			Group: *worldWeather, Phase: *worldPhase, TimeIndex: *worldTime,
			Date: sample.Date, DateZh: sample.DateZh, Names: names}
		worldSessionState = world.NewSession(worldScreenState)
	}

	if scene(*sceneName) == sceneBuilding {
		buildings, err := town.LoadBuildings(*buildingsPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		info := buildings.Town(*buildingTown)
		if info == nil {
			fmt.Fprintf(os.Stderr, "城鎮建築資料沒有城鎮 %d\n", *buildingTown)
			os.Exit(1)
		}
		var door *town.Door
		for i := range info.Doors {
			if info.Doors[i].Index == *buildingDoor {
				door = &info.Doors[i]
			}
		}
		if door == nil {
			fmt.Fprintf(os.Stderr, "城鎮 %d 沒有門 %d\n", *buildingTown, *buildingDoor)
			os.Exit(1)
		}
		texts, err := town.LoadText(*townTextPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		sample, err := town.LoadInn(*innPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		innState = town.EnterBuilding(fmt.Sprintf("TWN%dPICS", *buildingTown), door,
			texts.Building(*buildingTown, *buildingDoor), sample.Party)
		if *lang != "en" {
			tr, err := town.LoadTranslations(*townI18nPath)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			innState.Apply(tr.Door(*buildingTown, *buildingDoor))
		}
	}

	var playSession *town.Session
	if scene(*sceneName) == scenePlay || scene(*sceneName) == sceneFlow {
		state, err := town.LoadOutdoor(*outdoorPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		grid, err := loadGrid(lib, state.Archive())
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		state.SetGrid(grid)
		markerSrc, _, err := lib.Decode(town.MarkerArchive, 0)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		markers, err := town.LoadMarkers(markerSrc)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		state.SetMarkers(markers)
		buildings, err := town.LoadBuildings(*buildingsPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		info := buildings.Town(state.Town)
		if info == nil {
			fmt.Fprintf(os.Stderr, "城鎮建築資料沒有城鎮 %d\n", state.Town)
			os.Exit(1)
		}
		texts, err := town.LoadText(*townTextPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		trans, err := town.LoadTranslations(*townI18nPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if *lang == "en" {
			trans = nil
		}
		playSession = town.NewSession(state, info, texts, trans, state.Party)
		topics, err := town.LoadTopics(*topicsPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		playSession.Topics = topics
		playSession.CJK = *lang != "en"
		if rosterState, err := roster.Load(*rosterPath); err == nil {
			for _, slot := range rosterState.Occupied() {
				playSession.Roster = append(playSession.Roster, town.Candidate{
					Name: slot.Name, Portrait: slot.Portrait,
					Place: slot.Place, Town: slot.Town,
					Female: slot.Female, Wounds: slot.Wounds,
					Gold: slot.Gold, BuildingID: slot.Building,
					Sheet: townSheet(slot.Sheet)})
			}
			// 畫面樣本只寫隊伍有誰，錢包與所在地在角色檔裡——住宿與治療要花的
			// 就是這一份錢（RE-073），所以照名字補齊。
			for i := range playSession.Party {
				for _, c := range playSession.Roster {
					if c.Name != playSession.Party[i].Name {
						continue
					}
					if playSession.Party[i].Gold == 0 {
						playSession.Party[i].Gold = c.Gold
					}
					if playSession.Party[i].Place == "" {
						playSession.Party[i].Place = c.Place
						playSession.Party[i].TownCode = c.Town
						playSession.Party[i].BuildingID = c.BuildingID
					}
					if playSession.Party[i].Sheet == nil {
						playSession.Party[i].Sheet = c.Sheet
					}
					break
				}
			}
		} else {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		bank, err := town.LoadUIText(*uiTextPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		playSession.UIText = bank
		if worldScreenState != nil {
			playSession.TownCode = worldScreenState.Travel.Data.TownCode(
				state.Town, state.EntryPoint)
			playSession.World = worldScreenState
			playSession.OnLeaveTown = func(p town.Point) {
				enterWorld(worldScreenState, p)
				worldSessionState.Reset()
			}
			// 野外選 DOOR 就回到城鎮：鏡頭、夾限、入城點都由世界地圖的記錄給
			// （RE-069），格子陣列與建築資料則換成該城鎮那一份。
			worldSessionState.OnEnterTown = func(rec *world.Town) {
				info := buildings.Town(rec.Town)
				if info == nil {
					fmt.Fprintf(os.Stderr, "城鎮建築資料沒有城鎮 %d\n", rec.Town)
					os.Exit(1)
				}
				grid, err := loadGrid(lib, fmt.Sprintf("TOWN%d", rec.Town))
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				playSession.EnterFromWorld(town.TownEntry{
					Town: rec.Town, Camera: town.Point{X: rec.Camera.X, Y: rec.Camera.Y},
					MinX: rec.Bounds.MinX, MaxX: rec.Bounds.MaxX,
					MinY: rec.Bounds.MinY, MaxY: rec.Bounds.MaxY,
					EntryPoint: rec.EntryPoint,
				}, info, grid)
			}
		}
		// 住宿存檔（SPEC-006）。沒有給 -save 就不寫檔，流程其餘部分照跑——
		// 截圖與對拍不需要真的產生存檔。
		playSession.FaithfulInnStay = *faithfulInn
		if *savePath != "" {
			playSession.SaveGame = func() error {
				return save.Write(*savePath,
					buildSave(playSession, worldScreenState, *faithfulInn))
			}
		}
		// LISTEN 的挑槽在截圖與玩家路徑裡必須可重播，因此用固定序列而不是真亂數。
		listenPick := 0
		playSession.Rand = func(n int) int {
			v := listenPick % n
			listenPick++
			return v
		}
	}

	var flowState *flow.Flow
	if scene(*sceneName) == sceneFlow {
		if innState == nil {
			if innState, err = town.LoadInn(*innPath); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
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
		sprites:      &spriteCache{lib: lib, m: map[string]*ebiten.Image{}},
		sequence:     sequence,
		small:        small,
		large:        large,
		cjk:          cjk,
		strings:      table,
		canvas:       canvasFor(*lang),
		state:        charcreate.New(data, &lcg{state: uint32(*seed)}),
		scene:        scene(*sceneName),
		portrait:     *portrait,
		selected:     *menuSelected,
		roster:       slots,
		inn:          innState,
		outdoor:      outdoorState,
		play:         playSession,
		flow:         flowState,
		worldState:   worldScreenState,
		worldSession: worldSessionState,
		worldParty:   worldPartyState,
	}
	if slots != nil {
		g.remove = roster.NewDeleteFlow(slots)
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
			"frames":       frames, "start_seconds": times[:len(times)-1],
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
	case *titleFrame >= 0:
		if *shot == "" {
			fmt.Fprintln(os.Stderr, "-title-frame 需要搭配 -shot")
			os.Exit(2)
		}
		frames, err := sequence.Expand(&lcg{state: uint32(*seed)})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if *titleFrame >= len(frames) {
			fmt.Fprintf(os.Stderr, "-title-frame %d 超出 %d 個畫面\n", *titleFrame, len(frames))
			os.Exit(2)
		}
		g.scene = sceneTitle
		g.frame = frames[*titleFrame]
		g.jobs = []shotJob{{name: "title", path: *shot}}
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
				name:  fmt.Sprintf("title-%03d", i),
				path:  filepath.Join(*titleShots, fmt.Sprintf("title-%03d.png", i)),
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
	case *flowScript != "":
		if *shotDir == "" && *shot == "" {
			fmt.Fprintln(os.Stderr, "-flow-script 需要搭配 -shot-dir 或 -shot")
			os.Exit(2)
		}
		g.scene = sceneFlow
		g.flow = flow.New(g.roster, g.state, g.remove, innState, playSession)
		if *shotDir == "" {
			for _, token := range strings.Fields(*flowScript) {
				g.flow.Key(token)
			}
			g.jobs = []shotJob{{name: "flow", path: *shot}}
			break
		}
		g.jobs = append(g.jobs, shotJob{
			name: "step-000", path: filepath.Join(*shotDir, "step-000.png")})
		for i, token := range strings.Fields(*flowScript) {
			tok := token
			g.jobs = append(g.jobs, shotJob{
				name:  fmt.Sprintf("step-%03d", i+1),
				path:  filepath.Join(*shotDir, fmt.Sprintf("step-%03d.png", i+1)),
				apply: func() { g.flow.Key(tok) },
			})
		}
	case *worldScript != "":
		if *shotDir == "" && *shot == "" {
			fmt.Fprintln(os.Stderr, "-world-script 需要搭配 -shot-dir 或 -shot")
			os.Exit(2)
		}
		g.scene = sceneWorld
		if *shotDir == "" {
			for _, token := range strings.Fields(*worldScript) {
				if err := applyWorldToken(g.worldSession, token); err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(2)
				}
			}
			g.jobs = []shotJob{{name: "world", path: *shot}}
			break
		}
		g.jobs = append(g.jobs, shotJob{
			name: "step-000", path: filepath.Join(*shotDir, "step-000.png")})
		for i, token := range strings.Fields(*worldScript) {
			tok := token
			g.jobs = append(g.jobs, shotJob{
				name: fmt.Sprintf("step-%03d", i+1),
				path: filepath.Join(*shotDir, fmt.Sprintf("step-%03d.png", i+1)),
				apply: func() {
					if err := applyWorldToken(g.worldSession, tok); err != nil {
						fmt.Fprintln(os.Stderr, err)
						os.Exit(2)
					}
				},
			})
		}
	case *playScript != "":
		if *shotDir == "" && *shot == "" {
			fmt.Fprintln(os.Stderr, "-play-script 需要搭配 -shot-dir 或 -shot")
			os.Exit(2)
		}
		g.scene = scenePlay
		if *shotDir == "" {
			for _, token := range strings.Fields(*playScript) {
				if err := g.playToken(token); err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(2)
				}
			}
			g.jobs = []shotJob{{name: "play", path: *shot}}
			break
		}
		g.jobs = append(g.jobs, shotJob{
			name: "step-000", path: filepath.Join(*shotDir, "step-000.png")})
		for i, token := range strings.Fields(*playScript) {
			tok := token
			g.jobs = append(g.jobs, shotJob{
				name: fmt.Sprintf("step-%03d", i+1),
				path: filepath.Join(*shotDir, fmt.Sprintf("step-%03d.png", i+1)),
				apply: func() {
					if err := g.playToken(tok); err != nil {
						fmt.Fprintln(os.Stderr, err)
						os.Exit(2)
					}
				},
			})
		}
	case *deleteScript != "":
		if *shotDir == "" && *shot == "" {
			fmt.Fprintln(os.Stderr, "-delete-script 需要搭配 -shot-dir 或 -shot")
			os.Exit(2)
		}
		g.scene = sceneDelete
		if *shotDir == "" {
			// 只要最後一張：先把整串 token 套完，再截一張圖。
			for _, token := range strings.Fields(*deleteScript) {
				if err := applyDeleteToken(g.remove, token); err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(2)
				}
			}
			g.jobs = []shotJob{{name: "delete", path: *shot}}
			break
		}
		g.jobs = append(g.jobs, shotJob{
			name: "step-000", path: filepath.Join(*shotDir, "step-000.png")})
		for i, token := range strings.Fields(*deleteScript) {
			tok := token
			g.jobs = append(g.jobs, shotJob{
				name: fmt.Sprintf("step-%03d", i+1),
				path: filepath.Join(*shotDir, fmt.Sprintf("step-%03d.png", i+1)),
				apply: func() {
					if err := applyDeleteToken(g.remove, tok); err != nil {
						fmt.Fprintln(os.Stderr, err)
						os.Exit(2)
					}
				},
			})
		}
	case *script != "":
		if *shotDir == "" && *shot == "" {
			fmt.Fprintln(os.Stderr, "-script 需要搭配 -shot-dir 或 -shot")
			os.Exit(2)
		}
		g.scene = sceneScript
		if *lang != "en" {
			g.scene = sceneScriptCJK
		}
		if *shotDir == "" {
			// 只要最後一張：先把整串 token 套完，再截一張圖。
			for _, token := range strings.Fields(*script) {
				if err := applyToken(g.state, token); err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(2)
				}
			}
			g.jobs = []shotJob{{name: "script", path: *shot}}
			break
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
				name:  fmt.Sprintf("newchar-%03d", i),
				path:  filepath.Join(*shotDir, fmt.Sprintf("newchar-%03d.png", i)),
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
