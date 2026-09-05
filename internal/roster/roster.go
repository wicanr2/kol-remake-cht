// Package roster 讀取由 tools/extract_character_roster.py 抽出的角色槽位 typed data。
//
// 執行期不解析原版 `CHARDATA`；槽位版面與所在地組法都由原版角色清單畫面驗證過。
package roster

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Slot struct {
	Slot     int    `json:"slot"`
	Name     string `json:"name"`
	Place    string `json:"place"`
	Town     string `json:"town"`
	Occupied bool   `json:"occupied"`
	// Portrait 是這一格角色頭像命中的 CHARPICS 模板圖號。存檔的頭像逐像素與
	// 80 張模板比對，16 格全部命中；抽取器對不上就會失敗，不會靜靜留白。
	Portrait int `json:"portrait"`
	// Female 是角色記錄位移 0x1A；Wounds 是位移 0x1B 的傷勢，非 0 時旅店的招呼語
	// 換成「你受傷了，該去修道院」。
	Female bool `json:"female"`
	Wounds int  `json:"wounds"`
	// Gold 是角色記錄位移 0x2A–0x2C 的 24-bit 大端金幣；Building 是位移 0x2E／0x2F
	// 記下的留宿建築編號，與建築 payload 的 0x3A／0x3B 相同（RE-073）。
	Gold     int `json:"gold"`
	Building int `json:"building"`
	// Quests 是角色記錄 0x1E2 起的六個 byte：一個 byte 塞四個 2-bit 欄位，
	// 對應 24 個任務地點（RE-079）。
	Quests []int `json:"quests"`
	// Sheet 是 MIRROR 角色資料頁的欄位（RE-074）。
	Sheet *Sheet `json:"sheet"`
	// Spells 是角色記錄 0x200 起、每 16 byte 一筆、共十六筆的法術（RE-100）。
	// 版面與商品記錄相同，所以型別共用；空的那一格是 nil。
	Spells []*Item `json:"spells"`
	// Equipment 是穿戴欄十格：部位 0…8 在角色記錄 0x100 + 16 × 部位，
	// 第十格是手上那件（0x30）。空的那一格是 nil（RE-080）。
	Equipment []*Item `json:"equipment"`
	// Belt 是腰帶那一格（角色記錄 0x190）：立繪最左邊那一層畫的就是它，
	// 原版名冊裡只有兩名角色有，放的都是武器（RE-126／RE-127）。
	Belt *Item `json:"belt"`
	// Progress 是角色記錄 0x1DD 起那五格「已回報的任務數」（RE-107）。
	Progress []int `json:"progress"`
	// Carried 是交任務物品時要逐 byte 比的那十七格（RE-109）。typed 的
	// Equipment 丟掉了原始位元組，比對用不上。
	Carried []CarriedSlot `json:"carried"`
}

// CarriedSlot 是身上一格的原始十六 byte。Offset 是它在角色記錄裡的位移。
type CarriedSlot struct {
	Offset int    `json:"offset"`
	Bytes  string `json:"bytes"`
}

// Item 是穿戴欄的一格，欄位名與商店的商品記錄相同（RE-075）。
type Item struct {
	Name   string `json:"name"`
	Price  int    `json:"price"`
	Kind   int    `json:"kind"`
	Damage bool   `json:"damage"`
	Hands  int    `json:"hands"`
	// Class 是物品記錄 +0x0D 整個 byte：武器技能列與攻擊式都比對它。
	Class       int `json:"class"`
	Position    int `json:"position"`
	Encumbrance int `json:"encumbrance"`
	Condition   int `json:"condition"`
	Percent     int `json:"percent"`
	SizeClass   int `json:"size_class"`
	// Effect／TurnCode／PowerCode／Guards 只有法術那十六格會填（RE-100）：
	// 效果編號、持續回合的骰式、強度的骰式（bit 7 是「要不要選目標格」），
	// 以及擋招要比的四格護符。
	Effect    int   `json:"effect,omitempty"`
	TurnCode  int   `json:"turn_code,omitempty"`
	PowerCode int   `json:"power_code,omitempty"`
	Guards    []int `json:"guards,omitempty"`
	Dice      struct {
		Count int `json:"count"`
		Sides int `json:"sides"`
		Bonus int `json:"bonus"`
		Low   int `json:"low"`
		High  int `json:"high"`
	} `json:"dice"`
	Record string `json:"record"`
}

// Weapon 是武器熟練度的一列。
type Weapon struct {
	Row        int    `json:"row"`
	Name       string `json:"name"`
	Weapon     int    `json:"weapon"`
	Offence    int    `json:"offence"`
	Defence    int    `json:"defence"`
	OffenceRaw int    `json:"offence_raw"`
	DefenceRaw int    `json:"defence_raw"`
}

// Sheet 是角色資料頁要用到的欄位，欄位名與抽取器的輸出相同。
type Sheet struct {
	Stats           []int    `json:"stats"`
	Race            int      `json:"race"`
	Profession      int      `json:"profession"`
	Rank            int      `json:"rank"`
	AdventurePoints int      `json:"adventure_points"`
	MovementRate    int      `json:"movement_rate"`
	Nutrition       int      `json:"nutrition"`
	HeightFeet      int      `json:"height_feet"`
	HeightInches    int      `json:"height_inches"`
	Weight          int      `json:"weight"`
	Weapons         []Weapon `json:"weapons"`
	// Order 與 Title 是角色記錄 0x1A0／0x1A8 那兩個八 byte 的 5-bit 名字：
	// 所屬教派與頭銜，由 SCROLL 那一格入會時寫進去（RE-121）。
	Order string `json:"order"`
	Title string `json:"title"`
}

// Location 是清單右欄顯示的字串：固定長度的所在地、一個空白、三字城鎮代碼。
func (s Slot) Location() string {
	return s.Place + " " + s.Town
}

type Roster struct {
	Schema string `json:"schema"`
	Layout struct {
		SlotCount int `json:"slot_count"`
	} `json:"layout"`
	Slots []Slot `json:"slots"`
}

// Load 讀取角色槽位資料並做基本一致性檢查。
func Load(path string) (*Roster, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("讀取角色槽位失敗：%w", err)
	}
	var r Roster
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, fmt.Errorf("解析角色槽位失敗：%w", err)
	}
	if !strings.HasPrefix(r.Schema, "kol-character-roster-") {
		return nil, fmt.Errorf("非預期的角色槽位 schema：%s", r.Schema)
	}
	if len(r.Slots) != r.Layout.SlotCount {
		return nil, fmt.Errorf("槽位數 %d 與 layout 宣告的 %d 不符",
			len(r.Slots), r.Layout.SlotCount)
	}
	return &r, nil
}

// Occupied 回傳已建立角色的槽位，順序即槽位順序。
func (r *Roster) Occupied() []Slot {
	out := make([]Slot, 0, len(r.Slots))
	for _, s := range r.Slots {
		if s.Occupied {
			out = append(out, s)
		}
	}
	return out
}

// Full 表示 16 個槽位都已使用。原版在滿槽時不進 writer；remake 沿用同一條件。
func (r *Roster) Full() bool { return len(r.Occupied()) >= r.Layout.SlotCount }
