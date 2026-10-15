package ui

import (
	"github.com/wicanr2/kol-remake-cht/internal/assets"
	"image"
	"testing"
)

func TestAttachActionsDoesNotReplaceWithPartialGroup(t *testing.T) {
	a, err := assets.LoadHDActions("../../assets/hd/first-slice/actions")
	if err != nil {
		t.Fatal(err)
	}
	var sources []HDActionSource
	for _, s := range a.Manifest.Sources {
		sources = append(sources, HDActionSource{HDPageSource: HDPageSource{Archive: s.Archive, Index: s.Index, RGB: s.SourceRGB}, Rect: image.Rect(s.Rect[0], s.Rect[1], s.Rect[2], s.Rect[3])})
	}
	h := &HD{pages: &hdPages{}}
	if err = h.AttachActions(a.Image, sources); err != nil {
		t.Fatal(err)
	}
	old := h.actions
	for _, name := range []string{"missing", "mask", "last-rect", "duplicate", "no-atlas"} {
		t.Run(name, func(t *testing.T) {
			bad := append([]HDActionSource{}, sources...)
			atlas := a.Image
			switch name {
			case "missing":
				bad = bad[:67]
			case "mask":
				bad[67].Index = 28
			case "last-rect":
				bad[67].Rect.Max.X++
			case "duplicate":
				bad[67] = bad[0]
			case "no-atlas":
				atlas = nil
			}
			if err = h.AttachActions(atlas, bad); err == nil || h.actions != old {
				t.Fatalf("partial group changed active icons: %v", err)
			}
		})
	}
}
