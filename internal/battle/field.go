package battle

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

// Generate 是一般遭遇的地形生成（`sub_1E3A8:0x1E594`…`0x1E68C`）。
// 先逐格擲 1…4，中 1 的填 TileRough；再擲 100 次 1…6 加總，
// 照那個數把隨機格子改成 TileScatter。
func Generate(rng Rand) *Field {
	f := &Field{Tiles: make([]int, FieldSide*FieldSide)}
	for i := range f.Tiles {
		if rng.Roll(RoughOdds) == 1 {
			f.Tiles[i] = TileRough
		} else {
			f.Tiles[i] = TileGround
		}
	}
	total := 0
	for i := 0; i < ScatterRolls; i++ {
		total += rng.Roll(ScatterDie)
	}
	for i := 0; i < total; i++ {
		f.Tiles[rng.Roll(len(f.Tiles))-1] = TileScatter
	}
	return f
}

// At 回傳一格的磚塊編號；超出範圍回底色。
func (f *Field) At(x, y int) int {
	if x < 0 || y < 0 || x >= FieldSide || y >= FieldSide {
		return TileGround
	}
	return f.Tiles[y*FieldSide+x]
}

// View 回傳以 (cx, cy) 為中心的 9×9 視窗。
func (f *Field) View(cx, cy int) [][]int {
	rows := make([][]int, ViewRows)
	for r := range rows {
		line := make([]int, ViewCols)
		for c := range line {
			line[c] = f.At(cx-ViewCols/2+c, cy-ViewRows/2+r)
		}
		rows[r] = line
	}
	return rows
}
