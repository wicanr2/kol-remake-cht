package ui

import (
	"image"
	"image/color"
	"testing"
)

func TestHDSamplePixelCentersAndEdges(t *testing.T) {
	source := image.NewRGBA(image.Rect(0, 0, 2, 2))
	source.SetRGBA(0, 0, color.RGBA{0, 0, 0, 255})
	source.SetRGBA(1, 0, color.RGBA{120, 0, 0, 255})
	source.SetRGBA(0, 1, color.RGBA{0, 80, 0, 255})
	source.SetRGBA(1, 1, color.RGBA{120, 80, 0, 255})
	if got := resampleHD(source, 1, 1).RGBAAt(0, 0); got != (color.RGBA{60, 40, 0, 255}) {
		t.Fatalf("center: %v", got)
	}
	up := resampleHD(source, 3, 3)
	want := [3][3]color.RGBA{
		{{0, 0, 0, 255}, {60, 0, 0, 255}, {120, 0, 0, 255}},
		{{0, 40, 0, 255}, {60, 40, 0, 255}, {120, 40, 0, 255}},
		{{0, 80, 0, 255}, {60, 80, 0, 255}, {120, 80, 0, 255}},
	}
	for y, row := range want {
		for x, expected := range row {
			if got := up.RGBAAt(x, y); got != expected {
				t.Errorf("%d,%d: %v, want %v", x, y, got, expected)
			}
		}
	}
}

func TestHDSampleSubimageStrideAndPremultipliedAlpha(t *testing.T) {
	parent := image.NewRGBA(image.Rect(0, 0, 7, 6))
	sub := parent.SubImage(image.Rect(3, 2, 5, 4)).(*image.RGBA)
	sub.SetRGBA(3, 2, color.RGBA{100, 50, 0, 100})
	sub.SetRGBA(4, 2, color.RGBA{0, 0, 0, 0})
	sub.SetRGBA(3, 3, color.RGBA{0, 0, 0, 0})
	sub.SetRGBA(4, 3, color.RGBA{0, 0, 0, 0})
	if got := resampleHD(sub, 1, 1).RGBAAt(0, 0); got != (color.RGBA{25, 13, 0, 25}) {
		t.Fatalf("premultiplied subimage center: %v", got)
	}
}
