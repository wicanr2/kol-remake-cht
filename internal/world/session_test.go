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
