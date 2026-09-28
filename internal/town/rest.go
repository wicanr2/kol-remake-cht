package town

// REST 的三條路（RE-073、SPEC-006）。圖示列上的 REST 在客棧與修道院都會出現，
// 按下去之後看建築旗標分岔：bit6 是修道院（治療），bit3 是客棧（住宿存檔）。

// REST 用到的 KOLDATA 記錄。
const (
	// WholePartyRecord 是隊伍超過一人時的第一問 `Whole party?`。
	WholePartyRecord = 0x35
	// QuitRecord 是隊伍空了之後的 `Do you wish to quit?`。
	QuitRecord = 0x36
	// StayRecord 是客棧的開場 `"If you want to stay`。
	StayRecord = 0x37
	// FreeRoomRecord 是免費客棧的 `just grab an open room!"`。
	FreeRoomRecord = 0x38
	// HealOfferRecord 是修道院的開場 `"If you need healing`。
	HealOfferRecord = 0x39
	// CostPrefix 與 CostSuffix 夾住價錢：`it will cost you` ＋ N ＋ `gold, interested?"`。
	CostPrefix = 0x53
	CostSuffix = 0x54
	// NotEnoughRecord 是 `"Sorry, not enough gold crowns!"`。
	NotEnoughRecord = 0x56
	// LookFineRecord 是修道院看到沒受傷的人講的 `"You look fine!"`。
	LookFineRecord = 0x62
	// HealedRecord 是治好之後的 `"Drink and be healed!"`。
	HealedRecord = 0x63
	// SaveDiskRecord 是 `Save character to hard disk?`。
	SaveDiskRecord = 0x189
)

// SaveFailedZh 是 remake 自己多的一句：存檔失敗時的說明。原版沒有對應訊息，
// 因為原版失敗就是沒有存檔。這一句只在非對拍路徑出現（SPEC-006）。
const SaveFailedZh = "存檔失敗，這一晚不算錢。"

// restStage 是 REST 流程停在哪一問。
type restStage int

const (
	restNone restStage = iota
	// restWhole 是 `Whole party?`。
	restWhole
	// restPrice 是價錢問句，客棧與修道院共用。
	restPrice
	// restDisk 是 `Save character to hard disk?`（只有原版住宿模式才會出現）。
	restDisk
	// restQuit 是隊伍空了之後的 `Do you wish to quit?`。
	restQuit
)

// restState 是一次 REST 的狀態。
type restState struct {
	stage restStage
	abbey bool
	// queue 是還沒住宿的隊員名字。原版對六個槽位逐一跑同一支流程，
	// 每個人各自看到自己的價錢問句、花自己的錢。
	queue []string
	// whole 是「全隊嗎」答了 YES。原版只有這一條路在每位之前重畫隊員格
	// （`sub_11A93:0x127E5` 的 `sub_13E82`），所以白框會跟著跑；
	// 答 NO 或單人隊伍都直接住宿，白框不動。
	whole bool
	// target 是目前處理的隊員名字。
	target string
	price  int
	// next 是「按空白鍵讀完這一段之後」要做的事。
	next func()
}

// Resting 回報現在是不是停在 REST 的某一問。
func (s *Session) Resting() bool { return s.rest != nil }

// QuitAsked 回報玩家在「你要離開嗎」那一問答了 YES。上層用它收掉城鎮畫面。
func (s *Session) QuitAsked() bool { return s.quit }

// ConsumeQuit 讓外層只處理一次返回書本的請求。
func (s *Session) ConsumeQuit() bool {
	quit := s.quit
	s.quit = false
	return quit
}

// StartRest 是按下 REST。
func (s *Session) StartRest() {
	if s.Mode != ModeInside || s.Inside == nil || s.door == nil {
		return
	}
	switch {
	case s.door.IsAbbey():
		s.startHeal()
	case s.door.IsInn():
		s.startStay()
	default:
		// 旗標沒有這兩個位元卻畫出 REST，是資料與 InteriorIcons 不一致。
		s.note(iconName(IconRest))
	}
}

// startHeal 是修道院：傷勢乘上費率。沒受傷就只講一句，不收錢也不問。
func (s *Session) startHeal() {
	i := s.speakerIndex()
	if i < 0 {
		return
	}
	if s.Party[i].Wounds == 0 {
		s.say(nonEmpty(s.uiLines(LookFineRecord)), nil)
		return
	}
	s.rest = &restState{stage: restPrice, abbey: true,
		target: s.Party[i].Name,
		price:  s.Party[i].Wounds * s.door.HealRate}
	s.askPrice(HealOfferRecord)
}

// startStay 是客棧。隊伍超過一人就先問 `Whole party?`。
func (s *Session) startStay() {
	i := s.speakerIndex()
	if i < 0 {
		return
	}
	s.rest = &restState{target: s.Party[i].Name, price: s.door.Price}
	if len(s.Party) > 1 {
		s.rest.stage = restWhole
		s.Inside.Dialog = nonEmpty(s.uiLines(WholePartyRecord))
		s.Inside.DialogZh = nil
		s.setInsideIcons([]int{IconYes, IconNo})
		return
	}
	s.rest.queue = []string{s.Party[i].Name}
	s.stayNext()
}

// stayNext 開始處理佇列裡的下一位；佇列空了就收尾。
func (s *Session) stayNext() {
	for len(s.rest.queue) > 0 {
		name := s.rest.queue[0]
		s.rest.queue = s.rest.queue[1:]
		if s.memberIndex(name) < 0 {
			continue // 已經離隊了
		}
		s.rest.target = name
		s.rest.price = s.door.Price
		if s.rest.whole {
			s.setSpeaker(s.memberIndex(name))
		}
		if s.rest.price == 0 {
			// 免費客棧只講一句，不問價錢（`sub_14DB1:0x14F2E`）。
			lines := append(nonEmpty(s.uiLines(StayRecord)),
				nonEmpty(s.uiLines(FreeRoomRecord))...)
			s.restSay(lines, s.afterPrice)
			return
		}
		s.rest.stage = restPrice
		s.askPrice(StayRecord)
		return
	}
	s.finishRest()
}

// askPrice 把「開場 ＋ 要花你 N 枚金幣，有意思嗎？」畫進對話框，圖示列換成大拇指。
//
// 原版把數字與後半句拆到兩行，是因為對話框只有 24 個原版字格。中文模式的框是
// 6 行 × 22 個字格，數字與量詞放得下同一行，拆開反而讀不順——所以有譯文時
// 併成一行，沒有譯文（嚴格對拍的英文模式）就照原版拆。
func (s *Session) askPrice(intro int) {
	s.Inside.Dialog = s.priceLines(intro, s.rest.price)
	s.Inside.DialogZh = nil
	s.setInsideIcons([]int{IconYes, IconNo})
}

// restAnswer 是 REST 流程裡對大拇指的回答。原版 `sub_144EF` 回傳 0 是 YES、1 是 NO。
func (s *Session) restAnswer(yes bool) {
	if s.rest == nil {
		return
	}
	switch s.rest.stage {
	case restWhole:
		if yes {
			s.rest.whole = true
			for _, m := range s.Party {
				s.rest.queue = append(s.rest.queue, m.Name)
			}
		} else {
			s.rest.queue = []string{s.rest.target}
		}
		s.stayNext()
	case restPrice:
		if !yes {
			s.cancelRest()
			return
		}
		i := s.memberIndex(s.rest.target)
		if i < 0 || s.Party[i].Gold < s.rest.price {
			s.restSay(nonEmpty(s.uiLines(NotEnoughRecord)), s.skipTarget)
			return
		}
		s.afterPrice()
	case restDisk:
		// 兩個答案都存檔：原版的 NO 只是把磁碟機切到 A:（`sub_14DB1:0x14F6E`）。
		s.commitStay()
	case restQuit:
		if yes {
			s.quit = true
			s.endRest()
			return
		}
		// 答 NO 接著進名冊清單（`sub_11A93:0x128A5` 呼叫 `sub_147DD`），與 COMPANION 同一條路。
		s.rest = nil
		s.StartRecruit()
	}
}

// afterPrice 是付得起、也答應了之後的下一步。
func (s *Session) afterPrice() {
	if s.rest.abbey {
		s.commitHeal()
		return
	}
	if s.FaithfulInnStay {
		s.rest.stage = restDisk
		s.Inside.Dialog = nonEmpty(s.uiLines(SaveDiskRecord))
		s.Inside.DialogZh = nil
		s.setInsideIcons([]int{IconYes, IconNo})
		return
	}
	s.commitStay()
}

// commitHeal 是修道院成交：扣錢、傷勢歸零。原版不在這裡存檔。
func (s *Session) commitHeal() {
	i := s.memberIndex(s.rest.target)
	if i < 0 {
		s.endRest()
		return
	}
	s.Party[i].Gold -= s.rest.price
	s.Party[i].Wounds = 0
	if s.Party[i].Sheet != nil {
		s.Party[i].Sheet.Nutrition = 0
	}
	s.syncParty()
	s.restSay(nonEmpty(s.uiLines(HealedRecord)), s.endRest)
}

// commitStay 是客棧成交：扣錢、記下留宿的建築、寫存檔。存檔失敗就整筆退回。
func (s *Session) commitStay() {
	i := s.memberIndex(s.rest.target)
	if i < 0 {
		s.stayNext()
		return
	}
	beforeParty := append([]Member(nil), s.Party...)
	beforeRoster := append([]Candidate(nil), s.Roster...)
	s.Party[i].Gold -= s.rest.price
	s.Party[i].Place = placeKey(s.door.Name)
	s.Party[i].BuildingID = s.door.RecordID
	if s.TownCode != "" {
		s.Party[i].TownCode = s.TownCode
	}
	updated := CandidateFromMember(s.Party[i])
	found := false
	for j := range s.Roster {
		if s.Roster[j].Name == updated.Name {
			updated.Slot = s.Roster[j].Slot
			s.Roster[j] = updated
			found = true
			break
		}
	}
	if !found {
		s.Roster = append(s.Roster, updated)
	}
	if s.FaithfulInnStay {
		// 原版住宿即存檔即離隊（`sub_14DB1:0x15287`）。
		s.Party = append(s.Party[:i], s.Party[i+1:]...)
		if len(s.Party) > 0 {
			hasSpeaker := false
			for _, m := range s.Party {
				hasSpeaker = hasSpeaker || m.Speaker
			}
			if !hasSpeaker {
				s.Party[0].Speaker = true
			}
		}
	}
	if s.SaveGame != nil {
		if err := s.SaveGame(); err != nil {
			s.Party, s.Roster = beforeParty, beforeRoster
			s.syncParty()
			s.SaveError = err
			s.restSay([]string{SaveFailedZh}, s.endRest)
			return
		}
	}
	s.SaveError = nil
	s.syncParty()
	s.stayNext()
}

// skipTarget 是這一位沒住成（錢不夠），換下一位。
func (s *Session) skipTarget() { s.stayNext() }

// finishRest 是佇列跑完之後的收尾。原版把「目前這位」指回第一個還在的槽位
// （`sub_11A93:0x1282A`），隊伍空了才問是否離開遊戲。
func (s *Session) finishRest() {
	if len(s.Party) > 0 {
		s.setSpeaker(0)
	}
	if s.FaithfulInnStay && len(s.Party) == 0 {
		s.rest.stage = restQuit
		s.Inside.Dialog = nonEmpty(s.uiLines(QuitRecord))
		s.Inside.DialogZh = nil
		s.setInsideIcons([]int{IconYes, IconNo})
		return
	}
	s.endRest()
}

// cancelRest 是在價錢問句答 NO：回到建築內的一般畫面。
func (s *Session) cancelRest() { s.endRest() }

// endRest 結束流程，畫面回到招呼語與一般圖示列。
func (s *Session) endRest() {
	s.rest = nil
	if s.Inside == nil || s.door == nil {
		return
	}
	s.Inside.Dialog = s.greeting()
	s.Inside.DialogZh = nil
	s.setInsideIcons(InteriorIcons(s.door.IconFlags, len(s.Party)))
}

// restSay 顯示一段話，圖示列縮成一格 SPACEBAR；讀完之後接著做 next。
func (s *Session) restSay(lines []string, next func()) {
	s.say(lines, nil)
	if s.rest != nil {
		s.rest.next = next
	}
}

// setSpeaker 把發言者換成隊伍裡的第 i 位，並同步到畫面。
func (s *Session) setSpeaker(i int) {
	if i < 0 || i >= len(s.Party) {
		return
	}
	for k := range s.Party {
		s.Party[k].Speaker = k == i
	}
	s.syncParty()
}

// syncParty 把隊伍的異動同步到畫面狀態。
func (s *Session) syncParty() {
	if s.Outdoor != nil {
		s.Outdoor.Party = s.Party
	}
	if s.Inside != nil {
		s.Inside.Party = s.Party
	}
}

// speakerIndex 是目前發言者在隊伍裡的位置；隊伍空的時候回 -1。
func (s *Session) speakerIndex() int {
	if len(s.Party) == 0 {
		return -1
	}
	for i, m := range s.Party {
		if m.Speaker {
			return i
		}
	}
	return 0
}

// memberIndex 依名字找隊員；找不到回 -1。
func (s *Session) memberIndex(name string) int {
	for i, m := range s.Party {
		if m.Name == name {
			return i
		}
	}
	return -1
}
