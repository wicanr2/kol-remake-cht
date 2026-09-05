package town

import "github.com/wicanr2/kol-remake-cht/internal/portrait"

// 立繪要看的那幾格在角色記錄裡的位移（RE-126）。三層護甲是穿戴欄的前三格，
// 盾在部位 6，手上那件在 0x30，背著沒拿的那一件在 0x190。
const (
	PortraitHeadSlot  = 0x100
	PortraitTorsoSlot = 0x110
	PortraitLegsSlot  = 0x120
	PortraitShieldPos = 6
	PortraitSpareSlot = 0x190
)

// portraitLayerSlots 依 portrait 的層編號（0 頭、1 身、2 腿）排。
var portraitLayerSlots = [portrait.LayerCount]int{
	PortraitHeadSlot, PortraitTorsoSlot, PortraitLegsSlot,
}

// goodItem 把一件東西換成立繪要的四個欄位。
func goodItem(g *Good) *portrait.Item {
	if g == nil {
		return nil
	}
	return &portrait.Item{Class: g.Class, Style: g.SizeClass,
		Kind: g.Kind, TwoHand: g.Hands == 1}
}

// carriedItem 走另一條路：從身上那一格的原始十六 byte 讀同樣四個欄位。
// 讀檔回來的隊伍只有 Carried（存檔存的是原始位元組），typed 的 Worn 是空的。
func carriedItem(rec []byte) *portrait.Item {
	if len(rec) < 16 {
		return nil
	}
	empty := true
	for _, b := range rec {
		if b != 0 {
			empty = false
			break
		}
	}
	if empty {
		return nil
	}
	return &portrait.Item{Class: int(rec[13]), Style: int(rec[14] & 0x0F),
		Kind: int(rec[12] & 3), TwoHand: rec[13]&2 != 0}
}

// carriedAt 取身上某一個位移那一格。索引與 CarriedOffsets 對齊。
func (m Member) carriedAt(offset int) []byte {
	for i, off := range CarriedOffsets() {
		if off != offset {
			continue
		}
		if i < len(m.Carried) {
			return m.Carried[i]
		}
		return nil
	}
	return nil
}

// item 先看 typed 的穿戴欄，沒有再回頭讀原始位元組。
func (m Member) item(slot, offset int) *portrait.Item {
	if slot >= 0 && slot < len(m.Worn) {
		if it := goodItem(m.Worn[slot]); it != nil {
			return it
		}
	}
	return carriedItem(m.carriedAt(offset))
}

// Figure 是這名角色的立繪外觀（RE-126）。
func (m Member) Figure() portrait.Figure {
	f := portrait.Figure{Styles: [portrait.LayerCount]int{-1, -1, -1}}
	if m.Female {
		f.Gender = 1
	}
	if m.Sheet != nil {
		f.Race = m.Sheet.Race
	}
	for layer, offset := range portraitLayerSlots {
		if it := m.item(layer, offset); it != nil {
			f.Styles[layer] = it.Style
		}
	}
	// 手上那件在穿戴欄的第十格（`WeaponSlot`），與 0x30 是同一格。
	f.Weapon = m.item(WornSlots-1, CarriedFirst)
	f.Shield = m.item(PortraitShieldPos, PortraitHeadSlot+PortraitShieldPos*0x10)
	// 背著那一件只有原始位元組這一條路：typed 的穿戴欄沒有它。
	f.Spare = carriedItem(m.carriedAt(PortraitSpareSlot))
	return f
}

// PartyFigures 是隊伍那一頁要畫的東西（`sub_18D61`，RE-126）：層次大的先畫，
// 每個人往右挪的格數照畫的順序取。回傳的順序就是貼圖的順序。
func (s *Session) PartyFigures(l *portrait.Layout) []portrait.Piece {
	if l == nil || len(s.Party) == 0 {
		return nil
	}
	builds := make([]int, len(s.Party))
	for i, m := range s.Party {
		builds[i] = m.Figure().Build()
	}
	order, dx := l.PartyOrder(builds)
	var out []portrait.Piece
	for n, i := range order {
		out = append(out, l.Compose(s.Party[i].Figure(), dx[n])...)
	}
	return out
}
