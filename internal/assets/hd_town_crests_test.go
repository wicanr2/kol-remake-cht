package assets

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHDTownCrestsRealSources(t *testing.T) {
	root := filepath.Join("..", "..", "assets")
	m, err := LoadHDTownCrests(filepath.Join(root, "hd", "first-slice", "town-crests"))
	if err != nil {
		t.Fatal(err)
	}
	lib, err := Load(filepath.Join(root, "original-derived-png", "ega"))
	if err != nil {
		t.Fatal(err)
	}
	if err = m.ValidateSources(lib); err != nil {
		t.Fatal(err)
	}
	if len(m.Images) != 7 {
		t.Fatal("incomplete group")
	}
	copy := *m
	copy.Manifest = m.Manifest
	copy.Manifest.Pictures = append([]HDTownCrestPicture{}, m.Manifest.Pictures...)
	copy.Manifest.Pictures[6].RGBA = strings.Repeat("0", 64)
	if err = copy.ValidateSources(lib); err == nil {
		t.Fatal("accepted changed source RGBA")
	}
	copy = *m
	copy.Manifest = m.Manifest
	copy.Manifest.ArchiveCounts = map[string]int{"TWN7PICS": 46}
	if err = copy.ValidateSources(lib); err == nil {
		t.Fatal("accepted wrong source membership")
	}
}

func TestHDTownCrestsRejectBadGroups(t *testing.T) {
	source, err := filepath.Abs(filepath.Join("..", "..", "assets", "hd", "first-slice", "town-crests"))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(source, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name   string
		change func(*HDTownCrestsManifest)
	}{
		{"missing-town", func(m *HDTownCrestsManifest) { m.Pictures = m.Pictures[:6] }},
		{"wrong-index", func(m *HDTownCrestsManifest) { m.Pictures[0].Index = 1 }},
		{"duplicate-source", func(m *HDTownCrestsManifest) { m.Pictures[1].SourceRGB = m.Pictures[0].SourceRGB }},
		{"unsafe-path", func(m *HDTownCrestsManifest) { m.Pictures[0].File = "../crest.png" }},
		{"changed-preserved-crest", func(m *HDTownCrestsManifest) {
			p := m.Files["twn7pics-000.png"]
			p.SHA256 = strings.Repeat("0", 64)
			m.Files["twn7pics-000.png"] = p
		}},
		{"missing-last-png", func(m *HDTownCrestsManifest) {}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var m HDTownCrestsManifest
			if err := json.Unmarshal(raw, &m); err != nil {
				t.Fatal(err)
			}
			tc.change(&m)
			dir := t.TempDir()
			for name := range m.Files {
				if name == "twn7pics-000.png" && tc.name == "missing-last-png" {
					continue
				}
				if err := os.Symlink(filepath.Join(source, name), filepath.Join(dir, name)); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.Symlink(filepath.Join(source, "generation.json"), filepath.Join(dir, "generation.json")); err != nil {
				t.Fatal(err)
			}
			data, err := json.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "manifest.json"), data, 0600); err != nil {
				t.Fatal(err)
			}
			if got, err := LoadHDTownCrests(dir); err == nil || got != nil {
				t.Fatal("accepted invalid or partial group")
			}
		})
	}
}
