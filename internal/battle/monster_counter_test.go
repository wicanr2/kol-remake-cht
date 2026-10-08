package battle

import (
	"encoding/json"
	"os"
	"testing"
)

func TestMonsterCounterOriginal(t *testing.T) {
	path := os.Getenv("KOL_COUNTER_ORACLE_OUTPUT")
	if path == "" {
		t.Skip("需要固定原版有界計數收據")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var data struct {
		Cases int `json:"cases"`
		Rows  []struct {
			Actor  int `json:"actor"`
			Class  int `json:"alias_class_input"`
			Before int `json:"count_before"`
			After  int `json:"count_after"`
		} `json:"rows"`
	}
	if err = json.Unmarshal(raw, &data); err != nil {
		t.Fatal(err)
	}
	if data.Cases != 15360 || len(data.Rows) != data.Cases {
		t.Fatal("原版組數不符")
	}
	for _, row := range data.Rows {
		a := Actor{Side: SideMonster, Kind: row.Before, Command: CmdMissile, Flag1D: 1, Weapon: Weapon{Kind: row.Class}}
		a.ConsumeAmmunition()
		if a.Kind != row.After {
			t.Fatalf("slot%d class%d count%d: got%d original%d", row.Actor, row.Class, row.Before, a.Kind, row.After)
		}
		if row.Before == 0 && a.Command != 0 {
			t.Fatal("原值零未清除指令")
		}
	}
}

func TestMonsterShotCounterPaths(t *testing.T) {
	for _, wall := range []bool{false, true} {
		s := shootState()
		a := &s.Actors[0]
		a.Side = SideMonster
		a.Kind, a.Command, a.Flag1D, a.Weapon.Kind = 1, CmdMissile, 1, 87
		a.LoadLevel, a.ArmourWeight, a.Weapon.Weight = 1, 201, 120
		a.MonsterRanged = &Item{Encumbrance: 90, Class: 2}
		a.ReloadGear()
		before := a.Encumbrance
		s.Plan.Memo = map[int]Cell{0: {Row: 10, Col: 15}}
		if wall {
			s.Field.Tiles[10*FieldSide+13] = 0x29
		}
		result := s.shoot(0, &fixedRoll{values: []int{50}})
		if wall && len(result.Records) == 0 {
			t.Fatal("沒有走阻擋路徑")
		}
		if !wall && result.Message != MissileNothingRecord {
			t.Fatal("沒有走射空路徑")
		}
		if a.Kind != 0 {
			t.Fatal("怪物實際射擊未扣一次")
		}
		if a.WeaponSpeed != 84 || a.Encumbrance <= before {
			t.Fatal("計數零後未依RE-216切回手持負重")
		}
	}
	s := shootState()
	a := &s.Actors[0]
	a.Side = SideMonster
	a.Kind, a.Command, a.Flag1D, a.Weapon.Kind = 5, CmdLoad, 0, 18
	s.shoot(0, &fixedRoll{values: []int{50}})
	if a.Kind != 5 {
		t.Fatal("裝填不應扣計數")
	}
	empty := Actor{Side: SideEmpty, Kind: 19}
	empty.ConsumeAmmunition()
	if empty.Kind != 19 {
		t.Fatal("空槽不應扣計數")
	}
}
