package ui

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/kol-remake-cht/internal/book"
	"github.com/wicanr2/kol-remake-cht/internal/font"
	"github.com/wicanr2/kol-remake-cht/internal/roster"
)

// Roster 畫出角色清單頁：左欄名字、右欄所在地，兩欄都是黑字。
// 座標與標題圖號都由原版 list-characters-run1/frame-065 機械量得。
func Roster(dst *ebiten.Image, sprites Sprites, small *font.Font,
	slots []roster.Slot) error {
	bg, err := sprites.Get("BOOKPICS", book.OpenSpread)
	if err != nil {
		return err
	}
	draw(dst, bg, 0, 0)
	h := book.RosterHeadings
	names, err := sprites.Get("BOOKPICS", h.NamesIndex)
	if err != nil {
		return err
	}
	draw(dst, names, h.NamesX, h.NamesY)
	location, err := sprites.Get("BOOKPICS", h.LocationIndex)
	if err != nil {
		return err
	}
	draw(dst, location, h.LocationX, h.LocationY)
	for i, s := range slots {
		text(dst, small, book.RosterNames.X, book.RosterNames.Y+i*book.RosterNames.Step,
			s.Name, book.ColourValue)
		text(dst, small, book.RosterLocations.X,
			book.RosterLocations.Y+i*book.RosterLocations.Step,
			s.Location(), book.ColourValue)
	}
	return nil
}

// RosterCJK 是角色清單頁的中文模式。角色名字是玩家自取，維持原文；所在地查
// 顯示譯名（SPEC-033），標題與欄名改為中文。版面是 remake 自訂的，不對拍。
func RosterCJK(dst *ebiten.Image, sprites Sprites, f *font.Font, t Translator,
	slots []roster.Slot) error {
	bg, err := sprites.Get("BOOKPICS", book.OpenSpread)
	if err != nil {
		return err
	}
	drawScaled(dst, bg, 0, 0, book.CJKScale)
	l := book.CJKLayout
	text(dst, f, l.HeadingX, l.HeadingY, t.T("Character names"), book.ColourValue)
	text(dst, f, l.RightX, l.HeadingY, t.T("Location"), book.ColourValue)
	for i, s := range slots {
		y := l.RosterY + i*l.RosterStep
		text(dst, f, l.RosterNamesX, y, s.Name, book.ColourValue)
		text(dst, f, l.RosterLocationsX, y, locationCJK(t, s), book.ColourValue)
	}
	return nil
}
