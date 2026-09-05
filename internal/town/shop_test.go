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

// 商品表是空的就開不起清單——按下去只記一筆。
func TestEmptyShopNotesUnimplemented(t *testing.T) {
	s := shopSession(t, 3000, [2]int{0x21, 0x00}, nil)
	s.StartShop()
	if s.Shopping() || len(s.Unimplemented) != 1 {
		t.Fatalf("空商品表應該只記一筆未實作：%v", s.Unimplemented)
	}
}

// fitText 是合身那一段用得到的記錄。
func fitText() *UIText {
	t := shopText()
	t.Banks[0].Records = append(t.Banks[0].Records,
		uiRecord{Index: FitOfferRecord, Text: `"If you want this fitted@`},
		uiRecord{Index: CostPrefix, Text: "~H0Dit will cost you]"},
		uiRecord{Index: CostSuffix, Text: `@~H0Dgold, interested?"@`},
		uiRecord{Index: EquipRecord, Text: `"I'll put this on you."@`},
		uiRecord{Index: PackedRecord, Text: `"I'll pack it for you."@`})
	return t
}

// 尺寸剛好就直接穿上，不收錢也不問（原版的 `sub_157C6` 回 0）。
func TestShopFitsAndEquips(t *testing.T) {
	s := shopSession(t, 3000, [2]int{0x07, 0x02}, armourerGoods())
	s.UIText = fitText()
	s.Party[0].Sheet = &Sheet{Stats: [StatCount]int{0, 0, 100, 0, 0, 0, 0}}
	s.StartShop()
	yes(s) // 選毛皮帽
	yes(s) // 買
	if s.Party[0].Gold != 3000-32 {
		t.Fatalf("金幣是 %d", s.Party[0].Gold)
	}
	if s.Inside.Dialog[0] != `"I'll put this on you."` {
		t.Fatalf("畫的是 %q", s.Inside.Dialog)
	}
	if s.Party[0].Worn[0] == nil || s.Party[0].Worn[0].Name != "Fur Cap" {
		t.Fatalf("沒有穿上：%+v", s.Party[0].Worn[0])
	}
	if len(s.Party[0].Backpack) != 0 {
		t.Fatalf("穿上之後背包應該是空的，卻有 %d 件", len(s.Party[0].Backpack))
	}
}

// 尺寸不合就問要不要改：答應就付錢、改尺寸、穿上。
func TestShopFitCharges(t *testing.T) {
	s := shopSession(t, 3000, [2]int{0x07, 0x02}, armourerGoods())
	s.UIText = fitText()
	// 體型 150 → 毛皮帽該有的尺寸是 6 × 1 × 150 / 100 ＝ 9，記錄裡是 6。
	s.Party[0].Sheet = &Sheet{Stats: [StatCount]int{0, 0, 150, 0, 0, 0, 0}}
	s.StartShop()
	yes(s)
	yes(s)
	if s.Inside.Dialog[0] != `"If you want this fitted` {
		t.Fatalf("應該問要不要合身，畫的是 %q", s.Inside.Dialog)
	}
	if s.Inside.Dialog[1] != "it will cost you 5" {
		t.Fatalf("價錢那一行是 %q", s.Inside.Dialog[1])
	}
	yes(s)
	if s.Party[0].Gold != 3000-32-5 {
		t.Fatalf("金幣是 %d", s.Party[0].Gold)
	}
	if s.Party[0].Worn[0] == nil || s.Party[0].Worn[0].Encumbrance != 9 {
		t.Fatalf("穿上的是 %+v", s.Party[0].Worn[0])
	}
}

// 不改就留在背包，講記錄 0x5D。
func TestShopFitDeclined(t *testing.T) {
	s := shopSession(t, 3000, [2]int{0x07, 0x02}, armourerGoods())
	s.UIText = fitText()
	s.Party[0].Sheet = &Sheet{Stats: [StatCount]int{0, 0, 150, 0, 0, 0, 0}}
	s.StartShop()
	yes(s)
	yes(s)
	no(s)
	if s.Party[0].Worn[0] != nil {
		t.Fatal("不該穿上")
	}
	if len(s.Party[0].Backpack) != 1 {
		t.Fatalf("應該留在背包，背包有 %d 件", len(s.Party[0].Backpack))
	}
	if s.Inside.Dialog[0] != `"I'll pack it for you."` {
		t.Fatalf("畫的是 %q", s.Inside.Dialog)
	}
}

// 不會調整的店買完就結束，不問也不穿。
func TestShopWithoutFitting(t *testing.T) {
	s := shopSession(t, 3000, [2]int{0x07, 0x00}, armourerGoods())
	s.UIText = fitText()
	s.Party[0].Sheet = &Sheet{Stats: [StatCount]int{0, 0, 100, 0, 0, 0, 0}}
	s.StartShop()
	yes(s)
	yes(s)
	if s.Party[0].Worn[0] != nil || len(s.Party[0].Backpack) != 1 {
		t.Fatalf("不該穿上：穿 %+v 背包 %d", s.Party[0].Worn[0], len(s.Party[0].Backpack))
	}
	if s.Inside.List == nil {
		t.Fatal("買完應該回到清單")
	}
}

// adjustText 是甲冑調整用得到的記錄。
func adjustText() *UIText {
	t := fitText()
	t.Banks[0].Records = append(t.Banks[0].Records,
		uiRecord{Index: ArmourFitsRecord, Text: `"You don't need to fit!"`},
		uiRecord{Index: ArmourPrefixRecord, Text: `"The armor on your]`},
		uiRecord{Index: ArmourSuffixRecord, Text: `@~H0Dneeds to be adjusted!"`},
		uiRecord{Index: BodyPartRecord, Text: "head"},
		uiRecord{Index: BodyPartRecord + 1, Text: "torso"},
		uiRecord{Index: BodyPartRecord + 2, Text: "legs"})
	return t
}

// 身上的甲冑都合身就只講一句，不收錢也不問。
func TestAdjustNothingToDo(t *testing.T) {
	s := shopSession(t, 3000, [2]int{0x07, 0x02}, armourerGoods())
	s.UIText = adjustText()
	s.Party[0].Sheet = &Sheet{Stats: [StatCount]int{0, 0, 100, 0, 0, 0, 0}}
	cap := armourerGoods()[0]
	s.Party[0].Worn[0] = &cap
	s.StartTrain()
	if s.Shopping() {
		t.Fatal("都合身不該進問答")
	}
	if s.Inside.Dialog[0] != `"You don't need to fit!"` {
		t.Fatalf("畫的是 %q", s.Inside.Dialog)
	}
}

// 不合身就逐件講一句，講完問總價；答應之後扣錢並把尺寸改過來。
func TestAdjustCharges(t *testing.T) {
	s := shopSession(t, 3000, [2]int{0x07, 0x02}, armourerGoods())
	s.UIText = adjustText()
	// 體型 150：毛皮帽該有的尺寸 9（記錄 6）、毛皮上衣該有的 81（記錄 54）。
	s.Party[0].Sheet = &Sheet{Stats: [StatCount]int{0, 0, 150, 0, 0, 0, 0}}
	goods := armourerGoods()
	s.Party[0].Worn[0] = &goods[0]
	s.Party[0].Worn[1] = &goods[1]
	s.StartTrain()
	if s.Inside.Dialog[0] != `"The armor on your head` {
		t.Fatalf("第一句是 %q", s.Inside.Dialog)
	}
	s.Continue() // 讀完第一件
	if s.Inside.Dialog[0] != `"The armor on your torso` {
		t.Fatalf("第二句是 %q", s.Inside.Dialog)
	}
	s.Continue() // 讀完第二件，出現總價
	// 頭 5 ＋ 軀幹 50 ＝ 55。
	if s.Inside.Dialog[1] != "it will cost you 55" {
		t.Fatalf("總價那一行是 %q", s.Inside.Dialog[1])
	}
	yes(s)
	if s.Party[0].Gold != 3000-55 {
		t.Fatalf("金幣是 %d", s.Party[0].Gold)
	}
	if s.Party[0].Worn[0].Encumbrance != 9 || s.Party[0].Worn[1].Encumbrance != 81 {
		t.Fatalf("改完的尺寸是 %d／%d",
			s.Party[0].Worn[0].Encumbrance, s.Party[0].Worn[1].Encumbrance)
	}
}

// 錢不夠就講記錄 0x56，尺寸不變。
func TestAdjustNotEnoughGold(t *testing.T) {
	// 改一頂帽子要 5 枚，身上只有 1 枚。
	s := shopSession(t, 1, [2]int{0x07, 0x02}, armourerGoods())
	s.UIText = adjustText()
	s.Party[0].Sheet = &Sheet{Stats: [StatCount]int{0, 0, 150, 0, 0, 0, 0}}
	goods := armourerGoods()
	s.Party[0].Worn[0] = &goods[0]
	s.StartTrain()
	s.Continue()
	yes(s)
	if s.Party[0].Gold != 1 || s.Party[0].Worn[0].Encumbrance != 6 {
		t.Fatalf("買不起卻改了狀態：金幣 %d 尺寸 %d",
			s.Party[0].Gold, s.Party[0].Worn[0].Encumbrance)
	}
}

// stableGoods 是原版五間馬廄共用的那四匹（RE-119）。
func stableGoods() []Good {
	return []Good{
		{Index: 0, Name: "Draft Horse", Price: 200, Picture: 0, Movement: 10},
		{Index: 1, Name: "Lgt Warhorse", Price: 400, Picture: 1, Movement: 4},
		{Index: 2, Name: "Med Warhorse", Price: 800, Picture: 2, Movement: 6},
		{Index: 3, Name: "Hvy Warhorse", Price: 1600, Picture: 3, Movement: 8},
	}
}

// 馬廄（旗標 bit5）開得起清單，選一匹之後問的是價錢那一句，不是商品明細。
func TestStableBuysASteed(t *testing.T) {
	s := shopSession(t, 3000, [2]int{0x20, 0x00}, stableGoods())
	s.StartShop()
	if !s.Shopping() {
		t.Fatal("按下 TRADE 應該進馬廄")
	}
	if s.Inside.List == nil || s.Inside.List.Rows[0] != "Draft Horse" {
		t.Fatalf("清單是 %+v", s.Inside.List)
	}
	// 選第二匹（400）：明細那一格空著，問句帶著價錢。
	s.MoveShop(1)
	yes(s)
	if s.Inside.Detail != nil {
		t.Fatal("馬廄那一頁沒有商品明細")
	}
	if got := s.Inside.Dialog; len(got) == 0 {
		t.Fatal("問句是空的")
	}
	// 答應：扣掉 400，坐騎記到發言者身上。
	yes(s)
	if s.Party[0].Gold != 2600 {
		t.Fatalf("錢包是 %d，應為 2600", s.Party[0].Gold)
	}
	if s.Party[0].Steed == nil || s.Party[0].Steed.Name != "Lgt Warhorse" {
		t.Fatalf("坐騎是 %+v", s.Party[0].Steed)
	}
}

// 買下之後資料頁的移動率換成那一匹的（角色記錄 +0x1BA 落在坐騎那十六個 byte
// 裡）；名冊上那一份不能跟著動。
func TestStableSetsMovementRate(t *testing.T) {
	s := shopSession(t, 3000, [2]int{0x20, 0x00}, stableGoods())
	shared := theodoric()
	s.Party[0].Sheet = shared
	s.StartShop()
	s.MoveShop(1)
	yes(s)
	yes(s)
	if s.Party[0].Sheet.MovementRate != 4 {
		t.Fatalf("移動率是 %d，輕型戰馬應為 4", s.Party[0].Sheet.MovementRate)
	}
	if shared.MovementRate != 0 {
		t.Fatalf("名冊那一份被改成 %d", shared.MovementRate)
	}
}

// 身上已經有一匹就以舊換新：折抵舊的一半，折完不剩才收 50（`0x12B79`）。
func TestStableTradesInTheOldSteed(t *testing.T) {
	s := shopSession(t, 3000, [2]int{0x20, 0x00}, stableGoods())
	old := stableGoods()[3] // Hvy Warhorse，1600
	s.Party[0].Steed = &old
	// Draft Horse 標價 200，折抵 800 → 剩負的，收 50。
	if got := s.steedPrice(stableGoods()[0]); got != SteedFloor {
		t.Fatalf("折抵之後是 %d，應為 %d", got, SteedFloor)
	}
	// Hvy Warhorse 標價 1600，折抵 800 → 800。
	if got := s.steedPrice(stableGoods()[3]); got != 800 {
		t.Fatalf("折抵之後是 %d，應為 800", got)
	}
	// 差額落在 1…49 時照收差額，不是收 50（原版比的是差額大不大於 0）。
	cheap := Good{Index: 9, Name: "Nag", Price: 180}
	s.Party[0].Steed = &Good{Index: 9, Name: "Nag", Price: 300}
	if got := s.steedPrice(cheap); got != 30 {
		t.Fatalf("差額 30 收成 %d", got)
	}
	s.Party[0].Steed = &old
	s.StartShop()
	s.MoveShop(3)
	yes(s)
	yes(s)
	if s.Party[0].Gold != 2200 {
		t.Fatalf("錢包是 %d，應為 3000 − 800 ＝ 2200", s.Party[0].Gold)
	}
	if s.Party[0].Steed == nil || s.Party[0].Steed.Name != "Hvy Warhorse" {
		t.Fatalf("坐騎是 %+v", s.Party[0].Steed)
	}
}

// 錢不夠就退回清單，坐騎不變。
func TestStableRefusesWhenBroke(t *testing.T) {
	s := shopSession(t, 100, [2]int{0x20, 0x00}, stableGoods())
	s.StartShop()
	yes(s)
	yes(s)
	if s.Party[0].Gold != 100 {
		t.Fatalf("錢不夠不該扣錢，剩 %d", s.Party[0].Gold)
	}
	if s.Party[0].Steed != nil {
		t.Fatalf("錢不夠不該給坐騎：%+v", s.Party[0].Steed)
	}
}
