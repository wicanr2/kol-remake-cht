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
	Female     bool `json:"female"`
	Spellcaster bool `json:"spellcaster"`
	// Renown 是聲望等第 0…3。非 0 時招呼語改成訊息 26 + 等第。
	Renown int `json:"renown"`
	// Wounds 是角色記錄位移 0x1B 的傷勢。非 0 時招呼語整段換成 KOLDATA 的
	// 2Bh＋2Ch「你受傷了，該去修道院」；建築本身能治療（旗標 bit6）時換成 72h。
	Wounds int `json:"wounds"`
}

// Inn 是旅店畫面的完整狀態。
type Inn struct {
	Schema   string   `json:"schema"`
	Party    []Member `json:"party"`
	Portrait struct {
		Archive string `json:"archive"`
		Index   int    `json:"index"`
	} `json:"portrait"`
	Title    []string `json:"title"`
	Dialog   []string `json:"dialog"`
	DialogZh []string `json:"dialog_zh"`
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
