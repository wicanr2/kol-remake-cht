// Package chardata 讀取由 tools/extract_character_data.py 抽出的角色建立 typed data。
//
// 規則層只吃這份 JSON，執行期不解析原版 `BOOKDATA`。骰式、內部基值、名稱順序與
// selector 分組都來自已證實的原版契約。
package chardata

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Data struct {
	Schema            string             `json:"schema"`
	DiceSides         []int              `json:"dice_sides"`
	PrimaryAttributes []string           `json:"primary_attributes"`
	DerivedAttributes []Derived          `json:"derived_attributes"`
	DisplayScale      DisplayScale       `json:"display_scale"`
	Races             []Race             `json:"races"`
	Secondary         []Named            `json:"secondary"`
	ProfessionGroups  []ProfessionGroup  `json:"profession_groups"`
}

type Derived struct {
	Name string   `json:"name"`
	From []string `json:"from"`
}

type DisplayScale struct {
	Numerator   int `json:"numerator"`
	Denominator int `json:"denominator"`
}

type Named struct {
	Value int    `json:"value"`
	Name  string `json:"name"`
}

type Race struct {
	Value        int    `json:"value"`
	Name         string `json:"name"`
	HasSecondary bool   `json:"has_secondary"`
}

type ProfessionGroup struct {
	Race        int          `json:"race"`
	Secondary   *int         `json:"secondary"`
	Professions []Profession `json:"professions"`
}

type Profession struct {
	RecordIndex   int    `json:"record_index"`
	Name          string `json:"name"`
	DiceCodes     []int  `json:"dice_codes"`
	InternalBases []int  `json:"internal_bases"`
}

// Load 讀取 typed data 並做基本一致性檢查。
func Load(path string) (*Data, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("讀取角色資料失敗：%w", err)
	}
	var d Data
	if err := json.Unmarshal(raw, &d); err != nil {
		return nil, fmt.Errorf("解析角色資料失敗：%w", err)
	}
	if !strings.HasPrefix(d.Schema, "kol-character-creation-data-") {
		return nil, fmt.Errorf("非預期的角色資料 schema：%s", d.Schema)
	}
	if len(d.PrimaryAttributes) != 7 || len(d.DiceSides) != 8 {
		return nil, fmt.Errorf("角色資料的屬性或骰面數量不符")
	}
	total := 0
	for _, g := range d.ProfessionGroups {
		total += len(g.Professions)
		for _, p := range g.Professions {
			if len(p.DiceCodes) != 7 || len(p.InternalBases) != 7 {
				return nil, fmt.Errorf("職業 %s 的骰式或基值數量不符", p.Name)
			}
		}
	}
	if total != 39 {
		return nil, fmt.Errorf("職業記錄總數為 %d，不是 39", total)
	}
	return &d, nil
}

// Professions 回傳指定種族與 secondary 的職業清單。
// secondary 對沒有 secondary 的種族無作用。
func (d *Data) Professions(race, secondary int) []Profession {
	for _, g := range d.ProfessionGroups {
		if g.Race != race {
			continue
		}
		if g.Secondary == nil || *g.Secondary == secondary {
			return g.Professions
		}
	}
	return nil
}

// HasSecondary 回傳該種族是否需要選 secondary（性別）。
func (d *Data) HasSecondary(race int) bool {
	for _, r := range d.Races {
		if r.Value == race {
			return r.HasSecondary
		}
	}
	return false
}
