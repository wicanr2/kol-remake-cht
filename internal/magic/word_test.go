package magic

import (
	"encoding/json"
	"os"
	"testing"
)

func wordTable(t *testing.T) *WordTable {
	t.Helper()
	w, err := LoadWordTable("../../docs/i18n/spell-words-zh-Hant.json")
	if err != nil {
		t.Fatal(err)
	}
	return w
}

// 說明書印刷頁 74–75 的六個範例：每個都要唯一解析，種族與屬性照範例說明。
func TestParseManualExamples(t *testing.T) {
	w := wordTable(t)
	for name, want := range map[string][3]int{
		"TYANAWONTA":   {3, 6, 0},
		"ARNALYRTA":    {2, 6, 0},
		"KUMTWEYYRFE":  {6, 5, 2},
		"KELDAYYRITA":  {4, 3, 0},
		"VORVORWYRAMI": {8, 4, 4},
		"KUMNAFYRKE":   {6, 6, 9},
	} {
		got, ok := w.Parse(name)
		if !ok || [3]int{got.Class, got.Stat, got.Sub} != want {
			t.Errorf("%s 解析為 %+v（ok=%v），應為 %v", name, got, ok, want)
		}
	}
	if got, _ := w.Parse("KELDAYYRITA"); got.Gloss("") != "凱爾登・先見・強大・遠距離・持續中" {
		t.Errorf("注解是 %q", got.Gloss(""))
	}
}

// 城鎮商店賣的法術：名字解析出的種族／屬性／亞種要與法術記錄一致
// （`guards[0]` 高低 nibble 是 class／subclass、`effect` 是 stat，RE-121／128）。
func TestParseShopSpellsAgreesWithRecords(t *testing.T) {
	w := wordTable(t)
	body, err := os.ReadFile("../../assets/original-derived-data/town-buildings.json")
	if err != nil {
		t.Skip("沒有原版衍生資料：", err)
	}
	var doc struct {
		Towns []struct {
			Doors []struct {
				Spells []struct {
					Name   string `json:"name"`
					Effect int    `json:"effect"`
					Guards []int  `json:"guards"`
				} `json:"spells"`
			} `json:"doors"`
		} `json:"towns"`
	}
	if err := json.Unmarshal(body, &doc); err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, town := range doc.Towns {
		for _, door := range town.Doors {
			for _, s := range door.Spells {
				n++
				got, ok := w.Parse(s.Name)
				want := [3]int{s.Guards[0] >> 4, s.Effect, s.Guards[0] & 0x0F}
				if !ok || [3]int{got.Class, got.Stat, got.Sub} != want {
					t.Errorf("%s 解析為 %+v（ok=%v），記錄是 %v", s.Name, got, ok, want)
				}
			}
		}
	}
	if n != 48 {
		t.Fatalf("商店法術共 %d 筆，預期 48", n)
	}
}

// 玩家自取的名字拆不開就不加注解。
func TestParseRejectsPlayerNames(t *testing.T) {
	w := wordTable(t)
	for _, name := range []string{"Orc Killer", "Elf Restorer", "", "TYANA"} {
		if _, ok := w.Parse(name); ok {
			t.Errorf("%q 不應解析成功", name)
		}
	}
}
