package battle

import "testing"

// 上限加倍的遮罩：選項 0、1 讀到表前的 0x109A／0x109B（RE-155）。
func TestCapMask(t *testing.T) {
	for style, want := range map[int]int{0: 6, 1: 5, 2: 16, 5: 8, 11: 2} {
		if got := capMask(style); got != want {
			t.Errorf("選項 %d：遮罩 %d，應為 %d", style, got, want)
		}
	}
}

// 種類沒有那一位就把上限加倍；徒手不加倍。超出量是夾回之前的差。
func TestDamageRollCapDoubling(t *testing.T) {
	w := Weapon{Base: 5, Count: 1, Sides: 2} // 上限 8
	for _, c := range []struct {
		kind, cmd, total, over int
	}{
		{6, CmdWeapon, 8, 24},  // 32 夾到 8
		{0, CmdWeapon, 16, 16}, // 上限加倍成 16
		{0, CmdUnarmed, 8, 24}, // 怪物拿 Weapon；徒手指令不加倍
	} {
		a := Actor{Side: SideMonster, Strength: 70, Stamina: 100, Command: c.cmd}
		a.Weapon = w
		a.Weapon.Kind = c.kind
		// 基礎 5 ＋ 骰 2 ＋ 餘裕 200 ÷ 20 查到的 25 ＝ 32。
		total, over := damageRoll(a, 0, 200, &seqRand{values: []int{2}})
		if total != c.total || over != c.over {
			t.Errorf("種類 %d：傷害 %d、超出 %d，應為 %d、%d", c.kind, total, over, c.total, c.over)
		}
	}
}

func overCapState() *State {
	s := &State{Actors: make([]Actor, 7), Party: []Member{{WeaponName: "INGOT"}}}
	s.Actors[0] = Actor{Side: SideParty, Member: 0, Command: CmdWeapon}
	s.Actors[6] = Actor{Side: SideMonster, Command: CmdWeapon}
	return s
}

// 怪物扣的是穿戴欄第一格（`+0x30`），不是手上的武器；扣光連指令一起清掉。
func TestWearOverCapMonsterHood(t *testing.T) {
	s := overCapState()
	s.Actors[6].Armour[0] = testArmour(0, 255, 0x20)
	s.WearOverCap(6, 3)
	if s.Actors[6].Armour[0].Condition != 255-96 || s.Actors[6].Command != CmdWeapon {
		t.Fatalf("應該扣 96 點耐久，得到 %d", s.Actors[6].Armour[0].Condition)
	}
	s.WearOverCap(6, 6)
	if s.Actors[6].Armour[0] != nil || s.Actors[6].Command != 0 {
		t.Fatal("扣光了應該整件清掉、指令碼清成 0")
	}
}

// 沒有 0x20 的東西不扣。
func TestWearOverCapIgnoresPlainItems(t *testing.T) {
	s := overCapState()
	s.Actors[6].Armour[0] = testArmour(0, 10, 0)
	s.WearOverCap(6, 100)
	if s.Actors[6].Armour[0] == nil || s.Actors[6].Armour[0].Condition != 10 {
		t.Fatal("沒有 0x20 的護甲不該被扣")
	}
}

// 隊員頭槌扣 HEAD；用武器扣手上那件，扣光就沒武器了。
func TestWearOverCapParty(t *testing.T) {
	s := overCapState()
	s.Actors[0].Command, s.Actors[0].Style = CmdUnarmed, 3
	s.Actors[0].Armour[0] = testArmour(0, 40, 0x20)
	s.WearOverCap(0, 1)
	if s.Actors[0].Armour[0].Condition != 8 {
		t.Fatalf("頭槌應該扣 HEAD 32 點，得到 %d", s.Actors[0].Armour[0].Condition)
	}

	s = overCapState()
	s.Party[0].HasWeapon = true
	s.Party[0].HandItem = testArmour(0, 20, 0x20)
	s.Actors[0].Weapon = Weapon{Base: 1, Count: 1, Sides: 2}
	s.WearOverCap(0, 1)
	m := s.Party[0]
	if m.HandItem != nil || m.HasWeapon || s.Actors[0].Weapon != (Weapon{}) || s.Actors[0].Command != 0 {
		t.Fatal("手上那件扣光了應該清掉，指令碼清成 0")
	}
}
