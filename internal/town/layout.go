// Package town 保存城鎮畫面的版面與狀態。
//
// 版面數字全部由原版畫面機械量得：底圖由 tools/decompose_town_frame.py 逐張相減選出
// 並要求分群外殘餘為零；圖示、肖像與頭像的圖號由逐像素搜尋命中；文字座標由
// tools/read_text_from_frame.py 以抽出的原版字型逐格讀回。收據見 RE-054 與
// docs/verification/town-ui/。
package town

// Background 是城鎮畫面的底圖。`BOOKPICS[3]` 與 `KOLMPICS[0]` 是同一張紅磚圖，
// 書本與城鎮各存一份；城鎮走 KOLMPICS。
var Background = struct {
	Archive string
	Index   int
}{"KOLMPICS", 0}

// Shadow 是城鎮視窗共用的投影：往右下各偏移 4 px、厚 4 px 的黑塊。
// 肖像、標題框與對話框量到的都是同一組數字。
var Shadow = struct{ Offset, Thickness int }{4, 4}

// PartySlot 是畫面頂端的隊員格：上格是 40×38 的角色圖像，下格是名字前五個字元。
// 兩格共用一個外框；發言者的外框是白色，其餘隊員是亮藍色。
var PartySlot = struct {
	X, Y, Step, Width       int
	PortraitHeight          int
	PlateHeight             int
	PortraitX, PortraitY    int
	NameX, NameY, NameChars int
}{
	X: 14, Y: 1, Step: 48, Width: 43,
	PortraitHeight: 40, PlateHeight: 9,
	PortraitX: 16, PortraitY: 2,
	NameX: 16, NameY: 42, NameChars: 5,
}

// Portrait 是對話對象的肖像位置。素材本身已含外框，畫面上只多一層投影。
var Portrait = struct{ X, Y int }{8, 64}

// Box 是城鎮的文字視窗：白框黑底，右下有投影，內文以 8 px 字格排列。
type Box struct {
	X, Y, Width, Height int
	TextX, TextY        int
	LineStep, Cells     int
	Centred             bool
}

// TitleBox 是地點與對話對象的名稱框，兩行置中；DialogBox 是台詞框，靠左。
// 置中是以字格數計算的：起始格 ＝ (Cells − len) / 2，向下取整。
var (
	TitleBox  = Box{X: 102, Y: 58, Width: 195, Height: 20, TextX: 104, TextY: 60, LineStep: 8, Cells: 24, Centred: true}
	DialogBox = Box{X: 102, Y: 84, Width: 195, Height: 68, TextX: 104, TextY: 86, LineStep: 8, Cells: 24}
)

// IconBar 是畫面底部的圖示列。框內 22×20 是圖示的第 1..22 欄——整張 24×20 圖示畫在
// (X+Step*i, Y+1)，外框再蓋掉左右各一欄。
var IconBar = struct{ X, Y, Step, Width, Height, Slots int }{
	X: 32, Y: 170, Step: 32, Width: 24, Height: 22, Slots: 6,
}

// InnIcons 是旅店圖示列由左而右的 KOLMPICS 圖號，六格都逐像素命中。
var InnIcons = []int{5, 20, 7, 6, 8, 9}

// InnIconNames 是手冊給這六個圖示的名稱，順序與 InnIcons 相同。
var InnIconNames = []string{"DOOR", "MIRROR", "LISTEN", "TALK", "REST", "COMPANION"}

// 城鎮畫面的顏色語意：白框是發言者／目前選定的圖示，亮藍框是其餘隊員與未選圖示。
var (
	ColourSelected = [3]uint8{255, 255, 255}
	ColourIdle     = [3]uint8{85, 85, 255}
	ColourBoxEdge  = [3]uint8{255, 255, 255}
	ColourBoxFill  = [3]uint8{0, 0, 0}
	ColourText     = [3]uint8{255, 255, 255}
	ColourShadow   = [3]uint8{0, 0, 0}
)

// CJKInn 是旅店畫面的中文模式版面（放大後的像素）。與其他中文版面一樣，這一份是
// remake 自訂的，不對拍：位置是英文版面的兩倍，只有文字改用 16×18 的中文字格，
// 因此行距與框內留白要重排。
var CJKInn = struct {
	PartyX, PartyY, PartyStep, PartyWidth       int
	PortraitBoxHeight, PlateHeight              int
	PortraitX, PortraitY, NameX, NameY          int
	PicX, PicY                                  int
	TitleX, TitleY, TitleW, TitleH              int
	TitleTextY, TitleStep                       int
	DialogX, DialogY, DialogW, DialogH          int
	DialogTextX, DialogTextY, DialogStep        int
	DialogCells                                 int
	IconX, IconY, IconStep, IconW, IconH        int
	Scale                                       int
}{
	PartyX: 28, PartyY: 2, PartyStep: 96, PartyWidth: 86,
	PortraitBoxHeight: 80, PlateHeight: 18,
	PortraitX: 32, PortraitY: 4, NameX: 32, NameY: 84,
	PicX: 16, PicY: 128,
	TitleX: 204, TitleY: 116, TitleW: 390, TitleH: 40,
	TitleTextY: 117, TitleStep: 19,
	DialogX: 204, DialogY: 168, DialogW: 390, DialogH: 136,
	DialogTextX: 210, DialogTextY: 172, DialogStep: 22,
	DialogCells: 12,
	IconX: 64, IconY: 340, IconStep: 64, IconW: 48, IconH: 44,
	Scale: 2,
}
