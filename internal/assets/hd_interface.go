package assets

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"path/filepath"
)

type HDGlyph struct {
	X, Y    int
	Advance float64
}
type HDFont struct {
	Schema                     string `json:"schema"`
	Size, Cell, Baseline, Xpad int
	Glyphs                     map[string]HDGlyph
}
type HDInterface struct {
	Images           map[string]image.Image
	Font             HDFont
	NPCKey, CrestKey string
}

// LoadHDInterface 驗整組、字格邊界與固定檔名後才交給GPU，不接受部分套件。
func LoadHDInterface(dir string, options ...HDLoadOption) (*HDInterface, error) {
	var m struct {
		Schema, Rights string
		Files          map[string]HDImageInfo
		FontJSON       string `json:"font_json_sha256"`
		License        string `json:"font_license_sha256"`
		NPC            string `json:"npc_rgb_sha256"`
		Crest          string `json:"crest_rgb_sha256"`
	}
	raw, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	if m.Schema != "kol-hd-interface/1" || m.Rights == "" || !validHDHash(m.NPC) || !validHDHash(m.Crest) {
		return nil, fmt.Errorf("HD介面版本或來源無效")
	}
	wanted := map[string]image.Point{"background.png": {1920, 1200}, "crest.png": {320, 320}, "innkeeper-painted-v1.png": {1254, 1254}}
	for _, s := range []image.Point{{256, 288}, {144, 136}, {514, 372}, {1166, 118}, {1166, 406}, {400, 132}, {1164, 588}, {736, 450}, {808, 456}, {878, 882}, {192, 48}, {744, 712}, {664, 148}, {432, 72}} {
		for active := 0; active < 2; active++ {
			wanted[fmt.Sprintf("panel-%d-%d-%d.png", s.X, s.Y, active)] = image.Pt(s.X+8, s.Y+8)
		}
	}
	for _, s := range []string{"door", "back", "yes", "sword", "fist", "swap", "drop", "move", "rest", "shield", "arrow"} {
		wanted["icon-"+s+".png"] = image.Pt(100, 100)
	}
	// 字集可擴充，但atlas不得超過Ebitengine安全尺寸。
	fi, ok := m.Files["font-atlas.png"]
	if !ok || fi.Width != 6144 || fi.Height < 96 || fi.Height > 8192 || fi.Height%96 != 0 {
		return nil, fmt.Errorf("HD字模尺寸無效")
	}
	wanted["font-atlas.png"] = image.Pt(fi.Width, fi.Height)
	if len(m.Files) != len(wanted) {
		return nil, fmt.Errorf("HD介面圖數不符")
	}
	read := func(name, hash string) ([]byte, error) {
		b, e := os.ReadFile(filepath.Join(dir, name))
		if e != nil {
			return nil, e
		}
		sum := sha256.Sum256(b)
		if !validHDHash(hash) || hex.EncodeToString(sum[:]) != hash {
			return nil, fmt.Errorf("HD介面 %s 雜湊不符", name)
		}
		return b, nil
	}
	h := &HDInterface{Images: map[string]image.Image{}, NPCKey: m.NPC, CrestKey: m.Crest}
	for name, size := range wanted {
		info, ok := m.Files[name]
		if !ok || info.Width != size.X || info.Height != size.Y {
			return nil, fmt.Errorf("HD介面 %s 宣告尺寸不符", name)
		}
		if deferHDPixels(options) {
			pic, err := openHDImage(dir, name, info)
			if err != nil {
				return nil, err
			}
			h.Images[name] = pic
			continue
		}
		data, e := read(name, info.SHA256)
		if e != nil {
			return nil, e
		}
		conf, e := png.DecodeConfig(bytes.NewReader(data))
		if e != nil {
			return nil, e
		}
		if conf.Width != size.X || conf.Height != size.Y {
			return nil, fmt.Errorf("HD介面 %s 實際尺寸不符", name)
		}
		img, e := png.Decode(bytes.NewReader(data))
		if e != nil {
			return nil, e
		}
		h.Images[name] = img
	}
	raw, err = read("font-atlas.json", m.FontJSON)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(raw, &h.Font); err != nil {
		return nil, err
	}
	f := h.Font
	if f.Schema != "kol-hd-font/1" || f.Cell != 96 || f.Size != 64 || f.Baseline != 76 || f.Xpad != 10 || len(f.Glyphs) < 2608 {
		return nil, fmt.Errorf("HD字模契約不符")
	}
	occupied := map[image.Point]bool{}
	for ch, g := range f.Glyphs {
		p := image.Pt(g.X, g.Y)
		if len([]rune(ch)) != 1 || g.X < 0 || g.Y < 0 || g.X%96 != 0 || g.Y%96 != 0 || g.X+96 > fi.Width || g.Y+96 > fi.Height || g.Advance <= 0 || g.Advance > 96 || math.IsNaN(g.Advance) || math.IsInf(g.Advance, 0) || occupied[p] {
			return nil, fmt.Errorf("HD字格無效：%s", ch)
		}
		occupied[p] = true
	}
	for _, ch := range " …畫面設定切換後立即預覽同一場景，保留隊伍與操作位置" {
		if _, ok := f.Glyphs[string(ch)]; !ok {
			return nil, fmt.Errorf("HD缺字：%c", ch)
		}
	}
	if _, err = read("font-LICENSE.txt", m.License); err != nil {
		return nil, err
	}
	return h, nil
}
