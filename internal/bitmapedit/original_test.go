package bitmapedit

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
)

// 與原版同狀態比較（SPEC-036）：原版 KOLBOOK 編輯器從選角清單第一名進場，送同一串鍵、
// 答 y 存檔（`tools/oracle/book_editor_test.go`，收據 `bitmap-editor.json`）。remake 從
// 原版進場那一刻的工作緩衝起跑同一串鍵，存檔結果要與原版 CHARPICS 那一槽逐 byte 相同。
func TestOriginalEditorSameKeys(t *testing.T) {
	path := os.Getenv("KOL_BITMAP_EDITOR_INPUT")
	if path == "" {
		t.Skip("需要 dosgolem 原版編輯器收據")
	}
	var in struct {
		Schema         string   `json:"schema"`
		Keys           []string `json:"keys"`
		BeforeHex      string   `json:"before_hex"`
		ArrayBeforeHex string   `json:"array_before_hex"`
		AfterHex       string   `json:"after_hex"`
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(body, &in); err != nil {
		t.Fatal(err)
	}
	if in.Schema != "kol-bitmap-editor-original-v1" {
		t.Fatalf("schema %q", in.Schema)
	}
	before, _ := hex.DecodeString(in.BeforeHex)
	arrayBefore, _ := hex.DecodeString(in.ArrayBeforeHex)
	after, _ := hex.DecodeString(in.AfterHex)
	if len(before) != Bytes || len(after) != Bytes || string(before) != string(arrayBefore) {
		t.Fatal("收據的編輯前緩衝與陣列那一槽應相同且為 760 byte")
	}
	keys := map[string]int{
		"Up": KeyUp, "Down": KeyDown, "Left": KeyLeft, "Right": KeyRight,
		"Home": KeyHome, "PgUp": KeyPgUp, "End": KeyEnd, "PgDn": KeyPgDn, "KP5": KeyCenter,
		"Return": KeyEnter, "Escape": KeyEsc,
	}
	var pic Bitmap
	copy(pic[:], before)
	e := New(pic, nil)
	for _, k := range in.Keys {
		code, ok := keys[k]
		if !ok {
			if len(k) != 1 {
				t.Fatalf("不認得的鍵 %q", k)
			}
			code = int(k[0])
		}
		e.Key(code)
	}
	if !e.Done || !e.Save {
		t.Fatalf("按鍵走完應已存檔離開：%+v", e)
	}
	if string(e.Pix[:]) != string(after) {
		diff := 0
		for i := range after {
			if after[i] != e.Pix[i] {
				diff++
			}
		}
		t.Fatalf("存檔結果與原版不同：%d byte 不一樣", diff)
	}
}
