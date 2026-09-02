package town

import (
	"fmt"
	"image"
	"image/color"
)

// 隊伍標記的合成規則由 KOLTOWN 的 `sub_1D21D` 讀出（RE-060）。原版不存遮罩，是算出來的：
//
//  1. 把精靈四個 EGA 平面 OR 起來，得到「這個像素有沒有被精靈用到」的位圖。
//  2. `ax |= (ax << 1) | (ax >> 1)`，左右各膨脹一像素，再取反相當作 AND 遮罩。
//  3. `(背景磚 AND 遮罩) OR 精靈`，結果寫進暫存磚再照一般磚塊畫上去。
//
// 膨脹是在 **little-endian 的 16-bit word** 上做的，而畫面上左半 8 像素是低位元組——
// 所以位移不會跨越位元組邊界，第 7 欄與第 8 欄之間不互相膨脹。少了這一條，
// 原版畫面上會多出一個像素對不上。

// MarkerSide 是一個標記精靈的邊長，MarkerUnits 是 atlas 裡的單元數。
const (
	MarkerSide  = 16
	MarkerUnits = 10
)

// MarkerArchive 是隊伍標記 atlas 在資產 manifest 裡的 archive 名稱。
const MarkerArchive = "KOLPICS-MARKERS"

// MarkerFrames 是城鎮戶外畫面用的畫格數：`word_2DF24` 在 0…4 之間來回。
const MarkerFrames = 5

// Markers 是十個標記精靈的像素。
type Markers struct{ units [][]color.RGBA }

// LoadMarkers 從 16×160 的 atlas 建立標記集合，十個單元由上而下。
func LoadMarkers(img image.Image) (*Markers, error) {
	b := img.Bounds()
	if b.Dx() != MarkerSide || b.Dy() != MarkerSide*MarkerUnits {
		return nil, fmt.Errorf("隊伍標記 atlas 尺寸 %dx%d 不是 %d×%d",
			b.Dx(), b.Dy(), MarkerSide, MarkerSide*MarkerUnits)
	}
	m := &Markers{units: make([][]color.RGBA, MarkerUnits)}
	for unit := 0; unit < MarkerUnits; unit++ {
		px := make([]color.RGBA, MarkerSide*MarkerSide)
		for y := 0; y < MarkerSide; y++ {
			for x := 0; x < MarkerSide; x++ {
				r, g, bl, a := img.At(b.Min.X+x, b.Min.Y+unit*MarkerSide+y).RGBA()
				px[y*MarkerSide+x] = color.RGBA{uint8(r >> 8), uint8(g >> 8),
					uint8(bl >> 8), uint8(a >> 8)}
			}
		}
		m.units[unit] = px
	}
	return m, nil
}

// Unit 回傳第 index 個標記精靈的像素。
func (m *Markers) Unit(index int) []color.RGBA { return m.units[index] }

// black 是 EGA 索引 0 的顏色，也是膨脹出來那一圈外框的顏色。
var black = color.RGBA{0, 0, 0, 255}

// MarkerOverlay 回傳一個 16×16 的覆蓋層：膨脹範圍內取精靈的顏色（精靈沒用到的
// 像素是黑色外框），範圍外是透明。呼叫端只要把非透明的像素蓋上去就完成合成。
func (m *Markers) MarkerOverlay(index int) []color.RGBA {
	src := m.units[index]
	out := make([]color.RGBA, MarkerSide*MarkerSide)
	for y := 0; y < MarkerSide; y++ {
		for x := 0; x < MarkerSide; x++ {
			if !dilated(src, x, y) {
				continue
			}
			if used(src, x, y) {
				out[y*MarkerSide+x] = src[y*MarkerSide+x]
			} else {
				out[y*MarkerSide+x] = black
			}
		}
	}
	return out
}

// used 是「精靈有沒有用到這個像素」：四個平面 OR 起來就是「顏色索引不為 0」。
func used(src []color.RGBA, x, y int) bool {
	c := src[y*MarkerSide+x]
	return c.R != 0 || c.G != 0 || c.B != 0
}

// dilated 是左右各膨脹一像素，但不跨越 8 像素的位元組邊界。
func dilated(src []color.RGBA, x, y int) bool {
	if used(src, x, y) {
		return true
	}
	lo, hi := 0, 7
	if x >= 8 {
		lo, hi = 8, 15
	}
	if x-1 >= lo && used(src, x-1, y) {
		return true
	}
	if x+1 <= hi && used(src, x+1, y) {
		return true
	}
	return false
}
