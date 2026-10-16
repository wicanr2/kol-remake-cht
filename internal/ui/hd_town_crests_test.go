package ui

import (
	"fmt"
	"image"
	"reflect"
	"testing"
)

func TestHDTownCrestsAtomicFailure(t *testing.T) {
	var arts []HDTownCrestArt
	for i := 0; i < 7; i++ {
		arts = append(arts, HDTownCrestArt{Source: HDPageSource{Archive: fmt.Sprintf("TWN%dPICS", i+1), Index: 0, RGB: fmt.Sprintf("%064x", i+1)}, Paint: image.NewNRGBA(image.Rect(0, 0, 320, 320))})
	}
	cases := []struct {
		name   string
		change func([]HDTownCrestArt)
	}{
		{"last-missing", func(a []HDTownCrestArt) { a[6].Paint = nil }},
		{"swapped-town", func(a []HDTownCrestArt) { a[0], a[1] = a[1], a[0] }},
		{"wrong-size", func(a []HDTownCrestArt) { a[6].Paint = image.NewRGBA(image.Rect(0, 0, 56, 57)) }},
		{"duplicate-source", func(a []HDTownCrestArt) { a[5].Source.RGB = a[4].Source.RGB }},
		{"preserved-source-changed", func(a []HDTownCrestArt) { a[6].Source.RGB = fmt.Sprintf("%064x", 8) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			old := &hdTownCrests{}
			sources := []HDPageSource{{Archive: "keep", Index: 99, RGB: "keep"}}
			h := &HD{crests: old, pages: &hdPages{sources: sources, validated: true}, style: &hdInterface{crestKey: arts[6].Source.RGB}}
			input := append([]HDTownCrestArt{}, arts...)
			tc.change(input)
			if err := h.AttachTownCrests(input); err == nil {
				t.Fatal("accepted invalid group")
			}
			if h.crests != old || !h.pages.validated || !reflect.DeepEqual(h.pages.sources, sources) {
				t.Fatal("failed attachment changed active group")
			}
		})
	}
}
