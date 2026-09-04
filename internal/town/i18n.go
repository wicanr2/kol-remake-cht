package town

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// 城鎮建築的譯文與斷行。原文的斷行是原版 24 欄的結果，中文模式一行只有 22 個
// 16×18 字格，因此譯文自己帶斷行，不在執行期做斷字。

// DoorTranslation 是一個門的譯文。Messages 以訊息編號為鍵，放建築內對話的譯文；
// 進門敘述（訊息 0 與 1）合起來放在 Entry，不重複放進 Messages。
type DoorTranslation struct {
	Name     string              `json:"name"`
	Person   string              `json:"person"`
	Entry    []string            `json:"entry"`
	Messages map[string][]string `json:"messages"`
}

type townTranslation struct {
	Name  string                     `json:"name"`
	Doors map[string]DoorTranslation `json:"doors"`
}

// Translations 是城鎮文字的譯文表。
type Translations struct {
	Schema string                     `json:"schema"`
	Status string                     `json:"status"`
	Towns  map[string]townTranslation `json:"towns"`
	// UI 是三個 UI 文字庫的譯文，由 LoadUITranslations 另外載入
	// （`docs/i18n/ui-zh-Hant.json`），不放在城鎮譯文檔裡。
	UI *UITranslations
}

// UITranslations 是 KOLDATA／BOOKDATA／ENDTXT 三個文字庫的譯文，以 bank 名與
// 記錄號為鍵。記錄號是原版的識別，譯文改了字也對得回原文。
type UITranslations struct {
	Schema string `json:"schema"`
	Banks  map[string]map[string]uiRow `json:"banks"`
}

// uiRow 是一筆記錄的原文與譯文。原文一起存著，好在原文變動時看得出譯文過期了
// （`tools/verify_ui_translation.py` 會比對）。
type uiRow struct {
	En   string `json:"en"`
	Zh   string `json:"zh"`
	Note string `json:"note"`
}

// LoadUITranslations 讀取 UI 文字庫的譯文表。找不到檔案不是錯誤——沒有譯文就顯示原文。
func LoadUITranslations(path string) (*UITranslations, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &UITranslations{}, nil
		}
		return nil, fmt.Errorf("讀取 UI 譯文失敗：%w", err)
	}
	var t UITranslations
	if err := json.Unmarshal(body, &t); err != nil {
		return nil, fmt.Errorf("解析 UI 譯文失敗：%w", err)
	}
	if !strings.HasPrefix(t.Schema, "kol-ui-text-i18n-") {
		return nil, fmt.Errorf("非預期的 UI 譯文 schema：%s", t.Schema)
	}
	return &t, nil
}

// NewUITranslations 由「記錄號 → 譯文」的對照建一份 KOLDATA 譯文表，給測試用。
func NewUITranslations(koldata map[string]string) *UITranslations {
	t := &UITranslations{Banks: map[string]map[string]uiRow{BankUI: {}}}
	for index, zh := range koldata {
		t.Banks[BankUI][index] = uiRow{Zh: zh}
	}
	return t
}

// Record 回傳某個文字庫某一筆記錄的譯文；沒有就回空字串。
func (t *UITranslations) Record(bank string, index int) string {
	if t == nil || t.Banks == nil {
		return ""
	}
	return t.Banks[bank][strconv.Itoa(index)].Zh
}

// LoadTranslations 讀取譯文表。找不到檔案不是錯誤——沒有譯文就顯示原文。
func LoadTranslations(path string) (*Translations, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Translations{Towns: map[string]townTranslation{}}, nil
		}
		return nil, fmt.Errorf("讀取城鎮譯文失敗：%w", err)
	}
	var t Translations
	if err := json.Unmarshal(body, &t); err != nil {
		return nil, fmt.Errorf("解析城鎮譯文失敗：%w", err)
	}
	if !strings.HasPrefix(t.Schema, "kol-town-i18n-") {
		return nil, fmt.Errorf("非預期的城鎮譯文 schema：%s", t.Schema)
	}
	return &t, nil
}

// Door 回傳指定城鎮、指定門的譯文；沒有就回 nil。
func (t *Translations) Door(town, door int) *DoorTranslation {
	if t == nil {
		return nil
	}
	tw, ok := t.Towns[strconv.Itoa(town)]
	if !ok {
		return nil
	}
	d, ok := tw.Doors[strconv.Itoa(door)]
	if !ok {
		return nil
	}
	return &d
}

// Message 回傳指定城鎮、指定門、指定編號的訊息譯文；沒有就回 nil。
func (t *Translations) Message(town, door, index int) []string {
	d := t.Door(town, door)
	if d == nil || d.Messages == nil {
		return nil
	}
	return d.Messages[strconv.Itoa(index)]
}

// UIRecord 回傳 KOLDATA 某一筆記錄的譯文；沒有就回空字串。
func (t *Translations) UIRecord(index int) string {
	if t == nil {
		return ""
	}
	return t.UI.Record(BankUI, index)
}

// TownName 回傳城鎮的譯名；沒有就回空字串。
func (t *Translations) TownName(town int) string {
	if t == nil {
		return ""
	}
	return t.Towns[strconv.Itoa(town)].Name
}

// Apply 把譯文套進進門畫面：標題框換成譯名，對話框換成譯文的敘述。
// 沒有譯文的欄位保留原文，畫面因此不會出現空白。
func (in *Inn) Apply(tr *DoorTranslation) {
	if tr == nil {
		return
	}
	if tr.Name != "" || tr.Person != "" {
		title := make([]string, len(in.Title))
		copy(title, in.Title)
		if tr.Name != "" && len(title) > 0 {
			title[0] = tr.Name
		}
		if tr.Person != "" && len(title) > 1 {
			title[1] = tr.Person
		}
		in.Title = title
	}
	if len(tr.Entry) > 0 {
		in.DialogZh = tr.Entry
	}
}
