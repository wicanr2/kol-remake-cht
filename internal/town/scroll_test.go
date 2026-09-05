package town

import "testing"

// hostelOrder 是城 7 The Hostel 那一棟的教派（RE-121），欄位直接抄自建築 payload。
func hostelOrder() *Order {
	return &Order{
		Name: "Blue Gem", Title: "Illusionist", Intellect: 90,
		Record:      "8132a501cad00000b4b1959a5ee4ce80",
		Token:       "Blue Chain",
		TokenRecord: "a132a500d014b88201ae000002007000",
	}
}

// hostelSpells 是同一棟賣的前三筆。
func hostelSpells() []SpellOffer {
	return []SpellOffer{
		{Index: 0, Name: "Tyanalon", Price: 340, Effect: 6, TurnCode: 0,
			PowerCode: 10, Guards: []int{48, 0, 0, 0},
			Record: "8a642e0b1ee00c813000000006000a22"},
		{Index: 1, Name: "Tyatwelon", Price: 330, Effect: 5, TurnCode: 0,
			PowerCode: 10, Guards: []int{48, 0, 0, 0},
			Record: "9a6434b958f700643000000005000a21"},
		{Index: 2, Name: "Kelnalon", Price: 340, Effect: 6, TurnCode: 0,
			PowerCode: 10, Guards: []int{64, 0, 0, 0},
			Record: "85958e0b1ee00c814000000006000a22"},
	}
}

// scrollText 是 SCROLL 那一段用得到的記錄。
func scrollText() *UIText {
	t := restText()
	t.Banks[0].Records = append(t.Banks[0].Records,
		uiRecord{Index: JoinOfferRecord, Text: `"If you want to join@`},
		uiRecord{Index: JoinTheRecord, Text: "~H0Dthe]"},
		uiRecord{Index: JoinOrderRecord, Text: " Order]"},
		uiRecord{Index: MemberRecord, Text: `"You are now a member of@`},
		uiRecord{Index: SpellOfferRecord, Text: `"If you want this spell@`},
		uiRecord{Index: LearnSpellRecord, Text: "Learn a spell"},
		uiRecord{Index: ModifySpellRecord, Text: "Modify a spell"},
		uiRecord{Index: SpellbookFullRecord, Text: `"You've learned enough!"`},
		uiRecord{Index: CostPrefix, Text: "~H0Dit will cost you]"},
		uiRecord{Index: CostSuffix, Text: `@~H0Dgold, interested?"@`})
	return t
}

// scrollSession 把玩家放進一棟有 SCROLL 的建築，隊長是 Theodoric。
func scrollSession(t *testing.T, gold, intellect int) *Session {
	t.Helper()
	s := restSession(t, [2]int{0x00, 0x01}, 450, 0,
		Member{Name: "ddduhh", Gold: gold, Speaker: true})
	s.UIText = scrollText()
	sheet := theodoric()
	sheet.Stats[StatIntellect] = intellect
	s.Party[0].Sheet = sheet
	s.Buildings.Doors[0].Order = hostelOrder()
	s.Buildings.Doors[0].Spells = hostelSpells()
	s.door = &s.Buildings.Doors[0]
	return s
}

// 智力夠就先問要不要入會；答應之後扣掉入會費、教派記進角色，信物放在 +0x170。
func TestScrollJoinsTheOrder(t *testing.T) {
	s := scrollSession(t, 1000, 109)
	s.StartScroll()
	if !s.Scrolling() {
		t.Fatal("按下 SCROLL 應該進這一格")
	}
	if len(s.Inside.Dialog) == 0 {
		t.Fatal("入會的問句是空的")
	}
	yes(s)
	if s.Party[0].Gold != 550 {
		t.Fatalf("錢包是 %d，應為 1000 − 450 ＝ 550", s.Party[0].Gold)
	}
	if s.Party[0].Order == nil || s.Party[0].Order.Name != "Blue Gem" {
		t.Fatalf("教派是 %+v", s.Party[0].Order)
	}
	if s.Party[0].Token == nil || s.Party[0].Token.Name != "Blue Chain" {
		t.Fatalf("信物是 %+v", s.Party[0].Token)
	}
	if s.Party[0].Sheet.Order != "Blue Gem" || s.Party[0].Sheet.Title != "Illusionist" {
		t.Fatalf("資料頁那一列是 %q／%q", s.Party[0].Sheet.Order,
			s.Party[0].Sheet.Title)
	}
	// 讀完那一句回選單：只有一列，因為還沒學過法術。
	s.Continue()
	if s.Inside.List == nil || len(s.Inside.List.Rows) != 1 ||
		s.Inside.List.Rows[0] != "Learn a spell" {
		t.Fatalf("選單是 %+v", s.Inside.List)
	}
}

// 智力不夠就入不了會，但照樣賣法術——原版 `0x17321` 直接跳到清單那一段。
func TestScrollSkipsJoinWhenDim(t *testing.T) {
	s := scrollSession(t, 1000, 80)
	s.StartScroll()
	if s.Inside.List == nil || s.Inside.List.Rows[0] != "Tyanalon" {
		t.Fatalf("清單是 %+v", s.Inside.List)
	}
	if s.Party[0].Order != nil {
		t.Fatalf("智力不夠不該入會：%+v", s.Party[0].Order)
	}
}

// 買一筆法術：扣掉價錢，法術進角色的空槽。
func TestScrollBuysASpell(t *testing.T) {
	s := scrollSession(t, 1000, 80)
	s.StartScroll()
	s.MoveScroll(1) // Tyatwelon，330
	yes(s)
	if len(s.Inside.Dialog) == 0 {
		t.Fatal("問句是空的")
	}
	yes(s)
	if s.Party[0].Gold != 670 {
		t.Fatalf("錢包是 %d，應為 1000 − 330 ＝ 670", s.Party[0].Gold)
	}
	if len(s.Party[0].Spells) != 1 || s.Party[0].Spells[0].Name != "Tyatwelon" {
		t.Fatalf("法術是 %+v", s.Party[0].Spells)
	}
	if s.Party[0].Spells[0].Effect != 5 || s.Party[0].Spells[0].PowerCode != 10 {
		t.Fatalf("法術的欄位是 %+v", s.Party[0].Spells[0])
	}
	// 買完回清單，還可以再買一筆。
	if s.Inside.List == nil || s.Inside.List.Rows[0] != "Tyanalon" {
		t.Fatalf("買完應該回清單：%+v", s.Inside.List)
	}
}

// 錢不夠就講一句，法術不給、錢不扣。
func TestScrollRefusesWhenBroke(t *testing.T) {
	s := scrollSession(t, 100, 80)
	s.StartScroll()
	yes(s)
	yes(s)
	if s.Party[0].Gold != 100 || len(s.Party[0].Spells) != 0 {
		t.Fatalf("錢不夠不該成交：%d／%+v", s.Party[0].Gold, s.Party[0].Spells)
	}
}

// 十六格法術學滿了就講記錄 0x6D。
func TestScrollSpellbookFull(t *testing.T) {
	s := scrollSession(t, 10000, 80)
	for i := 0; i < SpellSlots; i++ {
		s.Party[0].Spells = append(s.Party[0].Spells, Spell{Name: "X"})
	}
	s.StartScroll()
	yes(s)
	if len(s.Party[0].Spells) != SpellSlots {
		t.Fatalf("學滿了不該再收：%d", len(s.Party[0].Spells))
	}
	if len(s.Inside.Dialog) == 0 ||
		s.Inside.Dialog[0] != `"You've learned enough!"` {
		t.Fatalf("講的是 %v", s.Inside.Dialog)
	}
}

// 已經有法術的角色，選單才有第二列；那一列還沒實作，按下去只記一筆。
func TestScrollMenuHasSpellbookRow(t *testing.T) {
	s := scrollSession(t, 1000, 109)
	s.Party[0].Order = hostelOrder()
	s.Party[0].Spells = []Spell{{Name: "Tyanalon"}}
	s.StartScroll()
	if s.Inside.List == nil || len(s.Inside.List.Rows) != 2 {
		t.Fatalf("選單是 %+v", s.Inside.List)
	}
	s.MoveScroll(1)
	yes(s)
	if len(s.Unimplemented) != 1 {
		t.Fatalf("法術書那一頁應該記一筆未實作：%v", s.Unimplemented)
	}
}
