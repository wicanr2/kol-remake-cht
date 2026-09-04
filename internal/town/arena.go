package town

// 競技場（RE-108）。它在世界地圖 (55,13) 那一格，不在城鎮裡，但畫面用的是建築
// 那一套：標題框、對話框與圖示列。走到那一格選 DOOR 就進來。
//
// 這裡只做「誰上場」到「開打」與打完的升階；下注那一段另外一塊。

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
)

// ArenaMaxRank 是位階上限（`0x2AC5C` 比的 0x18）。到頂就不用再打。
const ArenaMaxRank = 24

// arenaStage 是競技場停在哪一步。
type arenaStage int

const (
	// arenaChoosing 是多人隊伍在選誰上場。
	arenaChoosing arenaStage = iota
	// arenaConfirm 是「確定要進去嗎」。
	arenaConfirm
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
}

// ArenaFight 是談完之後要打的那一場。
type ArenaFight struct {
	// Member 是上場的隊員索引；Level 是對手等級（實得位階 ＋ 1）。
	Member, Level int
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

// ArenaAnswer 是「確定要進去嗎」的答案。答應就回傳那一場，答否印一句就結束。
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
	fight := &ArenaFight{Member: i, Level: s.Party[i].Sheet.EarnedRank() + 1}
	s.arena = nil
	return fight
}

// LeaveArena 收掉競技場的對話，回到野外。
func (s *Session) LeaveArena() {
	s.arena = nil
	s.Inside = nil
	s.Mode = ModeMap
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
