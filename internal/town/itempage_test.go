package town

import (
	"encoding/hex"
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/portrait"
)

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
		uiRecord{Index: TooSmallRecord, Text: "Too small to wear!"},
		uiRecord{Index: GiveToWhomRecord, Text: "Give to whom?"},
		uiRecord{Index: TheirPackFullRecord, Text: "Their backpack is full!"},
		uiRecord{Index: TradedRecord, Text: "The Object is traded!"},
		uiRecord{Index: DroppedRecord, Text: "It is dropped!"},
		uiRecord{Index: DropSureRecord, Text: "Are you sure?"},
		uiRecord{Index: YourPackFullRecord, Text: "Your backpack is full!"},
	)
	return t
}

// act 在動作層按下某一個圖示。
func act(s *Session, icon int) {
	for i, id := range s.Inside.Icons {
		if id == IconSprite(icon) {
			s.Inside.Selected = i
			s.Accept()
			return
		}
	}
	panic("圖示列上沒有那一格")
}

// use 在清單層按下那一格，進動作層。
func use(s *Session) {
	s.Inside.Selected = 0
	s.Accept()
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
	if len(s.Inside.Icons) != 2 || s.Inside.Icons[0] != IconSprite(IconItemUse) ||
		s.Inside.Icons[1] != IconSprite(IconUTurn) {
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
	use(s)

	worn := s.Inside.ItemPage.Message
	if len(worn) != 2 || worn[0] != "    YOU HAVE NOTHING" || worn[1] != "    ON YOUR HEAD!" {
		t.Fatalf("身上那一格的訊息是 %v", worn)
	}
	s.Key("RET")
	s.Inside.ItemPage.Selected = ItemSlotCount - 1
	use(s)
	pocket := s.Inside.ItemPage.Message
	if len(pocket) != 2 || pocket[1] != "    IN YOUR POCKET!" {
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
		if len(s.Inside.ItemPage.Message) > 0 {
			s.Key("RET")
		}
		s.Inside.ItemPage.Selected = c.row
		use(s)
		got := s.Inside.ItemPage.Message
		if len(got) != 2 || got[1] != "    ON YOUR "+strings.ToUpper(c.want)+"!" {
			t.Fatalf("第 %d 列印的是 %v，期望 %s", c.row, got, c.want)
		}
	}
}

// 空格訊息不是物品動作層；確認只回到同一列，不再次開訊息或離開資料頁。
func TestEmptySlotConfirmationReturnsToSameRow(t *testing.T) {
	for _, confirm := range []string{"RET", "SPACE"} {
		t.Run(confirm, func(t *testing.T) {
			s := itemSession(t)
			s.Party[s.speakerIndex()].Worn = [WornSlots]*Good{}
			openItems(s)
			s.Inside.ItemPage.Selected = 3
			use(s)
			icons, _ := s.Icons()
			if len(icons) != 1 || icons[0] != IconSprite(IconSpacebar) {
				t.Fatalf("空格確認圖示：%v", icons)
			}
			for _, ignored := range []string{"D", "U", ">", "A"} {
				s.Key(ignored)
				if s.Inside.ItemPage.Selected != 3 || len(s.Inside.ItemPage.Message) == 0 {
					t.Fatalf("%s不應移列或關閉空格訊息", ignored)
				}
			}
			s.Key(confirm)
			if !s.ShowingItemPage() || s.Inside.ItemPage.Selected != 3 || len(s.Inside.ItemPage.Message) != 0 {
				t.Fatal("確認應返回同一物品列")
			}
			wantIcons(t, s, IconItemUse, IconUTurn)
		})
	}
}

// 有東西的那一格顯示明細，與商店同一支（RE-075）。
func TestFilledSlotShowsDetail(t *testing.T) {
	s := itemSession(t)
	i := s.speakerIndex()
	s.Party[i].Worn[0] = &Good{Name: "Barbut Helm", Price: 250, Kind: 2,
		Position: 0, Encumbrance: 23, SizeClass: 8, Condition: 255}
	openItems(s)
	use(s)
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
	s.Inside.Selected = 1 // U-TURN 在那一格右邊
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

// 背包裡的武器可以拿到手上（`sub_18857` 的 0x16 → 0x30）。
func TestTakeItemIntoHand(t *testing.T) {
	s := itemSession(t)
	i := s.speakerIndex()
	s.Party[i].Backpack = []Good{{Name: "Long Spear", Kind: 1, Class: 158, Position: 9}}
	openItems(s)
	s.Inside.ItemPage.Selected = WornSlots + 1 // 第一個口袋
	use(s)
	act(s, IconItemTake)
	if s.Party[i].InHand == nil || s.Party[i].InHand.Name != "Long Spear" {
		t.Fatalf("手上是 %+v", s.Party[i].InHand)
	}
	if len(s.Party[i].Backpack) != 0 {
		t.Fatalf("背包應該空了：%+v", s.Party[i].Backpack)
	}
	if s.Inside.ItemPage.Selected != 0 {
		t.Fatal("拿取後完整重建應選回首欄")
	}
}

// 手上那件可以掛到腰帶（0x1C → 0x190）；條件是種類 3（Both）。
func TestBeltItemFromHand(t *testing.T) {
	s := itemSession(t)
	i := s.speakerIndex()
	s.Party[i].InHand = &Good{Name: "Greatsword", Kind: 3, Class: 87, Position: 9}
	openItems(s)
	s.Inside.ItemPage.Selected = WornSlots
	use(s)
	act(s, IconItemBelt)
	if s.Party[i].InHand != nil {
		t.Fatalf("手上應該空了：%+v", s.Party[i].InHand)
	}
	if s.Party[i].Worn[WornSlots-1] == nil ||
		s.Party[i].Worn[WornSlots-1].Name != "Greatsword" {
		t.Fatalf("腰帶是 %+v", s.Party[i].Worn[WornSlots-1])
	}
}

// 穿上：合身才搬進那件東西自己的部位，不合身印 Too small to wear!。
func TestWearItemChecksFit(t *testing.T) {
	s := itemSession(t)
	i := s.speakerIndex()
	helm := Good{Name: "Barbut Helm", Kind: 2, Position: 0, SizeClass: 8,
		Encumbrance: 23, Condition: 255}
	s.Party[i].Backpack = []Good{helm}
	openItems(s)
	s.Inside.ItemPage.Selected = WornSlots + 1
	use(s)
	act(s, IconItemWear)
	if s.Party[i].Worn[0] == nil {
		t.Fatalf("應該穿上了：%+v", s.Party[i].Worn[0])
	}

	// 尺寸不足的那一件穿不上。
	small := helm
	small.Encumbrance = 1
	s.Party[i].Worn[0] = nil
	s.Party[i].Backpack = []Good{small}
	s.Inside.ItemPage.Selected = WornSlots + 1
	use(s)
	act(s, IconItemWear)
	if s.Party[i].Worn[0] != nil {
		t.Fatalf("不合身不該穿上：%+v", s.Party[i].Worn[0])
	}
	if len(s.Inside.ItemPage.Note) == 0 ||
		s.Inside.ItemPage.Note[0] != "Too small to wear!" {
		t.Fatalf("應該印 Too small to wear!：%v", s.Inside.ItemPage.Note)
	}
}

// 放回背包（0x14）：滿了印 Your backpack is full!。
func TestStowItem(t *testing.T) {
	s := itemSession(t)
	i := s.speakerIndex()
	s.Party[i].Worn[0] = &Good{Name: "Barbut Helm", Kind: 2, Position: 0}
	openItems(s)
	s.Inside.ItemPage.Selected = 0
	use(s)
	act(s, IconItemStow)
	if s.Party[i].Worn[0] != nil || len(s.Party[i].Backpack) != 1 {
		t.Fatalf("應該搬進背包：%+v／%+v", s.Party[i].Worn[0], s.Party[i].Backpack)
	}

	s.Party[i].Worn[1] = &Good{Name: "Platemail T", Kind: 2, Position: 1}
	s.Party[i].Backpack = make([]Good, BackpackSlots)
	for j := range s.Party[i].Backpack {
		s.Party[i].Backpack[j] = Good{Name: "OCCUPIED"}
	}
	s.Inside.ItemPage.Selected = 1
	use(s)
	act(s, IconItemStow)
	if s.Party[i].Worn[1] == nil {
		t.Fatal("背包滿了不該搬走")
	}
	if len(s.Inside.ItemPage.Note) == 0 ||
		s.Inside.ItemPage.Note[0] != "Your backpack is full!" {
		t.Fatalf("應該印背包滿了：%v", s.Inside.ItemPage.Note)
	}
}

// 丟掉（0x15）：先確認YES才清空，訊息確認後完整返回首欄。
func TestDropItem(t *testing.T) {
	s := itemSession(t)
	i := s.speakerIndex()
	s.Party[i].Worn[0] = &Good{Name: "Barbut Helm", Kind: 2, Position: 0}
	openItems(s)
	s.Inside.ItemPage.Selected = 0
	use(s)
	act(s, IconItemDrop)
	if s.Party[i].Worn[0] == nil || !s.Inside.ItemPage.DropConfirm {
		t.Fatal("選丟棄時應先保留物品並等待YES／NO")
	}
	act(s, IconYes)
	if s.Party[i].Worn[0] != nil {
		t.Fatalf("應該丟掉了：%+v", s.Party[i].Worn[0])
	}
	if len(s.Inside.ItemPage.DropMessage) == 0 ||
		s.Inside.ItemPage.DropMessage[0] != "IT IS DROPPED!" || s.Inside.ItemPage.Detail == nil {
		t.Fatalf("應保留明細並印IT IS DROPPED!：%+v", s.Inside.ItemPage)
	}
	s.Key("SPACE")
	if len(s.Inside.ItemPage.DropMessage) != 0 || s.Inside.ItemPage.Selected != 0 {
		t.Fatal("完成訊息確認後應返回首欄")
	}
}

func TestDropItemCancelPreservesSource(t *testing.T) {
	s := itemSession(t)
	i := s.speakerIndex()
	g := &Good{Name: "Greatsword", Kind: 3, Class: 87, Position: 9}
	s.Party[i].Worn[9] = g
	openItems(s)
	s.Inside.ItemPage.Selected = 9
	use(s)
	act(s, IconItemDrop)
	for _, key := range []string{"U", "D", "SPACE"} {
		s.Key(key)
		if !s.Inside.ItemPage.DropConfirm || s.Inside.ItemPage.Selected != 9 || s.Party[i].Worn[9] != g {
			t.Fatalf("提問期間%s改變了來源或確認層", key)
		}
	}
	s.Key(">")
	s.Key("RET")
	if s.Party[i].Worn[9] != g || s.Inside.ItemPage.Selected != 0 || s.Inside.ItemPage.DropConfirm {
		t.Fatal("NO應保留物品，完整返回首欄")
	}
}

// 給隊友（0x0E）：對方背包滿了就不搬。
func TestGiveItem(t *testing.T) {
	for _, confirmation := range []string{"RET", "SPACE", "ICON"} {
		t.Run(confirmation, func(t *testing.T) { testGiveItemConfirmation(t, confirmation) })
	}
}

func testGiveItemConfirmation(t *testing.T, confirmation string) {
	s := itemSession(t)
	s.Party = append(s.Party, Member{Name: "Poontang"})
	s.Inside.Party = s.Party
	i := s.speakerIndex()
	record := "b10642ad0082b1a050fa17ff2a000840"
	s.Party[i].Worn[0] = &Good{Name: "Barbut Helm", Kind: 2, Position: 0, Record: record}
	s.Party[i].SyncCarried()
	openItems(s)
	s.Inside.ItemPage.Selected = 0
	use(s)
	act(s, IconItemGive)
	if !s.Inside.ItemPage.Giving {
		t.Fatal("應該在問給誰")
	}
	before, err := json.Marshal(s.Party)
	if err != nil {
		t.Fatal(err)
	}
	for _, ignored := range []string{"1", "6"} {
		s.Key(ignored)
		if !s.Inside.ItemPage.Giving || s.Inside.ItemPage.GiftConfirm {
			t.Fatalf("%s不應選收件人", ignored)
		}
	}
	s.Key("2")
	if !s.Inside.ItemPage.GiftConfirm || s.Inside.ItemPage.Giving || len(s.Inside.ItemPage.GiftMessage) == 0 {
		t.Fatal("選收件人應保留明細及完成訊息等待確認")
	}
	wantIcons(t, s, IconSpacebar)
	for _, ignored := range []string{"D", "U", "1", "2", "ESC"} {
		s.Key(ignored)
		after, err := json.Marshal(s.Party)
		if err != nil || string(after) != string(before) || !s.Inside.ItemPage.GiftConfirm {
			t.Fatalf("%s不應搬移物品或離開等待", ignored)
		}
	}
	if confirmation == "ICON" {
		act(s, IconSpacebar)
	} else {
		s.Key(confirmation)
	}
	if s.Party[i].Worn[0] != nil {
		t.Fatalf("應該給出去了：%+v", s.Party[i].Worn[0])
	}
	if len(s.Party[1].Backpack) != 1 {
		t.Fatalf("對方的背包是 %+v", s.Party[1].Backpack)
	}
	if hex.EncodeToString(s.Party[i].Carried[1+BackpackSlots]) != strings.Repeat("00", 16) ||
		len(s.Party[1].Carried) != ItemSlotCount || hex.EncodeToString(s.Party[1].Carried[1]) != record {
		t.Fatal("送出者與接收者的原始物品鏡像沒有同步")
	}
	if p := s.Inside.ItemPage; p.Selected != 0 || p.GiftConfirm || len(p.GiftMessage) > 0 || len(p.Note) > 0 {
		t.Fatalf("確認後應清訊息並回首欄：%+v", p)
	}
}

// RE-175正常原版的四次巨劍交換與NO／YES；鏡像須在重畫前同步。
func TestItemActionsSynchronizeCarriedBeforeFigure(t *testing.T) {
	s := itemSession(t)
	l, err := portrait.Load(filepath.Join("..", "..", "assets", "original-derived-data", "portrait-layout.json"))
	if err != nil {
		t.Fatal(err)
	}
	s.Portraits = l
	i := s.speakerIndex()
	record := "a3c8a1a4eef9100301904bff4f579600"
	g := &Good{Name: "Greatsword", Kind: 3, Class: 87, Hands: 1, SizeClass: 6, Position: 9, Record: record}
	s.Party[i].Worn[9] = g
	s.Party[i].SyncCarried()
	openItems(s)
	check := func(nonemptyOffset int) {
		t.Helper()
		for slot, offset := range CarriedOffsets() {
			want := strings.Repeat("00", 16)
			if offset == nonemptyOffset {
				want = record
			}
			if got := hex.EncodeToString(s.Party[i].Carried[slot]); got != want {
				t.Fatalf("位移0x%x物品鏡像=%s，預期%s", offset, got, want)
			}
		}
		if !reflect.DeepEqual(s.Inside.ItemPage.Figure, l.Compose(s.Party[i].Figure(), 0)) {
			t.Fatal("立繪在鏡像同步之前生成，仍含舊裝備")
		}
	}
	for _, step := range []struct{ row, icon, offset int }{
		{9, IconItemTake, 0x30}, {10, IconItemBelt, 0x190},
		{9, IconItemStow, 0xA0}, {11, IconItemTake, 0x30},
	} {
		s.Inside.ItemPage.Selected = step.row
		use(s)
		act(s, step.icon)
		check(step.offset)
	}
	s.Inside.ItemPage.Selected = 10
	use(s)
	act(s, IconItemDrop)
	act(s, IconNo)
	check(0x30)
	s.Inside.ItemPage.Selected = 10
	use(s)
	act(s, IconItemDrop)
	act(s, IconYes)
	if hex.EncodeToString(s.Party[i].Carried[0]) != strings.Repeat("00", 16) {
		t.Fatal("YES清除後、等待訊息時物品鏡像尚未清空")
	}
	s.Key("SPACE")
	check(-1)
	f := s.Party[i].Figure()
	if f.Belt != nil || f.Weapon != nil {
		t.Fatalf("丟棄後立繪仍有巨劍：%+v", f)
	}
}

// 吃掉（部位 10）：飢餓度減掉那一件的 Condition，夾在 0 以上。
func TestEatItem(t *testing.T) {
	s := itemSession(t)
	i := s.speakerIndex()
	s.Party[i].Sheet.Nutrition = 40
	s.Party[i].Backpack = []Good{{Name: "Rations", Kind: 0, Position: EatPosition,
		Condition: 25}}
	openItems(s)
	s.Inside.ItemPage.Selected = WornSlots + 1
	use(s)
	act(s, IconItemEat)
	if s.Party[i].Sheet.Nutrition != 15 {
		t.Fatalf("飢餓度是 %d，期望 15", s.Party[i].Sheet.Nutrition)
	}
	if len(s.Party[i].Backpack) != 0 {
		t.Fatalf("吃掉的那一件應該不見了：%+v", s.Party[i].Backpack)
	}
	if s.Inside.ItemPage.Selected != WornSlots+1 {
		t.Fatal("食用另支應保留既有選取，不套搬移的完整重建")
	}
}

// 圖示列依那一格與那件東西組出來（`sub_18857` 開頭那一串條件）。
func TestActionIconsFollowSlotAndKind(t *testing.T) {
	wield := Good{Kind: 1, Position: 9}
	wear := Good{Kind: 2, Position: 0}
	both := Good{Kind: 3, Position: 9}
	food := Good{Kind: 0, Position: EatPosition}
	for _, c := range []struct {
		what  string
		g     Good
		row   int
		party int
		want  []int
	}{
		{"背包裡的武器", wield, WornSlots + 1, 1,
			[]int{IconUTurn, IconItemTake, IconItemDrop}},
		{"背包裡的護甲", wear, WornSlots + 1, 1,
			[]int{IconUTurn, IconItemWear, IconItemDrop}},
		{"手上那件（Both）", both, WornSlots, 1,
			[]int{IconUTurn, IconItemBelt, IconItemDrop, IconItemStow}},
		{"隊伍不只一人", wear, 0, 2,
			[]int{IconUTurn, IconItemGive, IconItemDrop, IconItemStow}},
		{"食物", food, WornSlots + 1, 1,
			[]int{IconUTurn, IconItemDrop, IconItemEat}},
	} {
		got := itemIcons(c.g, c.row, c.party)
		if len(got) != len(c.want) {
			t.Fatalf("%s 的圖示列是 %v，期望 %v", c.what, got, c.want)
		}
		for i := range c.want {
			if got[i] != c.want[i] {
				t.Fatalf("%s 的圖示列是 %v，期望 %v", c.what, got, c.want)
			}
		}
	}
}

// RE-178正常來源：換列露出y7、空手確認返回露出完整快取，
// 搬移及重新開頁必須丟棄先前的還原歷史。
func TestItemEdgeNormalReturnAndRebuild(t *testing.T) {
	s := itemSession(t)
	l, err := portrait.Load(filepath.Join("..", "..", "assets", "original-derived-data", "portrait-layout.json"))
	if err != nil {
		t.Fatal(err)
	}
	s.Portraits = l
	openItems(s)
	p := s.Inside.ItemPage
	s.Key("D")
	if p.Selected != 1 || !p.RestoredEdge[7] || p.RestoredEdge[0] {
		t.Fatal("首次換列應只保留舊頭框的快取曝光")
	}
	for p.Selected < WornSlots {
		s.Key("D")
	}
	s.Party[s.speakerIndex()].InHand = nil
	use(s)
	s.Key("RET")
	if p.Selected != WornSlots || !p.RestoredEdge[0] || !p.RestoredEdge[157] || !p.RestoredEdge[163] || p.RestoredEdge[7] {
		t.Fatal("空手確認返回應複製已修改快取，再用完整框覆蓋y7")
	}
	s.Inside.Selected = 1
	s.Accept()
	openItems(s)
	p = s.Inside.ItemPage
	if p.RestoredEdge != [200]bool{} {
		t.Fatal("重新開頁不應保留前頁快取曝光")
	}
	i := s.speakerIndex()
	s.Party[i].Worn[9] = &Good{Name: "Greatsword", Kind: 3, Class: 87, Hands: 1, SizeClass: 6,
		Position: 9, Record: "a3c8a1a4eef9100301904bff4f579600"}
	s.Party[i].SyncCarried()
	for p.Selected < 9 {
		s.Key("D")
	}
	use(s)
	act(s, IconItemTake)
	if p.Selected != 0 || p.RestoredEdge != [200]bool{} || s.Party[i].InHand == nil {
		t.Fatal("成功搬移應重建首欄，清除舊快取曝光")
	}
}
