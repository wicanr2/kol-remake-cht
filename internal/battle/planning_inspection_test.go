package battle

import "testing"

type inspectionSequence struct{ calls int }

func (r *inspectionSequence) Roll(n int) int { r.calls++; return (r.calls-1)%n + 1 }

func TestPlanningInspectionEnteredOnce(t *testing.T) {
	r := &inspectionSequence{}
	s := &State{Stage: StageField, Rand: r, Party: []Member{{Name: "actor"}}, Actors: []Actor{
		{Side: SideParty, Head: 0, Chest: 10, Style: 3, Guard: 4},
		{Side: SideEmpty, Style: 2, Guard: 5},
		{Side: SideParty, Head: 10, Chest: 10, Member: 0},
	}, Round: Round{Order: []int{0, 1, 2}}}
	s.AdvanceToParty()
	if s.Current() != 2 || r.calls != 3 || s.Round.InspectionRoll != 3 {
		t.Fatalf("跳過順序 current=%d calls=%d roll=%d", s.Current(), r.calls, s.Round.InspectionRoll)
	}
	for _, i := range []int{0, 1} {
		if s.Actors[i].Style != 0 || s.Actors[i].Guard != 0 {
			t.Fatal("被跳過人物未清除招式／防禦")
		}
	}
	for n := 0; n < 3; n++ {
		s.BeginTurn()
		s.AdvanceToParty()
	}
	if r.calls != 3 || s.Round.InspectionRoll != 3 {
		t.Fatal("重複入口或取消重設門檻")
	}
	if !s.EndTurn() || r.calls != 3 {
		t.Fatal("規劃結束後額外擲骰")
	}
	s.BeginPlanning(r)
	if r.calls != 7 || s.Round.InspectionRoll != 7 {
		t.Fatalf("新回合未重設一次性狀態 calls=%d roll=%d", r.calls, s.Round.InspectionRoll)
	}
}
