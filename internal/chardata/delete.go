package chardata

import "fmt"

// DeleteSlot 依原版的刪除寫入把一個槽位標成空的。
//
// 原版 DELETE CHARACTER 只寫一個 byte：把該槽位名字的第一個 byte 寫成 0。收據是
// docs/verification/chardata-delete/：對固定原版 CHARDATA 刪掉第 2 名角色之後，整份
// 12288 bytes 只有 slot 1 的 offset 0x000 由 0x50 變成 0x00，其餘欄位與 CHARPICS、
// SHLDPICS 完全沒有變化。
func DeleteSlot(file []byte, slot int) error {
	if len(file) != RecordBytes*SlotCount {
		return fmt.Errorf("CHARDATA 長度 %d 不是 %d", len(file), RecordBytes*SlotCount)
	}
	if slot < 0 || slot >= SlotCount {
		return fmt.Errorf("槽位 %d 超出範圍", slot)
	}
	file[slot*RecordBytes+offName] = 0
	return nil
}

// DeletedOffsets 是刪除會寫到的槽位內位移，供收據逐項核對。
func DeletedOffsets() []int { return []int{offName} }
