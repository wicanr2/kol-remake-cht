package ui

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/kol-remake-cht/internal/town"
	"image"
	"image/color"
	"math"
)

type HDTownFacadeArt struct {
	Source HDPageSource
	Paint  image.Image
}
type HDTownFacadeSign struct {
	NativeSize           image.Point
	FontSize, LineHeight int
	Full, Clip           [3]image.Rectangle
}
type hdTownFacadeArt struct {
	source HDPageSource
	paint  *ebiten.Image
}
type hdTownFacades struct {
	pictures map[string]hdTownFacadeArt
	sign     HDTownFacadeSign
}

func townFacadeID(archive string, index int) string { return fmt.Sprintf("%s:%d", archive, index) }

// AttachTownFacades exchanges the complete group only after all bindings validate.
func (h *HD) AttachTownFacades(arts []HDTownFacadeArt, sign HDTownFacadeSign) error {
	if h.pages == nil || h.style == nil || len(arts) != 103 || sign.NativeSize != image.Pt(1774, 887) || sign.FontSize != 32 || sign.LineHeight != 34 {
		return fmt.Errorf("HD城鎮門面需要完整103圖與原生告示字模")
	}
	for _, r := range "請勿踐踏草坪" {
		if _, ok := h.style.glyphs[string(r)]; !ok {
			return fmt.Errorf("HD住宅告示缺字：%c", r)
		}
	}
	bounds := image.Rectangle{Max: sign.NativeSize}
	for i, r := range sign.Full {
		if r.Empty() || sign.Clip[i].Empty() || !sign.Clip[i].In(r) || !sign.Clip[i].In(bounds) {
			return fmt.Errorf("HD住宅告示安全矩形不符")
		}
	}
	wanted := []string{townFacadeID("MISCPICS", 0)}
	for t, count := range []int{33, 29, 25, 25, 29, 25, 45} {
		for i := 3; i <= count; i += 2 {
			wanted = append(wanted, townFacadeID(fmt.Sprintf("TWN%dPICS", t+1), i))
		}
	}
	sources := append([]HDPageSource{}, h.pages.sources...)
	seen := map[string]bool{}
	for i, a := range arts {
		if townFacadeID(a.Source.Archive, a.Source.Index) != wanted[i] || len(a.Source.RGB) != 64 || seen[a.Source.RGB] || a.Paint == nil {
			return fmt.Errorf("HD城鎮門面來源、順序或圖片不符")
		}
		b := a.Paint.Bounds()
		if b.Dx() < 1152 || b.Dy() < 576 || b.Dx() != b.Dy()*2 {
			return fmt.Errorf("HD城鎮門面原生尺寸不足")
		}
		if a.Source.Archive == "TWN7PICS" && a.Source.Index == 13 && a.Source.RGB != h.facadeKey {
			return fmt.Errorf("HD原核准旅店來源遭改動")
		}
		seen[a.Source.RGB] = true
		found := false
		for _, old := range h.pages.sources {
			if old.Archive == a.Source.Archive && old.Index == a.Source.Index {
				if old.RGB != a.Source.RGB {
					return fmt.Errorf("HD城鎮門面覆蓋既有來源")
				}
				found = true
			} else if old.RGB == a.Source.RGB {
				return fmt.Errorf("HD城鎮門面與其他來源指紋衝突")
			}
		}
		if !found {
			sources = append(sources, a.Source)
		}
	}
	next := &hdTownFacades{pictures: map[string]hdTownFacadeArt{}, sign: sign}
	for _, a := range arts {
		next.pictures[townFacadeID(a.Source.Archive, a.Source.Index)] = hdTownFacadeArt{a.Source, h.newPaint(a.Paint)}
	}
	h.facades = next
	h.pages.sources = sources
	h.pages.validated = false
	return nil
}
func townFacadeHDRect() image.Rectangle {
	l := town.CJKOutdoor
	return image.Rect(l.ViewX*2, l.ViewY*2, (l.ViewX+192*l.Scale)*2, (l.ViewY+96*l.Scale)*2)
}
func (h *HD) drawTownFacade(dst *ebiten.Image, sprites Sprites, out *town.Outdoor) (bool, error) {
	if h.facades == nil {
		return false, nil
	}
	archive, index := "", 0
	if out.AtEdge {
		archive = "MISCPICS"
	} else if door := out.AtDoor(); door != nil {
		archive, index = out.CrestArchive(), door.Picture
	} else {
		return false, nil
	}
	a, ok := h.facades.pictures[townFacadeID(archive, index)]
	if !ok {
		return false, fmt.Errorf("HD城鎮門面實際圖號不在完整組")
	}
	pic, err := sprites.Get(archive, index)
	if err != nil {
		return false, err
	}
	if pic == nil || pic.Bounds().Size() != image.Pt(192, 96) || h.fingerprint(pic) != a.source.RGB {
		return false, fmt.Errorf("HD城鎮門面實際來源尺寸或指紋不符")
	}
	h.drawHD(dst, a.paint, townFacadeHDRect())
	if archive == "TWN5PICS" && index == 19 {
		if err := h.drawTownResidenceSign(dst); err != nil {
			return false, err
		}
	}
	return true, nil
}
func (h *HD) drawTownResidenceSign(dst *ebiten.Image) error {
	sign := h.facades.sign
	viewport := townFacadeHDRect()
	native := func(r image.Rectangle) image.Rectangle {
		x := func(v int) int {
			return viewport.Min.X + int(math.Round(float64(v)*float64(viewport.Dx())/float64(sign.NativeSize.X)))
		}
		y := func(v int) int {
			return viewport.Min.Y + int(math.Round(float64(v)*float64(viewport.Dy())/float64(sign.NativeSize.Y)))
		}
		return image.Rect(x(r.Min.X), y(r.Min.Y), x(r.Max.X), y(r.Max.Y))
	}
	for i, b := range sign.Full {
		full, clip := native(b), native(sign.Clip[i])
		panel := dst.SubImage(clip).(*ebiten.Image)
		y := full.Min.Y + (full.Dy()-2*sign.LineHeight)/2
		if 2*sign.LineHeight > full.Dy() {
			return fmt.Errorf("HD住宅告示文字過高")
		}
		for lineNumber, line := range []string{"請勿", "踐踏草坪"} {
			if h.style.measure(line, float64(sign.FontSize)) > float64(full.Dx()) {
				return fmt.Errorf("HD住宅告示文字過寬")
			}
			r := image.Rect(full.Min.X, y+lineNumber*sign.LineHeight, full.Max.X, y+(lineNumber+1)*sign.LineHeight)
			h.style.label(panel, line, r, float64(sign.FontSize), color.RGBA{85, 85, 85, 255}, true)
		}
	}
	return nil
}
