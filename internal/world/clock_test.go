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
	for !c.Advance(0, 0) {
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
	want := (ClockStart / c.StepCost(0, 0)) + 1
	if steps+1 != want {
		t.Fatalf("換日走了 %d 步，預期 %d", steps+1, want)
	}
}

// 每步的成本是 `2 × ([0x12B8 + dy] + 隊伍移動率)`：索引是這一步的 Y 位移，
// 不是地形（RE-151）。dosgolem 在平原上量到移動率 20 時往北每步 40、往南與
// 東西每步 60。
func TestClockStepCost(t *testing.T) {
	c := NewClock()
	for _, v := range []struct{ dy, want int }{{-1, 40}, {0, 60}, {1, 60}} {
		if got := c.StepCost(v.dy, 0); got != v.want {
			t.Errorf("dy=%d 每步 %d，原版 %d", v.dy, got, v.want)
		}
	}
}

// clockStep 照原版 `sub_297F2` 的順序走一步：先依扣之前的時刻判時段，再扣。
func clockStep(c *Clock, dy int) {
	c.UpdatePart()
	c.Advance(dy, 20)
}

// 東西走十四步是 840，4200 − 840 = 3360 已經跨過 DAWN 的下限 3367，但時段要到
// 下一步開頭才換（RE-151）：第 14 步畫面上還是 DAWN，第 15 步換成 MID MORN。
func TestClockPartChangesOneStepLate(t *testing.T) {
	c := NewClock()
	for i := 0; i < 14; i++ {
		clockStep(c, 0)
	}
	if c.Ticks != 3360 || c.Part != 0 {
		t.Fatalf("第 14 步應為時刻 3360、還在 DAWN，得到 %d／%d", c.Ticks, c.Part)
	}
	clockStep(c, 0)
	if c.Part != 1 {
		t.Fatalf("第 15 步開頭應換成 MID MORN，得到 %d", c.Part)
	}
}

// `world-road-run1` 那一條：往北十二步（每步 40），循路先斜向西北三步拉回中線
// （每步 40），之後往西（每步 60）。第 7 次循路跨過門檻、第 8 次換時段；原版錄影
// 在第 9 次才看到，是擷取比按鍵晚一拍（出城那一下的畫格還停在城門提示）。
func TestRoadRunClock(t *testing.T) {
	c := NewClock()
	for i := 0; i < 12+3; i++ {
		clockStep(c, -1)
	}
	for i := 0; i < 4; i++ {
		clockStep(c, 0)
	}
	if c.Ticks != 3360 || c.Part != 0 {
		t.Fatalf("循路第 7 步應為時刻 3360、還在 DAWN，得到 %d／%d", c.Ticks, c.Part)
	}
	clockStep(c, 0)
	if c.Part != 1 {
		t.Fatalf("循路第 8 步應換成 MID MORN，得到 %d", c.Part)
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
	zh := NewClock().DateLinesZh()
	if len(zh) != 2 || zh[0] != "四月一日" || zh[1] != "西元 1100" {
		t.Fatalf("中文日期框是 %q", zh)
	}
	last := &Clock{Month: 11, Day: 30, Year: 1}
	if got := last.DateLinesZh(); got[0] != "十二月三十一日" || got[1] != "西元 1101" {
		t.Fatalf("年末中文日期框是 %q", got)
	}
}
