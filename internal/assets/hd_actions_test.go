package assets

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const actionAssetDir = "../../assets/hd/first-slice/actions"

func TestHDActionsRejectIncompleteOrCorruptGroups(t *testing.T) {
	good, err := LoadHDActions(actionAssetDir)
	if err != nil {
		t.Fatal(err)
	}
	lib, err := Load("../../assets/original-derived-png/ega")
	if err != nil {
		t.Fatal(err)
	}
	if err = good.ValidateSources(lib); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(actionAssetDir, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	atlas, err := os.ReadFile(filepath.Join(actionAssetDir, "actions.png"))
	if err != nil {
		t.Fatal(err)
	}
	license, err := os.ReadFile(filepath.Join(actionAssetDir, "font-LICENSE.txt"))
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"schema", "count", "mask28", "duplicate", "bounds", "hash", "size", "pixel", "alpha", "empty", "missing", "license", "source-png", "source-rgb"} {
		t.Run(name, func(t *testing.T) {
			var m HDActionsManifest
			if err := json.Unmarshal(raw, &m); err != nil {
				t.Fatal(err)
			}
			data := atlas
			switch name {
			case "schema":
				m.Schema = "invalid"
			case "count":
				m.Sources = m.Sources[:67]
			case "mask28":
				m.Sources[0].Index = 28
			case "duplicate":
				m.Sources[67] = m.Sources[0]
			case "bounds":
				m.Sources[67].Rect[2] = 10000
			case "hash":
				x := m.Files[m.Image]
				x.SHA256 = strings.Repeat("0", 64)
				m.Files[m.Image] = x
			case "size":
				x := m.Files[m.Image]
				x.Height = 961
				m.Files[m.Image] = x
			case "pixel":
				m.Sources[67].RGBA = strings.Repeat("0", 64)
			case "alpha":
				m.Sources[67].Alpha = [4]int{0, 0, 144, 120}
			case "empty":
				im := image.NewNRGBA(good.Image.Bounds())
				draw.Draw(im, im.Bounds(), good.Image, image.Point{}, draw.Src)
				draw.Draw(im, image.Rect(0, 0, 144, 120), image.Transparent, image.Point{}, draw.Src)
				var b bytes.Buffer
				if err := png.Encode(&b, im); err != nil {
					t.Fatal(err)
				}
				data = b.Bytes()
				x := m.Files[m.Image]
				x.SHA256 = hdTestHash(data)
				m.Files[m.Image] = x
				m.Sources[0].RGBA = hdTestHash(make([]byte, 144*120*4))
				m.Sources[0].Alpha = [4]int{}
			case "license":
				m.FontLicense = strings.Repeat("0", 64)
			case "source-png":
				m.Sources[67].SourcePNG = strings.Repeat("0", 64)
			case "source-rgb":
				m.Sources[67].SourceRGB = strings.Repeat("0", 64)
			}
			dir := t.TempDir()
			encoded, err := json.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			for file, body := range map[string][]byte{"manifest.json": encoded, "actions.png": data, "font-LICENSE.txt": license} {
				if name == "missing" && file == "actions.png" {
					continue
				}
				if err := os.WriteFile(filepath.Join(dir, file), body, 0644); err != nil {
					t.Fatal(err)
				}
			}
			a, err := LoadHDActions(dir)
			if name == "source-png" || name == "source-rgb" {
				if err != nil {
					t.Fatal(err)
				}
				if err = a.ValidateSources(lib); err == nil {
					t.Fatal("wrong source accepted")
				}
				return
			}
			if err == nil || a != nil {
				t.Fatalf("accepted corrupt group: %v", err)
			}
		})
	}
}

func hdTestHash(b []byte) string {
	const hex = "0123456789abcdef"
	s := sha256.Sum256(b)
	out := make([]byte, 64)
	for i, v := range s {
		out[i*2] = hex[v>>4]
		out[i*2+1] = hex[v&15]
	}
	return string(out)
}

func TestHDActionsCheckActualSourceBytes(t *testing.T) {
	a, err := LoadHDActions(actionAssetDir)
	if err != nil {
		t.Fatal(err)
	}
	lib, err := Load("../../assets/original-derived-png/ega")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err = os.Mkdir(filepath.Join(dir, "KOLMPICS"), 0755); err != nil {
		t.Fatal(err)
	}
	copyLib := *lib
	copyLib.root = dir
	for _, meta := range lib.index["KOLMPICS"] {
		path := filepath.Join(dir, meta.File)
		if meta.UnitIndex == 81 {
			im := image.NewNRGBA(image.Rect(0, 0, 24, 20))
			im.Set(0, 0, color.White)
			var b bytes.Buffer
			if err = png.Encode(&b, im); err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(path, b.Bytes(), 0644); err != nil {
				t.Fatal(err)
			}
		} else {
			source, err := filepath.Abs(filepath.Join(lib.root, meta.File))
			if err != nil {
				t.Fatal(err)
			}
			if err = os.Symlink(source, path); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err = a.ValidateSources(&copyLib); err == nil || !strings.Contains(err.Error(), "第81圖來源PNG雜湊不符") {
		t.Fatalf("actual modified PNG accepted or wrong failure: %v", err)
	}
	if err = a.ValidateSources(nil); err == nil {
		t.Fatal("nil library accepted")
	}
	wrongMother := *lib
	wrongMother.manifest.SourceZip.SHA256 = strings.Repeat("0", 64)
	if err = a.ValidateSources(&wrongMother); err == nil {
		t.Fatal("wrong source ZIP identity accepted")
	}
	short := *lib
	short.index = map[string]map[int]Image{"KOLMPICS": {}}
	for id, meta := range lib.index["KOLMPICS"] {
		if id != 81 {
			short.index["KOLMPICS"][id] = meta
		}
	}
	if err = a.ValidateSources(&short); err == nil {
		t.Fatal("70-PNG source archive accepted")
	}
}
