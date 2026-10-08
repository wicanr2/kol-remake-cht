package i18n

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"
)

// LoadParagraphs 載入完整段落的外部譯文，只供顯示查詢。沿用資料名稱 TSV
// 的五欄，與一般標籤／資料原名分開，避免敘述改寫其他字串的用途。
func (t *Table) LoadParagraphs(path string) error {
	if t == nil {
		return fmt.Errorf("字串表是空的")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("讀取段落譯表失敗：%w", err)
	}
	if !utf8.Valid(raw) {
		return fmt.Errorf("段落譯表不是有效 UTF-8")
	}
	r := csv.NewReader(strings.NewReader(string(raw)))
	r.Comma, r.FieldsPerRecord = '\t', 5
	rows, err := r.ReadAll()
	if err != nil {
		return fmt.Errorf("解析段落譯表失敗：%w", err)
	}
	if len(rows) < 2 || strings.Join(rows[0], "\t") != "key\ten\tzh-Hant\tstatus\tsource" {
		return fmt.Errorf("段落譯表欄位不符或為空")
	}
	values, keys := map[string]string{}, map[string]bool{}
	for i, row := range rows[1:] {
		key, en, zh, status, source := row[0], row[1], row[2], row[3], row[4]
		if key == "" || en == "" || zh == "" || source == "" || keys[key] ||
			(status != "核准" && status != "暫譯") || en != strings.Join(strings.Fields(en), " ") ||
			strings.ContainsAny(zh, "\r\n\t") {
			return fmt.Errorf("段落譯表第 %d 行內容無效", i+2)
		}
		if _, duplicate := values[en]; duplicate {
			return fmt.Errorf("段落原文重複：%q", en)
		}
		keys[key], values[en] = true, zh
	}
	t.paragraphs = values
	return nil
}

// Paragraph 只翻譯完整的顯示段落，查不到時沿用現有一般 UI 查表。
func (t *Table) Paragraph(en string) string {
	if t == nil {
		return en
	}
	if zh, ok := t.paragraphs[en]; ok {
		return zh
	}
	return t.T(en)
}
