package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/kol-remake-cht/internal/book"
	"github.com/wicanr2/kol-remake-cht/internal/screen"
	"github.com/wicanr2/kol-remake-cht/internal/ui"
	"image"
)

func (g *game) renderHDPage(name string) (*image.RGBA, error) {
	aux := screen.NewCanvasScaled(6)
	dst := aux.Image()
	e := g.hd.RenderUI(dst, func() error {
		switch name {
		case "cover":
			pic, e := g.sprites.Get("BOOKPICS", book.Cover)
			if e != nil {
				return e
			}
			ui.DrawNativePage(dst, pic, 0, 0, 3)
			drawCoverSubtitle(dst, g.cjk, g.titleSubs)
			return nil
		case "contents":
			return ui.ContentsCJK(dst, g.sprites, g.cjk, g.strings)
		case "menu", "menu-selected":
			cursor := -1
			if name == "menu-selected" {
				cursor = g.selected
			}
			return ui.CharacterOptionsCJK(dst, g.sprites, g.cjk, g.strings, cursor)
		case "roster":
			return ui.RosterCJK(dst, g.sprites, g.cjk, g.strings, g.roster.Occupied())
		}
		return fmt.Errorf("未知HD書頁%q", name)
	})
	if e != nil {
		return nil, e
	}
	return aux.Snapshot(), nil
}

func (g *game) renderCJKBookPage(name string) (*image.RGBA, error) {
	aux := screen.NewCanvasScaled(3)
	dst := aux.Image()
	var e error
	switch name {
	case "cover":
		var pic *ebiten.Image
		pic, e = g.sprites.Get("BOOKPICS", book.Cover)
		if e == nil {
			aux.DrawScaled(pic, 0, 0, 3)
			drawCoverSubtitle(dst, g.cjk, g.titleSubs)
		}
	case "contents":
		e = ui.ContentsCJK(dst, g.sprites, g.cjk, g.strings)
	case "menu", "menu-selected":
		cursor := -1
		if name == "menu-selected" {
			cursor = g.selected
		}
		e = ui.CharacterOptionsCJK(dst, g.sprites, g.cjk, g.strings, cursor)
	case "roster":
		e = ui.RosterCJK(dst, g.sprites, g.cjk, g.strings, g.roster.Occupied())
	default:
		e = fmt.Errorf("未知中文書頁%q", name)
	}
	if e != nil {
		return nil, e
	}
	return aux.Snapshot(), nil
}
func (g *game) renderHDBook(dst *ebiten.Image) error {
	if g.hdBookOriginal != g.book || g.hdBook == nil || g.hdBook.Index() > g.book.Index() {
		start, e := g.renderHDPage(g.bookFrom)
		if e != nil {
			return e
		}
		target, e := g.renderHDPage(g.bookTarget)
		if e != nil {
			return e
		}
		animation := *g.bookAnim
		animation.Geometry.FirstRow *= 6
		animation.Geometry.RowCount *= 6
		animation.Geometry.ColumnPixels *= 6
		player, e := book.NewOpenPlayer(&animation, start, target)
		if e != nil {
			return e
		}
		g.hdBook, g.hdBookOriginal = player, g.book
	}
	for g.hdBook.Index() < g.book.Index() {
		if !g.hdBook.Step() {
			return fmt.Errorf("HD翻書步驟與正式播放器不符")
		}
	}
	dst.WritePixels(g.hdBook.Image().Pix)
	return nil
}
