package battle

// 選目標格的游標（`sub_24131`，SPEC-010）。兩個 mode 的起點、夾限與禁區都不同：
// 近戰與移動只能指相鄰八格、不准停在自己那一格；射擊與用物品可以指整個 9×9 視窗，
// 但不准落在中間那一圈。

// 游標的兩個 mode。
const (
	// CursorAdjacent 是近戰與移動用的：起點是面前那一格，夾在 3…5。
	CursorAdjacent = 1
	// CursorRanged 是射擊與用物品用的：起點是上次記住的目標，夾在 0…8。
	CursorRanged = 2
)

// TargetCursor 是選目標格的游標狀態。Row／Col 是視窗內的列與欄（0…8），
// 中心 (4, 4) 就是行動者自己。
type TargetCursor struct {
	Mode     int
	Row, Col int
	// Icon 是圖示列白框停在第幾格。非零表示停在 U-TURN 上，Confirm 會當成取消
	// （`0x248F5`：`word_2DF3A != 0` 就回 0xFFFF）。
	Icon int
	// Memo 為真表示這一次進來有上次記住的目標可用（mode 2）。
	Memo bool
}

// NewAdjacentCursor 依行動者現在的方向擺好近戰／移動的游標（`0x241C7`）。
func NewAdjacentCursor(a *Actor) TargetCursor {
	c := TargetCursor{Mode: CursorAdjacent, Row: 3, Col: 3}
	if a.Facing >= 0 && a.Facing < 8 {
		n := FacingCell(a.Facing)
		c.Row, c.Col = n/3+3, n%3+3
	}
	return c
}

// NewRangedCursor 依上次記住的目標擺好射擊／用物品的游標（`0x2423D`）。
// memo 是那個目標的絕對座標；算出來的格落在 0…8 之外時 Memo 為 false。
//
// 原版在這裡也檢查了 0…8，但檢查唯一的效果是清掉一個馬上會被覆寫的區域變數，
// 所以越界的起點照樣留在游標上。這裡照抄那個行為，只是另外把結果報出來。
func NewRangedCursor(a *Actor, memo Cell) TargetCursor {
	c := TargetCursor{Mode: CursorRanged,
		Row: memo.Row + ViewCentre.Row - a.X,
		Col: memo.Col + ViewCentre.Col - a.Y}
	c.Memo = c.Row >= 0 && c.Row <= 8 && c.Col >= 0 && c.Col <= 8
	return c
}

// Move 走一步。dRow／dCol 是這一次按的方向（八個方向鍵都是 ±1），
// mouse 為真表示這一步來自滑鼠指格而不是方向鍵。
func (c *TargetCursor) Move(dRow, dCol int, mouse bool) {
	// 兩軸都沒動就整段跳過（`0x24411`）。
	if dRow == 0 && dCol == 0 {
		return
	}
	prevRow, prevCol := c.Row, c.Col
	c.Row += dRow
	c.Col += dCol
	if c.Mode == CursorAdjacent {
		// 停在自己那一格就再往同方向走一步（`0x2444C`）。
		if c.Row == ViewCentre.Row && c.Col == ViewCentre.Col {
			c.Row += dRow
			c.Col += dCol
		}
		c.Row = clampCursor(c.Row, 3, 5)
		c.Col = clampCursor(c.Col, 3, 5)
		return
	}
	c.Row = clampCursor(c.Row, 0, 8)
	c.Col = clampCursor(c.Col, 0, 8)
	// 中間那一圈是禁區（`0x244F1`）：方向鍵移進去就乘三跳到外面，
	// 滑鼠點進去則還原成上一格。乘三之後原版沒有再夾一次。
	if c.Row > 2 && c.Row < 6 && c.Col > 2 && c.Col < 6 {
		if mouse {
			c.Row, c.Col = prevRow, prevCol
			return
		}
		c.Row += 3 * dRow
		c.Col += 3 * dCol
	}
}

// Confirm 是按下 Return。白框停在 U-TURN 上算取消，回 false。
// 近戰／移動把選中的格換成方向寫回 +0x28；射擊／用物品回傳那一格的絕對座標。
func (c *TargetCursor) Confirm(a *Actor) (Cell, bool) {
	if c.Icon != 0 {
		return Cell{}, false
	}
	if c.Mode == CursorAdjacent {
		a.Facing = CellFacing(3*(c.Row-3) + (c.Col - 3))
		return Cell{Row: a.X, Col: a.Y}, true
	}
	return Cell{Row: a.X + c.Row - ViewCentre.Row,
		Col: a.Y + c.Col - ViewCentre.Col}, true
}

func clampCursor(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
