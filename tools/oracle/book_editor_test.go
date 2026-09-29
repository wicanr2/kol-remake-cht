//go:build oracle

package oracle

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
)

// editorKeys 是送進原版點陣圖編輯器的按鍵（SPEC-036）：工具列換色、ESC 切畫布、
// 落筆移動、抬筆、斜走、畫點、吸色、再切回工具列跑四支變換，最後離開答 y 存檔。
// remake 那一側用同一串（`cmd/kol/bitmap_editor_original_test.go`）。
var editorKeys = []string{
	"Right", "Right",
	"Escape",
	"d", "Right", "Right", "Down", "Down", "Home", "u", "PgDn", "x", "End", "KP5", "c", "Left", "x",
	"Escape",
	".", ".", ".", ".", "Return", ".", "Return", ".", "Return", ".", "Return", ".", "Return",
	"y",
}

// TestBookBitmapEditorOriginal 走正常書本路徑：封面 → 目錄頁 5（角色選項）→ 下移三格到
// EDIT PICTURE → 選角清單第一名 → 編輯器，送 editorKeys，記下編輯前的工作緩衝與存檔後
// CHARPICS 陣列那一槽，寫成 `bitmap-editor.json`。
func TestBookBitmapEditorOriginal(t *testing.T) {
	root := os.Getenv("KOL_ORACLE_ROOT")
	out := os.Getenv("KOL_ORACLE_CAPTURE")
	if root == "" || out == "" {
		t.Skip("需要玩家自備原版 KOL_ORACLE_ROOT 與 KOL_ORACLE_CAPTURE")
	}
	o, err := oracle.Load(filepath.Join(root, "KOL.EXE"), root)
	if err != nil {
		t.Fatal(err)
	}
	defer o.Close()
	defer captureBook(t, o, "editor-final.png")
	loaded := false
	for i := 0; i < 20 && !loaded; i++ {
		if i >= 4 {
			o.Type("\r")
		}
		if err = o.Run(50_000_000); err != nil {
			t.Fatal(err)
		}
		for _, name := range o.Opened() {
			if strings.EqualFold(filepath.Base(name), "KOLBOOK.EXE") {
				loaded = true
			}
		}
	}
	if !loaded {
		t.Fatal("沒有載入 KOLBOOK.EXE")
	}
	if err = o.Run(5_000_000); err != nil {
		t.Fatal(err)
	}
	press := func(name string) {
		t.Helper()
		switch {
		case len(name) == 1:
			err = o.TypeKeys(name)
		default:
			err = o.SendKeys(name)
		}
		if err != nil {
			t.Fatalf("送 %q：%v", name, err)
		}
		if err = o.Run(10_000_000); err != nil {
			t.Fatal(err)
		}
		for n := 0; o.KeysPending() > 0 && n < 40; n++ {
			if err = o.Run(10_000_000); err != nil {
				t.Fatal(err)
			}
		}
		if o.KeysPending() > 0 {
			t.Fatalf("%q 一直沒被讀走", name)
		}
		if err = o.Run(20_000_000); err != nil {
			t.Fatal(err)
		}
	}
	for _, k := range []string{"Space", "5", "Down", "Down", "Down", "Return"} {
		press(k)
	}
	if err = o.Run(60_000_000); err != nil {
		t.Fatalf("等選角清單畫完：%v", err)
	}
	captureBook(t, o, "editor-list.png")
	press("Return")
	if err = o.Run(60_000_000); err != nil {
		t.Fatalf("等編輯器畫完：%v", err)
	}
	captureBook(t, o, "editor-enter.png")
	o.SyncDGroupFromDS()
	// KOLBOOK DGROUP 在 IDA linear 0x19240。word_238FC 是選定那一槽在 CHARPICS 陣列裡的
	// 位移（`_main:0x10B26` 乘好 0x2F8），word_19B52 是工作緩衝的段。
	slotOff := o.Word(o.DS(0x238FC - 0x19240))
	workSeg := o.Word(o.DS(0x19B52 - 0x19240))
	before := o.Bytes(oracle.Addr{Seg: workSeg, Off: 0}, 760)
	arrayBefore := o.Bytes(o.DS(0x1844+slotOff), 760)
	for _, k := range editorKeys {
		press(k)
	}
	captureBook(t, o, "editor-after.png")
	after := o.Bytes(o.DS(0x1844+slotOff), 760)
	receipt := map[string]any{
		"schema": "kol-bitmap-editor-original-v1", "input": "KOLBOOK.EXE（正常書本路徑）",
		"evidence":    "SPEC-036；編輯前取工作緩衝 word_19B52:0，存檔後取 CHARPICS 陣列 DS:0x1844 + word_238FC",
		"slot_offset": slotOff, "keys": editorKeys,
		"before_hex": hex.EncodeToString(before), "array_before_hex": hex.EncodeToString(arrayBefore),
		"after_hex": hex.EncodeToString(after),
	}
	body, _ := json.MarshalIndent(receipt, "", "  ")
	if err = os.WriteFile(filepath.Join(out, "bitmap-editor.json"), append(body, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	if string(before) == string(after) {
		t.Fatal("存檔後那一槽沒有變，按鍵沒有生效")
	}
	t.Logf("槽位移 %#x，工作段 %#x；編輯前後不同", slotOff, workSeg)
}
