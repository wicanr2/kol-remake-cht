package battle

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/rng"
)

// RE-218 的正常三發命中。夾具保存原版觀察狀態，不能代替正常 GUI 驗收。
func TestMissileDamageNormalOriginal(t *testing.T) {
	path := os.Getenv("KOL_MISSILE_HIT_CASES")
	if path == "" {
		t.Skip("需要正常原版命中收據")
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Seed      uint32
		AfterSeed uint32 `json:"after_seed"`
		Actor     int
		Actors    []Actor
		Field     *Field
		Expected  map[string]int
		Hurt      int
		Damage    int
	}
	if err = json.Unmarshal(body, &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) != 3 {
		t.Fatal("必須包含正常三發命中")
	}
	for i, c := range cases {
		s := State{Actors: c.Actors, Field: c.Field}
		s.Plan.Memo = map[int]Cell{c.Actor: {Row: s.Actors[0].X, Col: s.Actors[0].Y}}
		random := rng.NewState(c.Seed)
		count := s.Actors[c.Actor].Kind
		out := s.shoot(c.Actor, random)
		if out.Target != 0 || out.Outcome == nil || out.Outcome.Damage != c.Damage {
			t.Fatalf("第%d發結算=%+v outcome=%+v，原版傷害%d", i+1, out, out.Outcome, c.Damage)
		}
		a := s.Actors[0]
		got := map[string]int{"Damage": a.Damage, "Head": a.Head, "Chest": a.Chest,
			"RightArm": a.RightArm, "LeftArm": a.LeftArm, "RightLeg": a.RightLeg,
			"LeftLeg": a.LeftLeg, "X": a.X, "Y": a.Y}
		for k, want := range c.Expected {
			if got[k] != want {
				t.Fatalf("第%d發%s=%d，原版%d", i+1, k, got[k], want)
			}
		}
		if a.Hurt != c.Hurt || random.State() != c.AfterSeed || s.Actors[c.Actor].Kind != count-1 {
			t.Fatalf("第%d發Hurt／seed／計數不符：%d/%d/%d", i+1, a.Hurt, random.State(), s.Actors[c.Actor].Kind)
		}
	}
}
