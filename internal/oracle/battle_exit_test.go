//go:build oracle

package oracle

import (
	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"testing"
)

// 原版共同出口純傷勢折算，沒有 RND 呼叫。入口橋只建立原函式的堆疊。
func TestBattleExitWoundsOriginal(t *testing.T) {
	o := prepare(t)
	initial := o.Save()
	cases := 0
	for _, old := range []int{0, 1, 127, 128, 254, 255} {
		for current := 0; current < 256; current++ {
			for side := 0; side < 3; side++ {
				o.Restore(initial)
				for slot := 0; slot < 6; slot++ {
					o.WriteU8(o.DS(uint16(actorBase+slot*actorStride+1)), 2)
					o.WriteU8(o.DS(uint16(0x8000+slot*0x300+0x1B)), byte(old))
				}
				o.WriteU8(o.DS(uint16(actorBase+3*actorStride+1)), byte(side))
				o.WriteU8(o.DS(uint16(actorBase+3*actorStride+0xE)), byte(current))
				calls := 0
				o.Stub(o.IDA(0x22278), func(p *oracle.Oracle) uint32 {
					calls++
					if p.Arg(0) != 3 {
						t.Fatalf("寫回非參戰槽 %d", p.Arg(0))
					}
					p.WriteU16(p.IDA(0x39A48), 0x8900)
					return 0
				})
				entry := oracle.Addr{Seg: 0x9000, Off: 0}
				target := o.IDA(0x2186E)
				o.WriteBytes(entry, []byte{0x55, 0x8B, 0xEC, 0x83, 0xEC, 4, 0x56, 0x57,
					0xEA, byte(target.Off), byte(target.Off >> 8), byte(target.Seg), byte(target.Seg >> 8)})
				if _, err := o.Call(entry); err != nil {
					t.Fatal(err)
				}
				want, wantCalls := old, 0
				if side == 0 {
					want = int(byte(old + int(uint16(current-old))/2))
					wantCalls = 1
				}
				if side == 0 && battle.PostBattleWounds(old, current) != want {
					t.Fatal("正式傷勢折算與原版契約不符")
				}
				for slot := 0; slot < 6; slot++ {
					expected := old
					if slot == 3 {
						expected = want
					}
					if got := int(o.Byte(o.DS(uint16(0x8000 + slot*0x300 + 0x1B)))); got != expected {
						t.Fatalf("舊=%d 戰中=%d 陣營=%d 槽=%d：原版=%d 預期=%d", old, current, side, slot, got, expected)
					}
				}
				if calls != wantCalls {
					t.Fatalf("角色取址次數=%d 預期=%d", calls, wantCalls)
				}
				cases++
			}
		}
	}
	t.Logf("原版共同出口傷勢折算與陣營跳過：%d 組通過；不含掉落或正常玩家路徑", cases)
}
