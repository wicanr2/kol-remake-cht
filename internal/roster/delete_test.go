package roster

import "testing"

func sample() *Roster {
	r := &Roster{Schema: "kol-character-roster-v1"}
	r.Layout.SlotCount = 4
	r.Slots = []Slot{
		{Slot: 0, Name: "Theodoric", Place: "HALFWAY HO", Town: "MIN", Occupied: true},
		{Slot: 1, Name: "Poontang", Place: "HALFWAY HO", Town: "MIN", Occupied: true},
		{Slot: 2, Name: "Olear", Place: "HALFWAY HO", Town: "MIN", Occupied: true},
		{Slot: 3, Name: "", Place: "TROLLSBANE", Town: "BRT", Occupied: false},
	}
	return r
}

// 刪除第二列之後名冊往上補齊，且只有那一個槽位受影響。
func TestDeleteCompactsList(t *testing.T) {
	f := NewDeleteFlow(sample())
	f.MoveCursor(1)
	f.Accept()
	if !f.Confirm {
		t.Fatal("Return 應該叫出確認視窗")
	}
	f.MoveCursor(1) // 視窗開著時方向鍵沒有作用
	if f.Cursor != 1 {
		t.Fatalf("視窗開著時反白條不應移動，得到 %d", f.Cursor)
	}
	f.Answer(true)
	if f.Confirm || f.Cursor != 0 || f.Deleted != 1 {
		t.Fatalf("回答之後狀態不符：confirm=%v cursor=%d deleted=%d",
			f.Confirm, f.Cursor, f.Deleted)
	}
	got := []string{}
	for _, s := range f.List() {
		got = append(got, s.Name)
	}
	if len(got) != 2 || got[0] != "Theodoric" || got[1] != "Olear" {
		t.Fatalf("刪除後的名冊不符：%v", got)
	}
}

// 回答 NO 不刪除任何角色，但一樣關閉視窗並把反白條重置。
func TestAnswerNoKeepsRoster(t *testing.T) {
	f := NewDeleteFlow(sample())
	f.MoveCursor(2)
	f.Accept()
	f.Answer(false)
	if f.Confirm || f.Cursor != 0 || f.Deleted != 0 {
		t.Fatalf("回答 NO 之後狀態不符：confirm=%v cursor=%d deleted=%d",
			f.Confirm, f.Cursor, f.Deleted)
	}
	if len(f.List()) != 3 {
		t.Fatalf("回答 NO 之後名冊長度應為 3，得到 %d", len(f.List()))
	}
}

// 反白條不繞圈。
func TestCursorClamps(t *testing.T) {
	f := NewDeleteFlow(sample())
	f.MoveCursor(-1)
	if f.Cursor != 0 {
		t.Fatalf("第 0 列往上應停在 0，得到 %d", f.Cursor)
	}
	f.MoveCursor(9)
	if f.Cursor != 2 {
		t.Fatalf("末列往下應停在 2，得到 %d", f.Cursor)
	}
}
