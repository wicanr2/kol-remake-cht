package magic

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// 法術名是精靈語音節串（說明書印刷頁 72–75）：
//
//	種族 + 屬性 + 強度 + 距離 +（持續時間）+ 亞種
//
// 中文模式保留原名，只用這張表解析出各段意義當注解（SPEC-034）。四個人形種族
// 在說明書要以 `TA` 結尾，遊戲資料卻省略，兩種都接受。

// WordTable 是 `docs/i18n/spell-words-zh-Hant.json`。
type WordTable struct {
	Schema   string                    `json:"schema"`
	Race     []wordRace                `json:"race"`
	Stat     []wordStat                `json:"stat"`
	Severity []wordPart                `json:"severity"`
	Range    []wordPart                `json:"range"`
	Duration []wordPart                `json:"duration"`
	Subclass map[string]map[string]int `json:"subclass"`
}

type wordPart struct {
	Syllable string `json:"syllable"`
	Zh       string `json:"zh"`
}

type wordRace struct {
	Syllable string `json:"syllable"`
	Class    int    `json:"class"`
	Zh       string `json:"zh"`
}

type wordStat struct {
	Syllable string `json:"syllable"`
	Stat     int    `json:"stat"`
	Zh       string `json:"zh"`
}

// Word 是一個法術名解析出的各段。Class／Stat／Sub 與 MAGIC.0 的索引相同；
// Duration 為空表示名字裡沒有持續時間那一段。
type Word struct {
	Class, Stat, Sub              int
	Race, StatZh, Severity, Range string
	Duration                      string
}

// LoadWordTable 讀取音節表。
func LoadWordTable(path string) (*WordTable, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("讀取法術音節表失敗：%w", err)
	}
	var t WordTable
	if err := json.Unmarshal(body, &t); err != nil {
		return nil, fmt.Errorf("解析法術音節表失敗：%w", err)
	}
	if t.Schema != "kol-spell-words-v1" || len(t.Race) != 8 || len(t.Stat) != 8 ||
		len(t.Severity) != 5 || len(t.Range) != 2 || len(t.Duration) != 5 {
		return nil, fmt.Errorf("法術音節表欄位不符")
	}
	return &t, nil
}

// Parse 把名字拆成各段。只有唯一一種拆法時才算成功；玩家自取、拆不開或有歧義
// 的名字回傳 false，呼叫端就不加注解。
func (t *WordTable) Parse(name string) (Word, bool) {
	if t == nil {
		return Word{}, false
	}
	w := strings.ToUpper(strings.TrimSpace(name))
	var found []Word
	for _, r := range t.Race {
		a, ok := strings.CutPrefix(w, r.Syllable)
		if !ok {
			continue
		}
		for _, s := range t.Stat {
			b, ok := strings.CutPrefix(a, s.Syllable)
			if !ok {
				continue
			}
			for _, v := range t.Severity {
				c, ok := strings.CutPrefix(b, v.Syllable)
				if !ok {
					continue
				}
				for _, g := range t.Range {
					d, ok := strings.CutPrefix(c, g.Syllable)
					if !ok {
						continue
					}
					base := Word{Class: r.Class, Stat: s.Stat, Race: r.Zh, StatZh: s.Zh,
						Severity: v.Zh, Range: g.Zh}
					// 沒有持續時間那一段，或有一段。
					tails := []struct{ rest, zh string }{{d, ""}}
					for _, u := range t.Duration {
						if e, ok := strings.CutPrefix(d, u.Syllable); ok {
							tails = append(tails, struct{ rest, zh string }{e, u.Zh})
						}
					}
					for _, tail := range tails {
						cand := base
						cand.Duration = tail.zh
						if r.Class <= 4 {
							if tail.rest == "" || tail.rest == "TA" {
								found = append(found, cand)
							}
							continue
						}
						if sub, ok := t.Subclass[fmt.Sprint(r.Class)][tail.rest]; ok {
							cand.Sub = sub
							found = append(found, cand)
						}
					}
				}
			}
		}
	}
	if len(found) != 1 {
		return Word{}, false
	}
	return found[0], true
}

// Gloss 是注解字串，各段以「・」相接；subName 把亞種換成顯示名稱。
func (w Word) Gloss(subName string) string {
	parts := []string{w.Race}
	if subName != "" {
		parts[0] = subName
	}
	parts = append(parts, w.StatZh, w.Severity, w.Range)
	if w.Duration != "" {
		parts = append(parts, "持續"+w.Duration)
	}
	return strings.Join(parts, "・")
}
