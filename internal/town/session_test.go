package town

import (
	"image"
	"image/color"
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
	out := &Outdoor{Town: 7, Icons: []int{DoorIcon}}
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

// 走到地圖邊界隊伍停住，並跳出城門提示：地圖視窗換林地圖、圖示列剩 DOOR 與 U-TURN。
func TestWalkStopsAtMapEdge(t *testing.T) {
	s := playSession(t)
	s.Outdoor.Camera.X, s.Outdoor.Camera.Y = 0, 0
	s.Outdoor.Marker.Row, s.Outdoor.Marker.Col = 0, 0
	if s.Walk(-1, 0) {
		t.Fatal("已經在左邊界，不該再往左")
	}
	if s.Mode != ModeGate {
		t.Fatal("撞到地圖邊界應該跳出城門提示")
	}
	icons, cursor := s.Icons()
	if len(icons) != 2 || icons[0] != DoorIcon || icons[1] != UTurnIcon || cursor != 0 {
		t.Fatalf("城門提示的圖示列是 %v／白框 %d", icons, cursor)
	}
	if x, _ := s.Outdoor.PartyPosition(); x != 0 {
		t.Fatalf("被夾限的是鏡頭，隊伍不該移動，卻在 x=%d", x)
	}
	s.MoveCursor(1)
	s.Accept() // U-TURN：回地圖
	if s.Mode != ModeMap || s.Outdoor.AtEdge {
		t.Fatal("選 U-TURN 應該回到地圖")
	}
	if !s.Walk(0, -1) && s.Mode != ModeGate {
		t.Fatal("上邊界也要跳出城門提示")
	}
	s.Accept() // DOOR：離開城鎮
	if !s.LeftTown {
		t.Fatal("在城門選 DOOR 應該離開城鎮")
	}
	if s.LeftGate.X != 0 || s.LeftGate.Y != 0 {
		t.Fatalf("離開的城門座標是 (%d,%d)", s.LeftGate.X, s.LeftGate.Y)
	}
}

// 鏡頭在邊界時隊伍還能往前站一格：原版夾限的是鏡頭，隊伍位置是鏡頭加偏移。
func TestPartyReachesOneCellPastCameraBound(t *testing.T) {
	s := playSession(t)
	s.Outdoor.Bounds.MaxX = 58
	s.Outdoor.Camera.X, s.Outdoor.Camera.Y = 58, 8
	s.Outdoor.Marker.Row, s.Outdoor.Marker.Col = 0, 0
	if !s.Walk(1, 0) {
		t.Fatal("偏移還是 0，第一步應該只動標記")
	}
	if x, _ := s.Outdoor.PartyPosition(); x != 59 {
		t.Fatalf("隊伍應該站到 59，實際 %d", x)
	}
	if s.Walk(1, 0) {
		t.Fatal("偏移已滿，第二步該被鏡頭夾限擋下")
	}
	if s.Mode != ModeGate {
		t.Fatal("鏡頭被夾限就要跳出城門提示")
	}
}

// 不可通行的磚塊擋住移動，而且連邊界判定都不做——原版是直接跳過整段移動。
func TestBlockedTileStopsMovement(t *testing.T) {
	s := playSession(t)
	img := image.NewGray(image.Rect(0, 0, GridSide, GridSide))
	img.SetGray(56, 8, color.Gray{Y: 9}) // 隊伍左邊那一格換成磚塊 9
	g, err := LoadGrid(img)
	if err != nil {
		t.Fatalf("LoadGrid：%v", err)
	}
	s.Outdoor.SetGrid(g)
	s.Buildings.Blocked[9] = true

	if s.Walk(-1, 0) {
		t.Fatal("左邊是不可通行的磚塊，不該走過去")
	}
	if x, _ := s.Outdoor.PartyPosition(); x != 57 {
		t.Fatalf("撞牆之後隊伍應該留在 57，實際 %d", x)
	}
	if s.Mode != ModeMap {
		t.Fatal("撞牆不是走到地圖邊界，不該跳出城門提示")
	}
	if !s.Walk(1, 0) {
		t.Fatal("右邊沒擋，應該走得過去")
	}
}

// 圖示列的生命週期：走離門口就清空、踩到門重建成兩格、選 U-TURN 之後只剩一格 DOOR。
// 三個狀態都由原版 town-map-edge-run1 的畫格證實（frame-070／074／122）。
func TestIconBarLifecycle(t *testing.T) {
	s := playSession(t)
	if icons, _ := s.Icons(); len(icons) != 1 || icons[0] != DoorIcon {
		t.Fatalf("剛從建築物出來時圖示列應該只有 DOOR，實際 %v", icons)
	}
	s.Walk(-1, 0)
	if icons, _ := s.Icons(); len(icons) != 0 {
		t.Fatalf("走離門口之後圖示列該清空，實際 %v", icons)
	}
	s.Walk(1, 0)
	icons, cursor := s.Icons()
	if len(icons) != 2 || icons[0] != DoorIcon || icons[1] != UTurnIcon || cursor != 0 {
		t.Fatalf("踩回門口的圖示列是 %v／白框 %d", icons, cursor)
	}
	s.MoveCursor(1)
	s.Accept() // U-TURN
	if icons, cursor := s.Icons(); len(icons) != 1 || icons[0] != DoorIcon || cursor != 0 {
		t.Fatalf("選 U-TURN 之後的圖示列是 %v／白框 %d", icons, cursor)
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

// 從野外進城：鏡頭、夾限、入城點都換成世界地圖那筆記錄，標記偏移歸零，
// 圖示列清空，模式回到地圖（RE-069，`sub_29CA3`）。
func TestEnterFromWorldResetsOutdoor(t *testing.T) {
	s := playSession(t)
	s.Walk(1, 0) // 先動一下，讓標記偏移不是 0
	info := &TownBuildings{Town: 3, Doors: []Door{{Index: 0, X: 9, Y: 9}}}
	s.EnterFromWorld(TownEntry{
		Town: 3, Camera: Point{X: 25, Y: 62},
		MinX: 0, MaxX: GridSide - 1, MinY: 0, MaxY: GridSide - 1,
		EntryPoint: 0,
	}, info, playGrid(t))

	o := s.Outdoor
	if o.Town != 3 || o.Camera.X != 25 || o.Camera.Y != 62 {
		t.Fatalf("進城後是城鎮 %d、鏡頭 (%d,%d)", o.Town, o.Camera.X, o.Camera.Y)
	}
	if o.Marker.Row != 0 || o.Marker.Col != 0 {
		t.Fatalf("標記偏移是 (%d,%d)，應歸零", o.Marker.Row, o.Marker.Col)
	}
	if x, y := o.PartyPosition(); x != 25 || y != 62 {
		t.Fatalf("隊伍落在 (%d,%d)，應為鏡頭那一格", x, y)
	}
	if s.Mode != ModeMap || len(o.Icons) != 0 || s.LeftTown {
		t.Fatalf("進城後模式 %v、圖示 %v、LeftTown %v", s.Mode, o.Icons, s.LeftTown)
	}
	if s.Buildings != info || o.CrestArchive() != "TWN3PICS" {
		t.Fatalf("建築資料或素材沒有換成城鎮 3：%s", o.CrestArchive())
	}
}
