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

func TestHDTownMapsRealSources(t *testing.T) {
	m, err := LoadHDTownMaps(filepath.Join("..", "..", "assets", "hd", "first-slice", "town-maps"))
	if err != nil {
		t.Fatal(err)
	}
	lib, err := Load(filepath.Join("..", "..", "assets", "original-derived-png", "ega"))
	if err != nil {
		t.Fatal(err)
	}
	if err = m.ValidateSources(lib); err != nil {
		t.Fatal(err)
	}
	if len(m.Images) != 336 || len(m.Grids) != 8 || m.Grids["TOWN2"] == nil || m.Grids["TOWN0"] == nil {
		t.Fatal("incomplete city")
	}
	before := m.Grids["TOWN1"]
	original := m.Manifest.Maps[0]
	cases := []struct {
		name   string
		change func(*HDTownMapSource)
	}{
		{"wrong-atlas-count", func(s *HDTownMapSource) { s.Count = 3 }},
		{"wrong-grid-png", func(s *HDTownMapSource) { s.Grid.SourcePNG = strings.Repeat("0", 64) }},
		{"wrong-grid-rgb", func(s *HDTownMapSource) { s.Grid.SourceRGB = strings.Repeat("0", 64) }},
		{"wrong-grid-rgba", func(s *HDTownMapSource) { s.GridRGBA = strings.Repeat("0", 64) }},
		{"wrong-atlas-rgba", func(s *HDTownMapSource) { s.AtlasRGBA = strings.Repeat("0", 64) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m.Manifest.Maps[0] = original
			tc.change(&m.Manifest.Maps[0])
			if m.ValidateSources(lib) == nil {
				t.Fatal("accepted different source")
			}
			if m.Grids["TOWN1"] != before {
				t.Fatal("partially replaced grids")
			}
		})
	}
	m.Manifest.Maps[0] = original
	second := m.Manifest.Maps[1]
	beforeSecond := m.Grids["TOWN2"]
	m.Manifest.Maps[1].Count = 3
	if m.ValidateSources(lib) == nil || m.Grids["TOWN1"] != before || m.Grids["TOWN2"] != beforeSecond {
		t.Fatal("invalid second city partially replaced the complete group")
	}
	m.Manifest.Maps[1] = second
	changedLib := *lib
	changedLib.index = map[string]map[int]Image{}
	for a, entries := range lib.index {
		changedLib.index[a] = entries
	}
	changedLib.index["TOWN1"] = map[int]Image{}
	for i, p := range lib.index["TOWN1"] {
		changedLib.index["TOWN1"][i] = p
	}
	changedLib.index["TOWN1"][99] = Image{}
	if m.ValidateSources(&changedLib) == nil || m.Grids["TOWN1"] != before {
		t.Fatal("accepted extra actual archive entry or replaced grids")
	}
	oldZip := m.Manifest.SourceZIP
	m.Manifest.SourceZIP = strings.Repeat("0", 64)
	if m.ValidateSources(lib) == nil || m.ValidateSources(nil) == nil {
		t.Fatal("accepted wrong/nil mother library")
	}
	m.Manifest.SourceZIP = oldZip
}

func TestHDTownMapsRejectPartial(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", "assets", "hd", "first-slice", "town-maps"))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name   string
		change func(*HDTownMapsManifest)
	}{
		{"gap", func(m *HDTownMapsManifest) { m.Regions[0].VisibleBounds[2]-- }},
		{"overlap", func(m *HDTownMapsManifest) { m.Regions[0].VisibleBounds[2]++ }},
		{"reversed", func(m *HDTownMapsManifest) { m.Regions[0].VisibleBounds = [4]int{6, 0, 0, 3} }},
		{"outside-source", func(m *HDTownMapsManifest) { m.Regions[0].VisibleBounds[0] = -1 }},
		{"source-too-small", func(m *HDTownMapsManifest) { m.Regions[0].SourceBounds[2]-- }},
		{"source-outside-map", func(m *HDTownMapsManifest) { m.Regions[0].SourceBounds = [4]int{-1, 0, 11, 12} }},
		{"duplicate-id", func(m *HDTownMapsManifest) { m.Regions[1].ID = m.Regions[0].ID }},
		{"unsafe-file", func(m *HDTownMapsManifest) { m.Regions[0].File = "../x.png" }},
		{"missing-file-entry", func(m *HDTownMapsManifest) { delete(m.Files, m.Regions[0].File) }},
		{"wrong-png-sha", func(m *HDTownMapsManifest) {
			n := m.Regions[0].File
			p := m.Files[n]
			p.SHA256 = strings.Repeat("0", 64)
			m.Files[n] = p
		}},
		{"wrong-rgba-sha", func(m *HDTownMapsManifest) { m.Regions[0].NativeRGBA = strings.Repeat("0", 64) }},
		{"wrong-size", func(m *HDTownMapsManifest) {
			n := m.Regions[0].File
			p := m.Files[n]
			p.Width = 1152
			p.Height = 1152
			m.Files[n] = p
		}},
		{"below-approved-delivery-size", func(m *HDTownMapsManifest) {
			n := m.Regions[0].File
			p := m.Files[n]
			p.Width, p.Height = 767, 767
			m.Files[n] = p
		}},
		{"wrong-source-index", func(m *HDTownMapsManifest) { m.Maps[0].Grid.Index = 1 }},
		{"wrong-source-count", func(m *HDTownMapsManifest) { m.Maps[0].Count = 256 }},
		{"wrong-generation", func(m *HDTownMapsManifest) { m.GenerationSHA = strings.Repeat("0", 64) }},
		{"missing-last-png", func(m *HDTownMapsManifest) {}},
		{"missing-second-city-region", func(m *HDTownMapsManifest) { m.Regions = append(m.Regions[:42], m.Regions[43:]...) }},
		{"missing-world-region", func(m *HDTownMapsManifest) { m.Regions = m.Regions[:len(m.Regions)-1] }},
		{"missing-second-city-png", func(m *HDTownMapsManifest) {}},
		{"transparent-native", func(m *HDTownMapsManifest) {}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var m HDTownMapsManifest
			if err := json.Unmarshal(raw, &m); err != nil {
				t.Fatal(err)
			}
			tc.change(&m)
			dir := t.TempDir()
			if tc.name == "transparent-native" {
				pic := image.NewNRGBA(image.Rect(0, 0, 1152, 1152))
				var buf bytes.Buffer
				if err := png.Encode(&buf, pic); err != nil {
					t.Fatal(err)
				}
				name := m.Regions[0].File
				if err := os.WriteFile(filepath.Join(dir, name), buf.Bytes(), 0600); err != nil {
					t.Fatal(err)
				}
				m.Files[name] = HDImageInfo{SHA256: monsterHash(buf.Bytes()), Width: 1152, Height: 1152}
				m.Regions[0].NativeRGBA = monsterHash(pic.Pix)
			}
			for name := range m.Files {
				if tc.name == "transparent-native" && name == m.Regions[0].File {
					continue
				}
				if tc.name == "missing-last-png" && name == "town1-r06-c05.png" {
					continue
				}
				if tc.name == "missing-second-city-png" && name == "town2-r06-c05.png" {
					continue
				}
				if err := os.Symlink(filepath.Join(root, name), filepath.Join(dir, name)); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.Symlink(filepath.Join(root, "generation.json"), filepath.Join(dir, "generation.json")); err != nil {
				t.Fatal(err)
			}
			data, err := json.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(filepath.Join(dir, "manifest.json"), data, 0600); err != nil {
				t.Fatal(err)
			}
			if got, err := LoadHDTownMaps(dir); err == nil || got != nil {
				t.Fatal("accepted invalid or partial city")
			}
		})
	}
}
