package battle

import (
	"bytes"
	"encoding/hex"
)

// EqualQuestItemRecords 對應 SPEC-014 sub_2886C 的完整 16-byte 比較。
// 缺少或損壞記錄不能由顯示名稱補猜；正式 24 個任務均提供完整記錄。
func EqualQuestItemRecords(a, b string) bool {
	x, err := hex.DecodeString(a)
	if err != nil || len(x) != 16 {
		return false
	}
	y, err := hex.DecodeString(b)
	return err == nil && len(y) == 16 && bytes.Equal(x, y)
}

func (s *State) matchesQuestItem(item *Item) bool {
	return s.Site != nil && s.Site.Index >= 0 && item != nil &&
		EqualQuestItemRecords(s.Site.QuestItem.Record, item.Record)
}
