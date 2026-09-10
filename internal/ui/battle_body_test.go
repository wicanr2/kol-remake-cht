package ui

import (
	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"testing"
)

func TestBodyMaskPreservesOtherPaletteColoursAndBounds(t *testing.T) {
	palette := [16][3]byte{{0, 0, 0}, {0, 0, 170}, {0, 170, 0}, {0, 170, 170}, {170, 0, 0}, {170, 0, 170}, {170, 85, 0}, {170, 170, 170}, {85, 85, 85}, {85, 85, 255}, {85, 255, 85}, {85, 255, 255}, {255, 85, 85}, {255, 85, 255}, {255, 255, 85}, {255, 255, 255}}
	pixels := make([]byte, bodyWidth*bodyHeight*4)
	for y := 0; y < bodyHeight; y++ {
		for x := 0; x < bodyWidth; x++ {
			p := (y*bodyWidth + x) * 4
			c := palette[x%16]
			copy(pixels[p:p+3], c[:])
			pixels[p+3] = 255
		}
	}
	shadeBody(pixels, [6]battle.BodyRect{{Left: bodyX, Top: bodyY, Right: bodyX + 16, Bottom: bodyY + 1}})
	for y := 0; y < bodyHeight; y++ {
		for x := 0; x < bodyWidth; x++ {
			want := palette[x%16]
			if y == 0 && x < 16 && (x == 7 || x == 8 || x == 15) {
				want = palette[12]
			}
			p := (y*bodyWidth + x) * 4
			if pixels[p] != want[0] || pixels[p+1] != want[1] || pixels[p+2] != want[2] || pixels[p+3] != 255 {
				t.Fatalf("遮罩意外改動像素 (%d,%d)", x, y)
			}
		}
	}
}
