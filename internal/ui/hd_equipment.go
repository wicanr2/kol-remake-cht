package ui

// 完整裝備頁只改HD呈現。素材先在資產層驗完整組，typed資料與保存保持。
import (
	"fmt"
	"github.com/wicanr2/kol-remake-cht/internal/hdart"
	"image"
	"math"
	"sort"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/kol-remake-cht/internal/portrait"
)

type hdEquipment struct {
	usedFrame        uint64
	SourceSizes      map[string][2]int
	Parts            map[string]hdart.Part
	Props            map[string]hdart.Prop
	HandPaint        map[string][]int
	defaultShieldRGB string
	Images           map[string]*ebiten.Image
	Sources          map[string]*ebiten.Image
	Base, Figure     *ebiten.Image
	Grips            map[string][]hdart.Grip
	Meshes           map[string]hdart.Mesh
	SourceHands      map[string][]int
	BaseOrigins      map[string]int
	BaseWaists       map[string]float64
	HeadBands        map[string][]int
	Warped           map[string]*ebiten.Image
	BodyShapes       map[string]*ebiten.Image
	BodyBounds       map[int]image.Rectangle
}

func (h *HD) AttachEquipment(images map[string]image.Image, m hdart.Equipment, binding hdart.Binding) error {
	if m.Schema != "kol-hd-equipment/1" || len(m.Files) != 46 || h.pages == nil || len(images) != 46 || len(m.Parts) != 169 || len(m.Props) != 32 || len(m.Sources) != 216 || len(binding.DefaultShieldRGB) != 64 {
		return fmt.Errorf("HD裝備需要完整素材與來源")
	}
	for _, x := range binding.BaseOrigins {
		if x <= 0 || x >= 40 {
			return fmt.Errorf("HD裝備來源身體原點無效")
		}
	}
	for _, a := range binding.SourceHands {
		if a[0] < 0 || a[1] < 0 || a[0] >= a[2] || a[1] >= a[3] || a[2] > 320 || a[3] > 200 {
			return fmt.Errorf("HD裝備來源雙手範圍無效")
		}
	}
	defaultFound := false
	for _, s := range m.Sources {
		if s.Archive == "BOOKDATA" && s.Index == 4 && s.RGB == binding.DefaultShieldRGB {
			defaultFound = true
		}
	}
	if !defaultFound {
		return fmt.Errorf("HD預設盾面來源不符")
	}
	for name, info := range m.Files {
		pic := images[name]
		if pic == nil || pic.Bounds() != image.Rect(0, 0, info.Width, info.Height) {
			return fmt.Errorf("HD裝備圖集缺漏或尺寸不符")
		}
	}
	for _, part := range m.Parts {
		pic := images[part.File]
		if pic == nil || !image.Rect(part.Rect[0], part.Rect[1], part.Rect[2], part.Rect[3]).In(pic.Bounds()) {
			return fmt.Errorf("HD裝備部件裁取越界")
		}
	}
	for _, part := range m.Props {
		pic := images[part.File]
		if pic == nil || !image.Rect(part.Rect[0], part.Rect[1], part.Rect[2], part.Rect[3]).In(pic.Bounds()) {
			return fmt.Errorf("HD裝備武器盾牌裁取越界")
		}
	}
	p := &hdEquipment{SourceSizes: map[string][2]int{}, Parts: m.Parts, Props: m.Props, Grips: m.Grips, Meshes: m.Meshes, BaseWaists: m.BaseWaists, HeadBands: map[string][]int{}, HandPaint: map[string][]int{}, BaseOrigins: map[string]int{}, SourceHands: map[string][]int{}, Images: map[string]*ebiten.Image{}, Sources: map[string]*ebiten.Image{}, Warped: map[string]*ebiten.Image{}, BodyShapes: map[string]*ebiten.Image{}, BodyBounds: map[int]image.Rectangle{}, defaultShieldRGB: binding.DefaultShieldRGB}
	for k, v := range m.HeadBands {
		p.HeadBands[k] = append([]int{}, v[:]...)
	}
	for k, v := range m.HandPaint {
		p.HandPaint[k] = append([]int{}, v[:]...)
	}
	for i, v := range binding.BaseOrigins {
		p.BaseOrigins[strconv.Itoa(i)] = v
	}
	for i, v := range binding.SourceHands {
		p.SourceHands[strconv.Itoa(i)] = append([]int{}, v[:]...)
	}
	names := []string{}
	for name := range images {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		p.Images[name] = h.newPaint(images[name])
	}
	for key, part := range p.Parts {
		p.Sources[key] = h.paintSubImage(p.Images[part.File], image.Rect(part.Rect[0], part.Rect[1], part.Rect[2], part.Rect[3]))
	}
	for key, part := range p.Props {
		p.Sources[key] = h.paintSubImage(p.Images[part.File], image.Rect(part.Rect[0], part.Rect[1], part.Rect[2], part.Rect[3]))
	}
	p.Base, p.Figure = ebiten.NewImage(1920, 1200), ebiten.NewImage(1920, 1200)
	for _, s := range m.Sources {
		p.SourceSizes[fmt.Sprintf("%s:%d", s.Archive, s.Index)] = s.Size
		h.pages.sources = append(h.pages.sources, HDPageSource{Archive: s.Archive, Index: s.Index, RGB: s.RenderedRGB})
	}
	h.equipment = p
	h.pages.validated = false
	return nil
}
func equipmentFor(n *hdNativeRender) (*hdEquipment, error) {
	if n.h.equipment == nil {
		return nil, fmt.Errorf("HD裝備素材尚未完整載入")
	}
	n.h.equipment.usedFrame = n.h.imageCache().frame
	return n.h.equipment, nil
}
func (p *hdEquipment) part(n *hdNativeRender, dst *ebiten.Image, key string) {
	a := p.Parts[key]
	src := p.Sources[key]
	n.h.drawHD(dst, src, equipmentRect(a.World))
}
func drawEquipmentBackdropHD(dst *ebiten.Image, wild bool) (bool, error) {
	n := activeHDUI
	if n == nil || dst != n.dst {
		return false, nil
	}
	p, err := equipmentFor(n)
	if err != nil {
		return true, err
	}
	name := "town-back-painted.png"
	if wild {
		name = "outdoor-back-painted.png"
	}
	n.h.drawHD(dst, p.Images[name], image.Rect(96, 18, 1824, 948))
	return true, nil
}
func drawEquipmentHD(dst *ebiten.Image, sprites Sprites, pieces []portrait.Piece, scale int) (bool, error) {
	n := activeHDUI
	if n == nil || dst != n.dst {
		return false, nil
	}
	p, err := equipmentFor(n)
	if err != nil {
		return true, err
	}
	for first := 0; first < len(pieces); {
		last := first + 1
		for last < len(pieces) && pieces[last].Archive != "PERPICS" {
			last++
		}
		group := pieces[first:last]
		if group[0].Archive == "PERPICS" && group[0].Index > 0 && group[0].Index < 6 {
			if err := p.figure(n, dst, sprites, group); err != nil {
				return true, err
			}
			first = last
			continue
		}
		if group[0].Archive != "PERPICS" || group[0].Index != 0 {
			return true, fmt.Errorf("HD裝備體型或起始圖層不符")
		}
		indices := map[string]int{"HEADPICS": -1, "TORPICS": -1, "LEGSPICS": -1}
		for _, q := range group {
			if _, ok := indices[q.Archive]; ok {
				if q.Index < 0 || q.Index >= 9 {
					return true, fmt.Errorf("HD體型0裝備圖號越界")
				}
				indices[q.Archive] = q.Index
			}
		}
		head, torso, legs := indices["HEADPICS"], indices["TORPICS"], indices["LEGSPICS"]
		p.Base.Clear()
		p.part(n, p.Base, "PERPICS:0")
		p.Figure.Clear()
		legTop := 924
		if legs >= 0 {
			legTop = int(math.Round(p.Parts[fmt.Sprintf("LEGSPICS:%d", legs)].World[1] * 6))
		}
		for y := 0; y < 1200; {
			visible := func(row int) bool {
				upper := 402
				if head >= 0 {
					upper = 390
				}
				return head < 0 && row < 402 || torso < 0 && row >= upper && row < legTop || legs < 0 && row >= 564
			}
			if !visible(y) {
				y++
				continue
			}
			end := y + 1
			for end < 1200 && visible(end) {
				end++
			}
			clip := image.Rect(0, y, 1920, end)
			p.Figure.SubImage(clip).(*ebiten.Image).DrawImage(p.Base, nil)
			y = end
		}
		if torso < 0 && legs >= 0 {
			// Removing the central hip/leg area must preserve exposed forearms.
			// Cache their actual separated alpha runs from the native body.
			clip := image.Rect(0, legTop, 1920, 1200)
			p.Figure.SubImage(clip).(*ebiten.Image).DrawImage(p.exposedArms(), nil)
		}
		for _, archive := range []string{"LEGSPICS", "TORPICS", "HEADPICS"} {
			if index := indices[archive]; index >= 0 {
				p.part(n, p.Figure, fmt.Sprintf("%s:%d", archive, index))
			}
		}
		dx := (group[0].X - p.BaseOrigins["0"]) * 48
		var op ebiten.DrawImageOptions
		op.GeoM.Translate(float64(dx), 0)
		dst.DrawImage(p.Figure, &op)
		for _, q := range group[1:] {
			if _, ok := indices[q.Archive]; ok {
				continue
			}
			if q.Archive == "SHLDPICS" || q.Archive == "WEAPPICS" {
				key := fmt.Sprintf("%s:%d", q.Archive, q.Index)
				a, ok := p.Props[key]
				if !ok || q.W*8 != a.SourceSize[0] || q.H != a.SourceSize[1] {
					return true, fmt.Errorf("HD武器盾牌來源尺寸不符：%s", key)
				}
				n.h.drawHD(dst, p.Sources[key], image.Rect((q.X*8+a.Paint[0])*6, (q.Y+a.Paint[1])*6, (q.X*8+a.Paint[2])*6, (q.Y+a.Paint[3])*6))
				continue
			}
			if q.Archive == "HANDPICS" {
				x, y := (q.X-(group[0].X-p.BaseOrigins["0"]))*48, q.Y*6
				b := p.HandPaint[fmt.Sprint(q.Index)]
				if len(b) != 4 || b[0] < 0 || b[1] < 0 || b[2] > q.W*8 || b[3] > q.H || b[2] <= b[0] || b[3] <= b[1] {
					return true, fmt.Errorf("HD手部著色範圍不符")
				}
				r := image.Rect(x+b[0]*6, y+b[1]*6, x+b[2]*6, y+b[3]*6).Intersect(p.Figure.Bounds())
				if r.Empty() {
					return true, fmt.Errorf("HD手部越界")
				}
				var hand ebiten.DrawImageOptions
				hand.GeoM.Translate(float64(r.Min.X+dx), float64(r.Min.Y))
				dst.DrawImage(p.Figure.SubImage(r).(*ebiten.Image), &hand)
				continue
			}
			if q.Archive == portrait.ArchiveDefaultShield {
				if err := p.shieldFace(n, dst, sprites, group, 0, torso, dx, q); err != nil {
					return true, err
				}
				continue
			}
			return true, fmt.Errorf("HD裝備不支援圖層%s", q.Archive)
		}
		first = last
	}
	return true, nil
}

// 網格只校準呈現腰線；外側手臂與拳頭保持等比，原typed資料不變。
func equipmentRect(a [4]float64) image.Rectangle {
	return image.Rect(int(math.Round(a[0]*6)), int(math.Round(a[1]*6)), int(math.Round(a[2]*6)), int(math.Round(a[3]*6)))
}
func (p *hdEquipment) torso(n *hdNativeRender, key string) *ebiten.Image {
	if pic := p.Warped[key]; pic != nil {
		return pic
	}
	part := p.Parts[key]
	src := n.h.paint(p.Sources[key])
	b := src.Bounds()
	// Cache both input and output outside the automatic texture atlas. Their
	// normalized origins keep the same mesh independent of atlas placement.
	pixels := make([]byte, b.Dx()*b.Dy()*4)
	src.ReadPixels(pixels)
	stable := ebiten.NewImageWithOptions(image.Rect(0, 0, b.Dx(), b.Dy()), &ebiten.NewImageOptions{Unmanaged: true})
	stable.WritePixels(pixels)
	p.Warped[key+":native"] = stable
	src, b = stable, stable.Bounds()
	r := equipmentRect(part.World)
	m := p.Meshes[key]
	sx := float64(r.Dx()) / float64(b.Dx())
	height := int(math.Ceil(math.Max(float64(r.Dy()), float64(b.Dy())*sx)))
	pic := ebiten.NewImageWithOptions(image.Rect(0, 0, r.Dx(), height), &ebiten.NewImageOptions{Unmanaged: true})
	xs := []float64{0, .25, .4, .6, .75, 1}
	ys := []float64{0, float64(b.Dy()) * .2, float64(b.Dy()) * .4, float64(m.NativeWaist), float64(b.Dy()) * .8, float64(b.Dy())}
	sort.Float64s(ys)
	verts := make([]ebiten.Vertex, 0, len(xs)*len(ys))
	for _, y := range ys {
		for _, x := range xs {
			weight := 1.0
			if x <= .25 || x >= .75 {
				weight = 0
			} else if x < .4 {
				weight = (x - .25) / .15
			} else if x > .6 {
				weight = (.75 - x) / .15
			}
			center := 0.0
			waist := m.WaistWorld*6 - float64(r.Min.Y)
			if y <= float64(m.NativeWaist) {
				center = y / float64(m.NativeWaist) * waist
			} else {
				center = waist + (y-float64(m.NativeWaist))/float64(b.Dy()-m.NativeWaist)*(float64(r.Dy())-waist)
			}
			dy := y*sx*(1-weight) + center*weight
			verts = append(verts, ebiten.Vertex{DstX: float32(x * float64(r.Dx())), DstY: float32(dy), SrcX: float32(float64(b.Min.X) + x*float64(b.Dx())), SrcY: float32(float64(b.Min.Y) + y), ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1})
		}
	}
	indices := []uint16{}
	for y := 0; y < len(ys)-1; y++ {
		for x := 0; x < len(xs)-1; x++ {
			a := uint16(y*len(xs) + x)
			b := a + 1
			c := a + uint16(len(xs))
			d := c + 1
			indices = append(indices, a, b, c, b, d, c)
		}
	}
	pic.DrawTriangles(verts, indices, src, &ebiten.DrawTrianglesOptions{Filter: ebiten.FilterLinear, Address: ebiten.AddressClampToZero})
	p.Warped[key] = pic
	return pic
}
func (p *hdEquipment) body(n *hdNativeRender, build, mask int) (*ebiten.Image, image.Rectangle) {
	key := fmt.Sprintf("%d:%d", build, mask)
	if pic := p.BodyShapes[key]; pic != nil {
		return pic, p.BodyBounds[build]
	}
	// This runs in Draw, after the graphics context is ready.
	p.Base.Clear()
	p.part(n, p.Base, fmt.Sprintf("PERPICS:%d", build))
	pixels := make([]byte, 1920*1200*4)
	p.Base.ReadPixels(pixels)
	r := equipmentRect(p.Parts[fmt.Sprintf("PERPICS:%d", build)].World)
	waist := int(math.Round(p.BaseWaists[fmt.Sprint(build)] * 6))
	head := p.HeadBands[fmt.Sprint(build)]
	arms := make(map[int][]image.Rectangle)
	limit := 0
	for side := 0; side < 2; side++ {
		if v := p.SourceHands[fmt.Sprint(build*2+side)][3]*6 + 60; v > limit {
			limit = v
		}
	}
	for y := waist; y < min(limit, r.Max.Y); y++ {
		runs := []image.Rectangle{}
		start := -1
		for x := r.Min.X; x < r.Max.X; x++ {
			on := pixels[(y*1920+x)*4+3] >= 8
			if on && start < 0 {
				start = x
			}
			if !on && start >= 0 {
				if x-start >= 3 {
					runs = append(runs, image.Rect(start, y, x, y+1))
				}
				start = -1
			}
		}
		if start >= 0 && r.Max.X-start >= 3 {
			runs = append(runs, image.Rect(start, y, r.Max.X, y+1))
		}
		if len(runs) >= 3 {
			arms[y] = []image.Rectangle{runs[0], runs[len(runs)-1]}
		}
	}
	rgba := make([]byte, r.Dx()*r.Dy()*4)
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			isHead := y < head[0]*6 && x >= head[1]*6 && x < head[2]*6
			keep := false
			if y < waist {
				keep = isHead && mask&1 == 0 || !isHead && mask&2 == 0
			} else {
				isArm := false
				for _, a := range arms[y] {
					if x >= a.Min.X && x < a.Max.X {
						isArm = true
						break
					}
				}
				keep = isArm && mask&2 == 0 || !isArm && mask&4 == 0
			}
			if keep {
				dst := ((y-r.Min.Y)*r.Dx() + x - r.Min.X) * 4
				src := (y*1920 + x) * 4
				copy(rgba[dst:dst+4], pixels[src:src+4])
			}
		}
	}
	pic := ebiten.NewImage(r.Dx(), r.Dy())
	pic.WritePixels(rgba)
	p.BodyShapes[key] = pic
	p.BodyBounds[build] = r
	return pic, r
}
func gripRect(g hdart.Grip) image.Rectangle {
	return image.Rect(g.Rect[0], g.Rect[1], g.Rect[2], g.Rect[3])
}
func rectCenter(r image.Rectangle) (float64, float64) {
	return float64(r.Min.X+r.Max.X) / 2, float64(r.Min.Y+r.Max.Y) / 2
}
func (p *hdEquipment) propRect(group []portrait.Piece, body, torso, dx int, q portrait.Piece, a hdart.Prop) image.Rectangle {
	r := image.Rect((q.X*8+a.Paint[0])*6, (q.Y+a.Paint[1])*6, (q.X*8+a.Paint[2])*6, (q.Y+a.Paint[3])*6)
	// RE-128: belt is a separate fixed page layer, independent of hand grips.
	if q.Archive == "WEAPPICS" && q.X == portrait.BeltCol && q.Y+q.H == portrait.BeltBaseline {
		return r
	}
	hands := []int{}
	for _, s := range group {
		if s.Archive == "HANDPICS" {
			hands = append(hands, s.Index)
		}
	}
	side := 0
	if q.Archive == "SHLDPICS" {
		side = 1
	} else if len(hands) == 0 {
		return r
	} else {
		side = hands[0] - body*2
	}
	source := func(side int) (float64, float64) {
		b := p.SourceHands[fmt.Sprint(body*2+side)]
		return float64(b[0]+b[2])*3 + float64(dx), float64(b[1]+b[3]) * 3
	}
	target := func(side int) (float64, float64) {
		x, y := rectCenter(gripRect(p.Grips[fmt.Sprintf("%d:%d", body, torso)][side]))
		return x + float64(dx), y
	}
	sx, sy := source(side)
	tx, ty := target(side)
	s := 1.0
	if q.Archive == "WEAPPICS" && len(hands) == 2 {
		sx, sy = source(0)
		x1, y1 := source(1)
		tx, ty = target(0)
		u1, v1 := target(1)
		s = (u1 - tx) / (x1 - sx)
		sy = (sy + y1) / 2
		ty = (ty + v1) / 2
	}
	ox, oy := tx-s*sx, ty-s*sy
	return image.Rect(int(math.Round(float64(r.Min.X)*s+ox)), int(math.Round(float64(r.Min.Y)*s+oy)), int(math.Round(float64(r.Max.X)*s+ox)), int(math.Round(float64(r.Max.Y)*s+oy)))
}
func (p *hdEquipment) figure(n *hdNativeRender, dst *ebiten.Image, sprites Sprites, group []portrait.Piece) error {
	body := group[0].Index
	indices := map[string]int{"HEADPICS": -1, "TORPICS": -1, "LEGSPICS": -1}
	for _, q := range group {
		if _, ok := indices[q.Archive]; ok {
			if q.Index < body*9 || q.Index >= body*9+9 {
				return fmt.Errorf("HD體型與裝備圖號不符")
			}
			indices[q.Archive] = q.Index
		}
	}
	mask := 0
	if indices["HEADPICS"] >= 0 {
		mask |= 1
	}
	if indices["TORPICS"] >= 0 {
		mask |= 2
	}
	if leg := indices["LEGSPICS"]; leg >= 0 && p.Parts[fmt.Sprintf("LEGSPICS:%d", leg)].Kind == "" {
		mask |= 4
	}
	pic, r := p.body(n, body, mask)
	p.Figure.Clear()
	if body == 3 {
		p.part(n, p.Figure, "WINGS:3")
	}
	var base ebiten.DrawImageOptions
	base.GeoM.Translate(float64(r.Min.X), float64(r.Min.Y))
	p.Figure.DrawImage(pic, &base)
	for _, archive := range []string{"LEGSPICS", "TORPICS", "HEADPICS"} {
		if i := indices[archive]; i >= 0 {
			key := fmt.Sprintf("%s:%d", archive, i)
			if _, ok := p.Meshes[key]; ok {
				r := equipmentRect(p.Parts[key].World)
				var op ebiten.DrawImageOptions
				op.GeoM.Translate(float64(r.Min.X), float64(r.Min.Y))
				p.Figure.DrawImage(p.torso(n, key), &op)
			} else {
				p.part(n, p.Figure, key)
			}
		}
	}
	dx := (group[0].X - p.BaseOrigins[fmt.Sprint(body)]) * 48
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(float64(dx), 0)
	dst.DrawImage(p.Figure, &op)
	for _, q := range group[1:] {
		if _, ok := indices[q.Archive]; ok {
			continue
		}
		if q.Archive == "WEAPPICS" || q.Archive == "SHLDPICS" {
			key := fmt.Sprintf("%s:%d", q.Archive, q.Index)
			a, ok := p.Props[key]
			if !ok || q.W*8 != a.SourceSize[0] || q.H != a.SourceSize[1] {
				return fmt.Errorf("HD武器來源不符")
			}
			r := p.propRect(group, body, indices["TORPICS"], dx, q, a)
			if r.Empty() {
				return fmt.Errorf("HD持握矩形不符")
			}
			n.h.drawHD(dst, p.Sources[key], r)
			continue
		}
		if q.Archive == "HANDPICS" {
			side := q.Index - body*2
			g := p.Grips[fmt.Sprintf("%d:%d", body, indices["TORPICS"])]
			if side < 0 || side >= len(g) {
				return fmt.Errorf("HD手部圖號不符")
			}
			r := gripRect(g[side]).Intersect(p.Figure.Bounds())
			if r.Empty() {
				return fmt.Errorf("HD手部矩形不符")
			}
			var hand ebiten.DrawImageOptions
			hand.GeoM.Translate(float64(r.Min.X+dx), float64(r.Min.Y))
			dst.DrawImage(p.Figure.SubImage(r).(*ebiten.Image), &hand)
			continue
		}
		if q.Archive == portrait.ArchiveDefaultShield {
			if err := p.shieldFace(n, dst, sprites, group, body, indices["TORPICS"], dx, q); err != nil {
				return err
			}
			continue
		}
		return fmt.Errorf("HD裝備不支援圖層%s", q.Archive)
	}
	return nil
}

// shieldFace keeps the editable bitmap's pixels and follows its own template.
// Only presentation coordinates change; Figure, Piece and saved bitmap stay intact.
func (p *hdEquipment) shieldFaceRect(group []portrait.Piece, body, torso, dx int, face portrait.Piece) (image.Rectangle, image.Rectangle, error) {
	var parent *portrait.Piece
	for i := range group {
		q := &group[i]
		if q.Archive == "SHLDPICS" && q.Index == portrait.PersonalShieldStyle {
			parent = q
			break
		}
	}
	if parent == nil {
		return image.Rectangle{}, image.Rectangle{}, fmt.Errorf("HD個人盾牌缺少模板")
	}
	a, ok := p.Props[fmt.Sprintf("SHLDPICS:%d", parent.Index)]
	if !ok {
		return image.Rectangle{}, image.Rectangle{}, fmt.Errorf("HD個人盾牌模板不符")
	}
	original := image.Rect((parent.X*8+a.Paint[0])*6, (parent.Y+a.Paint[1])*6, (parent.X*8+a.Paint[2])*6, (parent.Y+a.Paint[3])*6)
	target := original
	if body != 0 {
		target = p.propRect(group, body, torso, dx, *parent, a)
	}
	if original.Empty() || target.Empty() {
		return image.Rectangle{}, image.Rectangle{}, fmt.Errorf("HD個人盾牌矩形不符")
	}
	sx, sy := float64(target.Dx())/float64(original.Dx()), float64(target.Dy())/float64(original.Dy())
	x := func(v int) int { return int(math.Round(float64(target.Min.X) + float64(v-original.Min.X)*sx)) }
	y := func(v int) int { return int(math.Round(float64(target.Min.Y) + float64(v-original.Min.Y)*sy)) }
	child := image.Rect(x(face.X*48), y(face.Y*6), x((face.X*8+face.W*8)*6), y((face.Y+face.H)*6))
	return child, target, nil
}
func (p *hdEquipment) shieldFace(n *hdNativeRender, dst *ebiten.Image, sprites Sprites, group []portrait.Piece, body, torso, dx int, face portrait.Piece) error {
	rect, _, err := p.shieldFaceRect(group, body, torso, dx, face)
	if err != nil {
		return err
	}
	pic, err := pieceImage(sprites, face)
	if err != nil {
		return err
	}
	if pic.Bounds().Dx() != 40 || pic.Bounds().Dy() != 38 {
		return fmt.Errorf("HD個人盾牌必須為40×38")
	}
	if p.defaultShieldRGB == "" {
		return fmt.Errorf("HD預設盾面尚未核對")
	}
	if n.h.fingerprint(pic) == p.defaultShieldRGB {
		// The known empty default uses its already-painted HD shield template.
		// A changed bitmap keeps the complete original editable pixel grid below.
		return nil
	}
	var op ebiten.DrawImageOptions
	op.GeoM.Scale(float64(rect.Dx())/40, float64(rect.Dy())/38)
	op.GeoM.Translate(float64(rect.Min.X), float64(rect.Min.Y))
	op.Filter = ebiten.FilterNearest
	dst.DrawImage(pic, &op)
	return nil
}

func (p *hdEquipment) exposedArms() *ebiten.Image {
	const key = "0:exposed-arms"
	if pic := p.BodyShapes[key]; pic != nil {
		return pic
	}
	r := equipmentRect(p.Parts["PERPICS:0"].World)
	pixels := make([]byte, 1920*1200*4)
	p.Base.ReadPixels(pixels)
	result := make([]byte, len(pixels))
	limit := 0
	for side := 0; side < 2; side++ {
		limit = max(limit, p.SourceHands[fmt.Sprint(side)][3]*6+60)
	}
	for y := max(0, r.Min.Y); y < min(limit, r.Max.Y); y++ {
		runs := []image.Rectangle{}
		start := -1
		for x := max(0, r.Min.X); x < min(1920, r.Max.X); x++ {
			on := pixels[(y*1920+x)*4+3] >= 8
			if on && start < 0 {
				start = x
			}
			if !on && start >= 0 {
				if x-start >= 3 {
					runs = append(runs, image.Rect(start, y, x, y+1))
				}
				start = -1
			}
		}
		if start >= 0 && r.Max.X-start >= 3 {
			runs = append(runs, image.Rect(start, y, min(1920, r.Max.X), y+1))
		}
		if len(runs) < 3 {
			continue
		}
		for _, run := range []image.Rectangle{runs[0], runs[len(runs)-1]} {
			first, last := (y*1920+run.Min.X)*4, (y*1920+run.Max.X)*4
			copy(result[first:last], pixels[first:last])
		}
	}
	pic := ebiten.NewImage(1920, 1200)
	pic.WritePixels(result)
	p.BodyShapes[key] = pic
	return pic
}
