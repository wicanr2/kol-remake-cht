package town

import "testing"

func questDoor() *Door {
	return &Door{Index: 14, IconFlags: [2]int{1, QuestFlagBit},
		QuestSite: 4, QuestFinal: false, Name: "The Assembly Building"}
}

func questText() *BuildingText {
	return &BuildingText{Door: 14, Messages: []Message{
		{Index: 0, Display: []string{"進門一", "進門二"}},
		{Index: 1, Display: []string{"進門三", "進門四"}},
		{Index: 2, Display: []string{"敘述一", "敘述二"}},
		{Index: 3, Display: []string{"敘述三", "敘述四"}},
		{Index: 4, Display: []string{"敘述五", "敘述六"}},
		{Index: 5, Display: []string{"敘述七", "敘述八"}},
	}}
}

// 沒接過就講兩頁敘述，再問「要接嗎」；答應之後旗標從 0 變成 1。
func TestOfferQuestAccept(t *testing.T) {
	s := &Session{
		Party:  []Member{{Name: "甲", Speaker: true, Quests: make([]int, QuestFlagBytes), Progress: make([]int, 5)}},
		Inside: &Inn{}, door: questDoor(), Trans: &Translations{},
		UIText: &UIText{}, Mode: ModeInside,
	}
	s.Text = &Text{Towns: []townText{{Town: 0, Buildings: []BuildingText{*questText()}}}}
	s.Outdoor = &Outdoor{Town: 0}

	if !s.offerQuest() {
		t.Fatal("這一棟應該給任務")
	}
	if got := s.Inside.Dialog; len(got) != 4 || got[0] != "敘述一" {
		t.Fatalf("第一頁是 %v", got)
	}
	s.nextQuestPage()
	if got := s.Inside.Dialog; len(got) != 4 || got[0] != "敘述五" {
		t.Fatalf("第二頁是 %v", got)
	}
	s.nextQuestPage()
	if len(s.Inside.Icons) != 2 || s.Inside.Icons[0] != IconSprite(IconYes) {
		t.Fatalf("問句的圖示列是 %v", s.Inside.Icons)
	}
	s.questAnswer(true)
	if QuestFlag(s.Party[0].Quests, 4) != QuestTaken {
		t.Fatalf("旗標應該變成接了，得到 %d", QuestFlag(s.Party[0].Quests, 4))
	}
	if s.TookQuest == nil || s.TookQuest.Site != 4 {
		t.Fatalf("接下的任務記成 %+v", s.TookQuest)
	}
	if s.quest != nil {
		t.Fatal("流程應該結束了")
	}
}

// 拒絕不動旗標。
func TestOfferQuestRefuse(t *testing.T) {
	s := &Session{
		Party:  []Member{{Name: "甲", Speaker: true, Quests: make([]int, QuestFlagBytes), Progress: make([]int, 5)}},
		Inside: &Inn{}, door: questDoor(), Trans: &Translations{},
		UIText: &UIText{}, Mode: ModeInside,
		Outdoor: &Outdoor{Town: 0},
	}
	s.Text = &Text{Towns: []townText{{Town: 0, Buildings: []BuildingText{*questText()}}}}
	s.offerQuest()
	s.nextQuestPage()
	s.nextQuestPage()
	s.questAnswer(false)
	if QuestFlag(s.Party[0].Quests, 4) != QuestNone {
		t.Fatal("拒絕不該動旗標")
	}
	if s.TookQuest != nil {
		t.Fatalf("拒絕不該記成接下了：%+v", s.TookQuest)
	}
}

// 已經接了或完成了只再講一次敘述，不問；回報過了就當一般話題。
func TestOfferQuestByFlag(t *testing.T) {
	for _, tc := range []struct {
		name  string
		flag  int
		takes bool
		asks  bool
	}{
		{"沒接過", QuestNone, true, true},
		{"接了", QuestTaken, true, false},
		{"完成了", QuestDone, true, false},
		{"回報過", QuestReported, false, false},
	} {
		s := &Session{
			Party:  []Member{{Name: "甲", Speaker: true, Quests: make([]int, QuestFlagBytes), Progress: make([]int, 5)}},
			Inside: &Inn{}, door: questDoor(), Trans: &Translations{},
			UIText: &UIText{}, Mode: ModeInside, Outdoor: &Outdoor{Town: 0},
		}
		s.Text = &Text{Towns: []townText{{Town: 0, Buildings: []BuildingText{*questText()}}}}
		SetQuestFlag(s.Party[0].Quests, 4, tc.flag)
		if got := s.offerQuest(); got != tc.takes {
			t.Fatalf("%s：接手回 %v，期望 %v", tc.name, got, tc.takes)
		}
		if !tc.takes {
			continue
		}
		s.nextQuestPage()
		s.nextQuestPage()
		asked := s.quest != nil && s.quest.asking
		if asked != tc.asks {
			t.Fatalf("%s：問了 %v，期望 %v", tc.name, asked, tc.asks)
		}
	}
}

// 最終任務要先回報滿 0x17 件才接得到。
func TestOfferFinalQuestThreshold(t *testing.T) {
	door := questDoor()
	door.QuestFinal = true
	s := &Session{
		Party:  []Member{{Name: "甲", Speaker: true, Quests: make([]int, QuestFlagBytes), Progress: make([]int, 5)}},
		Inside: &Inn{}, door: door, Trans: &Translations{},
		UIText: &UIText{}, Mode: ModeInside, Outdoor: &Outdoor{Town: 0},
	}
	s.Text = &Text{Towns: []townText{{Town: 0, Buildings: []BuildingText{*questText()}}}}
	if s.offerQuest() {
		t.Fatal("回報數不足時不該給最終任務")
	}
	s.Party[0].Progress[0] = FinalQuestReports
	if !s.offerQuest() {
		t.Fatal("回報滿 0x17 件就該給")
	}
}
