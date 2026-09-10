package battle

import (
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
			body := `{"schema":"kol-monster-data-v1","monsters":[{` + tc.fields + `}]}`
			if err := os.WriteFile(path, []byte(body), 0600); err != nil {
				t.Fatal(err)
			}
			got, err := LoadMonsters(path)
			if tc.valid {
				if err != nil || len(got.List) != 1 {
					t.Fatalf("合法怪物資料載入失敗：%v", err)
				}
			} else if err == nil {
				t.Fatal("損壞或過期資料應拒絕載入")
			}
		})
	}
}
