//go:build oracle

package oracle

import (
	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"testing"
)

func TestBattleEnemyShootRingOriginal(t *testing.T) {
	o := prepare(t)
	// 只在射擊與移動出口回傳標記，保留原版距離、種類及右臂判斷。
	for ea, value := range map[uint32]byte{0x23F15: 1, 0x23F56: 0} {
		for n, b := range []byte{0xB8, value, 0, 0xCB} {
			o.WriteU8(o.IDA(ea+uint32(n)), b)
		}
	}
	o.WriteU16(o.IDA(0x37D2A), 6)
	for ring := 1; ring < 7; ring++ {
		for _, kind := range []byte{0, 1} {
			o.WriteU16(o.IDA(0x349B2), uint16(ring))
			o.WriteU8(o.DS(0x6712+6*50), kind)
			o.WriteU8(o.DS(0x6712+6*50+0x15), 50)
			result, err := o.Call(o.IDA(0x23EBB))
			if err != nil {
				t.Fatal(err)
			}
			want := uint16(0)
			if ring < 4 && kind != 0 {
				want = 1
			}
			if uint16(result) != want {
				t.Fatalf("原版環 %d 種類 %d 路由 %d，預期 %d", ring, kind, uint16(result), want)
			}
		}
	}
}

func TestBattleRingSearchOriginal(t *testing.T) {
	o := prepare(t)
	for ring := 0; ring < 7; ring++ {
		count := int(o.Byte(o.DS(uint16(0x0FC0 + ring))))
		t.Logf("原版環 %d 表值 %d；remake 同環格數 %d", ring, count, battle.RingSize(ring))
		if count != battle.RingSize(ring) {
			t.Fatalf("環長映射不同")
		}
		for _, first := range []int{1, count - 1} {
			rolls := 0
			o.WriteU16(o.IDA(0x349B2), uint16(ring))
			o.Stub(o.IDA(0x1B5A8), func(p *oracle.Oracle) uint32 {
				rolls++
				if p.Arg(0) != uint16(count-1) {
					t.Fatalf("原版起點骰上限 %d，預期 %d", p.Arg(0), count-1)
				}
				return uint32(first)
			})
			candidate := -1
			o.Stub(o.IDA(0x24BB6), func(p *oracle.Oracle) uint32 {
				if p.Arg(0) != uint16(ring) || p.Arg(2) != 0x11 {
					t.Fatal("候選查詢參數不同")
				}
				candidate = int(p.Arg(1))
				return 1
			})
			previous := uint16(0xFFFF)
			for n := 0; n < count; n++ {
				result, err := o.Call(o.IDA(0x2531C), previous, 0x11)
				if err != nil {
					t.Fatal(err)
				}
				want := (first - 1 + n) % count
				if int(uint16(result)) != want || candidate != want {
					t.Fatalf("第 %d 次續查：結果 %d 候選 %d，預期 %d", n, uint16(result), candidate, want)
				}
				previous = uint16(result)
			}
			candidate = -1
			result, err := o.Call(o.IDA(0x2531C), previous, 0x11)
			if err != nil {
				t.Fatal(err)
			}
			if uint16(result) != 0xFFFF || candidate != -1 || rolls != 1 {
				t.Fatalf("完整繞圈未結束或重擲：%x/%d/%d", uint16(result), candidate, rolls)
			}
		}
	}
}
