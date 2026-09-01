package ui

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/kol-remake-cht/internal/book"
	"github.com/wicanr2/kol-remake-cht/internal/font"
)

// HarborMaster 畫出 HARBOR MASTER 的訊息頁。原版在只安裝一個 region 時不畫標題圖，
// 只在左頁上方寫三行黑字，按任意鍵回到角色選項。
func HarborMaster(dst *ebiten.Image, sprites Sprites, small *font.Font) error {
	bg, err := sprites.Get("BOOKPICS", book.OpenSpread)
	if err != nil {
		return err
	}
	draw(dst, bg, 0, 0)
	for _, l := range book.HarborMasterLines {
		text(dst, small, l.X, l.Y, l.Text, [3]uint8{l.R, l.G, l.B})
	}
	return nil
}

// HarborMasterCJK 是訊息頁的中文模式。
func HarborMasterCJK(dst *ebiten.Image, sprites Sprites, f *font.Font, t Translator) error {
	bg, err := sprites.Get("BOOKPICS", book.OpenSpread)
	if err != nil {
		return err
	}
	drawScaled(dst, bg, 0, 0, book.CJKScale)
	// 原版把訊息斷成兩行是英文字距下的排版；中文模式的頁寬放得下整句，因此併成一行。
	l := book.CJKLayout
	for i, line := range []string{"ONLY ONE REGION IS INSTALLED.", "PRESS ANY KEY."} {
		y := l.RosterY + i*l.RosterStep*2
		text(dst, f, l.FieldX, y, t.T(line), book.ColourValue)
	}
	return nil
}
