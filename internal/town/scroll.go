package town

import (
	"encoding/hex"
	"strings"
)

// SCROLL（建築內圖示 0x0D，`sub_17318`，RE-121）做兩件事：加入這一棟代表的教派，
// 以及跟它買法術。原版的順序是
//
//	智力 < 門檻            → 直接跳到買法術（入不了會，照樣賣）
//	已經有教派             → 跳過入會，進選單
//	否則                   → 問要不要入會，答應就扣錢、寫進角色
//	選單                   → 第 0 列「Learn a spell」、第 1 列「Modify a spell」
//
// 第 1 列是法術編輯器（`sub_178E0`，RE-128）：挑一筆法術，改它的十三個欄位，
// 離開時按算出來的造價扣冒險點數與金冠。

// Order 是一個教派：建築 payload `0x50` 那十六個 byte，入會時整段抄進角色
// 記錄 `+0x1A0`。前八個 byte 是教派名、後八個是頭銜，兩個都是 5-bit 打包名，
// 而 `sub_1429C` 判斷「是不是這個教派的人」只比前八個。
type Order struct {
	Name string `json:"name"`
	// Title 是記錄後八個 byte 解出來的頭銜（Magician／Illusionist／Mystic…），
	// 入會之後與教派名一起印在角色資料頁那一列。
	Title string `json:"title"`
	// Intellect 是入會的智力門檻（payload `0x36`，`sub_17318:0x17321`）。
	Intellect int `json:"intellect"`
	// Allow 是這個教派准法術對哪幾類生效（payload `0x60` 起一類一個 byte，
	// `sub_178E0:0x17FB1`，RE-128）。六個教派剛好把八類分完：White Pearl 是
	// 人類與精靈、Blue Gem 是矮人與凱爾登、其餘四個各一類。
	Allow []int `json:"allow,omitempty"`
	// Record 是那十六個 byte 的十六進位；Title 從它的後半解出來。
	Record string `json:"record"`
	// Token 是入會信物（payload `0x2F0`）：一件十六 byte 的物品，
	// 優先放進角色 `+0x170`，那一格有東西才進背包。
	Token       string `json:"token"`
	TokenRecord string `json:"token_record"`
}

// SpellOffer 是 SCROLL 賣的一筆法術：建築 payload `0x100` 起，一筆 16 byte、
// 固定八筆，價錢是 `+0x0F` 的十倍。其餘欄位與角色記錄裡那十六格法術同版面。
type SpellOffer struct {
	Index     int    `json:"index"`
	Name      string `json:"name"`
	Price     int    `json:"price"`
	Effect    int    `json:"effect"`
	TurnCode  int    `json:"turn_code"`
	PowerCode int    `json:"power_code"`
	Guards    []int  `json:"guards,omitempty"`
	Record    string `json:"record"`
}

// SCROLL 用得到的 KOLDATA 記錄。
const (
	// JoinOfferRecord…JoinOrderRecord 三筆夾住教派名：
	// `"If you want to join@` ＋ `~H0Dthe]` ＋ 名字 ＋ ` Order]`。
	JoinOfferRecord = 0x44
	JoinTheRecord   = 0x45
	JoinOrderRecord = 0x46
	// MemberRecord 是 `"You are now a member of@`，後面接教派名與 `!"`。
	MemberRecord = 0x48
	// SpellOfferRecord 是 `"If you want this spell@`。
	SpellOfferRecord = 0x49
	// LearnSpellRecord／ModifySpellRecord 是選單那兩列。
	LearnSpellRecord  = 0x4E
	ModifySpellRecord = 0x4F
	// SpellbookFullRecord 是 `"You've learned enough!"`。
	SpellbookFullRecord = 0x6D
)

// SpellSlots 是角色記錄 `0x200`…`0x2FF` 那十六格法術（RE-100）。
const SpellSlots = 16

// scrollStage 是 SCROLL 流程停在哪一步。
type scrollStage int

const (
	scrollJoinAsk scrollStage = iota
	scrollMenu
	scrollList
	scrollConfirm
	// scrollEditPick 是「改哪一筆」的清單、scrollEdit 是編輯頁本身（RE-128）。
	scrollEditPick
	scrollEdit
)

type scrollState struct {
	stage       scrollStage
	top, cursor int
	// menu 是選單那一到兩列；rows 等於 1 時只有「Learn a spell」。
	rows int
	next func()
	// edit 是法術編輯器的狀態；nil 代表不在編輯（RE-128）。
	edit *spellEditState
}

// Scrolling 回報現在是不是停在 SCROLL 的畫面。
func (s *Session) Scrolling() bool { return s.scroll != nil }

// StartScroll 是按下 SCROLL。
func (s *Session) StartScroll() {
	if s.Mode != ModeInside || s.Inside == nil || s.door == nil {
		return
	}
	if s.door.Order == nil || len(s.door.Spells) == 0 {
		s.note(iconName(IconScroll))
		return
	}
	i := s.speakerIndex()
	if i < 0 {
		return
	}
	s.scroll = &scrollState{}
	switch {
	case s.intellect(i) < s.door.Order.Intellect:
		// 入不了會，但照樣賣法術（`0x17321` 直接跳到 `0x17677`）。
		s.showSpellList()
	case s.Party[i].Order != nil:
		s.showScrollMenu()
	default:
		s.askJoin()
	}
}

// intellect 是這名隊員的內部智力值（角色記錄 `+0x23`）。
func (s *Session) intellect(i int) int {
	if i < 0 || i >= len(s.Party) || s.Party[i].Sheet == nil {
		return 0
	}
	return s.Party[i].Sheet.Stats[StatIntellect]
}

// askJoin 問要不要入會：三筆記錄夾住教派名，再接價錢那一套句型。
func (s *Session) askJoin() {
	s.scroll.stage = scrollJoinAsk
	lines := nonEmpty(s.uiLines(JoinOfferRecord))
	lines = append(lines, joinLine(s.uiRaw(JoinTheRecord), s.shownName(s.door.Order.Name),
		s.uiRaw(JoinOrderRecord)))
	lines = append(lines, s.priceLines(0, s.door.Price)[1:]...)
	s.Inside.List = nil
	s.Inside.Detail = nil
	s.Inside.Dialog = lines
	s.Inside.DialogZh = nil
	s.setInsideIcons([]int{IconYes, IconNo})
}

// shownName 是代入對話句的資料名稱：中文模式查顯示譯名，英文照原名。
func (s *Session) shownName(en string) string {
	if s.DisplayName == nil {
		return en
	}
	return s.DisplayName(en)
}

// spellNotes 是法術清單每一列的中文注解；英文模式回 nil。
func (s *Session) spellNotes(rows []string) []string {
	if s.SpellGloss == nil {
		return nil
	}
	notes := make([]string, len(rows))
	for i, name := range rows {
		notes[i] = s.SpellGloss(name)
	}
	return notes
}

// shownExclaim 是名稱接驚嘆號的那一行；中文句用全形「！」。
func (s *Session) shownExclaim(en string) string {
	if s.DisplayName == nil {
		return en + "!"
	}
	return s.DisplayName(en) + "！"
}

// joinLine 把「the」、教派名與「Order」接成一行。兩筆記錄各自帶著前後空白，
// 中間補一個就接得起來。
func joinLine(the, name, order string) string {
	return strings.TrimSpace(the) + " " + name + strings.TrimRight(order, " ")
}

// showScrollMenu 是選單：有法術才畫第二列（`0x1755d` 掃 `0x200`…`0x300`）。
func (s *Session) showScrollMenu() {
	s.scroll.stage = scrollMenu
	s.scroll.rows = 1
	i := s.speakerIndex()
	if i >= 0 && len(s.Party[i].Spells) > 0 {
		s.scroll.rows = 2
	}
	s.scroll.cursor, s.scroll.top = 0, 0
	rows := []string{s.uiRecord(LearnSpellRecord)}
	if s.scroll.rows == 2 {
		rows = append(rows, s.uiRecord(ModifySpellRecord))
	}
	s.Inside.Dialog = nil
	s.Inside.DialogZh = nil
	s.Inside.Detail = nil
	s.Inside.List = &InnList{Rows: rows, Selected: 0, Total: len(rows)}
	s.setInsideIcons([]int{IconYes, IconUTurn})
}

// showSpellList 是八筆法術的清單，版面與商店的商品清單相同。
func (s *Session) showSpellList() {
	s.scroll.stage = scrollList
	s.scroll.cursor, s.scroll.top = 0, 0
	s.Inside.Dialog = nil
	s.Inside.DialogZh = nil
	s.Inside.Detail = nil
	s.Inside.List = s.spellView()
	s.setInsideIcons([]int{IconYes, IconUTurn})
}

func (s *Session) spellView() *InnList {
	sc := s.scroll
	offers := s.door.Spells
	end := sc.top + ListRows
	if end > len(offers) {
		end = len(offers)
	}
	rows := make([]string, 0, end-sc.top)
	for _, o := range offers[sc.top:end] {
		rows = append(rows, o.Name)
	}
	return &InnList{Rows: rows, Notes: s.spellNotes(rows), Selected: sc.cursor - sc.top,
		Total: len(offers)}
}

// MoveScroll 是 SCROLL 兩張清單上的方向鍵。
func (s *Session) MoveScroll(delta int) {
	sc := s.scroll
	if sc == nil {
		return
	}
	var total int
	switch sc.stage {
	case scrollMenu:
		total = sc.rows
	case scrollList:
		total = len(s.door.Spells)
	case scrollEditPick:
		total = len(s.ownSpells())
	case scrollEdit:
		s.MoveSpellField(delta)
		return
	default:
		return
	}
	if total == 0 {
		return
	}
	sc.cursor += delta
	if sc.cursor < 0 {
		sc.cursor = 0
	}
	if sc.cursor >= total {
		sc.cursor = total - 1
	}
	if sc.cursor < sc.top {
		sc.top = sc.cursor
	}
	if sc.cursor >= sc.top+ListRows {
		sc.top = sc.cursor - ListRows + 1
	}
	switch sc.stage {
	case scrollList:
		s.Inside.List = s.spellView()
		return
	case scrollEditPick:
		s.Inside.List = s.ownSpellView()
		return
	}
	s.Inside.List.Selected = sc.cursor
}

// scrollAnswer 是 SCROLL 流程裡對圖示的回應。
func (s *Session) scrollAnswer(id int) {
	sc := s.scroll
	if sc == nil {
		return
	}
	switch sc.stage {
	case scrollJoinAsk:
		if id != IconYes {
			s.endScroll()
			return
		}
		s.join()
	case scrollMenu:
		if id == IconUTurn {
			s.endScroll()
			return
		}
		if sc.cursor == 1 {
			s.showSpellPicker()
			return
		}
		s.showSpellList()
	case scrollList:
		if id == IconUTurn {
			s.endScroll()
			return
		}
		s.offerSpell()
	case scrollConfirm:
		if id != IconYes {
			s.endScroll()
			return
		}
		s.buySpell()
	case scrollEditPick:
		if id == IconUTurn {
			s.endScroll()
			return
		}
		s.startSpellEdit(sc.cursor)
	case scrollEdit:
		s.spellEditAnswer(id)
	}
}

// spellEditAnswer 是編輯頁上對圖示的回應。停在哪一層決定 YES 的意思：
// 名稱輸入收下字串、欄位選單選定一項、否則就是離開去結帳。
func (s *Session) spellEditAnswer(id int) {
	e := s.scroll.edit
	switch {
	case e.saving:
		if id != IconYes {
			s.endScroll()
			return
		}
		s.commitSpell()
	case e.typing != nil:
		if id != IconYes {
			e.typing = nil
			s.refreshSpellEdit()
			return
		}
		s.commitSpellName()
	case e.options != nil:
		if id != IconYes {
			e.options, e.top, e.cursor = nil, 0, 0
			s.refreshSpellEdit()
			return
		}
		s.chooseSpellOption()
	case id == IconYes:
		s.openSpellField()
	default:
		s.askRecordChanges()
	}
}

// join 是答應入會（`0x173fd`…`0x17553`）：錢不夠就講一句走人；夠就扣錢、
// 把教派那十六個 byte 記進角色，信物放進 `+0x170` 或背包。
func (s *Session) join() {
	i := s.speakerIndex()
	if i < 0 {
		s.endScroll()
		return
	}
	if s.Party[i].Gold < s.door.Price {
		s.scrollSay(nonEmpty(s.uiLines(NotEnoughRecord)), s.endScroll)
		return
	}
	slot := s.Party[i].emptyPackSlot()
	neckOccupied := s.Party[i].Worn[7] != nil && s.Party[i].Worn[7].Name != ""
	if neckOccupied && slot < 0 {
		s.scrollSay(nonEmpty(s.uiLines(BackpackFullRecord)), s.endScroll)
		return
	}
	order := *s.door.Order
	token, ok := orderTokenGood(order)
	if !ok {
		s.endScroll()
		return
	}
	s.Party[i].Gold -= s.door.Price
	if !neckOccupied {
		s.Party[i].Worn[7] = &token
	} else {
		s.Party[i].putPack(slot, token)
	}
	// 保留舊保存欄的信物資訊；實際位置由物品槽表示（RE-200）。
	s.Party[i].Token = &token
	s.Party[i].SyncCarried()
	s.Party[i].Order = &order
	if sheet := s.Party[i].Sheet; sheet != nil {
		copied := *sheet
		copied.Order = order.Name
		copied.Title = order.Title
		s.Party[i].Sheet = &copied
	}
	s.syncParty()
	lines := nonEmpty(s.uiLines(MemberRecord))
	lines = append(lines, s.shownExclaim(order.Name))
	s.scrollSay(lines, s.showScrollMenu)
}

// orderTokenGood 讀取既有 JSON 信物記錄的已證實欄位（RE-075／200）。
// 名稱已由抽取器提供，不讀 archive 或解壓原始資產。
func orderTokenGood(order Order) (Good, bool) {
	raw, err := hex.DecodeString(order.TokenRecord)
	if err != nil || len(raw) != CarriedStride || raw[0] == 0 || order.Token == "" {
		return Good{}, false
	}
	bonus := int(raw[8] >> 4)
	if bonus > 7 {
		bonus = -(bonus - 8)
	}
	count := int(raw[12]/32) + 1
	sides := [8]int{3, 4, 6, 8, 10, 12, 20, 100}[(raw[12]/4)&7]
	good := Good{Name: order.Token, Record: order.TokenRecord,
		Price: int(raw[8]&15)*256 + int(raw[9]), Kind: int(raw[12] & 3), Damage: raw[12]&1 != 0,
		Hands: int((raw[13] & 2) >> 1), Class: int(raw[13]), Position: int(raw[14] >> 4),
		Encumbrance: int(raw[10]), Condition: int(raw[11]), SizeClass: int(raw[14] & 15),
		Dice: Dice{Count: count, Sides: sides, Bonus: bonus, Low: max(count+bonus, 0), High: count*sides + bonus}}
	if good.Position == ShieldPosition {
		good.Percent = int(raw[14]&3)*10 + 20
	}
	return good, true
}

// offerSpell 是選中一筆法術：先看法術槽滿了沒，再問價錢。
func (s *Session) offerSpell() {
	i := s.speakerIndex()
	if i < 0 {
		s.endScroll()
		return
	}
	if len(s.Party[i].Spells) >= SpellSlots {
		s.scrollSay(nonEmpty(s.uiLines(SpellbookFullRecord)), s.endScroll)
		return
	}
	s.scroll.stage = scrollConfirm
	offer := s.door.Spells[s.scroll.cursor]
	s.Inside.List = nil
	s.Inside.DialogZh = nil
	s.Inside.Detail = nil
	s.Inside.Dialog = s.priceLines(SpellOfferRecord, offer.Price)
	s.setInsideIcons([]int{IconYes, IconNo})
}

// buySpell 是答應買下（`0x17870`…`0x178D5`）：錢不夠講一句，夠就扣錢並把那一筆
// 十六 byte 從建築 payload 抄進角色的空法術槽。
func (s *Session) buySpell() {
	i := s.speakerIndex()
	if i < 0 {
		s.endScroll()
		return
	}
	offer := s.door.Spells[s.scroll.cursor]
	if s.Party[i].Gold < offer.Price {
		s.scrollSay(nonEmpty(s.uiLines(NotEnoughRecord)), s.endScroll)
		return
	}
	s.Party[i].Gold -= offer.Price
	s.Party[i].Spells = append(s.Party[i].Spells, Spell{
		Name: offer.Name, Effect: offer.Effect, TurnCode: offer.TurnCode,
		PowerCode: offer.PowerCode,
		Guards:    append([]int(nil), offer.Guards...),
	})
	s.syncParty()
	// 抄完那十六個 byte 之後 `sub_17318` 就 `retf`（`0x178DA`）——買完直接回到
	// 建築的招呼語，不回清單。要再買一筆得重按 SCROLL。
	s.endScroll()
}

// scrollSay 顯示一段話，讀完之後接著做 next。
func (s *Session) scrollSay(lines []string, next func()) {
	s.Inside.Detail = nil
	s.say(lines, nil)
	if s.scroll != nil {
		s.scroll.next = next
	}
}

// endScroll 收掉 SCROLL 的畫面，回到建築的招呼語。
func (s *Session) endScroll() {
	s.scroll = nil
	if s.Inside == nil || s.door == nil {
		return
	}
	s.Inside.SpellEdit = nil
	s.Inside.List = nil
	s.Inside.Detail = nil
	s.Inside.Dialog = s.greeting()
	s.Inside.DialogZh = nil
	s.setInsideIcons(InteriorIcons(s.door.IconFlags, len(s.Party)))
}
