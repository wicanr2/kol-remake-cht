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

// 0x23…0x2D 這一段（牆與碎石）另有兩條例外，都在 `sub_25F8A:0x2602C`…`0x26095`：
// 飛在空中的人（記錄 +0x1E 非零）整條線都穿得過去，或者站在 RoadTileA／RoadTileB 上
// 時視線的第一步穿得過去。
const (
	seeOverLow  = 0x23
	seeOverHigh = 0x2E
	// RoadTileA／RoadTileB 是那兩塊「站上去就看得過去」的磚。
	RoadTileA = 0x24
	RoadTileB = 0x25
	// MountedSight 不是磚塊，是「飛在空中」這件事在 seesOver 裡的代號——原版那一條
	// 走的是參戰者記錄 +0x1E，不看腳下（`sub_25F8A:0x2606C`）。
	MountedSight = -1
)

// walkBits 是磚塊擋不擋**路**的位元圖，原樣抄自 DGROUP 0x0FC8 十個 byte
// （`sub_24E2B:0x24ED3`）。取位元的方式與 sightBits 相同。這是與擋視線
// （0x10CE）不同的另一張表：例如磚 0 擋視線但不擋路。
var walkBits = [10]byte{0x00, 0x10, 0x9A, 0xFF, 0xF1, 0xFF, 0x17, 0xFF, 0xD0, 0x7F}

// BlocksMovement 回報一塊磚擋不擋路。
func BlocksMovement(tile int) bool {
	if tile < 0 || tile >= len(walkBits)*8 {
		return true
	}
	return walkBits[tile/8]&(0x80>>uint(tile%8)) != 0
}

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
//
// mounted 是看的人記錄 +0x1E 非零，也就是飛在空中——那時牆與碎石整段都看得過去。
func (f *Field) Visible(x, y int, mounted bool) Sight {
	var state [ViewRows * ViewCols]int
	foot := f.At(clampCell(x), clampCell(y))
	if mounted {
		foot = MountedSight
	}
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

// walkLine 是 `sub_25D7A` 的走線：主軸每步走一格，誤差項滿了副軸才走一格。
// step 拿到第幾步與走到的格，回 false 就停。**起點那一格不會送進 step**，
// 原版也是先走一步才判（`0x25E19` 在迴圈開頭就位移）。
func walkLine(ox, oy, tx, ty int, step func(k, r, c int) bool) {
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
		if !step(k, r, c) {
			return
		}
	}
}

// ray 是視線那一種走法（`sub_25D7A(3)`）：線上的格子依序標記，
// 標到擋住的那一格就停。
func (f *Field) ray(ox, oy, tx, ty, foot int, state *[ViewRows * ViewCols]int) {
	walkLine(ox, oy, tx, ty, func(k, r, c int) bool {
		idx := (r-ox+ViewCentre.Row)*ViewCols + (c - oy + ViewCentre.Col)
		if idx < 0 || idx >= len(state) {
			return false
		}
		if state[idx] == sightUnknown {
			state[idx] = sightSeen
			tile := f.At(r, c)
			if Blocks(tile) && !seesOver(tile, foot, k) {
				state[idx] = sightBlocked
			}
		}
		return state[idx] != sightBlocked
	})
}

// seesOver 是牆與碎石那一段的例外。飛在空中的人（foot 為 MountedSight）不受步數與
// 腳下磚塊限制；其餘只有視線的第一步、而且站在路面上才穿得過去。
func seesOver(tile, foot, step int) bool {
	if tile < seeOverLow || tile >= seeOverHigh {
		return false
	}
	if foot == MountedSight {
		return true
	}
	return step == 0 && (foot == RoadTileA || foot == RoadTileB)
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

// LineOfSight 回報從 (ox, oy) 到 (tx, ty) 這條線通不通得過去
// （`sub_25D0F` ＋ `sub_25D7A(2)`，RE-102）。mode 2 只看地形：不比樹、不看人、
// 也不畫東西，遇到擋視線的磚就停。
//
// **原版那一支回 0 才是通得到**（`0x20F38` 拿 `jz` 接著畫戰場），這裡把回傳值
// 翻正過來，免得在 remake 這一側每個呼叫點都要記得反過來讀。起點與終點是同一格
// 時原版回 0，也就是「通」。
func (f *Field) LineOfSight(ox, oy, tx, ty, foot int) bool {
	if ox == tx && oy == ty {
		return true
	}
	clear := true
	walkLine(ox, oy, tx, ty, func(k, r, c int) bool {
		tile := f.At(r, c)
		if Blocks(tile) && !seesOver(tile, foot, k) {
			clear = false
			return false
		}
		return true
	})
	return clear
}
