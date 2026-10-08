package battle

import "testing"

func TestCommandPanelSourcesAndGate(t *testing.T) {
	s := &State{Stage: StageField, Party: []Member{{}}, Actors: []Actor{{Side: SideParty, Member: 0, Intellect: 60, Command: 7}, {Side: SideMonster, Level: 2, Command: 0x29, Style: 1, Guard: 6, Facing: 7}}, Round: Round{Order: []int{0}, InspectionRoll: 40}, Plan: Planning{Prompt: PromptWeaponForm, BodyActor: 1, Command: 0x21, Style: 2, Guard: 1}}
	got := s.CommandPanel()
	if len(got) != 4 || got[0].Code != 0x2C || got[1].Code != 0x2F || got[2].Code != 0x30 {
		t.Fatalf("規劃值來源=%v", got)
	}

	s.Plan.Prompt = PromptCommand
	s.Actors[0].Command, s.Actors[0].Style, s.Actors[0].Guard = 0x21, 3, 6
	got = s.CommandPanel()
	if len(got) != 4 || got[0].Code != 0x2C || got[1].Code != 0x41 || got[2].Code != 0x36 {
		t.Fatalf("主列未重載人物指令=%v", got)
	}
	if s.Plan.Command != 0x21 || s.Plan.Style != 2 || s.Plan.Guard != 1 {
		t.Fatal("顯示改變編輯暫存值")
	}
	s.Plan.BodyActor = 2
	if len(s.CommandPanel()) != 0 {
		t.Fatal("等於門檻不應顯示敵人命令")
	}
	s.Round.InspectionRoll = 39
	got = s.CommandPanel()
	if len(got) != 3 || got[0].Code != 0x37 || got[1].Code != 0x36 || got[2].Code != 0x4E {
		t.Fatalf("敵人記錄來源=%v", got)
	}
	if s.Round.InspectionRoll != 39 || s.Plan.Command != 0x21 {
		t.Fatal("顯示改變規劃狀態")
	}
	s.Plan.BodyActor = 0
	if len(s.CommandPanel()) != 0 {
		t.Fatal("空格應清框")
	}
}
