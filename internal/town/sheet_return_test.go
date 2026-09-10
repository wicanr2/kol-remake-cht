package town

import "testing"

// 由正式按鍵分派走內頁與返回，避免把不可達的孤立欄位狀態當成產品缺陷。
func TestSpellPageReturnsBeforeSheetCloses(t *testing.T) {
	s := recruitSession(t, 1)
	s.Party[0].Sheet = theodoric()
	s.Party[0].Spells = []Spell{{Name: "TEST SPELL"}}
	s.syncParty()
	choose := func(id int) {
		t.Helper()
		for n := 0; n < 12; n++ {
			icons, selected := s.Icons()
			if len(icons) > 0 && icons[selected] == IconSprite(id) {
				s.Key("RET")
				return
			}
			s.Key(">")
		}
		t.Fatalf("圖示 %d 未出現", id)
	}
	for n := 0; n < 2; n++ {
		choose(IconMirror)
		choose(IconSheetSpells)
		if !s.ShowingSpellList() {
			t.Fatal("未開啟法術清單")
		}
		choose(IconUTurn)
		if s.ShowingSpellList() || !s.ShowingSheet() {
			t.Fatal("法術頁返回未清理內頁")
		}
		choose(IconUTurn)
		if s.ShowingSheet() || s.Inside.SpellList != nil {
			t.Fatal("關頁後仍殘留法術畫面")
		}
	}
	// 模擬資料已更新後重開，不能重用上一輪畫面的快照。
	s.Party[0].Spells = nil
	choose(IconMirror)
	if s.ShowingSpellList() {
		t.Fatal("重開資料頁復活舊法術清單")
	}
	icons, _ := s.Icons()
	for _, id := range icons {
		if id == IconSprite(IconSheetSpells) {
			t.Fatal("舊法術圖示殘留")
		}
	}
}
