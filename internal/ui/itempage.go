package ui

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/kol-remake-cht/internal/portrait"

	"github.com/wicanr2/kol-remake-cht/internal/font"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

// drawItemPage 畫資料頁上的物品那一頁（`sub_16986`，RE-127）。畫的順序與原版相同：
// 先把分隔線右邊那一條清成黑，再畫整頁的框與那條線，然後才是清單與明細。
//
// 清單那十七列落在分隔線右邊的窄欄裡（字格欄 32 ＝ x 256），寬度剛好六個字元。
// 沒有反白的列是青色（EGA 索引 3）、選中的是白色（0x0F）。
func drawItemPage(dst *ebiten.Image, sprites Sprites, small, large *font.Font,
	p *town.ItemPage, party []town.Member) error {
	l := town.ItemPageLayout

	// 畫的順序照原版：整頁清掉 → 背景 → 立繪 → 右邊那條窄欄清成黑 → 框與分隔線
	// → 清單（`sub_1103F(0)`、`(0x21)`／`(0x64)`、`sub_18F3A(0)`、`sub_1B779`、
	// `sub_1B6DC`、`sub_1BB57`、`sub_196A9`）。
	fill(dst, l.Frame.X0, l.Frame.Y0, l.Frame.Width(), l.Frame.Height(),
		town.ColourBoxFill)
	if err := drawSheetPageBack(dst, sprites); err != nil {
		return err
	}

	// 左邊那個人：一層一層貼上去，順序就是原版畫的順序（RE-126）。
	// X 與寬的單位是 byte 欄，所以乘八才是像素。
	for _, piece := range p.Figure {
		pic, err := sprites.Get(piece.Archive, piece.Index)
		if err != nil {
			return err
		}
		draw(dst, pic, piece.X*PixelsPerByte, piece.Y)
	}

	fill(dst, l.RightPane.X0, l.RightPane.Y0,
		l.RightPane.Width(), l.RightPane.Height(), town.ColourBoxFill)
	frame(dst, l.Frame.X0, l.Frame.Y0, l.Frame.Width(), l.Frame.Height(),
		town.ColourSheetFrame)
	fill(dst, l.Divider.X0, l.Divider.Y0,
		l.Divider.Width(), l.Divider.Height(), town.ColourSheetFrame)

	for i, row := range p.Rows {
		colour := town.ColourListIdle
		if i == p.Selected {
			colour = town.ColourText
		}
		cellIn(dst, large, l.ListCol, l.ListRow+i*l.ListStep, row, colour)
	}

	// 清單移動時左上角那個小框（`sub_16D0B`）：三行——部位名稱加冒號、那件東西的
	// 名字、骰式與重量。重量靠右，位數少就往右挪。
	if m := p.Mark; m != nil {
		fill(dst, m.X+1, m.Y+1, town.MarkWidth-1, town.MarkHeight-1,
			town.ColourBoxFill)
		frame(dst, m.X, m.Y, town.MarkWidth+1, town.MarkHeight+1,
			town.ColourSheetHeader)
		col := m.X/small.Advance() + 1
		row := m.Y + town.MarkTextTop
		cell(dst, small, col, row, m.Slot)
		if m.Name != "" {
			cell(dst, small, col, row+town.MarkLineStep, m.Name)
			third := row + 2*town.MarkLineStep
			cell(dst, small, col, third, m.Dice)
			if m.Skill != "" {
				cell(dst, small, col+8, third, m.Skill)
			}
			w := col + town.MarkWeightCol - 1
			if m.Weight < 100 {
				w++
			}
			if m.Weight < 10 {
				w++
			}
			cell(dst, small, w, third, fmt.Sprintf("%d", m.Weight))
		}
	}

	// 明細、空格那幾句與動作做完那一句都畫在左邊那一框裡。
	if p.Detail == nil && len(p.Message) == 0 && len(p.Note) == 0 {
		return nil
	}
	// 明細框是 `sub_1BA9D` 畫的：填黑加一圈白，**沒有投影**（那是 `sub_1B7E9`，
	// 這一頁沒有呼叫它）。它的右緣壓在清單左緣上，原版也是這樣。
	fill(dst, l.Detail.X0+1, l.Detail.Y0+1,
		l.Detail.Width()-2, l.Detail.Height()-2, town.ColourBoxFill)
	frame(dst, l.Detail.X0, l.Detail.Y0, l.Detail.Width(), l.Detail.Height(),
		town.ColourSheetHeader)
	// 明細那幾段的座標是商店那一頁算好的（`sub_153FF`，RE-075），而這一頁的框在
	// 別的位置——`sub_153FF` 的模式 8 自己怎麼擺還沒讀。這裡照兩個框「內縮兩個
	// 像素」的同一個慣例平移過去（商店的框在 (102,84)、文字從 (104,86) 起），
	// 等模式 8 讀出來再換成它自己的座標。證據等級：強推論。
	f := small
	dc := (l.Detail.X0 + 2) / f.Advance()
	dr := l.Detail.Y0 + 2
	sc := (town.DialogBox.TextX) / f.Advance()
	sr := town.DialogBox.TextY
	// 動作做完那一句蓋在明細上面（原版是整頁重畫再印，效果相同）。
	if len(p.Note) > 0 {
		for i, line := range p.Note {
			cell(dst, f, dc, dr+i*town.LineStep, line)
		}
		if p.Giving {
			for i, m := range party {
				cell(dst, f, dc, dr+(i+2)*town.LineStep,
					fmt.Sprintf("%d %s", i+1, m.Name))
			}
		}
		return nil
	}
	if p.Detail != nil {
		for _, seg := range p.Detail {
			cell(dst, f, seg.Col-sc+dc, seg.Row-sr+dr, seg.Text)
		}
		return nil
	}
	for i, line := range p.Message {
		cell(dst, f, dc, dr+i*town.LineStep, line)
	}
	return nil
}

// PixelsPerByte 是 EGA 平面模式一個 byte 橫跨幾個像素。立繪的版面表用 byte 欄
// 當水平單位（RE-126）。
const PixelsPerByte = 8

// cellIn 與 cell 相同，但可以指定顏色——清單靠它區分反白與沒反白的列。
func cellIn(dst *ebiten.Image, f *font.Font, col, row int, s string, c [3]uint8) {
	if s == "" {
		return
	}
	x := col * f.Advance()
	fill(dst, x, row, len([]rune(s))*f.Advance(), f.Height(), town.ColourBoxFill)
	f.Draw(dst, x, row, s, c[0], c[1], c[2])
}

// drawPartyPage 畫隊伍立繪那一頁（`sub_18D61`，RE-126）：整頁清成黑、畫一圈框，
// 再把六個人由後往前貼上去。
func drawPartyPage(dst *ebiten.Image, sprites Sprites, pieces []portrait.Piece) error {
	r := town.PartyPageFrame
	fill(dst, r.X0, r.Y0, r.Width(), r.Height(), town.ColourBoxFill)
	if err := drawSheetPageBack(dst, sprites); err != nil {
		return err
	}
	frame(dst, r.X0, r.Y0, r.Width(), r.Height(), town.ColourSheetFrame)
	for _, piece := range pieces {
		pic, err := sprites.Get(piece.Archive, piece.Index)
		if err != nil {
			return err
		}
		draw(dst, pic, piece.X*PixelsPerByte, piece.Y)
	}
	return nil
}

// drawSheetPageBack 畫資料頁那兩頁的背景（`sub_1103F(0x21)` 與 `(0x64)`，RE-127）：
// 一間木屋的內裝，上下兩張各貼一次。
func drawSheetPageBack(dst *ebiten.Image, sprites Sprites) error {
	b := town.SheetPageBack
	for _, c := range [2]struct{ index, y int }{{b.TopIndex, b.TopY}, {b.LowIndex, b.LowY}} {
		pic, err := sprites.Get(b.Archive, c.index)
		if err != nil {
			return err
		}
		draw(dst, pic, b.X, c.y)
	}
	return nil
}
