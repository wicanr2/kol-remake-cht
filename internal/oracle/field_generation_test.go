//go:build oracle

package oracle

import (
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/rng"
)

// RE-080：原版緩衝區索引 Y*64+X，正式 Field 索引 X*64+Y。
// 橋接只建立區域變數堆疊並截在站位初始化前，保留真實 RND 呼叫。
func TestBattleFieldGenerationOriginal(t *testing.T) {
	o := prepare(t)
	initial := o.Save()
	for _, seed := range []int{0, 1, 65535} {
		for _, fixed := range []bool{false, true} {
			o.Restore(initial)
			o.WriteU16(o.DS(0x1600), uint16(seed))
			o.WriteU16(o.DS(0x1602), 0)
			field := oracle.Addr{Seg: 0x8000, Off: 0}
			o.WriteU16(o.DS(0x2B9A), field.Off)
			o.WriteU16(o.DS(0x2B9C), field.Seg)
			o.WriteU16(o.DS(0x6A98), 0)
			f := battle.NewField()
			if fixed {
				o.WriteU16(o.DS(0x6A98), 1)
				for x := 0; x < 64; x++ {
					for y := 0; y < 64; y++ {
						v := []int{9, 13, 0x24, 0x29}[(x*3+y)%4]
						f.Tiles[x*64+y] = v
						o.WriteU8(oracle.Addr{Seg: field.Seg, Off: uint16(y*64+x)}, byte(v))
					}
				}
			}
			entry := oracle.Addr{Seg: 0x9000, Off: 0}
			target := o.IDA(0x1E58A)
			o.WriteBytes(entry, []byte{0x55, 0x8B, 0xEC, 0x83, 0xEC, 4,
				0xEA, byte(target.Off), byte(target.Off >> 8), byte(target.Seg), byte(target.Seg >> 8)})
			o.WriteBytes(o.IDA(0x1E68C), []byte{0x8B, 0xE5, 0x5D, 0xCB})
			r := rng.New(seed)
			if fixed {
				f.Scatter(r)
			} else {
				f = battle.Generate(r)
			}
			if _, err := o.Call(entry); err != nil {
				t.Fatal(err)
			}
			diff := 0
			for x := 0; x < 64; x++ {
				for y := 0; y < 64; y++ {
					if int(o.Byte(oracle.Addr{Seg: field.Seg, Off: uint16(y*64+x)})) != f.At(x, y) {
						diff++
					}
				}
			}
			state := uint32(o.Word(o.DS(0x1600))) | uint32(o.Word(o.DS(0x1602)))<<16
			if diff != 0 || state != r.State() {
				t.Errorf("seed=%d 固定地圖=%t：差異格數=%d，原版亂數=%08x，重製版=%08x", seed, fixed, diff, state, r.State())
			}
		}
	}
}
