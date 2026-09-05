package ui

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/kol-remake-cht/internal/font"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

// drawItemPage 畫資料頁上的物品那一頁（`sub_16986`，RE-127）。畫的順序與原版相同：
// 先把分隔線右邊那一條清成黑，再畫整頁的框與那條線，然後才是清單與明細。
//
// 清單那十七列落在分隔線右邊的窄欄裡（字格欄 32 ＝ x 256），寬度剛好六個字元。
// 沒有反白的列是青色（EGA 索引 3）、選中的是白色（0x0F）。
func drawItemPage(dst *ebiten.Image, f *font.Font, p *town.ItemPage,
	party []town.Member) {
	l := town.ItemPageLayout

	// 原版這一頁是整頁重畫（`sub_1103F(0)` 之後才 `sub_18F3A(0)` 畫立繪），
	// 不是蓋在角色資料頁上，所以框裡先整塊清成黑。**立繪那一層還沒接**——
	// 合成規則已經有了（`internal/portrait`，RE-126），少的是那一頁的背景與外框
	// 沒有原版畫面可以對，所以左半邊目前是空的。
	fill(dst, l.Frame.X0, l.Frame.Y0, l.Frame.Width(), l.Frame.Height(),
		town.ColourBoxFill)
	frame(dst, l.Frame.X0, l.Frame.Y0, l.Frame.Width(), l.Frame.Height(),
		town.ColourSheetFrame)
	fill(dst, l.Divider.X0, l.Divider.Y0,
		l.Divider.Width(), l.Divider.Height(), town.ColourSheetFrame)

	for i, row := range p.Rows {
		colour := town.ColourListIdle
		if i == p.Selected {
			colour = town.ColourText
		}
		cellIn(dst, f, l.ListCol, l.ListRow+i*l.ListStep, row, colour)
	}

	// 明細、空格那幾句與動作做完那一句都畫在左邊那一框裡。
	if p.Detail == nil && len(p.Message) == 0 && len(p.Note) == 0 {
		return
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
		return
	}
	if p.Detail != nil {
		for _, seg := range p.Detail {
			cell(dst, f, seg.Col-sc+dc, seg.Row-sr+dr, seg.Text)
		}
		return
	}
	for i, line := range p.Message {
		cell(dst, f, dc, dr+i*town.LineStep, line)
	}
}

// cellIn 與 cell 相同，但可以指定顏色——清單靠它區分反白與沒反白的列。
func cellIn(dst *ebiten.Image, f *font.Font, col, row int, s string, c [3]uint8) {
	if s == "" {
		return
	}
	x := col * f.Advance()
	fill(dst, x, row, len([]rune(s))*f.Advance(), f.Height(), town.ColourBoxFill)
	f.Draw(dst, x, row, s, c[0], c[1], c[2])
}
