package world

import (
	"image"
	"testing"
)

func testGrid(t *testing.T, blocked map[[2]int]int) *Grid {
	t.Helper()
	img := image.NewGray(image.Rect(0, 0, GridSide, GridSide))
	for p, v := range blocked {
		img.Pix[p[1]*GridSide+p[0]] = uint8(v)
	}
	g, err := LoadGrid(img)
	if err != nil {
		t.Fatalf("LoadGrid：%v", err)
	}
	return g
}

func testData(t *testing.T) *Data {
	t.Helper()
	d := &Data{Schema: "kol-world-map-data-v1",
		Terrain: make([]int, 256), Road: make([]int, 256)}
	d.Terrain[0] = 7 // PLAINS
	d.Terrain[9] = 1 // RIVER
	d.Road[5] = 1
	d.blocked[9] = true
	d.towns = []Town{{Record: 0, X: 20, Y: 20, Code: "BRT", Town: 1}}
	return d
}

// 走八步才跨一格：次格座標在 0…7 之間走，繞回時世界座標才動。
func TestSubCellWrapsIntoWorldStep(t *testing.T) {
	tr := New(testData(t), testGrid(t, nil), 20, 20)
	if tr.SubX != 3 || tr.SubY != 3 {
		t.Fatalf("次格起點是 (%d,%d)", tr.SubX, tr.SubY)
	}
	// 次格從 3 起算，走到 7 還在同一格；第 5 步才繞回 0 並讓世界座標往東一格。
	for i := 0; i < 5; i++ {
		if !tr.Walk(1, 0) {
			t.Fatalf("第 %d 步該走得動", i+1)
		}
	}
	if tr.X != 21 || tr.SubX != 0 {
		t.Fatalf("往東五步之後是 (%d, 次格 %d)，應為 (21, 0)", tr.X, tr.SubX)
	}
	for i := 0; i < 4; i++ {
		tr.Walk(-1, 0)
	}
	if tr.X != 20 || tr.SubX != 4 {
		t.Fatalf("再往西四步之後是 (%d, 次格 %d)，應為 (20, 4)", tr.X, tr.SubX)
	}
}

// 地形只擋跨格的那一步：格內的次格移動照走，繞回到不可通行的格子時整個動作退回。
func TestBlockedTileStopsTravel(t *testing.T) {
	// 腳下那一格是鏡頭座標的下一列（Y＋1），所以擋路的河要放在 (21, 21)。
	grid := testGrid(t, map[[2]int]int{{21, 21}: 9})
	tr := New(testData(t), grid, 20, 20)
	for i := 0; i < 4; i++ {
		if !tr.Walk(1, 0) {
			t.Fatalf("第 %d 步還在格子裡，不該被地形擋住", i+1)
		}
	}
	if tr.X != 20 || tr.SubX != 7 {
		t.Fatalf("往東四步之後是 (%d, 次格 %d)，應為 (20, 7)", tr.X, tr.SubX)
	}
	if tr.Walk(1, 0) {
		t.Fatal("東邊是河，繞回的那一步不該走過去")
	}
	if tr.X != 20 || tr.SubX != 7 {
		t.Fatalf("撞牆之後是 (%d, 次格 %d)，應該退回 (20, 7)", tr.X, tr.SubX)
	}
	if !tr.Walk(0, 1) {
		t.Fatal("南邊沒擋，應該走得動")
	}
}

// 地形與道路看的是腳下那一格（Y＋1），城鎮入口看的是鏡頭座標本身。
func TestTerrainRoadAndTown(t *testing.T) {
	grid := testGrid(t, map[[2]int]int{{20, 22}: 5, {20, 23}: 9})
	tr := New(testData(t), grid, 20, 20)
	if got := tr.Terrain(); got != "PLAINS" {
		t.Fatalf("腳下地形是 %q", got)
	}
	if tr.OnRoad() {
		t.Fatal("平原不該算在路上")
	}
	if town := tr.TownHere(); town == nil || town.Code != "BRT" {
		t.Fatalf("腳下該是 BRT，實際 %v", town)
	}
	tr.Y, tr.SubY = 21, 3
	if !tr.OnRoad() {
		t.Fatal("磚塊 5 有道路旗標")
	}
	tr.Y = 22
	if got := tr.Terrain(); got != "RIVER" {
		t.Fatalf("磚塊 9 的地形是 %q", got)
	}
}

// 標記畫在視窗第 3 列第 5 欄那一格內，偏移是次格座標的兩倍像素。
func TestMarkerPixel(t *testing.T) {
	tr := New(testData(t), testGrid(t, nil), 20, 20)
	x, y := tr.MarkerPixel(16)
	if x != 5*16+6 || y != 3*16+6 {
		t.Fatalf("標記像素是 (%d,%d)", x, y)
	}
	tr.Walk(0, -1)
	if _, y = tr.MarkerPixel(16); y != 3*16+4 {
		t.Fatalf("往北一步之後標記的 y 是 %d", y)
	}
}

// 隊伍那個點是把底下四個像素的 EGA 索引與 15 互斥或，不是塗白。
// 原版畫格量到的兩組對應：0 → 15、8 → 7。
func TestMarkerInvertsEGAIndex(t *testing.T) {
	cases := []struct{ in, want int }{{0, 15}, {8, 7}, {7, 8}, {15, 0}}
	for _, c := range cases {
		got := EGAIndex(InvertEGA(egaPalette[c.in]))
		if got != c.want {
			t.Fatalf("索引 %d 反相之後是 %d，應為 %d", c.in, got, c.want)
		}
	}
}
