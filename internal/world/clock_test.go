package world

import "testing"

// 時段是「時刻落在哪一段」：門檻取自 DGROUP+0x12CC。
func TestClockDayParts(t *testing.T) {
	c := NewClock()
	cases := []struct {
		ticks int
		part  int
	}{{4200, 0}, {3367, 0}, {3366, 1}, {2527, 1}, {2526, 2},
		{1687, 2}, {1686, 3}, {847, 3}, {846, 4}, {0, 4}}
	for _, tc := range cases {
		if got := c.partFor(tc.ticks); got != tc.part {
			t.Errorf("時刻 %d 應該在第 %d 段，卻是 %d", tc.ticks, tc.part, got)
		}
	}
}

// 走到時刻用完就換日，時刻重設成滿格。
func TestClockNewDay(t *testing.T) {
	c := NewClock()
	steps := 0
	for !c.Advance(7, 0) {
		steps++
		if steps > 500 {
			t.Fatal("走了五百步還沒換日")
		}
	}
	// 日期不在 Advance 裡推——紮營那一張上面印的還是前一天。
	if c.Ticks != ClockStart || c.Day != 0 {
		t.Fatalf("換日之後是 %+v", c)
	}
	c.NextDay()
	if c.Day != 1 || c.Month != 3 {
		t.Fatalf("收掉紮營之後是 %+v", c)
	}
	want := (ClockStart / c.StepCost(7, 0)) + 1
	if steps+1 != want {
		t.Fatalf("換日走了 %d 步，預期 %d", steps+1, want)
	}
}

// 每步的成本是 2 × (地形成本 + 隊伍移動率)，平原加移動率 20 是 56。
// 原版實測：平原上第 15 步跨過 DAWN 的門檻，第 14 步還沒跨（world-clock-run2）。
func TestClockStepCost(t *testing.T) {
	c := NewClock()
	if got := c.StepCost(7, 0); got != 56 {
		t.Fatalf("平原每步是 %d", got)
	}
	if got := c.StepCost(4, 20); got != 52 {
		t.Fatalf("道路每步是 %d", got)
	}
	for i := 0; i < 14; i++ {
		c.Advance(7, 20)
	}
	if c.Part != 0 {
		t.Fatalf("第 14 步就換時段了：%d", c.Ticks)
	}
	c.Advance(7, 20)
	if c.Part != 1 {
		t.Fatalf("第 15 步沒換時段：%d", c.Ticks)
	}
}

// 月底換月、年底換年；原版沒有閏年。
func TestClockRollover(t *testing.T) {
	c := &Clock{Day: 30, Month: 11, Year: 0}
	c.NextDay()
	if c.Day != 0 || c.Month != 0 || c.Year != 1 {
		t.Fatalf("十二月三十一日的隔天是 %+v", c)
	}
	c = &Clock{Day: 27, Month: 1}
	c.NextDay()
	if c.Day != 0 || c.Month != 2 {
		t.Fatalf("二月二十八日的隔天是 %+v", c)
	}
}

// 天候是每月各自有上下限的隨機漫步；冬天多一段。
func TestClockWeather(t *testing.T) {
	c := &Clock{Month: 3, Weather: 1} // 四月：上限 3、下限 0
	if got := c.NextWeather(0); got != 2 {
		t.Fatalf("擲 0 應該變壞一段，卻是 %d", got)
	}
	if got := c.NextWeather(50); got != 0 {
		t.Fatalf("擲 50 應該變好一段，卻是 %d", got)
	}
	if got := c.NextWeather(120); got != 1 {
		t.Fatalf("擲 120 兩邊都不動，卻是 %d", got)
	}
	c = &Clock{Month: 0, Weather: 2} // 一月：0x8B，上限 3、下限 1，且第 3 段再往上推
	if got := c.NextWeather(0); got != 4 {
		t.Fatalf("一月從第 2 段變壞應該直接到第 4 段，卻是 %d", got)
	}
	c.Weather = 1
	if got := c.NextWeather(50); got != 1 {
		t.Fatalf("一月的下限是 1，不該再往下，卻是 %d", got)
	}
}

// 換日的營養消耗隨月份不同：一月最餓、六月與九月不餓。
func TestClockHunger(t *testing.T) {
	if got := (&Clock{Month: 0}).Hunger(); got != 4 {
		t.Fatalf("一月是 %d", got)
	}
	if got := (&Clock{Month: 5}).Hunger(); got != 0 {
		t.Fatalf("六月是 %d", got)
	}
}

// 日期框兩行的版面：月名在 byte 欄 3、日期在欄 7；AD 在欄 2、年份在欄 6。
func TestClockDateLines(t *testing.T) {
	got := NewClock().DateLines()
	if len(got) != 2 || got[0] != " Apr 1" || got[1] != "AD  1100" {
		t.Fatalf("日期框是 %q", got)
	}
}

// 循路一步扣的時刻與一般步伐不同：原版量到的是 RoadStepCost，
// 一般步伐在道路上是 2 × (6 + 20) = 52。
func TestRoadStepCost(t *testing.T) {
	c := NewClock()
	normal := c.StepCost(4, 20)
	if normal != 52 {
		t.Fatalf("道路上的一般步伐 %d", normal)
	}
	if RoadStepCost >= normal {
		t.Fatalf("循路一步 %d 不該比一般步伐貴", RoadStepCost)
	}
	// 平原十二步之後連按 ROAD：第 8 次還在 DAWN、第 9 次翻成 MID MORN
	// （`workplace/oracle/world-road-clock` 兩輪都量到第 9 次）。
	c = NewClock()
	for i := 0; i < 12; i++ {
		c.Advance(7, 20)
	}
	for i := 0; i < 8; i++ {
		c.Spend(RoadStepCost)
	}
	if c.Part != 0 {
		t.Fatalf("循路第 8 步時段 %d，預期還在 DAWN", c.Part)
	}
	c.Spend(RoadStepCost)
	if c.Part != 1 {
		t.Fatalf("循路第 9 步時段 %d，預期翻成 MID MORN", c.Part)
	}
}
