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
	// VerdictFled 是逃走，由 `Flee()`（戰場邊上）或 `BarFlee()`（逃走格）設。
	// 原版沒有對應的單一值：逃走格把兩個倒下計數設成 0xFF，戰場邊上那一條
	// 帶著本回合的計數出去（RE-146）。
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

// FleeRecords 是逃走之後那一句（`0x215B9`…`0x215EE`）：0x17B 接 0x17C 或 0x17D，
// 單複數看隊伍人數（`word_2DF2A`）。
func FleeRecords(party int) []int {
	tail := ManyLivesRecord
	if party == 1 {
		tail = OneLifeRecord
	}
	return []int{FleeForRecord, tail}
}

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
// 擲「骰數 ＋ 1」次面數 `sides` 的骰、重複 `waves` 輪，全部累加起來再除以隊伍人數；
// 每人不到 2 枚就一枚都不給。回傳的是**每人**的金額，不是總額。
//
// `waves` 是 `word_3B26A`，也就是**敵人的等級**：隨機遭遇在進戰鬥前把等級寫進去
// （`sub_2A4E5:0x2A7D1`），任務地點同理，而 `sub_1E3A8(0)` 會跳過那一整段預設值
// （`0x1E3B1` 的閘門），所以那個 1 落不到玩家身上（RE-097）。等級越高金冠越多。
//
// 骰式來自**第一隻怪**的記錄 +0x0F，同場其他怪的那一格原版沒有讀，所以敵人多寡
// 不影響金額（RE-096）。
func GoldReward(rolls, sides, waves, party int, rng Rand) int {
	if party <= 0 || sides <= 0 {
		return 0
	}
	total := 0
	for i := 0; i < waves; i++ {
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

// GoldRoll 把怪物記錄 +0x0F 拆成骰數與骰面（RE-091）。
func GoldRoll(raw int) (rolls, sides int) {
	return (raw / 16) & 7, GoldDice[raw&7]
}

// AdventureReward 是打贏之後每個人分到的冒險點數（`0x21270`…`0x212CF`）。
//
// 把每一個參戰敵人的攻擊技能與招架技能加起來，先除 2 再除隊伍人數。兩件事值得記住：
// **不看敵人死了沒**，站上場就算；用的是怪物記錄裡的原值，不是折算過等級的
// `Attack`／`Parry`。
func (s *State) AdventureReward() int {
	party := s.PartyCount()
	if party <= 0 {
		return 0
	}
	total := 0
	for i := range s.Actors {
		if s.Actors[i].Side == SideMonster {
			total += s.Actors[i].BaseAttack + s.Actors[i].BaseParry
		}
	}
	return total / 2 / party
}

// GoldAward 是打贏之後每個人分到的金冠。骰式取自場上第一隻敵人，輪數是這一場的
// 怪物等級（`word_3B26A`）——不是參戰者記錄那一格折算過的等級。
func (s *State) GoldAward(rng Rand) int {
	for i := range s.Actors {
		if s.Actors[i].Side == SideMonster {
			rolls, sides := GoldRoll(s.Actors[i].GoldRoll)
			return GoldReward(rolls, sides, s.MonsterLevel, s.PartyCount(), rng)
		}
	}
	return 0
}

// PluralSuffix 是**打贏**那一句補在名字後面的英文複數（`0x211F5`）：名字以 `S`
// 結尾補 `ES`、否則補 `S`，最後再接一個 `!`。
//
// 中文沒有這個變化，譯文把三個記錄都處理掉了，所以只有英文模式會用到。
// PluralBaseName 回傳英文複數拼接前的名稱底字（RE-256）。
// 原版先退一個欄位，把末尾 S 覆成 E，再寫 S；原始名稱保持。
func PluralBaseName(name string, count int) string {
	if count > 1 && name != "" && endsWithS(name) {
		return name[:len(name)-1]
	}
	return name
}

func PluralSuffix(name string, count int) string {
	if count <= 1 || name == "" {
		return ""
	}
	if endsWithS(name) {
		return "ES"
	}
	return "S"
}

// DefeatSuffix 是**打輸**那一句的複數（`0x2150D`）。它和打贏那一句不一樣：接在
// 後面的記錄 0x178 是 `s have`，本身已經帶了 s，所以名字結尾是 `S` 時只補一個 `E`
// （湊成 `…ES have`），不是 `S` 就什麼都不補（`…S have`）。
//
// 兩句用同一套後綴會印出 `THE BINDERAKSS HAVE DEFEATED YOU!`——實跑抓到的。
func DefeatSuffix(name string, count int) string {
	if count <= 1 || name == "" || !endsWithS(name) {
		return ""
	}
	return "E"
}

func endsWithS(name string) bool {
	c := name[len(name)-1]
	return c == 'S' || c == 's'
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

// RewardRecords 是打贏之後報獎勵那幾句（`0x212D2`…`0x21495`），拆成前後兩段。
//
// 隊伍人數大於 1 才印 `Each player]`；等於 1 時原版改逐一印那一位的名字，名字不是
// 記錄，所以由呼叫端插在 `before` 的前面。
//
// 拆兩段是因為兩句各有一個 `|N`，而它們讀的是同一個變數（`word_30D02`）——原版在
// 每一句印出去之前各設一次值。呼叫端要先用冒險點數展開 `before`、再用金冠展開
// `after`，一次展開整串會讓兩個數字變成同一個。
func RewardRecords(party int) (before, after []int) {
	if party > 1 {
		return []int{EachPlayerRecord, AdventureRecord}, []int{GoldRecord}
	}
	return []int{AdventureRecord}, []int{GoldRecord}
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

// 任務旗標（角色記錄 `0x1E2` 起，一個 byte 塞四個 2-bit 欄位，RE-079／RE-101）。
const (
	// QuestFlagTaken 是「接了這個任務」、QuestFlagDone 是「完成了」。
	QuestFlagTaken = 1
	QuestFlagDone  = 2
	// QuestFlagBytes 是一組旗標有幾個 byte（24 個地點 ÷ 4）。
	QuestFlagBytes = 6
)

// QuestDivisors 是任務編號的低兩位對應的除數（DGROUP `0x007C`）：那一個 byte 裡
// 第 n 個 2-bit 欄位要先除以它再取低兩位。
var QuestDivisors = [4]int{1, 4, 16, 64}

// QuestTaken 是「這名角色接了第 site 個任務嗎」（`sub_2B0FF:0x2B108`）。
func QuestTaken(flags []int, site int) bool {
	v, div, ok := questField(flags, site)
	return ok && (v/div)&3 == QuestFlagTaken
}

// MarkQuestDone 把一名角色的任務旗標標成「完成」（`sub_28F5D`）。回傳有沒有真的
// 改到——已經是完成的不再動。
//
// **沒接這個任務的人也會被標**：原版只跳過「已經是 2」的，值為 0 的照樣寫成 2
// （`0x28FB9` 只比 bit 1）。一起打完的整隊都算數，「接了」那一位元只是進得了任務
// 地點的條件（`sub_2B0FF:0x2B108`）。
//
// **只有任務戰鬥會做這件事**：原版看 `word_34968`，隨機遭遇把它設成 0
// （`sub_2A4E5:0x2A82B`）、任務地點設 1（`sub_2B0FF:0x2B1AD`）。remake 用
// `Site.Index >= 0` 表示同一件事。
func MarkQuestDone(flags []int, site int) bool {
	v, div, ok := questField(flags, site)
	if !ok || (v/div)&QuestFlagDone == QuestFlagDone {
		return false
	}
	// 原版是「先 and 掉那一位、再 or 上它的兩倍」（`0x28FCD`）——等於把 2-bit
	// 欄位從 1 換成 2。
	flags[site/4] = (v &^ div) | (div * 2)
	return true
}

func questField(flags []int, site int) (value, divisor int, ok bool) {
	if site < 0 || site >= QuestFlagBytes*4 {
		return 0, 0, false
	}
	i := site / 4
	if i >= len(flags) {
		return 0, 0, false
	}
	return flags[i], QuestDivisors[site&3], true
}

// markQuestDone 把隊伍裡的每個人都標成完成（`sub_28F5D`），回傳有幾個人被標到
// （已經完成的不算）。隨機遭遇（`Site.Index` 為負）什麼都不做。
func (s *State) markQuestDone() int {
	if s.Site == nil || s.Site.Index < 0 {
		return 0
	}
	n := 0
	for i := range s.Party {
		if MarkQuestDone(s.Party[i].Quests, s.Site.Index) {
			n++
		}
	}
	return n
}
