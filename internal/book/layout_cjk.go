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
	FieldX, FieldY, FieldStep int
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
}{
	HeadingX: 48, HeadingY: 32,
	FieldX: 48, FieldY: 72, FieldStep: 24,
	AttrY: 176, AttrStep: 20,
	ColonX: 128, ValueX: 144,
	NameCursorY: 88,
	RightX: 352, RightHeadingY: 32,
	RightListY: 104, RightStep: 24,
	PortraitFrameX: 412, PortraitFrameY: 210,
	PortraitFrameW: 86, PortraitFrameH: 96,
	PortraitDivider: 288,
	PortraitX: 416, PortraitY: 212,
	PortraitNameX: 416, PortraitNameY: 292,
	ActionsX: 352, ActionsY: 356,
	InstructionsY: 104,
}

// CJKTextBox 是中文模式左右兩頁的文字安全矩形（放大後的像素）。
// 交付前要求每個字串都畫在對應的矩形內，由 tools/verify_text_fit.py 檢查。
var CJKTextBox = struct {
	LeftX, LeftY, LeftW, LeftH     int
	RightX, RightY, RightW, RightH int
}{
	LeftX: 48, LeftY: 32, LeftW: 292, LeftH: 366,
	RightX: 352, RightY: 32, RightW: 272, RightH: 366,
}
