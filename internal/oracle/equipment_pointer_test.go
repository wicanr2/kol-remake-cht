//go:build oracle

package oracle

import "testing"

// 只驗記錄選擇及指標副作用，不將指定輸入當作正常玩家路徑。
func TestEquipmentRecordPointerOriginal(t *testing.T) {
	o := prepare(t)
	const pointer = 0xBB78
	for slot := 0; slot < 18; slot++ {
		flag := byte(0)
		want := 0xBB88 + slot*0x300
		if slot >= 6 {
			flag = 1
			want = 0xCD92 + (slot-6)*0x80
		}
		o.WriteU8(o.DS(uint16(actorBase+actorStride*slot+1)), flag)
		o.WriteU16(o.DS(pointer), 0xDEAD)
		result, err := o.Call(o.IDA(sideFn), uint16(slot))
		if err != nil {
			t.Fatal(err)
		}
		if uint16(result) != uint16(flag) || o.Word(o.DS(pointer)) != uint16(want) {
			t.Fatalf("索引=%d 旗標=%d：回傳=%d 指標=%04x，預期 %04x", slot, flag, uint16(result), o.Word(o.DS(pointer)), want)
		}
	}
	// 非零旗標採布林判準；這裡固定合法敵方索引，不建立索引小於六的敵方假資料。
	for _, flag := range []byte{2, 0x7F, 0x80, 0xFF} {
		o.WriteU8(o.DS(actorBase+6*actorStride+1), flag)
		result, err := o.Call(o.IDA(sideFn), 6)
		if err != nil || uint16(result) != 1 || o.Word(o.DS(pointer)) != 0xCD92 {
			t.Fatalf("非零旗標=%02x：回傳=%d 指標=%04x 錯誤=%v", flag, result, o.Word(o.DS(pointer)), err)
		}
	}
	t.Log("六個隊員槽、十二個敵方槽與四組非零旗標通過；指標副作用逐項符合原版公式")
}
