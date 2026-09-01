package chardata

import "testing"

// 刪除只動一個 byte：槽位名字的第一個字元。
func TestDeleteSlotWritesOneByte(t *testing.T) {
	before := make([]byte, RecordBytes*SlotCount)
	for i := range before {
		before[i] = byte(i%251 + 1) // 不留 0，才看得出唯一被清掉的那個 byte
	}
	after := make([]byte, len(before))
	copy(after, before)
	if err := DeleteSlot(after, 1); err != nil {
		t.Fatal(err)
	}
	changed := []int{}
	for i := range before {
		if before[i] != after[i] {
			changed = append(changed, i)
		}
	}
	if len(changed) != 1 || changed[0] != RecordBytes+offName {
		t.Fatalf("預期只有 offset %d 被寫，實際 %v", RecordBytes+offName, changed)
	}
	if after[changed[0]] != 0 {
		t.Fatalf("刪除應寫入 0，實際 0x%02x", after[changed[0]])
	}
}

func TestDeleteSlotRejectsBadInput(t *testing.T) {
	if err := DeleteSlot(make([]byte, 10), 0); err == nil {
		t.Fatal("長度不符應該失敗")
	}
	if err := DeleteSlot(make([]byte, RecordBytes*SlotCount), SlotCount); err == nil {
		t.Fatal("槽位超出範圍應該失敗")
	}
}
