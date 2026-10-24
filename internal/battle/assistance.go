package battle

// Assistance 是 SPEC-042 的可選規則。零值完整沿用原規則，不進入存檔。
type Assistance struct {
	NoFatigue    bool
	EnemyFatigue int
	FearImmune   bool
}

func (a Assistance) Multiplier() int {
	switch a.EnemyFatigue {
	case 2, 3, 6:
		return a.EnemyFatigue
	}
	return 1
}

func (s *State) ApplyAssistance() {
	if !s.Assistance.NoFatigue {
		return
	}
	for i := range s.Actors {
		if s.Actors[i].Side == SideParty {
			s.Actors[i].Fatigue = 0
		}
	}
}
