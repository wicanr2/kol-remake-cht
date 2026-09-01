package roster

// DeleteFlow 是 DELETE CHARACTER 的狀態機。
//
// 行為出自 docs/verification/character-ui/delete-character-run1 與
// delete-character-confirm-run1 的正常玩家路徑收據：
//
//   - 清單只列已建立的槽位，並且連續排列；刪除之後名冊會往上補齊。
//   - 反白條起點是第 0 列，上下移動不繞圈。
//   - Return 叫出確認視窗；視窗開著時清單沒有任何一列反白。
//   - 視窗不吃方向鍵與 Return（原版收據裡送出這兩個鍵，畫面逐像素沒有變化），
//     只認 Y 與 N。
//   - 兩種回答都回到刪除清單，反白條重置為第 0 列。
//
// 確認視窗按 ESC 的行為沒有量測過，因此這裡不實作。
type DeleteFlow struct {
	roster  *Roster
	Cursor  int
	Confirm bool
	Deleted int // 已完成的刪除次數
}

func NewDeleteFlow(r *Roster) *DeleteFlow { return &DeleteFlow{roster: r} }

// List 是畫面上要列出的槽位。
func (f *DeleteFlow) List() []Slot { return f.roster.Occupied() }

// MoveCursor 上下移動反白條；不繞圈。確認視窗開著時方向鍵沒有作用。
func (f *DeleteFlow) MoveCursor(delta int) {
	if f.Confirm {
		return
	}
	n := len(f.List())
	if n == 0 {
		return
	}
	f.Cursor += delta
	if f.Cursor < 0 {
		f.Cursor = 0
	}
	if f.Cursor >= n {
		f.Cursor = n - 1
	}
}

// Accept 是 Return：在清單上叫出確認視窗，在視窗內沒有作用。
func (f *DeleteFlow) Accept() {
	if f.Confirm || len(f.List()) == 0 {
		return
	}
	f.Confirm = true
}

// Answer 回答確認視窗。yes 為真才刪除；兩種回答都關閉視窗並把反白條重置為第 0 列。
func (f *DeleteFlow) Answer(yes bool) {
	if !f.Confirm {
		return
	}
	if yes && f.roster.DeleteOccupied(f.Cursor) {
		f.Deleted++
	}
	f.Confirm = false
	f.Cursor = 0
}

// DeleteOccupied 刪除第 i 個已建立的槽位。原版刪除只寫一個 byte：把該槽位名字的第一個
// byte 寫成 0（delete-character-confirm-run1 的 CHARDATA 差分只有 slot 1 的 offset
// 0x000 由 0x50 變成 0x00），所在地、屬性與 CHARPICS／SHLDPICS 都不動；名冊的「往上
// 補齊」是畫面只列已建立槽位造成的，不是資料搬移。
func (r *Roster) DeleteOccupied(i int) bool {
	seen := 0
	for idx := range r.Slots {
		if !r.Slots[idx].Occupied {
			continue
		}
		if seen == i {
			r.Slots[idx].Name = ""
			r.Slots[idx].Occupied = false
			return true
		}
		seen++
	}
	return false
}
