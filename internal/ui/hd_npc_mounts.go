package ui

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// AttachNPCMounts 先檢查完整組與四個保留NPC，再交換呈現快取。
func (h *HD) AttachNPCMounts(portraits, mounts map[string]image.Image, preserved, sources []HDPageSource) error {
	if h.pages == nil || h.style == nil || len(portraits) != 98 || len(mounts) != 4 || len(preserved) != 4 || len(sources) != 106 {
		return fmt.Errorf("HD NPC與坐騎掛接狀態或完整組數量不符")
	}
	kept := map[string]bool{}
	for i, s := range preserved {
		if kept[s.RGB] || i < 3 && h.pages.portraits[s.RGB] == nil || i == 3 && s.RGB != h.style.npcKey {
			return fmt.Errorf("HD保留NPC指紋或既有呈現不符")
		}
		kept[s.RGB] = true
	}
	seen := map[string]bool{}
	sizes := map[string]image.Point{}
	for _, s := range sources {
		if seen[s.RGB] {
			return fmt.Errorf("HD NPC與坐騎來源重複")
		}
		seen[s.RGB] = true
		size := image.Pt(88, 64)
		pic := portraits[s.RGB]
		minimum := image.Pt(528, 384)
		if s.Archive == "STEEDS" {
			pic = mounts[s.RGB]
			size = image.Pt(40, 38)
			minimum = image.Pt(240, 228)
		}
		sizes[s.RGB] = size
		if kept[s.RGB] {
			continue
		}
		if pic == nil || pic.Bounds().Dx() < minimum.X || pic.Bounds().Dy() < minimum.Y ||
			h.pages.portraits[s.RGB] != nil || h.pages.icons[s.RGB] != nil || s.RGB == h.style.npcKey {
			return fmt.Errorf("HD NPC與坐騎缺圖、原生尺寸不足或覆蓋既有圖片")
		}
		if _, ok := h.matching[s.RGB]; ok {
			return fmt.Errorf("HD NPC與坐騎覆蓋角色指紋")
		}
	}
	for key := range portraits {
		if !seen[key] || kept[key] || mounts[key] != nil {
			return fmt.Errorf("HD NPC圖片缺少來源或種類衝突")
		}
	}
	for key := range mounts {
		if !seen[key] || kept[key] {
			return fmt.Errorf("HD坐騎圖片缺少來源")
		}
	}
	oldSources := map[string]string{}
	for _, s := range h.pages.sources {
		oldSources[fmt.Sprintf("%s:%d", s.Archive, s.Index)] = s.RGB
	}
	nextSources := append([]HDPageSource{}, h.pages.sources...)
	for _, s := range sources {
		key := fmt.Sprintf("%s:%d", s.Archive, s.Index)
		if old, ok := oldSources[key]; ok {
			if old != s.RGB {
				return fmt.Errorf("HD NPC與坐騎覆蓋既有來源定位")
			}
		} else {
			nextSources = append(nextSources, s)
		}
	}
	nextPortraits := map[string]*ebiten.Image{}
	for key, pic := range h.pages.portraits {
		nextPortraits[key] = pic
	}
	nextIcons := map[string]*ebiten.Image{}
	for key, pic := range h.pages.icons {
		nextIcons[key] = pic
	}
	nextFit := map[*ebiten.Image]bool{}
	for pic, fit := range h.portraitFit {
		nextFit[pic] = fit
	}
	for key, pic := range portraits {
		nextPortraits[key] = h.newPaint(pic)
	}
	for key, pic := range mounts {
		painted := h.newPaint(pic)
		nextIcons[key] = painted
		nextFit[painted] = true
	}
	h.pages.portraits, h.pages.icons = nextPortraits, nextIcons
	h.pages.sources, h.pages.npcMountSizes = nextSources, sizes
	h.pages.validated = false
	h.portraitFit = nextFit
	return nil
}
