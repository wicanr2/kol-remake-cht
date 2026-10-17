package ui

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"image"
)

type HDMonsterArt struct {
	Source HDPageSource
	Size   image.Point
	Paint  image.Image
}
type hdMonsterArt struct {
	source HDPageSource
	size   image.Point
	paint  *ebiten.Image
}
type hdMonsters struct{ pictures [32]hdMonsterArt }
type HDMonsterPlacement struct {
	Index     int
	Original  *ebiten.Image
	Paint     *ebiten.Image
	Rect, Fit image.Rectangle
}

// AttachMonsters 全部檢查後才建立與交換本物件的快取。
func (h *HD) AttachMonsters(arts []HDMonsterArt) error {
	if h.pages == nil || len(arts) != 32 {
		return fmt.Errorf("HD怪物掛接需要完整32圖")
	}
	seen := map[string]bool{}
	nextSources := append([]HDPageSource{}, h.pages.sources...)
	for i, a := range arts {
		if a.Source.Archive != "MONPICS" || a.Source.Index != i || len(a.Source.RGB) != 64 || seen[a.Source.RGB] || a.Paint == nil {
			return fmt.Errorf("HD怪物來源、順序或圖片不符")
		}
		if _, err := MonsterSheetOrigin(a.Size); err != nil {
			return err
		}
		if a.Paint.Bounds().Dx() < a.Size.X*6 || a.Paint.Bounds().Dy() < a.Size.Y*6 {
			return fmt.Errorf("HD怪物原生尺寸不足")
		}
		seen[a.Source.RGB] = true
		found := false
		for _, old := range h.pages.sources {
			if old.Archive == a.Source.Archive && old.Index == a.Source.Index {
				if old.RGB != a.Source.RGB {
					return fmt.Errorf("HD怪物覆蓋既有來源")
				}
				found = true
			} else if old.RGB == a.Source.RGB {
				return fmt.Errorf("HD怪物與其他來源指紋衝突")
			}
		}
		if !found {
			nextSources = append(nextSources, a.Source)
		}
	}
	next := &hdMonsters{}
	for i, a := range arts {
		next.pictures[i] = hdMonsterArt{a.Source, a.Size, h.newPaint(a.Paint)}
	}
	h.monsters = next
	h.pages.sources = nextSources
	h.pages.validated = false
	return nil
}

// MonsterPlacements 沿既有資訊框與State.Spots順序，不改戰鬥狀態。
func (h *HD) MonsterPlacements(s *battle.State, sprites Sprites) ([]HDMonsterPlacement, error) {
	if h.monsters == nil || s == nil {
		return nil, nil
	}
	index := 0
	var origins []image.Point
	if s.Stage == battle.StageEncounter {
		if s.Site == nil {
			return nil, fmt.Errorf("HD怪物遭遇缺少場景")
		}
		index = s.Site.Monster
		for _, p := range s.Spots() {
			origins = append(origins, image.Pt(p[0]*6, p[1]*6))
		}
	} else {
		d, shown := s.EquipmentPanel()
		if !shown || !d.Monster || d.Equipment {
			return nil, nil
		}
		index = d.Portrait
	}
	if index < 0 || index >= len(h.monsters.pictures) {
		return nil, fmt.Errorf("HD怪物圖號越界")
	}
	a := h.monsters.pictures[index]
	original, err := sprites.Get("MONPICS", index)
	if err != nil {
		return nil, err
	}
	if original == nil || original.Bounds().Size() != a.size || h.fingerprint(original) != a.source.RGB {
		return nil, fmt.Errorf("HD怪物實際來源尺寸或指紋不符")
	}
	if s.Stage != battle.StageEncounter {
		p, err := MonsterSheetOrigin(a.size)
		if err != nil {
			return nil, err
		}
		origins = []image.Point{p.Mul(6)}
	}
	var out []HDMonsterPlacement
	for _, p := range origins {
		r := image.Rectangle{Min: p, Max: p.Add(a.size.Mul(6))}
		if !r.In(image.Rect(0, 0, 1920, 1200)) {
			return nil, fmt.Errorf("HD怪物顯示矩形越界")
		}
		out = append(out, HDMonsterPlacement{index, original, a.paint, r, fitHDPortrait(a.paint.Bounds(), r)})
	}
	return out, nil
}

func (h *HD) DrawMonsters(dst *ebiten.Image, placements []HDMonsterPlacement) {
	for _, p := range placements {
		h.drawHD(dst, p.Paint, p.Fit)
	}
}
