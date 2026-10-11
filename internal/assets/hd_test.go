package assets

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestHDWholePackValidation(t *testing.T) {
	const source = "../../assets/hd/first-slice"
	if _, err := LoadHD(source); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(source, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"missing-image", "corrupt-image", "wrong-dimensions", "duplicate-portrait"} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			for _, name := range []string{"inn-painted-v1.png", "party-painted-v1.png"} {
				data, e := os.ReadFile(filepath.Join(source, name))
				if e != nil {
					t.Fatal(e)
				}
				if e = os.WriteFile(filepath.Join(dir, name), data, 0600); e != nil {
					t.Fatal(e)
				}
			}
			var m HDManifest
			if e := json.Unmarshal(raw, &m); e != nil {
				t.Fatal(e)
			}
			switch kind {
			case "missing-image":
				err = os.Remove(filepath.Join(dir, "party-painted-v1.png"))
			case "corrupt-image":
				err = os.WriteFile(filepath.Join(dir, "party-painted-v1.png"), []byte("bad PNG"), 0600)
			case "wrong-dimensions":
				info := m.Files["party-painted-v1.png"]
				info.Width++
				m.Files["party-painted-v1.png"] = info
			case "duplicate-portrait":
				m.PortraitRGB[1] = m.PortraitRGB[0]
			}
			if err != nil {
				t.Fatal(err)
			}
			data, e := json.Marshal(m)
			if e != nil {
				t.Fatal(e)
			}
			if e = os.WriteFile(filepath.Join(dir, "manifest.json"), data, 0600); e != nil {
				t.Fatal(e)
			}
			if lib, e := LoadHD(dir); e == nil || lib != nil {
				t.Fatalf("%s不能交回部分素材：%v", kind, e)
			}
		})
	}
}
