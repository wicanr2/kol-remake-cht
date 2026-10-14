package font

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// Painter 只替換單次呈現的字模。Measure、字距及遊戲狀態保持原值。
// handled=false 時沿原字模繪製，讓離屏暫存畫布不受主畫布的呈現影響。
type Painter interface {
	DrawFont(dst *ebiten.Image, f *Font, x, y int, text string, ink color.RGBA, background *color.RGBA) (drawn int, handled bool)
}

var painter Painter

// WithPainter 限繪圖主執行緒使用；巢狀呼叫、錯誤及panic皆還原先前呈現。
func WithPainter(next Painter, render func() error) error {
	previous := painter
	painter = next
	defer func() { painter = previous }()
	return render()
}

// RuneAdvance 供HD呈現沿用同一字格，不另改命中或截斷量測。
func (f *Font) RuneAdvance(ch rune) int { return f.advance(ch) }
