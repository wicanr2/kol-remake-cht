package town

import "testing"

// 答應治療之後全隊的傷勢與飢餓都歸零，空位不算。
func TestHealPartyClearsWoundsAndHunger(t *testing.T) {
	party := []Member{
		{Name: "ddduhh", Wounds: 30, Sheet: &Sheet{Nutrition: 200}},
		{Name: "", Wounds: 9, Sheet: &Sheet{Nutrition: 9}},
		{Name: "Warder", Wounds: 20, Sheet: &Sheet{Nutrition: 128}},
	}
	if got := HealParty(party); got != 2 {
		t.Fatalf("治好 %d 個，應該是 2", got)
	}
	for _, i := range []int{0, 2} {
		if party[i].Wounds != 0 || party[i].Sheet.Nutrition != 0 {
			t.Fatalf("第 %d 個沒清乾淨：%d／%d", i,
				party[i].Wounds, party[i].Sheet.Nutrition)
		}
	}
	// 空位原樣不動。
	if party[1].Wounds != 9 || party[1].Sheet.Nutrition != 9 {
		t.Fatalf("空位被動到了：%+v", party[1])
	}
}
