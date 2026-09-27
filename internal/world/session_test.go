package world

import "testing"

// 剛進野外時圖示列是空的，人就算站在城鎮那一格上也不會彈出提示。
func TestFreshWorldHasNoPrompt(t *testing.T) {
	data := testData(t)
	data.towns = []Town{{Record: 0, X: 20, Y: 20, Index: 10, Town: 1}}
	sess := NewSession(&Screen{Travel: New(data, testGrid(t, nil), 20, 20)})
	if sess.Mode != ModeMap || len(sess.Screen.Icons) != 0 {
		t.Fatalf("剛進野外的模式是 %v、圖示 %v", sess.Mode, sess.Screen.Icons)
	}
}

// 走進城鎮入口那一格就跳出提示：地圖視窗換成外觀圖，圖示列剩 DOOR 與 U-TURN。
func TestWalkingOntoTownShowsPrompt(t *testing.T) {
	data := testData(t)
	data.towns = []Town{{Record: 0, X: 21, Y: 20, Index: 10, Code: "BRT", Town: 1}}
	s := &Screen{Travel: New(data, testGrid(t, nil), 20, 20)}
	sess := NewSession(s)
	if sess.Mode != ModeMap || len(s.Icons) != 0 {
		t.Fatalf("起點不是城鎮入口，模式 %v、圖示 %v", sess.Mode, s.Icons)
	}
	for i := 0; i < 4; i++ {
		sess.Walk(1, 0)
		if sess.Mode != ModeMap {
			t.Fatalf("第 %d 步還在同一格，不該跳出提示", i+1)
		}
	}
	sess.Walk(1, 0)
	if sess.Mode != ModeTown {
		t.Fatalf("走到 (%d,%d) 之後模式是 %v", s.Travel.X, s.Travel.Y, sess.Mode)
	}
	if got := sess.TownPicture(); got != 18 {
		t.Fatalf("外觀圖是 %d，應為 10 + 8 = 18", got)
	}
	want := []int{IconDoor + 3, IconUTurn + 3}
	if len(s.Icons) != 2 || s.Icons[0] != want[0] || s.Icons[1] != want[1] {
		t.Fatalf("圖示列是 %v，應為 %v", s.Icons, want)
	}
	if sess.Walk(1, 0) {
		t.Fatal("提示裡的方向鍵不該有作用")
	}
}

// U-TURN 回地圖，圖示列只剩一格 DOOR；站在原地不會再跳出提示。
func TestTownPromptUTurn(t *testing.T) {
	data := testData(t)
	data.towns = []Town{{Record: 0, X: 21, Y: 20, Index: 10, Town: 1}}
	sess := NewSession(&Screen{Travel: New(data, testGrid(t, nil), 20, 20)})
	for i := 0; i < 5; i++ {
		sess.Walk(1, 0)
	}
	if sess.Mode != ModeTown {
		t.Fatalf("走到入口上，模式卻是 %v", sess.Mode)
	}
	sess.MoveCursor(1)
	sess.Accept()
	if sess.Mode != ModeMap {
		t.Fatalf("U-TURN 之後模式是 %v", sess.Mode)
	}
	if len(sess.Screen.Icons) != 1 || sess.Screen.Icons[0] != IconDoor+3 {
		t.Fatalf("U-TURN 之後圖示列是 %v", sess.Screen.Icons)
	}
	sess.Walk(1, 0)
	if sess.Mode != ModeMap {
		t.Fatal("格內走一步不該再跳出提示")
	}
}

// DOOR 交出腳下那一筆記錄；記錄 6 以後全部進 TOWN7，用入城點區分。
func TestTownPromptDoorEnters(t *testing.T) {
	data := testData(t)
	data.towns = []Town{{Record: 8, X: 21, Y: 20, Index: 2, Town: 7, EntryPoint: 2}}
	sess := NewSession(&Screen{Travel: New(data, testGrid(t, nil), 20, 20)})
	var got *Town
	sess.OnEnterTown = func(rec *Town) { got = rec }
	for i := 0; i < 5; i++ {
		sess.Walk(1, 0)
	}
	sess.Accept()
	if got == nil || got.Town != 7 || got.EntryPoint != 2 {
		t.Fatalf("進城的記錄是 %+v", got)
	}
	if sess.TownPicture() != 10 {
		t.Fatalf("礦坑的外觀圖是 %d，應為 2 + 8 = 10", sess.TownPicture())
	}
}

// 站在路上時圖示列多出 ROAD 與 SIGNPOST。
func TestRoadIcons(t *testing.T) {
	data := testData(t)
	data.towns = nil
	// 道路旗標看的是腳下那一格（Y＋1）。
	grid := testGrid(t, map[[2]int]int{{21, 21}: 5})
	sess := NewSession(&Screen{Travel: New(data, grid, 20, 20)})
	if len(sess.Screen.Icons) != 0 {
		t.Fatalf("空曠處的圖示列是 %v", sess.Screen.Icons)
	}
	for i := 0; i < 5; i++ {
		sess.Walk(1, 0)
	}
	want := []int{IconRoad + 3, IconSignpost + 3}
	if len(sess.Screen.Icons) != 2 ||
		sess.Screen.Icons[0] != want[0] || sess.Screen.Icons[1] != want[1] {
		t.Fatalf("路上的圖示列是 %v，應為 %v", sess.Screen.Icons, want)
	}
}

// 地形標籤在走到城鎮入口的那一步不更新：原版此時跳進 `sub_29BBB`，
// 更新標籤的那一段被整個略過（RE-070）。
func TestTerrainLabelStaysOnTownStep(t *testing.T) {
	data := testData(t)
	data.towns = []Town{{Record: 0, X: 21, Y: 20, Index: 10, Town: 1}}
	// (21,21) 是往東一格之後腳下那一格；它的地形碼是 0，標籤真的更新就會看得出來。
	grid := testGrid(t, map[[2]int]int{{21, 21}: 5})
	s := &Screen{Travel: New(data, grid, 20, 20)}
	sess := NewSession(s)
	if s.TerrainCode != 7 {
		t.Fatalf("起點地形碼是 %d，應為 7（PLAINS）", s.TerrainCode)
	}
	for i := 0; i < 5; i++ {
		sess.Walk(1, 0)
	}
	if sess.Mode != ModeTown {
		t.Fatalf("走到入口上，模式卻是 %v", sess.Mode)
	}
	if s.TerrainCode != 7 {
		t.Fatalf("走到入口那一步地形碼變成 %d，應該留在 7", s.TerrainCode)
	}
	if got := s.TerrainLine(); got != "PLAINS" {
		t.Fatalf("標籤是 %q", got)
	}
}

// stepRand 讓天候那一擲固定，方便驗時段變換的回呼。
type stepRand struct{}

func (stepRand) Roll(int) int { return 1 }

// 時段變了才擲野外遭遇：同一個時段裡走幾步都不會擲（`sub_29A3E`，RE-078）。
func TestOnTimePartFiresOnlyWhenThePartChanges(t *testing.T) {
	data := testData(t)
	c := NewClock()
	s := &Screen{Travel: New(data, testGrid(t, nil), 20, 20),
		Clock: c, TerrainCode: 7, Movement: 20}
	sess := NewSession(s)
	sess.Rand = stepRand{}
	calls := 0
	sess.OnTimePart = func() { calls++ }

	// 停在同一個時段裡：時刻扣掉一點點，時段不動。
	part := c.Part
	sess.spendTime(1)
	if c.Part != part {
		t.Fatalf("扣一個時刻不該換時段，得到 %d", c.Part)
	}
	if calls != 0 {
		t.Fatalf("時段沒變不該擲遭遇，卻擲了 %d 次", calls)
	}

	// 一次扣到下一個時段。
	for i := 0; i < 40 && c.Part == part; i++ {
		sess.spendTime(100)
	}
	if c.Part == part {
		t.Fatal("扣了那麼多時刻應該換時段了")
	}
	if calls != 1 {
		t.Fatalf("換一次時段該擲一次，卻擲了 %d 次", calls)
	}
}

// 遭遇那一頁：問句用是／否，答完由呼叫端決定下一步；只等一個鍵那一頁換成
// 空白鍵那一格，按下去就收掉（RE-112）。
func TestRideAwayPromptIcons(t *testing.T) {
	data := testData(t)
	s := &Screen{Travel: New(data, testGrid(t, nil), 20, 20)}
	sess := NewSession(s)
	sess.SetPrompt(&Prompt{Kind: PromptRideAway,
		Lines: []string{"An extremely pungent"}, Ask: "Attempt to ride away?"})
	want := []int{IconYes + 3, IconNo + 3}
	if len(s.Icons) != 2 || s.Icons[0] != want[0] || s.Icons[1] != want[1] {
		t.Fatalf("圖示列是 %v，應為 %v", s.Icons, want)
	}
	answers := []bool{}
	sess.OnPromptAnswer = func(yes bool) { answers = append(answers, yes) }
	sess.Accept()
	if len(answers) != 1 || !answers[0] {
		t.Fatalf("白框停在 YES，該答是，得到 %v", answers)
	}
	if sess.Prompt != nil {
		t.Fatal("答完該收掉那一頁")
	}
}

// 回呼在答完之後換上下一頁時，那一頁不該被跟著收掉。
func TestPromptAnswerCanChainNextPage(t *testing.T) {
	data := testData(t)
	s := &Screen{Travel: New(data, testGrid(t, nil), 20, 20)}
	sess := NewSession(s)
	sess.SetPrompt(&Prompt{Kind: PromptRideAway, Ask: "Attempt to ride away?"})
	sess.OnPromptAnswer = func(yes bool) {
		if yes {
			sess.SetPrompt(&Prompt{Kind: PromptNotice,
				Lines: []string{"You cannot ride away."}})
		}
	}
	sess.Accept()
	if sess.Prompt == nil || sess.Prompt.Kind != PromptNotice {
		t.Fatalf("接上來那一頁不見了：%+v", sess.Prompt)
	}
	if len(s.Icons) != 1 || s.Icons[0] != IconSpacebar+3 {
		t.Fatalf("圖示列是 %v，應為只有空白鍵那一格", s.Icons)
	}
	// 只等一個鍵：按下去收掉，並且回呼收到 false。
	got := []bool{}
	sess.OnPromptAnswer = func(yes bool) { got = append(got, yes) }
	sess.Accept()
	if sess.Prompt != nil {
		t.Fatalf("那一頁該收掉了：%+v", sess.Prompt)
	}
	if len(got) != 1 || got[0] {
		t.Fatalf("收掉時該傳 false，得到 %v", got)
	}
}

// 那一頁蓋著地圖時方向鍵沒有作用（原版等的是 `sub_144EF`）。
func TestPromptBlocksWalking(t *testing.T) {
	data := testData(t)
	s := &Screen{Travel: New(data, testGrid(t, nil), 20, 20)}
	sess := NewSession(s)
	sess.SetPrompt(&Prompt{Kind: PromptRideAway, Ask: "Attempt to ride away?"})
	if sess.Walk(1, 0) {
		t.Fatal("那一頁蓋著地圖時不該走得動")
	}
	if s.Travel.X != 20 || s.Travel.Y != 20 {
		t.Fatalf("座標動到 (%d,%d)", s.Travel.X, s.Travel.Y)
	}
}

// 時段換了先擲遭遇、再擲天候（`sub_29987`：`sub_2A4E5` 在 `0x29A83` 的天候之前，
// RE-151）。
func TestTimePartRollsEncounterBeforeWeather(t *testing.T) {
	data := testData(t)
	c := NewClock()
	s := &Screen{Travel: New(data, testGrid(t, nil), 20, 20),
		Clock: c, TerrainCode: 7, Movement: 20}
	sess := NewSession(s)
	rng := &orderRand{}
	sess.Rand = rng
	sess.OnTimePart = func() { rng.log = append(rng.log, "遭遇") }
	for i := 0; i < 40 && c.Part == 0; i++ {
		sess.spendTime(100)
	}
	if len(rng.log) < 2 || rng.log[0] != "遭遇" || rng.log[1] != "天候" {
		t.Fatalf("順序應為遭遇、天候，得到 %v", rng.log)
	}
}

// 換日那一步不擲；按鍵收掉 CAMPING! 之後才判時段（換成 DAWN）並擲遭遇
// （`sub_297F2` 換日分支在 `sub_144C3` 之後才呼叫 `sub_29987`）。
func TestNewDayPartChangesAfterCamping(t *testing.T) {
	data := testData(t)
	c := NewClock()
	c.Ticks, c.Part = 10, DayParts-1
	s := &Screen{Travel: New(data, testGrid(t, nil), 20, 20),
		Clock: c, TerrainCode: 7, Movement: 20}
	sess := NewSession(s)
	sess.Rand = stepRand{}
	calls := 0
	sess.OnTimePart = func() { calls++ }
	sess.spendTime(60)
	if !sess.Camping || calls != 0 || c.Part != DayParts-1 {
		t.Fatalf("換日那一步應停在 CAMPING!、還沒擲：camping=%v 擲 %d 次 時段 %d", sess.Camping, calls, c.Part)
	}
	sess.EndCamping()
	if calls != 1 || c.Part != 0 {
		t.Fatalf("收掉之後應換成 DAWN 並擲一次：擲 %d 次 時段 %d", calls, c.Part)
	}
}

// orderRand 記下天候那一擲（`Roll(0x80)`）發生在什麼時候。
type orderRand struct{ log []string }

func (r *orderRand) Roll(n int) int {
	if n == 0x80 {
		r.log = append(r.log, "天候")
	}
	return 1
}
