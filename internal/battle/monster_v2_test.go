package battle

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestMonsterV2RejectsMalformedEquipment(t *testing.T) {
	source, err := os.ReadFile("../../assets/original-derived-data/monsters.json")
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"舊schema", func(d map[string]any) { d["schema"] = "kol-monster-data-v1" }},
		{"缺怪物", func(d map[string]any) { d["monsters"] = d["monsters"].([]any)[:31] }},
		{"重複索引", func(d map[string]any) { d["monsters"].([]any)[1].(map[string]any)["index"] = 0 }},
		{"缺裝備組", func(d map[string]any) {
			m := d["monsters"].([]any)[0].(map[string]any)
			m["equipment_variants"] = m["equipment_variants"].([]any)[:3]
		}},
		{"多裝備組", func(d map[string]any) {
			m := d["monsters"].([]any)[0].(map[string]any)
			v := m["equipment_variants"].([]any)
			m["equipment_variants"] = append(v, v[0])
		}},
		{"錯位移", func(d map[string]any) { firstMonsterGear(d)["source_offset"] = 33 }},
		{"缺物品格", func(d map[string]any) { g := firstMonsterGear(d); g["items"] = g["items"].([]any)[:5] }},
		{"缺來源格", func(d map[string]any) { g := firstMonsterGear(d); g["records"] = g["records"].([]any)[:5] }},
		{"非法來源hex", func(d map[string]any) { firstMonsterGear(d)["records"].([]any)[0] = "xx" }},
		{"有來源無物品", func(d map[string]any) { firstMonsterGear(d)["items"].([]any)[0] = nil }},
		{"非法hex", func(d map[string]any) { firstMonsterItem(d)["record"] = "xx" }},
		{"截短record", func(d map[string]any) { firstMonsterItem(d)["record"] = "0011" }},
		{"非null空格", func(d map[string]any) { firstMonsterItem(d)["record"] = "00000000000000000000000000000000" }},
		{"typed負重錯誤", func(d map[string]any) { i := firstMonsterItem(d); i["encumbrance"] = i["encumbrance"].(float64) + 1 }},
		{"typed武器碼錯誤", func(d map[string]any) { i := firstMonsterItem(d); i["class"] = i["class"].(float64) + 1 }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var d map[string]any
			if err := json.Unmarshal(source, &d); err != nil {
				t.Fatal(err)
			}
			tc.mutate(d)
			body, err := json.Marshal(d)
			if err != nil {
				t.Fatal(err)
			}
			p := filepath.Join(t.TempDir(), "monsters.json")
			if err = os.WriteFile(p, body, 0600); err != nil {
				t.Fatal(err)
			}
			if _, err = LoadMonsters(p); err == nil {
				t.Fatal("損壞v2資料應拒絕載入")
			}
		})
	}
}

func TestMonsterV2PaddedEmptyShield(t *testing.T) {
	mons, err := LoadMonsters("../../assets/original-derived-data/monsters.json")
	if err != nil {
		t.Fatal(err)
	}
	mon := mons.At(2)
	gear := mon.EquipmentVariants[2]
	if gear.Records[4] != "00000000008000000000000000000000" || gear.Items[4] != nil {
		t.Fatal("第一byte為零的盾格應保存填充並使用null")
	}
	s := NewQuest(&Site{Index: -1, Level: 1, Count: 1, MonsterName: mon.Name, MonsterX: []int{35}, MonsterY: []int{32}}, nil, mon, nil)
	s.initializeMonsterEquipment(&fixedRand{values: []int{3}})
	if a := s.Actors[6]; a.MonsterShield != nil || a.ShieldGrade != NoShield || a.ShieldWeight != 0 {
		t.Fatal("非零填充產生不存在的盾牌")
	}
}

func firstMonsterGear(d map[string]any) map[string]any {
	return d["monsters"].([]any)[0].(map[string]any)["equipment_variants"].([]any)[0].(map[string]any)
}
func firstMonsterItem(d map[string]any) map[string]any {
	for _, m := range d["monsters"].([]any) {
		for _, v := range m.(map[string]any)["equipment_variants"].([]any) {
			for _, it := range v.(map[string]any)["items"].([]any) {
				if it != nil {
					return it.(map[string]any)
				}
			}
		}
	}
	panic("真實v2資料沒有非空物品")
}

func TestMonsterV2EquipmentIsPerActorAndInitializedOnce(t *testing.T) {
	mons, err := LoadMonsters("../../assets/original-derived-data/monsters.json")
	if err != nil {
		t.Fatal(err)
	}
	mon := mons.At(9)
	s := NewQuest(&Site{Index: -1, Level: 1, Count: 2, MonsterName: mon.Name, MonsterX: []int{35, 37}, MonsterY: []int{32, 29}}, nil, mon, nil)
	s.initializeMonsterEquipment(&fixedRand{values: []int{3, 1}})
	if s.Actors[6].HandItem.Name != "Troll Maul" || s.Actors[7].HandItem.Name != "Long Spear" {
		t.Fatal("原版三巨魔樣本前兩組選擇不符")
	}
	before := *s.Actors[6].Armour[0]
	s.Actors[6].Armour[0].Condition--
	if *mon.EquipmentVariants[2].Items[1] != before {
		t.Fatal("可變耐久污染母表")
	}
	s.initializeMonsterEquipment(&fixedRand{values: nil})
	if s.Actors[6].Armour[0].Condition != before.Condition-1 {
		t.Fatal("重複初始化還原耐久")
	}
	s.Plan.BodyActor = 7
	s.Stage = StageField
	d, ok := s.EquipmentPanel()
	if !ok || d.Rows[0] != "Troll Maul" {
		t.Fatal("裝備面板讀到共用母表")
	}
	// DropItem只支援隊員；此處驗證怪物現行手持清空後的顯示，不宣稱新增掉落規則。
	s.Actors[6].HandItem = nil
	s.Actors[6].Unarmed = true
	d, ok = s.EquipmentPanel()
	if !ok || d.Rows[0] != "" {
		t.Fatal("丟下武器後面板回退母表")
	}
}
