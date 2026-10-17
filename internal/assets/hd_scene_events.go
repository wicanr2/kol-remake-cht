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
	"strings"
)

type HDSceneEventPicture struct {
	HDPageSource
	File         string `json:"file"`
	Size         [2]int `json:"size"`
	RGBA         string `json:"source_rgba_sha256"`
	DeliveryRGBA string `json:"delivery_rgba_sha256"`
}
type HDSceneEventsManifest struct {
	Schema        string                 `json:"schema"`
	Rights        string                 `json:"rights"`
	SourceZIP     string                 `json:"source_zip_sha256"`
	ArchiveCounts map[string]int         `json:"archive_counts"`
	Pictures      []HDSceneEventPicture  `json:"pictures"`
	Files         map[string]HDImageInfo `json:"files"`
	GenerationSHA string                 `json:"generation_sha256"`
}
type HDSceneEvents struct {
	Manifest HDSceneEventsManifest
	Images   []image.Image
}

func sceneEventExpected(i int) (string, int, [2]int, [2]int) {
	if i == 32 {
		return "MISCPICS", 1, [2]int{296, 32}, [2]int{768, 83}
	}
	if i < 25 {
		return "TWN0PICS", i, [2]int{192, 96}, [2]int{768, 384}
	}
	i -= 25
	if i < 5 {
		return "COMBPICS", i, [2]int{296, 144}, [2]int{768, 374}
	}
	return "COMBPICS", i, [2]int{88, 64}, [2]int{384, 280}
}
func LoadHDSceneEvents(dir string, options ...HDLoadOption) (*HDSceneEvents, error) {
	raw, e := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if e != nil {
		return nil, e
	}
	var m HDSceneEventsManifest
	if e = json.Unmarshal(raw, &m); e != nil {
		return nil, e
	}
	if m.Schema != "kol-hd-scene-events/1" || m.Rights == "" || !validHDHash(m.SourceZIP) || !validHDHash(m.GenerationSHA) || len(m.Pictures) != 33 || len(m.Files) != 33 || len(m.ArchiveCounts) != 3 || m.ArchiveCounts["TWN0PICS"] != 25 || m.ArchiveCounts["COMBPICS"] != 7 || m.ArchiveCounts["MISCPICS"] != 5 {
		return nil, fmt.Errorf("HD事件版本、完整組或來源數量不符")
	}
	seen := map[string]bool{}
	for i, p := range m.Pictures {
		archive, index, size, delivery := sceneEventExpected(i)
		name := fmt.Sprintf("%s-%03d.png", strings.ToLower(archive), index)
		info, ok := m.Files[name]
		if p.Archive != archive || p.Index != index || p.File != name || p.Size != size || !validHDHash(p.SourcePNG) || !validHDHash(p.SourceRGB) || !validHDHash(p.RGBA) || !validHDHash(p.DeliveryRGBA) || seen[p.SourceRGB] || !ok || !validHDHash(info.SHA256) || info.Width != delivery[0] || info.Height != delivery[1] {
			return nil, fmt.Errorf("HD事件圖號、來源、尺寸或檔名不符")
		}
		seen[p.SourceRGB] = true
	}
	generation, e := os.ReadFile(filepath.Join(dir, "generation.json"))
	if e != nil {
		return nil, e
	}
	if monsterHash(generation) != m.GenerationSHA {
		return nil, fmt.Errorf("HD事件生成清冊不符")
	}
	result := &HDSceneEvents{Manifest: m}
	for _, p := range m.Pictures {
		info := m.Files[p.File]
		if deferHDPixels(options) {
			pic, err := openHDImage(dir, p.File, m.Files[p.File])
			if err != nil {
				return nil, err
			}
			result.Images = append(result.Images, pic)
			continue
		}
		data, e := os.ReadFile(filepath.Join(dir, p.File))
		if e != nil {
			return nil, e
		}
		if monsterHash(data) != info.SHA256 {
			return nil, fmt.Errorf("HD事件PNG雜湊不符")
		}
		cfg, e := png.DecodeConfig(bytes.NewReader(data))
		if e != nil || cfg.Width != info.Width || cfg.Height != info.Height {
			return nil, fmt.Errorf("HD事件實際尺寸不符")
		}
		pic, e := png.Decode(bytes.NewReader(data))
		if e != nil {
			return nil, e
		}
		rgba := image.NewNRGBA(image.Rect(0, 0, info.Width, info.Height))
		draw.Draw(rgba, rgba.Bounds(), pic, pic.Bounds().Min, draw.Src)
		if monsterHash(rgba.Pix) != p.DeliveryRGBA {
			return nil, fmt.Errorf("HD事件RGBA不符")
		}
		for i := 3; i < len(rgba.Pix); i += 4 {
			if rgba.Pix[i] != 255 {
				return nil, fmt.Errorf("HD事件圖片須不透明")
			}
		}
		result.Images = append(result.Images, pic)
	}
	return result, nil
}
func (m *HDSceneEvents) ValidateSources(lib *Library) error {
	if lib == nil || lib.Manifest().SourceZip.SHA256 != m.Manifest.SourceZIP {
		return fmt.Errorf("HD事件來源母體不符")
	}
	for a, n := range m.Manifest.ArchiveCounts {
		if lib.Count(a) != n {
			return fmt.Errorf("HD事件實際來源數量不符")
		}
	}
	for _, p := range m.Manifest.Pictures {
		pic, meta, e := lib.Decode(p.Archive, p.Index)
		if e != nil {
			return e
		}
		if pic.Bounds().Size() != image.Pt(p.Size[0], p.Size[1]) || meta.SHA256 != p.SourcePNG {
			return fmt.Errorf("HD事件來源尺寸或登記雜湊不符")
		}
		data, e := os.ReadFile(filepath.Join(lib.root, meta.File))
		if e != nil {
			return e
		}
		if monsterHash(data) != p.SourcePNG {
			return fmt.Errorf("HD事件來源PNG不符")
		}
		rgba := image.NewNRGBA(image.Rect(0, 0, p.Size[0], p.Size[1]))
		draw.Draw(rgba, rgba.Bounds(), pic, pic.Bounds().Min, draw.Src)
		rgb := make([]byte, 0, p.Size[0]*p.Size[1]*3)
		for i := 0; i < len(rgba.Pix); i += 4 {
			if rgba.Pix[i+3] != 255 {
				return fmt.Errorf("HD事件來源須不透明")
			}
			rgb = append(rgb, rgba.Pix[i:i+3]...)
		}
		if monsterHash(rgb) != p.SourceRGB || monsterHash(rgba.Pix) != p.RGBA {
			return fmt.Errorf("HD事件來源RGB或RGBA不符")
		}
	}
	return nil
}
