package town

import "strings"

// 資料頁上的法術一覽（`sub_166CC`，RE-123）。原版先重畫抬頭那一層（`sub_16239`：
// 背景、隊員格、坐騎、抬頭框——屬性以下那幾框不畫），再把角色記錄 `0x200` 起
// 那十六格排成兩欄各八列：每一格畫一個白框，框裡印 5-bit 解出來的法術名。**遇到空的一格就整個停下來**——原版把兩層迴圈的計數
// 都推到界限（`0x16797`），不是跳過那一格繼續掃。圖示列縮成只有 U-TURN。

// 一格的版面（`0x16723`…`0x16791`）：
//
//	框   sub_1BA9D(0xC0×欄 + 0x0E, 0x0D×列 + 0x3A,
//	              0xC0×欄 + 0x70, 0x0D×列 + 0x42, 0x0F)
//	文字 word_2DF14 = 0x18×欄 + 2、word_2DF16 = 0x0D×列 + 0x3C
const (
	// SpellCellCols／SpellCellRows 是兩欄各八列，十六格對應十六個法術槽。
	SpellCellCols = 2
	SpellCellRows = 8
	// SpellCellColStep 是欄與欄的像素距離，SpellCellRowStep 是列距。
	SpellCellColStep = 0xC0
	SpellCellRowStep = 0x0D
	// 框的四個邊（相對於那一格的原點）。
	SpellCellX0, SpellCellY0 = 0x0E, 0x3A
	SpellCellX1, SpellCellY1 = 0x70, 0x42
	// 文字：欄是 byte 欄（× 8 才是像素），列是掃描線。
	SpellCellTextColStep = 0x18
	SpellCellTextCol     = 2
	SpellCellTextRow     = 0x3C
)

// SpellCell 是一格：框的矩形與名字的字格座標都算好了。
type SpellCell struct {
	Name             string
	FrameX, FrameY   int
	FrameW, FrameH   int
	TextCol, TextRow int
}

// SpellPage 是蓋在資料頁上的法術一覽。
type SpellPage struct {
	Cells []SpellCell
}

// SpellCellAt 依槽號算出那一格的版面。槽 0…7 是左欄，8…15 是右欄。
func SpellCellAt(slot int, name string) SpellCell {
	col, row := slot/SpellCellRows, slot%SpellCellRows
	x0 := SpellCellColStep*col + SpellCellX0
	y0 := SpellCellRowStep*row + SpellCellY0
	x1 := SpellCellColStep*col + SpellCellX1
	y1 := SpellCellRowStep*row + SpellCellY1
	return SpellCell{
		Name:   name,
		FrameX: x0, FrameY: y0,
		FrameW: x1 - x0 + 1, FrameH: y1 - y0 + 1,
		TextCol: SpellCellTextColStep*col + SpellCellTextCol,
		TextRow: SpellCellRowStep*row + SpellCellTextRow,
	}
}

// SpellCells 把一名角色的法術排成那十六格。空的一格之後全部不畫。
func SpellCells(spells []Spell) []SpellCell {
	limit := SpellCellCols * SpellCellRows
	if len(spells) < limit {
		limit = len(spells)
	}
	out := make([]SpellCell, 0, limit)
	for i := 0; i < limit; i++ {
		if spells[i].Name == "" {
			break
		}
		// 法術一覽畫在 `sub_16239` 之後，那一支在 `0x163FB` 把大寫旗標
		// `word_2DF54` 設成 1，所以名字印成大寫（原版畫面是 `DAYNALON`）。
		out = append(out, SpellCellAt(i, strings.ToUpper(spells[i].Name)))
	}
	return out
}

// ShowingSpellList 回報現在是不是停在法術一覽。
func (s *Session) ShowingSpellList() bool {
	return s.Inside != nil && s.Inside.SpellList != nil
}

// OpenSpellList 是資料頁上按下法術那一格（圖示 0x12）。
func (s *Session) OpenSpellList() {
	if s.Inside == nil || !s.ShowingSheet() {
		return
	}
	i := s.speakerIndex()
	if i < 0 {
		return
	}
	cells := SpellCells(s.Party[i].Spells)
	if len(cells) == 0 {
		return
	}
	s.Inside.SpellList = &SpellPage{Cells: cells}
	// 原版把圖示列清成只有 U-TURN，白框停在第 0 格（`0x167bb`）。
	s.setInsideIcons([]int{IconUTurn})
}

// closeSpellList 收掉法術一覽，圖示列回到資料頁那一列。
func (s *Session) closeSpellList() {
	if s.Inside == nil {
		return
	}
	s.Inside.SpellList = nil
	i := s.speakerIndex()
	if i >= 0 {
		s.setInsideIcons(sheetIcons(s.Party[i]))
	}
}
