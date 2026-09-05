package town

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// 城鎮居民的話題系統（RE-067）。一棟建築有六個話題槽，命中第 n 槽就顯示訊息 20 + n。
// LISTEN 隨機挑一個還沒講過的非空槽；MOUTH 讓玩家打字，用關鍵字字首比對。

// TopicMessageBase 是話題答覆的第一則訊息編號；第 n 槽對應 TopicMessageBase + n。
const TopicMessageBase = 20

// TopicSlots 是每棟建築的話題槽數。
const TopicSlots = 6

// 招呼語的訊息編號（RE-067）：`sub_11A93:0x12017` 起。
// 聲望等第為 0 時是「訊息 10 ＋ 性別×2 ＋ 會不會法術」，否則是訊息 26 ＋ 等第。
const (
	GreetMessageBase  = 10
	GreetRenownBase   = 26
	GreetPartyMessage = 16
	GreetFarewell     = 17
	MaxRenown         = 3
)

// GreetingRecord 是招呼語後面那一句 `"What can I do for you?"` 在 KOLDATA 的記錄號。
const GreetingRecord = 45

// 發言者身上有傷時的招呼語（RE-072）。WoundedRecord 與它的下一筆組成
// 「"You are hurt! You should go to the abbey."」；能治療的建築改用 WoundedHealRecord。
const (
	WoundedRecord     = 0x2B
	WoundedHealRecord = 0x72
)

// GreetMessage 依發言者選出招呼語的訊息編號。
func GreetMessage(m Member) int {
	if m.Renown > 0 {
		renown := m.Renown
		if renown > MaxRenown {
			renown = MaxRenown
		}
		return GreetRenownBase + renown
	}
	index := GreetMessageBase
	if m.Female {
		index += 2
	}
	if m.Spellcaster {
		index++
	}
	return index
}

// AskMessage／UnknownMessage 是問話的敘述與「我不知道」的答覆。
const (
	AskMessage     = 18
	UnknownMessage = 19
)

// ListenLeadInBase 是 LISTEN 開場白在 KOLDATA 的記錄號；加上話題槽的類別（0…3）。
const ListenLeadInBase = 0x9A

// 六個槽都講完之後的兩句：`"Any other gossip?"` 與 `"Not lately."`。
const (
	ListenExhaustedAsk   = 0x2E
	ListenExhaustedReply = 0x5B
)

// Keyword 是一個話題關鍵字。Prefix 最多四個字元，比對時看玩家輸入的詞是不是以它開頭。
type Keyword struct {
	Prefix   string `json:"prefix"`
	Code     int    `json:"code"`
	Topic    int    `json:"topic"`
	Category int    `json:"category"`
}

// TopicSlot 是建築的一個話題槽。Topic 為 0 代表這一格沒有掛關鍵字，
// 是 LISTEN 會講的傳聞；非 0 就要玩家用 MOUTH 問到那個關鍵字才會講。
type TopicSlot struct {
	Slot     int `json:"slot"`
	Raw      int `json:"raw"`
	Topic    int `json:"topic"`
	Category int `json:"category"`
}

// BuildingTopics 是一棟建築的六個話題槽。
type BuildingTopics struct {
	Door  int         `json:"door"`
	Slots []TopicSlot `json:"slots"`
}

type townTopics struct {
	Town      int              `json:"town"`
	Keywords  []Keyword        `json:"keywords"`
	Names     []string         `json:"names"`
	AskWho    map[string]int   `json:"ask_who"`
	Buildings []BuildingTopics `json:"buildings"`
}

// Topics 是全部城鎮的話題資料。
type Topics struct {
	Schema string       `json:"schema"`
	Towns  []townTopics `json:"towns"`
}

// LoadTopics 讀取抽取器產生的話題 typed data。
func LoadTopics(path string) (*Topics, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("讀取城鎮話題資料失敗：%w", err)
	}
	var t Topics
	if err := json.Unmarshal(body, &t); err != nil {
		return nil, fmt.Errorf("解析城鎮話題資料失敗：%w", err)
	}
	if !strings.HasPrefix(t.Schema, "kol-town-keywords-data-") {
		return nil, fmt.Errorf("非預期的城鎮話題 schema：%s", t.Schema)
	}
	return &t, nil
}

func (t *Topics) town(index int) *townTopics {
	for i := range t.Towns {
		if t.Towns[i].Town == index {
			return &t.Towns[i]
		}
	}
	return nil
}

// Building 回傳指定城鎮、指定門的話題槽；找不到就回 nil。
func (t *Topics) Building(town, door int) *BuildingTopics {
	tt := t.town(town)
	if tt == nil {
		return nil
	}
	for i := range tt.Buildings {
		if tt.Buildings[i].Door == door {
			return &tt.Buildings[i]
		}
	}
	return nil
}

// Lookup 在該城的關鍵字表裡找出玩家這句話命中的話題碼。
// 原版是整句轉小寫後以空白分詞，逐一比對關鍵字**字首**，回傳第一個命中的。
func (t *Topics) Lookup(town int, question string) (Keyword, bool) {
	tt := t.town(town)
	if tt == nil {
		return Keyword{}, false
	}
	words := strings.Fields(strings.ToLower(question))
	for _, k := range tt.Keywords {
		if strings.TrimSpace(k.Prefix) == "" {
			continue
		}
		for _, w := range words {
			if strings.HasPrefix(w, k.Prefix) {
				return k, true
			}
		}
	}
	return Keyword{}, false
}

// AskWho 回傳「這個話題該去問誰」的人名；沒有人知道就回空字串。
// 原版在霍布（城 7）不給這個提示。
func (t *Topics) AskWho(town, topic int) string {
	if town == 7 {
		return ""
	}
	tt := t.town(town)
	if tt == nil {
		return ""
	}
	who, ok := tt.AskWho[fmt.Sprint(topic)]
	if !ok || who < 0 || who >= len(tt.Names) {
		return ""
	}
	return tt.Names[who]
}

// SlotFor 回傳這棟建築哪一個話題槽對應該話題碼；沒有就回 -1。
// 原版是由第 5 槽往第 0 槽掃，所以重複時取編號最大的那一格。
func (b *BuildingTopics) SlotFor(topic int) int {
	for i := len(b.Slots) - 1; i >= 0; i-- {
		if b.Slots[i].Topic != 0 && b.Slots[i].Topic == topic {
			return b.Slots[i].Slot
		}
	}
	return -1
}

// Slot 回傳指定編號的話題槽；超出範圍就回 nil。
func (b *BuildingTopics) Slot(n int) *TopicSlot {
	for i := range b.Slots {
		if b.Slots[i].Slot == n {
			return &b.Slots[i]
		}
	}
	return nil
}
