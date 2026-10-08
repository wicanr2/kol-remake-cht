package world

import "testing"

func TestQuestArrivalWaitsBeforeTimePart(t *testing.T) {
	data := testData(t)
	data.towns = nil
	travel := New(data, testGrid(t, nil), 20, 20)
	travel.SubX = 7
	clock := &Clock{Ticks: 840, Part: 3, Month: 3}
	session := NewSession(&Screen{Travel: travel, Clock: clock, Movement: 20})
	parts := 0
	session.OnTimePart = func() { parts++ }
	session.OnArrive = func(x, y int) *Prompt { return &Prompt{Kind: PromptQuest} }
	if !session.Walk(1, 0) || session.Prompt == nil {
		t.Fatal("正常換格未進任務提示")
	}
	if clock.Ticks != 840 || clock.Part != 3 || parts != 0 {
		t.Fatalf("任務提示未回答就已更新時段：ticks=%d part=%d encounter=%d", clock.Ticks, clock.Part, parts)
	}
}

func questArrival(t *testing.T) (*Session, *orderRand) {
	t.Helper()
	data := testData(t)
	data.towns = nil
	travel := New(data, testGrid(t, nil), 20, 20)
	travel.SubX = 7
	clock := &Clock{Ticks: 840, Part: 3, Month: 3}
	s := NewSession(&Screen{Travel: travel, Clock: clock, Movement: 20})
	random := &orderRand{}
	s.Rand = random
	s.OnArrive = func(x, y int) *Prompt { return &Prompt{Kind: PromptQuest} }
	s.OnTimePart = func() { random.log = append(random.log, "遭遇") }
	if !s.Walk(1, 0) {
		t.Fatal("正常到達未移動")
	}
	return s, random
}

func TestDeclinedQuestResumesArrivalOnce(t *testing.T) {
	s, random := questArrival(t)
	s.MoveCursor(1)
	s.Accept()
	if s.Screen.Clock.Ticks != 780 || s.Screen.Clock.Part != 4 || s.Prompt != nil || s.pendingArrivalIndex != nil || s.pendingTimePart != nil {
		t.Fatalf("拒絕任務未完整接續：%+v", s.Screen.Clock)
	}
	if len(random.log) != 2 || random.log[0] != "遭遇" || random.log[1] != "天候" {
		t.Fatalf("任務拒絕後顺序錯誤：%v", random.log)
	}
	s.ResumeArrival()
	s.ResumeTimePart()
	if s.Screen.Clock.Ticks != 780 || len(random.log) != 2 || !s.Walk(-1, 0) {
		t.Fatal("拒絕後重複扣時間或不能繼續走")
	}
}

func TestAcceptedQuestWaitsForItsExit(t *testing.T) {
	s, random := questArrival(t)
	busy := false
	s.OnTimePartBusy = func() bool { return busy }
	s.OnPromptAnswer = func(yes bool) { busy = yes }
	s.Accept()
	s.ResumeAfterQuest(16)
	if !busy || s.Prompt != nil || s.Screen.Clock.Part != 3 || s.Screen.Clock.Ticks != 840 || len(random.log) != 0 || s.Walk(-1, 0) {
		t.Fatal("任務戰鬥仍等待就提前完成到達")
	}
	busy = false
	s.ResumeAfterQuest(16)
	if s.Screen.Clock.Ticks != 740 || s.pendingArrivalIndex != nil || s.pendingTimePart != nil || len(random.log) != 2 {
		t.Fatalf("任務出口未依實際索引接續：ticks=%d log=%v", s.Screen.Clock.Ticks, random.log)
	}
	s.ResumeAfterQuest(5)
	if s.Screen.Clock.Ticks != 740 || len(random.log) != 2 || !s.Walk(-1, 0) {
		t.Fatal("任務返回後重複扣時刻或不能走")
	}
}

func TestQuestReturnCanWaitForFollowingAmbush(t *testing.T) {
	s, random := questArrival(t)
	busy := false
	s.OnTimePartBusy = func() bool { return busy }
	s.OnPromptAnswer = func(bool) { busy = true }
	s.Accept()
	s.OnTimePart = func() { random.log = append(random.log, "遭遇"); busy = true }
	busy = false
	s.ResumeAfterQuest(16)
	if !busy || s.pendingArrivalIndex != nil || s.pendingTimePart == nil || s.Screen.Clock.Ticks != 840 || len(random.log) != 1 {
		t.Fatal("任務後隨機遭遇未保留自己的待續時間")
	}
	busy = false
	s.ResumeAfterBattle(5)
	if s.Screen.Clock.Ticks != 772 || len(random.log) != 2 || s.pendingTimePart != nil {
		t.Fatalf("後續遭遇未覆寫成本並完成一次：ticks=%d log=%v", s.Screen.Clock.Ticks, random.log)
	}
	s.ResumeAfterQuest(16)
	s.ResumeAfterBattle(5)
	if s.Screen.Clock.Ticks != 772 || len(random.log) != 2 {
		t.Fatal("到達或後續遭遇被重複接續")
	}
}
