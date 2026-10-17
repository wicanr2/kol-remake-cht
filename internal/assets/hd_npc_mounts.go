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
	"regexp"
)

type HDNPCMountSource struct {
	HDPageSource
	Size [2]int `json:"size"`
}

type HDNPCMountPortrait struct {
	HDNPCMountSource
	Kind string `json:"kind"`
	File string `json:"file"`
	Rect [4]int `json:"rect"`
	RGBA string `json:"native_rgba_sha256"`
}

type HDNPCPreserved struct {
	HDNPCMountSource
	File string `json:"hd_file"`
	PNG  string `json:"hd_png_sha256"`
}

type HDNPCMountsManifest struct {
	Schema        string                 `json:"schema"`
	Rights        string                 `json:"rights"`
	SourceZIP     string                 `json:"source_zip_sha256"`
	ArchiveCounts map[string]int         `json:"archive_counts"`
	Files         map[string]HDImageInfo `json:"files"`
	Preserved     []HDNPCPreserved       `json:"preserved"`
	Portraits     []HDNPCMountPortrait   `json:"portraits"`
	GenerationSHA string                 `json:"generation_sha256"`
}

type HDNPCMounts struct {
	Manifest  HDNPCMountsManifest
	Portraits map[string]image.Image
	Mounts    map[string]image.Image
}

func npcMountCounts() map[string]int {
	return map[string]int{"TWN1PICS": 33, "TWN2PICS": 29, "TWN3PICS": 25, "TWN4PICS": 25, "TWN5PICS": 29, "TWN6PICS": 25, "TWN7PICS": 45, "STEEDS": 4}
}

func npcMountHash(data []byte) string { return fmt.Sprintf("%x", sha256.Sum256(data)) }

// LoadHDNPCMounts 驗完整106個來源及102個新裁切；錯誤只回傳nil組。
func LoadHDNPCMounts(dir string, options ...HDLoadOption) (*HDNPCMounts, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return nil, err
	}
	var m HDNPCMountsManifest
	if err = json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	counts := npcMountCounts()
	if m.Schema != "kol-hd-npc-mounts/1" || m.Rights == "" || !validHDHash(m.SourceZIP) ||
		len(m.ArchiveCounts) != len(counts) || len(m.Files) != 27 || len(m.Preserved) != 4 || len(m.Portraits) != 102 || !validHDHash(m.GenerationSHA) {
		return nil, fmt.Errorf("HD NPC與坐騎版本、來源或完整組數量不符")
	}
	generation, err := os.ReadFile(filepath.Join(dir, "generation.json"))
	if err != nil {
		return nil, err
	}
	if npcMountHash(generation) != m.GenerationSHA {
		return nil, fmt.Errorf("HD NPC與坐騎生成清冊雜湊不符")
	}
	expected, seenRGB := map[string]bool{}, map[string]bool{}
	for archive, count := range counts {
		if m.ArchiveCounts[archive] != count {
			return nil, fmt.Errorf("HD NPC與坐騎來源%s數量不符", archive)
		}
		start, step := 2, 2
		if archive == "STEEDS" {
			start, step = 0, 1
		}
		for i := start; i < count; i += step {
			expected[fmt.Sprintf("%s:%d", archive, i)] = true
		}
	}
	validSource := func(s HDNPCMountSource) bool {
		key := fmt.Sprintf("%s:%d", s.Archive, s.Index)
		size := [2]int{88, 64}
		if s.Archive == "STEEDS" {
			size = [2]int{40, 38}
		}
		if !expected[key] || seenRGB[s.SourceRGB] || s.Size != size || !validHDHash(s.SourceRGB) || !validHDHash(s.SourcePNG) {
			return false
		}
		delete(expected, key)
		seenRGB[s.SourceRGB] = true
		return true
	}
	approved := []struct {
		archive    string
		index      int
		file, hash string
	}{
		{"TWN1PICS", 12, "pages/twn1pics-012-painted.png", "ed5966b0508182501c44ceb23ca3bbe990eefe2194dd8532f1032cfc57cd2d4f"},
		{"TWN1PICS", 18, "pages/twn1pics-018-painted.png", "095220b8023aeb05185ce907c9cb807b2d967f03b84981a6567241dd8818b92f"},
		{"TWN1PICS", 20, "pages/twn1pics-020-painted.png", "07c7b11725af9224de90c33980153a0fd0e9a81bdc01fe8a48e594a871cbcd84"},
		{"TWN7PICS", 12, "interface/innkeeper-painted-v1.png", "07ba04b2330327245e6f12518e096c42252f70394dc43fb9c95186e1ffcf399b"},
	}
	for i, p := range m.Preserved {
		a := approved[i]
		if p.Archive != a.archive || p.Index != a.index || p.File != a.file || p.PNG != a.hash || !validSource(p.HDNPCMountSource) {
			return nil, fmt.Errorf("HD保留NPC來源、順序或核准圖片不符")
		}
		data, err := os.ReadFile(filepath.Join(dir, "..", p.File))
		if err != nil {
			return nil, err
		}
		if npcMountHash(data) != p.PNG {
			return nil, fmt.Errorf("HD保留NPC圖片雜湊不符")
		}
	}
	namePattern := regexp.MustCompile(`^[a-z0-9][a-z0-9-]*\.png$`)
	images := map[string]image.Image{}
	for name, info := range m.Files {
		if !namePattern.MatchString(name) || !validHDHash(info.SHA256) || info.Width < 1 || info.Height < 1 || info.Width > 4096 || info.Height > 4096 {
			return nil, fmt.Errorf("HD NPC圖集檔名、尺寸或雜湊無效")
		}
		if deferHDPixels(options) {
			pic, err := openHDImage(dir, name, info)
			if err != nil {
				return nil, err
			}
			images[name] = pic
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		if npcMountHash(data) != info.SHA256 {
			return nil, fmt.Errorf("HD NPC圖集%s雜湊不符", name)
		}
		cfg, err := png.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		if cfg.Width != info.Width || cfg.Height != info.Height {
			return nil, fmt.Errorf("HD NPC圖集%s實際尺寸不符", name)
		}
		pic, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		images[name] = pic
	}
	result := &HDNPCMounts{Manifest: m, Portraits: map[string]image.Image{}, Mounts: map[string]image.Image{}}
	used := map[string]bool{}
	for _, p := range m.Portraits {
		mount := p.Archive == "STEEDS"
		if !validSource(p.HDNPCMountSource) || mount && p.Kind != "mount" || !mount && p.Kind != "npc" || !validHDHash(p.RGBA) {
			return nil, fmt.Errorf("HD NPC與坐騎來源、圖號、種類或裁切指紋不符")
		}
		pic := images[p.File]
		r := image.Rect(p.Rect[0], p.Rect[1], p.Rect[2], p.Rect[3])
		minSize := image.Pt(528, 384)
		if mount {
			minSize = image.Pt(240, 228)
		}
		if pic == nil || p.Rect[0] >= p.Rect[2] || p.Rect[1] >= p.Rect[3] || !r.In(pic.Bounds()) || r.Dx() < minSize.X || r.Dy() < minSize.Y {
			return nil, fmt.Errorf("HD NPC與坐騎裁切越界、缺件或原生尺寸不足")
		}
		if !deferHDPixels(options) {
			rgba := make([]byte, 0, r.Dx()*r.Dy()*4)
			for y := r.Min.Y; y < r.Max.Y; y++ {
				for x := r.Min.X; x < r.Max.X; x++ {
					c := color.NRGBAModel.Convert(pic.At(x, y)).(color.NRGBA)
					if c.A != 255 {
						return nil, fmt.Errorf("HD NPC與坐騎畫格必須不透明")
					}
					rgba = append(rgba, c.R, c.G, c.B, c.A)
				}
			}
			if npcMountHash(rgba) != p.RGBA {
				return nil, fmt.Errorf("HD NPC與坐騎原生裁切RGBA不符")
			}
		}
		crop := pic.(interface {
			SubImage(image.Rectangle) image.Image
		}).SubImage(r)
		if mount {
			result.Mounts[p.SourceRGB] = crop
		} else {
			result.Portraits[p.SourceRGB] = crop
		}
		used[p.File] = true
	}
	if len(expected) != 0 || len(result.Portraits) != 98 || len(result.Mounts) != 4 || len(used) != len(images) {
		return nil, fmt.Errorf("HD NPC與坐騎來源缺漏或圖集未引用")
	}
	return result, nil
}

// Sources 包含四個保留NPC，供CPU來源驗證與GPU首次繪圖核對。
func (n *HDNPCMounts) Sources() []HDNPCMountSource {
	sources := make([]HDNPCMountSource, 0, 106)
	for _, p := range n.Manifest.Preserved {
		sources = append(sources, p.HDNPCMountSource)
	}
	for _, p := range n.Manifest.Portraits {
		sources = append(sources, p.HDNPCMountSource)
	}
	return sources
}

// ValidateSources 在交換畫布前驗八組實際數量、PNG、原始RGB與尺寸。
func (n *HDNPCMounts) ValidateSources(lib *Library) error {
	if lib == nil || lib.Manifest().SourceZip.SHA256 != n.Manifest.SourceZIP {
		return fmt.Errorf("HD NPC與坐騎來源PNG庫母體不符")
	}
	for archive, count := range n.Manifest.ArchiveCounts {
		if lib.Count(archive) != count {
			return fmt.Errorf("HD NPC與坐騎來源%s實際數量不符", archive)
		}
	}
	for _, s := range n.Sources() {
		pic, meta, err := lib.Decode(s.Archive, s.Index)
		if err != nil {
			return err
		}
		if pic.Bounds().Size() != image.Pt(s.Size[0], s.Size[1]) || meta.SHA256 != s.SourcePNG {
			return fmt.Errorf("HD NPC與坐騎來源%s第%d圖尺寸或登記雜湊不符", s.Archive, s.Index)
		}
		data, err := os.ReadFile(filepath.Join(lib.root, meta.File))
		if err != nil {
			return err
		}
		if npcMountHash(data) != s.SourcePNG {
			return fmt.Errorf("HD NPC與坐騎來源PNG雜湊不符")
		}
		rgb := make([]byte, 0, s.Size[0]*s.Size[1]*3)
		b := pic.Bounds()
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				c := color.NRGBAModel.Convert(pic.At(x, y)).(color.NRGBA)
				rgb = append(rgb, c.R, c.G, c.B)
			}
		}
		if npcMountHash(rgb) != s.SourceRGB {
			return fmt.Errorf("HD NPC與坐騎來源%s第%d圖RGB不符", s.Archive, s.Index)
		}
	}
	return nil
}
