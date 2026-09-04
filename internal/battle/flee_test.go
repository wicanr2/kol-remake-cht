package battle

import "testing"

// 兩軸任一個到邊就算站到邊上（`0x21087`）。
func TestOnFieldEdge(t *testing.T) {
	cases := []struct {
		x, y int
		want bool
	}{
		{FieldMin, 30, true}, {30, FieldMin, true},
		{FieldMax, 30, true}, {30, FieldMax, true},
		{FieldMin, FieldMax, true},
		{30, 30, false}, {FieldMin + 1, 30, false}, {30, FieldMax - 1, false},
	}
	for _, c := range cases {
		if got := OnFieldEdge(c.x, c.y); got != c.want {
			t.Errorf("(%d, %d) 應該是 %v，得到 %v", c.x, c.y, c.want, got)
		}
	}
}

func fleeState() *State {
	s := &State{Field: NewField(), Actors: make([]Actor, 3),
		Site: &Site{Index: 0}, QuestItem: "CROWN",
		Party: []Member{{WeaponName: "SWORD"}, {WeaponName: "CROWN"}}}
	s.Actors[0] = Actor{Side: SideParty, Member: 0, X: 30, Y: 30, Threshold: 100}
	s.Actors[1] = Actor{Side: SideParty, Member: 1, X: 30, Y: 30, Threshold: 100}
	s.Actors[2] = Actor{Side: SideMonster, X: 20, Y: 20, Threshold: 100}
	for i := range s.Actors {
		s.Actors[i].SetParts()
	}
	s.Party[0].Quests = []int{QuestFlagTaken, 0, 0, 0, 0, 0}
	s.Party[1].Quests = []int{QuestFlagTaken, 0, 0, 0, 0, 0}
	return s
}

// 站在戰場中央就連問都不問。
func TestFleeNotOfferedInTheMiddle(t *testing.T) {
	s := fleeState()
	if o := s.FleeOffer(); o.Ask || o.Edge != 0 || o.Carrying != 0 {
		t.Fatalf("中央不該問，得到 %+v", o)
	}
}

// 帶著任務要的東西站到邊上就問，而且逃掉算完成。
func TestFleeWithQuestItem(t *testing.T) {
	s := fleeState()
	s.Actors[1].Y = FieldMax
	o := s.FleeOffer()
	if !o.Ask || o.Edge != 1 || o.Carrying != 1 {
		t.Fatalf("帶著東西到邊上應該問，得到 %+v", o)
	}
	if n := s.Flee(); n != 2 {
		t.Fatalf("逃掉應該把整隊標成完成，得到 %d", n)
	}
	if s.Verdict != VerdictFled || s.Stage != StageDone {
		t.Fatalf("結果應該是逃走，得到 %d／%d", s.Verdict, s.Stage)
	}
}

// 手上不是那件東西時只剩「全隊都到邊上」那一條，而且逃掉不算完成。
func TestFleeWithoutQuestItem(t *testing.T) {
	s := fleeState()
	s.Actors[0].X = FieldMin
	if o := s.FleeOffer(); o.Ask {
		t.Fatalf("只有一個人到邊上又沒帶東西，不該問，得到 %+v", o)
	}
	s.Actors[1].X = FieldMin
	o := s.FleeOffer()
	if !o.Ask || o.Edge != 2 || o.Carrying != 1 {
		t.Fatalf("全隊到邊上應該問，得到 %+v", o)
	}
	// 這一次沒有人帶著東西：任務不標。
	s.Party[1].WeaponName = "SWORD"
	if n := s.Flee(); n != 0 {
		t.Fatalf("沒帶東西不該標完成，得到 %d", n)
	}
}

// 隨機遭遇連比都不比（`sub_2886C:0x28875`）：只剩全隊到邊上那一條。
func TestFleeAmbushIgnoresQuestItem(t *testing.T) {
	s := fleeState()
	s.Site = &Site{Index: -1}
	s.Actors[1].Y = FieldMax
	if o := s.FleeOffer(); o.Ask || o.Carrying != 0 {
		t.Fatalf("隨機遭遇不看任務物品，得到 %+v", o)
	}
}

// 執行段走完、還沒分勝負又站到邊上，就停在逃走那一問；不逃就排下一回合。
func TestOfferFleeAfterRound(t *testing.T) {
	s := fleeState()
	s.Rand = &fixedRoll{values: []int{50}}
	s.Stage = StageField
	if s.OfferFlee() {
		t.Fatal("站在中央不該問")
	}
	s.Actors[0].X, s.Actors[1].X = FieldMin, FieldMin
	if !s.OfferFlee() {
		t.Fatal("全隊到邊上應該問")
	}
	if s.Stage != StageFlee || s.Plan.Prompt != PromptFlee {
		t.Fatalf("應該停在逃走那一問，得到 %v／%v", s.Stage, s.Plan.Prompt)
	}
	// 答否回戰場，繼續打。
	if s.Answer(false) || s.Stage != StageField {
		t.Fatalf("不逃應該回戰場，得到 %v", s.Stage)
	}
	// 答應就走。
	s.OfferFlee()
	if !s.Answer(true) || s.Verdict != VerdictFled {
		t.Fatalf("逃走應該結束這一場，得到 %v", s.Verdict)
	}
}
