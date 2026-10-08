package battle

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// 舊資料包若缺少遠程欄位，不能被解讀成所有怪物都不能射擊。
func TestLoadMonsterRangedContract(t *testing.T) {
	for _, tc := range []struct {
		name   string
		fields string
		valid  bool
	}{
		{"舊資料缺兩欄", ``, false},
		{"缺機率", `"ranged_code":0`, false},
		{"缺骰碼", `"ranged_chance":0`, false},
		{"空機率", `"ranged_chance":null,"ranged_code":0`, false},
		{"空骰碼", `"ranged_chance":0,"ranged_code":null`, false},
		{"負機率", `"ranged_chance":-1,"ranged_code":0`, false},
		{"超界機率", `"ranged_chance":256,"ranged_code":0`, false},
		{"負骰碼", `"ranged_chance":0,"ranged_code":-1`, false},
		{"超界骰碼", `"ranged_chance":0,"ranged_code":256`, false},
		{"字串機率", `"ranged_chance":"0","ranged_code":0`, false},
		{"小數骰碼", `"ranged_chance":0,"ranged_code":1.5`, false},
		{"合法零值", `"ranged_chance":0,"ranged_code":0`, true},
		{"合法位元組上限", `"ranged_chance":255,"ranged_code":255`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "monsters.json")
			fields := `"loot_chance":20`
			if tc.fields != "" {
				fields += "," + tc.fields
			}
			body := monsterLoadFixture(t, fields)
			if err := os.WriteFile(path, []byte(body), 0600); err != nil {
				t.Fatal(err)
			}
			got, err := LoadMonsters(path)
			if tc.valid {
				if err != nil || len(got.List) != MonsterCount {
					t.Fatalf("合法怪物資料載入失敗：%v", err)
				}
			} else if err == nil {
				t.Fatal("損壞或過期資料應拒絕載入")
			}
		})
	}
}

func TestLoadMonsterLootContract(t *testing.T) {
	for _, tc := range []struct {
		fields string
		valid  bool
	}{
		{``, false}, {`"loot_chance":null`, false}, {`"loot_chance":-1`, false},
		{`"loot_chance":256`, false}, {`"loot_chance":"20"`, false},
		{`"loot_chance":0`, true}, {`"loot_chance":20`, true}, {`"loot_chance":255`, true},
	} {
		path := filepath.Join(t.TempDir(), "monsters.json")
		fields := `"ranged_chance":0,"ranged_code":0`
		if tc.fields != "" {
			fields += "," + tc.fields
		}
		body := monsterLoadFixture(t, fields)
		if err := os.WriteFile(path, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
		_, err := LoadMonsters(path)
		if (err == nil) != tc.valid {
			t.Fatalf("搜刮資料%s：%v", tc.fields, err)
		}
	}
}

func monsterLoadFixture(t *testing.T, fields string) string {
	t.Helper()
	body, err := os.ReadFile("../../assets/original-derived-data/monsters.json")
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err = json.Unmarshal(body, &doc); err != nil {
		t.Fatal(err)
	}
	first := doc["monsters"].([]any)[0].(map[string]any)
	delete(first, "ranged_chance")
	delete(first, "ranged_code")
	delete(first, "loot_chance")
	var supplied map[string]any
	if err = json.Unmarshal([]byte("{"+fields+"}"), &supplied); err != nil {
		t.Fatal(err)
	}
	for k, v := range supplied {
		first[k] = v
	}
	out, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}
