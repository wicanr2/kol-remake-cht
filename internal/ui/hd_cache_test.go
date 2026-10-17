package ui

import (
	"image"
	"image/color"
	"testing"
)

type cacheTestProvider struct{ read func() image.Image }

func (p *cacheTestProvider) Bounds() image.Rectangle { return image.Rect(0, 0, 32, 16) }
func (p *cacheTestProvider) ColorModel() color.Model { return color.RGBAModel }
func (p *cacheTestProvider) At(int, int) color.Color {
	panic("provider must be resolved before reading pixels")
}
func (p *cacheTestProvider) ReadHDImage() (image.Image, error) { return p.read(), nil }

func TestHDCacheLoadsOnlyDrawnArtAndRestoresEvictedCrops(t *testing.T) {
	h := &HD{}
	loads := 0
	makeArt := func() image.Image {
		loads++
		pic := image.NewRGBA(image.Rect(0, 0, 32, 16))
		pic.SetRGBA(7, 5, color.RGBA{17, 43, 97, 255})
		return pic
	}
	root := h.newPaint(&cacheTestProvider{read: makeArt})
	unused := h.newPaintLoader(image.Rect(0, 0, 32, 16), func() (image.Image, error) { t.Fatal("unseen art decoded"); return nil, nil })
	_ = unused
	crop := h.paintSubImage(root, image.Rect(3, 2, 18, 10))
	if loads != 0 {
		t.Fatal("indexed art decoded")
	}
	h.BeginHDFrame()
	h.paint(crop)
	if loads != 1 || !h.cache.entries[root].resident {
		t.Fatal("requested crop not loaded")
	}
	if err := h.EndHDFrame(); err != nil {
		t.Fatal(err)
	}
	h.ReleaseHDMemory()
	if h.cache.bytes != 0 || h.cache.entries[root].resident {
		t.Fatal("EGA retained texture")
	}
	h.BeginHDFrame()
	h.paint(crop)
	if loads != 2 {
		t.Fatal("evicted crop did not reload")
	}
	// Charge a deliberately large image without allocating pixels. Verify the
	// actual budget policy evicts old data and keeps the current draw pinned.
	h.cache.entries[root].bytes = hdTextureBudget + 1
	h.cache.bytes = hdTextureBudget + 1
	if err := h.EndHDFrame(); err != nil {
		t.Fatal(err)
	}
	if !h.cache.entries[root].resident {
		t.Fatal("current frame was evicted")
	}
	h.BeginHDFrame()
	if err := h.EndHDFrame(); err != nil {
		t.Fatal(err)
	}
	if h.cache.bytes != 0 || h.cache.entries[root].resident {
		t.Fatal("old over-budget image not released")
	}
}
