package town

import "testing"

// armourerGoods 是原版城 1 魯德曼甲冑鋪的前三件，欄位直接抄自建築 payload。
func armourerGoods() []Good {
	return []Good{
		{Index: 0, Name: "Fur Cap", Price: 32, Kind: 2, Position: 0,
			Encumbrance: 6, Condition: 255,
			Dice: Dice{Count: 1, Sides: 6, Low: 1, High: 6}},
		{Index: 1, Name: "Fur Shirt", Price: 185, Kind: 2, Position: 1,
			Encumbrance: 54, Condition: 255,
			Dice: Dice{Count: 1, Sides: 6, Low: 1, High: 6}},
		{Index: 2, Name: "Fur Pants", Price: 180, Kind: 2, Position: 2,
			Encumbrance: 36, Condition: 255,
			Dice: Dice{Count: 1, Sides: 6, Low: 1, High: 6}},
	}
}

func shopSession(t *testing.T, gold int, flags [2]int, goods []Good) *Session {
	t.Helper()
	s := restSession(t, flags, 0, 0, Member{Name: "ddduhh", Gold: gold, Speaker: true})
	s.Buildings.Doors[0].Goods = goods
	s.door = &s.Buildings.Doors[0]
	return s
}

// 按下 TRADE 換成商品清單，圖示列是大拇指朝上與 U-TURN。
func TestShopListOpens(t *testing.T) {
	s := shopSession(t, 3000, [2]int{0x07, 0x02}, armourerGoods())
	s.StartShop()
	if !s.Shopping() {
		t.Fatal("按下 TRADE 應該進商店")
	}
	if s.Inside.List == nil || s.Inside.List.Rows[0] != "Fur Cap" {
		t.Fatalf("清單是 %+v", s.Inside.List)
	}
	icons, _ := s.Icons()
	if len(icons) != 2 || icons[0] != IconSprite(IconYes) || icons[1] != IconSprite(IconUTurn) {
		t.Fatalf("清單的圖示列是 %v", icons)
	}
}

// 明細面板的每一欄都由記錄算出來。
func TestShopDetailFields(t *testing.T) {
	s := shopSession(t, 3000, [2]int{0x07, 0x02}, armourerGoods())
	s.UIText = shopText()
	s.StartShop()
	yes(s) // 選第一件
	got := map[string]bool{}
	for _, seg := range s.Inside.Detail {
		got[seg.Text] = true
	}
	for _, want := range []string{"Fur Cap", "Typ:Wear", "Pos:Head", "Prt:1-6",
		"Enc:6", "GC :32", "Cnd:New"} {
		if !got[want] {
			t.Fatalf("明細少了 %q：%+v", want, s.Inside.Detail)
		}
	}
	icons, _ := s.Icons()
	if len(icons) != 2 || icons[1] != IconSprite(IconNo) {
		t.Fatalf("確認的圖示列是 %v", icons)
	}
}

// shopText 是明細面板用得到的那幾筆記錄。
func shopText() *UIText {
	recs := []uiRecord{
		{Index: AskBuyRecord, Text: `"Do you want this?"@@`},
		{Index: NotEnoughRecord, Text: `"Sorry, not enough@~H0Dgold crowns!"`},
		{Index: BackpackFullRecord, Text: `~H0D"Your backpack is full!"`},
		{Index: TypeLabelRecord, Text: "Typ:"},
		{Index: HandLabelRecord, Text: "Hnd:"},
		{Index: PositionLabelRecord, Text: "Pos:"},
		{Index: DamageLabelRecord, Text: "Dmg:"},
		{Index: ProtectLabelRecord, Text: "Prt:"},
		{Index: EncumbranceRecord, Text: "Enc:"},
		{Index: GoldLabelRecord, Text: "GC :"},
		{Index: NutritionLabelRecord, Text: "Nut:"},
		{Index: ConditionLabelRecord, Text: "Cnd:"},
	}
	for i, name := range []string{"Item", "Wield", "Wear", "Both"} {
		recs = append(recs, uiRecord{Index: TypeNameRecord + i, Text: name})
	}
	for i, name := range []string{"One", "Two"} {
		recs = append(recs, uiRecord{Index: HandNameRecord + i, Text: name})
	}
	for i, name := range []string{"Bad", "Poor", "Worn", "New"} {
		recs = append(recs, uiRecord{Index: ConditionNameRecord + i, Text: name})
	}
	for i, name := range []string{"Head", "Trso", "Legs", "Body", "Hnds", "Feet",
		"Shld", "Neck", "Fngr", "Belt", "Food"} {
		recs = append(recs, uiRecord{Index: PositionNameRecord + i, Text: name})
	}
	return &UIText{Schema: "kol-ui-text-data-v1",
		Banks: []uiBank{{Bank: BankUI, Records: recs}}}
}

// 買下之後扣錢、東西進背包，畫面回到清單。
func TestShopBuy(t *testing.T) {
	s := shopSession(t, 3000, [2]int{0x07, 0x00}, armourerGoods())
	s.UIText = shopText()
	s.StartShop()
	yes(s) // 選第一件
	yes(s) // 買
	if s.Party[0].Gold != 3000-32 {
		t.Fatalf("買完剩 %d", s.Party[0].Gold)
	}
	if len(s.Party[0].Backpack) != 1 || s.Party[0].Backpack[0].Name != "Fur Cap" {
		t.Fatalf("背包是 %+v", s.Party[0].Backpack)
	}
	if s.Inside.List == nil {
		t.Fatal("買完應該回到清單")
	}
}

// 錢不夠就講記錄 0x56，不扣錢也不進背包。
func TestShopNotEnoughGold(t *testing.T) {
	s := shopSession(t, 10, [2]int{0x07, 0x00}, armourerGoods())
	s.UIText = shopText()
	s.StartShop()
	yes(s)
	yes(s)
	if s.Party[0].Gold != 10 || len(s.Party[0].Backpack) != 0 {
		t.Fatalf("買不起卻改了狀態：金幣 %d 背包 %d",
			s.Party[0].Gold, len(s.Party[0].Backpack))
	}
	if s.Inside.Dialog[0] != `"Sorry, not enough` {
		t.Fatalf("畫的是 %q", s.Inside.Dialog)
	}
}

// 背包六格滿了就講記錄 0xA2；錢已經付掉了，與原版相同。
func TestShopBackpackFull(t *testing.T) {
	s := shopSession(t, 3000, [2]int{0x07, 0x00}, armourerGoods())
	s.UIText = shopText()
	for i := 0; i < BackpackSlots; i++ {
		s.Party[0].Backpack = append(s.Party[0].Backpack, Good{Name: "x"})
	}
	s.StartShop()
	yes(s)
	yes(s)
	if len(s.Party[0].Backpack) != BackpackSlots {
		t.Fatalf("背包變成 %d 格", len(s.Party[0].Backpack))
	}
	if s.Inside.Dialog[0] != `"Your backpack is full!"` {
		t.Fatalf("畫的是 %q", s.Inside.Dialog)
	}
}

// 清單的反白兩端夾住，超過八列才捲。
func TestShopListScrolls(t *testing.T) {
	goods := make([]Good, 0, 12)
	for i := 0; i < 12; i++ {
		goods = append(goods, Good{Index: i, Name: string(rune('a' + i))})
	}
	s := shopSession(t, 3000, [2]int{0x07, 0x00}, goods)
	s.StartShop()
	for i := 0; i < 20; i++ {
		s.MoveShop(1)
	}
	if s.shop.cursor != 11 || s.shop.top != 4 {
		t.Fatalf("按到底停在 %d，視窗第一列 %d", s.shop.cursor, s.shop.top)
	}
	for i := 0; i < 20; i++ {
		s.MoveShop(-1)
	}
	if s.shop.cursor != 0 || s.shop.top != 0 {
		t.Fatalf("按回頂端是 %d／%d", s.shop.cursor, s.shop.top)
	}
}

// 馬廄的商品表版面不同，還沒實作——按下去只記一筆。
func TestStableNotImplemented(t *testing.T) {
	s := shopSession(t, 3000, [2]int{0x21, 0x00}, nil)
	s.StartShop()
	if s.Shopping() || len(s.Unimplemented) != 1 {
		t.Fatalf("馬廄應該只記一筆未實作：%v", s.Unimplemented)
	}
}
