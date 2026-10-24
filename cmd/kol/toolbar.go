package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/font"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

//go:embed toolbarfont/*.png toolbarfont/manifest.json toolbarfont/OFL.txt toolbarfont/source.json
var toolbarFonts embed.FS

// 偏好只保存數字，不保存任何戰鬥輔助啟用狀態。
type toolbarPreferences struct {
	EnemyMultiplier int    `json:"enemy_multiplier"`
	Stats           string `json:"stats_floor"`
	Skills          string `json:"skills_floor"`
	Gold            string `json:"gold_floor"`
	GameFont        string `json:"game_font"`
}

func defaultToolbarPreferences() toolbarPreferences {
	return toolbarPreferences{6, "", "200", "9000", "current"}
}

func toolbarPreferencesPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "kol-remake-cht", "toolbar.json"), nil
}

func loadToolbarPreferences(path string) (toolbarPreferences, error) {
	p := defaultToolbarPreferences()
	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return p, nil
	}
	if err != nil {
		return p, err
	}
	if err = json.Unmarshal(raw, &p); err != nil {
		return defaultToolbarPreferences(), err
	}
	if p.GameFont != "current" && p.GameFont != "rounded" {
		return defaultToolbarPreferences(), fmt.Errorf("無效的遊戲字型")
	}
	if p.EnemyMultiplier != 1 && p.EnemyMultiplier != 2 && p.EnemyMultiplier != 3 && p.EnemyMultiplier != 6 {
		return defaultToolbarPreferences(), fmt.Errorf("無效的疲勞倍率")
	}
	for i, v := range []string{p.Stats, p.Skills, p.Gold} {
		if i == 0 && v == "" {
			continue
		}
		n, e := strconv.Atoi(v)
		max := []int{170, 200, 16777215}[i]
		if e != nil || n < 0 || n > max || i != 2 && n == 0 {
			return defaultToolbarPreferences(), fmt.Errorf("無效的角色下限")
		}
	}
	return p, nil
}

func saveToolbarPreferences(path string, p toolbarPreferences) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".toolbar-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	if _, err = tmp.Write(raw); err != nil {
		tmp.Close()
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}

type toolbarState struct {
	awaitFrame                                                  bool
	ready, hidden, quick, paused, help, character, all, release bool
	focus                                                       int
	message                                                     string
	prefs                                                       toolbarPreferences
	assist                                                      battle.Assistance
	face                                                        *font.Font
	gameLayer, uiLayer                                          *ebiten.Image
}

var toolbarTop = []image.Rectangle{
	image.Rect(8, 8, 104, 40), image.Rect(112, 8, 208, 40), image.Rect(216, 8, 352, 40), image.Rect(360, 8, 472, 40),
	image.Rect(480, 8, 592, 40), image.Rect(600, 8, 704, 40), image.Rect(712, 8, 832, 40), image.Rect(840, 8, 952, 40),
}
var toolbarQuick = []image.Rectangle{
	image.Rect(8, 56, 184, 88), image.Rect(192, 56, 400, 88), image.Rect(408, 56, 552, 88), image.Rect(560, 56, 752, 88), image.Rect(760, 56, 952, 88),
}
var toolbarCharacter = []image.Rectangle{
	image.Rect(320, 204, 536, 240), image.Rect(552, 204, 840, 240),
	image.Rect(448, 264, 648, 300), image.Rect(680, 264, 840, 300),
	image.Rect(448, 324, 648, 360), image.Rect(680, 324, 840, 360),
	image.Rect(448, 384, 648, 420), image.Rect(680, 384, 840, 420), image.Rect(704, 504, 856, 540),
}

func (g *game) toolbarHeader() int {
	if g.toolbar.hidden {
		return 0
	}
	if g.toolbar.quick {
		return 96
	}
	return 48
}
func (g *game) desktopSize() (int, int) {
	b := g.outputCanvas().Image().Bounds()
	w := b.Dx()
	if w < 960 {
		w = 960
	}
	return w, b.Dy()*w/b.Dx() + g.toolbarHeader()*w/960
}
func (g *game) gamePoint(x, y int) (int, int) {
	w, _ := g.desktopSize()
	b := g.outputCanvas().Image().Bounds()
	y -= g.toolbarHeader() * w / 960
	if y < 0 {
		return -1000, -1000
	}
	return x * b.Dx() / w, y * b.Dx() / w
}
func (g *game) toolbarSession() *town.Session {
	if g.play != nil {
		return g.play
	}
	if g.flow != nil {
		return g.flow.Town
	}
	return nil
}
func (g *game) syncAssistance() {
	if g.quest != nil && g.quest.state != nil {
		g.quest.state.Assistance = g.toolbar.assist
		g.quest.state.ApplyAssistance()
	}
}
func (g *game) rememberToolbar() {
	p, err := toolbarPreferencesPath()
	if err == nil {
		err = saveToolbarPreferences(p, g.toolbar.prefs)
	}
	if err != nil {
		g.toolbar.message = "偏好保存失敗：" + err.Error()
	}
}
func (g *game) toolbarAction(i int) {
	t := &g.toolbar
	defer g.resizeToolbarWindow()
	switch i {
	case 0, 1:
		mode := []string{"ega", "hd"}[i]
		if i == 1 && (g.strings == nil || !g.hdInstalled()) {
			t.message = "此版本沒有可用的HD素材"
			t.help = true
			return
		}
		if err := g.setGraphics(mode); err != nil {
			t.message = "畫面切換失敗"
			t.help = true
		}
		g.displayEvent(mode)
	case 2:
		t.quick = !t.quick
		if t.quick {
			t.hidden = false
			g.displaySettings.Open = false
		}
		t.character = false
		t.focus = 0
	case 3:
		if err := g.setMusic(!g.musicEnabled); err != nil {
			g.musicUnavailableNotice()
		}
		g.displayEvent(fmt.Sprintf("music-%t", g.musicEnabled))
	case 4:
		t.character, t.help, t.paused = false, false, false
		t.quick = false
		g.toggleDisplayPanel()
	case 5:
		t.paused = !t.paused
	case 6:
		t.help = !t.help
	case 7:
		t.hidden = !t.hidden
		if t.hidden {
			t.quick, t.character, t.help = false, false, false
		}
	case 8:
		t.assist.NoFatigue = !t.assist.NoFatigue
	case 9:
		next := map[int]int{0: 2, 1: 2, 2: 3, 3: 6, 6: 1}[t.assist.EnemyFatigue]
		if t.assist.EnemyFatigue == 0 && t.prefs.EnemyMultiplier > 1 {
			next = t.prefs.EnemyMultiplier
		}
		t.assist.EnemyFatigue = next
		t.prefs.EnemyMultiplier = next
		g.rememberToolbar()
	case 10:
		t.assist.FearImmune = !t.assist.FearImmune
	case 11:
		t.character = true
		t.focus = 0
		t.message = ""
	case 12:
		t.assist = battle.Assistance{}
	}
	t.release = true
	g.syncAssistance()
	if g.logInput {
		fmt.Printf("toolbar action=%d no_fatigue=%t enemy_multiplier=%d fear_immune=%t\n", i, t.assist.NoFatigue, t.assist.Multiplier(), t.assist.FearImmune)
	}
	g.displayEvent(fmt.Sprintf("toolbar-%d", i))
}

func (g *game) resizeToolbarWindow() {
	if !g.interactive {
		return
	}
	w, _ := ebiten.WindowSize()
	dw, dh := g.desktopSize()
	ebiten.SetWindowSize(w, dh*w/dw)
}

func (g *game) strengthenFromToolbar(kind string, value string) {
	t := &g.toolbar
	if g.quest != nil && g.quest.state != nil {
		t.message = "戰鬥結束後才能永久修改角色"
		return
	}
	s := g.toolbarSession()
	if s == nil {
		t.message = "目前沒有隊伍"
		return
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		t.message = "請輸入範圍內的整數"
		return
	}
	count, err := s.Strengthen(t.all, kind, n)
	if err != nil {
		t.message = err.Error()
		return
	}
	// 世界隊伍與城鎮持有同一份角色，回城／新戰鬥也應看到修改。
	if g.play == s {
		g.worldParty = s.Party
	}
	t.message = fmt.Sprintf("已套用%d名角色；住宿時保存", count)
	if g.logInput {
		raw, _ := json.Marshal(s.Party)
		fmt.Printf("toolbar-party %s\n", raw)
	}
	g.rememberToolbar()
	g.displayEvent("strengthen-" + kind)
}
func (g *game) characterAction(i int) {
	t := &g.toolbar
	switch i {
	case 0:
		t.all = false
	case 1:
		t.all = true
	case 3:
		g.strengthenFromToolbar("stats", t.prefs.Stats)
	case 5:
		g.strengthenFromToolbar("skills", t.prefs.Skills)
	case 7:
		g.strengthenFromToolbar("gold", t.prefs.Gold)
	case 8:
		t.character = false
		t.release = true
	}
}

// 回到遊戲前等待全部鍵鈕釋放，避免方向鍵／Enter 穿透到遭遇與世界移動。
func toolbarInputHeld() bool {
	return len(inpututil.AppendPressedKeys(nil)) > 0 || ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) || ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight)
}
func (g *game) updateToolbar() (bool, error) {
	t := &g.toolbar
	if !t.ready {
		root, _ := fs.Sub(toolbarFonts, "toolbarfont")
		f, err := font.LoadFS(root, "toolbar")
		if err != nil {
			return true, err
		}
		t.face = f
		t.prefs = defaultToolbarPreferences()
		path, err := toolbarPreferencesPath()
		if err == nil {
			t.prefs, err = loadToolbarPreferences(path)
		}
		if err != nil {
			t.message = "偏好無法讀取，已回預設"
			t.help = true
		}
		if err := g.setGameFont(t.prefs.GameFont); err != nil {
			t.prefs.GameFont = "current"
			t.message = "字模無法載入，保留目前字型。"
			t.help = true
		}
		t.ready = true
	}
	g.syncAssistance()
	for _, hot := range []struct {
		k      ebiten.Key
		action int
	}{{ebiten.KeyF8, 2}, {ebiten.KeyF7, 5}, {ebiten.KeyF12, 7}, {ebiten.KeyF9, 3}, {ebiten.KeyF10, 4}} {
		if inpututil.IsKeyJustPressed(hot.k) {
			g.toolbarAction(hot.action)
			return true, nil
		}
	}
	if t.release {
		// 設定頁持續攔截遊戲輸入，控制項間切換不需要等待釋放。
		// 釋放閘門只在返回遊戲時生效。
		if t.quick || t.character || t.help || t.paused || g.displaySettings.Open {
			t.release = false
		} else {
			if !toolbarInputHeld() {
				t.release = false
			}
			return true, nil
		}
	}
	x, y := ebiten.CursorPosition()
	w, _ := g.desktopSize()
	p := image.Pt(x*960/w, y*960/w)
	click := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	if click && !t.hidden && p.Y < g.toolbarHeader() {
		for i, r := range toolbarTop {
			if p.In(r) {
				g.toolbarAction(i)
				return true, nil
			}
		}
		if t.quick {
			for i, r := range toolbarQuick {
				if p.In(r) {
					g.toolbarAction(i + 8)
					return true, nil
				}
			}
		}
		return true, nil
	}
	if t.character {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			t.character = false
			t.release = true
			g.resizeToolbarWindow()
			g.displayEvent("toolbar-back-character")
			return true, nil
		}
		if click {
			for i, r := range toolbarCharacter {
				if p.In(r) {
					t.focus = i
					g.characterAction(i)
					g.displayEvent(fmt.Sprintf("toolbar-field-%d", i))
				}
			}
		}
		toolbarFocus(&t.focus, len(toolbarCharacter))
		var field *string
		switch t.focus {
		case 2:
			field = &t.prefs.Stats
		case 4:
			field = &t.prefs.Skills
		case 6:
			field = &t.prefs.Gold
		}
		if field != nil {
			for _, c := range ebiten.AppendInputChars(nil) {
				if c >= '0' && c <= '9' && len(*field) < 8 {
					*field += string(c)
				}
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && len(*field) > 0 {
				*field = (*field)[:len(*field)-1]
			}
			if (ebiten.IsKeyPressed(ebiten.KeyControlLeft) || ebiten.IsKeyPressed(ebiten.KeyControlRight)) && inpututil.IsKeyJustPressed(ebiten.KeyA) {
				*field = ""
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter) {
			g.characterAction(t.focus)
		}
		return true, nil
	}
	if g.displaySettings.Open {
		return false, nil
	}
	if t.help || t.paused || t.quick {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			t.help = false
			t.paused = false
			t.quick = false
			t.release = true
			g.resizeToolbarWindow()
			g.displayEvent("toolbar-back")
		}
		if t.quick && !t.help && !t.paused {
			toolbarFocus(&t.focus, len(toolbarQuick))
			if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
				g.toolbarAction(t.focus + 8)
			}
		}
		return true, nil
	}
	return false, nil
}
func toolbarFocus(index *int, count int) {
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) || inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) || inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		*index = (*index + 1) % count
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) || inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		*index = (*index + count - 1) % count
	}
}

func (g *game) Draw(dst *ebiten.Image) {
	if !g.interactive {
		g.drawGame(dst)
		return
	}
	t := &g.toolbar
	b := g.outputCanvas().Image().Bounds()
	if t.gameLayer == nil || t.gameLayer.Bounds() != b {
		t.gameLayer = ebiten.NewImage(b.Dx(), b.Dy())
	}
	t.gameLayer.Clear()
	g.drawGame(t.gameLayer)
	w, _ := g.desktopSize()
	scale := float64(w) / 960
	var op ebiten.DrawImageOptions
	op.GeoM.Scale(float64(w)/float64(b.Dx()), float64(w)/float64(b.Dx()))
	op.GeoM.Translate(0, float64(g.toolbarHeader())*scale)
	dst.DrawImage(t.gameLayer, &op)
	if t.face == nil {
		return
	}
	if t.uiLayer == nil {
		t.uiLayer = ebiten.NewImage(960, 696)
	}
	t.uiLayer.Clear()
	g.drawToolbar(t.uiLayer)
	op = ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	dst.DrawImage(t.uiLayer, &op)
	if t.awaitFrame {
		fmt.Printf("toolbar-frame hd=%t header=%d\n", g.hd != nil, g.toolbarHeader())
		t.awaitFrame = false
	}
}
func toolbarRect(dst *ebiten.Image, r image.Rectangle, c color.RGBA) {
	vector.DrawFilledRect(dst, float32(r.Min.X), float32(r.Min.Y), float32(r.Dx()), float32(r.Dy()), c, false)
}
func (g *game) toolbarText(dst *ebiten.Image, x, y int, s string) {
	g.toolbar.face.Draw(dst, x, y, s, 230, 235, 245)
}
func (g *game) toolbarButton(dst *ebiten.Image, r image.Rectangle, s string, on, focus bool) {
	c := color.RGBA{42, 52, 70, 255}
	if on {
		c = color.RGBA{40, 90, 115, 255}
	}
	toolbarRect(dst, r, c)
	if focus {
		vector.StrokeRect(dst, float32(r.Min.X), float32(r.Min.Y), float32(r.Dx()), float32(r.Dy()), 1, color.RGBA{240, 204, 115, 255}, false)
	}
	g.toolbarText(dst, r.Min.X+6, r.Min.Y+6, s)
}
func (g *game) drawToolbar(dst *ebiten.Image) {
	t := &g.toolbar
	if !t.hidden {
		toolbarRect(dst, image.Rect(0, 0, 960, g.toolbarHeader()), color.RGBA{22, 28, 40, 255})
		for i, s := range []string{"原貌", "HD畫面", "輔助 F8", "配樂 F9", "畫面 F10", "暫停", "快捷鍵", "收起 F12"} {
			on := i == 0 && g.hd == nil || i == 1 && g.hd != nil || i == 2 && t.quick || i == 3 && g.musicEnabled || i == 5 && t.paused || i == 6 && t.help
			g.toolbarButton(dst, toolbarTop[i], s, on, false)
		}
		if t.quick {
			for i, s := range []string{"玩家不疲勞：" + toolbarOn(t.assist.NoFatigue), fmt.Sprintf("敵疲勞：%d倍", t.assist.Multiplier()), "免怯戰：" + toolbarOn(t.assist.FearImmune), "角色強化（永久）", "戰鬥輔助歸零"} {
				g.toolbarButton(dst, toolbarQuick[i], s, i == 0 && t.assist.NoFatigue || i == 1 && t.assist.Multiplier() > 1 || i == 2 && t.assist.FearImmune, t.focus == i && !t.character)
			}
		}
	}
	if t.character || t.help || t.paused {
		toolbarRect(dst, image.Rect(184, 136, 888, 554), color.RGBA{25, 32, 46, 250})
		if t.character {
			g.toolbarText(dst, 208, 156, "角色強化（永久）")
			g.toolbarText(dst, 208, 210, "套用目標")
			labels := []string{"目前發言者", "全隊", t.prefs.Stats, "套用", t.prefs.Skills, "套用", t.prefs.Gold, "補足", "返回 Esc"}
			for i, s := range labels {
				g.toolbarButton(dst, toolbarCharacter[i], s, i == 0 && !t.all || i == 1 && t.all, t.focus == i)
			}
			g.toolbarText(dst, 208, 270, "屬性下限 1~170")
			g.toolbarText(dst, 208, 330, "已學技能 1~200")
			g.toolbarText(dst, 208, 390, "金幣下限")
			g.toolbarText(dst, 208, 440, "關閉輔助不還原角色變更；住宿時保存。")
			g.toolbarText(dst, 208, 470, toolbarMessage(t.message))
		} else {
			title := "快捷鍵與工具列"
			if t.paused {
				title = "遊戲已暫停"
			}
			g.toolbarText(dst, 208, 160, title)
			for i, s := range []string{"F8 輔助　F9 配樂　F10 畫面設定", "F7 暫停　F12 收合工具列", "方向鍵或Tab移焦，Enter執行，Esc返回", "角色欄位：數字輸入，Ctrl+A清空", "輔助每次啟動關閉，角色強化永久保存", "設定期間遊戲暫停，返回後先放開按鍵", toolbarMessage(t.message)} {
				g.toolbarText(dst, 208, 216+i*40, s)
			}
		}
	}
	g.drawDesktopDisplay(dst)
}
func toolbarOn(on bool) string {
	if on {
		return "開"
	}
	return "關"
}
func toolbarMessage(s string) string {
	r := []rune(s)
	if len(r) > 32 {
		return string(r[:31]) + "…"
	}
	return s
}
