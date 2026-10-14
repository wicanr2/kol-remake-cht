package assets

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
)

type HDTemplatePortrait struct {
	HDPageSource
	File     string `json:"file"`
	Rect     [4]int `json:"rect"`
	FrameRGB string `json:"frame_rgb_sha256"`
}

type HDTemplatesManifest struct {
	Schema        string                 `json:"schema"`
	Rights        string                 `json:"rights"`
	SourceZIP     string                 `json:"source_zip_sha256"`
	SourceCount   int                    `json:"source_count"`
	ApprovedAtlas string                 `json:"approved_atlas_sha256"`
	Preserved     []HDPageSource         `json:"preserved"`
	Files         map[string]HDImageInfo `json:"files"`
	Portraits     []HDTemplatePortrait   `json:"portraits"`
}

type HDTemplates struct {
	Images   map[string]image.Image
	Manifest HDTemplatesManifest
}

// LoadHDTemplates 先驗證80個來源的完整覆蓋，再原子載入74張新增手繪畫格。
func LoadHDTemplates(dir string) (*HDTemplates, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return nil, err
	}
	var m HDTemplatesManifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	if m.Schema != "kol-hd-templates/1" || m.Rights == "" || !validHDHash(m.SourceZIP) ||
		m.SourceCount != 80 || len(m.Preserved) != 6 || len(m.Portraits) != 74 || len(m.Files) != 20 ||
		m.ApprovedAtlas != "77e3d3e3a00dfe6144fe278eeda6c2b28b414e470f9ffb2033dfedc8c97cf016" {
		return nil, fmt.Errorf("HD角色模板版本、來源、80個索引或核准圖集不符")
	}
	indices, keys := map[int]bool{}, map[string]bool{}
	validSource := func(s HDPageSource) bool {
		if s.Archive != "CHARPICS" || s.Index < 0 || s.Index >= 80 || indices[s.Index] ||
			keys[s.SourceRGB] || !validHDHash(s.SourceRGB) || !validHDHash(s.SourcePNG) {
			return false
		}
		indices[s.Index], keys[s.SourceRGB] = true, true
		return true
	}
	approved := [...]int{47, 43, 21, 32, 35, 54}
	for i, s := range m.Preserved {
		if s.Index != approved[i] || !validSource(s) {
			return nil, fmt.Errorf("HD核准角色順序或來源指紋不符")
		}
	}
	for i := 0; i < 20; i++ {
		name := fmt.Sprintf("quartet-%02d.png", i)
		info, ok := m.Files[name]
		if !ok || !validHDHash(info.SHA256) || info.Width < 736 || info.Height < 736 || info.Width > 4096 || info.Height > 4096 {
			return nil, fmt.Errorf("HD角色圖集%s宣告尺寸、檔名或雜湊不符", name)
		}
	}
	for _, p := range m.Portraits {
		if !validSource(p.HDPageSource) || p.File != fmt.Sprintf("quartet-%02d.png", p.Index/4) || !validHDHash(p.FrameRGB) {
			return nil, fmt.Errorf("HD角色模板來源、重複索引、檔名或畫格指紋不符")
		}
		info := m.Files[p.File]
		r := image.Rect(p.Rect[0], p.Rect[1], p.Rect[2], p.Rect[3])
		if r.Min.X != p.Rect[0] || r.Min.Y != p.Rect[1] || r.Max.X != p.Rect[2] || r.Max.Y != p.Rect[3] ||
			r.Min.X < 0 || r.Min.Y < 0 || r.Max.X > info.Width || r.Max.Y > info.Height || r.Dx() < 368 || r.Dy() < 368 {
			return nil, fmt.Errorf("HD角色%d畫格越界或低於原生368像素", p.Index)
		}
	}
	if len(indices) != 80 {
		return nil, fmt.Errorf("HD角色來源索引有缺漏")
	}
	result := &HDTemplates{Images: map[string]image.Image{}, Manifest: m}
	for i := 0; i < 20; i++ {
		name := fmt.Sprintf("quartet-%02d.png", i)
		info := m.Files[name]
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != info.SHA256 {
			return nil, fmt.Errorf("HD角色圖集%s雜湊不符", name)
		}
		conf, err := png.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		if conf.Width != info.Width || conf.Height != info.Height {
			return nil, fmt.Errorf("HD角色圖集%s實際尺寸不符", name)
		}
		pic, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		if solid, ok := pic.(interface{ Opaque() bool }); !ok || !solid.Opaque() {
			return nil, fmt.Errorf("HD角色圖集%s須為不透明背景", name)
		}
		result.Images[name] = pic
	}
	for _, p := range m.Portraits {
		pic := result.Images[p.File]
		buf := make([]byte, 0, (p.Rect[2]-p.Rect[0])*(p.Rect[3]-p.Rect[1])*3)
		for y := p.Rect[1]; y < p.Rect[3]; y++ {
			for x := p.Rect[0]; x < p.Rect[2]; x++ {
				r, g, b, _ := pic.At(x, y).RGBA()
				buf = append(buf, byte(r>>8), byte(g>>8), byte(b>>8))
			}
		}
		sum := sha256.Sum256(buf)
		if hex.EncodeToString(sum[:]) != p.FrameRGB {
			return nil, fmt.Errorf("HD角色%d裁取指紋不符", p.Index)
		}
	}
	return result, nil
}
