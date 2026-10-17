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
	"regexp"
)

// HDTownMapSource 記錄實際地圖與圖集的來源，不以版本推測數量。
type HDTownMapSource struct {
	Archive   string       `json:"archive"`
	Count     int          `json:"archive_count"`
	Grid      HDPageSource `json:"grid"`
	Atlas     HDPageSource `json:"atlas"`
	GridRGBA  string       `json:"grid_rgba_sha256"`
	AtlasRGBA string       `json:"atlas_rgba_sha256"`
}

type HDTownMapRegion struct {
	ID            string `json:"id"`
	Archive       string `json:"archive"`
	File          string `json:"file"`
	SourceBounds  [4]int `json:"source_bounds"`
	VisibleBounds [4]int `json:"visible_bounds"`
	NativeRGBA    string `json:"native_rgba_sha256"`
}

type HDTownMapsManifest struct {
	Schema        string                 `json:"schema"`
	Rights        string                 `json:"rights"`
	SourceZIP     string                 `json:"source_zip_sha256"`
	Maps          []HDTownMapSource      `json:"maps"`
	Regions       []HDTownMapRegion      `json:"regions"`
	Files         map[string]HDImageInfo `json:"files"`
	GenerationSHA string                 `json:"generation_sha256"`
}

type HDTownMaps struct {
	Manifest HDTownMapsManifest
	Images   []image.Image
	Grids    map[string]*image.Gray
}

var hdTownMapArchive = regexp.MustCompile(`^TOWN[0-7]$`)
var hdTownMapRegionID = regexp.MustCompile(`^town[0-7]-r[0-9]{2}-c[0-9]{2}$`)

func townMapRect(a [4]int) image.Rectangle {
	return image.Rect(a[0], a[1], a[2], a[3])
}

func townMapOrdered(a [4]int) bool { return a[0] < a[2] && a[1] < a[3] }

// LoadHDTownMaps returns nothing unless all declared cities are fully covered.
func LoadHDTownMaps(dir string, options ...HDLoadOption) (*HDTownMaps, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return nil, err
	}
	var m HDTownMapsManifest
	if err = json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	if m.Schema != "kol-hd-town-maps/1" || m.Rights == "" || !validHDHash(m.SourceZIP) || !validHDHash(m.GenerationSHA) || len(m.Maps) < 1 || len(m.Maps) > 8 || len(m.Regions) == 0 || len(m.Files) != len(m.Regions) {
		return nil, fmt.Errorf("HD城鎮地圖版本、完整組或來源不符")
	}
	coverage := map[string][]byte{}
	for _, source := range m.Maps {
		if !hdTownMapArchive.MatchString(source.Archive) || source.Count != 2 || source.Grid.Archive != source.Archive || source.Atlas.Archive != source.Archive || source.Grid.Index != 0 || source.Atlas.Index != 1 || !validHDHash(source.Grid.SourcePNG) || !validHDHash(source.Grid.SourceRGB) || !validHDHash(source.GridRGBA) || !validHDHash(source.Atlas.SourcePNG) || !validHDHash(source.Atlas.SourceRGB) || !validHDHash(source.AtlasRGBA) || coverage[source.Archive] != nil {
			return nil, fmt.Errorf("HD城鎮地圖來源圖號、數量或重複城鎮不符")
		}
		coverage[source.Archive] = make([]byte, 64*64)
	}
	seen := map[string]bool{}
	for _, region := range m.Regions {
		cells, ok := coverage[region.Archive]
		source, visible := townMapRect(region.SourceBounds), townMapRect(region.VisibleBounds)
		info, fileOK := m.Files[region.File]
		if !ok || !hdTownMapRegionID.MatchString(region.ID) || region.ID[:5] != fmt.Sprintf("town%c", region.Archive[4]) || seen[region.ID] || region.File != region.ID+".png" || !fileOK || !validHDHash(region.NativeRGBA) || !validHDHash(info.SHA256) || !townMapOrdered(region.SourceBounds) || !townMapOrdered(region.VisibleBounds) || source.Dx() != 12 || source.Dy() != 12 || !source.In(image.Rect(0, 0, 64, 64)) || !visible.In(source) || info.Width != info.Height || info.Width < 768 || info.Width > 4096 {
			return nil, fmt.Errorf("HD城鎮地圖位置、原生尺寸、檔名或來源範圍不符")
		}
		seen[region.ID] = true
		for y := visible.Min.Y; y < visible.Max.Y; y++ {
			for x := visible.Min.X; x < visible.Max.X; x++ {
				if cells[y*64+x] != 0 {
					return nil, fmt.Errorf("HD城鎮地圖區塊重疊")
				}
				cells[y*64+x] = 1
			}
		}
	}
	for _, cells := range coverage {
		for _, covered := range cells {
			if covered != 1 {
				return nil, fmt.Errorf("HD城鎮地圖缺少可見格子")
			}
		}
	}
	generation, err := os.ReadFile(filepath.Join(dir, "generation.json"))
	if err != nil {
		return nil, err
	}
	if monsterHash(generation) != m.GenerationSHA {
		return nil, fmt.Errorf("HD城鎮地圖生成清冊不符")
	}
	result := &HDTownMaps{Manifest: m}
	for _, region := range m.Regions {
		if deferHDPixels(options) {
			pic, err := openHDImage(dir, region.File, m.Files[region.File])
			if err != nil {
				return nil, err
			}
			result.Images = append(result.Images, pic)
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, region.File))
		if err != nil {
			return nil, err
		}
		info := m.Files[region.File]
		if monsterHash(data) != info.SHA256 {
			return nil, fmt.Errorf("HD城鎮地圖PNG雜湊不符")
		}
		cfg, err := png.DecodeConfig(bytes.NewReader(data))
		if err != nil || cfg.Width != info.Width || cfg.Height != info.Height {
			return nil, fmt.Errorf("HD城鎮地圖實際原生尺寸不符")
		}
		pic, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		rgba := image.NewNRGBA(image.Rect(0, 0, info.Width, info.Height))
		draw.Draw(rgba, rgba.Bounds(), pic, pic.Bounds().Min, draw.Src)
		if monsterHash(rgba.Pix) != region.NativeRGBA {
			return nil, fmt.Errorf("HD城鎮地圖原生RGBA不符")
		}
		for i := 3; i < len(rgba.Pix); i += 4 {
			if rgba.Pix[i] != 255 {
				return nil, fmt.Errorf("HD城鎮地圖含透明像素")
			}
		}
		result.Images = append(result.Images, pic)
	}
	return result, nil
}

// ValidateSources retains typed grids only after every declared source passed.
func (m *HDTownMaps) ValidateSources(lib *Library) error {
	if lib == nil || lib.Manifest().SourceZip.SHA256 != m.Manifest.SourceZIP {
		return fmt.Errorf("HD城鎮地圖來源母體不符")
	}
	next := map[string]*image.Gray{}
	for _, source := range m.Manifest.Maps {
		if source.Count != 2 || lib.Count(source.Archive) != source.Count {
			return fmt.Errorf("HD城鎮地圖實際來源數量不符")
		}
		for _, item := range []struct {
			p    HDPageSource
			rgba string
			size image.Point
		}{{source.Grid, source.GridRGBA, image.Pt(64, 64)}, {source.Atlas, source.AtlasRGBA, image.Pt(256, 256)}} {
			pic, meta, err := lib.Decode(item.p.Archive, item.p.Index)
			if err != nil {
				return err
			}
			if pic.Bounds().Size() != item.size || meta.SHA256 != item.p.SourcePNG {
				return fmt.Errorf("HD城鎮地圖來源尺寸或登記雜湊不符")
			}
			data, err := os.ReadFile(filepath.Join(lib.root, meta.File))
			if err != nil {
				return err
			}
			if monsterHash(data) != item.p.SourcePNG {
				return fmt.Errorf("HD城鎮地圖來源PNG不符")
			}
			if item.p.Index == 0 {
				gray, ok := pic.(*image.Gray)
				if !ok || len(data) < 26 || data[24] != 8 || data[25] != 0 {
					return fmt.Errorf("HD城鎮地圖不是八位灰階索引PNG")
				}
				copy := image.NewGray(image.Rect(0, 0, 64, 64))
				draw.Draw(copy, copy.Bounds(), gray, gray.Bounds().Min, draw.Src)
				next[source.Archive] = copy
			}
			rgba := image.NewNRGBA(image.Rectangle{Max: item.size})
			draw.Draw(rgba, rgba.Bounds(), pic, pic.Bounds().Min, draw.Src)
			rgb := make([]byte, 0, item.size.X*item.size.Y*3)
			for i := 0; i < len(rgba.Pix); i += 4 {
				if rgba.Pix[i+3] != 255 {
					return fmt.Errorf("HD城鎮地圖來源含透明像素")
				}
				rgb = append(rgb, rgba.Pix[i:i+3]...)
			}
			if monsterHash(rgb) != item.p.SourceRGB || monsterHash(rgba.Pix) != item.rgba {
				return fmt.Errorf("HD城鎮地圖來源RGB或RGBA不符")
			}
		}
	}
	m.Grids = next
	return nil
}
