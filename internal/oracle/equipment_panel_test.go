//go:build oracle

package oracle

import (
	"fmt"
	"github.com/wicanr2/dosgolem/oracle"
	"reflect"
	"testing"
)

func TestBattlePanelSpaceOriginal(t *testing.T) {
	o := prepare(t)
	for _, ea := range []uint32{0x1103F, 0x1B90B} {
		o.StubValue(o.IDA(ea), 0)
	}
	for _, mode := range []uint16{0, 2, 3, 4, 5, 6} {
		for _, initial := range []uint16{0, 1} {
			t.Run(fmt.Sprintf("mode%d/from%d", mode, initial), func(t *testing.T) {
				o.WriteU16(o.IDA(0x2DF3C), 1)
				o.WriteU16(o.IDA(0x2DF3A), 0)
				o.WriteU8(o.DS(0x2DAE), 7)
				// IDA 0x30C7E 就是 DS:0x2DAE；它是第一個圖示，不是另一個獨立旗標。
				o.WriteU16(o.IDA(0x303FC), initial)
				reads, draws := 0, 0
				o.Stub(o.IDA(0x1B100), func(p *oracle.Oracle) uint32 {
					reads++
					if reads > 2 {
						t.Fatal("原版沒有在 Return 結束")
					}
					if reads == 1 {
						return 0x20
					}
					return 0x0D
				})
				o.Stub(o.IDA(0x22EFC), func(p *oracle.Oracle) uint32 { draws++; return 0 })
				result, err := o.Call(o.IDA(0x11722), mode)
				if err != nil {
					t.Fatal(err)
				}
				want, wantDraws := initial, 0
				if mode == 4 || mode == 5 {
					want ^= 1
					wantDraws = 1
				}
				if result != 7 || reads != 2 || draws != wantDraws || o.Word(o.IDA(0x303FC)) != want {
					t.Fatalf("返回=%d 讀鍵=%d 重畫=%d 模式=%d；預期重畫=%d 模式=%d", result, reads, draws, o.Word(o.IDA(0x303FC)), wantDraws, want)
				}
			})
		}
	}
}

func TestBattleTargetPanelSpaceOriginal(t *testing.T) {
	o := prepare(t)
	for _, ea := range []uint32{0x1BA9D, 0x1449C, 0x11703, 0x11722, 0x1C683, 0x1C601, 0x2324F, 0x1B90B, 0x223C0} {
		o.StubValue(o.IDA(ea), 0)
	}
	for _, occupied := range []bool{false, true} {
		t.Run(fmt.Sprintf("occupied%v", occupied), func(t *testing.T) {
			o.WriteU16(o.IDA(0x37D2A), 0)
			o.WriteU16(o.IDA(0x2EA18), 0)
			o.WriteU16(o.IDA(0x303FC), 0)
			o.WriteU16(o.IDA(0x2DF3A), 0)
			o.WriteU8(o.DS(0x6713), 0)
			o.WriteU8(o.DS(0x673A), 0)
			for i := 0; i < 81; i++ {
				v := byte(0)
				if occupied {
					v = 1
				}
				o.WriteU8(o.DS(uint16(0x6AE4+i)), v)
			}
			reads, draws := 0, 0
			o.Stub(o.IDA(0x1B100), func(p *oracle.Oracle) uint32 {
				reads++
				if reads > 2 {
					t.Fatal("目標迴圈沒有結束")
				}
				if reads == 1 {
					return 0x20
				}
				return 0x0D
			})
			o.Stub(o.IDA(0x22EFC), func(p *oracle.Oracle) uint32 { draws++; return 0 })
			if _, err := o.Call(o.IDA(0x24131), 1); err != nil {
				t.Fatal(err)
			}
			wantMode, wantDraws := uint16(0), 0
			if occupied {
				wantMode = 1
				wantDraws = 2
			}
			if reads != 2 || draws != wantDraws || o.Word(o.IDA(0x303FC)) != wantMode {
				t.Fatalf("讀鍵=%d 重畫=%d 模式=%d，預期 %d/%d", reads, draws, o.Word(o.IDA(0x303FC)), wantDraws, wantMode)
			}
		})
	}
}

func TestBattleEquipmentRowsOriginal(t *testing.T) {
	o := prepare(t)
	o.StubValue(o.IDA(0x1BA9D), 0)
	o.StubValue(o.IDA(0x1C2DA), 0)
	o.StubValue(o.IDA(0x1A8FD), 0)
	for _, slot := range []int{0, 5, 6, 17} {
		t.Run(fmt.Sprintf("slot%d", slot), func(t *testing.T) {
			o.WriteU16(o.IDA(0x37D2A), uint16(slot))
			o.WriteU16(o.IDA(0x303FC), 1)
			var offsets []int
			base := 0xBB88 + slot*0x300
			if slot < 6 {
				offsets = []int{0x30, 0x100, 0x110, 0x120, 0x160, 0x190}
				o.WriteU8(o.DS(uint16(base+0x3C)), 0)
			} else {
				base = 0xCD92 + (slot-6)*0x80
				offsets = []int{0x20, 0x30, 0x40, 0x50, 0x60}
			}
			var want [][4]uint16
			for row, off := range offsets {
				o.WriteU8(o.DS(uint16(base+off)), 1)
				want = append(want, [4]uint16{uint16(base + off), 1, 5, uint16(24 + 6*row)})
			}
			var got [][4]uint16
			o.Stub(o.IDA(0x1B94F), func(p *oracle.Oracle) uint32 {
				got = append(got, [4]uint16{p.Arg(0), p.Arg(1), p.Word(p.IDA(0x2DF14)), p.Word(p.IDA(0x2DF16))})
				return 0
			})
			if _, err := o.Call(o.IDA(0x22EFC)); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("裝備解名參數=%v，預期=%v", got, want)
			}
		})
	}
}

func TestBattlePortraitPositionOriginal(t *testing.T) {
	o := prepare(t)
	o.StubValue(o.IDA(0x1BA9D), 0)
	o.StubValue(o.IDA(0x1A8FD), 0)
	for _, slot := range []int{0, 5} {
		o.WriteU16(o.IDA(0x37D2A), uint16(slot))
		o.WriteU16(o.IDA(0x303FC), 0)
		calls := 0
		o.Stub(o.IDA(0x1C53D), func(p *oracle.Oracle) uint32 {
			calls++
			if p.Arg(1) != 7 || p.Arg(2) != 37 || p.Arg(3) != uint16(0x3254+slot*760) {
				t.Fatalf("肖像參數錯誤：%x %d %d %x", p.Arg(0), p.Arg(1), p.Arg(2), p.Arg(3))
			}
			return 0
		})
		if _, err := o.Call(o.IDA(0x22EFC)); err != nil {
			t.Fatal(err)
		}
		if calls != 1 {
			t.Fatalf("肖像繪製次數=%d", calls)
		}
	}
}
