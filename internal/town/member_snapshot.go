package town

// CloneMember 讓名冊與隊伍各自持有可變資料；住宿交易才提交更新。
func CloneMember(m Member) Member {
	n := m
	if m.Sheet != nil {
		v := *m.Sheet
		v.Weapons = append([]WeaponRow(nil), m.Sheet.Weapons...)
		n.Sheet = &v
	}
	item := func(g *Good) *Good {
		if g == nil {
			return nil
		}
		v := *g
		return &v
	}
	n.InHand = item(m.InHand)
	n.Token = item(m.Token)
	n.Steed = item(m.Steed)
	for i, g := range m.Worn {
		n.Worn[i] = item(g)
	}
	if m.Order != nil {
		v := *m.Order
		v.Allow = append([]int(nil), v.Allow...)
		n.Order = &v
	}
	n.Backpack = append([]Good(nil), m.Backpack...)
	n.Quests = append([]int(nil), m.Quests...)
	n.Progress = append([]int(nil), m.Progress...)
	n.Carried = nil
	for _, r := range m.Carried {
		n.Carried = append(n.Carried, append([]byte(nil), r...))
	}
	n.Spells = append([]Spell(nil), m.Spells...)
	for i := range n.Spells {
		n.Spells[i].Guards = append([]int(nil), m.Spells[i].Guards...)
	}
	return n
}

func CandidateFromMember(m Member) Candidate {
	m = CloneMember(m)
	m.Speaker = false
	return Candidate{ID: m.ID, Full: &m, Name: m.Name, Portrait: m.Portrait, Place: m.Place, Town: m.TownCode, Female: m.Female, Wounds: m.Wounds, Gold: m.Gold, BuildingID: m.BuildingID, Sheet: m.Sheet, Quests: m.Quests, Worn: m.Worn, InHand: m.InHand, Backpack: m.Backpack, Spells: m.Spells, Progress: m.Progress, Carried: m.Carried, Picture: m.Picture, Shield: m.Shield}
}

func (c Candidate) Member() Member { return c.member() }
