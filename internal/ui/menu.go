package ui

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/kol-remake-cht/internal/book"
	"github.com/wicanr2/kol-remake-cht/internal/font"
)

// CharacterOptions 畫出角色選項頁。反白項為黑色，其餘為青色，與書中其他清單同一套
// 顏色語意；座標與文字都由原版畫面機械量得。
func CharacterOptions(dst *ebiten.Image, sprites Sprites, small *font.Font,
	selected int) error {
	bg, err := sprites.Get("BOOKPICS", book.OpenSpread)
	if err != nil {
		return err
	}
	draw(dst, bg, 0, 0)
	heading, err := sprites.Get("BOOKPICS", book.CharacterOptionsHeading.ImageIndex)
	if err != nil {
		return err
	}
	draw(dst, heading, book.CharacterOptionsHeading.X, book.CharacterOptionsHeading.Y)
	for i, item := range book.CharacterOptionsItems {
		c := book.ColourInactive
		if i == selected {
			c = book.ColourValue
		}
		text(dst, small, book.CharacterOptionsList.X,
			book.CharacterOptionsList.Y+i*book.CharacterOptionsList.Step, item, c)
	}
	return nil
}

// CharacterOptionsCJK 是角色選項頁的中文模式。版面是 remake 自訂的，不對拍。
func CharacterOptionsCJK(dst *ebiten.Image, sprites Sprites, f *font.Font,
	t Translator, selected int) error {
	bg, err := sprites.Get("BOOKPICS", book.OpenSpread)
	if err != nil {
		return err
	}
	drawScaled(dst, bg, 0, 0, book.CJKScale)
	l := book.CJKLayout
	text(dst, f, l.HeadingX, l.HeadingY, t.T("Character options"), book.ColourValue)
	for i, item := range book.CharacterOptionsItems {
		c := book.ColourInactive
		if i == selected {
			c = book.ColourValue
		}
		text(dst, f, l.MenuX, l.MenuY+i*l.MenuStep, t.T(item), c)
	}
	return nil
}
