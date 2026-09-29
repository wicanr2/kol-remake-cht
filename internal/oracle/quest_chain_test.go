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

// 任務鏈的關鍵點接手（#14，SPEC-014）：從原版某一刻的狀態直接呼叫那一支常式，
// 比它寫回的欄位與 remake 同一條規則的結果。不跑整條路徑，也不比畫面。

// questRecords 是 `sub_28F5D` 走的那一組角色記錄：`[bx-4478h]`，DS:0xBB88 起每筆 0x300。
const questRecords = 0xBB88

// 標成完成（`sub_28F5D`）：24 個地點 × 目標那一名的旗標 byte 0…255，其他五名給不同
// 的值；比六名的旗標 byte 與 `word_34968` 是否被設成 0xFFFF。
func TestQuestDoneMarkOriginal(t *testing.T) {
	o := prepare(t)
	initial := o.Save()
	cases := 0
	for site := 0; site < battle.QuestFlagBytes*4; site++ {
		for v := 0; v < 256; v++ {
			for who := 0; who < 6; who++ {
				o.Restore(initial)
				o.WriteU16(o.IDA(0x2DF70), 0)
				o.WriteU16(o.IDA(0x30A72), uint16(site))
				o.WriteU16(o.IDA(0x34968), 1)
				want := make([][]int, 6)
				changed := false
				for k := 0; k < 6; k++ {
					flags := make([]int, battle.QuestFlagBytes)
					for j := range flags {
						flags[j] = (v*7 + k*37 + j*11) & 0xFF
					}
					if k == who {
						flags[site/4] = v
					}
					for j, f := range flags {
						o.WriteU8(o.DS(uint16(questRecords+k*0x300+0x1E2+j)), byte(f))
					}
					if battle.MarkQuestDone(flags, site) {
						changed = true
					}
					want[k] = flags
				}
				if _, err := o.Call(o.IDA(0x28F5D)); err != nil {
					t.Fatal(err)
				}
				for k := 0; k < 6; k++ {
					for j := range want[k] {
						got := int(o.Byte(o.DS(uint16(questRecords + k*0x300 + 0x1E2 + j))))
						if got != want[k][j] {
							t.Fatalf("地點=%d 值=%#x 目標=%d 記錄=%d byte=%d：原版=%#x remake=%#x",
								site, v, who, k, j, got, want[k][j])
						}
					}
				}
				flag := o.Word(o.IDA(0x34968))
				if (flag == 0xFFFF) != changed {
					t.Fatalf("地點=%d 值=%#x 目標=%d：word_34968=%#x，remake 有改=%t", site, v, who, flag, changed)
				}
				cases++
			}
		}
	}
	t.Logf("任務完成標記：%d 組與 remake 逐 byte 相同", cases)
}

// bridge 在 0x9000:0 放一段入口：照原函式的樣子建好堆疊框（`sub sp, frame`、
// 存 si／di），再遠跳到函式中段。每個出口暫時改成 `mov ax, 第幾個出口 ＋ 1;
// mov sp,bp; pop bp; retf`，所以 `Call` 會在那裡回來，回傳值說明走的是哪一個。
// 每一組都先 Restore，改過的程式碼跟著還原。
func bridge(o *oracle.Oracle, frame byte, entry uint32, exits ...uint32) oracle.Addr {
	at := oracle.Addr{Seg: 0x9000, Off: 0}
	target := o.IDA(entry)
	o.WriteBytes(at, []byte{0x55, 0x8B, 0xEC, 0x83, 0xEC, frame, 0x56, 0x57,
		0xEA, byte(target.Off), byte(target.Off >> 8), byte(target.Seg), byte(target.Seg >> 8)})
	for i, exit := range exits {
		o.WriteBytes(o.IDA(exit), []byte{0xB8, byte(i + 1), 0, 0x8B, 0xE5, 0x5D, 0xCB})
	}
	return at
}

// 戰場物品與參戰者記錄的落點（RE-094／RE-113）。
const (
	groundRecord = 0x9D20 // 十六格 × 16 byte 的物品記錄，第 0 byte 非零才算有東西
	groundX      = 0x2BAE
	groundY      = 0x2D9E
	actorRecord  = 0x6712 // 每名 0x32
	questWanted  = 0xB956 // 任務要的那一件（SPEC-014 §4）
)

type pickupCase struct {
	name      string
	foe       bool
	quest     bool // 任務戰鬥（word_34968 = 1）
	called    bool // 這一場已經喊過 0x113（word_32382 = 1）
	handFull  bool
	items     []int // 地上哪幾格有東西；都擺在腳下，第 0 個以外的名字不同
	questSlot int   // 哪一格是任務物品，-1 表示沒有
	keys      []byte
}

// 撿（`sub_1E3A8:0x202D6`…`0x2050E`）：同一組地上物品、手上狀態與回答，比訊息順序、
// 手上那十六 byte、地上那一格有沒有清掉、`+0x25` 與 `+0x2B`。
func TestQuestPickUpOriginal(t *testing.T) {
	o := prepare(t)
	initial := o.Save()
	quest := []byte{0x51, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 0x0F}
	other := []byte{0x52, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0, 1, 2, 3, 4, 5}
	held := []byte{0x53, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7}
	const who, x, y = 2, 20, 30
	cases := []pickupCase{
		{name: "任務物品、手空、答 y", quest: true, items: []int{0}, questSlot: 0, keys: []byte{'y'}},
		{name: "同一場第二次撿到", quest: true, called: true, items: []int{0}, questSlot: 0, keys: []byte{'Y'}},
		{name: "手上已經有東西", quest: true, handFull: true, items: []int{0}, questSlot: 0, keys: []byte{'y'}},
		{name: "答 n、腳下只有一件", quest: true, items: []int{0}, questSlot: 0, keys: []byte{'n'}},
		{name: "兩件、第一件答 n 第二件答 y", quest: true, items: []int{1, 4}, questSlot: 4, keys: []byte{'N', 'y'}},
		{name: "不是任務物品", quest: true, items: []int{3}, questSlot: -1, keys: []byte{'y'}},
		{name: "隨機遭遇撿到同一件", quest: false, items: []int{0}, questSlot: 0, keys: []byte{'y'}},
		{name: "腳下沒有東西", quest: true, items: nil, questSlot: -1, keys: nil},
		{name: "敵人撿", foe: true, quest: true, items: []int{0}, questSlot: 0, keys: nil},
	}
	for _, c := range cases {
		o.Restore(initial)
		entry := bridge(o, 4, 0x202D6, 0x2050E)
		// 原版這一側。
		o.WriteU16(o.IDA(0x37D2A), who)
		ar := uint16(actorRecord + who*0x32)
		side := byte(0)
		if c.foe {
			side = 1
		}
		o.WriteU8(o.DS(ar+1), side)
		o.WriteU8(o.DS(ar+0x1F), x)
		o.WriteU8(o.DS(ar+0x20), y)
		o.WriteU8(o.DS(ar+0x25), 0x26)
		o.WriteU8(o.DS(ar+0x2B), 1)
		rec := uint16(questRecords + who*0x300)
		o.WriteU8(o.DS(rec), 1)
		hand := make([]byte, 16)
		if c.handFull {
			hand = held
		}
		o.WriteBytes(o.DS(rec+0x30), hand)
		o.WriteU16(o.IDA(0x39A48), rec)
		for i := 0; i < 16; i++ {
			o.WriteBytes(o.DS(uint16(groundRecord+16*i)), make([]byte, 16))
			o.WriteU8(o.DS(uint16(groundX+i)), 0xFF)
			o.WriteU8(o.DS(uint16(groundY+i)), 0xFF)
		}
		for _, i := range c.items {
			item := other
			if i == c.questSlot {
				item = quest
			}
			o.WriteBytes(o.DS(uint16(groundRecord+16*i)), item)
			o.WriteU8(o.DS(uint16(groundX+i)), x)
			o.WriteU8(o.DS(uint16(groundY+i)), y)
		}
		flag := uint16(0)
		if c.quest {
			flag = 1
		}
		o.WriteU16(o.IDA(0x34968), flag)
		o.WriteBytes(o.DS(questWanted), quest)
		called := uint16(0)
		if c.called {
			called = 1
		}
		o.WriteU16(o.IDA(0x32382), called)
		var said []int
		keys := append([]byte(nil), c.keys...)
		o.Stub(o.IDA(0x1A8FD), func(p *oracle.Oracle) uint32 { said = append(said, int(p.Arg(0))); return 0 })
		o.Stub(o.IDA(0x1C1A4), func(*oracle.Oracle) uint32 { return 0 })
		o.Stub(o.IDA(0x1C1BF), func(*oracle.Oracle) uint32 { return 0 })
		// 回答用完之後的等鍵是 0x112／0x113 那兩句後面的「按一鍵繼續」，記成 -1。
		o.Stub(o.IDA(0x1B100), func(*oracle.Oracle) uint32 {
			if len(keys) == 0 {
				said = append(said, -1)
				return ' '
			}
			k := keys[0]
			keys = keys[1:]
			return uint32(k)
		})
		if _, err := o.Call(entry); err != nil {
			t.Fatalf("%s：%v", c.name, err)
		}
		gotHand := o.Bytes(o.DS(rec+0x30), 16)
		gotCmd := int(o.Byte(o.DS(ar + 0x25)))
		gotPicked := int(o.Byte(o.DS(ar + 0x2B)))
		gotCalled := o.Word(o.IDA(0x32382))
		cleared := map[int]bool{}
		for _, i := range c.items {
			cleared[i] = o.Byte(o.DS(uint16(groundX+i))) == 0xFF
		}

		// remake 這一側：同樣的地上物品、手上狀態與回答。
		s := &battle.State{Stage: battle.StageField, Actors: make([]battle.Actor, 7),
			Party: make([]battle.Member, 6)}
		site := -1
		if c.quest {
			site = 6
		}
		s.Site = &battle.Site{Index: site}
		s.Site.QuestItem.Name = "Q"
		s.Site.QuestItem.Record = hex.EncodeToString(quest)
		for i := range s.Actors {
			s.Actors[i] = battle.Actor{Side: battle.SideEmpty, Member: -1}
		}
		a := &s.Actors[who]
		a.X, a.Y, a.Command, a.PickedUp, a.Member = x, y, 0x26, 1, who
		if c.foe {
			a.Side, a.Member = battle.SideMonster, -1
		} else {
			a.Side = battle.SideParty
		}
		if c.handFull {
			s.Party[who].HasWeapon = true
			s.Party[who].WeaponName = "HELD"
			s.Party[who].HandItem = &battle.Item{Name: "HELD", Record: hex.EncodeToString(held)}
		}
		for _, i := range c.items {
			item := other
			if i == c.questSlot {
				item = quest
			}
			s.Ground[i] = battle.GroundItem{Name: fmt.Sprintf("ITEM%d", i), X: x, Y: y,
				Item: &battle.Item{Name: fmt.Sprintf("ITEM%d", i), Record: hex.EncodeToString(item)}}
		}
		if c.called {
			// 這一場已經喊過（word_32382 = 1）：先讓 remake 喊一次，把旗標立起來。
			s.QuestItemCall(&battle.Item{Record: hex.EncodeToString(quest)})
		}
		var want []int
		if c.foe {
			if slot := s.GroundUnder(who, 0); slot >= 0 {
				// 0x113 印完一樣要按一鍵才往下。
				if out := s.EnemyPickUp(who, slot); out.Message != 0 {
					want = append(want, out.Message, -1)
				}
			}
		} else if slot := s.GroundUnder(who, 0); slot >= 0 {
			want = append(want, battle.FoundItemRecord)
			s.PickUp = &battle.PickUpAsk{Actor: who, Slot: slot}
			for _, k := range c.keys {
				if s.AnswerPickUp(k == 'y' || k == 'Y') {
					out, _ := s.Act(nil)
					want = append(want, out.Message)
					// 執行段一句一個鍵：0x112／0x113 印完要按一鍵才往下。
					if out.Message != battle.FoundItemRecord {
						want = append(want, -1)
					}
				}
			}
		}
		if fmt.Sprint(said) != fmt.Sprint(want) {
			t.Fatalf("%s：訊息 原版=%#x remake=%#x", c.name, said, want)
		}
		if !c.foe {
			remakeHand := make([]byte, 16)
			if h := s.Party[who].HandItem; h != nil {
				remakeHand, _ = hex.DecodeString(h.Record)
			}
			if !bytes.Equal(gotHand, remakeHand) {
				t.Fatalf("%s：手上 原版=% x remake=% x", c.name, gotHand, remakeHand)
			}
		}
		for i, gone := range cleared {
			if gone != (s.Ground[i].Name == "") {
				t.Fatalf("%s：地上第 %d 格 原版清掉=%t remake=%q", c.name, i, gone, s.Ground[i].Name)
			}
		}
		if gotCmd != a.Command {
			t.Fatalf("%s：+0x25 原版=%#x remake=%#x", c.name, gotCmd, a.Command)
		}
		if c.foe && gotPicked != a.PickedUp {
			t.Fatalf("%s：+0x2B 原版=%d remake=%d", c.name, gotPicked, a.PickedUp)
		}
		shouted := false
		for _, m := range said {
			shouted = shouted || m == battle.QuestItemRecord
		}
		if int(gotCalled)-int(called) != map[bool]int{true: 1, false: 0}[shouted] {
			t.Fatalf("%s：word_32382 %d → %d，喊了 0x113=%t", c.name, called, gotCalled, shouted)
		}
		t.Logf("%s：訊息 %#x、+0x25=%#x 一致", c.name, said, gotCmd)
	}
}

// 交付（`sub_11A93:0x11B5A`…`0x11D0B`）：同一組隊伍身上物品、旗標與已回報數，
// 比交換落在哪一格、六名的旗標與已回報數，以及訊息 6／7 有沒有放。
func TestQuestDeliveryOriginal(t *testing.T) {
	o := prepare(t)
	initial := o.Save()
	wants := []byte{0x61, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	gives := []byte{0x62, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30}
	offsets := town.CarriedOffsets()
	type holder struct{ member, slot int }
	cases := []struct {
		name    string
		enabled bool
		site    int
		absent  []int    // 記錄第 0 byte 為零的槽
		holders []holder // 誰身上哪一格有 wants
		flags   [6]int   // 那一格任務欄位的值 0…3
		prog    [6]int
		noGift  bool
	}{
		{name: "持物者完成、他人未完成", enabled: true, site: 6, holders: []holder{{1, 0}}, flags: [6]int{2, 2, 1, 0, 3, 2}, prog: [6]int{0, 3, 5, 0, 7, 0x16}},
		{name: "非持物者也回報", enabled: true, site: 13, holders: []holder{{4, 5}}, flags: [6]int{2, 0, 2, 2, 1, 2}},
		{name: "兩人持同物只換第一件", enabled: true, site: 21, holders: []holder{{2, 7}, {0, 3}}, flags: [6]int{2, 2, 2, 2, 2, 2}},
		{name: "同一人兩格只換較小位移", enabled: true, site: 3, holders: []holder{{3, 9}, {3, 2}}, flags: [6]int{0, 0, 0, 2, 0, 0}},
		{name: "空槽身上有東西不算", enabled: true, site: 8, absent: []int{0}, holders: []holder{{0, 0}, {5, 1}}, flags: [6]int{2, 2, 2, 2, 2, 2}},
		{name: "不給東西的建築", enabled: true, site: 0, holders: []holder{{0, 0}}, flags: [6]int{2, 2, 0, 0, 0, 0}, noGift: true},
		{name: "這棟不收任務物品", enabled: false, site: 6, holders: []holder{{1, 0}}, flags: [6]int{2, 2, 2, 2, 2, 2}},
		{name: "沒有人帶著", enabled: true, site: 6, flags: [6]int{2, 2, 2, 2, 2, 2}},
	}
	for _, c := range cases {
		o.Restore(initial)
		entry := bridge(o, 0x14, 0x11B5A, 0x11D0B)
		icon := o.Byte(o.IDA(0x2FC01)) &^ 4
		if c.enabled {
			icon |= 4
		}
		o.WriteU8(o.IDA(0x2FC01), icon)
		o.WriteU8(o.IDA(0x2FBFA), byte(c.site))
		o.WriteU16(o.IDA(0x2DF70), 0)
		gift := gives
		if c.noGift {
			gift = make([]byte, 16)
		}
		o.WriteBytes(o.DS(0x1CF4+0x2E0), wants)
		o.WriteBytes(o.DS(0x1CF4+0x2F0), gift)
		div := town.QuestDivisors[c.site&3]
		carried := make([][][]byte, 6)
		flags := make([][]int, 6)
		prog := make([][]int, 6)
		for k := 0; k < 6; k++ {
			rec := uint16(questRecords + k*0x300)
			present := byte(1)
			for _, a := range c.absent {
				if a == k {
					present = 0
				}
			}
			o.WriteU8(o.DS(rec), present)
			carried[k] = make([][]byte, len(offsets))
			for j, off := range offsets {
				item := []byte{byte(0x70 + k), byte(j), 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
				for _, h := range c.holders {
					if h.member == k && h.slot == j {
						item = append([]byte(nil), wants...)
					}
				}
				o.WriteBytes(o.DS(rec+uint16(off)), item)
				carried[k][j] = item
			}
			flags[k] = make([]int, town.QuestFlagBytes)
			for j := range flags[k] {
				flags[k][j] = (k*29 + j*13) & 0xFF
			}
			v := flags[k][c.site/4]
			flags[k][c.site/4] = v&^(3*div) | c.flags[k]*div
			for j, f := range flags[k] {
				o.WriteU8(o.DS(rec+0x1E2+uint16(j)), byte(f))
			}
			prog[k] = []int{c.prog[k], 0, 0, 0, 0}
			o.WriteU8(o.DS(rec+0x1DD), byte(c.prog[k]))
		}
		var said []int
		o.Stub(o.IDA(0x14234), func(p *oracle.Oracle) uint32 { said = append(said, int(p.Arg(0))); return 0 })
		o.Stub(o.IDA(0x144C3), func(*oracle.Oracle) uint32 { return 0 })
		if _, err := o.Call(entry); err != nil {
			t.Fatalf("%s：%v", c.name, err)
		}

		// remake：第一件吻合物品（隊員 0…5、格位由小到大），再全隊回報。
		b := town.QuestBuilding{Site: c.site, Gives: gift}
		if c.enabled {
			b.Wants = wants
		}
		swapped := -1
		for k := 0; k < 6 && swapped < 0; k++ {
			if c.absent != nil && c.absent[0] == k {
				continue
			}
			if slot := town.SwapQuestItem(carried[k], b); slot >= 0 {
				swapped = k*100 + slot
			}
		}
		var want []int
		if swapped >= 0 {
			want = []int{0x200, 0x240}
			for k := 0; k < 6; k++ {
				if c.absent != nil && c.absent[0] == k {
					continue
				}
				town.ReportQuest(flags[k], prog[k], 0, c.site)
			}
		}
		if fmt.Sprint(said) != fmt.Sprint(want) {
			t.Fatalf("%s：訊息 原版=%#x remake=%#x", c.name, said, want)
		}
		for k := 0; k < 6; k++ {
			rec := uint16(questRecords + k*0x300)
			for j, off := range offsets {
				if got := o.Bytes(o.DS(rec+uint16(off)), 16); !bytes.Equal(got, carried[k][j]) {
					t.Fatalf("%s：隊員 %d 位移 %#x 原版=% x remake=% x", c.name, k, off, got, carried[k][j])
				}
			}
			for j, f := range flags[k] {
				if got := int(o.Byte(o.DS(rec + 0x1E2 + uint16(j)))); got != f {
					t.Fatalf("%s：隊員 %d 旗標 byte %d 原版=%#x remake=%#x", c.name, k, j, got, f)
				}
			}
			if got := int(o.Byte(o.DS(rec + 0x1DD))); got != prog[k][0] {
				t.Fatalf("%s：隊員 %d 已回報數 原版=%d remake=%d", c.name, k, got, prog[k][0])
			}
		}
		t.Logf("%s：交換=%d 訊息=%#x 一致", c.name, swapped, said)
	}
}

// 回合末的戰場邊判定與帶物逃走（`sub_1E3A8:0x21045`…`0x211AD`）：同一組站位、手上
// 物品、任務旗標與回答，比問不問 0x109、逃掉沒有，以及六名旗標（真的呼叫 `sub_28F5D`）。
// 出口 1 是 `0x2119B`（不問或答 n，排下一回合），出口 2 是 `0x211AD`（逃走的共同出口）。
func TestQuestFleeEdgeOriginal(t *testing.T) {
	o := prepare(t)
	initial := o.Save()
	quest := []byte{0x51, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 0x0F}
	other := []byte{0x52, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0, 1, 2, 3, 4, 5}
	type member struct {
		side, x, y int
		holds      bool
		down       bool
	}
	cases := []struct {
		name    string
		quest   bool
		members []member
		key     byte
	}{
		{name: "持物者在上緣、答 y", quest: true, key: 'y', members: []member{
			{0, 30, 4, true, false}, {0, 27, 51, false, true}, {0, 27, 52, false, true},
			{0, 22, 23, false, true}, {0, 23, 50, false, true}, {0, 21, 47, false, true}}},
		{name: "持物者在右緣、答 n", quest: true, key: 'N', members: []member{
			{0, 0x3B, 20, true, false}, {0, 27, 51, false, false}, {1, 4, 4, false, false},
			{2, 0, 0, false, false}, {2, 0, 0, false, false}, {2, 0, 0, false, false}}},
		{name: "持物者不在邊上", quest: true, key: 'y', members: []member{
			{0, 30, 5, true, false}, {0, 4, 20, false, false}, {0, 27, 52, false, false},
			{0, 22, 23, false, false}, {0, 23, 50, false, false}, {0, 21, 47, false, false}}},
		{name: "全隊在邊上、沒有任務物品", quest: true, key: 'y', members: []member{
			{0, 4, 9, false, false}, {0, 9, 4, false, false}, {0, 0x3B, 9, false, true},
			{0, 9, 0x3B, false, false}, {1, 30, 30, false, false}, {1, 31, 30, false, false}}},
		{name: "隨機遭遇、持物者在邊上", quest: false, key: 'y', members: []member{
			{0, 30, 4, true, false}, {0, 27, 51, false, false}, {0, 27, 52, false, false},
			{0, 22, 23, false, false}, {0, 23, 50, false, false}, {0, 21, 47, false, false}}},
		{name: "敵人拿著站在邊上不算", quest: true, key: 'y', members: []member{
			{1, 30, 4, true, false}, {0, 27, 51, false, false}, {0, 27, 52, false, false},
			{0, 22, 23, false, false}, {0, 23, 50, false, false}, {0, 21, 47, false, false}}},
	}
	const site = 6
	for _, c := range cases {
		o.Restore(initial)
		entry := bridge(o, 4, 0x21045, 0x2119B, 0x211AD)
		flag := uint16(0)
		if c.quest {
			flag = 1
		}
		o.WriteU16(o.IDA(0x34968), flag)
		o.WriteU16(o.IDA(0x30A72), site)
		o.WriteU16(o.IDA(0x2DF70), 0)
		o.WriteBytes(o.DS(questWanted), quest)
		s := &battle.State{Stage: battle.StageField, Actors: make([]battle.Actor, 6),
			Party: make([]battle.Member, 6)}
		index := -1
		if c.quest {
			index = site
		}
		s.Site = &battle.Site{Index: index}
		s.Site.QuestItem.Record = hex.EncodeToString(quest)
		party := 0
		for k, m := range c.members {
			ar := uint16(actorRecord + k*0x32)
			o.WriteU8(o.DS(ar+1), byte(m.side))
			o.WriteU8(o.DS(ar+0x1F), byte(m.x))
			o.WriteU8(o.DS(ar+0x20), byte(m.y))
			rec := uint16(questRecords + k*0x300)
			o.WriteU8(o.DS(rec), 1)
			item := other
			if m.holds {
				item = quest
			}
			o.WriteBytes(o.DS(rec+0x30), item)
			flags := make([]int, battle.QuestFlagBytes)
			for j := range flags {
				flags[j] = (k*41 + j*7) & 0xFF
			}
			flags[site/4] = flags[site/4]&^(3*16) | 1*16
			for j, f := range flags {
				o.WriteU8(o.DS(rec+0x1E2+uint16(j)), byte(f))
			}
			a := &s.Actors[k]
			a.Side, a.X, a.Y, a.Member = battle.Side(m.side), m.x, m.y, k
			// 倒下只記在 remake 這一側：原版這一段只比 +0x01 與座標，不看傷勢。
			if m.down {
				a.Head, a.Chest = 0, 0
			} else {
				a.Head, a.Chest = 10, 10
			}
			if m.side == 0 {
				party++
			}
			s.Party[k].Quests = flags
			s.Party[k].HandItem = &battle.Item{Record: hex.EncodeToString(item)}
		}
		o.WriteU16(o.IDA(0x2DF2A), uint16(party))
		var said []int
		o.Stub(o.IDA(0x1A8FD), func(p *oracle.Oracle) uint32 { said = append(said, int(p.Arg(0))); return 0 })
		o.Stub(o.IDA(0x1C1A4), func(*oracle.Oracle) uint32 { return 0 })
		o.Stub(o.IDA(0x1C1BF), func(*oracle.Oracle) uint32 { return 0 })
		o.Stub(o.IDA(0x1B100), func(*oracle.Oracle) uint32 { return uint32(c.key) })
		exit, err := o.Call(entry)
		if err != nil {
			t.Fatalf("%s：%v", c.name, err)
		}
		fled := uint16(exit) == 2

		offer := s.FleeOffer()
		var want []int
		wantFled := false
		if offer.Ask {
			want = []int{battle.FleeAskRecord}
			if c.key == 'y' || c.key == 'Y' {
				s.Flee()
				wantFled = true
			}
		}
		if fmt.Sprint(said) != fmt.Sprint(want) || fled != wantFled {
			t.Fatalf("%s：原版 訊息=%#x 逃走=%t；remake 訊息=%#x 逃走=%t（%+v）",
				c.name, said, fled, want, wantFled, offer)
		}
		for k := range c.members {
			rec := uint16(questRecords + k*0x300)
			for j, f := range s.Party[k].Quests {
				if got := int(o.Byte(o.DS(rec + 0x1E2 + uint16(j)))); got != f {
					t.Fatalf("%s：第 %d 名旗標 byte %d 原版=%#x remake=%#x", c.name, k, j, got, f)
				}
			}
		}
		t.Logf("%s：邊上 %d 名、帶物 %d 名、訊息 %#x、逃走=%t 一致", c.name, offer.Edge, offer.Carrying, said, fled)
	}
}
