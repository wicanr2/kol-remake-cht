package town

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Member 是隊伍中的一名角色。Portrait 是 CHARPICS 圖號；Speaker 為真時外框是白色。
type Member struct {
	Portrait int    `json:"portrait"`
	Name     string `json:"name"`
	Speaker  bool   `json:"speaker"`
	// Female 與 Spellcaster 決定店主的招呼語（RE-067）：角色記錄位移 0x1A 是性別，
	// 0x1A0 非零代表會法術，兩者一起選出訊息 10…13。
	Female      bool `json:"female"`
	Spellcaster bool `json:"spellcaster"`
	// Renown 是聲望等第 0…3。非 0 時招呼語改成訊息 26 + 等第。
	Renown int `json:"renown"`
	// Wounds 是角色記錄位移 0x1B 的傷勢。非 0 時招呼語整段換成 KOLDATA 的
	// 2Bh＋2Ch「你受傷了，該去修道院」；建築本身能治療（旗標 bit6）時換成 72h。
	Wounds int `json:"wounds"`
	// Gold 是角色自己的錢包（角色記錄 0x2A–0x2C 的 24-bit 大端）。原版沒有共用
	// 錢包，住宿與治療花的是當下發言者的錢（RE-073）。
	Gold int `json:"gold,omitempty"`
	// Place 是角色記錄 0x40–0x49 的所在地欄，10 個大寫字元；住宿之後會換成
	// 這一間客棧的名字。TownCode 是三字母城鎮代碼。
	Place    string `json:"place,omitempty"`
	TownCode string `json:"town_code,omitempty"`
	// BuildingID 是留宿建築的編號，來自建築 payload 的 0x3A／0x3B。
	BuildingID int `json:"building_id,omitempty"`
	// Sheet 是 MIRROR 那一頁要用的欄位（RE-074）；沒有就畫不出資料頁。
	Sheet *Sheet `json:"sheet,omitempty"`
	// Quests 是角色記錄 0x1E2 起的六個 byte：一個 byte 塞四個 2-bit 欄位，
	// 對應 24 個任務地點（RE-079）。
	Quests []int `json:"quests,omitempty"`
	// Backpack 是背包裡的東西，上限六格（RE-075）。
	Backpack []Good `json:"backpack,omitempty"`
	// Worn 是穿在身上的東西，一個部位一格；nil 代表那個部位是空的。
	Worn [WornSlots]*Good `json:"worn,omitempty"`
	// Spells 是角色記錄 0x200 起那十六筆法術（RE-100）；空的那一格不收進來。
	Spells []Spell `json:"spells,omitempty"`
	// Progress 是角色記錄 0x1DD 起那五格「已回報的任務數」（RE-107）。
	Progress []int `json:"progress,omitempty"`
	// Carried 是交任務物品時要逐 byte 比的那十七格，索引與 CarriedOffsets 對齊
	// （RE-109）。typed 的 Worn／Backpack 丟掉了原始位元組，比對用不上。
	Carried [][]byte `json:"carried,omitempty"`
}

// TurnInQuest 是這一名走進一棟收任務物品的建築時發生的事（RE-109）。
// 回傳交出去的是第幾格（沒有就是 −1）與有沒有記進「已回報數」。
func (m *Member) TurnInQuest(b *QuestBuilding) (slot int, reported bool) {
	if b == nil {
		return -1, false
	}
	return TurnIn(m.Carried, m.Quests, m.Progress, 0, *b)
}

// Inn 是旅店畫面的完整狀態。
type Inn struct {
	Schema   string   `json:"schema"`
	Party    []Member `json:"party"`
	Portrait struct {
		Archive string `json:"archive"`
		Index   int    `json:"index"`
	} `json:"portrait"`
	// Sheet 非 nil 時整個畫面換成角色資料頁（RE-074）。
	Sheet *SheetPage `json:"sheet,omitempty"`
	// SheetLabels 是 KOLDATA 記錄 73h 拆出來的帶座標標籤。
	SheetLabels []Segment `json:"sheet_labels,omitempty"`
	// Detail 是商品明細面板：畫在對話框裡的帶座標文字（RE-075）。
	Detail []Segment `json:"detail,omitempty"`
	Title []string   `json:"title"`
	Dialog   []string `json:"dialog"`
	DialogZh []string `json:"dialog_zh"`
	// QuestList 是蓋在角色資料頁上的任務一覽；nil 代表沒開（RE-111）。
	QuestList *QuestList `json:"quest_list,omitempty"`
	Icons    []int    `json:"icons"`
	Selected int      `json:"selected"`

	// List 非 nil 時對話框畫的是可選清單而不是台詞（招募同伴用）。
	List *InnList `json:"-"`
}

// LoadInn 讀取旅店畫面的樣本狀態。這份 JSON 只作對拍輸入，不是遊戲資料模型。
func LoadInn(path string) (*Inn, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("讀取旅店畫面樣本失敗：%w", err)
	}
	var in Inn
	if err := json.Unmarshal(raw, &in); err != nil {
		return nil, fmt.Errorf("解析旅店畫面樣本失敗：%w", err)
	}
	if !strings.HasPrefix(in.Schema, "kol-town-inn-") {
		return nil, fmt.Errorf("非預期的旅店畫面 schema：%s", in.Schema)
	}
	if len(in.Icons) != IconBar.Slots {
		return nil, fmt.Errorf("圖示數 %d 不是 %d", len(in.Icons), IconBar.Slots)
	}
	if in.Selected < 0 || in.Selected >= len(in.Icons) {
		return nil, fmt.Errorf("選定圖示 %d 超出範圍", in.Selected)
	}
	return &in, nil
}

// MoveIcon 依原版鍵位移動白框：`>` 往右、`<` 往左，兩端都繞回。
// 收據是 town-icon-move-run1：走到第 6 格再按 `>` 回到第 1 格。
func (in *Inn) MoveIcon(delta int) {
	n := len(in.Icons)
	if n == 0 {
		return
	}
	in.Selected = ((in.Selected+delta)%n + n) % n
}

// PlateName 是名牌上顯示的字串：名字的前 NameChars 個字元。
func (m Member) PlateName() string {
	r := []rune(m.Name)
	if len(r) > PartySlot.NameChars {
		r = r[:PartySlot.NameChars]
	}
	return string(r)
}

// Spell 是一筆法術（角色記錄 0x200 起、每 16 byte 一筆，RE-100）。城鎮這一側只
// 保存與轉交，規則在 `internal/battle`。
type Spell struct {
	Name string `json:"name"`
	// Effect 是效果編號；TurnCode／PowerCode 是持續回合與強度的骰式
	//（PowerCode 的 bit 7 是「要不要先選目標格」）；Guards 是擋招要比的四格。
	Effect    int   `json:"effect"`
	TurnCode  int   `json:"turn_code"`
	PowerCode int   `json:"power_code"`
	Guards    []int `json:"guards,omitempty"`
}
