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

// 天候動畫的節拍是 `sub_11390` 野外那一半（`0x11510`…`0x115C6`，RE-164）：野外主迴圈
// 等方向鍵時（`sub_1B100(3)`，`sub_29004:0x29105`）每轉一圈呼叫一次。
//
//	計數（`word_2DF28`）非零 → 等一個 BIOS tick、計數減一
//	計數為零            → 立刻畫一格、計數設 2
//	  雨雪（組 3／4）：畫第 `word_2DF5C` 格，再把它加一（到 3 歸零）
//	  其餘：先把雲層旋轉一格（`sub_2A18E`）再合成
//
// 雨雪的畫格索引整局延續，**換組不歸零**（`sub_2A0C6` 只寫 `word_2DF5A`）；雲層在換組時
// 重新載入，旋轉從 0 起算。地圖重畫（`sub_115D1` 結尾）把計數清零並立刻呼叫一次。
const WeatherWaitTicks = 2

// WeatherCall 是 `sub_11390` 野外那一半的一次呼叫。回傳這一次是不是等了一個 tick。
func (s *Screen) WeatherCall() bool {
	if !s.weatherStarted || s.weatherGroup != s.Group {
		// 換組（或第一次）：雲層重新載入，旋轉歸零；雨雪畫格索引不動。
		s.weatherGroup, s.weatherStarted = s.Group, true
		s.cloudTurns = 0
		if s.Group == 3 || s.Group == 4 {
			s.Phase = s.rainNext
		} else {
			s.Phase = 0
		}
	}
	if s.weatherCounter > 0 {
		s.weatherCounter--
		return true
	}
	s.weatherCounter = WeatherWaitTicks
	if s.Group == 3 || s.Group == 4 {
		s.Phase = s.rainNext
		s.rainNext = (s.rainNext + 1) % 3
	} else {
		s.cloudTurns = (s.cloudTurns + 1) % WeatherWidth
		s.Phase = s.cloudTurns
	}
	return false
}

// WeatherRedraw 是地圖重畫之後的那一次（`sub_115D1` 結尾：計數清零、立刻呼叫）。
func (s *Screen) WeatherRedraw() {
	s.weatherCounter = 0
	s.WeatherCall()
}

// AdvanceWeather 把經過的名義 BIOS ticks 餵給 `WeatherCall`：等的那一次吃掉一個 tick，
// 畫的那一次不花時間。不改動世界時間或亂數；呼叫端只在野外地圖等鍵時餵
// （CAMPING!、提示頁與戰鬥時原版不呼叫）。
func (s *Screen) AdvanceWeather(ticks float64) {
	if ticks <= 0 {
		return
	}
	if !s.weatherStarted {
		s.WeatherRedraw()
	}
	s.weatherTicks += ticks
	for {
		if s.weatherCounter > 0 && s.weatherTicks < 1 {
			return
		}
		if s.WeatherCall() {
			s.weatherTicks--
		}
	}
}

func rgba(c color.Color) color.RGBA {
	r, g, b, a := c.RGBA()
	return color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
}
