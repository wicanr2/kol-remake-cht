//go:build oracle

package oracle

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

// A05／A06／A09 的關鍵點接手（#15，RE-163）：腰帶切換、法術的 Ward 判定、修道院治療。
// 做法同 quest_chain_test.go：狀態寫進原版記憶體，從函式中段接手，比寫回的欄位與訊息順序。

// 切換（`sub_1E3A8:0x2052C`…`0x20621`）：手上 +0x30 與腰帶 +0x190 整段交換，之後依兩格
// 有沒有東西印 0x114／0x115／0x116 或補一個 `!`，最後等一鍵。
func TestBeltSwitchOriginal(t *testing.T) {
	o := prepare(t)
	initial := o.Save()
	sword := []byte{0x61, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 0x21}
	axe := []byte{0x62, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0, 1, 2, 3, 4, 0x22}
	empty := make([]byte, 16)
	const who = 3
	for _, c := range []struct {
		name       string
		hand, belt []byte
	}{
		{"手上與腰帶都有", sword, axe},
		{"只有手上", sword, empty},
		{"只有腰帶", empty, axe},
		{"兩格都空", empty, empty},
	} {
		o.Restore(initial)
		entry := bridge(o, 4, 0x2052C, 0x20621)
		o.WriteU16(o.IDA(0x37D2A), who)
		rec := uint16(questRecords + who*0x300)
		o.WriteU8(o.DS(rec), 1)
		o.WriteBytes(o.DS(rec+0x30), c.hand)
		o.WriteBytes(o.DS(rec+0x190), c.belt)
		var said []int
		o.Stub(o.IDA(0x1A8FD), func(p *oracle.Oracle) uint32 { said = append(said, int(p.Arg(0))); return 0 })
		o.Stub(o.IDA(0x1C2DA), func(p *oracle.Oracle) uint32 { said = append(said, -int(p.Arg(0))); return 0 })
		o.Stub(o.IDA(0x1C1BF), func(*oracle.Oracle) uint32 { return 0 })
		o.Stub(o.IDA(0x1B100), func(*oracle.Oracle) uint32 { said = append(said, -1); return ' ' })
		if _, err := o.Call(entry); err != nil {
			t.Fatalf("%s：%v", c.name, err)
		}
		gotHand := o.Bytes(o.DS(rec+0x30), 16)
		gotBelt := o.Bytes(o.DS(rec+0x190), 16)

		s := &battle.State{Actors: make([]battle.Actor, who+1), Party: make([]battle.Member, who+1)}
		s.Actors[who].Member = who
		m := &s.Party[who]
		if c.hand[0] != 0 {
			m.WeaponName, m.HasWeapon = "HAND", true
			m.HandItem = &battle.Item{Name: "HAND", Record: hex.EncodeToString(c.hand)}
		}
		if c.belt[0] != 0 {
			m.BeltName, m.HasSecondItem = "BELT", true
			m.BeltItem = &battle.Item{Name: "BELT", Record: hex.EncodeToString(c.belt)}
		}
		s.SwitchItem(who)
		records, bang := s.SwitchLine(who)
		want := append([]int(nil), records...)
		if bang {
			want = append(want, -0x21)
		}
		want = append(want, -1)
		if fmt.Sprint(said) != fmt.Sprint(want) {
			t.Fatalf("%s：訊息 原版=%#x remake=%#x", c.name, said, want)
		}
		record := func(it *battle.Item) []byte {
			if it == nil {
				return empty
			}
			b, _ := hex.DecodeString(it.Record)
			return b
		}
		if !bytes.Equal(gotHand, record(m.HandItem)) || !bytes.Equal(gotBelt, record(m.BeltItem)) {
			t.Fatalf("%s：交換後 原版手上=% x 腰帶=% x", c.name, gotHand, gotBelt)
		}
		t.Logf("%s：訊息 %#x 與兩格記錄一致", c.name, said)
	}
}

// 施法擋不擋得住（`sub_2818D:0x28284`…`0x2835F`）：出口 1 是 0x2834C（印 0x16A 沒有效果），
// 出口 2 是 0x2835F（印 0x16B 生效）。目標那一格：隊員是角色記錄 +0x28、敵人是怪物記錄
// +0x1D（`sub_22229` 分派）；倒下由 `sub_22866` 看參戰者記錄 +0x11／+0x13 是否 0xFF。
func TestSpellWardOriginal(t *testing.T) {
	o := prepare(t)
	initial := o.Save()
	guardSets := [][4]int{
		{0x10, 0x20, 0x30, 0x40}, {0x10, 0x10, 0x10, 0x10}, {0x20, 0, 0, 0},
		{0x30, 0x40, 0, 0}, {0x40, 0, 0, 0}, {0x50, 0, 0, 0}, {0x55, 0, 0, 0},
		{0x5F, 0x60, 0, 0}, {0x63, 0x67, 0, 0}, {0x6A, 0x7F, 0x80, 0x9C},
		{0x13, 0x27, 0x3E, 0x41}, {0xF0, 0xFF, 0xE5, 0xA1}, {0, 0, 0, 0}, {0x70, 0x75, 0x7A, 0x7F},
	}
	cases := 0
	for _, target := range []int{2, 9} {
		for _, down := range []bool{false, true} {
			for _, g := range guardSets {
				for ward := 0; ward < 256; ward++ {
					o.Restore(initial)
					entry := bridge(o, 0x12, 0x28284, 0x2834C, 0x2835F)
					o.WriteU16(o.IDA(0x3237E), uint16(target))
					ar := uint16(actorRecord + target*0x32)
					side := byte(0)
					if target >= 6 {
						side = 1
						o.WriteU8(o.DS(uint16(0xCD92+(target-6)*0x80+0x1D)), byte(ward))
					} else {
						o.WriteU8(o.DS(uint16(questRecords+target*0x300+0x28)), byte(ward))
					}
					o.WriteU8(o.DS(ar+1), side)
					head := byte(10)
					if down {
						head = 0xFF
					}
					o.WriteU8(o.DS(ar+0x11), head)
					o.WriteU8(o.DS(ar+0x13), 10)
					for i, v := range g {
						o.WriteU8(o.DS(uint16(0x9E30+i)), byte(v))
					}
					exit, err := o.Call(entry)
					if err != nil {
						t.Fatal(err)
					}
					blocked := uint16(exit) == 1
					if uint16(exit) != 1 && uint16(exit) != 2 {
						t.Fatalf("出口 %d 不在預期內", exit)
					}
					if got := battle.SpellBlocked(battle.Spell{Guards: g}, ward, down); got != blocked {
						t.Fatalf("目標=%d 倒下=%t 護符=%#x Ward=%#x：原版擋下=%t remake=%t",
							target, down, g, ward, blocked, got)
					}
					cases++
				}
			}
		}
	}
	t.Logf("法術 Ward 判定：%d 組一致", cases)
}

// 修道院治療（`sub_11A93:0x1264E`…`0x13D61`）：價錢、問答、金幣不夠、成交後傷勢與
// Nutrition 歸零。金幣經原版自己的 `sub_14CCF`／`sub_14D3A` 讀寫角色記錄 +0x2A…+0x2C。
func TestAbbeyHealOriginal(t *testing.T) {
	o := prepare(t)
	initial := o.Save()
	const who = 2
	cases := 0
	for _, wounds := range []int{0, 1, 2, 17, 255} {
		for _, rate := range []int{9, 12, 255} {
			price := town.HealPrice(wounds, rate)
			for _, gold := range []int{0, price - 1, price, price + 1, 0xFFFFFF} {
				if gold < 0 {
					continue
				}
				for _, accept := range []bool{true, false} {
					o.Restore(initial)
					entry := bridge(o, 0x14, 0x1264E, 0x13D61)
					o.WriteU8(o.IDA(0x2FC00), o.Byte(o.IDA(0x2FC00))|0x40)
					o.WriteU8(o.IDA(0x2FBFD), byte(rate))
					o.WriteU16(o.IDA(0x39A52), who*0x300)
					rec := uint16(questRecords + who*0x300)
					o.WriteU8(o.DS(rec), 1)
					o.WriteU8(o.DS(rec+0x1B), byte(wounds))
					o.WriteU8(o.DS(rec+0x9F), 37)
					o.WriteBytes(o.DS(rec+0x2A), []byte{byte(gold >> 16), byte(gold >> 8), byte(gold)})
					var said []string
					o.Stub(o.IDA(0x1A8FD), func(p *oracle.Oracle) uint32 {
						if r := p.Arg(0); r != 0 {
							said = append(said, fmt.Sprintf("%#x", r))
							return 0
						}
						// 記錄 0 是印 `sub_2D3FA` 排好的那一段（價錢）。
						buf := p.Bytes(p.DS(0x2D50), 16)
						if n := bytes.IndexByte(buf, 0); n >= 0 {
							buf = buf[:n]
						}
						said = append(said, "數字 "+string(buf))
						return 0
					})
					o.Stub(o.IDA(0x1452D), func(*oracle.Oracle) uint32 { return 0 })
					o.Stub(o.IDA(0x144C3), func(*oracle.Oracle) uint32 { said = append(said, "等鍵"); return 0 })
					answer := uint32(1)
					if accept {
						answer = 0
					}
					o.Stub(o.IDA(0x144EF), func(*oracle.Oracle) uint32 { said = append(said, "問"); return answer })
					if _, err := o.Call(entry); err != nil {
						t.Fatal(err)
					}
					g := o.Bytes(o.DS(rec+0x2A), 3)
					gotGold := int(g[0])<<16 | int(g[1])<<8 | int(g[2])
					gotWounds := int(o.Byte(o.DS(rec + 0x1B)))
					gotNutrition := int(o.Byte(o.DS(rec + 0x9F)))

					m := town.Member{Wounds: wounds, Gold: gold, Sheet: &town.Sheet{Nutrition: 37}}
					var want []string
					switch {
					case wounds == 0:
						want = []string{fmt.Sprintf("%#x", town.LookFineRecord), "等鍵"}
					default:
						want = []string{fmt.Sprintf("%#x", town.HealOfferRecord), fmt.Sprintf("%#x", town.CostPrefix),
							fmt.Sprintf("數字 %d", price), fmt.Sprintf("%#x", town.CostSuffix), "問"}
						switch {
						case !accept:
						case m.Gold < price:
							want = append(want, fmt.Sprintf("%#x", town.NotEnoughRecord), "等鍵")
						default:
							m.ApplyHeal(price)
							want = append(want, fmt.Sprintf("%#x", town.HealedRecord), "等鍵")
						}
					}
					if fmt.Sprint(said) != fmt.Sprint(want) {
						t.Fatalf("傷勢=%d 費率=%d 金幣=%d 答應=%t：原版 %q remake %q", wounds, rate, gold, accept, said, want)
					}
					if gotGold != m.Gold || gotWounds != m.Wounds || gotNutrition != m.Sheet.Nutrition {
						t.Fatalf("傷勢=%d 費率=%d 金幣=%d 答應=%t：原版 金幣 %d 傷勢 %d Nutrition %d；remake %d %d %d",
							wounds, rate, gold, accept, gotGold, gotWounds, gotNutrition, m.Gold, m.Wounds, m.Sheet.Nutrition)
					}
					cases++
				}
			}
		}
	}
	t.Logf("修道院治療：%d 組一致", cases)
}
