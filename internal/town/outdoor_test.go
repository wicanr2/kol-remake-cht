package town

import (
	"image"
	"image/color"
	"testing"
)

// 造一個格子陣列：值等於 (x*3 + y) mod 256，方便驗證索引與夾限。
func testGrid(t *testing.T) *Grid {
	t.Helper()
	img := image.NewGray(image.Rect(0, 0, GridSide, GridSide))
	for y := 0; y < GridSide; y++ {
		for x := 0; x < GridSide; x++ {
			img.SetGray(x, y, color.Gray{Y: uint8((x*3 + y) % 256)})
		}
	}
	g, err := LoadGrid(img)
	if err != nil {
		t.Fatalf("LoadGrid：%v", err)
	}
	return g
}

func newOutdoor(t *testing.T, camX, camY int) *Outdoor {
	t.Helper()
	o := &Outdoor{}
	o.Camera.X, o.Camera.Y = camX, camY
	o.Bounds.MaxX, o.Bounds.MaxY = GridSide-1, GridSide-1
	o.SetGrid(testGrid(t))
	return o
}

func TestViewOffsetsMatchOriginalLoop(t *testing.T) {
	o := newOutdoor(t, 20, 20)
	view := o.View()
	if len(view) != Viewport.Rows || len(view[0]) != Viewport.Cols {
		t.Fatalf("視窗尺寸 %d×%d 不是 %d×%d", len(view), len(view[0]),
			Viewport.Rows, Viewport.Cols)
	}
	for row := 0; row < Viewport.Rows; row++ {
		for col := 0; col < Viewport.Cols; col++ {
			x := 20 + col + ColBias
			y := 20 + row + RowBias
			if got, want := view[row][col], (x*3+y)%256; got != want {
				t.Fatalf("(列 %d, 欄 %d) 取到 %d，應為 %d", row, col, got, want)
			}
		}
	}
}

// 隊伍那一格永遠是鏡頭座標本身：欄偏移 −5 抵銷第 5 欄、列偏移 −2 抵銷第 2 列。
func TestMarkerCellIsCameraCell(t *testing.T) {
	o := newOutdoor(t, 31, 12)
	row, col := o.MarkerCell()
	if row != 2 || col != 5 {
		t.Fatalf("標記格是 (%d, %d)，應為 (2, 5)", row, col)
	}
	if got, want := o.View()[row][col], (31*3+12)%256; got != want {
		t.Fatalf("標記格的磚塊是 %d，應為鏡頭座標那一格的 %d", got, want)
	}
	x, y := o.PartyPosition()
	if x != 31 || y != 12 {
		t.Fatalf("隊伍座標 (%d, %d)，應為 (31, 12)", x, y)
	}
}

// 標記可以在視窗內偏移；此時隊伍座標是鏡頭加偏移，鏡頭本身不動。
func TestMarkerOffsetMovesPartyNotCamera(t *testing.T) {
	o := newOutdoor(t, 58, 8)
	o.Marker.Row, o.Marker.Col = 1, 1
	row, col := o.MarkerCell()
	if row != 3 || col != 6 {
		t.Fatalf("標記格是 (%d, %d)，應為 (3, 6)", row, col)
	}
	x, y := o.PartyPosition()
	if x != 59 || y != 9 {
		t.Fatalf("隊伍座標 (%d, %d)，應為 (59, 9)", x, y)
	}
	if o.Camera.X != 58 || o.Camera.Y != 8 {
		t.Fatal("標記偏移不該動到鏡頭")
	}
}

func TestViewClampsAtEdges(t *testing.T) {
	o := newOutdoor(t, 0, 0)
	view := o.View()
	// 左上角：欄 0..4 與列 0..1 都被夾到 0，因此與欄 5／列 2 同值。
	for row := 0; row < 2; row++ {
		for col := 0; col < 5; col++ {
			if view[row][col] != view[2][5] {
				t.Fatalf("(列 %d, 欄 %d) 沒有被夾限到原點", row, col)
			}
		}
	}
	o = newOutdoor(t, GridSide-1, GridSide-1)
	view = o.View()
	last := view[Viewport.Rows-1][Viewport.Cols-1]
	if view[2][5] != last {
		t.Fatal("右下角沒有被夾限到邊界")
	}
}

func TestMoveStopsAtBounds(t *testing.T) {
	o := newOutdoor(t, 0, 0)
	o.Move(-1, -1)
	if o.Camera.X != 0 || o.Camera.Y != 0 {
		t.Fatalf("鏡頭越過下界：(%d, %d)", o.Camera.X, o.Camera.Y)
	}
	o.Camera.X, o.Camera.Y = GridSide-1, GridSide-1
	o.Move(1, 1)
	if o.Camera.X != GridSide-1 || o.Camera.Y != GridSide-1 {
		t.Fatalf("鏡頭越過上界：(%d, %d)", o.Camera.X, o.Camera.Y)
	}
}

// 動畫畫格在 0…4 之間來回，與 word_2DF24／word_2DF26 的規則相同。
func TestAdvanceMarkerPingPongs(t *testing.T) {
	o := newOutdoor(t, 10, 10)
	step := 1
	got := []int{o.Marker.Frame}
	for i := 0; i < 9; i++ {
		o.AdvanceMarker(&step)
		got = append(got, o.Marker.Frame)
	}
	want := []int{0, 1, 2, 3, 4, 3, 2, 1, 0, 1}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("畫格序列 %v，應為 %v", got, want)
		}
	}
}

// 門口是狀態不是位置：原版是移動到門的座標時才觸發，
// 剛從建築物走出來、站在同一格上時仍然是地圖畫面。
func TestDoorStateNeedsMovement(t *testing.T) {
	o := newOutdoor(t, 57, 8)
	o.SetBuildings(&TownBuildings{Town: 7, Doors: []Door{
		{Index: 5, X: 57, Y: 8, Picture: 13, Portrait: 12, Name: "Halfway House"},
	}})
	o.Icons = []int{DoorIcon}

	if o.DoorUnderParty() == nil {
		t.Fatal("隊伍座標上應該有門")
	}
	if o.AtDoor() != nil {
		t.Fatal("還沒移動就不該是門口狀態")
	}
	icons, selected := o.IconRow()
	if len(icons) != 1 || selected != 0 {
		t.Fatalf("非門口狀態的圖示列是 %v／%d", icons, selected)
	}

	o.Move(0, 0)
	if o.AtDoor() == nil {
		t.Fatal("移動之後應該進入門口狀態")
	}
	icons, selected = o.IconRow()
	if len(icons) != 2 || icons[0] != DoorIcon || icons[1] != UTurnIcon || selected != 0 {
		t.Fatalf("門口狀態的圖示列是 %v／%d，應為 [DOOR U-TURN]／0", icons, selected)
	}

	o.Move(1, 0)
	if o.AtDoor() != nil {
		t.Fatal("走離門口之後應該回到地圖畫面")
	}
}

func TestDoorAtLooksUpCoordinates(t *testing.T) {
	b := &TownBuildings{Town: 7, Doors: []Door{
		{Index: 0, X: 13, Y: 9}, {Index: 5, X: 57, Y: 8},
	}}
	if d := b.DoorAt(57, 8); d == nil || d.Index != 5 {
		t.Fatalf("(57,8) 應該是門 5，取到 %v", d)
	}
	if b.DoorAt(57, 9) != nil {
		t.Fatal("(57,9) 不該有門")
	}
}
