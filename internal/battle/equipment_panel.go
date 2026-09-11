package battle

// ToggleEquipmentPanel 不執行目前圖示；目標格空白時空白鍵沒有作用（SPEC-019）。
func (s *State) ToggleEquipmentPanel() {
	if (s.Plan.Prompt == PromptCommand || s.Plan.Prompt == PromptTarget) && s.Plan.BodyActor > 0 {
		s.Plan.EquipmentPanel = !s.Plan.EquipmentPanel
	}
}

// EquipmentLabels 是裝備欄那六列的標籤（原版是 `W: H: T: L: S: B:`，依序是
// 武器、頭、軀幹、腿、盾、身體）。中文模式拿它去查譯文。
var EquipmentLabels = [6]string{"W:", "H:", "T:", "L:", "S:", "B:"}

type EquipmentPanelData struct {
	Name               string
	Monster, Equipment bool
	Portrait           int
	Rows               [6]string
	Count              int
	AmmunitionLabel    string
	Ammunition         int
}

func (s *State) EquipmentPanel() (EquipmentPanelData, bool) {
	if _, show := s.BodyPanel(); !show {
		return EquipmentPanelData{}, false
	}
	i := s.Plan.BodyActor - 1
	a := &s.Actors[i]
	d := EquipmentPanelData{Equipment: s.Plan.EquipmentPanel, Count: 6}
	for j, item := range a.Armour {
		if item != nil {
			d.Rows[j+1] = item.Name
		}
	}
	if a.Side == SideParty {
		m := s.member(i)
		d.Name, d.Portrait = m.Name, m.Portrait
		d.Rows[0], d.Rows[4], d.Rows[5] = m.WeaponName, m.Worn[4], m.BeltName
		if d.Equipment && m.HasWeapon && m.Weapon.Code&1 != 0 && m.Weapon.Kind&0xE0 == 0 {
			d.AmmunitionLabel = "ARROWS:"
			if m.Weapon.Kind == 0x12 || m.Weapon.Kind == 0x16 {
				d.AmmunitionLabel = "BOLTS:"
			}
			if a.Kind != 0xFF {
				d.Ammunition = a.Kind + 1
			}
		}
	} else {
		d.Monster, d.Portrait, d.Count = true, s.Monster.Index, 5
		if s.Monster.Weapon != nil && !a.Unarmed {
			d.Rows[0] = s.Monster.Weapon.Name
		}
		if s.Monster.Shield != nil {
			d.Rows[4] = s.Monster.Shield.Name
		}
	}
	return d, true
}
