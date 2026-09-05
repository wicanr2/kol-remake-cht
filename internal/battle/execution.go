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
	// BlockedByRecord 是下馬時那一格還有別人：`is blocked!`。
	// LandsRecord 是下來了：`lands!`。
	BlockedByRecord = 0x11A
	LandsRecord     = 0x11B
	// 射擊那一手的四句（RE-102）。MissileIntoRecord 後面要接 TreeRecord 或
	// WallRecord；MissileNothingRecord 是射過去空無一物。
	MissileIntoRecord    = 0x121
	TreeRecord           = 0x122
	WallRecord           = 0x123
	MissileNothingRecord = 0x124
	// ReloadRecord 是這一回合用來裝填那一句（`0x20803`，RE-093）。
	ReloadRecord = 0x15D
)

// NoReloadKinds 是不用裝填的三種武器種類（`0x207AF`…`0x207C8`）。
var NoReloadKinds = [3]int{6, 10, 11}

// NeedsReload 回報這一種武器射之前要不要先裝填。
func NeedsReload(kind int) bool {
	for _, k := range NoReloadKinds {
		if kind == k {
			return false
		}
	}
	return true
}

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

// ring0 是環 0 的九個位移（RE-081），第 9 項是 (0, 0)——自己站的那一格。
//
// **座標軸與原版是對調的，別「修正」它**（RE-095）：原版參戰者記錄的 +0x1F 是
// 欄、+0x20 是列，而 remake 的 `Actor.X` 是列（畫到畫面的垂直方向）、`Y` 是欄。
// 抄環位移表時把原版的 X 欄放進了 `Row`、Y 欄放進 `Col`，剛好與座標軸的對調抵消，
// 所以 **`ring0` 的索引在 remake 這邊等於方向編號**（0 是上、2 是右…）。
//
// 在原版那邊兩者是不同的：環索引 0 是「左」、方向編號 0 是「上」，中間差一個
// `i = (6 − d) mod 8` 的映射。戰場格子陣列也一樣是轉置過的（RE-080），
// 整條鏈一致，畫面因此正確——四對戰場對拍是它的證據。
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
	// Pending 是同一手還沒印完的幾句，一句等一次鍵。攻擊命中時原版分兩句
	// （`sub_26D76:0x2746A` 與 `sub_277F4:0x27B66`，RE-088），擊退成立還會再多
	// 一到兩句（RE-095），所以這裡是一個佇列而不是單獨一段。
	Pending [][]string
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
				// 打贏的第一件事是把任務標成完成（`0x211CB` 的 sub_28F5D，
				// RE-101）。隨機遭遇不做——原版看 word_34968，remake 看
				// Site.Index。
				s.QuestDone = s.markQuestDone()
				// 獎勵整段只在一般戰鬥發（`0x2124E` 比 `word_345DE == 0xFFFF`，
				// RE-096）：競技場與訓練戰跳過，改走升階那一條。
				// 順序照原版：冒險點數先算（不擲骰），金冠、掉落物後算。
				if !s.RankFight {
					s.Adventure = s.AdventureReward()
					s.Gold = s.GoldAward(rng)
					s.DropSpoils(s.SpoilItems, rng)
				}
			}
			return r, false
		}
		// 沒分勝負：先看要不要問「要逃嗎」（`0x21045`，RE-103），
		// 答完才排下一回合。
		if s.OfferFlee() {
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
	// Item 是這一手牽涉的物品名字。丟與撿都要它——原版印訊息時東西還在手上或
	// 還在地上，remake 這邊做完才印，名字得另外帶出來（RE-094）。
	Item string
	// Knock 是擊退或打暈要接著印的幾個記錄（RE-095）；沒有就是 nil。
	Knock []int
	// Cast 是施法打中之後的結果（RE-100）；沒施法或沒生效就是 nil。
	// 訊息要用它接「The spell has a … effect.」那一句。
	Cast *CastResult
	// Redraw 為真表示這一句之後原版會重畫戰場再等一次鍵
	// （`sub_26384:0x265D5`，移動成功才有）。
	Redraw bool
	// Records 是這一手要印的幾筆記錄，給一句話要接兩筆的情形用（射擊撞在樹上
	// 或牆上，RE-102）。有 Records 就不看 Message。
	Records []int
	// Plain 為真表示這一句**前面不加動作者的名字**。撿那三句是這樣：問句
	// （0x111）自己就是一句完整的話，而 0x112／0x113 開頭就是 `|G`，加了會變成
	// 兩次名字。原版對它們另外起一段訊息（`sub_1C1A4`，`0x20354`／`0x20417`／
	// `0x204B9`），不接在動作者的名字後面（RE-113）。
	Plain bool
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
	// 「要不要撿」那一問答完之後還有一句要印（RE-113）：先把它交出去再取下一手，
	// 順序才與原版一樣（問句、回答、結果三句接在同一手後面）。
	if s.resume != nil {
		out := *s.resume
		s.resume = nil
		return out, true
	}
	for s.Exec.Turn < len(s.Exec.Order) {
		i := s.Exec.Order[s.Exec.Turn]
		s.Exec.Turn++
		a := &s.Actors[i]
		// 已經退出戰鬥或空掉的位置跳過（`0x1FF77`）。
		if a.Side == SideEmpty || a.Down() {
			continue
		}
		out := s.act(i, rng)
		// 一手做完的收尾（`sub_1E3A8:0x20E64`）：這一手挨的傷搬進 +0x2A，
		// 挨傷計數與招架次數歸零——兩格都是「這一手」的計數，不是整場的。
		a.LastHurt, a.Fatigue, a.Parries = a.Fatigue, 0, 0
		return out, true
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
		// 命中就再判一次擊退（`0x27B77`）。落空或被擋下不判。
		if !outcome.Miss && !outcome.Parried {
			out.Knock = s.Knockback(i, target, rng)
		}
		return out
	case CmdRest:
		out.Message = RestingRecord
		return out
	case CmdLand:
		// 下馬（`0x2070f`…`0x20774`，RE-124）：那一格上還有別人就印記錄 0x11A
		// 走人，沒人才把 +0x1E 清掉並印「lands!」。
		//
		// 原版同時把記錄 `+0x25` 也清成 0；remake 沒有那一格（它在哪裡被讀還沒解），
		// 所以這裡只清 Flag1E。
		if s.sharesCell(i) {
			out.Message = BlockedByRecord
			return out
		}
		a.Flag1E = 0
		out.Message = LandsRecord
		// 原版接 `sub_1C1BF` 重畫戰場再等一次鍵。
		out.Redraw = true
		return out
	case CmdFlinch:
		// 怯戰：防禦選項改成 2 再印那一句（`0x2062B`）。
		a.Guard = 2
		out.Message = FlinchRecord
		return out
	case 0:
		out.Message = NothingRecord
		return out
	case CmdMove, CmdMountedMove:
		return s.move(i)
	case CmdDrop:
		out.Message, out.Item = s.DropItem(i)
		return out
	case CmdPickUp:
		slot := s.GroundUnder(i, 0)
		if slot < 0 {
			// 腳下沒東西：原版這一手就這樣過去，不印任何訊息。
			return out
		}
		// 隊員要先問「要不要撿」，敵人直接撿（`0x20332` 比參戰者記錄 +0x01）。
		if a.Side == SideParty {
			return s.askPickUp(i, slot)
		}
		out.Message, out.Item = s.TakeItem(i, slot)
		return out
	case CmdSwitch, CmdSwitchAlt1, CmdSwitchAlt2:
		out.Message = s.SwitchItem(i)
		return out
	case CmdCast:
		return s.cast(i, rng)
	case CmdMissile, CmdLoad:
		// 兩個指令碼走同一支，與原版一樣（`0x20785` 兩條都跳 `0x20799`）：
		// 該裝填的自己會變成裝填那一手（RE-105）。
		return s.shoot(i, rng)
	}
	// 剩下的幾個指令碼。**撿（0x15）與丟（0x24）不會走到這裡**——執行段的分派只認
	// 十個指令碼，那兩個不在其中，它們是規劃段的即時動作（RE-093）。真的落到這裡
	// 表示規劃段漏接了，標成還沒實作，不假裝做過。
	out.Unimplemented = true
	return out
}

// cast 是施法那一手（`sub_2818D`，RE-100）。
//
// 目標由法術自己決定：`+0x0E` 的 bit 7 沒設就打施法者面對的那一格（`sub_24BB6`），
// 設了就打規劃段選好的那一格——`sub_24C35` 讀的是 `[0x1A74]`／`[0x1B36]`，
// 也就是 remake 的 `Plan.Memo`。
//
// 效果 6（指定部位）在原版要玩家挑一個部位，需要在執行段中間插一次互動；
// remake 目前沒有那個機制，所以照原版「要選目標格的法術」那一條走隨機部位
// （`0x2846F` 的 `rand(6) − 1`）。
func (s *State) cast(i int, rng Rand) ActResult {
	a := &s.Actors[i]
	out := ActResult{Actor: i, Command: a.Command, Target: -1}
	m := s.member(i)
	var spell Spell
	if m != nil && a.SpellPick >= 0 && a.SpellPick < len(m.Spells) {
		spell = m.Spells[a.SpellPick]
	}
	target := s.FacingTarget(i)
	if spell.NeedsTarget() {
		c := s.memo(i)
		target = s.actorAt(c.Row, c.Col)
	}
	if target < 0 {
		out.Message = NothingThereRecord
		return out
	}
	area := -1
	if spell.Effect == SpellEffectArea {
		area = rng.Roll(6) - 1
	}
	res := s.Cast(i, target, area, rng)
	out.Target = target
	switch res.Verdict {
	case SpellNothingThere:
		out.Message = NothingThereRecord
	case SpellNoEffect:
		out.Message = NoEffectRecord
	default:
		out.Message = CastsAtRecord
		out.Cast = &res
	}
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

// sharesCell 回報這一格上還有沒有別人站著（`sub_227C0`，RE-124）：
// 比的是座標，退出戰場的那些不算。指令列的 `CommandEnv.Occupied` 走同一套。
func (s *State) sharesCell(i int) bool {
	if i < 0 || i >= len(s.Actors) {
		return false
	}
	a := s.Actors[i]
	for n := range s.Actors {
		if n == i {
			continue
		}
		o := s.Actors[n]
		if o.Side != SideEmpty && !o.Down() && o.X == a.X && o.Y == a.Y {
			return true
		}
	}
	return false
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

// shoot 是射擊那一手（指令碼 0x22 與 0x23，`0x20799` ＋ `sub_26D76(0)`，
// RE-102／RE-105）。
//
// 沒裝填的武器第一回合只能裝填，這一手就沒了（`0x207CD`，RE-093）；射出去之後
// 那一格又清回沒裝填（`0x20860`）。指令列上那兩格就是這兩個階段：0x22 是還沒
// 上弦、0x23 是可以射（RE-083 的條件表）。
//
// 目標是規劃段記住的那一格（`Plan.Memo`），中間拉一條線：撞在樹或牆上就印
// 「fires a missle into a …」，被半路的人接下就打那個人，一路到底就打停下來
// 那一格上的人，那一格空著印「fires a missle but nothing is there!」。
func (s *State) shoot(i int, rng Rand) ActResult {
	a := &s.Actors[i]
	out := ActResult{Actor: i, Command: a.Command, Target: -1}
	reload := NeedsReload(a.Weapon.Kind)
	if reload && a.Flag1D == 0 {
		a.Flag1D = 1
		a.Command = 0
		out.Command = 0
		out.Message = ReloadRecord
		return out
	}
	if reload {
		a.Flag1D = 0
	}
	c := s.memo(i)
	shot := s.Shoot(i, c.Row, c.Col, a.AttackValue()/10+rng.Roll(AttackRoll))
	if shot.Stopped && shot.Block != MissileActor {
		out.Records = []int{MissileIntoRecord, TreeRecord}
		if shot.Block == MissileWall {
			out.Records[1] = WallRecord
		}
		return out
	}
	target := shot.Target
	if target < 0 {
		// 停下來那一格先找敵方（`sub_24C35(1, …)`），沒有再找任何一名
		// （`sub_24C35(0, …)`）——**兩次都只看敵方陣營那一格**，射的人是誰不影響。
		target = s.enemyAt(shot.X, shot.Y)
		if target < 0 {
			target = s.actorAt(shot.X, shot.Y)
		}
	}
	if target < 0 || target == i || s.Actors[target].Down() {
		out.Message = MissileNothingRecord
		return out
	}
	out.Target = target
	outcome := ResolveMissile(a, &s.Actors[target], a.Style,
		shot.Score, shot.Block == MissileActor, rng)
	out.Outcome = &outcome
	if !outcome.Miss && !outcome.Parried {
		out.Knock = s.Knockback(i, target, rng)
	}
	return out
}

// enemyAt 是 `sub_24C35(1, x, y)`：那一格上的敵方參戰者，沒有就回 -1。
// 旗標的 bit 0 表示「我方那一邊跳過」，所以問的人是誰不影響結果。
func (s *State) enemyAt(x, y int) int {
	for i := range s.Actors {
		a := s.Actors[i]
		if a.Side == SideEmpty || a.Side == SideParty || a.Down() {
			continue
		}
		if a.X == x && a.Y == y {
			return i
		}
	}
	return -1
}
