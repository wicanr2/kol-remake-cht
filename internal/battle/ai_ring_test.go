package battle

import "testing"

type ringTraceRand struct{ limits []int }

func (r *ringTraceRand) Roll(n int) int { r.limits = append(r.limits, n); return 1 }

func TestEnemyContinuesSameRingAfterBlockedTarget(t *testing.T) {
	s := &State{Field: NewField(), Actors: []Actor{
		{Side: SideMonster, Kind: 1, X: 20, Y: 20, Head: 10, Chest: 10, RightArm: 10, LeftArm: 10, Threshold: 100, FleeAt: 100},
		{Side: SideParty, X: 18, Y: 20, Head: 10, Chest: 10, Threshold: 100},
		{Side: SideParty, X: 22, Y: 20, Head: 10, Chest: 10, Threshold: 100},
	}}
	s.Field.Tiles[19*FieldSide+20] = 0x29
	rng := &ringTraceRand{}
	d := s.PlanEnemy(0, rng)
	if d.Target != 2 || d.Ring != 1 || d.Command != CmdMissile {
		t.Fatalf("應選同環未被遮擋的隊員射擊，得到 %+v", d)
	}
	if len(rng.limits) != 3 {
		t.Fatalf("智力、鄰圈與外圈以外出現額外擲骰：%v", rng.limits)
	}
	if rng.limits[1] != 7 || rng.limits[2] != 15 {
		t.Fatalf("起點骰應為環長減一：%v", rng.limits)
	}
}
