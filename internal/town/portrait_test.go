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
	m.Worn[WornSlots-1] = &Good{Class: 87, SizeClass: 1, Kind: 1, Hands: 1}

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
