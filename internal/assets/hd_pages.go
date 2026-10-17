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

type HDPageSource struct {
	Archive   string `json:"archive"`
	Index     int    `json:"index"`
	SourceRGB string `json:"source_rgb_sha256"`
	SourcePNG string `json:"source_png_sha256"`
}

type HDPagePortrait struct {
	HDPageSource
	File string `json:"file"`
}

type HDPagesManifest struct {
	Schema     string                 `json:"schema"`
	Rights     string                 `json:"rights"`
	SourceZIP  string                 `json:"source_zip_sha256"`
	Files      map[string]HDImageInfo `json:"files"`
	Book       string                 `json:"book"`
	Books      []HDPageSource         `json:"books"`
	Background HDPageSource           `json:"background"`
	Portraits  []HDPagePortrait       `json:"portraits"`
}

type HDPages struct {
	Book      image.Image
	Portraits map[string]image.Image
	Manifest  HDPagesManifest
}

// Sources 是繪製前須核對的原PNG圖號，不以版本參數猜素材範圍。
func (p *HDPages) Sources() []HDPageSource {
	out := append([]HDPageSource{}, p.Manifest.Books...)
	out = append(out, p.Manifest.Background)
	for _, portrait := range p.Manifest.Portraits {
		out = append(out, portrait.HDPageSource)
	}
	return out
}

// LoadHDPages 原子載入本批書頁及三個服務頭像。原始圖像指紋在繪製前另核對。
func LoadHDPages(dir string, options ...HDLoadOption) (*HDPages, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return nil, err
	}
	var m HDPagesManifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	if m.Schema != "kol-hd-pages/1" || m.Rights == "" || !validHDHash(m.SourceZIP) ||
		m.Book != "book-open-painted.png" || len(m.Files) != 4 || len(m.Books) != 2 || len(m.Portraits) != 3 {
		return nil, fmt.Errorf("HD頁面素材版本、來源或四張圖片數量不符")
	}
	keys := map[string]bool{}
	validSource := func(p HDPageSource) bool {
		if !validHDHash(p.SourceRGB) || !validHDHash(p.SourcePNG) || keys[p.SourceRGB] {
			return false
		}
		keys[p.SourceRGB] = true
		return true
	}
	seenBook := map[int]bool{}
	for _, p := range m.Books {
		if p.Archive != "BOOKPICS" || p.Index != 1 && p.Index != 2 || seenBook[p.Index] || !validSource(p) {
			return nil, fmt.Errorf("HD書頁來源圖號或指紋無效")
		}
		seenBook[p.Index] = true
	}
	if m.Background.Archive != "KOLMPICS" || m.Background.Index != 0 || !validSource(m.Background) {
		return nil, fmt.Errorf("HD介面背景來源無效")
	}
	seenPortrait := map[int]bool{}
	for _, p := range m.Portraits {
		if p.Archive != "TWN1PICS" || p.Index != 12 && p.Index != 18 && p.Index != 20 ||
			seenPortrait[p.Index] || p.File != fmt.Sprintf("twn1pics-%03d-painted.png", p.Index) || !validSource(p.HDPageSource) {
			return nil, fmt.Errorf("HD服務頭像來源、圖號或檔名無效")
		}
		seenPortrait[p.Index] = true
	}
	load := func(name string) (image.Image, error) {
		info, ok := m.Files[name]
		if !ok || !validHDHash(info.SHA256) || info.Width < 1 || info.Height < 1 || info.Width > 4096 || info.Height > 4096 {
			return nil, fmt.Errorf("HD頁面 %s 宣告尺寸或雜湊無效", name)
		}
		if deferHDPixels(options) {
			return openHDImage(dir, name, info)
		}
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != info.SHA256 {
			return nil, fmt.Errorf("HD頁面 %s 雜湊不符", name)
		}
		conf, err := png.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		if conf.Width != info.Width || conf.Height != info.Height {
			return nil, fmt.Errorf("HD頁面 %s 實際尺寸不符", name)
		}
		return png.Decode(bytes.NewReader(data))
	}
	book, err := load(m.Book)
	if err != nil {
		return nil, err
	}
	p := &HDPages{Book: book, Portraits: map[string]image.Image{}, Manifest: m}
	for _, portrait := range m.Portraits {
		im, err := load(portrait.File)
		if err != nil {
			return nil, err
		}
		p.Portraits[portrait.SourceRGB] = im
	}
	return p, nil
}
