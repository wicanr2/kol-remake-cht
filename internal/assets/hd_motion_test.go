package assets

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestHDMotionAssetsAndSources(t *testing.T) {
	dir := "../../assets/hd/first-slice/motion"
	a, e := LoadHDMotionArt(dir)
	if e != nil {
		t.Fatal(e)
	}
	lib, e := Load("../../assets/original-derived-png/ega")
	if e != nil {
		t.Fatal(e)
	}
	if e = a.ValidateSources(lib); e != nil {
		t.Fatal(e)
	}
	if len(a.Images) != 65 || a.Manifest.ArchiveCounts["KOLPICS"] != 8 {
		t.Fatal("complete group or sparse archive count changed")
	}
	for i, p := range a.Images {
		b := p.Bounds()
		if b.Dx() > 768 || b.Dy() > 768 {
			t.Fatalf("unshrunk image%d %v", i, b)
		}
	}
}

func TestHDMotionRejectsBrokenCompleteGroup(t *testing.T) {
	dir := "../../assets/hd/first-slice/motion"
	raw, e := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if e != nil {
		t.Fatal(e)
	}
	cases := map[string]func(*HDMotionManifest){
		"missing record":     func(m *HDMotionManifest) { m.Pictures = m.Pictures[:63] },
		"wrong sparse count": func(m *HDMotionManifest) { m.ArchiveCounts["KOLPICS"] = 13 },
		"duplicate identity": func(m *HDMotionManifest) { m.Pictures[1] = m.Pictures[0] },
		"wrong source index": func(m *HDMotionManifest) { m.Pictures[0].Index = 99 },
		"unsafe file":        func(m *HDMotionManifest) { m.Pictures[0].File = "../outside.png" },
		"wrong size":         func(m *HDMotionManifest) { m.Pictures[0].Size = [2]int{768, 480} },
		"wrong transparency": func(m *HDMotionManifest) { m.Pictures[0].Transparent = true },
		"oversized delivery": func(m *HDMotionManifest) {
			f := m.Files[m.Pictures[0].File]
			f.Width = 1536
			m.Files[m.Pictures[0].File] = f
		},
		"bad source hash":       func(m *HDMotionManifest) { m.Pictures[0].SourceRGB = "bad" },
		"wrong generation hash": func(m *HDMotionManifest) { m.GenerationSHA = string(make([]byte, 64)) },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			var m HDMotionManifest
			if e := json.Unmarshal(raw, &m); e != nil {
				t.Fatal(e)
			}
			change(&m)
			temp := t.TempDir()
			data, e := json.Marshal(m)
			if e != nil {
				t.Fatal(e)
			}
			if e = os.WriteFile(filepath.Join(temp, "manifest.json"), data, 0600); e != nil {
				t.Fatal(e)
			}
			if _, e = LoadHDMotionArt(temp); e == nil {
				t.Fatal("broken motion group accepted")
			}
		})
	}
}
