package battle

import "sort"

// 回合結構（RE-080）。一個回合先算出手順序，再照那個順序讓每一名各行動一次，
// 走完整輪才問記錄 0x108「Begin next round?」。

// InitiativeRoll 是先攻擲的骰面數（`sub_1B5A8(0x64)`）。
const InitiativeRoll = 100

// LevelPoints 是升一級要的攻防值總和；LevelStep 是等級換算成先攻的權重。
const (
	LevelPoints = 1000
	LevelStep   = 10
)

// EmptyKey 是空位的先攻鍵值：排在所有人後面（`sub_1E3A8:0x1EF7B`）。
const EmptyKey = 0x7FFF

// CombatLevel 是隊員的戰鬥等級：四列武器的攻防原始值加總除以 1000 再加一
// （`sub_1E3A8:0x1EE64`…`0x1EF18` 把八個值加起來除以 0x3E8）。
func CombatLevel(skills []int) int {
	total := 0
	for _, v := range skills {
		total += v
	}
	return total/LevelPoints + 1
}

// Threshold 是昏迷門檻：體型與健康的平均（`sub_21AC4:0x21C1F`）。
// 六個部位的血量由它按 15、15、20、20、30、30 % 分配。
func Threshold(size, health int) int { return (size + health) / 2 }

// PartShare 是六個部位分到的百分比，順序同命中部位：頭、胸、右臂、左臂、右腿、左腿。
var PartShare = [6]int{15, 15, 20, 20, 30, 30}

// Parts 依昏迷門檻算出六個部位的血量上限。
func Parts(threshold int) [6]int {
	var out [6]int
	for i, share := range PartShare {
		out[i] = share * threshold / 100
	}
	return out
}

// InitiativeKey 是一名參戰者這一回合的先攻鍵值（`sub_1E3A8:0x1EF84`…`0x1F028`）：
//
//	基準 = (等級 − 1) × 10 + 洞察
//	隊員在第二回合起先除以 2 × 回合數
//	基準 −= 基準 × 傷勢 ÷ 200，傷勢 = 50 × 累積傷害 ÷ 昏迷門檻
//	鍵值 = 擲骰 + 基準
//
// roll 是 1…100 的擲骰結果。鍵值小的先動。
func InitiativeKey(a Actor, round int, roll int) int {
	base := (a.Level-1)*LevelStep + a.Foresight
	if a.Side == SideParty && round > 0 {
		base /= 2 * round
	}
	base -= base * a.Injury() / 200
	return roll + base
}

// Injury 是傷勢比例：50 × 累積傷害 ÷ 昏迷門檻（`sub_23683`）。
func (a Actor) Injury() int {
	if a.Threshold <= 0 {
		return 0
	}
	return 50 * a.Damage / a.Threshold
}

// Down 回報這一名退出戰鬥了沒有：頭或胸被打廢就算（`sub_22866`）。
func (a Actor) Down() bool { return a.Head <= 0 || a.Chest <= 0 }

// Round 是這一場的回合狀態。
type Round struct {
	// Number 是已經開始過幾個回合（原版 `word_30D00`，排完序才加一）。
	Number int
	// Order 是這一回合的出手順序，內容是參戰者索引。
	Order []int
	// Turn 是現在輪到 Order 的第幾個。
	Turn int
}

// BeginRound 依先攻鍵值排出這一回合的順序。空位排最後。
func (s *State) BeginRound(rng Rand) {
	keys := make([]int, len(s.Actors))
	order := make([]int, 0, len(s.Actors))
	for i, a := range s.Actors {
		if a.Side == SideEmpty {
			keys[i] = EmptyKey
			continue
		}
		keys[i] = InitiativeKey(a, s.Round.Number, rng.Roll(InitiativeRoll))
	}
	for i, a := range s.Actors {
		if a.Side != SideEmpty {
			order = append(order, i)
		}
	}
	// 原版是氣泡排序，相同鍵值維持原本的先後（`sub_2616C`）。
	sort.SliceStable(order, func(x, y int) bool {
		return keys[order[x]] < keys[order[y]]
	})
	s.Round.Order = order
	s.Round.Turn = 0
	s.Round.Number++
	s.skipDown()
}

// Current 是現在輪到誰；沒有人可以動就回 −1。
func (s *State) Current() int {
	if s.Round.Turn < 0 || s.Round.Turn >= len(s.Round.Order) {
		return -1
	}
	return s.Round.Order[s.Round.Turn]
}

// EndTurn 讓順序往前走一格，回報這一回合是不是走完了。
func (s *State) EndTurn() bool {
	s.Round.Turn++
	s.skipDown()
	return s.Round.Turn >= len(s.Round.Order)
}

// AdvanceToParty 把順序推到下一名還站著的隊員。敵人那一手要由 `sub_2374C`
// 的 AI 決定做什麼，那一段還沒解出來（RE-080 末段列為待解），因此這一版
// 直接跳過敵人的手——這是**已知的差異**，不是原版行為。
func (s *State) AdvanceToParty() {
	for s.Round.Turn < len(s.Round.Order) {
		i := s.Round.Order[s.Round.Turn]
		if s.Actors[i].Side == SideParty {
			break
		}
		s.Round.Turn++
		s.skipDown()
	}
	s.Cursor = s.Current()
	if s.Cursor < 0 {
		s.Cursor = 0
	}
}

// skipDown 跳過倒下或空掉的位置（`sub_1E3A8:0x1F085` 的 `sub_22866` 檢查）。
func (s *State) skipDown() {
	for s.Round.Turn < len(s.Round.Order) {
		a := s.Actors[s.Round.Order[s.Round.Turn]]
		if a.Side != SideEmpty && !a.Down() {
			return
		}
		s.Round.Turn++
	}
}
