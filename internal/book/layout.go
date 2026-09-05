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

// CharacterOptionsHeading 是角色選項頁的標題圖與位置。圖號由 tools/identify_frame_assets.py
// 在原版 accept-and-save-run1/frame-044 上逐像素命中。
var CharacterOptionsHeading = struct {
	ImageIndex int
	X, Y       int
}{ImageIndex: 8, X: 32, Y: 16}

// CharacterOptionsItems 是角色選項頁左半頁的選單項目。文字由
// tools/read_text_from_frame.py 以抽出的原版字型從 frame-044 逐格讀回。
var CharacterOptionsItems = []string{
	"LIST CHARACTERS",
	"CREATE CHARACTER",
	"DELETE CHARACTER",
	"EDIT PICTURE",
	"EDIT SHIELD",
	"HARBOR MASTER",
}

// CharacterOptionsList 是選單的起點與行距，與右頁清單同一組字模格規則。
var CharacterOptionsList = struct{ X, Y, Step int }{X: 24, Y: 52, Step: 6}

// RosterHeadings 是角色清單頁的兩個標題圖與位置。圖號由
// tools/identify_frame_assets.py 在 list-characters-run1/frame-065 上逐像素命中。
var RosterHeadings = struct {
	NamesIndex, NamesX, NamesY          int
	LocationIndex, LocationX, LocationY int
}{NamesIndex: 12, NamesX: 32, NamesY: 16,
	LocationIndex: 9, LocationX: 176, LocationY: 16}

// RosterNames 與 RosterLocations 是清單兩欄的起點與行距，與書中其他清單同一套規則。
var (
	RosterNames     = struct{ X, Y, Step int }{X: 24, Y: 52, Step: 6}
	RosterLocations = struct{ X, Y, Step int }{X: 176, Y: 52, Step: 6}
)

// SelectHeading 是右頁「Select one」標題圖與其位置。
var SelectHeading = struct {
	ImageIndex int
	X, Y       int
}{ImageIndex: 11, X: 176, Y: 16}

// 右頁清單：起點與行距。
var RightList = struct{ X, Y, Step int }{X: 176, Y: 52, Step: 6}

// PictureInstructions 是圖像層右頁的說明文字，排在右頁清單的同一個起點與行距。
// 內容由 tools/read_text_from_frame.py 以抽出的原版字型從
// picture-cycle-run1/frame-084 逐格讀回，不是由人眼抄的。
var PictureInstructions = []string{
	"USE ARROW KEYS",
	"TO SELECT A",
	"CHARACTER",
	"PRESS RETURN",
	"TO ACCEPT",
}

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

// PictureActions 是圖像層右頁底部的操作提示，棕色。字串與座標由
// tools/read_text_from_frame.py 從 picture-cycle-run1/frame-084 逐格讀回。
var PictureActions = Label{"CANCEL   ACCEPT", 176, 166, 170, 85, 0}

// PortraitFrame 是圖像層右頁的雙格外框：上格放 40×38 的角色圖像，下格放名字。
// 框線為 EGA 青，兩格內部填黑。座標量自 picture-cycle-run1/frame-084：
// 框線在 x=206 與 x=248、y=105／144／152。
var PortraitFrame = struct{ X, Y, Width, Height, Divider int }{
	X: 206, Y: 105, Width: 43, Height: 48, Divider: 144,
}

// PortraitName 是名牌文字起點。白字直接畫在黑底上，不另清字模格；字型與書頁內文相同。
var PortraitName = struct{ X, Y int }{208, 146}

// PortraitArrows 是圖像層右側的上下箭頭。兩個字模是 8×8 大字型的 codepoint 94 與 95
// （`^` 與 `_`），畫法與書頁文字相同：先把字模格清成白底再畫黑墨。這也是大字型第一個
// 被指認的畫面用途。
var PortraitArrows = struct {
	X, Up, Down      int
	UpRune, DownRune rune
}{X: 264, Up: 105, Down: 144, UpRune: '^', DownRune: '_'}

// NameCursor 是名字輸入游標：原版在下一個字模格的底列畫一條 8 px 寬的黑線。
// 由原版首次進入 New Character 頁的畫面量得（空名字時位於 x=24、y=63）。
var NameCursor = struct{ X, Y, Width, Height int }{24, 63, 8, 1}

// ColourPrompt 是確認視窗文字的 EGA 暗紅，量測自原版 delete-character-run1/frame-071。
var ColourPrompt = [3]uint8{170, 0, 0}

// ConfirmWindow 是刪除確認視窗的外框與投影。座標量自原版
// delete-character-run1/frame-071 與 delete-character-confirm-run1/frame-073：
// 兩次執行、清單長度不同（14 與 13 列），視窗位置完全相同。
//
// 外框是 132×24 的黑框白底；投影是往右下各偏移 4 px、厚 4 px 的黑塊，因此右側黑塊
// 佔 x 226..229、y 62..85，底部黑塊佔 y 82..85、x 98..229。
var ConfirmWindow = struct {
	X, Y, Width, Height, ShadowOffset, ShadowThickness int
}{X: 94, Y: 58, Width: 132, Height: 24, ShadowOffset: 4, ShadowThickness: 4}

// ConfirmLines 是視窗內的兩行文字。字串由 tools/read_text_from_frame.py 以抽出的原版
// 字型逐格讀回，座標是命中的字模格起點，不是人眼估的。
var ConfirmLines = []Label{
	{"ARE YOU SURE?", 104, 64, 170, 0, 0},
	{"YES/NO", 128, 70, 170, 0, 0},
}

// HarborMasterLines 是 HARBOR MASTER 的整頁訊息。這台機器上只安裝一個 region，
// 原版因此只畫這三行；第三行前空一列。文字與座標由 tools/read_text_from_frame.py
// 從 harbor-master-enter-run1/frame-090 以抽出的原版字型逐格讀回。
//
// 多個 region 時的畫面沒有量測過——那需要另一份原版磁碟。
var HarborMasterLines = []Label{
	{"ONLY ONE REGION", 24, 52, 0, 0, 0},
	{"IS INSTALLED.", 24, 58, 0, 0, 0},
	{"PRESS ANY KEY.", 24, 70, 0, 0, 0},
}
