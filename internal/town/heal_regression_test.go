package town

import "testing"

func TestAbbeyClearsWoundsAndNutrition(t *testing.T) {
	s := abbeySession(t, 10, Member{Name: "AUDIT", Gold: 100, Speaker: true, Wounds: 2, Sheet: &Sheet{Nutrition: 37}})
	s.StartRest()
	yes(s)
	if s.Party[0].Wounds != 0 || s.Party[0].Sheet.Nutrition != 0 || s.Party[0].Gold != 80 {
		t.Fatalf("治療結果不符 SPEC-006：%+v", s.Party[0])
	}
}
