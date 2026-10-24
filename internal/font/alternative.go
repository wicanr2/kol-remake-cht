package font

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// SetAlternatives 只替換繪製字模；原字距、量測、缺字及背景範圍保持。
// nil 清除呈現選項，還原原字模。兩個尺寸先完整檢查才交換。
func (f *Font) SetAlternatives(normal, large *Font) error {
	if normal == nil && large == nil {
		f.alternative, f.alternativeLarge = nil, nil
		return nil
	}
	if normal == nil || large == nil || normal == f || large == f ||
		normal.meta.Cell != f.meta.Cell || large.Height() != f.Height()*2 ||
		large.meta.Cell.Width != f.meta.Cell.Width*2 || len(normal.mapped) != len(large.mapped) {
		return fmt.Errorf("替代字模尺寸不符")
	}
	for ch, g := range normal.mapped {
		big, ok := large.mapped[ch]
		if ch < 0x2e80 || !f.HasRune(ch) || g.Advance != f.advance(ch) || !ok || big.Advance != g.Advance*2 {
			return fmt.Errorf("替代字模字距或字碼不符：%U", ch)
		}
	}
	f.alternative, f.alternativeLarge = normal, large
	return nil
}

// HasRune 不受呈現選項影響，供字模清冊及缺字驗證使用。
func (f *Font) HasRune(ch rune) bool {
	if len(f.mapped) > 0 {
		_, ok := f.mapped[ch]
		return ok
	}
	return int(ch) >= f.meta.FirstCodepoint && int(ch) < f.meta.FirstCodepoint+f.meta.GlyphCount
}

func (f *Font) glyphForDraw(ch rune) (*ebiten.Image, bool) {
	if f.alternative != nil && ch >= 0x2e80 {
		if glyph, ok := f.alternative.glyph(ch); ok {
			return glyph, true
		}
	}
	return f.glyph(ch)
}

// DrawAlternative 讓HD沿原文字座標與安全矩形繪製較大PNG；英文或缺字回原呈現。
// sx、sy 是相對原字格的倍率，clip 已由呼叫端依原截斷／換行政策決定。
func (f *Font) DrawAlternative(dst *ebiten.Image, ch rune, x, y, sx, sy float64, ink color.RGBA, clip image.Rectangle) bool {
	if f.alternative == nil || ch < 0x2e80 {
		return false
	}
	face := f.alternative
	if sx > 1 || sy > 1 {
		face = f.alternativeLarge
	}
	glyph, ok := face.glyph(ch)
	if !ok {
		return false
	}
	clip = clip.Intersect(dst.Bounds())
	if clip.Empty() {
		return true
	}
	var op ebiten.DrawImageOptions
	op.GeoM.Scale(sx*float64(f.meta.Cell.Width)/float64(face.meta.Cell.Width), sy*float64(f.Height())/float64(face.Height()))
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(ink)
	op.Filter = ebiten.FilterLinear
	dst.SubImage(clip).(*ebiten.Image).DrawImage(glyph, &op)
	return true
}
