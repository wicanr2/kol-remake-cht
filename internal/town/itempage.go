package town

import "strings"

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
}

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
	s.showItemSlot()
	s.setInsideIcons([]int{IconUTurn})
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
	s.showItemSlot()
}

// showItemSlot 把選中那一格的明細或空格訊息備好。原版是按下 Enter 才顯示，
// remake 一移動就更新——同一份資料，少一次按鍵。
func (s *Session) showItemSlot() {
	p := s.Inside.ItemPage
	i := s.speakerIndex()
	if i < 0 {
		return
	}
	g := s.Party[i].itemAt(p.Selected)
	if g != nil {
		p.Detail = s.goodDetail(*g)
		p.Message = nil
		s.Inside.Detail = p.Detail
		return
	}
	p.Detail = nil
	p.Message = s.emptySlotLines(p.Selected)
	s.Inside.Detail = nil
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
