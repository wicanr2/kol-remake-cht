// Package assets 載入由 tools/extract_ega_assets.py 產生的 runtime PNG。
//
// runtime 只讀這個目錄下的 manifest.json 與 PNG。任何原版 archive、executable 或
// 其他副檔名都會被拒絕，避免 remake 在執行期間又長出對原版格式的依賴。
package assets

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
)

// Manifest 是抽取器輸出的 provenance 索引。
type Manifest struct {
	Schema     string    `json:"schema"`
	Tool       string    `json:"tool"`
	SourceZip  SourceZip `json:"source_zip"`
	PNGCount   int       `json:"png_count"`
	ArchiveNum int       `json:"archive_count"`
	Archives   []Archive `json:"archives"`
}

type SourceZip struct {
	Name   string `json:"name"`
	SHA256 string `json:"sha256"`
}

type Archive struct {
	Name   string  `json:"archive"`
	Images []Image `json:"images"`
}

type Image struct {
	UnitIndex int    `json:"unit_index"`
	File      string `json:"file"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	SHA256    string `json:"png_sha256"`
}

// Library 是已載入的資產目錄。
type Library struct {
	root     string
	manifest Manifest
	index    map[string]map[int]Image
}

// Load 讀取 manifest 並建立索引；不預先解碼任何 PNG。
func Load(root string) (*Library, error) {
	raw, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		return nil, fmt.Errorf("讀取 manifest 失敗：%w", err)
	}
	var m Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("解析 manifest 失敗：%w", err)
	}
	if !strings.HasPrefix(m.Schema, "kol-ega-production-manifest-") {
		return nil, fmt.Errorf("非預期的 manifest schema：%s", m.Schema)
	}
	lib := &Library{root: root, manifest: m, index: map[string]map[int]Image{}}
	for _, a := range m.Archives {
		byIndex := make(map[int]Image, len(a.Images))
		for _, img := range a.Images {
			byIndex[img.UnitIndex] = img
		}
		lib.index[a.Name] = byIndex
	}
	return lib, nil
}

// Manifest 回傳已載入的索引內容。
func (l *Library) Manifest() Manifest { return l.manifest }

// Decode 解碼指定 archive 的圖號，回傳 Go 標準 image.Image。
func (l *Library) Decode(archive string, index int) (image.Image, Image, error) {
	byIndex, ok := l.index[archive]
	if !ok {
		return nil, Image{}, fmt.Errorf("manifest 沒有 archive %s", archive)
	}
	meta, ok := byIndex[index]
	if !ok {
		return nil, Image{}, fmt.Errorf("%s 沒有圖號 %d", archive, index)
	}
	path := filepath.Join(l.root, filepath.FromSlash(meta.File))
	if strings.ToLower(filepath.Ext(path)) != ".png" {
		return nil, Image{}, fmt.Errorf("拒絕載入非 PNG 檔：%s", meta.File)
	}
	rel, err := filepath.Rel(l.root, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		return nil, Image{}, fmt.Errorf("拒絕載入資產目錄之外的檔案：%s", meta.File)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, Image{}, err
	}
	defer f.Close()
	img, format, err := image.Decode(f)
	if err != nil {
		return nil, Image{}, fmt.Errorf("解碼 %s 失敗：%w", meta.File, err)
	}
	if format != "png" {
		return nil, Image{}, fmt.Errorf("%s 不是 PNG（%s）", meta.File, format)
	}
	b := img.Bounds()
	if b.Dx() != meta.Width || b.Dy() != meta.Height {
		return nil, Image{}, fmt.Errorf("%s 尺寸與 manifest 不符：%dx%d != %dx%d",
			meta.File, b.Dx(), b.Dy(), meta.Width, meta.Height)
	}
	return img, meta, nil
}

// Count 回傳 archive 的圖號數量。
func (l *Library) Count(archive string) int { return len(l.index[archive]) }
