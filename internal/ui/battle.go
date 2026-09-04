package ui

import (
	"strings"

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

	// 選目標格時把游標那一格也框起來。
	//
	// **這一圈還沒對拍**：原版是拿 `sub_1C683` 把那一格的精靈重畫一次
	// （`sub_24131:0x24603`），確切的視覺效果還沒解出來，這裡先用與白框同一個
	// 顏色標出位置，等取到選目標格的原版畫格再改（SPEC-010 的驗證那一節）。
	if s.Plan.Prompt == battle.PromptTarget {
		c := s.Plan.Cursor
		if c.Row >= 0 && c.Row < battle.ViewRows && c.Col >= 0 && c.Col < battle.ViewCols {
			frame(dst, v.X+c.Col*battle.TileSide, v.Y+c.Row*battle.TileSide,
				battle.TileSide, battle.TileSide, town.ColourSelected)
		}
	}

	for _, box := range []struct{ X, Y, Width, Height int }{
		{battle.SheetBox.X, battle.SheetBox.Y, battle.SheetBox.Width, battle.SheetBox.Height},
		{battle.BodyBox.X, battle.BodyBox.Y, battle.BodyBox.Width, battle.BodyBox.Height},
	} {
		shadow(dst, box.X, box.Y, box.Width, box.Height)
		fill(dst, box.X, box.Y, box.Width, box.Height, town.ColourBoxEdge)
		fill(dst, box.X+1, box.Y+1, box.Width-2, box.Height-2, town.ColourBoxFill)
	}

	// 執行段：訊息蓋掉圖示列那一塊（RE-087）。欄 3 是像素 24，行高 6 是小字型
	// 的高度，起點在掃描列 0xA8。
	// 執行段：整條訊息區換掉，圖示列與底圖都不見了（RE-089）。
	//
	// **底圖還沒認出來**：原版那一條的主色是白（5,488 個像素）加上一片亮黃的
	// 圖案（2,900 個），不是 KOLMPICS 圖 2——那一張是紅底，形狀也只對得上六成。
	// 這裡先鋪成白底，讓主色一致；黃色那一塊是已知差異。
	if s.Plan.Prompt == battle.PromptExecuting {
		box := battle.MessageBox
		fill(dst, box.X, box.Y, box.Width, box.Height, battle.MessageBack)
		area := battle.MessageArea
		for i, line := range s.Exec.Text {
			y := area.Y + i*area.LineHeight
			// 原版把訊息整段轉成大寫（KOLDATA 裡是小寫的 `moves!`，
			// 畫面上是 `MOVES!`）——與角色資料頁同一種處理（RE-074）。
			f.Draw(dst, area.Col*8, y, strings.ToUpper(line),
				battle.MessageInk[0], battle.MessageInk[1], battle.MessageInk[2])
		}
		return nil
	}

	return drawBattleIcons(dst, sprites, s)
}

// drawBattleIcons 畫指令圖示列。版面與城鎮同一組（RE-080：(32 + 32i, 171)、
// 一格 24×20），KOLMPICS 的圖號是命令碼加三。
func drawBattleIcons(dst *ebiten.Image, sprites Sprites, s *battle.State) error {
	b := town.IconBar
	for i, code := range s.Plan.Bar {
		icon, err := sprites.Get("KOLMPICS", code+battle.IconOffset)
		if err != nil {
			return err
		}
		x := b.X + i*b.Step
		draw(dst, icon, x, b.Y+1)
		c := town.ColourIdle
		if i == s.Plan.Selected {
			c = town.ColourSelected
		}
		frame(dst, x, b.Y, b.Width, b.Height, c)
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
