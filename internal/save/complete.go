package save

import (
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
	if c.Part != part {
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
