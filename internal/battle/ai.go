package battle

// 敵人那一手的抽樣器（RE-081／RE-092／RE-159）。攻擊式與命中部位共用同一個加權輪盤
// `sub_25A3A`；三張基礎權重表在戰鬥初始化時填好，見 `AIWeights`。

// AttackStyleBase 是武器攻擊式的預設基礎權重（DGROUP `0x1092`，10 筆）。
var AttackStyleBase = [10]int{6, 6, 5, 4, 6, 4, 4, 6, 6, 5}

// LocationWeights 是命中部位的基礎權重（DGROUP `0x10A6`，7 筆）：
// 頭、胸、右臂、左臂、右腿、左腿，最後一筆是第七個索引。
var LocationWeights = [7]int{1, 3, 4, 4, 4, 4, 2}

// MoveBase 是移動三式的預設基礎權重（DGROUP `0x10AE`）：步行、奔跑、衝刺。
var MoveBase = [3]int{12, 24, 72}

// AIWeights 是這一場戰鬥敵人 AI 用的三張基礎權重表（DGROUP `0x6AD2`、`0x1A9A`、
// `0x6A9A`）。`sub_1E3A8:0x1EC49` 看怪物記錄 `+0x1A0` 那一格：非零就從記錄
// `+0x1A0`／`+0x1AA`／`+0x1B1` 抄三張表，零就抄上面三張預設表（RE-159）。
// 旗標那一格同時是攻擊式權重的第 0 筆。
type AIWeights struct {
	Style    [10]int
	Location [7]int
	Move     [3]int
}

// DefaultAIWeights 是記錄 `+0x1A0` 為零時的三張預設表。
var DefaultAIWeights = AIWeights{Style: AttackStyleBase, Location: LocationWeights, Move: MoveBase}

// ContextWeights 是情境權重矩陣（DGROUP `0x0FF4`，16 列 × 10）。抽攻擊式時列號是
// `word_30916`（見 `enemyStrike`），抽部位時列號換成攻擊式。
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

// StyleMask 是攻擊式要的武器位元（DGROUP `0x109C` 的前 8 筆，`sub_2578F` 只讀這 8 筆）。
// 與怪物記錄 +0x2D 有交集的那一式權重乘 2，沒有乘 1；第 8、9 式見 `EnemyStyle`。
var StyleMask = [8]int{16, 16, 16, 8, 8, 8, 4, 4}

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
// RE-237：每次乘、加、減保留32位元，有號除法向零截斷。
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
	acc := make([]int32, n)
	for i := 0; i < n; i++ {
		w := int32(1)
		for k := 0; k < level; k++ {
			w *= int32(context[i])
		}
		w /= int32(levelDivisor[level])
		acc[i] = w * int32(base[i])
	}
	var total int32
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
	r := int32(rng.Roll(256) - 1)
	for i := range acc {
		r -= acc[i]
		if r < 0 {
			return i
		}
	}
	return 0
}

// EnemyStyle 是 `sub_2578F`：抽武器攻擊式。情境權重的列號是 `word_30916`（`guard`，
// 呼叫端決定，見 `enemyStrike`），再乘上「這一式能不能用」——前 8 筆
// `StyleMask[i]` 與怪物記錄 `+0x2D` 有交集算 2、沒有算 1。乘積存成一個 byte。
//
// **遮罩只寫前 8 筆**（`0x25841`：`< 8`），第 8、9 筆讀到的是共用緩衝 `byte_37CF8`
// +0x28／+0x29 當時的殘值——旅店名冊清單 `sub_147DD:0x149DD` 寫在那裡的索引（RE-159）。
// `tail` 就是這兩格（`State.StyleMaskTail`）。是 0 時那兩式的權重歸零，永遠抽不到。
//
// 最後一項（索引 9）另外乘「額度 ÷ 花費」，花費是 `sub_251B0(15, 1)`
// （`0x2589E`，RE-131）。比值夾在 1/3…1，所以那一項只會被壓低。
func EnemyStyle(a Actor, w AIWeights, guard, weaponBits int, tail [2]int, rng Rand) int {
	if guard < 0 || guard >= len(ContextWeights) {
		guard = 0
	}
	base := make([]int, len(w.Style))
	ctx := make([]int, len(w.Style))
	for i := range base {
		base[i] = w.Style[i]
		var mask int
		switch {
		case i >= len(StyleMask):
			mask = tail[i-len(StyleMask)]
		case StyleMask[i]&weaponBits != 0:
			mask = 2
		default:
			mask = 1
		}
		ctx[i] = ContextWeights[guard][i] * mask & 0xFF
	}
	num, den := ClampEffort(a.TimeAllowance(), ActionCost(a, StyleEffortTime, 1))
	if last := len(ctx) - 1; last >= 0 && den != 0 {
		ctx[last] = ctx[last] * num / den
	}
	return WeightedPick(base, ctx, a.Level, rng)
}

// EnemyLocation 是 `sub_2590F`：抽命中部位。情境權重同一張矩陣，但列號換成攻擊式。
//
// 索引 < 2 與 == 6 那三項另外乘「花費 ÷ 額度」，花費是
// `sub_251B0([0x10B2 + 攻擊選項], 2)`（`0x25994`，RE-131）。比值夾在 1…3，
// 所以動作越貴、額度越少，這三個部位越被偏向。
//
// **額度用的是這一手完整的那一份**（`0x2596C` 先把 `word_30D06` 設回
// `word_2FBB6`），不是花剩的。

// RE-237：原版高移動選項的七格位元組，見 SPEC-008 第8節勘誤。
var moveLocationContext = [3][7]int{
	{3, 4, 5, 0, 0, 0, 0},
	{2, 3, 2, 3, 4, 0, 96},
	{91, 96, 110, 116, 78, 97, 91},
}

func EnemyLocation(a Actor, w AIWeights, style int, rng Rand) int {
	var context []int
	if style >= IconMoveWalk && style <= IconMoveFast {
		context = moveLocationContext[style-IconMoveWalk][:]
	} else {
		if style < 0 || style >= len(ContextWeights) {
			style = 0
		}
		context = ContextWeights[style][:]
	}
	base := make([]int, len(w.Location))
	ctx := make([]int, len(w.Location))
	for i := range base {
		base[i] = w.Location[i]
		ctx[i] = context[i]
	}
	// 矩陣列用傳入的舊選項，花費卻查行動者寫回之前的 +0x23
	// （sub_2590F:0x25984..0x25994，RE-168）。
	time := 0
	if a.Style >= 0 && a.Style < len(commandTime) {
		time = commandTime[a.Style]
	} else if a.Style >= IconMoveWalk && a.Style <= IconMoveFast {
		// 原版查到 16 格表外：移動三式 60、61、62 讀出 0、0、250。
		time = [3]int{0, 0, 250}[a.Style-IconMoveWalk]
	}
	cost := ActionCost(a, time, 2)
	num, den := ClampEffort(a.TimeAllowance(), cost)
	if num != 0 {
		for _, i := range LocationEffortSlots {
			if i < len(ctx) {
				ctx[i] = ctx[i] * den / num & 0xFF
			}
		}
	}
	return WeightedPick(base, ctx, a.Level, rng)
}

// StyleEffortTime 是 `sub_2578F` 傳給 `sub_251B0` 的第一個引數（`0x2589A`）。
const StyleEffortTime = 0x0F

// LocationEffortSlots 是要套那個比例的三個部位索引（`0x259A9`：`< 2` 或 `== 6`）。
var LocationEffortSlots = [3]int{0, 1, 6}

// MoveFormEffortTime 是移動三式那一段傳給 `sub_251B0` 的引數（`0x23A81`：6, 0）。
const MoveFormEffortTime = 6

// EnemyMoveForm 是移動三式的抽樣（`0x239D7`…`0x23BB9`，RE-159）：步行 0、奔跑 1、衝刺 2。
//
// 基礎權重（`AIWeights.Move`）只調衝刺那一項：兩腿都還在（`< 0x80`）而且環號 < 7 就歸零；
// 否則環號 < 2 減半、疲勞比 > 0x31 減半，再乘兩次「額度 ÷ 花費」（花費
// `sub_251B0` 原始推入 6、0，實際花費是負重乘 6（RE-168），
// 比值由 `sub_2526C` 夾在 1/3…1）。乘除都在 16 位元裡做。
//
// 正規化只做索引 2 與 1（迴圈從 2 往下、到 0 就停），索引 0 維持原值；抽樣擲
// `rand(256) − 1` 從索引 0 起逐項扣，扣成負的就選中。沒有等級加權。
func EnemyMoveForm(a Actor, weights AIWeights, ring int, rng Rand) int {
	w := weights.Move
	if !partLost(a.RightLeg) && !partLost(a.LeftLeg) && ring < RingCount {
		w[2] = 0
	} else {
		if ring < 2 {
			w[2] /= 2
		}
		if a.FatigueRatio() > 0x31 {
			w[2] /= 2
		}
		num, den := ClampEffort(a.TimeAllowance(), ActionCost(a, 0, MoveFormEffortTime))
		if den != 0 {
			v := (w[2] * num & 0xFFFF) / den
			v = (v * num & 0xFFFF) / den
			w[2] = v & 0xFF
		}
	}
	sum := w[0] + w[1] + w[2]
	if sum == 0 {
		sum = 1
	}
	for i := 2; i > 0; i-- {
		w[i] = w[i] * 0xFF / sum & 0xFF
	}
	r := rng.Roll(256) - 1
	i := 0
	for ; i < 2; i++ {
		r -= w[i]
		if r < 0 {
			break
		}
	}
	return i
}

// 敵人那一手的決策（RE-081 的主幹）。
//
// 照原版做的：兩臂都廢就移動、同心環由內往外找目標而且起點隨機、貼上去就攻擊、
// `+0x00` 非零的怪物有 1/3 機率放棄攻擊改移動、攻擊式與部位走原版的加權輪盤、
// 移動方式走三式的權重（RE-159 以正常遭遇五個回合同狀態對拍）。
//
// 撤退（`0x237F5`…`0x238C1`）、反制掃描（`sub_253E7`）、側翼（`sub_25653`）與
// 空手去撿（`0x23BC2`，RE-116）也都照原版做了。
//
// 找目標的篩選遮罩 0x10（`sub_24C35`）：這一名的智力擲骰過關時，`+0x05` 非零的
// 隊員看不到。`+0x05` 由穿戴的 Shade Ring 設（RE-154）；`hiddenFrom`。
//
// 移動段的方向（RE-117）：先試往目標那一步、再試左右各一格，沒有目標時從隨機方向
// 起繞八格（`stepToward`），八格都站不下走三條退路（`enemyStuck`）。

// EnemyShootRing 是肯射擊的最遠環號。原版比的是 `word_349B2 < 4`（`0x23EBB`），
// 原版與 remake 的環號相同：0x23C1A 先設零搜尋鄰圈，0x23DDE 才從外圈 1 起。
// 環 0 在前面被貼身攻擊接走，實際會射的是環 1、2、3（SPEC-013 訂正）。
const EnemyShootRing = 4

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
// 行為相同。決策樹照 `sub_2374C` 的標籤逐段翻（RE-168，擾動對拍逐項核過）：
//
//	0x237C7  兩臂都廢            → 逃（`enemyFlee`）
//	0x237F5  傷勢比 FleeRecover／FleeAt 低 → 正常決策（`planTactics`）
//	0x23844  八個相鄰格有隊員貼著 → 傷勢達 FleeAt 逃，否則正常決策
//	0x23886  沒人貼著            → 在撤退而且上一手挨打才逃，否則原地喘息
func (s *State) PlanEnemy(i int, rng Rand) EnemyDecision {
	if i < 0 || i >= len(s.Actors) {
		return EnemyDecision{Target: -1}
	}
	a := &s.Actors[i]
	// `word_34A08` 是一次規劃之內的計數，`sub_2374C` 開頭歸零（RE-117）。
	a.Stuck = 0
	// AI 一進來就擲一次（`0x237A6` → `word_3AC5E`），在手臂與撤退判斷之前，不論後面用不用
	// 得到。找目標的篩選遮罩 0x10（Shade Ring）用的是這一次（`hiddenFrom`，RE-159）。
	s.aiRoll = rng.Roll(100)
	ratio := a.FatigueRatio()

	// 兩隻手臂都廢（`0x237B0`：兩臂記錄值都 ≥ 0x80）。
	if partLost(a.RightArm) && partLost(a.LeftArm) {
		return s.enemyFlee(i, rng)
	}
	if a.Retreating {
		if ratio < a.FleeRecover {
			a.Retreating = false
			return s.planTactics(i, rng)
		}
	} else if ratio < a.FleeAt {
		return s.planTactics(i, rng)
	}
	for _, c := range RingCells(0) {
		if s.candidateAt(i, a.X+c.Row, a.Y+c.Col, maskMelee) >= 0 {
			if ratio >= a.FleeAt {
				return s.enemyFlee(i, rng)
			}
			return s.planTactics(i, rng)
		}
	}
	return s.enemyCornered(i, rng)
}

// enemyCornered 是 `0x23886`：沒有路可走時，已在撤退而且上一手挨了打就繼續逃，
// 否則原地喘息並記下撤退狀態。
func (s *State) enemyCornered(i int, rng Rand) EnemyDecision {
	a := &s.Actors[i]
	if a.Retreating && a.LastHurt != 0 {
		return s.enemyFlee(i, rng)
	}
	return s.enemyRest(i)
}

// enemyRest 是 `0x238C1`：指令碼 5、兩個選項 0，並把撤退旗標設起來。
func (s *State) enemyRest(i int) EnemyDecision {
	a := &s.Actors[i]
	a.Retreating = true
	d := EnemyDecision{Command: CmdRest, Target: -1, Ring: -1, Facing: a.Facing}
	s.applyEnemy(i, d)
	return d
}

// planTactics 是決策樹從「手上有沒有武器」那個分岔起的整段（`0x23BC2`…`0x24030`）。
// 繞不出去的退路會**回到這裡重跑**（RE-117）。
func (s *State) planTactics(i int, rng Rand) EnemyDecision {
	a := &s.Actors[i]
	d := EnemyDecision{Target: -1, Ring: -1, Facing: a.Facing}

	// 手上那一格空著（`0x23BC2`，RE-116）：先看參戰者記錄 +0x2B，非零就清掉改成
	// 隨機走一步（`0x240B9`），否則出指令碼 0x26 去撿地上的東西，攻擊選項 0、防禦選項 2。
	// **原版沒有任何地方把 +0x2B 設成非零**（兩個寫入點都寫 0），所以移動那一條
	// 走不到；照著實作是為了讓這一段與原版逐條對得起來。
	if a.Unarmed {
		if a.PickedUp != 0 {
			a.PickedUp = 0
			return s.enemyRandom(i, rng)
		}
		d.Command = CmdPickUp
		d.Style = 0
		d.Guard = 2
		s.applyEnemy(i, d)
		return d
	}

	// 反制掃描（`0x23C1A`）：另擲一次 rand(100) 做智力檢定，過關才跑。有結論就直接走
	// 攻擊那一段，不再往外環找目標——抄來的攻擊選項會變成抽命中部位的列號。
	if EnemyIntellectCheck(a.Intellect, a.Level, rng.Roll(100)) {
		if prev, row, t, ok := s.counterScan(i, rng); ok {
			s.aiStyle, s.aiGuard = prev, row
			return s.enemyAttack(i, t, prev, row, rng)
		}
	}

	// 環 0：貼著的隊員直接打，方向就是那一格的方向，攻擊選項先填 0x0B、防禦選項 7。
	if t, cell := s.findInRingCell(i, 0, maskMelee, rng, nil); t >= 0 {
		a.Facing = cell
		s.aiStyle, s.aiGuard = EnemyNearStyle, EnemyNearGuard
		return s.enemyAttack(i, t, EnemyNearStyle, EnemyNearGuard, rng)
	}

	// 環 1…6（`0x23D94`…`0x23E4C`，RE-168）。篩選遮罩看種類（`+0x00`）：
	//
	//	種類 0    遮罩 0x19，只打站在地上的隊員，掃一趟
	//	種類非零  先擲 `rand(2)`：1 是遮罩 0x11（隊員，騎乘的也算）掃一趟；
	//	          2 先掃一趟遮罩 0x31（只找正在施法或射擊的隊員），找不到再掃一趟 0x11
	//
	// 找得到就先看線通不通，通得到才考慮射擊，否則靠近。
	foot := TileGround
	if s.Field != nil {
		foot = s.Field.At(a.X, a.Y)
	}
	if a.Flag1E != 0 {
		foot = MountedSight
	}
	mask, passes := maskMelee, 1
	if a.Kind != 0 {
		if rng.Roll(2) == 1 {
			mask = maskAny
		} else {
			mask, passes = maskShooter, 2
		}
	}
	for ; passes > 0; passes-- {
		for ring := 1; ring < RingCount; ring++ {
			t := s.findInRingMatching(i, ring, mask, rng, func(target int) bool {
				return s.Field == nil || s.Field.LineOfSight(a.X, a.Y, s.Actors[target].X, s.Actors[target].Y, foot)
			})
			if t < 0 {
				continue
			}
			if ring < EnemyShootRing && a.Kind != 0 && !partLost(a.RightArm) {
				return s.enemyShoot(i, t, ring)
			}
			return s.enemyApproach(i, t, ring, rng)
		}
		mask = maskAny
	}
	return s.enemyRandom(i, rng)
}

// enemyAttack 是 `0x23CA7`…`0x23D8B`：找到貼身的目標之後，種類非零的有 1/3 機率改成逃
// （`0x23CD3` 跳的是 `0x238E8`，不是朝目標走），否則填貼近打那一手。
func (s *State) enemyAttack(i, target, prev, row int, rng Rand) EnemyDecision {
	a := &s.Actors[i]
	if a.Kind != 0 && rng.Roll(3) == 1 {
		return s.enemyFlee(i, rng)
	}
	d := EnemyDecision{Command: CmdWeapon, Target: target, Ring: 0, Facing: a.Facing}
	s.enemyStrike(&d, i, target, prev, row, rng)
	s.applyEnemy(i, d)
	return d
}

// EnemyIntellectCheck 是 `sub_252CB`：`值 × 等級 ÷ 3 > 門檻`。反制掃描前用它擲一次
// （門檻是 `rand(100)`，`0x23C1A`）——等級越高、智力越高越容易過，這是原版把
// 「聰明的怪物」做出來的方式。
func EnemyIntellectCheck(value, level, roll int) bool {
	return value*level/3 > roll
}

// planned 回報那一名這一回合規劃過了沒有（記錄 `+0x22`）。原版在規劃段開始時把
// 每一名清 0、決定完一手設 1（`0x1F03D`／`0x1FAC4`）；這一版用規劃順序的位置判斷
// ——排在目前這一手之前的就是已經決定過的。
func (s *State) planned(i int) bool {
	for n, idx := range s.Round.Order {
		if idx == i {
			return n < s.Round.Turn
		}
	}
	return false
}

// counterScan 是 `sub_253E7`：繞著自己那一環，對每個找到的隊員算一個優先度，
// 取最高的那一組攻防選項抄過來。回傳的 style 會變成 `enemyStrike` 的 `prev`，
// 也就是抽命中部位的列號。
//
//	4  對方還沒規劃 → 0x0B；對方指令是武器攻擊而且擲中就改抄它的
//	5  對方已規劃、要用武器攻擊、而且方向對著我 → 抄對方的 +0x23
//	3  同上但不是武器攻擊（徒手）→ 0x0B
//
// 「已規劃」那兩列**接著一定會落到共同出口再算一次**（`0x2553C`）：對方要移動給
// 1、有防禦選項給 2、其餘給 4，攻擊選項一律 0，而且算出來的分數 `>=` 目前最高分
// 才覆蓋。所以 5 永遠留得住（共同出口最高只到 4），3 只在共同出口算出 1 或 2
// 時留得住。最高優先度是 1 才繞側面。
//
// **抄來的防禦選項留不到最後**：接下來一定走貼近打那三個選項（`0x23D08`）而那裡
// 無條件重設 `word_30916`，所以這一支只回攻擊選項那一格。
//
// 「方向對著我」的判準是 `對方的 +0x28 XOR 我看到他的方向 == 4`——方向編號差 4
// 就是正對面。每找到一個有結論的目標就把自己的方向轉過去（`0x2562A`）。
func (s *State) counterScan(i int, rng Rand) (style, row, target int, ok bool) {
	a := &s.Actors[i]
	cells := RingCells(0)
	best, bestStyle, last := 0, 0, -1
	// bestRow 是跟著攻擊選項一起記下的情境列號（`var_8`，出口寫進 `word_30916`）：
	// 還沒規劃那一條是 7，其餘都是對方的防禦選項 `+0x24`（RE-159）。
	bestRow := 0
	// 起點隨機，與 `sub_2531C` 第一次傳 0xFFFF 時的行為相同。
	start := rng.Roll(len(cells)-1) - 1
	for k := 0; k < len(cells); k++ {
		n := (start + k) % len(cells)
		c := cells[n]
		t := s.candidateAt(i, a.X+c.Row, a.Y+c.Col, maskMelee)
		if t < 0 {
			continue
		}
		d := &s.Actors[t]
		pri := 0
		if !s.planned(t) {
			// 還沒規劃（`0x255A6`）。
			if best < 4 {
				best, bestStyle, bestRow, pri = 4, EnemyNearStyle, EnemyNearGuard, 4
				if d.Command == CmdWeapon && rng.Roll(2) == 1 {
					bestStyle, bestRow = d.Style, d.Guard
				}
			}
		} else {
			if d.Command == CmdWeapon || d.Command == CmdUnarmed {
				if d.Facing^n == 4 {
					if d.Command == CmdWeapon {
						best, bestStyle, bestRow = 5, d.Style, d.Guard
					}
					if best < 3 {
						best, bestStyle, bestRow = 3, EnemyNearStyle, d.Guard
					}
				}
			}
			// 這幾條每一次都會重算優先度（`0x2553C` 是共同出口），
			// 所以上面設的 5 或 3 只有在這裡算出來的比不過時才留得住。
			switch {
			case d.Command == CmdMove:
				pri = 1
			case d.Guard != 0:
				pri = 2
			default:
				pri = 4
			}
			if pri >= best {
				bestStyle, bestRow, best = 0, d.Guard, pri
			}
		}
		if pri != 0 {
			last = t
			a.Facing = n
		}
	}
	if best == 0 || last < 0 {
		return 0, 0, -1, false
	}
	// 優先度 1（對方要移動）才繞側面（`0x25435`）。
	if best == 1 {
		s.flank(i, last)
	}
	return bestStyle, bestRow, last, true
}

// flank 是 `sub_25653`：從自己現在的方向出發，依方向編號的奇偶決定要掃幾格
// （奇數走 2 格、掃到第 5 格；偶數走 1 格、掃到第 6 格），找出「自己站在對方的
// 哪一側」再把方向偏過去，繞到目標側面。對方沒有攻擊選項就不繞。
//
// **只改方向，不改位置**——移動是之後執行段的事。原版把偏移後的方向夾在
// `> 8` 而不是 `> 7`，所以偏出 8 時它會落在界外；這一版照樣算，越界就放棄。
func (s *State) flank(i, target int) {
	a := &s.Actors[i]
	d := &s.Actors[target]
	if d.Style == 0 {
		return
	}
	face, mine := d.Facing, a.Facing
	span, limit := 1, 6
	if mine&1 != 0 {
		span, limit = 2, 5
	}
	for k := span; ; {
		k++
		v := k + mine
		if v > 7 {
			v -= 8
		}
		if v == face {
			if k == 4 {
				return
			}
			if k < 4 {
				face -= span
				if face < 0 {
					face += 8
				}
			} else {
				face += span
				if face > 8 {
					face -= 8
				}
			}
			if s.flankBlocked(i, face) {
				return
			}
			a.Facing = face
			return
		}
		if k == limit {
			return
		}
	}
}

// flankBlocked 判側翼要轉過去的那一格：磚塊擋不擋路（`sub_24E2B`），以及那一格
// 有沒有**自己人**（`sub_24BB6(0, 方向, 0x1A)` 的遮罩 0x02 是「+0x01 非零」，
// 對敵人來說就是同類）。方向落在 0…7 之外一律當成擋住。
func (s *State) flankBlocked(i, dir int) bool {
	cells := RingCells(0)
	if dir < 0 || dir >= len(cells) {
		return true
	}
	a := s.Actors[i]
	x, y := a.X+cells[dir].Row, a.Y+cells[dir].Col
	if x < FieldMin || x > FieldMax || y < FieldMin || y > FieldMax {
		return true
	}
	if s.Field != nil && BlocksMovement(s.Field.At(x, y)) {
		return true
	}
	if t := s.actorAt(x, y); t >= 0 && s.Actors[t].Side == SideMonster {
		return true
	}
	return false
}

// EnemyNearStyle／EnemyNearGuard 是環 0 找到目標時先填進去的攻擊選項與防禦選項
// （`0x23C81`）。攻擊式抽完會蓋掉前者，而**前者的舊值就是抽命中部位的列號**。
const (
	EnemyNearStyle = 0x0B
	EnemyNearGuard = 7
)

// enemyStrike 填貼近打那一手的攻擊選項與防禦選項（`0x23CD6`…`0x23D8B`）。
//
// `prev` 是進來時的攻擊選項：環 0 那條是 `EnemyNearStyle`，反制掃描那條是
// `counterScan` 從對手抄來的那一格。**`+0x24` 這一格兩種用途**——填進去的是
// 「命中部位 ＋1」，而同一格在這一名被攻擊時就是它的招架姿態（RE-081／RE-105）。
//
// `row` 是抽攻擊式時情境權重的列號（`word_30916`）：環 0 那條是 7
// （`0x23CA1`），反制掃描那條由 `counterScan` 帶回。
func (s *State) enemyStrike(d *EnemyDecision, i, target, prev, row int, rng Rand) {
	a := &s.Actors[i]
	d.Guard = EnemyNearGuard
	// 攻擊選項是 `sub_2578F()` 的結果加二；超過 10 就退回 1（`0x23CEE`）。
	// **這一次規劃裡被困住過就不抽了**，固定第 1 式（`0x23CD6` 先比 `word_34A08`，
	// RE-117）。
	style := 1
	if a.Stuck == 0 {
		style = EnemyStyle(*a, s.AIWeights(), row, a.Weapon.Kind, s.StyleMaskTail, rng) + 2
		if style > 0x0A {
			style = 1
		}
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
		loc := EnemyLocation(*a, s.AIWeights(), prev, rng)
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

// EnemyStuckRetries 是「走不出去」時重跑決策樹的上限。**原版沒有上限**
// ——它靠那 1/3 的擲骰跳出去（期望三次，`0x2396E`）；remake 加一道保險，
// 到頂就出指令碼 0（執行段印記錄 0x117「does nothing.」）。
const EnemyStuckRetries = 8

// enemyBoxed 是走不出去的三條退路（`0x2396E`…`0x23994`，RE-117）：
//
//	傷勢還沒到撤退門檻 → 原地揮武器（貼近打那一段，進來時的攻擊選項與防禦選項是上一個決策
//	                    留在全域的值，`s.aiStyle`／`s.aiGuard`）
//	否則 rand(3) == 1  → 原地喘息
//	否則               → `word_34A08` 加一，重跑決策樹
func (s *State) enemyBoxed(i int, rng Rand) EnemyDecision {
	a := &s.Actors[i]
	if a.FatigueRatio() < a.FleeAt {
		d := EnemyDecision{Command: CmdWeapon, Target: -1, Ring: 0, Facing: a.Facing}
		s.enemyStrike(&d, i, i, s.aiStyle, s.aiGuard, rng)
		s.applyEnemy(i, d)
		return d
	}
	if rng.Roll(3) == 1 {
		return s.enemyRest(i)
	}
	a.Stuck++
	if a.Stuck > EnemyStuckRetries {
		d := EnemyDecision{Command: 0, Target: -1, Ring: -1, Facing: a.Facing}
		s.applyEnemy(i, d)
		return d
	}
	return s.planTactics(i, rng)
}

// aiCellFree 是 AI 挑走向時「這個方向站得下」的兩道檢查：目的格沒有符合遮罩的參戰者
// （`sub_24BB6(0, 方向, 遮罩)`），而且地形不擋路（`sub_24E2B`，只看步行位元圖，
// 不看 +0x1E、也沒有邊界）。這**不是**執行段的 `moveBlocked`（RE-085），兩者不同（RE-168）。
func (s *State) aiCellFree(i, dir, mask int) bool {
	a := s.Actors[i]
	c := RingCells(0)[dir]
	x, y := a.X+c.Row, a.Y+c.Col
	if s.candidateAt(i, x, y, mask) >= 0 {
		return false
	}
	return s.Field == nil || !BlocksMovement(s.Field.At(x, y))
}

// partyNear 是 `sub_24EF9`：以 (x, y) 為中心的八個相鄰格有沒有隊員（遮罩 0x11）。
func (s *State) partyNear(i, x, y int) bool {
	for _, c := range RingCells(0) {
		if s.candidateAt(i, x+c.Row, y+c.Col, maskAny) >= 0 {
			return true
		}
	}
	return false
}

// enemyFlee 是 `0x238E8`：從隨機方向起（`rand(8) − 1`）依方向編號遞增找第一個站得下、
// 目的格沒人（遮罩 0x10）、而且目的格旁邊沒有隊員的方向。走法的環號是 0。
// 八個方向都不行就走 `0x2396E`。
func (s *State) enemyFlee(i int, rng Rand) EnemyDecision {
	a := s.Actors[i]
	start := rng.Roll(8) - 1
	for k := 0; k < 8; k++ {
		dir := (start + k) % 8
		c := RingCells(0)[dir]
		x, y := a.X+c.Row, a.Y+c.Col
		if s.Field != nil && BlocksMovement(s.Field.At(x, y)) {
			continue
		}
		if s.candidateAt(i, x, y, maskHidden) >= 0 || s.partyNear(i, x, y) {
			continue
		}
		return s.enemyStep(i, dir, 0, -1, rng)
	}
	return s.enemyBoxed(i, rng)
}

// enemyApproach 是靠近目標那一步（`0x23F56`…`0x240B6`）：先沿視線走一格（模式 1 的第一步兩軸
// 都走，所以就是兩軸偏移的正負號）換成方向，試這個方向、再試前一個、再試後一個，
// 三個都站不下就隨機挑（`enemyRandom`）。占用的判斷用遮罩 0x18（只算站在地上的）。
func (s *State) enemyApproach(i, target, ring int, rng Rand) EnemyDecision {
	a := s.Actors[i]
	t := s.Actors[target]
	dx, dy := sign(t.X-a.X), sign(t.Y-a.Y)
	base := 0
	for f, c := range RingCells(0) {
		if c.Row == dx && c.Col == dy {
			base = f
			break
		}
	}
	for _, dir := range [3]int{base, (base + 7) % 8, (base + 1) % 8} {
		if s.aiCellFree(i, dir, maskOnFoot|maskHidden) {
			return s.enemyStep(i, dir, ring, target, rng)
		}
	}
	return s.enemyRandom(i, rng)
}

// enemyRandom 是 `0x240B9`：從 `rand(8) − 1` 起依方向編號**遞減**找第一個站得下的方向
// （遮罩 0x18 加地形），走法的環號是 7。八個方向都不行就走 `0x23886`。
func (s *State) enemyRandom(i int, rng Rand) EnemyDecision {
	start := rng.Roll(8) - 1
	for k := 0; k < 8; k++ {
		dir := ((start-k)%8 + 8) % 8
		if s.aiCellFree(i, dir, maskOnFoot|maskHidden) {
			return s.enemyStep(i, dir, 7, -1, rng)
		}
	}
	return s.enemyCornered(i, rng)
}

// enemyStep 是 `0x2399C`：兩條腿都廢了走不動（記錄值都 ≥ 0x80）就走 `0x2396E`；
// 否則方向寫進 +0x28，移動方式走三式的加權輪盤（`0x239D7`）。
func (s *State) enemyStep(i, dir, ring, target int, rng Rand) EnemyDecision {
	a := &s.Actors[i]
	if partLost(a.RightLeg) && partLost(a.LeftLeg) {
		return s.enemyBoxed(i, rng)
	}
	d := EnemyDecision{Command: CmdMove, Target: target, Ring: ring, Facing: dir}
	form := EnemyMoveForm(*a, s.AIWeights(), ring, rng)
	d.Style = IconMoveWalk + form
	if form == 0 {
		d.Guard = 2
	}
	s.applyEnemy(i, d)
	return d
}

// enemyShoot 組出射擊那一手（`0x23F15`…`0x23F53`）：把目標那一格記進 Memo、
// 指令碼設 0x23、攻擊選項 0，並且**直接把 +0x1D 設成 1**——敵人不必先花一回合
// 裝填，原版就是這樣寫的。
func (s *State) enemyShoot(i, target, ring int) EnemyDecision {
	a := &s.Actors[i]
	// 原版射擊那一手不動 +0x28（`0x23F15`…`0x23F53` 只寫目標格、+0x1D、指令與兩個選項）。
	d := EnemyDecision{Command: CmdMissile, Target: target, Ring: ring, Facing: a.Facing}
	a.Flag1D = 1
	if s.Plan.Memo == nil {
		s.Plan.Memo = map[int]Cell{}
	}
	s.Plan.Memo[i] = Cell{Row: s.Actors[target].X, Col: s.Actors[target].Y}
	s.applyEnemy(i, d)
	return d
}

func (s *State) applyEnemy(i int, d EnemyDecision) {
	a := &s.Actors[i]
	a.Command, a.Style, a.Guard = d.Command, d.Style, d.Guard
	// 原版每個出口都把兩個選項留在全域（`word_30A78`／`word_30916`），下一次 AI 進來就是舊值。
	s.aiStyle, s.aiGuard = d.Style, d.Guard
	s.LastCommand = d.Command
	if d.Facing >= 0 {
		a.Facing = d.Facing
	}
}

// 找目標的篩選遮罩（`sub_24C35`，RE-168）。
const (
	maskParty   = 0x01                                // 只看隊員
	maskOnFoot  = 0x08                                // 目標的 +0x1E 必須是 0（騎乘的不算）
	maskHidden  = 0x10                                // Shade Ring 的隱身判斷（`hiddenFrom`）
	maskCaster  = 0x20                                // 目標的指令必須是施法（0x0D）或射擊（0x23）
	maskMelee   = maskParty | maskOnFoot | maskHidden // 0x19：貼身打、反制掃描、撤退時看有沒有人貼著
	maskAny     = maskParty | maskHidden              // 0x11
	maskShooter = maskParty | maskHidden | maskCaster // 0x31
)

// candidateAt 是 `sub_24C35`：依索引順序找站在 (x, y) 而且通過全部條件的參戰者，沒有回 -1。
// 空槽、自己與已經倒下的都不算；同一格上排在前面但沒通過的不會擋住後面的。
func (s *State) candidateAt(i, x, y, mask int) int {
	return s.candidateAtRoll(i, x, y, mask, s.aiRoll)
}

func (s *State) candidateAtRoll(i, x, y, mask, roll int) int {
	for t := range s.Actors {
		o := s.Actors[t]
		if o.Side == SideEmpty || t == i || o.X != x || o.Y != y || o.Down() {
			continue
		}
		if mask&maskParty != 0 && o.Side != SideParty {
			continue
		}
		if mask&maskOnFoot != 0 && o.Flag1E != 0 {
			continue
		}
		if mask&maskHidden != 0 && o.Flag05 != 0 && EnemyIntellectCheck(s.Actors[i].Intellect, s.Actors[i].Level, roll) {
			continue
		}
		if mask&maskCaster != 0 && o.Command != CmdCast && o.Command != CmdMissile {
			continue
		}
		return t
	}
	return -1
}

// findInRing 是貼身那一圈的搜尋（遮罩 0x19）：在指定的環上找一名還站著的隊員，
// 回傳索引；找不到回 -1。
//
// 起點是隨機的（`sub_2531C` 第一次傳 0xFFFF 時從 `rand(格數 − 1) − 1` 起跳），
// 所以同樣的戰況不會總是選同一個目標。
func (s *State) findInRing(i, ring int, rng Rand) int {
	return s.findInRingMatching(i, ring, maskMelee, rng, nil)
}

// 被擋住時延續同一圈與起點，不重擲（SPEC-013；原版 0x23EB8→0x23DEC）。
func (s *State) findInRingMatching(i, ring, mask int, rng Rand, accept func(int) bool) int {
	t, _ := s.findInRingCell(i, ring, mask, rng, accept)
	return t
}

// findInRingCell 同上，另外回傳找到的那一格在環上的編號（環 0 就是方向）。
func (s *State) findInRingCell(i, ring, mask int, rng Rand, accept func(int) bool) (target, cell int) {
	a := s.Actors[i]
	cells := RingCells(ring)
	if len(cells) == 0 {
		return -1, -1
	}
	start := rng.Roll(len(cells)-1) - 1
	for n := 0; n < len(cells); n++ {
		idx := (start + n) % len(cells)
		c := cells[idx]
		if t := s.candidateAt(i, a.X+c.Row, a.Y+c.Col, mask); t >= 0 && (accept == nil || accept(t)) {
			return t, idx
		}
	}
	return -1, -1
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

// facingToward 只算「朝哪一格」，不看那一格走不走得過去——攻擊與射擊要的是朝向，
// 不是一步走得到的位置。原版先用 `sub_25D7A(1)` 沿視線走一步再查 `[0x10D8]` 把
// 相對格差換成方向（RE-081）；這裡直接取符號，兩者在相鄰與正對角的情形相同，
// 斜向偏移大時可能差一格。
func (s *State) facingToward(i, target int) int {
	a := s.Actors[i]
	if target < 0 {
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

// hiddenFrom 是篩選遮罩 0x10（`sub_24C35:0x24D76`…`0x24DC8`）：`sub_252CB(看的人的智力,
// word_3AC5E)` 過關、而目標的 `+0x05` 非零（Shade Ring），就當那一格沒有人。
// `word_3AC5E` 是 AI 開頭擲的那一次（`PlanEnemy` 的 `aiRoll`），這裡不再擲。
func (s *State) hiddenFrom(i, t int, rng Rand) bool {
	a := s.Actors[i]
	return s.Actors[t].Flag05 != 0 && EnemyIntellectCheck(a.Intellect, a.Level, s.aiRoll)
}

// AIWeights 是這一場的三張基礎權重表：由這一場的怪物記錄決定（`Monster.AIWeights`）。
func (s *State) AIWeights() AIWeights {
	return s.Monster.AIWeights()
}
