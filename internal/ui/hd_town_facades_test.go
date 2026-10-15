package ui

import (
	"fmt"
	"image"
	"reflect"
	"testing"
)

func TestHDTownFacadesAtomicFailure(t *testing.T) {
	paint := image.NewNRGBA(image.Rect(0, 0, 1152, 576))
	var arts []HDTownFacadeArt
	add := func(archive string, index int) {
		arts = append(arts, HDTownFacadeArt{Source: HDPageSource{Archive: archive, Index: index, RGB: fmt.Sprintf("%064x", len(arts)+1)}, Paint: paint})
	}
	add("MISCPICS", 0)
	for town, count := range []int{33, 29, 25, 25, 29, 25, 45} {
		for i := 3; i <= count; i += 2 {
			add(fmt.Sprintf("TWN%dPICS", town+1), i)
		}
	}
	sign := HDTownFacadeSign{NativeSize: image.Pt(1774, 887), FontSize: 32, LineHeight: 34,
		Full: [3]image.Rectangle{image.Rect(770, 519, 1100, 625), image.Rect(-245, 626, 85, 740), image.Rect(1716, 519, 2046, 625)},
		Clip: [3]image.Rectangle{image.Rect(770, 519, 1100, 625), image.Rect(0, 626, 85, 740), image.Rect(1716, 519, 1774, 625)}}
	cases := []struct {
		name   string
		change func([]HDTownFacadeArt, *HDTownFacadeSign)
	}{
		{"last-image-missing", func(a []HDTownFacadeArt, s *HDTownFacadeSign) { a[len(a)-1].Paint = nil }},
		{"wrong-order", func(a []HDTownFacadeArt, s *HDTownFacadeSign) { a[0], a[1] = a[1], a[0] }},
		{"duplicate-source", func(a []HDTownFacadeArt, s *HDTownFacadeSign) { a[1].Source.RGB = a[0].Source.RGB }},
		{"small-native", func(a []HDTownFacadeArt, s *HDTownFacadeSign) {
			a[len(a)-1].Paint = image.NewRGBA(image.Rect(0, 0, 192, 96))
		}},
		{"sign-clip-outside", func(a []HDTownFacadeArt, s *HDTownFacadeSign) { s.Clip[0] = image.Rect(-1, 0, 10, 10) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			glyphs := map[string]HDGlyph{}
			for _, r := range "請勿踐踏草坪" {
				glyphs[string(r)] = HDGlyph{Advance: 64}
			}
			old := &hdTownFacades{}
			sources := []HDPageSource{{Archive: "keep", Index: 99, RGB: "keep"}}
			h := &HD{facades: old, pages: &hdPages{sources: sources, validated: true}, style: &hdInterface{glyphs: glyphs}}
			for _, a := range arts {
				if a.Source.Archive == "TWN7PICS" && a.Source.Index == 13 {
					h.facadeKey = a.Source.RGB
				}
			}
			input := append([]HDTownFacadeArt{}, arts...)
			cfg := sign
			tc.change(input, &cfg)
			if err := h.AttachTownFacades(input, cfg); err == nil {
				t.Fatal("accepted invalid group")
			}
			if h.facades != old || !h.pages.validated || !reflect.DeepEqual(h.pages.sources, sources) {
				t.Fatal("failed attachment changed active group")
			}
		})
	}
	if got := townFacadeHDRect(); got != image.Rect(624, 348, 1776, 924) {
		t.Fatalf("viewport changed: %v", got)
	}
}
