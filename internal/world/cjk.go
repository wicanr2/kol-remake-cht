package world

// 野外畫面的中文模式版面。與城鎮的中文版面一樣是 remake 自訂的（不對拍）：
// 素材放大兩倍，左欄由上而下是日期框、天候圖、時段／地形標籤，右邊是地圖視窗。
// 英文版面把標籤擠在 67×15 的框裡，中文兩個字就要 32 px，因此左欄整個重排。

// CJK 是中文模式的版面。座標都是 640×400 畫布上的像素。
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
	PartyX: 28, PartyY: 2, PartyStep: 96, PartyWidth: 86,
	PortraitBoxHeight: 80, PlateHeight: 18,
	PortraitX: 32, PortraitY: 4, NameX: 32, NameY: 84,
	DateX: 28, DateY: 114, DateW: 134, DateH: 44,
	DateTextX: 34, DateTextY: 118, DateStep: 20,
	WeatherX: 28, WeatherY: 162, WeatherW: 134, WeatherH: 130,
	WeatherPicX: 31, WeatherPicY: 165,
	LabelX: 28, LabelY: 296, LabelW: 134, LabelH: 44,
	LabelTextY: 300, LabelStep: 20,
	MapX: 206, MapY: 114, MapW: 388, MapH: 196,
	ViewX: 208, ViewY: 116, Tile: 32,
	MarkerSide: 4,
	IconX:      64, IconY: 348, IconStep: 64, IconW: 48, IconH: 44,
	Scale: 2,
}
