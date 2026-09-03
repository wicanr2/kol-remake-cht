package battle

import (
	"fmt"
	"image"
)

// 戰場的磚塊圖（RE-079）。一般遭遇是現場亂數生成的，任務地點不重新生成。

// 戰場用得到的磚塊編號。
const (
	// TileGround 是底色那一塊；TileRough 是散在裡面的另一種地面。
	TileGround = 0x09
	TileRough  = 0x0D
	// TileScatter 是後來灑上去的那一種。
	TileScatter = 0x0A
)

// ScatterRolls 是決定灑幾格的擲骰次數，ScatterDie 是骰面數
// （`sub_1E3A8:0x1E5EB` 擲 100 次 1…6 加總）。
const (
	ScatterRolls = 100
	ScatterDie   = 6
)

// RoughOdds 是底色擲骰的骰面數：擲出 1 就填 TileRough（`0x1E59C`）。
const RoughOdds = 4

// Rand 是可注入的亂數來源；n 為骰面數，回傳 1…n。與其他模組同一個介面。
type Rand interface {
	Roll(n int) int
}

// Field 是一張戰場。
type Field struct {
	Tiles []int
}

// NewField 是任務地點用的空戰場：整片底色，地形不重新生成。
func NewField() *Field {
	tiles := make([]int, FieldSide*FieldSide)
	for i := range tiles {
		tiles[i] = TileGround
	}
	return &Field{Tiles: tiles}
}

// LoadField 讀抽取器產生的戰場格子陣列 PNG。像素值即磚塊編號，
// 索引與世界地圖相同：`(X << 6) + Y`。
func LoadField(img image.Image) (*Field, error) {
	b := img.Bounds()
	if b.Dx() != FieldSide || b.Dy() != FieldSide {
		return nil, fmt.Errorf("戰場格子陣列應該是 %d×%d，卻是 %d×%d",
			FieldSide, FieldSide, b.Dx(), b.Dy())
	}
	f := &Field{Tiles: make([]int, FieldSide*FieldSide)}
	for row := 0; row < FieldSide; row++ {
		for col := 0; col < FieldSide; col++ {
			v, _, _, _ := img.At(b.Min.X+col, b.Min.Y+row).RGBA()
			f.Tiles[row*FieldSide+col] = int(v >> 8)
		}
	}
	return f, nil
}

// Generate 是一般遭遇的地形生成（`sub_1E3A8:0x1E594`…`0x1E5E2`）：
// 逐格擲 1…4，中 1 就填 TileRough、其餘 TileGround，再灑一次 TileScatter。
func Generate(rng Rand) *Field {
	f := &Field{Tiles: make([]int, FieldSide*FieldSide)}
	for i := range f.Tiles {
		if rng.Roll(RoughOdds) == 1 {
			f.Tiles[i] = TileRough
		} else {
			f.Tiles[i] = TileGround
		}
	}
	f.Scatter(rng)
	return f
}

// Scatter 是灑 TileScatter 的那一段（`sub_1E3A8:0x1E5E5`…`0x1E68C`）。
// **任務地點的戰場也會走這一段**：`word_34968` 只擋掉上面的底色生成
// （`0x1E58A` 的分支落在 `0x1E5E5`），灑的動作照跑——原版任務戰場上
// 那些散落的樹叢就是這樣來的。
//
// 先擲 100 次 1…6 加總得到次數，再照那個次數挑格子（`rand(0x1000) − 1`）；
// **只有原本是 TileGround 或 TileRough 的格子會被換掉**（`0x1E63C`／`0x1E65F`），
// 任務地圖上的路、牆、水面因此保持原樣。
func (f *Field) Scatter(rng Rand) {
	total := 0
	for i := 0; i < ScatterRolls; i++ {
		total += rng.Roll(ScatterDie)
	}
	for i := 0; i < total; i++ {
		at := rng.Roll(len(f.Tiles)) - 1
		if f.Tiles[at] == TileGround || f.Tiles[at] == TileRough {
			f.Tiles[at] = TileScatter
		}
	}
}

// At 回傳一格的磚塊編號；超出範圍回底色。索引是 `(X << 6) + Y`，
// 而**參戰者的 X 是畫面上的列、Y 是欄**——原版的陣型表就是這樣擺的：
// 第 0 組敵人的 X 是 26…28（隊伍在 31／33），也就是從上方壓下來。
func (f *Field) At(x, y int) int {
	if x < 0 || y < 0 || x >= FieldSide || y >= FieldSide {
		return TileGround
	}
	return f.Tiles[x*FieldSide+y]
}

// View 回傳以 (x, y) 為中心的視窗。x 走列、y 走欄；中心固定在 ViewCentre，
// 超出邊界的座標夾限在 0…63（`sub_223C0:0x2250C`…`0x2253B` 的四道夾限），
// 因此視窗邊緣會重複最外圈那一格。
func (f *Field) View(x, y int) [][]int {
	rows := make([][]int, ViewRows)
	for r := range rows {
		line := make([]int, ViewCols)
		for c := range line {
			line[c] = f.At(clampCell(x-ViewCentre.Row+r), clampCell(y-ViewCentre.Col+c))
		}
		rows[r] = line
	}
	return rows
}
