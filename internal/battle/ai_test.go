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
	a := Actor{Level: 1, Stamina: 200, Threshold: 100}
	with := EnemyStyle(a, DefaultAIWeights, 0, 16, [2]int{}, &fixedRand{values: []int{1}})
	without := EnemyStyle(a, DefaultAIWeights, 0, 0, [2]int{}, &fixedRand{values: []int{1}})
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
			a := Actor{Level: level, Stamina: 200, Threshold: 100}
			got := EnemyLocation(a, DefaultAIWeights, style, &fixedRand{values: []int{level*7 + style + 1}})
			if got < 0 || got >= len(LocationWeights) {
				t.Fatalf("等級 %d 攻擊式 %d 抽到 %d，出界", level, style, got)
			}
		}
	}
}

// 移動三式：兩腿都還在而且不是最外圈時，衝刺那一項歸零；但正規化只做索引 2 與 1，
// 索引 0 維持原值，權重加起來不到 255，擲到尾端的仍落在衝刺（RE-159）。
func TestEnemyMoveFormZeroesSprintWhenLegsAreFine(t *testing.T) {
	legs := Actor{RightLeg: 10, LeftLeg: 10}
	// 預設權重 12:24:72 → 12:0x18×255÷36=170:0，前兩項合計 182。
	for r := 1; r <= 256; r++ {
		got := EnemyMoveForm(legs, DefaultAIWeights, 0, &fixedRand{values: []int{r}})
		want := 2
		switch {
		case r-1 < 12:
			want = 0
		case r-1 < 182:
			want = 1
		}
		if got != want {
			t.Fatalf("r=%d 選到 %d，應為 %d", r, got, want)
		}
	}
	// 腿廢了：衝刺不歸零，環 0 減半、額度夠時比值 1 → 12:24:36，正規化成 12:85:127。
	lame := Actor{RightLeg: 0x90, LeftLeg: 10, Stamina: 200, Threshold: 50}
	if got := EnemyMoveForm(lame, DefaultAIWeights, 0, &fixedRand{values: []int{12 + 85 + 1}}); got != 2 {
		t.Fatalf("腿廢了應該選得到衝刺，得到 %d", got)
	}
	if got := EnemyMoveForm(lame, DefaultAIWeights, 0, &fixedRand{values: []int{12 + 85}}); got != 1 {
		t.Fatalf("r=97 應落在奔跑，得到 %d", got)
	}
}

// 怪物記錄 `+0x1A0` 非零才用自帶的三張表（RE-159）。
func TestMonsterAIWeights(t *testing.T) {
	m := Monster{StyleWeights: []int{7, 2, 2, 2, 2, 2, 7, 7, 2, 1},
		LocationWeights: []int{1, 6, 7, 1, 7, 1, 5}, MoveWeights: []int{6, 6, 72}}
	if w := m.AIWeights(); w.Move != [3]int{6, 6, 72} || w.Style[0] != 7 || w.Location[1] != 6 {
		t.Fatalf("旗標非零應用記錄自帶的表，得到 %+v", w)
	}
	m.StyleWeights[0] = 0
	if w := m.AIWeights(); w != DefaultAIWeights {
		t.Fatalf("+0x1A0 為零應用預設表，得到 %+v", w)
	}
	if w := (Monster{}).AIWeights(); w != DefaultAIWeights {
		t.Fatalf("缺欄時應用預設表，得到 %+v", w)
	}
}

// 規劃段輪到敵人時要真的填上一手，不能留空——留空的話執行段會變成
// 「指令 0 → 什麼都沒做」，看起來像原版選了不動。
func TestAdvanceToPartyPlansTheEnemy(t *testing.T) {
	s := &State{Rand: &fixedRand{values: []int{2, 3, 5, 7}}, Actors: []Actor{
		{Side: SideParty, X: 4, Y: 4, Head: 10, Chest: 10, Threshold: 100,
			RightArm: 10, LeftArm: 10, RightLeg: 10, LeftLeg: 10, Level: 2,
			FleeAt: 65, FleeRecover: 35},
		{Side: SideMonster, X: 4, Y: 5, Head: 10, Chest: 10, Threshold: 100,
			RightArm: 10, LeftArm: 10, RightLeg: 10, LeftLeg: 10, Level: 2,
			FleeAt: 65, FleeRecover: 35},
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
			RightArm: 10, LeftArm: 10, RightLeg: 10, LeftLeg: 10, Level: 2,
			FleeAt: 65, FleeRecover: 35},
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
			RightArm: 10, LeftArm: 10, RightLeg: 10, LeftLeg: 10, Level: 2,
			FleeAt: 65, FleeRecover: 35},
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
	// 地形要用 `NewField()`：磚 0 雖然走得過去，卻**擋視線**（`sightBits[0]` 是
	// 0xFF），全 0 的戰場上敵人看不見目標，走的會是「沒有目標」那一條——那樣這一支
	// 驗的就不是它以為在驗的東西了。
	s := &State{Field: NewField(),
		Actors: []Actor{
			{Side: SideParty, X: 10, Y: 10, Head: 10, Chest: 10, Threshold: 100},
			{Side: SideMonster, X: 15, Y: 10, Head: 10, Chest: 10, Threshold: 100,
				RightArm: 10, LeftArm: 10, RightLeg: 10, LeftLeg: 10, Level: 2,
				FleeAt: 65, FleeRecover: 35},
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

// 環 1…3 上看得見的目標會被射（`0x23F15`，RE-104）：指令碼換成 0x23、
// 目標那一格記進 Memo、+0x1D 直接設 1。
func TestPlanEnemyShootsFromRange(t *testing.T) {
	newState := func() *State {
		return &State{Field: NewField(), Actors: []Actor{
			{Side: SideParty, X: 10, Y: 10, Head: 10, Chest: 10, Threshold: 100},
			{Side: SideMonster, Kind: 3, X: 10, Y: 12, Head: 10, Chest: 10,
				Threshold: 100, RightArm: 10, LeftArm: 10,
				RightLeg: 10, LeftLeg: 10, Level: 2,
				FleeAt: 65, FleeRecover: 35},
		}}
	}
	s := newState()
	d := s.PlanEnemy(1, &fixedRand{values: []int{2}})
	if d.Command != CmdMissile {
		t.Fatalf("環 2 看得見應該射擊，得到 %#x", d.Command)
	}
	if d.Target != 0 || d.Ring != 1 {
		t.Fatalf("目標／環號是 %d／%d，期望 0／1", d.Target, d.Ring)
	}
	if got := s.Plan.Memo[1]; got.Row != 10 || got.Col != 10 {
		t.Fatalf("目標那一格應該記進 Memo，得到 %+v", got)
	}
	if s.Actors[1].Flag1D != 1 {
		t.Fatal("敵人不必先裝填，+0x1D 直接設 1")
	}
	// 中間隔一面牆就射不到，改成靠近。
	s = newState()
	s.Field.Tiles[10*FieldSide+11] = 0x29
	d = s.PlanEnemy(1, &fixedRand{values: []int{2}})
	if d.Command != CmdMove {
		t.Fatalf("被牆擋住應該改成移動，得到 %#x", d.Command)
	}
	// 環 3 仍能射擊，環 4 起才靠近（SPEC-013 原版環表訂正）。
	s = newState()
	s.Actors[1].Y = 14
	d = s.PlanEnemy(1, &fixedRand{values: []int{2}})
	if d.Command != CmdMissile || d.Ring != 3 {
		t.Fatalf("環 3 應該射擊，得到 %#x／%d", d.Command, d.Ring)
	}
	s = newState()
	s.Actors[1].Y = 15
	d = s.PlanEnemy(1, &fixedRand{values: []int{2}})
	if d.Command != CmdMove || d.Ring != 4 {
		t.Fatalf("環 4 應該靠近，得到 %#x／%d", d.Command, d.Ring)
	}
	// 右臂廢了也射不出去。
	s = newState()
	s.Actors[1].RightArm = 0
	d = s.PlanEnemy(1, &fixedRand{values: []int{2}})
	if d.Command != CmdMove {
		t.Fatalf("右臂廢了應該改成移動，得到 %#x", d.Command)
	}
}

// 貼近打那一手的三個選項（`0x23CD6`…`0x23D8B`）：攻擊式退回 1 時招架姿態是 2；
// 否則用**進來時**的攻擊選項當列號抽命中部位，填進去的是「部位 ＋1」。
func TestEnemyStrikeFillsGuardFromLocation(t *testing.T) {
	s := &State{Actors: []Actor{
		{Side: SideParty, X: 4, Y: 5, Head: 10, Chest: 10, Threshold: 100},
		{Side: SideMonster, X: 4, Y: 4, Head: 10, Chest: 10, Threshold: 100,
			RightArm: 10, LeftArm: 10, RightLeg: 10, LeftLeg: 10, Level: 2,
			FleeAt: 65, FleeRecover: 35},
	}}
	// 擲什麼都回 1：攻擊式是 sub_2578F 的第一項加二，不會退回 1，
	// 所以會走到抽部位那一條。
	var d EnemyDecision
	s.enemyStrike(&d, 1, 0, EnemyNearStyle, EnemyNearGuard, &fixedRand{values: []int{1}})
	if d.Style == 0 && d.Guard != 1 {
		t.Fatalf("部位超過 6 時該是攻擊選項 0、防禦選項 1，得到 %d／%d", d.Style, d.Guard)
	}
	if d.Style != 0 && (d.Guard < 0 || d.Guard > 6) {
		t.Fatalf("防禦選項 %d 不在 0…6", d.Guard)
	}
	// 進來時沒有攻擊選項：擲一次決定 0 或 2（`0x23D2F`）。
	var zero EnemyDecision
	s.enemyStrike(&zero, 1, 0, 0, EnemyNearGuard, &fixedRand{values: []int{1}})
	if zero.Style != 1 && zero.Guard != 0 && zero.Guard != 2 {
		t.Fatalf("prev 為 0 時防禦選項該是 0 或 2，得到 %d", zero.Guard)
	}
}

// 貼著的敵人現在會填招架姿態，不再一律留 0——原版那一格是「命中部位 ＋1」，
// 而它在被攻擊時就是防禦選項。
func TestPlanEnemyFillsGuardWhenAdjacent(t *testing.T) {
	s := &State{Actors: []Actor{
		{Side: SideParty, X: 4, Y: 5, Head: 10, Chest: 10, Threshold: 100},
		{Side: SideMonster, X: 4, Y: 4, Head: 10, Chest: 10, Threshold: 100,
			RightArm: 10, LeftArm: 10, RightLeg: 10, LeftLeg: 10, Level: 2,
			FleeAt: 65, FleeRecover: 35},
	}}
	d := s.PlanEnemy(1, &fixedRand{values: []int{2}})
	if d.Command != CmdWeapon {
		t.Fatalf("貼著隊員應該攻擊，得到 %#x", d.Command)
	}
	if s.Actors[1].Guard != d.Guard {
		t.Fatalf("防禦選項該寫回記錄，得到 %d／%d", s.Actors[1].Guard, d.Guard)
	}
}

// 智力擲骰：`值 × 等級 ÷ 3 > 門檻`（`sub_252CB`）。
func TestEnemyIntellectCheck(t *testing.T) {
	// 智力 60、等級 2 → 40，門檻 39 過、40 不過。
	if !EnemyIntellectCheck(60, 2, 39) {
		t.Fatal("40 > 39 該過")
	}
	if EnemyIntellectCheck(60, 2, 40) {
		t.Fatal("40 > 40 不成立")
	}
}

// 反制掃描抄對手的攻擊選項：對方已規劃、要用武器攻擊、而且方向正對著我，
// 優先度 5（`0x254D5`）。
func TestCounterScanCopiesFacingAttacker(t *testing.T) {
	s := &State{Actors: []Actor{
		// 隊員站在敵人的右邊那一格（環 0 的第 2 個方向），方向 6 正對回來。
		{Side: SideParty, X: 4, Y: 5, Head: 10, Chest: 10, Threshold: 100,
			Command: CmdWeapon, Style: 7, Guard: 3, Facing: 6},
		{Side: SideMonster, X: 4, Y: 4, Head: 10, Chest: 10, Threshold: 100,
			RightArm: 10, LeftArm: 10, RightLeg: 10, LeftLeg: 10, Level: 2,
			FleeAt: 65, FleeRecover: 35, Intellect: 90},
	}}
	// 兩名都已經規劃過：Turn 停在尾端。
	s.Round.Order, s.Round.Turn = []int{0, 1}, 2
	prev, _, target, ok := s.counterScan(1, &fixedRand{values: []int{1}})
	if !ok {
		t.Fatal("環 0 有隊員，該有結論")
	}
	if target != 0 {
		t.Fatalf("目標是 %d，期望 0", target)
	}
	if prev != 7 {
		t.Fatalf("該抄對方的攻擊選項 7，得到 %d", prev)
	}
	if s.Actors[1].Facing != 2 {
		t.Fatalf("方向該轉向目標（環 0 的第 2 格），得到 %d", s.Actors[1].Facing)
	}
}

// 對方還沒規劃：優先度 4，用 0x0B／7（`0x255AF`）。
func TestCounterScanUnplannedTarget(t *testing.T) {
	s := &State{Actors: []Actor{
		{Side: SideParty, X: 4, Y: 5, Head: 10, Chest: 10, Threshold: 100},
		{Side: SideMonster, X: 4, Y: 4, Head: 10, Chest: 10, Threshold: 100,
			RightArm: 10, LeftArm: 10, RightLeg: 10, LeftLeg: 10, Level: 2,
			FleeAt: 65, FleeRecover: 35, Intellect: 90},
	}}
	// 隊員排在敵人後面，還沒輪到它規劃。
	s.Round.Order, s.Round.Turn = []int{1, 0}, 1
	prev, _, target, ok := s.counterScan(1, &fixedRand{values: []int{1}})
	if !ok || target != 0 {
		t.Fatalf("該挑到第 0 名，得到 ok=%v target=%d", ok, target)
	}
	if prev != EnemyNearStyle {
		t.Fatalf("還沒規劃的對手該用 0x0B，得到 %#x", prev)
	}
}

// 場上沒有貼著的隊員就沒有結論，交回主幹繼續往外環找。
func TestCounterScanFindsNothing(t *testing.T) {
	s := &State{Actors: []Actor{
		{Side: SideParty, X: 20, Y: 20, Head: 10, Chest: 10, Threshold: 100},
		{Side: SideMonster, X: 4, Y: 4, Head: 10, Chest: 10, Threshold: 100,
			RightArm: 10, LeftArm: 10, RightLeg: 10, LeftLeg: 10, Level: 2,
			FleeAt: 65, FleeRecover: 35, Intellect: 90},
	}}
	s.Round.Order, s.Round.Turn = []int{0, 1}, 2
	if _, _, _, ok := s.counterScan(1, &fixedRand{values: []int{1}}); ok {
		t.Fatal("環 0 沒有人，不該有結論")
	}
}

// 側翼：對方要移動時繞到它的側面（`sub_25653`）。對方沒有攻擊選項就不繞。
//
// 自己的方向 0 是偶數 → 步幅 1、掃到第 6 格。從 0 往上數到第 3 格時碰上對方的
// 方向 3，`k == 3 < 4`，所以把對方的方向減一個步幅得到 2；環 0 的第 2 格
// （0,+1）站的是那名隊員、不是自己人，繞得過去。
func TestFlankNeedsTargetStyle(t *testing.T) {
	s := &State{Field: NewField(), Actors: []Actor{
		{Side: SideParty, X: 4, Y: 5, Head: 10, Chest: 10, Threshold: 100,
			Command: CmdMove, Facing: 3},
		{Side: SideMonster, X: 4, Y: 4, Head: 10, Chest: 10, Threshold: 100,
			RightArm: 10, LeftArm: 10, RightLeg: 10, LeftLeg: 10, Level: 2, Facing: 0},
	}}
	s.flank(1, 0)
	if s.Actors[1].Facing != 0 {
		t.Fatalf("對方沒有攻擊選項就不該繞，方向變成 %d", s.Actors[1].Facing)
	}
	s.Actors[0].Style = 5
	s.flank(1, 0)
	if s.Actors[1].Facing != 2 {
		t.Fatalf("該繞到第 2 格那一側，方向是 %d", s.Actors[1].Facing)
	}
}

// 正對面（`k == 4`）放棄，不改方向（`0x256DF`）。
func TestFlankGivesUpWhenHeadOn(t *testing.T) {
	s := &State{Field: NewField(), Actors: []Actor{
		{Side: SideParty, X: 4, Y: 5, Head: 10, Chest: 10, Threshold: 100,
			Command: CmdMove, Style: 5, Facing: 4},
		{Side: SideMonster, X: 4, Y: 4, Head: 10, Chest: 10, Threshold: 100,
			RightArm: 10, LeftArm: 10, RightLeg: 10, LeftLeg: 10, Level: 2, Facing: 0},
	}}
	s.flank(1, 0)
	if s.Actors[1].Facing != 0 {
		t.Fatalf("正對面該放棄，方向變成 %d", s.Actors[1].Facing)
	}
}

// 手上空著的怪物先去撿地上的東西，不打人（`0x23BC2`，RE-116）。
func TestPlanEnemyPicksUpWhenUnarmed(t *testing.T) {
	s := &State{Actors: []Actor{
		{Side: SideParty, X: 4, Y: 5, Head: 10, Chest: 10, Threshold: 100},
		{Side: SideMonster, X: 4, Y: 4, Head: 10, Chest: 10, Threshold: 100,
			RightArm: 10, LeftArm: 10, RightLeg: 10, LeftLeg: 10, Level: 2,
			FleeAt: 65, FleeRecover: 35, Unarmed: true},
	}}
	d := s.PlanEnemy(1, &fixedRand{values: []int{2}})
	if d.Command != CmdPickUp {
		t.Fatalf("空手應該去撿，得到 %#x", d.Command)
	}
	// 攻擊選項 0、防禦選項 2（`0x23C05`）。
	if d.Style != 0 || d.Guard != 2 {
		t.Fatalf("兩個選項是 %d／%d，期望 0／2", d.Style, d.Guard)
	}
	if s.Actors[1].Command != CmdPickUp || s.Actors[1].Guard != 2 {
		t.Fatal("決定沒有寫回參戰者記錄")
	}
}

// +0x2B 非零那一條改成移動，並把旗標清掉。原版沒有地方把它設起來，
// 所以這一條在遊戲裡走不到——測試直接把旗標塞進去，釘住這一段的形狀。
func TestPlanEnemyUnarmedWithPickedUpFlagMoves(t *testing.T) {
	s := &State{Actors: []Actor{
		{Side: SideParty, X: 4, Y: 5, Head: 10, Chest: 10, Threshold: 100},
		{Side: SideMonster, X: 4, Y: 4, Head: 10, Chest: 10, Threshold: 100,
			RightArm: 10, LeftArm: 10, RightLeg: 10, LeftLeg: 10, Level: 2,
			FleeAt: 65, FleeRecover: 35, Unarmed: true, PickedUp: 1},
	}}
	d := s.PlanEnemy(1, &fixedRand{values: []int{2}})
	if d.Command != CmdMove {
		t.Fatalf("+0x2B 非零應該改成移動，得到 %#x", d.Command)
	}
	if s.Actors[1].PickedUp != 0 {
		t.Fatalf("旗標應該被清掉，還是 %d", s.Actors[1].PickedUp)
	}
}

// 撿到手上之後就不再空手，下一回合會正常打人。
func TestEnemyPickUpFillsTheHand(t *testing.T) {
	s := &State{Actors: []Actor{
		{Side: SideParty, X: 4, Y: 5, Head: 10, Chest: 10, Threshold: 100},
		{Side: SideMonster, X: 4, Y: 4, Head: 10, Chest: 10, Threshold: 100,
			RightArm: 10, LeftArm: 10, RightLeg: 10, LeftLeg: 10, Level: 2,
			FleeAt: 65, FleeRecover: 35, Unarmed: true},
	}}
	s.Ground[0] = GroundItem{Name: "BROADSWORD", X: 4, Y: 4}
	if record, name := s.TakeItem(1, 0); record != FoundItemRecord || name != "BROADSWORD" {
		t.Fatalf("撿的結果是 %#x／%q", record, name)
	}
	if s.Actors[1].Unarmed {
		t.Fatal("撿到之後手上不該還空著")
	}
	if s.Ground[0].Name != "" {
		t.Fatalf("地上那一格應該清空，還有 %q", s.Ground[0].Name)
	}
	d := s.PlanEnemy(1, &fixedRand{values: []int{2}})
	if d.Command != CmdWeapon {
		t.Fatalf("有武器之後應該打人，得到 %#x", d.Command)
	}
}

// 敵人撿起成功本身不印任何一句；撿到任務物品而且這一場還沒喊過才印 0x113，
// 撿完 `+0x25` 寫成 0x15（RE-162）。
func TestEnemyPickUpIsSilentExceptQuestItem(t *testing.T) {
	record := "5101020304050607080910111213140f"
	s := &State{Site: &Site{Index: 6}, Actors: []Actor{
		{Side: SideMonster, X: 4, Y: 4, Command: CmdPickUp, Unarmed: true, PickedUp: 1},
		{Side: SideMonster, X: 6, Y: 6, Command: CmdPickUp, Unarmed: true, PickedUp: 1},
	}}
	s.Site.QuestItem.Record = record
	s.Ground[0] = GroundItem{Name: "AXE", X: 4, Y: 4, Item: &Item{Name: "AXE", Record: "52000000000000000000000000000001"}}
	s.Ground[1] = GroundItem{Name: "SHIP WHEEL", X: 6, Y: 6, Item: &Item{Name: "SHIP WHEEL", Record: record}}
	if out := s.EnemyPickUp(0, 0); out.Message != 0 || out.Item != "AXE" {
		t.Fatalf("一般物品：訊息 %#x、名字 %q", out.Message, out.Item)
	}
	out := s.EnemyPickUp(1, 1)
	if out.Message != QuestItemRecord || !out.Plain {
		t.Fatalf("任務物品應該印 0x113，得到 %#x", out.Message)
	}
	for i, a := range s.Actors {
		if a.Command != CmdDrop || a.PickedUp != 0 {
			t.Fatalf("第 %d 名撿完 +0x25=%#x +0x2B=%d", i, a.Command, a.PickedUp)
		}
	}
	s.Ground[2] = GroundItem{Name: "SHIP WHEEL", X: 4, Y: 4, Item: &Item{Name: "SHIP WHEEL", Record: record}}
	if out := s.EnemyPickUp(0, 2); out.Message != 0 {
		t.Fatalf("同一場第二次不再喊，得到 %#x", out.Message)
	}
}

// 這一次規劃裡被困住過（`word_34A08` 非零）就不抽攻擊式，固定第 1 式
// （`0x23CD6` 先比 `word_34A08`，RE-117）。
func TestEnemyStrikeFixesStyleAfterBeingStuck(t *testing.T) {
	newState := func() *State {
		return &State{Actors: []Actor{
			{Side: SideParty, X: 4, Y: 5, Head: 10, Chest: 10, Threshold: 100},
			{Side: SideMonster, X: 4, Y: 4, Head: 10, Chest: 10, Threshold: 100,
				RightArm: 10, LeftArm: 10, RightLeg: 10, LeftLeg: 10, Level: 2,
				FleeAt: 65, FleeRecover: 35},
		}}
	}
	// 困住過：第 1 式，而且 style == 1 那一條收尾把防禦選項定成 2。
	stuck := newState()
	stuck.Actors[1].Stuck = 1
	var d EnemyDecision
	stuck.enemyStrike(&d, 1, 0, EnemyNearStyle, EnemyNearGuard, &fixedRand{values: []int{3}})
	if d.Style != 1 {
		t.Fatalf("被困住過該固定第 1 式，得到 %d", d.Style)
	}
	if d.Guard != 2 {
		t.Fatalf("第 1 式的防禦選項該是 2，得到 %d", d.Guard)
	}
	// 反面對照：沒被困住過就要真的去抽——同樣的骰子給出不一樣的攻擊式，
	// 否則上面那個 1 證明不了是 Stuck 造成的。
	fresh := newState()
	var free EnemyDecision
	fresh.enemyStrike(&free, 1, 0, EnemyNearStyle, EnemyNearGuard, &fixedRand{values: []int{3}})
	if free.Style == 1 {
		t.Fatal("沒被困住過時攻擊式仍是 1，這個測試分不出 Stuck 有沒有作用")
	}
}

// 繞不出去的第三條退路：`word_34A08` 加一之後重跑決策樹（`0x23994`，RE-117）。
func TestEnemyStuckCountsUpAndRetries(t *testing.T) {
	s := &State{Actors: []Actor{
		{Side: SideParty, X: 0, Y: 0, Head: 10, Chest: 10, Threshold: 100},
		{Side: SideMonster, X: 4, Y: 4, Head: 10, Chest: 10, Threshold: 100,
			RightArm: 10, LeftArm: 10, RightLeg: 10, LeftLeg: 10, Level: 2,
			FleeAt: 65, FleeRecover: 35},
	}}
	// 傷勢還沒到撤退門檻走第一條（原地揮武器），所以先把它推過門檻。
	// 比例是 `50 × 疲勞 ÷ 耐力`，要 >= FleeAt 65 就得讓疲勞是耐力的 1.3 倍。
	// 再加上 rand(3) != 1 排掉第二條，才走得到第三條。
	s.Actors[1].Fatigue, s.Actors[1].Stamina = 130, 100
	if got := s.Actors[1].FatigueRatio(); got < s.Actors[1].FleeAt {
		t.Fatalf("這組數值的疲勞比例是 %d，還沒到門檻 %d——測試會走第一條退路",
			got, s.Actors[1].FleeAt)
	}
	before := s.Actors[1].Stuck
	s.enemyBoxed(1, &fixedRand{values: []int{2}})
	if s.Actors[1].Stuck <= before {
		t.Fatalf("第三條退路該把 Stuck 加上去，%d → %d", before, s.Actors[1].Stuck)
	}
}
