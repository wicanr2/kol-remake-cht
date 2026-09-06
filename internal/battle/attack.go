package battle

// 近戰的攻擊結算（RE-080／SPEC-009）。這裡只做規則層：算攻擊分數與防禦分數、
// 判命中、擲傷害與部位，再把傷害記到身上。指令列、選目標與敵人 AI 不在這裡。

// 七張表原樣抄自 KOLTOWN 的 DGROUP，位移寫在各自的註解裡。

// optionModifier 是 0x110A：一列七個 16-bit 有號數，列由攻擊選項決定、
// 欄由對手的防禦選項決定，直接加到攻擊分數上。
var optionModifier = [9][7]int{
	{70, -10, 20, 13, -15, -2, 15},
	{80, 0, 30, 18, 30, 13, 20},
	{75, -5, 25, 18, 30, 13, -20},
	{56, -24, 6, -21, -44, -6, 21},
	{55, -25, 5, -27, -10, -2, 15},
	{48, -32, -2, -29, -12, -4, -27},
	{37, -43, -13, 0, -58, -50, -18},
	{55, -25, 5, 13, -5, -27, -5},
	{43, -37, -7, 6, -12, -34, -52},
}

// styleWeaponMask 是 0x109C：攻擊選項要的武器位元。武器的種類 byte
// （物品記錄 +0x0D）沒有那個位元，攻擊分數就砍半。
var styleWeaponMask = [16]int{16, 16, 16, 8, 8, 8, 4, 4, 4, 2, 1, 3, 4, 4, 4, 4}

// styleDamage 是 0x1200：攻擊選項的傷害加成。
var styleDamage = [16]int{0, 4, 0, 0, 0, 2, 2, 2, 1, 1, 1, 0, 0, 1, 2, 4}

// marginBonus 是 0x120C：命中餘裕除以 20 之後查到的傷害加成。
var marginBonus = [16]int{0, 1, 2, 4, 6, 8, 11, 14, 17, 21, 25, 30, 35, 40, 45, 50}

// locationOdds 是 0x121C：命中部位的累積機率，一列六個，
// 順序是頭、胸、右臂、左臂、右腿、左腿。
var locationOdds = [3][6]int{
	{25, 70, 85, 100, 0, 0},
	{5, 60, 75, 90, 95, 100},
	{0, 10, 10, 10, 55, 100},
}

// styleLocation 是 0x122E：攻擊選項用哪一列部位機率。
var styleLocation = [16]int{1, 1, 0, 1, 2, 0, 1, 2, 0, 1, 2, 0, 6, 7, 0, 5}

// diceSides 是 0x00D4：傷害碼的高位挑骰面數。
var diceSides = [8]int{2, 3, 5, 7, 9, 11, 19, 99}

// parryMessage 是 0x11B2：被擋下時印哪一句（記錄 **0x15F** + 值，RE-088），
// 列是防禦選項、欄是攻擊選項。
var parryMessage = [8][11]int{
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
	{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2},
	{3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3},
	{4, 4, 4, 2, 9, 4, 2, 9, 4, 2, 9},
	{5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5},
	{7, 7, 2, 7, 6, 8, 8, 6, 8, 8, 6},
	{0, 0, 0, 4, 0, 0, 0, 0, 2, 2, 2},
}

// AttackRoll 是攻擊、防禦與部位各自擲的骰面數。
const AttackRoll = 100

// NoShield 是「沒有盾牌」：原版用 0xFF 表示（`sub_24FD3:0x250CD`）。
const NoShield = 0xFF

// Weapon 是手上那件武器在戰鬥要用的數字，全部取自物品記錄（RE-075）。
type Weapon struct {
	// Kind 是 +0x0D 的種類 byte：武器技能列與攻擊式的可用性都比對它。
	Kind int
	// Base 是 +0x08 高 nibble 的基礎傷害；Count／Sides 由 +0x0C 的傷害碼算出來。
	Base  int
	Count int
	Sides int
	// Code 是 +0x0C 的傷害碼原值。組指令列時要用它的 bit 0（`sub_228B7:0x22951`），
	// 而 Count／Sides 只保留高位元的部分，取不回來。
	Code int
	// Weight 是 +0x0A 的負重。
	Weight int
	// TwoHand 為真時不能同時拿盾牌（+0x0D 的 bit 1）。
	TwoHand bool
}

// UnarmedSlots 是徒手各式拿哪一格穿戴品當「武器」：`sub_26AC2:0x26BDA` 用攻擊選項
// 去查 `[0x0FD2]`（`00 50 60 00 40 10`）再加 `0x100`，而穿戴欄十格就是從 `0x100`
// 起、一格 `0x10`（RE-127）。所以除以 16 得到的槽號是：
//
//	式 1 → 5 FEET（踢）、式 2 → 6 SHIELD（盾擊）、式 3 → 0 HEAD（頭槌）、
//	式 4 → 4 HANDS（拳）
//
// 式 0 在執行段就被歸零（`0x2067C`），式 5 用不到；兩個都指回 HEAD 與 TORSO。
// 十格的順序是 HEAD／TORSO／LEGS／BODY／HANDS／FEET／SHIELD／NECK／RING／BELT。
var UnarmedSlots = [6]int{0, 5, 6, 0, 4, 1}

// AttackWeapon 是這一手真正用到的那件東西：徒手拿的是對應那一格穿戴品，
// 其餘拿手上那件（`sub_26AC2` 依指令碼挑 `var_4`）。**傷害骰、種類與重量都從它來**。
func (a Actor) AttackWeapon(style int) Weapon {
	if a.Command != CmdUnarmed {
		return a.Weapon
	}
	if style < 0 || style >= len(a.UnarmedWeapon) {
		return Weapon{}
	}
	return a.UnarmedWeapon[style]
}

// DamageCode 把物品記錄 +0x0C 的傷害碼換成骰數與骰面（`sub_26D76:0x274D6`）。
func DamageCode(code int) (count, sides int) {
	return code/32 + 1, diceSides[(code/4)&7]
}

// Nourished 是力量依飢餓折算之後的值：`力量 × (510 − 飢餓) ÷ 510`
// （`sub_2198A:0x219CF`）。角色記錄 +0x9F 是飢餓值——野外每換一天就往上加
// （RE-077），0 表示吃飽，255 時力氣只剩一半。
func Nourished(strength, nutrition int) int {
	return strength * (510 - nutrition) / 510
}

// EncumbranceFactor 是負重的係數：70 − (等級 − 1) × 5（`sub_24FD3:0x24FEA`）。
func EncumbranceFactor(level int) int { return 70 - (level-1)*5 }

// Encumbrance 算負重與武器速度。armour 是頭、軀幹、腿三件的重量，
// shield 是盾牌重量（沒有盾牌傳 0）。`sub_24FD3`。
func Encumbrance(level int, w Weapon, armour, shield int) (enc, speed int) {
	factor := EncumbranceFactor(level)
	speed = w.Weight * factor / 100
	total := w.Weight/2 + armour + shield
	return factor * total / 100, speed
}

// ReloadGear 依現在手上那件武器重算負重與武器速度。
//
// **原版每次算一手的花費都重跑一遍**：`sub_251B0` 開頭就呼叫 `sub_24FD3`
// （實跑量到，`internal/oracle` 的 `TestActionCostMatchesOriginal`——不擋掉它，
// 寫進去的速度與負重會被空白記錄算出來的 0 蓋掉）。所以戰鬥中撿起、丟下或
// 換走手上那件武器，兩個數字都要立刻跟著變：花費 → 疲勞 → 撤退判斷是一條鏈，
// 而攻擊分數與執行鍵值也各自減去負重。
func (a *Actor) ReloadGear() {
	a.Encumbrance, a.WeaponSpeed = Encumbrance(
		a.LoadLevel, a.Weapon, a.ArmourWeight, a.ShieldWeight)
}

// SkillPenaltyScale 是罰值折算成技能的倍率：`sub_26CD5` 扣的是那一格的十倍。
const SkillPenaltyScale = 10

// OffenceSkill／DefenceSkill 是這一刻真正算數的攻防技能：手上武器那一列的值
// 扣掉法術效果 7、8 累積的罰值乘十，下限 0（`sub_26AC2` → `sub_26CD5`）。
// 原版每次要用都重算一遍，所以罰值不改記錄裡的技能本身。
func (a Actor) OffenceSkill() int { return skillAfterPenalty(a.Attack, a.SkillPenaltyOff) }

// DefenceSkill 同上，對應防禦那一格。
func (a Actor) DefenceSkill() int { return skillAfterPenalty(a.Parry, a.SkillPenaltyDef) }

func skillAfterPenalty(skill, penalty int) int {
	if v := skill - penalty*SkillPenaltyScale; v > 0 {
		return v
	}
	return 0
}

// AttackValue 是攻擊分數的骰前部分（`sub_26D76:0x26E08`…`0x26E7E`）。
func (a Actor) AttackValue() int {
	v := a.Strength + a.OffenceSkill()
	v -= v*a.FatigueRatio()/200 + a.Damage*10 + a.Hurt*20
	return v - a.Encumbrance
}

// DefenceValue 是防禦分數的骰前部分（`sub_26D76:0x27229`…`0x272FB`）。
// 招架次數是這一回合已經擋過幾次，愈多擋得愈差；至少算一次。
func (a Actor) DefenceValue() int {
	parries := a.Parries
	if parries < 1 {
		parries = 1
	}
	v := a.DefenceSkill()/parries + a.Quickness
	v -= v*a.FatigueRatio()/200 + a.Damage*10 + a.Encumbrance
	if a.ShieldGrade != NoShield {
		v += 100*a.ShieldGrade + 200
	}
	return v
}

// AttackScore 是這一擊的攻擊分數。style 是攻擊選項，guard 是對手的防禦選項。
// row 用來挑 optionModifier 的列：原版在攻擊選項為 1 時改擲 `rand(3) − 1`
// （`sub_26D76:0x26FE7`），因此那一種要由呼叫端把擲出來的值傳進來。
func AttackScore(a Actor, style, guard, row, roll int) int {
	score := a.AttackValue()/10 + roll
	if row >= 0 && row < len(optionModifier) && guard >= 0 && guard < len(optionModifier[0]) {
		score += optionModifier[row][guard]
	}
	if score >= 0 && style > 1 {
		if mask := styleWeaponMask[(style-2)&0x0F]; a.AttackWeapon(style).Kind&mask == 0 {
			score /= 2
		}
	}
	return score
}

// ModifierRow 是 optionModifier 要用的列號（`sub_26D76:0x26FDD`）。
// 攻擊選項為 1 時是 `rand(3) − 1`，其餘是選項加一。
func ModifierRow(style int, rng Rand) int {
	if style == 1 {
		return rng.Roll(3) - 1
	}
	return style + 1
}

// DefenceScore 是對手的防禦分數。防禦選項為 0 表示不招架，分數直接是 0
// （`sub_26D76:0x27204`）；算出來是負的也當 0（`0x27454`）。
func DefenceScore(d Actor, roll int) int {
	if d.Guard == 0 {
		return 0
	}
	return clampDefence(d.DefenceValue()/10 + roll)
}

// MissileDefenceValue 是被射中時防禦分數的骰前部分（`sub_26D76:0x271E3`）：
// 招架與招架次數都不算，換成快捷的一半，其餘折算與盾牌照舊。射擊那一條在
// `0x271FE` 直接跳到 `0x2728C`，連「防禦選項為 0」的檢查都跳過了。
func (a Actor) MissileDefenceValue() int {
	v := a.Quickness / 2
	v -= v*a.FatigueRatio()/200 + a.Damage*10 + a.Encumbrance
	if a.ShieldGrade != NoShield {
		v += 100*a.ShieldGrade + 200
	}
	return v
}

// MissileDefenceScore 是被射中時的防禦分數。
func MissileDefenceScore(d Actor, roll int) int {
	return clampDefence(d.MissileDefenceValue()/10 + roll)
}

// clampDefence 是 `0x27382` 的那一條：防禦分數是負的就當 0 再比，
// 所以命中餘裕不會因為對手的分數很負而灌水。
func clampDefence(v int) int {
	if v < 0 {
		return 0
	}
	return v
}

// HitLocation 依攻擊選項挑一列累積機率，再用 1…100 的擲骰選部位
// （`sub_26D76:0x27665`）。回傳 0…5，順序是頭、胸、右臂、左臂、右腿、左腿。
func HitLocation(style, roll int) int {
	row := styleLocation[style&0x0F]
	if row >= len(locationOdds) {
		row = 0
	}
	for i, odds := range locationOdds[row] {
		if roll <= odds {
			return i
		}
	}
	return len(locationOdds[row]) - 1
}

// DamageRoll 擲這一擊的傷害（`sub_26D76:0x274D6`…`0x275B4`）。
// margin 是攻擊分數減防禦分數。回傳值已夾在 0 與武器上限之間。
func DamageRoll(a Actor, style, margin int, rng Rand) int {
	w := a.AttackWeapon(style)
	bonus := (a.Strength - 70) / 8
	bonus = (70 - a.FatigueRatio()) * bonus / 100
	total := w.Base
	for i := 0; i < w.Count; i++ {
		total += rng.Roll(w.Sides)
	}
	total += bonus + styleDamage[style&0x0F]
	if step := margin / 20; step >= 0 {
		if step >= len(marginBonus) {
			step = len(marginBonus) - 1
		}
		total += marginBonus[step]
	}
	if limit := w.Count*w.Sides + 1 + w.Base; total > limit {
		total = limit
	}
	if total < 0 {
		total = 0
	}
	return total
}

// Outcome 是一次攻擊的結果。
type Outcome struct {
	// Miss 為真代表攻擊分數是負的，整個落空；Margin 是差多少（正值）。
	Miss   bool
	Margin int
	// Parried 為真代表被擋下；Message 是記錄 0x14F 起的第幾句。
	Parried bool
	Message int
	// Damage／Location 是打中之後的傷害與部位（0…5）。
	Damage   int
	Location int
	// Down 為真代表這一擊把對手打倒；Limb 為真代表那一肢廢了。
	Down bool
	Limb bool
}

// Resolve 打一次近戰。style 是攻擊選項，rng 依序擲：攻擊、（選項為 1 時的列）、
// 防禦、傷害骰、部位。攻擊者與被攻擊者的狀態都會被更新。
func Resolve(a *Actor, d *Actor, style int, rng Rand) Outcome {
	row := ModifierRow(style, rng)
	attack := AttackScore(*a, style, d.Guard, row, rng.Roll(AttackRoll))
	if attack < 0 {
		return Outcome{Miss: true, Margin: -attack}
	}
	if d.Guard != 0 {
		d.Parries++
	}
	defence := DefenceScore(*d, rng.Roll(AttackRoll))
	if attack < defence {
		guard := d.Guard & 7
		msg := parryMessage[guard][style&0x0F]
		// 碼 4 而且對方壯很多時換成 2（`sub_26D76:0x27417`）。
		if msg == 4 && d.Size-a.Size > 100 {
			msg = 2
		}
		return Outcome{Parried: true, Message: msg}
	}
	damage := DamageRoll(*a, style, attack-defence, rng)
	location := HitLocation(style, rng.Roll(AttackRoll))
	out := Outcome{Damage: damage, Location: location}
	out.Down, out.Limb = d.Take(damage, location)
	return out
}

// ResolveMissile 是射中之後的結算（`sub_26D76` 的 0x23 那一條，RE-102）。
// attack 是彈道扣完剩下的命中值——沿路的樹與人已經先扣過了。
//
// stray 為真表示這一發是被擋在半路的人接下的（`word_30A74 == 2`）：原版那時把
// 兩邊的分數都換成 `rand(100)`（`0x27153`），彈道算出來的命中值不算數。
//
// 射擊不會被招架：分數不夠時原版印的是「but misses completely!」而不是招架句，
// 所以這裡照樣回 Parried，訊息由 `AttackLine` 依指令碼分（RE-088）。
func ResolveMissile(a *Actor, d *Actor, style, attack int, stray bool, rng Rand) Outcome {
	defence := 0
	if stray {
		attack = rng.Roll(AttackRoll)
		defence = clampDefence(rng.Roll(AttackRoll))
	} else {
		if attack < 0 {
			return Outcome{Miss: true, Margin: -attack}
		}
		defence = MissileDefenceScore(*d, rng.Roll(AttackRoll))
	}
	if attack < defence {
		return Outcome{Parried: true}
	}
	damage := DamageRoll(*a, style, attack-defence, rng)
	location := HitLocation(style, rng.Roll(AttackRoll))
	out := Outcome{Damage: damage, Location: location}
	out.Down, out.Limb = d.Take(damage, location)
	return out
}

// Take 把一次傷害記到身上（`sub_277F4`）：累積傷害到頂就倒下，
// 部位血量扣光那一肢就廢了。回傳是不是倒下、那一肢是不是廢了。
func (a *Actor) Take(damage, location int) (down, limb bool) {
	a.Damage += damage
	if a.Damage > 0x7F {
		a.Damage = 0x7F
	}
	if a.Threshold > 0 && a.Damage >= a.Threshold {
		a.Head, a.Chest = 0, 0
		down = true
	}
	parts := [6]*int{&a.Head, &a.Chest, &a.RightArm, &a.LeftArm, &a.RightLeg, &a.LeftLeg}
	if location >= 0 && location < len(parts) {
		p := parts[location]
		if *p > 0 {
			*p -= damage
			if *p <= 0 {
				*p = 0
				limb = true
			}
		}
	}
	a.Hurt += damage
	return down, limb
}
