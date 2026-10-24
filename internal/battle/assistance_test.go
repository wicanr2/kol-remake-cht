package battle

import (
	"reflect"
	"testing"
)

func TestAssistanceFatigueRules(t *testing.T) {
	base := Actor{Side: SideMonster, Head: 10, Chest: 10, Stamina: 200, Threshold: 100, Damage: 50, Hurt: 1, Fatigue: 20, Command: CmdRest}
	for _, mult := range []int{0, 1, 2, 3, 6, 99} {
		s := &State{Actors: []Actor{base}, Assistance: Assistance{EnemyFatigue: mult}}
		s.EndTurnFatigue(0)
		want := 20 + 12*s.Assistance.Multiplier()
		if s.Actors[0].Fatigue != want {
			t.Fatalf("mult%d: got%d want%d", mult, s.Actors[0].Fatigue, want)
		}
	}
	for _, side := range []Side{SideParty, SideMonster} {
		a := base
		a.Side = side
		a.Hurt = 0
		a.Fatigue = 100
		s := &State{Actors: []Actor{a}, Assistance: Assistance{EnemyFatigue: 6}}
		s.EndTurnFatigue(0)
		if s.Actors[0].Fatigue != 70 {
			t.Fatal("倍率改變了休息回復")
		}
	}
	base.Side = SideParty
	base.Fatigue = 500
	s := &State{Actors: []Actor{base}, Assistance: Assistance{NoFatigue: true}}
	if s.EndTurnFatigue(0) != 0 || s.Actors[0].Fatigue != 0 || s.Actors[0].Down() {
		t.Fatal("玩家仍因疲勞倒下")
	}
	s.Actors[0].Head = 0
	s.ApplyAssistance()
	if s.Actors[0].Head != 0 {
		t.Fatal("輔助復活了角色")
	}
	// 零值不得動角色，也不得在回復路徑改成新的下限。
	s.Assistance = Assistance{}
	s.Actors[0].Fatigue = -5
	before := s.Actors[0]
	s.ApplyAssistance()
	if !reflect.DeepEqual(before, s.Actors[0]) {
		t.Fatal("關閉仍改變狀態")
	}
}

func TestAssistanceFearAndSpell(t *testing.T) {
	for _, immune := range []bool{false, true} {
		s := &State{Monster: Monster{Fearsome: true}, Actors: []Actor{{Side: SideParty, Command: CmdWeapon, Strength: 1, Intellect: 1, Level: 1}}, Assistance: Assistance{FearImmune: immune}}
		r := &fixedRoll{values: []int{99}}
		s.checkFlinch(0, r)
		if immune && (r.i != 0 || s.Actors[0].Command != CmdWeapon) {
			t.Fatal("免怯戰仍擲骰或改指令")
		}
		if !immune && (r.i != 1 || s.Actors[0].Command != CmdFlinch) {
			t.Fatal("關閉沒有沿原規則")
		}
	}
	s := &State{Actors: []Actor{{Side: SideParty, Fatigue: 50}, {Side: SideMonster, Fatigue: 50}}, Assistance: Assistance{NoFatigue: true, EnemyFatigue: 6}}
	s.castFatigue(0, 3, true)
	s.castFatigue(1, 3, true)
	if s.Actors[0].Fatigue != 0 || s.Actors[1].Fatigue != 86 {
		t.Fatal("法術疲勞未套用輔助")
	}
	s.castFatigue(1, 3, false)
	if s.Actors[1].Fatigue != 80 {
		t.Fatal("倍率改變了法術回復")
	}
}
