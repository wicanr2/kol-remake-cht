//go:build oracle

package oracle

import (
	"github.com/wicanr2/dosgolem/oracle"
	"testing"
)

func TestBattleCounterScanStartOriginal(t *testing.T) {
	o := prepare(t)
	for _, occupied := range []bool{false, true} {
		for first := 1; first <= 7; first++ {
			o.WriteU16(o.IDA(0x349B2), 0)
			o.WriteU16(o.IDA(0x37D2A), 6)
			o.WriteU8(o.DS(0x6712+6*50+0x28), 3)
			for slot := 0; slot < 2; slot++ {
				o.WriteU8(o.DS(uint16(0x6712+slot*50+0x22)), 0)
				o.WriteU8(o.DS(uint16(0x6712+slot*50+0x25)), 0)
			}
			rolls, visits := 0, 0
			o.Stub(o.IDA(0x1B5A8), func(p *oracle.Oracle) uint32 {
				rolls++
				if p.Arg(0) != 7 {
					t.Fatalf("反制搜尋起點骰=%d，預期 7", p.Arg(0))
				}
				return uint32(first)
			})
			o.Stub(o.IDA(0x24BB6), func(p *oracle.Oracle) uint32 {
				if p.Arg(0) != 0 || p.Arg(2) != 0x19 {
					t.Fatal("反制候選環或遮罩不同")
				}
				position := int(p.Arg(1))
				want := (first - 1 + visits) % 8
				if position != want {
					t.Fatalf("續查位置 %d，預期 %d", position, want)
				}
				visits++
				if occupied && (position == 0 || position == 7) {
					slot := 0
					if position == 7 {
						slot = 1
					}
					p.WriteU16(p.IDA(0x3237E), uint16(slot))
					return 1
				}
				return 0
			})
			result, err := o.Call(o.IDA(0x253E7))
			if err != nil {
				t.Fatal(err)
			}
			if rolls != 1 || visits != 8 {
				t.Fatalf("擲骰／候選次數 %d/%d", rolls, visits)
			}
			wantResult, wantFacing := uint16(0), byte(3)
			if occupied {
				wantResult, wantFacing = 1, 7
				if first == 1 {
					wantFacing = 0
				}
				if o.Word(o.IDA(0x30A78)) != 0x0B {
					t.Fatal("未規劃隊員的招式不同")
				}
			}
			if uint16(result) != wantResult || o.Byte(o.DS(0x6712+6*50+0x28)) != wantFacing {
				t.Fatal("反制搜尋結果或選中方向不同")
			}
		}
	}
}
