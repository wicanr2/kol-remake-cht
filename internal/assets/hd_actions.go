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

type HDActionSource struct {
	HDPageSource
	Rect  [4]int `json:"rect"`
	RGBA  string `json:"native_rgba_sha256"`
	Alpha [4]int `json:"alpha_bbox"`
}
type HDActionsManifest struct {
	Schema      string                 `json:"schema"`
	Rights      string                 `json:"rights"`
	SourceZIP   string                 `json:"source_zip_sha256"`
	Image       string                 `json:"image"`
	Files       map[string]HDImageInfo `json:"files"`
	Sources     []HDActionSource       `json:"sources"`
	FontLicense string                 `json:"font_license_sha256"`
}
type HDActions struct {
	Image    image.Image
	Manifest HDActionsManifest
}

// HDActionIndices 是實際存在的68張24×20圖示。第28圖是徽章AND遮罩。
func HDActionIndices() []int {
	var ids []int
	for _, span := range [][2]int{{3, 27}, {31, 31}, {36, 69}, {74, 81}} {
		for i := span[0]; i <= span[1]; i++ {
			ids = append(ids, i)
		}
	}
	return ids
}

// LoadHDActions 全組驗證後才回傳；不提供部分圖集。
func LoadHDActions(dir string) (*HDActions, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return nil, err
	}
	var m HDActionsManifest
	if err = json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	if m.Schema != "kol-hd-actions/1" || m.Rights == "" || !validHDHash(m.SourceZIP) || m.Image != "actions.png" || len(m.Files) != 1 || len(m.Sources) != 68 || !validHDHash(m.FontLicense) {
		return nil, fmt.Errorf("HD動作圖示版本、來源、數量或授權無效")
	}
	license, err := os.ReadFile(filepath.Join(dir, "font-LICENSE.txt"))
	if err != nil {
		return nil, err
	}
	if fmt.Sprintf("%x", sha256.Sum256(license)) != m.FontLicense {
		return nil, fmt.Errorf("HD動作圖示字型授權雜湊不符")
	}
	info, ok := m.Files[m.Image]
	if !ok || info.Width != 1296 || info.Height != 960 || !validHDHash(info.SHA256) {
		return nil, fmt.Errorf("HD動作圖集尺寸或雜湊宣告無效")
	}
	data, err := os.ReadFile(filepath.Join(dir, m.Image))
	if err != nil {
		return nil, err
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != info.SHA256 {
		return nil, fmt.Errorf("HD動作圖集雜湊不符")
	}
	cfg, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	if cfg.Width != info.Width || cfg.Height != info.Height {
		return nil, fmt.Errorf("HD動作圖集實際尺寸不符")
	}
	pic, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	ids := HDActionIndices()
	seen := map[string]bool{}
	for i, s := range m.Sources {
		x, y := i%9*144, i/9*120
		if s.Archive != "KOLMPICS" || s.Index != ids[i] || !validHDHash(s.SourcePNG) || !validHDHash(s.SourceRGB) || seen[s.SourceRGB] || !validHDHash(s.RGBA) || s.Rect != [4]int{x, y, x + 144, y + 120} {
			return nil, fmt.Errorf("HD動作第%d格圖號、來源或矩形無效", i)
		}
		seen[s.SourceRGB] = true
		rgba := make([]byte, 0, 144*120*4)
		bb := image.Rectangle{}
		for py := 0; py < 120; py++ {
			for px := 0; px < 144; px++ {
				c := color.NRGBAModel.Convert(pic.At(x+px, y+py)).(color.NRGBA)
				rgba = append(rgba, c.R, c.G, c.B, c.A)
				if c.A >= 8 {
					bb = bb.Union(image.Rect(px, py, px+1, py+1))
				}
			}
		}
		if bb.Empty() || bb.Min.X <= 0 || bb.Min.Y <= 0 || bb.Max.X >= 144 || bb.Max.Y >= 120 || s.Alpha != [4]int{bb.Min.X, bb.Min.Y, bb.Max.X, bb.Max.Y} || fmt.Sprintf("%x", sha256.Sum256(rgba)) != s.RGBA {
			return nil, fmt.Errorf("HD動作第%d圖空白、接觸格緣或像素指紋不符", s.Index)
		}
	}
	return &HDActions{Image: pic, Manifest: m}, nil
}

// ValidateSources 在交換HD畫布前核對目前PNG庫，不以版號猜圖號範圍。
func (a *HDActions) ValidateSources(lib *Library) error {
	if lib == nil || lib.Count("KOLMPICS") != 71 || lib.Manifest().SourceZip.SHA256 != a.Manifest.SourceZIP {
		return fmt.Errorf("HD動作來源PNG庫數量或母體不符")
	}
	for _, s := range a.Manifest.Sources {
		pic, meta, err := lib.Decode(s.Archive, s.Index)
		if err != nil {
			return err
		}
		if pic.Bounds().Dx() != 24 || pic.Bounds().Dy() != 20 || meta.SHA256 != s.SourcePNG {
			return fmt.Errorf("HD動作第%d圖來源尺寸或登記雜湊不符", s.Index)
		}
		data, err := os.ReadFile(filepath.Join(lib.root, meta.File))
		if err != nil {
			return err
		}
		if fmt.Sprintf("%x", sha256.Sum256(data)) != s.SourcePNG {
			return fmt.Errorf("HD動作第%d圖來源PNG雜湊不符", s.Index)
		}
		rgb := make([]byte, 0, 24*20*3)
		b := pic.Bounds()
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				c := color.NRGBAModel.Convert(pic.At(x, y)).(color.NRGBA)
				rgb = append(rgb, c.R, c.G, c.B)
			}
		}
		if fmt.Sprintf("%x", sha256.Sum256(rgb)) != s.SourceRGB {
			return fmt.Errorf("HD動作第%d圖來源RGB不符", s.Index)
		}
	}
	return nil
}
