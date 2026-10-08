package battle

import "testing"

type re273ShotRand struct{ calls int }

func (r *re273ShotRand) Roll(n int) int { r.calls++; return n }

// RE-273 的正常反例：先倒地的人不能遮住同格仍站立的射擊目標。
func TestMissileFallenStackTarget(t *testing.T) {
	for _, standing := range []bool{true, false} {
		actors := []Actor{
			{Side: SideMonster, X: 30, Y: 30, Head: 10, Chest: 10, Command: CmdMissile, Strength: 100, Flag1D: 1},
			{Side: SideParty, X: 30, Y: 33, Head: 10, Chest: 0},
			{Side: SideParty, X: 30, Y: 33, Head: 10, Chest: 10},
		}
		if !standing {
			actors[2].Chest = 0
		}
		s := State{Actors: actors, Field: NewField()}
		s.Plan.Memo = map[int]Cell{0: {Row: 30, Col: 33}}
		random := &re273ShotRand{}
		out := s.shootPhase(0, random, true)
		if standing {
			if out.Target != 2 || out.Message == MissileNothingRecord || random.calls != 2 {
				t.Fatalf("站立目標被倒地者遮住：out=%+v rolls=%d", out, random.calls)
			}
		} else if out.Message != MissileNothingRecord || random.calls != 1 {
			t.Fatalf("只有倒地者仍成為目標：out=%+v rolls=%d", out, random.calls)
		}
	}
}
