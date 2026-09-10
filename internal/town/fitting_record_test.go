package town

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestFittingUpdatesOriginalRecordAndCarried(t *testing.T) {
	for _, purchase := range []bool{false, true} {
		goods := armourerGoods()
		raw := bytes.Repeat([]byte{0x55}, CarriedStride)
		raw[0x0A] = byte(goods[0].Encumbrance)
		goods[0].Record = hex.EncodeToString(raw)
		s := shopSession(t, 3000, [2]int{0x07, 0x02}, goods)
		s.Party[0].Sheet = &Sheet{Stats: [StatCount]int{0, 0, 150, 0, 0, 0, 0}}
		if purchase {
			s.UIText = fitText()
			s.StartShop()
			yes(s)
			yes(s)
			yes(s)
		} else {
			s.UIText = adjustText()
			s.Party[0].Worn[0] = &goods[0]
			s.StartTrain()
			s.Continue()
			yes(s)
		}
		got := s.Party[0].Worn[0]
		if got == nil || got.Encumbrance != 9 {
			t.Fatalf("購買=%v：合身流程未完成", purchase)
		}
		raw[0x0A] = 9
		if got.Record != hex.EncodeToString(raw) || !bytes.Equal(s.Party[0].Carried[1+BackpackSlots], raw) {
			t.Fatalf("購買=%v：原始物品／隨身記錄未同步重量或改壞其他欄位", purchase)
		}
	}
}
