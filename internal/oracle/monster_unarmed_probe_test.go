//go:build oracle

package oracle

import (
	"fmt"
	"sort"
	"testing"
)

// 徒手那四式，怪物到底從哪裡取那件穿戴品？
//
// RE-131 解出隊員那一側：`sub_26AC2` 依攻擊選項查 `0x0FD2`（`00 50 60 00 40 10`）
// 加 `0x100`，除以一格 `0x10` 得槽號，四式對上 FEET／SHIELD／HEAD／HANDS。
// 同一份筆記在結尾留了一個沒查的洞：
//
//	怪物那一側沒接——同一張表加 `+0x30` 落在 `0x30`／`0x80`／`0x90`／`0x70`／`0x40`，
//	其中三個在怪物記錄的已知版面之外，還沒查那幾格是什麼。
//
// 這個洞不是紙上問題：怪物本體記錄的**步幅是 128**（`monsterStride`，RE-080 與
// SPEC-017 一致），而 `+0x80`／`+0x90` 已經越過那一格的邊界——真要那樣取，讀到的
// 會是**下一隻怪物記錄的開頭**。而 remake 的 `NewQuest` 目前用
// `unarmedFrom(monster.Worn)` 填了怪物的徒手武器，取的是 `monsters.json` 那份
// **母表**的七格穿戴欄（母表記錄大得多，`+0x80`／`+0x90` 在它裡面是有內容的）。
//
// 兩者只可能有一邊對。靜態掃描回答不了這個問題——用算出來的指標取的存取在位元組
// 層面看不見——所以這裡開讀取監看，讓原版自己說它讀了哪裡。
func TestProbeMonsterUnarmedSlotReads(t *testing.T) {
	const (
		slot       = 6    // 第一個敵人格
		cmdUnarmed = 0x29 // 徒手的指令碼（+0x25）
		optionOff  = 0x23 // 攻擊選項
		commandOff = 0x25
		watchLo    = 0x20 // 從武器那一格開始看
		watchHi    = 0xC0 // 看到超過本體記錄邊界 0x80 之後兩格
	)
	// 攻擊選項 → RE-131 那張表算出來的穿戴欄槽號與位移。
	styles := []struct {
		option int
		name   string
		slot   int
	}{
		{1, "踢（FEET）", 5},
		{2, "盾擊（SHIELD）", 6},
		{3, "頭槌（HEAD）", 0},
		{4, "拳（HANDS）", 4},
	}

	// 兩種怪物：手上有武器的，以及武器那一格整格清零的。
	// 後者是要排除「因為牠有武器所以走武器路徑」這個可能——remake 判斷徒手用的
	// 正是 `monster.Weapon == nil`，那一格空了才是真正的赤手空拳。
	for _, armed := range []bool{true, false} {
		for _, st := range styles {
			name := fmt.Sprintf("有武器_選項%d_%s", st.option, st.name)
			if !armed {
				name = fmt.Sprintf("空手_選項%d_%s", st.option, st.name)
			}
			runProbe(t, name, armed, st.option, st.slot)
		}
	}
}

func runProbe(t *testing.T, name string, armed bool, option, wornSlot int) {
	const (
		slot       = 6
		cmdUnarmed = 0x29
		optionOff  = 0x23
		commandOff = 0x25
		watchLo    = 0x20
		watchHi    = 0xC0
	)
	{
		st := struct {
			option int
			slot   int
		}{option, wornSlot}
		t.Run(name, func(t *testing.T) {
			o := prepare(t)
			rec := uint16(actorBase + actorStride*slot)
			body := uint16(monsterBase + monsterStride*(slot-6))

			o.WriteU8(o.DS(rec+sideFlag), 1) // 這一格是敵人
			o.WriteU8(o.DS(rec+penaltyOff), 0)
			o.WriteU8(o.DS(rec+penaltyDef), 0)
			o.WriteU16(o.IDA(bodyPtr), body)
			o.WriteU16(o.IDA(encounterLevel), 5)
			// 走徒手那條分支：規劃過了（+0x22）、指令碼是徒手、選這一式。
			o.WriteU8(o.DS(rec+plannedFlag), 1)
			o.WriteU8(o.DS(rec+commandOff), cmdUnarmed)
			o.WriteU8(o.DS(rec+optionOff), uint8(st.option))
			// 每一格穿戴欄放不同的可辨識值，才看得出讀到的是哪一格。
			for off := uint16(watchLo); off < watchHi; off += 0x10 {
				for k := uint16(0); k < 0x10; k++ {
					o.WriteU8(o.DS(body+off+k), uint8(off))
				}
			}
			if !armed {
				// 武器那一格整格清零＝赤手空拳。
				for k := uint16(0); k < 0x10; k++ {
					o.WriteU8(o.DS(body+0x20+k), 0)
				}
			}

			lo := o.DS(body + watchLo).Linear()
			hi := o.DS(body + watchHi).Linear()
			log := o.WatchReadsAt(lo, hi)
			_, err := o.Call(o.IDA(skillFn), uint16(slot))
			o.StopWatchingReads()
			if err != nil {
				t.Fatalf("呼叫 sub_26AC2：%v", err)
			}

			// 讀到的位址收斂成「本體記錄的哪幾格」。
			touched := map[uint16]int{}
			for _, r := range *log {
				touched[(uint16(watchLo)+r.Off)&0xF0]++
			}
			var slots []uint16
			for off := range touched {
				slots = append(slots, off)
			}
			sort.Slice(slots, func(i, j int) bool { return slots[i] < slots[j] })

			want := uint16(0x30 + st.slot*0x10)
			for _, off := range slots {
				mark := ""
				if off == want {
					mark = "  ← RE-131 那張表算出來的那一格"
				}
				if off >= 0x80 {
					mark += "  ⚠ 越過本體記錄邊界 0x80"
				}
				t.Logf("  本體記錄 +%#04x：讀了 %d 次%s", off, touched[off], mark)
			}
			if len(slots) == 0 {
				t.Log("  這一式完全沒有碰本體記錄的穿戴欄區——" +
					"表示 sub_26AC2 這條路不取穿戴品，取的地方在別處")
			}
			t.Logf("選項 %d（%s）預期會碰 +%#04x（槽 %d）；實際碰到 %v",
				st.option, map[bool]string{true: "有武器", false: "空手"}[armed],
				want, st.slot, slots)
		})
	}
}

// 正對照：同一套設定用在隊員身上，四式必須各自讀到不同的穿戴格。
//
// 沒有這一支，上面那個結果證明不了任何事——四式都只讀 `+0x20` 有兩種可能：
// 怪物那條路本來就只取武器（結論），或者**我的設定根本沒讓它走進徒手分支**
// （假陰性，而且它安靜得跟成功一模一樣）。隊員那一側 RE-131 已經解出來了，
// 拿它當已知答案：同樣的旗標與選項設下去，如果隊員讀得到四個不同的槽，
// 就表示設定方式沒問題，怪物那一側的「沒碰穿戴欄」是真的。
//
// 隊員的穿戴欄基底是角色記錄 `+0x100`（`[0x0FD2 + 選項] + 0x100`，RE-131／RE-127），
// 與怪物的 `+0x30` 不同——同一個槽號、不同的基底。
func TestProbePartyUnarmedSlotReads(t *testing.T) {
	const (
		slot       = 0
		cmdUnarmed = 0x29
		optionOff  = 0x23
		commandOff = 0x25
		charBase   = 0xBB88 // 角色記錄表（RE-080／129）
		charStride = 0x300
		wornBase   = 0x100 // 隊員的穿戴欄起點
		watchSpan  = 0xA0  // 十格
	)
	styles := []struct {
		option int
		name   string
		slot   int
	}{
		{1, "踢（FEET）", 5},
		{2, "盾擊（SHIELD）", 6},
		{3, "頭槌（HEAD）", 0},
		{4, "拳（HANDS）", 4},
	}

	hits := map[int][]uint16{}
	for _, st := range styles {
		t.Run(fmt.Sprintf("選項%d_%s", st.option, st.name), func(t *testing.T) {
			o := prepare(t)
			rec := uint16(actorBase + actorStride*slot)
			char := uint16(charBase + charStride*slot)

			o.WriteU8(o.DS(rec+sideFlag), 0) // 這一格是隊員
			o.WriteU8(o.DS(rec+penaltyOff), 0)
			o.WriteU8(o.DS(rec+penaltyDef), 0)
			o.WriteU8(o.DS(rec+plannedFlag), 1)
			o.WriteU8(o.DS(rec+commandOff), cmdUnarmed)
			o.WriteU8(o.DS(rec+optionOff), uint8(st.option))
			for off := uint16(0); off < watchSpan; off += 0x10 {
				for k := uint16(0); k < 0x10; k++ {
					o.WriteU8(o.DS(char+wornBase+off+k), uint8(off))
				}
			}

			lo := o.DS(char + wornBase).Linear()
			log := o.WatchReadsAt(lo, o.DS(char+wornBase+watchSpan).Linear())
			_, err := o.Call(o.IDA(skillFn), uint16(slot))
			o.StopWatchingReads()
			if err != nil {
				t.Fatalf("呼叫 sub_26AC2：%v", err)
			}

			touched := map[uint16]int{}
			for _, r := range *log {
				touched[r.Off&0xF0]++
			}
			var slots []uint16
			for off := range touched {
				slots = append(slots, off)
			}
			sort.Slice(slots, func(i, j int) bool { return slots[i] < slots[j] })
			hits[st.option] = slots
			want := uint16(st.slot * 0x10)
			t.Logf("選項 %d 預期會碰穿戴欄 +%#04x（槽 %d）；實際碰到 %v（相對 +0x100）",
				st.option, want, st.slot, slots)
		})
	}
}
