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
	// BlockedRecord／MovedRecord 是移動走不動與走了一步那兩句（RE-085）。
	BlockedRecord = 0x10A
	MovedRecord   = 0x10B
)

// 戰場上走得到的範圍：兩軸都夾在 4…0x3B（`sub_26384:0x2650B`）。
// 陣列是 64×64，但最外圈四格進不去。
const (
	FieldMin = 4
	FieldMax = 0x3B
)

// 移動方式的第一個圖號；方式 = 圖號 − MoveFormBase，0x35 會落成負數再換成 5。
const MoveFormBase = 0x3C

// MountingForm 是會把 Flag1E 設起來的第一個移動方式（`0x263BD`）。
const MountingForm = 3

// StuckTile 是 Flag1E 非零時腳下踩到就走不動的磚（`0x2647D`）。
const StuckTile = 0x0C

// MoveForm 把攻擊選項欄放的移動方式圖號換成方式編號 0…5（`0x263A6`）。
func MoveForm(style int) int {
	form := style - MoveFormBase
	if form < 0 {
		return 5
	}
	return form
}

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
	// Text 是現在顯示的那一句，已經展開代換碼並斷好行。battle 這一層不認得
	// 文字庫，由呼叫端填（RE-086／RE-087）。
	Text []string
	// Pending 是同一手的下一句。攻擊命中時原版分兩句、各等一次鍵
	// （`sub_26D76:0x2746A` 與 `sub_277F4:0x27B66`，RE-088）。
	Pending []string
}

// Step 做出下一手。原版每一手印一句訊息再等玩家按一個鍵（`sub_1B100(5)`），
// 所以這裡一次只走一手；回報執行段是不是還沒走完。
//
// 走完之後自己收尾：數兩邊倒下的人，依 RE-091 的四個條件決定這一場的去向，
// 沒分勝負才排下一回合的規劃。
func (s *State) Step(rng Rand) (ActResult, bool) {
	r, ok := s.Act(rng)
	if !ok {
		s.Exec.Text = nil
		switch v := s.RoundVerdict(); v {
		case VerdictLost, VerdictWon:
			s.Verdict = v
			s.Stage = StageDone
			if v == VerdictWon {
				s.Gold = GoldReward(s.Monster.Gold.Rolls, s.Monster.Gold.Sides,
					s.EnemyCount(), s.PartyCount(), rng)
			}
			return r, false
		}
		s.BeginPlanning(rng)
		s.AdvanceToParty()
		s.BeginTurn()
		return r, false
	}
	s.Exec.Log = append(s.Exec.Log, r)
	return r, true
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
	// Redraw 為真表示這一句之後原版會重畫戰場再等一次鍵
	// （`sub_26384:0x265D5`，移動成功才有）。
	Redraw bool
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
	// 規劃段應該已經替敵人填好指令了（RE-081／RE-092）。真的空著就標成還沒實作，
	// 不要讓畫面看起來像「原版選了什麼都不做」——那是 AI 沒接上時的舊症狀。
	if a.Side != SideParty && a.Command == 0 {
		out.Unimplemented = true
		return out
	}
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
	case CmdMove, CmdDismount:
		return s.move(i)
	}
	// 剩下的幾個指令碼。**撿（0x15）與丟（0x24）不會走到這裡**——執行段的分派只認
	// 十個指令碼，那兩個不在其中，它們是規劃段的即時動作（RE-093）。真的落到這裡
	// 表示規劃段漏接了，標成還沒實作，不假裝做過。
	out.Unimplemented = true
	return out
}

// move 是移動那一手（`sub_26384`，RE-085）。一次只走一格，走不動就印記錄 0x10A。
func (s *State) move(i int) ActResult {
	a := &s.Actors[i]
	out := ActResult{Actor: i, Command: a.Command, Target: -1}
	// 方式 3…5 會把 Flag1E 設起來（`0x263BD`）。
	if MoveForm(a.Style) > MountingForm-1 {
		a.Flag1E = 1
	}
	if a.Facing < 0 || a.Facing >= len(ring0) {
		out.Message = BlockedRecord
		return out
	}
	step := ring0[a.Facing]
	x, y := a.X+step.Row, a.Y+step.Col
	if s.moveBlocked(i, x, y) {
		out.Message = BlockedRecord
		return out
	}
	a.X, a.Y = x, y
	out.Message = MovedRecord
	// 走成功之後原版印完訊息等一次鍵、重畫戰場、再等一次（`0x265C9`／`0x265E1`）。
	out.Redraw = true
	return out
}

// moveBlocked 是那四道判斷（RE-085）：地形、腳下、有人、界限。
func (s *State) moveBlocked(i, x, y int) bool {
	a := s.Actors[i]
	// 界限：兩軸都要落在 4…0x3B。
	if x < FieldMin || x > FieldMax || y < FieldMin || y > FieldMax {
		return true
	}
	// 地形。Flag1E 非零時只有三種磚真的擋得住。
	if s.Field != nil {
		tile := s.Field.At(x, y)
		if BlocksMovement(tile) {
			if a.Flag1E == 0 {
				return true
			}
			if (tile > 0x1F && tile < 0x23) || tile == 0x2E || tile > 0x37 {
				return true
			}
		}
	}
	// 腳下：Flag1E 非零時踩到 0x0C 就走不動。
	if a.Flag1E != 0 && s.Field != nil && s.Field.At(a.X, a.Y) == StuckTile {
		return true
	}
	// 有人：沒有 Flag1E 的擋沒有 Flag1E 的，有的擋有的。
	for j := range s.Actors {
		if j == i {
			continue
		}
		b := s.Actors[j]
		if b.Side == SideEmpty || b.Down() || b.X != x || b.Y != y {
			continue
		}
		if (a.Flag1E == 0) == (b.Flag1E == 0) {
			return true
		}
	}
	return false
}
