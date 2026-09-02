package town

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// 城鎮建築的門座標與配對規則由 KOLTOWN 讀出（RE-061）：隊伍走到門的座標時，
// 地圖視窗換成 `TWN{n}PICS[2i+3]`（192×96），圖示列只剩 DOOR 與 U-TURN 兩格。
// 圖示列畫的是 `KOLMPICS[id + 3]`，DOOR 的 id 是 2、U-TURN 是 15。

// DoorIcon／UTurnIcon 是門口狀態的兩個 KOLMPICS 圖號。
const (
	DoorIcon  = 5
	UTurnIcon = 18
)

// Door 是一個門：地圖座標與它對應的建築素材與文字。
type Door struct {
	Index    int    `json:"index"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
	Picture  int    `json:"picture"`
	Portrait int    `json:"portrait"`
	Name     string `json:"name"`
	Person   string `json:"person"`
	// TextMissing 記的是原版資料本身的缺口：TOWN7 的門 21 有圖與肖像，沒有文字。
	TextMissing bool `json:"text_missing"`
}

// TownBuildings 是一座城鎮的門、入城點與地形通行規則。
type TownBuildings struct {
	Town        int     `json:"town"`
	Archive     string  `json:"archive"`
	EntryPoints []Point `json:"entry_points"`
	Doors       []Door  `json:"doors"`
	// Blocked 是「磚塊編號 → 不可通行」的查表。原版存成 32 byte 的位元圖
	// （payload 3 的位移 0x00–0x1F），移動前查目的地格子的磚塊編號（RE-065）。
	Blocked [256]bool
}

// Point 是地圖座標。
type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Buildings 是全部城鎮的建築資料。
type Buildings struct {
	Schema string          `json:"schema"`
	Towns  []TownBuildings `json:"towns"`
}

// rawDoor 對應 JSON 的欄位名；Door 自己的 X/Y 是拆開的兩個欄位。
type rawDoor struct {
	Index       int    `json:"index"`
	X           int    `json:"x"`
	Y           int    `json:"y"`
	Picture     int    `json:"picture"`
	Portrait    int    `json:"portrait"`
	Name        string `json:"name"`
	Person      string `json:"person"`
	TextMissing bool   `json:"text_missing"`
}

type rawTown struct {
	Town         int       `json:"town"`
	Archive      string    `json:"archive"`
	EntryPoints  []Point   `json:"entry_points"`
	Doors        []rawDoor `json:"doors"`
	BlockedTiles []int     `json:"blocked_tiles"`
}

type rawBuildings struct {
	Schema string    `json:"schema"`
	Towns  []rawTown `json:"towns"`
}

// LoadBuildings 讀取抽取器產生的建築 typed data。
func LoadBuildings(path string) (*Buildings, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("讀取城鎮建築資料失敗：%w", err)
	}
	var raw rawBuildings
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("解析城鎮建築資料失敗：%w", err)
	}
	if !strings.HasPrefix(raw.Schema, "kol-town-buildings-data-") {
		return nil, fmt.Errorf("非預期的城鎮建築 schema：%s", raw.Schema)
	}
	out := &Buildings{Schema: raw.Schema}
	for _, t := range raw.Towns {
		town := TownBuildings{Town: t.Town, Archive: t.Archive, EntryPoints: t.EntryPoints}
		for _, tile := range t.BlockedTiles {
			if tile < 0 || tile >= len(town.Blocked) {
				return nil, fmt.Errorf("城鎮 %d 的不可通行磚塊編號 %d 超出範圍", t.Town, tile)
			}
			town.Blocked[tile] = true
		}
		for _, d := range t.Doors {
			if d.X < 0 || d.X >= GridSide || d.Y < 0 || d.Y >= GridSide {
				return nil, fmt.Errorf("城鎮 %d 的門 %d 座標 (%d, %d) 超出地圖",
					t.Town, d.Index, d.X, d.Y)
			}
			town.Doors = append(town.Doors, Door{
				Index: d.Index, X: d.X, Y: d.Y, Picture: d.Picture,
				Portrait: d.Portrait, Name: d.Name, Person: d.Person,
				TextMissing: d.TextMissing,
			})
		}
		out.Towns = append(out.Towns, town)
	}
	return out, nil
}

// Town 回傳指定城鎮的建築資料。
func (b *Buildings) Town(index int) *TownBuildings {
	for i := range b.Towns {
		if b.Towns[i].Town == index {
			return &b.Towns[i]
		}
	}
	return nil
}

// Passable 回傳某個磚塊編號可不可以走進去。
func (t *TownBuildings) Passable(tile int) bool {
	if tile < 0 || tile >= len(t.Blocked) {
		return false
	}
	return !t.Blocked[tile]
}

// DoorAt 回傳位於地圖座標 (x, y) 的門；沒有就回 nil。
// 原版是由小到大掃同一張表，第一個吻合的就成立。
func (t *TownBuildings) DoorAt(x, y int) *Door {
	for i := range t.Doors {
		if t.Doors[i].X == x && t.Doors[i].Y == y {
			return &t.Doors[i]
		}
	}
	return nil
}
