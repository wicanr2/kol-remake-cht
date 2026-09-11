package ui

import (
	"fmt"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/kol-remake-cht/internal/book"
	"github.com/wicanr2/kol-remake-cht/internal/font"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

// drawSheetCJK 畫中文模式的角色資料頁（SPEC-030）。
//
// 與 `drawSheet` 畫的是**同一份 `SheetPage`**，而且用的是**同一份版面**
// （`town.SheetLayout`）——每一個座標乘上 `book.CJKScale` 就是這一頁的位置。
// 三倍畫布下原版最小行距 6 剛好變成 18，正是中文字模格的高度，所以放得下；
// 兩倍時只有 12，那才是要重排的原因。
//
// 水平方向原版一格 8 px，三倍之後 24 px，而中文字寬 16——字在格子裡靠左，
// 欄與欄的相對位置與原版一致。
//
// `labels` 是 `Session.sheetLabels()` 解出來的那十四段，**中文譯文已經在記錄
// 0x73 裡**（`docs/i18n/ui-zh-Hant.json` 的 KOLDATA 115），所以標籤不另外查表：
// 這裡只取它的文字。順序與原版記錄相同：八個屬性、武器表頭、
// 金冠／冒險點數／移動率、傷勢、營養。
//
// 這一頁不對拍，原版沒有中文畫面。
func drawSheetCJK(dst *ebiten.Image, f *font.Font, labels []town.Segment, p *town.SheetPage) {
	l := town.SheetLayout
	s := book.CJKScale

	boxAt(dst, scaleRect(l.Header, s), town.ColourSheetHeader)
	rule(dst, scaleRect(l.HeaderRule, s), town.ColourSheetHeader)
	boxAt(dst, scaleRect(l.Stats, s), town.ColourSheetFrame)
	boxAt(dst, scaleRect(l.Weapons, s), town.ColourSheetFrame)
	rule(dst, scaleRect(l.WeaponRule, s), town.ColourSheetFrame)
	boxAt(dst, scaleRect(l.Gold, s), town.ColourSheetFrame)
	boxAt(dst, scaleRect(l.Bars, s), town.ColourSheetFrame)

	barCJK(dst, scaleRect(l.DamageBar, s), p.DamageCells, l.BarCell*s, l.BarInset*s)
	barCJK(dst, scaleRect(l.NutritionBar, s), p.NutritionCells, l.BarCell*s, l.BarInset*s)

	// 抬頭六列照原版排。中文字比原版字窄，置中與靠右都用實際量到的寬度算。
	head := scaleRect(l.Header, s)
	centredCJK(dst, f, head, l.NameRow*s, p.Name)
	centredCJK(dst, f, head, l.GenderRow*s, p.Gender)
	cellCJK(dst, f, s, l.LeftCol, l.RaceRow, p.Race)
	rightCJK(dst, f, s, l.RightCol, l.RaceRow, p.Profession)
	cellCJK(dst, f, s, l.LeftCol, l.RankRow, p.RankLeft)
	if p.RankRight != "" {
		rightCJK(dst, f, s, l.RightCol, l.RankRow, p.RankRight)
	}
	cellCJK(dst, f, s, l.HeightCol, l.HeightRow, p.Height)
	cellCJK(dst, f, s, l.WeightCol, l.WeightRow, p.Weight)

	// 八個屬性：標籤靠左、值在原版那一欄。
	statSafe := (l.StatValueCol - 2) * town.CellWidth * s
	for i, v := range p.Stats {
		row := (town.SheetStatFirstRow + i*town.LineStep) * s
		cellCJKAt(dst, f, 2*town.CellWidth*s, row, clip(f, label(labels, i), statSafe))
		cellCJKAt(dst, f, l.StatValueCol*town.CellWidth*s, row, fmt.Sprintf("%3d", v))
	}

	// 武器表頭那一段是「武器技能　攻擊　防禦」一整串，原版靠串裡的空白把後兩個字
	// 推到欄位上。中文字比原版窄，照空白拆開分別擺才對得齊底下的數值。
	header := strings.Fields(labelRaw(labels, sheetLabelWeaponHeader))
	for i, col := range [3]int{l.WeaponNameCol, l.WeaponOffCol, l.WeaponDefCol} {
		if i < len(header) {
			cellCJK(dst, f, s, col, town.SheetWeaponHeaderRow, header[i])
		}
	}
	nameSafe := (l.WeaponOffCol - l.WeaponNameCol) * town.CellWidth * s
	for i, w := range p.Weapons {
		if i >= l.WeaponRows {
			break
		}
		row := l.WeaponFirstRow + i*l.WeaponRowStep
		// 武器名是原版資料裡的物品名，不在譯文表的記錄裡——與英文版一樣照原文印。
		cellCJK(dst, f, s, l.WeaponNameCol, row, clip(f, w.Name, nameSafe))
		cellCJK(dst, f, s, l.WeaponOffCol, row, fmt.Sprintf("%3d", w.Offence))
		cellCJK(dst, f, s, l.WeaponDefCol, row, fmt.Sprintf("%3d", w.Defence))
	}

	// 金幣那一框的三列。數值不截斷——那是遊戲數字，截掉會讀成另一個值。
	labelSafe := (l.GoldValueCol - l.WeaponNameCol) * town.CellWidth * s
	for i, row := range [3]struct {
		y     int
		col   int
		value string
	}{
		{l.GoldRow, l.GoldValueCol, fmt.Sprintf("% 9d", p.Gold)},
		{l.AdventureRow, l.AdventureCol, fmt.Sprintf("%5d", p.Adventure)},
		{l.MovementRow, l.MovementCol, p.Movement},
	} {
		cellCJK(dst, f, s, l.WeaponNameCol, row.y, clip(f, label(labels, sheetLabelGold+i), labelSafe))
		cellCJK(dst, f, s, row.col, row.y, row.value)
	}
	if p.Order != "" {
		cellCJK(dst, f, s, l.OrderCol, l.EquipRow, p.Order)
	}
	if p.Title != "" {
		rightCJK(dst, f, s, l.TitleCol, l.EquipRow, p.Title)
	}

	// 兩條長條的標籤在原版記錄裡就自帶位置（`~V84~H05`／`~V92`），照那一組放。
	for i := range [2]int{} {
		seg := labelSeg(labels, sheetLabelBars+i)
		cellCJKAt(dst, f, seg.Col*town.CellWidth*s, seg.Row*s, label(labels, sheetLabelBars+i))
	}
}

// scaleRect 把英文版面的矩形放大到中文畫布。
func scaleRect(r town.Rect, s int) town.Rect {
	return town.Rect{X0: r.X0 * s, Y0: r.Y0 * s, X1: r.X1*s + s - 1, Y1: r.Y1*s + s - 1}
}

func boxAt(dst *ebiten.Image, r town.Rect, c [3]uint8) {
	fill(dst, r.X0+1, r.Y0+1, r.Width()-2, r.Height()-2, town.ColourBoxFill)
	frame(dst, r.X0, r.Y0, r.Width(), r.Height(), c)
	shadow(dst, r.X0, r.Y0, r.Width(), r.Height())
}

func rule(dst *ebiten.Image, r town.Rect, c [3]uint8) {
	fill(dst, r.X0, r.Y0, r.Width(), r.Height(), c)
}

// cellCJK 在原版字格座標畫一段文字，位置乘上倍率。
func cellCJK(dst *ebiten.Image, f *font.Font, s, col, row int, text string) {
	cellCJKAt(dst, f, col*town.CellWidth*s, row*s, text)
}

// cellCJKAt 直接以像素座標畫。清底的寬度用 `Measure` 逐字量，不用「字數 × 固定
// 字寬」——中文字模表裡半形與全形並存，固定字寬會清錯範圍。
func cellCJKAt(dst *ebiten.Image, f *font.Font, x, y int, text string) {
	if text == "" {
		return
	}
	w, h := f.Measure(text)
	fill(dst, x, y, w, h, town.ColourBoxFill)
	f.Draw(dst, x, y, text,
		town.ColourText[0], town.ColourText[1], town.ColourText[2])
}

// centredCJK 把一段文字放在某個框的水平中線上。
func centredCJK(dst *ebiten.Image, f *font.Font, r town.Rect, y int, text string) {
	if text == "" {
		return
	}
	w, _ := f.Measure(text)
	cellCJKAt(dst, f, r.X0+(r.Width()-w)/2, y, text)
}

// rightCJK 把一段文字的右緣靠到某一欄。
func rightCJK(dst *ebiten.Image, f *font.Font, s, col, row int, text string) {
	if text == "" {
		return
	}
	w, _ := f.Measure(text)
	cellCJKAt(dst, f, col*town.CellWidth*s-w, row*s, text)
}

// barCJK 與 `bar` 同樣是「青底鋪滿、紅色從左緣蓋上 n 格」，格寬與內縮跟著倍率走。
func barCJK(dst *ebiten.Image, r town.Rect, cells, cell, inset int) {
	fill(dst, r.X0, r.Y0, r.Width(), r.Height(), town.ColourSheetFrame)
	if cells <= 0 {
		return
	}
	w := cells*cell + 1
	if w > r.Width() {
		w = r.Width()
	}
	fill(dst, r.X0, r.Y0+inset, w, r.Height()-2*inset, town.ColourSheetBar)
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
// 譯文那一串的全形空白是**替原版字寬對齊用的**（「力量　：」讓冒號落在原版那一
// 欄）；中文字比原版窄，留著只會把冒號推開一段。表頭那一段例外，見 `labelRaw`
// ——它的空白是欄與欄的間隔。
//
// 段數不足就回空字串——**不回填英文**：中文模式混排英文標籤比缺一格更難讀，
// 而缺格在畫面上看得出來，混排看起來像是故意的。
func label(labels []town.Segment, n int) string {
	return strings.ReplaceAll(labelRaw(labels, n), "　", "")
}

func labelRaw(labels []town.Segment, n int) string { return labelSeg(labels, n).Text }

func labelSeg(labels []town.Segment, n int) town.Segment {
	if n < 0 || n >= len(labels) {
		return town.Segment{}
	}
	return labels[n]
}

// clip 把一段文字截到安全寬度以內。**逐字截斷，不換行也不縮字**（SPEC-030）：
// 這一頁每一格都有固定位置，換行會推擠下一列，縮字會讓同一頁出現兩種字級。
func clip(f *font.Font, text string, safe int) string {
	if text == "" || safe <= 0 {
		return text
	}
	if w, _ := f.Measure(text); w <= safe {
		return text
	}
	runes := []rune(text)
	for n := len(runes) - 1; n > 0; n-- {
		if w, _ := f.Measure(string(runes[:n])); w <= safe {
			return string(runes[:n])
		}
	}
	return ""
}
