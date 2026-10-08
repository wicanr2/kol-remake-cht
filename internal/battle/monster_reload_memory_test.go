package battle

import (
	"encoding/hex"
	"encoding/json"
	"github.com/wicanr2/kol-remake-cht/internal/rng"
	"os"
	"testing"
)

func TestRE269HeadSourceLifetime(t *testing.T) {
	m := &MonsterReloadMemory{}
	if v, k := m.value(); !k || v != 0 {
		t.Fatalf("startup %d %v", v, k)
	}
	if _, k := (*MonsterReloadMemory)(nil).value(); k {
		t.Fatal("unbound cache inferred")
	}
	for _, source := range []byte{0, 6, 10, 11, 33, 255} {
		for _, present := range []bool{false, true} {
			raw := make([]byte, 16)
			raw[3] = source
			a := &Actor{Side: SideMonster}
			if present {
				raw[0] = 1
				a.Armour[0] = &Item{Encumbrance: 1, Record: hex.EncodeToString(raw)}
			}
			gear := MonsterEquipment{Records: make([]string, 6)}
			gear.Records[1] = hex.EncodeToString(raw)
			m.capture(a, gear)
			s := &State{Actors: make([]Actor, MaxActors), MonsterReloadMemory: m}
			s.Actors[7] = Actor{Side: SideMonster, Kind: 2, Command: CmdMissile, Flag1D: 1}
			if v, k := s.monsterReloadByte(7); !k || v != int(source) {
				t.Fatal("copied head source lost")
			}
			s.consumeShotAmmunition(7)
			want := CmdLoad
			if source == 6 || source == 10 || source == 11 {
				want = CmdMissile
			}
			if s.Actors[7].Command != want || s.Actors[7].Kind != 1 {
				t.Fatal("full-byte reload predicate lost")
			}
			// A later small battle does not overwrite this row.
			next := &State{Actors: make([]Actor, MaxActors), MonsterReloadMemory: m}
			next.Actors[7] = Actor{Side: SideMonster}
			if v, k := next.monsterReloadByte(7); !k || v != int(source) {
				t.Fatal("cross-battle value lost")
			}
			if present {
				a.breakArmour(0)
			}
			wantValue := int(source)
			if present {
				wantValue = 0
			}
			if v, k := next.monsterReloadByte(7); !k || v != wantValue {
				t.Fatal("head clearing or empty padding lost")
			}
			if _, k := next.monsterReloadByte(8); k {
				t.Fatal("third alias inferred")
			}
		}
	}
	m.capture(&Actor{}, MonsterEquipment{})
	if _, k := m.value(); k {
		t.Fatal("malformed record inferred")
	}
}

func TestRE269V2HistoryAgainstOriginal(t *testing.T) {
	path := os.Getenv("KOL_RE269_LIFETIME")
	if path == "" {
		t.Skip("隔離原版收據由KOL_RE269_LIFETIME指定")
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var original struct {
		Rows []struct {
			Monster  int    `json:"monster"`
			Variants []int  `json:"variants"`
			Seed     uint32 `json:"seed_after_seven"`
			Head     string `json:"head_hex"`
			Source   int    `json:"source_byte"`
		} `json:"copy_rows"`
	}
	if err = json.Unmarshal(body, &original); err != nil {
		t.Fatal(err)
	}
	if len(original.Rows) != 32 {
		t.Fatal("原版母體不完整")
	}
	monsters, err := LoadMonsters("../../assets/original-derived-data/monsters.json")
	if err != nil {
		t.Fatal(err)
	}
	m := &MonsterReloadMemory{}
	for _, row := range original.Rows {
		mon := monsters.At(row.Monster)
		makeState := func(count int) *State {
			s := NewQuest(&Site{Index: -1, Level: 1, Count: count, MonsterX: make([]int, count), MonsterY: make([]int, count)}, nil, mon, nil)
			s.MonsterReloadMemory = m
			return s
		}
		s := makeState(7)
		random := rng.NewState(1)
		s.initializeMonsterEquipment(random)
		if random.State() != row.Seed {
			t.Fatal("複製骰序不符")
		}
		for i, v := range row.Variants {
			if s.Actors[6+i].EquipmentVariant != v {
				t.Fatal("選定裝備不符")
			}
		}
		if v, k := m.value(); !k || v != row.Source {
			t.Fatalf("monster%d source=%d/%v expected%d", row.Monster, v, k, row.Source)
		}
		next := makeState(3)
		next.initializeMonsterEquipment(random)
		if v, k := m.value(); !k || v != row.Source {
			t.Fatal("跨場記錄被覆寫")
		}
	}
	if v, k := (&MonsterReloadMemory{}).value(); !k || v != 0 {
		t.Fatal("新程序初值錯誤")
	}
}

func TestRE269FirstDestroyedHead(t *testing.T) {
	s := &State{Actors: make([]Actor, MaxActors), Monster: Monster{EquipmentVariants: make([]MonsterEquipment, 4)}}
	raw := make([]byte, 16)
	raw[0] = 1
	raw[3] = 6
	s.Monster.EquipmentVariants[0].Records = make([]string, 6)
	s.Monster.EquipmentVariants[0].Records[1] = hex.EncodeToString(raw)
	s.Actors[6] = Actor{Side: SideMonster, EquipmentVariant: 0, Armour: [3]*Item{{Record: hex.EncodeToString(raw)}}}
	if v, k := s.monsterReloadByte(6); !k || v != 6 {
		t.Fatal("intact head changed")
	}
	s.Actors[6].breakArmour(0)
	if v, k := s.monsterReloadByte(6); !k || v != 0 {
		t.Fatal("destroyed head source not cleared")
	}
}
