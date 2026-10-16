package assets

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHDSceneEventsSources(t *testing.T) {
	dir := filepath.Join("..", "..", "assets", "hd", "first-slice", "scene-events")
	events, e := LoadHDSceneEvents(dir)
	if e != nil {
		t.Fatal(e)
	}
	lib, e := Load(filepath.Join("..", "..", "assets", "original-derived-png", "ega"))
	if e != nil {
		t.Fatal(e)
	}
	if e = events.ValidateSources(lib); e != nil {
		t.Fatal(e)
	}
	if len(events.Images) != 33 {
		t.Fatal("partial group")
	}
	original := events.Manifest.Pictures[0]
	for _, change := range []func(*HDSceneEventPicture){func(p *HDSceneEventPicture) { p.SourcePNG = strings.Repeat("0", 64) }, func(p *HDSceneEventPicture) { p.SourceRGB = strings.Repeat("0", 64) }, func(p *HDSceneEventPicture) { p.RGBA = strings.Repeat("0", 64) }, func(p *HDSceneEventPicture) { p.Size = [2]int{191, 96} }} {
		events.Manifest.Pictures[0] = original
		change(&events.Manifest.Pictures[0])
		if events.ValidateSources(lib) == nil {
			t.Fatal("accepted different source")
		}
	}
	events.Manifest.Pictures[0] = original
	events.Manifest.ArchiveCounts["COMBPICS"] = 8
	if events.ValidateSources(lib) == nil {
		t.Fatal("accepted actual count mismatch")
	}
	if events.ValidateSources(nil) == nil {
		t.Fatal("accepted nil library")
	}
}
func TestHDSceneEventsRejectIncomplete(t *testing.T) {
	root, e := filepath.Abs(filepath.Join("..", "..", "assets", "hd", "first-slice", "scene-events"))
	if e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile(filepath.Join(root, "manifest.json"))
	if e != nil {
		t.Fatal(e)
	}
	cases := []struct {
		name   string
		change func(*HDSceneEventsManifest)
	}{
		{"partial-group", func(m *HDSceneEventsManifest) { m.Pictures = m.Pictures[:31] }},
		{"wrong-source-count", func(m *HDSceneEventsManifest) { m.ArchiveCounts["TWN0PICS"] = 24 }},
		{"wrong-index", func(m *HDSceneEventsManifest) { m.Pictures[0].Index = 1 }},
		{"duplicate-rgb", func(m *HDSceneEventsManifest) { m.Pictures[1].SourceRGB = m.Pictures[0].SourceRGB }},
		{"unsafe-file", func(m *HDSceneEventsManifest) { m.Pictures[0].File = "../other.png" }},
		{"undersized-delivery", func(m *HDSceneEventsManifest) {
			n := m.Pictures[0].File
			p := m.Files[n]
			p.Width = 767
			m.Files[n] = p
		}},
		{"wrong-rgba", func(m *HDSceneEventsManifest) { m.Pictures[0].DeliveryRGBA = strings.Repeat("0", 64) }},
		{"wrong-generation", func(m *HDSceneEventsManifest) { m.GenerationSHA = strings.Repeat("0", 64) }},
		{"missing-last-png", func(*HDSceneEventsManifest) {}},
		{"transparent", func(*HDSceneEventsManifest) {}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var m HDSceneEventsManifest
			if e := json.Unmarshal(raw, &m); e != nil {
				t.Fatal(e)
			}
			tc.change(&m)
			dir := t.TempDir()
			if tc.name == "transparent" {
				pic := image.NewNRGBA(image.Rect(0, 0, 768, 384))
				var buf bytes.Buffer
				if e := png.Encode(&buf, pic); e != nil {
					t.Fatal(e)
				}
				name := m.Pictures[0].File
				if e := os.WriteFile(filepath.Join(dir, name), buf.Bytes(), 0600); e != nil {
					t.Fatal(e)
				}
				m.Files[name] = HDImageInfo{SHA256: monsterHash(buf.Bytes()), Width: 768, Height: 384}
				m.Pictures[0].DeliveryRGBA = monsterHash(pic.Pix)
			}
			for name := range m.Files {
				if tc.name == "missing-last-png" && name == "combpics-006.png" || tc.name == "transparent" && name == m.Pictures[0].File {
					continue
				}
				if e := os.Symlink(filepath.Join(root, name), filepath.Join(dir, name)); e != nil {
					t.Fatal(e)
				}
			}
			if e := os.Symlink(filepath.Join(root, "generation.json"), filepath.Join(dir, "generation.json")); e != nil {
				t.Fatal(e)
			}
			data, e := json.Marshal(m)
			if e != nil {
				t.Fatal(e)
			}
			if e = os.WriteFile(filepath.Join(dir, "manifest.json"), data, 0600); e != nil {
				t.Fatal(e)
			}
			if got, e := LoadHDSceneEvents(dir); e == nil || got != nil {
				t.Fatal("accepted incomplete or invalid event group")
			}
		})
	}
}
