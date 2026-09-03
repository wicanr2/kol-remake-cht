package ui

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/font"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

// BattleEncounter 畫遭遇那一張滿版怪物圖（RE-079）。原版先把 (14,2)–(305,156)
// 清掉，畫第一隻怪物，再把畫好的那一塊複製到 `MonsterSpots` 的其餘位置，
// 最後在上面蓋一個黑底白框的文字框。
func BattleEncounter(dst *ebiten.Image, sprites Sprites, f *font.Font,
	s *battle.State, lines []string) error {
	bg, err := sprites.Get(town.Background.Archive, town.Background.Index)
	if err != nil {
		return err
	}
	draw(dst, bg, 0, 0)
	// 原版先在 (14,2)–(305,156) 畫一個青色外框的框，裡面填黑，右下有黑色投影
	// （`sub_2A975:0x2A992` 的 `sub_1BA9D(0Eh, 2, 131h, 9Ch, 3)`）。
	c := battle.EncounterClear
	shadow(dst, c.X, c.Y, c.Width, c.Height)
	fill(dst, c.X, c.Y, c.Width, c.Height, battle.ColourFieldEdge)
	fill(dst, c.X+1, c.Y+1, c.Width-2, c.Height-2, town.ColourBoxFill)

	pic, err := sprites.Get(battle.MonsterArchive, s.Site.Monster)
	if err != nil {
		return err
	}

	for _, spot := range s.Spots() {
		draw(dst, pic, spot[0], spot[1])
	}

	box := battle.EncounterBox
	fill(dst, box.X, box.Y, box.Width, box.Height, town.ColourBoxFill)
	fr := battle.EncounterFrame
	frame(dst, fr.X, fr.Y, fr.Width, fr.Height, town.ColourBoxEdge)
	t := battle.EncounterText
	for i, line := range lines {
		f.Draw(dst, t.X, t.Y+i*t.Step, line,
			town.ColourText[0], town.ColourText[1], town.ColourText[2])
	}

	// 圖示列只剩一格「按空白鍵繼續」。
	icon, err := sprites.Get("KOLMPICS", town.IconSprite(town.IconSpacebar))
	if err != nil {
		return err
	}
	b := town.IconBar
	draw(dst, icon, b.X, b.Y+1)
	frame(dst, b.X, b.Y, b.Width, b.Height, town.ColourSelected)
	return nil
}

// BattleField 畫戰場那一頁（RE-079）：右邊是 9×9 格的戰場視窗，
// 左上是裝備欄、左下是人形框，底下是指令圖示列。
//
// tiles 是 `tools/extract_battle_tiles.py` 產生的磚圖。任務地點的戰場只畫
// 視線通得到的格子，看不到的畫成 TileUnseen；目前輪到的那一格永遠畫得出來，
// 外面再套一圈白框（`sub_223C0:0x2279E`）。
func BattleField(dst *ebiten.Image, sprites Sprites, tiles *BattleTiles,
	f *font.Font, s *battle.State) error {
	bg, err := sprites.Get(town.Background.Archive, town.Background.Index)
	if err != nil {
		return err
	}
	draw(dst, bg, 0, 0)

	w := battle.FieldWindow
	shadow(dst, w.X, w.Y, w.Width, w.Height)
	fill(dst, w.X, w.Y, w.Width, w.Height, town.ColourBoxEdge)
	cx, cy := s.Camera()
	sight := s.Field.Visible(cx, cy)
	v := battle.FieldView
	for r := 0; r < battle.ViewRows; r++ {
		for c := 0; c < battle.ViewCols; c++ {
			row := clampField(cx - battle.ViewCentre.Row + r)
			col := clampField(cy - battle.ViewCentre.Col + c)
			centre := r == battle.ViewCentre.Row && c == battle.ViewCentre.Col
			img := tiles.Tile(battle.TileUnseen)
			if sight[r][c] || centre {
				tile := s.Field.At(row, col)
				if sprite := s.SpriteAt(row, col); sprite >= 0 {
					img = tiles.WithSprite(tile, sprite)
				} else {
					img = tiles.Tile(tile)
				}
			}
			draw(dst, img, v.X+c*battle.TileSide, v.Y+r*battle.TileSide)
		}
	}
	cur := battle.FieldCursor
	frame(dst, cur.X, cur.Y, cur.Width, cur.Height, town.ColourSelected)

	for _, box := range []struct{ X, Y, Width, Height int }{
		{battle.SheetBox.X, battle.SheetBox.Y, battle.SheetBox.Width, battle.SheetBox.Height},
		{battle.BodyBox.X, battle.BodyBox.Y, battle.BodyBox.Width, battle.BodyBox.Height},
	} {
		shadow(dst, box.X, box.Y, box.Width, box.Height)
		fill(dst, box.X, box.Y, box.Width, box.Height, town.ColourBoxEdge)
		fill(dst, box.X+1, box.Y+1, box.Width-2, box.Height-2, town.ColourBoxFill)
	}
	return nil
}

// clampField 把戰場座標夾在 0…63，與原版視窗的四道夾限相同。
func clampField(v int) int {
	if v < 0 {
		return 0
	}
	if v > battle.FieldSide-1 {
		return battle.FieldSide - 1
	}
	return v
}
