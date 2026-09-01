// Package screen 提供固定 320×200 的原版解析度畫布。
//
// remake 一律先畫進這塊畫布再放大到視窗，對拍才能直接和原版 EGA 畫面逐像素比較，
// 不必處理縮放造成的插值差異。
package screen

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
)

// 原版 BIOS mode 0Dh 的可見解析度。
const (
	Width  = 320
	Height = 200
)

// Canvas 是原版解析度的離屏畫布。
type Canvas struct {
	img *ebiten.Image
}

func NewCanvas() *Canvas {
	return &Canvas{img: ebiten.NewImage(Width, Height)}
}

// NewCanvasScaled 建立放大 n 倍的畫布。中文模式需要更大的字模格，因此在放大後的
// 畫布上作畫；英文嚴格對拍模式仍用原版解析度。
func NewCanvasScaled(scale int) *Canvas {
	if scale < 1 {
		scale = 1
	}
	return &Canvas{img: ebiten.NewImage(Width*scale, Height*scale)}
}

// DrawScaled 把素材放大 n 倍後畫到指定座標，採最近鄰，不產生新的中間色。
func (c *Canvas) DrawScaled(src *ebiten.Image, x, y, scale int) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(scale), float64(scale))
	op.GeoM.Translate(float64(x), float64(y))
	op.Filter = ebiten.FilterNearest
	c.img.DrawImage(src, op)
}

func (c *Canvas) Image() *ebiten.Image { return c.img }

// Clear 以透明清空畫布。
func (c *Canvas) Clear() { c.img.Clear() }

// DrawAt 把素材畫到指定座標，不做縮放或濾波。
func (c *Canvas) DrawAt(src *ebiten.Image, x, y int) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	op.Filter = ebiten.FilterNearest
	c.img.DrawImage(src, op)
}

// Snapshot 讀回畫布像素，供對拍使用。
func (c *Canvas) Snapshot() *image.RGBA {
	b := c.img.Bounds()
	buf := make([]byte, b.Dx()*b.Dy()*4)
	c.img.ReadPixels(buf)
	out := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	copy(out.Pix, buf)
	return out
}

// WritePNG 把目前畫布寫成 PNG；對拍工具讀的就是這個檔案。
func (c *Canvas) WritePNG(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, c.Snapshot()); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}
