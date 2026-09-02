package town

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// 三份 UI 文字庫（RE-063 的版面）。`sub_1A8FD(記錄號, 文字庫)` 依 0x0D 分隔取第 N 段。
// 標記語法 `~Vxx`／`~Hxx`／`~Cxx`／`@`／`|X` 在抽取時原樣保留，這裡按需要剝掉。

// 文字庫名稱。
const (
	BankUI   = "KOLDATA"
	BankBook = "BOOKDATA"
	BankEnd  = "ENDTXT"
)

// 組出「我聽過某某談起這件事」用到的三筆記錄（RE-067）。
const (
	HeardPrefix   = 0x2F // `"I heard`
	HeardMiddle   = 0x30 // `talking about`
	HeardObject   = 0x9E // that／them／him，加上話題槽的類別
)

type uiRecord struct {
	Index int    `json:"index"`
	Text  string `json:"text"`
}

type uiBank struct {
	Bank    string     `json:"bank"`
	Records []uiRecord `json:"records"`
}

// UIText 是三份 UI 文字庫。
type UIText struct {
	Schema string   `json:"schema"`
	Banks  []uiBank `json:"banks"`
}

// LoadUIText 讀取抽取器產生的 UI 文字庫。
func LoadUIText(path string) (*UIText, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("讀取 UI 文字庫失敗：%w", err)
	}
	var t UIText
	if err := json.Unmarshal(body, &t); err != nil {
		return nil, fmt.Errorf("解析 UI 文字庫失敗：%w", err)
	}
	if !strings.HasPrefix(t.Schema, "kol-ui-text-data-") {
		return nil, fmt.Errorf("非預期的 UI 文字庫 schema：%s", t.Schema)
	}
	return &t, nil
}

// RecordPadded 取一筆記錄但**保留前後空白**。原版有些欄位靠字串自帶的空白置中，
// 例如野外標籤的 `"  DAWN ]"`，去掉空白就會往左移兩個字。
func (t *UIText) RecordPadded(bank string, index int) string {
	return stripKeepSpaces(t.raw(bank, index))
}

func (t *UIText) raw(bank string, index int) string {
	for i := range t.Banks {
		if t.Banks[i].Bank != bank {
			continue
		}
		for _, r := range t.Banks[i].Records {
			if r.Index == index {
				return r.Text
			}
		}
	}
	return ""
}

// Record 取一筆記錄，並把版面標記剝掉——標記只影響原版的游標位置，不是內容。
func (t *UIText) Record(bank string, index int) string {
	for i := range t.Banks {
		if t.Banks[i].Bank != bank {
			continue
		}
		for _, r := range t.Banks[i].Records {
			if r.Index == index {
				return strip(r.Text)
			}
		}
	}
	return ""
}

// Heard 組出 MOUTH 沒對上時的那一句：「I heard <某某> talking about <that>.」
func (t *UIText) Heard(who string, category int) string {
	parts := []string{t.Record(BankUI, HeardPrefix), who,
		t.Record(BankUI, HeardMiddle), t.Record(BankUI, HeardObject+category)}
	var out []string
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, " ")
}

// strip 拿掉版面標記並去掉前後空白。
func strip(text string) string {
	return strings.TrimSpace(stripKeepSpaces(text))
}

// stripKeepSpaces 拿掉 `~Vxx`／`~Hxx`／`~Cxx` 座標與顏色標記，把 `@` 當空白，
// 並丟掉 `]` 這個原版用來標示接續的記號；前後空白原樣保留。
func stripKeepSpaces(text string) string {
	var b strings.Builder
	for i := 0; i < len(text); {
		switch {
		// `~V0D` 這種標記是四個字元：波浪號、V／H／C，再兩位十六進位。
		case text[i] == '~' && i+3 < len(text) &&
			(text[i+1] == 'V' || text[i+1] == 'H' || text[i+1] == 'C'):
			i += 4
		case text[i] == '@':
			b.WriteByte(' ')
			i++
		case text[i] == ']':
			i++
		default:
			b.WriteByte(text[i])
			i++
		}
	}
	return b.String()
}
