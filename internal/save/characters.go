package save

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
)

// Characters 是 SPEC-023 的保存核心，**正式遊戲用的就是它**（`CompleteSchema`，
// 也就是 v2）；v1 刻意保留為匯入入口，見 `save.go` 的 `Schema`。
// 槽位不壓縮，隊伍只引用槽位，避免兩份角色資料互相覆蓋。
type Characters struct {
	Characters [16]*Character `json:"characters"`
	Party      []int          `json:"party"`
	Speaker    int            `json:"speaker"`
}

// Character 不含畫面、資產或狀態機。空值也是保存內容，不使用 omitempty。
type Character struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Portrait    int              `json:"portrait"`
	Female      bool             `json:"female"`
	Spellcaster bool             `json:"spellcaster"`
	Renown      int              `json:"renown"`
	Wounds      int              `json:"wounds"`
	Gold        int              `json:"gold"`
	Place       string           `json:"place"`
	TownCode    string           `json:"town_code"`
	BuildingID  int              `json:"building_id"`
	Sheet       *CharacterSheet  `json:"sheet"`
	Quests      [6]int           `json:"quests"`
	Progress    [5]int           `json:"progress"`
	Carried     [17]string       `json:"carried"`
	InHand      *Item            `json:"in_hand"`
	Worn        [10]*Item        `json:"worn"`
	Backpack    []Item           `json:"backpack"`
	Token       *Item            `json:"token"`
	Steed       *Item            `json:"steed"`
	Spells      []CharacterSpell `json:"spells"`
	Order       *CharacterOrder  `json:"order"`
	// Picture／Shield 是在書本 EDIT PICTURE／EDIT SHIELD 改過的圖（SPEC-036）：760 byte
	// 的 40 × 38 4 位元索引，以十六進位字串保存。null 表示沒改過，圖像沿用 Portrait 那一張
	// 模板、盾牌沿用新角色的預設盾牌。v3 新增。
	Picture *string `json:"picture"`
	Shield  *string `json:"shield"`
}

// BitmapBytes 是自訂圖像一筆的長度（原版 CHARPICS／SHLDPICS 的一筆）。
const BitmapBytes = 760

type CharacterSheet struct {
	Ward            int               `json:"ward"`
	Stats           [7]int            `json:"stats"`
	Race            int               `json:"race"`
	Profession      int               `json:"profession"`
	Rank            int               `json:"rank"`
	AdventurePoints int               `json:"adventure_points"`
	MovementRate    int               `json:"movement_rate"`
	Nutrition       int               `json:"nutrition"`
	HeightFeet      int               `json:"height_feet"`
	HeightInches    int               `json:"height_inches"`
	Weight          int               `json:"weight"`
	Weapons         []CharacterWeapon `json:"weapons"`
	Order           string            `json:"order"`
	Title           string            `json:"title"`
	HasItems        bool              `json:"has_items"`
	HasNotes        bool              `json:"has_notes"`
}

type CharacterWeapon struct {
	Row        int    `json:"row"`
	Name       string `json:"name"`
	Weapon     int    `json:"weapon"`
	Offence    int    `json:"offence"`
	Defence    int    `json:"defence"`
	OffenceRaw int    `json:"offence_raw"`
	DefenceRaw int    `json:"defence_raw"`
}

type Item struct {
	Index       int      `json:"index"`
	Name        string   `json:"name"`
	Price       int      `json:"price"`
	Kind        int      `json:"kind"`
	Damage      bool     `json:"damage"`
	Hands       int      `json:"hands"`
	Class       int      `json:"class"`
	Position    int      `json:"position"`
	Encumbrance int      `json:"encumbrance"`
	Condition   int      `json:"condition"`
	Percent     int      `json:"percent"`
	Dice        ItemDice `json:"dice"`
	SizeClass   int      `json:"size_class"`
	Record      string   `json:"record"`
	Picture     int      `json:"picture"`
	Movement    int      `json:"movement"`
}

type ItemDice struct {
	Count int `json:"count"`
	Sides int `json:"sides"`
	Bonus int `json:"bonus"`
	Low   int `json:"low"`
	High  int `json:"high"`
}

type CharacterSpell struct {
	Name      string `json:"name"`
	Effect    int    `json:"effect"`
	TurnCode  int    `json:"turn_code"`
	PowerCode int    `json:"power_code"`
	Guards    []int  `json:"guards"`
}

type CharacterOrder struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Intellect   int    `json:"intellect"`
	Allow       []int  `json:"allow"`
	Record      string `json:"record"`
	Token       string `json:"token"`
	TokenRecord string `json:"token_record"`
}

// DecodeCharacters 先檢查 JSON 形狀，再解碼；encoding/json 單獨使用會把
// 少於定長陣列的輸入補零、忽略過長部分，且容許 null 成為整數零。
func DecodeCharacters(raw []byte) (*Characters, error) {
	if err := requireShape(raw, reflect.TypeOf(Characters{}), "characters"); err != nil {
		return nil, err
	}
	var out Characters
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	if err := out.Validate(); err != nil {
		return nil, err
	}
	return &out, nil
}

func requireShape(raw json.RawMessage, typ reflect.Type, path string) error {
	isNull := bytes.Equal(bytes.TrimSpace(raw), []byte("null"))
	if typ.Kind() == reflect.Pointer {
		if isNull {
			return nil
		}
		return requireShape(raw, typ.Elem(), path)
	}
	if isNull {
		if typ.Kind() == reflect.Slice {
			return nil
		}
		return fmt.Errorf("%s 不可為 null", path)
	}
	switch typ.Kind() {
	case reflect.Struct:
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(raw, &fields); err != nil {
			return fmt.Errorf("%s：%w", path, err)
		}
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			name := field.Tag.Get("json")
			value, ok := fields[name]
			if !ok {
				return fmt.Errorf("%s 缺少 %s", path, name)
			}
			if err := requireShape(value, field.Type, path+"."+name); err != nil {
				return err
			}
			delete(fields, name)
		}
		if len(fields) != 0 {
			return fmt.Errorf("%s 含未知欄位", path)
		}
	case reflect.Array, reflect.Slice:
		var values []json.RawMessage
		if err := json.Unmarshal(raw, &values); err != nil {
			return fmt.Errorf("%s：%w", path, err)
		}
		if typ.Kind() == reflect.Array && len(values) != typ.Len() {
			return fmt.Errorf("%s 長度必須為 %d", path, typ.Len())
		}
		for i, v := range values {
			if err := requireShape(v, typ.Elem(), fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Characters) Validate() error {
	if s == nil {
		return fmt.Errorf("缺少角色保存內容")
	}
	if len(s.Party) > MaxParty {
		return fmt.Errorf("隊伍超過六人")
	}
	if (len(s.Party) == 0 && s.Speaker != -1) || (len(s.Party) > 0 && (s.Speaker < 0 || s.Speaker >= len(s.Party))) {
		return fmt.Errorf("發言者索引錯誤")
	}
	seen := map[int]bool{}
	for _, slot := range s.Party {
		if slot < 0 || slot >= len(s.Characters) || s.Characters[slot] == nil || seen[slot] {
			return fmt.Errorf("隊伍槽位 %d 無效或重複", slot)
		}
		seen[slot] = true
	}
	names := map[string]bool{}
	ids := map[string]bool{}
	for slot, c := range s.Characters {
		if c == nil {
			continue
		}
		if c.ID == "" || ids[c.ID] {
			return fmt.Errorf("角色槽位 %d 身分空白或重複", slot)
		}
		ids[c.ID] = true
		if c.Name == "" || names[c.Name] {
			return fmt.Errorf("角色槽位 %d 名稱空白或重複", slot)
		}
		names[c.Name] = true
		if c.Gold < 0 || c.Gold > MaxGold || c.Wounds < 0 {
			return fmt.Errorf("角色槽位 %d 金冠或傷勢無效", slot)
		}
		if c.Sheet == nil {
			return fmt.Errorf("角色槽位 %d 缺少資料頁", slot)
		}
		if len(c.Backpack) > 6 || len(c.Spells) > 16 || len(c.Sheet.Weapons) > 4 {
			return fmt.Errorf("角色槽位 %d 欄位數量超界", slot)
		}
		rows := map[int]bool{}
		for _, w := range c.Sheet.Weapons {
			if w.Row < 0 || w.Row >= 4 || rows[w.Row] {
				return fmt.Errorf("角色槽位 %d 武器列錯誤", slot)
			}
			rows[w.Row] = true
		}
		for _, v := range append(c.Quests[:], c.Progress[:]...) {
			if v < 0 || v > 255 {
				return fmt.Errorf("角色槽位 %d 任務欄位超界", slot)
			}
		}
		for _, raw := range c.Carried {
			b, err := hex.DecodeString(raw)
			if err != nil || len(b) != 16 {
				return fmt.Errorf("角色槽位 %d 物品記錄必須為 16 bytes", slot)
			}
		}
		for _, bm := range []*string{c.Picture, c.Shield} {
			if bm == nil {
				continue
			}
			b, err := hex.DecodeString(*bm)
			if err != nil || len(b) != BitmapBytes {
				return fmt.Errorf("角色槽位 %d 自訂圖像必須為 %d bytes", slot, BitmapBytes)
			}
		}
	}
	return nil
}
