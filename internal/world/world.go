// Package world 是野外地圖這一段的狀態機（RE-066／RE-068）。
//
// 版面與城鎮共用同一個地圖視窗（12×6 格、每格 16×16），但兩件事不同：
//
//   - 隊伍在世界地圖上有**次格座標**：`word_2DF36`／`word_2DF38` 在 0…7 之間繞回，
//     繞回時世界座標才走一格。標記是 2×2 像素的點，畫在視窗第 3 列第 5 欄那一格內、
//     偏移「次格座標 × 2」像素。
//   - 左側面板換成天候圖與「時段／地形」標籤，兩者都由腳下那一格的磚塊編號決定。
package world

import (
	"encoding/json"
	"fmt"
	"image"
	"os"
	"strings"
)

// GridSide 是世界地圖的邊長，與城鎮相同。
const GridSide = 64

// SubCells 是一格之內的次格數；次格座標在 0…SubCells-1 之間繞回。
const SubCells = 8

// SubPixels 是一個次格在畫面上的像素寬度。8 × 2 ＝ 16，正好是一格磚。
const SubPixels = 2

// PartyCell 是隊伍標記固定佔的視窗格。原版讀目的地磚塊時的基底是 0x29 ＝ 3×12＋5。
var PartyCell = struct{ Row, Col int }{Row: 3, Col: 5}

// TerrainNames／TimeNames 是原版 KOLDATA 記錄 0xDA／0xD5 起的名稱。
var (
	TerrainNames = []string{"DESERT", "RIVER", "SWAMP", "HILL",
		"ROAD", "FOREST", "SHORE", "PLAINS"}
	TimeNames = []string{"DAWN", "MID MORN", "NOON", "AFT NOON", "TWILIGHT"}
)

// TerrainRecordBase／TimeRecordBase 是那些名稱在 KOLDATA 的記錄號。
const (
	TerrainRecordBase = 0xDA
	TimeRecordBase    = 0xD5
)

// Town 是世界地圖上的一個城鎮入口。
type Town struct {
	Record     int    `json:"record"`
	X, Y       int    `json:"-"`
	Index      int    `json:"index"`
	Code       string `json:"code"`
	Disk       string `json:"disk"`
	Town       int    `json:"town"`
	EntryPoint int    `json:"entry_point"`
	Camera     struct{ X, Y int } `json:"camera"`
	Bounds     struct {
		MinX int `json:"min_x"`
		MaxX int `json:"max_x"`
		MinY int `json:"min_y"`
		MaxY int `json:"max_y"`
	} `json:"bounds"`
}

type rawTown struct {
	Town
	RawX int `json:"x"`
	RawY int `json:"y"`
}

// Data 是野外地圖的數值表。
type Data struct {
	Schema  string `json:"schema"`
	Terrain []int  `json:"terrain_by_tile"`
	Road    []int  `json:"road_by_tile"`
	Blocked []int  `json:"blocked_tiles"`
	Towns   []rawTown `json:"towns"`
	// Arena 是競技場那一格（`sub_29004:0x29416` 比的 `payload3[0x320]`／`[0x321]`，
	// RE-077／RE-108）。走到那裡跳出 DOOR／U-TURN，選 DOOR 就進場。
	Arena struct{ X, Y int } `json:"arena"`

	blocked [256]bool
	towns   []Town
}

// LoadData 讀取抽取器產生的野外地圖數值表。
func LoadData(path string) (*Data, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("讀取野外地圖資料失敗：%w", err)
	}
	var d Data
	if err := json.Unmarshal(body, &d); err != nil {
		return nil, fmt.Errorf("解析野外地圖資料失敗：%w", err)
	}
	if !strings.HasPrefix(d.Schema, "kol-world-map-data-") {
		return nil, fmt.Errorf("非預期的野外地圖 schema：%s", d.Schema)
	}
	if len(d.Terrain) != 256 || len(d.Road) != 256 {
		return nil, fmt.Errorf("逐磚塊的表長度不是 256")
	}
	for _, t := range d.Blocked {
		if t < 0 || t >= len(d.blocked) {
			return nil, fmt.Errorf("不可通行磚塊編號 %d 超出範圍", t)
		}
		d.blocked[t] = true
	}
	for _, r := range d.Towns {
		t := r.Town
		t.X, t.Y = r.RawX, r.RawY
		d.towns = append(d.towns, t)
	}
	return &d, nil
}

// Passable 回傳某個磚塊編號可不可以走進去。
func (d *Data) Passable(tile int) bool {
	if tile < 0 || tile >= len(d.blocked) {
		return false
	}
	return !d.blocked[tile]
}

// TerrainCode 回傳某個磚塊編號的地形碼。
func (d *Data) TerrainCode(tile int) int {
	if tile < 0 || tile >= len(d.Terrain) {
		return -1
	}
	return d.Terrain[tile]
}

// TerrainName 回傳某個磚塊編號的地形名稱。
func (d *Data) TerrainName(tile int) string {
	if tile < 0 || tile >= len(d.Terrain) {
		return ""
	}
	code := d.Terrain[tile]
	if code < 0 || code >= len(TerrainNames) {
		return ""
	}
	return TerrainNames[code]
}

// OnRoad 回傳腳下這一格算不算在路上；原版靠它決定要不要給 ROAD／SIGNPOST。
func (d *Data) OnRoad(tile int) bool {
	return tile >= 0 && tile < len(d.Road) && d.Road[tile] != 0
}

// TownAt 回傳世界座標上的城鎮入口；沒有就回 nil。
// ArenaAt 回報 (x, y) 是不是競技場那一格。
func (d *Data) ArenaAt(x, y int) bool { return d.Arena.X == x && d.Arena.Y == y }

func (d *Data) TownAt(x, y int) *Town {
	for i := range d.towns {
		if d.towns[i].X == x && d.towns[i].Y == y {
			return &d.towns[i]
		}
	}
	return nil
}

// TownCode 回傳某座城鎮、某個入城點的三字母代碼；查不到就回空字串。
func (d *Data) TownCode(town, entry int) string {
	for i := range d.towns {
		if d.towns[i].Town == town && d.towns[i].EntryPoint == entry {
			return d.towns[i].Code
		}
	}
	return ""
}

// TownEntry 回傳某座城鎮、某個入城點的完整記錄；查不到就回 nil。
// 讀檔要靠它把鏡頭夾限放回進城時的那一組（SPEC-006）。
func (d *Data) TownEntry(town, entry int) *Town {
	for i := range d.towns {
		if d.towns[i].Town == town && d.towns[i].EntryPoint == entry {
			return &d.towns[i]
		}
	}
	return nil
}

// TownCount 是有資料的城鎮入口數。
func (d *Data) TownCount() int { return len(d.towns) }

// Grid 是世界地圖的格子陣列，一格一 byte ＝ 磚塊編號。
type Grid struct{ cells []byte }

// LoadGrid 從抽取出的 64×64 灰階 PNG 建立格子陣列。
func LoadGrid(img image.Image) (*Grid, error) {
	b := img.Bounds()
	if b.Dx() != GridSide || b.Dy() != GridSide {
		return nil, fmt.Errorf("世界地圖尺寸 %dx%d 不是 %d×%d",
			b.Dx(), b.Dy(), GridSide, GridSide)
	}
	gray, ok := img.(*image.Gray)
	if !ok {
		return nil, fmt.Errorf("世界地圖不是 8-bit 灰階 PNG")
	}
	cells := make([]byte, GridSide*GridSide)
	for y := 0; y < GridSide; y++ {
		for x := 0; x < GridSide; x++ {
			cells[y*GridSide+x] = gray.GrayAt(b.Min.X+x, b.Min.Y+y).Y
		}
	}
	return &Grid{cells: cells}, nil
}

// At 回傳世界座標的磚塊編號，座標先夾限在地圖之內（繪製迴圈也是這樣做的）。
func (g *Grid) At(x, y int) int {
	return int(g.cells[clamp(y, 0, GridSide-1)*GridSide+clamp(x, 0, GridSide-1)])
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
