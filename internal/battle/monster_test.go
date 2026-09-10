package battle

import "testing"

// 怪物徒手仍然拿手上那件，不挑穿戴品。
//
// 這一支取代了早先的 `TestMonsterUnarmedFromWorn`——那一支釘的是「怪物的徒手
// 四式也拿穿戴欄那四格」，2026-09-10 的讀取監看推翻了它：四式的 `var_4` 一律
// 指向本體記錄 `+0x20`（`docs/re/20260910-monster-unarmed-slots.md`）。
func TestMonsterUnarmedUsesHeldWeapon(t *testing.T) {
	club := Weapon{Kind: 60, Count: 2, Sides: 5}
	a := Actor{Side: SideMonster, Command: CmdUnarmed, Weapon: club}
	// 就算穿戴品那幾格有東西，四式也不看它們。
	a.UnarmedWeapon[1] = Weapon{Kind: 8, Count: 1, Sides: 3}
	a.UnarmedWeapon[3] = Weapon{Kind: 4, Count: 1, Sides: 3}
	for style := 0; style < len(a.UnarmedWeapon); style++ {
		if got := a.AttackWeapon(style); got != club {
			t.Fatalf("怪物第 %d 式拿到 %+v，該是手上那件 %+v", style, got, club)
		}
	}
	// 反面：同一份設定換成隊員，就會挑穿戴品——判準確實分得出兩側。
	a.Side = SideParty
	if got := a.AttackWeapon(3); got.Kind != 4 {
		t.Fatalf("隊員第 3 式該拿穿戴品，得到 %+v", got)
	}
	// 空手的怪物（本體記錄 `+0x20` 全零）拿到的就是空武器，不會退回穿戴品。
	b := Actor{Side: SideMonster, Command: CmdUnarmed}
	b.UnarmedWeapon[3] = Weapon{Kind: 4, Count: 1, Sides: 3}
	if got := b.AttackWeapon(3); got.Kind != 0 || got.Count != 0 {
		t.Fatalf("空手怪物該拿到空武器，得到 %+v", got)
	}
}
