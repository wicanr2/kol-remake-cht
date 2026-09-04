package save

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func sample() *Save {
	s := New()
	s.SavedAt = "2026-09-03T12:00:00Z"
	s.Settings.FaithfulInnStay = true
	s.World = World{X: 40, Y: 21, SubX: 3, SubY: 3}
	s.Location = Location{Kind: LocationTown, Town: 7, EntryPoint: 3, X: 57, Y: 8}
	s.Party = []Member{{Name: "Theodoric", Portrait: 47, Gold: 953,
		Place: "HALFWAY HO", TownCode: "MIN", BuildingID: 0x56, Speaker: true}}
	return s
}

// 寫出去再讀回來，內容一模一樣。
func TestRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "slot1.json")
	if err := Write(path, sample()); err != nil {
		t.Fatalf("寫入失敗：%v", err)
	}
	got, err := Read(path)
	if err != nil {
		t.Fatalf("讀取失敗：%v", err)
	}
	if got.Party[0].Gold != 953 || got.Party[0].Place != "HALFWAY HO" {
		t.Fatalf("讀回來的隊員是 %+v", got.Party[0])
	}
	if !got.Settings.FaithfulInnStay {
		t.Fatal("設定沒有讀回來")
	}
	if got.Location.Town != 7 || got.World.X != 40 {
		t.Fatalf("位置讀成 %+v／%+v", got.Location, got.World)
	}
}

// 任務旗標與已回報數要跟著存檔往返。少了這兩欄，接了任務、打完任務、
// 回報過的進度在重載之後就沒了。
func TestQuestProgressRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "slot1.json")
	in := sample()
	in.Party[0].Quests = []int{0x03, 0x00, 0x0C, 0x00, 0x00, 0x00}
	in.Party[0].Progress = []int{5, 0, 0, 0, 0}
	in.Party[0].Carried = []string{"89a0235b0b300000000a000000000000"}
	if err := Write(path, in); err != nil {
		t.Fatalf("寫入失敗：%v", err)
	}
	got, err := Read(path)
	if err != nil {
		t.Fatalf("讀取失敗：%v", err)
	}
	if len(got.Party[0].Quests) != 6 || got.Party[0].Quests[2] != 0x0C {
		t.Fatalf("任務旗標讀成 %v", got.Party[0].Quests)
	}
	if len(got.Party[0].Progress) != 5 || got.Party[0].Progress[0] != 5 {
		t.Fatalf("已回報數讀成 %v", got.Party[0].Progress)
	}
	if len(got.Party[0].Carried) != 1 ||
		got.Party[0].Carried[0] != "89a0235b0b300000000a000000000000" {
		t.Fatalf("身上帶的東西讀成 %v", got.Party[0].Carried)
	}
}

// 沒有這兩欄的舊存檔照樣讀得開，欄位留空交給角色檔補。
func TestQuestProgressOptional(t *testing.T) {
	path := filepath.Join(t.TempDir(), "slot1.json")
	if err := Write(path, sample()); err != nil {
		t.Fatalf("寫入失敗：%v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("讀檔失敗：%v", err)
	}
	if strings.Contains(string(raw), "quests") {
		t.Fatal("沒有任務旗標時不該寫出 quests 欄")
	}
	got, err := Read(path)
	if err != nil {
		t.Fatalf("讀取失敗：%v", err)
	}
	if got.Party[0].Quests != nil || got.Party[0].Progress != nil ||
		got.Party[0].Carried != nil {
		t.Fatalf("空的欄位讀成 %v／%v／%v", got.Party[0].Quests,
			got.Party[0].Progress, got.Party[0].Carried)
	}
}

// 不認得的設定鍵要原樣留著，不能靜默丟掉。
func TestUnknownSettingsSurvive(t *testing.T) {
	path := filepath.Join(t.TempDir(), "slot1.json")
	s := sample()
	s.Settings.Extra = map[string]json.RawMessage{"future_toggle": json.RawMessage("42")}
	if err := Write(path, s); err != nil {
		t.Fatalf("寫入失敗：%v", err)
	}
	got, err := Read(path)
	if err != nil {
		t.Fatalf("讀取失敗：%v", err)
	}
	if string(got.Settings.Extra["future_toggle"]) != "42" {
		t.Fatalf("未知設定沒有留下來：%v", got.Settings.Extra)
	}
	keys := got.Settings.SettingKeys()
	if len(keys) != 2 || keys[0] != "faithful_inn_stay" || keys[1] != "future_toggle" {
		t.Fatalf("設定鍵是 %v", keys)
	}
}

// schema 不符一律失敗，不做寬鬆讀取。
func TestSchemaMismatchFails(t *testing.T) {
	dir := t.TempDir()
	for _, schema := range []string{"", "kol-save-v2", "something-else"} {
		path := filepath.Join(dir, "bad.json")
		body, _ := json.Marshal(map[string]any{"schema": schema})
		if err := os.WriteFile(path, body, 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := Read(path); err == nil {
			t.Fatalf("schema %q 應該被擋下來", schema)
		}
	}
}

// 欄位超界、隊伍超過六人、重名一律失敗，不補預設值。
func TestValidateRejectsOutOfRange(t *testing.T) {
	cases := []struct {
		name  string
		build func(*Save)
	}{
		{"金幣是負的", func(s *Save) { s.Party[0].Gold = -1 }},
		{"金幣超過 24-bit", func(s *Save) { s.Party[0].Gold = MaxGold + 1 }},
		{"傷勢是負的", func(s *Save) { s.Party[0].Wounds = -1 }},
		{"沒有名字", func(s *Save) { s.Party[0].Name = "" }},
		{"位置種類不對", func(s *Save) { s.Location.Kind = "dungeon" }},
		{"來源不對", func(s *Save) { s.Origin.Kind = "guess" }},
		{"隊伍超過上限", func(s *Save) {
			for i := 0; i <= MaxParty; i++ {
				s.Party = append(s.Party, Member{Name: string(rune('A' + i))})
			}
		}},
		{"隊伍裡有兩個同名", func(s *Save) {
			s.Party = append(s.Party, Member{Name: "Theodoric"})
		}},
	}
	for _, c := range cases {
		s := sample()
		c.build(s)
		if err := s.Validate(); err == nil {
			t.Fatalf("%s 應該被擋下來", c.name)
		}
	}
}

// 寫入失敗不留半個檔，也不動既有的存檔。
func TestWriteRejectsInvalidWithoutTouchingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "slot1.json")
	if err := Write(path, sample()); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	bad := sample()
	bad.Party[0].Gold = -5
	if err := Write(path, bad); err == nil {
		t.Fatal("超界的存檔應該寫不進去")
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("寫入失敗卻改動了既有的存檔")
	}
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.Contains(e.Name(), ".tmp-") {
			t.Fatalf("留下了暫存檔 %s", e.Name())
		}
	}
}

// 存檔是人看得懂的 JSON：兩個空格縮排、結尾一個換行。
func TestWriteIsReadableJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "slot1.json")
	if err := Write(path, sample()); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if !strings.HasSuffix(text, "}\n") {
		t.Fatal("結尾應該有一個換行")
	}
	if !strings.Contains(text, "\n  \"schema\": \"kol-save-v1\",") {
		t.Fatalf("縮排不是兩個空格：\n%s", text[:120])
	}
}

// 發言者沒標記時預設是隊長；隊伍空的時候回 -1。
func TestSpeakerFallsBackToLeader(t *testing.T) {
	s := sample()
	s.Party[0].Speaker = false
	if s.Speaker() != 0 {
		t.Fatalf("沒有標記發言者時應該回 0，回的是 %d", s.Speaker())
	}
	s.Party = nil
	if s.Speaker() != -1 {
		t.Fatalf("隊伍空的時候應該回 -1，回的是 %d", s.Speaker())
	}
}
