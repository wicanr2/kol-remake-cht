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
	PartyX: 42, PartyY: 3, PartyStep: 144, PartyWidth: 129,
	PortraitBoxHeight: 120, PlateHeight: 27,
	PortraitX: 48, PortraitY: 6, NameX: 48, NameY: 126,
	PicX: 24, PicY: 192,
	TitleX: 306, TitleY: 174, TitleW: 585, TitleH: 60,
	TitleTextY: 175, TitleStep: 24,
	DialogX: 306, DialogY: 252, DialogW: 585, DialogH: 204,
	DialogTextX: 315, DialogTextY: 258, DialogStep: 24,
	DialogCells: 24,
	IconX:       96, IconY: 510, IconStep: 96, IconW: 72, IconH: 66,
	Scale: 3,
}

// CJKOutdoor 是戶外畫面的中文模式版面，與 CJKInn 一樣是 remake 自訂的，不對拍：
// 位置是英文版面的三倍（`book.CJKScale`），行距照原版等比推得。
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
	PartyX: 42, PartyY: 3, PartyStep: 144, PartyWidth: 129,
	PortraitBoxHeight: 120, PlateHeight: 27,
	PortraitX: 48, PortraitY: 6, NameX: 48, NameY: 126,
	DateX: 42, DateY: 171, DateW: 201, DateH: 66,
	DateTextX: 51, DateTextY: 177, DateStep: 24,
	CrestX: 72, CrestY: 255,
	MapX: 309, MapY: 171, MapW: 582, MapH: 294,
	ViewX: 312, ViewY: 174, Tile: 48,
	IconX: 96, IconY: 510, IconStep: 96, IconW: 72, IconH: 66,
	Scale: 3,
}

// SheetStatFirstRow 是八個屬性的第一列；每一列往下 8 px，
// 與 KOLDATA 記錄 73h 的 ~V 座標相同。
// CellWidth 是原版一個字格的寬度。英文字模就是這個寬；中文模式把座標乘上
// `book.CJKScale` 之後，一格變成 24 px 而中文字寬 16，字在格子裡靠左。
const CellWidth = 8

// SheetWeaponHeaderRow 是武器框那一列表頭（`WEAPON SKILL  OFF  DEF`）的列號，
// 取自記錄 0x73 的 `~V36`。
const SheetWeaponHeaderRow = 0x36

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
	// 法術編輯器那一頁的三種顏色（RE-128）。原版存的是 EGA 索引 0x0F 與 3，
	// 但這一頁的底圖換過調色盤——量原版畫面得到的是黑與亮紅。
	ColourSpellField  = [3]uint8{0, 0, 0}
	ColourSpellCursor = [3]uint8{255, 85, 85}
	ColourSpellRule   = [3]uint8{0, 0, 0}
	ColourSheetFrame  = [3]uint8{0, 170, 170}
	ColourSheetBar    = [3]uint8{255, 85, 85}
)

// ItemPageLayout 是資料頁上物品那一頁的版面（`sub_16986`，RE-127）。
// 座標是 320×200 畫布上的像素，全部來自反組譯裡的常數。
var ItemPageLayout = struct {
	// Frame 是整頁那一圈（`sub_1B6DC(0x0F, 2, 0x131, 0x9C, 3)`）。
	Frame Rect
	// Divider 是把畫面切成左右的那一條直線（`sub_1BB57(0xFE, 3, 0xFE, 0x9C, 3)`）。
	Divider Rect
	// RightPane 是直線右邊那一條先清成黑的窄欄
	// （`sub_1B779(0xFF, 3, 0x131, 0x9B, 0)`）。
	RightPane Rect
	// Detail 是選中那一格的明細框（`sub_1BA9D(0x3E, 0x3A, 0x101, 0x7D, 0x0F)`）。
	Detail Rect
	// 清單畫在字格欄 32（＝ x 256）、第 4 列，列高 9（`sub_196A9` 的參數）。
	// 那一欄正好落在分隔線右邊的窄欄裡，寬度剛好六個字元——`SHIELD`、`WEAPON`、
	// `POCKET` 都是六個字。
	ListCol, ListRow, ListStep int
}{
	Frame:     Rect{15, 2, 305, 156},
	Divider:   Rect{254, 3, 254, 156},
	RightPane: Rect{255, 3, 305, 155},
	Detail:    Rect{62, 58, 257, 125},
	ListCol:   32, ListRow: 4, ListStep: 9,
}

// PartyPageFrame 是隊伍立繪那一頁的外框（`sub_18D61` 收尾的
// `sub_1B6DC(0x0E, 2, 0x131, 0x9C, 3)`，RE-126）。比物品頁那一圈往左一個像素。
var PartyPageFrame = Rect{14, 2, 305, 156}

// SheetPageBack 是資料頁那兩頁（物品、隊伍立繪）的背景（`sub_1103F`，RE-127）。
// 在城裡是 KOLPICS 第 5 張（上半）加第 13 張（下半），兩張各 288 px 寬；
// 在野外換成第 4 張整張（288×155）。開關 `word_2DF2C` 是目前所在的城鎮編號，
// 野外迴圈 `sub_29004` 把它設成 0（RE-150）。第 4 張貼在 (2 欄, 3)，與上半那張同一個
// 原點（`sub_1103F(0x20)`，RE-127 的表）。
var SheetPageBack = struct {
	Archive            string
	TopIndex, LowIndex int
	OutdoorIndex       int
	X, TopY, LowY      int
}{
	Archive: "KOLPICS", TopIndex: 5, LowIndex: 13, OutdoorIndex: 4,
	X: 16, TopY: 3, LowY: 0x50,
}

// SheetBackPiece 是背板的一張圖與貼上去的掃描列。
type SheetBackPiece struct{ Index, Y int }

// SheetPageBackPieces 是背板要貼的幾張（`sub_16986:0x169C1`）：城裡（含城鎮戶外）是上下
// 兩半，野外是第 4 張整張（RE-150）。
func SheetPageBackPieces(wild bool) []SheetBackPiece {
	b := SheetPageBack
	if wild {
		return []SheetBackPiece{{b.OutdoorIndex, b.TopY}}
	}
	return []SheetBackPiece{{b.TopIndex, b.TopY}, {b.LowIndex, b.LowY}}
}
