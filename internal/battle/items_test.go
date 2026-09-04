package battle

import "testing"

func itemState() *State {
	s := &State{Party: []Member{{
		Name: "POONTANG", HasWeapon: true, WeaponName: "LONG SPEAR",
		BeltName: "DAGGER", HasSecondItem: true,
	}}}
	s.Party[0].Worn[0] = "LONG SPEAR"
	s.Party[0].Worn[5] = "DAGGER"
	s.Actors = []Actor{{Side: SideParty, Member: 0, X: 12, Y: 34,
		Head: 10, Chest: 10, Threshold: 100}}
	return s
}

// 丟：東西掉在自己腳下那一格，手上變空。
func TestDropItemLandsUnderfoot(t *testing.T) {
	s := itemState()
	got, name := s.DropItem(0)
	if got != DropsRecord {
		t.Fatalf("訊息是 %#x，期望 %#x", got, DropsRecord)
	}
	if name != "LONG SPEAR" {
		t.Fatalf("丟掉的名字是 %q——訊息要用它展開 |E", name)
	}
	if s.Ground[0].Name != "LONG SPEAR" {
		t.Fatalf("第 0 格是 %q", s.Ground[0].Name)
	}
	if s.Ground[0].X != 12 || s.Ground[0].Y != 34 {
		t.Fatalf("掉在 (%d,%d)，應該在 (12,34)", s.Ground[0].X, s.Ground[0].Y)
	}
	if s.Party[0].HasWeapon || s.Party[0].WeaponName != "" {
		t.Fatal("丟完手上應該是空的")
	}
}

// 十六格都滿了就丟不下去，只印 0x10C。
func TestDropItemWhenGroundIsFull(t *testing.T) {
	s := itemState()
	for i := range s.Ground {
		s.Ground[i] = GroundItem{Name: "ROCK", X: 1, Y: 1}
	}
	if got, _ := s.DropItem(0); got != CannotDropRecord {
		t.Fatalf("訊息是 %#x，期望 %#x", got, CannotDropRecord)
	}
	if !s.Party[0].HasWeapon {
		t.Fatal("丟不下去的時候武器應該還在手上")
	}
}

// 撿：只撿得到腳下那一格，不同格的碰不到。
func TestPickUpOnlyReachesTheSameCell(t *testing.T) {
	s := itemState()
	s.Party[0].HasWeapon = false
	s.Party[0].WeaponName = ""
	s.Ground[0] = GroundItem{Name: "MACE", X: 12, Y: 35}   // 隔壁
	s.Ground[1] = GroundItem{Name: "SHIELD", X: 12, Y: 34} // 腳下

	msg, slot, name := s.PickUpItem(0)
	if name != "SHIELD" {
		t.Fatalf("撿到的名字是 %q——訊息要用它展開 |H", name)
	}
	if msg != FoundItemRecord {
		t.Fatalf("訊息是 %#x，期望 %#x", msg, FoundItemRecord)
	}
	if slot != 1 {
		t.Fatalf("撿到第 %d 格，應該是第 1 格", slot)
	}
	if s.Party[0].WeaponName != "SHIELD" {
		t.Fatalf("手上是 %q", s.Party[0].WeaponName)
	}
	if s.Ground[1].Name != "" {
		t.Fatal("撿走之後那一格應該空了")
	}
	if s.Ground[0].Name != "MACE" {
		t.Fatal("隔壁那一格不該被動到")
	}
}

// 手上已經有東西就撿不起來，印 0x112。
func TestPickUpNeedsAnEmptyHand(t *testing.T) {
	s := itemState()
	s.Ground[0] = GroundItem{Name: "MACE", X: 12, Y: 34}
	msg, slot, _ := s.PickUpItem(0)
	if msg != MustDropRecord || slot != -1 {
		t.Fatalf("手上有東西時得到 %#x／%d", msg, slot)
	}
	if s.Ground[0].Name != "MACE" {
		t.Fatal("撿不起來的東西應該還在地上")
	}
}

// 腳下沒東西就什麼都不印。
func TestPickUpNothingThere(t *testing.T) {
	s := itemState()
	if msg, slot, _ := s.PickUpItem(0); msg != 0 || slot != -1 {
		t.Fatalf("腳下沒東西時得到 %#x／%d", msg, slot)
	}
}

// 切換：手上與腰帶對調。
func TestSwitchItemSwapsHandAndBelt(t *testing.T) {
	s := itemState()
	if got := s.SwitchItem(0); got != PutsInBeltRecord {
		t.Fatalf("訊息是 %#x，期望 %#x", got, PutsInBeltRecord)
	}
	m := s.Party[0]
	if m.WeaponName != "DAGGER" || m.BeltName != "LONG SPEAR" {
		t.Fatalf("交換後手上 %q、腰帶 %q", m.WeaponName, m.BeltName)
	}
	if m.Worn[0] != "DAGGER" || m.Worn[5] != "LONG SPEAR" {
		t.Fatalf("裝備欄沒跟著換：%q／%q", m.Worn[0], m.Worn[5])
	}
	// 換回去要回到原狀。
	s.SwitchItem(0)
	if s.Party[0].WeaponName != "LONG SPEAR" {
		t.Fatal("換兩次應該回到原狀")
	}
}

// 三手都走執行段的 act，不再被標成還沒實作。
func TestActHandlesTheThreeItemCommands(t *testing.T) {
	for _, cmd := range []int{CmdDrop, CmdPickUp, CmdSwitch, CmdSwitchAlt1, CmdSwitchAlt2} {
		s := itemState()
		s.Actors[0].Command = cmd
		r := s.act(0, &fixedRand{values: []int{1}})
		if r.Unimplemented {
			t.Fatalf("%#x 不該被標成還沒實作", cmd)
		}
	}
}
