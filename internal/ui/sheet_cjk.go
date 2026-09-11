package ui

import (
	"fmt"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/kol-remake-cht/internal/font"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

// drawSheetCJK 畫中文模式的角色資料頁（SPEC-030）。
//
// 與 `drawSheet` 畫的是**同一份 `SheetPage`**——欄位、計算與資料來源一個都沒變，
// 差別只在版面：中文字模格高 18 而原版行距只有 6…8，等比放大會溢出，所以垂直重排，
// 抬頭從六列併成四列。這一頁不對拍，原版沒有中文畫面。
//
// `labels` 是 `Session.sheetLabels()` 解出來的那十四段，**中文譯文已經在記錄
// 0x73 裡**（`docs/i18n/ui-zh-Hant.json` 的 KOLDATA 115），所以標籤不另外查表：
// 這裡只取它的文字，座標改用本規格的那一組。順序與原版記錄相同：
// 八個屬性、武器表頭、金冠／冒險點數／移動率、傷勢、營養。
func drawSheetCJK(dst *ebiten.Image, f *font.Font, labels []town.Segment, p *town.SheetPage) {
	l := town.CJKSheet

	box(dst, l.Header, town.ColourSheetHeader)
	box(dst, l.Stats, town.ColourSheetFrame)
	box(dst, l.Weapons, town.ColourSheetFrame)
	box(dst, l.Gold, town.ColourSheetFrame)
	box(dst, l.Bars, town.ColourSheetFrame)

	// 兩條長條：標籤在同一列的左側，條從標籤右緣起。
	barCJK(dst, l.DamageBar, p.DamageCells)
	barCJK(dst, l.NutritionBar, p.NutritionCells)
	for i := range l.BarRows {
		cellCJK(dst, f, l.BarLabelCol[i], l.BarRows[i], label(labels, sheetLabelBars+i))
	}

	// 抬頭四列。名字置中，其餘照三欄排。
	safe := l.Header.Width() - 2*f.Advance()
	centredCJK(dst, f, l.Header, l.NameRow, clip(f, p.Name, safe))
	third := safe / 3
	for i, s := range [3]string{p.Gender, p.Race, p.Profession} {
		cellCJK(dst, f, l.HeaderCols[i], l.IdentityRow, clip(f, s, third))
	}
	cellCJK(dst, f, l.HeaderCols[0], l.RankRow, clip(f, p.RankLeft, third))
	if p.RankRight != "" {
		cellCJK(dst, f, l.HeaderCols[1], l.RankRow, clip(f, p.RankRight, third))
	}
	cellCJK(dst, f, l.HeaderCols[0], l.BodyRow, p.Height)
	cellCJK(dst, f, l.HeaderCols[1], l.BodyRow, p.Weight)

	// 八個屬性：標籤靠左、值靠右。標籤走譯文，取不到就留空——不回填英文，
	// 中文模式混排英文標籤比缺一格更難讀。
	statSafe := (l.StatValueCol - l.StatLabelCol) * f.Advance()
	for i, v := range p.Stats {
		row := l.StatFirstRow + i*l.LineStep
		cellCJK(dst, f, l.StatLabelCol, row, clip(f, label(labels, i), statSafe))
		cellCJK(dst, f, l.StatValueCol, row, fmt.Sprintf("%3d", v))
	}

	// 武器熟練度：表頭線在框頂下方，資料列往下排。
	// 表頭那一段是「武器技能　攻擊　防禦」一整串，原版靠串裡的空白把後兩個字
	// 推到英文版面的欄位上。中文版面的欄位自己定位，所以照空白拆開分別擺，
	// 才會與底下的數值對齊。
	header := strings.Fields(labelRaw(labels, sheetLabelWeaponHeader))
	for i, col := range [3]int{l.WeaponNameCol, l.WeaponOffCol, l.WeaponDefCol} {
		if i < len(header) {
			cellCJK(dst, f, col, l.HeaderRow, header[i])
		}
	}
	fill(dst, l.Weapons.X0+1, l.RuleRow, l.Weapons.Width()-2, 2, town.ColourSheetFrame)
	nameSafe := (l.WeaponOffCol - l.WeaponNameCol) * f.Advance()
	for i, w := range p.Weapons {
		if i >= l.WeaponRows {
			break
		}
		row := l.WeaponFirstRow + i*l.LineStep
		// 武器名是原版資料裡的物品名，不在譯文表的記錄裡——與英文版一樣照原文印。
		cellCJK(dst, f, l.WeaponNameCol, row, clip(f, w.Name, nameSafe))
		cellCJK(dst, f, l.WeaponOffCol, row, fmt.Sprintf("%3d", w.Offence))
		cellCJK(dst, f, l.WeaponDefCol, row, fmt.Sprintf("%3d", w.Defence))
	}

	// 金幣那一框的四列。數值不截斷——那是遊戲數字，截掉會讀成另一個值。
	for i, row := range []struct {
		y     int
		value string
	}{
		{l.GoldRow, fmt.Sprintf("% 9d", p.Gold)},
		{l.AdventureRow, fmt.Sprintf("%5d", p.Adventure)},
		{l.MovementRow, p.Movement},
	} {
		labelSafe := (l.ValueCol - l.LabelCol) * f.Advance()
		cellCJK(dst, f, l.LabelCol, row.y, clip(f, label(labels, sheetLabelGold+i), labelSafe))
		cellCJK(dst, f, l.ValueCol, row.y, row.value)
	}
	if p.Order != "" {
		cellCJK(dst, f, l.OrderCol, l.EquipRow, p.Order)
	}
	if p.Title != "" {
		cellCJK(dst, f, l.TitleCol-len([]rune(p.Title)), l.EquipRow, p.Title)
	}
}

// cellCJK 與 `cell` 的差別只有一處：清底的寬度用 `Measure` 逐字量，
// 不用「字數 × 固定字寬」。中文字模表裡半形與全形並存，固定字寬會清錯範圍。
func cellCJK(dst *ebiten.Image, f *font.Font, col, row int, s string) {
	if s == "" {
		return
	}
	x := col * f.Advance()
	w, h := f.Measure(s)
	fill(dst, x, row, w, h, town.ColourBoxFill)
	f.Draw(dst, x, row, s,
		town.ColourText[0], town.ColourText[1], town.ColourText[2])
}

// centredCJK 把一段文字放在某個框的水平中線上。
func centredCJK(dst *ebiten.Image, f *font.Font, r town.Rect, row int, s string) {
	if s == "" {
		return
	}
	w, _ := f.Measure(s)
	x := r.X0 + (r.Width()-w)/2
	cellCJK(dst, f, x/f.Advance(), row, s)
}

// barCJK 與 `bar` 同樣是「青底鋪滿、紅色從左緣蓋上 n 格」，格寬與內縮改用 CJK 那一組。
func barCJK(dst *ebiten.Image, r town.Rect, cells int) {
	l := town.CJKSheet
	fill(dst, r.X0, r.Y0, r.Width(), r.Height(), town.ColourSheetFrame)
	if cells <= 0 {
		return
	}
	w := cells*l.BarCell + 1
	if max := r.Width(); w > max {
		w = max
	}
	fill(dst, r.X0, r.Y0+l.BarInset, w, r.Height()-2*l.BarInset, town.ColourSheetBar)
}

// clip 把一段文字截到安全寬度以內。**逐字截斷，不換行也不縮字**（SPEC-030）：
// 這一頁每一格都有固定位置，換行會推擠下一列，縮字會讓同一頁出現兩種字級。
func clip(f *font.Font, s string, safe int) string {
	if s == "" || safe <= 0 {
		return s
	}
	if w, _ := f.Measure(s); w <= safe {
		return s
	}
	runes := []rune(s)
	for n := len(runes) - 1; n > 0; n-- {
		if w, _ := f.Measure(string(runes[:n])); w <= safe {
			return string(runes[:n])
		}
	}
	return ""
}

// 記錄 0x73 那十四段的索引：0…7 是八個屬性，8 是武器表頭，
// 9…11 是金冠／冒險點數／移動率，12…13 是傷勢與營養。
const (
	sheetLabelWeaponHeader = 8
	sheetLabelGold         = 9
	sheetLabelBars         = 12
)

// label 取第 n 段的文字，並去掉全形空白。
//
// 譯文那一串的全形空白是**替英文版面對齊用的**（「力量　：」讓冒號落在原版
// 那一欄）；中文版面的欄位自己定位，留著只會把冒號推開一段。表頭那一段例外，
// 見 `labelRaw`——它的空白是欄與欄的間隔，去掉就擠在一起了。
//
// 段數不足就回空字串——**不回填英文**：中文模式混排英文標籤比缺一格更難讀，
// 而缺格在畫面上看得出來，混排看起來像是故意的。
func label(labels []town.Segment, n int) string {
	return strings.ReplaceAll(labelRaw(labels, n), "\u3000", "")
}

// labelRaw 取第 n 段的原文，空白照留。
func labelRaw(labels []town.Segment, n int) string {
	if n < 0 || n >= len(labels) {
		return ""
	}
	return labels[n].Text
}
