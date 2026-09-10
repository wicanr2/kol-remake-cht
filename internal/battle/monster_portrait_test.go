package battle

import "testing"

func TestMonsterPortraitIndependentOfAbility(t *testing.T) {
	s := &State{Stage: StageField, Monster: Monster{Index: 13}, Actors: make([]Actor, 7)}
	s.Actors[6] = Actor{Side: SideMonster, Head: 10, Chest: 10}
	s.Plan.BodyActor = 7
	for _, ability := range []int{0, 1, 99, 255} {
		s.Actors[6].Kind = ability
		panel, ok := s.EquipmentPanel()
		if !ok || !panel.Monster || panel.Portrait != 13 {
			t.Fatalf("能力 %d 影響怪物肖像：%+v，顯示=%v", ability, panel, ok)
		}
	}
}
