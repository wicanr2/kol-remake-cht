package battle

import "testing"

// 固定骰：依序回傳給定的值，用完之後一律回最後一個。
type fixedRoll struct {
	values []int
	i      int
}

func (f *fixedRoll) Roll(int) int {
	if f.i < len(f.values) {
		v := f.values[f.i]
		f.i++
		return v
	}
	if len(f.values) == 0 {
		return 1
	}
	return f.values[len(f.values)-1]
}

// 指令的時間成本：三個固定值加上查表，攻擊選項為 1 時再除以 2（`0x1FDDD`）。
func TestTimeCost(t *testing.T) {
	if got := TimeCost(CmdUnarmed, 5); got != 8 {
		t.Errorf("徒手應該是 8，得到 %d", got)
	}
	if got := TimeCost(CmdMove, 0x3C); got != 6 {
		t.Errorf("移動應該是 6，得到 %d", got)
	}
	if got := TimeCost(CmdMissile, 3); got != 1 {
		t.Errorf("射擊應該是 1，得到 %d", got)
	}
	// 表是 00 0f 07 07 07 06 06 06 08 08 08 …
	if got := TimeCost(CmdWeapon, 2); got != 7 {
		t.Errorf("攻擊選項 2 應該查到 7，得到 %d", got)
	}
	// 攻擊選項 1 查到 0x0F 之後再除以 2。
	if got := TimeCost(CmdWeapon, 1); got != 7 {
		t.Errorf("攻擊選項 1 應該是 0x0F ÷ 2 ＝ 7，得到 %d", got)
	}
	if got := TimeCost(CmdWeapon, 8); got != 8 {
		t.Errorf("攻擊選項 8 應該查到 8，得到 %d", got)
	}
}

// 執行鍵值的四項扣除與一次擲骰（`0x1FD6C`）。
func TestExecKey(t *testing.T) {
	a := Actor{Quickness: 80, Encumbrance: 200, WeaponSpeed: 40,
		Command: CmdWeapon, Style: 2, Threshold: 100}
	// 沒受傷：80 − 0 − 200÷10 − (40÷2 × 7 ÷ 10) ＝ 80 − 20 − 14 ＝ 46，加擲骰 10。
	if got := ExecKey(a, 10); got != 56 {
		t.Errorf("鍵值應該是 56，得到 %d", got)
	}
	// 受傷之後快捷先打折：傷勢 ＝ 50 × 50 ÷ 100 ＝ 25，80 − 80×25÷200 ＝ 70。
	a.Damage = 50
	if got := ExecKey(a, 10); got != 46 {
		t.Errorf("受傷之後應該是 46，得到 %d", got)
	}
	// 擲骰是先加進去、之後才夾 0（`0x1FE6F` 加、`0x1FE84` 夾），
	// 所以扣到很負的人加了骰也還是 0。
	slow := Actor{Quickness: 1, Encumbrance: 900, WeaponSpeed: 200,
		Command: CmdUnarmed, Threshold: 100}
	if got := ExecKey(slow, 1); got != 0 {
		t.Errorf("扣成負數再加擲骰仍應夾成 0，得到 %d", got)
	}
}

// 射擊時武器種類不是 0x12／0x16 就把鍵值除以 2（`0x1FEDE`）。
func TestExecKeyMissileHalves(t *testing.T) {
	a := Actor{Quickness: 100, Command: CmdMissile, Threshold: 100,
		Weapon: Weapon{Kind: 0x20}}
	if got := ExecKey(a, 0); got != 50 {
		t.Errorf("一般射擊武器應該減半成 50，得到 %d", got)
	}
	a.Weapon.Kind = 0x12
	if got := ExecKey(a, 0); got != 100 {
		t.Errorf("種類 0x12 不減半，應該是 100，得到 %d", got)
	}
	a.Weapon.Kind = 0x16
	if got := ExecKey(a, 0); got != 100 {
		t.Errorf("種類 0x16 不減半，應該是 100，得到 %d", got)
	}
}

// 鍵值大的先動；空位不進順序。
func TestBeginExecutionOrder(t *testing.T) {
	s := &State{Actors: make([]Actor, 4)}
	for i := range s.Actors {
		s.Actors[i] = Actor{Side: SideParty, Threshold: 100, Quickness: 10 * (i + 1)}
	}
	s.Actors[3].Side = SideEmpty
	s.BeginExecution(&fixedRoll{values: []int{0}})
	want := []int{2, 1, 0} // 快捷 30、20、10
	if len(s.Exec.Order) != len(want) {
		t.Fatalf("順序長度 %d，應該是 %d", len(s.Exec.Order), len(want))
	}
	for i := range want {
		if s.Exec.Order[i] != want[i] {
			t.Fatalf("順序是 %v，應該是 %v", s.Exec.Order, want)
		}
	}
}

// 面前那一格站著誰：遮罩只看 Flag1E，不分敵我（`sub_26D76:0x26DA5`）。
func TestFacingTarget(t *testing.T) {
	s := &State{Actors: make([]Actor, 3)}
	s.Actors[0] = Actor{Side: SideParty, X: 10, Y: 10, Facing: 2, Threshold: 100}
	s.Actors[0].SetParts()
	// 方向 2 是「右」＝ 欄 +1。
	s.Actors[1] = Actor{Side: SideMonster, X: 10, Y: 11, Threshold: 100}
	s.Actors[1].SetParts()
	s.Actors[2] = Actor{Side: SideEmpty}
	if got := s.FacingTarget(0); got != 1 {
		t.Fatalf("面前應該是 1，得到 %d", got)
	}
	// Flag1E 非零的人不算目標。
	s.Actors[1].Flag1E = 1
	if got := s.FacingTarget(0); got != -1 {
		t.Fatalf("Flag1E 非零不該算目標，得到 %d", got)
	}
	s.Actors[1].Flag1E = 0
	// 自己人站在面前一樣打得到。
	s.Actors[1].Side = SideParty
	if got := s.FacingTarget(0); got != 1 {
		t.Fatalf("自己人也該算目標，得到 %d", got)
	}
	// 面前沒人。
	s.Actors[0].Facing = 6
	if got := s.FacingTarget(0); got != -1 {
		t.Fatalf("面前沒人應該回 −1，得到 %d", got)
	}
}

// 一手的分派：攻擊選項 0 什麼都不做、面前沒人就揮空、有人就結算。
func TestActDispatch(t *testing.T) {
	newState := func() *State {
		s := &State{Actors: make([]Actor, 2)}
		s.Actors[0] = Actor{Side: SideParty, X: 10, Y: 10, Facing: 2,
			Threshold: 100, Quickness: 50, Strength: 90, Level: 3,
			Attack: 400, Weapon: Weapon{Base: 3, Count: 1, Sides: 5}}
		s.Actors[0].SetParts()
		s.Actors[1] = Actor{Side: SideMonster, X: 10, Y: 11, Threshold: 100}
		s.Actors[1].SetParts()
		s.Exec.Order = []int{0}
		return s
	}
	// 攻擊選項 0。
	s := newState()
	s.Actors[0].Command, s.Actors[0].Style = CmdWeapon, 0
	r, ok := s.Act(&fixedRoll{values: []int{50}})
	if !ok || r.Command != 0 || r.Message != NothingRecord {
		t.Fatalf("攻擊選項 0 應該歸零並印 %#x，得到 %#x／%#x", NothingRecord, r.Command, r.Message)
	}
	// 面前沒人。
	s = newState()
	s.Actors[0].Command, s.Actors[0].Style = CmdUnarmed, 2
	s.Actors[0].Facing = 6
	s.Exec.Turn = 0
	r, _ = s.Act(&fixedRoll{values: []int{50}})
	if r.Message != MissUnarmedRecord || r.Target != -1 {
		t.Fatalf("徒手揮空應該印 %#x，得到 %#x", MissUnarmedRecord, r.Message)
	}
	s = newState()
	s.Actors[0].Command, s.Actors[0].Style = CmdWeapon, 2
	s.Actors[0].Facing = 6
	r, _ = s.Act(&fixedRoll{values: []int{50}})
	if r.Message != MissWeaponRecord {
		t.Fatalf("武器揮空應該印 %#x，得到 %#x", MissWeaponRecord, r.Message)
	}
	// 面前有人：走結算。
	s = newState()
	s.Actors[0].Command, s.Actors[0].Style = CmdWeapon, 2
	r, _ = s.Act(&fixedRoll{values: []int{99, 1, 5, 50}})
	if r.Target != 1 || r.Outcome == nil {
		t.Fatalf("面前有人應該走結算，得到 target %d outcome %v", r.Target, r.Outcome)
	}
	// 休息與怯戰。
	s = newState()
	s.Actors[0].Command = CmdRest
	r, _ = s.Act(&fixedRoll{values: []int{50}})
	if r.Message != RestingRecord {
		t.Fatalf("休息應該印 %#x，得到 %#x", RestingRecord, r.Message)
	}
	s = newState()
	s.Actors[0].Command = CmdFlinch
	r, _ = s.Act(&fixedRoll{values: []int{50}})
	if r.Message != FlinchRecord || s.Actors[0].Guard != 2 {
		t.Fatalf("怯戰應該印 %#x 並把防禦選項設 2，得到 %#x／%d",
			FlinchRecord, r.Message, s.Actors[0].Guard)
	}
	// 還沒實作的那幾手要被標出來，不能靜靜跳過。
	s = newState()
	s.Actors[0].Command = CmdPickUp
	r, _ = s.Act(&fixedRoll{values: []int{50}})
	if !r.Unimplemented {
		t.Fatal("撿東西這一手應該標成還沒實作")
	}
}

// 移動方式的圖號換算：0x3C…0x40 是 0…4，0x35 落成負數再換成 5（`0x263A6`）。
func TestMoveForm(t *testing.T) {
	want := map[int]int{0x3C: 0, 0x3D: 1, 0x3E: 2, 0x3F: 3, 0x40: 4, 0x35: 5}
	for style, form := range want {
		if got := MoveForm(style); got != form {
			t.Errorf("圖號 %#x 應該是方式 %d，得到 %d", style, form, got)
		}
	}
}

// 移動一次只走一格，而且方式 3 起會把 Flag1E 設起來（RE-085）。
func TestMoveOneStep(t *testing.T) {
	s := &State{Actors: make([]Actor, 1)}
	s.Actors[0] = Actor{Side: SideParty, X: 20, Y: 20, Facing: 2,
		Threshold: 100, Command: CmdMove, Style: 0x3C}
	s.Actors[0].SetParts()
	s.Exec.Order = []int{0}
	r, ok := s.Act(&fixedRoll{values: []int{50}})
	if !ok || r.Message != MovedRecord {
		t.Fatalf("應該走得動並印 %#x，得到 %#x", MovedRecord, r.Message)
	}
	// 方向 2 是右＝欄 +1。
	if s.Actors[0].X != 20 || s.Actors[0].Y != 21 {
		t.Fatalf("應該走到 (20, 21)，得到 (%d, %d)", s.Actors[0].X, s.Actors[0].Y)
	}
	if s.Actors[0].Flag1E != 0 {
		t.Fatal("方式 0 不該把 Flag1E 設起來")
	}
	// 方式 3 起會設 Flag1E。
	s.Exec.Turn = 0
	s.Actors[0].Style = 0x3F
	s.Act(&fixedRoll{values: []int{50}})
	if s.Actors[0].Flag1E != 1 {
		t.Fatalf("方式 3 應該把 Flag1E 設成 1，得到 %d", s.Actors[0].Flag1E)
	}
}

// 界限是 4…0x3B，最外圈四格進不去（`0x2650B`）。
func TestMoveClampedToField(t *testing.T) {
	for _, c := range []struct {
		x, y, facing int
	}{
		{FieldMin, 20, 0},  // 往上出界
		{20, FieldMin, 6},  // 往左出界
		{FieldMax, 20, 4},  // 往下出界
		{20, FieldMax, 2},  // 往右出界
	} {
		s := &State{Actors: make([]Actor, 1)}
		s.Actors[0] = Actor{Side: SideParty, X: c.x, Y: c.y, Facing: c.facing,
			Threshold: 100, Command: CmdMove, Style: 0x3C}
		s.Actors[0].SetParts()
		s.Exec.Order = []int{0}
		r, _ := s.Act(&fixedRoll{values: []int{50}})
		if r.Message != BlockedRecord {
			t.Errorf("(%d, %d) 往 %d 應該走不動，得到 %#x", c.x, c.y, c.facing, r.Message)
		}
		if s.Actors[0].X != c.x || s.Actors[0].Y != c.y {
			t.Errorf("走不動卻動了：(%d, %d)", s.Actors[0].X, s.Actors[0].Y)
		}
	}
}

// 誰擋誰：沒有 Flag1E 的擋沒有 Flag1E 的，有的擋有的（`0x2649C` 的遮罩 8／4）。
func TestMoveBlockedByActor(t *testing.T) {
	build := func(mine, theirs int) *State {
		s := &State{Actors: make([]Actor, 2)}
		s.Actors[0] = Actor{Side: SideParty, X: 20, Y: 20, Facing: 2,
			Threshold: 100, Command: CmdMove, Style: 0x3C, Flag1E: mine}
		s.Actors[0].SetParts()
		s.Actors[1] = Actor{Side: SideMonster, X: 20, Y: 21,
			Threshold: 100, Flag1E: theirs}
		s.Actors[1].SetParts()
		s.Exec.Order = []int{0}
		return s
	}
	// 兩邊都沒有 Flag1E：擋。
	s := build(0, 0)
	if r, _ := s.Act(&fixedRoll{values: []int{50}}); r.Message != BlockedRecord {
		t.Errorf("同類應該擋住，得到 %#x", r.Message)
	}
	// 我沒有、對方有：不擋。注意移動方式 0x3C 不會改我的 Flag1E。
	s = build(0, 1)
	if r, _ := s.Act(&fixedRoll{values: []int{50}}); r.Message != MovedRecord {
		t.Errorf("不同類不該擋，得到 %#x", r.Message)
	}
	// 兩邊都有：擋。用方式 0x3F 讓自己的 Flag1E 也是 1。
	s = build(1, 1)
	s.Actors[0].Style = 0x3F
	if r, _ := s.Act(&fixedRoll{values: []int{50}}); r.Message != BlockedRecord {
		t.Errorf("同類應該擋住，得到 %#x", r.Message)
	}
}

// 擋路與擋視線是兩張不同的位元圖（RE-085）。
func TestBlocksMovementDiffersFromSight(t *testing.T) {
	// 磚 0 擋視線但不擋路。
	if !Blocks(0) || BlocksMovement(0) {
		t.Error("磚 0 應該擋視線但不擋路")
	}
	// 牆（0x29）兩張表都擋。
	if !Blocks(0x29) || !BlocksMovement(0x29) {
		t.Error("磚 0x29 兩張表都該擋")
	}
	// 草地（0x09）兩張表都不擋。
	if Blocks(0x09) || BlocksMovement(0x09) {
		t.Error("磚 0x09 兩張表都不該擋")
	}
	// 超出範圍一律當成擋。
	if !BlocksMovement(-1) || !BlocksMovement(80) {
		t.Error("超出範圍應該當成擋")
	}
}

// 倒下的人與空位跳過（`0x1FF77`）。
func TestActSkipsDown(t *testing.T) {
	s := &State{Actors: make([]Actor, 3)}
	s.Actors[0] = Actor{Side: SideParty, Threshold: 100, Command: CmdRest}
	s.Actors[0].SetParts()
	s.Actors[0].Head = 0 // 倒下
	s.Actors[1] = Actor{Side: SideEmpty}
	s.Actors[2] = Actor{Side: SideParty, Threshold: 100, Command: CmdRest}
	s.Actors[2].SetParts()
	s.Exec.Order = []int{0, 1, 2}
	r, ok := s.Act(&fixedRoll{values: []int{50}})
	if !ok || r.Actor != 2 {
		t.Fatalf("應該跳到第 2 名，得到 actor %d ok %v", r.Actor, ok)
	}
	if _, ok := s.Act(&fixedRoll{values: []int{50}}); ok {
		t.Fatal("後面沒有人了，應該回 false")
	}
}

// 移動成功之後原版要按兩次鍵：印完訊息等一次、重畫戰場再等一次（`0x265C9`／
// `0x265E1`）。走不動只有一次。
func TestMoveRedraw(t *testing.T) {
	build := func(facing int) *State {
		s := &State{Actors: make([]Actor, 1)}
		s.Actors[0] = Actor{Side: SideParty, X: 20, Y: 20, Facing: facing,
			Threshold: 100, Command: CmdMove, Style: 0x3C}
		s.Actors[0].SetParts()
		s.Exec.Order = []int{0}
		return s
	}
	s := build(2)
	r, _ := s.Act(&fixedRoll{values: []int{50}})
	if r.Message != MovedRecord || !r.Redraw {
		t.Fatalf("走成功應該要重畫再等一次，得到 %#x redraw=%v", r.Message, r.Redraw)
	}
	// 走到界外：只印一句，不重畫。
	s = build(0)
	s.Actors[0].X = FieldMin
	s.Exec.Turn = 0
	r, _ = s.Act(&fixedRoll{values: []int{50}})
	if r.Message != BlockedRecord || r.Redraw {
		t.Fatalf("走不動不該重畫，得到 %#x redraw=%v", r.Message, r.Redraw)
	}
}
