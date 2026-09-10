package town

import "encoding/hex"

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
// 只使用物品已保存的記錄，不從名稱猜測原版資料。
func (m *Member) SyncCarried() {
	rows := make([][]byte, 1+BackpackSlots+WornSlots)
	put := func(i int, g *Good) {
		rows[i] = make([]byte, CarriedStride)
		if g == nil {
			return
		}
		if raw, err := hex.DecodeString(g.Record); err == nil && len(raw) == CarriedStride {
			copy(rows[i], raw)
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
