package battle

import (
	"slices"
	"testing"
)

func TestPickedWeaponUpdatesCommandsAndAmmunitionPanel(t *testing.T) {
	s := planningState(Member{HasWeapon: true, WeaponName: "Sword", Weapon: Weapon{Kind: 0x58, Code: 1}})
	s.ToggleEquipmentPanel()
	s.DropItem(0)
	if s.Party[0].Weapon != (Weapon{}) {
		t.Fatal("丟棄後選單仍保留舊武器數值")
	}
	bow := Weapon{Kind: 11, Code: 25, Count: 1, Sides: 20, Base: 2, Weight: 40}
	s.Ground[1] = GroundItem{Name: "Elf Bow", X: s.Actors[0].X, Y: s.Actors[0].Y, Weapon: bow}
	if _, name := s.TakeItem(0, 1); name != "Elf Bow" {
		t.Fatal("未拾取弓")
	}
	if s.Party[0].Weapon != bow || s.Actors[0].Weapon != bow {
		t.Fatal("拾取後兩份武器數值不同")
	}
	bar := CommandBar(&s.Actors[0], &s.Party[0], CommandEnv{})
	if !slices.Contains(bar, CmdMissile) || slices.Contains(bar, CmdWeapon) {
		t.Fatalf("拾取弓後指令仍依舊武器：%v", bar)
	}
	d, _ := s.EquipmentPanel()
	if d.Rows[0] != "Elf Bow" || d.AmmunitionLabel != "ARROWS:" || d.Ammunition != 20 {
		t.Fatal(d)
	}
}

func TestFullPackSpoilUpdatesMemberWeapon(t *testing.T) {
	s := planningState(Member{HasWeapon: true, WeaponName: "Sword", Weapon: Weapon{Kind: 0x58, Code: 1}})
	for i := range s.Party[0].Pack {
		s.Party[0].Pack[i] = "Occupied"
	}
	bow := Weapon{Kind: 18, Code: 25, Weight: 30}
	s.Ground[0] = GroundItem{Name: "Crossbow", Weapon: bow}
	s.TakeSpoil(0, 0)
	if s.Party[0].Weapon != bow || s.Actors[0].Weapon != bow {
		t.Fatal("背包滿時戰利品交換未同步武器")
	}
	if s.Ground[0].Name != "Sword" || s.Ground[0].Weapon.Kind != 0x58 {
		t.Fatal("原武器交換至地面時資料遺失")
	}
}
