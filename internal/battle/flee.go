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
	quest := s.Site != nil && s.Site.Index >= 0 && s.QuestItem != ""
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
		if quest {
			if m := s.member(i); m != nil && m.WeaponName == s.QuestItem {
				out.Carrying++
			}
		}
	}
	out.Ask = out.Carrying > 0 || (party > 0 && out.Edge == party)
	return out
}

// Flee 是玩家答應逃走那一下（`0x2116C`…`0x2117B`）。帶著任務要的東西的話
// 任務就標成完成——**不必把怪物打完**（RE-101）。回傳標到幾個人。
func (s *State) Flee() int {
	if s.FleeOffer().Carrying > 0 {
		s.QuestDone = s.markQuestDone()
	}
	s.Verdict = VerdictFled
	s.Stage = StageDone
	return s.QuestDone
}
