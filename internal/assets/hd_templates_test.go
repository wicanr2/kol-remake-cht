package assets

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func templateFixture() string {
	if path := os.Getenv("KOL_HD_TEMPLATE_FIXTURE"); path != "" {
		return path
	}
	return filepath.Join("..", "..", "assets", "hd", "first-slice", "templates")
}

func TestHDTemplatesSourceCoverageAndNativeFrames(t *testing.T) {
	lib, err := LoadHDTemplates(templateFixture())
	if err != nil {
		t.Fatal(err)
	}
	if len(lib.Images) != 20 || len(lib.Manifest.Portraits) != 74 || len(lib.Manifest.Preserved) != 6 {
		t.Fatal("incomplete source coverage")
	}
	for _, p := range lib.Manifest.Portraits {
		if p.Rect[2]-p.Rect[0] < 368 || p.Rect[3]-p.Rect[1] < 368 {
			t.Fatalf("undersized source %d", p.Index)
		}
	}
}

func TestHDTemplatesRejectInvalidBindings(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(templateFixture(), "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name   string
		mutate func(*HDTemplatesManifest)
	}{
		{"schema", func(m *HDTemplatesManifest) { m.Schema = "other" }},
		{"count", func(m *HDTemplatesManifest) { m.SourceCount = 81 }},
		{"missing", func(m *HDTemplatesManifest) { m.Portraits = m.Portraits[:73] }},
		{"preserved-order", func(m *HDTemplatesManifest) { m.Preserved[0], m.Preserved[1] = m.Preserved[1], m.Preserved[0] }},
		{"duplicate-index", func(m *HDTemplatesManifest) { m.Portraits[1].Index = m.Portraits[0].Index }},
		{"duplicate-rgb", func(m *HDTemplatesManifest) { m.Portraits[0].SourceRGB = m.Preserved[0].SourceRGB }},
		{"wrong-archive", func(m *HDTemplatesManifest) { m.Portraits[0].Archive = "HEADPICS" }},
		{"path", func(m *HDTemplatesManifest) { m.Portraits[0].File = "../quartet-00.png" }},
		{"outside", func(m *HDTemplatesManifest) { m.Portraits[0].Rect[2] = 4097 }},
		{"undersized", func(m *HDTemplatesManifest) { m.Portraits[0].Rect[2] = m.Portraits[0].Rect[0] + 367 }},
		{"reversed", func(m *HDTemplatesManifest) {
			m.Portraits[0].Rect[0], m.Portraits[0].Rect[2] = m.Portraits[0].Rect[2], m.Portraits[0].Rect[0]
		}},
		{"unapproved-six", func(m *HDTemplatesManifest) { m.ApprovedAtlas = m.SourceZIP }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var m HDTemplatesManifest
			if err := json.Unmarshal(raw, &m); err != nil {
				t.Fatal(err)
			}
			tc.mutate(&m)
			data, err := json.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "manifest.json"), data, 0600); err != nil {
				t.Fatal(err)
			}
			if got, err := LoadHDTemplates(dir); err == nil || got != nil {
				t.Fatal("invalid binding accepted")
			}
		})
	}
}
