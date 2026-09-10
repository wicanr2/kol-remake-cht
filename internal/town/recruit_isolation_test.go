package town

import "testing"

func TestJoinedCharacterDoesNotMutateRoster(t *testing.T) {
	for _, recruit := range []bool{false, true} {
		s := recruitSession(t, 1)
		c := &s.Roster[0]
		c.Sheet = &Sheet{Nutrition: 17, Weapons: []WeaponRow{{OffenceRaw: 31}}}
		c.InHand = &Good{Condition: 42}
		c.Worn[0] = &Good{Condition: 43}
		c.Backpack = []Good{{Condition: 44}}
		c.Quests, c.Progress = []int{1}, []int{2}
		c.Carried = [][]byte{{3}}
		c.Spells = []Spell{{Guards: []int{4}}}
		if recruit {
			s.StartRecruit()
			s.Accept()
			s.Accept()
		} else if !s.SelectLeader(c.Name) {
			t.Fatal("無法選角")
		}
		m := &s.Party[len(s.Party)-1]
		m.Sheet.Nutrition = 99
		m.Sheet.Weapons[0].OffenceRaw = 99
		m.InHand.Condition, m.Worn[0].Condition, m.Backpack[0].Condition = 99, 99, 99
		m.Quests[0], m.Progress[0], m.Carried[0][0], m.Spells[0].Guards[0] = 99, 99, 99, 99
		if c.Sheet.Nutrition != 17 || c.Sheet.Weapons[0].OffenceRaw != 31 ||
			c.InHand.Condition != 42 || c.Worn[0].Condition != 43 || c.Backpack[0].Condition != 44 ||
			c.Quests[0] != 1 || c.Progress[0] != 2 || c.Carried[0][0] != 3 || c.Spells[0].Guards[0] != 4 {
			t.Fatalf("招募=%v：未保存的隊員修改污染名冊", recruit)
		}
		if len(s.Outdoor.Party) != len(s.Party) || len(s.Inside.Party) != len(s.Party) {
			t.Fatalf("招募=%v：畫面隊伍未同步", recruit)
		}
	}
}
