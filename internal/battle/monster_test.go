package battle

import "testing"

// 怪物的徒手四式也拿穿戴欄那四格（RE-131）。
func TestMonsterUnarmedFromWorn(t *testing.T) {
	worn := make([]*Item, 7)
	worn[0] = &Item{Name: "Conical Helm", Class: 4, Encumbrance: 9}
	worn[4] = &Item{Name: "Self Bow", Class: 0x12, Encumbrance: 20}
	worn[5] = &Item{Name: "Short Spear", Class: 8, Encumbrance: 30}
	got := unarmedFrom(worn)
	// 式 3 是頭槌 → 槽 0；式 4 是拳 → 槽 4；式 1 是踢 → 槽 5。
	if got[3].Kind != 4 || got[4].Kind != 0x12 || got[1].Kind != 8 {
		t.Fatalf("四式拿到 %+v", got)
	}
	// 式 2 是盾擊 → 槽 6，這一份沒有，留空。
	if got[2].Kind != 0 {
		t.Fatalf("空格該留空，得到 %+v", got[2])
	}
	// 短的切片不會越界。
	if out := unarmedFrom(nil); out[3].Kind != 0 {
		t.Fatalf("空的穿戴欄該回全空，得到 %+v", out)
	}
}
