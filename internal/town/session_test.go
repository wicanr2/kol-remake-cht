package town

import (
	"image"
	"testing"
)

func playGrid(t *testing.T) *Grid {
	t.Helper()
	img := image.NewGray(image.Rect(0, 0, GridSide, GridSide))
	g, err := LoadGrid(img)
	if err != nil {
		t.Fatalf("LoadGrid：%v", err)
	}
	return g
}

func playSession(t *testing.T) *Session {
	t.Helper()
	out := &Outdoor{Town: 7}
	out.Camera.X, out.Camera.Y = 57, 8
	out.Bounds.MaxX, out.Bounds.MaxY = GridSide-1, GridSide-1
	out.SetGrid(playGrid(t))
	buildings := &TownBuildings{Town: 7, Doors: []Door{
		{Index: 5, X: 57, Y: 8, Picture: 13, Portrait: 12,
			Name: "Halfway House", Person: "Devin Torlock"},
	}}
	text := &Text{Towns: []townText{{Town: 7, Buildings: []BuildingText{{
		Door: 5, Messages: []Message{
			{Index: 0, Display: []string{"a1", "a2", "a3", "a4"}},
			{Index: 1, Display: []string{"b1", "b2", "b3", "b4"}},
		}}}}}}
	return NewSession(out, buildings, text, nil,
		[]Member{{Portrait: 47, Name: "Theodoric"}})
}

// 直接重播 town-outdoor-move-run1 量到的移動序列（RE-060）：
// 鏡頭與標記偏移每一步都要與原版相同。
func TestWalkReplaysOriginalCameraFollow(t *testing.T) {
	s := playSession(t)
	steps := []struct {
		dx, dy               int
		camX, camY, row, col int
	}{
		{1, 0, 57, 8, 0, 1},
		{1, 0, 58, 8, 0, 1},
		{0, 1, 58, 8, 1, 1},
		{0, 1, 58, 9, 1, 1},
		{0, -1, 58, 9, 0, 1},
	}
	for i, step := range steps {
		s.Walk(step.dx, step.dy)
		o := s.Outdoor
		if o.Camera.X != step.camX || o.Camera.Y != step.camY ||
			o.Marker.Row != step.row || o.Marker.Col != step.col {
			t.Fatalf("第 %d 步之後鏡頭 (%d,%d) 偏移 (%d,%d)，應為 (%d,%d)／(%d,%d)",
				i+1, o.Camera.X, o.Camera.Y, o.Marker.Row, o.Marker.Col,
				step.camX, step.camY, step.row, step.col)
		}
	}
}

// 走出去再走回來會踩到門，此時才進門口狀態——原版的掃門迴圈接在移動後面。
func TestSessionEnterAndLeaveBuilding(t *testing.T) {
	s := playSession(t)
	if s.Mode != ModeMap {
		t.Fatal("剛開始應該是地圖模式（人雖然站在門的格子上）")
	}

	s.Walk(-1, 0)
	if s.Mode != ModeMap {
		t.Fatal("走離門口仍是地圖模式")
	}
	s.Walk(1, 0)
	if s.Mode != ModeDoor {
		t.Fatalf("走回門口應該進門口狀態，取到 %v", s.Mode)
	}
	icons, cursor := s.Icons()
	if len(icons) != 2 || icons[0] != DoorIcon || icons[1] != UTurnIcon || cursor != 0 {
		t.Fatalf("門口圖示列是 %v／%d", icons, cursor)
	}

	// 白框移到 U-TURN 再按 Return：回到地圖，不進建築。
	s.MoveCursor(1)
	s.Accept()
	if s.Mode != ModeMap || s.Inside != nil {
		t.Fatal("選 U-TURN 應該回到地圖")
	}

	// 再走回門口，這次選 DOOR。
	s.Walk(-1, 0)
	s.Walk(1, 0)
	s.Accept()
	if s.Mode != ModeInside || s.Inside == nil {
		t.Fatal("選 DOOR 應該進建築")
	}
	if s.Inside.Title[0] != "Halfway House" {
		t.Fatalf("標題框是 %v", s.Inside.Title)
	}
	if len(s.Inside.Dialog) != 8 {
		t.Fatalf("進門敘述有 %d 行，應為 8", len(s.Inside.Dialog))
	}
	icons, _ = s.Icons()
	if len(icons) != 1 || icons[0] != BuildingIcon {
		t.Fatalf("建築內圖示列是 %v", icons)
	}

	// 出來之後人還站在門的格子上，畫面卻是地圖。
	s.Accept()
	if s.Mode != ModeMap || s.Inside != nil {
		t.Fatal("從建築出來應該回到地圖")
	}
	if x, y := s.Outdoor.PartyPosition(); x != 57 || y != 8 {
		t.Fatalf("出來之後隊伍在 (%d,%d)，應為 (57,8)", x, y)
	}
	if s.Outdoor.DoorUnderParty() == nil {
		t.Fatal("腳下還是那個門")
	}
}

// 走到地圖邊界會停住（離開城鎮的路徑還沒實作）。
func TestWalkStopsAtMapEdge(t *testing.T) {
	s := playSession(t)
	s.Outdoor.Camera.X, s.Outdoor.Camera.Y = 0, 0
	s.Outdoor.Marker.Row, s.Outdoor.Marker.Col = 0, 0
	if s.Walk(-1, 0) {
		t.Fatal("已經在左邊界，不該再往左")
	}
	if s.Walk(0, -1) {
		t.Fatal("已經在上邊界，不該再往上")
	}
}

// 非地圖模式時方向鍵沒有作用；圖示白框只在門口狀態會動。
func TestInputIsModeScoped(t *testing.T) {
	s := playSession(t)
	s.Walk(-1, 0)
	s.Walk(1, 0)
	if s.Mode != ModeDoor {
		t.Fatal("應該在門口狀態")
	}
	before, _ := s.Outdoor.PartyPosition()
	if s.Walk(1, 0) {
		t.Fatal("門口狀態不該吃方向鍵")
	}
	if after, _ := s.Outdoor.PartyPosition(); after != before {
		t.Fatal("門口狀態下隊伍不該移動")
	}
	s.Accept()
	s.MoveCursor(1)
	if _, cursor := s.Icons(); cursor != 0 {
		t.Fatal("建築內不該移動圖示白框")
	}
}
