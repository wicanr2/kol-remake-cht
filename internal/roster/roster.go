// Package roster 讀取由 tools/extract_character_roster.py 抽出的角色槽位 typed data。
//
// 執行期不解析原版 `CHARDATA`；槽位版面與所在地組法都由原版角色清單畫面驗證過。
package roster

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Slot struct {
	Slot     int    `json:"slot"`
	Name     string `json:"name"`
	Place    string `json:"place"`
	Town     string `json:"town"`
	Occupied bool   `json:"occupied"`
	// Portrait 是這一格角色頭像命中的 CHARPICS 模板圖號。存檔的頭像逐像素與
	// 80 張模板比對，16 格全部命中；抽取器對不上就會失敗，不會靜靜留白。
	Portrait int `json:"portrait"`
	// Female 是角色記錄位移 0x1A；Wounds 是位移 0x1B 的傷勢，非 0 時旅店的招呼語
	// 換成「你受傷了，該去修道院」。
	Female bool `json:"female"`
	Wounds int  `json:"wounds"`
}

// Location 是清單右欄顯示的字串：固定長度的所在地、一個空白、三字城鎮代碼。
func (s Slot) Location() string {
	return s.Place + " " + s.Town
}

type Roster struct {
	Schema string `json:"schema"`
	Layout struct {
		SlotCount int `json:"slot_count"`
	} `json:"layout"`
	Slots []Slot `json:"slots"`
}

// Load 讀取角色槽位資料並做基本一致性檢查。
func Load(path string) (*Roster, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("讀取角色槽位失敗：%w", err)
	}
	var r Roster
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, fmt.Errorf("解析角色槽位失敗：%w", err)
	}
	if !strings.HasPrefix(r.Schema, "kol-character-roster-") {
		return nil, fmt.Errorf("非預期的角色槽位 schema：%s", r.Schema)
	}
	if len(r.Slots) != r.Layout.SlotCount {
		return nil, fmt.Errorf("槽位數 %d 與 layout 宣告的 %d 不符",
			len(r.Slots), r.Layout.SlotCount)
	}
	return &r, nil
}

// Occupied 回傳已建立角色的槽位，順序即槽位順序。
func (r *Roster) Occupied() []Slot {
	out := make([]Slot, 0, len(r.Slots))
	for _, s := range r.Slots {
		if s.Occupied {
			out = append(out, s)
		}
	}
	return out
}

// Full 表示 16 個槽位都已使用。原版在滿槽時不進 writer；remake 沿用同一條件。
func (r *Roster) Full() bool { return len(r.Occupied()) >= r.Layout.SlotCount }
