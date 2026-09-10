package battle

import (
	"encoding/json"
	"fmt"
	"os"
)

// 怪物的數值（RE-078／RE-080）。名稱與地形遮罩早就解出來了，屬性是這一輪
// 才從記錄 +0x11 起讀出來的六個 byte：戰鬥要用洞察、體型與健康。

// Monster 是一種怪物，欄位名與 `tools/extract_monsters.py` 的輸出相同。
type Monster struct {
	// SPEC-029：初始化射擊資格，原版怪物記錄 +0x1E／+0x1F。
	RangedChance int `json:"ranged_chance"`
	RangedCode int `json:"ranged_code"`
	Index     int    `json:"index"`
	Name      string `json:"name"`
	Strength  int    `json:"strength"`
	Quickness int    `json:"quickness"`
	Foresight int    `json:"foresight"`
	Size      int    `json:"size"`
	Health    int    `json:"health"`
	// TerrainMask／SeasonMask 是這一隻出沒的地形與季節遮罩（`TOWN0` payload 4 的
	// `0x40`／`0x60`，RE-078）：bit 7 是地形 0、bit 3 是冬。
	TerrainMask int `json:"terrain_mask"`
	SeasonMask  int `json:"season_mask"`
	Intellect   int `json:"intellect"`
	// AttackSkill／ParrySkill 是記錄 +0x17 起的兩個大端 16-bit；戰鬥再乘等級除以 5。
	// 戰後的冒險點數用的是這裡的原值，不折算等級（RE-096）。
	AttackSkill int `json:"attack_skill"`
	ParrySkill  int `json:"parry_skill"`
	// FleeAt／FleeRecover 是 AI 的兩個撤退門檻（記錄 +0x0D／+0x0E，RE-081）：
	// **疲勞比例**達到 FleeAt 就考慮撤退，撤退中降到 FleeRecover 以下就回到正常
	// 判斷（RE-130 更正，當初讀成傷勢比例）。不死系與元素系是 100／0——比例到
	// 100 的那一刻人已經力竭倒下了（RE-131），所以那幾隻永遠不撤退。
	FleeAt      int `json:"flee_at"`
	FleeRecover int `json:"flee_recover"`
	// Shatters 是記錄 +0x10 的 bit 7：力竭倒下那一刻印哪一句
	// （`shatters into pieces!` 而不是 `dies from blood loss!`，RE-131）。
	// 三十二隻裡只有 SKELETON 是這一種。Flags 整格留著。
	Flags    int  `json:"flags"`
	Shatters bool `json:"shatters"`
	// Fearsome 是記錄 +0x10 的 bit 0：這一種怪物在場時，隊員每要出手一次就得過
	// 一次意志檢定，沒過那一手變成怯戰（RE-135）。CYCLOPS、CLIFF TROLL、SLEDGE
	// 與 MIST GRUB 四隻是這一種。
	Fearsome bool `json:"fearsome"`
	// Gold 是打贏之後掉多少金冠的骰式（記錄 +0x0F，RE-091／RE-096）：
	// 高 nibble 的低三位是骰數，低三位查骰面表，擲「骰數 ＋ 1」次。
	// 原版只讀第一隻怪的骰式，同場其他怪的這一格沒有讀。
	Gold struct {
		Raw   int `json:"raw"`
		Rolls int `json:"rolls"`
		Sides int `json:"sides"`
	} `json:"gold"`
	// Weapon 是手上那件武器（記錄 +0x20）；ArmourWeight 是三件護甲的重量和；
	// Shield 是盾牌（記錄 +0x60），沒有就是 nil。
	Weapon       *Item `json:"weapon"`
	ArmourWeight int   `json:"armour_weight"`
	Shield       *Item `json:"shield"`
	// Worn 是記錄 `+0x30` 起那七格穿戴欄，順序與角色記錄的十格相同
	// （HEAD／TORSO／LEGS／BODY／HANDS／FEET／SHIELD，RE-131）。徒手四式拿的
	// 就是其中四格；射擊那一手拿的是 HANDS。
	Worn []*Item `json:"worn"`
	// Encounter 是遇到這一隻時印的四行敘述（記錄 +0x1C0 起那四格，RE-112）。
	// 原版只在**問要不要騎走那一頁**印它——全隊都有坐騎才問得出那一句，
	// 有人用走的就整段跳過，那四行也就看不到（`sub_2A4E5:0x2A667`）。
	Encounter []string `json:"encounter"`
}

// Item 是一格裝備，欄位名與商店的商品記錄相同（RE-075）。
type Item struct {
	Price       int    `json:"price"`
	Kind        int    `json:"kind"`
	Damage      bool   `json:"damage"`
	Position    int    `json:"position"`
	Condition   int    `json:"condition"`
	Percent     int    `json:"percent"`
	Name        string `json:"name"`
	Class       int    `json:"class"`
	Hands       int    `json:"hands"`
	Encumbrance int    `json:"encumbrance"`
	SizeClass   int    `json:"size_class"`
	Dice        struct {
		Count int `json:"count"`
		Sides int `json:"sides"`
		Bonus int `json:"bonus"`
		Low   int `json:"low"`
		High  int `json:"high"`
	} `json:"dice"`
	Record string `json:"record"`
}

// CombatWeapon 把一格裝備換成戰鬥要用的武器數字。骰面數用的是戰鬥那張表
// （`0x00D4`），與商店顯示的那一張差一（RE-080）。
func (i *Item) CombatWeapon() Weapon {
	if i == nil {
		return Weapon{}
	}
	w := Weapon{Kind: i.Class, Base: i.Dice.Bonus,
		Weight: i.Encumbrance, TwoHand: i.Hands == 1}
	w.Count, w.Sides = DamageCode(i.damageCode())
	return w
}

// damageCode 從記錄的 hex 取回 +0x0C 的傷害碼。
func (i *Item) damageCode() int {
	if i == nil || len(i.Record) < 26 {
		return 0
	}
	var v int
	for _, c := range i.Record[24:26] {
		v <<= 4
		switch {
		case c >= '0' && c <= '9':
			v |= int(c - '0')
		case c >= 'a' && c <= 'f':
			v |= int(c-'a') + 10
		case c >= 'A' && c <= 'F':
			v |= int(c-'A') + 10
		default:
			return 0
		}
	}
	return v
}

// MonsterCount 是怪物的種數（`TOWN0` 的 payload 5…36）。挑怪物是 `rand(0x20) − 1`。
const MonsterCount = 32

// SeasonChance 是一個季節的遭遇機率：白天與夜裡各八個地形。值要**除以 2** 才是
// 百分比（`0x2A541`）。
type SeasonChance struct {
	Season string `json:"season"`
	Day    []int  `json:"day"`
	Night  []int  `json:"night"`
}

// Monsters 是抽取器產生的怪物表。
type Monsters struct {
	Schema string         `json:"schema"`
	List   []Monster      `json:"monsters"`
	Chance []SeasonChance `json:"encounter_chance"`
}

// EncounterChance 取某個季節、地形、時段的機率原值（還沒除以 2）。
// 表缺了就回 0，也就是不會遇到。
func (m *Monsters) EncounterChance(season, terrain int, night bool) int {
	if m == nil || season < 0 || season >= len(m.Chance) {
		return 0
	}
	row := m.Chance[season].Day
	if night {
		row = m.Chance[season].Night
	}
	if terrain < 0 || terrain >= len(row) {
		return 0
	}
	return row[terrain]
}

// LoadMonsters 讀取怪物數值表。
func LoadMonsters(path string) (*Monsters, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("讀取怪物資料：%w", err)
	}
	var m Monsters
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, fmt.Errorf("解析怪物資料：%w", err)
	}
	if m.Schema != "kol-monster-data-v1" {
		return nil, fmt.Errorf("怪物資料 schema 不符：%q", m.Schema)
	}
	var fields struct { List []map[string]json.RawMessage `json:"monsters"` }
	if err := json.Unmarshal(body, &fields); err != nil { return nil, err }
	for i, mon := range m.List {
		if len(fields.List[i]["ranged_chance"]) == 0 || len(fields.List[i]["ranged_code"]) == 0 ||
			string(fields.List[i]["ranged_chance"]) == "null" || string(fields.List[i]["ranged_code"]) == "null" ||
			mon.RangedChance < 0 || mon.RangedChance > 255 || mon.RangedCode < 0 || mon.RangedCode > 255 {
			return nil, fmt.Errorf("怪物 %d 缺少合法遠程初始化欄位，請重新抽取資料", mon.Index)
		}
	}
	return &m, nil
}

// At 回傳某個編號的怪物；沒有就回零值。
func (m *Monsters) At(index int) Monster {
	for _, mon := range m.List {
		if mon.Index == index {
			return mon
		}
	}
	return Monster{}
}
