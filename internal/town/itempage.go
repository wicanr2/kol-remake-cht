package town

import (
	"fmt"
	"strings"

	"github.com/wicanr2/kol-remake-cht/internal/portrait"
)

// 物品那一頁（`sub_16986`，RE-127）。清單十七列，每一列對應角色記錄裡的一格。
const (
	// ItemSlotLabels 是清單的文字：十七項用 `@` 分隔。
	ItemSlotLabels = 0x8D
	// SlotNameRecord 起是單獨的部位名稱（Head…Weapon），空格那一句要接它。
	SlotNameRecord = 0x8E
	// SlotNameLast 是名稱表夾住的上限：超過就用最後那一筆（Pocket）。
	SlotNameLast = 0x0B
	// NothingRecord／OnYourRecord／InPocketRecord 是空格那三句。
	NothingRecord  = 0xCA
	OnYourRecord   = 0xCB
	InPocketRecord = 0xCC
	// PackFirst／PackLast 是背包那六格在角色記錄裡的位移。
	PackFirst = 0xA0
	PackLast  = 0xF0
	// WornBase／WornStride 是穿戴欄十格。
	WornBase   = 0x100
	WornStride = 0x10
	// ItemSlotCount 是清單有幾列。
	ItemSlotCount = 17
	// LastWornRow 是清單上最後一列「身上」的那一格（WEAPON）。它之後是背包，
	// 空格那一句因此換成 `in your pocket!`。
	LastWornRow = 0x0A
)

// ItemSlotOffsets 是清單每一列對應的角色記錄位移（DGROUP 0x618，RE-127）：
// 穿戴欄十格、手上那件、背包六格。順序就是清單上的順序。
func ItemSlotOffsets() []int {
	out := make([]int, 0, ItemSlotCount)
	for pos := 0; pos < WornSlots; pos++ {
		out = append(out, WornBase+pos*WornStride)
	}
	out = append(out, CarriedFirst)
	for off := PackFirst; off <= PackLast; off += CarriedStride {
		out = append(out, off)
	}
	return out
}

// ItemPage 是角色資料頁上的物品那一頁。
type ItemPage struct {
	// Rows 是十七列的部位名稱，順序同 ItemSlotOffsets。
	Rows []string `json:"rows"`
	// Selected 是反白的那一列。
	Selected int `json:"selected"`
	// Detail 是選中那一格的明細（與商店同一支，RE-075）；那一格空著就是 nil。
	Detail []Segment `json:"detail,omitempty"`
	// Message 是空格那幾句。
	Message []string `json:"message,omitempty"`
	// Acting 為真時圖示列換成那一格能做的事（`sub_18857`）。
	Acting bool `json:"acting,omitempty"`
	// Giving 為真時在等玩家選要給誰。
	Giving bool `json:"giving,omitempty"`
	// Note 是動作做完之後那一句（丟掉了、給出去了、穿不下…）。
	Note []string `json:"note,omitempty"`
	// Figure 是左邊那個人要貼的幾張圖（`sub_18F3A(0)`，RE-126）。
	Figure []portrait.Piece `json:"figure,omitempty"`
	// Mark 是清單移動時左上角那個小框（`sub_16D0B`，RE-127）。
	Mark *ItemMark `json:"mark,omitempty"`
}

// 動作用得到的幾筆訊息（RE-127）。
const (
	TooSmallRecord      = 0x27
	GiveToWhomRecord    = 0x28
	TheirPackFullRecord = 0x29
	TradedRecord        = 0x2A
	DroppedRecord       = 0xD2
	YourPackFullRecord  = 0xD3
)

// itemRows 把清單那一筆記錄拆成十七列。少一列就整頁不開——寧可看得出來壞了。
func (s *Session) itemRows() []string {
	rows := s.uiLines(ItemSlotLabels)
	if len(rows) != ItemSlotCount {
		return nil
	}
	out := make([]string, len(rows))
	for i, r := range rows {
		out[i] = strings.ToUpper(strings.TrimSpace(r))
	}
	return out
}

// itemAt 取清單第 row 列那一格的東西。穿戴欄十格與手上那件走 typed data，
// 背包那六格照 Backpack 的順序。
func (m Member) itemAt(row int) *Good {
	switch {
	case row < 0 || row >= ItemSlotCount:
		return nil
	case row < WornSlots:
		return m.Worn[row]
	case row == WornSlots:
		return m.InHand
	}
	i := row - WornSlots - 1
	if i < len(m.Backpack) {
		return &m.Backpack[i]
	}
	return nil
}

// OpenItemPage 是資料頁上按下物品那一格（`sub_16986`）：整頁換成十七列的清單。
func (s *Session) OpenItemPage() {
	if s.Inside == nil || s.Inside.Sheet == nil {
		return
	}
	rows := s.itemRows()
	if rows == nil {
		s.note(sheetIconName(IconSheetItems))
		return
	}
	s.Inside.ItemPage = &ItemPage{Rows: rows}
	s.refreshItemFigure()
	s.refreshItemMark()
	// 清單層的圖示列是那一格加 U-TURN（`sub_16986` 的 `sub_11703(9)`／`(0x0F)`）。
	s.setInsideIcons([]int{IconItemUse, IconUTurn})
}

// ShowingItemPage 回報現在是不是停在物品那一頁。
func (s *Session) ShowingItemPage() bool {
	return s.Inside != nil && s.Inside.ItemPage != nil
}

// closeItemPage 是物品頁上選 U-TURN：回角色資料頁。
func (s *Session) closeItemPage() {
	if s.Inside == nil {
		return
	}
	s.Inside.ItemPage = nil
	s.Inside.Detail = nil
	i := s.speakerIndex()
	if i < 0 {
		s.closeSheet()
		return
	}
	s.setInsideIcons(sheetIcons(s.Party[i]))
}

// MoveItemSelection 是上下鍵：到頂到底都繞回去（`0x198BA`／`0x198EA`）。
func (s *Session) MoveItemSelection(delta int) {
	if !s.ShowingItemPage() {
		return
	}
	p := s.Inside.ItemPage
	n := len(p.Rows)
	if n == 0 {
		return
	}
	p.Selected = ((p.Selected+delta)%n + n) % n
	// 清單層只有立繪、清單與左上角那個小框：明細框要按下那一格才彈出來
	// （原版是 `sub_196A9` 回來之後才 `sub_1BA9D`）。
	p.Detail, p.Message, p.Note = nil, nil, nil
	s.Inside.Detail = nil
	s.refreshItemMark()
}

// emptySlotLines 是那一格空著時的三句（`0x18C66`…`0x18CE7`）：
// 身上那十一格印「You have nothing on your <部位>」，背包六格印
// 「You have nothing in your pocket!」。
func (s *Session) emptySlotLines(row int) []string {
	out := nonEmpty(s.uiLines(NothingRecord))
	if row > LastWornRow {
		return append(out, nonEmpty(s.uiLines(InPocketRecord))...)
	}
	out = append(out, nonEmpty(s.uiLines(OnYourRecord))...)
	name := row
	if name > SlotNameLast {
		name = SlotNameLast
	}
	return append(out, nonEmpty(s.uiLines(SlotNameRecord+name))...)
}

// 物品頁的動作圖示（`sub_18857` 的跳表，RE-127）。編號就是原版的圖號。
const (
	// IconItemUse 是清單層那一格：按下去進動作層（`sub_16986` 的 `sub_11703(9)`）。
	IconItemUse = 0x09
	// 動作層那幾格，順序與原版加進圖示列的順序相同。
	IconItemTake = 0x16 // 拿到手上
	IconItemWear = 0x10 // 穿上
	IconItemBelt = 0x1C // 掛到腰帶
	IconItemGive = 0x0E // 給隊友
	IconItemDrop = 0x15 // 丟掉
	IconItemStow = 0x14 // 放回背包
	IconItemEat  = 0x03 // 吃掉
)

// EatPosition 是「吃得下」的那個部位：原版比 `+0x0E` 的高 nibble 等於 10。
const EatPosition = 10

// itemIcons 是選中那一格能做的事（`sub_18857` 開頭那一串 `sub_11703`）。
func itemIcons(g Good, row, party int) []int {
	icons := []int{IconUTurn}
	if row != WornSlots && g.Kind&1 != 0 {
		icons = append(icons, IconItemTake)
	}
	if row > WornSlots && g.Kind&3 == 2 {
		icons = append(icons, IconItemWear)
	}
	if row == WornSlots && g.Kind&3 == 3 {
		icons = append(icons, IconItemBelt)
	}
	if party > 1 {
		icons = append(icons, IconItemGive)
	}
	icons = append(icons, IconItemDrop)
	if row <= WornSlots {
		icons = append(icons, IconItemStow)
	}
	if g.Position == EatPosition {
		icons = append(icons, IconItemEat)
	}
	return icons
}

// removeItem 把清單第 row 列那一格清空，回傳原本放在那裡的東西。
//
// 背包在 remake 是緊湊的 slice（原版是固定六格、中間可以有洞）。兩者在「沒有洞」
// 的狀態下等價，而 remake 只會從尾端補，所以洞不會出現。
func (m *Member) removeItem(row int) *Good {
	switch {
	case row < 0 || row >= ItemSlotCount:
		return nil
	case row < WornSlots:
		g := m.Worn[row]
		m.Worn[row] = nil
		return g
	case row == WornSlots:
		g := m.InHand
		m.InHand = nil
		return g
	}
	i := row - WornSlots - 1
	if i >= len(m.Backpack) {
		return nil
	}
	g := m.Backpack[i]
	m.Backpack = append(m.Backpack[:i], m.Backpack[i+1:]...)
	return &g
}

// AcceptItemSlot 是清單層按下那一格：換成動作圖示列（`sub_18857`）。
func (s *Session) AcceptItemSlot() {
	if !s.ShowingItemPage() {
		return
	}
	p := s.Inside.ItemPage
	i := s.speakerIndex()
	if i < 0 {
		return
	}
	g := s.Party[i].itemAt(p.Selected)
	if g == nil {
		// 空的那一格只印那幾句，沒有動作可做。
		p.Detail, p.Note = nil, nil
		p.Message = s.emptySlotLines(p.Selected)
		s.Inside.Detail = nil
		return
	}
	// 資料頁那一整頁走大寫旗標（`word_2DF54`，RE-074），明細跟著轉大寫——
	// 原版畫面收據的 frame-079 印的是 `TYP:WEAR  POS:TRSO`。
	p.Detail = upperSegments(s.goodDetail(*g))
	p.Message, p.Note = nil, nil
	s.Inside.Detail = p.Detail
	p.Acting = true
	s.setInsideIcons(itemIcons(*g, p.Selected, len(s.Party)))
}

// closeItemAction 回到清單層：框裡那幾行收掉，圖示列換回那兩格。
func (s *Session) closeItemAction() {
	p := s.Inside.ItemPage
	p.Acting = false
	p.Giving = false
	p.Detail, p.Message, p.Note = nil, nil, nil
	s.Inside.Detail = nil
	s.refreshItemMark()
	s.setInsideIcons([]int{IconItemUse, IconUTurn})
}

// ItemAction 執行動作層選到的那一格（`sub_18857` 的跳表）。
func (s *Session) ItemAction(id int) {
	if !s.ShowingItemPage() {
		return
	}
	p := s.Inside.ItemPage
	i := s.speakerIndex()
	if i < 0 {
		return
	}
	m := &s.Party[i]
	g := m.itemAt(p.Selected)
	if g == nil {
		s.closeItemAction()
		return
	}
	switch id {
	case IconUTurn:
		s.closeItemAction()
	case IconItemTake:
		s.moveItem(m, p.Selected, WornSlots)
	case IconItemBelt:
		s.moveItem(m, p.Selected, WornSlots-1)
	case IconItemWear:
		s.wearItem(m, p.Selected, *g)
	case IconItemStow:
		s.stowItem(m, p.Selected)
	case IconItemDrop:
		m.removeItem(p.Selected)
		p.Note = nonEmpty(s.uiLines(DroppedRecord))
		s.afterItemAction()
	case IconItemEat:
		s.eatItem(m, p.Selected, *g)
	case IconItemGive:
		p.Giving = true
		p.Note = nonEmpty(s.uiLines(GiveToWhomRecord))
	}
}

// moveItem 把那一格的東西搬到另一格（`sub_18CD5`）。目標有東西就對調。
func (s *Session) moveItem(m *Member, from, to int) {
	if from == to {
		s.afterItemAction()
		return
	}
	g := m.removeItem(from)
	if g == nil {
		s.afterItemAction()
		return
	}
	old := m.removeItem(to)
	m.putItem(to, g)
	if old != nil {
		m.putItem(from, old)
	}
	s.afterItemAction()
}

// putItem 把一件東西放進清單第 row 列那一格。
func (m *Member) putItem(row int, g *Good) {
	switch {
	case row < WornSlots:
		m.Worn[row] = g
	case row == WornSlots:
		m.InHand = g
	default:
		if len(m.Backpack) < BackpackSlots {
			m.Backpack = append(m.Backpack, *g)
		}
	}
}

// wearItem 是穿上：先看合不合身（`sub_157C6`），過了才搬進那件東西自己的部位。
func (s *Session) wearItem(m *Member, from int, g Good) {
	p := s.Inside.ItemPage
	if FitSize(g, s.memberSize(s.speakerIndex())) > g.Encumbrance {
		p.Note = nonEmpty(s.uiLines(TooSmallRecord))
		return
	}
	if g.Position < 0 || g.Position >= WornSlots {
		p.Note = nonEmpty(s.uiLines(TooSmallRecord))
		return
	}
	s.moveItem(m, from, g.Position)
}

// stowItem 是放回自己的背包：滿了就印那一句（記錄 0xD3）。
func (s *Session) stowItem(m *Member, from int) {
	p := s.Inside.ItemPage
	if len(m.Backpack) >= BackpackSlots {
		p.Note = nonEmpty(s.uiLines(YourPackFullRecord))
		return
	}
	g := m.removeItem(from)
	if g == nil {
		s.afterItemAction()
		return
	}
	m.Backpack = append(m.Backpack, *g)
	s.afterItemAction()
}

// eatItem 是吃掉：飢餓度減掉那一件的 `rec[11]`，夾在 0 以上（`sub_15896`，RE-075）。
func (s *Session) eatItem(m *Member, from int, g Good) {
	if m.Sheet != nil {
		n := m.Sheet.Nutrition - g.Condition
		if n < 0 {
			n = 0
		}
		sheet := *m.Sheet
		sheet.Nutrition = n
		m.Sheet = &sheet
	}
	m.removeItem(from)
	s.afterItemAction()
}

// GiveItem 把選中那一格給隊伍中的第 n 位（`sub_18857` 的 `'1'`…`'6'`）。
func (s *Session) GiveItem(n int) {
	if !s.ShowingItemPage() || !s.Inside.ItemPage.Giving {
		return
	}
	p := s.Inside.ItemPage
	i := s.speakerIndex()
	if n < 0 || n >= len(s.Party) || n == i || i < 0 {
		return
	}
	if len(s.Party[n].Backpack) >= BackpackSlots {
		p.Note = nonEmpty(s.uiLines(TheirPackFullRecord))
		return
	}
	g := s.Party[i].removeItem(p.Selected)
	if g == nil {
		s.closeItemAction()
		return
	}
	s.Party[n].Backpack = append(s.Party[n].Backpack, *g)
	p.Note = nonEmpty(s.uiLines(TradedRecord))
	s.afterItemAction()
}

// afterItemAction 是動作做完：回清單層並重新算明細。原版是把 `byte_2DF50` 設成 1
// 讓物品頁整頁重畫，效果相同。
func (s *Session) afterItemAction() {
	p := s.Inside.ItemPage
	note := p.Note
	s.closeItemAction()
	p.Note = note
	s.refreshItemFigure()
	s.syncParty()
}

// Digit 是物品頁吃的數字鍵：正在問「給誰」的時候選人（`sub_18857` 的 `'1'`…`'6'`）。
func (s *Session) Digit(n int) {
	if s.ShowingItemPage() && s.Inside.ItemPage.Giving {
		s.GiveItem(n)
	}
}

// refreshItemFigure 重算左邊那個人。裝備一動立繪就跟著變——原版是整頁重畫，
// 這裡只重算那一串貼圖計畫。
func (s *Session) refreshItemFigure() {
	if !s.ShowingItemPage() || s.Portraits == nil {
		return
	}
	i := s.speakerIndex()
	if i < 0 {
		return
	}
	s.Inside.ItemPage.Figure = s.Portraits.Compose(s.Party[i].Figure(), 0)
}

// 左上角那個小框（`sub_16D0B`，RE-127）。框寬 0x63、高 0x15；一般畫在 x = 6，
// WEAPON 那一列往右挪到 x = 0x6E。
const (
	MarkX      = 6
	MarkWeapon = 0x6E
	MarkWidth  = 0x63
	MarkHeight = 0x15
	// 三行的行距：第一行在框頂 ＋2，之後每行 ＋6。
	MarkTextTop  = 2
	MarkLineStep = 6
	// 重量印在第 10 欄起（相對框的左緣），位數少就往右挪，等於靠右對齊。
	MarkWeightCol = 0x0A
	// ShieldRow 是唯一不印骰式的那一列（`sub_16D0B` 比的是 `arg_2 != 6`）。
	ShieldRow = 6
)

// ItemMark 是清單移動時左上角那個小框的內容。
type ItemMark struct {
	X, Y int    `json:"-"`
	Slot string `json:"slot"`
	Name string `json:"name,omitempty"`
	// Dice 是骰式（SHIELD 那一列不印）；Skill 是 WEAPON 那一列多印的技能列號。
	Dice   string `json:"dice,omitempty"`
	Skill  string `json:"skill,omitempty"`
	Weight int    `json:"weight,omitempty"`
}

// refreshItemMark 重算左上角那個小框。位置來自版面表裡「那個部位在立繪上的高度」。
func (s *Session) refreshItemMark() {
	if !s.ShowingItemPage() {
		return
	}
	p := s.Inside.ItemPage
	i := s.speakerIndex()
	if i < 0 || s.Portraits == nil {
		p.Mark = nil
		return
	}
	m := s.Party[i]
	race := 0
	if m.Sheet != nil {
		race = m.Sheet.Race
	}
	if race < 0 || race >= len(s.Portraits.SlotMarks) {
		p.Mark = nil
		return
	}
	x := MarkX
	if p.Selected == WornSlots {
		x = MarkWeapon
	}
	name := p.Selected
	if name > SlotNameLast {
		name = SlotNameLast
	}
	mark := &ItemMark{X: x, Y: s.Portraits.SlotMarks[race][p.Selected],
		Slot: strings.ToUpper(s.uiRecord(SlotNameRecord+name)) + ":"}
	if g := m.itemAt(p.Selected); g != nil {
		mark.Name = strings.ToUpper(g.Name)
		mark.Weight = g.Encumbrance
		if p.Selected != ShieldRow {
			mark.Dice = fmt.Sprintf("%d-%d", g.Dice.Low, g.Dice.High)
		}
		// 手上那件多印一個「用哪一列武器技能」的號碼（`0x16F0E` 的迴圈）。
		if p.Selected == WornSlots && m.Sheet != nil {
			for n, row := range m.Sheet.Weapons {
				if row.Weapon == g.Class {
					mark.Skill = fmt.Sprintf("%d", n+1)
					break
				}
			}
		}
	}
	p.Mark = mark
}

// upperSegments 把明細那幾段轉大寫。
func upperSegments(segs []Segment) []Segment {
	out := make([]Segment, len(segs))
	for i, s := range segs {
		s.Text = strings.ToUpper(s.Text)
		out[i] = s
	}
	return out
}
