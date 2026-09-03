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
	s.Actors[0].Command = CmdMove
	r, _ = s.Act(&fixedRoll{values: []int{50}})
	if !r.Unimplemented {
		t.Fatal("移動這一手應該標成還沒實作")
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
