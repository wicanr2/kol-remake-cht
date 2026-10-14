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

type HDScrollManifest struct {
	Schema    string                 `json:"schema"`
	Rights    string                 `json:"rights"`
	SourceZIP string                 `json:"source_zip_sha256"`
	Image     string                 `json:"image"`
	Source    HDPageSource           `json:"source"`
	Files     map[string]HDImageInfo `json:"files"`
}

type HDScroll struct {
	Image    image.Image
	Manifest HDScrollManifest
}

// LoadHDScroll 驗證單張完整不透明卷軸；來源PNG指紋在繪製前另核對。
func LoadHDScroll(dir string) (*HDScroll, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return nil, err
	}
	var m HDScrollManifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	if m.Schema != "kol-hd-scroll/1" || m.Rights == "" || !validHDHash(m.SourceZIP) ||
		m.Image != "spell-scroll-painted.png" || len(m.Files) != 1 ||
		m.Source.Archive != "KOLPICS" || m.Source.Index != 1 ||
		!validHDHash(m.Source.SourceRGB) || !validHDHash(m.Source.SourcePNG) {
		return nil, fmt.Errorf("HD法術卷軸版本、來源、圖號或單張圖片數量不符")
	}
	info, ok := m.Files[m.Image]
	if !ok || !validHDHash(info.SHA256) || info.Width < 1 || info.Height < 1 || info.Width > 4096 || info.Height > 4096 {
		return nil, fmt.Errorf("HD法術卷軸宣告尺寸或雜湊無效")
	}
	data, err := os.ReadFile(filepath.Join(dir, m.Image))
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(data)
	if hex.EncodeToString(sum[:]) != info.SHA256 {
		return nil, fmt.Errorf("HD法術卷軸雜湊不符")
	}
	conf, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	if conf.Width != info.Width || conf.Height != info.Height {
		return nil, fmt.Errorf("HD法術卷軸實際尺寸不符")
	}
	pic, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	if solid, ok := pic.(interface{ Opaque() bool }); !ok || !solid.Opaque() {
		return nil, fmt.Errorf("HD法術卷軸須為完整不透明背景")
	}
	return &HDScroll{Image: pic, Manifest: m}, nil
}
