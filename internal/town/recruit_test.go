package town

import "testing"

func recruitSession(t *testing.T, party int) *Session {
	t.Helper()
	s := playSession(t)
	s.TownCode = "MIN"
	names := []string{"Poontang", "Olear", "Moraine", "Fighting Bull", "Kenneth",
		"Jax", "Sharakan", "Warder", "Shyla", "Aventhalas", "One tooth"}
	for i, n := range names {
		s.Roster = append(s.Roster, Candidate{Name: n, Portrait: i,
			Place: "HALFWAY HO", Town: "MIN"})
	}
	// 別間店的人與別座城的人都不該出現在清單裡。
	s.Roster = append(s.Roster,
		Candidate{Name: "Mula", Place: "TROLLSBANE", Town: "BRT"},
		Candidate{Name: "Elsewhere", Place: "HALFWAY HO", Town: "BRT"})
	for len(s.Party) < party {
		s.Party = append(s.Party, Member{Name: names[len(s.Party)-1]})
	}
	s.Walk(-1, 0)
	s.Walk(1, 0)
	s.Accept()   // 進門
	s.Continue() // 讀完敘述，圖示列展開
	return s
}

// 候選人是這間建築裡不在隊伍中的角色，清單一次只放得下八列。
func TestRecruitCandidates(t *testing.T) {
	s := recruitSession(t, 1)
	s.StartRecruit()
	if !s.Recruiting() {
		t.Fatal("按下 COMPANION 應該進招募流程")
	}
	icons, _ := s.Icons()
	if len(icons) != 2 || icons[0] != IconSprite(IconYes) || icons[1] != IconSprite(IconNo) {
		t.Fatalf("問句的圖示列是 %v", icons)
	}
	s.Accept() // 回答之後進清單
	if s.recruit == nil || len(s.recruit.Candidates) != 11 {
		t.Fatalf("候選人有 %d 位，應為 11", len(s.recruit.Candidates))
	}
	if s.Inside.List == nil || len(s.Inside.List.Rows) != ListRows {
		t.Fatalf("清單一次應該只畫 %d 列", ListRows)
	}
	if s.Inside.List.Rows[0] != "Poontang" || s.Inside.List.Selected != 0 {
		t.Fatalf("清單第一列是 %q、反白在 %d", s.Inside.List.Rows[0], s.Inside.List.Selected)
	}
}

// 反白到底之後清單才捲；兩端都夾住，不繞回。
func TestRecruitListScrolls(t *testing.T) {
	s := recruitSession(t, 1)
	s.StartRecruit()
	s.Accept()
	for i := 0; i < 7; i++ {
		s.MoveList(1)
	}
	if s.recruit.Top != 0 || s.Inside.List.Selected != 7 {
		t.Fatalf("第八列之前不該捲：Top %d、反白 %d", s.recruit.Top, s.Inside.List.Selected)
	}
	s.MoveList(1)
	if s.recruit.Top != 1 || s.Inside.List.Selected != 7 {
		t.Fatalf("捲一列之後：Top %d、反白 %d", s.recruit.Top, s.Inside.List.Selected)
	}
	for i := 0; i < 10; i++ {
		s.MoveList(1)
	}
	if s.recruit.Cursor != 10 || s.recruit.Top != 3 {
		t.Fatalf("按到底之後游標 %d、Top %d，應為 10 與 3", s.recruit.Cursor, s.recruit.Top)
	}
	for i := 0; i < 20; i++ {
		s.MoveList(-1)
	}
	if s.recruit.Cursor != 0 || s.recruit.Top != 0 {
		t.Fatalf("按回頂端之後游標 %d、Top %d", s.recruit.Cursor, s.recruit.Top)
	}
}

// 選定之後新成員成為發言者，圖示列回到建築內的一般狀態。
func TestRecruitJoin(t *testing.T) {
	s := recruitSession(t, 1)
	s.StartRecruit()
	s.Accept()
	s.MoveList(1)
	s.Accept() // 白框在第一格＝加入
	if len(s.Party) != 2 || s.Party[1].Name != "Olear" {
		t.Fatalf("隊伍是 %v", s.Party)
	}
	if !s.Party[1].Speaker || s.Party[0].Speaker {
		t.Fatal("新成員應該成為發言者")
	}
	if s.Recruiting() || s.Inside.List != nil {
		t.Fatal("加入之後應該離開招募流程")
	}
	if len(s.Inside.Icons) < 2 {
		t.Fatalf("加入之後圖示列是 %v", s.Inside.Icons)
	}
}

// 第二格是取消：回到建築內，隊伍不變。
func TestRecruitCancel(t *testing.T) {
	s := recruitSession(t, 1)
	s.StartRecruit()
	s.Accept()
	s.Inside.Selected = 1
	s.Accept()
	if s.Recruiting() || len(s.Party) != 1 {
		t.Fatalf("取消之後還在招募或隊伍變成 %d 人", len(s.Party))
	}
}

// 隊伍滿六人時 COMPANION 不出現在圖示列裡。
func TestFullPartyHasNoCompanionIcon(t *testing.T) {
	icons := InteriorIcons([2]int{0x08, 0}, MaxParty)
	for _, id := range icons {
		if id == IconCompanion {
			t.Fatal("滿員仍出現 COMPANION")
		}
	}
}

// 原版名冊清單把第 8、9 名的槽號留在共用緩衝 +0x28／+0x29，戰鬥 AI 會讀到；
// 清單不到九名時那兩格保留上一次的值（RE-159）。
func TestRosterListTail(t *testing.T) {
	s := recruitSession(t, 1)
	for i := range s.Roster {
		s.Roster[i].Slot = i + 2
	}
	s.StartRecruit()
	s.Accept()
	// 清單從 Poontang（槽 2）起共 11 名，第 8、9 名（從 0 數）是槽 10、11。
	if s.RosterListTail != [2]int{10, 11} {
		t.Fatalf("殘值是 %v，應為 [10 11]", s.RosterListTail)
	}
	s.cancelRecruit()
	// 清單變短：隊伍多了十人的名字，候選只剩一名，殘值不動。
	for _, n := range []string{"Poontang", "Olear", "Moraine", "Fighting Bull", "Kenneth", "Jax", "Sharakan", "Warder", "Shyla", "Aventhalas"} {
		s.Party = append(s.Party, Member{Name: n})
	}
	s.openList()
	if len(s.recruit.Candidates) != 1 || s.RosterListTail != [2]int{10, 11} {
		t.Fatalf("候選 %d 名、殘值 %v；清單不夠長時應保留上一次的值", len(s.recruit.Candidates), s.RosterListTail)
	}
}
