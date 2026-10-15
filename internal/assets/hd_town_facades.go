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
	"reflect"
	"strings"
)

type HDTownFacadePicture struct {
	HDPageSource
	File       string `json:"file"`
	Size       [2]int `json:"size"`
	RGBA       string `json:"source_rgba_sha256"`
	NativeRGBA string `json:"native_rgba_sha256"`
}
type HDTownFacadeSign struct {
	Picture    string    `json:"picture"`
	NativeSize [2]int    `json:"native_size"`
	FontSize   int       `json:"font_size"`
	LineHeight int       `json:"line_height"`
	Full       [3][4]int `json:"full"`
	Clip       [3][4]int `json:"clip"`
}
type HDTownFacadesManifest struct {
	Schema        string                 `json:"schema"`
	Rights        string                 `json:"rights"`
	SourceZIP     string                 `json:"source_zip_sha256"`
	Count         int                    `json:"source_count"`
	ArchiveCounts map[string]int         `json:"archive_counts"`
	Files         map[string]HDImageInfo `json:"files"`
	Pictures      []HDTownFacadePicture  `json:"pictures"`
	GenerationSHA string                 `json:"generation_sha256"`
	Sign          HDTownFacadeSign       `json:"residence_sign"`
}
type HDTownFacades struct {
	Manifest HDTownFacadesManifest
	Images   []image.Image
}

var hdTownFacadeCounts = []int{33, 29, 25, 25, 29, 25, 45}

const preservedTownFacadeSHA = "0ec14587368570c7c84aa20b695335b022c7da8ef1b8f3a07255b1b73adb2b9c"

func hdTownFacadeSources() []HDPageSource {
	out := []HDPageSource{{Archive: "MISCPICS", Index: 0}}
	for t, count := range hdTownFacadeCounts {
		for index := 3; index <= count; index += 2 {
			out = append(out, HDPageSource{Archive: fmt.Sprintf("TWN%dPICS", t+1), Index: index})
		}
	}
	return out
}
func hdTownFacadeSignContract() HDTownFacadeSign {
	return HDTownFacadeSign{Picture: "twn5pics-019.png", NativeSize: [2]int{1774, 887}, FontSize: 32, LineHeight: 34,
		Full: [3][4]int{{770, 519, 1100, 625}, {-245, 626, 85, 740}, {1716, 519, 2046, 625}},
		Clip: [3][4]int{{770, 519, 1100, 625}, {0, 626, 85, 740}, {1716, 519, 1774, 625}}}
}

// LoadHDTownFacades validates all103 bindings before decoding or returning a group.
func LoadHDTownFacades(dir string) (*HDTownFacades, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return nil, err
	}
	var m HDTownFacadesManifest
	if err = json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	wanted := hdTownFacadeSources()
	if m.Schema != "kol-hd-town-facades/1" || m.Rights == "" || !validHDHash(m.SourceZIP) || m.Count != 103 || len(m.Files) != 103 || len(m.Pictures) != 103 || len(m.ArchiveCounts) != 8 || !validHDHash(m.GenerationSHA) || !reflect.DeepEqual(m.Sign, hdTownFacadeSignContract()) {
		return nil, fmt.Errorf("HD城鎮門面版本、完整數量、來源或告示綁定不符")
	}
	if m.ArchiveCounts["MISCPICS"] != 5 {
		return nil, fmt.Errorf("HD城門來源數量不符")
	}
	for t, count := range hdTownFacadeCounts {
		if m.ArchiveCounts[fmt.Sprintf("TWN%dPICS", t+1)] != count {
			return nil, fmt.Errorf("HD城鎮來源數量不符")
		}
	}
	seen := map[string]bool{}
	for i, p := range m.Pictures {
		file := fmt.Sprintf("%s-%03d.png", strings.ToLower(p.Archive), p.Index)
		if p.Archive != wanted[i].Archive || p.Index != wanted[i].Index || p.File != file || p.Size != [2]int{192, 96} || !validHDHash(p.SourcePNG) || !validHDHash(p.SourceRGB) || !validHDHash(p.RGBA) || !validHDHash(p.NativeRGBA) || seen[p.SourceRGB] {
			return nil, fmt.Errorf("HD城鎮門面圖號、來源、尺寸或檔名不符")
		}
		seen[p.SourceRGB] = true
		info, ok := m.Files[file]
		if !ok || !validHDHash(info.SHA256) || info.Width < 1152 || info.Height < 576 || info.Width != info.Height*2 || info.Width > 4096 {
			return nil, fmt.Errorf("HD城鎮門面原生尺寸或雜湊不符")
		}
		if p.Archive == "TWN7PICS" && p.Index == 13 && info.SHA256 != preservedTownFacadeSHA {
			return nil, fmt.Errorf("HD原核准旅店遭改動")
		}
		if file == m.Sign.Picture && (info.Width != m.Sign.NativeSize[0] || info.Height != m.Sign.NativeSize[1]) {
			return nil, fmt.Errorf("HD住宅告示原生尺寸不符")
		}
	}
	generation, err := os.ReadFile(filepath.Join(dir, "generation.json"))
	if err != nil {
		return nil, err
	}
	if monsterHash(generation) != m.GenerationSHA {
		return nil, fmt.Errorf("HD城鎮門面生成清冊雜湊不符")
	}
	result := &HDTownFacades{Manifest: m}
	for _, p := range m.Pictures {
		info := m.Files[p.File]
		data, err := os.ReadFile(filepath.Join(dir, p.File))
		if err != nil {
			return nil, err
		}
		if monsterHash(data) != info.SHA256 {
			return nil, fmt.Errorf("HD城鎮門面PNG雜湊不符")
		}
		cfg, err := png.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		if cfg.Width != info.Width || cfg.Height != info.Height {
			return nil, fmt.Errorf("HD城鎮門面PNG實際尺寸不符")
		}
		pic, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		rgba := image.NewNRGBA(image.Rect(0, 0, info.Width, info.Height))
		draw.Draw(rgba, rgba.Bounds(), pic, pic.Bounds().Min, draw.Src)
		for i := 3; i < len(rgba.Pix); i += 4 {
			if rgba.Pix[i] != 255 {
				return nil, fmt.Errorf("HD城鎮門面須完整不透明")
			}
		}
		if monsterHash(rgba.Pix) != p.NativeRGBA {
			return nil, fmt.Errorf("HD城鎮門面原生RGBA不符")
		}
		result.Images = append(result.Images, pic)
	}
	return result, nil
}

// ValidateSources uses real PNG membership and counts, including the missing index1.
func (m *HDTownFacades) ValidateSources(lib *Library) error {
	if lib == nil || lib.Manifest().SourceZip.SHA256 != m.Manifest.SourceZIP {
		return fmt.Errorf("HD城鎮門面來源母體不符")
	}
	for archive, count := range m.Manifest.ArchiveCounts {
		if lib.Count(archive) != count {
			return fmt.Errorf("HD城鎮門面實際來源數量不符")
		}
	}
	for _, p := range m.Manifest.Pictures {
		pic, meta, err := lib.Decode(p.Archive, p.Index)
		if err != nil {
			return err
		}
		if pic.Bounds().Size() != image.Pt(192, 96) || meta.SHA256 != p.SourcePNG {
			return fmt.Errorf("HD城鎮門面來源尺寸或登記雜湊不符")
		}
		raw, err := os.ReadFile(filepath.Join(lib.root, meta.File))
		if err != nil {
			return err
		}
		if monsterHash(raw) != p.SourcePNG {
			return fmt.Errorf("HD城鎮門面來源PNG雜湊不符")
		}
		rgba := image.NewNRGBA(image.Rect(0, 0, 192, 96))
		draw.Draw(rgba, rgba.Bounds(), pic, pic.Bounds().Min, draw.Src)
		rgb := make([]byte, 0, 192*96*3)
		for i := 0; i < len(rgba.Pix); i += 4 {
			if rgba.Pix[i+3] != 255 {
				return fmt.Errorf("HD城鎮門面來源須不透明")
			}
			rgb = append(rgb, rgba.Pix[i:i+3]...)
		}
		if monsterHash(rgb) != p.SourceRGB || monsterHash(rgba.Pix) != p.RGBA {
			return fmt.Errorf("HD城鎮門面來源RGB或RGBA不符")
		}
	}
	return nil
}
