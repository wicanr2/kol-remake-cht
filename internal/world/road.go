package world

// 道路的自動循路（RE-070）。`payload3[0x200 + 磚塊]` 不是布林值，是**八個方向的
// 位元遮罩**：方向 0 是正北，順時針數到 7 是西北，方向 d 對應位元 `1 << (7 - d)`
// （`byte_1312` 那張表就是 80h、40h…01h）。
//
// ROAD（圖示 17h）與 SIGNPOST（18h）都走 `sub_29DDE`，回傳一個模擬的方向鍵碼，
// 由主迴圈當成玩家按鍵處理；`word_37D28` 在那一步設起來，踏出道路就退回。

// Directions 是八個方向的位移，順序與原版的 `word_12F2`／`word_1302` 相同。
var Directions = [8]struct{ DX, DY int }{
	{0, -1}, {1, -1}, {1, 0}, {1, 1}, {0, 1}, {-1, 1}, {-1, 0}, {-1, -1},
}

// NoDirection 是原版的 0FFh：目前沒有在循任何一條路。
const NoDirection = -1

// DirectionMask 是方向在道路位元遮罩裡的位元。
func DirectionMask(dir int) int { return 1 << (7 - uint(dir)) }

// RoadMask 回傳腳下那一格的道路方向遮罩（原版的 `word_2FA2A`）。
func (t *Travel) RoadMask() int {
	tile := t.Tile()
	if tile < 0 || tile >= len(t.Data.Road) {
		return 0
	}
	return t.Data.Road[tile]
}

// hasRoad 回傳某個方向在腳下這一格有沒有路。
func (t *Travel) hasRoad(dir int) bool {
	return dir >= 0 && dir < 8 && t.RoadMask()&DirectionMask(dir) != 0
}

// pickRoad 挑出這一步要走的方向，並更新 `Direction`／`Excluded`。
//
//   - ROAD（other 為 false）：目前方向還有路就沿用；沒有就往兩側各找三格。
//   - SIGNPOST（other 為 true）：先把目前方向記成排除項，再從那裡逆時針掃一圈，
//     找**另一條**路。
//
// 找不到路時 `Direction` 回到 `NoDirection`，回傳 false。
func (s *Session) pickRoad(other bool) bool {
	t := s.Screen.Travel
	if !other && s.Direction != NoDirection {
		if t.hasRoad(s.Direction) {
			return true
		}
		base := s.Direction
		for step := 1; step < 4; step++ {
			if s.Direction = (base - step + 8) % 8; t.hasRoad(s.Direction) {
				return true
			}
			if s.Direction = (base + step) % 8; t.hasRoad(s.Direction) {
				return true
			}
		}
		s.Excluded, s.Direction = NoDirection, NoDirection
	}
	if other {
		s.Excluded = s.Direction
	}
	s.Direction = NoDirection
	start := 7
	if s.Excluded >= 0 && s.Excluded < 8 {
		start = s.Excluded
	}
	for dir, first := start, true; first || dir != start; dir = (dir + 7) % 8 {
		first = false
		if dir != s.Excluded && t.hasRoad(dir) {
			s.Direction = dir
			return true
		}
	}
	return false
}

// roadStep 是這一步真正要走的方向。走正東西南北時，隊伍如果不在格子中線上，
// 原版改走一個斜向把它拉回中線——這就是 `sub_29DDE` 尾端那四組比較。
func (s *Session) roadStep() (dx, dy int) {
	t := s.Screen.Travel
	dir := s.Direction
	switch dir {
	case 0: // 北
		dir = centre(t.SubX, 7, 1, dir)
	case 2: // 東
		dir = centre(t.SubY, 1, 3, dir)
	case 4: // 南
		dir = centre(t.SubX, 5, 3, dir)
	case 6: // 西
		dir = centre(t.SubY, 7, 5, dir)
	}
	return Directions[dir].DX, Directions[dir].DY
}

// centre 依次格座標偏中線的哪一側，換成 high／low 這兩個斜向之一。
// 原版的判斷是「> 4」與「< 3」，也就是次格 3、4 視為已經在中線上。
func centre(sub, high, low, dir int) int {
	switch {
	case sub > 4:
		return high
	case sub < 3:
		return low
	}
	return dir
}

// FollowRoad 是 ROAD 與 SIGNPOST 兩個圖示：挑方向、走一步。
// 那一步只准落在道路磚塊上，否則整個動作退回（`word_37D28`）。
func (s *Session) FollowRoad(other bool) bool {
	if s.Mode != ModeMap || !s.pickRoad(other) {
		return false
	}
	dx, dy := s.roadStep()
	t := s.Screen.Travel
	x, y := t.X, t.Y
	moved := t.WalkRoad(dx, dy)
	if t.X != x || t.Y != y {
		s.Sync()
		if s.Mode == ModeTown {
			return moved
		}
		s.Screen.SyncTerrain()
	}
	if moved {
		s.advanceTime()
	}
	return moved
}
