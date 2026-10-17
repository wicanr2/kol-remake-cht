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

// HDManifest 是 SPEC-039 首批純呈現素材，與規則及保存資料分開。
type HDManifest struct {
	Schema      string                 `json:"schema"`
	Rights      string                 `json:"rights"`
	SourceZIP   string                 `json:"source_zip_sha256"`
	Files       map[string]HDImageInfo `json:"files"`
	FacadeRGB   string                 `json:"facade_rgb_sha256"`
	PortraitRGB []string               `json:"portrait_rgb_sha256"`
	Sources     map[string]string      `json:"sources"`
}

type HDImageInfo struct {
	SHA256 string `json:"sha256"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type HDLibrary struct {
	Manifest HDManifest
	Facade   image.Image
	Party    image.Image
}

func validHDHash(s string) bool {
	b, err := hex.DecodeString(s)
	return err == nil && len(b) == sha256.Size
}

// LoadHD 驗完整組後才交給呈現層；不讀原版archive，也不改既有Library。
func LoadHD(dir string, options ...HDLoadOption) (*HDLibrary, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return nil, err
	}
	var m HDManifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	if m.Schema != "kol-hd-first-slice/1" || m.Rights == "" || !validHDHash(m.SourceZIP) ||
		!validHDHash(m.FacadeRGB) || len(m.PortraitRGB) != 6 || len(m.Files) != 2 || len(m.Sources) == 0 {
		return nil, fmt.Errorf("HD素材包的版本、來源或圖數不符SPEC-039")
	}
	seen := map[string]bool{}
	for _, h := range m.PortraitRGB {
		if !validHDHash(h) || seen[h] {
			return nil, fmt.Errorf("HD人物指紋無效或重複")
		}
		seen[h] = true
	}
	for _, h := range m.Sources {
		if !validHDHash(h) {
			return nil, fmt.Errorf("HD來源雜湊無效")
		}
	}
	load := func(name string, width, height int) (image.Image, error) {
		info, ok := m.Files[name]
		if !ok || !validHDHash(info.SHA256) || info.Width != width || info.Height != height {
			return nil, fmt.Errorf("HD %s 宣告尺寸不符", name)
		}
		if deferHDPixels(options) {
			return openHDImage(dir, name, info)
		}
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		h := sha256.Sum256(data)
		if hex.EncodeToString(h[:]) != info.SHA256 {
			return nil, fmt.Errorf("HD %s 雜湊不符", name)
		}
		config, err := png.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		if config.Width != width || config.Height != height {
			return nil, fmt.Errorf("HD %s 實際尺寸不符", name)
		}
		img, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		if img.Bounds().Dx() != width || img.Bounds().Dy() != height {
			return nil, fmt.Errorf("HD %s 實際尺寸不符", name)
		}
		return img, nil
	}
	facade, err := load("inn-painted-v1.png", 1774, 887)
	if err != nil {
		return nil, err
	}
	party, err := load("party-painted-v1.png", 1536, 1024)
	if err != nil {
		return nil, err
	}
	return &HDLibrary{Manifest: m, Facade: facade, Party: party}, nil
}
