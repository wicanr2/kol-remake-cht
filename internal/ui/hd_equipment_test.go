package ui

import (
	"encoding/json"
	"image"
	"strconv"
	"strings"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/assets"
	"github.com/wicanr2/kol-remake-cht/internal/hdart"
	"github.com/wicanr2/kol-remake-cht/internal/portrait"
)

func TestHDEquipmentAttachRejectsPartialReplacement(t *testing.T) {
	e, err := assets.LoadHDEquipment("../../assets/hd/first-slice/equipment")
	if err != nil {
		t.Fatal(err)
	}
	l, err := portrait.Load("../../assets/original-derived-data/portrait-layout.json")
	if err != nil {
		t.Fatal(err)
	}
	binding := hdart.Binding{}
	for i, b := range l.Body {
		binding.BaseOrigins[i] = b.X
	}
	for i, h := range l.Hands {
		b := e.Manifest.HandPaint[strconv.Itoa(i)]
		binding.SourceHands[i] = [4]int{h.X*8 + b[0], h.Y + b[1], h.X*8 + b[2], h.Y + b[3]}
	}
	for _, s := range e.Manifest.Sources {
		if s.Archive == "BOOKDATA" {
			binding.DefaultShieldRGB = s.RGB
		}
	}
	h := &HD{pages: &hdPages{}}
	if err = h.AttachEquipment(e.Images, e.Manifest, binding); err != nil {
		t.Fatal(err)
	}
	old, count := h.equipment, len(h.pages.sources)
	raw, err := json.Marshal(e.Manifest)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"sources", "last-crop", "missing-image", "hand-binding", "default-binding"} {
		t.Run(name, func(t *testing.T) {
			var m hdart.Equipment
			if err := json.Unmarshal(raw, &m); err != nil {
				t.Fatal(err)
			}
			images := map[string]image.Image{}
			for k, v := range e.Images {
				images[k] = v
			}
			b := binding
			switch name {
			case "sources":
				m.Sources = m.Sources[:215]
			case "last-crop":
				p := m.Parts["LEGSPICS:53"]
				p.Rect[2] = 10000
				m.Parts["LEGSPICS:53"] = p
			case "missing-image":
				delete(images, "outdoor-back-painted.png")
			case "hand-binding":
				b.SourceHands[11] = [4]int{}
			case "default-binding":
				b.DefaultShieldRGB = strings.Repeat("0", 64)
			}
			if err := h.AttachEquipment(images, m, b); err == nil || h.equipment != old || len(h.pages.sources) != count {
				t.Fatalf("partial replacement changed active equipment: %v", err)
			}
		})
	}
}
