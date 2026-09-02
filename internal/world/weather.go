package world

import (
	"fmt"
	"image"
	"image/color"
)

// 天候面板（RE-068）。`WEATPICS` 每三個 payload 是一組：底圖、覆蓋層與遮罩。
// `sub_2A18E:0x2A1DE` 的合成式是 `(底圖 AND 遮罩) OR 覆蓋層`；抽取出來的遮罩
// 只有 0 與 15 兩種值，而且**遮罩為 0 的像素覆蓋層一定非零、遮罩為 15 的一定為零**
// （64×62 共 3,968 個像素逐一驗過），因此合成等價於一張蠟版：
// 遮罩是 0 就畫覆蓋層，否則畫底圖。
//
// 同一段程式在合成之前先把覆蓋層**整列向右旋轉一個像素**（`rcr` 帶進位、繞回），
// 雲就是這樣飄的。
const (
	// WeatherWidth／WeatherHeight 是天候圖的尺寸。
	WeatherWidth  = 64
	WeatherHeight = 62
	// WeatherGroups 是六組天候／時段背景。
	WeatherGroups = 6
	// WeatherArchive 是天候圖的 archive 名稱。
	WeatherArchive = "WEATPICS"
)

// WeatherLayers 是一組天候圖的三個圖層索引：底圖、覆蓋層、遮罩。
func WeatherLayers(group int) (base, overlay, mask int) {
	g := group * 3
	return g, g + 1, g + 2
}

// Weather 是一組已經展開成像素的天候圖。
type Weather struct {
	base    []color.RGBA
	overlay []color.RGBA
	stencil []bool // true 代表這個像素要畫覆蓋層
}

// NewWeather 把三張圖層變成可以逐畫格合成的資料。
func NewWeather(base, overlay, mask image.Image) (*Weather, error) {
	for _, img := range []image.Image{base, overlay, mask} {
		b := img.Bounds()
		if b.Dx() != WeatherWidth || b.Dy() != WeatherHeight {
			return nil, fmt.Errorf("天候圖層尺寸 %dx%d 不是 %d×%d",
				b.Dx(), b.Dy(), WeatherWidth, WeatherHeight)
		}
	}
	w := &Weather{
		base:    make([]color.RGBA, WeatherWidth*WeatherHeight),
		overlay: make([]color.RGBA, WeatherWidth*WeatherHeight),
		stencil: make([]bool, WeatherWidth*WeatherHeight),
	}
	bb, ob, mb := base.Bounds(), overlay.Bounds(), mask.Bounds()
	for y := 0; y < WeatherHeight; y++ {
		for x := 0; x < WeatherWidth; x++ {
			i := y*WeatherWidth + x
			w.base[i] = rgba(base.At(bb.Min.X+x, bb.Min.Y+y))
			w.overlay[i] = rgba(overlay.At(ob.Min.X+x, ob.Min.Y+y))
			r, g, b, _ := mask.At(mb.Min.X+x, mb.Min.Y+y).RGBA()
			w.stencil[i] = r == 0 && g == 0 && b == 0
		}
	}
	return w, nil
}

// Composite 回傳把覆蓋層向右旋轉 phase 個像素之後的畫面。
func (w *Weather) Composite(phase int) []color.RGBA {
	out := make([]color.RGBA, len(w.base))
	copy(out, w.base)
	shift := ((phase % WeatherWidth) + WeatherWidth) % WeatherWidth
	for y := 0; y < WeatherHeight; y++ {
		row := y * WeatherWidth
		for x := 0; x < WeatherWidth; x++ {
			src := row + (x-shift+WeatherWidth)%WeatherWidth
			if w.stencil[src] {
				out[row+x] = w.overlay[src]
			}
		}
	}
	return out
}

func rgba(c color.Color) color.RGBA {
	r, g, b, a := c.RGBA()
	return color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
}
