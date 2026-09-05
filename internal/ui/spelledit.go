package ui

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/kol-remake-cht/internal/font"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

// drawSpellEdit 畫法術編輯器那一頁（`sub_1874E` ＋ `sub_18350`，RE-128）。
//
// 這一頁蓋掉整個畫面：`sub_1103F(0x1D)` 與 `sub_1103F(0)` 走同一條分支——
// 圖 0、y=0、寬 0x28 個 byte 欄、高 0xC8，也就是整張 320×200 的底圖
// （`0x11048`…`0x11072`）。
//
// **這一頁的繪字是透明的**：`word_2DF1C` 在進來時設成 0xFF、離開時歸零，
// 所以字不清底，底色是那張圖本身的白。字的顏色來自工作副本後面那十三個 byte，
// 游標那一欄是 3、其餘 0x0F——在這一頁的調色盤下量出來分別是亮紅與黑。
func drawSpellEdit(dst *ebiten.Image, sprites Sprites, large *font.Font,
	p *town.SpellEditPage) error {
	pic, err := sprites.Get(town.SpellEditArchive, town.SpellEditPicture)
	if err != nil {
		return err
	}
	draw(dst, pic, 0, 0)

	for _, seg := range p.Labels {
		over(dst, large, seg.Col, seg.Row, seg.Text, town.ColourSpellField)
	}
	for _, r := range p.Rules {
		// `sub_1BAFF` 畫的是含兩端的線段，所以寬度要多一個像素。
		fill(dst, r.X1, r.Y, r.X2-r.X1+1, 1, town.ColourSpellRule)
	}
	for _, cell := range p.Fields {
		colour := town.ColourSpellField
		if cell.Color == town.SpellCursorColor {
			colour = town.ColourSpellCursor
		}
		over(dst, large, cell.Col, cell.Row, cell.Text, colour)
	}
	// 兩行造價：上面那一行印的是金冠、下面那一行是冒險點數——標籤與內容對調是
	// 原版自己的接法，見 `town.SpellEditPage`。
	for _, seg := range p.Costs {
		over(dst, large, seg.Col, seg.Row, seg.Text, town.ColourSpellField)
	}
	return nil
}

// over 是透明繪字：不先把字格清成底色，直接畫上去。
func over(dst *ebiten.Image, f *font.Font, col, row int, s string, c [3]uint8) {
	if s == "" {
		return
	}
	f.Draw(dst, col*f.Advance(), row, s, c[0], c[1], c[2])
}
