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
	Index     int    `json:"index"`
	Name      string `json:"name"`
	Strength  int    `json:"strength"`
	Quickness int    `json:"quickness"`
	Foresight int    `json:"foresight"`
	Size      int    `json:"size"`
	Health    int    `json:"health"`
	Intellect int    `json:"intellect"`
	// AttackSkill／ParrySkill 是記錄 +0x17 起的兩個大端 16-bit；戰鬥再乘等級除以 5。
	AttackSkill int `json:"attack_skill"`
	ParrySkill  int `json:"parry_skill"`
	// Gold 是打贏之後這一隻怪會掉多少金冠的骰式（記錄 +0x0F，RE-091）：
	// 高 nibble 的低三位是骰數，低三位查骰面表。每一隻擲「骰數 ＋ 1」次。
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
}

// Item 是一格裝備，欄位名與商店的商品記錄相同（RE-075）。
type Item struct {
	Name        string `json:"name"`
	Class       int    `json:"class"`
	Hands       int    `json:"hands"`
	Encumbrance int    `json:"encumbrance"`
	SizeClass   int    `json:"size_class"`
	Dice        struct {
		Count int `json:"count"`
		Sides int `json:"sides"`
		Bonus int `json:"bonus"`
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

// Monsters 是抽取器產生的怪物表。
type Monsters struct {
	Schema string    `json:"schema"`
	List   []Monster `json:"monsters"`
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
