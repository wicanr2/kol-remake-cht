package i18n

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParagraphCatalogKeepsNamesAndLabelsSeparate(t *testing.T) {
	tab, err := Load("../../docs/i18n/newchar-zh-Hant.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := tab.LoadDisplayNames("../../docs/i18n/display-names-zh-Hant.tsv"); err != nil {
		t.Fatal(err)
	}
	before := tab.T("NAME")
	if err := tab.LoadParagraphs("../../docs/i18n/world-prompts-zh-Hant.tsv"); err != nil {
		t.Fatal(err)
	}
	if got := tab.Paragraph("Will you partake?"); got != "要接下嗎？" {
		t.Fatal(got)
	}
	if tab.T("NAME") != before || tab.DisplayName("Dagger") != "匕首" || tab.T("Will you partake?") != "Will you partake?" {
		t.Fatal("完整段落不得改寫一般標籤或原名查詢")
	}
	var english *Table
	if english.Paragraph("Will you partake?") != "Will you partake?" {
		t.Fatal("英文原值改變")
	}
	if tab.Paragraph("UNLISTED-PARAGRAPH") != "UNLISTED-PARAGRAPH" {
		t.Fatal("未知原文被改變")
	}
}

func TestParagraphRejectsInvalidCatalogAtomically(t *testing.T) {
	tab, err := Load("../../docs/i18n/newchar-zh-Hant.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := tab.LoadParagraphs("../../docs/i18n/world-prompts-zh-Hant.tsv"); err != nil {
		t.Fatal(err)
	}
	header := "key\ten\tzh-Hant\tstatus\tsource\n"
	for _, raw := range []string{
		header + "a\tText\t譯文\t暫譯\tsource\na\tOther\t譯文\t暫譯\tsource\n",
		header + "a\tText\t譯文\t暫譯\tsource\nb\tText\t另一譯文\t暫譯\tsource\n",
		header + "a\tText\t\t暫譯\tsource\n",
		header + "a\tText\t譯文\t待查\tsource\n",
		header + "a\t Text\t譯文\t暫譯\tsource\n",
		header + "a\tText\t譯文\t暫譯\n",
		header + string([]byte{0xff}),
	} {
		p := filepath.Join(t.TempDir(), "invalid.tsv")
		if err := os.WriteFile(p, []byte(raw), 0600); err != nil {
			t.Fatal(err)
		}
		if err := tab.LoadParagraphs(p); err == nil {
			t.Fatal("不合法譯表被接受")
		}
		if got := tab.Paragraph("Will you partake?"); got != "要接下嗎？" {
			t.Fatal("失敗載入污染舊譯表", got)
		}
	}
}
