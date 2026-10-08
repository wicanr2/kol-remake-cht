package battle

import (
	"encoding/hex"
	"encoding/json"
	"github.com/wicanr2/kol-remake-cht/internal/rng"
	"os"
	"testing"
)

func TestSpellLifecycleOriginalCases(t *testing.T) {
	path := os.Getenv("KOL_SPELL_LIFECYCLE_SOURCE")
	if path == "" {
		t.Skip("需要固定種子的原版法術生命週期收據")
	}
	type snapshot struct {
		TargetHex                   string `json:"target_hex"`
		CharacterHex                string `json:"character_hex"`
		Fatigue, EffectiveCountdown int
		Seed                        uint32
	}
	var data struct {
		Cases []struct {
			Name                          string
			Effect, Negative, Target      int
			Blocked                       bool
			Initial, AfterCast, AfterTick snapshot
		}
	}
	// JSON 使用明示 snake_case，避免把未解到的欄位零值當作原版結果。
	var raw struct{ Cases []json.RawMessage }
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(body, &raw); err != nil {
		t.Fatal(err)
	}
	decode := func(b json.RawMessage) snapshot {
		var x struct {
			TargetHex          string `json:"target_hex"`
			CharacterHex       string `json:"character_hex"`
			Fatigue            int    `json:"fatigue"`
			EffectiveCountdown int    `json:"effective_countdown"`
			Seed               uint32 `json:"seed"`
		}
		if err := json.Unmarshal(b, &x); err != nil {
			t.Fatal(err)
		}
		return snapshot{x.TargetHex, x.CharacterHex, x.Fatigue, x.EffectiveCountdown, x.Seed}
	}
	for _, b := range raw.Cases {
		var x struct {
			Name                     string
			Effect, Negative, Target int
			Blocked                  bool
			Initial                  json.RawMessage
			AfterCast                json.RawMessage `json:"after_cast"`
			AfterTick                json.RawMessage `json:"after_tick"`
		}
		if err = json.Unmarshal(b, &x); err != nil {
			t.Fatal(err)
		}
		data.Cases = append(data.Cases, struct {
			Name                          string
			Effect, Negative, Target      int
			Blocked                       bool
			Initial, AfterCast, AfterTick snapshot
		}{x.Name, x.Effect, x.Negative, x.Target, x.Blocked, decode(x.Initial), decode(x.AfterCast), decode(x.AfterTick)})
	}
	if len(data.Cases) != 21 {
		t.Fatal("原版案例形狀不符")
	}
	for _, c := range data.Cases {
		t.Run(c.Name, func(t *testing.T) {
			bytes := func(value string) []byte {
				b, e := hex.DecodeString(value)
				if e != nil {
					t.Fatal(e)
				}
				return b
			}
			initial, character := bytes(c.Initial.TargetHex), bytes(c.Initial.CharacterHex)
			m := Member{Strength: int(character[0x1D]), Nutrition: int(character[0x9F]), Quickness: int(character[0x1E]), Foresight: int(character[0x21]), Intellect: int(character[0x23])}
			s := &State{Party: []Member{{Spells: []Spell{{Effect: c.Effect, TurnCode: 0, PowerCode: 130}}}, m}, Actors: make([]Actor, 7), Monster: Monster{Strength: 70, Quickness: 80, Foresight: 90, Intellect: 110}}
			s.Actors[0] = Actor{Side: SideParty, Member: 0, Negative: c.Negative != 0}
			a := &s.Actors[c.Target]
			*a = Actor{Side: SideParty, Member: 1, Strength: 150, Quickness: 150, Foresight: 150, Intellect: 150,
				SpellTurns: c.Initial.EffectiveCountdown, Fatigue: c.Initial.Fatigue, SkillPenaltyOff: 7, SkillPenaltyDef: 9, Threshold: 109, Flag03: int(initial[3]), Damage: int(initial[14]), Hurt: int(initial[45])}
			if c.Target == 6 {
				a.Side, a.Member = SideMonster, -1
			}
			a.Head, a.Chest, a.RightArm, a.LeftArm, a.RightLeg, a.LeftLeg = int(initial[17]), int(initial[19]), int(initial[21]), int(initial[23]), int(initial[25]), int(initial[27])
			if c.Blocked {
				s.Party[0].Spells[0].Guards = [4]int{16, 16, 16, 16}
			}
			random := rng.New(1)
			compare := func(want snapshot) {
				t.Helper()
				b := bytes(want.TargetHex)
				got := []int{a.Strength, a.Quickness, a.Foresight, a.Intellect, a.SkillPenaltyOff, a.SkillPenaltyDef, a.Damage, a.Hurt, a.Head, a.Chest, a.RightArm, a.LeftArm, a.RightLeg, a.LeftLeg}
				offsets := []int{8, 9, 10, 12, 46, 47, 14, 45, 17, 19, 21, 23, 25, 27}
				for i, v := range got {
					w := int(b[offsets[i]])
					if offsets[i] >= 17 && offsets[i] <= 27 && w == 255 {
						w = 0
					}
					if v != w {
						t.Errorf("欄位+%02X=%d，原版%d", offsets[i], v, w)
					}
				}
				if a.SpellTurns != want.EffectiveCountdown || a.Fatigue != want.Fatigue || random.State() != want.Seed {
					t.Errorf("有效倒數/疲勞/seed=%d/%d/%d，原版%d/%d/%d", a.SpellTurns, a.Fatigue, random.State(), want.EffectiveCountdown, want.Fatigue, want.Seed)
				}
			}
			if c.Effect != 0 {
				s.Cast(0, c.Target, AreaRandom, random)
			}
			compare(c.AfterCast)
			s.WearsOff(c.Target)
			compare(c.AfterTick)
		})
	}
}

func lifecycleRoundState(edge, won bool) *State {
	s := &State{RankFight: true, Party: []Member{{Name: "TARGET", Strength: 100, Quickness: 80, Foresight: 90, Intellect: 70}}, Actors: make([]Actor, 7)}
	s.Actors[0] = Actor{Name: "TARGET", Side: SideParty, Member: 0, Strength: 150, Quickness: 80, Foresight: 90, Intellect: 70, SpellTurns: 1, Head: 16, Chest: 16, Threshold: 109, Stamina: 200, X: 10, Y: 10}
	s.Actors[6] = Actor{Side: SideMonster, Member: -1, Head: 16, Chest: 16, Threshold: 109, Stamina: 200, X: 12, Y: 10}
	for i := 1; i < 6; i++ {
		s.Actors[i].Side = SideEmpty
	}
	if edge {
		s.Actors[0].X = FieldMin
	}
	if won {
		s.Actors[6].Head = 0
	}
	s.Plan.Prompt = PromptExecuting
	return s
}

func TestSpellLifecycleExpiryBeforeRoundDecision(t *testing.T) {
	for _, c := range []struct {
		name      string
		edge, won bool
	}{{"next-round", false, false}, {"flee", true, false}, {"victory", false, true}} {
		t.Run(c.name, func(t *testing.T) {
			s := lifecycleRoundState(c.edge, c.won)
			out, ok := s.Step(maxRand{})
			if !ok || out.Message != WearsOffRecord || out.Actor != 0 || s.Actors[0].SpellTurns != 0 || s.Actors[0].Strength != 100 || s.Stage == StageDone || s.Plan.Prompt == PromptFlee || s.Round.Number != 0 {
				t.Fatalf("失效頁晚於回合判定：%+v state %+v", out, s)
			}
			_, ok = s.Step(maxRand{})
			if ok {
				t.Fatal("結算後仍重複失效")
			}
			if c.won && s.Verdict != VerdictWon || c.edge && s.Plan.Prompt != PromptFlee || !c.edge && !c.won && s.Round.Number != 1 {
				t.Fatal("回合出口不符")
			}
		})
	}
}

func TestSpellLifecycleExhaustionWaitBeforeExpiry(t *testing.T) {
	s := lifecycleRoundState(false, false)
	s.Actors[0].Command, s.Actors[0].Fatigue = CmdRest, 600
	out, ok := s.Step(maxRand{})
	if !ok || out.Exhausted != ExhaustedPartyRecord || s.Actors[0].SpellTurns != 1 {
		t.Fatalf("力竭等待前已跑倒數：%+v", out)
	}
	fatigue := s.Actors[0].Fatigue
	out, ok = s.Step(maxRand{})
	if !ok || out.Message != WearsOffRecord || s.Actors[0].Fatigue != fatigue || s.Actors[0].SpellTurns != 0 {
		t.Fatalf("力竭後未先失效或重複算疲勞：%+v", out)
	}
}

func TestRoundCommandResetSlotBoundary(t *testing.T) {
	for _, c := range []struct {
		name  string
		slot  int
		side  Side
		down  bool
		clear bool
	}{
		{"slot5-party", 5, SideParty, false, false},
		{"slot5-monster", 5, SideMonster, false, false},
		{"slot6-monster", 6, SideMonster, false, true},
		{"slot6-party", 6, SideParty, false, true},
		{"slot17-monster", 17, SideMonster, false, true},
		{"slot6-down", 6, SideMonster, true, true},
		{"slot6-empty", 6, SideEmpty, false, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			s := &State{Actors: make([]Actor, MaxActors)}
			for i := range s.Actors {
				s.Actors[i].Side = SideEmpty
			}
			a := &s.Actors[c.slot]
			*a = Actor{Side: c.side, Command: CmdRest, Head: 16, Chest: 16, Stamina: 200, Fatigue: 100}
			if c.down {
				a.Head = 0
			}
			if _, ok := s.Act(maxRand{}); ok {
				t.Fatal("unexpected wait")
			}
			want := CmdRest
			if c.clear {
				want = 0
			}
			if a.Command != want {
				t.Fatalf("Command=%d want %d", a.Command, want)
			}
		})
	}
}

func TestRoundCommandResetAfterSpellWait(t *testing.T) {
	for _, slot := range []int{5, 6} {
		s := &State{Actors: make([]Actor, 7), Party: make([]Member, 6), Monster: Monster{Strength: 70}}
		for i := range s.Actors {
			s.Actors[i].Side = SideEmpty
		}
		s.Party[5] = Member{Strength: 100, Nutrition: 100}
		a := &s.Actors[slot]
		*a = Actor{Side: SideMonster, Member: -1, Command: CmdRest, SpellTurns: 1, Strength: 150,
			Head: 16, Chest: 16, Stamina: 200, Fatigue: 100, Hurt: 0}
		if slot == 5 {
			a.Side, a.Member = SideParty, 5
		}
		out, ok := s.Act(maxRand{})
		if !ok || out.Message != WearsOffRecord || out.Actor != slot || a.Command != CmdRest || a.SpellTurns != 0 {
			t.Fatalf("slot%d wait=%+v command%d turns%d", slot, out, a.Command, a.SpellTurns)
		}
		fatigue, strength := a.Fatigue, a.Strength
		if _, ok = s.Act(maxRand{}); ok {
			t.Fatal("spell wait repeated")
		}
		want := CmdRest
		if slot == 6 {
			want = 0
		}
		if a.Command != want || a.Fatigue != fatigue || a.Strength != strength || a.SpellTurns != 0 {
			t.Fatalf("slot%d resumed command%d fatigue%d strength%d turns%d", slot, a.Command, a.Fatigue, a.Strength, a.SpellTurns)
		}
	}
}
