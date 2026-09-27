package battle

import (
	"encoding/hex"
	"fmt"
)

// Armour 是前三個穿戴槽的完整獨立記錄；規則見 SPEC-017。
func cloneArmour(items []*Item) (out [3]*Item) {
	for i := range out {
		if i < len(items) && items[i] != nil {
			copy := *items[i]
			out[i] = &copy
		}
	}
	return
}

func armourBytes(item *Item) (raw [16]byte) {
	if item == nil {
		return
	}
	b, err := hex.DecodeString(item.Record)
	if err != nil || len(b) != len(raw) {
		panic(fmt.Sprintf("護甲 %q 缺少有效的 16-byte 原始記錄", item.Name))
	}
	copy(raw[:], b)
	return
}

// AbsorbArmour 回傳護甲後傷害，並依原版後段修改命中槽的耐久。
// 空槽仍擲骰，負加值也不提前夾零。
func (a *Actor) AbsorbArmour(damage, location int, rng Rand) int {
	if location < 0 || location >= 6 {
		panic("護甲命中部位越界")
	}
	pos := [6]int{0, 1, 1, 1, 2, 2}[location]
	if a.Side == SideMonster {
		pos = [6]int{0, 1, 2, 2, 2, 2}[location]
	}
	item := a.Armour[pos]
	raw := armourBytes(item)
	count, sides := DamageCode(int(raw[12]))
	absorbed := int(raw[8]&0x70) >> 4
	if raw[8]&0x80 != 0 {
		absorbed = -absorbed
	}
	for i := 0; i < count; i++ {
		absorbed += rng.Roll(sides)
	}
	if absorbed >= damage {
		return 0
	}
	remaining := damage - absorbed
	if item != nil {
		condition := int(raw[11]) - int(raw[15]&0x20)*remaining
		if condition < 0 {
			a.breakArmour(pos)
		} else {
			raw[11] = byte(condition)
			item.Condition = condition
			item.Record = hex.EncodeToString(raw[:])
		}
	}
	return remaining
}

// breakArmour 是耐久扣成負的那一件整格清零（`sub_22084:0x220D3`）：連帶負重與徒手
// 拿那一格的各式都要跟著空掉。
func (a *Actor) breakArmour(pos int) {
	item := a.Armour[pos]
	a.Armour[pos] = nil
	a.ArmourWeight -= item.Encumbrance
	for style, slot := range UnarmedSlots {
		if slot == pos {
			a.UnarmedWeapon[style] = Weapon{}
		}
	}
	a.ReloadGear()
}

// wearItem 是 `sub_22084(格, 次數)`：`+0x0F` 有 0x20 才扣，耐久 `+0x0B` 減
// 32 × 次數，扣成負的回 true（呼叫端負責清掉）。沒有這一位的東西原封不動。
func wearItem(item *Item, times int) bool {
	if item == nil {
		return false
	}
	raw := armourBytes(item)
	condition := int(raw[11]) - int(raw[15]&0x20)*times
	if condition < 0 {
		return true
	}
	raw[11] = byte(condition)
	item.Condition = condition
	item.Record = hex.EncodeToString(raw[:])
	return false
}

// WearOverCap 是傷害超過上限之後扣攻擊者那一格（`sub_26D76:0x275C2`…`0x2761A`，RE-155）：
// 隊員用武器或射擊扣手上那件（`+0x30`），徒手扣那一式用的穿戴品；怪物一律扣記錄
// `+0x30`——那是穿戴欄第一格，不是手上的武器。整件扣光了，這一手的指令碼清成 0。
// 徒手的 FEET／SHIELD／HANDS 三格這一版沒帶完整記錄，資料裡那三格也沒有會扣的東西。
func (s *State) WearOverCap(i, over int) {
	if over <= 0 || i < 0 || i >= len(s.Actors) {
		return
	}
	a := &s.Actors[i]
	broken := false
	switch {
	case a.Side == SideMonster:
		if wearItem(a.Armour[0], over) {
			a.breakArmour(0)
			broken = true
		}
	case a.Command == CmdUnarmed:
		if pos := UnarmedSlots[a.Style%len(UnarmedSlots)]; pos < len(a.Armour) && wearItem(a.Armour[pos], over) {
			a.breakArmour(pos)
			broken = true
		}
	default:
		if m := s.member(i); m != nil && m.HasWeapon && wearItem(m.HandItem, over) {
			s.breakHandItem(i)
			broken = true
		}
	}
	if broken {
		a.Command = 0
	}
}
