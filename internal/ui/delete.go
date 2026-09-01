package ui

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/kol-remake-cht/internal/book"
	"github.com/wicanr2/kol-remake-cht/internal/font"
	"github.com/wicanr2/kol-remake-cht/internal/roster"
)

// DeleteList 畫出 DELETE CHARACTER 的畫面。
//
// 版面與角色清單頁相同（同樣的兩個標題圖、同樣的兩欄起點與行距），差別只在顏色語意：
// 反白的那一列是黑字，其餘是青字，兩欄同時反白。確認視窗開著時沒有任何一列反白。
// 三件事都由原版 delete-character-run1 的畫面逐像素量得。
//
// 這張清單是書本四個功能共用的角色選擇畫面：`DELETE CHARACTER`、`EDIT PICTURE`、
// `EDIT SHIELD`、`HARBOR MASTER` 進入後的第一張畫面是同一個 PNG（四份原版收據的
// SHA-256 完全相同），不同的只有選定角色之後走到哪裡。
func DeleteList(dst *ebiten.Image, sprites Sprites, small *font.Font,
	slots []roster.Slot, cursor int, confirm bool) error {
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
		c := book.ColourInactive
		if !confirm && i == cursor {
			c = book.ColourValue
		}
		text(dst, small, book.RosterNames.X,
			book.RosterNames.Y+i*book.RosterNames.Step, s.Name, c)
		text(dst, small, book.RosterLocations.X,
			book.RosterLocations.Y+i*book.RosterLocations.Step, s.Location(), c)
	}
	if confirm {
		drawConfirmWindow(dst, small, book.ConfirmLines)
	}
	return nil
}

// drawConfirmWindow 畫確認視窗：先畫投影，再畫黑框白底，最後畫文字。
// 文字直接畫在視窗白底上，不另清字模格。
func drawConfirmWindow(dst *ebiten.Image, f *font.Font, lines []book.Label) {
	w := book.ConfirmWindow
	// 右側投影：由視窗右緣往右 ShadowThickness px，起點往下偏移 ShadowOffset。
	fill(dst, w.X+w.Width, w.Y+w.ShadowOffset, w.ShadowThickness, w.Height,
		book.ColourValue)
	// 底部投影：由視窗下緣往下 ShadowThickness px，起點往右偏移 ShadowOffset。
	fill(dst, w.X+w.ShadowOffset, w.Y+w.Height, w.Width, w.ShadowThickness,
		book.ColourValue)
	// 外框與白底。
	fill(dst, w.X, w.Y, w.Width, w.Height, book.ColourValue)
	fill(dst, w.X+1, w.Y+1, w.Width-2, w.Height-2,
		[3]uint8{book.PageInk.R, book.PageInk.G, book.PageInk.B})
	for _, l := range lines {
		f.Draw(dst, l.X, l.Y, l.Text, l.R, l.G, l.B)
	}
}

// DeleteListCJK 是刪除畫面的中文模式。角色名字與所在地是原版資料，維持原文；
// 標題、欄名與確認視窗改為中文。版面是 remake 自訂的，不對拍。
func DeleteListCJK(dst *ebiten.Image, sprites Sprites, f *font.Font, t Translator,
	slots []roster.Slot, cursor int, confirm bool) error {
	bg, err := sprites.Get("BOOKPICS", book.OpenSpread)
	if err != nil {
		return err
	}
	drawScaled(dst, bg, 0, 0, book.CJKScale)
	l := book.CJKLayout
	text(dst, f, l.HeadingX, l.HeadingY, t.T("Character names"), book.ColourValue)
	text(dst, f, l.RightX, l.HeadingY, t.T("Location"), book.ColourValue)
	for i, s := range slots {
		c := book.ColourInactive
		if !confirm && i == cursor {
			c = book.ColourValue
		}
		y := l.RosterY + i*l.RosterStep
		text(dst, f, l.RosterNamesX, y, s.Name, c)
		text(dst, f, l.RosterLocationsX, y, s.Location(), c)
	}
	if confirm {
		w := book.CJKConfirm
		fill(dst, w.X+w.Width, w.Y+w.ShadowOffset, w.ShadowThickness, w.Height,
			book.ColourValue)
		fill(dst, w.X+w.ShadowOffset, w.Y+w.Height, w.Width, w.ShadowThickness,
			book.ColourValue)
		fill(dst, w.X, w.Y, w.Width, w.Height, book.ColourValue)
		fill(dst, w.X+1, w.Y+1, w.Width-2, w.Height-2,
			[3]uint8{book.PageInk.R, book.PageInk.G, book.PageInk.B})
		p := book.ColourPrompt
		f.Draw(dst, w.Line1X, w.Line1Y, t.T("ARE YOU SURE?"), p[0], p[1], p[2])
		f.Draw(dst, w.Line2X, w.Line2Y, t.T("YES/NO"), p[0], p[1], p[2])
	}
	return nil
}

// SelectCharacter 是四個書本功能共用的角色選擇清單，等同沒有確認視窗的刪除清單。
func SelectCharacter(dst *ebiten.Image, sprites Sprites, small *font.Font,
	slots []roster.Slot, cursor int) error {
	return DeleteList(dst, sprites, small, slots, cursor, false)
}
