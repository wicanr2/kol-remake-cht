package battle

import "testing"

func TestInitialIntellectMatchesRestoredStats(t *testing.T) {
	party := []Member{{Name: "Poontang", Intellect: 23}}
	monster := testMonster()
	monster.Intellect = 61
	s := NewQuest(testSite(), party, monster, NewField())
	for _, c := range []struct{ slot, want int }{{0, 23}, {PartySlots, 61}} {
		a := &s.Actors[c.slot]
		if a.Intellect != c.want {
			t.Fatalf("slot=%d 初始智力=%d，預期=%d", c.slot, a.Intellect, c.want)
		}
		// 未受法術時的判定，必須與恢復原始屬性後一致。
		before := EnemyIntellectCheck(a.Intellect, 3, c.want-1)
		a.SpellTurns = 1
		if !s.WearsOff(c.slot) {
			t.Fatal("法術到期前置無效")
		}
		if a.Intellect != c.want || EnemyIntellectCheck(a.Intellect, 3, c.want-1) != before || !before {
			t.Fatalf("slot=%d 初始與恢復後智力判定不一致", c.slot)
		}
	}
}
