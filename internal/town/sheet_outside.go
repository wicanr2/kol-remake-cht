package town

// 城鎮戶外與野外按 1…6 打開角色資料頁（RE-150）。
//
// 原版兩處（`sub_10000:loc_10DA4`、`sub_29004:loc_29652`）相同：鍵值 0x31…0x36 減 0x31
// 得到 n，掃六個角色槽找第 n 名在籍的、把 `word_39A52` 指過去，保存圖示列
// （`word_2DF3C`）與兩個 byte 旗標，呼叫 `sub_15ABF`，回來再還原那幾樣。
// `word_39A52` **不還原**——它就是發言者，所以關掉之後發言者停在第 n 名。
//
// 資料頁本身（`sub_16239` 開頭的 `sub_1103F(0)` 與 `sub_13EEE`）整頁重畫，看起來與
// 建築內用 MIRROR 打開的一樣；只有物品與隊伍立繪那兩頁的背板在野外換一張。

// sheetHost 記下打開資料頁之前的畫面，關掉時回去。
type sheetHost struct {
	mode   Mode
	inside *Inn
	cursor int
}

// sheetDigit 把 `1`…`6` 換成隊伍裡的位置。
func sheetDigit(token string) (int, bool) {
	if len(token) != 1 || token[0] < '1' || token[0] > '6' {
		return 0, false
	}
	return int(token[0] - '1'), true
}

// SheetDigit 是給野外那一層用的同一份換算。
func SheetDigit(token string) (int, bool) { return sheetDigit(token) }

// OpenSheetOutside 打開第 n 名的資料頁。隊伍沒有那麼多人就不做事（原版掃完六格找不到
// 第 n 名，跳回迴圈，`0x2964F`）。wild 為真時是在野外打開的。
func (s *Session) OpenSheetOutside(n int, wild bool) bool {
	if s.sheetHost != nil || n < 0 || n >= len(s.Party) {
		return false
	}
	page := s.buildSheet(s.Party[n])
	if page == nil {
		return false
	}
	s.setSpeaker(n)
	s.sheetHost = &sheetHost{mode: s.Mode, inside: s.Inside, cursor: s.Cursor}
	s.Inside = &Inn{Party: s.Party, Wild: wild, Sheet: page, SheetLabels: s.sheetLabels()}
	s.Mode = ModeInside
	s.restoreSpeaker()
	s.setInsideIcons(sheetIcons(s.Party[n]))
	return true
}

// ShowingSheetOutside 回報現在是不是停在戶外或野外打開的那一份資料頁（含子頁）。
func (s *Session) ShowingSheetOutside() bool { return s.sheetHost != nil }

// closeSheetOutside 是資料頁上選 U-TURN：回到打開之前的畫面。
func (s *Session) closeSheetOutside() {
	h := s.sheetHost
	s.sheetHost = nil
	s.Inside = h.inside
	s.Mode = h.mode
	s.Cursor = h.cursor
	s.syncParty()
}
