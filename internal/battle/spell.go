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

// 攻防技能罰值那兩格；`SpellEffectField` 的第 7、8 項就是它們。
const (
	SpellFieldSkillOff = 0x2E
	SpellFieldSkillDef = 0x2F
)

// 兩個有自己分支的效果編號。
const (
	// SpellEffectStat 之外的兩個特例：5 動的是疲勞（`[0x322E + 2i]`，
	// `MAGIC.0` 屬性表的 `Fatigue`），6 是「指定部位治療或傷害」（`Body`）。
	SpellEffectFatigue = 5
	SpellEffectArea    = 6
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
	// NeedsArea 為真表示骰已經擲完、效果還沒套，正等玩家挑一個部位。
	NeedsArea bool
}

// AreaPrompt 是問部位那一句要用的三筆記錄（`0x2845E`…`0x284F1`）：
// 「Select an area to」＋ hurt／heal ＋ 部位清單。
func AreaPrompt(negative bool) []int {
	verb := HealRecord
	if negative {
		verb = HurtRecord
	}
	return []int{SelectAreaRecord, verb, AreaListRecord}
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

// AreaAsk 是傳給 `Cast` 的哨兵：效果 6 而且不用選目標格時，先把持續與強度擲出來，
// **停在這裡等玩家挑部位**，不套效果。擲骰要在問之前——原版的順序是
// `0x283B2`（持續）、`0x28420`（強度）、`0x2845E`（問），順序換了整串亂數就不同。
const AreaAsk = -2

// Cast 是施法那一手（`sub_2818D`）。target 是目標的參戰者索引、area 是效果 6 要打
// 哪一個部位（0…5，其他效果傳 −1，`AreaAsk` 表示要問）。
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
	case sp.Effect == SpellEffectArea && area == AreaAsk:
		// 停在問部位那一層：效果留到 `AnswerArea` 才套。
		res.NeedsArea = true
	case sp.Effect == SpellEffectArea:
		res.Area = area
		s.castArea(target, area, res.Power, a.Negative)
	case sp.Effect == SpellEffectFatigue:
		s.castFatigue(target, res.Power, a.Negative)
	case sp.Effect >= 1 && sp.Effect <= len(SpellEffectField):
		s.castStat(target, sp.Effect, res.Power, res.Turns, a.Negative)
	}
	return res
}

// castFatigue 是效果 5：疲勞加減「強度 × 2」（`0x286AB`）。這一個效果的正負**是反
// 過來的**——負面往上加、正面往下減，因為疲勞越高越糟（對照 `castStat`：那一段
// 對六個欄位一視同仁，沒有反過來，RE-130）。
//
// 原版下限那一段是死碼：`cmp` 比的是位址常數 `322Eh` 而不是那一格的值，`jb` 永遠
// 不成立，所以**疲勞減得到負數**。照抄——比例跟著變成負的，攻防因此小幅變好。
// 這一條也不掛倒數：效果 5 走的是自己的分支，不碰 `+0x30`。
func (s *State) castFatigue(target, power int, negative bool) {
	t := &s.Actors[target]
	if negative {
		t.Fatigue += power * 2
		return
	}
	t.Fatigue -= power * 2
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
// 效果 7、8 動的是參戰者記錄 `+0x2E`／`+0x2F`，也就是攻防技能的**罰值**
// （`Off. Skill`／`Def. Skill`，RE-129）。原版這一段對六個欄位一視同仁：
// 正面加、負面減，夾在 0…0xFF——**沒有為罰值那兩格把正負反過來**（效果 5 的
// 疲勞就有反，見 `0x286AB`）。所以照原版跑出來的結果是：正面的法術把罰值加上去
// ＝把對方的技能扣掉強度的十倍，負面的法術反而是把罰值減回來（罰值本來就是 0
// 的話什麼都不會發生）。這一版照抄，不代它修。
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
	case SpellFieldSkillOff:
		p = &t.SkillPenaltyOff
	case SpellFieldSkillDef:
		p = &t.SkillPenaltyDef
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
// （要印記錄 0x173）。
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
// （`0x287E4`…`0x2885E`），拆成前後兩段。
//
// **中間要插一個空白**：原版在強度與方向之間另外印一次 `sub_1C2DA(' ')`，
// 那個空白不在任何一筆記錄裡，接起來會變成 `mildlyPositive`。
func SpellEffectLine(power int, negative bool) (before, after []int) {
	return []int{SpellHasRecord, StrengthWord(power)},
		[]int{DirectionWord(negative), EffectRecord}
}

// Member 取一名參戰者對應的隊員資料；不是隊員就回 nil。訊息那一層要用它取法術名。
func (s *State) Member(i int) *Member { return s.member(i) }

// 選法術那兩張清單的標題（KOLDATA）。
const (
	WhichSpellRecord   = 0x104 // "~V07~H03Which Spell?"
	SelectAffectRecord = 0x105 // "~V07~H03Select Affect"
)

// SpellListBox 是那兩張清單的框（`0x1F3C1` 的 `sub_1BA9D(10h, 2, 88h, 4Ch, 0Fh)`）
// ——與裝備欄同一個框，清單直接蓋在上面。
var SpellListBox = SheetBox

// SpellListRows 是清單第一行的位置與行距。
//
// 標題的座標是記錄自己帶的（`~H03~V07`）；清單是 `sub_1E3A8:0x1F4DE` 呼叫
// `sub_1AAE8(3, 4, 11h, 可見列數, 總筆數, 2E3Ah, 0Dh)`，依 RE-115 的參數語意：
// 第一列左緣在第 4 個 byte 欄、上緣在第 0x11 條掃描線，可見列數上限 7（`0x1F4D8`），
// 行距 `6 + 2 × word_2DF18`，戰鬥裡是大字的 8（與戰利品清單相同）。
var SpellListRows = struct{ Col, Row, LineStep, Visible int }{
	Col: 4, Row: 0x11, LineStep: 8, Visible: 7,
}

// SpellListTitle 是目前這一張清單的標題記錄。
func (s *State) SpellListTitle() int {
	if s.Plan.Prompt == PromptAffect {
		return SelectAffectRecord
	}
	return WhichSpellRecord
}

// SpellNames 是「Which Spell?」那張清單的內容：那一名帶著的法術名字，順序同
// 角色記錄裡的欄位（`0x1F411`）。不是隊員或沒有法術就回 nil。
func (s *State) SpellNames(i int) []string {
	m := s.member(i)
	if m == nil {
		return nil
	}
	out := make([]string, 0, len(m.Spells))
	for _, sp := range m.Spells {
		out = append(out, sp.Name)
	}
	return out
}

// AffectRecords 是「Select Affect」那兩項的記錄（`0x1F5E5`）：先 Positive、
// 後 Negative，選中的索引就寫進參戰者記錄 +0x27。
func AffectRecords() []int { return []int{PositiveRecord, NegativeRecord} }

// beginSpellList 進「Which Spell?」那張清單（`0x1F3C1`）。
func (s *State) beginSpellList(i int) {
	s.Plan.Prompt = PromptSpell
	s.Plan.List = s.SpellNames(i)
	s.Plan.ListPick = 0
	s.Plan.Bar = nil
}

// chooseSpell 選好法術之後接著問正負（`0x1F553`）。清單的文字由呼叫端填——
// 那兩項是 KOLDATA 記錄，規則層不碰文字。
func (s *State) chooseSpell(i int) bool {
	if s.Plan.ListPick < 0 || s.Plan.ListPick >= len(s.Plan.List) {
		return false
	}
	s.Actors[i].SpellPick = s.Plan.ListPick
	s.Plan.Prompt = PromptAffect
	s.Plan.List = []string{"", ""}
	s.Plan.ListPick = 0
	return false
}

// chooseAffect 選好正負之後去選目標格（`0x1F639`）。
//
// 用哪一種游標由法術自己決定：`+0x0E` 的 bit 7 沒設就是相鄰、設了就是遠距
// （`0x1F677` 的 `mode = 記錄[+0x0E] ÷ 0x80 + 1`）。
func (s *State) chooseAffect(i int) bool {
	s.Actors[i].Negative = s.Plan.ListPick != 0
	mode := CursorAdjacent
	if m := s.member(i); m != nil {
		pick := s.Actors[i].SpellPick
		if pick >= 0 && pick < len(m.Spells) && m.Spells[pick].NeedsTarget() {
			mode = CursorRanged
		}
	}
	s.Plan.List, s.Plan.ListPick = nil, 0
	s.beginTarget(i, mode, PromptCommand)
	return false
}
