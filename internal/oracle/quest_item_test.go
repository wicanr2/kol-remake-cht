//go:build oracle

package oracle

import (
	"encoding/hex"
	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"testing"
)

// 真實呼叫 sub_2886C，無常式替代或 RND；包含零 byte 以排除 C 字串比對。
func TestQuestItemRecordOriginal(t *testing.T) {
	o := prepare(t)
	initial := o.Save()
	target := []byte{1, 0, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 0}
	cases := 0
	for mutation := -1; mutation < 128; mutation++ {
		item := append([]byte(nil), target...)
		if mutation >= 0 {
			item[mutation/8] ^= 1 << (mutation % 8)
		}
		for _, flag := range []uint16{0, 1} {
			o.Restore(initial)
			o.WriteU16(o.IDA(0x34968), flag)
			o.WriteU16(o.IDA(0x39A48), 0x8000)
			o.WriteBytes(o.DS(0x8030), item)
			o.WriteBytes(o.DS(0xB956), target)
			result, err := o.Call(o.IDA(0x2886C), 0x30)
			if err != nil {
				t.Fatal(err)
			}
			// 原版回傳 16-bit AX；DX 不屬於此函式的結果。
			got := uint16(result)
			want := flag != 0 && mutation == -1
			if (got != 0) != want {
				t.Fatalf("位元差異=%d 任務旗標=%d 原版=%d", mutation, flag, got)
			}
			actual := flag != 0 && battle.EqualQuestItemRecords(hex.EncodeToString(target), hex.EncodeToString(item))
			if actual != (got != 0) {
				t.Fatal("正式比對與原版不同")
			}
			cases++
		}
	}
	t.Logf("完整 16-byte 任務物品比對：%d 組通過", cases)
}
