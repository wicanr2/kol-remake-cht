package town

// 接任務（RE-110）。入口是建築內的問話：玩家問到的關鍵字命中**第 5 槽**，
// 而且這一棟的 `icon_flags[1] & 4` 開著（與收任務物品同一位，同一棟給也收），
// 就走這條路而不是放一般的關鍵字回應。

// QuestSlot 是接任務那一槽。原版的關鍵字迴圈由第 5 槽往第 0 槽掃，
// 只有第 5 槽會走接任務（`0x12414`）。
const QuestSlot = 5

// 接任務用的 KOLDATA 記錄（RE-110）。
const (
	// TakeQuestRecord 是「Will you do this?」。
	TakeQuestRecord = 0x42
	// QuestAcceptedRecord 是「Very well. Be off!」。
	QuestAcceptedRecord = 0x43
	// QuestRefusedRecord 是「You lack valor.」。
	QuestRefusedRecord = 0x71
)

// 任務敘述的訊息編號：建築 payload 的 0x100…0x1FF，四則各四行，
// 原版分兩頁放（`sub_14564:0x145F5`，每頁兩則接成八行）。
const (
	QuestMessageFirst = 2
	QuestMessageCount = 4
)

// questOffer 是接任務走到哪一步。
type questOffer struct {
	// pages 是還沒講完的敘述，一頁八行。
	pages [][]string
	// asking 為真時停在「Will you do this?」。
	asking bool
	// site 是這一棟給的任務編號；member 是接的那一名。
	site   int
	member int
}

// OffersQuest 回報這一棟給不給任務。與收任務物品共用 `icon_flags[1]` 的同一位
// （RE-109／RE-110）——原版兩段各自檢查一次，資料側只有一個旗標。
func (d *Door) OffersQuest() bool { return d.IconFlags[1]&QuestFlagBit != 0 }

// QuestLines 是任務敘述的兩頁，一頁八行。缺頁就少一頁，不補空白。
func (b *BuildingText) QuestLines() [][]string {
	var pages [][]string
	for i := 0; i < QuestMessageCount; i += 2 {
		var page []string
		for j := 0; j < 2; j++ {
			m := b.Message(QuestMessageFirst + i + j)
			if m == nil {
				continue
			}
			page = append(page, m.Display...)
		}
		if len(page) > 0 {
			pages = append(pages, page)
		}
	}
	return pages
}

// questPages 取兩頁敘述，中文模式優先用譯文。
func (s *Session) questPages() [][]string {
	var pages [][]string
	for i := 0; i < QuestMessageCount; i += 2 {
		var page []string
		for j := 0; j < 2; j++ {
			page = append(page, s.message(QuestMessageFirst+i+j)...)
		}
		if len(page) > 0 {
			pages = append(pages, page)
		}
	}
	return pages
}

// offerQuest 是問話命中第 5 槽時發生的事。回傳有沒有接手——沒接手的話
// 呼叫端照常放那一槽的關鍵字回應（原版是 `sub_14564` 回非 0 的那兩條路）。
//
// 旗標的四個狀態各走各的（`0x145BD`／`0x146BA`／`0x1471C`）：
// 沒接過就講敘述再問；接了或完成了只再講一次敘述；回報過了就當一般話題。
func (s *Session) offerQuest() bool {
	if s.door == nil || !s.door.OffersQuest() {
		return false
	}
	i := s.speakerIndex()
	if i < 0 {
		return false
	}
	site := s.door.QuestSite
	switch QuestFlag(s.Party[i].Quests, site) {
	case QuestReported:
		// 已經回報過：原版回 5，呼叫端改放一般回應。
		return false
	case QuestNone:
		// 最終任務要先回報滿 0x17 件才接得到（`0x145E7`）。不夠就當一般話題。
		if s.door.QuestFinal && !CanTakeFinal(s.Party[i].Progress, 0) {
			return false
		}
	}
	pages := s.questPages()
	if len(pages) == 0 {
		return false
	}
	s.quest = &questOffer{pages: pages, site: site, member: i}
	// 接過了就只講敘述，不問（`0x146CC`）。
	if QuestFlag(s.Party[i].Quests, site) != QuestNone {
		s.quest.asking = false
		s.quest.member = -1
	} else {
		s.quest.asking = true
	}
	s.nextQuestPage()
	return true
}

// nextQuestPage 放下一頁敘述；講完了就問「Will you do this?」，
// 不問的那條路直接收掉。
func (s *Session) nextQuestPage() {
	if s.quest == nil || s.Inside == nil {
		return
	}
	if len(s.quest.pages) > 0 {
		page := s.quest.pages[0]
		s.quest.pages = s.quest.pages[1:]
		s.Inside.Dialog = page
		s.Inside.DialogZh = nil
		s.setInsideIcons([]int{IconSpacebar})
		return
	}
	if !s.quest.asking {
		s.quest = nil
		s.setInsideIcons(InteriorIcons(s.door.IconFlags, len(s.Party)))
		return
	}
	s.Inside.Dialog = nonEmpty(s.uiLines(TakeQuestRecord))
	s.Inside.DialogZh = nil
	s.setInsideIcons([]int{IconYes, IconNo})
}

// questAnswer 是「Will you do this?」的答案。答應就把旗標從「沒接」設成
// 「接了」——原版用 `or` 把除數疊上去（`0x146AB`），欄位本來就是 0。
func (s *Session) questAnswer(yes bool) {
	if s.quest == nil {
		return
	}
	record := QuestRefusedRecord
	if yes {
		record = QuestAcceptedRecord
		if i := s.quest.member; i >= 0 && i < len(s.Party) {
			SetQuestFlag(s.Party[i].Quests, s.quest.site, QuestTaken)
			s.TookQuest = &TakenQuest{Member: i, Site: s.quest.site,
				Final: s.door != nil && s.door.QuestFinal}
		}
	}
	s.quest = nil
	s.say(nonEmpty(s.uiLines(record)), nil)
}

// TakenQuest 是這一次接下的任務；沒接就是 nil。（`QuestTaken` 這個名字被旗標的
// 其中一個狀態佔用了。）
type TakenQuest struct {
	Member int
	Site   int
	Final  bool
}

// TakingQuest 回報現在停在接任務的哪一步——講敘述或問句。
func (s *Session) TakingQuest() bool { return s.quest != nil }
