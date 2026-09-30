package town

import (
	"fmt"
	"strings"
)

// MIRROR 的角色資料頁（RE-074）。畫面上的每一個數字都由角色記錄算出來，
// 版面座標直接來自 `sub_15ABF` 與 `sub_16239` 推給繪圖 helper 的引數，
// 不是量出來的：`word_2DF14` 是字格欄（× 8 ＝ x），`word_2DF16` 是像素列。

// SheetIcons 是角色資料頁底下的圖示編號。原版固定推 U-TURN、16 與 6，
// 另外兩個看角色記錄有沒有東西（`sub_15ABF:0x1607E`）。
const (
	IconSheetItems  = 16
	IconSheetSpells = 18
	IconSheetNotes  = 19
)

// 屬性在角色記錄裡的位移與順序（0x1D–0x23）。體型不畫在頁面上。
const (
	StatStrength = iota
	StatQuickness
	StatSize
	StatHealth
	StatForesight
	StatCharisma
	StatIntellect
	StatCount
)

// StatScaleNum／StatScaleDen 是內部值換成畫面值的比例：`× 100 / 150`。
const (
	StatScaleNum = 100
	StatScaleDen = 150
)

// WeaponRow 是武器熟練度的一列。角色記錄裡是 0x4A + 13 × 列（RE-074）。
type WeaponRow struct {
	// Row 是這一列在角色記錄裡的編號 0…3。
	Row  int    `json:"row"`
	Name string `json:"name"`
	// Weapon 是武器識別（記錄位移 +8）；訓練時用它比對練過沒有（RE-076）。
	Weapon  int `json:"weapon"`
	Offence int `json:"offence"`
	Defence int `json:"defence"`
	// OffenceRaw／DefenceRaw 是沒有除以 10 的原始值，訓練加的是這一個。
	OffenceRaw int `json:"offence_raw"`
	DefenceRaw int `json:"defence_raw"`
}

// Sheet 是一名角色在資料頁上要用到的全部欄位。
type Sheet struct {
	Ward            int            `json:"ward"`
	Stats           [StatCount]int `json:"stats"`
	Race            int            `json:"race"`
	Profession      int            `json:"profession"`
	Rank            int            `json:"rank"`
	AdventurePoints int            `json:"adventure_points"`
	MovementRate    int            `json:"movement_rate"`
	Nutrition       int            `json:"nutrition"`
	HeightFeet      int            `json:"height_feet"`
	HeightInches    int            `json:"height_inches"`
	Weight          int            `json:"weight"`
	Weapons         []WeaponRow    `json:"weapons"`
	Order           string         `json:"order"`
	Title           string         `json:"title"`
	// HasItems／HasNotes 決定圖示列要不要多那兩格：原版看 `char[0x200]`
	// 與 `char[0x1DD]`…`char[0x1E1]` 有沒有東西。
	HasItems bool `json:"has_items"`
	HasNotes bool `json:"has_notes"`
}

// scaleStat 把內部值換成畫面值。
func scaleStat(v int) int { return v * StatScaleNum / StatScaleDen }

// Displayed 是頁面上那八個數字，順序與 KOLDATA 記錄 73h 的標籤相同。
// 後兩個是算出來的：平衡是力量與智力的平均、耐力是力量與健康的平均。
func (s *Sheet) Displayed() [8]int {
	str, hea, inl := s.Stats[StatStrength], s.Stats[StatHealth], s.Stats[StatIntellect]
	return [8]int{
		scaleStat(str),
		scaleStat(s.Stats[StatQuickness]),
		scaleStat(hea),
		scaleStat(s.Stats[StatForesight]),
		scaleStat(s.Stats[StatCharisma]),
		scaleStat(inl),
		scaleStat((inl + str) / 2),
		scaleStat((hea + str) / 2),
	}
}

// WeaponRows 是武器列的上限。
const WeaponRows = 4

// EarnedRank 是這名角色「應得」的位階：八個攻防值加起來除以 200
// （`sub_14371`，RE-076）。AWARD 的訓練要求記錄裡的 Rank 與它相同。
func (s *Sheet) EarnedRank() int {
	total := 0
	for _, w := range s.Weapons {
		total += w.OffenceRaw + w.DefenceRaw
	}
	return total / 200
}

// MovementLetter 是畫面上那個字母：`chr(0x41 + 移動率)`，移動率 0 當成 20。
func (s *Sheet) MovementLetter() string {
	rate := s.MovementRate
	if rate == 0 {
		rate = SheetDefaultMovement
	}
	return string(rune('A' + rate))
}

// SheetDefaultMovement 是移動率為 0 時代入的值（`sub_15ABF:0x15F04`）。
const SheetDefaultMovement = 20

// Height 是 `7' 4"` 這個字串，格式與原版的 `%1d'%2d"` 相同。
func (s *Sheet) Height() string {
	return fmt.Sprintf("%1d'%2d\"", s.HeightFeet, s.HeightInches)
}

// Weight 是 `296 LBS`，格式同原版的 `%3d LBS`。
func (s *Sheet) WeightText() string { return fmt.Sprintf("%3d LBS", s.Weight) }

// BarCells 是傷勢／營養長條的格數：值右移三位，非 0 但不足一格算一格。
func BarCells(v int) int {
	n := v >> 3
	if v != 0 && n == 0 {
		n = 1
	}
	return n
}

// SheetPage 是畫面上要畫的東西，已經把記錄號換成字串。
type SheetPage struct {
	Name           string      `json:"name"`
	Gender         string      `json:"gender"`
	Race           string      `json:"race"`
	Profession     string      `json:"profession"`
	RankLeft       string      `json:"rank_left"`
	RankRight      string      `json:"rank_right"`
	Height         string      `json:"height"`
	Weight         string      `json:"weight"`
	Stats          [8]int      `json:"stats"`
	Weapons        []WeaponRow `json:"weapons"`
	Gold           int         `json:"gold"`
	Adventure      int         `json:"adventure"`
	Movement       string      `json:"movement"`
	Order          string      `json:"order"`
	Title          string      `json:"title"`
	DamageCells    int         `json:"damage_cells"`
	NutritionCells int         `json:"nutrition_cells"`
	Portrait       int         `json:"portrait"`
	// Steed 非 nil 時右上角多一塊坐騎（RE-120）：原版看角色記錄 `+0x1B0` 是不是
	// 零，非零才畫圖與名牌（`sub_16239:0x1625e`）。
	Steed *SteedPic `json:"steed,omitempty"`
}

// KOLDATA 上與角色資料頁有關的記錄。
const (
	// GenderRecord 是性別名稱的起點：0 男 → 記錄 4，1 女 → 記錄 5。
	GenderRecord = 4
	// RaceRecord 是種族名稱的起點；角色記錄位移 0x24 直接就是記錄號（0…3）。
	RaceRecord = 0
	// ProfessionOffset 是職業記錄的偏移：`char[0x1C] + 6`，非人族與女性精靈再減 6。
	ProfessionOffset = 6
	// RankRecord 是位階名稱的起點。
	RankRecord = 0x74
	// SheetLabels 是整頁標籤的那一筆記錄，帶 ~V／~H 座標。
	SheetLabels = 0x73
)

// professionRecord 是職業名稱的 KOLDATA 記錄號：`char[0x1C] + 6` 再過 `sub_14307`
// 的換算（見 CanonicalRecord）。
func professionRecord(profession, race, gender int) int {
	return CanonicalRecord(profession+ProfessionOffset, race, gender)
}

// buildSheet 把一名隊員的資料頁組出來。記錄號在這裡就換成字串，
// 畫面層只管排版。
func (s *Session) buildSheet(m Member) *SheetPage {
	sh := m.Sheet
	if sh == nil {
		return nil
	}
	gender := 0
	if m.Female {
		gender = 1
	}
	page := &SheetPage{
		Name: strings.ToUpper(m.Name),
		// 抬頭整段是大寫：原版在畫名字之前把 `word_2DF54` 設成 1，
		// 之後那幾筆記錄都照這個旗標轉大寫，離開頁面才歸零（`sub_15ABF:0x1616E`）。
		Gender:     strings.ToUpper(s.uiRecord(GenderRecord + gender)),
		Race:       strings.ToUpper(s.uiRecord(RaceRecord + sh.Race)),
		Profession: strings.ToUpper(s.uiRecord(professionRecord(sh.Profession, sh.Race, gender))),
		RankLeft:   strings.ToUpper(s.uiRecord(RankRecord + sh.Rank/5*5)),
		Height:     sh.Height(),
		Weight:     sh.WeightText(),
		Stats:      sh.Displayed(),
		Weapons:    upperWeapons(sh.Weapons),
		Gold:       m.Gold,
		Adventure:  sh.AdventurePoints,
		Movement:   sh.MovementLetter(),
		// 教派與頭銜也在大寫旗標底下（原版畫面是 `WHITE PEARL`，RE-123 的對拍）。
		Order:          strings.ToUpper(sh.Order),
		Title:          strings.ToUpper(sh.Title),
		DamageCells:    BarCells(m.Wounds),
		NutritionCells: BarCells(sh.Nutrition),
		Portrait:       m.Portrait,
	}
	if m.Steed != nil {
		page.Steed = &SteedPic{Picture: m.Steed.Picture,
			Name: steedPlateName(m.Steed.Name), FullName: m.Steed.Name}
	}
	if sh.Rank%5 != 0 {
		page.RankRight = strings.ToUpper(s.uiRecord(RankRecord + sh.Rank))
	}
	return page
}

// steedPlateName 是資料頁名牌上那一格：大寫、只留前四個字元（RE-120）。
func steedPlateName(name string) string {
	up := []rune(strings.ToUpper(name))
	if len(up) > SheetSteed.NameChars {
		up = up[:SheetSteed.NameChars]
	}
	return strings.TrimRight(string(up), " ")
}

// sheetIcons 是資料頁的圖示列。後兩個要有東西才出現，本輪的原版存檔全部是空的，
// 因此只有前三個會畫出來（`sub_15ABF:0x1607E`）。
func sheetIcons(m Member) []int {
	icons := []int{IconUTurn, IconSheetItems, IconCompanion}
	// 法術那一格看角色記錄 `0x200` 有沒有東西，任務一覽那一格看 `0x1DD`…`0x1E1`
	// 有沒有東西（`sub_15ABF:0x1607E`）——remake 這一側就是會不會法術與
	// 已回報的任務數。`Sheet.HasItems`／`HasNotes` 是舊的旗標欄位，
	// typed data 沒有來源，所以由 Member 自己的兩塊資料算。
	if len(m.Spells) > 0 || (m.Sheet != nil && m.Sheet.HasItems) {
		icons = append(icons, IconSheetSpells)
	}
	if hasReports(m.Progress) || (m.Sheet != nil && m.Sheet.HasNotes) {
		icons = append(icons, IconSheetNotes)
	}
	return icons
}

// hasReports 回報那五格「已回報的任務數」有沒有東西。
func hasReports(progress []int) bool {
	for _, n := range progress {
		if n != 0 {
			return true
		}
	}
	return false
}

// OpenSheet 是按下 MIRROR：畫面換成發言者的角色資料頁。
func (s *Session) OpenSheet() {
	if s.Mode != ModeInside || s.Inside == nil {
		return
	}
	i := s.speakerIndex()
	if i < 0 {
		return
	}
	page := s.buildSheet(s.Party[i])
	if page == nil {
		s.note(iconName(IconMirror))
		return
	}
	s.Inside.Sheet = page
	s.Inside.SheetLabels = s.sheetLabels()
	s.setInsideIcons(sheetIcons(s.Party[i]))
}

// closeSheet 是資料頁上選 U-TURN：回到建築內的一般畫面。
func (s *Session) closeSheet() {
	if s.sheetHost != nil {
		s.closeSheetOutside()
		return
	}
	if s.Inside == nil || s.door == nil {
		return
	}
	s.Inside.Sheet = nil
	s.Inside.SheetLabels = nil
	s.Inside.QuestList = nil
	s.Inside.ItemPage = nil
	s.Inside.PartyPage = nil
	s.Inside.Dialog = s.greeting()
	s.Inside.DialogZh = nil
	s.setInsideIcons(InteriorIcons(s.door.IconFlags, len(s.Party)))
}

// ShowingSheet 回報現在是不是停在角色資料頁。
func (s *Session) ShowingSheet() bool { return s.Inside != nil && s.Inside.Sheet != nil }

// upperWeapons 把武器名轉大寫。資料裡存的是 `Greatsword`（5-bit 解碼器的 mode 1），
// 資料頁整段走大寫旗標，所以畫出來是 `GREATSWORD`（RE-074／RE-075）。
func upperWeapons(rows []WeaponRow) []WeaponRow {
	out := make([]WeaponRow, len(rows))
	for i, w := range rows {
		w.Name = strings.ToUpper(w.Name)
		out[i] = w
	}
	return out
}

// sheetLabels 是頁面上那些固定標籤，中文模式優先用譯文（譯文自己帶同一套座標）。
func (s *Session) sheetLabels() []Segment {
	if zh := s.Trans.UIRecord(SheetLabels); zh != "" {
		return ParseSegments(zh)
	}
	if s.UIText == nil {
		return nil
	}
	return s.UIText.Segments(BankUI, SheetLabels)
}

// sheetIconName 是資料頁上那幾個還沒實作的圖示名稱，只用在 Unimplemented 紀錄。
func sheetIconName(id int) string {
	switch id {
	case IconSheetItems:
		return "角色資料頁：物品"
	case IconCompanion:
		return "角色資料頁：換人"
	case IconSheetSpells:
		return "角色資料頁：法術"
	case IconSheetNotes:
		return "角色資料頁：筆記"
	}
	return iconName(id)
}
