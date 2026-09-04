package battle

import "testing"

// 輪盤：權重全在第一項就選第一項，全在第二項就選第二項。
func TestWeightedPickFollowsWeights(t *testing.T) {
	// Roll 回 1 → r ＝ 0，第一個非零的項就會把它扣成負的。
	one := func() Rand { return &fixedRand{values: []int{1}} }
	if got := WeightedPick([]int{1, 0, 0}, []int{1, 1, 1}, 1, one()); got != 0 {
		t.Fatalf("權重全在第 0 項，卻選了 %d", got)
	}
	if got := WeightedPick([]int{0, 1, 0}, []int{1, 1, 1}, 1, one()); got != 1 {
		t.Fatalf("權重全在第 1 項，卻選了 %d", got)
	}
}

// 等級放大情境權重的差距：`w^等級 ÷ 2^(等級−1)`。
// 情境權重 2 比 1，等級 1 時是 2:1，等級 3 時是 8:0——高等級的怪物幾乎不亂選。
func TestWeightedPickLevelSharpensTheOdds(t *testing.T) {
	base, ctx := []int{1, 1}, []int{2, 1}
	// 等級 1：權重 2:1 → 正規化成 170:85。r ＝ 200 落在第二項。
	low := &fixedRand{values: []int{201}}
	if got := WeightedPick(base, ctx, 1, low); got != 1 {
		t.Fatalf("等級 1 時 r=200 應該落在第 1 項，得到 %d", got)
	}
	// 等級 3：第二項的權重被除成 0，同樣的 r 只能落在第 0 項。
	high := &fixedRand{values: []int{201}}
	if got := WeightedPick(base, ctx, 3, high); got != 0 {
		t.Fatalf("等級 3 時第 1 項的權重應該被除成 0，卻選了 %d", got)
	}
}

// 權重全零不會除以零，也不會挑出界。
func TestWeightedPickAllZero(t *testing.T) {
	if got := WeightedPick([]int{0, 0}, []int{0, 0}, 2, &fixedRand{values: []int{1}}); got != 0 {
		t.Fatalf("權重全零應該回 0，得到 %d", got)
	}
}

// 攻擊式：武器位元對得上的那一式情境權重加倍。
func TestEnemyStyleDoublesUsableForms(t *testing.T) {
	// StyleMask[0] 是 16；武器位元帶 16 的話第 0 式權重加倍。
	with := EnemyStyle(1, 0, 16, &fixedRand{values: []int{1}})
	without := EnemyStyle(1, 0, 0, &fixedRand{values: []int{1}})
	if with < 0 || with >= len(AttackStyleBase) {
		t.Fatalf("攻擊式 %d 出界", with)
	}
	if without < 0 || without >= len(AttackStyleBase) {
		t.Fatalf("攻擊式 %d 出界", without)
	}
}

// 部位：回傳一定落在七項之內。
func TestEnemyLocationInRange(t *testing.T) {
	for level := 0; level < 8; level++ {
		for style := 0; style < 10; style++ {
			got := EnemyLocation(level, style, &fixedRand{values: []int{level*7 + style + 1}})
			if got < 0 || got >= len(LocationWeights) {
				t.Fatalf("等級 %d 攻擊式 %d 抽到 %d，出界", level, style, got)
			}
		}
	}
}

// 移動三式：兩腿都還在而且不是最外圈時，衝刺那一項歸零。
func TestEnemyMoveFormZeroesSprintWhenLegsAreFine(t *testing.T) {
	// 權重變成 12:24:0，r 掃過整個範圍都不該選到 2。
	for r := 1; r <= 256; r++ {
		got := EnemyMoveForm(1, 0, true, 0, &fixedRand{values: []int{r}})
		if got == 2 {
			t.Fatalf("腿還在的時候不該選衝刺（r=%d）", r)
		}
	}
	// 腿廢了就選得到衝刺——權重 12:24:72，大部分的 r 會落在第 2 項。
	sprinted := false
	for r := 1; r <= 256; r++ {
		if EnemyMoveForm(1, 0, false, 0, &fixedRand{values: []int{r}}) == 2 {
			sprinted = true
			break
		}
	}
	if !sprinted {
		t.Fatal("腿廢了應該選得到衝刺")
	}
}

// 規劃段輪到敵人時要真的填上一手，不能留空——留空的話執行段會變成
// 「指令 0 → 什麼都沒做」，看起來像原版選了不動。
func TestAdvanceToPartyPlansTheEnemy(t *testing.T) {
	s := &State{Rand: &fixedRand{values: []int{2, 3, 5, 7}}, Actors: []Actor{
		{Side: SideParty, X: 4, Y: 4, Head: 10, Chest: 10, Threshold: 100,
			RightArm: 10, LeftArm: 10, RightLeg: 10, LeftLeg: 10, Level: 2},
		{Side: SideMonster, X: 4, Y: 5, Head: 10, Chest: 10, Threshold: 100,
			RightArm: 10, LeftArm: 10, RightLeg: 10, LeftLeg: 10, Level: 2},
	}}
	// 敵人先動：順序裡把它排在前面。
	s.Round.Order = []int{1, 0}
	s.Round.Turn = 0
	s.AdvanceToParty()

	enemy := s.Actors[1]
	if enemy.Command != CmdWeapon && enemy.Command != CmdMove {
		t.Fatalf("敵人的指令是 %#x，應該是武器攻擊或移動", enemy.Command)
	}
	// 停在隊員身上等玩家選。
	if s.Round.Turn != 1 || s.Actors[s.Round.Order[s.Round.Turn]].Side != SideParty {
		t.Fatalf("應該停在隊員那一手，Turn ＝ %d", s.Round.Turn)
	}
}

// 貼在旁邊的敵人會挑到那名隊員，而且方向指過去。
func TestPlanEnemyTargetsTheAdjacentMember(t *testing.T) {
	s := &State{Actors: []Actor{
		{Side: SideParty, X: 4, Y: 5, Head: 10, Chest: 10, Threshold: 100},
		{Side: SideMonster, X: 4, Y: 4, Head: 10, Chest: 10, Threshold: 100,
			RightArm: 10, LeftArm: 10, RightLeg: 10, LeftLeg: 10, Level: 2},
	}}
	// Roll 一律回 2：找目標的起點固定，1/3 那一擲回 2 所以不會放棄攻擊。
	d := s.PlanEnemy(1, &fixedRand{values: []int{2}})
	if d.Command != CmdWeapon {
		t.Fatalf("貼著隊員應該攻擊，得到 %#x", d.Command)
	}
	if d.Target != 0 {
		t.Fatalf("目標應該是第 0 名隊員，得到 %d", d.Target)
	}
	if d.Ring != 0 {
		t.Fatalf("相鄰應該在環 0 找到，得到 %d", d.Ring)
	}
	// (4,4) 往 (4,5)：Row 差 0、Col 差 +1，也就是環 0 的第 2 格。
	if want := 2; d.Facing != want {
		t.Fatalf("方向是 %d，期望 %d", d.Facing, want)
	}
}

// 場上沒有隊員可打就移動，不會卡住。
func TestPlanEnemyMovesWhenNobodyIsAround(t *testing.T) {
	s := &State{Actors: []Actor{
		{Side: SideMonster, X: 4, Y: 4, Head: 10, Chest: 10, Threshold: 100,
			RightArm: 10, LeftArm: 10, RightLeg: 10, LeftLeg: 10, Level: 2},
	}}
	d := s.PlanEnemy(0, &fixedRand{values: []int{2}})
	if d.Command != CmdMove {
		t.Fatalf("沒有目標應該移動，得到 %#x", d.Command)
	}
	if d.Style < IconMoveWalk || d.Style > IconMoveWalk+2 {
		t.Fatalf("移動方式 %#x 出界", d.Style)
	}
}

// 離得遠的敵人會一回合走一格靠過來，貼上之後改成攻擊。
// 這一支把規劃與移動串起來跑，驗的是「AI 決定的方向真的把它帶到隊員旁邊」。
func TestEnemyClosesInThenAttacks(t *testing.T) {
	s := &State{Field: &Field{Tiles: make([]int, FieldSide*FieldSide)},
		Actors: []Actor{
			{Side: SideParty, X: 10, Y: 10, Head: 10, Chest: 10, Threshold: 100},
			{Side: SideMonster, X: 15, Y: 10, Head: 10, Chest: 10, Threshold: 100,
				RightArm: 10, LeftArm: 10, RightLeg: 10, LeftLeg: 10, Level: 2},
		}}
	// Roll 一律回 2：找目標的起點固定，而且 1/3 那一擲不會中，敵人每回合都往前走。
	rng := &fixedRand{values: []int{2}}

	attacked := false
	for round := 0; round < 8; round++ {
		d := s.PlanEnemy(1, rng)
		if d.Command == CmdWeapon {
			attacked = true
			break
		}
		if d.Command != CmdMove {
			t.Fatalf("第 %d 回合的指令是 %#x", round, d.Command)
		}
		s.move(1)
	}
	if !attacked {
		t.Fatalf("走了八個回合還沒貼上來，敵人停在 (%d,%d)",
			s.Actors[1].X, s.Actors[1].Y)
	}
	// 貼上來之後距離應該是 1。
	dx, dy := s.Actors[1].X-s.Actors[0].X, s.Actors[1].Y-s.Actors[0].Y
	if dx*dx > 1 || dy*dy > 1 {
		t.Fatalf("攻擊時距離還有 (%d,%d)", dx, dy)
	}
}
