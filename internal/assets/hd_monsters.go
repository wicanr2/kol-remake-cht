package assets

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
)

type HDMonsterSource struct {
	HDPageSource
	Size [2]int `json:"size"`
	RGBA string `json:"source_rgba_sha256"`
}
type HDMonsterPicture struct {
	HDMonsterSource
	File       string   `json:"file"`
	NativeRGBA string   `json:"native_rgba_sha256"`
	AlphaBBox  [4]int   `json:"visible_alpha_bbox"`
	AlphaRange [2]uint8 `json:"alpha_range"`
}
type HDMonstersManifest struct {
	Schema        string                 `json:"schema"`
	Rights        string                 `json:"rights"`
	SourceZIP     string                 `json:"source_zip_sha256"`
	Count         int                    `json:"source_count"`
	Files         map[string]HDImageInfo `json:"files"`
	Pictures      []HDMonsterPicture     `json:"pictures"`
	GenerationSHA string                 `json:"generation_sha256"`
}
type HDMonsters struct {
	Manifest HDMonstersManifest
	Images   [32]image.Image
}

func monsterHash(data []byte) string { return fmt.Sprintf("%x", sha256.Sum256(data)) }

// LoadHDMonsters 驗完整32張原生透明圖片後才回傳，不提供部分組。
func LoadHDMonsters(dir string, options ...HDLoadOption) (*HDMonsters, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return nil, err
	}
	var m HDMonstersManifest
	if err = json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	if m.Schema != "kol-hd-monsters/1" || m.Rights == "" || !validHDHash(m.SourceZIP) || m.Count != 32 || len(m.Files) != 32 || len(m.Pictures) != 32 || !validHDHash(m.GenerationSHA) {
		return nil, fmt.Errorf("HD怪物版本、來源或完整數量不符")
	}
	generation, err := os.ReadFile(filepath.Join(dir, "generation.json"))
	if err != nil {
		return nil, err
	}
	if monsterHash(generation) != m.GenerationSHA {
		return nil, fmt.Errorf("HD怪物生成清冊雜湊不符")
	}
	result := &HDMonsters{Manifest: m}
	seen := map[string]bool{}
	for i, p := range m.Pictures {
		if p.Archive != "MONPICS" || p.Index != i || p.File != fmt.Sprintf("monster-%02d.png", i) || seen[p.SourceRGB] || !validHDHash(p.SourcePNG) || !validHDHash(p.SourceRGB) || !validHDHash(p.RGBA) || !validHDHash(p.NativeRGBA) {
			return nil, fmt.Errorf("HD怪物第%d圖來源、順序或檔名不符", i)
		}
		seen[p.SourceRGB] = true
		if p.Size[0] <= 0 || p.Size[0]%8 != 0 || p.Size[0] > 120 || p.Size[1] <= 0 || p.Size[1] > 74 {
			return nil, fmt.Errorf("HD怪物來源尺寸無效")
		}
		info, ok := m.Files[p.File]
		if !ok || !validHDHash(info.SHA256) || info.Width < p.Size[0]*6 || info.Height < p.Size[1]*6 || info.Width > 4096 || info.Height > 4096 {
			return nil, fmt.Errorf("HD怪物原生尺寸不足或雜湊無效")
		}
		if deferHDPixels(options) {
			pic, err := openHDImage(dir, p.File, m.Files[p.File])
			if err != nil {
				return nil, err
			}
			result.Images[i] = pic
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, p.File))
		if err != nil {
			return nil, err
		}
		if monsterHash(data) != info.SHA256 {
			return nil, fmt.Errorf("HD怪物PNG雜湊不符")
		}
		cfg, err := png.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		if cfg.Width != info.Width || cfg.Height != info.Height {
			return nil, fmt.Errorf("HD怪物PNG實際尺寸不符")
		}
		pic, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		pixels := make([]byte, 0, info.Width*info.Height*4)
		bb := image.Rectangle{}
		lo, hi := uint8(255), uint8(0)
		for y := 0; y < info.Height; y++ {
			for x := 0; x < info.Width; x++ {
				c := color.NRGBAModel.Convert(pic.At(x, y)).(color.NRGBA)
				pixels = append(pixels, c.R, c.G, c.B, c.A)
				if c.A < lo {
					lo = c.A
				}
				if c.A > hi {
					hi = c.A
				}
				if c.A >= 8 {
					bb = bb.Union(image.Rect(x, y, x+1, y+1))
				}
			}
		}
		if lo != 0 || hi < 250 || p.AlphaRange != [2]uint8{lo, hi} || bb.Empty() || bb.Min.X <= 0 || bb.Min.Y <= 0 || bb.Max.X >= info.Width || bb.Max.Y >= info.Height || p.AlphaBBox != [4]int{bb.Min.X, bb.Min.Y, bb.Max.X, bb.Max.Y} || monsterHash(pixels) != p.NativeRGBA {
			return nil, fmt.Errorf("HD怪物透明輪廓、邊界或原生RGBA不符")
		}
		result.Images[i] = pic
	}
	return result, nil
}

// ValidateSources 驗實際母體、每張PNG、原始RGBA與GPU可比的二值alpha。
func (m *HDMonsters) ValidateSources(lib *Library) error {
	if lib == nil || lib.Count("MONPICS") != 32 || lib.Manifest().SourceZip.SHA256 != m.Manifest.SourceZIP {
		return fmt.Errorf("HD怪物來源PNG庫數量或母體不符")
	}
	for _, s := range m.Manifest.Pictures {
		pic, meta, err := lib.Decode(s.Archive, s.Index)
		if err != nil {
			return err
		}
		if pic.Bounds().Size() != image.Pt(s.Size[0], s.Size[1]) || meta.SHA256 != s.SourcePNG {
			return fmt.Errorf("HD怪物來源尺寸或登記雜湊不符")
		}
		data, err := os.ReadFile(filepath.Join(lib.root, meta.File))
		if err != nil {
			return err
		}
		if monsterHash(data) != s.SourcePNG {
			return fmt.Errorf("HD怪物來源PNG雜湊不符")
		}
		rgb := make([]byte, 0, s.Size[0]*s.Size[1]*3)
		rgba := make([]byte, 0, s.Size[0]*s.Size[1]*4)
		b := pic.Bounds()
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				c := color.NRGBAModel.Convert(pic.At(x, y)).(color.NRGBA)
				if c.A != 0 && c.A != 255 || c.A == 0 && (c.R != 0 || c.G != 0 || c.B != 0) {
					return fmt.Errorf("HD怪物來源alpha或隱藏RGB契約不符")
				}
				rgb = append(rgb, c.R, c.G, c.B)
				rgba = append(rgba, c.R, c.G, c.B, c.A)
			}
		}
		if monsterHash(rgb) != s.SourceRGB || monsterHash(rgba) != s.RGBA {
			return fmt.Errorf("HD怪物來源RGB或RGBA不符")
		}
	}
	return nil
}
