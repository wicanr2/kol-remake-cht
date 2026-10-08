package town

import "encoding/hex"

// emptyPackSlot 找原版六格中的第一個空位；中間空格不以slice長度判滿。
func (m Member) emptyPackSlot() int {
	for i := 0; i < BackpackSlots; i++ {
		if i >= len(m.Backpack) || m.Backpack[i].Name == "" {
			return i
		}
	}
	return -1
}

func (m *Member) putPack(slot int, good Good) {
	if slot < 0 || slot >= BackpackSlots {
		return
	}
	for len(m.Backpack) <= slot {
		m.Backpack = append(m.Backpack, Good{})
	}
	m.Backpack[slot] = good
}

func (m *Member) clearPack(slot int) *Good {
	if slot < 0 || slot >= len(m.Backpack) || m.Backpack[slot].Name == "" {
		return nil
	}
	good := m.Backpack[slot]
	m.Backpack[slot] = Good{}
	for len(m.Backpack) > 0 && m.Backpack[len(m.Backpack)-1].Name == "" {
		m.Backpack = m.Backpack[:len(m.Backpack)-1]
	}
	return &good
}

// setEncumbrance 同步已知的 +0x0A 欄位（RE-075／SPEC-009）。
// 沒有完整來源記錄的物品只改 typed data，不補造其餘未知 bytes。
func (g *Good) setEncumbrance(value int) {
	g.Encumbrance = value
	if raw, err := hex.DecodeString(g.Record); err == nil && len(raw) == CarriedStride {
		raw[0x0A] = byte(value)
		g.Record = hex.EncodeToString(raw)
	}
}

// SyncCarried 同步任務比對的十七格：手持、背包六格、穿戴十格（RE-109）。
// 只使用已保存的記錄；空槽保留原始殘留，有物槽移除仍清零（RE-170／182）。
func (m *Member) SyncCarried() {
	rows := make([][]byte, 1+BackpackSlots+WornSlots)
	put := func(i int, g *Good) {
		rows[i] = make([]byte, CarriedStride)
		if g == nil {
			if i < len(m.Carried) && len(m.Carried[i]) == CarriedStride && m.Carried[i][0] == 0 {
				copy(rows[i], m.Carried[i])
			}
			return
		}
		if raw, err := hex.DecodeString(g.Record); err == nil && len(raw) == CarriedStride {
			copy(rows[i], raw)
		} else if g.Name == "" && i < len(m.Carried) && len(m.Carried[i]) == CarriedStride && m.Carried[i][0] == 0 {
			copy(rows[i], m.Carried[i])
		}
	}
	put(0, m.InHand)
	for i := 0; i < BackpackSlots; i++ {
		var g *Good
		if i < len(m.Backpack) {
			g = &m.Backpack[i]
		}
		put(1+i, g)
	}
	for i, g := range m.Worn {
		put(1+BackpackSlots+i, g)
	}
	m.Carried = rows
}
