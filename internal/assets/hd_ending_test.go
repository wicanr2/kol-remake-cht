package assets

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestHDEndingMatchesSourcePictures(t *testing.T) {
	const dir = "../../assets/hd/first-slice/events"
	h, err := LoadHDEnding(dir)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 7; i++ {
		file, err := os.Open(fmt.Sprintf("../../assets/original-derived-png/ega/ENDPICS/endpics-%03d.png", i))
		if err != nil {
			t.Fatal(err)
		}
		im, err := png.Decode(file)
		file.Close()
		if err != nil {
			t.Fatal(err)
		}
		var rgb []byte
		for y := im.Bounds().Min.Y; y < im.Bounds().Max.Y; y++ {
			for x := im.Bounds().Min.X; x < im.Bounds().Max.X; x++ {
				r, g, b, _ := im.At(x, y).RGBA()
				rgb = append(rgb, byte(r>>8), byte(g>>8), byte(b>>8))
			}
		}
		sum := sha256.Sum256(rgb)
		if h.Images[i] == nil || h.SourceRGB[i] != hex.EncodeToString(sum[:]) {
			t.Fatalf("第%d段未對應真實原圖", i)
		}
	}
}

func TestHDEndingRejectsIncompleteOrMisindexedPack(t *testing.T) {
	const source = "../../assets/hd/first-slice/events"
	for _, kind := range []string{"missing", "corrupt", "size", "index", "duplicate", "path"} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			entries, err := os.ReadDir(source)
			if err != nil {
				t.Fatal(err)
			}
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				data, err := os.ReadFile(filepath.Join(source, entry.Name()))
				if err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(filepath.Join(dir, entry.Name()), data, 0600); err != nil {
					t.Fatal(err)
				}
			}
			raw, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
			if err != nil {
				t.Fatal(err)
			}
			var m HDEndingManifest
			if err = json.Unmarshal(raw, &m); err != nil {
				t.Fatal(err)
			}
			name := m.Panels[6].File
			switch kind {
			case "missing":
				err = os.Remove(filepath.Join(dir, name))
			case "corrupt":
				err = os.WriteFile(filepath.Join(dir, name), []byte("not a PNG"), 0600)
			case "size":
				i := m.Files[name]
				i.Width++
				m.Files[name] = i
			case "index":
				m.Panels[6].Index = 7
			case "duplicate":
				m.Panels[6].Index = 0
			case "path":
				m.Panels[6].File = "../endpics-006-painted.png"
			}
			if err != nil {
				t.Fatal(err)
			}
			raw, err = json.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(filepath.Join(dir, "manifest.json"), raw, 0600); err != nil {
				t.Fatal(err)
			}
			if h, err := LoadHDEnding(dir); err == nil || h != nil {
				t.Fatal("無效子組不可交回部分素材")
			}
		})
	}
}
