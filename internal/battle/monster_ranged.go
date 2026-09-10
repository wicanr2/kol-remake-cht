package battle

// MonsterRangedRoll 對應 sub_222BF；SPEC-029 的獨立三位元骰數與骰面，總和保留低 byte。
func MonsterRangedRoll(code int, r Rand) int {
	sides := [...]int{2, 3, 5, 7, 9, 11, 19, 99}
	value := 0
	for n := 0; n <= (code>>3)&7; n++ {
		value = (value + r.Roll(sides[code&7])) & 255
	}
	return value
}

func MonsterRangedValue(chance, code int, r Rand) int {
	if chance > r.Roll(100) {
		return MonsterRangedRoll(code, r)
	}
	return 0
}

// InitializeMonsterRanged 在首次規劃前依槽位初始化一次；取消或後續回合不重擲。
func (s *State) InitializeMonsterRanged(r Rand) {
	if !s.rangedPending || r == nil { return }
	for i := PartySlots; i < len(s.Actors); i++ {
		if s.Actors[i].Side == SideMonster {
			s.Actors[i].Kind = MonsterRangedValue(s.Monster.RangedChance, s.Monster.RangedCode, r)
		}
	}
	s.rangedPending = false
}
