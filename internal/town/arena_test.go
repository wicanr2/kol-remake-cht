package town

import "testing"

func arenaMember(name string, rank, offence int) Member {
	return Member{Name: name, Sheet: &Sheet{Rank: rank,
		Weapons: []WeaponRow{{OffenceRaw: offence, DefenceRaw: offence}}}}
}

// 帳面位階與實得一樣就不用升；到 24 也不用再打（RE-108）。
func TestArenaCandidates(t *testing.T) {
	party := []Member{
		arenaMember("追得上", 0, 300),             // 實得 3、帳面 0 → 合格
		arenaMember("已經齊了", 3, 300),            // 實得 3、帳面 3 → 不用
		arenaMember("到頂了", ArenaMaxRank, 9000), // 帳面 24 → 不用
	}
	got := ArenaCandidates(party)
	if len(got) != 1 || got[0] != 0 {
		t.Fatalf("合格的是 %v", got)
	}
}

// 沒有人夠格就只講一句。
func TestArenaNobodyReady(t *testing.T) {
	s := &Session{Party: []Member{arenaMember("甲", 3, 300)},
		Trans: &Translations{}, UIText: &UIText{}, Outdoor: &Outdoor{}}
	if !s.EnterArena(&Inn{}) {
		t.Fatal("還是要開起畫面講一句")
	}
	if s.ArenaAnswer(true) != nil {
		t.Fatal("沒有人夠格就不該打")
	}
}

// 一個人就直接問「確定要進去嗎」；答應之後回傳那一場，對手等級是實得位階加一。
func TestArenaSinglePartyFights(t *testing.T) {
	s := &Session{Party: []Member{arenaMember("甲", 0, 300)},
		Trans: &Translations{}, UIText: &UIText{}, Outdoor: &Outdoor{}}
	s.EnterArena(&Inn{})
	if len(s.Inside.Icons) != 2 || s.Inside.Icons[0] != IconSprite(IconYes) {
		t.Fatalf("確認那一問的圖示列是 %v", s.Inside.Icons)
	}
	fight := s.ArenaAnswer(true)
	if fight == nil || fight.Member != 0 || fight.Level != 4 {
		t.Fatalf("那一場是 %+v", fight)
	}
	if s.InArena() {
		t.Fatal("談完就不在競技場的對話裡了")
	}
}

// 答否只印一句，不打。
func TestArenaRefuse(t *testing.T) {
	s := &Session{Party: []Member{arenaMember("甲", 0, 300)},
		Trans: &Translations{}, UIText: &UIText{}, Outdoor: &Outdoor{}}
	s.EnterArena(&Inn{})
	if s.ArenaAnswer(false) != nil {
		t.Fatal("答否不該打")
	}
}

// 多人要先選誰上場，選到不夠格的要重選。
func TestArenaChoose(t *testing.T) {
	s := &Session{Party: []Member{
		arenaMember("齊了", 3, 300),
		arenaMember("要升", 0, 300),
	}, Trans: &Translations{}, UIText: &UIText{}, Outdoor: &Outdoor{}}
	s.EnterArena(&Inn{})
	if s.Inside.List == nil || len(s.Inside.List.Rows) != 2 {
		t.Fatalf("清單是 %+v", s.Inside.List)
	}
	// 選第 0 個（不夠格）不會往下走。
	s.ArenaChoose()
	if s.Inside.List == nil {
		t.Fatal("選到不夠格的應該留在清單上")
	}
	s.MoveArenaList(1)
	s.ArenaChoose()
	if s.Inside.List != nil {
		t.Fatal("選到夠格的就該換成確認那一問")
	}
	if p := s.Inside.Portrait; p.Archive != ArenaPortraitArchive || p.Index != ArenaPortraitTester {
		t.Fatalf("競技場的肖像應該是 COMBPICS 第 6 張，得到 %s[%d]", p.Archive, p.Index)
	}
	// 多人隊伍答應之後先換到賭場，答「不下注」才開打。
	if s.ArenaAnswer(true) != nil {
		t.Fatal("多人隊伍要先進賭場")
	}
	if p := s.Inside.Portrait; p.Archive != ArenaPortraitArchive || p.Index != ArenaPortraitBetter {
		t.Fatalf("賭場的肖像應該換成 COMBPICS 第 5 張（`0x2ADCB`），得到 %s[%d]", p.Archive, p.Index)
	}
	s.Continue()
	fight := s.BetAnswer(false)
	if fight == nil || fight.Member != 1 {
		t.Fatalf("上場的是 %+v", fight)
	}
}

// 下注：押他會贏是正值、押他會輸是負值；押到上場那一位或押第二次都會被擋。
func TestArenaBetting(t *testing.T) {
	s := &Session{Party: []Member{
		arenaMember("上場", 0, 300),
		arenaMember("押注", 3, 600),
	}, Trans: &Translations{}, UIText: &UIText{}, Outdoor: &Outdoor{}}
	s.Party[1].Gold = 50
	s.EnterArena(&Inn{})
	s.ArenaChoose() // 選第 0 名上場
	s.ArenaAnswer(true)
	s.Continue()
	s.askBet()

	// 押上場那一位會被擋。
	s.BetAnswer(true)
	s.Inside.List.Selected = 0
	s.BetChoose()
	if s.arena.stage != arenaBetAsk {
		t.Fatal("押到上場那一位應該回去重問")
	}

	// 押另一位：金額 25，押他會贏。
	s.askBet()
	s.BetAnswer(true)
	s.Inside.List.Selected = 1
	s.BetChoose()
	if !s.BettingAmount() {
		t.Fatal("應該在打金額")
	}
	s.TypeBet('2')
	s.TypeBet('5')
	s.TypeBet('9') // 只收兩位
	s.SubmitBet()
	s.BetSide(true)
	if got := s.ArenaBets()[1]; got != 25 {
		t.Fatalf("押的是 %d", got)
	}

	// 同一位押第二次會被擋。
	s.BetAnswer(true)
	s.Inside.List.Selected = 1
	s.BetChoose()
	if s.arena.stage != arenaBetAsk {
		t.Fatal("押第二次應該回去重問")
	}
}

// 押他會輸是負值；金額超過錢包會被擋。
func TestArenaBetSideAndPurse(t *testing.T) {
	s := &Session{Party: []Member{
		arenaMember("上場", 0, 300),
		arenaMember("押注", 3, 600),
	}, Trans: &Translations{}, UIText: &UIText{}, Outdoor: &Outdoor{}}
	s.Party[1].Gold = 10
	s.EnterArena(&Inn{})
	s.ArenaChoose()
	s.ArenaAnswer(true)
	s.Continue()
	s.askBet()
	s.BetAnswer(true)
	s.Inside.List.Selected = 1
	s.BetChoose()
	// 超過錢包。
	s.TypeBet('9')
	s.TypeBet('9')
	s.SubmitBet()
	if s.arena.stage != arenaBetAsk || s.ArenaBets()[1] != 0 {
		t.Fatal("錢不夠不該押得成")
	}
	// 押他會輸。
	s.askBet()
	s.BetAnswer(true)
	s.Inside.List.Selected = 1
	s.BetChoose()
	s.TypeBet('5')
	s.SubmitBet()
	s.BetSide(false)
	if got := s.ArenaBets()[1]; got != -5 {
		t.Fatalf("押他會輸應該是負值，得到 %d", got)
	}
}

// 打贏升一階，一次只加一級。
func TestArenaWinAddsOneRank(t *testing.T) {
	s := &Session{Party: []Member{arenaMember("甲", 0, 300)},
		Trans: &Translations{}, UIText: &UIText{}, Outdoor: &Outdoor{}}
	s.ArenaWin(0)
	if s.Party[0].Sheet.Rank != 1 {
		t.Fatalf("升成 %d", s.Party[0].Sheet.Rank)
	}
	s.ArenaWin(0)
	if s.Party[0].Sheet.Rank != 2 {
		t.Fatalf("再打一場升成 %d", s.Party[0].Sheet.Rank)
	}
}

// 戰後結算：打贏時押他會贏的人拿錢，打輸時整排先取負，押他會輸的人反而拿錢。
func TestSettleBets(t *testing.T) {
	newSession := func() *Session {
		s := &Session{Party: []Member{
			arenaMember("上場", 0, 300),
			arenaMember("押贏", 3, 600),
			arenaMember("押輸", 3, 600),
		}, Trans: &Translations{}, UIText: &UIText{}, Outdoor: &Outdoor{}}
		s.Party[1].Gold, s.Party[2].Gold = 100, 100
		return s
	}
	bets := []int{0, 20, -30}

	won := newSession()
	lines := won.SettleBets(bets, true)
	if won.Party[1].Gold != 120 || won.Party[2].Gold != 70 {
		t.Fatalf("打贏之後的錢包是 %d／%d", won.Party[1].Gold, won.Party[2].Gold)
	}
	if len(lines) != 2 {
		t.Fatalf("應該印兩句，得到 %d", len(lines))
	}

	lost := newSession()
	lost.SettleBets(bets, false)
	if lost.Party[1].Gold != 80 || lost.Party[2].Gold != 130 {
		t.Fatalf("打輸之後的錢包是 %d／%d", lost.Party[1].Gold, lost.Party[2].Gold)
	}
}
