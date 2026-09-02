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

// DoorTranslation 是一個門的譯文。
type DoorTranslation struct {
	Name   string   `json:"name"`
	Person string   `json:"person"`
	Entry  []string `json:"entry"`
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
