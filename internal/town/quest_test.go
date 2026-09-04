package town

import "testing"

// 身上要掃的那幾格：0x30 起、0x1A0 止，跳過 0x40…0x9F（`0x11B8D`）。
func TestCarriedOffsets(t *testing.T) {
	got := CarriedOffsets()
	if got[0] != 0x30 {
		t.Fatalf("第一格是 %#x，期望 0x30", got[0])
	}
	if last := got[len(got)-1]; last != 0x190 {
		t.Fatalf("最後一格是 %#x，期望 0x190", last)
	}
	for _, off := range got {
		if off >= CarriedSkipLo && off <= CarriedSkipHi {
			t.Fatalf("%#x 那一格不該被掃", off)
		}
	}
	// 0x30 一格，加上 0xA0…0x190 共十六格。
	if len(got) != 17 {
		t.Fatalf("應該掃十七格，得到 %d", len(got))
	}
}

// 四個狀態讀寫都對得上，而且不會動到同一個 byte 裡的別人。
func TestQuestFlagRoundTrip(t *testing.T) {
	flags := make([]int, QuestFlagBytes)
	for _, site := range []int{0, 1, 2, 3, 23} {
		for _, state := range []int{QuestTaken, QuestDone, QuestReported, QuestNone} {
			if !SetQuestFlag(flags, site, state) {
				t.Fatalf("設不了地點 %d 的旗標", site)
			}
			if got := QuestFlag(flags, site); got != state {
				t.Fatalf("地點 %d 讀回 %d，期望 %d", site, got, state)
			}
		}
	}
	// 同一個 byte 裡的四格互不干擾。
	flags = make([]int, QuestFlagBytes)
	SetQuestFlag(flags, 0, QuestReported)
	SetQuestFlag(flags, 1, QuestTaken)
	SetQuestFlag(flags, 2, QuestDone)
	if QuestFlag(flags, 0) != QuestReported || QuestFlag(flags, 1) != QuestTaken ||
		QuestFlag(flags, 2) != QuestDone || QuestFlag(flags, 3) != QuestNone {
		t.Fatalf("四格互相干擾了：%v", flags)
	}
	if SetQuestFlag(flags, 24, QuestDone) || SetQuestFlag(flags, -1, QuestDone) {
		t.Fatal("界外不該寫得進去")
	}
}

func carriedFixture(want []byte, at int) [][]byte {
	out := make([][]byte, len(CarriedOffsets()))
	for i := range out {
		out[i] = make([]byte, CarriedStride)
	}
	if at >= 0 {
		copy(out[at], want)
	}
	return out
}

// 帶著那件東西走進去：換成建築給的那件，旗標變成回報過，已回報數加一。
func TestTurnInSwapsAndReports(t *testing.T) {
	want := []byte("SWORD-----------")
	give := []byte("CLOAK-----------")
	b := QuestBuilding{Site: 3, Wants: want, Gives: give}
	carried := carriedFixture(want, 5)
	flags := make([]int, QuestFlagBytes)
	SetQuestFlag(flags, 3, QuestDone)
	progress := make([]int, 5)

	slot, reported := TurnIn(carried, flags, progress, 0, b)
	if slot != 5 || !reported {
		t.Fatalf("應該在第 5 格交出去並回報，得到 %d／%v", slot, reported)
	}
	if string(carried[5]) != string(give) {
		t.Fatalf("那一格應該換成 %q，得到 %q", give, carried[5])
	}
	if QuestFlag(flags, 3) != QuestReported {
		t.Fatalf("旗標應該是回報過，得到 %d", QuestFlag(flags, 3))
	}
	if progress[0] != 1 {
		t.Fatalf("已回報數應該是 1，得到 %d", progress[0])
	}
}

// 建築沒有要給的東西時那一格清空；旗標不是「完成」就只換東西不記帳。
func TestTurnInWithoutGiveAndWithoutFlag(t *testing.T) {
	want := []byte("FLAG------------")
	b := QuestBuilding{Site: 1, Wants: want}
	carried := carriedFixture(want, 0)
	flags := make([]int, QuestFlagBytes)
	progress := make([]int, 5)

	slot, reported := TurnIn(carried, flags, progress, 0, b)
	if slot != 0 {
		t.Fatalf("應該在第 0 格交出去，得到 %d", slot)
	}
	if reported {
		t.Fatal("旗標不是「完成」就不記帳")
	}
	for _, v := range carried[0] {
		if v != 0 {
			t.Fatalf("沒有要給的東西時那一格該清空，得到 %v", carried[0])
		}
	}
	if progress[0] != 0 {
		t.Fatalf("已回報數不該動，得到 %d", progress[0])
	}
}

// 身上沒有那件東西就什麼都不做。
func TestTurnInWithoutTheItem(t *testing.T) {
	b := QuestBuilding{Site: 2, Wants: []byte("QUILL-----------")}
	carried := carriedFixture(nil, -1)
	flags := make([]int, QuestFlagBytes)
	SetQuestFlag(flags, 2, QuestDone)
	progress := make([]int, 5)
	if slot, reported := TurnIn(carried, flags, progress, 0, b); slot >= 0 || reported {
		t.Fatalf("不該交出任何東西，得到 %d／%v", slot, reported)
	}
	if QuestFlag(flags, 2) != QuestDone {
		t.Fatal("旗標不該動")
	}
}

// 最終任務的門檻。
func TestCanTakeFinal(t *testing.T) {
	progress := []int{FinalQuestReports - 1, 0, 0, 0, 0}
	if CanTakeFinal(progress, 0) {
		t.Fatal("差一個就不該開放")
	}
	progress[0] = FinalQuestReports
	if !CanTakeFinal(progress, 0) {
		t.Fatal("滿了應該開放")
	}
	if CanTakeFinal(progress, 9) {
		t.Fatal("群組界外不該開放")
	}
}

// 進門時的交付：這一棟收的那件在誰身上就換誰的，旗標與已回報數跟著動。
func TestSessionTurnInOnEnter(t *testing.T) {
	want := []byte("SWORD-----------")
	give := []byte("CLOAK-----------")
	door := &Door{IconFlags: [2]int{0, QuestFlagBit}, QuestSite: 3,
		QuestWants: "Truth Sword", QuestWantsBytes: hexOf(want),
		QuestGives: "Flying Cloak", QuestGivesBytes: hexOf(give)}
	if door.Quest() == nil {
		t.Fatal("這一棟應該收任務物品")
	}
	// 沒有那一位就不算收任務物品。
	plain := &Door{QuestWantsBytes: hexOf(want)}
	if plain.Quest() != nil {
		t.Fatal("旗標沒開就不該收")
	}

	s := &Session{Party: []Member{
		{Name: "甲", Quests: make([]int, QuestFlagBytes), Progress: make([]int, 5),
			Carried: carriedFixture(nil, -1)},
		{Name: "乙", Quests: make([]int, QuestFlagBytes), Progress: make([]int, 5),
			Carried: carriedFixture(want, 2)},
	}}
	SetQuestFlag(s.Party[1].Quests, 3, QuestDone)
	s.turnInQuests(door)

	if len(s.TurnedIn) != 1 {
		t.Fatalf("應該只有一名交得出來，得到 %d", len(s.TurnedIn))
	}
	got := s.TurnedIn[0]
	if got.Member != 1 || got.Slot != 2 || !got.Reported || got.Site != 3 {
		t.Fatalf("交付結果不對：%+v", got)
	}
	if string(s.Party[1].Carried[2]) != string(give) {
		t.Fatalf("那一格應該換成 %q", give)
	}
	if QuestFlag(s.Party[1].Quests, 3) != QuestReported {
		t.Fatal("旗標應該變成回報過")
	}
	if s.Party[1].Progress[0] != 1 {
		t.Fatalf("已回報數應該是 1，得到 %d", s.Party[1].Progress[0])
	}
	// 不收任務物品的建築什麼都不做。
	s.turnInQuests(plain)
	if s.TurnedIn != nil {
		t.Fatal("不收任務物品的建築不該有交付紀錄")
	}
}

func hexOf(b []byte) string {
	const digits = "0123456789abcdef"
	out := make([]byte, 0, len(b)*2)
	for _, v := range b {
		out = append(out, digits[v>>4], digits[v&0x0F])
	}
	return string(out)
}
