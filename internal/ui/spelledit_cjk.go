package ui

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/kol-remake-cht/internal/book"
	"github.com/wicanr2/kol-remake-cht/internal/font"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

// drawSpellEditCJK 是法術編輯器那一頁的中文模式（SPEC-035）。版面是英文頁
// （`drawSpellEdit`，RE-128）乘上 `book.CJKScale`：同一張卷軸底圖、同樣的透明
// 繪字與分隔線；欄名取 KOLDATA 0x4D 的譯文，欄位值查顯示譯名（SPEC-033），
// 法術名保留原文並在底下加音節注解（SPEC-034）。
//
// 英文名稱輸入框已依RE-212還原；中文頁把選單畫成黑底清單框蓋在卷軸中央，
// 名稱輸入則在名稱欄接一條橫槓游標。這兩者是 remake 的版面，不對拍。
func drawSpellEditCJK(dst *ebiten.Image, sprites Sprites, f *font.Font, t Translator,
	p *town.SpellEditPage) error {
	s := book.CJKScale
	pic, err := sprites.Get(town.SpellEditArchive, town.SpellEditPicture)
	if err != nil {
		return err
	}
	drawScaled(dst, pic, 0, 0, s)

	ink := town.ColourSpellField
	for _, seg := range p.Labels {
		f.Draw(dst, seg.Col*town.CellWidth*s, seg.Row*s, seg.Text, ink[0], ink[1], ink[2])
	}
	for _, r := range p.Rules {
		fill(dst, r.X1*s, r.Y*s, (r.X2-r.X1+1)*s, s, town.ColourSpellRule)
	}
	for i, cell := range p.Fields {
		c := town.ColourSpellField
		if cell.Color == town.SpellCursorColor {
			c = town.ColourSpellCursor
		}
		text := displayName(t, cell.Text)
		x := cell.Col * town.CellWidth * s
		if i == town.SpellFieldName {
			if p.Typing != nil {
				text = *p.Typing + "_"
			}
			// 名字照原版置中的字格算式，但中文模式以實際寬度置中在同一段欄位。
			w, _ := f.Measure(text)
			left := town.SpellNameLeft * town.CellWidth * s
			x = left + (town.SpellNameCols*town.CellWidth*s-w)/2
			if gloss := spellGloss(t, cell.Text); gloss != "" && p.Typing == nil {
				gw, _ := f.Measure(gloss)
				f.Draw(dst, left+(town.SpellNameCols*town.CellWidth*s-gw)/2,
					cell.Row*s+f.Height(), gloss, ink[0], ink[1], ink[2])
			}
		}
		f.Draw(dst, x, cell.Row*s, text, c[0], c[1], c[2])
	}
	// 兩行造價的標籤與內容對調是原版自己的接法（RE-128），中文照樣保留。
	for _, seg := range p.Costs {
		f.Draw(dst, seg.Col*town.CellWidth*s, seg.Row*s, seg.Text, ink[0], ink[1], ink[2])
	}
	if p.Menu != nil {
		// 選單開在被編輯欄位的另一側、欄名列下方：正在改的那一欄與欄名都看得到。
		x := 520
		if p.Cursor >= 0 && p.Cursor < len(p.Fields) && p.Fields[p.Cursor].Col*town.CellWidth*s >= 480 {
			x = 150
		}
		drawSpellMenuCJK(dst, f, t, p.Menu, x, 186)
	}
	return nil
}

// drawSpellMenuCJK 把欄位選單畫成黑底清單框，左上角在 (x, y)。
func drawSpellMenuCJK(dst *ebiten.Image, f *font.Font, t Translator, m *town.InnList, x, y int) {
	if n := activeHDUI; n != nil && n.dst == dst {
		paper := n.paper
		n.paper = false
		defer func() { n.paper = paper }()
	}
	const w, rowH, pad = 280, 18, 12
	h := town.ListRows*rowH + 2*pad
	shadow(dst, x, y, w, h)
	fill(dst, x, y, w, h, town.ColourBoxEdge)
	fill(dst, x+1, y+1, w-2, h-2, town.ColourBoxFill)
	for i, row := range m.Rows {
		if i >= town.ListRows {
			break
		}
		c := town.ColourListIdle
		if i == m.Selected {
			c = town.ColourText
		}
		f.Draw(dst, x+pad, y+pad+i*rowH, displayName(t, row), c[0], c[1], c[2])
	}
	if m.Scrolls() {
		a := town.ColourListArrow
		f.Draw(dst, x+w-24, y+pad, "▲", a[0], a[1], a[2])
		f.Draw(dst, x+w-24, y+pad+(town.ListRows-1)*rowH, "▼", a[0], a[1], a[2])
	}
}
