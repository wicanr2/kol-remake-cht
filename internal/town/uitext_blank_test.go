package town

import "testing"

// 刻意留空（空譯文但有 note）的記錄要照空字串用；沒有 note 的空譯文仍算沒譯，
// 退回原文。
func TestUIRawHonoursIntentionalBlank(t *testing.T) {
	bank := &UIText{Banks: []uiBank{{Bank: BankUI, Records: []uiRecord{
		{Index: 0x140, Text: "The]"}, {Index: 0x141, Text: "blow]"}}}}}
	tr := &Translations{UI: &UITranslations{Banks: map[string]map[string]uiRow{BankUI: {
		"320": {En: "The]", Zh: "", Note: "中文不加冠詞"},
		"321": {En: "blow]", Zh: ""},
	}}}}
	if got := UIRaw(bank, tr, BankUI, 0x140); got != "" {
		t.Fatalf("刻意留空的記錄退回了原文：%q", got)
	}
	if got := UIRaw(bank, tr, BankUI, 0x141); got == "" {
		t.Fatal("沒有 note 的空譯文應退回原文")
	}
}
