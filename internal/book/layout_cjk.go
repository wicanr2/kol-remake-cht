package book

// 中文模式的版面。**這一份不是量自原版**，原版沒有中文畫面；它是 remake 自訂的
// 版面，用來容納 16×18 的中文字格。畫布放大到原版的**三倍**（960×600），書本美術以
// 最近鄰放大，因此書頁比例與原版一致。
//
// **三倍是算出來的，不是挑的。** 原版最小行距 6 px，而中文字模格高 18——兩倍畫布
// 只給得起 12 px，放不下，於是每一處行距都得另外加大，文字因此與書頁美術的線條
// 錯開；三倍剛好 `6 × 3 = 18`，位置與行距都照原版等比推得出來。
//
// 英文嚴格對拍模式仍用 320×200 與原版的 8×6 字格，兩者共用同一份狀態機。

// CJKScale 是中文模式相對原版解析度的放大倍率。
const CJKScale = 3

// CJKCell 是中文字模格的邊長。
const CJKCell = 16

// CJKLayout 是中文 New Character 頁的座標表（單位是放大後的像素）。
var CJKLayout = struct {
	HeadingX, HeadingY             int
	FieldX                         int
	NameLabelY, NameValueY         int
	RaceY, SecondaryY, ClassY      int
	AttrY, AttrStep                int
	ColonX, ValueX                 int
	NameCursorY                    int
	RightX, RightHeadingY          int
	RightListY, RightStep          int
	PortraitFrameX, PortraitFrameY int
	PortraitFrameW, PortraitFrameH int
	PortraitDivider                int
	PortraitX, PortraitY           int
	PortraitNameX, PortraitNameY   int
	ActionsX, ActionsY             int
	InstructionsY                  int
	MenuX, MenuY, MenuStep         int
	RosterNamesX, RosterLocationsX int
	RosterY, RosterStep            int
}{
	HeadingX: 72, HeadingY: 48,
	FieldX: 72,
	// 名字最長 16 個 ASCII 字元，和標籤同一列放不下，因此另起一行畫在整頁寬度上——
	// 原版英文版面也是這樣處理名字的。
	NameLabelY: 108, NameValueY: 138,
	RaceY: 174, SecondaryY: 210, ClassY: 246,
	AttrY: 285, AttrStep: 18,
	ColonX: 192, ValueX: 216,
	NameCursorY: 162,
	RightX:      528, RightHeadingY: 48,
	RightListY: 156, RightStep: 18,
	PortraitFrameX: 618, PortraitFrameY: 315,
	PortraitFrameW: 129, PortraitFrameH: 144,
	PortraitDivider: 432,
	PortraitX:       624, PortraitY: 318,
	PortraitNameX: 624, PortraitNameY: 438,
	ActionsX: 528, ActionsY: 534,
	InstructionsY: 156,
	MenuX:         72, MenuY: 156, MenuStep: 18,
	RosterNamesX: 72, RosterLocationsX: 528, RosterY: 138, RosterStep: 18,
}

// CJKTextBox 是中文模式左右兩頁的文字安全矩形（放大後的像素）。
// 交付前要求每個字串都畫在對應的矩形內，由 tools/verify_text_fit.py 檢查。
var CJKTextBox = struct {
	LeftX, LeftY, LeftW, LeftH     int
	RightX, RightY, RightW, RightH int
}{
	// 書頁底緣的裝飾從原版第 190 列開始，放大三倍即 570，因此安全矩形到 y=570 為止。
	LeftX: 72, LeftY: 48, LeftW: 438, LeftH: 522,
	RightX: 528, RightY: 48, RightW: 408, RightH: 522,
}

// CJKConfirm 是中文模式的刪除確認視窗。位置與尺寸是原版視窗放大三倍之後再加高，
// 讓兩行 16×16 中文字上下各留 8 px；投影厚度與偏移同樣是原版的兩倍。
// 這一份和其他中文版面一樣是 remake 自訂的，不對拍。
var CJKConfirm = struct {
	X, Y, Width, Height            int
	ShadowOffset, ShadowThickness  int
	Line1X, Line1Y, Line2X, Line2Y int
}{
	X: 282, Y: 174, Width: 396, Height: 84,
	ShadowOffset: 12, ShadowThickness: 12,
	Line1X: 312, Line1Y: 192, Line2X: 384, Line2Y: 216,
}

// CJKPictureInstructions 是圖像層右頁說明的中文版。原版把一句話拆成五行是英文字距下的
// 排版；中文一行放得下較多字，逐行硬譯會讀成「選擇角色／圖像」這種斷句，因此中文模式
// 另給一組自己的行。
var CJKPictureInstructions = []string{
	"USE ARROW KEYS TO BROWSE",
	"CHARACTER PICTURES",
	"PRESS RETURN TO ACCEPT",
}
