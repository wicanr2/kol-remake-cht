package assets

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
)

type HDTownCrestPicture struct {
	HDPageSource
	File       string `json:"file"`
	Size       [2]int `json:"size"`
	RGBA       string `json:"source_rgba_sha256"`
	NativeRGBA string `json:"native_rgba_sha256"`
}
type HDTownCrestsManifest struct {
	Schema        string                 `json:"schema"`
	Rights        string                 `json:"rights"`
	SourceZIP     string                 `json:"source_zip_sha256"`
	Count         int                    `json:"source_count"`
	ArchiveCounts map[string]int         `json:"archive_counts"`
	Files         map[string]HDImageInfo `json:"files"`
	Pictures      []HDTownCrestPicture   `json:"pictures"`
	GenerationSHA string                 `json:"generation_sha256"`
}
type HDTownCrests struct {
	Manifest HDTownCrestsManifest
	Images   []image.Image
}

const preservedTownCrestSHA = "3e9fa444a362271b2978c508f7a40c04ba9a14cf212d55cd2ff5fcf9683d30f4"

// LoadHDTownCrests rejects an incomplete group before returning any image.
func LoadHDTownCrests(dir string, options ...HDLoadOption) (*HDTownCrests, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return nil, err
	}
	var m HDTownCrestsManifest
	if err = json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	if m.Schema != "kol-hd-town-crests/1" || m.Rights == "" || !validHDHash(m.SourceZIP) || m.Count != 7 || len(m.Pictures) != 7 || len(m.Files) != 7 || len(m.ArchiveCounts) != 7 || !validHDHash(m.GenerationSHA) {
		return nil, fmt.Errorf("HD七城徽章版本、數量或來源不符")
	}
	seen := map[string]bool{}
	for i, p := range m.Pictures {
		archive, file := fmt.Sprintf("TWN%dPICS", i+1), fmt.Sprintf("twn%dpics-000.png", i+1)
		info, ok := m.Files[file]
		if p.Archive != archive || p.Index != 0 || p.File != file || p.Size != [2]int{56, 57} || m.ArchiveCounts[archive] != hdTownFacadeCounts[i] || !validHDHash(p.SourcePNG) || !validHDHash(p.SourceRGB) || !validHDHash(p.RGBA) || !validHDHash(p.NativeRGBA) || seen[p.SourceRGB] || !ok || info.Width != 320 || info.Height != 320 || !validHDHash(info.SHA256) {
			return nil, fmt.Errorf("HD七城徽章綁定、尺寸、來源或原生檔不符")
		}
		if i == 6 && info.SHA256 != preservedTownCrestSHA {
			return nil, fmt.Errorf("HD原核准霍布城徽章遭改動")
		}
		seen[p.SourceRGB] = true
	}
	generation, err := os.ReadFile(filepath.Join(dir, "generation.json"))
	if err != nil {
		return nil, err
	}
	if monsterHash(generation) != m.GenerationSHA {
		return nil, fmt.Errorf("HD七城徽章生成清冊不符")
	}
	result := &HDTownCrests{Manifest: m}
	for _, p := range m.Pictures {
		if deferHDPixels(options) {
			pic, err := openHDImage(dir, p.File, m.Files[p.File])
			if err != nil {
				return nil, err
			}
			result.Images = append(result.Images, pic)
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, p.File))
		if err != nil {
			return nil, err
		}
		if monsterHash(data) != m.Files[p.File].SHA256 {
			return nil, fmt.Errorf("HD七城徽章PNG雜湊不符")
		}
		cfg, err := png.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		if cfg.Width != 320 || cfg.Height != 320 {
			return nil, fmt.Errorf("HD七城徽章實際原生尺寸不符")
		}
		pic, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		rgba := image.NewNRGBA(image.Rect(0, 0, 320, 320))
		draw.Draw(rgba, rgba.Bounds(), pic, pic.Bounds().Min, draw.Src)
		if monsterHash(rgba.Pix) != p.NativeRGBA || rgba.NRGBAAt(0, 0).A != 0 || rgba.NRGBAAt(160, 160).A != 255 {
			return nil, fmt.Errorf("HD七城徽章原生RGBA或透明框不符")
		}
		result.Images = append(result.Images, pic)
	}
	return result, nil
}

func (m *HDTownCrests) ValidateSources(lib *Library) error {
	if lib == nil || lib.Manifest().SourceZip.SHA256 != m.Manifest.SourceZIP {
		return fmt.Errorf("HD七城徽章來源母體不符")
	}
	for archive, count := range m.Manifest.ArchiveCounts {
		if lib.Count(archive) != count {
			return fmt.Errorf("HD七城徽章實際來源數量不符")
		}
	}
	for _, p := range m.Manifest.Pictures {
		pic, meta, err := lib.Decode(p.Archive, p.Index)
		if err != nil {
			return err
		}
		if pic.Bounds().Size() != image.Pt(56, 57) || meta.SHA256 != p.SourcePNG {
			return fmt.Errorf("HD七城徽章來源尺寸或登記雜湊不符")
		}
		raw, err := os.ReadFile(filepath.Join(lib.root, meta.File))
		if err != nil {
			return err
		}
		if monsterHash(raw) != p.SourcePNG {
			return fmt.Errorf("HD七城徽章來源PNG不符")
		}
		rgba := image.NewNRGBA(image.Rect(0, 0, 56, 57))
		draw.Draw(rgba, rgba.Bounds(), pic, pic.Bounds().Min, draw.Src)
		rgb := make([]byte, 0, 56*57*3)
		for i := 0; i < len(rgba.Pix); i += 4 {
			a := rgba.Pix[i+3]
			if (a != 0 && a != 255) || (a == 0 && (rgba.Pix[i] != 0 || rgba.Pix[i+1] != 0 || rgba.Pix[i+2] != 0)) {
				return fmt.Errorf("HD七城徽章來源透明遮罩不符")
			}
			rgb = append(rgb, rgba.Pix[i:i+3]...)
		}
		if monsterHash(rgb) != p.SourceRGB || monsterHash(rgba.Pix) != p.RGBA {
			return fmt.Errorf("HD七城徽章來源RGB或RGBA不符")
		}
	}
	return nil
}
