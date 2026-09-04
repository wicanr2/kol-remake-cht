package town

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
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

// AskDiskRecord 是 COMPANION 的第一問：`Look for characters on@~H0Dhard disk?`。
const AskDiskRecord = 392

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

// UIRaw 取某個文字庫的一筆記錄並**保留前後空白**，中文模式優先用譯文。
// 原文與譯文用同一套語法，所以走同一套版面碼處理。
//
// 戰鬥的訊息是靠記錄號串接的碎片（RE-088），少了自帶的空白就會黏在一起，
// 所以這裡不能用 Record——那一支會把兩端的空白 trim 掉。
func UIRaw(bank *UIText, tr *Translations, bankName string, index int) string {
	if zh := tr.UIBank().Record(bankName, index); zh != "" {
		return stripKeepSpaces(zh)
	}
	if bank == nil {
		return ""
	}
	return bank.RecordPadded(bankName, index)
}

// UISegments 取一筆記錄並**保留版面碼**，拆成幾段帶座標的文字；中文模式優先用譯文。
//
// 與 UIRaw 的差別只在這裡：UIRaw 是給「串接成一句話」用的，版面碼對它沒有意義所以
// 剝掉；帶框的清單標題（`~H02~V09Items Found:`）反過來，座標就是那一筆記錄的內容，
// 剝掉之後整段會畫到畫面左上角去。
func UISegments(bank *UIText, tr *Translations, bankName string, index int) []Segment {
	if zh := tr.UIBank().Record(bankName, index); zh != "" {
		return ParseSegments(zh)
	}
	if bank == nil {
		return nil
	}
	return bank.Segments(bankName, index)
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

// stripKeepSpaces 拿掉 `~Vxx`／`~Hxx`／`~Cxx` 座標與顏色標記，把 `@` 與 `]`
// 都當成一個空白；前後空白原樣保留。
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
			// `]` 是「接著往下畫」的記號，畫面上留一個空白：原版的
			// `How many]` ＋ `Offensive` 顯示成 `How many Offensive`（RE-076）。
			b.WriteByte(' ')
			i++
		default:
			b.WriteByte(text[i])
			i++
		}
	}
	return b.String()
}

// splitUILines 把一筆記錄依 `@` 拆成幾行並剝掉版面標記。原文與譯文共用。
func splitUILines(raw string) []string {
	if raw == "" {
		return nil
	}
	var out []string
	for _, part := range strings.Split(raw, "@") {
		out = append(out, strip(part))
	}
	return out
}

// Lines 取一筆記錄並依 `@` 拆成幾行；每一行的版面標記照樣剝掉。
// 原版的 `@` 是換行，後面常接一個 `~Hxx` 把游標移回框的左緣。
func (t *UIText) Lines(bank string, index int) []string {
	return splitUILines(t.raw(bank, index))
}

// Segment 是一段帶座標的文字：Col 是字格欄（× 8 就是 x），Row 是像素列。
type Segment struct {
	Col, Row int
	Text     string
}

// Segments 把一筆帶 `~Vxx`／`~Hxx` 座標的記錄拆成幾段。角色資料頁的標籤
// （KOLDATA 記錄 73h）就是這種記錄：每一段自己帶垂直座標，水平座標只在
// 段落開頭出現一次，`@` 之後回到最後一次 `~H` 設的左緣、並往下一行（RE-074）。
func (t *UIText) Segments(bank string, index int) []Segment {
	return parseSegments(t.raw(bank, index))
}

// ParseSegments 讓譯文表也能用同一套座標語法。
func ParseSegments(raw string) []Segment { return parseSegments(raw) }

func parseSegments(raw string) []Segment {
	if raw == "" {
		return nil
	}
	var out []Segment
	var buf strings.Builder
	col, row, margin := 0, 0, 0
	flush := func() {
		if buf.Len() > 0 {
			out = append(out, Segment{Col: col, Row: row, Text: buf.String()})
			buf.Reset()
		}
	}
	for i := 0; i < len(raw); {
		switch {
		case raw[i] == '~' && i+3 < len(raw) &&
			(raw[i+1] == 'V' || raw[i+1] == 'H' || raw[i+1] == 'C'):
			value, err := strconv.ParseInt(raw[i+2:i+4], 16, 32)
			if err == nil {
				switch raw[i+1] {
				case 'V':
					flush()
					row = int(value)
				case 'H':
					flush()
					col, margin = int(value), int(value)
				}
			}
			i += 4
		case raw[i] == '@':
			flush()
			row += LineStep
			col = margin
			i++
		case raw[i] == ']':
			// `]` 是「接著往下畫」的記號，原版把它換成一個空白再印
			// （`sub_1A8FD:0x1A990`）——不是丟掉。記錄 0xE7 的
			// `…AMBUSHED@BY]` 少了它會印成 `BY2`（RE-097）。
			buf.WriteByte(' ')
			i++
		default:
			buf.WriteByte(raw[i])
			i++
		}
	}
	flush()
	return out
}

// LineStep 是原版文字的行距。
const LineStep = 8
