package i18n

import "testing"

// 文字庫的譯文疊上來之後，書本標籤走的是文字庫那一份（RE-063）。
func TestLoadBankOverridesLabels(t *testing.T) {
	table, err := Load("../../docs/i18n/newchar-zh-Hant.json")
	if err != nil {
		t.Fatal(err)
	}
	before := table.T("NAME")
	n, err := table.LoadBank("../../docs/i18n/ui-zh-Hant.json", "BOOKDATA")
	if err != nil {
		t.Fatal(err)
	}
	if n < 50 {
		t.Fatalf("只疊上 %d 段，太少", n)
	}
	for _, want := range [][2]string{
		{"NAME", "姓名"}, {"RACE", "種族"}, {"SEX", "性別"}, {"CLASS", "職業"},
		{"STRENGTH", "力量"}, {"BODY PTS.", "身體點數"},
		{"HUMAN", "人類"}, {"MALE", "男"}, {"BARBARIAN", "野蠻人"},
	} {
		if got := table.T(want[0]); got != want[1] {
			t.Fatalf("%q 譯成 %q，文字庫那一份是 %q", want[0], got, want[1])
		}
	}
	if before == table.T("NAME") {
		t.Fatalf("NAME 應該從 %q 換成文字庫那一份", before)
	}
	// 文字庫沒有的字串仍然走原本那一份。
	if table.T("CANCEL") == "CANCEL" {
		t.Fatal("remake 自己的字串不該被清掉")
	}
}
