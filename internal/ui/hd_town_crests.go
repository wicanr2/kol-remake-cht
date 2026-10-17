package ui

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/kol-remake-cht/internal/town"
	"image"
)

type HDTownCrestArt struct {
	Source HDPageSource
	Paint  image.Image
}
type hdTownCrestArt struct {
	source HDPageSource
	paint  *ebiten.Image
}
type hdTownCrests struct {
	pictures map[string]hdTownCrestArt
	byRGB    map[string]*ebiten.Image
}

func (h *HD) AttachTownCrests(arts []HDTownCrestArt) error {
	if h.pages == nil || h.style == nil || len(arts) != 7 {
		return fmt.Errorf("HD七城徽章缺少完整組或介面")
	}
	sources := append([]HDPageSource{}, h.pages.sources...)
	seen := map[string]bool{}
	for i, a := range arts {
		if a.Source.Archive != fmt.Sprintf("TWN%dPICS", i+1) || a.Source.Index != 0 || len(a.Source.RGB) != 64 || seen[a.Source.RGB] || a.Paint == nil || a.Paint.Bounds().Size() != image.Pt(320, 320) {
			return fmt.Errorf("HD七城徽章順序、來源或原生尺寸不符")
		}
		if i == 6 && a.Source.RGB != h.style.crestKey {
			return fmt.Errorf("HD原核准徽章來源遭改動")
		}
		seen[a.Source.RGB] = true
		found := false
		for _, old := range h.pages.sources {
			if old.Archive == a.Source.Archive && old.Index == a.Source.Index {
				if old.RGB != a.Source.RGB {
					return fmt.Errorf("HD七城徽章覆蓋既有來源")
				}
				found = true
			} else if old.RGB == a.Source.RGB {
				return fmt.Errorf("HD七城徽章來源指紋衝突")
			}
		}
		if !found {
			sources = append(sources, a.Source)
		}
	}
	next := &hdTownCrests{pictures: map[string]hdTownCrestArt{}, byRGB: map[string]*ebiten.Image{}}
	for _, a := range arts {
		paint := h.newPaint(a.Paint)
		next.pictures[a.Source.Archive] = hdTownCrestArt{a.Source, paint}
		next.byRGB[a.Source.RGB] = paint
	}
	h.crests = next
	h.pages.sources = sources
	h.pages.validated = false
	return nil
}

func (h *HD) drawTownCrest(dst *ebiten.Image, sprites Sprites, out *town.Outdoor) (bool, error) {
	if h.crests == nil {
		return false, nil
	}
	a, ok := h.crests.pictures[out.CrestArchive()]
	if !ok {
		return false, fmt.Errorf("HD七城徽章缺少目前城鎮")
	}
	pic, err := sprites.Get(out.CrestArchive(), town.Crest.Index)
	if err != nil {
		return false, err
	}
	if pic == nil || pic.Bounds().Size() != image.Pt(56, 57) || h.fingerprint(pic) != a.source.RGB {
		return false, fmt.Errorf("HD七城徽章實際來源尺寸或指紋不符")
	}
	h.drawHD(dst, a.paint, hdRect(130, 502, 320, 320))
	return true, nil
}
