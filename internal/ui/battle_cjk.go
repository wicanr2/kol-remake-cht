package ui

import (
	"fmt"
	"image"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/book"
	"github.com/wicanr2/kol-remake-cht/internal/font"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

// BattleFieldCJK 畫中文模式的戰場那一頁。
//
// 與 `BattleField` 用**同一份版面**，每一個座標乘上 `book.CJKScale`。三倍之後
// 訊息區的行高 6 變成 18，正好是中文字模格的高度，所以四行都放得下；戰場、
// 三個框與圖示列也都照原版的相對位置。
//
// 這一頁不對拍——原版沒有中文畫面。英文那一版（逐畫面對拍用）不動。
func BattleFieldCJK(dst *ebiten.Image, sprites Sprites, tiles *BattleTiles,
	f *font.Font, t Translator, s *battle.State, body *BattleBody) error {
	sc := book.CJKScale
	bg, err := sprites.Get(town.Background.Archive, town.Background.Index)
	if err != nil {
		return err
	}
	drawScaled(dst, bg, 0, 0, sc)

	w := battle.FieldWindow
	shadow(dst, w.X*sc, w.Y*sc, w.Width*sc, w.Height*sc)
	fill(dst, w.X*sc, w.Y*sc, w.Width*sc, w.Height*sc, town.ColourBoxEdge)

	cx, cy := s.Camera()
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
			drawScaled(dst, img, (v.X+c*battle.TileSide)*sc, (v.Y+r*battle.TileSide)*sc, sc)
		}
	}
	cur := battle.FieldCursor
	frame(dst, cur.X*sc, cur.Y*sc, cur.Width*sc, cur.Height*sc, town.ColourSelected)
	if s.Plan.Prompt == battle.PromptTarget {
		c := s.Plan.Cursor
		if c.Row >= 0 && c.Row < battle.ViewRows && c.Col >= 0 && c.Col < battle.ViewCols {
			frame(dst, (v.X+c.Col*battle.TileSide)*sc, (v.Y+c.Row*battle.TileSide)*sc,
				battle.TileSide*sc, battle.TileSide*sc, town.ColourSelected)
		}
	}

	for _, box := range [2]struct{ X, Y, Width, Height int }{
		{battle.SheetBox.X, battle.SheetBox.Y, battle.SheetBox.Width, battle.SheetBox.Height},
		{battle.BodyBox.X, battle.BodyBox.Y, battle.BodyBox.Width, battle.BodyBox.Height},
	} {
		x, y, bw, bh := box.X*sc, box.Y*sc, box.Width*sc, box.Height*sc
		shadow(dst, x, y, bw, bh)
		fill(dst, x, y, bw, bh, town.ColourBoxEdge)
		fill(dst, x+1, y+1, bw-2, bh-2, town.ColourBoxFill)
	}

	if err := body.DrawScaled(dst, sprites, s, sc); err != nil {
		return err
	}
	if err := drawBattleSheetCJK(dst, sprites, f, t, s, sc); err != nil {
		return err
	}
	if err := drawBattleCommandPanelCJK(dst, sprites, f, t, s, sc); err != nil {
		return err
	}

	// 執行段：整條訊息區換成一張卷軸，圖示列被蓋掉（RE-089、RE-090）。
	if s.Plan.Prompt == battle.PromptExecuting ||
		s.Plan.Prompt == battle.PromptFlee ||
		s.Plan.Prompt == battle.PromptPickUp ||
		s.Plan.Prompt == battle.PromptArea {
		box := battle.MessageBox
		scroll, err := sprites.Get(battle.MessageScroll.Archive, battle.MessageScroll.Index)
		if err != nil {
			return err
		}
		drawScaled(dst, scroll, box.X*sc, box.Y*sc, sc)
		area := battle.MessageArea
		for i, line := range s.Exec.Text {
			y := (area.Y + i*area.LineHeight) * sc
			// 中文不轉大寫——原版轉大寫是因為它的小字模只有大寫好看，
			// 中文沒有大小寫，轉了只會讓夾在句子裡的英文名字變形。
			cellInAt(dst, f, area.Col*town.CellWidth*sc, y, t.T(line), battle.MessageInk)
		}
		return nil
	}
	return drawBattleIconsCJK(dst, sprites, s, sc)
}

// BattleEncounterCJK 是中文模式的遭遇那一頁：同一份版面乘上倍率，訊息改用
// 中文字模。`lines` 已經是譯好的敘述（`quest.lines` 走 KOLDATA 的譯文），
// 所以這裡不再查表。
func BattleEncounterCJK(dst *ebiten.Image, sprites Sprites, f *font.Font,
	s *battle.State, lines []string) error {
	sc := book.CJKScale
	bg, err := sprites.Get(town.Background.Archive, town.Background.Index)
	if err != nil {
		return err
	}
	drawScaled(dst, bg, 0, 0, sc)

	c := battle.EncounterClear
	shadow(dst, c.X*sc, c.Y*sc, c.Width*sc, c.Height*sc)
	fill(dst, c.X*sc, c.Y*sc, c.Width*sc, c.Height*sc, battle.ColourFieldEdge)
	fill(dst, c.X*sc+1, c.Y*sc+1, c.Width*sc-2, c.Height*sc-2, town.ColourBoxFill)

	pic, err := sprites.Get(battle.MonsterArchive, s.Site.Monster)
	if err != nil {
		return err
	}
	for _, spot := range s.Spots() {
		drawScaled(dst, pic, spot[0]*sc, spot[1]*sc, sc)
	}

	box := battle.EncounterBox
	fill(dst, box.X*sc, box.Y*sc, box.Width*sc, box.Height*sc, town.ColourBoxFill)
	fr := battle.EncounterFrame
	frame(dst, fr.X*sc, fr.Y*sc, fr.Width*sc, fr.Height*sc, town.ColourBoxEdge)
	tx := battle.EncounterText
	for i, line := range lines {
		cellInAt(dst, f, tx.X*sc, (tx.Y+i*tx.Step)*sc, line, town.ColourText)
	}

	icon, err := sprites.Get("KOLMPICS", town.IconSprite(town.IconSpacebar))
	if err != nil {
		return err
	}
	b := town.IconBar
	drawScaled(dst, icon, b.X*sc, (b.Y+1)*sc, sc)
	frame(dst, b.X*sc, b.Y*sc, b.Width*sc, b.Height*sc, town.ColourSelected)
	return nil
}

// BattleSpoilsCJK 畫中文模式的戰利品清單。戰場底層與英文 renderer 共用同一份狀態，
// 但清單本身直接在 960×600 畫布上畫中文字，避免把原版小字模放大後再當成中文頁。
// 原版沒有中文畫面；這裡只保留已知的框、列距、選中顏色與圖示覆蓋範圍。
func BattleSpoilsCJK(dst *ebiten.Image, sprites Sprites, tiles *BattleTiles,
	f *font.Font, t Translator, s *battle.State, title []town.Segment, body *BattleBody) error {
	if err := BattleFieldCJK(dst, sprites, tiles, f, t, s, body); err != nil {
		return err
	}
	sc := book.CJKScale
	b := battle.SpoilBox
	drawBattleBoxCJK(dst, b.X, b.Y, b.Width, b.Height, sc)
	drawBattleTitleCJK(dst, f, title, b.X, b.Width, sc)

	l := battle.SpoilList
	textX := l.Col * town.CellWidth
	safe := battleCJKListSafeWidth(b.X, b.Width, textX, sc)
	for i, line := range s.SpoilLines() {
		if s.Spoil.Phase == battle.SpoilPickItem {
			line = displayName(t, line)
		}
		y := l.Row + i*l.LineStep
		ink := town.ColourListIdle
		if i == s.Spoil.Cursor {
			ink = town.ColourText
		}
		drawBattleListTextCJK(dst, f, textX*sc, y*sc, line, safe, ink)
	}
	return drawSpoilIconsCJK(dst, sprites, s, sc)
}

// BattleSpellListCJK 畫中文模式的法術／正負效果清單。rows 已由 cmd/kol 依狀態機
// 選好；法術名保持資料原文，正負效果文字則沿用已翻譯的 KOLDATA。
func BattleSpellListCJK(dst *ebiten.Image, sprites Sprites, tiles *BattleTiles,
	f *font.Font, t Translator, s *battle.State, title []town.Segment,
	rows []string, body *BattleBody) error {
	if err := BattleFieldCJK(dst, sprites, tiles, f, t, s, body); err != nil {
		return err
	}
	sc := book.CJKScale
	b := battle.SpellListBox
	drawBattleBoxCJK(dst, b.X, b.Y, b.Width, b.Height, sc)
	drawBattleTitleCJK(dst, f, title, b.X, b.Width, sc)

	l := battle.SpellListRows
	textX := l.Col * town.CellWidth
	safe := battleCJKListSafeWidth(b.X, b.Width, textX, sc)
	for i, row := range rows {
		if i >= l.Visible {
			break
		}
		y := l.Row + i*l.LineStep
		if i == s.Plan.ListPick {
			// 英文版目前以白框表示重製版游標；原版外觀尚未完全對拍，中文只等比
			// 保留這個既有視覺契約。
			frame(dst, (textX-2)*sc, (y-1)*sc,
				(b.Width-(textX-2-b.X)-4)*sc, l.LineStep*sc,
				town.ColourSelected)
		}
		drawBattleListTextCJK(dst, f, textX*sc, y*sc, row, safe, town.ColourText)
	}
	return nil
}

// drawBattleBoxCJK 與英文戰鬥框使用相同的投影、外框與內填色，只把座標乘上倍率。
func drawBattleBoxCJK(dst *ebiten.Image, x, y, width, height, scale int) {
	x, y, width, height = x*scale, y*scale, width*scale, height*scale
	shadow(dst, x, y, width, height)
	fill(dst, x, y, width, height, town.ColourBoxEdge)
	fill(dst, x+1, y+1, width-2, height-2, town.ColourBoxFill)
}

// battleCJKListSafeWidth 給固定列的文字安全矩形：框右側保留四個原版像素，
// 與英文法術游標框及戰利品內緣一致。
func battleCJKListSafeWidth(boxX, boxWidth, textX, scale int) int {
	return (boxX + boxWidth - 4 - textX) * scale
}

func drawBattleTitleCJK(dst *ebiten.Image, f *font.Font, title []town.Segment,
	boxX, boxWidth, scale int) {
	for _, seg := range title {
		x := seg.Col * town.CellWidth * scale
		y := seg.Row * scale
		safe := battleCJKListSafeWidth(boxX, boxWidth, seg.Col*town.CellWidth, scale)
		drawBattleListTextCJK(dst, f, x, y, seg.Text, safe, town.ColourText)
	}
}

func drawBattleListTextCJK(dst *ebiten.Image, f *font.Font,
	x, y int, text string, safe int, ink [3]uint8) {
	cellInAt(dst, f, x, y, clip(f, text, safe), ink)
}

// drawSpoilIconsCJK 清掉報獎勵卷軸所在的同一塊底圖，再放回 YES／U-TURN 圖示。
func drawSpoilIconsCJK(dst *ebiten.Image, sprites Sprites, s *battle.State, scale int) error {
	bg, err := sprites.Get(town.Background.Archive, town.Background.Index)
	if err != nil {
		return err
	}
	m := battle.MessageBox
	strip := image.Rect(m.X, m.Y, m.X+m.Width, m.Y+m.Height)
	drawScaled(dst, bg.SubImage(strip).(*ebiten.Image), m.X*scale, m.Y*scale, scale)

	b := town.IconBar
	for i, code := range battle.SpoilIcons {
		icon, err := sprites.Get("KOLMPICS", town.IconSprite(code))
		if err != nil {
			return err
		}
		x := (b.X + i*b.Step) * scale
		drawScaled(dst, icon, x, (b.Y+1)*scale, scale)
		ink := town.ColourIdle
		if i == s.Spoil.Icon {
			ink = town.ColourSelected
		}
		frame(dst, x, b.Y*scale, b.Width*scale, b.Height*scale, ink)
	}
	return nil
}

// drawBattleSheetCJK 是左上那一格的裝備欄。六列標籤走譯文，值照原文印
// ——那是物品名，不在譯文表裡。
func drawBattleSheetCJK(dst *ebiten.Image, sprites Sprites, f *font.Font,
	t Translator, s *battle.State, sc int) error {
	d, show := s.EquipmentPanel()
	if !show {
		return nil
	}
	// 英文版把整塊畫進 119×73 的離屏再貼上去，座標因此帶著 −17／−3 的位移。
	// 這裡直接算絕對位置：離屏的原點是 (17, 3)，所以 x、y 就是原版的座標。
	at := func(x, y int, value string, up bool) {
		if value == "" {
			return
		}
		if up {
			value = strings.ToUpper(value)
		}
		cellInAt(dst, f, x*sc, y*sc, value, [3]uint8{255, 255, 255})
	}
	if !d.Monster {
		at(24, 7, d.Name, true)
	}
	if d.Equipment {
		for i := 0; i < d.Count; i++ {
			at(24, 24+6*i, t.T(battle.EquipmentLabels[i]), false)
			at(48, 24+6*i, displayName(t, d.Rows[i]), false)
		}
		if d.AmmunitionLabel != "" {
			at(24, 60, fmt.Sprintf("%s%d", t.T(d.AmmunitionLabel), d.Ammunition), false)
		}
		return nil
	}
	archive, x, y := "CHARPICS", 56, 37
	if d.Monster {
		archive = "MONPICS"
	}
	pic, err := sprites.Get(archive, d.Portrait)
	if err != nil {
		return err
	}
	if pic == nil {
		return fmt.Errorf("%s 肖像 %d 缺少圖片", archive, d.Portrait)
	}
	if d.Monster {
		pw, ph := pic.Bounds().Dx(), pic.Bounds().Dy()
		x, y = (3+(15-pw/8)/2)*8, 76-ph
	}
	drawScaled(dst, pic, x*sc, y*sc, sc)
	return nil
}

// drawBattleCommandPanelCJK 是左下那一列命令圖示與它們的標籤。
func drawBattleCommandPanelCJK(dst *ebiten.Image, sprites Sprites, f *font.Font,
	t Translator, s *battle.State, sc int) error {
	for _, item := range s.CommandPanel() {
		pic, err := sprites.Get("KOLMPICS", item.Code+battle.IconOffset)
		if err != nil {
			return err
		}
		if pic == nil || pic.Bounds().Dx() != 24 || pic.Bounds().Dy() != 20 {
			return fmt.Errorf("命令圖示 %d 缺少 24×20 圖片", item.Code)
		}
		// 標籤在圖示上方一行。中文字高 18 而原版那一段只留 8×3 = 24 px，
		// 剛好放得下。
		cellInAt(dst, f, item.X*sc, (item.Y-8)*sc, t.T(item.Label), [3]uint8{255, 255, 255})
		drawScaled(dst, pic, item.X*sc, item.Y*sc, sc)
	}
	return nil
}

// drawBattleIconsCJK 是底下那一條圖示列，與英文版同一組位置乘上倍率。
func drawBattleIconsCJK(dst *ebiten.Image, sprites Sprites, s *battle.State, sc int) error {
	b := town.IconBar
	for i, code := range s.Plan.Bar {
		icon, err := sprites.Get("KOLMPICS", town.IconSprite(code))
		if err != nil {
			return err
		}
		x := (b.X + i*b.Step) * sc
		drawScaled(dst, icon, x, (b.Y+1)*sc, sc)
		c := town.ColourIdle
		if i == s.Plan.Selected {
			c = town.ColourSelected
		}
		frame(dst, x, b.Y*sc, b.Width*sc, b.Height*sc, c)
	}
	return nil
}

var _ = ebiten.NewImage
