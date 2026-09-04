package town

import "testing"

func arenaMember(name string, rank, offence int) Member {
	return Member{Name: name, Sheet: &Sheet{Rank: rank,
		Weapons: []WeaponRow{{OffenceRaw: offence, DefenceRaw: offence}}}}
}

// 帳面位階與實得一樣就不用升；到 24 也不用再打（RE-108）。
func TestArenaCandidates(t *testing.T) {
	party := []Member{
		arenaMember("追得上", 0, 300),        // 實得 3、帳面 0 → 合格
		arenaMember("已經齊了", 3, 300),      // 實得 3、帳面 3 → 不用
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
	fight := s.ArenaAnswer(true)
	if fight == nil || fight.Member != 1 {
		t.Fatalf("上場的是 %+v", fight)
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
