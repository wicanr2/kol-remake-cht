package battle

import "testing"

// retreatState 造一隻累得很慘的敵人與一名隊員，兩者的距離由呼叫端擺。
// 撤退門檻比的是**疲勞比例** `50 × Fatigue ÷ Stamina`（RE-130），
// 所以疲勞 140／耐力 100 是 70。
func retreatState(px, py int, fatigue int) *State {
	return &State{Actors: []Actor{
		{Side: SideParty, X: px, Y: py, Head: 10, Chest: 10, Threshold: 100},
		{Side: SideMonster, X: 10, Y: 10, Head: 10, Chest: 10, Threshold: 100,
			RightArm: 10, LeftArm: 10, RightLeg: 10, LeftLeg: 10, Level: 2,
			FleeAt: 65, FleeRecover: 35, Stamina: 100, Fatigue: fatigue},
	}}
}

// 傷得夠重、旁邊沒有人貼著：原地喘息，並記下撤退狀態（`0x238C1`）。
func TestEnemyRestsWhenHurtAndAlone(t *testing.T) {
	s := retreatState(20, 20, 140)
	d := s.PlanEnemy(1, &fixedRand{values: []int{2}})
	if d.Command != CmdRest {
		t.Fatalf("該原地喘息，得到 %#x", d.Command)
	}
	if !s.Actors[1].Retreating {
		t.Fatal("撤退狀態該被記起來")
	}
	if s.Actors[1].Command != CmdRest {
		t.Fatalf("指令該寫回記錄，得到 %#x", s.Actors[1].Command)
	}
}

// 傷得夠重、有隊員貼著：往外逃（`0x23873` 跳到移動段）。
func TestEnemyFleesWhenHurtAndCornered(t *testing.T) {
	s := retreatState(10, 11, 140)
	d := s.PlanEnemy(1, &fixedRand{values: []int{2}})
	if d.Command != CmdMove {
		t.Fatalf("該往外逃，得到 %#x", d.Command)
	}
}

// 已經在撤退、上一手又挨了打：繼續逃，不再喘息（`0x238BE`）。
func TestRetreatingEnemyKeepsFleeingWhileHurt(t *testing.T) {
	s := retreatState(20, 20, 140)
	s.Actors[1].Retreating = true
	s.Actors[1].LastHurt = 12
	d := s.PlanEnemy(1, &fixedRand{values: []int{2}})
	if d.Command != CmdMove {
		t.Fatalf("該繼續逃，得到 %#x", d.Command)
	}
}

// 已經在撤退、但上一手沒挨打：停下來喘息（`0x238C1` 的第二個入口）。
func TestRetreatingEnemyRestsWhenNoLongerBeingHit(t *testing.T) {
	s := retreatState(20, 20, 140)
	s.Actors[1].Retreating = true
	s.Actors[1].LastHurt = 0
	d := s.PlanEnemy(1, &fixedRand{values: []int{2}})
	if d.Command != CmdRest {
		t.Fatalf("該喘息，得到 %#x", d.Command)
	}
}

// 傷勢降到 FleeRecover 以下就解除撤退，回到正常判斷（`0x2381D`）。
func TestRetreatEndsWhenWoundsRecover(t *testing.T) {
	s := retreatState(10, 11, 60) // 傷勢 30 < FleeRecover 35
	s.Actors[1].Retreating = true
	d := s.PlanEnemy(1, &fixedRand{values: []int{2}})
	if s.Actors[1].Retreating {
		t.Fatal("傷勢回穩該清掉撤退狀態")
	}
	if d.Command != CmdWeapon {
		t.Fatalf("該回到正常判斷去打貼著的隊員，得到 %#x", d.Command)
	}
}

// 不死系的門檻是 100／0，而傷勢最多只到「累積傷害等於昏迷門檻」那一刻的 50——
// 走不進撤退那一段。
func TestUndeadNeverRetreats(t *testing.T) {
	s := retreatState(10, 11, 100) // 傷勢 50
	s.Actors[1].FleeAt, s.Actors[1].FleeRecover = 100, 0
	d := s.PlanEnemy(1, &fixedRand{values: []int{2}})
	if d.Command != CmdWeapon {
		t.Fatalf("不死系該照常攻擊，得到 %#x", d.Command)
	}
	if s.Actors[1].Retreating {
		t.Fatal("不該進入撤退狀態")
	}
}

// 一手做完的收尾：挨傷計數搬進 LastHurt，它與招架次數都歸零（`0x20E64`）。
func TestActClearsPerTurnCounters(t *testing.T) {
	s := &State{Actors: []Actor{
		{Side: SideParty, X: 5, Y: 5, Head: 10, Chest: 10, Threshold: 100,
			Command: CmdRest, Hurt: 9, Parries: 3},
	}}
	s.Exec.Order = []int{0}
	drainAct(s, &fixedRand{values: []int{1}})
	a := s.Actors[0]
	if a.LastHurt != 9 {
		t.Fatalf("這一手挨的傷該搬進 LastHurt，得到 %d", a.LastHurt)
	}
	if a.Hurt != 0 || a.Parries != 0 {
		t.Fatalf("兩個計數該歸零，得到 %d／%d", a.Hurt, a.Parries)
	}
}
