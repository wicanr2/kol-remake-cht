package town

import "github.com/wicanr2/kol-remake-cht/internal/magic"

// 法術編輯器（SCROLL 選單第 1 列，`sub_178E0`，RE-128）的十三個欄位。編號就是
// 原版的欄位游標：上下左右在 0…12 之間繞圈（`0x17ABD`／`0x17AE5`）。
const (
	// SpellFieldCount 是欄位數。
	SpellFieldCount = 13
	// SpellFieldName 是法術名稱：`sub_1B5C8` 收 12 個字，`sub_1C0F6` 打包成
	// 記錄開頭那八個 byte。這一欄不查 `MAGIC.0`，也不花錢。
	SpellFieldName = 0
	// 欄位 1…4 是四格護符的高 nibble（`+0x08`…`+0x0B`），也就是「對哪一類有效」。
	SpellFieldClassFirst = 1
	SpellFieldClassLast  = 4
	// 欄位 5…8 是同四格的低 nibble，「那一類底下的哪一種」。
	SpellFieldSubFirst = 5
	SpellFieldSubLast  = 8
	// SpellFieldStat 是影響哪個屬性（`+0x0C`）。
	SpellFieldStat = 9
	// SpellFieldTargetFlag 是 `+0x0E` 的 bit7：施法前要不要先選一格。
	SpellFieldTargetFlag = 10
	// SpellFieldTurn／SpellFieldPower 是持續回合與強度的骰式（`+0x0D`／`+0x0E`）。
	SpellFieldTurn  = 11
	SpellFieldPower = 12
)

// SpellGuards 是那四格護符。
const SpellGuards = 4

// SpellNameMax 是名稱的字數上限（`sub_1B5C8` 的第二個引數）。
const SpellNameMax = 12

// StatNoTurnsFirst…StatNoTurnsLast 是「這個屬性的效果不持續」的那一段：
// 屬性落在 3…6 時原版不畫持續回合、也不收那一欄的錢，改印一句 KOLDATA 0x52
// （`0x1869E`…`0x186CA`）。
const (
	StatNoTurnsFirst = 3
	StatNoTurnsLast  = 6
)

// guardAt 取第 i 格護符；沒有那一格就當 0。
func guardAt(sp Spell, i int) int {
	if i < 0 || i >= len(sp.Guards) {
		return 0
	}
	return sp.Guards[i]
}

// setGuard 寫回第 i 格護符，必要時把那一串補到四格。
func setGuard(sp *Spell, i, value int) {
	if i < 0 || i >= SpellGuards {
		return
	}
	for len(sp.Guards) < SpellGuards {
		sp.Guards = append(sp.Guards, 0)
	}
	sp.Guards[i] = value & 0xFF
}

// spellNoTurns 是「這一筆法術的屬性不吃持續回合」。
func spellNoTurns(sp Spell) bool {
	return sp.Effect >= StatNoTurnsFirst && sp.Effect <= StatNoTurnsLast
}

// SpellFieldValue 取一個欄位現在停在選單第幾項。名稱那一欄沒有選單，回 0。
func SpellFieldValue(d *magic.Dictionary, sp Spell, field int) int {
	switch {
	case field == SpellFieldName:
		return 0
	case field <= SpellFieldClassLast:
		return guardAt(sp, field-SpellFieldClassFirst) >> 4
	case field <= SpellFieldSubLast:
		return guardAt(sp, field-SpellFieldSubFirst) & 0x0F
	case field == SpellFieldStat:
		return sp.Effect
	case field == SpellFieldTargetFlag:
		return (sp.PowerCode >> 7) & 1
	case field == SpellFieldTurn:
		return diceIndex(d, sp.TurnCode)
	case field == SpellFieldPower:
		return diceIndex(d, sp.PowerCode)
	}
	return 0
}

func diceIndex(d *magic.Dictionary, code int) int {
	if d == nil {
		return 0
	}
	return d.DiceIndex(code)
}

// SetSpellField 把選單挑中的第 n 項寫回工作副本（`0x18093`…`0x1811F`）。骰式那
// 兩欄是把 `MAGIC.0` 的 `0x200` 表查出來的值 or 回去，原本的 bit0 與 bit7 留著
// ——bit7 是另一欄（要不要先選目標）在管的。
func SetSpellField(d *magic.Dictionary, sp *Spell, field, value int) {
	switch {
	case field == SpellFieldName:
		return
	case field <= SpellFieldClassLast:
		i := field - SpellFieldClassFirst
		setGuard(sp, i, guardAt(*sp, i)&0x0F|(value&0x0F)<<4)
	case field <= SpellFieldSubLast:
		i := field - SpellFieldSubFirst
		setGuard(sp, i, guardAt(*sp, i)&0xF0|value&0x0F)
	case field == SpellFieldStat:
		sp.Effect = value & 0xFF
	case field == SpellFieldTargetFlag:
		sp.PowerCode = sp.PowerCode&0x7F | (value&1)<<7
	case field == SpellFieldTurn:
		sp.TurnCode = sp.TurnCode&0x81 | diceValue(d, value)
	case field == SpellFieldPower:
		sp.PowerCode = sp.PowerCode&0x81 | diceValue(d, value)
	}
}

func diceValue(d *magic.Dictionary, n int) int {
	if d == nil {
		return 0
	}
	return d.DiceValue(n) & 0x7E
}

// SpellFieldOptions 是一個欄位的選單內容（`0x17E0B`…`0x17E9D` 挑的那幾段）。
// 名稱那一欄回 nil——它走的是文字輸入，不是選單。
func SpellFieldOptions(d *magic.Dictionary, sp Spell, field int) []string {
	if d == nil {
		return nil
	}
	switch {
	case field == SpellFieldName:
		return nil
	case field <= SpellFieldClassLast:
		return d.Class
	case field <= SpellFieldSubLast:
		// 那一格自己的高 nibble 決定選單是哪一組。
		return d.SubclassGroup(guardAt(sp, field-SpellFieldSubFirst) >> 4)
	case field == SpellFieldStat:
		return d.Stat
	case field == SpellFieldTargetFlag:
		// 兩個選項是 KOLDATA 記錄 0x50／0x51，不在 `MAGIC.0` 裡。
		return nil
	case field == SpellFieldTurn:
		return trim(d.Dice, magic.TurnDiceOptions)
	case field == SpellFieldPower:
		return d.Dice
	}
	return nil
}

func trim(rows []string, n int) []string {
	if n > len(rows) {
		n = len(rows)
	}
	return rows[:n]
}

// SpellFieldText 是一個欄位現在顯示什麼。名稱與目標旗標那兩欄的字不在
// `MAGIC.0` 裡，由呼叫端補。
func SpellFieldText(d *magic.Dictionary, sp Spell, field int) string {
	if d == nil {
		return ""
	}
	v := SpellFieldValue(d, sp, field)
	switch {
	case field == SpellFieldName:
		return sp.Name
	case field <= SpellFieldClassLast:
		return d.ClassName(v)
	case field <= SpellFieldSubLast:
		return d.SubclassName(guardAt(sp, field-SpellFieldSubFirst)>>4, v)
	case field == SpellFieldStat:
		return d.StatName(v)
	case field == SpellFieldTurn:
		if spellNoTurns(sp) {
			return ""
		}
		return d.DiceName(v)
	case field == SpellFieldPower:
		return d.DiceName(v)
	}
	return ""
}

// SpellCost 是這一筆法術的兩種造價（`sub_18350` 邊畫邊累加）：四格護符各自依
// 「整個 byte」查表，屬性那一格另一張表，「要不要先選目標」是固定級距，兩個
// 骰式各自拆成骰數與骰面再查。屬性落在不吃持續回合的那一段時，持續那一欄
// 連算都不算。
func SpellCost(d *magic.Dictionary, sp Spell) (adventure, gold int) {
	if d == nil {
		return 0, 0
	}
	add := func(ap, gc int) {
		adventure += ap
		gold += gc
	}
	for i := 0; i < SpellGuards; i++ {
		add(d.TargetCost(guardAt(sp, i)))
	}
	add(d.StatCostOf(sp.Effect))
	add(d.TargetFlagCost(sp.PowerCode&0x80 != 0))
	if !spellNoTurns(sp) {
		add(d.DiceCostOf(sp.TurnCode))
	}
	add(d.DiceCostOf(sp.PowerCode))
	return adventure, gold
}
