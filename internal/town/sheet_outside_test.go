package town

import "testing"

func outsideSheetSession(t *testing.T) *Session {
	t.Helper()
	s := playSession(t)
	s.UIText = itemText()
	s.Party = []Member{
		{Name: "Theodoric", Portrait: 47, Speaker: true, Sheet: theodoric()},
		{Name: "Olear", Portrait: 12, Sheet: theodoric()},
		{Name: "Moraine", Portrait: 30, Sheet: theodoric()},
	}
	s.syncParty()
	return s
}

// 城鎮戶外按 `2` 開第二名的資料頁，發言者換過去；U-TURN 回到戶外，發言者不還原
// （`sub_10000:loc_10DA4` 不還原 `word_39A52`，RE-150）。
func TestSheetDigitOutdoor(t *testing.T) {
	s := outsideSheetSession(t)
	if s.Mode != ModeMap {
		t.Fatalf("夾具應該在戶外，得到模式 %d", s.Mode)
	}
	s.Cursor = 0
	if !s.Key("2") || !s.ShowingSheet() || !s.ShowingSheetOutside() || s.Inside.Wild {
		t.Fatal("戶外按 2 應該打開資料頁，而且不是野外那一份")
	}
	if s.speakerIndex() != 1 || !s.Inside.Party[1].Speaker {
		t.Fatal("發言者應該換成第二名")
	}
	s.Inside.Selected = 0 // U-TURN
	s.Key("RET")
	if s.ShowingSheetOutside() || s.Mode != ModeMap || s.Inside != nil {
		t.Fatal("U-TURN 應該回到戶外")
	}
	if s.speakerIndex() != 1 {
		t.Fatal("原版不還原發言者")
	}
}

// 隊伍沒有第 n 名就不做事（`0x2964F` 跳回迴圈）。
func TestSheetDigitBeyondParty(t *testing.T) {
	s := outsideSheetSession(t)
	if s.Key("5") || s.ShowingSheet() || s.Mode != ModeMap || s.speakerIndex() != 0 {
		t.Fatal("只有三人時按 5 不該有反應")
	}
}

// 從戶外那一份也進得去物品頁，返回兩層回到戶外；野外那一份標 Wild，背板換第 4 張。
func TestSheetOutsideSubPages(t *testing.T) {
	s := outsideSheetSession(t)
	if !s.OpenSheetOutside(0, true) || !s.Inside.Wild {
		t.Fatal("野外打開的資料頁應該標 Wild")
	}
	openItems(s)
	if !s.ShowingItemPage() {
		t.Fatal("應該進到物品頁")
	}
	s.Inside.Selected = 1 // 清單層的 U-TURN
	s.Accept()
	if s.ShowingItemPage() || !s.ShowingSheet() {
		t.Fatal("物品頁的 U-TURN 應該回資料頁")
	}
	s.Inside.Selected = 0
	s.Accept()
	if s.ShowingSheetOutside() || s.Mode != ModeMap {
		t.Fatal("資料頁的 U-TURN 應該回到原畫面")
	}
	if got := SheetPageBackPieces(true); len(got) != 1 || got[0].Index != 4 || got[0].Y != 3 {
		t.Fatalf("野外背板應該是 KOLPICS 第 4 張貼在掃描列 3，得到 %v", got)
	}
	if got := SheetPageBackPieces(false); len(got) != 2 || got[0].Index != 5 || got[1].Index != 13 {
		t.Fatalf("城裡背板應該是第 5、13 張，得到 %v", got)
	}
}
