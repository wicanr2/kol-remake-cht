package main

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/wicanr2/kol-remake-cht/internal/assets"
	"github.com/wicanr2/kol-remake-cht/internal/bitmapedit"
	"github.com/wicanr2/kol-remake-cht/internal/flow"
	"github.com/wicanr2/kol-remake-cht/internal/roster"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

// 書本 EDIT PICTURE／EDIT SHIELD 的接線（SPEC-036）：編輯器本身在 internal/bitmapedit，
// 這裡只管從哪裡取圖、改好寫到哪裡，以及這個畫面的輸入。

// memberBitmap 是一名角色目前的圖：改過就是存檔裡那一張；沒改過，圖像是 CHARPICS
// 的模板、盾牌是預設盾牌。兩張 PNG 都只用 EGA 16 色，所以能無損轉回索引。
func memberBitmap(lib *assets.Library, c town.Candidate, shield bool) (bitmapedit.Bitmap, error) {
	custom := c.Picture
	if shield {
		custom = c.Shield
	}
	if custom != nil {
		return bitmapedit.ParseHex(*custom)
	}
	var img image.Image
	var err error
	if shield {
		img, err = lib.DecodeDefaultShield()
	} else {
		img, _, err = lib.Decode("CHARPICS", c.Portrait)
	}
	if err != nil {
		return bitmapedit.Bitmap{}, err
	}
	return bitmapedit.FromImage(img)
}

// attachBitmapEditor 把編輯器的取圖與寫回接到城鎮名冊與存檔。寫回同時改名冊與隊伍裡
// 同名的那一名，再存檔；存檔失敗就整筆退回。
func attachBitmapEditor(f *flow.Flow, lib *assets.Library, s *town.Session) {
	find := func(name string) int {
		for i := range s.Roster {
			if s.Roster[i].Name == name {
				return i
			}
		}
		return -1
	}
	f.LoadBitmap = func(slot roster.Slot, shield bool) (bitmapedit.Bitmap, error) {
		i := find(slot.Name)
		if i < 0 {
			return bitmapedit.Bitmap{}, fmt.Errorf("名冊裡沒有 %s", slot.Name)
		}
		return memberBitmap(lib, s.Roster[i], shield)
	}
	f.SaveBitmap = func(slot roster.Slot, shield bool, b bitmapedit.Bitmap) error {
		i := find(slot.Name)
		if i < 0 {
			return fmt.Errorf("名冊裡沒有 %s", slot.Name)
		}
		hex := b.Hex()
		beforeRoster := s.Roster[i]
		beforeParty := append([]town.Member(nil), s.Party...)
		set := func(pic, shd **string) {
			if shield {
				*shd = &hex
			} else {
				*pic = &hex
			}
		}
		set(&s.Roster[i].Picture, &s.Roster[i].Shield)
		if s.Roster[i].Full != nil {
			m := town.CloneMember(*s.Roster[i].Full)
			set(&m.Picture, &m.Shield)
			s.Roster[i].Full = &m
		}
		for k := range s.Party {
			if s.Party[k].Name == slot.Name {
				set(&s.Party[k].Picture, &s.Party[k].Shield)
			}
		}
		if s.SaveGame != nil {
			if err := s.SaveGame(); err != nil {
				s.Roster[i], s.Party = beforeRoster, beforeParty
				return err
			}
		}
		return nil
	}
}

// editorKeys 是編輯器畫面的鍵盤。方向鍵沿用 `U D L R`，所以字母在這個畫面送小寫；
// 另外多四個斜走鍵與數字鍵盤 5（原版 `sub_146CE` 讀的是延伸鍵掃描碼）。
var editorKeys = []struct {
	key   ebiten.Key
	token string
}{
	{ebiten.KeyArrowUp, "U"}, {ebiten.KeyArrowDown, "D"},
	{ebiten.KeyArrowLeft, "L"}, {ebiten.KeyArrowRight, "R"},
	{ebiten.KeyHome, "HOME"}, {ebiten.KeyPageUp, "PGUP"},
	{ebiten.KeyEnd, "END"}, {ebiten.KeyPageDown, "PGDN"},
	{ebiten.KeyNumpad7, "HOME"}, {ebiten.KeyNumpad9, "PGUP"},
	{ebiten.KeyNumpad1, "END"}, {ebiten.KeyNumpad3, "PGDN"},
	{ebiten.KeyNumpad8, "U"}, {ebiten.KeyNumpad2, "D"},
	{ebiten.KeyNumpad4, "L"}, {ebiten.KeyNumpad6, "R"},
	{ebiten.KeyNumpad5, "5"},
	{ebiten.KeyEnter, "RET"}, {ebiten.KeyNumpadEnter, "RET"},
	{ebiten.KeyEscape, "ESC"},
	{ebiten.KeyComma, "<"}, {ebiten.KeyPeriod, ">"},
	{ebiten.KeyD, "d"}, {ebiten.KeyU, "u"}, {ebiten.KeyX, "x"}, {ebiten.KeyC, "c"},
	{ebiten.KeyY, "y"}, {ebiten.KeyN, "n"},
}

// editorMouse 記住上一次送出的滑鼠位置，拖曳時只在位置變了才再送。
type editorMouse struct {
	held         bool
	lastX, lastY int
}

// editorTokens 讀編輯器畫面這一幀的輸入。左鍵按下送 `CLICK:x,y`（320 × 200 座標）；
// 按住拖曳只在畫布一帶繼續送——原版在那一帶不等放開，其他地方要等放開才算一下。
func (g *game) editorTokens() []string {
	var out []string
	for _, k := range editorKeys {
		if inpututil.IsKeyJustPressed(k.key) {
			out = append(out, k.token)
		}
	}
	x, y := g.cursorPosition()
	if g.strings != nil {
		x, y = x/3, y/3
	}
	pressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	switch {
	case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft):
		out = append(out, fmt.Sprintf("CLICK:%d,%d", x, y))
		g.editMouse = editorMouse{held: true, lastX: x, lastY: y}
	case pressed && g.editMouse.held && (x != g.editMouse.lastX || y != g.editMouse.lastY) && x < 185 && y < 165:
		out = append(out, fmt.Sprintf("CLICK:%d,%d", x, y))
		g.editMouse.lastX, g.editMouse.lastY = x, y
	case !pressed:
		g.editMouse.held = false
	}
	return out
}
