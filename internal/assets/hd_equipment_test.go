package assets

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/hdart"
)

const equipmentDir = "../../assets/hd/first-slice/equipment"

func TestHDEquipmentCompleteSourceGroup(t *testing.T) {
	e, err := LoadHDEquipment(equipmentDir)
	if err != nil {
		t.Fatal(err)
	}
	lib, err := Load("../../assets/original-derived-png/ega")
	if err != nil {
		t.Fatal(err)
	}
	if err = e.ValidateSources(lib); err != nil {
		t.Fatal(err)
	}
	if err = e.ValidateSources(nil); err == nil {
		t.Fatal("nil source accepted")
	}
	wrong := *lib
	wrong.manifest.SourceZip.SHA256 = strings.Repeat("0", 64)
	if err = e.ValidateSources(&wrong); err == nil {
		t.Fatal("wrong ZIP source accepted")
	}
	short := *lib
	short.index = map[string]map[int]Image{}
	for name, index := range lib.index {
		short.index[name] = index
	}
	short.index["HANDPICS"] = map[int]Image{}
	for id, meta := range lib.index["HANDPICS"] {
		if id != 11 {
			short.index["HANDPICS"][id] = meta
		}
	}
	if err = e.ValidateSources(&short); err == nil {
		t.Fatal("11 hands accepted")
	}
}

func TestHDEquipmentRejectsIncompleteOrUnsafeInputs(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(equipmentDir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	lib, err := Load("../../assets/original-derived-png/ega")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"schema", "part-count", "prop-count", "source-count", "mask-id", "duplicate-source", "path", "file-hash", "file-size", "crop", "empty-crop", "world", "native-size", "kind", "grip", "mesh", "head", "hand-paint", "missing-file", "source-png", "source-rgb", "source-rendered-rgb", "raw-anchors"} {
		t.Run(name, func(t *testing.T) {
			var m hdart.Equipment
			if err := json.Unmarshal(raw, &m); err != nil {
				t.Fatal(err)
			}
			switch name {
			case "schema":
				m.Schema = "wrong"
			case "part-count":
				delete(m.Parts, "WINGS:3")
			case "prop-count":
				delete(m.Props, "SHLDPICS:3")
			case "source-count":
				m.Sources = m.Sources[:215]
			case "mask-id":
				m.Sources[0].Archive = "KOLMPICS"
				m.Sources[0].Index = 28
			case "duplicate-source":
				m.Sources[215] = m.Sources[0]
			case "path":
				for k, v := range m.Files {
					delete(m.Files, k)
					m.Files["../outside.png"] = v
					break
				}
			case "file-hash":
				for k, v := range m.Files {
					v.SHA256 = strings.Repeat("0", 64)
					m.Files[k] = v
					break
				}
			case "file-size":
				for k, v := range m.Files {
					v.Width++
					m.Files[k] = v
					break
				}
			case "crop":
				v := m.Parts["HEADPICS:53"]
				v.Rect[2] = 10000
				m.Parts["HEADPICS:53"] = v
			case "empty-crop":
				v := m.Parts["HEADPICS:53"]
				v.Rect[2] = v.Rect[0]
				m.Parts["HEADPICS:53"] = v
			case "world":
				v := m.Parts["HEADPICS:53"]
				v.World[1] = v.World[3]
				m.Parts["HEADPICS:53"] = v
			case "native-size":
				v := m.Parts["HEADPICS:53"]
				v.World[2] = 639
				m.Parts["HEADPICS:53"] = v
			case "kind":
				v := m.Parts["LEGSPICS:0"]
				v.Kind = "armor-only"
				m.Parts["LEGSPICS:0"] = v
			case "grip":
				v := m.Grips["5:53"]
				v[1].Rect[2] = 10000
				m.Grips["5:53"] = v
			case "mesh":
				v := m.Meshes["TORPICS:53"]
				v.NativeWaist = 0
				m.Meshes["TORPICS:53"] = v
			case "head":
				m.HeadBands["5"] = [3]int{0, 0, 0}
			case "hand-paint":
				m.HandPaint["11"] = [4]int{1, 1, 2, 2}
			case "source-png":
				m.Sources[215].PNG = strings.Repeat("0", 64)
			case "source-rgb":
				m.Sources[215].RGB = strings.Repeat("0", 64)
			case "source-rendered-rgb":
				m.Sources[215].RenderedRGB = strings.Repeat("0", 64)
			}
			dir := t.TempDir()
			for file := range m.Files {
				if name == "path" || name == "missing-file" && file == "outdoor-back-painted.png" {
					continue
				}
				source, err := filepath.Abs(filepath.Join(equipmentDir, file))
				if err != nil {
					t.Fatal(err)
				}
				if err = os.Symlink(source, filepath.Join(dir, file)); err != nil {
					t.Fatal(err)
				}
			}
			encoded, err := json.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			if name == "raw-anchors" {
				var fields map[string]any
				if err = json.Unmarshal(encoded, &fields); err != nil {
					t.Fatal(err)
				}
				fields["source_hands"] = map[string]any{}
				encoded, err = json.Marshal(fields)
				if err != nil {
					t.Fatal(err)
				}
			}
			if err = os.WriteFile(filepath.Join(dir, "manifest.json"), encoded, 0644); err != nil {
				t.Fatal(err)
			}
			e, err := LoadHDEquipment(dir)
			if name == "source-png" || name == "source-rgb" || name == "source-rendered-rgb" || name == "hand-paint" {
				if err != nil {
					t.Fatal(err)
				}
				if err = e.ValidateSources(lib); err == nil {
					t.Fatal("wrong actual source accepted")
				}
				return
			}
			if err == nil || e != nil {
				t.Fatalf("accepted invalid group: %v", err)
			}
		})
	}
}

func TestHDEquipmentBindingDoesNotSerializeOriginalAnchors(t *testing.T) {
	b, err := json.Marshal(hdart.Binding{BaseOrigins: [6]int{1}, SourceHands: [12][4]int{{1, 2, 3, 4}}, DefaultShieldRGB: "private"})
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "{}" {
		t.Fatalf("transient source geometry serialized: %s", b)
	}
}
