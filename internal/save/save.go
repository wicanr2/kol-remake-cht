// Package save 是 remake 自有的存檔格式（SPEC-006）。
//
// remake 不回寫原版檔案：`CHARDATA`、`CHARPICS`、`SHLDPICS`、`STEEDPIC` 一律唯讀，
// 而且 runtime 也不解析它們——把原版角色搬進來是離線工具
// `tools/import_original_characters.py` 的事。這裡只讀寫下面這個 JSON。
package save

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Schema 是目前的存檔格式版本。讀取時只接受同一個大版本，不做寬鬆讀取。
const Schema = "kol-save-v1"

// SchemaPrefix 是版本檢查用的字首。
const SchemaPrefix = "kol-save-v"

// MaxParty 是隊伍人數上限。角色儲存區的 16 槽是另一個上限，兩者不可混用。
const MaxParty = 6

// MaxGold 是金幣欄位的上限：原版存成 24-bit 大端（角色記錄 0x2A–0x2C）。
const MaxGold = 1<<24 - 1

// Origin 記下這份存檔怎麼來的。
type Origin struct {
	// Kind 是 "remake"（遊戲自己存的）或 "imported"（由原版角色檔匯入）。
	Kind string `json:"kind"`
	// SourceSHA256 與 Tool 只有匯入的存檔才有。
	SourceSHA256 string `json:"source_sha256,omitempty"`
	Tool         string `json:"tool,omitempty"`
	Note         string `json:"note,omitempty"`
}

// Settings 是 remake 真的會遵守的開關。未知的鍵原樣保留，不靜默丟掉——
// 舊版遊戲讀到新版存檔時不應該把玩家的設定吃掉。
type Settings struct {
	// FaithfulInnStay 開啟時客棧住宿完全照原版：問「存到硬碟嗎」、角色離隊、
	// 隊伍空了問是否離開遊戲。與原版對拍一律在開啟的狀態下進行（SPEC-006）。
	FaithfulInnStay bool
	// Extra 是這一版還不認得的設定鍵。
	Extra map[string]json.RawMessage
}

const settingFaithfulInnStay = "faithful_inn_stay"

// UnmarshalJSON 把認得的鍵拆出來，其餘留在 Extra。
func (s *Settings) UnmarshalJSON(raw []byte) error {
	var all map[string]json.RawMessage
	if err := json.Unmarshal(raw, &all); err != nil {
		return err
	}
	if v, ok := all[settingFaithfulInnStay]; ok {
		if err := json.Unmarshal(v, &s.FaithfulInnStay); err != nil {
			return fmt.Errorf("設定 %s 不是布林值：%w", settingFaithfulInnStay, err)
		}
		delete(all, settingFaithfulInnStay)
	}
	if len(all) > 0 {
		s.Extra = all
	}
	return nil
}

// MarshalJSON 把認得的鍵與 Extra 併回同一個物件。map 的鍵會排序，輸出是穩定的。
func (s Settings) MarshalJSON() ([]byte, error) {
	all := map[string]json.RawMessage{}
	for k, v := range s.Extra {
		all[k] = v
	}
	flag, err := json.Marshal(s.FaithfulInnStay)
	if err != nil {
		return nil, err
	}
	all[settingFaithfulInnStay] = flag
	return json.Marshal(all)
}

// World 是野外的位置。座標與次格座標的語意見 RE-068／RE-069。
// 世界時間還沒有規格，所以格式裡先沒有日期欄位（SPEC-006）。
type World struct {
	X    int `json:"x"`
	Y    int `json:"y"`
	SubX int `json:"sub_x"`
	SubY int `json:"sub_y"`
}

// 位置的兩種形態。
const (
	LocationTown  = "town"
	LocationWorld = "world"
)

// Location 是存檔當下人在哪裡。
type Location struct {
	Kind string `json:"kind"`
	// 以下四個只有 Kind 是 town 時有意義，對應角色記錄的 0x97／0x13／0x99／0x9C。
	Town       int `json:"town"`
	EntryPoint int `json:"entry_point"`
	X          int `json:"x"`
	Y          int `json:"y"`
}

// Member 是隊伍裡的一名角色。欄位只放有 RE 證據的那些（SPEC-006 的欄位表）。
type Member struct {
	Name     string `json:"name"`
	Portrait int    `json:"portrait"`
	Female   bool   `json:"female"`
	Wounds   int    `json:"wounds"`
	// Gold 是角色自己的錢包：原版沒有共用錢包，花的是當下發言者的錢（RE-073）。
	Gold int `json:"gold"`
	// Place 是角色記錄 0x40–0x49 的所在地欄，10 個大寫字元；招募清單靠它比對建築。
	Place string `json:"place"`
	// TownCode 是三字母城鎮代碼。
	TownCode string `json:"town_code"`
	// BuildingID 是角色留宿的建築編號，來自建築 payload 的 0x3A／0x3B。
	BuildingID  int  `json:"building_id"`
	Renown      int  `json:"renown"`
	Spellcaster bool `json:"spellcaster"`
	Speaker     bool `json:"speaker"`
	// Quests 是角色記錄 0x1E2 起那六個 byte（一個 byte 四個 2-bit 欄位，RE-101）；
	// Progress 是 0x1DD 起那五格已回報的任務數（RE-107）。兩者都會在遊玩中改變，
	// 不寫進存檔的話接了任務、打完任務、回報過的進度重載之後就沒了。
	Quests   []int `json:"quests,omitempty"`
	Progress []int `json:"progress,omitempty"`
	// Carried 是角色記錄裡那十七格「身上帶著的東西」（RE-109），一格 16 個 byte
	// 的物品記錄，這裡存 hex。交任務物品要逐 byte 比，所以存的是原始位元組而不是
	// 名字；交出去那一格會被換掉，因此它也是會變的欄位。
	Carried []string `json:"carried,omitempty"`
	// Rank 是角色記錄 `+0x29` 的帳面位階（RE-074）。競技場打贏會加一（RE-108），
	// 所以它也是會變的。用指標是因為 0 是合法值——沒帶這一欄才照角色檔補。
	Rank *int `json:"rank,omitempty"`
	// AdventurePoints 是角色記錄的冒險點數（RE-096）：打贏一場加一份、訓練花掉
	// 一份（RE-076），所以它也是會變的。同樣用指標，0 是合法值。
	AdventurePoints *int `json:"adventure_points,omitempty"`
	// Steed 是身上那一匹坐騎（RE-119）：原版買下之後抄進角色記錄 `+0x1B0`，
	// 沒買過就是空的。存檔帶著它，重載才知道以舊換新要折抵多少。
	Steed *Steed `json:"steed,omitempty"`
}

// Steed 是一匹坐騎。欄位對應馬廄那一筆記錄裡讀得出來的那幾欄（RE-119）：
// 圖號、名字、標價與移動率。移動率要帶著，重載之後隊伍腳程才不會退回用走的。
type Steed struct {
	Index    int    `json:"index"`
	Name     string `json:"name"`
	Price    int    `json:"price"`
	Picture  int    `json:"picture"`
	Movement int    `json:"movement,omitempty"`
}

// Save 是一份存檔。
type Save struct {
	Schema   string   `json:"schema"`
	SavedAt  string   `json:"saved_at"`
	Origin   Origin   `json:"origin"`
	Settings Settings `json:"settings"`
	World    World    `json:"world"`
	Location Location `json:"location"`
	Party    []Member `json:"party"`
}

// New 建立一份空的存檔骨架。
func New() *Save {
	return &Save{Schema: Schema, Origin: Origin{Kind: "remake"},
		Location: Location{Kind: LocationWorld}}
}

// Validate 檢查一份存檔是否合乎格式。缺漏、超界一律失敗，不補預設值。
func (s *Save) Validate() error {
	if !strings.HasPrefix(s.Schema, SchemaPrefix) {
		return fmt.Errorf("非預期的存檔 schema：%q", s.Schema)
	}
	if s.Schema != Schema {
		return fmt.Errorf("存檔版本 %q 不是這一版的 %q", s.Schema, Schema)
	}
	switch s.Origin.Kind {
	case "remake", "imported":
	default:
		return fmt.Errorf("非預期的存檔來源：%q", s.Origin.Kind)
	}
	switch s.Location.Kind {
	case LocationTown, LocationWorld:
	default:
		return fmt.Errorf("非預期的位置種類：%q", s.Location.Kind)
	}
	if len(s.Party) > MaxParty {
		return fmt.Errorf("隊伍有 %d 人，超過上限 %d", len(s.Party), MaxParty)
	}
	seen := map[string]bool{}
	for _, m := range s.Party {
		if m.Name == "" {
			return fmt.Errorf("隊伍裡有沒有名字的角色")
		}
		if seen[m.Name] {
			return fmt.Errorf("隊伍裡有兩個 %s", m.Name)
		}
		seen[m.Name] = true
		if m.Gold < 0 || m.Gold > MaxGold {
			return fmt.Errorf("%s 的金幣 %d 超出 0…%d", m.Name, m.Gold, MaxGold)
		}
		if m.Wounds < 0 {
			return fmt.Errorf("%s 的傷勢 %d 是負的", m.Name, m.Wounds)
		}
	}
	return nil
}

// Speaker 回傳目前的發言者索引；沒有人被標記就是第一位。隊伍空的時候回 -1。
func (s *Save) Speaker() int {
	if len(s.Party) == 0 {
		return -1
	}
	for i := range s.Party {
		if s.Party[i].Speaker {
			return i
		}
	}
	return 0
}

// Read 讀一份存檔。
func Read(path string) (*Save, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("讀取存檔失敗：%w", err)
	}
	var out Save
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("解析存檔失敗：%w", err)
	}
	if err := out.Validate(); err != nil {
		return nil, err
	}
	return &out, nil
}

// Write 原子寫入：先寫同目錄的暫存檔再 rename。中途失敗不會留下半個存檔，
// 也不會把既有的存檔改成壞的。
func Write(path string, s *Save) error {
	if err := s.Validate(); err != nil {
		return err
	}
	body, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化存檔失敗：%w", err)
	}
	body = append(body, '\n')
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("建立存檔目錄失敗：%w", err)
	}
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("建立暫存檔失敗：%w", err)
	}
	name := tmp.Name()
	if _, err := tmp.Write(body); err != nil {
		tmp.Close()
		os.Remove(name)
		return fmt.Errorf("寫入暫存檔失敗：%w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(name)
		return fmt.Errorf("同步暫存檔失敗：%w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(name)
		return fmt.Errorf("關閉暫存檔失敗：%w", err)
	}
	if err := os.Rename(name, path); err != nil {
		os.Remove(name)
		return fmt.Errorf("換上存檔失敗：%w", err)
	}
	return nil
}

// SettingKeys 回傳這份設定裡所有的鍵，排序過，方便收據列出來。
func (s Settings) SettingKeys() []string {
	keys := []string{settingFaithfulInnStay}
	for k := range s.Extra {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
