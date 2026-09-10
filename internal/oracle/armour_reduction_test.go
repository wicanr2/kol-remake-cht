//go:build oracle

package oracle

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/kol-remake-cht/internal/battle"
)

// 此窄測試的入口橋只建立原版區域變數，再跳到 IDA 0x276B1。
// 原版的指標選擇、護甲、耐久與結尾 CALL 均執行；亂數及扣傷常式以參數觀測替代。
// 不驗命中之前的攻防，也不冒充正常玩家路徑或完整傷勢回寫。
func TestArmourReductionOriginal(t *testing.T) {
	o := prepare(t)
	initial := o.Save()
	for _, enemy := range []bool{false, true} {
		for part := 0; part < 6; part++ {
			for _, c := range []struct {
				name                                                          string
				code, rawBonus, rolls, sides, bonus, condition, flags, damage int
				empty                                                         bool
			}{
				{"扣耐久", 0, 0, 1, 2, 0, 200, 0x20, 5, false},
				{"耐久歸零保留", 0, 0, 1, 2, 0, 128, 0x20, 5, false},
				{"耐久負數清除", 0, 0, 1, 2, 0, 127, 0x20, 5, false},
				{"完全阻擋", 32, 0x50, 2, 2, 5, 200, 0x20, 5, false},
				{"負加值無耐久旗標", 76, 0xD0, 3, 7, -5, 200, 0, 5, false},
				{"零傷害負加值", 76, 0xD0, 3, 7, -5, 200, 0, 0, false},
				{"空槽仍擲骰", 0, 0, 1, 2, 0, 0, 0, 5, true},
			} {
				t.Run(fmt.Sprintf("敵方=%v/部位=%d/%s", enemy, part, c.name), func(t *testing.T) {
					o.Restore(initial)
					slot, base, offset := 1, 0xBE88, 0x100
					armor := []int{0, 1, 1, 1, 2, 2}[part]
					if enemy {
						slot, base, offset = 6, 0xCD92, 0x30
						armor = []int{0, 1, 2, 2, 2, 2}[part]
						o.WriteU8(o.DS(uint16(actorBase+slot*actorStride+1)), 1)
					}
					o.WriteU16(o.IDA(0x3237E), uint16(slot))
					before := bytes.Repeat([]byte{0xA5}, 50)
					start := 1 + armor*16
					if c.empty {
						clear(before[start : start+16])
					}
					before[start+8], before[start+11] = byte(c.rawBonus), byte(c.condition)
					before[start+12], before[start+15] = byte(c.code), byte(c.flags)
					addr := o.DS(uint16(base + offset - 1))
					o.WriteBytes(addr, before)
					want := append([]byte(nil), before...)
					remaining := max(0, c.damage-c.bonus-c.rolls)
					condition := c.condition - (c.flags&0x20)*remaining
					if condition < 0 {
						clear(want[start : start+16])
					} else {
						want[start+11] = byte(condition)
					}
					rolls, takes := 0, 0
					o.Stub(o.IDA(rollEnter), func(p *oracle.Oracle) uint32 {
						rolls++
						if int(p.Arg(0)) != c.sides {
							t.Errorf("骰面=%d，預期 %d", p.Arg(0), c.sides)
						}
						return 1
					})
					o.Stub(o.IDA(0x277F4), func(p *oracle.Oracle) uint32 {
						takes++
						if p.Arg(0) != uint16(remaining) || p.Arg(1) != uint16(part) || p.Arg(2) != 0 {
							t.Errorf("扣傷參數=%d,%d,%d，預期 %d,%d,0", p.Arg(0), p.Arg(1), p.Arg(2), remaining, part)
						}
						return 0
					})
					entry := oracle.Addr{Seg: 0x9000, Off: 0}
					target := o.IDA(0x276B1)
					// push bp; mov bp,sp; sub sp,1Ch; push si; push di;
					// mov [bp-8],part; mov [bp-0Eh],damage; jmp far 原版後段。
					bridge := []byte{0x55, 0x8B, 0xEC, 0x83, 0xEC, 0x1C, 0x56, 0x57,
						0xC7, 0x46, 0xF8, byte(part), 0, 0xC7, 0x46, 0xF2, byte(c.damage), 0,
						0xEA, byte(target.Off), byte(target.Off >> 8), byte(target.Seg), byte(target.Seg >> 8)}
					o.WriteBytes(entry, bridge)
					if _, err := o.Call(entry); err != nil {
						t.Fatal(err)
					}
					if rolls != c.rolls || takes != 1 || !bytes.Equal(o.Bytes(addr, 50), want) {
						t.Fatalf("骰數=%d/%d 扣傷呼叫=%d；護甲及鄰格=%x，預期 %x", rolls, c.rolls, takes, o.Bytes(addr, 50), want)
					}
					actor := battle.Actor{Side: battle.SideParty}
					if enemy {
						actor.Side = battle.SideMonster
					}
					for pos := range actor.Armour {
						raw := before[1+pos*16 : 1+(pos+1)*16]
						if c.empty && pos == armor {
							continue
						}
						actor.Armour[pos] = &battle.Item{Record: hex.EncodeToString(raw), Condition: int(raw[11]), Percent: 40}
					}
					if got := actor.AbsorbArmour(c.damage, part, &recordingRand{values: []int{1, 1, 1}}); got != remaining {
						t.Fatalf("remake 傷害=%d，原版=%d", got, remaining)
					}
					for pos, item := range actor.Armour {
						raw := make([]byte, 16)
						if item != nil {
							var err error
							raw, err = hex.DecodeString(item.Record)
							if err != nil || item.Condition != int(raw[11]) || item.Percent != 40 {
								t.Fatalf("remake typed／原始記錄不同步：%+v", item)
							}
						}
						if !bytes.Equal(raw, want[1+pos*16:1+(pos+1)*16]) {
							t.Fatalf("remake 護甲槽 %d 與原版不同：%x", pos, raw)
						}
					}
				})
			}
		}
	}
}
