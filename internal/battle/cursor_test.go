package battle

import "testing"

// 近戰的起點就是面前那一格（`0x241C7`）。
func TestAdjacentCursorStart(t *testing.T) {
	cases := map[int]Cell{
		0: {3, 4}, // 上
		1: {3, 5}, // 右上
		2: {4, 5}, // 右
		3: {5, 5}, // 右下
		4: {5, 4}, // 下
		5: {5, 3}, // 左下
		6: {4, 3}, // 左
		7: {3, 3}, // 左上
	}
	for facing, want := range cases {
		c := NewAdjacentCursor(&Actor{Facing: facing})
		if c.Row != want.Row || c.Col != want.Col {
			t.Errorf("方向 %d 的起點是 (%d, %d)，應該是 (%d, %d)",
				facing, c.Row, c.Col, want.Row, want.Col)
		}
	}
}

// 近戰不准停在自己那一格：移到中心就再往同方向走一步（`0x2444C`）。
func TestAdjacentCursorSkipsCentre(t *testing.T) {
	c := NewAdjacentCursor(&Actor{Facing: 0}) // (3, 4)
	c.Move(1, 0, false)
	if c.Row != 5 || c.Col != 4 {
		t.Fatalf("往下走過中心應該落在 (5, 4)，得到 (%d, %d)", c.Row, c.Col)
	}
	// 斜著穿過中心也一樣。
	c = TargetCursor{Mode: CursorAdjacent, Row: 3, Col: 3}
	c.Move(1, 1, false)
	if c.Row != 5 || c.Col != 5 {
		t.Fatalf("斜穿中心應該落在 (5, 5)，得到 (%d, %d)", c.Row, c.Col)
	}
}

// 近戰夾在 3…5（`0x2446E`）。
func TestAdjacentCursorClamp(t *testing.T) {
	c := TargetCursor{Mode: CursorAdjacent, Row: 3, Col: 3}
	c.Move(-1, -1, false)
	if c.Row != 3 || c.Col != 3 {
		t.Fatalf("左上角再往外應該留在 (3, 3)，得到 (%d, %d)", c.Row, c.Col)
	}
	c = TargetCursor{Mode: CursorAdjacent, Row: 5, Col: 5}
	c.Move(1, 1, false)
	if c.Row != 5 || c.Col != 5 {
		t.Fatalf("右下角再往外應該留在 (5, 5)，得到 (%d, %d)", c.Row, c.Col)
	}
}

// 射擊的起點是上次記住的目標換算成視窗座標（`0x2423D`）。
func TestRangedCursorStart(t *testing.T) {
	a := &Actor{X: 20, Y: 30}
	c := NewRangedCursor(a, Cell{Row: 18, Col: 33})
	if c.Row != 2 || c.Col != 7 || !c.Memo {
		t.Fatalf("起點應該是 (2, 7) 且有記憶目標，得到 (%d, %d) memo=%v",
			c.Row, c.Col, c.Memo)
	}
	// 算出來落在 0…8 之外就沒有可用的起點。
	c = NewRangedCursor(a, Cell{Row: 4, Col: 30})
	if c.Memo {
		t.Fatal("越界的記憶目標不該算成可用")
	}
}

// 射擊不准落在 2…6 的內圈：方向鍵乘三跳出去、滑鼠還原（`0x244F1`、`0x24519`）。
func TestRangedCursorInnerRing(t *testing.T) {
	c := TargetCursor{Mode: CursorRanged, Row: 2, Col: 4}
	c.Move(1, 0, false)
	if c.Row != 6 || c.Col != 4 {
		t.Fatalf("方向鍵移進內圈應該乘三跳到 (6, 4)，得到 (%d, %d)", c.Row, c.Col)
	}
	c = TargetCursor{Mode: CursorRanged, Row: 2, Col: 4}
	c.Move(1, 0, true)
	if c.Row != 2 || c.Col != 4 {
		t.Fatalf("滑鼠點內圈應該還原成 (2, 4)，得到 (%d, %d)", c.Row, c.Col)
	}
	// 外圈照常走。
	c = TargetCursor{Mode: CursorRanged, Row: 0, Col: 0}
	c.Move(1, 1, false)
	if c.Row != 1 || c.Col != 1 {
		t.Fatalf("外圈走一步應該到 (1, 1)，得到 (%d, %d)", c.Row, c.Col)
	}
	// 夾在 0…8。
	c = TargetCursor{Mode: CursorRanged, Row: 8, Col: 8}
	c.Move(1, 1, false)
	if c.Row != 8 || c.Col != 8 {
		t.Fatalf("右下角再往外應該留在 (8, 8)，得到 (%d, %d)", c.Row, c.Col)
	}
}

// Return 的兩種結果，以及白框停在 U-TURN 上就是取消（`0x248F5`）。
func TestCursorConfirm(t *testing.T) {
	a := &Actor{X: 20, Y: 30, Facing: 0}
	c := NewAdjacentCursor(a)
	c.Move(0, 1, false) // (3, 5) ＝ 右上
	if _, ok := c.Confirm(a); !ok {
		t.Fatal("白框在第一格時 Return 應該成立")
	}
	if a.Facing != 1 {
		t.Fatalf("選右上那一格應該把方向寫成 1，得到 %d", a.Facing)
	}
	r := TargetCursor{Mode: CursorRanged, Row: 1, Col: 7}
	cell, ok := r.Confirm(a)
	if !ok || cell.Row != 17 || cell.Col != 33 {
		t.Fatalf("射擊確定應該回 (17, 33)，得到 (%d, %d) ok=%v",
			cell.Row, cell.Col, ok)
	}
	r.Icon = 1
	if _, ok := r.Confirm(a); ok {
		t.Fatal("白框停在 U-TURN 上應該當成取消")
	}
}
