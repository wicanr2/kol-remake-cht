package battle

import "testing"

func TestAmmunitionPanelTracksShotsAndWeaponChanges(t *testing.T) {
	s := planningState(Member{HasWeapon: true, Weapon: Weapon{Kind: 6, Code: 1}})
	s.ToggleEquipmentPanel()
	d, _ := s.EquipmentPanel()
	if d.AmmunitionLabel != "ARROWS:" || d.Ammunition != 20 {
		t.Fatal(d)
	}
	s.Actors[0].ConsumeAmmunition()
	d, _ = s.EquipmentPanel()
	if d.Ammunition != 19 {
		t.Fatal("彈藥顯示沒有跟隨消耗")
	}
	s.ToggleEquipmentPanel()
	d, _ = s.EquipmentPanel()
	if d.AmmunitionLabel != "" {
		t.Fatal("肖像模式不應顯示彈藥")
	}
	s.ToggleEquipmentPanel()
	s.DropItem(0)
	d, _ = s.EquipmentPanel()
	if d.AmmunitionLabel != "" {
		t.Fatal("丟掉武器仍顯示彈藥")
	}
}

func TestEquipmentPanelTracksMutableItems(t *testing.T) {
	m := Member{Name: "PLAYER", Portrait: 12, Size: 80, Health: 80, Strength: 80,
		HasWeapon: true, HasSecondItem: true, WeaponName: "Sword", BeltName: "Axe", Weapon: Weapon{Kind: 0x20, Code: 1},
		Armour: [3]*Item{{Name: "Helmet", Record: "01000000000000000000000000000020"}}, Worn: [6]string{4: "Shield"}}
	s := planningState(m)
	d, ok := s.EquipmentPanel()
	if !ok || d.Name != "PLAYER" || d.Portrait != 12 || d.Rows != [6]string{"Sword", "Helmet", "", "", "Shield", "Axe"} {
		t.Fatalf("初始顯示=%+v", d)
	}
	s.Actors[0].AbsorbArmour(3, 0, &fixedRand{values: []int{1}})
	d, _ = s.EquipmentPanel()
	if d.Rows[1] != "" {
		t.Fatal("破損護甲不可回退成開場名稱")
	}
	s.SwitchItem(0)
	d, _ = s.EquipmentPanel()
	if d.Rows[0] != "Axe" || d.Rows[5] != "Sword" {
		t.Fatalf("腰帶交換未同步：%v", d.Rows)
	}
	s.DropItem(0)
	d, _ = s.EquipmentPanel()
	if d.Rows[0] != "" {
		t.Fatal("丟棄武器後仍顯示舊名稱")
	}
}

func TestEquipmentModePreservesCancelAndResetsNewActor(t *testing.T) {
	s := planningState(Member{HasWeapon: true, Weapon: Weapon{Kind: 0x20, Code: 1}})
	s.ToggleEquipmentPanel()
	if !s.Plan.EquipmentPanel || s.Actors[0].Command != 0 {
		t.Fatal("切換不應選取指令")
	}
	selectIcon(t, s, CmdWeapon)
	s.Choose()
	if _, show := s.EquipmentPanel(); show {
		t.Fatal("空目標應清空兩個資訊框")
	}
	s.ToggleEquipmentPanel()
	if !s.Plan.EquipmentPanel {
		t.Fatal("空格不應切換模式")
	}
	selectIcon(t, s, CmdCancel)
	s.Choose()
	if !s.Plan.EquipmentPanel {
		t.Fatal("取消應保留模式")
	}
	s.AdvanceToParty()
	s.BeginTurn()
	if s.Plan.EquipmentPanel {
		t.Fatal("新隊員規劃應恢復肖像模式")
	}
}

func TestEnemyEquipmentPanelUsesFiveRows(t *testing.T) {
	s := planningState(Member{})
	s.Actors[PartySlots] = Actor{Side: SideMonster, Kind: 3, Armour: [3]*Item{{Name: "Enemy helm"}}}
	s.Monster.Index = 3
	s.Monster.Weapon = &Item{Name: "Claw"}
	s.Monster.Shield = &Item{Name: "Enemy shield"}
	s.Plan.BodyActor = PartySlots + 1
	d, ok := s.EquipmentPanel()
	if !ok || !d.Monster || d.Name != "" || d.Portrait != 3 || d.Count != 5 || d.Rows != [6]string{"Claw", "Enemy helm", "", "", "Enemy shield", ""} {
		t.Fatalf("敵人顯示=%+v", d)
	}
}
