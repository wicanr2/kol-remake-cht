package battle

import "testing"

type rangedCountRand struct{ calls int }

func (r *rangedCountRand) Roll(int) int { r.calls++; return 1 }

func TestMonsterRangedInitializeOnce(t *testing.T) {
	for _, chance := range []int{0, 255} {
		site := &Site{Index: -1, Count: 2, Monster: 27, MonsterX: []int{10, 12}, MonsterY: []int{10, 12}}
		s := NewQuest(site, nil, Monster{RangedChance: chance}, nil)
		r := &rangedCountRand{}
		s.InitializeMonsterRanged(r)
		want, calls := 0, 2
		if chance == 255 {
			want, calls = 1, 4
		}
		for _, i := range []int{6, 7} {
			if s.Actors[i].Kind != want {
				t.Fatalf("槽 %d 能力 %d，預期 %d", i, s.Actors[i].Kind, want)
			}
		}
		if r.calls != calls {
			t.Fatalf("初始化消耗 %d，預期 %d", r.calls, calls)
		}
		s.Actors[6].Kind = 77
		s.InitializeMonsterRanged(r)
		if r.calls != calls || s.Actors[6].Kind != 77 {
			t.Fatal("重複初始化重擲或覆寫能力")
		}
	}
}

func TestInitializedAbilityControlsEnemyShooting(t *testing.T) {
	for _, chance := range []int{0, 255} {
		site := &Site{Index: -1, Monster: 27, Level: 3, Count: 1, PartyX: []int{30}, PartyY: []int{32}, MonsterX: []int{30}, MonsterY: []int{30}}
		monster := Monster{Index: 27, Strength: 100, Health: 100, Size: 100, FleeAt: 100, RangedChance: chance, Weapon: &Item{Name: "WEAPON", Class: 0x21}}
		s := NewQuest(site, []Member{{Name: "TARGET", Strength: 100, Health: 100, Size: 100}}, monster, nil)
		r := &rangedCountRand{}
		s.InitializeMonsterRanged(r)
		d := s.PlanEnemy(6, r)
		if (d.Command == CmdMissile) != (chance == 255) {
			t.Fatalf("機率 %d、能力 %d：AI 指令 %x", chance, s.Actors[6].Kind, d.Command)
		}
	}
}

func TestFirstPlanningInitializesMonsterAbility(t *testing.T) {
	party := []Member{{Name: "PLAYER", Strength: 100, Size: 100, Health: 100}}
	monster := Monster{Index: 27, Name: "SLEDGE", Strength: 150, Size: 148, Health: 100}
	quest := NewQuest(&Site{Index: -1, Monster: 27, Count: 1, PartyX: []int{20}, PartyY: []int{20}, MonsterX: []int{30}, MonsterY: []int{30}}, party, monster, nil)
	arena := NewArena(0, 27, 3, monster, party, nil)
	for _, s := range []*State{quest, arena} {
		r := &rangedCountRand{}
		if !s.rangedPending {
			t.Fatal("建構時未設定待初始化")
		}
		s.BeginPlanning(r)
		if s.rangedPending || s.Actors[6].Kind != 0 {
			t.Fatal("首次規劃未依原版初始化 SLEDGE")
		}
		s.Actors[6].Kind = 77
		s.BeginPlanning(r)
		if s.Actors[6].Kind != 77 {
			t.Fatal("後續規劃重擲能力")
		}
	}
}
