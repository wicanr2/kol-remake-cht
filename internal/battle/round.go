package battle

import "sort"

// 回合結構（RE-080／SPEC-008）。一個回合分成兩段：**規劃**時每一名依洞察排出來的
// 順序各挑一個指令，整輪挑完問記錄 0x108「Begin next round?」；答應之後才進**執行**段，
// 依快捷重排一次、由快的先做出來。**這一支只做規劃段的順序**；執行段的鍵值在
// `execution.go`，它用的負重與時間成本各在 `Encumbrance`（`attack.go`）與
// `ActionCost`（`fatigue.go`），兩者都有原版對拍。

// PlanRoll 是規劃順序擲的骰面數（`sub_1B5A8(0x64)`）。
const PlanRoll = 100

// LevelPoints 是升一級要的攻防值總和；LevelStep 是等級在規劃鍵值裡的權重。
const (
	LevelPoints = 1000
	LevelStep   = 10
)

// EmptyKey 是空位的鍵值：排在所有人後面（`sub_1E3A8:0x1EF7B`）。
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

// PlanKey 是一名參戰者這一回合的**規劃**鍵值（`sub_1E3A8:0x1EF84`…`0x1F028`）：
//
//	基準 = (等級 − 1) × 10 + 洞察
//	隊員在這一回合重排過時（`word_30D00` ≠ 0）先除以 2 × word_30D00（RE-145）
//	基準 −= 基準 × 傷勢 ÷ 200，傷勢 = 50 × 累積傷害 ÷ 昏迷門檻
//	鍵值 = 擲骰 + 基準
//
// roll 是 1…100 的擲骰結果。**鍵值小的先規劃**——洞察低的先決定、高的後決定。
func PlanKey(a Actor, replans int, roll int) int {
	base := (a.Level-1)*LevelStep + a.Foresight
	if a.Side == SideParty && replans > 0 {
		base /= 2 * replans
	}
	base -= base * a.FatigueRatio() / 200
	return roll + base
}

// FatigueRatio 是疲勞比例：`50 × 疲勞 ÷ (力量 ＋ 健康)`（`sub_23683`）。
// 規劃鍵值、攻防分數、力量加成與敵人 AI 的撤退門檻都拿它折算。
//
// 分子是 `[0x322E + 2i]`、分母是 `[0x1C50 + 2i]`，**不是累積傷害與昏迷門檻**
// （RE-130 更正）：累積傷害另有自己的線性項（各處的 `Damage × 10`）。
// 沒有人施效果 5 的法術時疲勞是 0，這一項就整個消失。
func (a Actor) FatigueRatio() int {
	if a.Stamina <= 0 {
		return 0
	}
	return 50 * a.Fatigue / a.Stamina
}

// Down 回報這一名退出戰鬥了沒有：頭或胸被打廢就算（`sub_22866`）。
func (a Actor) Down() bool { return a.Head <= 0 || a.Chest <= 0 }

// Round 是這一場的回合狀態。這一版只走規劃段。
type Round struct {
	// Number 是已經開始過幾個回合。原版沒有對應的全域，remake 用來記錄與顯示。
	Number int
	// Replans 是原版的 `word_30D00`：每個新回合在 `0x2119B` 歸零，排完序
	// （`0x1F04C`）加一，所以只有 U-TURN 重排時才會大於 0；規劃鍵值拿它除
	// 隊員的基準（RE-145）。
	Replans int
	// Order 是這一回合的規劃順序，內容是參戰者索引。
	Order []int
	// Turn 是現在輪到 Order 的第幾個。
	Turn int
	// InspectionRoll 是目前規劃位置的檢視門檻；位置進入時擲一次，取消不重設。
	InspectionRoll     int
	inspectionPosition int
	inspectionRand     Rand
}

// BeginPlanning 開新的一回合：先跑一次法術倒數，再排出誰先決定。空位排最後。
func (s *State) BeginPlanning(rng Rand) {
	// 排順序之前先跑一次法術倒數（`sub_1E3A8:0x20F01`）：過期的把屬性重算回原值，
	// 因為規劃鍵值要用洞察，而洞察本身可能正被法術改著。
	s.WoreOff = s.WoreOff[:0]
	for i := range s.Actors {
		if s.WearsOff(i) {
			s.WoreOff = append(s.WoreOff, i)
		}
	}
	s.Round.Number++
	s.Round.Replans = 0 // `0x2119B`
	s.Replan(rng)
}

// Replan 從 `0x1EF3A` 起重排這一回合：逐名重擲規劃鍵值、排序（`sub_2616C`）、
// `word_30D00` 加一（`0x1F04C`）。「Begin next round?」選 U-TURN 或在
// 逃走格答 n 都跳回這裡（RE-143），**不經過**回合結束的法術倒數 `0x20F01`。
func (s *State) Replan(rng Rand) {
	s.InitializeMonsterRanged(rng)
	s.Round.inspectionRand = rng
	s.Round.inspectionPosition = 0
	s.Round.InspectionRoll = 0
	keys := make([]int, len(s.Actors))
	order := make([]int, 0, len(s.Actors))
	for i, a := range s.Actors {
		if a.Side == SideEmpty {
			keys[i] = EmptyKey
			continue
		}
		keys[i] = PlanKey(a, s.Round.Replans, rng.Roll(PlanRoll))
	}
	limit := len(s.Actors)
	if s.Site != nil && PartySlots+s.Site.Count < limit {
		limit = PartySlots + s.Site.Count
	}
	for i := 0; i < limit; i++ {
		order = append(order, i)
	}
	// 原版是氣泡排序，相同鍵值維持原本的先後（`sub_2616C`）。
	sort.SliceStable(order, func(x, y int) bool {
		return keys[order[x]] < keys[order[y]]
	})
	s.Round.Order = order
	s.Round.Turn = 0
	s.Round.Replans++
	s.skipDown()
}

// Current 是現在輪到誰；沒有人可以動就回 −1。
func (s *State) Current() int {
	if s.Round.Turn < 0 || s.Round.Turn >= len(s.Round.Order) {
		return -1
	}
	return s.Round.Order[s.Round.Turn]
}

// EndTurn 讓規劃順序往前走一格，回報這一輪是不是決定完了。
func (s *State) EndTurn() bool {
	s.Round.Turn++
	s.skipDown()
	return s.Round.Turn >= len(s.Round.Order)
}

// AdvanceToParty 處理途中敵人的 AI，推進至下一名還能行動的隊員。
func (s *State) AdvanceToParty() {
	s.Plan.EquipmentPanel = false
	s.skipDown()
	for s.Round.Turn < len(s.Round.Order) {
		i := s.Round.Order[s.Round.Turn]
		if s.Actors[i].Side == SideParty {
			break
		}
		// 輪到敵人就讓 AI 決定那一手（RE-081／RE-092）。原版是規劃段呼叫
		// `sub_2374C`，把指令碼與兩個選項寫進參戰者記錄，執行段照樣做出來。
		if s.Actors[i].Side == SideMonster && !s.Actors[i].Down() && s.Rand != nil {
			s.PlanEnemy(i, s.Rand)
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
		if s.Round.inspectionPosition != s.Round.Turn+1 {
			rng := s.Round.inspectionRand
			if rng == nil {
				rng = s.Rand
			}
			if rng != nil {
				s.Round.InspectionRoll = rng.Roll(100)
			}
			s.Round.inspectionPosition = s.Round.Turn + 1
		}
		a := &s.Actors[s.Round.Order[s.Round.Turn]]
		if a.Side != SideEmpty && !a.Down() {
			return
		}
		a.Style, a.Guard = 0, 0
		s.Round.Turn++
	}
}
