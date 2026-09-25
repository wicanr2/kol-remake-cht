package ui

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/kol-remake-cht/internal/book"
	"github.com/wicanr2/kol-remake-cht/internal/font"
	"github.com/wicanr2/kol-remake-cht/internal/portrait"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

// drawItemPageCJK 畫中文模式的物品那一頁（RE-127／SPEC-030）。
//
// 與 `drawItemPage` 用**同一份版面**（`town.ItemPageLayout`），每一個座標乘上
// `book.CJKScale`。右側那張清單三倍之後剛好放得下十七列（最後一列底 462、右欄底
// 465），欄寬容得下九個中文字。
//
// 這一頁不對拍，原版沒有中文畫面。
func drawItemPageCJK(dst *ebiten.Image, sprites Sprites, f *font.Font, t Translator,
	p *town.ItemPage) error {
	l := town.ItemPageLayout
	s := book.CJKScale
	interactionBase := itemPageUsesInteractionBase(p)

	// 背景與背板：整頁美術，放大同樣的倍率。
	bg, err := sprites.Get(town.Background.Archive, town.Background.Index)
	if err != nil {
		return err
	}
	drawScaled(dst, bg, 0, 0, s)
	if !interactionBase {
		b := town.SheetPageBack
		for _, c := range [2]struct{ index, y int }{{b.TopIndex, b.TopY}, {b.LowIndex, b.LowY}} {
			pic, err := sprites.Get(b.Archive, c.index)
			if err != nil {
				return err
			}
			drawScaled(dst, pic, b.X*s, c.y*s, s)
		}

		right := scaleRect(l.RightPane, s)
		fill(dst, right.X0, right.Y0, right.Width(), right.Height(), town.ColourBoxFill)
		fr := scaleRect(l.Frame, s)
		frame(dst, fr.X0, fr.Y0, fr.Width(), fr.Height(), town.ColourSheetFrame)
		div := scaleRect(l.Divider, s)
		fill(dst, div.X0, div.Y0, div.Width(), div.Height(), town.ColourSheetFrame)
		shadow(dst, fr.X0, fr.Y0, fr.Width(), fr.Height())

		// 槽位清單。名稱走譯文，取不到就留原文——這幾個是原版的欄位名，
		// 缺一格比混排更難看出是哪一格。
		for i, row := range p.Rows {
			colour := town.ColourListIdle
			if i == p.Selected {
				colour = town.ColourText
			}
			y := (l.ListRow + i*l.ListStep) * s
			text := t.T(row)
			cellInAt(dst, f, l.ListCol*town.CellWidth*s, y, text, colour)
		}

		// 選中那一格的小框：部位名、物品名、骰式與重量。
		if m := p.Mark; m != nil {
			x, y := m.X*s, m.Y*s
			w, h := (town.MarkWidth+1)*s, (town.MarkHeight+1)*s
			fill(dst, x+1, y+1, w-2, h-2, town.ColourBoxFill)
			frame(dst, x, y, w, h, town.ColourSheetHeader)
			shadow(dst, x, y, w, h)
			tx := x + 2*s
			ty := y + town.MarkTextTop*s
			step := town.MarkLineStep * s
			cellCJKAt(dst, f, tx, ty, t.T(m.Slot))
			if m.Name != "" {
				cellCJKAt(dst, f, tx, ty+step, displayName(t, m.Name))
				third := ty + 2*step
				cellCJKAt(dst, f, tx, third, m.Dice)
				if m.Skill != "" {
					cellCJKAt(dst, f, tx+8*f.Advance(), third, m.Skill)
				}
				weight := fmt.Sprintf("%d", m.Weight)
				ww, _ := f.Measure(weight)
				cellCJKAt(dst, f, x+w-2*s-ww, third, weight)
			}
		}
	}

	if p.Detail == nil && len(p.Message) == 0 && len(p.Note) == 0 {
		return nil
	}
	// 明細框：填黑、白框，再畫 `sub_1B7E9` 對應的右／下四像素投影。
	de := scaleRect(l.Detail, s)
	fill(dst, de.X0+1, de.Y0+1, de.Width()-2, de.Height()-2, town.ColourBoxFill)
	frame(dst, de.X0, de.Y0, de.Width(), de.Height(), town.ColourSheetHeader)
	shadow(dst, de.X0, de.Y0, de.Width(), de.Height())
	// 對應英文版 `sub_153FF` 模式 8 的文字基準；中文只放大這個已量得的
	// 英文版面，原版沒有中文畫面可作像素對拍。
	dx, dy := de.X0+2*s, (l.Detail.Y0-8)*s
	step := town.LineStep * s
	// 動作做完那一句蓋在明細上面（原版是整頁重畫再印，效果相同）。
	if len(p.Note) > 0 {
		for i, line := range p.Note {
			cellCJKAt(dst, f, dx, dy+i*step, t.T(line))
		}
		return nil
	}
	// 明細那幾段自帶座標，但那是商店那一頁的框算出來的；這一頁的框在別的位置，
	// 所以照英文版同一個平移換算過來（見 `drawItemPage` 的說明）。
	if p.Detail != nil {
		sc, sr := town.DialogBox.TextX, town.DialogBox.TextY
		for _, seg := range p.Detail {
			x := dx + (seg.Col*town.CellWidth-sc)*s
			y := dy + (seg.Row-sr)*s
			cellCJKAt(dst, f, x, y, displayName(t, t.T(seg.Text)))
		}
		return nil
	}
	for i, line := range p.Message {
		cellCJKAt(dst, f, dx, dy+i*step, t.T(line))
	}
	return nil
}

// cellInAt 以像素座標畫一段文字，可指定顏色。
func cellInAt(dst *ebiten.Image, f *font.Font, x, y int, text string, c [3]uint8) {
	if text == "" {
		return
	}
	w, h := f.Measure(text)
	fill(dst, x, y, w, h, town.ColourBoxFill)
	f.Draw(dst, x, y, text, c[0], c[1], c[2])
}

// drawPartyPageCJK 畫中文模式的隊伍立繪那一頁（RE-126）。這一頁沒有文字，
// 與英文版是同一份版面乘上倍率。
func drawPartyPageCJK(dst *ebiten.Image, sprites Sprites, pieces []portrait.Piece) error {
	s := book.CJKScale
	bg, err := sprites.Get(town.Background.Archive, town.Background.Index)
	if err != nil {
		return err
	}
	drawScaled(dst, bg, 0, 0, s)
	b := town.SheetPageBack
	for _, c := range [2]struct{ index, y int }{{b.TopIndex, b.TopY}, {b.LowIndex, b.LowY}} {
		pic, err := sprites.Get(b.Archive, c.index)
		if err != nil {
			return err
		}
		drawScaled(dst, pic, b.X*s, c.y*s, s)
	}
	r := scaleRect(town.PartyPageFrame, s)
	frame(dst, r.X0, r.Y0, r.Width(), r.Height(), town.ColourSheetFrame)
	shadow(dst, r.X0, r.Y0, r.Width(), r.Height())
	for _, piece := range pieces {
		pic, err := sprites.Get(piece.Archive, piece.Index)
		if err != nil {
			return err
		}
		drawScaled(dst, pic, piece.X*PixelsPerByte*s, piece.Y*s, s)
	}
	return nil
}
