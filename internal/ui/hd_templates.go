package ui

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

type HDTemplatePortrait struct {
	HDPageSource
	File string
	Rect image.Rectangle
}

// AttachTemplates 保留原核准六圖與優先序，只補齊其餘74個原版指紋。
func (h *HD) AttachTemplates(images map[string]image.Image, preserved []HDPageSource, portraits []HDTemplatePortrait) error {
	if h.pages == nil || len(h.matching) != 6 || len(preserved) != 6 || len(portraits) != 74 || len(images) != 20 {
		return fmt.Errorf("HD角色模板掛接狀態不符")
	}
	for i, p := range preserved {
		if got, ok := h.matching[p.RGB]; !ok || got != i {
			return fmt.Errorf("HD核准六圖的來源或優先序不符")
		}
	}
	var sources []HDPageSource
	for _, p := range preserved {
		sources = append(sources, p)
	}
	for _, p := range portraits {
		if _, ok := h.matching[p.RGB]; ok {
			return fmt.Errorf("HD新增角色覆蓋既有指紋")
		}
		sources = append(sources, p.HDPageSource)
	}
	// 跟書頁共用首次繪圖前驗證，避免RunGame之前呼叫GPU ReadPixels。
	h.pages.sources = append(h.pages.sources, sources...)
	h.pages.validated = false
	atlases := map[string]*ebiten.Image{}
	for name, pic := range images {
		atlases[name] = ebiten.NewImageFromImage(pic)
	}
	h.portraitFit = map[*ebiten.Image]bool{}
	for j, p := range portraits {
		i := j + 6
		h.portraits[i] = atlases[p.File].SubImage(p.Rect).(*ebiten.Image)
		h.matching[p.RGB] = i
		h.portraitFit[h.portraits[i]] = true
	}
	return nil
}

// fitHDPortrait 不裁寬披風或橫向武器；空白留在原頭像矩形內。
func fitHDPortrait(source image.Rectangle, target image.Rectangle) image.Rectangle {
	w, h := target.Dx(), target.Dy()
	if source.Dx()*h > source.Dy()*w {
		h = max(1, source.Dy()*w/source.Dx())
	} else {
		w = max(1, source.Dx()*h/source.Dy())
	}
	x, y := target.Min.X+(target.Dx()-w)/2, target.Min.Y+(target.Dy()-h)/2
	return image.Rect(x, y, x+w, y+h)
}
