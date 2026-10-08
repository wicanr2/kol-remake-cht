package battle

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/rng"
)

type spellReactionRand struct {
	calls []int
}

func (r *spellReactionRand) Roll(n int) int {
	r.calls = append(r.calls, n)
	return n
}

func spellReactionState(power int) *State {
	sp := Spell{Effect: SpellEffectArea, PowerCode: power}
	s := castState(sp)
	s.Actors[0].Command, s.Actors[0].Negative = CmdCast, true
	s.Actors[0].X, s.Actors[0].Y, s.Actors[0].Size = 10, 10, 255
	s.Actors[1].X, s.Actors[1].Y, s.Actors[1].Size = 10, 11, 1
	s.Actors[1].Head, s.Actors[1].Chest = 127, 127
	s.Actors[1].RightArm, s.Actors[1].LeftArm = 127, 127
	s.Actors[1].RightLeg, s.Actors[1].LeftLeg = 127, 127
	s.Actors[1].SpellTurns = 5
	s.Actors[1].Command = CmdRest
	s.Plan.Memo = map[int]Cell{0: {Row: 10, Col: 11}}
	s.Exec.Order = []int{0, 1}
	return s
}

func TestSpellReactionRandomAreaOrder(t *testing.T) {
	s := spellReactionState(138)
	rng := &spellReactionRand{}
	out, ok := s.Act(rng)
	if !ok || out.Cast == nil || !reflect.DeepEqual(rng.calls, []int{2, 11, 6}) {
		t.Fatalf("持續、強度、部位骰序：%v，結果 %+v", rng.calls, out)
	}
	if out.Cast.Area != 5 || out.Cast.Power != 11 || s.Actors[1].LeftLeg != 116 || s.Actors[1].SpellTurns != 0 {
		t.Fatal("部位、傷害或法術倒數未依原版更新")
	}
}

func TestSpellReactionBlockedConsumesNoDice(t *testing.T) {
	for _, guard := range []int{0x10, 0x40} {
		s := spellReactionState(138)
		s.Actors[1].Ward = guard
		rng := &spellReactionRand{}
		out, _ := s.Act(rng)
		if out.Message != NoEffectRecord || len(rng.calls) != 0 || s.Actors[1].SpellTurns != 5 {
			t.Fatalf("被擋仍擲骰或改倒數：%v %+v", rng.calls, out)
		}
	}
}

func TestSpellReactionDownDropsHeldItem(t *testing.T) {
	for _, area := range []int{0, 1, 4} {
		s := spellReactionState(156)
		s.Actors[1].Side, s.Actors[1].Member = SideParty, 1
		item := &Item{Name: "test weapon", Record: "b9a1f2a027028644105528dc139c9004"}
		s.Party = append(s.Party, Member{HasWeapon: true, WeaponName: item.Name, HandItem: item})
		s.Actors[1].Weapon = Weapon{Kind: 19}
		if area == 4 {
			s.Actors[1].Threshold = 10
		} else {
			s.Actors[1].Head, s.Actors[1].Chest = 10, 10
		}
		s.Area = &AreaQuestion{Caster: 0, Target: 1, Power: 27, Negative: true}
		res, _, ok := s.AnswerArea(area)
		if !ok || res.Damage == nil || !s.Actors[1].Down() || s.Actors[1].Command != 0 || s.Party[1].HandItem != nil || s.Ground[0].Item != item {
			t.Fatalf("部位%d未取消並掉落完整手持物品", area)
		}
	}
}

func TestSpellReactionWaitsBeforeKnockback(t *testing.T) {
	s := spellReactionState(156)
	s.Exec.Turn = 1
	s.Area = &AreaQuestion{Caster: 0, Target: 1, Power: 27, Negative: true}
	res, _, ok := s.AnswerArea(4)
	if !ok || res.Damage == nil || s.Actors[1].X != 10 || s.Actors[1].Y != 11 || s.Exec.Turn != 1 {
		t.Fatal("傷害頁前已移動或推進下一手")
	}
	rng := &spellReactionRand{}
	out, ok := s.Act(rng)
	if !ok || !reflect.DeepEqual(rng.calls, []int{100, 3}) || !reflect.DeepEqual(out.Records, []int{KnocksRecord, BackwardsRecord}) || !out.Plain || out.Target != 1 || s.Exec.Turn != 1 || s.Actors[1].Command != 0 {
		t.Fatalf("傷害頁後未先結算反應：%v %+v", rng.calls, out)
	}
}

func TestSpellReactionExcludedDamage(t *testing.T) {
	for _, c := range []struct {
		power, part int
		negative    bool
	}{
		{0, 127, true}, {15, 127, true}, {27, 10, true}, {27, 0, true}, {27, 127, false},
	} {
		s := spellReactionState(156)
		s.Actors[1].RightLeg = c.part
		s.Exec.Turn = 1
		s.Area = &AreaQuestion{Caster: 0, Target: 1, Power: c.power, Negative: c.negative}
		s.AnswerArea(4)
		rng := &spellReactionRand{}
		out, _ := s.Act(rng)
		if len(rng.calls) != 0 || len(out.Records) != 0 {
			t.Fatalf("排除案例仍判擊退：%+v %v", c, rng.calls)
		}
	}
}

// 原版局部收據由 dosgolem 重生。比較明示同狀態與固定 seed，不以內部自洽代替原版。
func TestSpellReactionOriginalCases(t *testing.T) {
	path := os.Getenv("KOL_SPELL_REACTION_SOURCE")
	if path == "" {
		t.Skip("需要原版部位法術局部收據")
	}
	var source struct {
		Cases []struct {
			Name           string
			PowerCode      int `json:"power_code"`
			Negative       int
			Blocked, Armed bool
			Initial, Final map[string]json.RawMessage
			Waits          []map[string]json.RawMessage
		}
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(body, &source); err != nil {
		t.Fatal(err)
	}
	if len(source.Cases) != 8 {
		t.Fatal("原版案例形狀不符")
	}
	for _, c := range source.Cases {
		t.Run(c.Name, func(t *testing.T) {
			decode := func(m map[string]json.RawMessage, key string) []byte {
				var value string
				if err := json.Unmarshal(m[key], &value); err != nil {
					t.Fatal(err)
				}
				b, err := hex.DecodeString(value)
				if err != nil {
					t.Fatal(err)
				}
				return b
			}
			seed := func(m map[string]json.RawMessage) uint32 {
				var value uint32
				if err := json.Unmarshal(m["seed"], &value); err != nil {
					t.Fatal(err)
				}
				return value
			}
			initial := decode(c.Initial, "target_hex")
			s := spellReactionState(c.PowerCode)
			// RE-208 勘誤：前批 +0x0D=5 是旗標；有效倒數另讀 +0x30。
			s.Actors[1].SpellTurns = 0
			if initial[13] != 0 {
				s.Actors[1].SpellTurns = int(initial[48])
			}
			s.Actors[1].Threshold = int(initial[16])
			partsInitial := []*int{&s.Actors[1].Head, &s.Actors[1].Chest, &s.Actors[1].RightArm, &s.Actors[1].LeftArm, &s.Actors[1].RightLeg, &s.Actors[1].LeftLeg}
			for i, p := range partsInitial {
				*p = int(initial[17+2*i])
				if *p == 255 {
					*p = 0
				}
			}
			s.Actors[1].Side, s.Actors[1].Member = SideParty, 1
			s.Actors[1].Command = int(initial[37])
			s.Actors[1].Flag03 = int(initial[3])
			s.Actors[0].Negative = c.Negative != 0
			s.Party = append(s.Party, Member{})
			if c.Blocked {
				s.Party[0].Spells[0].Guards = [4]int{16, 16, 16, 16}
			}
			if c.Armed {
				hand := decode(c.Initial, "character_hex")[48:64]
				item := &Item{Name: "Fixture", Record: hex.EncodeToString(hand)}
				s.Party[1].HasWeapon, s.Party[1].WeaponName, s.Party[1].HandItem = true, item.Name, item
				s.Actors[1].Weapon = Weapon{Kind: int(hand[13])}
			}
			random := rng.New(1)
			res := s.Cast(0, 1, AreaRandom, random)
			if len(c.Waits) == 2 {
				if random.State() != seed(c.Waits[1]) || s.Actors[1].X != int(initial[32]) || s.Actors[1].Y != int(initial[31]) || s.spellReaction == nil {
					t.Fatal("傷害頁等待前狀態不符")
				}
				s.Act(random)
			}
			want := decode(c.Final, "target_hex")
			a := s.Actors[1]
			parts := []int{a.Head, a.Chest, a.RightArm, a.LeftArm, a.RightLeg, a.LeftLeg}
			for i, v := range parts {
				expected := int(want[17+2*i])
				if expected == 255 {
					expected = 0
				}
				if v != expected {
					t.Fatalf("部位%d：%d，原版%d", i, v, expected)
				}
			}
			turns := 0
			if want[13] != 0 {
				turns = int(want[48])
			}
			if a.Damage != int(want[14]) || a.Hurt != int(want[45]) || a.SpellTurns != turns || a.Command != int(want[37]) || a.X != int(want[32]) || a.Y != int(want[31]) || random.State() != seed(c.Final) {
				t.Fatalf("角色／seed不符：%+v seed%d", a, random.State())
			}
			if c.Armed {
				ground := decode(c.Final, "ground_hex")[:16]
				if s.Party[1].HandItem != nil || s.Ground[0].Item == nil || s.Ground[0].Item.Record != hex.EncodeToString(ground) || s.Ground[0].Y != int(decode(c.Final, "ground_x")[0]) || s.Ground[0].X != int(decode(c.Final, "ground_y")[0]) {
					t.Fatal("物品／地上座標不符")
				}
			}
			if c.Blocked && res.Verdict != SpellNoEffect {
				t.Fatal("阻擋結果不符")
			}
		})
	}
}
