package town

import (
	"errors"
	"testing"
)

// restText 是 REST 流程用得到的那幾筆 KOLDATA 記錄，內容與原版逐字相同。
func restText() *UIText {
	return &UIText{Schema: "kol-ui-text-data-v1", Banks: []uiBank{{
		Bank: BankUI, Records: []uiRecord{
			{Index: GreetingRecord, Text: `"What can I do for you?"`},
			{Index: WholePartyRecord, Text: "Whole party?"},
			{Index: QuitRecord, Text: "Do you wish to quit?"},
			{Index: StayRecord, Text: `"If you want to stay@`},
			{Index: FreeRoomRecord, Text: `~H0Djust grab an open room!"@`},
			{Index: HealOfferRecord, Text: `"If you need healing@`},
			{Index: CostPrefix, Text: "~H0Dit will cost you]"},
			{Index: CostSuffix, Text: `@~H0Dgold, interested?"@`},
			{Index: NotEnoughRecord, Text: `"Sorry, not enough@~H0Dgold crowns!"`},
			{Index: LookFineRecord, Text: `"You look fine!"`},
			{Index: HealedRecord, Text: `"Drink and be healed!"`},
			{Index: SaveDiskRecord, Text: "Save character to hard@disk?"},
		}}}}
}

// restSession 把玩家放進半途之家，隊伍是指定的那幾位。
func restSession(t *testing.T, flags [2]int, price, healRate int, party ...Member) *Session {
	t.Helper()
	s := playSession(t)
	s.UIText = restText()
	s.Buildings.Doors[0].IconFlags = flags
	s.Buildings.Doors[0].Price = price
	s.Buildings.Doors[0].HealRate = healRate
	s.Buildings.Doors[0].RecordID = 0x56
	if len(party) > 0 {
		s.Party = party
	}
	s.Walk(-1, 0)
	s.Walk(1, 0)
	s.Accept()   // 進門
	s.Continue() // 讀完敘述，圖示列展開
	s.Inside.Party = s.Party
	return s
}

func innSession(t *testing.T, price int, party ...Member) *Session {
	t.Helper()
	return restSession(t, [2]int{0x09, 0x00}, price, 0, party...)
}

func abbeySession(t *testing.T, rate int, party ...Member) *Session {
	t.Helper()
	return restSession(t, [2]int{0x41, 0x00}, 0, rate, party...)
}

// yes／no 按下大拇指那一格。
func yes(s *Session) { s.Inside.Selected = 0; s.Accept() }
func no(s *Session)  { s.Inside.Selected = 1; s.Accept() }

// 收費客棧：價錢問句是三行，答應之後扣錢並寫存檔。
func TestInnStayCharges(t *testing.T) {
	saved := 0
	s := innSession(t, 56, Member{Name: "Theodoric", Gold: 1009, Speaker: true})
	s.SaveGame = func() error { saved++; return nil }
	s.StartRest()
	if !s.Resting() {
		t.Fatal("按下 REST 應該進住宿流程")
	}
	want := []string{`"If you want to stay`, "it will cost you 56", `gold, interested?"`}
	for i, line := range want {
		if s.Inside.Dialog[i] != line {
			t.Fatalf("第 %d 行是 %q，應為 %q", i+1, s.Inside.Dialog[i], line)
		}
	}
	icons, _ := s.Icons()
	if len(icons) != 2 || icons[0] != IconSprite(IconYes) {
		t.Fatalf("價錢問句的圖示列是 %v", icons)
	}
	yes(s)
	if s.Party[0].Gold != 1009-56 {
		t.Fatalf("付完錢剩 %d，應為 %d", s.Party[0].Gold, 1009-56)
	}
	if saved != 1 {
		t.Fatalf("存檔被呼叫 %d 次，應為 1 次", saved)
	}
	if s.Party[0].Place != "HALFWAY HO" || s.Party[0].BuildingID != 0x56 {
		t.Fatalf("留宿之後所在地是 %q／建築 %d",
			s.Party[0].Place, s.Party[0].BuildingID)
	}
	if s.Resting() {
		t.Fatal("流程應該結束了")
	}
	if len(s.Party) != 1 {
		t.Fatalf("預設不離隊，隊伍卻剩 %d 人", len(s.Party))
	}
}

// 答 NO 不收錢也不存檔。
func TestInnStayDeclined(t *testing.T) {
	s := innSession(t, 56, Member{Name: "Theodoric", Gold: 1009, Speaker: true})
	s.SaveGame = func() error { t.Fatal("答 NO 不該存檔"); return nil }
	s.StartRest()
	no(s)
	if s.Party[0].Gold != 1009 || s.Resting() {
		t.Fatalf("答 NO 之後金幣 %d、流程 %v", s.Party[0].Gold, s.Resting())
	}
}

// 免費客棧只講一句，不問價錢，讀完就存檔。
func TestFreeInnStay(t *testing.T) {
	saved := 0
	s := innSession(t, 0, Member{Name: "Theodoric", Gold: 100, Speaker: true})
	s.SaveGame = func() error { saved++; return nil }
	s.StartRest()
	icons, _ := s.Icons()
	if len(icons) != 1 || icons[0] != IconSprite(IconSpacebar) {
		t.Fatalf("免費客棧不該有大拇指，圖示列是 %v", icons)
	}
	if got := s.Inside.Dialog[1]; got != `just grab an open room!"` {
		t.Fatalf("第二行是 %q", got)
	}
	s.Continue()
	if saved != 1 || s.Party[0].Gold != 100 {
		t.Fatalf("免費住宿存檔 %d 次、金幣 %d", saved, s.Party[0].Gold)
	}
}

// 錢不夠就講 KOLDATA 0x56，交易不成立。
func TestInnStayNotEnoughGold(t *testing.T) {
	s := innSession(t, 120, Member{Name: "Aventhalas", Gold: 59, Speaker: true})
	s.SaveGame = func() error { t.Fatal("買不起不該存檔"); return nil }
	s.StartRest()
	yes(s)
	if s.Inside.Dialog[0] != `"Sorry, not enough` {
		t.Fatalf("第一行是 %q", s.Inside.Dialog[0])
	}
	if s.Party[0].Gold != 59 {
		t.Fatalf("買不起卻扣了錢，剩 %d", s.Party[0].Gold)
	}
}

// 隊伍超過一人先問「全隊嗎」：答 YES 每個人各自付自己的錢。
func TestInnStayWholeParty(t *testing.T) {
	saves := 0
	s := innSession(t, 56,
		Member{Name: "Theodoric", Gold: 1009, Speaker: true},
		Member{Name: "Olear", Gold: 1777})
	s.SaveGame = func() error { saves++; return nil }
	s.StartRest()
	if s.Inside.Dialog[0] != "Whole party?" {
		t.Fatalf("第一問是 %q", s.Inside.Dialog[0])
	}
	yes(s) // 全隊
	yes(s) // 第一位的價錢
	yes(s) // 第二位的價錢
	if s.Party[0].Gold != 1009-56 || s.Party[1].Gold != 1777-56 {
		t.Fatalf("兩人的金幣是 %d／%d", s.Party[0].Gold, s.Party[1].Gold)
	}
	if saves != 2 {
		t.Fatalf("存檔 %d 次，應為 2 次", saves)
	}
}

// 答 NO 只有發言者住宿。
func TestInnStaySpeakerOnly(t *testing.T) {
	s := innSession(t, 56,
		Member{Name: "Theodoric", Gold: 1009},
		Member{Name: "Olear", Gold: 1777, Speaker: true})
	s.SaveGame = func() error { return nil }
	s.StartRest()
	no(s) // 不是全隊
	yes(s)
	if s.Party[0].Gold != 1009 || s.Party[1].Gold != 1777-56 {
		t.Fatalf("兩人的金幣是 %d／%d", s.Party[0].Gold, s.Party[1].Gold)
	}
	if s.Resting() {
		t.Fatal("只有一位要住，流程應該結束了")
	}
}

// 原版住宿模式：多一問「存到硬碟嗎」，住完離隊，隊伍空了問是否離開遊戲。
func TestFaithfulInnStayRetires(t *testing.T) {
	s := innSession(t, 56, Member{Name: "Theodoric", Gold: 1009, Speaker: true})
	s.FaithfulInnStay = true
	s.SaveGame = func() error { return nil }
	s.StartRest()
	yes(s)
	if s.Inside.Dialog[0] != "Save character to hard" {
		t.Fatalf("硬碟問句第一行是 %q", s.Inside.Dialog[0])
	}
	yes(s)
	if len(s.Party) != 0 {
		t.Fatalf("原版模式住完應該離隊，隊伍還有 %d 人", len(s.Party))
	}
	if s.Inside.Dialog[0] != "Do you wish to quit?" {
		t.Fatalf("隊伍空了應該問是否離開，畫的是 %q", s.Inside.Dialog[0])
	}
	yes(s)
	if !s.QuitAsked() {
		t.Fatal("答 YES 應該記下要離開遊戲")
	}
}

// 存檔失敗就整筆退回：不扣錢、不改所在地、不離隊。
func TestInnStaySaveFailureRollsBack(t *testing.T) {
	s := innSession(t, 56, Member{Name: "Theodoric", Gold: 1009, Speaker: true})
	s.FaithfulInnStay = true
	boom := errors.New("磁碟滿了")
	s.SaveGame = func() error { return boom }
	s.StartRest()
	yes(s) // 價錢
	yes(s) // 硬碟問句
	if s.Party[0].Gold != 1009 || s.Party[0].Place != "" || len(s.Party) != 1 {
		t.Fatalf("存檔失敗卻改了狀態：金幣 %d、所在地 %q、人數 %d",
			s.Party[0].Gold, s.Party[0].Place, len(s.Party))
	}
	if !errors.Is(s.SaveError, boom) {
		t.Fatalf("沒有記下存檔失敗的原因：%v", s.SaveError)
	}
	if s.Inside.Dialog[0] != SaveFailedZh {
		t.Fatalf("存檔失敗的訊息是 %q", s.Inside.Dialog[0])
	}
}

// 修道院看到沒受傷的人只講一句，不收錢也不問。
func TestAbbeyLooksFine(t *testing.T) {
	s := abbeySession(t, 10, Member{Name: "Theodoric", Gold: 1009, Speaker: true})
	s.StartRest()
	if s.Resting() {
		t.Fatal("沒受傷不該進問答")
	}
	if s.Inside.Dialog[0] != `"You look fine!"` {
		t.Fatalf("畫的是 %q", s.Inside.Dialog[0])
	}
	if s.Party[0].Gold != 1009 {
		t.Fatalf("不該收錢，剩 %d", s.Party[0].Gold)
	}
}

// 修道院的價錢是傷勢乘費率；付完傷勢歸零。
func TestAbbeyHealCharges(t *testing.T) {
	s := abbeySession(t, 10, Member{Name: "Warder", Gold: 467, Wounds: 20, Speaker: true})
	s.SaveGame = func() error { t.Fatal("治療不存檔"); return nil }
	s.StartRest()
	if s.Inside.Dialog[1] != "it will cost you 200" {
		t.Fatalf("價錢那一行是 %q", s.Inside.Dialog[1])
	}
	yes(s)
	if s.Party[0].Gold != 467-200 || s.Party[0].Wounds != 0 {
		t.Fatalf("治療後金幣 %d、傷勢 %d", s.Party[0].Gold, s.Party[0].Wounds)
	}
	if s.Inside.Dialog[0] != `"Drink and be healed!"` {
		t.Fatalf("治好之後畫的是 %q", s.Inside.Dialog[0])
	}
}

// 治療買不起同樣講 0x56，傷勢不變。
func TestAbbeyHealNotEnoughGold(t *testing.T) {
	s := abbeySession(t, 12, Member{Name: "Poontang", Gold: 5, Wounds: 1, Speaker: true})
	s.StartRest()
	yes(s)
	if s.Party[0].Wounds != 1 || s.Party[0].Gold != 5 {
		t.Fatalf("買不起卻改了狀態：金幣 %d、傷勢 %d", s.Party[0].Gold, s.Party[0].Wounds)
	}
}

// 中文模式的價錢句併成一行：中文的框放得下，拆開反而讀不順。
func TestPriceLineJoinsInChinese(t *testing.T) {
	s := innSession(t, 56, Member{Name: "Theodoric", Gold: 1009, Speaker: true})
	s.Trans = &Translations{UI: map[string]string{
		"55": "「你要住下的話，",
		"83": "要付",
		"84": "枚金幣，願意嗎？」",
	}}
	s.StartRest()
	want := []string{"「你要住下的話，", "要付 56 枚金幣，願意嗎？」"}
	if len(s.Inside.Dialog) != len(want) {
		t.Fatalf("中文的價錢句是 %q", s.Inside.Dialog)
	}
	for i, line := range want {
		if s.Inside.Dialog[i] != line {
			t.Fatalf("第 %d 行是 %q，應為 %q", i+1, s.Inside.Dialog[i], line)
		}
	}
}
