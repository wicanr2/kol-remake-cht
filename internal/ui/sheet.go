package ui

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/kol-remake-cht/internal/font"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

// drawSheet 畫 MIRROR 的角色資料頁（RE-074）。畫的順序與原版相同：
// 先五個框與兩條青色長條，再把標籤那一筆記錄按 `~V`／`~H` 座標放上去，
// 最後才是各欄的數值——長條上的紅色也在標籤之前，因為原版是先畫框再畫字。
func drawSheet(dst *ebiten.Image, sprites Sprites, f *font.Font,
	labels []town.Segment, p *town.SheetPage) error {
	l := town.SheetLayout

	box(dst, l.Header, town.ColourSheetHeader)
	fill(dst, l.HeaderRule.X0, l.HeaderRule.Y0,
		l.HeaderRule.Width(), l.HeaderRule.Height(), town.ColourSheetHeader)
	box(dst, l.Stats, town.ColourSheetFrame)
	box(dst, l.Weapons, town.ColourSheetFrame)
	fill(dst, l.WeaponRule.X0, l.WeaponRule.Y0,
		l.WeaponRule.Width(), l.WeaponRule.Height(), town.ColourSheetFrame)
	box(dst, l.Gold, town.ColourSheetFrame)
	box(dst, l.Bars, town.ColourSheetFrame)

	bar(dst, l.DamageBar, p.DamageCells)
	bar(dst, l.NutritionBar, p.NutritionCells)

	for _, seg := range labels {
		cell(dst, f, seg.Col, seg.Row, seg.Text)
	}

	// 抬頭：名字與性別置中，種族與位階靠左、職業與副位階靠右。
	centred(dst, f, l.NameRow, p.Name)
	centred(dst, f, l.GenderRow, p.Gender)
	cell(dst, f, l.LeftCol, l.RaceRow, p.Race)
	right(dst, f, l.RaceRow, p.Profession)
	cell(dst, f, l.LeftCol, l.RankRow, p.RankLeft)
	if p.RankRight != "" {
		right(dst, f, l.RankRow, p.RankRight)
	}
	cell(dst, f, l.HeightCol, l.HeightRow, p.Height)
	cell(dst, f, l.WeightCol, l.WeightRow, p.Weight)

	// 八個屬性：與標籤同一列，值靠第 12 欄，格式 %3d。
	for i, v := range p.Stats {
		row := town.SheetStatFirstRow + i*town.LineStep
		cell(dst, f, l.StatValueCol, row, fmt.Sprintf("%3d", v))
	}

	// 武器熟練度：只畫有東西的那幾列，每一列往下 6 px。
	for i, w := range p.Weapons {
		if i >= l.WeaponRows {
			break
		}
		row := l.WeaponFirstRow + i*l.WeaponRowStep
		cell(dst, f, l.WeaponNameCol, row, w.Name)
		cell(dst, f, l.WeaponOffCol, row, fmt.Sprintf("%3d", w.Offence))
		cell(dst, f, l.WeaponDefCol, row, fmt.Sprintf("%3d", w.Defence))
	}

	cell(dst, f, l.GoldValueCol, l.GoldRow, fmt.Sprintf("% 9d", p.Gold))
	cell(dst, f, l.AdventureCol, l.AdventureRow, fmt.Sprintf("%5d", p.Adventure))
	cell(dst, f, l.MovementCol, l.MovementRow, p.Movement)
	if p.Armour != "" {
		cell(dst, f, l.ArmourCol, l.EquipRow, p.Armour)
	}
	if p.Shield != "" {
		cell(dst, f, l.ShieldCol-len([]rune(p.Shield)), l.EquipRow, p.Shield)
	}
	// 有坐騎才有右上角那一塊（RE-120）。
	if p.Steed != nil {
		return drawSheetSteed(dst, sprites, f, p.Steed)
	}
	return nil
}

// drawSheetSteed 畫資料頁右上角那一匹：先清黑，再放 40×38 的圖，
// 上下兩個白框共用一道投影，名字印在下面那格（RE-120）。
func drawSheetSteed(dst *ebiten.Image, sprites Sprites, f *font.Font,
	s *town.SteedPic) error {
	l := town.SheetSteed
	pic, err := sprites.Get("STEEDS", s.Picture)
	if err != nil {
		return err
	}
	fill(dst, l.Fill.X0, l.Fill.Y0, l.Fill.Width(), l.Fill.Height(),
		town.ColourBoxFill)
	draw(dst, pic, l.X, l.Y)
	frame(dst, l.Frame.X0, l.Frame.Y0, l.Frame.Width(), l.Frame.Height(),
		town.ColourBoxEdge)
	frame(dst, l.Plate.X0, l.Plate.Y0, l.Plate.Width(), l.Plate.Height(),
		town.ColourBoxEdge)
	shadow(dst, l.Frame.X0, l.Frame.Y0, l.Frame.Width(),
		l.Plate.Y1-l.Frame.Y0+1)
	cell(dst, f, l.NameCol, l.NameRow, s.Name)
	return nil
}

// box 畫一個框：原版 `sub_1BA9D` 先把內部填黑、再畫外框，最後補右下投影
// （`sub_1B7E9`），與城鎮的文字視窗同一組數字。
func box(dst *ebiten.Image, r town.Rect, c [3]uint8) {
	fill(dst, r.X0+1, r.Y0+1, r.Width()-2, r.Height()-2, town.ColourBoxFill)
	frame(dst, r.X0, r.Y0, r.Width(), r.Height(), c)
	shadow(dst, r.X0, r.Y0, r.Width(), r.Height())
}

// bar 畫一條長條：先鋪青色底，再從左緣蓋上 n 格紅色。
func bar(dst *ebiten.Image, r town.Rect, cells int) {
	l := town.SheetLayout
	fill(dst, r.X0, r.Y0, r.Width(), r.Height(), town.ColourSheetFrame)
	if cells <= 0 {
		return
	}
	// 原版的紅色矩形是含端點的 `(39, y0+1)`…`(39 + 8n, y1-1)`，所以寬度是 8n + 1。
	fill(dst, r.X0, r.Y0+l.BarInset, cells*l.BarCell+1, r.Height()-2*l.BarInset,
		town.ColourSheetBar)
}

// cell 在字格座標畫一段文字：欄 × 8 是 x，列本身就是 y。
//
// 原版的繪字是**不透明**的：每一個字格連同空白都先清成底色。金幣那一行看得出來
// ——`GOLD CROWNS  :` 的冒號在第 29 欄，而 `% 9ld` 的值也從第 29 欄開始，
// 第一個字元是空白，把冒號蓋掉了。原版畫面上那個冒號因此不見，另外兩行都還在。
func cell(dst *ebiten.Image, f *font.Font, col, row int, s string) {
	if s == "" {
		return
	}
	x := col * f.Advance()
	fill(dst, x, row, len([]rune(s))*f.Advance(), f.Height(), town.ColourBoxFill)
	f.Draw(dst, x, row, s,
		town.ColourText[0], town.ColourText[1], town.ColourText[2])
}

// centred 把一段文字放在抬頭的 22 格內置中。
func centred(dst *ebiten.Image, f *font.Font, row int, s string) {
	l := town.SheetLayout
	cell(dst, f, (l.CentreCells-len([]rune(s)))/2+l.LeftCol, row, s)
}

// right 把一段文字靠齊抬頭右緣。
func right(dst *ebiten.Image, f *font.Font, row int, s string) {
	l := town.SheetLayout
	cell(dst, f, l.RightCol-len([]rune(s)), row, s)
}
