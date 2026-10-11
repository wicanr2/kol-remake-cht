package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	"strconv"

	"github.com/wicanr2/kol-remake-cht/internal/book"
	"github.com/wicanr2/kol-remake-cht/internal/font"
)

// ContentsCJK 以空白書頁取代烘在右頁圖片內的英文，沿用七個選項的順序及編號。
// 左頁的 Origin 標誌與版權原樣保留。中文安全矩形沿 CJKTextBox.Right。
func ContentsCJK(dst *ebiten.Image, sprites Sprites, f *font.Font, t Translator) error {
	bg, err := sprites.Get("BOOKPICS", book.Contents)
	if err != nil {
		return err
	}
	drawScaled(dst, bg, 0, 0, book.CJKScale)
	blank, err := sprites.Get("BOOKPICS", book.OpenSpread)
	if err != nil {
		return err
	}
	const pageX = 160
	right := blank.SubImage(image.Rect(pageX, 0, 320, 200)).(*ebiten.Image)
	drawScaled(dst, right, pageX*book.CJKScale, 0, book.CJKScale)
	l := book.CJKTextBox
	text(dst, f, l.RightX, l.RightY, t.T("Contents"), book.ColourValue)
	items := []string{"Credits", "Foreword", "Dedication", "Install new region", "Character options", "Play the game", "Run the intro"}
	for i, item := range items {
		y := l.RightY + 60 + i*54
		text(dst, f, l.RightX, y, t.T(item), book.ColourValue)
		number := strconv.Itoa(i + 1)
		width, _ := f.Measure(number)
		// 紙張右緣在原版 x=303，目錄數字留在 x=294 以內，避開裝飾邊。
		text(dst, f, 294*book.CJKScale-width, y, number, book.ColourValue)
	}
	return nil
}

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
