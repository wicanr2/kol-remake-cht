package town

import (
	"fmt"
	"reflect"
	"testing"
)

// KOLDATA 記錄 0…38 的原文名稱表（`sub_14189` 掃的那一段）。
var nameTable = []string{
	"Human", "Elven", "Dwarven", "Kelden", "Male", "Female",
	"Barbarian", "Ranger", "Warrion", "Squire", "Darkguard", "Watchman",
	"Plainsman", "Hunter", "Regular", "Highwayman", "Pirate", "Rogue",
	"Tigress", "Amazon", "Huntress", "Plainswoman", "Brekland", "Klvar",
	"Melod", "Pyar", "Thism", "Usip", "Tunneler", "Ratguard",
	"Trollbane", "Militia", "Digger", "Spider Guard", "Orcbane", "Levy",
	"Cliff Guard", "Rock Ranger", "Far Seeker",
}

func greetText() *UIText {
	var records []uiRecord
	for i, name := range nameTable {
		records = append(records, uiRecord{Index: i, Text: name})
	}
	records = append(records,
		uiRecord{Index: GreetingRecord, Text: `"What can I do for you?"`},
		uiRecord{Index: WhoRecord, Text: `"Which one needs help?"`},
		uiRecord{Index: ExpelRecord, Text: `"Get out, scum!"`})
	return &UIText{Schema: "kol-ui-text-data-v1", Banks: []uiBank{{Bank: BankUI, Records: records}}}
}

// member 造一位帶種族與職業的隊員。職業記錄號是 profession + 6。
func member(name string, race, profession int, female bool) Member {
	return Member{Name: name, Female: female, Sheet: &Sheet{Race: race, Profession: profession}}
}

// greetSession 進到門 5，讀完進門敘述，停在招呼流程的第一步。
func greetSession(t *testing.T, dislikes, likes string, flags [2]int, party ...Member) *Session {
	t.Helper()
	return greetSessionWith(t, func(d *Door) {
		d.IconFlags = flags
		d.Dislikes, d.Likes = dislikes, likes
	}, party...)
}

// greetSessionWith 讓呼叫端在進門之前改門的資料。
func greetSessionWith(t *testing.T, mod func(*Door), party ...Member) *Session {
	t.Helper()
	s := playSession(t)
	s.UIText = greetText()
	mod(&s.Buildings.Doors[0])
	b := &s.Text.Towns[0].Buildings[0]
	for _, i := range []int{10, 11, 12, 13, GreetPartyMessage, HostileMessage, FavoredMessage, GreetFarewell} {
		b.Messages = append(b.Messages, Message{Index: i, Display: []string{fmt.Sprintf("msg%d", i)}})
	}
	s.Party = party
	s.Party[0].Speaker = true
	s.Walk(-1, 0)
	s.Walk(1, 0)
	s.Accept()
	for i := 0; i < 5 && !s.greeted; i++ {
		s.Continue()
	}
	if !s.greeted {
		t.Fatal("進門敘述讀不完")
	}
	return s
}

func dialog(s *Session) []string { return s.Inside.Dialog }

func wantIcons(t *testing.T, s *Session, ids ...int) {
	t.Helper()
	var want []int
	for _, id := range ids {
		want = append(want, IconSprite(id))
	}
	if !reflect.DeepEqual(s.Inside.Icons, want) {
		t.Fatalf("圖示列是 %v，應為 %v", s.Inside.Icons, want)
	}
}

// `sub_14307` 的換算：`sub_11A93` 用它把建築的名稱記錄號換成隊員自己那一套。
func TestCanonicalRecord(t *testing.T) {
	for _, c := range []struct {
		record, race, gender, want int
	}{
		{0x7F, 3, 1, 0x7F}, // 沒有這個名稱
		{0x0B, 2, 0, 0x0B}, // 0…11 原樣
		{0x00, 3, 1, 0x00},
		{0x0C, 0, 0, 0x0C}, // 人類：不換
		{0x0C, 0, 1, 0x0C},
		{0x0C, 1, 0, 0x0C}, // 男精靈：不換
		{0x0C, 1, 1, 0x06}, // 女精靈：減 6
		{0x0C, 2, 0, 0x06}, // 矮人：減 6
		{0x0C, 3, 1, 0x06}, // 凱爾登：減 6
		{0x1A, 2, 0, 0x14},
	} {
		if got := CanonicalRecord(c.record, c.race, c.gender); got != c.want {
			t.Errorf("CanonicalRecord(%#x, 種族 %d, 性別 %d) = %#x，應為 %#x", c.record, c.race, c.gender, got, c.want)
		}
	}
}

func TestNameMatches(t *testing.T) {
	s := greetSession(t, "", "", [2]int{0x09, 0}, member("A", 0, 0, false))
	elf := member("Elf", 1, 3, false) // 精靈，職業 3 → 記錄 9 Squire
	dwarf := member("Dwarf", 2, 5, false)
	for _, c := range []struct {
		name string
		m    Member
		want bool
	}{
		{"Elven", elf, true}, // 種族
		{"Human", elf, false},
		{"Squire", elf, true}, // 職業 3 + 6 = 9
		{"Darkguard", elf, false},
		{"Dwarven", dwarf, true},
		{"Watchman", dwarf, true}, // 職業 5 + 6 = 11
		{"None", elf, false},      // 找不到 → 0x7F
		{"", elf, false},
		{"Elven", Member{Name: "沒有資料頁"}, false},
	} {
		if got := s.nameMatches(c.name, c.m); got != c.want {
			t.Errorf("nameMatches(%q, %s) = %t，應為 %t", c.name, c.m.Name, got, c.want)
		}
	}
}

// 隊伍不只一人：先問誰要開口，圖示列只剩 DOOR，畫面上沒有人是發言者。
func TestGreetingAsksWhoForParty(t *testing.T) {
	s := greetSession(t, "None", "None", [2]int{0x09, 0},
		member("A", 0, 0, false), member("B", 0, 1, true))
	if !s.AskingWho() {
		t.Fatal("多人隊伍應該停在問誰要開口")
	}
	want := []string{"msg16", `"Which one needs help?"`}
	if !reflect.DeepEqual(dialog(s), want) {
		t.Fatalf("對話是 %q，應為 %q", dialog(s), want)
	}
	wantIcons(t, s, IconDoor)
	for _, m := range s.Inside.Party {
		if m.Speaker {
			t.Fatal("問誰要開口的時候不該有白框")
		}
	}
	// 沒有第三位；DOOR 之外的鍵不改變狀態。
	if s.ChooseSpeaker(2) || !s.AskingWho() {
		t.Fatal("超出隊伍人數的數字不該有作用")
	}
	// 選第二位：換成他，招呼語是他的（女性、沒有法術 → 訊息 12）。
	if !s.Key("2") {
		t.Fatal("數字鍵沒有被接手")
	}
	if s.AskingWho() {
		t.Fatal("選完人不該還在問")
	}
	if !s.Party[1].Speaker || s.Party[0].Speaker {
		t.Fatal("發言者應換成第二位")
	}
	want = []string{"msg12", `"What can I do for you?"`}
	if !reflect.DeepEqual(dialog(s), want) {
		t.Fatalf("招呼語是 %q，應為 %q", dialog(s), want)
	}
	wantIcons(t, s, InteriorIcons([2]int{0x09, 0}, 2)...)
}

// 一人的隊伍不問，直接招呼。
func TestGreetingSoleMemberSkipsPrompt(t *testing.T) {
	s := greetSession(t, "None", "None", [2]int{0x09, 0}, member("A", 0, 0, false))
	if s.AskingWho() {
		t.Fatal("一人隊伍不該問誰要開口")
	}
	if want := []string{"msg10", `"What can I do for you?"`}; !reflect.DeepEqual(dialog(s), want) {
		t.Fatalf("招呼語是 %q，應為 %q", dialog(s), want)
	}
}

// 在問誰要開口時選 DOOR：多人隊伍照常說告別（訊息 17）。
func TestGreetingDoorInPromptSaysFarewell(t *testing.T) {
	s := greetSession(t, "None", "None", [2]int{0x09, 0},
		member("A", 0, 0, false), member("B", 0, 1, true))
	s.Accept()
	if want := []string{"msg17"}; !reflect.DeepEqual(dialog(s), want) {
		t.Fatalf("告別是 %q，應為 %q", dialog(s), want)
	}
	s.Continue()
	if s.Mode != ModeMap {
		t.Fatalf("讀完告別應回到地圖，現在是 %v", s.Mode)
	}
}

// 全隊都被討厭：一句 Get out，讀完不說告別直接離開。
func TestGreetingExpelsWholeParty(t *testing.T) {
	s := greetSession(t, "Elven", "None", [2]int{0x09, 0},
		member("A", 1, 0, false), member("B", 1, 2, true))
	if s.AskingWho() {
		t.Fatal("全隊被討厭時不問誰要開口")
	}
	if want := []string{`"Get out, scum!"`}; !reflect.DeepEqual(dialog(s), want) {
		t.Fatalf("對話是 %q，應為 %q", dialog(s), want)
	}
	wantIcons(t, s, IconSpacebar)
	s.Continue()
	if s.Mode != ModeMap || s.Inside != nil {
		t.Fatal("讀完應直接回到地圖")
	}
}

// 隊伍裡只有一部分人被討厭：照常問；選到被討厭的人是訊息 14，讀完回頭再問。
func TestGreetingRefusesDislikedSpeakerAndAsksAgain(t *testing.T) {
	s := greetSession(t, "Elven", "None", [2]int{0x09, 0},
		member("A", 0, 0, false), member("B", 1, 2, true))
	if !s.AskingWho() {
		t.Fatal("只有部分人被討厭時要問誰要開口")
	}
	s.Key("2")
	if want := []string{"msg14"}; !reflect.DeepEqual(dialog(s), want) {
		t.Fatalf("被拒時是 %q，應為 %q", dialog(s), want)
	}
	wantIcons(t, s, IconSpacebar)
	s.Continue()
	if !s.AskingWho() {
		t.Fatal("多人隊伍被拒之後應回頭問誰要開口")
	}
	s.Key("1")
	if want := []string{"msg10", `"What can I do for you?"`}; !reflect.DeepEqual(dialog(s), want) {
		t.Fatalf("換人之後的招呼語是 %q，應為 %q", dialog(s), want)
	}
}

// 一人的隊伍被討厭：訊息 14 之後直接離開。
func TestGreetingSoleMemberDislikedLeaves(t *testing.T) {
	s := greetSession(t, "Elven", "None", [2]int{0x09, 0}, member("A", 1, 0, false))
	if want := []string{"msg14"}; !reflect.DeepEqual(dialog(s), want) {
		t.Fatalf("被拒時是 %q，應為 %q", dialog(s), want)
	}
	s.Continue()
	if s.Mode != ModeMap {
		t.Fatalf("一人被拒之後應離開，現在是 %v", s.Mode)
	}
}

// 建築欣賞他的職業：訊息 15 取代一般招呼；有傷的話仍是受傷那一句。
func TestGreetingFavoredMessage(t *testing.T) {
	s := greetSession(t, "None", "Squire", [2]int{0x09, 0}, member("A", 0, 3, false))
	if want := []string{"msg15", `"What can I do for you?"`}; !reflect.DeepEqual(dialog(s), want) {
		t.Fatalf("被欣賞時是 %q，應為 %q", dialog(s), want)
	}
}

// 在一般圖示列停著時按數字換發言者，並且重新判斷好惡；換到同一位不做事。
func TestDigitChangesSpeakerInsideBuilding(t *testing.T) {
	s := greetSession(t, "Elven", "Squire", [2]int{0x09, 0},
		member("A", 0, 0, false), member("B", 0, 3, false), member("C", 1, 2, false))
	s.Key("1")
	if s.AskingWho() || !s.interiorIdle() {
		t.Fatal("選完第一位應停在一般圖示列")
	}
	before := append([]string(nil), dialog(s)...)
	s.Key("1")
	if !reflect.DeepEqual(dialog(s), before) {
		t.Fatal("選到同一位不該改變畫面")
	}
	s.Key("2")
	if want := []string{"msg15", `"What can I do for you?"`}; !reflect.DeepEqual(dialog(s), want) {
		t.Fatalf("換成被欣賞的 B：%q", dialog(s))
	}
	s.Key("3")
	if want := []string{"msg14"}; !reflect.DeepEqual(dialog(s), want) {
		t.Fatalf("換成被討厭的 C：%q", dialog(s))
	}
	s.Continue()
	if !s.AskingWho() {
		t.Fatal("被拒之後回頭問誰要開口")
	}
}

// SCROLL 建築（旗標 byte 1 的 bit 0）只招呼自己教派的人：有別的教派就被拒，沒教派的不算。
func TestGreetingRefusesOtherOrder(t *testing.T) {
	here := &Order{Name: "White Pearl", Record: "bba134282050c9808684e91a42e00000"}
	other := &Order{Name: "Blue Gem", Record: "0102030405060708090a0b0c0d0e0f10"}
	a := member("A", 0, 0, false)
	b := member("B", 0, 0, false)
	c := member("C", 0, 0, false)
	a.Order, b.Order = here, other
	s := greetSessionWith(t, func(d *Door) {
		d.IconFlags = [2]int{0x00, 0x01}
		d.Dislikes, d.Likes = "None", "None"
		d.Order = here
	}, a, b, c)
	for _, step := range []struct {
		key  string
		want []string
	}{
		{"1", []string{"msg10", `"What can I do for you?"`}},
		{"2", []string{"msg14"}},
	} {
		s.Key(step.key)
		if !reflect.DeepEqual(dialog(s), step.want) {
			t.Fatalf("按 %s：%q，應為 %q", step.key, dialog(s), step.want)
		}
		if step.key == "2" {
			s.Continue()
		}
	}
	s.Key("3")
	if want := []string{"msg10", `"What can I do for you?"`}; !reflect.DeepEqual(dialog(s), want) {
		t.Fatalf("沒有教派的人：%q", dialog(s))
	}
}

// 點肖像只在原版收的那幾個輸入迴圈有效（RE-167）：戶外走路、問誰開口、建築內一般圖示列。
func TestPortraitClicksState(t *testing.T) {
	s := playSession(t)
	if s.Mode != ModeMap || !s.PortraitClicks() {
		t.Fatal("戶外走路應該收點肖像")
	}
	s.Walk(-1, 0)
	s.Walk(1, 0)
	if s.Mode != ModeDoor || s.PortraitClicks() {
		t.Fatal("門口的圖示列不收點肖像")
	}
	s2 := greetSession(t, "None", "None", [2]int{0x09, 0},
		member("A", 0, 0, false), member("B", 0, 1, true))
	if !s2.AskingWho() || !s2.PortraitClicks() {
		t.Fatal("問誰開口應該收點肖像")
	}
	s2.Key("1")
	if !s2.PortraitClicks() {
		t.Fatal("建築內一般圖示列應該收點肖像")
	}
	s2.StartRest()
	if s2.PortraitClicks() {
		t.Fatal("住宿流程進行中不收點肖像")
	}
}
