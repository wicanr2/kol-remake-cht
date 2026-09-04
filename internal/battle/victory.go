package battle

// 一場戰鬥怎麼結束（RE-091）。原版每走完一個回合就數兩個數——倒下的隊員與倒下的
// 敵人——再依序判斷四個條件。remake 先前只判「隊伍全倒」，把怪物全部打死不會結束。

// Verdict 是一個回合走完之後這一場的去向。（`Outcome` 這個名字已經被
// 攻擊結算佔用了，那是一次攻擊的結果。）
type Verdict int

const (
	// VerdictContinue 是還沒分勝負，排下一回合。
	VerdictContinue Verdict = iota
	// VerdictLost 是倒下的隊員等於隊伍人數。
	VerdictLost
	// VerdictWon 是倒下的敵人等於敵人數。
	VerdictWon
	// VerdictFled 是逃走。原版用兩個計數同時等於 0xFF 表示，remake 由 `Flee()` 設。
	VerdictFled
)

// 打贏、打輸與逃走各自的訊息記錄（KOLDATA，RE-091）。
const (
	DefeatedThemRecord = 0x174 // "You have defeated|F"
	EachPlayerRecord   = 0x175 // "Each player]"
	AdventureRecord    = 0x176 // "receives|N adventure points, and"
	GoldRecord         = 0x177 // "|N gold crowns!"
	PluralHaveRecord   = 0x178 // "s have"
	SingularHasRecord  = 0x179 // " has"
	DefeatedYouRecord  = 0x17A // " defeated you!"
	FleeForRecord      = 0x17B // "You flee for your]"
	OneLifeRecord      = 0x17C // "life!"
	ManyLivesRecord    = 0x17D // "lives!"
	GainedLevelRecord  = 0x17E // "you have gained a level!"
)

// GoldDice 是金冠骰面表（DGROUP `0x00D4`）。怪物記錄 +0x0F 的低三位選面數，
// 高 nibble 的低三位是骰數——每一隻怪擲「骰數 ＋ 1」次。
var GoldDice = [8]int{2, 3, 5, 7, 9, 11, 19, 99}

// Down 之外的第二個判準：原版看參戰者記錄 +0x01 是不是 2，那一格就不算數。
// remake 用 Side 表示同一件事——空格子的 Side 是 SideEmpty。

// RoundVerdict 數完兩邊倒下的人，回報這一場的去向。
//
// 原版的四個判斷依序是（`0x21016`…`0x21042`）：倒下的隊員等於隊伍人數就是打輸、
// 倒下的敵人等於敵人數就是打贏、倒下的隊員是 0 就繼續、`word_345DE` 不是 `0xFFFF`
// 就先處理升級。**順序有意義**：兩邊同時全倒時算打輸。
func (s *State) RoundVerdict() Verdict {
	partyDown, enemyDown, party, enemies := 0, 0, 0, 0
	for i := range s.Actors {
		a := s.Actors[i]
		switch a.Side {
		case SideParty:
			party++
			if a.Down() {
				partyDown++
			}
		case SideMonster:
			enemies++
			if a.Down() {
				enemyDown++
			}
		}
	}
	switch {
	case party > 0 && partyDown == party:
		return VerdictLost
	case enemies > 0 && enemyDown == enemies:
		return VerdictWon
	default:
		return VerdictContinue
	}
}

// GoldReward 是打贏之後每個人分到的金冠（`0x21360`…`0x21413`）。
//
// 每一隻怪擲「骰數 ＋ 1」次面數 `sides` 的骰，全部累加起來再除以隊伍人數；
// 每人不到 2 枚就一枚都不給。回傳的是**每人**的金額，不是總額。
func GoldReward(rolls, sides, count, party int, rng Rand) int {
	if party <= 0 || sides <= 0 {
		return 0
	}
	total := 0
	for i := 0; i < count; i++ {
		for r := 0; r <= rolls; r++ {
			total += rng.Roll(sides)
		}
	}
	each := total / party
	if each < 2 {
		return 0
	}
	return each
}

// PluralSuffix 是敵人數大於 1 時補在名字後面的英文複數。名字以 `S` 結尾補 `ES`，
// 否則補 `S`（`0x211F5`；打輸那一條的 `0x2150D` 是同一套）。
//
// 中文沒有這個變化，譯文把三個記錄都處理掉了，所以只有英文模式會用到。
func PluralSuffix(name string, count int) string {
	if count <= 1 || name == "" {
		return ""
	}
	if c := name[len(name)-1]; c == 'S' || c == 's' {
		return "ES"
	}
	return "S"
}

// VictoryRecords 是打贏那一句的記錄（`0x211DB`）。記錄 0x174 本身帶 `|F`，
// 名字由代換展開；之後還要補英文複數（PluralSuffix）與 `!`。
func VictoryRecords() []int { return []int{DefeatedThemRecord} }

// DefeatRecords 是打輸那一句，拆成名字前後兩段（`0x214DE`）：
// `The]` ＋ 名字 ＋ 複數 ＋ `s have`／` has` ＋ ` defeated you!`。
//
// 名字不是記錄，所以由呼叫端插在中間——原版也是分開印的。
func DefeatRecords(enemies int) (before, after []int) {
	before = []int{TheRecord}
	if enemies > 1 {
		return before, []int{PluralHaveRecord, DefeatedYouRecord}
	}
	return before, []int{SingularHasRecord, DefeatedYouRecord}
}

// RewardRecords 是打贏之後報獎勵那一句（`0x212D2`、`0x2148D`）。隊伍人數大於 1
// 才印 `Each player]`，單人跳過。
//
// **只有金冠。**冒險點數那一句（記錄 0x176）的數字要 `sub_26C5F(0x17)` 才算得出來，
// 那一支還沒解讀（RE-091），寧可不報也不報錯的數。
func RewardRecords(party int) []int {
	if party > 1 {
		return []int{EachPlayerRecord, GoldRecord}
	}
	return []int{GoldRecord}
}

// EnemyCount 是這一場放了幾隻敵人；PartyCount 是隊伍有幾個人。獎勵的兩個除數
// 都用它們（RE-091 的 `word_30C6C` 與 `word_2DF2A`）。
func (s *State) EnemyCount() int { return s.countSide(SideMonster) }

// PartyCount 是隊伍人數。
func (s *State) PartyCount() int { return s.countSide(SideParty) }

func (s *State) countSide(side Side) int {
	n := 0
	for i := range s.Actors {
		if s.Actors[i].Side == side {
			n++
		}
	}
	return n
}
