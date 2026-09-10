package save

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCompleteFileRoundTripAndFailedWritePreservesFile(t *testing.T) {
	s := New()
	s.Schema = CompleteSchema
	s.Characters = characterSample()
	s.Clock = &Clock{Ticks: 840, Part: 4, Day: 27, Month: 1, Year: 3, Weather: 4}
	s.Settings.Extra = map[string]json.RawMessage{"future": json.RawMessage(`{"x":1}`)}
	path := filepath.Join(t.TempDir(), "save.json")
	if err := Write(path, s); err != nil {
		t.Fatal(err)
	}
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(s, got) {
		t.Fatalf("完整保存往返不符：%s", original)
	}
	s.Characters.Party = []int{0}
	if err = Write(path, s); err == nil {
		t.Fatal("錯誤參照不應覆寫存檔")
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(original) {
		t.Fatal("失敗破壞舊檔")
	}
	var bad map[string]json.RawMessage
	if err = json.Unmarshal(original, &bad); err != nil {
		t.Fatal(err)
	}
	delete(bad, "clock")
	raw, err := json.Marshal(bad)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = decodeSave(raw); err == nil {
		t.Fatal("不可補回缺少的世界時間")
	}
}

func TestClockRejectsInvalidDateAndPart(t *testing.T) {
	for _, c := range []Clock{{Ticks: 4200, Part: 1, Month: 3}, {Ticks: 4200, Day: 28, Month: 1}, {Ticks: 4201, Month: 3}, {Ticks: 4200, Month: 12}, {Ticks: 4200, Month: 3, Weather: 5}} {
		if err := c.Validate(); err == nil {
			t.Fatalf("應拒絕 %+v", c)
		}
	}
}
