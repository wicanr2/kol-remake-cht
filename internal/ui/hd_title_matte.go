package ui

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	"math"
)

// Source overlays contain copied background. The union of the six source
// differences defines foreground geometry; closing and filling repair dither
// holes without changing the timing, placement or generated painting files.
func (m *hdMotion) prepareTitleMattes() error {
	for _, g := range []struct {
		name       string
		x, y, w, h int
	}{{"set0pics", 112, 53, 88, 70}, {"set1pics", 96, 53, 104, 84}, {"set2pics", 48, 53, 160, 125}} {
		union := make([]bool, g.w*g.h)
		for phase := 0; phase < 6; phase++ {
			back := 0
			if phase >= 4 {
				back = phase - 3
			}
			source := m.originals[fmt.Sprintf("%s-%d", g.name, phase)]
			background := m.originals[fmt.Sprintf("intro-%02d", back)]
			if source == nil || background == nil || source.Bounds().Size() != image.Pt(g.w, g.h) {
				return fmt.Errorf("HD片頭輪廓來源不符")
			}
			for y := 0; y < g.h; y++ {
				for x := 0; x < g.w; x++ {
					r, gg, b, _ := source.At(x, y).RGBA()
					rr, g2, b2, _ := background.At(g.x+x, g.y+y).RGBA()
					union[y*g.w+x] = union[y*g.w+x] || r != rr || gg != g2 || b != b2
				}
			}
		}
		mask := closeAndFillTitleMask(union, g.w, g.h)
		for phase := 0; phase < 6; phase++ {
			id := fmt.Sprintf("%s-%d", g.name, phase)
			original := m.cpu[id]
			out := image.NewNRGBA(original.Bounds())
			copy(out.Pix, original.Pix)
			for y := 0; y < out.Bounds().Dy(); y++ {
				for x := 0; x < out.Bounds().Dx(); x++ {
					sx := (float64(x)+.5)*float64(g.w)/float64(out.Bounds().Dx()) - .5
					sy := (float64(y)+.5)*float64(g.h)/float64(out.Bounds().Dy()) - .5
					out.Pix[out.PixOffset(x, y)+3] = titleMatteAlpha(mask, g.w, g.h, sx, sy)
				}
			}
			m.images[id] = ebiten.NewImageFromImage(out)
		}
	}
	return nil
}
func closeAndFillTitleMask(src []bool, w, h int) []bool {
	clamp := func(v, n int) int {
		if v < 0 {
			return 0
		}
		if v >= n {
			return n - 1
		}
		return v
	}
	dilate, closed := make([]bool, len(src)), make([]bool, len(src))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					dilate[y*w+x] = dilate[y*w+x] || src[clamp(y+dy, h)*w+clamp(x+dx, w)]
				}
			}
		}
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			closed[y*w+x] = true
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					closed[y*w+x] = closed[y*w+x] && dilate[clamp(y+dy, h)*w+clamp(x+dx, w)]
				}
			}
		}
	}
	outside := make([]bool, len(src))
	var queue []int
	push := func(i int) {
		if !closed[i] && !outside[i] {
			outside[i] = true
			queue = append(queue, i)
		}
	}
	for x := 0; x < w; x++ {
		push(x)
		push((h-1)*w + x)
	}
	for y := 0; y < h; y++ {
		push(y * w)
		push(y*w + w - 1)
	}
	for i := 0; i < len(queue); i++ {
		p := queue[i]
		x, y := p%w, p/w
		if x > 0 {
			push(p - 1)
		}
		if x+1 < w {
			push(p + 1)
		}
		if y > 0 {
			push(p - w)
		}
		if y+1 < h {
			push(p + w)
		}
	}
	for i := range closed {
		closed[i] = !outside[i]
	}
	return closed
}
func titleMatteAlpha(mask []bool, w, h int, x, y float64) uint8 {
	x0, y0 := int(math.Floor(x)), int(math.Floor(y))
	fx, fy := x-float64(x0), y-float64(y0)
	at := func(xx, yy int) float64 {
		if xx < 0 {
			xx = 0
		}
		if yy < 0 {
			yy = 0
		}
		if xx >= w {
			xx = w - 1
		}
		if yy >= h {
			yy = h - 1
		}
		if mask[yy*w+xx] {
			return 255
		}
		return 0
	}
	v := (at(x0, y0)*(1-fx)+at(x0+1, y0)*fx)*(1-fy) + (at(x0, y0+1)*(1-fx)+at(x0+1, y0+1)*fx)*fy
	return uint8(math.Round(v))
}
