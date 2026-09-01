// Package i18n 載入 remake 的中文 UI 字串。
//
// 字串表與術語表同一套狀態：核准／暫譯／待查。翻譯不寫死在程式裡，因為術語還在
// 審查中；查不到的字串會原樣回傳英文，寧可漏譯也不要猜。
package i18n

import (
	"encoding/json"
	"fmt"
	"os"
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
	Language string
	entries  map[string]Entry
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
