package town

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// 城鎮文字的來源與解碼契約見 RE-062：`TOWNn.0` 每棟建築的 payload 從位移 128 起是
// 16 byte 一格的文字格，一格就是畫面上的一行，四格（64 byte）是一則訊息。
// 抽取器已經把 5-bit 字元解成文字並套上畫面規則，runtime 只讀 JSON。

// Message 是一則四行訊息。Display 是套過畫面規則的版本，Lines 是解碼後的原文。
type Message struct {
	Index   int      `json:"index"`
	Lines   []string `json:"lines"`
	Display []string `json:"display"`
}

// BuildingText 是一棟建築的全部訊息。
type BuildingText struct {
	Door     int       `json:"door"`
	Messages []Message `json:"messages"`
}

type townText struct {
	Town      int            `json:"town"`
	Buildings []BuildingText `json:"buildings"`
}

// Text 是全部城鎮的文字。
type Text struct {
	Schema string     `json:"schema"`
	Towns  []townText `json:"towns"`
}

// LoadText 讀取抽取器產生的城鎮文字。
func LoadText(path string) (*Text, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("讀取城鎮文字失敗：%w", err)
	}
	var t Text
	if err := json.Unmarshal(body, &t); err != nil {
		return nil, fmt.Errorf("解析城鎮文字失敗：%w", err)
	}
	if !strings.HasPrefix(t.Schema, "kol-town-text-data-") {
		return nil, fmt.Errorf("非預期的城鎮文字 schema：%s", t.Schema)
	}
	return &t, nil
}

// Building 回傳指定城鎮、指定門的建築文字；找不到就回 nil。
func (t *Text) Building(town, door int) *BuildingText {
	for i := range t.Towns {
		if t.Towns[i].Town != town {
			continue
		}
		for j := range t.Towns[i].Buildings {
			if t.Towns[i].Buildings[j].Door == door {
				return &t.Towns[i].Buildings[j]
			}
		}
	}
	return nil
}

// Message 回傳指定編號的訊息；沒有就回 nil。
func (b *BuildingText) Message(index int) *Message {
	for i := range b.Messages {
		if b.Messages[i].Index == index {
			return &b.Messages[i]
		}
	}
	return nil
}

// EntryMessages 是進門時顯示的敘述：訊息 0 與 1 合起來八行，正好填滿對話框。
const (
	EntryMessageFirst = 0
	EntryMessageCount = 2
)

// EntryLines 回傳進門敘述的八行（畫面版本）。
func (b *BuildingText) EntryLines() []string {
	var out []string
	for i := 0; i < EntryMessageCount; i++ {
		m := b.Message(EntryMessageFirst + i)
		if m == nil {
			continue
		}
		out = append(out, m.Display...)
	}
	return out
}

// BuildingIcon 是建築內部圖示列唯一那一格的 KOLMPICS 圖號（原版圖示 id 1，畫的是 id + 3）。
const BuildingIcon = 4

// EnterBuilding 依門與建築文字組出進門那一張畫面的狀態。版面與旅店畫面相同：
// 肖像在 (8,64)、標題框放建築名與人物名、對話框放進門敘述。
func EnterBuilding(archive string, door *Door, text *BuildingText, party []Member) *Inn {
	// 進門當下說話的是建築裡的人（畫在肖像格），隊員都不是發言者，外框是亮藍色。
	members := make([]Member, len(party))
	copy(members, party)
	for i := range members {
		members[i].Speaker = false
	}
	in := &Inn{Schema: "kol-town-inn-building-v1", Party: members}
	in.Portrait.Archive = archive
	in.Portrait.Index = door.Portrait
	in.Title = []string{door.Name, door.Person}
	if text != nil {
		in.Dialog = text.EntryLines()
	}
	in.Icons = []int{BuildingIcon}
	in.Selected = 0
	return in
}
