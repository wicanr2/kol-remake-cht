package battle

// 一手的時間額度與疲勞（RE-130）。
//
// 原版每一手收尾都做同一串事：`sub_236AB` 先算出這一手的額度 `word_30D06`，
// 三個地方從裡面扣掉花費，剩下的除以十扣進疲勞（`0x20D40`…`0x20D8C`）。額度是**負**
// 的時候疲勞反而上升——傷得越重、動作越貴，額度越容易變負，這就是「受傷 → 疲勞 →
// 撤退」那條迴路。沒有它，`FatigueRatio()` 一整場都是 0，撤退與力竭都不會發生。
//
// 選了休息而且這一手沒挨打的人走另一條（`0x20AE0`）：疲勞剩七成，而且**跳過**
// 這一整段。

// guardTime 是 `[0x10BE]`：防禦姿態的負重乘數（`sub_268D4:0x26980`）。
var guardTime = [8]int{0, 2, 1, 2, 2, 2, 2, 0}

// moveTime 是 `[0x10E4]`：移動方式的負重乘數（`sub_268D4:0x26953`）。
var moveTime = [6]int{2, 3, 6, 2, 4, 4}

// 施法要多久：三張表相加（`0x20BCE`…`0x20D1E`）。
//
// `guardCastTime` 是 `[0x10EA]`，索引是護符那一格的**高 nibble**（類別）；四格裡
// 取值最大的那一格，它的低 nibble 乘十六一起算進去。`effectCastTime` 是
// `[0x10F8]`，索引是效果編號。`diceCastTime` 是 `[0x1102]`，索引是骰面碼**加一**。
//
// **最後一格是原版讀出界的結果**：骰面碼 7（99 面）查到的是表尾之後那一個 byte，
// 也就是攻擊選項修正表 `[0x110A]` 的第一格 0x46。照抄。
var (
	guardCastTime  = [14]int{100, 200, 150, 250, 0, 0, 250, 80, 200, 0, 0, 120, 100, 0}
	effectCastTime = [10]int{0, 120, 150, 140, 130, 200, 200, 100, 100, 0}
	diceCastTime   = [9]int{30, 50, 70, 90, 110, 130, 210, 0, 0x46}
)

// 疲勞的兩個係數：這一手挨的傷一點五個單位，剩下的額度十分之一進疲勞。
const (
	hurtTimeCost  = 5
	fatigueDivide = 10
)

// TimeAllowance 是這一手的額度，還沒扣任何花費（`sub_236AB`）：
//
//	額度 = 耐力 + 耐力 ÷ 2 − 疲勞 − 累積傷害 × 10 + (昏迷門檻 − 累積傷害) × 2
//
// 除法先做再相加，不是乘一點五。
func (a Actor) TimeAllowance() int {
	free := a.Threshold - a.Damage
	return a.Stamina + a.Stamina/2 - a.Fatigue - a.Damage*10 + free*2
}

// commandTimeIndex 是這一手指令的時間成本索引（`0x20B4A`…`0x20BA2`）。
// 移動與騎乘移動是 0、射擊 9、徒手 8、武器攻擊查 `[0x10B2 + 攻擊選項]`，其餘 0。
func commandTimeIndex(a Actor) int {
	switch a.Command {
	case CmdMove, CmdMountedMove:
		return 0
	case CmdMissile:
		return 9
	case CmdUnarmed:
		return 8
	case CmdWeapon:
		if a.Style >= 0 && a.Style < len(commandTime) {
			return commandTime[a.Style]
		}
	}
	return 0
}

// encumbranceMultiplier 是負重那一項的乘數（`sub_268D4`）：移動查移動方式那張表、
// 下馬固定 2、武器與徒手查防禦姿態那張表，其餘 1；騎著再加倍。
func encumbranceMultiplier(a Actor) int {
	m := 1
	switch a.Command {
	case CmdMove, CmdMountedMove:
		// 攻擊選項在移動時放的是圖號；減掉基準得到 0…2，為負就當第 5 格。
		k := a.Style - IconMoveWalk
		if k < 0 {
			k = 5
		}
		if k < len(moveTime) {
			m = moveTime[k]
		}
	case CmdLand:
		m = 2
	case CmdWeapon, CmdUnarmed:
		if a.Guard >= 0 && a.Guard < len(guardTime) {
			m = guardTime[a.Guard]
		}
	}
	if a.Flag1E != 0 {
		m *= 2
	}
	return m
}

// CastTime 是施法要花掉的額度（`0x20BCE`…`0x20D1E`）。四格護符裡挑類別最貴的那一
// 格，加上它的種、效果與骰式；要選目標格的法術（強度碼 bit 7）再加倍。
func CastTime(sp Spell) int {
	best, at := 0, 0
	for k, g := range sp.Guards {
		if c := castLookup(guardCastTime[:], g>>4); c > best {
			best, at = c, k
		}
	}
	cost := best + (sp.Guards[at]&0x0F)<<4 + castLookup(effectCastTime[:], sp.Effect)
	// 骰數與骰面各查一次：骰數是 `(強度碼 >> 4) & 7` 加一，骰面碼加一查表。
	dice := ((sp.PowerCode>>4)&7 + 1) * castLookup(diceCastTime[:], (sp.PowerCode>>1)&7+1)
	if sp.PowerCode > 0x7F {
		dice *= 2
	}
	return cost + dice
}

func castLookup(table []int, i int) int {
	if i < 0 || i >= len(table) {
		return 0
	}
	return table[i]
}

// spendTime 是這一手花掉之後剩下的額度（`0x20B27`…`0x20D1E`）。
//
// 負重那一項的第一個乘數（`sub_251B0` 的 `var_2`）用武器速度。原版在
// 「`+0x22` 非零而且這一手是徒手」時改成從本體記錄取一格，但 `+0x22` 的語意還沒解，
// 所以那一條沒實作——它只影響徒手那一手的額度。
func (s *State) spendTime(i int) int {
	a := &s.Actors[i]
	left := a.TimeAllowance()
	left -= a.Hurt * hurtTimeCost
	left -= commandTimeIndex(*a)*a.WeaponSpeed + encumbranceMultiplier(*a)*a.Encumbrance
	if a.Command == CmdCast {
		if m := s.member(i); m != nil &&
			a.SpellPick >= 0 && a.SpellPick < len(m.Spells) {
			left -= CastTime(m.Spells[a.SpellPick])
		}
	}
	return left
}

// ExhaustedRatio 是力竭的門檻（`0x20D9A`）。到了就**當場倒下**——原版把頭與胸
// 都設成 0xFF、指令清掉、重畫戰場（`0x20E05`…`0x20E52`），與被打倒同一個結果。
const ExhaustedRatio = 100

// 力竭那三句（`0x20DCF`…`0x20DFD`）。隊員與怪物分開，怪物再看記錄 +0x10 的 bit 7。
const (
	ExhaustedPartyRecord    = 0x11C // collapses from exhaustion!
	ExhaustedMonsterRecord  = 0x11D // dies from blood loss!
	ExhaustedShattersRecord = 0x11E // shatters into pieces!
)

// ActionCost 是 `sub_251B0(a, b)`：`a × 武器速度 + b × 負重`。敵人 AI 拿它跟這一手
// 的額度比，決定攻擊式與命中部位的權重要偏多少。
//
// 原版在「`+0x22` 非零（這一回合規劃過了）而且這一手是徒手」時，速度那一項改成
// 從角色記錄取徒手那一式的物品記錄第 0x0A 格。remake 的 typed data 沒有那四筆
// 徒手記錄，所以這一條沒實作——缺口記在 SPEC-008。
func ActionCost(a Actor, timeIndex, weightIndex int) int {
	return timeIndex*a.WeaponSpeed + weightIndex*a.Encumbrance
}

// ClampEffort 是 `sub_2526C`：把「額度 ÷ 花費」夾進 1/3…1，回傳夾過的分子與分母。
// 額度是負的、或者比花費還大，就直接當 1；花費超過額度三倍就當 1/3。
func ClampEffort(allow, cost int) (int, int) {
	switch {
	case allow < 0:
		return 1, 3
	case allow > cost:
		return 1, 1
	case cost > 3*allow:
		return 1, 3
	}
	return allow, cost
}

// EndTurnFatigue 是一手收尾的疲勞更新（`0x20AE0` 與 `0x20D40`）。回傳力竭那一句的
// 記錄編號，沒力竭就回 0。
//
// 呼叫點要在清掉 `Hurt` **之前**——休息那一條與挨傷那一項都要讀它。
func (s *State) EndTurnFatigue(i int) int {
	a := &s.Actors[i]
	if a.Hurt == 0 && a.Command == CmdRest {
		// 休息那一條跳過額度那一整段。
		a.Fatigue = a.Fatigue * RestRecoveryNum / RestRecoveryDen
	} else {
		left := s.spendTime(i)
		// 騎著的人另外折算：額度為負就加倍、否則折半（`0x20D30` 看 `+0x1E`）。
		if a.Flag1E != 0 {
			if left < 0 {
				left *= 2
			} else {
				left /= 2
			}
		}
		if a.Fatigue -= left / fatigueDivide; a.Fatigue < 0 {
			a.Fatigue = 0
		}
	}
	if a.FatigueRatio() < ExhaustedRatio {
		return 0
	}
	// 倒下：頭與胸清掉（remake 用 0 表示廢掉），指令也清掉。
	a.Head, a.Chest, a.Command = 0, 0, 0
	if a.Side == SideParty {
		return ExhaustedPartyRecord
	}
	if a.Shatters {
		return ExhaustedShattersRecord
	}
	return ExhaustedMonsterRecord
}
