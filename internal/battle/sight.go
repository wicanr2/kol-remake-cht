package battle

// 戰場的視線（RE-079）。任務地點的戰場只畫得出視線通得到的格子，其餘畫成
// TileUnseen；一般遭遇整片都看得到——`sub_223C0:0x2240A` 用 `word_34968`
// 分這兩條路，只有任務那一條會先把 81 格的可見度全設成「還沒看到」。

// TileUnseen 是還沒看到那一格畫的磚塊（`sub_223C0:0x22597` 填的 0x4A）。
const TileUnseen = 0x4A

// sightBits 是磚塊擋不擋視線的位元圖，原樣抄自 KOLTOWN 的 DGROUP 0x10CE
// 十個 byte；取位元用的是 `byte_1312` 的 `80 40 20 10 08 04 02 01`
// （`sub_25F8A:0x26001`…`0x2601D`）。位元有設就是擋。
var sightBits = [10]byte{0xFF, 0x80, 0x01, 0xFF, 0xF1, 0xFF, 0x17, 0xF9, 0x76, 0x7F}

// 0x23…0x2D 這一段（牆與碎石）另有例外：站在 RoadTileA／RoadTileB 上時，
// 視線的第一步穿得過去（`sub_25F8A:0x2602C`…`0x26095`）。
const (
	seeOverLow  = 0x23
	seeOverHigh = 0x2E
	// RoadTileA／RoadTileB 是那兩塊「站上去就看得過去」的磚。
	RoadTileA = 0x24
	RoadTileB = 0x25
)

// Blocks 回報一塊磚擋不擋視線。
func Blocks(tile int) bool {
	if tile < 0 || tile >= len(sightBits)*8 {
		return true
	}
	return sightBits[tile/8]&(0x80>>uint(tile%8)) != 0
}

// Sight 是視窗每一格看不看得到，索引與視窗相同：[列][欄]。
type Sight [ViewRows][ViewCols]bool

// 每一格在計算中的三種狀態。0 是還沒算到，也就是畫成 TileUnseen 的那一種。
const (
	sightUnknown = 0
	sightSeen    = 1
	sightBlocked = 3
)

// Visible 算出以 (x, y) 為中心的視窗可見度。原版對 81 格逐格拉一條線
// （`sub_223C0:0x22436` 兩層迴圈呼叫 `sub_25D7A(3)`）：線上的格子依序標記，
// 標到擋住的那一格就停，後面的格子維持「還沒看到」。
func (f *Field) Visible(x, y int) Sight {
	var state [ViewRows * ViewCols]int
	foot := f.At(clampCell(x), clampCell(y))
	for row := 0; row < ViewRows; row++ {
		for col := 0; col < ViewCols; col++ {
			f.ray(x, y,
				clampCell(x+row-ViewCentre.Row),
				clampCell(y+col-ViewCentre.Col), foot, &state)
		}
	}
	var out Sight
	for i, v := range state {
		out[i/ViewCols][i%ViewCols] = v != sightUnknown
	}
	return out
}

// ray 是 `sub_25D7A` 的走線：主軸每步走一格，誤差項滿了副軸才走一格。
func (f *Field) ray(ox, oy, tx, ty, foot int, state *[ViewRows * ViewCols]int) {
	dRow, dCol := tx-ox, ty-oy
	aRow, aCol := abs(dRow), abs(dCol)
	steps, rowMajor := aCol, false
	if aCol < aRow {
		steps, rowMajor = aRow, true
	}
	if steps == 0 {
		return
	}
	err := steps / 2
	r, c := ox, oy
	for k := 0; k < steps; k++ {
		if rowMajor {
			err += aCol
		} else {
			err += aRow
		}
		if err >= steps {
			if rowMajor {
				c += sign(dCol)
			} else {
				r += sign(dRow)
			}
			err -= steps
		}
		if rowMajor {
			r += sign(dRow)
		} else {
			c += sign(dCol)
		}
		idx := (r-ox+ViewCentre.Row)*ViewCols + (c - oy + ViewCentre.Col)
		if idx < 0 || idx >= len(state) {
			return
		}
		if state[idx] == sightUnknown {
			state[idx] = sightSeen
			tile := f.At(r, c)
			if Blocks(tile) && !seesOver(tile, foot, k) {
				state[idx] = sightBlocked
			}
		}
		if state[idx] == sightBlocked {
			return
		}
	}
}

// seesOver 是牆與碎石那一段的例外：只有視線的第一步、而且站在路面上才穿得過去。
func seesOver(tile, foot, step int) bool {
	return tile >= seeOverLow && tile < seeOverHigh && step == 0 &&
		(foot == RoadTileA || foot == RoadTileB)
}

func clampCell(v int) int {
	if v < 0 {
		return 0
	}
	if v > FieldSide-1 {
		return FieldSide - 1
	}
	return v
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func sign(v int) int {
	switch {
	case v < 0:
		return -1
	case v > 0:
		return 1
	}
	return 0
}
