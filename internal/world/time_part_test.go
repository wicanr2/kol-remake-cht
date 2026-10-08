package world

import "testing"

func delayedTimeSession(t *testing.T) (*Session, *orderRand) {
	t.Helper()
	c := NewClock()
	c.Ticks = 3360
	s := NewSession(&Screen{Travel: New(testData(t), testGrid(t, nil), 20, 20),
		Clock: c, Movement: 20})
	rng := &orderRand{}
	s.Rand = rng
	return s, rng
}

func TestTimePartWaitsForBattleAndResumesOnce(t *testing.T) {
	s, rng := delayedTimeSession(t)
	busy := true
	s.OnTimePart = func() { rng.log = append(rng.log, "遭遇") }
	s.OnTimePartBusy = func() bool { return busy }
	s.spendTime(40)
	if s.Screen.Clock.Ticks != 3360 || len(rng.log) != 1 {
		t.Fatal("戰鬥未完成便擲天候或扣時刻")
	}
	before := *s.Screen.Travel
	if s.Walk(1, 0) || s.FollowRoad(true) || *s.Screen.Travel != before {
		t.Fatal("待處理的時段允許世界移動")
	}
	s.ResumeTimePart()
	s.spendTime(60)
	if s.Screen.Clock.Ticks != 3360 || len(rng.log) != 1 {
		t.Fatal("等待期間被重複扣時刻")
	}
	busy = false
	s.ResumeTimePart()
	s.ResumeTimePart()
	if s.Screen.Clock.Ticks != 3320 || len(rng.log) != 2 || rng.log[1] != "天候" {
		t.Fatalf("接續結果錯誤：ticks=%d log=%v", s.Screen.Clock.Ticks, rng.log)
	}
}

func TestTimePartWaitsAcrossPromptChainAndBattle(t *testing.T) {
	s, rng := delayedTimeSession(t)
	busy := false
	s.OnTimePartBusy = func() bool { return busy }
	s.OnTimePart = func() { s.SetPrompt(&Prompt{Kind: PromptRideAway}) }
	s.OnPromptAnswer = func(bool) { s.SetPrompt(&Prompt{Kind: PromptNotice}) }
	s.spendTime(60)
	s.Accept()
	if s.Prompt == nil || s.Prompt.Kind != PromptNotice || len(rng.log) != 0 {
		t.Fatal("提示链未完成已擲天候")
	}
	s.OnPromptAnswer = func(bool) { busy = true }
	s.Accept()
	if s.Prompt != nil || s.Screen.Clock.Ticks != 3360 || len(rng.log) != 0 {
		t.Fatal("提示接戰鬥時提早完成時間操作")
	}
	busy = false
	s.ResumeTimePart()
	if s.Screen.Clock.Ticks != 3300 || len(rng.log) != 1 {
		t.Fatal("整條提示與戰鬥完成後未接續")
	}
}

func TestHealingPromptFinishesTimePart(t *testing.T) {
	s, rng := delayedTimeSession(t)
	s.OnTimePart = func() { s.SetPrompt(&Prompt{Kind: PromptHealing}) }
	s.spendTime(60)
	s.Accept()
	if s.Prompt != nil || s.Screen.Clock.Ticks != 3300 || len(rng.log) != 1 {
		t.Fatal("治療問答完成後未接續世界時間")
	}
}

func TestCampingDateWaitsForEncounterPrompt(t *testing.T) {
	s, rng := delayedTimeSession(t)
	c := s.Screen.Clock
	c.Ticks, c.Part = 10, DayParts-1
	s.spendTime(60)
	s.OnTimePart = func() { s.SetPrompt(&Prompt{Kind: PromptHealing}) }
	s.EndCamping()
	if c.Day != 0 || s.Screen.Group != CampWeather || len(rng.log) != 0 {
		t.Fatal("換日遭遇未完成便推日期或重設天候")
	}
	s.Accept()
	s.ResumeTimePart()
	if c.Day != 1 || c.Weather != ClearWeather || s.Screen.Group != ClearWeather || len(rng.log) != 1 {
		t.Fatal("換日遭遇完成後的日期／天候接續錯誤")
	}
}
