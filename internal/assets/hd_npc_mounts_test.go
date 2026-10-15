package assets

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const npcMountDir = "../../assets/hd/first-slice/npc-mounts"

func TestHDNPCMountsCompleteSourceGroup(t *testing.T) {
	n, err := LoadHDNPCMounts(npcMountDir)
	if err != nil {
		t.Fatal(err)
	}
	lib, err := Load("../../assets/original-derived-png/ega")
	if err != nil {
		t.Fatal(err)
	}
	if err = n.ValidateSources(lib); err != nil {
		t.Fatal(err)
	}
	if len(n.Portraits) != 98 || len(n.Mounts) != 4 || len(n.Sources()) != 106 {
		t.Fatal("incomplete group")
	}
	if err = n.ValidateSources(nil); err == nil {
		t.Fatal("nil source accepted")
	}
	wrong := *lib
	wrong.manifest.SourceZip.SHA256 = strings.Repeat("0", 64)
	if err = n.ValidateSources(&wrong); err == nil {
		t.Fatal("wrong mother ZIP accepted")
	}
	short := *lib
	short.index = map[string]map[int]Image{}
	for a, idx := range lib.index {
		short.index[a] = idx
	}
	short.index["TWN7PICS"] = map[int]Image{}
	for i, meta := range lib.index["TWN7PICS"] {
		if i != 44 {
			short.index["TWN7PICS"][i] = meta
		}
	}
	if err = n.ValidateSources(&short); err == nil {
		t.Fatal("short actual archive accepted")
	}
	for _, field := range []string{"png", "rgb", "size"} {
		t.Run(field, func(t *testing.T) {
			bad := *n
			bad.Manifest.Portraits = append([]HDNPCMountPortrait{}, n.Manifest.Portraits...)
			s := &bad.Manifest.Portraits[len(bad.Manifest.Portraits)-1]
			switch field {
			case "png":
				s.SourcePNG = strings.Repeat("0", 64)
			case "rgb":
				s.SourceRGB = strings.Repeat("0", 64)
			case "size":
				s.Size[0]++
			}
			if err = bad.ValidateSources(lib); err == nil {
				t.Fatal("wrong source accepted")
			}
		})
	}
}

func TestHDNPCMountsRejectsPartialOrUnsafeGroups(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(npcMountDir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	base, err := filepath.Abs(filepath.Dir(npcMountDir))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"schema", "archive-count", "source-count", "duplicate-source", "duplicate-rgb", "wrong-id", "source-size", "kind", "unsafe-path", "png-hash", "png-size", "missing-png", "crop", "reversed-crop", "native-size", "rgba", "preserved-order", "preserved-path", "preserved-hash", "generation"} {
		t.Run(name, func(t *testing.T) {
			var m HDNPCMountsManifest
			if err = json.Unmarshal(raw, &m); err != nil {
				t.Fatal(err)
			}
			switch name {
			case "schema":
				m.Schema = "wrong"
			case "archive-count":
				m.ArchiveCounts["TWN7PICS"]--
			case "source-count":
				m.Portraits = m.Portraits[:101]
			case "duplicate-source":
				m.Portraits[101].HDNPCMountSource = m.Portraits[0].HDNPCMountSource
			case "duplicate-rgb":
				m.Portraits[101].SourceRGB = m.Portraits[0].SourceRGB
			case "wrong-id":
				m.Portraits[0].Index = 1
			case "source-size":
				m.Portraits[0].Size[0]++
			case "kind":
				m.Portraits[101].Kind = "invalid"
			case "unsafe-path":
				for k, v := range m.Files {
					delete(m.Files, k)
					m.Files["../escape.png"] = v
					break
				}
			case "png-hash":
				for k, v := range m.Files {
					v.SHA256 = strings.Repeat("0", 64)
					m.Files[k] = v
					break
				}
			case "png-size":
				for k, v := range m.Files {
					v.Width++
					m.Files[k] = v
					break
				}
			case "crop":
				m.Portraits[101].Rect[2] = 10000
			case "reversed-crop":
				m.Portraits[101].Rect[0] = m.Portraits[101].Rect[2] + 1
			case "native-size":
				m.Portraits[0].Rect[2] = m.Portraits[0].Rect[0] + 527
			case "rgba":
				m.Portraits[101].RGBA = strings.Repeat("0", 64)
			case "preserved-order":
				m.Preserved[0], m.Preserved[1] = m.Preserved[1], m.Preserved[0]
			case "preserved-path":
				m.Preserved[3].File = "../../escape.png"
			case "preserved-hash":
				m.Preserved[3].PNG = strings.Repeat("0", 64)
			case "generation":
				m.GenerationSHA = strings.Repeat("0", 64)
			}
			root := t.TempDir()
			dir := filepath.Join(root, "npc-mounts")
			if err = os.Mkdir(dir, 0755); err != nil {
				t.Fatal(err)
			}
			for _, s := range []string{"pages", "interface"} {
				if err = os.Symlink(filepath.Join(base, s), filepath.Join(root, s)); err != nil {
					t.Fatal(err)
				}
			}
			for file := range m.Files {
				if name == "unsafe-path" || name == "missing-png" && file == "steeds-00-03-painted-v1.png" {
					continue
				}
				if err = os.Symlink(filepath.Join(base, "npc-mounts", file), filepath.Join(dir, file)); err != nil {
					t.Fatal(err)
				}
			}
			if err = os.Symlink(filepath.Join(base, "npc-mounts/generation.json"), filepath.Join(dir, "generation.json")); err != nil {
				t.Fatal(err)
			}
			data, err := json.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(filepath.Join(dir, "manifest.json"), data, 0644); err != nil {
				t.Fatal(err)
			}
			got, err := LoadHDNPCMounts(dir)
			if err == nil || got != nil {
				t.Fatalf("partial group accepted: %v", err)
			}
		})
	}
}
