package battle

import (
	"reflect"
	"testing"
)

func TestWeaponAttackAreaAndGuardFlow(t *testing.T) {
	for _, area := range []struct{ icon, add int }{{IconAttackHigh, 0}, {IconAttackMiddle, 1}, {IconAttackLow, 2}} {
		s := planningState(Member{HasWeapon: true, Weapon: Weapon{Kind: 0x20, Code: 1}, Size: 100, Health: 100, Strength: 100})
		i := s.Current()
		a := &s.Actors[i]
		a.Facing = 0
		s.Actors = append(s.Actors, Actor{Side: SideMonster, Member: -1, X: a.X - 1, Y: a.Y, Size: 100, Head: 20, Chest: 20})
		s.Plan.Command = CmdWeapon
		s.Plan.Prompt = PromptWeaponForm
		s.Plan.Bar = WeaponForms()
		s.Plan.Selected = 3 // 選項 2
		s.Choose()
		if s.Plan.Prompt != PromptAttackArea || !reflect.DeepEqual(s.Plan.Bar, []int{CmdCancel, IconAttackHigh, IconAttackMiddle, IconAttackLow}) {
			t.Fatalf("漏部位：%+v", s.Plan)
		}
		for j, icon := range s.Plan.Bar {
			if icon == area.icon {
				s.Plan.Selected = j
			}
		}
		s.Choose()
		if s.Plan.Prompt != PromptGuardForm || s.Plan.Style != 2+area.add {
			t.Fatal("部位未寫入攻擊式")
		}
		s.Plan.Selected = 3
		s.Choose()
		if s.Actors[i].Style != 2+area.add || s.Actors[i].Guard != 2 {
			t.Fatalf("%+v", s.Actors[i])
		}
	}
}

func TestWeaponAttackAreaSizeBoundaryAndCancel(t *testing.T) {
	s := planningState(Member{Size: 100, Health: 100, Strength: 100})
	i := s.Current()
	a := s.Actors[i]
	s.Actors[i].Facing = 0
	s.Actors = append(s.Actors, Actor{Side: SideMonster, Member: -1, X: a.X - 1, Y: a.Y, Size: 200, Head: 20, Chest: 20})
	if got := s.attackAreaForms(i); !reflect.DeepEqual(got, []int{CmdCancel, IconAttackMiddle, IconAttackLow}) {
		t.Fatalf("體型差100：%v", got)
	}
	s.Actors[len(s.Actors)-1].Flag1E = 1
	if got := s.attackAreaForms(i); !reflect.DeepEqual(got, []int{CmdCancel, IconAttackMiddle}) {
		t.Fatalf("空中目標：%v", got)
	}
	s.Plan.Prompt = PromptAttackArea
	s.Plan.Bar = []int{CmdCancel, IconAttackMiddle}
	s.Plan.Selected = 0
	s.Choose()
	if s.Plan.Prompt != PromptWeaponForm {
		t.Fatal("取消未退回攻擊式")
	}
	// 瘋狂攻擊保留原版跳過部位及防禦的規則。
	s.Plan.Command = CmdWeapon
	s.Plan.Selected = 2
	s.Choose()
	if s.Actors[i].Style != 1 || s.Actors[i].Guard != 2 || s.Plan.Prompt != PromptCommand {
		t.Fatal("瘋狂攻擊流程改變")
	}
}
