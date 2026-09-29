package ui

import (
	"image"
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
	f *font.Font, s *battle.State, body *BattleBody, sheet *BattleSheet) error {
	bg, err := sprites.Get(town.Background.Archive, town.Background.Index)
	if err != nil {
		return err
	}
	draw(dst, bg, 0, 0)

	w := battle.FieldWindow
	shadow(dst, w.X, w.Y, w.Width, w.Height)
	fill(dst, w.X, w.Y, w.Width, w.Height, town.ColourBoxEdge)
	cx, cy := s.Camera()
	// 騎著的人看得過牆（參戰者記錄 +0x1E，RE-102）。
	mounted := false
	if s.Cursor >= 0 && s.Cursor < len(s.Actors) {
		mounted = s.Actors[s.Cursor].Flag1E != 0
	}
	sight := s.Field.Visible(cx, cy, mounted)
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

	// 主指令列重畫時標出這一名已決定的目標格（`0x1F16A`，RE-160）。原版同樣是用
	// `sub_1C683` 重畫那一格的精靈，視覺效果與選目標格那一圈同一個未對拍的近似。
	if s.Plan.Prompt == battle.PromptCommand && s.Plan.SummaryShown {
		c := s.Plan.Summary
		frame(dst, v.X+c.Col*battle.TileSide, v.Y+c.Row*battle.TileSide,
			battle.TileSide, battle.TileSide, town.ColourSelected)
	}

	for _, box := range []struct{ X, Y, Width, Height int }{
		{battle.SheetBox.X, battle.SheetBox.Y, battle.SheetBox.Width, battle.SheetBox.Height},
		{battle.BodyBox.X, battle.BodyBox.Y, battle.BodyBox.Width, battle.BodyBox.Height},
	} {
		shadow(dst, box.X, box.Y, box.Width, box.Height)
		fill(dst, box.X, box.Y, box.Width, box.Height, town.ColourBoxEdge)
		fill(dst, box.X+1, box.Y+1, box.Width-2, box.Height-2, town.ColourBoxFill)
	}

	if err := body.Draw(dst, sprites, s); err != nil {
		return err
	}
	if err := sheet.Draw(dst, sprites, f, s); err != nil {
		return err
	}
	if err := drawBattleCommandPanel(dst, sprites, f, s); err != nil {
		return err
	}
	if err := drawBattleResult(dst, sprites, s); err != nil {
		return err
	}

	// 執行段：整條訊息區換成一張卷軸，圖示列被蓋掉（RE-089、RE-090）。底圖是
	// MISCPICS 的圖 1，296×32，白紙面配黃色的捲邊與軸心；文字寫在紙面上，
	// 欄 3（像素 24）、掃描列 0xA8 起、行高 6（RE-087）。
	// 逃走那一問（RE-103）與撿那一問（RE-113）走同一條路：原版問句也是印在
	// 執行段那張卷軸上，沒有另外開框。
	if s.Plan.Prompt == battle.PromptExecuting ||
		s.Plan.Prompt == battle.PromptFlee ||
		s.Plan.Prompt == battle.PromptPickUp ||
		s.Plan.Prompt == battle.PromptArea {
		box := battle.MessageBox
		scroll, err := sprites.Get(battle.MessageScroll.Archive,
			battle.MessageScroll.Index)
		if err != nil {
			return err
		}
		draw(dst, scroll, box.X, box.Y)
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

// drawBattleResult 在收尾時把戰鬥結果圖蓋在裝備欄、人形框與戰場視窗上
// （`sub_288D4`，RE-147）：貼圖、白框、投影。底下的訊息卷軸不受影響。
func drawBattleResult(dst *ebiten.Image, sprites Sprites, s *battle.State) error {
	return drawBattleResultScaled(dst, sprites, s, 1)
}

// drawBattleResultScaled 是同一張圖乘上倍率，中文模式用。
func drawBattleResultScaled(dst *ebiten.Image, sprites Sprites, s *battle.State, sc int) error {
	n := s.ResultPicture()
	if n == battle.ResultNone {
		return nil
	}
	pic, err := sprites.Get(battle.ResultArchive, n)
	if err != nil {
		return err
	}
	at := battle.ResultPictureAt
	drawScaled(dst, pic, at.X*sc, at.Y*sc, sc)
	fr := battle.ResultFrame
	frame(dst, fr.X*sc, fr.Y*sc, fr.Width*sc, fr.Height*sc, town.ColourBoxEdge)
	shadow(dst, fr.X*sc, fr.Y*sc, fr.Width*sc, fr.Height*sc)
	return nil
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

// BattleSpoils 是打贏之後那張戰利品清單（RE-096）。清單疊在戰場畫面上，
// 原版沒有把底下重畫掉。
//
// title 是標題那一筆記錄拆好的幾段：那一筆自己帶著 `~H02~V09` 兩個座標，
// 所以不能走會剝掉版面碼的 UIRaw（見 `town.UISegments`）。
func BattleSpoils(dst *ebiten.Image, sprites Sprites, tiles *BattleTiles,
	f *font.Font, s *battle.State, title []town.Segment, body *BattleBody, sheet *BattleSheet) error {
	if err := BattleField(dst, sprites, tiles, f, s, body, sheet); err != nil {
		return err
	}
	b := battle.SpoilBox
	shadow(dst, b.X, b.Y, b.Width, b.Height)
	fill(dst, b.X, b.Y, b.Width, b.Height, town.ColourBoxEdge)
	fill(dst, b.X+1, b.Y+1, b.Width-2, b.Height-2, town.ColourBoxFill)

	ink := town.ColourText
	for _, seg := range title {
		f.Draw(dst, seg.Col*8, seg.Row, strings.ToUpper(seg.Text),
			ink[0], ink[1], ink[2])
	}

	// 選中那一列是換文字顏色，不是框起來：原版的清單元件用 `sub_1B033` 把整列
	// 重畫一次，選中傳顏色 0x0F、其餘傳 3（RE-115）。與招募、商店那兩張清單同一
	// 套語意——它們本來就走同一支 `sub_1AAE8`。
	l := battle.SpoilList
	for i, line := range s.SpoilLines() {
		y := l.Row + i*l.LineStep
		c := town.ColourListIdle
		if i == s.Spoil.Cursor {
			c = town.ColourText
		}
		f.Draw(dst, l.Col*8, y, strings.ToUpper(line), c[0], c[1], c[2])
	}
	// 背包塞不下、手上那件被換下來的那一句（記錄 0x183）由訊息區顯示，
	// 這裡不重複。

	// 圖示列換成兩格：原版開清單之前先 `sub_1449C()` 把整條清掉，再放 YES 與
	// U-TURN 並把白框擺在第 0 格（RE-115）。
	return drawSpoilIcons(dst, sprites, s)
}

// drawSpoilIcons 把戰場底下那一條換成戰利品畫面的兩格圖示。
//
// 清掉的範圍是 `sub_1103F(2)` 的 (1, 0xA4)–(0x26, 0xC4)，也就是 x = 8…304、
// y = 164…196——**與執行段那張卷軸是同一塊**（`MessageBox` 是 (8, 0xA4) 296×32），
// 所以報完獎勵的那幾句在清單出現的同時就被蓋掉了。
func drawSpoilIcons(dst *ebiten.Image, sprites Sprites, s *battle.State) error {
	bg, err := sprites.Get(town.Background.Archive, town.Background.Index)
	if err != nil {
		return err
	}
	m := battle.MessageBox
	strip := image.Rect(m.X, m.Y, m.X+m.Width, m.Y+m.Height)
	draw(dst, bg.SubImage(strip).(*ebiten.Image), m.X, m.Y)
	b := town.IconBar
	for i, code := range battle.SpoilIcons {
		icon, err := sprites.Get("KOLMPICS", town.IconSprite(code))
		if err != nil {
			return err
		}
		x := b.X + i*b.Step
		draw(dst, icon, x, b.Y+1)
		c := town.ColourIdle
		if i == s.Spoil.Icon {
			c = town.ColourSelected
		}
		frame(dst, x, b.Y, b.Width, b.Height, c)
	}
	return nil
}

// BattleSpellList 是規劃段選法術那兩張清單（RE-100）：「Which Spell?」與
// 「Select Affect」。框與裝備欄同一個，清單直接蓋在上面。
//
// title 是標題那一筆記錄拆好的幾段（它自己帶 `~H03~V07`）；rows 是清單的內容。
func BattleSpellList(dst *ebiten.Image, sprites Sprites, tiles *BattleTiles,
	f *font.Font, s *battle.State, title []town.Segment, rows []string, body *BattleBody, sheet *BattleSheet) error {
	if err := BattleField(dst, sprites, tiles, f, s, body, sheet); err != nil {
		return err
	}
	b := battle.SpellListBox
	shadow(dst, b.X, b.Y, b.Width, b.Height)
	fill(dst, b.X, b.Y, b.Width, b.Height, town.ColourBoxEdge)
	fill(dst, b.X+1, b.Y+1, b.Width-2, b.Height-2, town.ColourBoxFill)

	ink := town.ColourText
	for _, seg := range title {
		f.Draw(dst, seg.Col*8, seg.Row, strings.ToUpper(seg.Text),
			ink[0], ink[1], ink[2])
	}
	l := battle.SpellListRows
	for i, row := range rows {
		if i >= l.Visible {
			break
		}
		y := l.Row + i*l.LineStep
		if i == s.Plan.ListPick {
			// **游標的樣子還沒對拍**，與戰利品那張清單同一個狀況：先用與指令列
			// 白框同一個顏色框起來。
			frame(dst, l.Col*8-2, y-1, b.Width-(l.Col*8-2-b.X)-4, l.LineStep,
				town.ColourSelected)
		}
		f.Draw(dst, l.Col*8, y, strings.ToUpper(row), ink[0], ink[1], ink[2])
	}
	return nil
}
