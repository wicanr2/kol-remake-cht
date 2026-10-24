package battle

import "testing"

func TestBodyPanelFollowsPlanningInspection(t *testing.T) {
	s := planningState(Member{Name: "測試", Female: true, HasWeapon: true, Weapon: Weapon{Kind: 0x20, Code: 1}, Size: 80, Health: 80, Strength: 80})
	a := &s.Actors[0]
	s.Actors[PartySlots] = Actor{Side: SideMonster, Member: -1, X: a.X - 1, Y: a.Y, Threshold: 80, Stamina: 160}
	s.Actors[PartySlots].SetParts()
	check := func(kind int, visible bool) {
		t.Helper()
		d, ok := s.BodyPanel()
		if ok != visible || ok && d.Kind != kind {
			t.Fatalf("人形=%+v 顯示=%v，預期種類=%d 顯示=%v", d, ok, kind, visible)
		}
		if s.Cursor != 0 {
			t.Fatal("檢視別人不得改動行動者")
		}
	}
	check(1, true)
	selectIcon(t, s, CmdWeapon)
	s.Choose()
	check(2, true) // 初始面前格有敵人。
	s.MoveCursor(0, 1, false)
	check(0, false)
	s.MoveCursor(0, -1, true)
	check(2, true)
	s.Choose()
	check(2, true) // 選完目標仍保留被檢視者。
	selectIcon(t, s, 0x2D)
	s.Choose()
	check(2, true)
	selectIcon(t, s, IconAttackMiddle)
	s.Choose()
	check(2, true)
	selectIcon(t, s, 0x32)
	s.Choose()
	check(1, true) // 寫回後停在同一名，人形框回到自己（`0x1F16A`）。
	s.Choose()     // YES 換人：這一輪決定完。
	check(0, false)
	s.Round.Turn = 0 // 開始新一輪後才能重新輪到這名隊員。
	s.BeginTurn()
	check(1, true)
	selectIcon(t, s, CmdWeapon)
	s.Choose()
	selectIcon(t, s, CmdCancel)
	s.Choose()
	check(1, true)
	s.Plan.Prompt = PromptExecuting
	check(0, false)
}

func TestBodyPanelDoesNotExposeHiddenActors(t *testing.T) {
	s := planningState(Member{Size: 80, Health: 80, Strength: 80})
	s.Actors[0].X, s.Actors[0].Y = 10, 10
	s.Actors[PartySlots] = Actor{Side: SideMonster, X: 10, Y: 14, Threshold: 80, Stamina: 160}
	s.Actors[PartySlots].SetParts()
	s.Field.Tiles[10*FieldSide+12] = 0x29
	s.Plan.Prompt = PromptTarget
	s.Plan.Cursor = TargetCursor{Mode: CursorRanged, Row: 4, Col: 8}
	s.inspectBodyTarget()
	if _, ok := s.BodyPanel(); ok {
		t.Fatal("不可透過牆顯示敵人人形")
	}
}

func TestBodyPanelRejectsInvalidDenominators(t *testing.T) {
	if _, err := BodyDamageRects(0, [6]int{}, [6]int{}); err == nil {
		t.Fatal("零部位上限須回報錯誤")
	}
	if _, err := BodyFatigueBoundary(0, 0); err == nil {
		t.Fatal("零體力須回報錯誤")
	}
}

// 問部位那一頁的人形框換成法術的目標（`sub_2818D:0x28484`，RE-153）。
func TestBodyPanelShowsAreaTarget(t *testing.T) {
	s := fleeState()
	s.Stage = StageField
	s.Plan.Prompt = PromptExecuting
	if _, show := s.BodyPanel(); show {
		t.Fatal("執行段一般時候人形框不畫")
	}
	s.Plan.Prompt = PromptArea
	s.Area = &AreaQuestion{Caster: 0, Target: 1, Power: 3}
	d, show := s.BodyPanel()
	if !show {
		t.Fatal("問部位那一頁應該畫目標的人形")
	}
	want := Parts(s.Actors[1].Threshold)
	if d.Maximum != want {
		t.Fatalf("人形應該是目標那一名：%v，應為 %v", d.Maximum, want)
	}
	if d.FatigueBar {
		t.Fatal("問部位那一頁只畫人形與傷勢，不畫疲勞直條")
	}
}

func TestTargetInspectionKeepsNonActorTileAndClearsEmpty(t *testing.T) {
	s := planningState(Member{Size: 80, Health: 80, Strength: 80})
	a := &s.Actors[0]
	a.X, a.Y, a.Facing = 10, 10, 0
	s.Actors[PartySlots] = Actor{Side: SideMonster, Member: -1, X: 9, Y: 10, Threshold: 80, Stamina: 160}
	s.Actors[PartySlots].SetParts()
	s.Actors[PartySlots].Head = 0
	s.Actors[PartySlots+1] = Actor{Side: SideMonster, Member: -1, X: 9, Y: 11, Threshold: 80, Stamina: 160}
	s.Actors[PartySlots+1].SetParts()
	s.beginTarget(0, CursorAdjacent, PromptMoveForm)
	if s.Plan.BodyActor != 0 {
		t.Fatalf("新進倒地格須空框，得到 %d", s.Plan.BodyActor)
	}
	s.MoveCursor(0, 1, false)
	if s.Plan.BodyActor != PartySlots+2 {
		t.Fatalf("站立者資訊錯誤：%d", s.Plan.BodyActor)
	}
	s.MoveCursor(0, -1, false)
	if s.Plan.BodyActor != PartySlots+2 {
		t.Fatalf("倒地格須保留前框：%d", s.Plan.BodyActor)
	}
	s.MoveCursor(0, -1, false)
	if s.Plan.BodyActor != 0 {
		t.Fatalf("空格須清框：%d", s.Plan.BodyActor)
	}
}

func TestTargetInspectionUsesStandingSpriteOverDownActor(t *testing.T) {
	s := planningState(Member{Size: 80, Health: 80, Strength: 80})
	s.Actors[0].X, s.Actors[0].Y, s.Actors[0].Facing = 10, 10, 0
	s.Actors[1] = Actor{Side: SideParty, X: 9, Y: 10, Threshold: 80}
	s.Actors[PartySlots] = Actor{Side: SideMonster, Member: -1, X: 9, Y: 10, Threshold: 80, Stamina: 160}
	s.Actors[PartySlots].SetParts()
	s.beginTarget(0, CursorAdjacent, PromptMoveForm)
	if s.Plan.BodyActor != PartySlots+1 {
		t.Fatalf("應檢視實際站立精靈：%d", s.Plan.BodyActor)
	}
}

func TestTargetInspectionObjectKeepsPreviousPanel(t *testing.T) {
	s := planningState(Member{Size: 80, Health: 80, Strength: 80})
	s.Actors[0].X, s.Actors[0].Y = 10, 10
	s.Plan.Prompt, s.Plan.BodyActor = PromptTarget, 1
	s.Plan.Cursor = TargetCursor{Mode: CursorAdjacent, Row: 3, Col: 3}
	s.Ground[0] = GroundItem{Name: "Dagger", X: 9, Y: 9}
	s.inspectBodyTarget()
	if s.Plan.BodyActor != 1 {
		t.Fatalf("物品格須保留前框：%d", s.Plan.BodyActor)
	}
	s.Ground[0] = GroundItem{}
	s.inspectBodyTarget()
	if s.Plan.BodyActor != 0 {
		t.Fatalf("空格須清框：%d", s.Plan.BodyActor)
	}
}
