package town

// 進門之後誰來開口、建築喜不喜歡這個人（#32，RE-166）。
//
// 原版讀完進門敘述後不是直接招呼，而是先看隊伍與建築的好惡（`sub_11A93:0x11D0B`…
// `0x12004`）：
//
//  1. 隊伍不只一人，而且每一位都符合建築討厭的種族或職業：只印 `"Get out, scum!"`，
//     按鍵之後不說告別就走出建築。
//  2. 隊伍不只一人：印訊息 16 加 `"Which one needs help?"`，圖示列只剩 DOOR，
//     按 1…6 選第 n 位在籍的隊員開口；選 DOOR 走出去（多人隊伍會說告別）。
//  3. 開口的人被討厭，或這棟建築是 SCROLL 而他屬於別的教派：印訊息 14，按鍵之後
//     單人隊伍走出建築，多人隊伍回到第 2 步。
//  4. 開口的人有傷是既有的規則（RE-072）；沒有傷而被欣賞：印訊息 15；否則照聲望、
//     性別與會不會法術挑訊息（RE-067）。
//
// 建築內一般圖示列停著的時候按 1…6 一樣換發言者，並從第 3 步重新判斷
// （`sub_11A93:0x1222F` 跳回 `0x11E8E`）。

// 招呼流程用到的訊息與記錄編號。
const (
	// HostileMessage／FavoredMessage 是建築訊息 14 與 15（`sub_14234` 的位移 0x400、0x440）。
	HostileMessage = 14
	FavoredMessage = 15
	// WhoRecord 是 `"Which one needs help?"`，ExpelRecord 是 `"Get out, scum!"`。
	WhoRecord   = 0x55
	ExpelRecord = 0x5A
	// NameRecords 是 `sub_14189` 掃的名稱表長度：KOLDATA 記錄 0…38。
	NameRecords = 39
	// NameNone 是 `sub_14189` 找不到名稱時回的 0x7F。
	NameNone = 0x7F
	// orderMatchHex 是 `sub_1429C` 比對教派的八個 byte，換成十六進位是 16 個字元。
	orderMatchHex = 16
)

// greetWait 記的是「讀完這一句之後要做什麼」；greetNone 是沒有在等。
type greetWait int

const (
	greetNone  greetWait = iota
	greetLeave           // 不說告別直接走出建築
	greetReask           // 回到「誰要開口」
)

// CanonicalRecord 是 `sub_14307`：名稱記錄號依隊員的種族與性別換成他自己那一套的號碼。
// 0…11 與 0x7F 原樣回傳；之後矮人與凱爾登（種族 ≥ 2）與女性精靈用低六號的那一筆。
func CanonicalRecord(record, race, gender int) int {
	if record == NameNone || record <= 0x0B {
		return record
	}
	if race >= 2 || (race == 1 && gender != 0) {
		return record - ProfessionOffset
	}
	return record
}

// GreetRule 是一棟建築對隊員的好惡與教派要求（`sub_11A93:0x11D0B`…`0x12004`），
// 只含判斷用的欄位，讓對拍與 Session 共用同一份規則。
type GreetRule struct {
	// Dislike 與 Like 是 KOLDATA 名稱記錄號（0…38），找不到名稱是 NameNone。
	Dislike, Like int
	// Scroll 為真時只招呼自己教派的人；Order 是教派記錄前八個 byte 的十六進位。
	Scroll bool
	Order  string
}

// RecordMatches 是建築名稱與隊員的比對：換算後的記錄號等於角色 `+0x24` 的種族，
// 或等於 `+0x1C` 的職業加 6（`sub_11A93:0x11D2F`）。
func RecordMatches(record int, m Member) bool {
	if m.Sheet == nil {
		return false
	}
	gender := 0
	if m.Female {
		gender = 1
	}
	record = CanonicalRecord(record, m.Sheet.Race, gender)
	return record == m.Sheet.Race || record == m.Sheet.Profession+ProfessionOffset
}

// Dislikes 回報建築討厭這位隊員的種族或職業。
func (g GreetRule) Dislikes(m Member) bool { return RecordMatches(g.Dislike, m) }

// Likes 回報建築欣賞這位隊員的種族或職業。
func (g GreetRule) Likes(m Member) bool { return RecordMatches(g.Like, m) }

// WrongOrder 是 `sub_1429C`：SCROLL 建築只招呼自己教派的人。沒有教派的隊員不算數，
// 有教派卻與這棟不同就被拒。教派比的是記錄前八個 byte。
func (g GreetRule) WrongOrder(m Member) bool {
	if !g.Scroll || m.Order == nil {
		return false
	}
	mine := m.Order.Record
	if len(mine) < orderMatchHex || len(g.Order) < orderMatchHex {
		return mine != g.Order
	}
	return mine[:orderMatchHex] != g.Order[:orderMatchHex]
}

// Refuses 回報這位隊員開口會被拒絕：被討厭，或屬於別的教派。
func (g GreetRule) Refuses(m Member) bool { return g.Dislikes(m) || g.WrongOrder(m) }

// Expels 回報全隊都被趕出去：隊伍不只一人，而且每一位都被討厭（`sub_11A93:0x11DE0`
// 要求被討厭的人數不為零、其他人數為零；教派不算在內）。
func (g GreetRule) Expels(party []Member) bool {
	if len(party) <= 1 {
		return false
	}
	for _, m := range party {
		if !g.Dislikes(m) {
			return false
		}
	}
	return true
}

// nameRecord 是 `sub_14189`：把建築的名稱在 KOLDATA 記錄 0…38 裡找同名的那一筆。
// 原版逐字比對解開後的名字，找不到回 0x7F。這裡用原文（不是譯文）比。
func (s *Session) nameRecord(name string) int {
	if s.UIText == nil || name == "" {
		return NameNone
	}
	for i := 0; i < NameRecords; i++ {
		if s.UIText.Record(BankUI, i) == name {
			return i
		}
	}
	return NameNone
}

// rule 把目前這棟建築的好惡換成規則；不在建築裡就沒有任何要求。
func (s *Session) rule() GreetRule {
	g := GreetRule{Dislike: NameNone, Like: NameNone}
	if s.door == nil {
		return g
	}
	g.Dislike, g.Like = s.nameRecord(s.door.Dislikes), s.nameRecord(s.door.Likes)
	if s.door.IsScroll() && s.door.Order != nil {
		g.Scroll, g.Order = true, s.door.Order.Record
	}
	return g
}

// nameMatches 回報建築的那個名稱是不是這位隊員的種族或職業。
func (s *Session) nameMatches(name string, m Member) bool {
	return RecordMatches(s.nameRecord(name), m)
}

// startGreeting 是讀完進門敘述之後的第一步。
func (s *Session) startGreeting() {
	s.greeted = true
	s.greetWait, s.askingWho = greetNone, false
	if len(s.Party) > 1 {
		if s.rule().Expels(s.Party) {
			s.expel()
			return
		}
		s.askWho()
		return
	}
	s.speakerGreeting()
}

// expel 是全隊都被討厭：一句話之後不說告別就離開。
func (s *Session) expel() {
	s.greetWait = greetLeave
	s.Inside.Dialog = nonEmpty(s.uiLines(ExpelRecord))
	s.Inside.DialogZh = nil
	s.setInsideIcons([]int{IconSpacebar})
}

// askWho 問誰要開口。這一刻沒有任何人是發言者（`word_39A52` 是 0xFFFF），
// 隊伍列的白框全部收掉，圖示列只剩 DOOR。
func (s *Session) askWho() {
	lines := append([]string(nil), s.message(GreetPartyMessage)...)
	if ask := s.uiRecord(WhoRecord); ask != "" {
		lines = append(lines, ask)
	}
	s.Inside.Dialog = lines
	s.Inside.DialogZh = nil
	s.Inside.Party = append([]Member(nil), s.Party...)
	for i := range s.Inside.Party {
		s.Inside.Party[i].Speaker = false
	}
	s.askingWho = true
	s.setInsideIcons([]int{IconDoor})
}

// speakerGreeting 是選定發言者之後的判斷：被拒就印訊息 14，否則是他的招呼語。
func (s *Session) speakerGreeting() {
	s.askingWho = false
	if i := s.speakerIndex(); i >= 0 {
		s.setSpeaker(i)
	}
	s.restoreSpeaker()
	if i := s.speakerIndex(); i >= 0 && s.rule().Refuses(s.Party[i]) {
		s.refuse()
		return
	}
	s.Inside.Dialog = s.greeting()
	s.Inside.DialogZh = nil
	s.setInsideIcons(InteriorIcons(s.door.IconFlags, len(s.Party)))
}

// refuse 印訊息 14。一個人的隊伍讀完就走；多人隊伍回頭問誰要開口。
func (s *Session) refuse() {
	s.greetWait = greetReask
	if len(s.Party) <= 1 {
		s.greetWait = greetLeave
	}
	s.Inside.Dialog = nonEmpty(s.message(HostileMessage))
	s.Inside.DialogZh = nil
	s.setInsideIcons([]int{IconSpacebar})
}

// continueGreeting 處理 greetWait 等著的那一步；沒有在等就回 false。
func (s *Session) continueGreeting() bool {
	wait := s.greetWait
	if wait == greetNone {
		return false
	}
	s.greetWait = greetNone
	switch wait {
	case greetLeave:
		s.finishLeaveDoor()
	case greetReask:
		s.askWho()
	}
	return true
}

// ChooseSpeaker 是按 1…6：換成第 n 位（從 0 算）開口。只有在問誰開口，或建築內一般圖示列
// 停著的時候才有效；選到目前的發言者什麼都不做（原版掃到同一位就繼續等鍵）。
func (s *Session) ChooseSpeaker(n int) bool {
	if n < 0 || n >= len(s.Party) || !(s.askingWho || s.interiorIdle()) {
		return false
	}
	if !s.askingWho && s.Party[n].Speaker {
		return true
	}
	s.setSpeaker(n)
	s.speakerGreeting()
	return true
}

// AskingWho 回報畫面是不是停在「誰要開口」。
func (s *Session) AskingWho() bool { return s.askingWho }

// interiorIdle 回報建築內是不是停在一般圖示列：沒有任何流程在進行，圖示列就是
// 這棟建築自己的那一列。
func (s *Session) interiorIdle() bool {
	if s.Mode != ModeInside || s.Inside == nil || s.door == nil || !s.greeted || s.leaving ||
		s.greetWait != greetNone || s.pending != nil ||
		s.recruit != nil || s.rest != nil || s.shop != nil || s.train != nil || s.scroll != nil ||
		s.quest != nil || s.mouth != nil || s.questList != nil || s.arena != nil || s.asking ||
		s.ShowingSheet() || s.ShowingItemPage() || s.ShowingPartyPage() || s.sheetHost != nil {
		return false
	}
	want := InteriorIcons(s.door.IconFlags, len(s.Party))
	if len(s.Inside.Icons) != len(want) {
		return false
	}
	for i, id := range want {
		if s.Inside.Icons[i] != IconSprite(id) {
			return false
		}
	}
	return true
}
