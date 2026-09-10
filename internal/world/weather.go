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
// SPEC-026 訂正：雲層邏輯列寬為 128，只有畫面上方 31 列合成；下半保留底圖。
// 雨雪三張均為完整畫格，不能當作三層。
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
	frames  [][]color.RGBA
	groups  []*Weather
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
	if len(w.frames) != 0 {
		return append([]color.RGBA(nil), w.frames[((phase%3)+3)%3]...)
	}
	const width = WeatherWidth * 2
	shift := ((phase % width) + width) % width
	for y := 0; y < WeatherHeight/2; y++ {
		row := y * WeatherWidth
		for x := 0; x < WeatherWidth; x++ {
			src := y*width + (x-shift+width)%width
			if w.stencil[src] {
				out[row+x] = w.overlay[src]
			}
		}
	}
	return out
}

// LoadWeatherGroups 預載六組；雨雪三張是畫格，其餘才是底圖、雲層與遮罩。
func LoadWeatherGroups(decode func(int) (image.Image, error)) (*Weather, error) {
	all := &Weather{groups: make([]*Weather, WeatherGroups)}
	for group := range all.groups {
		var layers [3]image.Image
		for i := range layers {
			img, err := decode(group*3 + i)
			if err != nil {
				return nil, err
			}
			layers[i] = img
		}
		w, err := NewWeather(layers[0], layers[1], layers[2])
		if err != nil {
			return nil, err
		}
		if group == 3 || group == 4 {
			w.frames = make([][]color.RGBA, 3)
			for i, img := range layers {
				b := img.Bounds()
				for y := 0; y < WeatherHeight; y++ {
					for x := 0; x < WeatherWidth; x++ {
						w.frames[i] = append(w.frames[i], rgba(img.At(b.Min.X+x, b.Min.Y+y)))
					}
				}
			}
		}
		all.groups[group] = w
	}
	return all, nil
}

// CompositeGroup 使天候值的改變同時切換實際圖層。
func (w *Weather) CompositeGroup(group, phase int) []color.RGBA {
	if len(w.groups) == 0 {
		return w.Composite(phase)
	}
	if group < 0 || group >= len(w.groups) {
		return nil
	}
	return w.groups[group].Composite(phase)
}

// AdvanceWeather 只累積動畫用的名義 BIOS ticks，不改動世界時間或亂數。
func (s *Screen) AdvanceWeather(ticks float64) {
	if ticks <= 0 {
		return
	}
	if s.weatherStarted && s.weatherGroup != s.Group {
		s.Phase, s.weatherTicks = 0, 0
	}
	s.weatherGroup, s.weatherStarted = s.Group, true
	s.weatherTicks += ticks
	for s.weatherTicks >= 2 {
		s.weatherTicks -= 2
		s.Phase++
		period := 128
		if s.Group == 3 || s.Group == 4 {
			period = 3
		}
		s.Phase %= period
	}
}

func rgba(c color.Color) color.RGBA {
	r, g, b, a := c.RGBA()
	return color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
}
