package ui

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/font"
	"strings"
)

// BattleSheet 使用固定大小的文字安全區，保留外框與相鄰戰場。
type BattleSheet struct{ canvas *ebiten.Image }

func (p *BattleSheet) Draw(dst *ebiten.Image, sprites Sprites, f *font.Font, s *battle.State) error {
	d, show := s.EquipmentPanel()
	if !show {
		return nil
	}
	if p.canvas == nil {
		p.canvas = ebiten.NewImage(119, 73)
	}
	p.canvas.Clear()
	text := func(x, y int, value string) { f.Draw(p.canvas, x-17, y-3, strings.ToUpper(value), 255, 255, 255) }
	if !d.Monster {
		text(24, 7, d.Name)
	}
	if d.Equipment {
		for i := 0; i < d.Count; i++ {
			text(24, 24+6*i, battle.EquipmentLabels[i])
			text(40, 24+6*i, d.Rows[i])
		}
		if d.AmmunitionLabel != "" {
			text(24, 60, fmt.Sprintf("%s%d", d.AmmunitionLabel, d.Ammunition))
		}
	} else {
		archive, x, y := "CHARPICS", 56, 37
		if d.Monster {
			archive = "MONPICS"
		}
		var pic *ebiten.Image
		var err error
		if d.Monster {
			pic, err = sprites.Get(archive, d.Portrait)
		} else {
			pic, err = memberPortrait(sprites, d.Portrait, d.Picture)
		}
		if err != nil {
			return err
		}
		if pic == nil {
			return fmt.Errorf("%s 肖像 %d 缺少圖片", archive, d.Portrait)
		}
		w, h := pic.Bounds().Dx(), pic.Bounds().Dy()
		if d.Monster {
			if w <= 0 || w%8 != 0 || w > 120 || h <= 0 || h > 74 {
				return fmt.Errorf("MONPICS %d 尺寸無法放入原版資訊框：%d×%d", d.Portrait, w, h)
			}
			x, y = (3+(15-w/8)/2)*8, 76-h
		} else if w != 40 || h != 38 {
			return fmt.Errorf("CHARPICS %d 肖像尺寸須為 40×38", d.Portrait)
		}
		draw(p.canvas, pic, x-17, y-3)
	}
	draw(dst, p.canvas, 17, 3)
	return nil
}
