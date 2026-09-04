package battle

// 施法（RE-100）。執行段的指令碼 `0x0D` 走這一支（RE-093）。法術不是獨立的資料檔，
// 而是存在**角色記錄 0x200 起、每 16 byte 一筆、共十六筆**——版面與商品記錄相同。

// SpellSlots 是角色帶得動的法術筆數；SpellBase 是它們在角色記錄裡的起點。
const (
	SpellSlots = 16
	SpellBase  = 0x200
)

// 施法會用到的 KOLDATA 記錄。
const (
	NothingThereRecord = 0x169 // "casts a spell, but nothing is there!"
	NoEffectRecord     = 0x16A // "casts a spell, but it has no effect!"
	CastsAtRecord      = 0x16B // "casts|S at|F."
	AreaListRecord     = 0x16C // ", 1 head, 2 chest, … or 6 left leg."
	SpellHasRecord     = 0x16D // "The spell has a]"
	MildlyRecord       = 0x16E // mildly／markedly／strongly／powerfully
	EffectRecord       = 0x172 // " effect."
	WearsOffRecord     = 0x173 // "'s spell wears off!"
	SelectAreaRecord   = 0x184 // "Select an area to]"
	HealRecord         = 0x185 // "heal"
	HurtRecord         = 0x186 // "hurt"
	PositiveRecord     = 0x106 // "Positive"（+0x27 為 0）
	NegativeRecord     = 0x107 // "Negative"（+0x27 非零）
)

// SpellStrength 是強度形容詞的段數：強度 ÷ 8，上限 3（`0x287F9`）。
const SpellStrength = 8

// Spell 是一筆法術記錄。程式只讀四處，其餘與商品記錄相同。
type Spell struct {
	// Name 是法術的名字（`|S` 展開的那一段）。
	Name string
	// Effect 是 +0x0C 的效果編號。0 或大於 8 什麼都不做。
	Effect int
	// TurnCode 是 +0x0D（持續回合的骰式）、PowerCode 是 +0x0E（效果強度的骰式）。
	// PowerCode 的 bit 7 另外表示「這一手要不要先選一個目標格」。
	TurnCode, PowerCode int
	// Guards 是 +0x08…+0x0B 那四格，擋招要用（`0x282EC`）。
	Guards [4]int
}

// NeedsTarget 是「要不要先選目標格」（`+0x0E` 的 bit 7）。沒設的法術朝施法者
// 面對的方向施放。
func (s Spell) NeedsTarget() bool { return s.PowerCode&0x80 != 0 }

// SpellDice 把法術的骰式拆成骰數與骰面（`0x28380`…`0x283AC`）。
//
// **與武器傷害碼那一套（`DamageCode`）差一個位移**：法術先右移一位再取低六位，
// 低三位查骰面表、整段除以 8 得骰數。兩邊各有各的反組譯出處，不要合併。
func SpellDice(code int) (count, sides int) {
	v := (code >> 1) & 0x3F
	return v/8 + 1, diceSides[v&7]
}

// Turns 是這一次施法的持續回合、Power 是效果強度。原版先擲持續、再擲強度
// （`0x283B2` 與 `0x28420`），順序不能對調——共用同一顆亂數。
func (s Spell) Turns(rng Rand) int { return rollDice(s.TurnCode, rng) }

// Power 是效果強度。
func (s Spell) Power(rng Rand) int { return rollDice(s.PowerCode, rng) }

func rollDice(code int, rng Rand) int {
	count, sides := SpellDice(code)
	total := 0
	for i := 0; i < count; i++ {
		total += rng.Roll(sides)
	}
	return total
}

// SpellEffectField 是效果 1…8 對應的參戰者欄位（DGROUP `0x12A4`）。
// 第 5、6 項用不到——那兩個效果在前面就被自己的分支接走了。
var SpellEffectField = [8]int{0x08, 0x09, 0x0A, 0x0C, 0x2C, 0x2C, 0x2E, 0x2F}

// 兩個有自己分支的效果編號。
const (
	// SpellEffectStat 之外的兩個特例：5 改一個還沒解出語意的全域，
	// 6 是「指定部位治療或傷害」。
	SpellEffectGlobal = 5
	SpellEffectArea   = 6
)

// SpellVerdict 是一次施法的去向。
type SpellVerdict int

const (
	// SpellNothingThere 是目標格上沒有人（記錄 0x169）。
	SpellNothingThere SpellVerdict = iota
	// SpellNoEffect 是擋下來了（記錄 0x16A）。
	SpellNoEffect
	// SpellHits 是施法成功。
	SpellHits
)

// CastResult 是一次施法的結果。
type CastResult struct {
	Verdict SpellVerdict
	// Turns 是持續回合、Power 是效果強度；沒生效時兩個都是 0。
	Turns, Power int
	// Area 是效果 6 選中的部位 0…5，其餘效果是 −1。
	Area int
}

// SpellBlocked 是「目標擋不擋得住」（`0x28284`…`0x2835C`）。
//
// 倒下的目標一定擋不住。否則拿法術的四格護符與目標那一格（隊員看 `+0x1D`、
// 敵人看 `+0x28`）比高低 nibble：高 nibble 要相同；那一格 ≥ 0x50 時低 nibble
// 還要夠大。四格裡有一格中就算穿得過去。
func SpellBlocked(sp Spell, guard int, down bool) bool {
	if down {
		return false
	}
	high, low := guard&0xF0, guard&0x0F
	for _, g := range sp.Guards {
		gh, gl := g&0xF0, g&0x0F
		if gh != high {
			continue
		}
		if gh >= 0x50 && gl < low {
			continue
		}
		return false
	}
	return true
}

// StrengthWord 是「The spell has a …」那一句要用的記錄：強度除以 8，上限 3。
func StrengthWord(power int) int {
	n := power / SpellStrength
	if n > 3 {
		n = 3
	}
	return MildlyRecord + n
}

// DirectionWord 是 Positive 或 Negative 那一格（`0x28844`）。
func DirectionWord(negative bool) int {
	if negative {
		return NegativeRecord
	}
	return PositiveRecord
}

// Cast 是施法那一手（`sub_2818D`）。target 是目標的參戰者索引、area 是效果 6 要打
// 哪一個部位（0…5，其他效果傳 −1）。
//
// 擲骰的順序照原版：**先擲持續回合、再擲強度**（`0x283B2` 與 `0x28420`），
// 兩者共用同一顆亂數，對調會讓整串結果不同。
func (s *State) Cast(caster, target, area int, rng Rand) CastResult {
	res := CastResult{Verdict: SpellNothingThere, Area: -1}
	if caster < 0 || caster >= len(s.Actors) {
		return res
	}
	a := &s.Actors[caster]
	m := s.member(caster)
	if m == nil || a.SpellPick < 0 || a.SpellPick >= len(m.Spells) {
		return res
	}
	sp := m.Spells[a.SpellPick]
	if target < 0 || target >= len(s.Actors) || s.Actors[target].Side == SideEmpty {
		return res
	}
	if SpellBlocked(sp, s.Actors[target].Ward, s.Actors[target].Down()) {
		res.Verdict = SpellNoEffect
		return res
	}
	res.Verdict = SpellHits
	res.Turns = sp.Turns(rng)
	res.Power = sp.Power(rng)
	switch {
	case sp.Effect == SpellEffectArea:
		res.Area = area
		s.castArea(target, area, res.Power, a.Negative)
	case sp.Effect == SpellEffectGlobal:
		// 效果 5 改的那個全域（`[0x322E]`）語意還沒解（RE-100），先不動任何東西。
	case sp.Effect >= 1 && sp.Effect <= len(SpellEffectField):
		s.castStat(target, sp.Effect, res.Power, res.Turns, a.Negative)
	}
	return res
}

// castArea 是效果 6：治療或傷害指定的部位（`0x2859E`…`0x2869A`）。
func (s *State) castArea(target, area, power int, negative bool) {
	if area < 0 || area > 5 {
		return
	}
	t := &s.Actors[target]
	if negative {
		t.Take(power, area)
		return
	}
	// 治療：補到那個部位的上限為止。廢掉的部位（原版是 0xFF）從零開始補。
	caps := Parts(t.Threshold)
	parts := [6]*int{&t.Head, &t.Chest, &t.RightArm, &t.LeftArm, &t.RightLeg, &t.LeftLeg}
	v := *parts[area]
	if v < 0 {
		v = 0
	}
	v += power
	if v > caps[area] {
		v = caps[area]
	}
	*parts[area] = v
	// 部位那一條把「身上有法術」設回 0（`0x283EA`）——它不是持續效果。
	t.SpellTurns = 0
}

// castStat 是效果 1…4、7、8：把某一個屬性加減強度，並掛上倒數（`0x28704`）。
//
// 效果 7、8 動的是參戰者記錄 `+0x2E`／`+0x2F`，那兩格的語意還沒解（RE-100），
// 所以這一版對它們什麼都不做——寧可不動，也不要動錯一格。
func (s *State) castStat(target, effect, power, turns int, negative bool) {
	t := &s.Actors[target]
	var p *int
	switch SpellEffectField[effect-1] {
	case 0x08:
		p = &t.Strength
	case 0x09:
		p = &t.Quickness
	case 0x0A:
		p = &t.Foresight
	case 0x0C:
		p = &t.Intellect
	default:
		return
	}
	v := *p
	if negative {
		if v -= power; v < 0 {
			v = 0
		}
	} else {
		if v += power; v > 0xFF {
			v = 0xFF
		}
	}
	*p = v
	t.SpellTurns = turns + 1
}

// WearsOff 是回合開始時的法術倒數（`sub_1E3A8:0x20F01`）。
//
// 倒數到 0 就把屬性重算回原值——原版的做法是重新呼叫 `sub_2198A`（隊員）或
// `sub_2215A`（敵人）把角色／怪物記錄的數值整批搬回參戰者記錄，所以法術的加成
// 不必個別記帳，過期就整個重來。回傳這一位的效果是不是剛好在這一回合結束
//（要印記錄 0x173）。
func (s *State) WearsOff(i int) bool {
	if i < 0 || i >= len(s.Actors) {
		return false
	}
	a := &s.Actors[i]
	if a.SpellTurns <= 0 {
		return false
	}
	a.SpellTurns--
	if a.SpellTurns > 0 {
		return false
	}
	s.restoreStats(i)
	return true
}

// restoreStats 把一名參戰者的屬性重算回原值。
func (s *State) restoreStats(i int) {
	a := &s.Actors[i]
	if m := s.member(i); m != nil {
		a.Strength = Nourished(m.Strength, m.Nutrition)
		a.Quickness = m.Quickness
		a.Foresight = m.Foresight
		a.Intellect = m.Intellect
		return
	}
	a.Strength = s.Monster.Strength
	a.Quickness = s.Monster.Quickness
	a.Foresight = s.Monster.Foresight
	a.Intellect = s.Monster.Intellect
}

// SpellEffectLine 是「The spell has a <強度> <方向> effect.」那一句
//（`0x287E4`…`0x2885E`），拆成前後兩段。
//
// **中間要插一個空白**：原版在強度與方向之間另外印一次 `sub_1C2DA(' ')`，
// 那個空白不在任何一筆記錄裡，接起來會變成 `mildlyPositive`。
func SpellEffectLine(power int, negative bool) (before, after []int) {
	return []int{SpellHasRecord, StrengthWord(power)},
		[]int{DirectionWord(negative), EffectRecord}
}

// Member 取一名參戰者對應的隊員資料；不是隊員就回 nil。訊息那一層要用它取法術名。
func (s *State) Member(i int) *Member { return s.member(i) }
