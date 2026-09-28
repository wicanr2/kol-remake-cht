package town

// 競技場（RE-108）。它在世界地圖 (55,13) 那一格，不在城鎮裡，但畫面用的是建築
// 那一套：標題框、對話框與圖示列。走到那一格選 DOOR 就進來。
//
// 這裡只做「誰上場」到「開打」與打完的升階；下注那一段另外一塊。

// 競技場與賭場的肖像（`sub_2B454`，RE-014／RE-108）：`comb%s` 加模式後綴 ＝ COMBPICS，
// 貼在欄 [1, 12)、掃描列 [64, 128)，與建築肖像同一格。進競技場（`0x2ABD2`）畫第 6 張，
// 換到賭場、印完記錄 0xF0 之後（`0x2ADCB`）畫第 5 張。
const (
	ArenaPortraitArchive = "COMBPICS"
	ArenaPortraitTester  = 6
	ArenaPortraitBetter  = 5
)

// 競技場用到的 KOLDATA 記錄。
const (
	// ArenaTitleRecord 是 `The Arena@The Tester`。
	ArenaTitleRecord = 0xEB
	// ArenaBoothRecord 是賭場的 `A Gambling Booth@The Bet Taker`。
	ArenaBoothRecord = 0xEC
	// ArenaWhichRecord 是 `Which one wishs to fight?`。
	ArenaWhichRecord = 0xED
	// ArenaSureRecord 是 `Are you sure you wish to enter?`。
	ArenaSureRecord = 0xEE
	// ArenaCourageRecord 是 `Return when you have gained courage!`。
	ArenaCourageRecord = 0xEF
	// ArenaNoneRecord／ArenaNotYouRecord 是「沒有人夠格」的兩句（多人／單人）。
	ArenaNoneRecord   = 0xFA
	ArenaNotYouRecord = 0xF9
	// ArenaLevelRecord 是打贏之後的 `you have gained a level!`。
	ArenaLevelRecord = 0x17E
	// 下注那一段的記錄（RE-108）。
	// ArenaBetAskOne／ArenaBetAskMany 是單人／多人的「要不要下注」。
	ArenaBetAskOne  = 0xF1
	ArenaBetAskMany = 0xF2
	// ArenaBetWhoRecord 是「哪一位要押」。
	ArenaBetWhoRecord = 0xF3
	// ArenaBetSelfRecord 是押到上場那一位；ArenaBetAgainRecord 是同一位押第二次。
	ArenaBetSelfRecord  = 0xF4
	ArenaBetAgainRecord = 0xF5
	// ArenaBetHowMuchRecord 是「多少錢」；ArenaBetFriendRecord 是「押他會贏嗎」；
	// ArenaBetGoodRecord 是押他會輸時那一句。
	ArenaBetHowMuchRecord = 0xF6
	ArenaBetFriendRecord  = 0xF7
	ArenaBetGoodRecord    = 0xF8
	// ArenaBetTunnelRecord 是進賭場那一段敘述。
	ArenaBetTunnelRecord = 0xF0
	// ArenaBetWinRecord／ArenaBetLoseRecord 是戰後結算那兩句。
	ArenaBetWinRecord  = 0x17F
	ArenaBetLoseRecord = 0x180
)

// ArenaBetMaxDigits 是金額讀幾位數（`sub_1B5C8(0x2E3A, 2)`），所以上限是 99。
const ArenaBetMaxDigits = 2

// ArenaMaxRank 是位階上限（`0x2AC5C` 比的 0x18）。到頂就不用再打。
const ArenaMaxRank = 24

// arenaStage 是競技場停在哪一步。
type arenaStage int

const (
	// arenaChoosing 是多人隊伍在選誰上場。
	arenaChoosing arenaStage = iota
	// arenaConfirm 是「確定要進去嗎」。
	arenaConfirm
	// arenaBetAsk 是賭場的「要不要下注」。
	arenaBetAsk
	// arenaBetWho 是「哪一位要押」的清單。
	arenaBetWho
	// arenaBetAmount 是「多少錢」，讀兩位數。
	arenaBetAmount
	// arenaBetSide 是「押他會贏嗎」。
	arenaBetSide
	// arenaDone 是這一次談完了（沒人夠格、或答否）。
	arenaDone
)

// arenaState 是一次競技場對話。
type arenaState struct {
	stage arenaStage
	// candidates 是夠格上場的隊員索引。
	candidates []int
	// chosen 是選中的那一名；−1 表示還沒選。
	chosen int
	// bets 是每一名押的金額：正值押他會贏、負值押他會輸、0 是還沒押。
	bets []int
	// better 是這一輪正在押的那一名；digits 是打到一半的金額。
	better int
	digits []rune
}

// ArenaBets 是這一場的賭注，索引與隊伍相同。
func (s *Session) ArenaBets() []int {
	if s.arena == nil {
		return nil
	}
	return s.arena.bets
}

// ArenaFight 是談完之後要打的那一場。
type ArenaFight struct {
	// Member 是上場的隊員索引；Level 是對手等級（實得位階 ＋ 1）。
	Member, Level int
	// Bets 是每一名押的金額：正值押他會贏、負值押他會輸、0 是沒押。
	Bets []int
}

// PendingFight 是談完之後等著上層開打的那一場；沒有就是 nil。
func (s *Session) PendingFight() *ArenaFight { return s.pendingFight }

// ClearPendingFight 在上層把那一場開起來之後清掉。
func (s *Session) ClearPendingFight() { s.pendingFight = nil }

// ArenaCandidates 是夠格上場的隊員（`0x2AC0F`…`0x2AC77`）：帳面位階與實得位階
// 一樣就不用升，帳面位階到 24 也不用再打。
func ArenaCandidates(party []Member) []int {
	var out []int
	for i, m := range party {
		if m.Sheet == nil {
			continue
		}
		if m.Sheet.Rank == m.Sheet.EarnedRank() || m.Sheet.Rank >= ArenaMaxRank {
			continue
		}
		out = append(out, i)
	}
	return out
}

// EnterArena 是在競技場那一格選了 DOOR。回傳有沒有真的開起對話——
// 沒有人夠格時只講一句就結束。
func (s *Session) EnterArena(inside *Inn) bool {
	if inside == nil {
		return false
	}
	s.Inside = inside
	s.Mode = ModeInside
	s.arena = &arenaState{chosen: -1}
	s.Inside.Portrait.Archive, s.Inside.Portrait.Index = ArenaPortraitArchive, ArenaPortraitTester
	s.Inside.Title = s.uiLines(ArenaTitleRecord)
	cands := ArenaCandidates(s.Party)
	if len(cands) == 0 {
		record := ArenaNotYouRecord
		if len(s.Party) > 1 {
			record = ArenaNoneRecord
		}
		s.arena.stage = arenaDone
		s.say(nonEmpty(s.uiLines(record)), nil)
		return true
	}
	s.arena.candidates = cands
	// 判斷依據是**隊伍人數**而不是有幾個人夠格：一個人就直接用那一名，
	// 多人就印記錄 0xED 讓玩家選（`0x2AC8A`）。
	if len(s.Party) == 1 {
		s.arena.chosen = cands[0]
		s.askArenaSure()
		return true
	}
	s.arena.stage = arenaChoosing
	s.Inside.List = s.arenaList()
	s.Inside.Dialog = nonEmpty(s.uiLines(ArenaWhichRecord))
	s.Inside.DialogZh = nil
	s.setInsideIcons([]int{IconYes, IconUTurn})
	return true
}

// arenaList 是「誰要上場」那張清單，內容是隊伍全員——原版讓玩家在整隊裡選，
// 選到不夠格的印 `You are not ready`（`0x2ACC0`）再問一次。
func (s *Session) arenaList() *InnList {
	rows := make([]string, 0, len(s.Party))
	for _, m := range s.Party {
		rows = append(rows, m.Name)
	}
	return &InnList{Rows: rows}
}

// InArena 回報現在是不是在競技場的對話裡。
func (s *Session) InArena() bool { return s.arena != nil }

// MoveArenaList 移動「誰要上場」那張清單的反白。
func (s *Session) MoveArenaList(delta int) {
	if s.arena == nil || s.Inside == nil || s.Inside.List == nil {
		return
	}
	n := len(s.Inside.List.Rows)
	if n == 0 {
		return
	}
	s.Inside.List.Selected = ((s.Inside.List.Selected+delta)%n + n) % n
}

// ArenaChoose 是清單上按下確定。夠格就問「確定要進去嗎」，不夠格就再問一次。
func (s *Session) ArenaChoose() {
	if s.arena == nil || s.Inside == nil || s.Inside.List == nil {
		return
	}
	i := s.Inside.List.Selected
	if !s.arenaReady(i) {
		s.Inside.Dialog = nonEmpty(s.uiLines(ArenaNotYouRecord))
		s.Inside.DialogZh = nil
		return
	}
	s.arena.chosen = i
	s.Inside.List = nil
	s.askArenaSure()
}

func (s *Session) arenaReady(i int) bool {
	for _, c := range s.arena.candidates {
		if c == i {
			return true
		}
	}
	return false
}

func (s *Session) askArenaSure() {
	s.arena.stage = arenaConfirm
	s.Inside.List = nil
	s.Inside.Dialog = nonEmpty(s.uiLines(ArenaSureRecord))
	s.Inside.DialogZh = nil
	s.setInsideIcons([]int{IconYes, IconNo})
}

// ArenaAnswer 是「確定要進去嗎」的答案。答否印一句就結束；答應之後多人隊伍先換到
// 賭場（RE-108），一個人就直接開打。
func (s *Session) ArenaAnswer(yes bool) *ArenaFight {
	if s.arena == nil || s.arena.stage != arenaConfirm {
		return nil
	}
	i := s.arena.chosen
	if !yes || i < 0 || i >= len(s.Party) || s.Party[i].Sheet == nil {
		s.arena.stage = arenaDone
		s.say(nonEmpty(s.uiLines(ArenaCourageRecord)), nil)
		return nil
	}
	if len(s.Party) > 1 {
		s.openBooth()
		return nil
	}
	return s.startFight()
}

// startFight 收掉對話，回傳要打的那一場。
func (s *Session) startFight() *ArenaFight {
	i := s.arena.chosen
	fight := &ArenaFight{Member: i, Level: s.Party[i].Sheet.EarnedRank() + 1,
		Bets: append([]int(nil), s.arena.bets...)}
	s.arena = nil
	return fight
}

// openBooth 換到賭場：標題改成 `A Gambling Booth`，先講一段敘述再問要不要下注。
func (s *Session) openBooth() {
	s.arena.stage = arenaBetAsk
	s.arena.bets = make([]int, len(s.Party))
	s.arena.better = -1
	s.Inside.Title = s.uiLines(ArenaBoothRecord)
	s.say(nonEmpty(s.uiLines(ArenaBetTunnelRecord)), nil)
	s.Inside.Portrait.Index = ArenaPortraitBetter
}

// askBet 問「要不要下注」。單人隊伍走不到這裡，記錄仍照原版分兩筆。
func (s *Session) askBet() {
	s.arena.stage = arenaBetAsk
	s.arena.better = -1
	record := ArenaBetAskMany
	if len(s.Party) == 1 {
		record = ArenaBetAskOne
	}
	s.Inside.List = nil
	s.Inside.Dialog = nonEmpty(s.uiLines(record))
	s.Inside.DialogZh = nil
	s.setInsideIcons([]int{IconYes, IconNo})
}

// BetAnswer 是「要不要下注」的答案。答否就開打，答是就問哪一位要押。
func (s *Session) BetAnswer(yes bool) *ArenaFight {
	if s.arena == nil || s.arena.stage != arenaBetAsk {
		return nil
	}
	if !yes {
		return s.startFight()
	}
	s.arena.stage = arenaBetWho
	s.Inside.List = s.arenaList()
	s.Inside.Dialog = nonEmpty(s.uiLines(ArenaBetWhoRecord))
	s.Inside.DialogZh = nil
	s.setInsideIcons([]int{IconYes, IconUTurn})
	return nil
}

// BetChoose 是「哪一位要押」的清單上按下確定。押到上場那一位或押過第二次都會
// 被擋下來，講一句再問一次。
func (s *Session) BetChoose() {
	if s.arena == nil || s.arena.stage != arenaBetWho || s.Inside.List == nil {
		return
	}
	i := s.Inside.List.Selected
	switch {
	case i == s.arena.chosen:
		s.arena.stage = arenaBetAsk
		s.Inside.List = nil
		s.say(nonEmpty(s.uiLines(ArenaBetSelfRecord)), nil)
		return
	case i < len(s.arena.bets) && s.arena.bets[i] != 0:
		s.arena.stage = arenaBetAsk
		s.Inside.List = nil
		s.say(nonEmpty(s.uiLines(ArenaBetAgainRecord)), nil)
		return
	}
	s.arena.better = i
	s.arena.digits = nil
	s.arena.stage = arenaBetAmount
	s.Inside.List = nil
	s.refreshBet()
}

// TypeBet 收一個數字。只收兩位（`sub_1B5C8(0x2E3A, 2)`），所以上限是 99。
func (s *Session) TypeBet(r rune) {
	if s.arena == nil || s.arena.stage != arenaBetAmount {
		return
	}
	if r < '0' || r > '9' || len(s.arena.digits) >= ArenaBetMaxDigits {
		return
	}
	s.arena.digits = append(s.arena.digits, r)
	s.refreshBet()
}

// BackspaceBet 退一格。
func (s *Session) BackspaceBet() {
	if s.arena == nil || s.arena.stage != arenaBetAmount || len(s.arena.digits) == 0 {
		return
	}
	s.arena.digits = s.arena.digits[:len(s.arena.digits)-1]
	s.refreshBet()
}

// SubmitBet 送出金額。第一位是零就取消，超過那一名的錢包就講一句錢不夠，
// 兩種都回去問「要不要下注」。
func (s *Session) SubmitBet() {
	if s.arena == nil || s.arena.stage != arenaBetAmount {
		return
	}
	amount := 0
	for _, r := range s.arena.digits {
		amount = amount*10 + int(r-'0')
	}
	if len(s.arena.digits) == 0 || s.arena.digits[0] == '0' {
		s.askBet()
		return
	}
	i := s.arena.better
	if i < 0 || i >= len(s.Party) || amount > s.Party[i].Gold {
		s.arena.stage = arenaBetAsk
		s.say(nonEmpty(s.uiLines(NotEnoughRecord)), nil)
		return
	}
	s.arena.stage = arenaBetSide
	s.Inside.Dialog = nonEmpty(s.uiLines(ArenaBetFriendRecord))
	s.Inside.DialogZh = nil
	s.setInsideIcons([]int{IconYes, IconNo})
}

// BetSide 是「押他會贏嗎」的答案。押他會輸就把金額取負（`0x2AFAF` 的 `neg`），
// 並多講一句。
func (s *Session) BetSide(win bool) {
	if s.arena == nil || s.arena.stage != arenaBetSide {
		return
	}
	amount := 0
	for _, r := range s.arena.digits {
		amount = amount*10 + int(r-'0')
	}
	if !win {
		amount = -amount
	}
	if i := s.arena.better; i >= 0 && i < len(s.arena.bets) {
		s.arena.bets[i] = amount
	}
	s.arena.digits = nil
	if win {
		s.askBet()
		return
	}
	s.arena.stage = arenaBetAsk
	s.say(nonEmpty(s.uiLines(ArenaBetGoodRecord)), nil)
}

// refreshBet 把「多少錢」那一問與打到一半的數字放進對話框。金額那格與 MOUTH、
// 訓練共用原版的 `sub_1B5C8`（呼叫點 `0x1353E`），所以筆尖那格畫的是同一條橫槓
// （見 TextCaret）——原版沒有取到競技場的畫格，這一條是從共用同一支函式推出來的。
func (s *Session) refreshBet() {
	lines := nonEmpty(s.uiLines(ArenaBetHowMuchRecord))
	lines = append(lines, string(s.arena.digits)+TextCaret)
	s.Inside.Dialog = lines
	s.Inside.DialogZh = nil
	s.setInsideIcons(nil)
}

// BettingAmount 回報現在是不是在打金額。
func (s *Session) BettingAmount() bool {
	return s.arena != nil && s.arena.stage == arenaBetAmount
}

// LeaveArena 收掉競技場的對話，回到野外。
func (s *Session) LeaveArena() {
	s.arena = nil
	s.Inside = nil
	s.Mode = ModeMap
}

// SettleBets 是戰後的賭金結算（`0x216C9`…`0x2175B`，RE-103）：每一格非零的賭注
// 加進那一名的錢包，再依正負印 `wins a bet!` 或 `loses a bet!`。輸掉那一場時
// 整排先取負再結算——押他會輸（負值）的人因此變成賭中。
//
// **那兩句怎麼與名字接起來沒有對拍**：原版的記錄本身只有 `wins a bet!`，
// 前面接誰、怎麼接沒有讀出來，這一版照最直白的「名字 ＋ 那一句」組。
func (s *Session) SettleBets(bets []int, won bool) []string {
	var out []string
	for i, bet := range bets {
		if bet == 0 || i >= len(s.Party) {
			continue
		}
		if !won {
			bet = -bet
		}
		s.Party[i].Gold += bet
		record := ArenaBetWinRecord
		if bet < 0 {
			record = ArenaBetLoseRecord
			if s.Party[i].Gold < 0 {
				s.Party[i].Gold = 0
			}
		}
		out = append(out, s.Party[i].Name+" "+s.uiRecord(record))
	}
	return out
}

// ArenaWin 是那一場打贏之後的升階（`0x21611`）：帳面位階加一，一次只加一級，
// 所以帳面追上實得要打好幾場。回傳升階那一句。
func (s *Session) ArenaWin(member int) []string {
	if member < 0 || member >= len(s.Party) || s.Party[member].Sheet == nil {
		return nil
	}
	s.Party[member].Sheet.Rank++
	return nonEmpty(s.uiLines(ArenaLevelRecord))
}
