package town

import "fmt"

// 法術編輯器那一頁的版面（`sub_18350` 邊算邊畫，RE-128）。x 是 byte 欄、y 是像素。
const (
	// SpellEditPicture 是背景圖（`sub_1874E`：`sub_1103F(0x1D)`）。
	SpellEditPicture = 0x1D
	// SpellTitleRecord 是標題那一筆，自己帶座標：CLASS／SUBCLASS／STAT／
	// RANGE DUR DMG 與兩行造價。
	SpellTitleRecord = 0x4D
	// SpellSaveRecord 是「Record changes in Spellbook?」。
	SpellSaveRecord = 0x4A
	// SpellDeniedRecord…SpellDeniedOrderRecord 是教派不准這一類時的兩列。
	SpellDeniedRecord      = 0x4B
	SpellDeniedOrderRecord = 0x4C
	// SpellRangeRecord 是 RANGE 那一欄的兩個選項：Short／Long。
	SpellRangeRecord = 0x50
	// SpellNoDurationRecord 是 `----`：屬性不吃持續時那一欄印的東西。
	SpellNoDurationRecord = 0x52
)

// 十三個欄位畫在哪。名字那一欄是置中的，所以只記那一列的 y。
const (
	spellNameRow  = 0x29
	spellNameCols = 0x0C
	spellNameLeft = 0x0E
	spellGuardRow = 0x40
	spellGuardGap = 8
	spellClassCol = 0x07
	spellSubCol   = 0x15
	spellLowerRow = 0x72
	spellStatCol  = 0x07
	spellRangeCol = 0x12
	spellTurnCol  = 0x18
	spellPowerCol = 0x1D
	spellCostCol  = 0x1D
	spellCostRow  = 0x82
	spellGoldRow  = 0x8A
)

// 欄位的兩種顏色：游標停著的那一欄是 3，其餘 0x0F（`0x18377`…`0x18395`）。
const (
	SpellFieldColor  = 0x0F
	SpellCursorColor = 3
)

// SpellEditPage 是法術編輯器畫出來的一頁。
type SpellEditPage struct {
	// Picture 是背景圖號；Labels 是標題那一筆拆出來的帶座標文字。
	Picture int       `json:"picture"`
	Labels  []Segment `json:"labels,omitempty"`
	// Rules 是六條分隔線（x1、y、x2 都是像素）。
	Rules []SpellRule `json:"rules,omitempty"`
	// Fields 是十三個欄位，順序就是欄位編號。
	Fields []SpellFieldCell `json:"fields"`
	// Cursor 是游標停在第幾欄。
	Cursor int `json:"cursor"`
	// Adventure／Gold 是算出來的兩種造價。**畫面上那兩行的標籤是反的**：
	// 原版在 `Adventure Point Cost..` 那一行印的是金冠（`0x18702`），
	// 在 `Gold Crown Cost.......` 那一行印的是冒險點數（`0x18736`），
	// 而扣款與「不夠」那兩句都照真正的資源走。remake 照原樣畫。
	Adventure int `json:"adventure"`
	Gold      int `json:"gold"`
	// Menu 非 nil 時某一欄的選單開著，蓋在這一頁上。
	Menu *InnList `json:"menu,omitempty"`
	// Typing 非 nil 時停在名稱輸入；內容是目前打了什麼。
	Typing *string `json:"typing,omitempty"`
}

// SpellRule 是一條分隔線。
type SpellRule struct {
	X1 int `json:"x1"`
	Y  int `json:"y"`
	X2 int `json:"x2"`
}

// SpellFieldCell 是一個欄位畫出來的樣子。
type SpellFieldCell struct {
	Col   int    `json:"col"`
	Row   int    `json:"row"`
	Text  string `json:"text"`
	Color int    `json:"color"`
}

// spellRules 是 `sub_1874E` 畫的六條線：標題底下兩段、下半部四段。
var spellRules = []SpellRule{
	{0x38, 0x3E, 0xA0}, {0xA8, 0x3E, 0x108},
	{0x38, 0x70, 0x88}, {0x90, 0x70, 0xB8},
	{0xC0, 0x70, 0xE0}, {0xE8, 0x70, 0x108},
}

// spellFieldAt 是第 field 欄畫在哪一格。
func spellFieldAt(field int) (col, row int) {
	switch {
	case field == SpellFieldName:
		return spellNameLeft, spellNameRow
	case field <= SpellFieldClassLast:
		return spellClassCol, spellGuardRow + spellGuardGap*(field-SpellFieldClassFirst)
	case field <= SpellFieldSubLast:
		return spellSubCol, spellGuardRow + spellGuardGap*(field-SpellFieldSubFirst)
	case field == SpellFieldStat:
		return spellStatCol, spellLowerRow
	case field == SpellFieldTargetFlag:
		return spellRangeCol, spellLowerRow
	case field == SpellFieldTurn:
		return spellTurnCol, spellLowerRow
	}
	return spellPowerCol, spellLowerRow
}

// spellEditPage 把工作副本畫成一頁。
func (s *Session) spellEditPage() *SpellEditPage {
	e := s.scroll.edit
	page := &SpellEditPage{
		Picture: SpellEditPicture,
		Labels:  s.spellTitleLabels(),
		Rules:   spellRules,
		Cursor:  e.field,
		Fields:  make([]SpellFieldCell, SpellFieldCount),
	}
	for field := 0; field < SpellFieldCount; field++ {
		col, row := spellFieldAt(field)
		cell := SpellFieldCell{Col: col, Row: row, Color: SpellFieldColor,
			Text: s.spellCellText(e.draft, field)}
		if field == SpellFieldName {
			// 名字置中：原版把 12 欄減掉字數再折半（`0x183C7`）。
			cell.Col = (spellNameCols-len([]rune(cell.Text)))/2 + spellNameLeft
		}
		if field == e.field {
			cell.Color = SpellCursorColor
		}
		page.Fields[field] = cell
	}
	page.Adventure, page.Gold = SpellCost(s.Magic, e.draft)
	page.Menu = e.menuView()
	page.Typing = e.typing
	return page
}

// menuView 是某一欄開著的選單；沒開就是 nil。捲動的規則與其他清單相同。
func (e *spellEditState) menuView() *InnList {
	if e.options == nil {
		return nil
	}
	end := e.top + ListRows
	if end > len(e.options) {
		end = len(e.options)
	}
	return &InnList{Rows: append([]string(nil), e.options[e.top:end]...),
		Selected: e.cursor - e.top, Total: len(e.options)}
}

// spellCellText 是一個欄位畫出來的字。名稱、RANGE 與「不吃持續」那三種字不在
// `MAGIC.0` 裡，由 KOLDATA 補。
func (s *Session) spellCellText(sp Spell, field int) string {
	switch field {
	case SpellFieldName:
		return sp.Name
	case SpellFieldTargetFlag:
		return s.uiRecord(SpellRangeRecord + SpellFieldValue(s.Magic, sp, field))
	case SpellFieldTurn:
		if spellNoTurns(sp) {
			return s.uiRecord(SpellNoDurationRecord)
		}
	}
	return SpellFieldText(s.Magic, sp, field)
}

// spellCostLines 是最後兩列的數字，給收據看的。上面那一行印的是金冠、
// 下面那一行印的是冒險點數——標籤與內容對調是原版自己的接法。
func (s *Session) spellCostLines(page *SpellEditPage) []Segment {
	return []Segment{
		{Col: spellCostCol, Row: spellCostRow, Text: fmt.Sprint(page.Gold)},
		{Col: spellCostCol, Row: spellGoldRow, Text: fmt.Sprint(page.Adventure)},
	}
}

// spellTitleLabels 是標題那一筆，中文模式優先用譯文——版面座標寫在記錄裡，
// 兩邊走同一套語法。
func (s *Session) spellTitleLabels() []Segment {
	if zh := s.Trans.UIRecord(SpellTitleRecord); zh != "" {
		return ParseSegments(zh)
	}
	if s.UIText == nil {
		return nil
	}
	return s.UIText.Segments(BankUI, SpellTitleRecord)
}
