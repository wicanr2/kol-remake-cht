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
			a.Armour[pos] = nil
			a.ArmourWeight -= item.Encumbrance
			for style, slot := range UnarmedSlots {
				if slot == pos {
					a.UnarmedWeapon[style] = Weapon{}
				}
			}
			a.ReloadGear()
		} else {
			raw[11] = byte(condition)
			item.Condition = condition
			item.Record = hex.EncodeToString(raw[:])
		}
	}
	return remaining
}
