package save

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSettingsReplacePreviousDecode(t *testing.T) {
	var s Settings
	if err := json.Unmarshal([]byte(`{"faithful_inn_stay":true,"future_toggle":42}`), &s); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(`{}`), &s); err != nil {
		t.Fatal(err)
	}
	if s.FaithfulInnStay || len(s.Extra) != 0 {
		t.Fatalf("新設定沿用了先前存檔：%+v", s)
	}
}

func TestReadRejectsInvalidSettings(t *testing.T) {
	for _, raw := range []string{`null`, `[]`, `{"faithful_inn_stay":null}`, `{"faithful_inn_stay":1}`} {
		t.Run(raw, func(t *testing.T) {
			body, err := json.Marshal(sample())
			if err != nil {
				t.Fatal(err)
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(body, &fields); err != nil {
				t.Fatal(err)
			}
			fields["settings"] = json.RawMessage(raw)
			body, err = json.Marshal(fields)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(t.TempDir(), "invalid.json")
			if err := os.WriteFile(path, body, 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := Read(path); err == nil {
				t.Fatal("讀檔接受了非規格設定")
			}
		})
	}
}
