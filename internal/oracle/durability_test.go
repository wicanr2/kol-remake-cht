//go:build oracle

package oracle

import (
	"bytes"
	"testing"
)

// 僅驗證原版子常式，不作正常玩家路徑或 remake 完成證據。
func TestDurabilityBoundariesOriginal(t *testing.T) {
	o := prepare(t)
	const base, slot, pointer = 0xBB88, 0x30, 0xBB78
	o.WriteU16(o.DS(pointer), base)
	for _, c := range []struct{ condition, flags, count int }{
		{200, 0x20, 1}, {200, 0x20, 2}, {200, 0, 7},
		{40, 0x20, 1}, {32, 0x20, 1}, {31, 0x20, 1},
		{255, 0x20, 7}, {255, 0x20, 8}, {0, 0x20, 0},
	} {
		before := bytes.Repeat([]byte{0xA5}, 18)
		before[1], before[12], before[16] = 1, byte(c.condition), byte(c.flags)
		for i, value := range before {
			o.WriteU8(o.DS(base+slot-1+uint16(i)), value)
		}
		want := append([]byte(nil), before...)
		remaining := c.condition - (c.flags&0x20)*c.count
		if remaining < 0 {
			clear(want[1:17])
		} else {
			want[12] = byte(remaining)
		}
		if _, err := o.Call(o.IDA(0x22084), slot, uint16(c.count)); err != nil {
			t.Fatal(err)
		}
		got := o.Bytes(o.DS(base+slot-1), 18)
		if !bytes.Equal(got, want) {
			t.Fatalf("耐久=%d 旗標=%x 次數=%d：得到 %x，預期 %x", c.condition, c.flags, c.count, got, want)
		}
		t.Logf("耐久=%d 旗標=%x 次數=%d：剩餘=%d 清除=%v；相鄰資料不變", c.condition, c.flags, c.count, remaining, remaining < 0)
	}
}
