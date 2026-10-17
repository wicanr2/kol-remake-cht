package ui

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/kol-remake-cht/internal/ending"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

func (h *HD) AttachEnding(images map[int]image.Image, sourceRGB map[int]string) {
	h.endingImages = make(map[int]*ebiten.Image, len(images))
	h.endingKeys = make(map[int]string, len(sourceRGB))
	for i, im := range images {
		h.endingImages[i] = h.newPaint(im)
		h.endingKeys[i] = sourceRGB[i]
	}
}

// Ending 沿同一結局狀態與文字座標呈現七段，不更動事件或保存。
func (h *HD) Ending(dst *ebiten.Image, sprites Sprites, state *ending.State, segments []town.Segment) error {
	if state == nil || state.Current() == nil {
		return nil
	}
	p := state.Current()
	art := h.endingImages[p.Picture]
	if art == nil || h.style == nil {
		return fmt.Errorf("HD結局第%d段素材未載入", p.Picture)
	}
	source, err := sprites.Get(ending.Archive, p.Picture)
	if err != nil {
		return err
	}
	if h.fingerprint(source) != h.endingKeys[p.Picture] {
		return fmt.Errorf("HD結局第%d段來源圖像不符", p.Picture)
	}
	dst.Fill(color.RGBA{50, 24, 27, 255})
	vector.StrokeRect(dst, 18, 18, 1884, 1164, 2, hdGold, false)
	b := p.Box
	r := image.Rect(b.X()*6, b.Y()*6, (b.X()+b.Width())*6, (b.Y()+b.Height())*6)
	vector.DrawFilledRect(dst, float32(r.Min.X-4), float32(r.Min.Y-4), float32(r.Dx()+8), float32(r.Dy()+8), hdGold, false)
	vector.DrawFilledRect(dst, float32(r.Min.X), float32(r.Min.Y), float32(r.Dx()), float32(r.Dy()), color.RGBA{15, 12, 11, 255}, false)
	w, height := r.Dx(), r.Dx()*art.Bounds().Dy()/art.Bounds().Dx()
	if height > r.Dy() {
		height, w = r.Dy(), r.Dy()*art.Bounds().Dx()/art.Bounds().Dy()
	}
	x, y := r.Min.X+(r.Dx()-w)/2, r.Min.Y+(r.Dy()-height)/2
	h.drawHD(dst, art, image.Rect(x, y, x+w, y+height))
	for _, seg := range segments {
		text := strings.ToUpper(seg.Text)
		if strings.TrimSpace(text) == "" {
			continue
		}
		x, y := seg.Col*48, seg.Row*6
		height := 40
		for _, next := range segments {
			if step := (next.Row - seg.Row) * 6; step > 0 && step < height {
				height = step
			}
		}
		safe := image.Rect(x, y, x+int(h.style.measure(text, 32)+1), y+height)
		if !safe.In(dst.Bounds()) || safe.Overlaps(r) {
			return fmt.Errorf("HD結局第%d段文字超出安全區：%q", p.Picture, text)
		}
		for _, ch := range text {
			if _, ok := h.style.glyphs[string(ch)]; !ok {
				return fmt.Errorf("HD結局缺字：%c", ch)
			}
		}
		h.style.label(dst, text, safe, 32, hdInk, false)
	}
	return nil
}
