package battle

import "testing"

// fixedRand 讓規劃順序的擲骰固定，方便驗證。
type fixedRand struct {
	values []int
	i      int
}

func (r *fixedRand) Roll(n int) int {
	if len(r.values) == 0 {
		return 1
	}
	v := r.values[r.i%len(r.values)]
	r.i++
	return v
}

// 規劃鍵值：(等級 − 1) × 10 + 洞察，再加擲骰；隊員在這一回合重排過時先除以
// 2 × word_30D00（RE-145）。
func TestPlanKey(t *testing.T) {
	a := Actor{Side: SideParty, Level: 3, Foresight: 40, Threshold: 100}
	if got := PlanKey(a, 0, 7); got != 7+(3-1)*10+40 {
		t.Fatalf("第一回合鍵值 %d", got)
	}
	if got := PlanKey(a, 1, 7); got != 7+((3-1)*10+40)/2 {
		t.Fatalf("重排過一次的鍵值 %d", got)
	}
	// 疲勞會再折損：比例 = 50 × 疲勞 ÷ 耐力，基準扣掉基準 × 比例 ÷ 200。
	// 累積傷害在這一條式子裡不出現（RE-130）。
	tired := a
	tired.Stamina, tired.Fatigue = 200, 100
	base := (3-1)*10 + 40
	if got := PlanKey(tired, 0, 7); got != 7+base-base*25/200 {
		t.Fatalf("疲勞之後鍵值 %d", got)
	}
	hurt := a
	hurt.Damage = 50
	if got := PlanKey(hurt, 0, 7); got != 7+base {
		t.Fatalf("累積傷害不該動規劃鍵值，得到 %d", got)
	}
	// 敵人不吃重排那一條。
	m := Actor{Side: SideMonster, Level: 3, Foresight: 40, Threshold: 100}
	if got := PlanKey(m, 2, 7); got != 7+(3-1)*10+40 {
		t.Fatalf("敵人重排兩次的鍵值 %d", got)
	}
}

// 鍵值小的先動；有效範圍內的空隊員槽排最後，仍會消耗檢視門檻。
func TestBeginPlanningOrder(t *testing.T) {
	s := NewQuest(testSite(), []Member{{Name: "Poontang", Size: 100, Health: 100}},
		testMonster(), NewField())
	s.BeginPlanning(&fixedRand{values: []int{50, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}})
	if s.Round.Number != 1 {
		t.Fatalf("回合數 %d", s.Round.Number)
	}
	if len(s.Round.Order) != PartySlots+testSite().Count {
		t.Fatalf("順序長度 %d", len(s.Round.Order))
	}
	emptySeen := false
	for _, i := range s.Round.Order {
		if s.Actors[i].Side == SideEmpty {
			emptySeen = true
		} else if emptySeen {
			t.Fatal("空槽沒有排在有效人物之後")
		}
	}
	// 走完一輪之後 EndTurn 要回報結束。
	for range s.Round.Order {
		s.EndTurn()
	}
	if !s.EndTurn() {
		t.Fatal("走完一輪之後 EndTurn 應該回報結束")
	}
}

// 部位血量照 15、15、20、20、30、30 % 分配。
func TestParts(t *testing.T) {
	p := Parts(100)
	if p != [6]int{15, 15, 20, 20, 30, 30} {
		t.Fatalf("部位血量 %v", p)
	}
	a := Actor{Threshold: 100}
	a.SetParts()
	if a.Head != 15 || a.RightLeg != 30 || a.Down() {
		t.Fatalf("填好的部位不對：%+v", a)
	}
	a.Head = 0
	if !a.Down() {
		t.Fatal("頭被打廢應該算退出戰鬥")
	}
}

// 戰鬥等級：八個攻防原始值加總除以 1000 再加一。
func TestCombatLevel(t *testing.T) {
	if got := CombatLevel([]int{500, 500, 400, 400, 300, 300, 200, 200}); got != 3 {
		t.Fatalf("戰鬥等級 %d", got)
	}
	if got := CombatLevel(nil); got != 1 {
		t.Fatalf("沒有技能時 %d", got)
	}
}

// RE-248：原版掃描槽位與最後一次戰場重畫各自保留。
func TestPlanningCursorRetainsScannedTailAndPartyView(t *testing.T) {
	for _, side := range []Side{SideParty, SideEmpty} {
		for _, withEnemy := range []bool{false, true} {
			s := &State{Actors: make([]Actor, 18)}
			for i := range s.Actors {
				s.Actors[i].Side = SideEmpty
			}
			s.Actors[4] = Actor{Side: SideParty, Head: 10, Chest: 10, X: 30, Y: 31}
			s.Actors[6] = Actor{Side: SideMonster, Head: 10, Chest: 10, X: 34, Y: 35}
			s.Actors[1] = Actor{Side: side, Head: 0, Chest: 10, X: 20, Y: 21}
			s.Round.Order = []int{4, 1}
			if withEnemy {
				s.Round.Order = []int{4, 6, 1}
			}
			rng := &fixedRand{values: []int{1}}
			s.Round.inspectionRand = rng
			s.AdvanceToParty()
			if s.Cursor != 4 {
				t.Fatalf("隊員索引 %d", s.Cursor)
			}
			done := s.EndTurn()
			if !done {
				s.AdvanceToParty()
			}
			if s.Cursor != 1 {
				t.Errorf("末端槽位 %d，應保留 1；side=%d enemy=%v", s.Cursor, side, withEnemy)
			}
			x, y := s.Camera()
			if x != 30 || y != 31 {
				t.Errorf("鏡頭 %d,%d，應保留 30,31；side=%d enemy=%v", x, y, side, withEnemy)
			}
			if rng.i != len(s.Round.Order) {
				t.Fatalf("檢視擲骰 %d，預期 %d", rng.i, len(s.Round.Order))
			}
			if s.Current() != -1 {
				t.Fatalf("規劃尚未走完 %d", s.Current())
			}
		}
	}
}
