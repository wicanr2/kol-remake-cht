// Package i18n 載入 remake 的中文 UI 字串。
//
// 字串表與術語表同一套狀態：核准／暫譯／待查。翻譯不寫死在程式裡，因為術語還在
// 審查中；查不到的字串會原樣回傳英文，寧可漏譯也不要猜。
package i18n

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/wicanr2/kol-remake-cht/internal/magic"
)

type Entry struct {
	En     string `json:"en"`
	Zh     string `json:"zh"`
	Status string `json:"status"`
	Note   string `json:"note"`
}

type table struct {
	Schema      string  `json:"schema"`
	Language    string  `json:"language"`
	Labels      []Entry `json:"labels"`
	Races       []Entry `json:"races"`
	Secondary   []Entry `json:"secondary"`
	Professions []Entry `json:"professions"`
	UI          []Entry `json:"ui"`
}

// Table 是一種語言的 UI 字串查表。
type Table struct {
	Language   string
	entries    map[string]Entry
	names      map[string]string
	upperNames map[string]string
	paragraphs map[string]string
	spellWords *magic.WordTable
	spellSub   func(class, sub int) string
}

// Load 讀取字串表。
func Load(path string) (*Table, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("讀取 UI 字串表失敗：%w", err)
	}
	var t table
	if err := json.Unmarshal(raw, &t); err != nil {
		return nil, fmt.Errorf("解析 UI 字串表失敗：%w", err)
	}
	if t.Schema != "kol-ui-i18n-v1" {
		return nil, fmt.Errorf("非預期的 UI 字串表 schema：%s", t.Schema)
	}
	out := &Table{Language: t.Language, entries: map[string]Entry{}}
	for _, group := range [][]Entry{t.Labels, t.Races, t.Secondary, t.Professions, t.UI} {
		for _, e := range group {
			out.entries[e.En] = e
		}
	}
	if len(out.entries) == 0 {
		return nil, fmt.Errorf("UI 字串表是空的")
	}
	return out, nil
}

// T 回傳譯文；查不到就原樣回傳，不猜。
func (t *Table) T(en string) string {
	if t == nil {
		return en
	}
	if e, ok := t.entries[en]; ok && e.Zh != "" {
		return e.Zh
	}
	return en
}

// LoadDisplayNames 讀取資料名稱的繁中顯示譯名。這張表只由繪製端查詢；
// 原始資料名稱仍用於任務比較、戰鬥狀態與存檔。
func (t *Table) LoadDisplayNames(path string) error {
	if t == nil {
		return fmt.Errorf("字串表是空的")
	}
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("讀取資料名稱譯表失敗：%w", err)
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.Comma = '\t'
	r.FieldsPerRecord = 5
	rows, err := r.ReadAll()
	if err != nil {
		return fmt.Errorf("解析資料名稱譯表失敗：%w", err)
	}
	if len(rows) < 2 || strings.Join(rows[0], "\t") != "key\ten\tzh-Hant\tstatus\tsource" {
		return fmt.Errorf("資料名稱譯表欄位不符或為空")
	}
	names := make(map[string]string, len(rows)-1)
	upperNames := make(map[string]string, len(rows)-1)
	keys := make(map[string]bool, len(rows)-1)
	for i, row := range rows[1:] {
		key, en, zh, status, source := row[0], row[1], row[2], row[3], row[4]
		if key == "" || en == "" || zh == "" || source == "" ||
			(status != "核准" && status != "暫譯") || keys[key] {
			return fmt.Errorf("資料名稱譯表第 %d 行內容無效", i+2)
		}
		keys[key] = true
		if old, ok := names[en]; ok && old != zh {
			return fmt.Errorf("資料名稱 %q 的譯文衝突", en)
		}
		names[en] = zh
		upper := strings.ToUpper(en)
		if old, ok := upperNames[upper]; ok && old != zh {
			return fmt.Errorf("資料名稱 %q 的大寫譯文衝突", en)
		}
		upperNames[upper] = zh
	}
	t.names = names
	t.upperNames = upperNames
	return nil
}

// LoadSpellWords 讀法術音節表（SPEC-034）。subName 把 MAGIC.0 的亞種索引換成
// 英文名，再由 DisplayName 轉成顯示譯名。
func (t *Table) LoadSpellWords(path string, subName func(class, sub int) string) error {
	if t == nil {
		return fmt.Errorf("字串表是空的")
	}
	w, err := magic.LoadWordTable(path)
	if err != nil {
		return err
	}
	t.spellWords, t.spellSub = w, subName
	return nil
}

// SpellGloss 是法術名的中文注解；名字不是可解析的精靈語音節串就回空字串。
// 法術名本身不翻譯：它就是法術構成規則，玩家在編輯器輸入的也是這些詞。
func (t *Table) SpellGloss(name string) string {
	if t == nil || t.spellWords == nil {
		return ""
	}
	w, ok := t.spellWords.Parse(name)
	if !ok {
		return ""
	}
	sub := ""
	if w.Sub > 0 && t.spellSub != nil {
		if en := t.spellSub(w.Class, w.Sub); en != "" {
			sub = t.DisplayName(strings.ToUpper(en))
		}
	}
	return w.Gloss(sub)
}

// DisplayName 僅翻譯完整的資料原名；玩家自取的姓名與法術名不在表內。
func (t *Table) DisplayName(en string) string {
	if t == nil {
		return en
	}
	if zh, ok := t.names[en]; ok {
		return zh
	}
	if zh, ok := t.upperNames[en]; ok {
		return zh
	}
	return en
}

// Missing 回傳給定字串裡沒有譯文的那些，供交付前檢查。
func (t *Table) Missing(all []string) []string {
	var out []string
	for _, en := range all {
		if e, ok := t.entries[en]; !ok || e.Zh == "" {
			out = append(out, en)
		}
	}
	return out
}

// Entries 回傳全部條目，供狀態統計使用。
func (t *Table) Entries() map[string]Entry { return t.entries }

// uiBanks 是 `docs/i18n/ui-zh-Hant.json` 的版面：三個文字庫，記錄號當鍵。
type uiBanks struct {
	Schema string `json:"schema"`
	Banks  map[string]map[string]struct {
		En string `json:"en"`
		Zh string `json:"zh"`
	} `json:"banks"`
}

// markup 把 `~Vxx`／`~Hxx`／`~Cxx` 這類定位碼剝掉。原版用它們在畫面上跳位置，
// remake 的書本自己排版，比對與顯示都不需要（RE-063）。
var markup = regexp.MustCompile(`~[VHCvhc][0-9A-Fa-f]{2}`)

// LoadBank 把一份 UI 文字庫的譯文疊到現有的表上：以**英文段**當鍵。
//
// 書本畫面的標籤是從原版畫面量測出來的（RE-043），文字庫則是逐筆翻好的
// （486 筆全過），兩邊本來是各自為政的兩份譯文——`NAME` 在一邊是「名字」、
// 在另一邊是「姓名」。這一支把文字庫那一份接上來：剝掉定位碼、按 `@` 切段、
// 去掉尾端的冒號與空白，英文段就是書本標籤的那個字串。
//
// **中英段數不同的記錄整筆跳過**：對不上就寧可漏譯，不猜哪一段配哪一段。
func (t *Table) LoadBank(path, bank string) (int, error) {
	if t == nil {
		return 0, fmt.Errorf("字串表是空的")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("讀取 UI 文字庫譯文失敗：%w", err)
	}
	var doc uiBanks
	if err := json.Unmarshal(raw, &doc); err != nil {
		return 0, fmt.Errorf("解析 UI 文字庫譯文失敗：%w", err)
	}
	records, ok := doc.Banks[bank]
	if !ok {
		return 0, fmt.Errorf("譯文裡沒有文字庫 %s", bank)
	}
	added := 0
	for _, rec := range records {
		en := strings.Split(markup.ReplaceAllString(rec.En, ""), "@")
		zh := strings.Split(markup.ReplaceAllString(rec.Zh, ""), "@")
		if rec.Zh == "" || len(en) != len(zh) {
			continue
		}
		for i := range en {
			key := strings.TrimRight(strings.TrimSpace(en[i]), ": ")
			key = strings.TrimSpace(key)
			value := strings.TrimSpace(zh[i])
			if key == "" || value == "" {
				continue
			}
			// 譯文那一段同樣要去掉尾端的冒號，再把補齊寬度用的全形空白收掉——
			// 原版英文靠半形空白把冒號對齊，譯文照做，而 remake 的書本是把
			// 標籤與冒號各自定位的，那些空白不能跟著印出去。
			value = strings.TrimSpace(strings.TrimRight(value, "：:"))
			if value == "" {
				continue
			}
			t.entries[key] = Entry{En: key, Zh: value, Status: "核准",
				Note: bank + " 文字庫"}
			added++
		}
	}
	return added, nil
}
