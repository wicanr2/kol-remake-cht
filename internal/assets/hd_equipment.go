package assets

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/wicanr2/kol-remake-cht/internal/hdart"
)

type HDEquipment struct {
	Manifest hdart.Equipment
	Images   map[string]image.Image
}

func equipmentBounds(r [4]int) image.Rectangle { return image.Rect(r[0], r[1], r[2], r[3]) }
func equipmentHash(data []byte) string         { return fmt.Sprintf("%x", sha256.Sum256(data)) }

// LoadHDEquipment 驗完整圖集及校準；任何缺件或錯誤只回傳nil組。
func LoadHDEquipment(dir string) (*HDEquipment, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err = json.Unmarshal(raw, &fields); err != nil {
		return nil, err
	}
	for _, key := range []string{"source_hands", "base_origins"} {
		if _, ok := fields[key]; ok {
			return nil, fmt.Errorf("HD裝備不得包含原始錨點%s", key)
		}
	}
	var m hdart.Equipment
	if err = json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	if m.Schema != "kol-hd-equipment/1" || m.Rights == "" || !validHDHash(m.SourceZIP) || !validHDHash(m.LayoutSHA256) || len(m.Files) != 46 || len(m.Parts) != 169 || len(m.Props) != 32 || len(m.Sources) != 216 || len(m.Grips) != 50 || len(m.Meshes) != 9 || len(m.BaseWaists) != 5 || len(m.HeadBands) != 5 || len(m.HandPaint) != 12 {
		return nil, fmt.Errorf("HD裝備版本、來源或完整組數量不符")
	}
	counts := map[string]int{"PERPICS": 6, "HEADPICS": 54, "TORPICS": 54, "LEGSPICS": 54, "WEAPPICS": 28, "SHLDPICS": 4, "HANDPICS": 12}
	if len(m.ArchiveCounts) != len(counts) {
		return nil, fmt.Errorf("HD裝備來源archive數量不符")
	}
	expected := map[string]bool{}
	for archive, count := range counts {
		if m.ArchiveCounts[archive] != count {
			return nil, fmt.Errorf("HD裝備來源%s數量不符", archive)
		}
		for i := 0; i < count; i++ {
			expected[fmt.Sprintf("%s:%d", archive, i)] = true
		}
	}
	for _, i := range []int{4, 5, 13} {
		expected[fmt.Sprintf("KOLPICS:%d", i)] = true
	}
	expected["BOOKDATA:4"] = true
	for _, s := range m.Sources {
		key := fmt.Sprintf("%s:%d", s.Archive, s.Index)
		if !expected[key] || s.Size[0] < 1 || s.Size[1] < 1 || !validHDHash(s.PNG) || !validHDHash(s.RGB) || !validHDHash(s.RenderedRGB) {
			return nil, fmt.Errorf("HD裝備來源%s重複、缺漏或雜湊無效", key)
		}
		delete(expected, key)
	}
	if len(expected) != 0 {
		return nil, fmt.Errorf("HD裝備來源缺漏")
	}
	images := map[string]image.Image{}
	namePattern := regexp.MustCompile(`^[a-z0-9][a-z0-9-]*\.png$`)
	for name, info := range m.Files {
		if !namePattern.MatchString(name) || !validHDHash(info.SHA256) || info.Width < 1 || info.Height < 1 || info.Width > 4096 || info.Height > 4096 {
			return nil, fmt.Errorf("HD裝備圖集檔名、尺寸或雜湊無效：%s", name)
		}
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		if equipmentHash(data) != info.SHA256 {
			return nil, fmt.Errorf("HD裝備圖集雜湊不符：%s", name)
		}
		cfg, err := png.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		if cfg.Width != info.Width || cfg.Height != info.Height {
			return nil, fmt.Errorf("HD裝備圖集實際尺寸不符：%s", name)
		}
		pic, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		images[name] = pic
	}
	used := map[string]bool{}
	checkCrop := func(file string, rect [4]int, hash string) error {
		pic := images[file]
		if pic == nil {
			return fmt.Errorf("HD裝備缺少圖集%s", file)
		}
		r := equipmentBounds(rect)
		if r.Empty() || rect[0] >= rect[2] || rect[1] >= rect[3] || !r.In(pic.Bounds()) || !validHDHash(hash) {
			return fmt.Errorf("HD裝備裁取越界或指紋無效：%s", file)
		}
		pixels := make([]byte, 0, r.Dx()*r.Dy()*4)
		visible := false
		for y := r.Min.Y; y < r.Max.Y; y++ {
			for x := r.Min.X; x < r.Max.X; x++ {
				c := color.NRGBAModel.Convert(pic.At(x, y)).(color.NRGBA)
				pixels = append(pixels, c.R, c.G, c.B, c.A)
				visible = visible || c.A >= 8
			}
		}
		if !visible || equipmentHash(pixels) != hash {
			return fmt.Errorf("HD裝備裁取空白或RGBA指紋不符：%s", file)
		}
		used[file] = true
		return nil
	}
	parts := map[string]bool{"WINGS:3": true}
	for _, archive := range []string{"PERPICS", "HEADPICS", "TORPICS", "LEGSPICS"} {
		count := 54
		if archive == "PERPICS" {
			count = 6
		}
		for i := 0; i < count; i++ {
			parts[fmt.Sprintf("%s:%d", archive, i)] = true
		}
	}
	for key, p := range m.Parts {
		if !parts[key] {
			return nil, fmt.Errorf("HD裝備未知或重複部件%s", key)
		}
		delete(parts, key)
		if err := checkCrop(p.File, p.Rect, p.RGBA); err != nil {
			return nil, err
		}
		for _, v := range p.World {
			if math.IsNaN(v) || math.IsInf(v, 0) || math.Abs(v) > 640 {
				return nil, fmt.Errorf("HD裝備呈現座標無效：%s", key)
			}
		}
		if p.World[0] >= p.World[2] || p.World[1] >= p.World[3] || p.Rect[2]-p.Rect[0] < int(math.Ceil((p.World[2]-p.World[0])*6)) || p.Rect[3]-p.Rect[1] < int(math.Ceil((p.World[3]-p.World[1])*6)) {
			return nil, fmt.Errorf("HD裝備原生尺寸不足或呈現矩形無效：%s", key)
		}
		armor := key == "LEGSPICS:45" || key == "LEGSPICS:51" || key == "LEGSPICS:53"
		if (p.Kind != "") != armor {
			return nil, fmt.Errorf("HD獨立護具分類不符：%s", key)
		}
	}
	if len(parts) != 0 {
		return nil, fmt.Errorf("HD裝備部件缺漏")
	}
	props := map[string]bool{}
	for archive, count := range map[string]int{"WEAPPICS": 28, "SHLDPICS": 4} {
		for i := 0; i < count; i++ {
			props[fmt.Sprintf("%s:%d", archive, i)] = true
		}
	}
	for key, p := range m.Props {
		if !props[key] {
			return nil, fmt.Errorf("HD武器盾牌圖號不符：%s", key)
		}
		delete(props, key)
		if err := checkCrop(p.File, p.Rect, p.RGBA); err != nil {
			return nil, err
		}
		b := equipmentBounds(p.Paint)
		if b.Empty() || p.Paint[0] >= p.Paint[2] || p.Paint[1] >= p.Paint[3] || p.SourceSize[0] < 1 || p.SourceSize[1] < 1 || !b.In(image.Rect(0, 0, p.SourceSize[0], p.SourceSize[1])) {
			return nil, fmt.Errorf("HD武器盾牌來源著色範圍無效：%s", key)
		}
	}
	for _, name := range []string{"town-back-painted.png", "outdoor-back-painted.png"} {
		pic := images[name]
		if pic == nil {
			return nil, fmt.Errorf("HD裝備背板缺漏")
		}
		opaque, ok := pic.(interface{ Opaque() bool })
		if !ok || !opaque.Opaque() {
			return nil, fmt.Errorf("HD裝備背板須不透明")
		}
		used[name] = true
	}
	if len(used) != len(images) {
		return nil, fmt.Errorf("HD裝備包含未引用圖片")
	}
	for i := 0; i < 12; i++ {
		b, ok := m.HandPaint[strconv.Itoa(i)]
		if !ok || b[0] < 0 || b[1] < 0 || b[0] >= b[2] || b[1] >= b[3] {
			return nil, fmt.Errorf("HD手部著色框缺漏或無效")
		}
	}
	for build := 1; build < 6; build++ {
		key := strconv.Itoa(build)
		waist, ok := m.BaseWaists[key]
		head, hok := m.HeadBands[key]
		if !ok || !hok || math.IsNaN(waist) || math.IsInf(waist, 0) || waist <= 0 || waist >= 200 || head[0] <= 0 || head[0] >= 200 || head[1] < 0 || head[1] >= head[2] || head[2] > 320 {
			return nil, fmt.Errorf("HD身體腰線或頭框無效")
		}
		for style := -1; style < 9; style++ {
			torso := -1
			if style >= 0 {
				torso = build*9 + style
			}
			grips, ok := m.Grips[fmt.Sprintf("%d:%d", build, torso)]
			if !ok || len(grips) != 2 {
				return nil, fmt.Errorf("HD雙手握點缺漏")
			}
			for side, g := range grips {
				b := equipmentBounds(g.Rect)
				if g.HandIndex != build*2+side || b.Empty() || g.Rect[0] >= g.Rect[2] || g.Rect[1] >= g.Rect[3] || !b.In(image.Rect(0, 0, 1920, 1200)) {
					return nil, fmt.Errorf("HD雙手握點越界或圖號不符")
				}
			}
		}
	}
	for i := 45; i < 54; i++ {
		key := fmt.Sprintf("TORPICS:%d", i)
		mesh, ok := m.Meshes[key]
		p := m.Parts[key]
		if !ok || mesh.NativeWaist <= 0 || mesh.NativeWaist >= p.Rect[3]-p.Rect[1] || math.IsNaN(mesh.WaistWorld) || math.IsInf(mesh.WaistWorld, 0) || mesh.WaistWorld <= p.World[1] || mesh.WaistWorld >= p.World[3] {
			return nil, fmt.Errorf("HD上身網格腰線無效：%s", key)
		}
	}
	return &HDEquipment{Manifest: m, Images: images}, nil
}

// ValidateSources 核對實際PNG位元組及著色框，成功前不交換HD快取。
func (e *HDEquipment) ValidateSources(lib *Library) error {
	if lib == nil || lib.Manifest().SourceZip.SHA256 != e.Manifest.SourceZIP {
		return fmt.Errorf("HD裝備來源PNG庫母體不符")
	}
	for archive, count := range e.Manifest.ArchiveCounts {
		if lib.Count(archive) != count {
			return fmt.Errorf("HD裝備來源%s實際數量不符", archive)
		}
	}
	for _, s := range e.Manifest.Sources {
		var pic image.Image
		var data []byte
		var err error
		if s.Archive == "BOOKDATA" {
			pic, err = lib.DecodeDefaultShield()
			if err == nil {
				data, err = os.ReadFile(filepath.Join(filepath.Dir(lib.root), "bookdata/default-shield.png"))
			}
		} else {
			var meta Image
			pic, meta, err = lib.Decode(s.Archive, s.Index)
			if err == nil {
				data, err = os.ReadFile(filepath.Join(lib.root, meta.File))
			}
		}
		if err != nil {
			return err
		}
		if pic.Bounds().Dx() != s.Size[0] || pic.Bounds().Dy() != s.Size[1] || equipmentHash(data) != s.PNG {
			return fmt.Errorf("HD裝備來源%s第%d圖PNG或尺寸不符", s.Archive, s.Index)
		}
		rgb := make([]byte, 0, s.Size[0]*s.Size[1]*3)
		rendered := make([]byte, 0, s.Size[0]*s.Size[1]*3)
		paint := image.Rectangle{}
		b := pic.Bounds()
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				c := color.NRGBAModel.Convert(pic.At(x, y)).(color.NRGBA)
				rgb = append(rgb, c.R, c.G, c.B)
				premultiplied := color.RGBAModel.Convert(pic.At(x, y)).(color.RGBA)
				rendered = append(rendered, premultiplied.R, premultiplied.G, premultiplied.B)
				if c.A > 0 && (c.R > 0 || c.G > 0 || c.B > 0) {
					paint = paint.Union(image.Rect(x-b.Min.X, y-b.Min.Y, x-b.Min.X+1, y-b.Min.Y+1))
				}
			}
		}
		if equipmentHash(rgb) != s.RGB {
			return fmt.Errorf("HD裝備來源%s第%d圖RGB不符", s.Archive, s.Index)
		}
		if equipmentHash(rendered) != s.RenderedRGB {
			return fmt.Errorf("HD裝備來源%s第%d圖呈現RGB不符", s.Archive, s.Index)
		}
		actual := [4]int{paint.Min.X, paint.Min.Y, paint.Max.X, paint.Max.Y}
		if s.Archive == "HANDPICS" && e.Manifest.HandPaint[strconv.Itoa(s.Index)] != actual {
			return fmt.Errorf("HD手部著色框與來源不符")
		}
		if s.Archive == "WEAPPICS" || s.Archive == "SHLDPICS" {
			p := e.Manifest.Props[fmt.Sprintf("%s:%d", s.Archive, s.Index)]
			if p.Paint != actual || p.SourceSize != s.Size {
				return fmt.Errorf("HD武器盾牌著色框與來源不符")
			}
		}
	}
	return nil
}
