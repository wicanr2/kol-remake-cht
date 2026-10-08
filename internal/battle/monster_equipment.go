package battle

// SPEC-011第13節：RE-203與2026-10-04核准的v2資料契約。
// RE-203：先逐怪物選四套裝備，再進行 SPEC-029 初始化。
func (s *State) initializeMonsterEquipment(r Rand) {
	if !s.equipmentPending || r == nil {
		return
	}
	for i := PartySlots; i < len(s.Actors); i++ {
		a := &s.Actors[i]
		if a.Side != SideMonster {
			continue
		}
		v := r.Roll(4) - 1
		gear := s.Monster.EquipmentVariants[v]
		clone := func(it *Item) *Item {
			if it == nil {
				return nil
			}
			value := *it
			return &value
		}
		a.EquipmentVariant = v
		a.HandItem = clone(gear.Items[0])
		a.Weapon = a.HandItem.CombatWeapon()
		a.Unarmed = a.HandItem == nil
		a.Armour = cloneArmour(gear.Items[1:4])
		a.ArmourWeight = 0
		for _, it := range a.Armour {
			if it != nil {
				a.ArmourWeight += it.Encumbrance
			}
		}
		a.MonsterShield = clone(gear.Items[4])
		a.MonsterRanged = clone(gear.Items[5])
		a.ShieldWeight, a.ShieldGrade = 0, NoShield
		if a.MonsterShield != nil {
			a.ShieldWeight = a.MonsterShield.Encumbrance
			a.ShieldGrade = a.MonsterShield.SizeClass
		}
		a.ReloadGear()
		if i == PartySlots+6 {
			s.MonsterReloadMemory.capture(a, gear)
		}
	}
	s.equipmentPending = false
}
