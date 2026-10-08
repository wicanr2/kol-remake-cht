package town

import "strings"

// spellEditState 是法術編輯器停在哪。原版把選中的那十六個 byte 抄進一份工作副本
// （DGROUP `0x9E28`），全部的修改都只動副本；離開時答應存檔才寫回角色記錄
// （`0x1826D`…`0x1827E`）。
type spellEditState struct {
	// slot 是這名隊員第幾筆法術；draft 是那一筆的工作副本。
	slot  int
	draft Spell
	// field 是欄位游標 0…12。
	field int
	// options 是正在開著的那一欄的選單；nil 代表沒開。
	options []string
	// top、cursor 是選單捲到哪、停在第幾項。
	top, cursor int
	// typing 非 nil 時停在名稱輸入。
	typing *string
	// saving 為真時停在「Record changes in Spellbook?」。
	saving bool
}

// SpellEditing 回報現在是不是停在法術編輯器。
func (s *Session) SpellEditing() bool {
	return s.scroll != nil && s.scroll.edit != nil
}

// SpellNameTyping 回報名稱欄位正接收文字；鍵盤方向鍵不應在此時變成字母。
func (s *Session) SpellNameTyping() bool {
	return s.SpellEditing() && s.scroll.edit.typing != nil
}

// showSpellPicker 是選單第 1 列：先列出這名隊員的法術，選一筆才進編輯頁
// （`0x178FB`…`0x17A82`）。
func (s *Session) showSpellPicker() {
	s.scroll.stage = scrollEditPick
	s.scroll.cursor, s.scroll.top = 0, 0
	s.Inside.Dialog = nil
	s.Inside.DialogZh = nil
	s.Inside.Detail = nil
	s.Inside.List = s.ownSpellView()
	s.setInsideIcons([]int{IconYes, IconUTurn})
}

func (s *Session) ownSpellView() *InnList {
	sc := s.scroll
	spells := s.ownSpells()
	end := sc.top + ListRows
	if end > len(spells) {
		end = len(spells)
	}
	rows := make([]string, 0, end-sc.top)
	for _, sp := range spells[sc.top:end] {
		rows = append(rows, sp.Name)
	}
	return &InnList{Rows: rows, Notes: s.spellNotes(rows), Selected: sc.cursor - sc.top,
		Total: len(spells)}
}

// ownSpells 是這名隊員身上有名字的那幾筆——原版跳過空槽（`0x17917`）。
func (s *Session) ownSpells() []Spell {
	i := s.speakerIndex()
	if i < 0 {
		return nil
	}
	out := make([]Spell, 0, len(s.Party[i].Spells))
	for _, sp := range s.Party[i].Spells {
		if sp.Name != "" {
			out = append(out, sp)
		}
	}
	return out
}

// startSpellEdit 把選中的那一筆抄成工作副本，進編輯頁。
func (s *Session) startSpellEdit(slot int) {
	spells := s.ownSpells()
	if slot < 0 || slot >= len(spells) {
		s.endScroll()
		return
	}
	draft := spells[slot]
	draft.Guards = append([]int(nil), draft.Guards...)
	for len(draft.Guards) < SpellGuards {
		draft.Guards = append(draft.Guards, 0)
	}
	s.scroll.stage = scrollEdit
	s.scroll.edit = &spellEditState{slot: slot, draft: draft}
	s.refreshSpellEdit()
}

// refreshSpellEdit 重畫編輯頁。原版每一輪都重算一次造價，所以改完馬上看得到。
func (s *Session) refreshSpellEdit() {
	s.Inside.Dialog = nil
	s.Inside.DialogZh = nil
	s.Inside.List = nil
	page := s.spellEditPage()
	s.Inside.SpellEdit = page
	s.Inside.Detail = page.Costs
	if page.Menu != nil {
		s.setInsideIcons([]int{IconYes, IconUTurn})
		return
	}
	// 兩個圖示：留在這一頁繼續改，或者離開去結帳（`sub_1874E` 的 7 與 0x0F）。
	s.setInsideIcons([]int{IconYes, IconUTurn})
}

// MoveSpellField 是編輯頁上的方向鍵。沒開選單時走欄位游標，開著就走選單。
// 欄位游標會繞圈：往前越過 0 回到 12、往後越過 12 回到 0（`0x17ABD`）。
func (s *Session) MoveSpellField(delta int) {
	if !s.SpellEditing() {
		return
	}
	e := s.scroll.edit
	if e.options != nil {
		s.moveSpellOption(delta)
		return
	}
	e.field = (e.field + delta) % SpellFieldCount
	if e.field < 0 {
		e.field += SpellFieldCount
	}
	s.refreshSpellEdit()
}

func (s *Session) moveSpellOption(delta int) {
	e := s.scroll.edit
	total := len(e.options)
	if total == 0 {
		return
	}
	e.cursor += delta
	if e.cursor < 0 {
		e.cursor = 0
	}
	if e.cursor >= total {
		e.cursor = total - 1
	}
	if e.cursor < e.top {
		e.top = e.cursor
	}
	if e.cursor >= e.top+ListRows {
		e.top = e.cursor - ListRows + 1
	}
	s.refreshSpellEdit()
}

// openSpellField 是在編輯頁上按下確認：名稱那一欄開文字輸入，RANGE 那一欄的
// 兩個選項來自 KOLDATA，其餘從 `MAGIC.0` 挑一段當選單（`0x17C50`…`0x17F70`）。
func (s *Session) openSpellField() {
	e := s.scroll.edit
	switch e.field {
	case SpellFieldName:
		// 原版sub_1B5C8有效輸入長度從0開始；副本仍保留舊名（RE-212）。
		typed := ""
		e.typing = &typed
	case SpellFieldTargetFlag:
		e.options = []string{s.uiRecord(SpellRangeRecord),
			s.uiRecord(SpellRangeRecord + 1)}
	default:
		e.options = SpellFieldOptions(s.Magic, e.draft, e.field)
	}
	// sub_1AAE8 在 IDA linear 0x1AB6E 清零選列，不採用欄位現值（RE-209）。
	e.cursor, e.top = 0, 0
	s.refreshSpellEdit()
}

// chooseSpellOption 是在欄位選單裡選定一項。四個類別那幾欄要先過教派這一關：
// 允許表那一格是零就不寫回，改講一句話（`0x17FA8`…`0x18043`）。
func (s *Session) chooseSpellOption() {
	e := s.scroll.edit
	if e.cursor < 0 || e.cursor >= len(e.options) {
		e.options = nil
		s.refreshSpellEdit()
		return
	}
	choice := e.cursor
	e.options, e.top, e.cursor = nil, 0, 0
	if e.field >= SpellFieldClassFirst && e.field <= SpellFieldClassLast &&
		!s.orderAllows(choice) {
		s.spellDenied()
		return
	}
	SetSpellField(s.Magic, &e.draft, e.field, choice)
	s.refreshSpellEdit()
}

// orderAllows 是「這個教派准不准這一類」。原版查 `[0x1CF4 + 0x60 + 選項]`，
// 那一格非零就放行；選 None（第 0 項）一律放行（`0x17FB1`…`0x17FCC`）。
// `0x1CF4` 是建築 payload 的緩衝——同一支在 `0x17684` 用 `[0x1CF4 + 0x100]`
// 讀這一棟賣的法術，所以那張允許表就在教派記錄後面。
//
// 查的是**這一棟**的教派，不是角色身上那一個：走進別家的塔就照別家的規矩。
func (s *Session) orderAllows(choice int) bool {
	if choice == 0 {
		return true
	}
	if s.door == nil || s.door.Order == nil {
		return true
	}
	for _, allowed := range s.door.Order.Allow {
		if allowed == choice {
			return true
		}
	}
	return false
}

// spellDenied 是不准的那一句：`"That one is unknown to@` ＋ `the Order of the@`
// ＋ 教派名。
func (s *Session) spellDenied() {
	lines := nonEmpty(s.uiLines(SpellDeniedRecord))
	lines = append(lines, nonEmpty(s.uiLines(SpellDeniedOrderRecord))...)
	if i := s.speakerIndex(); i >= 0 && s.Party[i].Order != nil {
		lines = append(lines, s.shownExclaim(s.Party[i].Order.Name))
	}
	s.scrollSay(lines, s.refreshSpellEdit)
}

// TypeSpellName 收名稱輸入的一個字。原版收滿12字便確認並關閉輸入框（RE-212）。
func (s *Session) TypeSpellName(r rune) {
	if !s.SpellEditing() || s.scroll.edit.typing == nil {
		return
	}
	typed := *s.scroll.edit.typing
	if len([]rune(typed)) >= SpellNameMax {
		return
	}
	typed += string(r)
	s.scroll.edit.typing = &typed
	if len([]rune(typed)) == SpellNameMax {
		s.commitSpellName()
		return
	}
	s.refreshSpellEdit()
}

// BackspaceSpellName 退一個字。
func (s *Session) BackspaceSpellName() {
	if !s.SpellEditing() || s.scroll.edit.typing == nil {
		return
	}
	runes := []rune(*s.scroll.edit.typing)
	if len(runes) == 0 {
		return
	}
	typed := string(runes[:len(runes)-1])
	s.scroll.edit.typing = &typed
	s.refreshSpellEdit()
}

// commitSpellName 收下打好的名字。原版要打了東西才接受（`byte_30D0A` 非零），
// 空的就當沒改。
func (s *Session) commitSpellName() {
	e := s.scroll.edit
	if e.typing == nil {
		return
	}
	if typed := strings.TrimSpace(*e.typing); typed != "" {
		// 原版5-bit名稱的顯示模式為各單字字首大寫，其餘小寫（RE-212）。
		name := []rune(typed)
		wordStart := true
		for i, r := range name {
			if r >= 'A' && r <= 'Z' && !wordStart {
				name[i] = r + ('a' - 'A')
			} else if r >= 'a' && r <= 'z' && wordStart {
				name[i] = r - ('a' - 'A')
			}
			wordStart = r == ' '
		}
		e.draft.Name = string(name)
	}
	e.typing = nil
	s.refreshSpellEdit()
}

// askRecordChanges 是離開編輯頁時的那一問（`0x18122`）。
func (s *Session) askRecordChanges() {
	s.scroll.edit.saving = true
	s.Inside.SpellEdit = nil
	s.Inside.Detail = nil
	s.Inside.List = nil
	s.Inside.Dialog = nonEmpty(s.uiLines(SpellSaveRecord))
	s.Inside.DialogZh = nil
	s.setInsideIcons([]int{IconYes, IconNo})
}

// commitSpell 是答應存檔：先比金冠、再比冒險點數，兩關都過才扣款並寫回
// （`0x1816C`…`0x1825D`）。訊息與扣的資源是對得上的——對調的只有畫面上
// 那兩行標籤，見 `SpellEditPage`。
func (s *Session) commitSpell() {
	i := s.speakerIndex()
	if i < 0 {
		s.endScroll()
		return
	}
	e := s.scroll.edit
	ap, gc := SpellCost(s.Magic, e.draft)
	sheet := s.Party[i].Sheet
	if s.Party[i].Gold < gc {
		s.scrollSay(nonEmpty(s.uiLines(NotEnoughRecord)), s.endScroll)
		return
	}
	if sheet == nil || sheet.AdventurePoints < ap {
		s.scrollSay(nonEmpty(s.uiLines(NoPointsRecord)), s.endScroll)
		return
	}
	s.Party[i].Gold -= gc
	copied := *sheet
	copied.AdventurePoints -= ap
	s.Party[i].Sheet = &copied
	s.writeSpellBack(i, e.slot, e.draft)
	s.syncParty()
	s.endScroll()
}

// writeSpellBack 把工作副本蓋回那一格。清單跳過了空槽，所以要按名字找回原本
// 在哪一格，不能拿清單的序號直接當索引。
func (s *Session) writeSpellBack(member, slot int, draft Spell) {
	seen := 0
	for k := range s.Party[member].Spells {
		if s.Party[member].Spells[k].Name == "" {
			continue
		}
		if seen == slot {
			s.Party[member].Spells[k] = draft
			return
		}
		seen++
	}
}
