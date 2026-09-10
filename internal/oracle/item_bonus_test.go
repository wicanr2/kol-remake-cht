//go:build oracle

package oracle

import "testing"

// 驗證護甲／武器共用加值解碼；不代表整條攻擊或玩家路徑已對拍。
func TestSignedItemBonusOriginal(t *testing.T) {
	o := prepare(t)
	for raw := 0; raw < 256; raw++ {
		want := (raw & 0x70) >> 4
		if raw&0x80 != 0 {
			want = -want
		}
		got, err := o.Call(o.IDA(0x26D3C), uint16(raw))
		if err != nil {
			t.Fatal(err)
		}
		if int(int16(got)) != want {
			t.Fatalf("原始值=%02x：得到 %d，預期 %d", raw, int16(got), want)
		}
	}
	t.Log("全部 256 個位元組值通過，包含正負零、正負七及不影響結果的低四位")
}
