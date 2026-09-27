package battle

// 戰鬥中的逃走（RE-103）。每個回合走完、還沒分勝負時，原版先數誰站到了戰場
// 邊上（`0x21045`…`0x21129`），符合條件才問記錄 **0x109**
// （`Do you want to flee?`）。

// OnFieldEdge 回報一格是不是可走區的最外圈。原版比的是 `+0x1F` 或 `+0x20`
// 等於 4 或 0x3B（`0x21087`…`0x210DB`）——**兩軸任一個到邊就算**，
// 不是要站在角上。
func OnFieldEdge(x, y int) bool {
	return x == FieldMin || x == FieldMax || y == FieldMin || y == FieldMax
}

// FleeOffer 是逃走那一問的狀態。
type FleeOffer struct {
	// Ask 是這一回合要不要問。
	Ask bool
	// Edge 是幾名隊員站到了戰場邊上（`word_3AC5A`）。
	Edge int
	// Carrying 是其中幾名手上拿著任務要的那件東西（`word_39A54`）。
	Carrying int
}

// FleeOffer 算這一回合的逃走條件。
//
// 兩條路任一條成立就問：**有人帶著任務要的東西站到邊上**（`word_39A54!= 0`），
// 或者**全隊都站到了邊上**（`word_3AC5A == word_2DF2A`）。兩條都不成立連問都
// 不問，直接排下一回合。
//
// 「帶著東西」是 `sub_2886C(0x30)`：把角色記錄 0x30 起的十六個 byte 與
// `[0xB926]` 逐 byte 比，也就是手上那件的名字等於任務要的那件。隨機遭遇那一支
// 一律回 0（`0x28875` 先看 `word_34968`），所以野外遭遇只剩「全隊到邊上」
// 那一條——remake 用 `Site.Index < 0` 表示同一件事。
func (s *State) FleeOffer() FleeOffer {
	var out FleeOffer
	party := 0
	for i := range s.Actors {
		a := s.Actors[i]
		if a.Side != SideParty {
			continue
		}
		party++
		// 原版在這一段不看倒下的人，只看座標（`0x21059` 只比 +0x01）。
		if !OnFieldEdge(a.X, a.Y) {
			continue
		}
		out.Edge++
		if m := s.member(i); m != nil && s.matchesQuestItem(m.HandItem) {
			out.Carrying++
		}
	}
	out.Ask = out.Carrying > 0 || (party > 0 && out.Edge == party)
	return out
}

// FleeAskRecord 是問句那一筆（`0x21135`）。
const FleeAskRecord = 0x109

// OfferFlee 在符合條件時把規劃段停在逃走那一問，回報有沒有停下來
// （`0x21116`…`0x21146`）。
func (s *State) OfferFlee() bool {
	if !s.FleeOffer().Ask {
		return false
	}
	s.Stage = StageFlee
	s.Plan.Prompt = PromptFlee
	s.Plan.Bar = nil
	s.Plan.Selected = 0
	return true
}

// Flee 是玩家答應逃走那一下（`0x2116C`…`0x2117B`）。帶著任務要的東西的話
// 任務就標成完成——**不必把怪物打完**（RE-101）。回傳標到幾個人。
//
// 這條路帶著本回合數到的兩個倒下計數跳到共同出口 `0x211AD`，兩個都不是 0xFF，
// 出口不印逃走那一句，直接離開戰場（RE-146，`FleeLine`）。
func (s *State) Flee() int {
	if s.FleeOffer().Carrying > 0 {
		s.QuestDone = s.markQuestDone()
	}
	s.Verdict = VerdictFled
	s.Stage = StageDone
	s.fledFromBar = false
	// 回到執行段那一層；訊息清空讓呼叫端知道出口這一步輪到它處理。
	s.Plan.Prompt = PromptExecuting
	s.Exec.Text = nil
	return s.QuestDone
}

// IconFleeBar 是「Begin next round?」那一列多出來的逃走格（RE-143）。
const IconFleeBar = 0x30

// NextRoundBar 是整輪規劃完那一列（`sub_1E3A8:0x1FB3x`）：YES、U-TURN，
// `word_345DE == 0xFFFF`（競技場與訓練戰以外的戰鬥，含任務戰鬥）再加逃走格
// （`loc_1FB5F`）。
func (s *State) NextRoundBar() []int {
	bar := []int{CmdYes, CmdCancel}
	if !s.RankFight {
		bar = append(bar, IconFleeBar)
	}
	return bar
}

// NextRoundChoice 是「Begin next round?」那一列選下去之後往哪裡走。
type NextRoundChoice int

const (
	// NextRoundExecute 是 YES：進執行段。
	NextRoundExecute NextRoundChoice = iota
	// NextRoundReplan 是 U-TURN：跳回 `0x1EF3A` 重新規劃這一回合。
	NextRoundReplan
	// NextRoundAskFlee 是逃走格：印記錄 0x109 問要不要逃（`loc_1FBAD`）。
	NextRoundAskFlee
)

// ChooseNextRound 依白框停在哪一格處理「Begin next round?」（`0x1FB6B`）。
func (s *State) ChooseNextRound() NextRoundChoice {
	switch s.icon() {
	case CmdCancel:
		s.Answer(false)
		return NextRoundReplan
	case IconFleeBar:
		s.Stage = StageFlee
		s.fleeFromBar = true
		s.Plan.Prompt = PromptFlee
		s.Plan.Bar = nil
		s.Plan.Selected = 0
		return NextRoundAskFlee
	}
	s.Answer(true)
	return NextRoundExecute
}

// BarFleeChance 是從逃走格逃走時第 i 名的判定門檻（`0x1FC01`…`0x1FC90`）：
// （智力 `+0x0C` ＋ 力量 `+0x08`）÷ 2 × 等級 `+0x02` ÷ 3。
func (s *State) BarFleeChance(i int) int {
	a := s.Actors[i]
	return (a.Intellect + a.Strength) / 2 * a.Level / 3
}

// BarFlee 是在逃走格答應之後（`loc_1FBF9`…`0x1FCF5`，RE-143）：
//
// 全隊一定逃得掉，六格逐名擲 `rand(100)`，**大於**門檻的那一名，手上那件的種類
// byte 高三位元非零（是武器）就丟在腳下（`sub_269BA`）。原版在這裡另看
// `[word_37D2A]+0x06` 與「手上是不是任務要的那件」，但 `+0x06` 在現行資料下恆為 0
// （RE-141），那一層等於不作用，remake 以這一名自己的 `Flag06` 代入。
//
// 這條路直接跳到共同出口 `0x211AD`，**不經過** `0x21176`——所以與戰場邊上的逃走
// 不同，帶著任務物品從這裡逃走不會把任務標成完成。
func (s *State) BarFlee(rng Rand) {
	for i := 0; i < PartySlots && i < len(s.Actors); i++ {
		a := &s.Actors[i]
		if a.Side != SideParty {
			continue
		}
		if rng.Roll(100) <= s.BarFleeChance(i) {
			continue
		}
		m := s.member(i)
		if m == nil || !m.HasWeapon || m.Weapon.Kind&0xE0 == 0 {
			continue
		}
		if a.Flag06 != 0 && !s.matchesQuestItem(m.HandItem) {
			continue
		}
		s.DropItem(i)
	}
	s.fleeFromBar = false
	s.fledFromBar = true
	s.Verdict = VerdictFled
	s.Stage = StageDone
	s.Plan.Prompt = PromptExecuting
	s.Exec.Text = nil
}

// FleeLine 回報戰鬥出口要不要印「You flee for your life!」（`0x21592`…`0x215EE`）。
// 出口要 `word_345DA` 與 `word_34966` 都等於 0xFF 才印，而這兩個計數只在逃走格答 y
// 的 `0x1FCF5` 設成 0xFF；戰場邊上的逃走帶著本回合的倒下計數出去，一般戰鬥落到
// `0x21607` 什麼都不印（RE-146）。
func (s *State) FleeLine() bool {
	return s.Verdict == VerdictFled && s.fledFromBar
}
