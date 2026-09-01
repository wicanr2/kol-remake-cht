package book

// 中文模式的版面。**這一份不是量自原版**，原版沒有中文畫面；它是 remake 自訂的
// 版面，用來容納 16×16 的中文字格。畫布放大到原版的兩倍（640×400），書本美術以
// 最近鄰放大，因此書頁比例與原版一致，只有文字格變大。
//
// 英文嚴格對拍模式仍用 320×200 與原版的 8×6 字格，兩者共用同一份狀態機。

// CJKScale 是中文模式相對原版解析度的放大倍率。
const CJKScale = 2

// CJKCell 是中文字模格的邊長。
const CJKCell = 16

// CJKLayout 是中文 New Character 頁的座標表（單位是放大後的像素）。
var CJKLayout = struct {
	HeadingX, HeadingY       int
	FieldX                   int
	NameLabelY, NameValueY   int
	RaceY, SecondaryY, ClassY int
	AttrY, AttrStep          int
	ColonX, ValueX           int
	NameCursorY              int
	RightX, RightHeadingY    int
	RightListY, RightStep    int
	PortraitFrameX, PortraitFrameY int
	PortraitFrameW, PortraitFrameH int
	PortraitDivider          int
	PortraitX, PortraitY     int
	PortraitNameX, PortraitNameY int
	ActionsX, ActionsY       int
	InstructionsY            int
	MenuX, MenuY, MenuStep   int
	RosterNamesX, RosterLocationsX int
	RosterY, RosterStep      int
}{
	HeadingX: 48, HeadingY: 32,
	FieldX: 48,
	// 名字最長 16 個 ASCII 字元，和標籤同一列放不下，因此另起一行畫在整頁寬度上——
	// 原版英文版面也是這樣處理名字的。
	NameLabelY: 72, NameValueY: 92,
	RaceY: 116, SecondaryY: 140, ClassY: 164,
	AttrY: 190, AttrStep: 19,
	ColonX: 128, ValueX: 144,
	NameCursorY: 108,
	RightX: 352, RightHeadingY: 32,
	RightListY: 104, RightStep: 22,
	PortraitFrameX: 412, PortraitFrameY: 210,
	PortraitFrameW: 86, PortraitFrameH: 96,
	PortraitDivider: 288,
	PortraitX: 416, PortraitY: 212,
	PortraitNameX: 416, PortraitNameY: 292,
	ActionsX: 352, ActionsY: 356,
	InstructionsY: 104,
	MenuX: 48, MenuY: 104, MenuStep: 24,
	RosterNamesX: 48, RosterLocationsX: 352, RosterY: 92, RosterStep: 18,
}

// CJKTextBox 是中文模式左右兩頁的文字安全矩形（放大後的像素）。
// 交付前要求每個字串都畫在對應的矩形內，由 tools/verify_text_fit.py 檢查。
var CJKTextBox = struct {
	LeftX, LeftY, LeftW, LeftH     int
	RightX, RightY, RightW, RightH int
}{
	// 書頁底緣的裝飾從原版第 190 列開始，放大兩倍即 380，因此安全矩形到 y=380 為止。
	LeftX: 48, LeftY: 32, LeftW: 292, LeftH: 348,
	RightX: 352, RightY: 32, RightW: 272, RightH: 348,
}

// CJKConfirm 是中文模式的刪除確認視窗。位置與尺寸是原版視窗放大兩倍之後再加高，
// 讓兩行 16×16 中文字上下各留 8 px；投影厚度與偏移同樣是原版的兩倍。
// 這一份和其他中文版面一樣是 remake 自訂的，不對拍。
var CJKConfirm = struct {
	X, Y, Width, Height             int
	ShadowOffset, ShadowThickness   int
	Line1X, Line1Y, Line2X, Line2Y  int
}{
	X: 188, Y: 116, Width: 264, Height: 56,
	ShadowOffset: 8, ShadowThickness: 8,
	Line1X: 264, Line1Y: 124, Line2X: 296, Line2Y: 148,
}
