package battle

// 敵人那一手的抽樣器（RE-081／RE-092）。攻擊式與命中部位共用同一個加權輪盤
// `sub_25A3A`；三張基礎權重表由 RE-092 從 DGROUP 的靜態表解出來。

// AttackStyleBase 是武器攻擊式的基礎權重（DGROUP `0x1092`，10 筆）。
var AttackStyleBase = [10]int{6, 6, 5, 4, 6, 4, 4, 6, 6, 5}

// LocationWeights 是命中部位的基礎權重（DGROUP `0x10A6`，7 筆）：
// 頭、胸、右臂、左臂、右腿、左腿，最後一筆是第七個索引。
var LocationWeights = [7]int{1, 3, 4, 4, 4, 4, 2}

// MoveBase 是移動三式的基礎權重（DGROUP `0x10AE`）：步行、奔跑、衝刺。
// 比例約 1 : 2 : 6——原版的敵人偏好衝過來。
var MoveBase = [3]int{12, 24, 72}

// ContextWeights 是情境權重矩陣（DGROUP `0x0FF4`，16 列 × 10）。抽攻擊式時列號是
// 對手的防禦選項，抽部位時列號換成攻擊式。
var ContextWeights = [16][10]int{
	{8, 8, 8, 6, 8, 7, 7, 8, 9, 11},
	{4, 4, 3, 2, 4, 2, 2, 4, 4, 6},
	{7, 7, 6, 5, 7, 6, 6, 7, 8, 11},
	{4, 3, 3, 7, 9, 7, 3, 3, 4, 9},
	{2, 5, 5, 1, 6, 5, 3, 9, 10, 9},
	{6, 6, 6, 1, 3, 3, 4, 6, 8, 8},
	{10, 9, 3, 4, 6, 1, 4, 5, 2, 7},
	{6, 6, 6, 6, 6, 6, 6, 6, 6, 10},
	{1, 2, 3, 3, 4, 4, 3, 1, 3, 5},
	{7, 3, 1, 2, 1, 3, 5, 3, 3, 2},
	{2, 1, 3, 5, 4, 3, 5, 3, 1, 4},
	{3, 9, 8, 4, 4, 1, 3, 2, 3, 5},
	{3, 2, 1, 4, 2, 4, 6, 8, 3, 1},
	{4, 6, 6, 4, 4, 3, 1, 3, 5, 2},
	{3, 3, 2, 1, 2, 4, 1, 2, 7, 2},
	{2, 4, 6, 6, 6, 6, 3, 0, 6, 6},
}

// StyleMask 是攻擊式要的武器位元（DGROUP `0x109C`，10 筆）。與怪物記錄 +0x2D
// 有交集的那一式權重乘 2，沒有乘 1。
var StyleMask = [10]int{16, 16, 16, 8, 8, 8, 4, 4, 4, 2}

// levelDivisor 是 `w^等級` 的除數（DGROUP `0x10C6`）＝ 2^(等級−1)。
var levelDivisor = [8]int{1, 1, 2, 4, 8, 16, 32, 64}

// WeightedPick 是 `sub_25A3A`：權重抽樣，回傳選中的索引。
//
// 兩段權重——`context` 是情境權重、`base` 是基礎權重：
//
//	w = context[i]^等級 ÷ 2^(等級−1)      ; 等級越高越偏向權重高的選項
//	acc[i] = w × base[i]
//	正規化成總和 255，再擲 rand(256) − 1 逐項扣，扣成負的就選中
//
// 正規化那一步原版分兩條路避免 32-bit 溢位：總和超過 16 bit 就先把總和除以 255、
// 每一項直接除；否則每一項先乘 255 再除總和（`0x25BBF`…`0x25C92`）。兩條路的
// 語意相同，這裡照抄，因為溢位與否會影響取整的結果。
func WeightedPick(base, context []int, level int, rng Rand) int {
	n := len(base)
	if n == 0 || len(context) < n {
		return 0
	}
	if level < 0 {
		level = 0
	}
	if level >= len(levelDivisor) {
		level = len(levelDivisor) - 1
	}
	acc := make([]int64, n)
	for i := 0; i < n; i++ {
		w := int64(1)
		for k := 0; k < level; k++ {
			w *= int64(context[i])
		}
		w /= int64(levelDivisor[level])
		acc[i] = w * int64(base[i])
	}
	var total int64
	for _, v := range acc {
		total += v
	}
	if total == 0 {
		total = 1
	}
	wide := total > 0xFFFF
	if wide {
		total /= 255
		if total == 0 {
			total = 1
		}
	}
	for i := range acc {
		if !wide {
			acc[i] *= 255
		}
		acc[i] /= total
	}
	r := int64(rng.Roll(256) - 1)
	for i := range acc {
		r -= acc[i]
		if r < 0 {
			return i
		}
	}
	return 0
}

// EnemyStyle 是 `sub_2578F`：抽武器攻擊式。情境權重的列號是對手的防禦選項，
// 再乘上「這一式能不能用」——`StyleMask[i]` 與怪物的武器位元有交集算 2、沒有算 1。
func EnemyStyle(level, guard, weaponBits int, rng Rand) int {
	if guard < 0 || guard >= len(ContextWeights) {
		guard = 0
	}
	base := make([]int, len(AttackStyleBase))
	ctx := make([]int, len(AttackStyleBase))
	for i := range base {
		base[i] = AttackStyleBase[i]
		ctx[i] = ContextWeights[guard][i]
		if StyleMask[i]&weaponBits != 0 {
			ctx[i] *= 2
		}
	}
	return WeightedPick(base, ctx, level, rng)
}

// EnemyLocation 是 `sub_2590F`：抽命中部位。情境權重同一張矩陣，但列號換成攻擊式。
//
// **少了一個修正**：原版對索引 < 2 或 == 6 的三項另外乘 `word_2F940 ÷ word_30D06`
// （負重與行動餘力的比），那兩個全域的語意還沒解（RE-081），所以這裡不套。
func EnemyLocation(level, style int, rng Rand) int {
	if style < 0 || style >= len(ContextWeights) {
		style = 0
	}
	base := make([]int, len(LocationWeights))
	ctx := make([]int, len(LocationWeights))
	for i := range base {
		base[i] = LocationWeights[i]
		ctx[i] = ContextWeights[style][i]
	}
	return WeightedPick(base, ctx, level, rng)
}

// EnemyMoveForm 是移動三式的抽樣（`0x239D7` 之後）：步行 0、奔跑 1、衝刺 2。
// 三個基礎權重照狀況打折之後正規化，這裡只做腿還在不在那一條——其餘幾條要
// `word_30D06`／`word_2F940` 的語意才做得了。
func EnemyMoveForm(level, ring int, legsOK bool, hurt int, rng Rand) int {
	w := [3]int{MoveBase[0], MoveBase[1], MoveBase[2]}
	switch {
	case legsOK && ring < RingCount:
		// 兩腿都還在而且不是最外圈：衝刺這一項歸零（`0x23A3B`）。
		w[2] = 0
	case ring < 2:
		w[2] /= 2
	}
	if hurt > 49 {
		w[2] /= 2
	}
	base := []int{w[0], w[1], w[2]}
	return WeightedPick(base, base, level, rng)
}

// 敵人那一手的決策（RE-081 的主幹）。
//
// **這是簡化版，不是整棵樹。**已經照原版做的：兩臂都廢就移動、同心環由內往外找
// 目標而且起點隨機、貼上去就攻擊、有 1/3 機率放棄攻擊改移動、攻擊式與部位走原版
// 的加權輪盤、移動方式走三式的權重。
//
// 撤退那一段也照原版做了（`0x237F5`…`0x238C1`，RE-081）。
//
// **還沒做的**，都需要語意還沒解出來的旗標（RE-081 的「還沒走完的」那一節）：
// 反制掃描 `sub_253E7`、側翼 `sub_25653` 與沒有武器時的指令 0x26。這些情形目前
// 一律落到「攻擊或移動」。射擊那一條已經照原版做了（RE-104）。

// EnemyShootRing 是肯射擊的最遠環號。原版比的是 `word_349B2 < 4`（`0x23EBB`），
// 而那個計數從 1 起算相鄰的那一圈（`0x23DDE`）——remake 的環號從 0 起算同一圈，
// 所以門檻少一。環 0 在前面就被貼身攻擊接走了，實際會射的是環 1 與環 2。
const EnemyShootRing = 3

// EnemyDecision 是敵人這一回合要做什麼。
type EnemyDecision struct {
	Command int
	Style   int
	Guard   int
	Facing  int
	// Target 是選中的目標索引，沒有目標時是 -1；Ring 是在第幾環找到的。
	Target int
	Ring   int
}

// PlanEnemy 決定第 i 名敵人這一回合的一手，並寫進它的參戰者記錄。
//
// 原版把結果放進三個全域再由呼叫端寫回（`0x1F104`），remake 直接寫——兩邊的可觀察
// 行為相同，中間那三個全域沒有別的讀者。
//
// 主幹（`0x23C1A`…`0x24030`）：
//
//	環 0（八個相鄰格）找隊員 → 找到就攻擊，但有 1/3 機率改成移動
//	環 0 沒人 → 往外掃環 1…6
//	    線通不到就換一個目標（`0x23EAB` 的 `sub_25D0F`）
//	    環號 < 4、種類非零又有右臂 → 射擊（`0x23F15`，RE-104）
//	    否則 → 朝目標走一步
//	七環都沒人 → 隨機挑一個站得下的方向移動
func (s *State) PlanEnemy(i int, rng Rand) EnemyDecision {
	if i < 0 || i >= len(s.Actors) {
		return EnemyDecision{Target: -1}
	}
	a := &s.Actors[i]
	d := EnemyDecision{Target: -1, Ring: -1, Facing: a.Facing}

	// 兩隻手臂都廢就只能移動（`0x237E9`）。
	if a.RightArm <= 0 && a.LeftArm <= 0 {
		return s.enemyMove(i, -1, -1, rng)
	}

	// 撤退（`0x237F5`…`0x238C1`）。傷勢是 `50 × 累積傷害 ÷ 昏迷門檻`，
	// 兩個門檻來自怪物記錄的 +0x0D／+0x0E——不死系是 100／0，永遠走不進來。
	if d, ok := s.enemyRetreat(i, rng); ok {
		return d
	}

	// 環 0：貼著的隊員直接打。種類非零的怪物有 1/3 機率放棄攻擊改成移動
	// （`0x23CA7`），所以敵人不會每回合都撲上來。
	if t := s.findInRing(i, 0, rng); t >= 0 {
		if rng.Roll(3) == 1 {
			return s.enemyMove(i, t, 0, rng)
		}
		d.Command = CmdWeapon
		d.Target = t
		d.Ring = 0
		d.Facing = s.stepToward(i, t)
		s.enemyStrike(&d, i, t, EnemyNearStyle, rng)
		s.applyEnemy(i, d)
		return d
	}

	// 環 1…6：找得到就先看線通不通，通得到才考慮射擊，否則靠近。
	foot := TileGround
	if s.Field != nil {
		foot = s.Field.At(a.X, a.Y)
	}
	if a.Flag1E != 0 {
		foot = MountedSight
	}
	for ring := 1; ring < RingCount; ring++ {
		t := s.findInRing(i, ring, rng)
		if t < 0 {
			continue
		}
		// 線通不到就跳過。**原版是回頭在同一環繼續找下一個**
		// （`0x23EB8` 跳回 `0x23DEC`），這裡換下一環——`findInRing` 一次只回
		// 一名，兩者只在同一環有多名目標而近的那名被擋住時不同。
		if s.Field != nil &&
			!s.Field.LineOfSight(a.X, a.Y, s.Actors[t].X, s.Actors[t].Y, foot) {
			continue
		}
		if ring < EnemyShootRing && a.Kind != 0 && a.RightArm > 0 {
			return s.enemyShoot(i, t, ring)
		}
		return s.enemyMove(i, t, ring, rng)
	}
	return s.enemyMove(i, -1, -1, rng)
}

// EnemyNearStyle／EnemyNearGuard 是環 0 找到目標時先填進去的攻擊選項與防禦選項
// （`0x23C81`）。攻擊式抽完會蓋掉前者，而**前者的舊值就是抽命中部位的列號**。
const (
	EnemyNearStyle = 0x0B
	EnemyNearGuard = 7
)

// enemyStrike 填貼近打那一手的攻擊選項與防禦選項（`0x23CD6`…`0x23D8B`）。
//
// `prev` 是進來時的攻擊選項：環 0 那條是 `EnemyNearStyle`，反制掃描那條是從對手
// 抄來的（還沒實作）。**`+0x24` 這一格兩種用途**——填進去的是「命中部位 ＋1」，
// 而同一格在這一名被攻擊時就是它的招架姿態（RE-081／RE-105）。
//
// **與原版的一個差別**：原版在移動段繞完八個方向都站不下時會把 `word_34A08` 加一，
// 之後攻擊式就固定是 1；這一版沒有那條重試路徑，所以攻擊式一律抽。
func (s *State) enemyStrike(d *EnemyDecision, i, target, prev int, rng Rand) {
	a := &s.Actors[i]
	d.Guard = EnemyNearGuard
	// 攻擊選項是 `sub_2578F()` 的結果加二；超過 10 就退回 1（`0x23CEE`）。
	style := EnemyStyle(a.Level, s.Actors[target].Guard, a.Weapon.Kind, rng) + 2
	if style > 0x0A {
		style = 1
	}
	d.Style = style
	switch {
	case style == 1:
		d.Guard = 2
	case prev == 0:
		// 進來時沒有攻擊選項就擲一次決定招架姿態（`0x23D2F`）。
		if rng.Roll(2) == 1 {
			d.Guard = 0
		} else {
			d.Guard = 2
		}
	default:
		// 列號是**進來時**那個值，不是剛抽出來的攻擊式（`0x23D58`）。
		loc := EnemyLocation(a.Level, prev, rng)
		switch {
		case loc == 0:
			d.Guard = 0
		case loc+1 > 6:
			d.Style, d.Guard = 0, 1
		default:
			d.Guard = loc + 1
		}
	}
}

// enemyRetreat 是傷得太重那一段（`0x237F5`…`0x238C1`）。第二個回傳值為 false
// 表示這一手不走撤退，交回主幹繼續判。
//
//	已在撤退而且傷勢降到 FleeRecover 以下 → 清掉旗標，回正常判斷
//	已在撤退，或傷勢達到 FleeAt →
//	    八個方向有隊員貼著 → 往外逃（隨機挑一個站得下的方向）
//	    八個方向都沒人 →
//	        還沒進入撤退狀態，或上一手沒挨打 → 原地喘息，並記下撤退狀態
//	        已在撤退而且上一手挨了打 → 繼續逃
func (s *State) enemyRetreat(i int, rng Rand) (EnemyDecision, bool) {
	a := &s.Actors[i]
	hurt := a.Injury()
	if a.Retreating && hurt < a.FleeRecover {
		a.Retreating = false
		return EnemyDecision{}, false
	}
	if !a.Retreating && hurt < a.FleeAt {
		return EnemyDecision{}, false
	}
	if s.adjacentParty(i) {
		return s.enemyMove(i, -1, -1, rng), true
	}
	if !a.Retreating || a.LastHurt == 0 {
		a.Retreating = true
		d := EnemyDecision{Command: CmdRest, Target: -1, Ring: -1, Facing: a.Facing}
		s.applyEnemy(i, d)
		return d, true
	}
	return s.enemyMove(i, -1, -1, rng), true
}

// adjacentParty 回報八個相鄰格有沒有站著的隊員（`sub_24BB6(0, 方向, 0x19)`）。
// 撤退那一段是按方向編號順序掃的，不像找目標那樣從隨機的一格起跳。
func (s *State) adjacentParty(i int) bool {
	a := s.Actors[i]
	for _, c := range RingCells(0) {
		if t := s.actorAt(a.X+c.Row, a.Y+c.Col); t >= 0 &&
			s.Actors[t].Side == SideParty && !s.Actors[t].Down() {
			return true
		}
	}
	return false
}

// enemyShoot 組出射擊那一手（`0x23F15`…`0x23F53`）：把目標那一格記進 Memo、
// 指令碼設 0x23、攻擊選項 0，並且**直接把 +0x1D 設成 1**——敵人不必先花一回合
// 裝填，原版就是這樣寫的。
func (s *State) enemyShoot(i, target, ring int) EnemyDecision {
	a := &s.Actors[i]
	d := EnemyDecision{Command: CmdMissile, Target: target, Ring: ring,
		Facing: s.stepToward(i, target)}
	a.Flag1D = 1
	if s.Plan.Memo == nil {
		s.Plan.Memo = map[int]Cell{}
	}
	s.Plan.Memo[i] = Cell{Row: s.Actors[target].X, Col: s.Actors[target].Y}
	s.applyEnemy(i, d)
	return d
}

// enemyMove 組出移動那一手：方向朝目標（沒有目標就隨機挑一個站得下的），
// 移動方式走三式的加權輪盤（`0x239D7`）。
func (s *State) enemyMove(i, target, ring int, rng Rand) EnemyDecision {
	a := &s.Actors[i]
	d := EnemyDecision{Command: CmdMove, Target: target, Ring: ring}
	d.Facing = s.stepToward(i, target)
	legsOK := a.RightLeg > 0 && a.LeftLeg > 0
	form := EnemyMoveForm(a.Level, max(ring, 0), legsOK, a.Injury(), rng)
	d.Style = IconMoveWalk + form
	if form == 0 {
		d.Guard = 2
	}
	s.applyEnemy(i, d)
	return d
}

func (s *State) applyEnemy(i int, d EnemyDecision) {
	a := &s.Actors[i]
	a.Command, a.Style, a.Guard = d.Command, d.Style, d.Guard
	if d.Facing >= 0 {
		a.Facing = d.Facing
	}
}

// findInRing 在指定的環上找一名還站著的隊員，回傳索引；找不到回 -1。
//
// 起點是隨機的（`sub_2531C` 第一次傳 0xFFFF 時從 `rand(格數 − 1) − 1` 起跳），
// 所以同樣的戰況不會總是選同一個目標。
func (s *State) findInRing(i, ring int, rng Rand) int {
	a := s.Actors[i]
	cells := RingCells(ring)
	if len(cells) == 0 {
		return -1
	}
	start := rng.Roll(len(cells)) - 1
	for n := 0; n < len(cells); n++ {
		c := cells[(start+n)%len(cells)]
		if t := s.actorAt(a.X+c.Row, a.Y+c.Col); t >= 0 &&
			s.Actors[t].Side == SideParty && !s.Actors[t].Down() {
			return t
		}
	}
	return -1
}

// actorAt 回傳站在那一格上的參戰者索引；沒有人就回 -1。
func (s *State) actorAt(x, y int) int {
	for i := range s.Actors {
		a := s.Actors[i]
		if a.Side != SideEmpty && a.X == x && a.Y == y {
			return i
		}
	}
	return -1
}

// stepToward 是「往目標走一步」的方向。原版先用 `sub_25D7A(1)` 沿視線走一步再查
// `[0x10D8]` 把相對格差換成方向（RE-081）；這裡直接取符號，兩者在相鄰與正對角的
// 情形相同，斜向偏移大時可能差一格。目標是 -1 就隨機挑一個站得下的方向。
func (s *State) stepToward(i, target int) int {
	a := s.Actors[i]
	if target < 0 {
		for f := 0; f < 8; f++ {
			c := RingCells(0)[f]
			if !s.moveBlocked(i, a.X+c.Row, a.Y+c.Col) {
				return f
			}
		}
		return a.Facing
	}
	t := s.Actors[target]
	dx, dy := sign(t.X-a.X), sign(t.Y-a.Y)
	for f, c := range RingCells(0) {
		if c.Row == dx && c.Col == dy {
			return f
		}
	}
	return a.Facing
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
