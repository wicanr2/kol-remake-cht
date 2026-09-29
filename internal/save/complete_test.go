package save

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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

// v2 完整存檔讀進來補上自訂圖像兩欄（null），當成 v3；寫出是 v3，自訂圖像能往返（SPEC-036）。
func TestCompleteV2MigratesAndPictureRoundTrips(t *testing.T) {
	s := New()
	s.Schema = CompleteSchema
	s.Characters = characterSample()
	s.Clock = &Clock{Ticks: 840, Part: 4, Day: 27, Month: 1, Year: 3, Weather: 4}
	slot := 0
	for slot < len(s.Characters.Characters) && s.Characters.Characters[slot] == nil {
		slot++
	}
	raw, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err = json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	fields["schema"], _ = json.Marshal(CompleteSchemaV2)
	var slots []map[string]json.RawMessage
	if err = json.Unmarshal(fields["characters"], &slots); err != nil {
		t.Fatal(err)
	}
	for _, c := range slots {
		if c != nil {
			delete(c, "picture")
			delete(c, "shield")
		}
	}
	fields["characters"], _ = json.Marshal(slots)
	v2, _ := json.Marshal(fields)
	got, err := decodeSave(v2)
	if err != nil {
		t.Fatalf("v2 應讀得進來：%v", err)
	}
	if got.Schema != CompleteSchema || got.Characters.Characters[slot].Picture != nil {
		t.Fatalf("遷移後 schema %q、圖像 %v", got.Schema, got.Characters.Characters[slot].Picture)
	}
	pic := strings.Repeat("3a", BitmapBytes)
	s.Characters.Characters[slot].Picture = &pic
	path := filepath.Join(t.TempDir(), "save.json")
	if err = Write(path, s); err != nil {
		t.Fatal(err)
	}
	back, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if back.Characters.Characters[slot].Picture == nil || *back.Characters.Characters[slot].Picture != pic {
		t.Fatal("自訂圖像沒有往返")
	}
	short := "00"
	s.Characters.Characters[slot].Shield = &short
	if err = Write(path, s); err == nil {
		t.Fatal("長度不是 760 bytes 的自訂圖像應寫不出去")
	}
}
