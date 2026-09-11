package world

// 野外畫面的中文模式版面。與城鎮的中文版面一樣是 remake 自訂的（不對拍）：
// 素材放大三倍（`book.CJKScale`），左欄由上而下是日期框、天候圖、時段／地形標籤，
// 右邊是地圖視窗。左欄的框比原版高——英文版面把標籤擠在 67×15 裡，中文一行要 18 px。

// CJK 是中文模式的版面。座標都是 960×600 畫布上的像素。
var CJK = struct {
	PartyX, PartyY, PartyStep, PartyWidth  int
	PortraitBoxHeight, PlateHeight         int
	PortraitX, PortraitY, NameX, NameY     int
	DateX, DateY, DateW, DateH             int
	DateTextX, DateTextY, DateStep         int
	WeatherX, WeatherY, WeatherW, WeatherH int
	WeatherPicX, WeatherPicY               int
	LabelX, LabelY, LabelW, LabelH         int
	LabelTextY, LabelStep                  int
	MapX, MapY, MapW, MapH                 int
	ViewX, ViewY, Tile                     int
	MarkerSide                             int
	IconX, IconY, IconStep, IconW, IconH   int
	Scale                                  int
}{
	PartyX: 42, PartyY: 3, PartyStep: 144, PartyWidth: 129,
	PortraitBoxHeight: 120, PlateHeight: 27,
	PortraitX: 48, PortraitY: 6, NameX: 48, NameY: 126,
	DateX: 42, DateY: 171, DateW: 201, DateH: 66,
	DateTextX: 51, DateTextY: 177, DateStep: 24,
	WeatherX: 42, WeatherY: 243, WeatherW: 201, WeatherH: 195,
	WeatherPicX: 46, WeatherPicY: 247,
	LabelX: 42, LabelY: 444, LabelW: 201, LabelH: 66,
	LabelTextY: 450, LabelStep: 24,
	MapX: 309, MapY: 171, MapW: 582, MapH: 294,
	ViewX: 312, ViewY: 174, Tile: 48,
	MarkerSide: 6,
	IconX:      96, IconY: 522, IconStep: 96, IconW: 72, IconH: 66,
	Scale: 3,
}
