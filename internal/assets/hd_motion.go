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

type HDMotionPicture struct {
	HDPageSource
	ID           string `json:"id"`
	Transparent  bool   `json:"transparent"`
	File         string `json:"file"`
	Size         [2]int `json:"size"`
	RGBA         string `json:"source_rgba_sha256"`
	DeliveryRGBA string `json:"delivery_rgba_sha256"`
}
type HDMotionAtlas struct {
	File  string `json:"file"`
	PNG   string `json:"png_sha256"`
	RGB   string `json:"rgb_sha256"`
	RGBA  string `json:"rgba_sha256"`
	Size  [2]int `json:"size"`
	Cells int    `json:"cells"`
}
type HDMotionManifest struct {
	TitleCards    []HDMotionPicture        `json:"title_cards"`
	Stencils      []HDMotionPicture        `json:"weather_stencils"`
	BattleAtlases map[string]HDMotionAtlas `json:"battle_atlases"`
	Schema        string                   `json:"schema"`
	Rights        string                   `json:"rights"`
	SourceZIP     string                   `json:"source_zip_sha256"`
	ArchiveCounts map[string]int           `json:"archive_counts"`
	Pictures      []HDMotionPicture        `json:"pictures"`
	Files         map[string]HDImageInfo   `json:"files"`
	GenerationSHA string                   `json:"generation_sha256"`
}
type HDMotionArt struct {
	Manifest HDMotionManifest
	Images   []image.Image
}

func motionExpected(i int) (string, string, int, [2]int, [2]int, bool) {
	switch i {
	case 0:
		return "intro-00", "BACKPICS", 0, [2]int{320, 200}, [2]int{768, 480}, false
	case 1:
		return "intro-01", "BACKPICS", 1, [2]int{320, 200}, [2]int{768, 480}, false
	case 2:
		return "intro-02", "BACKPICS", 2, [2]int{320, 200}, [2]int{768, 480}, false
	case 3:
		return "intro-03", "BACKPICS", 3, [2]int{320, 200}, [2]int{768, 480}, false
	case 4:
		return "intro-04", "BACKPICS", 4, [2]int{320, 200}, [2]int{768, 480}, false
	case 5:
		return "intro-05", "BACKPICS", 5, [2]int{320, 200}, [2]int{768, 480}, false
	case 6:
		return "intro-06", "BACKPICS", 6, [2]int{320, 200}, [2]int{768, 480}, false
	case 7:
		return "book-cover", "BOOKPICS", 0, [2]int{320, 200}, [2]int{768, 480}, false
	case 8:
		return "body-3", "MISCPICS", 3, [2]int{56, 70}, [2]int{336, 420}, false
	case 9:
		return "body-4", "MISCPICS", 4, [2]int{56, 70}, [2]int{336, 420}, false
	case 10:
		return "body-5", "MISCPICS", 5, [2]int{56, 70}, [2]int{336, 420}, false
	case 11:
		return "weather-base-0", "WEATPICS", 0, [2]int{64, 62}, [2]int{384, 372}, false
	case 12:
		return "weather-cloud-0", "WEATPICS", 1, [2]int{64, 62}, [2]int{768, 186}, false
	case 13:
		return "weather-base-1", "WEATPICS", 3, [2]int{64, 62}, [2]int{384, 372}, false
	case 14:
		return "weather-cloud-1", "WEATPICS", 4, [2]int{64, 62}, [2]int{768, 186}, false
	case 15:
		return "weather-base-2", "WEATPICS", 6, [2]int{64, 62}, [2]int{384, 372}, false
	case 16:
		return "weather-cloud-2", "WEATPICS", 7, [2]int{64, 62}, [2]int{768, 186}, false
	case 17:
		return "weather-base-5", "WEATPICS", 15, [2]int{64, 62}, [2]int{384, 372}, false
	case 18:
		return "weather-cloud-5", "WEATPICS", 16, [2]int{64, 62}, [2]int{768, 186}, false
	case 19:
		return "weather-3-0", "WEATPICS", 9, [2]int{64, 62}, [2]int{384, 372}, false
	case 20:
		return "weather-3-1", "WEATPICS", 10, [2]int{64, 62}, [2]int{384, 372}, false
	case 21:
		return "weather-3-2", "WEATPICS", 11, [2]int{64, 62}, [2]int{384, 372}, false
	case 22:
		return "weather-4-0", "WEATPICS", 12, [2]int{64, 62}, [2]int{384, 372}, false
	case 23:
		return "weather-4-1", "WEATPICS", 13, [2]int{64, 62}, [2]int{384, 372}, false
	case 24:
		return "weather-4-2", "WEATPICS", 14, [2]int{64, 62}, [2]int{384, 372}, false
	case 25:
		return "walking-helmets", "KOLPICS-MARKERS", 0, [2]int{16, 160}, [2]int{640, 128}, true
	case 26:
		return "battle-terrain-10", "KOLPICS", 8, [2]int{64, 160}, [2]int{640, 256}, false
	case 27:
		return "battle-terrain-20", "KOLPICS", 8, [2]int{64, 160}, [2]int{640, 256}, false
	case 28:
		return "battle-terrain-30", "KOLPICS", 8, [2]int{64, 160}, [2]int{640, 256}, false
	case 29:
		return "battle-terrain-40", "KOLPICS", 8, [2]int{64, 160}, [2]int{640, 256}, false
	case 30:
		return "battle-terrain-50", "KOLPICS", 8, [2]int{64, 160}, [2]int{640, 256}, false
	case 31:
		return "battle-terrain-60", "KOLPICS", 8, [2]int{64, 160}, [2]int{640, 256}, false
	case 32:
		return "battle-terrain-70", "KOLPICS", 8, [2]int{64, 160}, [2]int{640, 256}, false
	case 33:
		return "battle-sprite-0", "KOLPICS", 8, [2]int{64, 160}, [2]int{128, 128}, true
	case 34:
		return "battle-sprite-6", "KOLPICS", 8, [2]int{64, 160}, [2]int{128, 128}, true
	case 35:
		return "battle-sprite-7", "KOLPICS", 8, [2]int{64, 160}, [2]int{128, 128}, true
	case 36:
		return "battle-sprite-8", "KOLPICS", 8, [2]int{64, 160}, [2]int{128, 128}, true
	case 37:
		return "battle-sprite-38", "KOLPICS", 8, [2]int{64, 160}, [2]int{128, 128}, true
	case 38:
		return "battle-alternate-0", "KOLPICS", 12, [2]int{64, 160}, [2]int{640, 256}, false
	case 39:
		return "battle-alternate-1", "KOLPICS", 12, [2]int{64, 160}, [2]int{512, 128}, false
	case 40:
		return "set0pics-0", "SET0PICS", 0, [2]int{88, 70}, [2]int{528, 420}, false
	case 41:
		return "set0pics-1", "SET0PICS", 1, [2]int{88, 70}, [2]int{528, 420}, false
	case 42:
		return "set0pics-2", "SET0PICS", 2, [2]int{88, 70}, [2]int{528, 420}, false
	case 43:
		return "set0pics-3", "SET0PICS", 3, [2]int{88, 70}, [2]int{528, 420}, false
	case 44:
		return "set0pics-4", "SET0PICS", 4, [2]int{88, 70}, [2]int{528, 420}, false
	case 45:
		return "set0pics-5", "SET0PICS", 5, [2]int{88, 70}, [2]int{528, 420}, false
	case 46:
		return "set1pics-0", "SET1PICS", 0, [2]int{104, 84}, [2]int{624, 504}, false
	case 47:
		return "set1pics-1", "SET1PICS", 1, [2]int{104, 84}, [2]int{624, 504}, false
	case 48:
		return "set1pics-2", "SET1PICS", 2, [2]int{104, 84}, [2]int{624, 504}, false
	case 49:
		return "set1pics-3", "SET1PICS", 3, [2]int{104, 84}, [2]int{624, 504}, false
	case 50:
		return "set1pics-4", "SET1PICS", 4, [2]int{104, 84}, [2]int{624, 504}, false
	case 51:
		return "set1pics-5", "SET1PICS", 5, [2]int{104, 84}, [2]int{624, 504}, false
	case 52:
		return "set2pics-0", "SET2PICS", 0, [2]int{160, 125}, [2]int{768, 600}, false
	case 53:
		return "set2pics-1", "SET2PICS", 1, [2]int{160, 125}, [2]int{768, 600}, false
	case 54:
		return "set2pics-2", "SET2PICS", 2, [2]int{160, 125}, [2]int{768, 600}, false
	case 55:
		return "set2pics-3", "SET2PICS", 3, [2]int{160, 125}, [2]int{768, 600}, false
	case 56:
		return "set2pics-4", "SET2PICS", 4, [2]int{160, 125}, [2]int{768, 600}, false
	case 57:
		return "set2pics-5", "SET2PICS", 5, [2]int{160, 125}, [2]int{768, 600}, false
	case 58:
		return "facepics-0", "FACEPICS", 0, [2]int{112, 80}, [2]int{672, 480}, false
	case 59:
		return "facepics-1", "FACEPICS", 1, [2]int{112, 80}, [2]int{672, 480}, false
	case 60:
		return "facepics-2", "FACEPICS", 2, [2]int{112, 80}, [2]int{672, 480}, false
	case 61:
		return "facepics-3", "FACEPICS", 3, [2]int{112, 80}, [2]int{672, 480}, false
	case 62:
		return "facepics-4", "FACEPICS", 4, [2]int{112, 80}, [2]int{672, 480}, false
	case 63:
		return "facepics-5", "FACEPICS", 5, [2]int{112, 80}, [2]int{672, 480}, false
	case 64:
		return "battle-ground-9", "KOLPICS", 8, [2]int{64, 160}, [2]int{128, 128}, false
	}
	panic("invalid motion asset index")
}

func LoadHDMotionArt(dir string) (*HDMotionArt, error) {
	raw, e := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if e != nil {
		return nil, e
	}
	var m HDMotionManifest
	if e = json.Unmarshal(raw, &m); e != nil {
		return nil, e
	}
	if m.Schema != "kol-hd-motion/1" || m.Rights == "" || !validHDHash(m.SourceZIP) || !validHDHash(m.GenerationSHA) || len(m.Pictures) != 65 || len(m.Files) != 65 {
		return nil, fmt.Errorf("HD動畫版本或完整組不符")
	}
	counts := map[string]int{"BACKPICS": 14, "BOOKPICS": 14, "MISCPICS": 5, "WEATPICS": 18, "KOLPICS": 8, "KOLPICS-MARKERS": 1, "SET0PICS": 6, "SET1PICS": 6, "SET2PICS": 6, "FACEPICS": 6}
	if len(m.ArchiveCounts) != len(counts) {
		return nil, fmt.Errorf("HD動畫來源組不符")
	}
	for a, n := range counts {
		if m.ArchiveCounts[a] != n {
			return nil, fmt.Errorf("HD動畫來源實際數量不符")
		}
	}

	if len(m.Stencils) != 4 || len(m.BattleAtlases) != 2 {
		return nil, fmt.Errorf("HD動畫資料來源不完整")
	}
	for i, p := range m.Stencils {
		indices := []int{2, 5, 8, 17}
		if p.Archive != "WEATPICS" || p.Index != indices[i] || p.Size != [2]int{64, 62} || !validHDHash(p.SourcePNG) || !validHDHash(p.SourceRGB) || !validHDHash(p.RGBA) {
			return nil, fmt.Errorf("HD雲層遮罩來源不符")
		}
	}
	for _, v := range []string{"8", "12"} {
		p := m.BattleAtlases[v]
		name := "BATTLE/battle-tiles-08.png"
		pin := "51d83c983ff621dae991111af5ce8ce4601cd6ed14c05a73da76700eb6b08d3b"
		if v == "12" {
			name = "BATTLE/battle-tiles-12.png"
			pin = "6617265036b93bac7b433e1040c4dae5bb28f92990686360f38ae296206365c8"
		}
		if p.File != name || p.PNG != pin || p.Cells != 80 || p.Size != [2]int{160, 128} || !validHDHash(p.RGB) || !validHDHash(p.RGBA) {
			return nil, fmt.Errorf("HD戰場來源版本或80格形狀不符")
		}
	}
	if len(m.TitleCards) != 7 {
		return nil, fmt.Errorf("HD字幕卡來源不完整")
	}
	for i, p := range m.TitleCards {
		if p.Archive != "BACKPICS" || p.Index != i+7 || p.Size != [2]int{320, 200} || !validHDHash(p.SourcePNG) || !validHDHash(p.SourceRGB) || !validHDHash(p.RGBA) {
			return nil, fmt.Errorf("HD字幕卡來源不符")
		}
	}
	seen := map[string]bool{}
	for i, p := range m.Pictures {
		id, archive, index, size, delivery, transparent := motionExpected(i)
		name := id + ".png"
		info, ok := m.Files[name]
		if p.ID != id || p.Transparent != transparent || p.Archive != archive || p.Index != index || p.File != name || p.Size != size || !validHDHash(p.SourcePNG) || !validHDHash(p.SourceRGB) || !validHDHash(p.RGBA) || !validHDHash(p.DeliveryRGBA) || seen[p.File] || !ok || !validHDHash(info.SHA256) || info.Width != delivery[0] || info.Height != delivery[1] {
			return nil, fmt.Errorf("HD動畫圖號、來源、尺寸或檔名不符")
		}
		seen[p.File] = true
	}
	generation, e := os.ReadFile(filepath.Join(dir, "generation.json"))
	if e != nil {
		return nil, e
	}
	if monsterHash(generation) != m.GenerationSHA {
		return nil, fmt.Errorf("HD動畫生成清冊不符")
	}
	result := &HDMotionArt{Manifest: m}
	for _, p := range m.Pictures {
		info := m.Files[p.File]
		data, e := os.ReadFile(filepath.Join(dir, p.File))
		if e != nil {
			return nil, e
		}
		if monsterHash(data) != info.SHA256 {
			return nil, fmt.Errorf("HD動畫PNG雜湊不符")
		}
		cfg, e := png.DecodeConfig(bytes.NewReader(data))
		if e != nil || cfg.Width != info.Width || cfg.Height != info.Height {
			return nil, fmt.Errorf("HD動畫實際尺寸不符")
		}
		pic, e := png.Decode(bytes.NewReader(data))
		if e != nil {
			return nil, e
		}
		rgba := image.NewNRGBA(image.Rect(0, 0, info.Width, info.Height))
		draw.Draw(rgba, rgba.Bounds(), pic, pic.Bounds().Min, draw.Src)
		if monsterHash(rgba.Pix) != p.DeliveryRGBA {
			return nil, fmt.Errorf("HD動畫RGBA不符")
		}
		transparentPixel, opaquePixel := false, false
		for i := 3; i < len(rgba.Pix); i += 4 {
			transparentPixel = transparentPixel || rgba.Pix[i] == 0
			opaquePixel = opaquePixel || rgba.Pix[i] == 255
			if !p.Transparent && rgba.Pix[i] != 255 {
				return nil, fmt.Errorf("HD動畫圖片須不透明")
			}
		}
		if p.Transparent && (!transparentPixel || !opaquePixel) {
			return nil, fmt.Errorf("HD動畫透明背景或前景不符")
		}

		result.Images = append(result.Images, pic)
	}
	return result, nil
}
func (m *HDMotionArt) ValidateSources(lib *Library) error {
	if lib == nil || lib.Manifest().SourceZip.SHA256 != m.Manifest.SourceZIP {
		return fmt.Errorf("HD動畫來源母體不符")
	}
	for a, n := range m.Manifest.ArchiveCounts {
		if lib.Count(a) != n {
			return fmt.Errorf("HD動畫實際來源數量不符")
		}
	}
	for _, p := range m.Manifest.BattleAtlases {
		data, e := os.ReadFile(filepath.Join(lib.root, p.File))
		if e != nil {
			return e
		}
		if monsterHash(data) != p.PNG {
			return fmt.Errorf("HD戰場來源PNG不符")
		}
		pic, e := png.Decode(bytes.NewReader(data))
		if e != nil {
			return e
		}
		if pic.Bounds().Size() != image.Pt(160, 128) {
			return fmt.Errorf("HD戰場來源實際尺寸不符")
		}
		rgba := image.NewNRGBA(image.Rect(0, 0, 160, 128))
		draw.Draw(rgba, rgba.Bounds(), pic, pic.Bounds().Min, draw.Src)
		var rgb []byte
		for i := 0; i < len(rgba.Pix); i += 4 {
			rgb = append(rgb, rgba.Pix[i:i+3]...)
		}
		if monsterHash(rgba.Pix) != p.RGBA || monsterHash(rgb) != p.RGB {
			return fmt.Errorf("HD戰場來源RGB或RGBA不符")
		}
	}
	sourcePictures := append([]HDMotionPicture{}, m.Manifest.Pictures...)
	sourcePictures = append(sourcePictures, m.Manifest.Stencils...)
	sourcePictures = append(sourcePictures, m.Manifest.TitleCards...)
	for _, p := range sourcePictures {
		pic, meta, e := lib.Decode(p.Archive, p.Index)
		if e != nil {
			return e
		}
		if pic.Bounds().Size() != image.Pt(p.Size[0], p.Size[1]) || meta.SHA256 != p.SourcePNG {
			return fmt.Errorf("HD動畫來源尺寸或登記雜湊不符")
		}
		data, e := os.ReadFile(filepath.Join(lib.root, meta.File))
		if e != nil {
			return e
		}
		if monsterHash(data) != p.SourcePNG {
			return fmt.Errorf("HD動畫來源PNG不符")
		}
		rgba := image.NewNRGBA(image.Rect(0, 0, p.Size[0], p.Size[1]))
		draw.Draw(rgba, rgba.Bounds(), pic, pic.Bounds().Min, draw.Src)
		rgb := make([]byte, 0, p.Size[0]*p.Size[1]*3)
		for i := 0; i < len(rgba.Pix); i += 4 {
			rgb = append(rgb, rgba.Pix[i:i+3]...)
		}
		if monsterHash(rgb) != p.SourceRGB || monsterHash(rgba.Pix) != p.RGBA {
			return fmt.Errorf("HD動畫來源RGB或RGBA不符")
		}
	}
	return nil
}
