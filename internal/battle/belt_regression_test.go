package battle

import "testing"

func TestBeltSwapRetainsWeaponData(t *testing.T) {
	a := Weapon{Kind: 0x21, Base: 3, Count: 2, Sides: 6, Weight: 10}
	b := Weapon{Kind: 0x22, Base: 5, Count: 1, Sides: 8, Weight: 20}
	for _, occupied := range []bool{false, true} {
		belt := Weapon{}
		name := ""
		if occupied {
			belt, name = b, "AXE"
		}
		s := &State{Actors: []Actor{{Member: 0, Weapon: a}}, Party: []Member{{HasWeapon: true, WeaponName: "SWORD", HasSecondItem: occupied, BeltName: name, BeltWeapon: belt}}}
		s.SwitchItem(0)
		if s.Actors[0].Weapon != belt || s.Party[0].BeltWeapon != a {
			t.Fatalf("首次交換遺失數值：%+v", s.Party[0])
		}
		s.SwitchItem(0)
		if s.Actors[0].Weapon != a || s.Party[0].BeltWeapon != belt || s.Party[0].WeaponName != "SWORD" {
			t.Fatalf("再次交換未還原：%+v", s.Party[0])
		}
	}
}
