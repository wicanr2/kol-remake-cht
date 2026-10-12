package assets

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestHDInterfacePack(t *testing.T) {
	source := "../../assets/hd/first-slice/interface"
	h, err := LoadHDInterface(source)
	if err != nil {
		t.Fatal(err)
	}
	var old struct {
		Fonts []struct{ Glyphs []struct{ Codepoint rune } }
	}
	b, err := os.ReadFile("../../assets/ui-font/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &old); err != nil {
		t.Fatal(err)
	}
	for _, f := range old.Fonts {
		for _, g := range f.Glyphs {
			if _, ok := h.Font.Glyphs[string(g.Codepoint)]; !ok {
				t.Fatalf("正式CJK字集缺字：%U", g.Codepoint)
			}
		}
	}
	for _, bad := range []string{"missing-npc", "font-out-of-bounds", "duplicate-glyph", "missing-font-license", "corrupt-icon"} {
		t.Run(bad, func(t *testing.T) {
			dir := t.TempDir()
			entries, e := os.ReadDir(source)
			if e != nil {
				t.Fatal(e)
			}
			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}
				data, e := os.ReadFile(filepath.Join(source, entry.Name()))
				if e != nil {
					t.Fatal(e)
				}
				if e = os.WriteFile(filepath.Join(dir, entry.Name()), data, 0600); e != nil {
					t.Fatal(e)
				}
			}
			switch bad {
			case "missing-npc":
				err = os.Remove(filepath.Join(dir, "innkeeper-painted-v1.png"))
			case "missing-font-license":
				err = os.Remove(filepath.Join(dir, "font-LICENSE.txt"))
			case "corrupt-icon":
				err = os.WriteFile(filepath.Join(dir, "icon-door.png"), []byte("bad"), 0600)
			default:
				var f HDFont
				raw, e := os.ReadFile(filepath.Join(dir, "font-atlas.json"))
				if e != nil {
					t.Fatal(e)
				}
				if e = json.Unmarshal(raw, &f); e != nil {
					t.Fatal(e)
				}
				if bad == "font-out-of-bounds" {
					g := f.Glyphs["畫"]
					g.X = 6144
					f.Glyphs["畫"] = g
				} else {
					f.Glyphs["畫"] = f.Glyphs["面"]
				}
				raw, e = json.Marshal(f)
				if e != nil {
					t.Fatal(e)
				}
				err = os.WriteFile(filepath.Join(dir, "font-atlas.json"), raw, 0600)
				sum := sha256.Sum256(raw)
				var m map[string]any
				data, e := os.ReadFile(filepath.Join(dir, "manifest.json"))
				if e != nil {
					t.Fatal(e)
				}
				if e = json.Unmarshal(data, &m); e != nil {
					t.Fatal(e)
				}
				m["font_json_sha256"] = hex.EncodeToString(sum[:])
				data, e = json.Marshal(m)
				if e != nil {
					t.Fatal(e)
				}
				if err == nil {
					err = os.WriteFile(filepath.Join(dir, "manifest.json"), data, 0600)
				}
			}
			if err != nil {
				t.Fatal(err)
			}
			if result, e := LoadHDInterface(dir); e == nil || result != nil {
				t.Fatal("破損或越界介面不可部分載入")
			}
		})
	}
}
