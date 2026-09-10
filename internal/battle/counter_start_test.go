package battle

import "testing"

type counterStartRand struct {
	t            *testing.T
	first, rolls int
}

func (r *counterStartRand) Roll(n int) int {
	r.rolls++
	if n != 7 {
		r.t.Fatalf("反制搜尋起點骰為 %d，預期 7", n)
	}
	return r.first
}

func TestCounterScanUsesOriginalStartRange(t *testing.T) {
	for _, occupied := range []bool{false, true} {
		for first := 1; first <= 7; first++ {
			s := &State{Actors: []Actor{
				{Side: SideParty, X: 3, Y: 4, Head: 10, Chest: 10, Threshold: 100},
				{Side: SideParty, X: 3, Y: 3, Head: 10, Chest: 10, Threshold: 100},
				{Side: SideMonster, X: 4, Y: 4, Facing: 3},
			}}
			if !occupied {
				s.Actors[0].Side = SideEmpty
				s.Actors[1].Side = SideEmpty
			}
			rng := &counterStartRand{t: t, first: first}
			style, target, ok := s.counterScan(2, rng)
			if rng.rolls != 1 || ok != occupied {
				t.Fatalf("起點 %d 結果 %v，擲骰 %d", first, ok, rng.rolls)
			}
			wantTarget, wantFacing := -1, 3
			if occupied {
				wantTarget, wantFacing = 1, 7
				if first == 1 {
					wantTarget, wantFacing = 0, 0
				}
				if style != EnemyNearStyle {
					t.Fatal("招式不同")
				}
			}
			if target != wantTarget || s.Actors[2].Facing != wantFacing {
				t.Fatalf("起點 %d 目標／方向 %d/%d，預期 %d/%d", first, target, s.Actors[2].Facing, wantTarget, wantFacing)
			}
		}
	}
}
