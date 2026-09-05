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

// StablePic 是馬廄明細頁那一塊（RE-120）。進這一頁先呼叫 `sub_1452D`，它把對話框
// 內部清成黑並把游標設回 (0x0D, 0x56)，之後三段都從那裡推：
//
//	圖   `sub_1C53D(段, 0x0E, 0x56, 0)`      → (0x0E × 8, 0x56) = (112, 86)
//	名字 游標 +(7, 0x10)                     → 欄 0x14、列 0x66
//	問句 游標欄設回 0x0D、列再 +0x18         → 欄 0x0D、列 0x7E
var StablePic = struct {
	X, Y             int
	NameCol, NameRow int
	AskRow           int
}{
	X: 112, Y: 86,
	NameCol: 0x14, NameRow: 0x66,
	AskRow: 0x7E,
}

// SheetSteed 是角色資料頁右上角那一塊坐騎（`sub_16239:0x16275`…`0x1633b`，RE-120）：
//
//	sub_1B779(0x106, 2, 0x131, 0x2F, 0)     ; 先把 (262,2)–(305,47) 清成黑
//	sub_1C53D(ds, 0x21, 2, 坐騎圖槽)         ; 圖畫在 (0x21 × 8, 2) = (264, 2)
//	sub_1B6DC(0x105, 1, 0x132, 0x28, 0x0F)  ; 圖的白框 (261,1)–(306,40)
//	sub_1B6DC(0x105, 0x28, 0x132, 0x30, 0x0F) ; 名牌白框 (261,40)–(306,48)
//	sub_1B7E9(0x105, 1, 0x132, 0x30)        ; 兩格共用的投影
//	word_2DF16 = 0x2A、word_2DF14 = 0x22    ; 名字印在 (272, 42)
//
// 名字**只印四個字元**：`0x1634c` 的迴圈界限寫死在 4，遇到 0 就提早停
// （`Lgt Warhorse` 因此變成 `LGT`），這樣才塞得進 46 px 寬的名牌。
var SheetSteed = struct {
	Fill, Frame, Plate Rect
	X, Y               int
	NameCol, NameRow   int
	NameChars          int
}{
	Fill:  Rect{262, 2, 305, 47},
	Frame: Rect{261, 1, 306, 40},
	Plate: Rect{261, 40, 306, 48},
	X:     264, Y: 2,
	NameCol: 0x22, NameRow: 0x2A, NameChars: 4,
}

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
	// ColourListIdle 是可選清單裡沒有反白的那幾列：EGA 索引 3。
	ColourListIdle = [3]uint8{0, 170, 170}
	// ColourListArrow 是清單右側捲動箭頭的顏色：EGA 索引 4。
	ColourListArrow = [3]uint8{170, 0, 0}
	// ColourQuestBadge 是任務一覽那一格的外框：EGA 索引 0x0A
	//（`sub_168F8` 傳給 `sub_1BA9D` 的最後一個參數，RE-111）。
	ColourQuestBadge = [3]uint8{85, 255, 85}
	ColourShadow     = [3]uint8{0, 0, 0}
)

// CJKInn 是旅店畫面的中文模式版面（放大後的像素）。與其他中文版面一樣，這一份是
// remake 自訂的，不對拍：位置是英文版面的兩倍，只有文字改用 16×18 的中文字格，
// 因此行距與框內留白要重排。
var CJKInn = struct {
	PartyX, PartyY, PartyStep, PartyWidth int
	PortraitBoxHeight, PlateHeight        int
	PortraitX, PortraitY, NameX, NameY    int
	PicX, PicY                            int
	TitleX, TitleY, TitleW, TitleH        int
	TitleTextY, TitleStep                 int
	DialogX, DialogY, DialogW, DialogH    int
	DialogTextX, DialogTextY, DialogStep  int
	DialogCells                           int
	IconX, IconY, IconStep, IconW, IconH  int
	Scale                                 int
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
	IconX:       64, IconY: 340, IconStep: 64, IconW: 48, IconH: 44,
	Scale: 2,
}

// CJKOutdoor 是戶外畫面的中文模式版面，與 CJKInn 一樣是 remake 自訂的，不對拍：
// 位置大致是英文版面的兩倍，日期框改放兩行 16 px 中文字因此加高。
var CJKOutdoor = struct {
	PartyX, PartyY, PartyStep, PartyWidth int
	PortraitBoxHeight, PlateHeight        int
	PortraitX, PortraitY, NameX, NameY    int
	DateX, DateY, DateW, DateH            int
	DateTextX, DateTextY, DateStep        int
	CrestX, CrestY                        int
	MapX, MapY, MapW, MapH                int
	ViewX, ViewY, Tile                    int
	IconX, IconY, IconStep, IconW, IconH  int
	Scale                                 int
}{
	PartyX: 28, PartyY: 2, PartyStep: 96, PartyWidth: 86,
	PortraitBoxHeight: 80, PlateHeight: 18,
	PortraitX: 32, PortraitY: 4, NameX: 32, NameY: 84,
	DateX: 28, DateY: 114, DateW: 134, DateH: 44,
	DateTextX: 34, DateTextY: 118, DateStep: 20,
	CrestX: 48, CrestY: 170,
	MapX: 206, MapY: 114, MapW: 388, MapH: 196,
	ViewX: 208, ViewY: 116, Tile: 32,
	IconX: 64, IconY: 340, IconStep: 64, IconW: 48, IconH: 44,
	Scale: 2,
}

// SheetStatFirstRow 是八個屬性的第一列；每一列往下 8 px，
// 與 KOLDATA 記錄 73h 的 ~V 座標相同。
const SheetStatFirstRow = 58

// Rect 是含端點的矩形，與原版繪圖 helper 的引數同一種寫法。
type Rect struct{ X0, Y0, X1, Y1 int }

// Width／Height 是像素寬高。
func (r Rect) Width() int  { return r.X1 - r.X0 + 1 }
func (r Rect) Height() int { return r.Y1 - r.Y0 + 1 }

// SheetLayout 是 MIRROR 角色資料頁的版面（RE-074）。座標全部是
// `sub_15ABF`／`sub_16239` 推給繪圖 helper 的引數，不是量出來的。
var SheetLayout = struct {
	Header, Stats, Weapons, Gold, Bars Rect
	HeaderRule, WeaponRule             Rect
	DamageBar, NutritionBar            Rect
	NameRow, GenderRow                 int
	RaceRow, RankRow                   int
	HeightRow, WeightRow               int
	LeftCol, RightCol, CentreCells     int
	HeightCol, WeightCol               int
	StatValueCol                       int
	WeaponNameCol, WeaponOffCol        int
	WeaponDefCol, WeaponFirstRow       int
	WeaponRowStep, WeaponRows          int
	GoldValueCol, GoldRow              int
	AdventureCol, AdventureRow         int
	MovementCol, MovementRow           int
	// EquipRow 那一列印的是所屬教派與頭銜（RE-121 更正 RE-074 的「盔甲／盾牌」）。
	EquipRow, OrderCol, TitleCol int
	BarCell, BarInset            int
}{
	Header:     Rect{70, 4, 249, 46},
	HeaderRule: Rect{70, 14, 248, 14},
	Stats:      Rect{14, 56, 121, 120},
	Weapons:    Rect{126, 52, 305, 88},
	WeaponRule: Rect{127, 61, 304, 61},
	Gold:       Rect{126, 93, 305, 126},
	Bars:       Rect{38, 130, 288, 158},
	// 兩條長條的青色底：原版先填滿再把紅色蓋上去。
	DamageBar:    Rect{39, 138, 287, 144},
	NutritionBar: Rect{39, 152, 287, 158},

	NameRow: 7, GenderRow: 16, RaceRow: 22, RankRow: 28,
	HeightRow: 34, WeightRow: 40,
	// 抬頭的名字與性別在第 9 欄起的 22 格內置中，右側欄位靠齊第 31 欄。
	LeftCol: 9, RightCol: 31, CentreCells: 22,
	HeightCol: 18, WeightCol: 16,

	StatValueCol: 12,

	WeaponNameCol: 16, WeaponOffCol: 30, WeaponDefCol: 35,
	WeaponFirstRow: 63, WeaponRowStep: 6, WeaponRows: 4,

	GoldValueCol: 29, GoldRow: 95,
	AdventureCol: 33, AdventureRow: 103,
	MovementCol: 37, MovementRow: 111,
	EquipRow: 119, OrderCol: 16, TitleCol: 38,

	// 長條一格 8 px，紅色比青色底上下各縮 1 px、左緣同樣從 39 起。
	BarCell: 8, BarInset: 1,
}

// 角色資料頁的顏色：抬頭框是白的，其餘的框與長條底是青色，長條本身是亮紅。
var (
	ColourSheetHeader = [3]uint8{255, 255, 255}
	ColourSheetFrame  = [3]uint8{0, 170, 170}
	ColourSheetBar    = [3]uint8{255, 85, 85}
)
