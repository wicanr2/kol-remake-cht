package ui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

type hdSampleKey struct {
	source        *ebiten.Image
	width, height int
}

func (h *HD) sampledHD(src *ebiten.Image, width, height int) *ebiten.Image {
	// Equal-size draws need no resampling. Dynamic weather/body images must
	// retain their latest pixels rather than an immutable cached copy.
	if src.Bounds().Size() == image.Pt(width, height) {
		return src
	}
	key := hdSampleKey{src, width, height}
	if result := h.samples[key]; result != nil {
		return result
	}
	b := src.Bounds()
	raw := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	src.ReadPixels(raw.Pix)
	result := ebiten.NewImageFromImage(resampleHD(raw, width, height))
	if h.samples == nil {
		h.samples = make(map[hdSampleKey]*ebiten.Image)
	}
	h.samples[key] = result
	return result
}

// resampleHD依SPEC-039以固定整數雙線性取樣，避免不同GPU驅動的取樣捨入差。
// 權重對準像素中心；RGBA保持預乘透明度，只在四鄰點合併後捨入一次。
func resampleHD(src *image.RGBA, width, height int) *image.RGBA {
	b := src.Bounds()
	sw, sh := b.Dx(), b.Dy()
	result := image.NewRGBA(image.Rect(0, 0, width, height))
	axis := func(p, n, d int) (int, int, int64, int64) {
		denominator := int64(2 * d)
		numerator := int64((2*p+1)*n - d)
		if numerator <= 0 {
			return 0, 0, 0, denominator
		}
		if numerator >= int64(n-1)*denominator {
			return n - 1, n - 1, 0, denominator
		}
		start := int(numerator / denominator)
		return start, start + 1, numerator % denominator, denominator
	}
	for y := 0; y < height; y++ {
		y0, y1, fy, dy := axis(y, sh, height)
		for x := 0; x < width; x++ {
			x0, x1, fx, dx := axis(x, sw, width)
			divisor := dx * dy
			p00 := src.PixOffset(b.Min.X+x0, b.Min.Y+y0)
			p10 := src.PixOffset(b.Min.X+x1, b.Min.Y+y0)
			p01 := src.PixOffset(b.Min.X+x0, b.Min.Y+y1)
			p11 := src.PixOffset(b.Min.X+x1, b.Min.Y+y1)
			for channel := 0; channel < 4; channel++ {
				top := int64(src.Pix[p00+channel])*(dx-fx) + int64(src.Pix[p10+channel])*fx
				bottom := int64(src.Pix[p01+channel])*(dx-fx) + int64(src.Pix[p11+channel])*fx
				result.Pix[y*result.Stride+x*4+channel] = byte((top*(dy-fy) + bottom*fy + divisor/2) / divisor)
			}
		}
	}
	return result
}
