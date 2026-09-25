package town

import (
	"os"
	"path/filepath"
	"testing"
)

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

// 物品交換與全隊旗標回報分成兩個按鍵階段。
func TestTurnInSwapsAndReports(t *testing.T) {
	want := []byte("SWORD-----------")
	give := []byte("CLOAK-----------")
	b := QuestBuilding{Site: 3, Wants: want, Gives: give}
	carried := carriedFixture(want, 5)
	flags := make([]int, QuestFlagBytes)
	SetQuestFlag(flags, 3, QuestDone)
	progress := make([]int, 5)

	slot := SwapQuestItem(carried, b)
	if slot != 5 {
		t.Fatalf("應該在第 5 格交出去，得到 %d", slot)
	}
	if string(carried[5]) != string(give) {
		t.Fatalf("那一格應該換成 %q，得到 %q", give, carried[5])
	}
	if QuestFlag(flags, 3) != QuestDone || progress[0] != 0 {
		t.Fatal("交換物品時不該提前回報")
	}
	if !ReportQuest(flags, progress, 0, b.Site) {
		t.Fatal("台詞讀完後應回報已完成的任務")
	}
	if QuestFlag(flags, 3) != QuestReported {
		t.Fatalf("旗標應該是回報過，得到 %d", QuestFlag(flags, 3))
	}
	if progress[0] != 1 {
		t.Fatalf("已回報數應該是 1，得到 %d", progress[0])
	}
}

// 建築沒有要給的東西時那一格清空；旗標不是「完成」就不記帳。
func TestTurnInWithoutGiveAndWithoutFlag(t *testing.T) {
	want := []byte("FLAG------------")
	b := QuestBuilding{Site: 1, Wants: want}
	carried := carriedFixture(want, 0)
	flags := make([]int, QuestFlagBytes)
	progress := make([]int, 5)

	slot := SwapQuestItem(carried, b)
	if slot != 0 {
		t.Fatalf("應該在第 0 格交出去，得到 %d", slot)
	}
	for _, v := range carried[0] {
		if v != 0 {
			t.Fatalf("沒有要給的東西時那一格該清空，得到 %v", carried[0])
		}
	}
	if ReportQuest(flags, progress, 0, b.Site) || progress[0] != 0 {
		t.Fatal("旗標不是完成時不得記帳")
	}
}

// 身上沒有那件東西就什麼都不做。
func TestTurnInWithoutTheItem(t *testing.T) {
	b := QuestBuilding{Site: 2, Wants: []byte("QUILL-----------")}
	carried := carriedFixture(nil, -1)
	flags := make([]int, QuestFlagBytes)
	SetQuestFlag(flags, 2, QuestDone)
	if slot := SwapQuestItem(carried, b); slot >= 0 {
		t.Fatalf("不該交出任何東西，得到 %d", slot)
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

// 原版先讓玩家讀進門敘述，再交換第一件物品並顯示訊息 6／7；第二次按鍵後
// 才替所有已完成這個任務的人回報，持物者不必是其中之一（RE-139）。
func TestSessionQuestTurnInKeySequence(t *testing.T) {
	want := []byte("SWORD-----------")
	give := []byte("CLOAK-----------")
	door := &Door{Index: 0, IconFlags: [2]int{0, QuestFlagBit}, QuestSite: 3,
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

	s := &Session{
		Party: []Member{
			{Name: "甲", Speaker: true, Quests: make([]int, QuestFlagBytes),
				Progress: make([]int, 5), Carried: carriedFixture(nil, -1)},
			{Name: "乙", Quests: make([]int, QuestFlagBytes),
				Progress: make([]int, 5), Carried: carriedFixture(want, 2)},
			{Name: "丙", Quests: make([]int, QuestFlagBytes),
				Progress: make([]int, 5), Carried: carriedFixture(want, 3)},
		},
		Mode: ModeInside, Inside: &Inn{Icons: []int{IconSprite(IconSpacebar)},
			Dialog: []string{"進門敘述"}},
		Outdoor: &Outdoor{Town: 0}, door: door, Trans: &Translations{},
		Text: &Text{Towns: []townText{{Town: 0, Buildings: []BuildingText{{Door: 0,
			Messages: []Message{
				{Index: 6, Display: []string{"交付第一段"}},
				{Index: 7, Display: []string{"交付第二段"}},
			}}}}}},
	}
	SetQuestFlag(s.Party[0].Quests, 3, QuestDone)
	SetQuestFlag(s.Party[2].Quests, 3, QuestDone)
	if s.TurnedIn != nil || string(s.Party[1].Carried[2]) != string(want) {
		t.Fatal("讀進門敘述之前不得交付")
	}
	s.Key("SPACE")

	if len(s.TurnedIn) != 1 {
		t.Fatalf("一次只應交換第一件，得到 %d", len(s.TurnedIn))
	}
	got := s.TurnedIn[0]
	if got.Member != 1 || got.Slot != 2 || got.Reported || got.Site != 3 {
		t.Fatalf("交付結果不對：%+v", got)
	}
	if len(s.Inside.Dialog) != 2 || s.Inside.Dialog[0] != "交付第一段" ||
		s.Inside.Dialog[1] != "交付第二段" {
		t.Fatalf("未顯示訊息 6／7：%v", s.Inside.Dialog)
	}
	if string(s.Party[1].Carried[2]) != string(give) {
		t.Fatalf("那一格應該換成 %q", give)
	}
	if string(s.Party[2].Carried[3]) != string(want) {
		t.Fatal("同一次進門不該交換第二件")
	}
	if s.Party[0].Progress[0] != 0 || s.Party[2].Progress[0] != 0 {
		t.Fatal("交付台詞未讀完，不得提前回報")
	}
	s.Key("SPACE")
	if !s.TurnedIn[0].Reported || QuestFlag(s.Party[0].Quests, 3) != QuestReported ||
		QuestFlag(s.Party[2].Quests, 3) != QuestReported ||
		s.Party[0].Progress[0] != 1 || s.Party[2].Progress[0] != 1 ||
		s.Party[1].Progress[0] != 0 {
		t.Fatalf("全隊回報狀態不符：%+v", s.Party)
	}
	// 不收任務物品的建築什麼都不做。
	if s.swapFirstQuestItem(plain) || string(s.Party[2].Carried[3]) != string(want) {
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

// 抽取器輸出的任務欄位要一路帶到 Door。這一條擋的是實際踩過的坑：
// `LoadBuildings` 建 Door 時漏抄那六個欄位，結果 typed data 有、規則層也對，
// 但玩家走進收任務的建築什麼都不會發生——單元測試全綠，遊戲裡從來沒生效過。
func TestLoadBuildingsKeepsQuestFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "buildings.json")
	body := `{"schema":"kol-town-buildings-data-v1","towns":[{"town":7,
      "archive":"TOWN7","doors":[{"index":14,"x":42,"y":23,"icon_flags":[1,4],
      "name":"The Assembly Building","quest_site":4,"quest_final":true,
      "quest_wants":"Shackles",
      "quest_wants_bytes":"89a0235b0b300000000a000000000000",
      "quest_gives":"Black Ingot",
      "quest_gives_bytes":"b130235812e3be8403e828ff54769420"}]}]}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("寫入失敗：%v", err)
	}
	loaded, err := LoadBuildings(path)
	if err != nil {
		t.Fatalf("讀取失敗：%v", err)
	}
	door := loaded.Town(7).DoorAt(42, 23)
	if door == nil {
		t.Fatal("找不到那扇門")
	}
	q := door.Quest()
	if q == nil {
		t.Fatal("這一棟應該收任務物品——載入器把 quest 欄位丟掉了")
	}
	if q.Site != 4 || !q.Final || len(q.Wants) != CarriedStride ||
		len(q.Gives) != CarriedStride {
		t.Fatalf("任務欄位讀成 %+v", q)
	}
	if door.QuestWants != "Shackles" {
		t.Fatalf("要的東西讀成 %q", door.QuestWants)
	}
}
