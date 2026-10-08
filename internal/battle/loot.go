package battle

// LootDownParty 實作SPEC-027／RE-182的共同出口搜刮；勝利領取及賭金結算完成後呼叫。
// 原版0x211BF依怪物倒地計數設出口勝利標記，因此兩側同倒也跳過搜刮。
func (s *State) LootDownParty(rng Rand) {
	if s.lootSettled || s.Stage != StageDone || s.Verdict == VerdictContinue || rng == nil {
		return
	}
	s.lootSettled = true
	if s.Verdict == VerdictWon || (!s.fledFromBar && s.EnemyCount() > 0 && s.allEnemiesDown()) {
		return
	}
	for i := 0; i < PartySlots && i < len(s.Actors); i++ {
		a := &s.Actors[i]
		if a.Side != SideParty || !a.Down() {
			continue
		}
		m := s.member(i)
		if m == nil {
			continue
		}
		// 戰中護甲以Actor的副本為準，避免重新帶回開場耐久。
		for pos, item := range a.Armour {
			m.WornItems[pos] = item
		}
		m.WornItems[9] = m.BeltItem
		for pos := 0; pos < BackpackSlots+len(m.WornItems); pos++ {
			if pos == BackpackSlots+7 {
				continue
			}
			lost := rng.Roll(100) < s.Monster.LootChance
			var item **Item
			if pos < BackpackSlots {
				item = &m.PackItems[pos]
			} else {
				item = &m.WornItems[pos-BackpackSlots]
			}
			if lost && !s.matchesQuestItem(*item) {
				s.trackLootClear()
				m.LootCleared[pos] = true
				*item = nil
				if pos < BackpackSlots {
					m.Pack[pos] = ""
				} else if worn := pos - BackpackSlots; worn < len(a.Armour) {
					a.Armour[worn] = nil
				} else if worn == 9 {
					m.BeltItem, m.BeltName, m.HasSecondItem = nil, "", false
				}
			}
		}
		m.GoldLooted = rng.Roll(100) < s.Monster.LootChance
	}
}

func (s *State) allEnemiesDown() bool {
	for _, a := range s.Actors {
		if a.Side == SideMonster && !a.Down() {
			return false
		}
	}
	return true
}
