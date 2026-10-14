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

// HDEndingPanel 對應 SPEC-015 的原圖號，來源指紋只供呈現驗證。
type HDEndingPanel struct {
	Index     int    `json:"index"`
	File      string `json:"file"`
	SourceRGB string `json:"source_rgb_sha256"`
	SourcePNG string `json:"source_png_sha256"`
}

type HDEndingManifest struct {
	Schema    string                 `json:"schema"`
	Rights    string                 `json:"rights"`
	SourceZIP string                 `json:"source_zip_sha256"`
	Files     map[string]HDImageInfo `json:"files"`
	Panels    []HDEndingPanel        `json:"panels"`
}

type HDEnding struct {
	Images    map[int]image.Image
	SourceRGB map[int]string
}

// LoadHDEnding 驗證七段的完整子組，不交回部分圖片。
func LoadHDEnding(dir string) (*HDEnding, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return nil, err
	}
	var m HDEndingManifest
	if err = json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	if m.Schema != "kol-hd-ending/1" || m.Rights == "" || !validHDHash(m.SourceZIP) || len(m.Files) != 7 || len(m.Panels) != 7 {
		return nil, fmt.Errorf("HD結局素材版本、來源或七段圖數不符")
	}
	h := &HDEnding{Images: map[int]image.Image{}, SourceRGB: map[int]string{}}
	seen := map[string]bool{}
	for _, p := range m.Panels {
		if p.Index < 0 || p.Index >= 7 || h.Images[p.Index] != nil ||
			p.File != fmt.Sprintf("endpics-%03d-painted.png", p.Index) ||
			!validHDHash(p.SourceRGB) || !validHDHash(p.SourcePNG) || seen[p.SourceRGB] {
			return nil, fmt.Errorf("HD結局圖號、檔名或來源指紋無效")
		}
		seen[p.SourceRGB] = true
		info, ok := m.Files[p.File]
		if !ok || !validHDHash(info.SHA256) || info.Width < 1 || info.Height < 1 || info.Width > 4096 || info.Height > 4096 {
			return nil, fmt.Errorf("HD結局 %s 宣告尺寸或雜湊無效", p.File)
		}
		data, e := os.ReadFile(filepath.Join(dir, p.File))
		if e != nil {
			return nil, e
		}
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != info.SHA256 {
			return nil, fmt.Errorf("HD結局 %s 雜湊不符", p.File)
		}
		conf, e := png.DecodeConfig(bytes.NewReader(data))
		if e != nil {
			return nil, e
		}
		if conf.Width != info.Width || conf.Height != info.Height {
			return nil, fmt.Errorf("HD結局 %s 實際尺寸不符", p.File)
		}
		im, e := png.Decode(bytes.NewReader(data))
		if e != nil {
			return nil, e
		}
		h.Images[p.Index], h.SourceRGB[p.Index] = im, p.SourceRGB
	}
	return h, nil
}
