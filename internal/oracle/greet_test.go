//go:build oracle

package oracle

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

// 進門招呼的關鍵點接手（#32，RE-166）：名稱換算、全隊被趕、發言者被拒、被欣賞。
// 做法同 quest_chain_test.go：狀態寫進原版記憶體，從 `sub_11A93` 中段接手，
// 看它走到哪個出口，並與 remake 的 `town.GreetRule` 比。

// greetMember 是一位隊員的種族、性別、職業與傷勢，寫進角色記錄 `+0x24`、`+0x1A`、`+0x1C`、`+0x1B`。
type greetMember struct {
	race, female, profession, wounds int
}

func (g greetMember) member() town.Member {
	return town.Member{Female: g.female != 0, Wounds: g.wounds,
		Sheet: &town.Sheet{Race: g.race, Profession: g.profession}}
}

// greetTypes 是六種職業裡取三種、四個種族、兩種性別的組合。
func greetTypes() []greetMember {
	var out []greetMember
	for race := 0; race < 4; race++ {
		for female := 0; female < 2; female++ {
			for _, profession := range []int{0, 3, 5} {
				out = append(out, greetMember{race: race, female: female, profession: profession})
			}
		}
	}
	return out
}

const (
	partyCount   = 0x2DF2A // word_2DF2A：隊員數
	speakerWord  = 0x39A52 // word_39A52：發言者的記錄位移
	iconFlagsLo  = 0x2FC00 // byte_2FC00／01：建築圖示旗標
	buildingBase = 0x1CF4  // 建築 payload 在 DS 的位置
	orderOffset  = 0x50    // payload 裡教派記錄的位移
)

// putParty 把隊員寫進前幾格角色記錄，其餘格的「在籍」旗標與教派清掉。
func putParty(o *oracle.Oracle, members []greetMember) {
	for k := 0; k < 6; k++ {
		rec := uint16(questRecords + k*0x300)
		o.WriteBytes(o.DS(rec+0x1A0), make([]byte, 16))
		if k >= len(members) {
			o.WriteU8(o.DS(rec), 0)
			continue
		}
		m := members[k]
		o.WriteU8(o.DS(rec), 1)
		o.WriteU8(o.DS(rec+0x1A), byte(m.female))
		o.WriteU8(o.DS(rec+0x1B), byte(m.wounds))
		o.WriteU8(o.DS(rec+0x1C), byte(m.profession))
		o.WriteU8(o.DS(rec+0x24), byte(m.race))
	}
	o.WriteU16(o.IDA(partyCount), uint16(len(members)))
}

// sub_14307：任何名稱記錄號、種族與性別，換算結果都與 remake 相同。
func TestNameCanonicalOriginal(t *testing.T) {
	o := prepare(t)
	cases := 0
	for race := 0; race < 5; race++ {
		for female := 0; female < 2; female++ {
			o.WriteU8(o.DS(questRecords+0x24), byte(race))
			o.WriteU8(o.DS(questRecords+0x1A), byte(female))
			for record := 0; record < 0x80; record++ {
				v, err := o.Call(o.IDA(0x14307), uint16(record), 0)
				if err != nil {
					t.Fatal(err)
				}
				if got, want := int(uint16(v)), town.CanonicalRecord(record, race, female); got != want {
					t.Fatalf("記錄 %#x 種族 %d 性別 %d：原版 %#x，remake %#x", record, race, female, got, want)
				}
				cases++
			}
		}
	}
	t.Logf("名稱換算：%d 組一致", cases)
}

// 進門後第一個判斷（`0x11D0B`…`0x11E25`）：隊伍不只一人、而且每一位都被討厭，才會走到
// 「Get out」（出口 1，`0x11DF2`）；否則進入問誰開口／招呼（出口 2，`0x11E25`）。
// 建築討厭的名稱由 `sub_14189` 解出，這裡直接替它回記錄號，名稱本身的解碼另有證據。
func TestGreetingExpelOriginal(t *testing.T) {
	o := prepare(t)
	initial := o.Save()
	types := greetTypes()
	cases, expelled := 0, 0
	run := func(dislike int, party []greetMember) {
		o.Restore(initial)
		entry := bridge(o, 0x14, 0x11D0B, 0x11DF2, 0x11E25)
		putParty(o, party)
		o.Stub(o.IDA(0x14189), func(p *oracle.Oracle) uint32 { return uint32(dislike) })
		exit, err := o.Call(entry)
		if err != nil {
			t.Fatal(err)
		}
		var members []town.Member
		for _, m := range party {
			members = append(members, m.member())
		}
		want := 2
		if (town.GreetRule{Dislike: dislike}).Expels(members) {
			want = 1
			expelled++
		}
		if int(uint16(exit)) != want {
			t.Fatalf("討厭 %#x 隊伍 %+v：原版出口 %d，remake %d", dislike, party, exit, want)
		}
		cases++
	}
	dislikes := []int{town.NameNone}
	for i := 0; i < town.NameRecords; i++ {
		dislikes = append(dislikes, i)
	}
	for _, d := range dislikes {
		for _, a := range types {
			run(d, []greetMember{a})
			for _, b := range types {
				run(d, []greetMember{a, b})
			}
		}
	}
	if expelled == 0 {
		t.Fatal("沒有任何一組被趕出去，測不到那一條路")
	}
	t.Logf("全隊被趕：%d 組一致（其中 %d 組被趕）", cases, expelled)
}

// 發言者被選定之後（`0x11E8E`）：被討厭或教派不合是出口 1（`0x11EF3`，訊息 14）；
// 有傷是出口 2（`0x11F5D`）；被欣賞是出口 3（`0x12004`，訊息 15）；其餘是出口 4（`0x12017`）。
func TestGreetingSpeakerOriginal(t *testing.T) {
	o := prepare(t)
	initial := o.Save()
	types := greetTypes()
	const (
		orderNone = iota
		orderSame
		orderOther
	)
	hereOrder := []byte{0xBB, 0xA1, 0x34, 0x28, 0x20, 0x50, 0xC9, 0x80, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88}
	counts := map[int]int{}
	run := func(dislike, like int, m greetMember, scroll bool, order int) {
		o.Restore(initial)
		entry := bridge(o, 0x14, 0x11E8E, 0x11EF3, 0x11F5D, 0x12004, 0x12017)
		putParty(o, []greetMember{m, m})
		o.WriteU16(o.IDA(speakerWord), 0)
		flags := byte(0)
		if scroll {
			flags = 1
		}
		o.WriteU8(o.IDA(iconFlagsLo+1), flags)
		o.WriteBytes(o.DS(buildingBase+orderOffset), hereOrder)
		rec := uint16(questRecords)
		orderHex := ""
		switch order {
		case orderSame:
			// 前八個 byte 相同、後八個不同：原版只比前八個。
			mine := append(append([]byte(nil), hereOrder[:8]...), 9, 9, 9, 9, 9, 9, 9, 9)
			o.WriteBytes(o.DS(rec+0x1A0), mine)
			orderHex = hex.EncodeToString(mine)
		case orderOther:
			mine := append([]byte(nil), hereOrder...)
			mine[7] ^= 0x01
			o.WriteBytes(o.DS(rec+0x1A0), mine)
			orderHex = hex.EncodeToString(mine)
		}
		o.Stub(o.IDA(0x13E82), func(*oracle.Oracle) uint32 { return 0 })
		o.Stub(o.IDA(0x14189), func(p *oracle.Oracle) uint32 {
			if p.Arg(0) == 0x40 {
				return uint32(dislike)
			}
			return uint32(like)
		})
		exit, err := o.Call(entry)
		if err != nil {
			t.Fatal(err)
		}
		mm := m.member()
		if orderHex != "" {
			mm.Order = &town.Order{Record: orderHex}
		}
		rule := town.GreetRule{Dislike: dislike, Like: like, Scroll: scroll, Order: hex.EncodeToString(hereOrder)}
		want := 4
		switch {
		case rule.Refuses(mm):
			want = 1
		case mm.Wounds != 0:
			want = 2
		case rule.Likes(mm):
			want = 3
		}
		if int(uint16(exit)) != want {
			t.Fatalf("討厭 %#x 欣賞 %#x 隊員 %+v 教派要求 %t 隊員教派 %d：原版出口 %d，remake %d",
				dislike, like, m, scroll, order, exit, want)
		}
		counts[want]++
	}
	names := []int{town.NameNone}
	for i := 0; i < town.NameRecords; i++ {
		names = append(names, i)
	}
	// 只看討厭、只看欣賞，各配上有傷與沒傷。
	for _, n := range names {
		for _, m := range types {
			for _, wounds := range []int{0, 2} {
				m.wounds = wounds
				run(n, town.NameNone, m, false, orderNone)
				run(town.NameNone, n, m, false, orderNone)
			}
		}
	}
	// 兩者並存與教派：挑幾個代表性的名稱。
	few := []int{town.NameNone, 1, 2, 9, 14, 20, 30}
	for _, d := range few {
		for _, l := range few {
			for _, m := range types {
				for _, scroll := range []bool{false, true} {
					for _, order := range []int{orderNone, orderSame, orderOther} {
						run(d, l, m, scroll, order)
					}
				}
			}
		}
	}
	for want := 1; want <= 4; want++ {
		if counts[want] == 0 {
			t.Fatalf("出口 %d 一次都沒走到，測不到那一條路", want)
		}
	}
	t.Log(fmt.Sprintf("發言者好惡：被拒 %d、有傷 %d、被欣賞 %d、一般招呼 %d，與 remake 全部一致",
		counts[1], counts[2], counts[3], counts[4]))
}
