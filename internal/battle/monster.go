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
