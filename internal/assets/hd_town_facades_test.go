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

func TestHDTownFacadesRealSources(t *testing.T) {
	root := filepath.Join("..", "..", "assets")
	m, err := LoadHDTownFacades(filepath.Join(root, "hd", "first-slice", "town-facades"))
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
	if len(m.Images) != 103 {
		t.Fatal("incomplete group")
	}
	// A real highest image index equals the number of members, because index1 is absent.
	if _, _, err = lib.Decode("TWN2PICS", 29); err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"zip", "count", "source-rgb"} {
		copy := *m
		copy.Manifest = m.Manifest
		copy.Manifest.Pictures = append([]HDTownFacadePicture{}, m.Manifest.Pictures...)
		switch kind {
		case "zip":
			copy.Manifest.SourceZIP = strings.Repeat("0", 64)
		case "count":
			copy.Manifest.ArchiveCounts = map[string]int{"MISCPICS": 6}
		case "source-rgb":
			copy.Manifest.Pictures[0].SourceRGB = strings.Repeat("0", 64)
		}
		if err := copy.ValidateSources(lib); err == nil {
			t.Fatalf("accepted bad %s", kind)
		}
	}
}

func TestHDTownFacadesRejectIncompleteAndUnsafeGroups(t *testing.T) {
	source, err := filepath.Abs(filepath.Join("..", "..", "assets", "hd", "first-slice", "town-facades"))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(source, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name   string
		mutate func(*HDTownFacadesManifest)
	}{
		{"schema", func(m *HDTownFacadesManifest) { m.Schema = "unknown" }},
		{"count", func(m *HDTownFacadesManifest) { m.Count-- }},
		{"archive-count", func(m *HDTownFacadesManifest) { m.ArchiveCounts["TWN1PICS"]++ }},
		{"missing-picture", func(m *HDTownFacadesManifest) { m.Pictures = m.Pictures[:102] }},
		{"order", func(m *HDTownFacadesManifest) { m.Pictures[0], m.Pictures[1] = m.Pictures[1], m.Pictures[0] }},
		{"missing-index1", func(m *HDTownFacadesManifest) { m.Pictures[1].Index = 1 }},
		{"duplicate-source", func(m *HDTownFacadesManifest) { m.Pictures[1].SourceRGB = m.Pictures[0].SourceRGB }},
		{"unsafe-path", func(m *HDTownFacadesManifest) { m.Pictures[0].File = "../miscpics-000.png" }},
		{"source-size", func(m *HDTownFacadesManifest) { m.Pictures[0].Size = [2]int{191, 96} }},
		{"native-shape", func(m *HDTownFacadesManifest) {
			q := m.Files[m.Pictures[0].File]
			q.Width--
			m.Files[m.Pictures[0].File] = q
		}},
		{"changed-preserved-inn", func(m *HDTownFacadesManifest) {
			q := m.Files["twn7pics-013.png"]
			q.SHA256 = strings.Repeat("0", 64)
			m.Files["twn7pics-013.png"] = q
		}},
		{"shifted-sign", func(m *HDTownFacadesManifest) { m.Sign.Full[0][0]++ }},
		{"missing-png", func(m *HDTownFacadesManifest) {}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var m HDTownFacadesManifest
			if err := json.Unmarshal(raw, &m); err != nil {
				t.Fatal(err)
			}
			tc.mutate(&m)
			dir := t.TempDir()
			if err := os.Symlink(filepath.Join(source, "generation.json"), filepath.Join(dir, "generation.json")); err != nil {
				t.Fatal(err)
			}
			data, _ := json.Marshal(m)
			if err := os.WriteFile(filepath.Join(dir, "manifest.json"), data, 0600); err != nil {
				t.Fatal(err)
			}
			if got, err := LoadHDTownFacades(dir); err == nil || got != nil {
				t.Fatal("accepted partial or invalid group")
			}
		})
	}
	t.Run("transparent-native-png", func(t *testing.T) {
		var m HDTownFacadesManifest
		if err := json.Unmarshal(raw, &m); err != nil {
			t.Fatal(err)
		}
		im := image.NewNRGBA(image.Rect(0, 0, 1152, 576))
		var b bytes.Buffer
		if err := png.Encode(&b, im); err != nil {
			t.Fatal(err)
		}
		name := m.Pictures[0].File
		m.Files[name] = HDImageInfo{SHA256: monsterHash(b.Bytes()), Width: 1152, Height: 576}
		m.Pictures[0].NativeRGBA = monsterHash(im.Pix)
		dir := t.TempDir()
		if err := os.Symlink(filepath.Join(source, "generation.json"), filepath.Join(dir, "generation.json")); err != nil {
			t.Fatal(err)
		}
		data, _ := json.Marshal(m)
		if err := os.WriteFile(filepath.Join(dir, "manifest.json"), data, 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), b.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}
		if got, err := LoadHDTownFacades(dir); err == nil || got != nil {
			t.Fatal("accepted transparent facade")
		}
	})
}
