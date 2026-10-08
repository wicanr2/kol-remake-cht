package battle

import "testing"

// planningState 擺一場只有一名隊員的戰鬥，並把規劃段推到那一名頭上。
func planningState(m Member) *State {
	// 部位耐久由體型與健康算出；兩者都是 0 會讓六個部位全是 0，等於手腳都廢了。
	if m.Size == 0 && m.Health == 0 {
		m.Size, m.Health = 40, 40
	}
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
	if !commitThenYes(t, s) {
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
	if !commitThenYes(t, s) {
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
	if !commitThenYes(t, s) {
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
	if s.Plan.Cursor.Row != 0 || s.Plan.Cursor.Col != 0 {
		t.Fatal("首次射擊應從左上角開始")
	}
	// RE-171 gui10：沿左邊、下邊正常走到右下角，越界方向鍵留在邊界。
	for n := 0; n < 9; n++ {
		s.MoveCursor(1, 0, false)
	}
	for n := 0; n < 9; n++ {
		s.MoveCursor(0, 1, false)
	}
	if !commitThenYes(t, s) {
		t.Fatal("射擊確定目標就該收尾")
	}
	a := s.Actors[0]
	if a.Command != CmdMissile || a.Style != 0 || a.Guard != 0 {
		t.Fatalf("射擊應該寫 %#x／0／0，得到 %#x／%d／%d",
			CmdMissile, a.Command, a.Style, a.Guard)
	}
	want := Cell{Row: a.X + 4, Col: a.Y + 4}
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
	if !commitThenYes(t, s) {
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
		if !commitThenYes(t, s) {
			t.Fatalf("%#x 應該直接收尾", icon)
		}
		a := s.Actors[0]
		if a.Command != icon || a.Style != 0 || a.Guard != 0 {
			t.Fatalf("%#x 應該寫 %#x／0／0，得到 %#x／%d／%d",
				icon, icon, a.Command, a.Style, a.Guard)
		}
	}
}

// 指令列被改壞時，分派表認不得的圖號會被記下來，不默默吞掉。0x30 是「Begin next
// round?」那一列的逃走格，不會出現在主指令列，拿來當認不得的樣本。
func TestPlanningUnimplemented(t *testing.T) {
	const notYet = 0x30
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
	if !commitThenYes(t, s) {
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

// 主指令列產生得出的每一個圖號，分派表都有 case：選下去不會落到 default（#67）。
// 圖號清單照 `CommandBar` 會 append 的那幾個常數。
func TestPlanningEveryBarIconHandled(t *testing.T) {
	icons := []int{CmdMissile, CmdLoad, CmdWeapon, CmdUnarmed, CmdCast, CmdSwitchAlt2,
		CmdSwitchAlt1, CmdSwitch, CmdDrop, CmdPickUp, CmdLand, CmdMountedMove, CmdMove, CmdRest}
	for _, icon := range icons {
		s := planningState(Member{HasWeapon: true, Weapon: Weapon{Kind: 0x20, Code: 1},
			Spells: []Spell{{Name: "Test"}}})
		s.Plan.Bar = []int{CmdYes, icon}
		s.Plan.Selected = 1
		s.Choose()
		if len(s.Plan.Unimplemented) != 0 {
			t.Errorf("圖號 %#x 落到 default：%#x", icon, s.Plan.Unimplemented)
		}
	}
}

// U-TURN 跳回 `0x1EF3A` 重排這一回合：`word_30D00` 加一（`0x1F04C`），但不經過回合
// 結束的法術倒數（`0x20F01`），身上法術的剩餘回合不能多扣，回合數也不變
// （RE-143／RE-145，#69）。
func TestUTurnReplansWithoutSpellCountdown(t *testing.T) {
	s := planningState(Member{HasWeapon: true, Weapon: Weapon{Kind: 0x20, Code: 1}})
	s.Rand = &fixedRoll{values: []int{50, 50, 50, 50}}
	s.Actors[0].SpellTurns = 3
	round, replans := s.Round.Number, s.Round.Replans
	s.Plan.Prompt = PromptNextRound
	s.Plan.Bar = s.NextRoundBar()
	s.Plan.Selected = 1 // U-TURN
	if c := s.ChooseNextRound(); c != NextRoundReplan {
		t.Fatalf("U-TURN 應該重新規劃，得到 %d", c)
	}
	if s.Actors[0].SpellTurns != 3 {
		t.Fatalf("U-TURN 不該扣法術回合，剩 %d", s.Actors[0].SpellTurns)
	}
	if s.Round.Number != round || s.Round.Replans != replans+1 {
		t.Fatalf("U-TURN 只加 word_30D00：回合 %d → %d，重排 %d → %d",
			round, s.Round.Number, replans, s.Round.Replans)
	}
}

// 每個新回合 `word_30D00` 在 `0x2119B` 歸零，所以第二回合起的第一次規劃不除基準；
// 只有同一回合 U-TURN 重排時才除（RE-145）。
func TestNewRoundResetsReplans(t *testing.T) {
	s := planningState(Member{HasWeapon: true, Weapon: Weapon{Kind: 0x20, Code: 1}})
	s.Rand = &fixedRoll{values: []int{50}}
	s.BeginPlanning(s.Rand)
	s.BeginPlanning(s.Rand)
	if s.Round.Replans != 1 {
		t.Fatalf("新回合排完序之後 word_30D00 應為 1，得到 %d", s.Round.Replans)
	}
}

// commitThenYes 按下收尾那一個 Return：原版寫回之後停在同一名、白框在 YES（RE-158），
// 再按一次 Return（YES）才換人。回報第二下之後這一輪是不是決定完了。
func commitThenYes(t *testing.T, s *State) bool {
	t.Helper()
	who := s.Current()
	if s.Choose() {
		t.Fatal("寫回之後不該直接換人")
	}
	if s.Current() != who || s.Plan.Prompt != PromptCommand || s.Plan.Selected != 0 || s.Plan.Bar[0] != CmdYes {
		t.Fatalf("寫回之後應停在同一名的主指令列、白框在 YES：輪到 %d（原本 %d）、層 %d、白框 %d",
			s.Current(), who, s.Plan.Prompt, s.Plan.Selected)
	}
	return s.Choose()
}

// 新回合一開始白框在 YES，直接按下就是沿用記錄裡上一回合那一手（RE-158）。
func TestPlanningYesRepeatsLastCommand(t *testing.T) {
	s := planningState(Member{Name: "Theodoric"})
	s.Actors[0].Command, s.Actors[0].Style, s.Actors[0].Guard = CmdUnarmed, 1, 2
	s.BeginTurn()
	if s.icon() != CmdYes {
		t.Fatalf("白框應停在 YES，停在 %#x", s.icon())
	}
	if !s.Choose() {
		t.Fatal("唯一一名按 YES，這一輪就該決定完")
	}
	if a := s.Actors[0]; a.Command != CmdUnarmed || a.Style != 1 || a.Guard != 2 {
		t.Fatalf("YES 不該改記錄：%#x／%d／%d", a.Command, a.Style, a.Guard)
	}
}

// RE-171：主指令摘要不依格上有人與否，人物條件只控制合成暫存重建。
func TestPlanningSummaryMarksTarget(t *testing.T) {
	s := planningState(Member{Name: "Poontang", HasWeapon: true, Weapon: Weapon{Kind: 0x20, Code: 1}})
	a := &s.Actors[0]
	s.Actors[PartySlots] = Actor{Side: SideMonster, Member: -1, X: a.X - 1, Y: a.Y, Threshold: 80, Stamina: 160}
	s.Actors[PartySlots].SetParts()
	selectIcon(t, s, CmdWeapon)
	s.Choose()
	s.Choose() // 面向上方那一格
	selectIcon(t, s, 0x2D)
	s.Choose()
	selectIcon(t, s, 0x32)
	s.Choose()
	if !s.Plan.SummaryShown || s.Plan.Summary != (Cell{Row: 3, Col: 4}) {
		t.Fatalf("應標出 (3,4)，得到 %v／%v", s.Plan.Summary, s.Plan.SummaryShown)
	}
	// 原版正常空摘要的回歸：格上沒有人仍標，取消或休息才清掉。
	s.Actors[PartySlots].Side = SideEmpty
	s.BeginTurn()
	if !s.Plan.SummaryShown {
		t.Fatal("空目標格仍應標示圓環")
	}
	s.LastCommand = CmdRest
	s.Actors[0].Command = CmdRest
	s.BeginTurn()
	if s.Plan.SummaryShown {
		t.Fatal("休息命令不應標示圓環")
	}
}

// 原版gui7 plan-09→10：前一人移動，換到指令0的Theodoric主列後摘要必須清掉。
func TestPlanningSummaryReloadsCurrentActorCommand(t *testing.T) {
	s := planningState(Member{Name: "Theodoric"})
	s.LastCommand = CmdMove
	s.Actors[0].Command = 0
	s.BeginTurn()
	if s.LastCommand != 0 || s.Plan.SummaryShown {
		t.Fatal("主列重建不應沿用前一人物／AI的移動指令")
	}
	s.LastCommand = CmdRest
	s.Actors[0].Command = CmdWeapon
	s.BeginTurn()
	if s.LastCommand != CmdWeapon || !s.Plan.SummaryShown {
		t.Fatal("主列應從目前人物重新載入已決定的指令")
	}
}
