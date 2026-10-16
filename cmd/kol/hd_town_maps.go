package main

import (
	"fmt"
	"github.com/wicanr2/kol-remake-cht/internal/assets"
	"github.com/wicanr2/kol-remake-cht/internal/ui"
	"image"
	"path/filepath"
)

// 完整來源與畫稿驗後才接入；呼叫端最後交換整個HD快取。
func attachHDTownMaps(hd *ui.HD, dir string, original *assets.Library) error {
	maps, e := assets.LoadHDTownMaps(filepath.Join(dir, "town-maps"))
	if e != nil {
		return fmt.Errorf("HD城鎮地圖載入失敗：%w", e)
	}
	if e = maps.ValidateSources(original); e != nil {
		return e
	}
	var sources []ui.HDTownMapSource
	var arts []ui.HDTownMapArt
	for _, s := range maps.Manifest.Maps {
		sources = append(sources, ui.HDTownMapSource{GridSource: ui.HDPageSource{Archive: s.Archive, Index: 0, RGB: s.Grid.SourceRGB}, AtlasSource: ui.HDPageSource{Archive: s.Archive, Index: 1, RGB: s.Atlas.SourceRGB}, Grid: maps.Grids[s.Archive]})
	}
	for i, r := range maps.Manifest.Regions {
		a, b := r.SourceBounds, r.VisibleBounds
		arts = append(arts, ui.HDTownMapArt{Archive: r.Archive, SourceBounds: image.Rect(a[0], a[1], a[2], a[3]), VisibleBounds: image.Rect(b[0], b[1], b[2], b[3]), Paint: maps.Images[i]})
	}
	return hd.AttachTownMaps(sources, arts)
}
