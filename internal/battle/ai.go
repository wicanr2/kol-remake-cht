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
// 撤退（`0x237F5`…`0x238C1`）、反制掃描（`sub_253E7`）、側翼（`sub_25653`）與
// 空手去撿（`0x23BC2`，RE-116）也都照原版做了。
//
// 反制掃描的篩選遮罩 0x10 要的 `+0x05` **在原版也恆為零**（RE-118：那一格只由
// 穿戴欄部位 3／5／7／8 那件東西的種類 byte 決定，而 `ITEM.LST` 裡放得進那四格的
// 三件種類 byte 都是 0），所以那個條件兩邊都恆真——不是缺口。
//
// **解出來了但還沒接上**：移動段的方向要先試往目標那一步、再試左右各一格，沒有
// 目標時從隨機方向起繞八格，八格都站不下另有三條退路（RE-117）。接上去會改變
// 戰況，`arena` 與 `full-chain-ending` 那兩條長路徑的固定按鍵序列得整條重生
// （各有兩到三場戰鬥，中間還夾著城鎮互動），所以留待專門的一輪處理。

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

	// 兩隻手臂都廢就只能移動（`0x237E9`）。
	if a.RightArm <= 0 && a.LeftArm <= 0 {
		return s.enemyMove(i, -1, -1, rng)
	}

	// 撤退（`0x237F5`…`0x238C1`）。傷勢是 `50 × 累積傷害 ÷ 昏迷門檻`，
	// 兩個門檻來自怪物記錄的 +0x0D／+0x0E——不死系是 100／0，永遠走不進來。
	if d, ok := s.enemyRetreat(i, rng); ok {
		return d
	}
	// `word_34A08` 是一次規劃之內的計數（`sub_2374C` 開頭歸零，RE-117）。
	a.Stuck = 0
	return s.planTactics(i, rng)
}

// planTactics 是決策樹從「手上有沒有武器」那個分岔起的整段（`0x23BC2`…`0x24030`）。
// 抽出來是因為繞不出去的第三條退路會**回到這裡重跑**（RE-117）。
func (s *State) planTactics(i int, rng Rand) EnemyDecision {
	a := &s.Actors[i]
	d := EnemyDecision{Target: -1, Ring: -1, Facing: a.Facing}

	// 手上那一格空著（`0x23BC2`，RE-116）：先看參戰者記錄 +0x2B，非零就清掉改成
	// 移動，否則出指令碼 0x26 去撿地上的東西，攻擊選項 0、防禦選項 2。
	// **原版沒有任何地方把 +0x2B 設成非零**（兩個寫入點都寫 0），所以移動那一條
	// 走不到；照著實作是為了讓這一段與原版逐條對得起來。
	if a.Unarmed {
		if a.PickedUp != 0 {
			a.PickedUp = 0
			return s.enemyMove(i, -1, -1, rng)
		}
		d.Command = CmdPickUp
		d.Style = 0
		d.Guard = 2
		s.applyEnemy(i, d)
		return d
	}

	// 反制掃描（`0x23C1A`）：智力擲骰過關才跑。有結論就直接走攻擊那一段，
	// 不再往外環找目標——抄來的攻擊選項會變成抽命中部位的列號。
	if EnemyIntellectCheck(a.Intellect, a.Level, rng.Roll(100)) {
		if prev, t, ok := s.counterScan(i, rng); ok {
			if rng.Roll(3) == 1 {
				return s.enemyMove(i, t, 0, rng)
			}
			d.Command = CmdWeapon
			d.Target = t
			d.Ring = 0
			d.Facing = a.Facing
			s.enemyStrike(&d, i, t, prev, rng)
			s.applyEnemy(i, d)
			return d
		}
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
		d.Facing = s.facingToward(i, t)
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
func (s *State) counterScan(i int, rng Rand) (style, target int, ok bool) {
	a := &s.Actors[i]
	cells := RingCells(0)
	best, bestStyle, last := 0, 0, -1
	// 起點隨機，與 `sub_2531C` 第一次傳 0xFFFF 時的行為相同。
	start := rng.Roll(len(cells)) - 1
	for k := 0; k < len(cells); k++ {
		n := (start + k) % len(cells)
		c := cells[n]
		t := s.actorAt(a.X+c.Row, a.Y+c.Col)
		if t < 0 || s.Actors[t].Side != SideParty || s.Actors[t].Down() {
			continue
		}
		d := &s.Actors[t]
		pri := 0
		if !s.planned(t) {
			// 還沒規劃（`0x255A6`）。
			if best < 4 {
				best, bestStyle, pri = 4, EnemyNearStyle, 4
				if d.Command == CmdWeapon && rng.Roll(2) == 1 {
					bestStyle = d.Style
				}
			}
		} else {
			if d.Command == CmdWeapon || d.Command == CmdUnarmed {
				if d.Facing^n == 4 {
					if d.Command == CmdWeapon {
						best, bestStyle = 5, d.Style
					}
					if best < 3 {
						best, bestStyle = 3, EnemyNearStyle
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
				bestStyle, best = 0, pri
			}
		}
		if pri != 0 {
			last = t
			a.Facing = n
		}
	}
	if best == 0 || last < 0 {
		return 0, -1, false
	}
	// 優先度 1（對方要移動）才繞側面（`0x25435`）。
	if best == 1 {
		s.flank(i, last)
	}
	return bestStyle, last, true
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
	// **這一次規劃裡被困住過就不抽了**，固定第 1 式（`0x23CD6` 先比 `word_34A08`，
	// RE-117）。
	style := 1
	if a.Stuck == 0 {
		style = EnemyStyle(a.Level, s.Actors[target].Guard, a.Weapon.Kind, rng) + 2
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

// EnemyStuckRetries 是「八個方向都走不了」時重跑決策樹的上限。**原版沒有上限**
// ——它靠那 1/3 的擲骰跳出去（期望三次，`0x2396E`）；remake 加一道保險，
// 到頂就出指令碼 0（執行段印記錄 0x117「does nothing.」）。
const EnemyStuckRetries = 8

// enemyStuck 是繞不出去的三條退路（`0x2396E`…`0x23994`，RE-117）：
//
//	傷勢還沒到撤退門檻 → 原地揮武器（貼近打那一段，式固定第 1 式）
//	否則 rand(3) == 1  → 回撤退段
//	否則               → `word_34A08` 加一，重跑決策樹
func (s *State) enemyStuck(i int, rng Rand) EnemyDecision {
	a := &s.Actors[i]
	if a.Injury() < a.FleeAt {
		// 原地揮武器。目標是自己面前那一格，沒有實際目標時就打空氣——
		// 原版跳的是 `0x23CD6`，也就是貼近打那一段的開頭。
		d := EnemyDecision{Command: CmdWeapon, Target: -1, Ring: 0,
			Facing: a.Facing}
		s.enemyStrike(&d, i, i, EnemyNearStyle, rng)
		s.applyEnemy(i, d)
		return d
	}
	if rng.Roll(3) == 1 {
		a.Retreating = true
		d := EnemyDecision{Command: CmdRest, Target: -1, Ring: -1, Facing: a.Facing}
		s.applyEnemy(i, d)
		return d
	}
	a.Stuck++
	if a.Stuck > EnemyStuckRetries {
		d := EnemyDecision{Command: 0, Target: -1, Ring: -1, Facing: a.Facing}
		s.applyEnemy(i, d)
		return d
	}
	return s.planTactics(i, rng)
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
		Facing: s.facingToward(i, target)}
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
	facing := s.stepToward(i, target, rng)
	if facing < 0 {
		// 八個方向都站不下：走那三條退路（RE-117）。
		return s.enemyStuck(i, rng)
	}
	d := EnemyDecision{Command: CmdMove, Target: target, Ring: ring}
	d.Facing = facing
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

// stepToward 是移動段挑方向那一段（`0x23F56`…`0x240B9` 與 `0x238E8`…`0x2394F`，
// RE-117）：有目標就先試往目標那一步，不通再試左右各一格；三個都不通、或者根本
// 沒有目標，就從**隨機**方向起繞八格。八格都站不下回 -1。
func (s *State) stepToward(i, target int, rng Rand) int {
	a := s.Actors[i]
	if target >= 0 {
		// 有目標（`0x23F56`…`0x240B9`）：先試往目標那一步，不通再試左右各一格。
		t := s.Actors[target]
		dx, dy := sign(t.X-a.X), sign(t.Y-a.Y)
		base := -1
		for f, c := range RingCells(0) {
			if c.Row == dx && c.Col == dy {
				base = f
				break
			}
		}
		if base >= 0 {
			for _, f := range [3]int{base, (base + 7) % 8, (base + 1) % 8} {
				c := RingCells(0)[f]
				if !s.moveBlocked(i, a.X+c.Row, a.Y+c.Col) {
					return f
				}
			}
		}
	}
	// 沒有目標，或者那三個方向都不通（`0x240B9` 退回這一條）：從**隨機方向**起
	// 繞一圈（`0x238E8` 的 `rand(8) − 1`）。八格都站不下就回 -1，由呼叫端走
	// 三條退路。
	start := rng.Roll(8) - 1
	for k := 0; k < 8; k++ {
		f := (start + k) % 8
		c := RingCells(0)[f]
		if !s.moveBlocked(i, a.X+c.Row, a.Y+c.Col) {
			return f
		}
	}
	return -1
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
