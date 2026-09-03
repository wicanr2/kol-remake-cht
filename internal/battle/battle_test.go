package battle

import "testing"

func testSite() *Site {
	return &Site{Index: 5, X: 48, Y: 21, Monster: 19, MonsterName: "BINDERAK",
		Level: 3, Count: 3, QuestByte: 0x1E3, QuestDiv: 4,
		MonsterX: []int{25, 27, 30}, MonsterY: []int{41, 37, 42},
		PartyX: []int{14, 12}, PartyY: []int{41, 41},
		Narrative: []string{"From the rim of one of"}}
}

// 任務旗標：一個 byte 塞四個 2-bit 欄位，除數決定看哪一個。
func TestSiteTaken(t *testing.T) {
	s := testSite()
	if !s.Taken([]int{0, 4, 0, 0, 0, 0}) {
		t.Fatal("Poontang 的 [0x1E3] = 4 應該帶著第 5 個任務")
	}
	if s.Taken([]int{0, 0, 0, 0, 0, 0}) {
		t.Fatal("沒有旗標卻算成帶著任務")
	}
	if s.Taken([]int{4, 0, 0, 0, 0, 0}) {
		t.Fatal("旗標在錯的 byte 上卻算數")
	}
	// 除數 4 看的是第 1 對位元：值 8 是 10b，低兩位不是 1。
	if s.Taken([]int{0, 8, 0, 0, 0, 0}) {
		t.Fatal("2-bit 欄位是 2 卻算成 1")
	}
}

// 站位直接照記錄擺：前六個位置是隊伍，其餘是敵人。
func TestNewQuestPlacement(t *testing.T) {
	party := []Member{{Name: "Poontang"}, {Name: "Theodoric"}}
	s := NewQuest(testSite(), party, NewField())
	if s.Stage != StageIntro {
		t.Fatalf("開場停在 %v", s.Stage)
	}
	if s.Actors[0].X != 14 || s.Actors[0].Y != 41 || s.Actors[0].Side != SideParty {
		t.Fatalf("第一名隊員是 %+v", s.Actors[0])
	}
	if s.Actors[1].Name != "Theodoric" || s.Actors[1].Kind != PartyKind {
		t.Fatalf("第二名隊員是 %+v", s.Actors[1])
	}
	if s.Actors[2].Side != SideEmpty {
		t.Fatalf("第三個位置沒空著：%+v", s.Actors[2])
	}
	if s.Actors[PartySlots].Side != SideMonster || s.Actors[PartySlots].X != 25 {
		t.Fatalf("第一隻敵人是 %+v", s.Actors[PartySlots])
	}
	if s.Actors[PartySlots+2].Y != 42 || s.Actors[PartySlots+3].Side != SideEmpty {
		t.Fatalf("敵人數量不對：%+v", s.Actors[PartySlots+2:PartySlots+4])
	}
}

// 開場問「要不要參一腳」，答否就直接結束；答應之後走到滿版怪物圖再到戰場。
func TestStages(t *testing.T) {
	s := NewQuest(testSite(), []Member{{Name: "Poontang"}}, NewField())
	if !s.Answer(false) || s.Stage != StageDone {
		t.Fatalf("答否之後停在 %v", s.Stage)
	}
	s = NewQuest(testSite(), []Member{{Name: "Poontang"}}, NewField())
	s.Answer(true)
	if s.Stage != StageEncounter {
		t.Fatalf("答應之後停在 %v", s.Stage)
	}
	s.Continue()
	if s.Stage != StageField {
		t.Fatalf("讀完之後停在 %v", s.Stage)
	}
	s.Answer(false)
	if s.Stage != StageFlee {
		t.Fatalf("不開下一回合之後停在 %v", s.Stage)
	}
	s.Answer(false)
	if s.Stage != StageField {
		t.Fatalf("不逃之後停在 %v", s.Stage)
	}
	s.Answer(false)
	if !s.Answer(true) || s.Stage != StageDone {
		t.Fatalf("逃走之後停在 %v", s.Stage)
	}
}

// cornerRand 把灑點全部擲到最後一格，底色則交替 1 與 2，
// 這樣底色與灑點兩件事可以分開看。
type cornerRand struct{ i int }

func (r *cornerRand) Roll(n int) int {
	if n == FieldSide*FieldSide {
		return n
	}
	r.i++
	if n == RoughOdds {
		return 2 - r.i%2
	}
	return 1
}

// 一般遭遇的地形：逐格擲 1…4，中 1 填 TileRough，再灑 TileScatter。
func TestGenerate(t *testing.T) {
	f := Generate(&cornerRand{})
	if f.At(0, 0) != TileRough || f.At(0, 1) != TileGround {
		t.Fatalf("底色不對：%d %d", f.At(0, 0), f.At(0, 1))
	}
	if f.At(FieldSide-1, FieldSide-1) != TileScatter {
		t.Fatalf("灑點沒落在最後一格：%d", f.At(FieldSide-1, FieldSide-1))
	}
	if len(f.Tiles) != FieldSide*FieldSide {
		t.Fatalf("戰場大小是 %d", len(f.Tiles))
	}
}

// 視窗是以誰為中心的 9×9。
func TestView(t *testing.T) {
	f := NewField()
	f.Tiles[14*FieldSide+41] = TileScatter
	v := f.View(14, 41)
	if len(v) != ViewRows || len(v[0]) != ViewCols {
		t.Fatalf("視窗是 %d×%d", len(v), len(v[0]))
	}
	if v[ViewCentre.Row][ViewCentre.Col] != TileScatter {
		t.Fatalf("中心那一格是 %d", v[ViewCentre.Row][ViewCentre.Col])
	}
}
