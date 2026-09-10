//go:build oracle

package oracle

import (
	"strings"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/kol-remake-cht/internal/battle"
)

func TestBattleAmmunitionTextOriginal(t *testing.T) {
	o := prepare(t)
	bank := strings.Repeat("\r", 257) + "~V18~H03W:@H:@T:@L:@S:@B:@\rARROWS:\rBOLTS:\r"
	o.WriteBytes(o.DS(0x9E5C), []byte(bank))
	o.WriteBytes(o.DS(0xBB88), []byte("PLAYER\r"))
	o.WriteU16(o.IDA(0x37D2A), 0)
	o.WriteU16(o.IDA(0x303FC), 1)
	o.StubValue(o.IDA(0x1BA9D), 0)
	o.Stub(o.IDA(0x1B94F), func(p *oracle.Oracle) uint32 {
		p.WriteU8(p.DS(0x2D50), 13)
		return 0
	})
	for _, kind := range []byte{6, 18} {
		for _, count := range []byte{0, 19, 255} {
			o.WriteU8(o.DS(0xBB88+0x30), 1)
			o.WriteU8(o.DS(0xBB88+0x3C), 1)
			o.WriteU8(o.DS(0xBB88+0x3D), kind)
			o.WriteU8(o.DS(0x6712), count)
			var chars []byte
			o.Stub(o.IDA(0x1C2DA), func(p *oracle.Oracle) uint32 {
				if p.Word(p.IDA(0x2DF16)) == 60 {
					if p.Word(p.IDA(0x2DF14)) != 3 || p.Word(p.IDA(0x2DF54)) != 1 {
						t.Fatal("彈藥標頭未使用第 3 欄及小字型")
					}
					chars = append(chars, byte(p.Arg(0)))
				}
				// 只替換字元出口，不模擬字寬或聲稱像素已驗。
				return 0
			})
			if _, err := o.Call(o.IDA(0x22EFC)); err != nil {
				t.Fatal(err)
			}
			label := "ARROWS:"
			if kind == 18 {
				label = "BOLTS:"
			}
			digits := map[byte]string{0: "1", 19: "20", 255: "0"}[count]
			if string(chars) != label+digits {
				t.Fatalf("文字 %q，預期 %q", chars, label+digits)
			}
		}
	}
}

// 只執行射擊結算後段；不替換扣除、裝填重設或耗盡判定。
func TestBattleAmmunitionConsumeOriginal(t *testing.T) {
	o := prepare(t)
	// 在原函式入口建立有界堆疊框，再跳到 0x2082C；0x208D2 收尾返回。
	entry := uint32(0x1E3A8)
	delta := uint16(0x2082C - (entry + 9))
	for i, b := range []byte{0x55, 0x89, 0xE5, 0x83, 0xEC, 8, 0xE9, byte(delta), byte(delta >> 8)} {
		o.WriteU8(o.IDA(entry+uint32(i)), b)
	}
	for i, b := range []byte{0x89, 0xEC, 0x5D, 0xCB} {
		o.WriteU8(o.IDA(0x208D2+uint32(i)), b)
	}
	for _, slot := range []int{0, 5} {
		for _, kind := range []byte{6, 10, 11, 18, 22} {
			for count := 0; count < 255; count++ {
				base := uint16(0x6712 + 50*slot)
				o.WriteU16(o.IDA(0x37D2A), uint16(slot))
				o.WriteU8(o.DS(uint16(0xBB88+0x300*slot+0x3D)), kind)
				o.WriteU8(o.DS(base), byte(count))
				o.WriteU8(o.DS(base+0x1D), 1)
				o.WriteU8(o.DS(base+0x25), 0x23)
				if _, err := o.Call(o.IDA(entry)); err != nil {
					t.Fatal(err)
				}
				wantCount, wantCommand, wantLoaded := byte(count-1), byte(0x23), byte(1)
				if kind == 18 || kind == 22 {
					wantCommand, wantLoaded = 0x22, 0
				}
				if count == 0 {
					wantCommand = 0
				}
				if o.Byte(o.DS(base)) != wantCount || o.Byte(o.DS(base+0x25)) != wantCommand || o.Byte(o.DS(base+0x1D)) != wantLoaded {
					t.Fatalf("槽 %d 種類 %d 原值 %d：扣除／指令／裝填結果不符", slot, kind, count)
				}
				a := battle.Actor{Side: battle.SideParty, Kind: count, Flag1D: 1, Command: battle.CmdMissile, Weapon: battle.Weapon{Kind: int(kind)}}
				a.ConsumeAmmunition()
				if a.Kind != int(wantCount) || a.Command != int(wantCommand) || a.Flag1D != int(wantLoaded) {
					t.Fatalf("remake 與原版不同：槽 %d 種類 %d 原值 %d", slot, kind, count)
				}
			}
		}
	}
}

func TestBattleAmmunitionPanelOriginal(t *testing.T) {
	o := prepare(t)
	for _, ea := range []uint32{0x1BA9D, 0x1C2DA, 0x1B94F} {
		o.StubValue(o.IDA(ea), 0)
	}
	for _, slot := range []int{0, 5} {
		for _, kind := range []byte{6, 18, 22, 32} {
			for _, present := range []byte{0, 1} {
				for _, flags := range []byte{0, 1} {
					for _, count := range []byte{0, 19, 254, 255} {
						o.WriteU16(o.IDA(0x37D2A), uint16(slot))
						o.WriteU16(o.IDA(0x303FC), 1)
						base := uint16(0xBB88 + slot*0x300)
						o.WriteU8(o.DS(base+0x30), present)
						o.WriteU8(o.DS(base+0x3C), flags)
						o.WriteU8(o.DS(base+0x3D), kind)
						o.WriteU8(o.DS(uint16(0x6712+50*slot)), count)
						header, number := uint16(0), -1
						o.Stub(o.IDA(0x1A8FD), func(p *oracle.Oracle) uint32 {
							if p.Arg(0) == 0x102 || p.Arg(0) == 0x103 {
								header = p.Arg(0)
							}
							return 0
						})
						o.Stub(o.IDA(0x1B8B8), func(p *oracle.Oracle) uint32 { number = int(p.Arg(0)); return 0 })
						if _, err := o.Call(o.IDA(0x22EFC)); err != nil {
							t.Fatal(err)
						}
						wantHeader, wantNumber := uint16(0), -1
						if present != 0 && flags&1 != 0 && kind&0xE0 == 0 {
							wantHeader, wantNumber = 0x102, int(count)+1
							if kind == 18 || kind == 22 {
								wantHeader = 0x103
							}
							if count == 255 {
								wantNumber = 0
							}
						}
						if header != wantHeader || number != wantNumber {
							t.Fatalf("槽 %d 種類 %d 有物品 %d 旗標 %d 原值 %d：標頭 %x 數字 %d，預期 %x/%d", slot, kind, present, flags, count, header, number, wantHeader, wantNumber)
						}
						s := battle.State{Stage: battle.StageField, Actors: make([]battle.Actor, 6), Party: make([]battle.Member, 6)}
						s.Plan.BodyActor, s.Plan.EquipmentPanel = slot+1, true
						s.Actors[slot].Kind = int(count)
						s.Actors[slot].Member = slot
						s.Party[slot].HasWeapon = present != 0
						s.Party[slot].Weapon = battle.Weapon{Kind: int(kind), Code: int(flags)}
						d, show := s.EquipmentPanel()
						label := map[uint16]string{0: "", 0x102: "ARROWS:", 0x103: "BOLTS:"}[header]
						if !show || d.AmmunitionLabel != label || (header != 0 && d.Ammunition != number) {
							t.Fatalf("remake 彈藥顯示與原版不同：%+v，原版 %x/%d", d, header, number)
						}
					}
				}
			}
		}
	}
}
