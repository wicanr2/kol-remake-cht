package battle

import "testing"

// planningState 擺一場只有一名隊員的戰鬥，並把規劃段推到那一名頭上。
func planningState(m Member) *State {
	site := testSite()
	site.Count = 0
	s := NewQuest(site, []Member{m}, testMonster(), NewField())
	s.Stage = StageField
	s.Round.Order = []int{0}
	s.Round.Turn = 0
	s.Cursor = 0
	s.BeginTurn()
	return s
}

// 走完一整條武器攻擊：主指令列 → 選目標格 → 武器攻擊式 → 防禦方式。
func TestPlanningWeaponAttack(t *testing.T) {
	s := planningState(Member{Name: "Poontang", HasWeapon: true,
		Weapon: Weapon{Kind: 0x20, Code: 1}})
	if s.Plan.Prompt != PromptCommand {
		t.Fatalf("一開始應該停在主指令列，得到 %d", s.Plan.Prompt)
	}
	selectIcon(t, s, CmdWeapon)
	if s.Choose() {
		t.Fatal("選武器攻擊不該直接結束這一輪")
	}
	if s.Plan.Prompt != PromptTarget {
		t.Fatalf("選武器之後應該進選目標格，得到 %d", s.Plan.Prompt)
	}
	// 面向預設是 0（上），起點就在 (3, 4)。
	if s.Choose() {
		t.Fatal("確定目標格不該結束這一輪")
	}
	if s.Plan.Prompt != PromptWeaponForm {
		t.Fatalf("確定目標之後應該進武器攻擊式，得到 %d", s.Plan.Prompt)
	}
	selectIcon(t, s, 0x2D) // 攻擊選項 5
	if s.Choose() {
		t.Fatal("選攻擊式不該結束這一輪")
	}
	if s.Plan.Prompt != PromptGuardForm {
		t.Fatalf("選完攻擊式應該進防禦方式，得到 %d", s.Plan.Prompt)
	}
	selectIcon(t, s, 0x32) // 防禦選項 3
	if !s.Choose() {
		t.Fatal("只有一名隊員，選完防禦式這一輪就該結束")
	}
	a := s.Actors[0]
	if a.Command != CmdWeapon || a.Style != 5 || a.Guard != 3 {
		t.Fatalf("寫回的三個值是 %#x／%d／%d，應該是 %#x／5／3",
			a.Command, a.Style, a.Guard, CmdWeapon)
	}
	if s.Plan.Prompt != PromptNextRound {
		t.Fatalf("整輪決定完應該停在 Begin next round?，得到 %d", s.Plan.Prompt)
	}
}

// 武器攻擊式的選項為 1 時原版不問防禦式，直接給 2 收尾（`0x1F738`）。
func TestPlanningWeaponFormOneSkipsGuard(t *testing.T) {
	s := planningState(Member{HasWeapon: true, Weapon: Weapon{Kind: 0x20, Code: 1}})
	selectIcon(t, s, CmdWeapon)
	s.Choose()
	s.Choose() // 確定目標格
	selectIcon(t, s, 0x2B)
	if !s.Choose() {
		t.Fatal("攻擊選項 1 應該直接收尾")
	}
	a := s.Actors[0]
	if a.Style != 1 || a.Guard != 2 {
		t.Fatalf("應該是選項 1／防禦 2，得到 %d／%d", a.Style, a.Guard)
	}
}

// 移動：攻擊選項欄放圖號本身，防禦選項只有走路那一格是 2。
func TestPlanningMove(t *testing.T) {
	s := planningState(Member{})
	selectIcon(t, s, CmdMove)
	s.Choose()
	if s.Plan.Prompt != PromptTarget {
		t.Fatalf("選移動之後應該進選目標格，得到 %d", s.Plan.Prompt)
	}
	s.Choose()
	if s.Plan.Prompt != PromptMoveForm {
		t.Fatalf("確定目標之後應該進移動方式，得到 %d", s.Plan.Prompt)
	}
	selectIcon(t, s, IconMoveWalk)
	if !s.Choose() {
		t.Fatal("選完移動方式這一輪就該結束")
	}
	a := s.Actors[0]
	if a.Command != CmdMove || a.Style != IconMoveWalk || a.Guard != 2 {
		t.Fatalf("得到 %#x／%#x／%d，應該是 %#x／%#x／2",
			a.Command, a.Style, a.Guard, CmdMove, IconMoveWalk)
	}
}

// 射擊選完目標就收尾，並把那一格記進 Memo（`0x1F7BF`、`0x24936`）。
func TestPlanningMissileRemembersTarget(t *testing.T) {
	s := planningState(Member{HasWeapon: true, Weapon: Weapon{Kind: 6, Code: 1}})
	selectIcon(t, s, CmdMissile)
	s.Choose()
	if s.Plan.Cursor.Mode != CursorRanged {
		t.Fatalf("射擊應該用 mode 2 的游標，得到 %d", s.Plan.Cursor.Mode)
	}
	s.MoveCursor(-2, 0, false) // 從中心往上兩格，落在外圈
	if !s.Choose() {
		t.Fatal("射擊確定目標就該收尾")
	}
	a := s.Actors[0]
	if a.Command != CmdMissile || a.Style != 0 || a.Guard != 0 {
		t.Fatalf("射擊應該寫 %#x／0／0，得到 %#x／%d／%d",
			CmdMissile, a.Command, a.Style, a.Guard)
	}
	want := Cell{Row: a.X - 2, Col: a.Y}
	if got := s.Plan.Memo[0]; got != want {
		t.Fatalf("記住的目標是 (%d, %d)，應該是 (%d, %d)",
			got.Row, got.Col, want.Row, want.Col)
	}
}

// 選目標格把白框移到 U-TURN 再確定，就退回主指令列。
func TestPlanningTargetCancel(t *testing.T) {
	s := planningState(Member{HasWeapon: true, Weapon: Weapon{Kind: 0x20, Code: 1}})
	selectIcon(t, s, CmdWeapon)
	s.Choose()
	s.MoveSelection(1) // 白框移到 U-TURN
	if s.Choose() {
		t.Fatal("取消不該結束這一輪")
	}
	if s.Plan.Prompt != PromptCommand {
		t.Fatalf("取消應該退回主指令列，得到 %d", s.Plan.Prompt)
	}
	if s.Actors[0].Command != 0 {
		t.Fatalf("取消不該寫回指令，得到 %#x", s.Actors[0].Command)
	}
}

// 防禦方式取消退回攻擊式那一列（`0x1F772`）。
func TestPlanningGuardCancelGoesBack(t *testing.T) {
	s := planningState(Member{HasWeapon: true, Weapon: Weapon{Kind: 0x20, Code: 1}})
	selectIcon(t, s, CmdWeapon)
	s.Choose()
	s.Choose()
	selectIcon(t, s, 0x2D)
	s.Choose()
	selectIcon(t, s, CmdCancel)
	s.Choose()
	if s.Plan.Prompt != PromptWeaponForm {
		t.Fatalf("防禦式取消應該退回武器攻擊式，得到 %d", s.Plan.Prompt)
	}
}

// 兩個選項都是 0 時整個指令改成休息（`0x1F785`）。
func TestPlanningZeroOptionsBecomeRest(t *testing.T) {
	s := planningState(Member{HasWeapon: true, Weapon: Weapon{Kind: 0x20, Code: 1}})
	selectIcon(t, s, CmdWeapon)
	s.Choose()
	s.Choose()
	selectIcon(t, s, 0x2A) // 攻擊選項 0
	s.Choose()
	selectIcon(t, s, 0x2A) // 防禦選項 0
	if !s.Choose() {
		t.Fatal("這一輪應該結束")
	}
	if s.Actors[0].Command != CmdRest {
		t.Fatalf("兩個選項都 0 應該改成休息，得到 %#x", s.Actors[0].Command)
	}
}

// 直接收尾的那幾格：兩個選項都 0，指令碼就是圖號本身（`0x1F3B2`）。
func TestPlanningDirectCommands(t *testing.T) {
	for _, icon := range []int{CmdRest, CmdDrop} {
		s := planningState(Member{HasWeapon: true, Weapon: Weapon{Kind: 0x20, Code: 1}})
		selectIcon(t, s, icon)
		if !s.Choose() {
			t.Fatalf("%#x 應該直接收尾", icon)
		}
		a := s.Actors[0]
		if a.Command != icon || a.Style != 0 || a.Guard != 0 {
			t.Fatalf("%#x 應該寫 %#x／0／0，得到 %#x／%d／%d",
				icon, icon, a.Command, a.Style, a.Guard)
		}
	}
}

// 還沒實作的圖示按下去會被記下來，不默默吞掉。分派表（`0x1FA02`）認得的圖號
// 現在都做了，所以這裡拿一個表上沒有、remake 也還沒接的圖號當樣本。
func TestPlanningUnimplemented(t *testing.T) {
	const notYet = 0x30 // 一般遭遇那一列多出來的那一格，語意未定
	s := planningState(Member{HasWeapon: true,
		Weapon: Weapon{Kind: 0x20, Code: 1}})
	s.Plan.Bar = []int{CmdYes, notYet}
	s.Plan.Selected = 1
	if s.Choose() {
		t.Fatal("還沒實作的圖示不該結束這一輪")
	}
	if len(s.Plan.Unimplemented) != 1 || s.Plan.Unimplemented[0] != notYet {
		t.Fatalf("應該記下 %#x，得到 %#x", notYet, s.Plan.Unimplemented)
	}
	if s.Plan.Prompt != PromptCommand {
		t.Fatalf("應該留在主指令列，得到 %d", s.Plan.Prompt)
	}
}

// 0x22 那一格是「還沒上弦」：選下去直接收尾，執行段才變成裝填（RE-105）。
func TestPlanningLoadCommits(t *testing.T) {
	s := planningState(Member{HasWeapon: true,
		Weapon: Weapon{Kind: 0x20, Code: 1}})
	s.Plan.Bar = []int{CmdYes, CmdLoad}
	s.Plan.Selected = 1
	who := s.Current()
	if !s.Choose() {
		t.Fatal("選 0x22 應該直接收尾")
	}
	if len(s.Plan.Unimplemented) != 0 {
		t.Fatalf("不該記成還沒實作，得到 %#x", s.Plan.Unimplemented)
	}
	if got := s.Actors[who].Command; got != CmdLoad {
		t.Fatalf("指令碼應該是 %#x，得到 %#x", CmdLoad, got)
	}
}

// 白框在圖示列上繞回，與城鎮同一條規則。
func TestPlanningSelectionWraps(t *testing.T) {
	s := planningState(Member{})
	n := len(s.Plan.Bar)
	s.MoveSelection(-1)
	if s.Plan.Selected != n-1 {
		t.Fatalf("往左應該繞到末格 %d，得到 %d", n-1, s.Plan.Selected)
	}
	s.MoveSelection(1)
	if s.Plan.Selected != 0 {
		t.Fatalf("再往右應該繞回第 0 格，得到 %d", s.Plan.Selected)
	}
}

// selectIcon 把白框移到某個圖示上；找不到就讓測試失敗。
func selectIcon(t *testing.T, s *State, icon int) {
	t.Helper()
	for i, v := range s.Plan.Bar {
		if v == icon {
			s.Plan.Selected = i
			if s.Plan.Prompt == PromptTarget {
				s.Plan.Cursor.Icon = i
			}
			return
		}
	}
	t.Fatalf("圖示列 %#x 裡沒有 %#x", s.Plan.Bar, icon)
}
