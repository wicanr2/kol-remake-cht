package battle

import "testing"

// fixedRand 讓規劃順序的擲骰固定，方便驗證。
type fixedRand struct{ values []int; i int }

func (r *fixedRand) Roll(n int) int {
	if len(r.values) == 0 {
		return 1
	}
	v := r.values[r.i%len(r.values)]
	r.i++
	return v
}

// 規劃鍵值：(等級 − 1) × 10 + 洞察，再加擲骰；隊員第二回合起先除以 2 × 回合數。
func TestPlanKey(t *testing.T) {
	a := Actor{Side: SideParty, Level: 3, Foresight: 40, Threshold: 100}
	if got := PlanKey(a, 0, 7); got != 7+(3-1)*10+40 {
		t.Fatalf("第一回合鍵值 %d", got)
	}
	if got := PlanKey(a, 1, 7); got != 7+((3-1)*10+40)/2 {
		t.Fatalf("第二回合鍵值 %d", got)
	}
	// 受傷會再折損：傷勢 = 50 × 傷害 ÷ 門檻，基準扣掉基準 × 傷勢 ÷ 200。
	hurt := a
	hurt.Damage = 50
	base := (3-1)*10 + 40
	if got := PlanKey(hurt, 0, 7); got != 7+base-base*25/200 {
		t.Fatalf("受傷之後鍵值 %d", got)
	}
	// 敵人不吃回合數那一條。
	m := Actor{Side: SideMonster, Level: 3, Foresight: 40, Threshold: 100}
	if got := PlanKey(m, 2, 7); got != 7+(3-1)*10+40 {
		t.Fatalf("敵人第三回合鍵值 %d", got)
	}
}

// 鍵值小的先動；空位不進順序。
func TestBeginPlanningOrder(t *testing.T) {
	s := NewQuest(testSite(), []Member{{Name: "Poontang", Size: 100, Health: 100}},
		testMonster(), NewField())
	s.BeginPlanning(&fixedRand{values: []int{50, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}})
	if s.Round.Number != 1 {
		t.Fatalf("回合數 %d", s.Round.Number)
	}
	if len(s.Round.Order) != 1+testSite().Count {
		t.Fatalf("順序長度 %d", len(s.Round.Order))
	}
	prev := -1 << 30
	for _, i := range s.Round.Order {
		if s.Actors[i].Side == SideEmpty {
			t.Fatalf("空位進了順序：%d", i)
		}
		_ = prev
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
