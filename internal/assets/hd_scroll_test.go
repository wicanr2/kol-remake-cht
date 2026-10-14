package assets

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestHDScrollSource(t *testing.T) {
	p, err := LoadHDScroll("../../assets/hd/first-slice/scroll")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("../../assets/original-derived-png/ega/KOLPICS/kolpics-001.png")
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(data)
	if hex.EncodeToString(sum[:]) != p.Manifest.Source.SourcePNG {
		t.Fatal("來源PNG雜湊不符")
	}
	im, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	var rgb []byte
	for y := im.Bounds().Min.Y; y < im.Bounds().Max.Y; y++ {
		for x := im.Bounds().Min.X; x < im.Bounds().Max.X; x++ {
			c := color.NRGBAModel.Convert(im.At(x, y)).(color.NRGBA)
			rgb = append(rgb, c.R, c.G, c.B)
		}
	}
	sum = sha256.Sum256(rgb)
	if hex.EncodeToString(sum[:]) != p.Manifest.Source.SourceRGB {
		t.Fatal("來源RGB雜湊不符")
	}
}

func TestHDScrollRejectsIncompletePack(t *testing.T) {
	for _, kind := range []string{"missing", "corrupt", "schema", "index", "archive", "source", "size", "extra", "path", "transparent"} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			root := "../../assets/hd/first-slice/scroll"
			raw, err := os.ReadFile(filepath.Join(root, "manifest.json"))
			if err != nil {
				t.Fatal(err)
			}
			var m HDScrollManifest
			if err := json.Unmarshal(raw, &m); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(filepath.Join(root, m.Image))
			if err != nil {
				t.Fatal(err)
			}
			switch kind {
			case "missing":
				data = nil
			case "corrupt":
				data = []byte("broken PNG")
			case "schema":
				m.Schema = "kol-hd-scroll/unknown"
			case "index":
				m.Source.Index = 14
			case "archive":
				m.Source.Archive = "BOOKPICS"
			case "source":
				m.Source.SourceRGB = "invalid"
			case "size":
				info := m.Files[m.Image]
				info.Height++
				m.Files[m.Image] = info
			case "extra":
				m.Files["extra.png"] = m.Files[m.Image]
			case "path":
				m.Image = "../spell-scroll-painted.png"
			case "transparent":
				var buf bytes.Buffer
				if err := png.Encode(&buf, image.NewNRGBA(image.Rect(0, 0, 2, 2))); err != nil {
					t.Fatal(err)
				}
				data = buf.Bytes()
				sum := sha256.Sum256(data)
				m.Files[m.Image] = HDImageInfo{SHA256: hex.EncodeToString(sum[:]), Width: 2, Height: 2}
			}
			if data != nil {
				if err := os.WriteFile(filepath.Join(dir, "spell-scroll-painted.png"), data, 0600); err != nil {
					t.Fatal(err)
				}
			}
			raw, err = json.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "manifest.json"), raw, 0600); err != nil {
				t.Fatal(err)
			}
			if pack, err := LoadHDScroll(dir); err == nil || pack != nil {
				t.Fatalf("%s accepted incomplete pack", kind)
			}
		})
	}
}
