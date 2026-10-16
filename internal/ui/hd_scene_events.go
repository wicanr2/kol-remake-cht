package ui

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
)

type HDSceneEventArt struct {
	Source HDPageSource
	Size   image.Point
	Paint  image.Image
}
type hdSceneEventArt struct {
	size     image.Point
	paint    *ebiten.Image
	portrait bool
	paper    bool
}
type hdSceneEvents struct{ byRGB map[string]hdSceneEventArt }

func (h *HD) AttachSceneEvents(arts []HDSceneEventArt) error {
	if h.pages == nil || h.style == nil || len(arts) != 33 {
		return fmt.Errorf("HD事件需要完整介面與33張圖片")
	}
	sources := append([]HDPageSource{}, h.pages.sources...)
	seen := map[string]bool{}
	for i, a := range arts {
		archive, index, size, delivery := "TWN0PICS", i, image.Pt(192, 96), image.Pt(768, 384)
		if i >= 25 {
			archive, index = "COMBPICS", i-25
			size, delivery = image.Pt(296, 144), image.Pt(768, 374)
			if index >= 5 {
				size, delivery = image.Pt(88, 64), image.Pt(384, 280)
			}
		}
		if i == 32 {
			archive, index = "MISCPICS", 1
			size, delivery = image.Pt(296, 32), image.Pt(768, 83)
		}
		if a.Source.Archive != archive || a.Source.Index != index || len(a.Source.RGB) != 64 || seen[a.Source.RGB] || a.Size != size || a.Paint == nil || a.Paint.Bounds().Size() != delivery {
			return fmt.Errorf("HD事件來源、尺寸或完整組不符")
		}
		seen[a.Source.RGB] = true
		found := false
		for _, old := range sources {
			if old.Archive == archive && old.Index == index {
				if old.RGB != a.Source.RGB {
					return fmt.Errorf("HD事件覆蓋既有來源")
				}
				found = true
			} else if old.RGB == a.Source.RGB {
				return fmt.Errorf("HD事件來源指紋衝突")
			}
		}
		if !found {
			sources = append(sources, a.Source)
		}
	}
	next := &hdSceneEvents{byRGB: map[string]hdSceneEventArt{}}
	for _, a := range arts {
		next.byRGB[a.Source.RGB] = hdSceneEventArt{a.Size, ebiten.NewImageFromImage(a.Paint), a.Source.Archive == "COMBPICS" && a.Source.Index >= 5, a.Source.Archive == "MISCPICS" && a.Source.Index == 1}
	}
	h.sceneEvents = next
	h.pages.sources = sources
	h.pages.validated = false
	return nil
}
