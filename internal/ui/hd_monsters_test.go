package ui

import (
	"github.com/wicanr2/kol-remake-cht/internal/assets"
	"image"
	"reflect"
	"strings"
	"testing"
)

func TestMonsterGeometryCompletePopulation(t *testing.T) {
	lib, err := assets.Load("../../assets/original-derived-png/ega")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 32; i++ {
		pic, _, err := lib.Decode("MONPICS", i)
		if err != nil {
			t.Fatal(err)
		}
		p, err := MonsterSheetOrigin(pic.Bounds().Size())
		if err != nil {
			t.Fatal(err)
		}
		if !(image.Rectangle{Min: p, Max: p.Add(pic.Bounds().Size())}).In(image.Rect(17, 3, 136, 76)) {
			t.Fatalf("monster%d leaves original sheet", i)
		}
	}
	for _, size := range []image.Point{{0, 1}, {41, 38}, {128, 38}, {40, 75}, {40, 0}} {
		if _, err = MonsterSheetOrigin(size); err == nil {
			t.Fatal("bad source geometry accepted")
		}
	}
}

func TestAttachMonstersFailureIsAtomic(t *testing.T) {
	m, err := assets.LoadHDMonsters("../../assets/hd/first-slice/monsters")
	if err != nil {
		t.Fatal(err)
	}
	var arts []HDMonsterArt
	for i, p := range m.Manifest.Pictures {
		arts = append(arts, HDMonsterArt{HDPageSource{p.Archive, p.Index, p.SourceRGB}, image.Pt(p.Size[0], p.Size[1]), m.Images[i]})
	}
	for _, name := range []string{"missing", "duplicate", "nil-last", "small-last", "source-conflict"} {
		t.Run(name, func(t *testing.T) {
			h := &HD{pages: &hdPages{validated: true}}
			a := append([]HDMonsterArt{}, arts...)
			switch name {
			case "missing":
				a = a[:31]
			case "duplicate":
				a[31].Source = a[0].Source
			case "nil-last":
				a[31].Paint = nil
			case "small-last":
				a[31].Paint = image.NewNRGBA(image.Rect(0, 0, 1, 1))
			case "source-conflict":
				h.pages.sources = []HDPageSource{{"MONPICS", 31, strings.Repeat("0", 64)}}
			}
			old := append([]HDPageSource{}, h.pages.sources...)
			if err = h.AttachMonsters(a); err == nil {
				t.Fatal("invalid group accepted")
			}
			if h.monsters != nil || !h.pages.validated || !reflect.DeepEqual(append([]HDPageSource{}, h.pages.sources...), old) {
				t.Fatal("failed group mutated HD")
			}
		})
	}
	h := &HD{pages: &hdPages{validated: true}}
	if err = h.AttachMonsters(arts); err != nil {
		t.Fatal(err)
	}
	if h.monsters == nil || h.pages.validated || len(h.pages.sources) != 32 {
		t.Fatal("complete group not attached")
	}
}
