package save

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func characterSample() *Characters {
	c := &Character{ID: "test-olear", Name: "Olear", Gold: 0, Sheet: &CharacterSheet{
		Stats: [7]int{77, 0, 81, 63, 59, 52, 44}, Ward: 14, AdventurePoints: 0,
		Weapons: []CharacterWeapon{{Row: 2, Name: "BOW", OffenceRaw: 123, DefenceRaw: 45}},
	}, Backpack: []Item{}, Spells: []CharacterSpell{}}
	for i := range c.Carried {
		c.Carried[i] = strings.Repeat("00", 16)
	}
	c.Worn[9] = &Item{Name: "BELT", Condition: 1, Record: strings.Repeat("01", 16), Dice: ItemDice{Bonus: -1}}
	c.InHand = &Item{Name: "BOW", Condition: 19, Class: 7, Record: strings.Repeat("02", 16)}
	c.Progress[4] = 3
	c.Quests[5] = 128
	s := &Characters{Party: []int{7}, Speaker: 0}
	s.Characters[7] = c
	return s
}

func TestCharactersRoundTripPreservesChangedAndEmptyState(t *testing.T) {
	want := characterSample()
	raw, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	got, err := DecodeCharacters(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("角色往返不同：%s", raw)
	}
	got.Characters[7].Sheet.Weapons[0].OffenceRaw = 0
	got.Characters[7].InHand = nil
	got.Characters[7].Worn[9] = nil
	got.Characters[7].Carried[0] = strings.Repeat("00", 16)
	raw, err = json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	again, err := DecodeCharacters(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, again) {
		t.Fatal("零值或空裝備被補回")
	}
	if want.Characters[7].Sheet.Weapons[0].OffenceRaw != 123 {
		t.Fatal("重載資料與來源共用記憶體")
	}
	// 離隊角色留在原槽；刪除才將該槽清空，不把其他角色向前搬。
	again.Party = nil
	again.Speaker = -1
	if err := again.Validate(); err != nil {
		t.Fatal(err)
	}
	again.Characters[7] = nil
	if err := again.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestCharactersRejectIncompleteOrInvalidData(t *testing.T) {
	cases := map[string]func(map[string]any){
		"少槽":    func(m map[string]any) { m["characters"] = m["characters"].([]any)[:15] },
		"多槽":    func(m map[string]any) { m["characters"] = append(m["characters"].([]any), nil) },
		"空槽參照":  func(m map[string]any) { m["party"] = []int{0} },
		"重複參照":  func(m map[string]any) { m["party"] = []int{7, 7} },
		"缺少隊伍":  func(m map[string]any) { delete(m, "party") },
		"發言者空值": func(m map[string]any) { m["speaker"] = nil },
		"未知欄位":  func(m map[string]any) { m["unknown"] = 1 },
		"缺少空手欄": func(m map[string]any) { delete(characterMap(m), "in_hand") },
		"少穿戴格":  func(m map[string]any) { characterMap(m)["worn"] = []any{} },
		"空金冠":   func(m map[string]any) { characterMap(m)["gold"] = nil },
		"缺少資料頁": func(m map[string]any) { characterMap(m)["sheet"] = nil },
		"短原始記錄": func(m map[string]any) { characterMap(m)["carried"].([]any)[0] = "00" },
		"七格背包": func(m map[string]any) {
			c := characterMap(m)
			c["backpack"] = []any{c["in_hand"], c["in_hand"], c["in_hand"], c["in_hand"], c["in_hand"], c["in_hand"], c["in_hand"]}
		},
		"少屬性": func(m map[string]any) { characterMap(m)["sheet"].(map[string]any)["stats"] = []int{1} },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			raw, err := json.Marshal(characterSample())
			if err != nil {
				t.Fatal(err)
			}
			var m map[string]any
			if err = json.Unmarshal(raw, &m); err != nil {
				t.Fatal(err)
			}
			mutate(m)
			raw, err = json.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			if _, err = DecodeCharacters(raw); err == nil {
				t.Fatal("應拒絕不完整或無效資料")
			}
		})
	}
}

func characterMap(m map[string]any) map[string]any {
	return m["characters"].([]any)[7].(map[string]any)
}
