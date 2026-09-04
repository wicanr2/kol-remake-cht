package town

import (
	"encoding/hex"
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
	// IconFlags 是建築 payload 位移 0x3C／0x3D 的兩個旗標 byte，決定內部圖示列。
	IconFlags [2]int `json:"icon_flags"`
	// Price 是 payload 位移 0x38／0x39 的價錢（高位元組在前）。客棧用整個 16-bit
	// 當住一晚的總價，0 代表免費；修道院只用低位元組，見 HealRate（RE-073）。
	Price int `json:"price"`
	// HealRate 是 payload 位移 0x39：修道院每一點傷勢的價錢。
	HealRate int `json:"heal_rate"`
	// RecordID 是 payload 位移 0x3A／0x3B 的建築編號。原版住宿存檔時把它寫進
	// 角色記錄位移 0x2E／0x2F，記下這名角色留在哪一間客棧。
	RecordID int `json:"record_id"`
	// Goods 是商店的商品表（RE-075）。馬廄的表版面不同，只有名字。
	Goods []Good `json:"goods"`
	// Skills 是訓練建築教的四門技能（RE-076）。
	Skills []TrainSkill `json:"skills"`
	// 收任務物品那一組（RE-109）。`icon_flags[1]` 的 bit 2 開啟才有；
	// QuestWantsBytes 是要收的那件物品十六 byte 的十六進位，
	// QuestGivesBytes 是換給你的那件（沒有就是空字串）。
	QuestSite       int    `json:"quest_site"`
	QuestFinal      bool   `json:"quest_final"`
	QuestWants      string `json:"quest_wants"`
	QuestWantsBytes string `json:"quest_wants_bytes"`
	QuestGives      string `json:"quest_gives"`
	QuestGivesBytes string `json:"quest_gives_bytes"`
}

// QuestFlagBit 是 `icon_flags[1]` 裡「收任務物品」那一位（`sub_11A93:0x11B5A`）。
const QuestFlagBit = 4

// Quest 把一扇門的任務欄位換成規則層要的形狀；這一棟不收任務物品就回 nil。
func (d *Door) Quest() *QuestBuilding {
	if d.IconFlags[1]&QuestFlagBit == 0 || d.QuestWantsBytes == "" {
		return nil
	}
	wants, err := hex.DecodeString(d.QuestWantsBytes)
	if err != nil || len(wants) != CarriedStride {
		return nil
	}
	out := &QuestBuilding{Site: d.QuestSite, Final: d.QuestFinal, Wants: wants}
	if gives, err := hex.DecodeString(d.QuestGivesBytes); err == nil &&
		len(gives) == CarriedStride {
		out.Gives = gives
	}
	return out
}

// IsTrainer 回傳這棟建築有沒有 AWARD 的訓練（旗標 bit7）。
func (d *Door) IsTrainer() bool { return d.IconFlags[0]&0x80 != 0 }

// IsInn 回傳這棟建築是不是能住宿存檔的客棧（旗標 bit3）。
func (d *Door) IsInn() bool { return d.IconFlags[0]&0x08 != 0 }

// IsAbbey 回傳這棟建築是不是能治療的修道院（旗標 bit6）。
func (d *Door) IsAbbey() bool { return d.IconFlags[0]&0x40 != 0 }

// IsShop 回傳這棟建築有沒有 TRADE：旗標 bit1／bit4／bit5 任一（RE-067）。
func (d *Door) IsShop() bool { return d.IconFlags[0]&0x32 != 0 }

// IsStable 回傳這是不是馬廄（旗標 bit5）。馬廄的商品表版面與一般商店不同。
func (d *Door) IsStable() bool { return d.IconFlags[0]&0x20 != 0 }

// IsTavern 回傳這是不是賣食物的店（旗標 bit4）。買完會問要不要當場吃掉。
func (d *Door) IsTavern() bool { return d.IconFlags[0]&0x10 != 0 }

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
	IconFlags   []int  `json:"icon_flags"`
	Price       int    `json:"price"`
	HealRate    int    `json:"heal_rate"`
	RecordID    int          `json:"record_id"`
	Goods       []Good       `json:"goods"`
	Skills      []TrainSkill `json:"skills"`
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
			door := Door{
				Index: d.Index, X: d.X, Y: d.Y, Picture: d.Picture,
				Portrait: d.Portrait, Name: d.Name, Person: d.Person,
				TextMissing: d.TextMissing, Price: d.Price,
				HealRate: d.HealRate, RecordID: d.RecordID, Goods: d.Goods,
				Skills: d.Skills,
			}
			if len(d.IconFlags) == 2 {
				door.IconFlags = [2]int{d.IconFlags[0], d.IconFlags[1]}
			}
			town.Doors = append(town.Doors, door)
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
