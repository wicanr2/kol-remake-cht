package town

import "testing"

// itemText 是物品那一頁要用的幾筆記錄，內容照原版 KOLDATA（RE-127）。
func itemText() *UIText {
	t := restText()
	t.Banks[0].Records = append(t.Banks[0].Records,
		uiRecord{Index: ItemSlotLabels, Text: "HEAD@TORSO@LEGS@BODY@HANDS@FEET@" +
			"SHIELD@NECK@RING@BELT@WEAPON@POCKET@POCKET@POCKET@POCKET@POCKET@POCKET"},
		uiRecord{Index: NothingRecord, Text: "    You have nothing@"},
		uiRecord{Index: OnYourRecord, Text: "    on your]"},
		uiRecord{Index: InPocketRecord, Text: "    in your pocket!"},
		uiRecord{Index: SlotNameRecord + 0, Text: "Head"},
		uiRecord{Index: SlotNameRecord + 9, Text: "Belt"},
		uiRecord{Index: SlotNameRecord + 10, Text: "Weapon"},
		uiRecord{Index: SlotNameRecord + SlotNameLast, Text: "Pocket"},
	)
	return t
}

// itemSession 停在角色資料頁上，圖示列是 U-TURN／物品／換人。
func itemSession(t *testing.T) *Session {
	t.Helper()
	s := innSession(t, 56, Member{Name: "Theodoric", Gold: 1009,
		Portrait: 47, Speaker: true, Sheet: theodoric()})
	s.UIText = itemText()
	s.OpenSheet()
	if !s.ShowingSheet() {
		t.Fatal("沒進到角色資料頁")
	}
	return s
}

// openItems 按下資料頁圖示列的第二格（物品）。
func openItems(s *Session) {
	s.Inside.Selected = 1
	s.Accept()
}

func TestItemSlotOffsetsMatchCarried(t *testing.T) {
	got := ItemSlotOffsets()
	if len(got) != ItemSlotCount {
		t.Fatalf("清單有 %d 列", len(got))
	}
	seen := map[int]bool{}
	for _, off := range got {
		if seen[off] {
			t.Fatalf("位移 0x%x 出現兩次", off)
		}
		seen[off] = true
	}
	for _, off := range CarriedOffsets() {
		if !seen[off] {
			t.Fatalf("身上那一格 0x%x 不在清單裡", off)
		}
	}
	if got[0] != WornBase || got[WornSlots-1] != WornBase+(WornSlots-1)*WornStride {
		t.Fatalf("穿戴欄那十格是 0x%x…0x%x", got[0], got[WornSlots-1])
	}
	if got[WornSlots] != CarriedFirst {
		t.Fatalf("第十一列是 0x%x，應該是手上那件", got[WornSlots])
	}
}

func TestOpenItemPageListsSeventeenRows(t *testing.T) {
	s := itemSession(t)
	openItems(s)
	if !s.ShowingItemPage() {
		t.Fatalf("沒開起來；未實作紀錄：%v", s.Unimplemented)
	}
	p := s.Inside.ItemPage
	if len(p.Rows) != ItemSlotCount {
		t.Fatalf("清單有 %d 列：%v", len(p.Rows), p.Rows)
	}
	if p.Rows[0] != "HEAD" || p.Rows[WornSlots-1] != "BELT" ||
		p.Rows[WornSlots] != "WEAPON" || p.Rows[ItemSlotCount-1] != "POCKET" {
		t.Fatalf("清單的名稱對不上：%v", p.Rows)
	}
	if len(s.Inside.Icons) != 1 || s.Inside.Icons[0] != IconSprite(IconUTurn) {
		t.Fatalf("圖示列是 %v", s.Inside.Icons)
	}
}

func TestItemSelectionWrapsAround(t *testing.T) {
	s := itemSession(t)
	openItems(s)
	p := s.Inside.ItemPage
	s.MoveItemSelection(-1)
	if p.Selected != ItemSlotCount-1 {
		t.Fatalf("往上應該繞到最後一列，得到 %d", p.Selected)
	}
	s.MoveItemSelection(1)
	if p.Selected != 0 {
		t.Fatalf("往下應該繞回第一列，得到 %d", p.Selected)
	}
}

// 空的那一格印「You have nothing on your <部位>」；背包那六格換成 pocket 那一句。
func TestEmptySlotMessage(t *testing.T) {
	s := itemSession(t)
	i := s.speakerIndex()
	s.Party[i].Worn = [WornSlots]*Good{}
	s.Party[i].InHand = nil
	s.Party[i].Backpack = nil
	openItems(s)

	worn := s.Inside.ItemPage.Message
	if len(worn) < 2 || worn[len(worn)-1] != "Head" {
		t.Fatalf("身上那一格的訊息是 %v", worn)
	}
	s.Inside.ItemPage.Selected = ItemSlotCount - 1
	s.MoveItemSelection(0)
	pocket := s.Inside.ItemPage.Message
	if len(pocket) < 2 || pocket[len(pocket)-1] != "in your pocket!" {
		t.Fatalf("背包那一格的訊息是 %v", pocket)
	}
}

// 腰帶那一列印的是 Belt，手上那一列印的是 Weapon——兩格不是同一個。
func TestBeltAndWeaponAreDifferentRows(t *testing.T) {
	s := itemSession(t)
	i := s.speakerIndex()
	s.Party[i].Worn = [WornSlots]*Good{}
	s.Party[i].InHand = nil
	openItems(s)
	for _, c := range []struct {
		row  int
		want string
	}{{WornSlots - 1, "Belt"}, {WornSlots, "Weapon"}} {
		s.Inside.ItemPage.Selected = c.row
		s.MoveItemSelection(0)
		got := s.Inside.ItemPage.Message
		if len(got) == 0 || got[len(got)-1] != c.want {
			t.Fatalf("第 %d 列印的是 %v，期望 %s", c.row, got, c.want)
		}
	}
}

// 有東西的那一格顯示明細，與商店同一支（RE-075）。
func TestFilledSlotShowsDetail(t *testing.T) {
	s := itemSession(t)
	i := s.speakerIndex()
	s.Party[i].Worn[0] = &Good{Name: "Barbut Helm", Price: 250, Kind: 2,
		Position: 0, Encumbrance: 23, SizeClass: 8, Condition: 255}
	openItems(s)
	if s.Inside.ItemPage.Detail == nil {
		t.Fatal("有東西的那一格應該有明細")
	}
	if s.Inside.ItemPage.Message != nil {
		t.Fatalf("有東西就不該印空格那幾句：%v", s.Inside.ItemPage.Message)
	}
}

// U-TURN 回到角色資料頁，圖示列跟著回來。
func TestItemPageUTurnReturnsToSheet(t *testing.T) {
	s := itemSession(t)
	openItems(s)
	s.Inside.Selected = 0
	s.Accept()
	if s.ShowingItemPage() {
		t.Fatal("應該關掉物品頁")
	}
	if !s.ShowingSheet() {
		t.Fatal("應該回到角色資料頁")
	}
	for _, id := range s.Inside.Icons {
		if id == IconSprite(IconSheetItems) {
			return
		}
	}
	t.Fatalf("資料頁的圖示列沒回來：%v", s.Inside.Icons)
}
