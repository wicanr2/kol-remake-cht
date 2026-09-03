package battle

import "sort"

// 執行段（RE-084）。整輪規劃做完、玩家在「Begin next round?」答應之後，原版
// **另外排一次順序**：這一次看的是快捷、負重與指令的時間成本，鍵值大的先動。
// 排完由陣列尾端往回走，一名一名把 +0x25 讀回來做出來。

// commandTime 是 `[0x10B2]`：指令的時間成本，索引是攻擊選項。
var commandTime = [16]int{
	0x00, 0x0F, 0x07, 0x07, 0x07, 0x06, 0x06, 0x06,
	0x08, 0x08, 0x08, 0x00, 0x00, 0x02, 0x01, 0x02,
}

// 幾個指令有固定的時間成本，不查表（`0x1FDE0`…`0x1FE16`）。
const (
	timeUnarmed = 8
	timeMove    = 6
	timeMissile = 1
)

// 射擊不打折的兩種武器種類（`0x1FEDE`）。
const (
	missileKindA = 0x12
	missileKindB = 0x16
)

// 執行段會用到的 KOLDATA 記錄。
const (
	// PackFullRecord 是撿東西時十六格都滿了那一句。
	PackFullRecord = 0x10C
	// NothingRecord 是指令被歸零之後那一句。
	NothingRecord = 0x117
	// FlinchRecord 是怯戰那一句；RestingRecord 是「is resting!」。
	FlinchRecord   = 0x118
	RestingRecord  = 0x119
	// MissUnarmedRecord／MissWeaponRecord 是面前沒有人時揮空那兩句。
	MissUnarmedRecord = 0x11F
	MissWeaponRecord  = 0x120
)

// CmdFlinch 是怯戰檢定沒過時被換上的指令碼（`0x2016A`）。
const CmdFlinch = 0x0B

// ring0 是環 0 的九個位移（RE-081）：八個方向照方向編號排，第 9 項是 (0, 0)，
// 也就是自己站的那一格。Row 是原版的 X、Col 是 Y。
var ring0 = [9]Cell{
	{-1, 0}, {-1, 1}, {0, 1}, {1, 1}, {1, 0}, {1, -1}, {0, -1}, {-1, -1}, {0, 0},
}

// Execution 是執行段的出手順序。
type Execution struct {
	// Order 是依鍵值排好的參戰者索引，**由先動到後動**。
	Order []int
	// Turn 是現在輪到 Order 的第幾個。
	Turn int
	// Log 是這一回合每一手的結果，依實際出手順序。
	Log []ActResult
}

// RunAll 把這一回合的執行段一次跑完，結果收進 Log。
//
// 原版每一手都會印一句訊息再等玩家按鍵（`sub_1B100(5)`），那一段還沒實作，
// 所以這裡先一次跑完——結算本身照原版做，只有「逐句顯示」沒有做（SPEC-008）。
func (s *State) RunAll(rng Rand) []ActResult {
	s.Exec.Log = nil
	for {
		r, ok := s.Act(rng)
		if !ok {
			return s.Exec.Log
		}
		s.Exec.Log = append(s.Exec.Log, r)
	}
}

// TimeCost 是一個指令的時間成本（`0x1FDDD`…`0x1FE50`）。
func TimeCost(command, style int) int {
	switch command {
	case CmdUnarmed:
		return timeUnarmed
	case CmdMove:
		return timeMove
	case CmdMissile:
		return timeMissile
	}
	t := 0
	if style >= 0 && style < len(commandTime) {
		t = commandTime[style]
	}
	if style == 1 {
		t /= 2
	}
	return t
}

// ExecKey 是一名參戰者這一回合的**執行**鍵值（`0x1FD6C`…`0x1FF06`）：
//
//	鍵值 = 快捷 − 快捷 × 傷勢 ÷ 200 − 負重 ÷ 10
//	鍵值 −= (武器速度 ÷ 2) × 時間成本 ÷ 10
//	鍵值 += 擲骰，負數夾成 0
//	射擊而且武器種類不是 0x12／0x16 → 再除以 2
//
// roll 是 1…100 的擲骰結果。**鍵值大的先動**。
//
// 原版在移動且攻擊選項為 0 時也算了一次除以 2，但結果沒有寫回，所以那一條
// 實際上不生效（RE-084）；這裡照原版的行為，不實作那一條。
func ExecKey(a Actor, roll int) int {
	key := a.Quickness - a.Quickness*a.Injury()/200
	key -= a.Encumbrance / 10
	key -= (a.WeaponSpeed / 2) * TimeCost(a.Command, a.Style) / 10
	key += roll
	if key < 0 {
		key = 0
	}
	if a.Command == CmdMissile &&
		a.Weapon.Kind != missileKindA && a.Weapon.Kind != missileKindB {
		key /= 2
	}
	return key
}

// BeginExecution 依執行鍵值排出這一回合誰先動。空位不排。
func (s *State) BeginExecution(rng Rand) {
	keys := make([]int, len(s.Actors))
	order := make([]int, 0, len(s.Actors))
	for i, a := range s.Actors {
		if a.Side == SideEmpty {
			continue
		}
		keys[i] = ExecKey(a, rng.Roll(PlanRoll))
		order = append(order, i)
	}
	// 原版排小到大再從尾端往回走。等價於由大到小，而相同鍵值時索引大的先動。
	sort.SliceStable(order, func(x, y int) bool {
		if keys[order[x]] != keys[order[y]] {
			return keys[order[x]] > keys[order[y]]
		}
		return order[x] > order[y]
	})
	s.Exec.Order = order
	s.Exec.Turn = 0
}

// ActResult 是一手做出來的結果。
type ActResult struct {
	// Actor 是誰動的；Command 是實際做出來的指令碼（可能被歸零或換成怯戰）。
	Actor   int
	Command int
	// Target 是打到誰；−1 表示沒有目標。
	Target int
	// Outcome 是打中的話那一次結算的結果。
	Outcome *Outcome
	// Message 是要印的 KOLDATA 記錄；0 表示沒有訊息。
	Message int
	// Unimplemented 為真表示這一手 remake 還沒做出來，不是原版沒有動作。
	Unimplemented bool
}

// FacingTarget 回傳面前那一格站著誰（`sub_26D76:0x26DA5` 的
// `sub_24BB6(0, 方向, 8)`）。遮罩 8 只要求對方的 Flag1E 為零，**不分敵我**——
// 面前站著自己人一樣打得到。沒有人就回 −1。
func (s *State) FacingTarget(i int) int {
	if i < 0 || i >= len(s.Actors) {
		return -1
	}
	a := s.Actors[i]
	facing := a.Facing
	if facing < 0 || facing >= len(ring0) {
		return -1
	}
	x, y := a.X+ring0[facing].Row, a.Y+ring0[facing].Col
	for j := range s.Actors {
		if j == i {
			continue
		}
		b := s.Actors[j]
		if b.Side == SideEmpty || b.Down() || b.Flag1E != 0 {
			continue
		}
		if b.X == x && b.Y == y {
			return j
		}
	}
	return -1
}

// Act 做出下一手，回報做完了沒有。第二個回傳值為 false 表示這一回合的執行段
// 已經走完。
func (s *State) Act(rng Rand) (ActResult, bool) {
	for s.Exec.Turn < len(s.Exec.Order) {
		i := s.Exec.Order[s.Exec.Turn]
		s.Exec.Turn++
		a := &s.Actors[i]
		// 已經退出戰鬥或空掉的位置跳過（`0x1FF77`）。
		if a.Side == SideEmpty || a.Down() {
			continue
		}
		return s.act(i, rng), true
	}
	return ActResult{}, false
}

// act 是一手的分派（`0x2065E`…`0x2070F`）。
func (s *State) act(i int, rng Rand) ActResult {
	a := &s.Actors[i]
	out := ActResult{Actor: i, Command: a.Command, Target: -1}
	switch a.Command {
	case CmdWeapon, CmdUnarmed:
		// 攻擊選項為 0 就把指令歸零，什麼都不做（`0x2067C`）。
		if a.Style == 0 {
			out.Command = 0
			out.Message = NothingRecord
			return out
		}
		target := s.FacingTarget(i)
		if target < 0 {
			out.Message = MissWeaponRecord
			if a.Command == CmdUnarmed {
				out.Message = MissUnarmedRecord
			}
			return out
		}
		outcome := Resolve(a, &s.Actors[target], a.Style, rng)
		out.Target = target
		out.Outcome = &outcome
		return out
	case CmdRest:
		out.Message = RestingRecord
		return out
	case CmdFlinch:
		// 怯戰：防禦選項改成 2 再印那一句（`0x2062B`）。
		a.Guard = 2
		out.Message = FlinchRecord
		return out
	case 0:
		out.Message = NothingRecord
		return out
	}
	// 移動、撿、丟與背包那幾手的內容還沒解完（SPEC-010 末段），不假裝做過。
	out.Unimplemented = true
	return out
}
