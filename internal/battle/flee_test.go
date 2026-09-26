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
	s.Site.QuestItem.Record = "01000000000000000000000000000000"
	s.Party[1].HandItem = &Item{Name: "CROWN", Record: s.Site.QuestItem.Record}
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
	s.Party[1].HandItem = nil
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

// 逃走那一句：單人隊伍用單數那一筆（`0x215C9`）。
func TestFleeRecords(t *testing.T) {
	one := FleeRecords(1)
	if len(one) != 2 || one[0] != FleeForRecord || one[1] != OneLifeRecord {
		t.Fatalf("單人應該是 %#x + %#x，得到 %v", FleeForRecord, OneLifeRecord, one)
	}
	many := FleeRecords(3)
	if many[1] != ManyLivesRecord {
		t.Fatalf("多人應該用 %#x，得到 %#x", ManyLivesRecord, many[1])
	}
}

// 答應逃走之後回到執行段那一層，訊息清空——呼叫端拿它當「那一句還沒印」的判準。
func TestFleeReturnsToExecuting(t *testing.T) {
	s := fleeState()
	s.Actors[0].X, s.Actors[1].X = FieldMin, FieldMin
	s.Exec.Text = []string{"上一句"}
	s.OfferFlee()
	s.Flee()
	if s.Plan.Prompt != PromptExecuting {
		t.Fatalf("應該回到執行段那一層，得到 %d", s.Plan.Prompt)
	}
	if len(s.Exec.Text) != 0 {
		t.Fatalf("訊息應該清空，得到 %v", s.Exec.Text)
	}
}

// 任務地點那一場會從地點記錄填上「任務要的那一件」（RE-113）；隨機遭遇不填。
func TestNewQuestTakesQuestItemName(t *testing.T) {
	site := &Site{Index: 3, Count: 0,
		QuestItem: QuestItem{Item: Item{Name: "Truth Sword"}},
		PartyX:    []int{20}, PartyY: []int{20}}
	s := NewQuest(site, []Member{{Name: "POONTANG", Size: 100, Health: 100}},
		Monster{}, Generate(&fixedRand{values: []int{1}}))
	if s == nil {
		t.Fatal("該開得起來")
	}
	if s.QuestItem != "TRUTH SWORD" {
		t.Fatalf("任務物品是 %q，期望 TRUTH SWORD（畫面上整段大寫）", s.QuestItem)
	}
	// 隨機遭遇（Site.Index < 0）不填——原版 `sub_2886C` 先看 `word_34968`。
	site.Index = -1
	s = NewQuest(site, []Member{{Name: "POONTANG", Size: 100, Health: 100}},
		Monster{}, Generate(&fixedRand{values: []int{1}}))
	if s.QuestItem != "" {
		t.Fatalf("隨機遭遇不該有任務物品，得到 %q", s.QuestItem)
	}
}

// 「Begin next round?」那一列：一般與任務戰鬥多一格逃走，競技場與訓練戰沒有（RE-143）。
func TestNextRoundBarFleeIcon(t *testing.T) {
	s := fleeState()
	if bar := s.NextRoundBar(); len(bar) != 3 || bar[2] != IconFleeBar {
		t.Fatalf("一般戰鬥應該有逃走格，得到 %#x", bar)
	}
	s.RankFight = true
	if bar := s.NextRoundBar(); len(bar) != 2 {
		t.Fatalf("競技場不該有逃走格，得到 %#x", bar)
	}
}

// 白框停在哪一格決定往哪裡走：U-TURN 重新規劃、逃走格進問句、其餘進執行段。
func TestChooseNextRound(t *testing.T) {
	s := fleeState()
	s.Stage = StageField
	s.Plan.Prompt = PromptNextRound
	s.Plan.Bar = s.NextRoundBar()
	s.Plan.Selected = 2
	if c := s.ChooseNextRound(); c != NextRoundAskFlee || s.Stage != StageFlee || s.Plan.Prompt != PromptFlee {
		t.Fatalf("逃走格應該進問句，得到 %d／%d／%d", c, s.Stage, s.Plan.Prompt)
	}
}

// 答應之後全隊都逃掉；擲出大於門檻、手上是武器的人把武器丟在腳下。
// 從這一格逃走不經過 0x21176，帶著任務物品也不算完成。
func TestBarFleeDropsWeaponOnFailedRoll(t *testing.T) {
	for _, c := range []struct {
		roll    int
		dropped bool
	}{{roll: 100, dropped: true}, {roll: 1, dropped: false}} {
		s := fleeState()
		s.Actors[0].Intellect, s.Actors[0].Strength, s.Actors[0].Level = 60, 60, 3 // 門檻 60
		s.Party[0].HasWeapon = true
		s.Party[0].Weapon = Weapon{Kind: 0x20, Code: 1}
		s.Party[1].HasWeapon = false // 帶著任務物品的那一名手上不是武器，不丟
		s.Stage, s.fleeFromBar = StageFlee, true
		s.Rand = &fixedRoll{values: []int{c.roll, c.roll}}
		if !s.Answer(true) {
			t.Fatal("答應逃走應該回報處理了")
		}
		if s.Verdict != VerdictFled || s.Stage != StageDone || s.QuestDone != 0 {
			t.Fatalf("應該逃走而且不標任務完成，得到 %d／%d／%d", s.Verdict, s.Stage, s.QuestDone)
		}
		on := s.Ground[0].Name != ""
		if on != c.dropped || s.Party[0].HasWeapon == c.dropped {
			t.Fatalf("擲 %d：丟下=%v，應為 %v", c.roll, on, c.dropped)
		}
	}
}

// 不逃就跳回重新規劃這一回合（`0x1FD1B` 走 `0x1EF3A`），與 U-TURN 相同。
func TestBarFleeDeclined(t *testing.T) {
	s := fleeState()
	s.Stage, s.fleeFromBar = StageFlee, true
	s.Answer(false)
	if s.Stage != StageField || s.fleeFromBar || s.Verdict != 0 {
		t.Fatalf("不逃應該回戰場重新規劃，得到 %d／%v／%d", s.Stage, s.fleeFromBar, s.Verdict)
	}
}
