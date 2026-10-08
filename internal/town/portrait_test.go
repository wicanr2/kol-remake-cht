package town

import (
	"path/filepath"
	"sort"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/portrait"
)

// rec 造一格身上的原始十六 byte：種類、class 與外觀號放在原版的位置。
func rec(kind, class, style int) []byte {
	out := make([]byte, 16)
	out[0] = 1 // 名字那一格非零，代表這一格有東西
	out[12] = byte(kind)
	out[13] = byte(class)
	out[14] = byte(style)
	return out
}

func TestFigureFromWorn(t *testing.T) {
	m := Member{Female: true, Sheet: &Sheet{Race: 1}}
	m.Worn[0] = &Good{SizeClass: 3}
	m.Worn[1] = &Good{SizeClass: 4}
	m.Worn[2] = &Good{SizeClass: 5}
	m.Worn[PortraitShieldPos] = &Good{SizeClass: 2}
	m.Worn[WornSlots-1] = &Good{Class: 87, SizeClass: 1, Kind: 1} // 腰帶
	m.InHand = &Good{Class: 87, SizeClass: 1, Kind: 1, Hands: 1}  // 手上那件

	f := m.Figure()
	if f.Gender != 1 || f.Race != 1 || f.Build() != 5 {
		t.Fatalf("體型組合算錯：%+v", f)
	}
	if f.Styles != [3]int{3, 4, 5} {
		t.Fatalf("三層外觀號是 %v", f.Styles)
	}
	if f.Shield == nil || f.Shield.Style != 2 {
		t.Fatalf("盾沒讀到：%+v", f.Shield)
	}
	if f.Weapon == nil || !f.Weapon.TwoHand || f.Weapon.Class != 87 {
		t.Fatalf("手上那件沒讀到：%+v", f.Weapon)
	}
	if f.Belt == nil || f.Belt.TwoHand {
		t.Fatalf("腰帶那一格沒讀到，或讀成手上那件：%+v", f.Belt)
	}
}

// 讀檔回來的隊伍只有原始位元組，typed 的穿戴欄是空的——立繪仍然要組得出來。
func TestFigureFromCarried(t *testing.T) {
	offsets := CarriedOffsets()
	m := Member{Sheet: &Sheet{Race: 3}, Carried: make([][]byte, len(offsets))}
	for i, off := range offsets {
		switch off {
		case CarriedFirst:
			// 手上空著，與 Theodoric 一樣。
		case PortraitHeadSlot:
			m.Carried[i] = rec(2, 0, 0x08)
		case PortraitTorsoSlot:
			m.Carried[i] = rec(2, 0, 0x18)
		case PortraitLegsSlot:
			m.Carried[i] = rec(2, 0, 0x28)
		case PortraitBeltSlot:
			m.Carried[i] = rec(1, 87, 0x91)
		}
	}
	f := m.Figure()
	if f.Build() != 3 {
		t.Fatalf("體型組合是 %d", f.Build())
	}
	// 外觀號只取低 nibble：0x08／0x18／0x28 都是 8。
	if f.Styles != [3]int{8, 8, 8} {
		t.Fatalf("三層外觀號是 %v", f.Styles)
	}
	if f.Weapon != nil {
		t.Fatalf("手上應該空著：%+v", f.Weapon)
	}
	if f.Belt == nil || f.Belt.Class != 87 || f.Belt.Style != 1 {
		t.Fatalf("腰帶那一格沒讀到：%+v", f.Belt)
	}
}

// 空的那一格是負數，畫的時候整層跳過。
func TestFigureEmptySlotsAreNegative(t *testing.T) {
	f := Member{Sheet: &Sheet{}}.Figure()
	if f.Styles != [3]int{-1, -1, -1} {
		t.Fatalf("空的穿戴欄應該是 -1，得到 %v", f.Styles)
	}
	if f.Weapon != nil || f.Shield != nil || f.Belt != nil {
		t.Fatalf("空的那幾格應該是 nil：%+v", f)
	}
}

// RE-175：原版只看首byte；空欄未使用尾端可保留舊值。
func TestFigureIgnoresEmptyRecordTails(t *testing.T) {
	m := Member{Sheet: &Sheet{}, Carried: make([][]byte, len(CarriedOffsets()))}
	for i := range m.Carried {
		m.Carried[i] = rec(3, 87, 6)
		m.Carried[i][0] = 0
	}
	f := m.Figure()
	if f.Styles != [3]int{-1, -1, -1} || f.Weapon != nil || f.Shield != nil || f.Belt != nil {
		t.Fatalf("空欄尾端不得生成裝備圖：%+v", f)
	}
	if carriedItem(rec(3, 87, 6)[:15]) != nil || carriedItem(rec(3, 87, 6)) == nil {
		t.Fatal("短記錄須拒絕，完整非空記錄須保留")
	}
}

// 物品那一頁的清單就是身上那十七格，只是順序不同（RE-127）。
func TestItemSlotsAreTheCarriedSlots(t *testing.T) {
	l, err := portrait.Load(filepath.Join("..", "..", "assets",
		"original-derived-data", "portrait-layout.json"))
	if err != nil {
		t.Skip("沒有 portrait-layout.json，跳過")
	}
	want := CarriedOffsets()
	got := append([]int(nil), l.ItemSlots...)
	sort.Ints(got)
	if len(got) != len(want) {
		t.Fatalf("清單有 %d 項，身上有 %d 格", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("第 %d 格是 0x%x，身上那一組是 0x%x", i, got[i], want[i])
		}
	}
}

// 隊伍立繪那一頁：層次大的先畫，人多就有多份貼圖，按一個鍵回資料頁。
func TestOpenPartyPage(t *testing.T) {
	s := itemSession(t)
	l, err := portrait.Load(filepath.Join("..", "..", "assets",
		"original-derived-data", "portrait-layout.json"))
	if err != nil {
		t.Skip("沒有 portrait-layout.json，跳過")
	}
	s.Portraits = l
	s.Party = append(s.Party, Member{Name: "Poontang", Female: true,
		Sheet: &Sheet{Race: 1}})
	s.Inside.Party = s.Party

	s.Inside.Selected = 2 // 資料頁圖示列的第三格是換人
	s.Accept()
	if !s.ShowingPartyPage() {
		t.Fatalf("沒開起來；未實作紀錄：%v", s.Unimplemented)
	}
	// 兩個人至少各有體型與一隻手。
	if len(s.Inside.PartyPage) < 4 {
		t.Fatalf("只貼了 %d 張", len(s.Inside.PartyPage))
	}
	// 第一張是層次大的那一位：Theodoric 是男性種族 3（層次 6），
	// Poontang 是女性種族 1（體型組合 5，層次 4），所以先畫前者。
	if s.Inside.PartyPage[0].Archive != portrait.ArchiveBody ||
		s.Inside.PartyPage[0].Index != 3 {
		t.Fatalf("第一張是 %+v", s.Inside.PartyPage[0])
	}
	s.Inside.Selected = 0
	s.Accept()
	if s.ShowingPartyPage() {
		t.Fatal("按一個鍵應該回資料頁")
	}
	if !s.ShowingSheet() {
		t.Fatal("應該回到角色資料頁")
	}
	// 空白鍵走另一條正式按鍵分派，不能只把圖示列還原而留下立繪頁。
	s.Inside.Selected = 2
	s.Key("RET")
	if !s.ShowingPartyPage() {
		t.Fatal("重新開啟隊伍立繪頁失敗")
	}
	s.Key("SPACE")
	if s.ShowingPartyPage() || !s.ShowingSheet() {
		t.Fatal("空白鍵沒有從隊伍立繪頁返回角色資料頁")
	}
}
