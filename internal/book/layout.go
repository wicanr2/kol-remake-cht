package book

// PageInk 是書頁上文字格的底色。原版畫字時會把字模格清成白底，因此書頁原有的黃色
// 斑點不會透出來；量測自原版空白 New Character 畫面的標籤框。
var PageInk = struct{ R, G, B uint8 }{0xFF, 0xFF, 0xFF}

// NewCharacterHeading 是 New Character 頁的標題圖與其位置。圖號與座標由
// tools/locate_asset_in_frame.py 在原版畫面搜出，差異為 0。
var NewCharacterHeading = struct {
	ImageIndex int
	X, Y       int
}{ImageIndex: 10, X: 32, Y: 16}

// Label 是原版畫在書頁上的一段文字：座標與 EGA 顏色都由原版畫面機械量得。
type Label struct {
	Text    string
	X, Y    int
	R, G, B uint8
}

// NewCharacterLabels 是 New Character 頁左半頁的靜態標籤。
//
// 座標與顏色出自 docs/re/text-layout.json，由 tools/extract_text_layout.py 以抽出的
// 原版字型在 docs/verification/character-ui/escape-layers-run1 的空白 New Character
// 畫面上定位，每個字串都要求唯一命中。這裡不放任何推測值。
var NewCharacterLabels = []Label{
	{"NAME", 24, 52, 170, 85, 0},
	{"RACE", 24, 68, 0, 170, 170},
	{"SEX", 24, 84, 0, 170, 170},
	{"CLASS", 24, 100, 0, 170, 170},
	{"STRENGTH", 24, 118, 0, 170, 170},
	{"QUICKNESS", 24, 124, 0, 170, 170},
	{"SIZE", 24, 130, 0, 170, 170},
	{"HEALTH", 24, 136, 0, 170, 170},
	{"FORESIGHT", 24, 142, 0, 170, 170},
	{"CHARISMA", 24, 148, 0, 170, 170},
	{"INTELLECT", 24, 154, 0, 170, 170},
	{"BALANCE", 24, 160, 0, 170, 170},
	{"ENDURANCE", 24, 166, 0, 170, 170},
	{"BODY PTS.", 24, 172, 0, 170, 170},
}

// NewCharacterSeparators 是每列的冒號。原版把冒號放在固定字元欄而不是緊接標籤：
// 名字／種族／性別／職業列在字元欄 5（x=64），屬性列在字元欄 10（x=104）。
// 顏色跟隨該列標籤，量測自原版空白 New Character 畫面。
var NewCharacterSeparators = []Label{
	{":", 64, 52, 170, 85, 0},
	{":", 64, 68, 0, 170, 170},
	{":", 64, 84, 0, 170, 170},
	{":", 64, 100, 0, 170, 170},
	{":", 104, 118, 0, 170, 170},
	{":", 104, 124, 0, 170, 170},
	{":", 104, 130, 0, 170, 170},
	{":", 104, 136, 0, 170, 170},
	{":", 104, 142, 0, 170, 170},
	{":", 104, 148, 0, 170, 170},
	{":", 104, 154, 0, 170, 170},
	{":", 104, 160, 0, 170, 170},
	{":", 104, 166, 0, 170, 170},
	{":", 104, 172, 0, 170, 170},
}

// 以下座標與顏色同樣由原版畫面機械量得（tools/read_text_from_frame.py、
// tools/locate_asset_in_frame.py），記於 docs/re/text-layout.json 與 RE-043。

// SelectHeading 是右頁「Select one」標題圖與其位置。
var SelectHeading = struct {
	ImageIndex int
	X, Y       int
}{ImageIndex: 11, X: 176, Y: 16}

// 右頁清單：起點與行距。
var RightList = struct{ X, Y, Step int }{X: 176, Y: 52, Step: 6}

// 左頁的欄位值位置。名字與職業的值另起一行，種族與性別的值和標籤同一行。
var (
	NameValue      = struct{ X, Y int }{24, 58}
	RaceValue      = struct{ X, Y int }{72, 68}
	SecondaryValue = struct{ X, Y int }{72, 84}
	ClassValue     = struct{ X, Y int }{24, 106}
)

// 屬性數值靠右對齊於 x=112 起的三格，行距與標籤相同。
var AttributeValue = struct{ X, Y, Step, Cells int }{112, 118, 6, 3}

// 顏色語意：作用中的欄位標籤為棕色，其餘為青色；已定值為黑色；
// 選單反白項為黑色、未選為青色；職業層在左頁的即時預覽為亮紅色。
var (
	ColourActive   = [3]uint8{170, 85, 0}
	ColourInactive = [3]uint8{0, 170, 170}
	ColourValue    = [3]uint8{0, 0, 0}
	ColourPreview  = [3]uint8{255, 85, 85}
)

// LabelRow 是四個欄位標籤在 NewCharacterLabels 內的索引。
const (
	RowName = iota
	RowRace
	RowSecondary
	RowClass
)

// RowFirstAttribute 是十列屬性標籤在 NewCharacterLabels 內的起始索引。擲屬性層會把
// 這十列同時標成作用中，量測自原版 picture-cycle-run1/frame-082（1668 px 由青轉棕）。
const RowFirstAttribute = 4

// NameCursor 是名字輸入游標：原版在下一個字模格的底列畫一條 8 px 寬的黑線。
// 由原版首次進入 New Character 頁的畫面量得（空名字時位於 x=24、y=63）。
var NameCursor = struct{ X, Y, Width, Height int }{24, 63, 8, 1}
