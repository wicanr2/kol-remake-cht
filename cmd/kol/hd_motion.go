package main

import (
	"fmt"
	"github.com/wicanr2/kol-remake-cht/internal/assets"
	"github.com/wicanr2/kol-remake-cht/internal/ui"
	"image"
	"path/filepath"
)

func attachHDMotion(h *ui.HD, dir string, lib *assets.Library) error {
	group, e := assets.LoadHDMotionArt(filepath.Join(dir, "motion"), assets.DeferHDPixels)
	if e != nil {
		return fmt.Errorf("HD動畫載入失敗：%w", e)
	}
	if e = group.ValidateSources(lib); e != nil {
		return e
	}
	var arts []ui.HDMotionImage
	for i, p := range group.Manifest.Pictures {
		source, _, e := lib.Decode(p.Archive, p.Index)
		if e != nil {
			return e
		}
		arts = append(arts, ui.HDMotionImage{Name: p.ID, Source: ui.HDPageSource{Archive: p.Archive, Index: p.Index, RGB: p.SourceRGB}, Size: image.Pt(p.Size[0], p.Size[1]), Original: source, Paint: group.Images[i]})
	}
	masks := map[int]image.Image{}
	for _, g := range []int{0, 1, 2, 5} {
		pic, _, e := lib.Decode("WEATPICS", g*3+2)
		if e != nil {
			return e
		}
		masks[g] = pic
	}
	if e := h.AttachMotion(arts, masks); e != nil {
		return e
	}
	var cards []ui.HDPageSource
	for _, c := range group.Manifest.TitleCards {
		cards = append(cards, ui.HDPageSource{Archive: c.Archive, Index: c.Index, RGB: c.SourceRGB})
	}
	return h.AttachTitleCards(cards)
}
