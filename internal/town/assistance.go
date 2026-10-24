package town

import "fmt"

// Strengthen 永久補足指定欄位。先驗證全部目標，避免只修改半隊。
func (s *Session) Strengthen(all bool, kind string, floor int) (int, error) {
	max := 0
	switch kind {
	case "stats":
		max = 170
	case "skills":
		max = 200
	case "gold":
		max = 16777215
	default:
		return 0, fmt.Errorf("未知強化項目")
	}
	if floor < 0 || floor > max || kind != "gold" && floor == 0 {
		return 0, fmt.Errorf("數值超出範圍")
	}
	if len(s.Party) == 0 {
		return 0, fmt.Errorf("目前沒有隊伍")
	}
	target := 0
	for i, m := range s.Party {
		if m.Speaker {
			target = i
			break
		}
	}
	for i, m := range s.Party {
		if (all || i == target) && kind != "gold" && m.Sheet == nil {
			return 0, fmt.Errorf("角色缺少資料頁")
		}
	}
	n := 0
	for i := range s.Party {
		if !all && i != target {
			continue
		}
		m := &s.Party[i]
		switch kind {
		case "stats":
			v := (floor*150 + 99) / 100
			for k := range m.Sheet.Stats {
				if m.Sheet.Stats[k] < v {
					m.Sheet.Stats[k] = v
				}
			}
		case "skills":
			for k := range m.Sheet.Weapons {
				w := &m.Sheet.Weapons[k]
				if w.Offence < floor {
					w.Offence = floor
				}
				if w.Defence < floor {
					w.Defence = floor
				}
				if w.OffenceRaw < w.Offence*10 {
					w.OffenceRaw = w.Offence * 10
				}
				if w.DefenceRaw < w.Defence*10 {
					w.DefenceRaw = w.Defence * 10
				}
			}
		case "gold":
			if m.Gold < floor {
				m.Gold = floor
			}
		}
		for k := range s.Roster {
			c := &s.Roster[k]
			if m.ID != "" && c.ID == m.ID || m.ID == "" && c.Name == m.Name {
				c.Sheet = m.Sheet
				c.Gold = m.Gold
				if c.Full != nil {
					c.Full.Sheet = m.Sheet
					c.Full.Gold = m.Gold
				}
			}
		}
		n++
	}
	s.syncParty()
	if s.Inside != nil {
		s.Inside.Sheet = nil
	}
	return n, nil
}
