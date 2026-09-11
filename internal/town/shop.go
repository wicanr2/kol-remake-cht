package town

import (
	"strconv"
	"strings"
)

// 商店（RE-075）。TRADE 圖示由建築旗標的 bit1／bit4／bit5 給；bit5 是馬廄、
// bit4 是賣食物的店。商品表在建築 payload 的 0x100 起，一筆 16 byte；
// 清單用的是招募同伴那一支同樣的元件，明細欄位與骰子算式見 RE-075。

// 商店會用到的 KOLDATA 記錄。
const (
	// AskBuyRecord 是選定商品之後的 `"Do you want this?"`。
	AskBuyRecord = 0x31
	// AskEatRecord 是酒館買完之後的 `"Eat here?"`。
	AskEatRecord = 0x32
	// BackpackFullRecord 是 `"Your backpack is full!"`。
	BackpackFullRecord = 0xA2
	// 明細面板的標籤與名稱表。
	TypeLabelRecord      = 0xA3
	TypeNameRecord       = 0xA4 // Item／Wield／Wear／Both
	HandLabelRecord      = 0xA8
	HandNameRecord       = 0xA9 // One／Two
	PositionLabelRecord  = 0xAB
	DamageLabelRecord    = 0xAC
	ProtectLabelRecord   = 0xAD
	EncumbranceRecord    = 0xAE
	GoldLabelRecord      = 0xAF
	NutritionLabelRecord = 0xB0
	ConditionLabelRecord = 0xB1
	FitLabelRecord       = 0xB2
	NutritionNameRecord  = 0xB3 // Poor／Fair／Good／Great
	ConditionNameRecord  = 0xB7 // Bad／Poor／Worn／New
	PositionNameRecord   = 0xBB // Head…Food
	// FitOfferRecord 是 `"If you want this fitted`。
	FitOfferRecord = 0x33
	// 馬廄那三句（RE-119）：明細那一頁的 `"If you want this steed`，
	// 以及交出舊坐騎的 `"Give me your reins."` 與交付的 `"Take these reins."`。
	SteedOfferRecord = 0xCD
	GiveReinsRecord  = 0x65
	TakeReinsRecord  = 0x66
	// EquipRecord 是 `"I'll put this on you."`；PackedRecord 是 `"I'll pack it for you."`。
	EquipRecord  = 0x5C
	PackedRecord = 0x5D
)

// BackpackSlots 是背包格數：角色記錄 0xA0 起、一格 16 byte、共六格。
const BackpackSlots = 6

// ShieldPosition 是盾牌的穿戴部位編號；只有它的防護顯示成百分比。
const ShieldPosition = 6

// FoodPosition 是食物的部位編號；只有它顯示營養。
const FoodPosition = 0x0A

// Dice 是傷害或防護的骰子範圍。
type Dice struct {
	Count int `json:"count"`
	Sides int `json:"sides"`
	Bonus int `json:"bonus"`
	Low   int `json:"low"`
	High  int `json:"high"`
}

// Good 是一件商品，也是背包裡的一格——商店賣的與背包裝的是同一種記錄。
type Good struct {
	Index  int    `json:"index"`
	Name   string `json:"name"`
	Price  int    `json:"price"`
	Kind   int    `json:"kind"`
	Damage bool   `json:"damage"`
	Hands  int    `json:"hands"`
	// Class 是物品記錄 +0x0D 整個 byte：戰鬥比對它挑武器技能列與攻擊式（RE-080）。
	Class       int  `json:"class"`
	Position    int  `json:"position"`
	Encumbrance int  `json:"encumbrance"`
	Condition   int  `json:"condition"`
	Percent     int  `json:"percent"`
	Dice        Dice `json:"dice"`
	// SizeClass 是 rec[14] 的低 nibble：合身的三張表都用它當索引（RE-075）。
	SizeClass int    `json:"size_class"`
	Record    string `json:"record"`
	// Picture 只有馬廄那四匹有：記錄第 0 個 byte，是 `STEEDPIC` 的圖號（RE-119）。
	Picture int `json:"picture,omitempty"`
	// Movement 也只有那四匹有：記錄 `+0x7A`。買下之後整段 `+0x70`…`+0x7F` 抄進
	// 角色記錄 `+0x1B0`，而算隊伍腳程的 `sub_29004:0x290A6` 讀的就是 `+0x1BA`
	// ＝ 這一格。數字越大越慢：輕型戰馬 4、中型 6、重型 8、駄馬 10，用走的是 20。
	Movement int `json:"movement,omitempty"`
}

// ConditionStep 是狀態名稱的檔次：超過 0x20／0x40／0x80 各加一。
func (g Good) ConditionStep() int {
	step := 0
	for _, threshold := range [...]int{0x20, 0x40, 0x80} {
		if g.Condition > threshold {
			step++
		}
	}
	return step
}

// NutritionStep 是營養名稱的檔次。
func (g Good) NutritionStep() int { return g.Condition / 25 }

// shopStage 是商店流程停在哪一步。
type shopStage int

const (
	shopList shopStage = iota
	shopConfirm
	shopEat
	// shopFit 是「要合身嗎」那一問。
	shopFit
	// shopAdjust 是 TRAIN 的甲冑調整。
	shopAdjust
)

// shopState 是一次逛商店的狀態。
type shopState struct {
	goods  []Good
	cursor int
	top    int
	stage  shopStage
	// next 是「按空白鍵讀完這一段之後」要做的事。
	next func()
	// member／slot／fit 是合身那一問的對象與該有的尺寸；
	// 甲冑調整時 fit 是總價、adjust 是要改的那幾件。
	member, slot, fit int
	adjust            []adjustEntry
}

// Shopping 回報現在是不是停在商店的畫面。
func (s *Session) Shopping() bool { return s.shop != nil }

// StartShop 是按下 TRADE。
func (s *Session) StartShop() {
	if s.Mode != ModeInside || s.Inside == nil || s.door == nil {
		return
	}
	if len(s.door.Goods) == 0 {
		s.note(iconName(IconTrade))
		return
	}
	s.shop = &shopState{goods: s.door.Goods}
	s.showShopList()
}

// showShopList 把對話框換成商品清單，圖示列是大拇指朝上與 U-TURN，
// 與招募同伴那一份相同（RE-072／RE-075）。
func (s *Session) showShopList() {
	s.shop.stage = shopList
	s.Inside.Dialog = nil
	s.Inside.DialogZh = nil
	s.Inside.Detail = nil
	s.clearSteedPic()
	s.Inside.List = s.shopView()
	s.setInsideIcons([]int{IconYes, IconUTurn})
}

// shopView 是清單視窗內那幾列。
func (s *Session) shopView() *InnList {
	sh := s.shop
	end := sh.top + ListRows
	if end > len(sh.goods) {
		end = len(sh.goods)
	}
	rows := make([]string, 0, end-sh.top)
	for _, g := range sh.goods[sh.top:end] {
		rows = append(rows, g.Name)
	}
	return &InnList{Rows: rows, Selected: sh.cursor - sh.top, Total: len(sh.goods)}
}

// MoveShop 是清單上的方向鍵。兩端夾住不繞回，視窗跟著捲。
func (s *Session) MoveShop(delta int) {
	sh := s.shop
	if sh == nil || sh.stage != shopList || len(sh.goods) == 0 {
		return
	}
	sh.cursor += delta
	if sh.cursor < 0 {
		sh.cursor = 0
	}
	if sh.cursor >= len(sh.goods) {
		sh.cursor = len(sh.goods) - 1
	}
	if sh.cursor < sh.top {
		sh.top = sh.cursor
	}
	if sh.cursor >= sh.top+ListRows {
		sh.top = sh.cursor - ListRows + 1
	}
	s.Inside.List = s.shopView()
}

// shopAnswer 是商店流程裡對圖示的回應。
func (s *Session) shopAnswer(id int) {
	sh := s.shop
	if sh == nil {
		return
	}
	switch sh.stage {
	case shopList:
		if id == IconUTurn {
			s.endShop()
			return
		}
		s.showGood()
	case shopConfirm:
		if id != IconYes {
			s.showShopList()
			return
		}
		s.buy()
	case shopEat:
		// 兩個答案都已經付過錢了：答是就當場吃掉，答否才進背包。
		if id == IconYes {
			s.finishBuy(true)
			return
		}
		s.finishBuy(false)
	case shopFit:
		s.fitAnswer(id == IconYes)
	case shopAdjust:
		s.adjustAnswer(id == IconYes)
	}
}

// showGood 把對話框換成商品明細與「要買嗎」。
//
// 馬廄那一頁不一樣（RE-119）：沒有明細欄位，問句是記錄 0xCD ＋ 0x53 ＋ 價錢 ＋ 0x54
// （與住宿、治療、訓練同一套），而價錢先扣掉身上那一匹的一半。
func (s *Session) showGood() {
	s.shop.stage = shopConfirm
	g := s.shop.goods[s.shop.cursor]
	s.Inside.List = nil
	s.Inside.DialogZh = nil
	if s.door.IsStable() {
		s.Inside.Detail = nil
		s.Inside.Steed = &SteedPic{Picture: g.Picture, Name: g.Name}
		s.Inside.DialogRow = StablePic.AskRow
		s.Inside.Dialog = s.priceLines(SteedOfferRecord, s.steedPrice(g))
		s.setInsideIcons([]int{IconYes, IconNo})
		return
	}
	s.Inside.Dialog = nonEmpty(s.uiLines(AskBuyRecord))
	s.Inside.Detail = s.goodDetail(g)
	s.setInsideIcons([]int{IconYes, IconNo})
}

// SteedFloor 是折抵之後一毛不剩時收的價（`0x12B79` 的 0x32）。
const SteedFloor = 50

// steedPrice 是這一匹對發言者的實收價（`0x12B1C`…`0x12B94`，RE-119）：
// 標價減掉身上那一匹價錢的一半，**減完不大於 0 才收 50**——原版的判準是
// 差額本身，不是「不足 50」，所以差額 1…49 照收差額。四匹的價錢是 200 起
// 的倍增數列，這兩種寫法在原版資料上結果一樣，但公式要照原版。
func (s *Session) steedPrice(g Good) int {
	price := g.Price
	if i := s.speakerIndex(); i >= 0 && s.Party[i].Steed != nil {
		price -= s.Party[i].Steed.Price / 2
	}
	if price <= 0 {
		return SteedFloor
	}
	return price
}

// buy 是答應要買：先看錢夠不夠。
func (s *Session) buy() {
	g := s.shop.goods[s.shop.cursor]
	i := s.speakerIndex()
	if i < 0 {
		s.endShop()
		return
	}
	if s.door.IsStable() {
		s.buySteed(i, g)
		return
	}
	if s.Party[i].Gold < g.Price {
		s.Inside.Detail = nil
		s.shopSay(nonEmpty(s.uiLines(NotEnoughRecord)), s.showShopList)
		return
	}
	s.Party[i].Gold -= g.Price
	s.syncParty()
	if s.door.IsTavern() {
		s.shop.stage = shopEat
		s.Inside.Detail = nil
		s.Inside.Dialog = nonEmpty(s.uiLines(AskEatRecord))
		s.Inside.DialogZh = nil
		s.setInsideIcons([]int{IconYes, IconNo})
		return
	}
	s.finishBuy(false)
}

// buySteed 是馬廄那一條（`0x12C11`…`0x12CF2`，RE-119）：錢不夠就照商店那一句退回
// 清單；錢夠就扣錢，身上本來有一匹先講記錄 0x65（交出韁繩），再講 0x66 交付新的，
// 最後把那一匹記進角色。
//
// **這一版沒有搬圖**：原版還會把 `STEEDPIC` 那張圖抄進角色圖像區（一槽 0x2F8
// byte），而 remake 的坐騎圖是照 `Picture` 現查的，不必先搬。
//
// 抄進去的十六個 byte 裡有一格是移動率（`+0x1BA`），隊伍腳程與「要不要騎走」
// 那一問都看它（RE-112／RE-119），所以買下的同時要改資料頁上的那個數字。
func (s *Session) buySteed(i int, g Good) {
	price := s.steedPrice(g)
	if s.Party[i].Gold < price {
		s.Inside.Detail = nil
		s.shopSay(nonEmpty(s.uiLines(NotEnoughRecord)), s.showShopList)
		return
	}
	s.Party[i].Gold -= price
	lines := nonEmpty(s.uiLines(TakeReinsRecord))
	if s.Party[i].Steed != nil {
		lines = append(nonEmpty(s.uiLines(GiveReinsRecord)), lines...)
	}
	steed := g
	s.Party[i].Steed = &steed
	// 資料頁那一份與招募清單共用指標，複製一份再改移動率。
	if sheet := s.Party[i].Sheet; sheet != nil {
		copied := *sheet
		copied.MovementRate = g.Movement
		s.Party[i].Sheet = &copied
	}
	s.syncParty()
	s.Inside.Detail = nil
	s.shopSay(lines, s.showShopList)
}

// finishBuy 把東西放進背包；當場吃掉就不佔格子。背包滿了要講記錄 0xA2。
func (s *Session) finishBuy(eaten bool) {
	g := s.shop.goods[s.shop.cursor]
	if eaten {
		s.showShopList()
		return
	}
	i := s.speakerIndex()
	if i < 0 {
		s.endShop()
		return
	}
	if len(s.Party[i].Backpack) >= BackpackSlots {
		s.Inside.Detail = nil
		s.shopSay(nonEmpty(s.uiLines(BackpackFullRecord)), s.showShopList)
		return
	}
	s.Party[i].Backpack = append(s.Party[i].Backpack, g)
	s.syncParty()
	s.offerFit(i, len(s.Party[i].Backpack)-1)
}

// clearSteedPic 把馬廄那一頁的坐騎圖與挪下去的問句位置收回來。
func (s *Session) clearSteedPic() {
	if s.Inside == nil {
		return
	}
	s.Inside.Steed = nil
	s.Inside.DialogRow = 0
}

// shopSay 顯示一段話，讀完之後接著做 next。
func (s *Session) shopSay(lines []string, next func()) {
	s.clearSteedPic()
	s.say(lines, nil)
	if s.shop != nil {
		s.shop.next = next
	}
}

// endShop 離開商店，回到建築內的一般畫面。
func (s *Session) endShop() {
	s.shop = nil
	if s.Inside == nil || s.door == nil {
		return
	}
	s.Inside.List = nil
	s.Inside.Detail = nil
	s.clearSteedPic()
	s.Inside.Dialog = s.greeting()
	s.Inside.DialogZh = nil
	s.setInsideIcons(InteriorIcons(s.door.IconFlags, len(s.Party)))
}

// 明細面板的版面（`sub_153FF`，引數是對話框的文字欄 13）。原版先畫記錄 0x31
// 的兩行，游標因此落在對話框的第三列；名稱置中，其餘欄位排成兩欄。
const (
	detailNameRow  = 102
	detailFirstRow = 118
	detailRowStep  = 8
	detailLeftCol  = 16
	detailRightCol = 26
	detailNameCol  = 14
	detailCells    = 24
)

// goodDetail 把一件商品排成帶座標的文字。標籤與名稱都走譯文優先。
func (s *Session) goodDetail(g Good) []Segment {
	row := detailFirstRow
	var out []Segment
	add := func(col int, text string) {
		if text != "" {
			out = append(out, Segment{Col: col, Row: row, Text: text})
		}
	}
	name := g.Name
	out = append(out, Segment{
		Col: detailNameCol + (detailCells-len([]rune(name)))/2,
		Row: detailNameRow, Text: name})

	add(detailLeftCol, s.uiRecord(TypeLabelRecord)+s.uiRecord(TypeNameRecord+g.Kind))
	switch {
	case g.Kind == 1:
		add(detailRightCol, s.uiRecord(HandLabelRecord)+s.uiRecord(HandNameRecord+g.Hands))
	case g.Kind != 0:
		add(detailRightCol,
			s.uiRecord(PositionLabelRecord)+s.uiRecord(PositionNameRecord+g.Position))
	}

	row += detailRowStep
	label := s.uiRecord(ProtectLabelRecord)
	if g.Damage {
		label = s.uiRecord(DamageLabelRecord)
	}
	add(detailLeftCol, label+g.Range())
	add(detailRightCol, s.uiRecord(EncumbranceRecord)+itoa(g.Encumbrance))

	row += detailRowStep
	add(detailLeftCol, s.uiRecord(GoldLabelRecord)+itoa(g.Price))
	if g.Position == FoodPosition && g.Kind == 0 {
		add(detailRightCol,
			s.uiRecord(NutritionLabelRecord)+s.uiRecord(NutritionNameRecord+g.NutritionStep()))
	} else {
		add(detailRightCol,
			s.uiRecord(ConditionLabelRecord)+s.uiRecord(ConditionNameRecord+g.ConditionStep()))
	}
	return out
}

// Range 是傷害或防護那一欄的字串：盾牌是百分比，其餘是骰子範圍。
func (g Good) Range() string {
	if g.Position == ShieldPosition {
		return itoa(g.Percent) + "%"
	}
	return itoa(g.Dice.Low) + "-" + itoa(g.Dice.High)
}

// itoa 是數字轉字串，原版用的是 `%d`。
func itoa(v int) string { return strconv.Itoa(v) }

// 合身與穿戴（RE-075）。買完之後只有「會調整的店」（旗標 byte 1 的 bit1）
// 遇到頭／軀幹／腿的穿戴品才走這一段，其餘情形買完就結束。

// WornSlots 是穿戴槽位數：部位 0…9 各一格。原版是角色記錄 0x100 + 16 × 部位，
// 部位 9（腰帶）例外，換到 0x30 那一格。
const WornSlots = 10

// 合身要用的三張表都在 KOLTOWN 的 DGROUP（RE-075）。
var (
	// fitCost 是合身的價錢：[部位][rec[14] 的低 nibble]。
	fitCost = [3][9]int{
		{5, 5, 10, 10, 25, 25, 25, 50, 80},
		{50, 50, 100, 100, 250, 250, 250, 500, 800},
		{30, 30, 60, 60, 150, 150, 150, 300, 480},
	}
	// fitSizeFactor 是 DGROUP+0x116，用低 nibble 查。
	fitSizeFactor = [9]int{6, 5, 8, 10, 10, 12, 15, 15, 20}
	// fitPositionFactor 是 DGROUP+0x120，用部位查。
	fitPositionFactor = [3]int{1, 9, 6}
)

// FitPositions 是需要合身的部位上限：0 頭、1 軀幹、2 腿。
const FitPositions = 3

// FitSize 是這件衣物穿在這名角色身上該有的尺寸。
// `[0x116 + 低 nibble] × [0x120 + 部位] × 體型 / 100`。
func FitSize(g Good, size int) int {
	if g.Position >= FitPositions {
		return 0
	}
	return fitSizeFactor[g.Kind4()] * fitPositionFactor[g.Position] * size / 100
}

// Kind4 是 rec[14] 的低 nibble，合身的三張表都用它當索引。
func (g Good) Kind4() int {
	if g.SizeClass < 0 || g.SizeClass >= len(fitSizeFactor) {
		return 0
	}
	return g.SizeClass
}

// FitCost 是把這件衣物改到合身要花的錢。
func FitCost(g Good) int {
	if g.Position >= FitPositions {
		return 0
	}
	return fitCost[g.Position][g.Kind4()]
}

// offerFit 是買完之後的合身與穿戴。只有會調整的店（旗標 byte 1 的 bit1）
// 遇到頭／軀幹／腿的東西才走這一段；合身就直接穿上，不合身才問要不要改。
func (s *Session) offerFit(member, slot int) {
	g := s.Party[member].Backpack[slot]
	if s.door.IconFlags[1]&0x02 == 0 || g.Position >= FitPositions {
		s.showShopList()
		return
	}
	want := FitSize(g, s.memberSize(member))
	if want == g.Encumbrance {
		// 本來就合身：直接穿上。
		s.equip(member, slot)
		return
	}
	s.shop.stage = shopFit
	s.shop.member, s.shop.slot, s.shop.fit = member, slot, want
	s.Inside.Detail = nil
	s.Inside.Dialog = s.priceLines(FitOfferRecord, FitCost(g))
	s.Inside.DialogZh = nil
	s.setInsideIcons([]int{IconYes, IconNo})
}

// fitAnswer 是「要合身嗎」的回答。
func (s *Session) fitAnswer(yes bool) {
	sh := s.shop
	member, slot := sh.member, sh.slot
	g := s.Party[member].Backpack[slot]
	cost := FitCost(g)
	if !yes {
		s.packIt()
		return
	}
	if s.Party[member].Gold < cost {
		s.shopSay(nonEmpty(s.uiLines(NotEnoughRecord)), s.packIt)
		return
	}
	s.Party[member].Gold -= cost
	s.Party[member].Backpack[slot].setEncumbrance(sh.fit)
	s.syncParty()
	s.equip(member, slot)
	s.Party[member].SyncCarried()
}

// packIt 是不改也不穿：東西留在背包裡（記錄 0x5D）。
func (s *Session) packIt() {
	s.Inside.Detail = nil
	s.shopSay(nonEmpty(s.uiLines(PackedRecord)), s.showShopList)
}

// equip 把背包那一格與對應的穿戴槽對調，與原版 `sub_19A87` 相同。
func (s *Session) equip(member, slot int) {
	m := &s.Party[member]
	g := m.Backpack[slot]
	if g.Position < 0 || g.Position >= WornSlots {
		s.showShopList()
		return
	}
	worn := m.Worn[g.Position]
	m.Backpack = append(m.Backpack[:slot], m.Backpack[slot+1:]...)
	if worn != nil {
		m.Backpack = append(m.Backpack, *worn)
	}
	item := g
	m.Worn[g.Position] = &item
	s.syncParty()
	s.Inside.Detail = nil
	s.shopSay(nonEmpty(s.uiLines(EquipRecord)), s.showShopList)
}

// memberSize 是角色的體型（角色記錄 0x1F）。沒有資料頁就當 0，合身一定不成立。
func (s *Session) memberSize(i int) int {
	if s.Party[i].Sheet == nil {
		return 0
	}
	return s.Party[i].Sheet.Stats[StatSize]
}

// priceLines 組出「開場 ＋ 要花你 N 枚金幣，有意思嗎？」，與住宿、治療同一套。
func (s *Session) priceLines(intro, price int) []string {
	lines := nonEmpty(s.uiLines(intro))
	// 記錄 0x53 結尾的 `]` 就是數字前面那個空白，不用自己補（RE-076）；
	// 譯文沒有那個記號，就自己補一個。
	cost := strings.TrimLeft(s.uiRaw(CostPrefix), " ")
	if cost != "" && !strings.HasSuffix(cost, " ") {
		cost += " "
	}
	amount := strconv.Itoa(price)
	suffix := nonEmpty(s.uiLines(CostSuffix))
	if s.Trans.UIRecord(CostSuffix) != "" && len(suffix) == 1 {
		return append(lines, strings.TrimSpace(cost+amount+" "+suffix[0]))
	}
	lines = append(lines, strings.TrimRight(cost+amount, " "))
	return append(lines, suffix...)
}

// 甲冑調整（RE-075）。TRAIN 圖示由旗標 byte 1 的 bit1 與 bit3 給：
// bit1 是把身上的甲冑改到合身，bit3 是鍛造。
//
// **鍛造刻意不做**（RE-134）：入口判準是物品記錄 +0x0F 的 bit5，而三個母體逐筆
// 掃過之後，唯一叫 Ingot 的東西（城 5 的 Copper Ingot）那一位是零，有那一位的
// 全是 Cloth Pants 與 Leather Hood——**原版玩家走不到這一條**。說明書寫的是設計
// 意圖，資料沒跟上。照抄流程容易，但做出來是一條走不到的路，而入口條件錯了會讓
// 錯的東西變成武器。

// 甲冑調整用得到的記錄。
const (
	// ArmourFitsRecord 是 `"You don't need to fit!"`。
	ArmourFitsRecord = 0x67
	// ArmourPrefixRecord 與 ArmourSuffixRecord 夾住部位名：
	// `"The armor on your` ＋ head／torso／legs ＋ `needs to be adjusted!"`。
	ArmourPrefixRecord = 0x3A
	ArmourSuffixRecord = 0x3B
	// BodyPartRecord 是 head／torso／legs 三個小寫部位名。
	BodyPartRecord = 0xC7
)

// adjustEntry 是一件要改的甲冑：部位與改完該有的尺寸。
type adjustEntry struct {
	Position int
	Size     int
}

// StartTrain 是按下 TRAIN。
func (s *Session) StartTrain() {
	if s.Mode != ModeInside || s.Inside == nil || s.door == nil {
		return
	}
	if s.door.IconFlags[1]&0x02 != 0 {
		s.startAdjust()
		return
	}
	// 鍛造那一條原版走不到（見上面），只記一筆。
	s.note(iconName(IconTrain))
}

// startAdjust 掃過頭、軀幹、腿三個部位，把不合身的挑出來。
func (s *Session) startAdjust() {
	i := s.speakerIndex()
	if i < 0 {
		return
	}
	size := s.memberSize(i)
	var pending []adjustEntry
	total := 0
	for pos := 0; pos < FitPositions; pos++ {
		g := s.Party[i].Worn[pos]
		if g == nil {
			continue
		}
		want := FitSize(*g, size)
		if want == g.Encumbrance {
			continue
		}
		total += FitCost(*g)
		pending = append(pending, adjustEntry{Position: pos, Size: want})
	}
	if total == 0 {
		s.say(nonEmpty(s.uiLines(ArmourFitsRecord)), nil)
		return
	}
	s.shop = &shopState{stage: shopAdjust, member: i, adjust: pending, fit: total}
	s.adjustNotice(0)
}

// adjustNotice 逐件講「你某處的甲冑該調整了」，每一句都等空白鍵，
// 講完才問總價。
func (s *Session) adjustNotice(n int) {
	sh := s.shop
	if n >= len(sh.adjust) {
		s.Inside.Dialog = s.priceLines(FitOfferRecord, sh.fit)
		s.Inside.DialogZh = nil
		s.setInsideIcons([]int{IconYes, IconNo})
		return
	}
	// 記錄 0x3A 結尾的 `]` 是部位名前面那個空白。
	part := s.uiRecord(BodyPartRecord + sh.adjust[n].Position)
	lines := []string{strings.TrimRight(s.uiRaw(ArmourPrefixRecord)+part, " ")}
	lines = append(lines, nonEmpty(s.uiLines(ArmourSuffixRecord))...)
	next := n + 1
	s.shopSay(lines, func() { s.adjustNotice(next) })
}

// adjustAnswer 是總價那一問的回答。
func (s *Session) adjustAnswer(yes bool) {
	sh := s.shop
	if !yes {
		s.endShop()
		return
	}
	i := sh.member
	if s.Party[i].Gold < sh.fit {
		s.shopSay(nonEmpty(s.uiLines(NotEnoughRecord)), s.endShop)
		return
	}
	s.Party[i].Gold -= sh.fit
	for _, entry := range sh.adjust {
		if g := s.Party[i].Worn[entry.Position]; g != nil {
			g.setEncumbrance(entry.Size)
		}
	}
	s.Party[i].SyncCarried()
	s.syncParty()
	s.endShop()
}
