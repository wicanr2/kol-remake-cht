package ui

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/font"
)

func drawBattleCommandPanel(dst *ebiten.Image, sprites Sprites, f *font.Font, s *battle.State) error {
	for _, item := range s.CommandPanel() {
		pic, err := sprites.Get("KOLMPICS", item.Code+battle.IconOffset)
		if err != nil {
			return err
		}
		if pic == nil || pic.Bounds().Dx() != 24 || pic.Bounds().Dy() != 20 {
			return fmt.Errorf("命令圖示 %d 缺少 24×20 圖片", item.Code)
		}
		f.Draw(dst, item.X, item.Y-8, item.Label, 255, 255, 255)
		draw(dst, pic, item.X, item.Y)
	}
	return nil
}
