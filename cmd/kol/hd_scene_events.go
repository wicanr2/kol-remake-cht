package main

import (
	"fmt"
	"github.com/wicanr2/kol-remake-cht/internal/assets"
	"github.com/wicanr2/kol-remake-cht/internal/ui"
	"image"
	"path/filepath"
)

func attachHDSceneEvents(hd *ui.HD, dir string, original *assets.Library) error {
	events, e := assets.LoadHDSceneEvents(filepath.Join(dir, "scene-events"))
	if e != nil {
		return fmt.Errorf("HD事件載入失敗：%w", e)
	}
	if e = events.ValidateSources(original); e != nil {
		return e
	}
	var arts []ui.HDSceneEventArt
	for i, p := range events.Manifest.Pictures {
		arts = append(arts, ui.HDSceneEventArt{Source: ui.HDPageSource{Archive: p.Archive, Index: p.Index, RGB: p.SourceRGB}, Size: image.Pt(p.Size[0], p.Size[1]), Paint: events.Images[i]})
	}
	return hd.AttachSceneEvents(arts)
}
