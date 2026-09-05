package world

// 野外的移動（RE-068／RE-069）。原版 `sub_29004:0x291F0` 起的順序：
//
//  1. 先存下移動前的世界座標與次格座標。
//  2. 次格座標往那個方向走一格；走出 0…7 就繞回，並讓世界座標走 `Step`。
//  3. 世界座標夾限在 0…63。
//  4. **只有世界座標真的變了**（`word_3AC5A != 0`）才檢查地形：目的地磚塊
//     不可通行時，世界座標與次格座標一起退回移動前的值（`0x2958F`）。
//     格內的次格移動不受地形限制——`0x293B7` 的分支直接跳過整段檢查。

// Travel 是野外地圖上的一次旅行狀態。
type Travel struct {
	Data *Data
	Grid *Grid

	// X／Y 是世界座標；SubX／SubY 是格內的次格座標（0…7）。
	X, Y       int
	SubX, SubY int
	// Step 是次格繞回時世界座標走幾格。原版初值 1，選 ROAD 之後會變成 10。
	Step int
}

// New 建立一次野外旅行。次格座標的起點與原版相同，都是 (3, 3)。
func New(data *Data, grid *Grid, x, y int) *Travel {
	return &Travel{Data: data, Grid: grid, X: x, Y: y,
		SubX: SubCells/2 - 1, SubY: SubCells/2 - 1, Step: 1}
}

// Tile 是隊伍腳下那一格的磚塊編號。
//
// 原版讀的是「畫出來的那一格」：`byte_30941` 就是視窗緩衝區的第 0x29 個 byte，
// 也就是第 3 列第 5 欄（0x29 ＝ 3×12＋5）。繪製迴圈與城鎮共用，列偏移固定 −2，
// 所以第 3 列畫的是 Y＋1——隊伍腳下那一格在地圖上比鏡頭座標低一列（RE-070）。
// 地形名稱、道路旗標與地形擋路全部走這一格。
func (t *Travel) Tile() int { return t.Grid.At(t.X, t.Y+RowBias+PartyCell.Row) }

// Terrain 是腳下那一格的地形名稱。
func (t *Travel) Terrain() string { return t.Data.TerrainName(t.Tile()) }

// OnRoad 回傳隊伍是不是站在路上。
func (t *Travel) OnRoad() bool { return t.Data.OnRoad(t.Tile()) }

// TownHere 回傳腳下的城鎮入口；沒有就回 nil。
func (t *Travel) TownHere() *Town { return t.Data.TownAt(t.X, t.Y) }

// Walk 依方向鍵走一步。回傳是否真的動了——撞到不可通行的磚塊會原地不動。
func (t *Travel) Walk(dx, dy int) bool { return t.walk(dx, dy, false) }

// WalkRoad 是 ROAD／SIGNPOST 模擬出來的那一步。原版在那一步設 `word_37D28`，
// 目的地不是道路磚塊就把整個動作退回（`sub_29004:0x2930C`）。
func (t *Travel) WalkRoad(dx, dy int) bool { return t.walk(dx, dy, true) }

func (t *Travel) walk(dx, dy int, roadOnly bool) bool {
	if dx == 0 && dy == 0 {
		return false
	}
	x, y, sx, sy := t.X, t.Y, t.SubX, t.SubY
	t.X, t.SubX = step(t.X, t.SubX, dx, t.Step)
	t.Y, t.SubY = step(t.Y, t.SubY, dy, t.Step)
	t.X = clamp(t.X, 0, GridSide-1)
	t.Y = clamp(t.Y, 0, GridSide-1)
	if t.X == x && t.Y == y {
		return true
	}
	// 擋路查的也是那一格：原版讀 `viewbuf[0x29 + dx + 12*dy]`，基底同樣是第 3 列。
	if roadOnly && !t.Data.OnRoad(t.Tile()) {
		t.X, t.Y, t.SubX, t.SubY = x, y, sx, sy
		return false
	}
	if !t.Data.Passable(t.Tile()) {
		t.X, t.Y, t.SubX, t.SubY = x, y, sx, sy
		return false
	}
	return true
}

// step 次格走一格；走出 0…SubCells-1 就繞回，並讓世界座標走 stride 格。
func step(cell, sub, delta, stride int) (int, int) {
	switch {
	case delta < 0:
		sub--
		if sub < 0 {
			sub = SubCells - 1
			cell -= stride
		}
	case delta > 0:
		sub++
		if sub >= SubCells {
			sub = 0
			cell += stride
		}
	}
	return cell, sub
}

// MarkerPixel 回傳隊伍標記在地圖視窗內容區裡的像素座標（左上角為 0,0）。
func (t *Travel) MarkerPixel(tileSide int) (x, y int) {
	return PartyCell.Col*tileSide + t.SubX*SubPixels,
		PartyCell.Row*tileSide + t.SubY*SubPixels
}

// Camera 是畫在視窗第 2 列第 5 欄那一格的世界座標。原版的繪製迴圈是
// 共用的，偏移固定 −5／−2，而野外把隊伍畫在第 3 列，因此鏡頭正是隊伍那一格。
func (t *Travel) Camera() (x, y int) { return t.X, t.Y }
