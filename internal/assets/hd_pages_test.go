package assets

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const pageFixture = "../../assets/hd/first-slice/pages"

func TestHDPageSourcePictures(t *testing.T) {
	p, err := LoadHDPages(pageFixture)
	if err != nil {
		t.Fatal(err)
	}
	for _, source := range p.Sources() {
		name := filepath.Join("../../assets/original-derived-png/ega", source.Archive,
			fmt.Sprintf("%s-%03d.png", strings.ToLower(source.Archive), source.Index))
		data, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != source.SourcePNG {
			t.Fatalf("%s PNG來源不符", name)
		}
		im, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		h := sha256.New()
		for y := im.Bounds().Min.Y; y < im.Bounds().Max.Y; y++ {
			for x := im.Bounds().Min.X; x < im.Bounds().Max.X; x++ {
				r, g, b, _ := im.At(x, y).RGBA()
				h.Write([]byte{byte(r >> 8), byte(g >> 8), byte(b >> 8)})
			}
		}
		if hex.EncodeToString(h.Sum(nil)) != source.SourceRGB {
			t.Fatalf("%s RGB來源不符", name)
		}
	}
}

func TestHDPagesRejectPartialOrWrongPack(t *testing.T) {
	for _, name := range []string{"missing", "corrupt", "size", "index", "duplicate-source", "path"} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			raw, err := os.ReadFile(filepath.Join(pageFixture, "manifest.json"))
			if err != nil {
				t.Fatal(err)
			}
			var m HDPagesManifest
			if err := json.Unmarshal(raw, &m); err != nil {
				t.Fatal(err)
			}
			for file := range m.Files {
				data, err := os.ReadFile(filepath.Join(pageFixture, file))
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, file), data, 0600); err != nil {
					t.Fatal(err)
				}
			}
			switch name {
			case "missing":
				err = os.Remove(filepath.Join(dir, m.Portraits[2].File))
			case "corrupt":
				err = os.WriteFile(filepath.Join(dir, m.Book), []byte("broken"), 0600)
			case "size":
				info := m.Files[m.Book]
				info.Width--
				m.Files[m.Book] = info
			case "index":
				m.Portraits[0].Index = 13
			case "duplicate-source":
				m.Portraits[0].SourceRGB = m.Books[0].SourceRGB
			case "path":
				m.Portraits[0].File = "../book-open-painted.png"
			}
			if err != nil {
				t.Fatal(err)
			}
			raw, err = json.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "manifest.json"), raw, 0600); err != nil {
				t.Fatal(err)
			}
			if p, err := LoadHDPages(dir); err == nil || p != nil {
				t.Fatalf("%s 必須拒絕整組：%v", name, err)
			}
		})
	}
}
