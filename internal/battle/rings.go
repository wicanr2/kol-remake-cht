package battle

// 同心方環（RE-081）。戰場的搜尋單位是以行動者為中心的方環：環號 0…6，
// 格數在 `[0x0FC0]` ＝ `08 10 18 20 28 30 38`，位移由 `[0x0FD8]`（X）與
// `[0x0FE6]`（Y）兩張指標表指出去。
//
// 半徑不等於環號——**環 r 的半徑是 r ＋ 1**，所以環 0 是八個相鄰格、環 6 是
// 56 格的最外圈。起點在 `(−R, 0)`，沿邊順時針走一圈。位移用公式產生而不是抄表，
// 因為 224 對數字抄錯一個很難發現；`rings_test.go` 拿原版那七張表逐項驗過。

// RingCount 是環的數量（`[0x0FC0]` 有七筆）。
const RingCount = 7

// RingCells 是環 r 的位移，從 `(−R, 0)` 起順時針，共 8R 格。
func RingCells(ring int) []Cell {
	if ring < 0 || ring >= RingCount {
		return nil
	}
	r := ring + 1
	out := make([]Cell, 0, 8*r)
	for y := 0; y <= r; y++ {
		out = append(out, Cell{Row: -r, Col: y})
	}
	for x := -r + 1; x <= r; x++ {
		out = append(out, Cell{Row: x, Col: r})
	}
	for y := r - 1; y >= -r; y-- {
		out = append(out, Cell{Row: r, Col: y})
	}
	for x := r - 1; x >= -r; x-- {
		out = append(out, Cell{Row: x, Col: -r})
	}
	for y := -r + 1; y < 0; y++ {
		out = append(out, Cell{Row: -r, Col: y})
	}
	return out
}

// RingSize 是環 r 的格數（8R）。
func RingSize(ring int) int {
	if ring < 0 || ring >= RingCount {
		return 0
	}
	return 8 * (ring + 1)
}
