package town

import "testing"

func TestStrengthenTargetsAndBounds(t *testing.T) {
	a := Member{ID: "a", Name: "A", Gold: 9500, Sheet: &Sheet{Weapons: []WeaponRow{{Offence: 230, OffenceRaw: 2300, Defence: 4, DefenceRaw: 47}}}}
	b := Member{ID: "b", Name: "B", Speaker: true, Sheet: &Sheet{Weapons: []WeaponRow{{Offence: 8, Defence: 9}}}}
	s := &Session{Party: []Member{a, b}, Roster: []Candidate{{ID: "a", Name: "A", Full: &a}, {ID: "b", Name: "B", Full: &b}}, Inside: &Inn{}}
	if n, e := s.Strengthen(false, "skills", 200); e != nil || n != 1 {
		t.Fatal(n, e)
	}
	if s.Party[0].Sheet.Weapons[0].Defence != 4 || s.Party[1].Sheet.Weapons[0].OffenceRaw != 2000 {
		t.Fatal("目標或原始技能值錯誤")
	}
	if s.Roster[1].Full.Sheet.Weapons[0].Defence != 200 {
		t.Fatal("名冊未同步")
	}
	if _, e := s.Strengthen(true, "gold", 9000); e != nil || s.Party[0].Gold != 9500 || s.Party[1].Gold != 9000 {
		t.Fatal("補足金幣不正確")
	}
	if _, e := s.Strengthen(true, "stats", 170); e != nil {
		t.Fatal(e)
	}
	for _, m := range s.Party {
		for _, v := range m.Sheet.Stats {
			if v != 255 {
				t.Fatal("屬性超界或換算錯誤")
			}
		}
	}
	for _, v := range []struct {
		k string
		n int
	}{{"stats", 171}, {"skills", 201}, {"skills", 0}, {"gold", 16777216}, {"gold", -1}} {
		if _, e := s.Strengthen(true, v.k, v.n); e == nil {
			t.Fatal("接受非法數值", v)
		}
	}
	s.Party[1].Sheet = nil
	s.Party[0].Sheet.Stats[0] = 1
	if _, e := s.Strengthen(true, "stats", 100); e == nil || s.Party[0].Sheet.Stats[0] != 1 {
		t.Fatal("驗證失敗仍部分改值")
	}
}
