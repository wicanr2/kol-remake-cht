package town

import "testing"

func topicSession(t *testing.T) *Session {
	t.Helper()
	s := playSession(t)
	s.Buildings.Doors[0].IconFlags = [2]int{0x09, 0x00} // 旅店：LISTEN／MOUTH ＋ REST／COMPANION
	s.Topics = &Topics{Schema: "kol-town-keywords-data-v1", Towns: []townTopics{{
		Town: 7,
		Keywords: []Keyword{
			{Prefix: "elve", Code: 8, Topic: 8, Category: 0},
			{Prefix: "bron", Code: 3, Topic: 3, Category: 0},
			{Prefix: "ghou", Code: 0x4C, Topic: 12, Category: 1},
		},
		Names:  []string{"elmu", "benjamin"},
		AskWho: map[string]int{"12": 1},
		Buildings: []BuildingTopics{{Door: 5, Slots: []TopicSlot{
			{Slot: 0, Raw: 0x40, Category: 1}, // 沒有關鍵字：LISTEN 的傳聞
			{Slot: 1, Raw: 8, Topic: 8},       // elve：要用 MOUTH 問
			{Slot: 2, Raw: 3, Topic: 3},       // bron：要用 MOUTH 問
			{Slot: 3, Raw: 0x40, Category: 1},
			{Slot: 4, Raw: 8, Topic: 8},
			{Slot: 5, Raw: 0x40, Category: 1},
		}}},
	}}}
	s.UIText = &UIText{Schema: "kol-ui-text-data-v1", Banks: []uiBank{{
		Bank: BankUI, Records: []uiRecord{
			{Index: 0x2F, Text: `"I heard]`},
			{Index: 0x30, Text: "talking about]"},
			{Index: 0x2E, Text: `"Any other gossip?"`},
			{Index: 0x5B, Text: `"Not lately."@`},
			{Index: 0x9A, Text: `"What do you know?"`},
			{Index: 0x9B, Text: `"How are things?"`},
			{Index: 0x9E, Text: "that"},
			{Index: 0x9F, Text: "them"},
		},
	}}}
	b := s.Text.Building(7, 5)
	b.Messages = append(b.Messages,
		Message{Index: 18, Display: []string{"You ask,"}},
		Message{Index: 19, Display: []string{"I don't know."}},
		Message{Index: 20, Display: []string{"Elves live in the Klvar Wood."}},
		Message{Index: 22, Display: []string{"Rain again tomorrow."}},
		Message{Index: 23, Display: []string{"The roof needs fixing."}},
		Message{Index: 25, Display: []string{"I hunted Bronk as a youth."}})
	return s
}

func enterInn(t *testing.T, s *Session) {
	t.Helper()
	s.Walk(-1, 0)
	s.Walk(1, 0)
	s.Accept()
	if s.Mode != ModeInside {
		t.Fatal("應該進了建築")
	}
}

// 進門先只有 SPACEBAR；按過之後才換成這棟建築的圖示列。
func TestInteriorIconBarAppearsAfterSpacebar(t *testing.T) {
	s := topicSession(t)
	enterInn(t, s)
	if icons, _ := s.Icons(); len(icons) != 1 || icons[0] != IconSprite(IconSpacebar) {
		t.Fatalf("進門的圖示列是 %v", icons)
	}
	s.Continue()
	icons, cursor := s.Icons()
	want := []int{IconDoor, IconMirror, IconListen, IconMouth, IconRest, IconCompanion}
	if len(icons) != len(want) || cursor != 0 {
		t.Fatalf("圖示列是 %v／白框 %d", icons, cursor)
	}
	for i, id := range want {
		if icons[i] != IconSprite(id) {
			t.Fatalf("第 %d 格是 %d，應為 %d", i, icons[i], IconSprite(id))
		}
	}
}

// LISTEN 挑到非空槽就先講開場白再講內容，而且同一個槽不會講第二次。
func TestListenTellsEachTopicOnce(t *testing.T) {
	s := topicSession(t)
	enterInn(t, s)
	s.Continue()
	s.Rand = func(int) int { return 0 }
	if !s.Listen() {
		t.Fatal("第 0 槽是沒有關鍵字的傳聞，LISTEN 應該講得出來")
	}
	got := s.Inside.Dialog
	if len(got) != 2 || got[0] != "You ask," || got[1] != `"How are things?"` {
		t.Fatalf("LISTEN 第一段是 %v", got)
	}
	if icons, _ := s.Icons(); len(icons) != 1 || icons[0] != IconSprite(IconSpacebar) {
		t.Fatalf("等玩家讀完時圖示列應該只剩 SPACEBAR，實際 %v", icons)
	}
	s.Continue()
	if got := s.Inside.Dialog; len(got) != 1 ||
		got[0] != "Elves live in the Klvar Wood." {
		t.Fatalf("按空白鍵之後是 %v", got)
	}
	s.Continue() // 沒有下一段了，圖示列換回來
	if icons, _ := s.Icons(); len(icons) != 6 {
		t.Fatalf("讀完之後圖示列應該回到六格，實際 %v", icons)
	}
	// 起點那一格講過了，就往回找：5 → 4 → 3…，第 5 槽也是傳聞。
	if !s.Listen() {
		t.Fatal("第 5 槽是傳聞，應該往回找得到")
	}
	s.Continue()
	if got := s.Inside.Dialog; len(got) != 1 ||
		got[0] != "I hunted Bronk as a youth." {
		t.Fatalf("第二次 LISTEN 的內容是 %v", got)
	}
	s.Continue()
	// 還有第 3 槽這個傳聞。
	if !s.Listen() {
		t.Fatal("第 3 槽也是傳聞，應該講得出來")
	}
	s.Continue()
	if got := s.Inside.Dialog; len(got) != 1 || got[0] != "The roof needs fixing." {
		t.Fatalf("第三次 LISTEN 的內容是 %v", got)
	}
	s.Continue()
	// 三個傳聞都講完，六格繞一圈之後講「都說完了」。
	if s.Listen() {
		t.Fatal("沒有話題了，不該再講出內容")
	}
	if got := s.Inside.Dialog; len(got) != 2 || got[1] != `"Any other gossip?"` {
		t.Fatalf("講完之後的第一段是 %v", got)
	}
	s.Continue()
	if got := s.Inside.Dialog; len(got) != 1 || got[0] != `"Not lately."` {
		t.Fatalf("講完之後的第二段是 %v", got)
	}
}

// MOUTH 用關鍵字字首比對；命中這棟建築的槽就講那一則。
func TestAskMatchesKeywordPrefix(t *testing.T) {
	s := topicSession(t)
	enterInn(t, s)
	s.Continue()
	s.Ask("what are bronks?")
	s.Continue()
	if got := s.Inside.Dialog; len(got) != 1 || got[0] != "Rain again tomorrow." {
		t.Fatalf("問 bronk 得到 %v", got)
	}
	s.Continue()
	// 霍布（城 7）不給「去問誰」的提示，因此照樣是「我不知道」。
	s.Ask("tell me about ghouls")
	s.Continue()
	if got := s.Inside.Dialog; len(got) != 1 || got[0] != "I don't know." {
		t.Fatalf("在霍布問 ghoul 得到 %v", got)
	}
	s.Continue()
	// 整張表都沒中就講「我不知道」。
	s.Ask("where is the cheese")
	s.Continue()
	if got := s.Inside.Dialog; len(got) != 1 || got[0] != "I don't know." {
		t.Fatalf("問無關的字得到 %v", got)
	}
}

// 其他城鎮會指出誰知道這個話題；霍布不給這個提示（原版 `word_2DF2C == 7` 那條）。
func TestAskWhoOnlyOutsideHobe(t *testing.T) {
	topics := &Topics{Schema: "kol-town-keywords-data-v1", Towns: []townTopics{
		{Town: 1, Names: []string{"elmu", "benjamin"}, AskWho: map[string]int{"12": 1}},
		{Town: 7, Names: []string{"elmu", "benjamin"}, AskWho: map[string]int{"12": 1}},
	}}
	if got := topics.AskWho(1, 12); got != "benjamin" {
		t.Fatalf("城 1 的話題 12 該去問 benjamin，得到 %q", got)
	}
	if got := topics.AskWho(7, 12); got != "" {
		t.Fatalf("霍布不該給提示，得到 %q", got)
	}
	if got := topics.AskWho(1, 3); got != "" {
		t.Fatalf("沒有人知道的話題該回空字串，得到 %q", got)
	}
}

// 文字庫的版面標記要剝掉，內容不能動。
func TestUITextStripsMarkup(t *testing.T) {
	bank := &UIText{Schema: "kol-ui-text-data-v1", Banks: []uiBank{{
		Bank: BankUI, Records: []uiRecord{
			{Index: 1, Text: `"What news have you of@~H0Dthe world?"`},
			{Index: 2, Text: `"I heard]`},
		},
	}}}
	if got := bank.Record(BankUI, 1); got != `"What news have you of the world?"` {
		t.Fatalf("剝標記之後是 %q", got)
	}
	if got := bank.Record(BankUI, 2); got != `"I heard` {
		t.Fatalf("剝接續記號之後是 %q", got)
	}
}
