package save

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
)

// Clock 保存規則時間，不包含系統時鐘或亂數執行序列（RE-077）。
type Clock struct {
	Ticks   int `json:"ticks"`
	Part    int `json:"part"`
	Day     int `json:"day"`
	Month   int `json:"month"`
	Year    int `json:"year"`
	Weather int `json:"weather"`
}

func (c *Clock) Validate() error {
	if c == nil {
		return fmt.Errorf("缺少世界時間")
	}
	days := [12]int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	if c.Month < 0 || c.Month >= 12 || c.Day < 0 || c.Day >= days[c.Month] || c.Year < 0 || c.Ticks < 0 || c.Ticks > 4200 || c.Weather < 0 || c.Weather > 4 {
		return fmt.Errorf("世界時間超界")
	}
	part := 4
	for i, floor := range []int{3367, 2527, 1687, 847, 0} {
		if c.Ticks >= floor {
			part = i
			break
		}
	}
	// RE-151：時段在扣時刻之前更新，合法保存可能仍保留上一時段。
	// 保存原值，讓重載後的正常 UpdatePart 處理尚未發生的變換。
	if c.Part < 0 || c.Part > part {
		return fmt.Errorf("世界時段與剩餘時刻不一致")
	}
	return nil
}

type plainSave Save

func (s Save) MarshalJSON() ([]byte, error) {
	raw, err := json.Marshal(plainSave(s))
	if err != nil || s.Schema != CompleteSchema {
		return raw, err
	}
	if err := s.Characters.Validate(); err != nil {
		return nil, err
	}
	var fields map[string]json.RawMessage
	if err = json.Unmarshal(raw, &fields); err != nil {
		return nil, err
	}
	state, err := json.Marshal(s.Characters)
	if err != nil {
		return nil, err
	}
	var members map[string]json.RawMessage
	if err = json.Unmarshal(state, &members); err != nil {
		return nil, err
	}
	for k, v := range members {
		fields[k] = v
	}
	return json.Marshal(fields)
}

func decodeSave(raw []byte) (*Save, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return nil, err
	}
	var schema string
	if err := json.Unmarshal(fields["schema"], &schema); err != nil {
		return nil, err
	}
	var state *Characters
	var clock *Clock
	if schema == CompleteSchemaV2 {
		migrated, err := migrateV2Characters(fields["characters"])
		if err != nil {
			return nil, err
		}
		fields["characters"] = migrated
		schema = CompleteSchema
		fields["schema"], _ = json.Marshal(schema)
	}
	if schema == CompleteSchema {
		subset := map[string]json.RawMessage{}
		for _, k := range []string{"characters", "party", "speaker"} {
			v, ok := fields[k]
			if !ok {
				return nil, fmt.Errorf("缺少 %s", k)
			}
			subset[k] = v
			delete(fields, k)
		}
		b, err := json.Marshal(subset)
		if err != nil {
			return nil, err
		}
		state, err = DecodeCharacters(b)
		if err != nil {
			return nil, err
		}
		if err = requireShape(fields["clock"], reflect.TypeOf(Clock{}), "clock"); err != nil {
			return nil, err
		}
		if err = json.Unmarshal(fields["clock"], &clock); err != nil {
			return nil, err
		}
		for _, k := range []string{"settings", "world", "location", "origin", "saved_at"} {
			if _, ok := fields[k]; !ok {
				return nil, fmt.Errorf("缺少 %s", k)
			}
		}
		if err = requireShape(fields["world"], reflect.TypeOf(World{}), "world"); err != nil {
			return nil, err
		}
		if err = requireShape(fields["location"], reflect.TypeOf(Location{}), "location"); err != nil {
			return nil, err
		}
		raw, err = json.Marshal(fields)
		if err != nil {
			return nil, err
		}
	}
	var out plainSave
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	s := Save(out)
	s.Characters = state
	if schema == CompleteSchema {
		s.Clock = clock
	}
	return &s, nil
}

// migrateV2Characters 在 v2 每一名角色上補 v3 的兩個欄位（null）。空槽維持 null。
func migrateV2Characters(raw json.RawMessage) (json.RawMessage, error) {
	if raw == nil {
		return nil, fmt.Errorf("缺少 characters")
	}
	var slots []json.RawMessage
	if err := json.Unmarshal(raw, &slots); err != nil {
		return nil, fmt.Errorf("characters：%w", err)
	}
	for i, slot := range slots {
		if bytes.Equal(bytes.TrimSpace(slot), []byte("null")) {
			continue
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(slot, &fields); err != nil {
			return nil, fmt.Errorf("characters[%d]：%w", i, err)
		}
		for _, k := range []string{"picture", "shield"} {
			if _, ok := fields[k]; ok {
				return nil, fmt.Errorf("v2 存檔的 characters[%d] 不該有 %s", i, k)
			}
			fields[k] = json.RawMessage("null")
		}
		out, err := json.Marshal(fields)
		if err != nil {
			return nil, err
		}
		slots[i] = out
	}
	return json.Marshal(slots)
}
