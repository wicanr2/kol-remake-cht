package ui

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

type HDGlyph struct {
	X, Y    int
	Advance float64
}
type hdInterface struct {
	images           map[string]*ebiten.Image
	glyphs           map[string]HDGlyph
	t                Translator
	npcKey, crestKey string
	scratch          *ebiten.Image
}

var hdInk = color.RGBA{238, 225, 202, 255}
var hdGold = color.RGBA{212, 176, 112, 255}
var hdMuted = color.RGBA{182, 169, 154, 255}

func hdRect(x, y, w, h int) image.Rectangle { return image.Rect(x, y, x+w, y+h) }

func (h *HD) AttachInterface(images map[string]image.Image, glyphs map[string]HDGlyph, npcKey, crestKey string, t Translator) {
	s := &hdInterface{images: map[string]*ebiten.Image{}, glyphs: glyphs, t: t, npcKey: npcKey, crestKey: crestKey, scratch: ebiten.NewImage(1920, 1200)}
	for n, img := range images {
		s.images[n] = ebiten.NewImageFromImage(img)
	}
	s.images["innkeeper-cropped.png"] = s.images["innkeeper-painted-v1.png"].SubImage(image.Rect(0, 90, 1254, 1084)).(*ebiten.Image)
	h.style = s
}
func (s *hdInterface) word(text string) string {
	if s.t != nil {
		return s.t.T(text)
	}
	return text
}
func (s *hdInterface) panel(dst *ebiten.Image, r image.Rectangle, active bool) {
	n := 0
	if active {
		n = 1
	}
	pic := s.images[fmt.Sprintf("panel-%d-%d-%d.png", r.Dx(), r.Dy(), n)]
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(float64(r.Min.X-4), float64(r.Min.Y-4))
	dst.DrawImage(pic, &op)
}
func (s *hdInterface) measure(text string, size float64) float64 {
	w := 0.0
	for _, ch := range text {
		g, ok := s.glyphs[string(ch)]
		if !ok {
			g = s.glyphs["?"]
		}
		w += g.Advance * size / 64
	}
	return w
}

// label 使用核准原型的比例字距與基線；名牌過長以省略號顯示，裁切限於文字安全矩形。
func (s *hdInterface) label(dst *ebiten.Image, text string, r image.Rectangle, size float64, c color.RGBA, center bool) {
	if text == "" {
		return
	}
	rs := []rune(text)
	truncated := false
	for len(rs) > 0 && s.measure(string(rs), size) > float64(r.Dx()) {
		rs = rs[:len(rs)-1]
		truncated = true
	}
	if truncated {
		for len(rs) > 0 && s.measure(string(rs)+"…", size) > float64(r.Dx()) {
			rs = rs[:len(rs)-1]
		}
		rs = append(rs, '…')
	}
	x := float64(r.Min.X)
	if center {
		x += (float64(r.Dx()) - s.measure(string(rs), size)) / 2
	}
	scale := size / 64
	baseline := float64(r.Min.Y) + (float64(r.Dy())-size)/2 + size*.88
	target := dst.SubImage(r).(*ebiten.Image)
	for _, ch := range rs {
		g, ok := s.glyphs[string(ch)]
		if !ok {
			g = s.glyphs["?"]
		}
		var op ebiten.DrawImageOptions
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(x-10*scale, baseline-76*scale)
		op.ColorScale.ScaleWithColor(c)
		op.Filter = ebiten.FilterLinear
		target.DrawImage(s.images["font-atlas.png"].SubImage(hdRect(g.X, g.Y, 96, 96)).(*ebiten.Image), &op)
		x += g.Advance * scale
	}
}
func (s *hdInterface) begin(dst *ebiten.Image) {
	s.scratch.Clear()
	s.scratch.DrawImage(dst, nil)
	dst.DrawImage(s.images["background.png"], nil)
}
func (s *hdInterface) preserve(dst *ebiten.Image, r image.Rectangle) {
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(float64(r.Min.X), float64(r.Min.Y))
	dst.DrawImage(s.scratch.SubImage(r).(*ebiten.Image), &op)
}
func (h *HD) partyHD(dst *ebiten.Image, sprites Sprites, members []town.Member) error {
	s := h.style
	for i, m := range members {
		pic, err := memberPortrait(sprites, m.Portrait, m.Picture)
		if err != nil {
			return err
		}
		r := hdRect(84+288*i, 6, 256, 288)
		s.panel(dst, r, m.Speaker)
		if index, ok := h.matching[h.fingerprint(pic)]; ok {
			pic = h.portraits[index]
		}
		h.drawHD(dst, pic, hdRect(r.Min.X+8, r.Min.Y+8, 240, 230))
		ink := hdInk
		if m.Speaker {
			ink = hdGold
		}
		s.label(dst, m.PlateName(), hdRect(r.Min.X+12, r.Min.Y+244, 232, 34), 28, ink, true)
	}
	return nil
}

type hdCommand struct{ key, title string }

func hdCommandFor(code int, battling bool) hdCommand {
	if battling {
		switch code {
		case battle.CmdYes:
			return hdCommand{"yes", "確認"}
		case battle.CmdWeapon:
			return hdCommand{"sword", "武器攻擊"}
		case battle.CmdUnarmed:
			return hdCommand{"fist", "徒手"}
		case battle.CmdSwitch, battle.CmdSwitchAlt1, battle.CmdSwitchAlt2:
			return hdCommand{"swap", "切換武器"}
		case battle.CmdDrop:
			return hdCommand{"drop", "丟棄"}
		case battle.CmdMove:
			return hdCommand{"move", "移動"}
		case battle.CmdRest:
			return hdCommand{"rest", "休息"}
		case battle.CmdCancel:
			return hdCommand{"back", "返回"}
		}
	}
	if !battling {
		switch code {
		case town.IconDoor:
			return hdCommand{"door", "進門"}
		case town.IconUTurn:
			return hdCommand{"back", "返回"}
		case town.IconRest:
			return hdCommand{"rest", "休息"}
		case town.IconSpacebar:
			return hdCommand{"", "繼續"}
		case 7:
			return hdCommand{"yes", "確認"}
		}
	}
	return hdCommand{}
}
func (h *HD) toolbarHD(dst *ebiten.Image, sprites Sprites, indices []int, selected int, battling, inside bool) error {
	s := h.style
	for i, index := range indices {
		r := hdRect(192+i*192, 1020, 144, 136)
		s.panel(dst, r, i == selected)
		command := hdCommandFor(index-3, battling)
		if inside && index-3 == town.IconDoor {
			command.title = "返回"
		}
		if command.key != "" {
			h.drawHD(dst, s.images["icon-"+command.key+".png"], hdRect(r.Min.X+35, r.Min.Y+12, 74, 74))
		} else {
			pic, err := sprites.Get("KOLMPICS", index)
			if err != nil {
				return err
			}
			if h.pages != nil {
				if icon, ok := h.pages.icons[h.fingerprint(pic)]; ok {
					pic = icon
				}
			}
			h.drawHD(dst, pic, hdRect(r.Min.X+24, r.Min.Y+8, 96, 80))
		}
		s.label(dst, command.title, hdRect(r.Min.X+10, r.Min.Y+96, 124, 30), 26, hdInk, true)
	}
	return nil
}
func (h *HD) innHD(dst *ebiten.Image, sprites Sprites, in *town.Inn) error {
	s := h.style
	lines := in.DialogZh
	if len(lines) == 0 {
		lines = in.Dialog
	}
	if len(lines) > 9 || len(in.Title) > 2 {
		return h.party(dst, sprites, sheetVisibleParty(in))
	}
	// 輸入游標子頁仍用現有文字引擎，避免改變哨符或遮蔽游標。
	for _, line := range append(append([]string{}, in.Dialog...), in.DialogZh...) {
		if strings.Contains(line, town.TextCaret) {
			return h.party(dst, sprites, sheetVisibleParty(in))
		}
	}
	s.begin(dst)
	if err := h.partyHD(dst, sprites, sheetVisibleParty(in)); err != nil {
		return err
	}
	s.panel(dst, hdRect(54, 390, 514, 372), false)
	if in.Portrait.Archive != "" {
		pic, err := sprites.Get(in.Portrait.Archive, in.Portrait.Index)
		if err != nil {
			return err
		}
		if h.fingerprint(pic) == s.npcKey {
			pic = s.images["innkeeper-cropped.png"]
		} else if painted, ok := h.pagePortrait(pic); ok {
			pic = painted
		}
		h.drawHD(dst, pic, hdRect(62, 398, 498, 356))
	}
	s.panel(dst, hdRect(612, 350, 1166, 118), false)
	for i, line := range in.Title {
		if i > 1 {
			break
		}
		size, ink, y, height := 34.0, hdGold, 360, 44
		if i == 1 {
			size, ink, y, height = 28, hdMuted, 412, 38
		}
		s.label(dst, s.word(line), hdRect(642, y, 1106, height), size, ink, true)
	}
	s.panel(dst, hdRect(612, 504, 1166, 406), false)
	size, step := 34.0, 50
	if len(lines) > 7 {
		size, step = 30, 42
	}
	for i, line := range lines {
		if i >= 9 {
			break
		}
		s.label(dst, line, hdRect(650, 526+i*step, 1090, step-6), size, hdInk, false)
	}
	return h.toolbarHD(dst, sprites, in.Icons, in.Selected, false, true)
}
func (h *HD) outdoorHD(dst *ebiten.Image, sprites Sprites, out *town.Outdoor) error {
	s := h.style
	s.begin(dst)
	if err := h.partyHD(dst, sprites, out.Party); err != nil {
		return err
	}
	s.panel(dst, hdRect(84, 342, 400, 132), false)
	lines := out.DateZh
	if len(lines) == 0 {
		lines = out.Date
	}
	for i, line := range lines {
		if i > 1 {
			break
		}
		size, ink, y, height := 32.0, hdInk, 356, 42
		if i == 1 {
			size, ink, y, height = 30, hdMuted, 412, 40
		}
		s.label(dst, s.word(line), hdRect(110, y, 344, height), size, ink, false)
	}
	crest, err := sprites.Get(out.CrestArchive(), town.Crest.Index)
	if err != nil {
		return err
	}
	if h.fingerprint(crest) == s.crestKey {
		h.drawHD(dst, s.images["crest.png"], hdRect(130, 502, 320, 320))
	} else {
		s.preserve(dst, hdRect(120, 490, 430, 450))
	}
	s.panel(dst, hdRect(618, 342, 1164, 588), false)
	s.preserve(dst, hdRect(624, 348, 1152, 576))
	if door := out.AtDoor(); door != nil && !out.AtEdge {
		pic, err := sprites.Get(out.CrestArchive(), door.Picture)
		if err != nil {
			return err
		}
		if h.fingerprint(pic) == h.facadeKey {
			h.drawHD(dst, h.facade, hdRect(624, 348, 1152, 576))
		}
	}
	icons, sel := out.IconRow()
	return h.toolbarHD(dst, sprites, icons, sel, false, false)
}
func (h *HD) battleHD(dst *ebiten.Image, sprites Sprites, state *battle.State) error {
	s := h.style
	d, _ := state.EquipmentPanel()
	s.begin(dst)
	s.panel(dst, hdRect(96, 12, 736, 450), false)
	s.label(dst, d.Name, hdRect(144, 32, 644, 50), 34, hdInk, false)
	pic, err := memberPortrait(sprites, d.Portrait, d.Picture)
	if err != nil {
		return err
	}
	if i, ok := h.matching[h.fingerprint(pic)]; ok {
		pic = h.portraits[i]
	}
	h.drawHD(dst, pic, hdRect(260, 80, 368, 368))
	s.panel(dst, hdRect(48, 492, 808, 456), false)
	s.panel(dst, hdRect(906, 42, 878, 882), false)
	s.preserve(dst, hdRect(906, 42, 878, 882))
	// 保留已驗部位遮罩與真正指令圖碼。原型的箭頭／健康人形不能當規則資料。
	s.preserve(dst, hdRect(468, 510, 360, 426))
	for _, item := range state.CommandPanel() {
		x, y := item.X*6, (item.Y-8)*6-2
		s.label(dst, s.word(item.Label), hdRect(x, y, 150, 44), 32, hdInk, false)
		s.preserve(dst, hdRect(item.X*6, item.Y*6, 144, 120))
	}
	indices := make([]int, len(state.Plan.Bar))
	for i, code := range state.Plan.Bar {
		indices[i] = code + 3
	}
	return h.toolbarHD(dst, sprites, indices, state.Plan.Selected, true, false)
}

func (h *HD) DrawMusicControl(dst *ebiten.Image, enabled bool) {
	s := h.style
	s.panel(dst, hdRect(1488, 1152, 192, 48), false)
	s.label(dst, MusicControlLabel(enabled), hdRect(1500, 1158, 168, 36), 26, color.RGBA{255, 255, 255, 255}, true)
}

func (h *HD) DrawDisplaySettings(dst *ebiten.Image, state DisplaySettings) {
	s := h.style
	s.panel(dst, hdRect(1704, 1152, 192, 48), false)
	s.label(dst, s.word("Display F10"), hdRect(1716, 1158, 168, 36), 28, hdInk, true)
	if !state.Open {
		return
	}
	vector.DrawFilledRect(dst, 0, 0, 1920, 1200, color.RGBA{0, 0, 0, 140}, false)
	s.panel(dst, hdRect(1144, 352, 744, 712), false)
	s.label(dst, s.word("Display settings"), hdRect(1184, 384, 664, 52), 38, hdInk, false)
	s.label(dst, s.word("Changes preview immediately"), hdRect(1184, 440, 664, 44), 30, hdMuted, false)
	for i := 0; i < 2; i++ {
		r := hdRect(1184, 496+i*168, 664, 148)
		s.panel(dst, r, i == 1 || state.Cursor == i)
		title, detail := s.word("EGA Classic"), s.word("Original pixel artwork")
		if i == 1 {
			title, detail = s.word("HD Painted"), "介面、人物與旅店"
		}
		s.label(dst, title, hdRect(r.Min.X+28, r.Min.Y+18, 420, 42), 34, hdInk, false)
		s.label(dst, detail, hdRect(r.Min.X+28, r.Min.Y+82, 596, 42), 30, hdMuted, false)
		if i == 1 {
			s.label(dst, s.word("In use"), hdRect(r.Min.X+510, r.Min.Y+20, 126, 40), 28, hdGold, true)
		}
	}
	message := "同一場景，保留隊伍與操作位置"
	if state.Problem {
		message = s.word("HD images are unavailable.")
	}
	s.label(dst, message, hdRect(1184, 836, 664, 44), 28, hdMuted, false)
	s.panel(dst, hdRect(1416, 952, 432, 72), state.Cursor == 2)
	s.label(dst, s.word("Back to game  Esc"), hdRect(1432, 966, 400, 42), 32, hdInk, true)
}
