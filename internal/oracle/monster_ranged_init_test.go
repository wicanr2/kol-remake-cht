//go:build oracle

package oracle

import (
	"testing"
	"github.com/wicanr2/kol-remake-cht/internal/rng"
	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/dosgolem/oracle"
)

// SPEC-029 草案的獨立模型；尚未接入正式規則。雙方固定相同測試種子。
func TestMonsterRangedConversionOriginal(t *testing.T) {
	o := prepare(t)
	initial := o.Save()
	sides := [...]int{2, 3, 5, 7, 9, 11, 19, 99}
	for i, value := range sides {
		if int(o.Byte(o.DS(uint16(0xD4+i)))) != value { t.Fatal("原版骰面表不符") }
	}
	for _, seed := range []int{0, 1, 65535} {
		for code := 0; code < 256; code++ {
			o.Restore(initial)
			o.WriteU16(o.DS(0x1600), uint16(seed))
			o.WriteU16(o.DS(0x1602), 0)
			r := rng.New(seed)
			want := battle.MonsterRangedRoll(code, r)
			result, err := o.Call(o.IDA(0x222BF), uint16(code))
			if err != nil { t.Fatal(err) }
			state := uint32(o.Word(o.DS(0x1600))) | uint32(o.Word(o.DS(0x1602)))<<16
			if uint16(result) != uint16(want) || state != r.State() {
				t.Fatalf("seed=%d code=%02x：原版值／狀態 %d/%08x，模型 %d/%08x", seed, code, uint16(result), state, want, r.State())
			}
		}
	}
	t.Log("256 種代碼 × 3 組固定種子，返回值與亂數消耗均符合；尚未驗初始化機率閘門及正常 GUI")
}

// 橋接只建立原函式區域變數堆疊，出口截在下一階段前；保留機率與轉換的真實呼叫。
func TestMonsterRangedGateOriginal(t *testing.T) {
	o := prepare(t)
	initial := o.Save()
	cases := 0
	for _, seed := range []int{0, 1, 65535} {
		for chance := 0; chance < 256; chance++ {
			for _, code := range []int{0, 0x3F, 0xFF} {
				o.Restore(initial)
				o.WriteU16(o.DS(0x1600), uint16(seed))
				o.WriteU16(o.DS(0x1602), 0)
				o.WriteU16(o.IDA(0x2FBC0), 6)
				o.WriteU8(o.DS(0xCD92+0x1E), byte(chance))
				o.WriteU8(o.DS(0xCD92+0x1F), byte(code))
				o.WriteU8(o.DS(0x6712+6*50), 0xA5)
				entry := oracle.Addr{Seg: 0x9000, Off: 0}
				target := o.IDA(0x1ED38)
				o.WriteBytes(entry, []byte{0x55, 0x8B, 0xEC, 0x83, 0xEC, 4,
					0xEA, byte(target.Off), byte(target.Off>>8), byte(target.Seg), byte(target.Seg>>8)})
				o.WriteBytes(o.IDA(0x1EDDC), []byte{0x8B, 0xE5, 0x5D, 0xCB})
				r := rng.New(seed)
				want := battle.MonsterRangedValue(chance, code, r)
				if _, err := o.Call(entry); err != nil { t.Fatal(err) }
				got := int(o.Byte(o.DS(0x6712+6*50)))
				state := uint32(o.Word(o.DS(0x1600))) | uint32(o.Word(o.DS(0x1602)))<<16
				if got != want || state != r.State() {
					t.Fatalf("seed=%d chance=%d code=%02x：原版 %d/%08x，模型 %d/%08x", seed, chance, code, got, state, want, r.State())
				}
				cases++
			}
		}
	}
	t.Logf("初始化機率與條件式轉換 %d 組固定種子比較通過", cases)
}
