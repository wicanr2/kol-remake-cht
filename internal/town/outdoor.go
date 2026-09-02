package town

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"os"
	"strings"
)

// 戶外地圖的版面與捲動規則全部由 KOLTOWN 的繪製迴圈 `sub_115D1` 讀出（RE-059）：
// 視窗是 12×6 格、每格 16×16；格子索引是 `(X << 6) + Y`，X 與 Y 各自夾限。
// 螢幕座標由 `tools/decompose_town_frame.py` 的分群結果量得。

// Viewport 是地圖視窗的內容區：左上角與格數。
var Viewport = struct{ X, Y, Cols, Rows, Tile int }{X: 104, Y: 58, Cols: 12, Rows: 6, Tile: 16}

// ColBias／RowBias 是原版迴圈裡的 0FFFBh／0FFFEh：
// 第 col 欄看的是鏡頭 X−5＋col，第 row 列看的是鏡頭 Y−2＋row，
// 所以隊伍所在的那一格永遠是 (row 2, col 5)。
const (
	ColBias = -5
	RowBias = -2
)

// MapWindow 是地圖視窗的白框：內容區四周各一像素，右下有共用投影。
var MapWindow = Box{X: 103, Y: 57, Width: 194, Height: 98}

// DateBox 是左上的日期框，兩行小字。行距 6 px 比城鎮文字框的 8 px 密。
var DateBox = Box{X: 14, Y: 57, Width: 67, Height: 15, TextX: 16, TextY: 59, LineStep: 6, Cells: 8}

// Crest 是日期框下方的城鎮徽章：`TWN{n}PICS` 的圖號 0，帶 KOLMPICS[28] 共用遮罩。
var Crest = struct{ Index, X, Y int }{Index: 0, X: 24, Y: 85}

// GridSide 是格子陣列的邊長。Y 只有 64 個值——`word_2DF6C` 靜態就是 3Fh。
const GridSide = 64

// Grid 是一個城鎮的地圖格子陣列，一格一 byte ＝ 磚塊編號。
type Grid struct{ cells []byte }

// LoadGrid 從抽取出的 64×64 灰階 PNG 建立格子陣列；像素值即磚塊編號。
func LoadGrid(img image.Image) (*Grid, error) {
	b := img.Bounds()
	if b.Dx() != GridSide || b.Dy() != GridSide {
		return nil, fmt.Errorf("地圖格子陣列尺寸 %dx%d 不是 %d×%d",
			b.Dx(), b.Dy(), GridSide, GridSide)
	}
	gray, ok := img.(*image.Gray)
	if !ok {
		return nil, fmt.Errorf("地圖格子陣列不是 8-bit 灰階 PNG")
	}
	cells := make([]byte, GridSide*GridSide)
	for y := 0; y < GridSide; y++ {
		for x := 0; x < GridSide; x++ {
			cells[y*GridSide+x] = gray.GrayAt(b.Min.X+x, b.Min.Y+y).Y
		}
	}
	return &Grid{cells: cells}, nil
}

// At 回傳地圖座標 (x, y) 的磚塊編號。PNG 是 X 橫向、Y 縱向，對應原版的 `(X << 6) + Y`。
func (g *Grid) At(x, y int) int { return int(g.cells[y*GridSide+x]) }

// TileAtlas 是磚塊 atlas 的排列：16 欄 × 16 列，每塊 16×16。
var TileAtlas = struct{ Index, Columns, Side int }{Index: 1, Columns: 16, Side: 16}

// Outdoor 是戶外畫面的完整狀態。
type Outdoor struct {
	Schema  string   `json:"schema"`
	Town    int      `json:"town"`
	Party   []Member `json:"party"`
	Camera  struct{ X, Y int } `json:"camera"`
	Bounds  struct{ MinX, MaxX, MinY, MaxY int } `json:"bounds"`
	Date    []string `json:"date"`
	DateZh  []string `json:"date_zh"`
	Icons    []int `json:"icons"`
	Selected int    `json:"selected"`

	// Marker 是隊伍標記：Base 對應 word_2DF48（城鎮戶外畫面是 0），
	// Frame 對應 word_2DF24 的動畫畫格，在 0…4 之間來回，
	// Row／Col 對應 word_2DF38／word_2DF36——標記在視窗內的列／欄偏移。
	// 隊伍先在視窗內走，走到一定位置鏡頭才跟著捲動，所以兩者要分開記。
	Marker struct{ Base, Frame, Row, Col int } `json:"marker"`

	// AtDoorCell 是「站在門口」這個狀態。原版是**移動到**門的座標時才觸發
	// （`sub_10000` 的掃門迴圈在移動之後才跑），所以剛從建築物走出來、
	// 站在同一格上時仍然是地圖畫面。位置本身不足以決定，要獨立記。
	AtDoorCell bool `json:"at_door"`

	grid      *Grid
	markers   *Markers
	buildings *TownBuildings
}

// LoadOutdoor 讀取戶外畫面的樣本狀態。這份 JSON 只作對拍輸入，不是遊戲資料模型。
func LoadOutdoor(path string) (*Outdoor, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("讀取戶外畫面樣本失敗：%w", err)
	}
	var out Outdoor
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("解析戶外畫面樣本失敗：%w", err)
	}
	if !strings.HasPrefix(out.Schema, "kol-town-outdoor-") {
		return nil, fmt.Errorf("非預期的戶外畫面 schema：%s", out.Schema)
	}
	if out.Town < 0 || out.Town > 7 {
		return nil, fmt.Errorf("城鎮編號 %d 超出範圍", out.Town)
	}
	if out.Bounds.MaxX == 0 && out.Bounds.MaxY == 0 {
		out.Bounds.MaxX, out.Bounds.MaxY = GridSide-1, GridSide-1
	}
	if out.Marker.Base < 0 || out.Marker.Base+MarkerFrames > MarkerUnits {
		return nil, fmt.Errorf("標記基底 %d 超出範圍", out.Marker.Base)
	}
	if out.Marker.Frame < 0 || out.Marker.Frame >= MarkerFrames {
		return nil, fmt.Errorf("標記畫格 %d 超出範圍", out.Marker.Frame)
	}
	row, col := -RowBias+out.Marker.Row, -ColBias+out.Marker.Col
	if row < 0 || row >= Viewport.Rows || col < 0 || col >= Viewport.Cols {
		return nil, fmt.Errorf("標記偏移 (%d, %d) 讓標記落在視窗外",
			out.Marker.Row, out.Marker.Col)
	}
	// 圖示是看情境出現的：不在門口、不在路上時整列都是空的。
	if len(out.Icons) == 0 {
		out.Selected = 0
	} else if out.Selected < 0 || out.Selected >= len(out.Icons) {
		return nil, fmt.Errorf("選定圖示 %d 超出範圍", out.Selected)
	}
	return &out, nil
}

// SetGrid 掛上該城鎮的格子陣列。
func (o *Outdoor) SetGrid(g *Grid) { o.grid = g }

// SetMarkers 掛上隊伍標記精靈。
func (o *Outdoor) SetMarkers(m *Markers) { o.markers = m }

// SetBuildings 掛上該城鎮的門與建築資料。
func (o *Outdoor) SetBuildings(b *TownBuildings) { o.buildings = b }

// IconRow 回傳目前該畫的圖示列與白框位置。站在門口時原版把圖示列重建成
// DOOR 與 U-TURN 兩格，白框回到第一格。
func (o *Outdoor) IconRow() ([]int, int) {
	if o.AtDoor() != nil {
		return []int{DoorIcon, UTurnIcon}, 0
	}
	return o.Icons, o.Selected
}

// AtDoor 回傳目前門口狀態所在的門；不在門口狀態就回 nil。
func (o *Outdoor) AtDoor() *Door {
	if !o.AtDoorCell || o.buildings == nil {
		return nil
	}
	x, y := o.PartyPosition()
	return o.buildings.DoorAt(x, y)
}

// DoorUnderParty 回傳隊伍座標上的門（不看狀態），移動之後用它決定要不要進門口狀態。
func (o *Outdoor) DoorUnderParty() *Door {
	if o.buildings == nil {
		return nil
	}
	x, y := o.PartyPosition()
	return o.buildings.DoorAt(x, y)
}

// MarkerCell 是隊伍標記目前佔的視窗格。偏移為 0 時落在 (列 2, 欄 5)，
// 也就是鏡頭座標本身那一格。
func (o *Outdoor) MarkerCell() (row, col int) {
	return -RowBias + o.Marker.Row, -ColBias + o.Marker.Col
}

// PartyPosition 是隊伍在地圖上的座標：鏡頭座標加上標記在視窗內的偏移。
func (o *Outdoor) PartyPosition() (x, y int) {
	return o.Camera.X + o.Marker.Col, o.Camera.Y + o.Marker.Row
}

// MarkerOverlay 回傳目前畫格的覆蓋層像素；沒掛標記時回傳 nil。
func (o *Outdoor) MarkerOverlay() []color.RGBA {
	if o.markers == nil {
		return nil
	}
	return o.markers.MarkerOverlay(o.Marker.Base + o.Marker.Frame)
}

// AdvanceMarker 依原版的來回規則推進動畫畫格：走到 4 往回、走到 0 往前。
func (o *Outdoor) AdvanceMarker(step *int) {
	o.Marker.Frame += *step
	if o.Marker.Frame >= MarkerFrames-1 {
		o.Marker.Frame = MarkerFrames - 1
		*step = -1
	}
	if o.Marker.Frame <= 0 {
		o.Marker.Frame = 0
		*step = 1
	}
}

// Archive 是該城鎮的資產 archive 名稱。
func (o *Outdoor) Archive() string { return fmt.Sprintf("TOWN%d", o.Town) }

// CrestArchive 是該城鎮的徽章所在 archive。
func (o *Outdoor) CrestArchive() string { return fmt.Sprintf("TWN%dPICS", o.Town) }

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// View 回傳目前視窗內 6 列 × 12 欄的磚塊編號，夾限規則與原版迴圈相同。
func (o *Outdoor) View() [][]int {
	rows := make([][]int, Viewport.Rows)
	for row := range rows {
		line := make([]int, Viewport.Cols)
		for col := range line {
			x := clamp(o.Camera.X+col+ColBias, o.Bounds.MinX, o.Bounds.MaxX)
			y := clamp(o.Camera.Y+row+RowBias, o.Bounds.MinY, o.Bounds.MaxY)
			line[col] = o.grid.At(x, y)
		}
		rows[row] = line
	}
	return rows
}

// Move 依方向鍵移動鏡頭。夾限與繪製時相同，走到邊界就停住。
// 移動之後重新判斷門口狀態——原版的掃門迴圈就是接在移動後面。
func (o *Outdoor) Move(dx, dy int) {
	o.Camera.X = clamp(o.Camera.X+dx, o.Bounds.MinX, o.Bounds.MaxX)
	o.Camera.Y = clamp(o.Camera.Y+dy, o.Bounds.MinY, o.Bounds.MaxY)
	o.AtDoorCell = o.DoorUnderParty() != nil
}
