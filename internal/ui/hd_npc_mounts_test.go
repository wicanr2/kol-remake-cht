package ui

import (
	"image"
	"reflect"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/kol-remake-cht/internal/assets"
)

func TestAttachNPCMountsKeepsExistingGroupOnError(t *testing.T) {
	n, err := assets.LoadHDNPCMounts("../../assets/hd/first-slice/npc-mounts")
	if err != nil {
		t.Fatal(err)
	}
	var kept, sources []HDPageSource
	for _, s := range n.Manifest.Preserved {
		kept = append(kept, HDPageSource{s.Archive, s.Index, s.SourceRGB})
	}
	for _, s := range n.Sources() {
		sources = append(sources, HDPageSource{s.Archive, s.Index, s.SourceRGB})
	}
	var mountKey string
	for _, s := range sources {
		if s.Archive == "STEEDS" {
			mountKey = s.RGB
		}
	}
	newBase := func() *HD {
		h := &HD{pages: &hdPages{portraits: map[string]*ebiten.Image{}, icons: map[string]*ebiten.Image{}, validated: true}, style: &hdInterface{npcKey: kept[3].RGB}}
		for _, p := range kept[:3] {
			h.pages.portraits[p.RGB] = ebiten.NewImage(1, 1)
			h.pages.sources = append(h.pages.sources, p)
		}
		return h
	}
	for _, name := range []string{"missing-source", "duplicate-last-source", "missing-image", "small-image", "preserved-overwrite", "existing-source-conflict"} {
		t.Run(name, func(t *testing.T) {
			h := newBase()
			s := append([]HDPageSource{}, sources...)
			pics := map[string]image.Image{}
			for k, p := range n.Portraits {
				pics[k] = p
			}
			mounts := map[string]image.Image{}
			for k, p := range n.Mounts {
				mounts[k] = p
			}
			switch name {
			case "missing-source":
				s = s[:105]
			case "duplicate-last-source":
				s[105] = s[0]
			case "missing-image":
				delete(mounts, mountKey)
			case "small-image":
				mounts[mountKey] = image.NewRGBA(image.Rect(0, 0, 239, 228))
			case "preserved-overwrite":
				for k := range pics {
					delete(pics, k)
					break
				}
				pics[kept[0].RGB] = image.NewRGBA(image.Rect(0, 0, 528, 384))
			case "existing-source-conflict":
				h.pages.sources = append(h.pages.sources, HDPageSource{s[105].Archive, s[105].Index, "wrong"})
			}
			oldSources := append([]HDPageSource{}, h.pages.sources...)
			if err = h.AttachNPCMounts(pics, mounts, kept, s); err == nil {
				t.Fatal("invalid group accepted")
			}
			if len(h.pages.portraits) != 3 || len(h.pages.icons) != 0 || !h.pages.validated || len(h.portraitFit) != 0 || !reflect.DeepEqual(h.pages.sources, oldSources) {
				t.Fatal("failed attachment mutated existing group")
			}
		})
	}
	h := newBase()
	if err = h.AttachNPCMounts(n.Portraits, n.Mounts, kept, sources); err != nil {
		t.Fatal(err)
	}
	if len(h.pages.portraits) != 101 || len(h.pages.icons) != 4 || len(h.pages.sources) != 106 || len(h.pages.npcMountSizes) != 106 || h.pages.validated {
		t.Fatal("complete group not attached")
	}
	for _, pic := range h.pages.icons {
		if !h.portraitFit[pic] {
			t.Fatal("mount would be stretched")
		}
	}
	for _, p := range kept[:3] {
		if h.pages.portraits[p.RGB].Bounds() != image.Rect(0, 0, 1, 1) {
			t.Fatal("approved NPC changed")
		}
	}
}
