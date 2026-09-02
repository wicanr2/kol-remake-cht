package world

// 野外畫面的版面（RE-068）。矩形都直接來自 `sub_2A39B` 與 `sub_29B40` 的引數。

// Box 是一個框：左上角與寬高。
type Box struct{ X, Y, Width, Height int }

// MapWindow 是地圖視窗的白框，與城鎮同一個矩形（`sub_1B7E9(67h,39h,128h,9Ah)`）。
var MapWindow = Box{X: 103, Y: 57, Width: 194, Height: 98}

// Viewport 是地圖視窗的內容區。
var Viewport = struct{ X, Y, Cols, Rows, Tile int }{X: 104, Y: 58, Cols: 12, Rows: 6, Tile: 16}

// ColBias／RowBias 是繪製迴圈的固定偏移，與城鎮共用同一支 `sub_115D1`。
const (
	ColBias = -5
	RowBias = -2
)

// WeatherBox 是天候圖的白框：`sub_1B7E9(0Fh, 4Bh, 50h, 8Ah)` ＝ (15,75)–(80,138)。
var WeatherBox = Box{X: 15, Y: 75, Width: 66, Height: 64}

// WeatherPicture 是天候圖本身的左上角，落在白框內緣。
var WeatherPicture = struct{ X, Y int }{X: 16, Y: 76}

// LabelBox 是時段／地形的標籤框：`sub_1BA9D(0Eh, 8Eh, 50h, 9Ch, 0Fh)`。
var LabelBox = Box{X: 14, Y: 142, Width: 67, Height: 15}

// LabelLines 是兩行標籤的起點。原版用的是 byte 欄（乘 8）與掃描列。
var LabelLines = [2]struct{ X, Y int }{{X: 16, Y: 144}, {X: 24, Y: 150}}

// MarkerSide 是隊伍那個閃爍的點的邊長。
const MarkerSide = 2

// Screen 是野外畫面的完整狀態。
type Screen struct {
	Travel  *Travel
	Weather *Weather
	// Group 是天候組（0…5）；Phase 是雲層旋轉過的像素數。
	Group, Phase int
	// TimeIndex 是時段（0…4）。
	TimeIndex int
	Date      []string
	Icons     []int
	Selected  int
	// MarkerOff 是閃爍的另一半：點暫時不畫。原版每隔幾個 tick 就切一次。
	MarkerOff bool
	// Names 供應標籤的兩行文字。原版直接畫 KOLDATA 的記錄，而那些記錄**自帶空白**
	// 用來置中（`"  DAWN ]"`），所以不能先去空白。
	Names Namer
}

// Namer 供應時段與地形的顯示文字。
type Namer interface {
	TimeText(index int) string
	TerrainText(code int) string
}

// TimeLine／TerrainLine 是標籤框的兩行。
func (s *Screen) TimeLine() string {
	if s.Names == nil {
		return name(TimeNames, s.TimeIndex)
	}
	return s.Names.TimeText(s.TimeIndex)
}

func (s *Screen) TerrainLine() string {
	code := s.Travel.Data.TerrainCode(s.Travel.Tile())
	if s.Names == nil {
		return name(TerrainNames, code)
	}
	return s.Names.TerrainText(code)
}

func name(list []string, index int) string {
	if index < 0 || index >= len(list) {
		return ""
	}
	return list[index]
}

// View 回傳視窗內 6 列 × 12 欄的磚塊編號。
func (s *Screen) View() [][]int {
	x, y := s.Travel.Camera()
	rows := make([][]int, Viewport.Rows)
	for row := range rows {
		line := make([]int, Viewport.Cols)
		for col := range line {
			line[col] = s.Travel.Grid.At(x+col+ColBias, y+row+RowBias)
		}
		rows[row] = line
	}
	return rows
}
