package battle

import "testing"

// 這一份地形是任務地點 5 的戰場（`TOWN0` payload 0x3D+5）在隊伍那一格附近
// 9×9 格的原樣，列是 10…18、欄是 37…45。隊伍站在 (14, 41)。
var siteFiveWindow = [9][9]int{
	{9, 9, 9, 9, 9, 14, 9, 9, 9},
	{21, 9, 9, 9, 14, 14, 9, 9, 9},
	{16, 9, 9, 9, 14, 9, 9, 9, 9},
	{16, 9, 9, 9, 14, 9, 9, 9, 9},
	{16, 9, 9, 9, 14, 9, 9, 9, 9},
	{11, 21, 9, 9, 14, 9, 9, 9, 9},
	{11, 16, 45, 40, 14, 43, 39, 9, 9},
	{11, 11, 21, 41, 36, 41, 9, 9, 9},
	{22, 11, 11, 41, 36, 41, 9, 9, 9},
}

func siteFiveField() *Field {
	f := NewField()
	for r, line := range siteFiveWindow {
		for c, tile := range line {
			f.Tiles[(10+r)*FieldSide+37+c] = tile
		}
	}
	return f
}

// 原版 `world-quest-run1` 的 frame-886：視窗最下面兩列只看得到隊伍那一欄，
// 左右都被 (16, 39)…(16, 43) 那一排石板擋住。
func TestVisibleMatchesOriginalFrame(t *testing.T) {
	sight := siteFiveField().Visible(14, 41)
	for row := 0; row < ViewRows; row++ {
		for col := 0; col < ViewCols; col++ {
			want := true
			switch {
			case row >= 7:
				want = col == ViewCentre.Col
			case row == ViewCentre.Row && col == ViewCentre.Col:
				// 中心格原版另外畫（`sub_223C0:0x2278A` 起），視線本身不標記。
				want = false
			}
			if got := sight[row][col]; got != want {
				t.Errorf("視窗 (%d,%d) 可見度 %v，預期 %v", row, col, got, want)
			}
		}
	}
}

// 位元圖直接對幾塊已知的磚：草地與石板路不擋，牆與碎石擋。
func TestBlocksKnownTiles(t *testing.T) {
	for tile, want := range map[int]bool{
		TileGround: false, TileScatter: false, 14: false, 0x24: false,
		41: true, 45: true, 40: true, TileUnseen: true,
	} {
		if got := Blocks(tile); got != want {
			t.Errorf("磚 %d 擋視線 %v，預期 %v", tile, got, want)
		}
	}
}

// 站在路面上時，視線的第一步穿得過 0x23…0x2D 那一段。
func TestSeesOverFromRoad(t *testing.T) {
	if !seesOver(41, RoadTileA, 0) {
		t.Error("站在 0x24 上、第一步應該看得過磚 41")
	}
	if seesOver(41, 14, 0) {
		t.Error("站在磚 14 上不該看得過磚 41")
	}
	if seesOver(41, RoadTileA, 1) {
		t.Error("只有第一步能看過去")
	}
}
