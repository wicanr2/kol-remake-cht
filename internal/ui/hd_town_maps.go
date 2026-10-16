package ui

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/kol-remake-cht/internal/town"
	"github.com/wicanr2/kol-remake-cht/internal/world"
	"image"
	"image/color"
	imagedraw "image/draw"
)

// HDTownMapSource 保留驗證過的索引地圖；手繪圖以地圖位置查區塊。
type HDTownMapSource struct {
	GridSource, AtlasSource HDPageSource
	Grid                    *image.Gray
}
type HDTownMapArt struct {
	Archive                     string
	SourceBounds, VisibleBounds image.Rectangle
	Paint                       image.Image
}
type hdTownMapCity struct {
	source HDTownMapSource
	owners [4096]int
}
type hdTownMapRegion struct {
	source image.Rectangle
	paint  *ebiten.Image
}
type hdTownMaps struct {
	cities  map[string]*hdTownMapCity
	regions []hdTownMapRegion
}

func (h *HD) AttachTownMaps(sources []HDTownMapSource, arts []HDTownMapArt) error {
	if h.pages == nil || h.style == nil || len(sources) == 0 || len(arts) == 0 {
		return fmt.Errorf("HD城鎮地圖缺少介面或完整組")
	}
	next := &hdTownMaps{cities: map[string]*hdTownMapCity{}}
	for _, s := range sources {
		archive := s.GridSource.Archive
		if len(archive) != 5 || archive[:4] != "TOWN" || archive[4] < '0' || archive[4] > '7' || s.AtlasSource.Archive != archive || s.GridSource.Index != 0 || s.AtlasSource.Index != 1 || len(s.GridSource.RGB) != 64 || len(s.AtlasSource.RGB) != 64 || s.Grid == nil || s.Grid.Bounds() != image.Rect(0, 0, 64, 64) || next.cities[archive] != nil {
			return fmt.Errorf("HD城鎮地圖來源或完整組不符")
		}
		grid := image.NewGray(s.Grid.Bounds())
		imagedraw.Draw(grid, grid.Bounds(), s.Grid, grid.Bounds().Min, imagedraw.Src)
		s.Grid = grid
		city := &hdTownMapCity{source: s}
		for i := range city.owners {
			city.owners[i] = -1
		}
		next.cities[archive] = city
	}
	for i, a := range arts {
		city := next.cities[a.Archive]
		if city == nil || a.Paint == nil || a.Paint.Bounds().Dx() != a.Paint.Bounds().Dy() || a.Paint.Bounds().Dx() < 768 || a.Paint.Bounds().Dx() > 4096 || a.SourceBounds.Dx() != 12 || a.SourceBounds.Dy() != 12 || !a.SourceBounds.In(image.Rect(0, 0, 64, 64)) || a.VisibleBounds.Empty() || !a.VisibleBounds.In(a.SourceBounds) {
			return fmt.Errorf("HD城鎮地圖畫稿尺寸或覆蓋範圍不符")
		}
		for y := a.VisibleBounds.Min.Y; y < a.VisibleBounds.Max.Y; y++ {
			for x := a.VisibleBounds.Min.X; x < a.VisibleBounds.Max.X; x++ {
				p := y*64 + x
				if city.owners[p] != -1 {
					return fmt.Errorf("HD城鎮地圖區塊重疊")
				}
				city.owners[p] = i
			}
		}
	}
	var required []HDPageSource
	for _, city := range next.cities {
		for _, owner := range city.owners {
			if owner < 0 {
				return fmt.Errorf("HD城鎮地圖覆蓋缺格")
			}
		}
		required = append(required, city.source.GridSource, city.source.AtlasSource)
	}
	allSources := append([]HDPageSource{}, h.pages.sources...)
	for _, s := range required {
		found := false
		for _, old := range allSources {
			if old.Archive == s.Archive && old.Index == s.Index {
				if old.RGB != s.RGB {
					return fmt.Errorf("HD城鎮地圖覆蓋既有來源")
				}
				found = true
			}
		}
		if !found {
			allSources = append(allSources, s)
		}
	}
	for _, a := range arts {
		next.regions = append(next.regions, hdTownMapRegion{a.SourceBounds, ebiten.NewImageFromImage(a.Paint)})
	}
	h.townMaps = next
	h.pages.sources = allSources
	h.pages.validated = false
	return nil
}

func (h *HD) drawTownMap(dst *ebiten.Image, sprites Sprites, out *town.Outdoor) (bool, error) {
	if h.townMaps == nil || out == nil || out.AtEdge || out.AtDoor() != nil {
		return false, nil
	}
	city := h.townMaps.cities[out.Archive()]
	if city == nil {
		return false, nil
	}
	b := out.Bounds
	if b.MinX < 0 || b.MinY < 0 || b.MaxX > 63 || b.MaxY > 63 || b.MinX > b.MaxX || b.MinY > b.MaxY {
		return false, fmt.Errorf("HD城鎮地圖鏡頭界限不符")
	}
	for _, s := range []struct {
		p    HDPageSource
		size image.Point
	}{{city.source.GridSource, image.Pt(64, 64)}, {city.source.AtlasSource, image.Pt(256, 256)}} {
		pic, err := sprites.Get(s.p.Archive, s.p.Index)
		if err != nil {
			return false, err
		}
		if pic == nil || pic.Bounds().Size() != s.size || h.fingerprint(pic) != s.p.RGB {
			return false, fmt.Errorf("HD城鎮地圖實際來源尺寸或指紋不符")
		}
	}
	clamp := func(v, lo, hi int) int {
		if v < lo {
			return lo
		}
		if v > hi {
			return hi
		}
		return v
	}
	type cell struct{ region, x, y, col, row int }
	cells := make([]cell, 0, 72)
	// Check the actual Outdoor grid before any map pixel is drawn.
	for row, line := range out.View() {
		for col, index := range line {
			x := clamp(out.Camera.X+town.ColBias+col, b.MinX, b.MaxX)
			y := clamp(out.Camera.Y+town.RowBias+row, b.MinY, b.MaxY)
			if int(city.source.Grid.GrayAt(x, y).Y) != index {
				return false, fmt.Errorf("HD城鎮地圖目前格子與驗證來源不符")
			}
			owner := city.owners[y*64+x]
			if owner < 0 || owner >= len(h.townMaps.regions) {
				return false, fmt.Errorf("HD城鎮地圖目前格子缺少畫稿")
			}
			cells = append(cells, cell{owner, x, y, col, row})
		}
	}
	l := town.CJKOutdoor
	side := l.Tile * 2
	for _, c := range cells {
		region := h.townMaps.regions[c.region]
		whole := h.sampledHD(region.paint, 12*side, 12*side)
		x := (c.x - region.source.Min.X) * side
		y := (c.y - region.source.Min.Y) * side
		tile := whole.SubImage(image.Rect(x, y, x+side, y+side)).(*ebiten.Image)
		var op ebiten.DrawImageOptions
		op.GeoM.Translate(float64(l.ViewX*2+c.col*side), float64(l.ViewY*2+c.row*side))
		op.Filter = ebiten.FilterNearest
		dst.DrawImage(tile, &op)
	}
	if h.drawMotionTownMarker(dst, out) {
		return true, nil
	}
	if overlay := out.MarkerOverlay(); overlay != nil {
		row, col := out.MarkerCell()
		for y := 0; y < town.MarkerSide; y++ {
			for x := 0; x < town.MarkerSide; x++ {
				c := overlay[y*town.MarkerSide+x]
				if c.A == 0 {
					continue
				}
				for dy := 0; dy < 6; dy++ {
					for dx := 0; dx < 6; dx++ {
						dst.Set(l.ViewX*2+col*side+x*6+dx, l.ViewY*2+row*side+y*6+dy, c)
					}
				}
			}
		}
	}
	return true, nil
}

// drawHDWorldMap runs inside the native UI pass, before prompts and controls.
func drawHDWorldMap(dst *ebiten.Image, sprites Sprites, s *world.Screen) (bool, error) {
	n := activeHDUI
	if n == nil || n.dst != dst || n.h.townMaps == nil {
		return false, nil
	}
	h := n.h
	city := h.townMaps.cities["TOWN0"]
	if city == nil {
		return false, nil
	}
	if s == nil || s.Travel == nil || s.Travel.Grid == nil {
		return false, fmt.Errorf("HD世界地圖缺少目前旅行狀態")
	}
	for _, source := range []struct {
		p    HDPageSource
		size image.Point
	}{{city.source.GridSource, image.Pt(64, 64)}, {city.source.AtlasSource, image.Pt(256, 256)}} {
		pic, err := sprites.Get(source.p.Archive, source.p.Index)
		if err != nil {
			return false, err
		}
		if pic == nil || pic.Bounds().Size() != source.size || h.fingerprint(pic) != source.p.RGB {
			return false, fmt.Errorf("HD世界地圖實際來源尺寸或指紋不符")
		}
	}
	clamp := func(v int) int {
		if v < 0 {
			return 0
		}
		if v > 63 {
			return 63
		}
		return v
	}
	type cell struct{ owner, x, y, col, row int }
	var cells []cell
	cx, cy := s.Travel.Camera()
	for row, line := range s.View() {
		for col, index := range line {
			x, y := clamp(cx+world.ColBias+col), clamp(cy+world.RowBias+row)
			if int(city.source.Grid.GrayAt(x, y).Y) != index {
				return false, fmt.Errorf("HD世界地圖目前格子與驗證來源不符")
			}
			owner := city.owners[y*64+x]
			if owner < 0 || owner >= len(h.townMaps.regions) {
				return false, fmt.Errorf("HD世界地圖目前格子缺少畫稿")
			}
			cells = append(cells, cell{owner, x, y, col, row})
		}
	}
	l := world.CJK
	side := l.Tile * 2
	for _, c := range cells {
		region := h.townMaps.regions[c.owner]
		whole := h.sampledHD(region.paint, side*12, side*12)
		x, y := (c.x-region.source.Min.X)*side, (c.y-region.source.Min.Y)*side
		tile := whole.SubImage(image.Rect(x, y, x+side, y+side)).(*ebiten.Image)
		var op ebiten.DrawImageOptions
		op.GeoM.Translate(float64(l.ViewX*2+c.col*side), float64(l.ViewY*2+c.row*side))
		dst.DrawImage(tile, &op)
	}
	// Preserve the point's size, sub-cell position and blink. EGAIndex cannot invert hand-painted RGB colors.
	if !s.MarkerOff {
		mx, my := s.Travel.MarkerPixel(l.Tile / l.Scale)
		x0, y0 := (l.ViewX+mx*l.Scale)*2, (l.ViewY+my*l.Scale)*2
		for y := 0; y < l.MarkerSide*2; y++ {
			for x := 0; x < l.MarkerSide*2; x++ {
				c := dst.At(x0+x, y0+y)
				r, g, b, a := c.RGBA()
				dst.Set(x0+x, y0+y, color.RGBA{255 - uint8(r>>8), 255 - uint8(g>>8), 255 - uint8(b>>8), uint8(a >> 8)})
			}
		}
	}
	return true, nil
}
