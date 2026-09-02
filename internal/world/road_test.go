package world

import "testing"

// roadSession 建一個腳下有道路的野外狀態。tiles 是「地圖座標 → 磚塊編號」，
// masks 是「磚塊編號 → 道路方向遮罩」。
func roadSession(t *testing.T, tiles map[[2]int]int, masks map[int]int) *Session {
	t.Helper()
	data := testData(t)
	data.towns = nil
	for tile, mask := range masks {
		data.Road[tile] = mask
		data.Terrain[tile] = 4
	}
	return NewSession(&Screen{Travel: New(data, testGrid(t, tiles), 20, 20)})
}

// 第一次按 ROAD 時還沒有方向：原版由方向 7 逆時針掃一圈，取第一條有路的。
func TestFirstRoadPickPicksHighestDirection(t *testing.T) {
	// 腳下那一格是 (20,21)；遮罩 0x22 ＝ 方向 2（東）與方向 6（西）。
	s := roadSession(t, map[[2]int]int{{20, 21}: 3}, map[int]int{3: 0x22})
	if !s.FollowRoad(false) {
		t.Fatal("腳下有路，應該走得動")
	}
	if s.Direction != 6 {
		t.Fatalf("挑到方向 %d，由 7 往回掃該先遇到 6（西）", s.Direction)
	}
	if s.Screen.Travel.SubX != 2 {
		t.Fatalf("往西一步之後次格 X 是 %d", s.Screen.Travel.SubX)
	}
}

// 目前方向還有路就沿用，不重挑。
func TestRoadKeepsCurrentDirection(t *testing.T) {
	s := roadSession(t, map[[2]int]int{{20, 21}: 3}, map[int]int{3: 0x22})
	s.Direction = 2 // 東
	if !s.FollowRoad(false) || s.Direction != 2 {
		t.Fatalf("方向變成 %d，該留在 2（東）", s.Direction)
	}
	if s.Screen.Travel.SubX != 4 {
		t.Fatalf("往東一步之後次格 X 是 %d", s.Screen.Travel.SubX)
	}
}

// 目前方向沒路時往兩側各找三格，先找逆時針那一側。
func TestRoadSearchesNeighbouringDirections(t *testing.T) {
	// 遮罩 0x08 ＝ 只有方向 4（南）。
	s := roadSession(t, map[[2]int]int{{20, 21}: 3}, map[int]int{3: 0x08})
	s.Direction = 2 // 東：沒有路
	if !s.FollowRoad(false) {
		t.Fatal("南邊有路，應該轉過去")
	}
	if s.Direction != 4 {
		t.Fatalf("轉到方向 %d，應為 4（南）", s.Direction)
	}
}

// SIGNPOST 排除目前方向，挑另一條路。
func TestSignpostPicksAnotherRoad(t *testing.T) {
	s := roadSession(t, map[[2]int]int{{20, 21}: 3}, map[int]int{3: 0x22})
	s.Direction = 2 // 東
	if !s.FollowRoad(true) {
		t.Fatal("還有另一條路，應該轉過去")
	}
	if s.Direction != 6 || s.Excluded != 2 {
		t.Fatalf("轉到方向 %d、排除 %d，應為 6 與 2", s.Direction, s.Excluded)
	}
}

// 腳下沒有路就什麼都不做。
func TestRoadWithoutRoadDoesNothing(t *testing.T) {
	s := roadSession(t, nil, nil)
	if s.FollowRoad(false) {
		t.Fatal("腳下沒有路，不該走")
	}
	if s.Direction != NoDirection {
		t.Fatalf("方向變成 %d", s.Direction)
	}
}

// 走正東西南北時，隊伍偏離格子中線就改走斜向拉回來。
func TestRoadStepCentresParty(t *testing.T) {
	s := roadSession(t, map[[2]int]int{{20, 21}: 3}, map[int]int{3: 0x20})
	s.Direction = 2 // 東
	s.Screen.Travel.SubY = 6
	if dx, dy := s.roadStep(); dx != 1 || dy != -1 {
		t.Fatalf("次格 Y 偏下時該走東北，實際 (%d,%d)", dx, dy)
	}
	s.Screen.Travel.SubY = 1
	if dx, dy := s.roadStep(); dx != 1 || dy != 1 {
		t.Fatalf("次格 Y 偏上時該走東南，實際 (%d,%d)", dx, dy)
	}
	s.Screen.Travel.SubY = 3
	if dx, dy := s.roadStep(); dx != 1 || dy != 0 {
		t.Fatalf("在中線上時該走正東，實際 (%d,%d)", dx, dy)
	}
}

// 循路那一步只准落在道路磚塊上，踏出去就整個退回。
func TestRoadStepRefusesToLeaveTheRoad(t *testing.T) {
	// 腳下 (20,21) 是路；往東跨格之後腳下會變成 (21,21)，那裡不是路。
	s := roadSession(t, map[[2]int]int{{20, 21}: 3, {21, 21}: 0}, map[int]int{3: 0x20})
	s.Direction = 2
	tr := s.Screen.Travel
	tr.SubX, tr.SubY = 7, 3
	if s.FollowRoad(false) {
		t.Fatal("東邊那一格不是路，這一步該被退回")
	}
	if tr.X != 20 || tr.SubX != 7 {
		t.Fatalf("退回之後是 (%d, 次格 %d)，應為 (20, 7)", tr.X, tr.SubX)
	}
}
