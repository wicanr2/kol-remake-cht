package assets

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const monsterDir = "../../assets/hd/first-slice/monsters"

func TestHDMonstersCompleteSourceAndAlpha(t *testing.T) {
	m, err := LoadHDMonsters(monsterDir)
	if err != nil {
		t.Fatal(err)
	}
	lib, err := Load("../../assets/original-derived-png/ega")
	if err != nil {
		t.Fatal(err)
	}
	if err = m.ValidateSources(lib); err != nil {
		t.Fatal(err)
	}
	if m.Manifest.Pictures[13].AlphaRange[1] != 254 {
		t.Fatal("valid alpha254 foreground lost")
	}
	if err = m.ValidateSources(nil); err == nil {
		t.Fatal("nil source accepted")
	}
	for _, field := range []string{"png", "rgb", "rgba", "size", "count", "zip"} {
		t.Run(field, func(t *testing.T) {
			bad := *m
			bad.Manifest.Pictures = append([]HDMonsterPicture{}, m.Manifest.Pictures...)
			s := &bad.Manifest.Pictures[31]
			switch field {
			case "png":
				s.SourcePNG = strings.Repeat("0", 64)
			case "rgb":
				s.SourceRGB = strings.Repeat("0", 64)
			case "rgba":
				s.RGBA = strings.Repeat("0", 64)
			case "size":
				s.Size[1]++
			case "zip":
				bad.Manifest.SourceZIP = strings.Repeat("0", 64)
			}
			if field == "count" {
				short := *lib
				short.index = map[string]map[int]Image{"MONPICS": {}}
				if err = bad.ValidateSources(&short); err == nil {
					t.Fatal("short archive accepted")
				}
			} else if err = bad.ValidateSources(lib); err == nil {
				t.Fatal("wrong source accepted")
			}
		})
	}
}

func TestHDMonstersRejectsIncompleteOrDamagedGroup(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(monsterDir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	base, err := filepath.Abs(monsterDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"schema", "count", "missing-entry", "duplicate", "unsafe-path", "native-size", "png-hash", "missing-png", "actual-size", "rgba", "alpha", "border", "generation"} {
		t.Run(name, func(t *testing.T) {
			var m HDMonstersManifest
			if err = json.Unmarshal(raw, &m); err != nil {
				t.Fatal(err)
			}
			p := &m.Pictures[31]
			switch name {
			case "schema":
				m.Schema = "bad"
			case "count":
				m.Count--
			case "missing-entry":
				m.Pictures = m.Pictures[:31]
			case "duplicate":
				p.SourceRGB = m.Pictures[0].SourceRGB
			case "unsafe-path":
				p.File = "../escape.png"
			case "native-size":
				p.Size = [2]int{120, 74}
				f := m.Files[p.File]
				f.Width = 719
				m.Files[p.File] = f
			case "png-hash":
				f := m.Files[p.File]
				f.SHA256 = strings.Repeat("0", 64)
				m.Files[p.File] = f
			case "actual-size":
				f := m.Files[p.File]
				f.Width++
				m.Files[p.File] = f
			case "rgba":
				p.NativeRGBA = strings.Repeat("0", 64)
			case "alpha":
				p.AlphaRange[0] = 1
			case "border":
				p.AlphaBBox[0] = 0
			case "generation":
				m.GenerationSHA = strings.Repeat("0", 64)
			}
			dir := t.TempDir()
			for file := range m.Files {
				if name == "missing-png" && file == "monster-31.png" {
					continue
				}
				if err = os.Symlink(filepath.Join(base, file), filepath.Join(dir, file)); err != nil {
					t.Fatal(err)
				}
			}
			if err = os.Symlink(filepath.Join(base, "generation.json"), filepath.Join(dir, "generation.json")); err != nil {
				t.Fatal(err)
			}
			data, err := json.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(filepath.Join(dir, "manifest.json"), data, 0644); err != nil {
				t.Fatal(err)
			}
			got, err := LoadHDMonsters(dir)
			if err == nil || got != nil {
				t.Fatal("invalid partial group returned")
			}
		})
	}
}
