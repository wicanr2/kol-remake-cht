package ui

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/kol-remake-cht/internal/battle"
)

// BattleTiles 是戰場磚圖：一列 10 塊、一塊 16×16，由
// `tools/extract_battle_tiles.py` 從 `KOLPICS` 的 payload 8 產生。
// 疊過精靈的磚塊會留著重複用——一場戰鬥只有幾種組合。
type BattleTiles struct {
	src   image.Image
	atlas *ebiten.Image
	tiles map[int]*ebiten.Image
	comp  map[[2]int]*ebiten.Image
}

// NewBattleTiles 收下抽取器產生的 atlas。
func NewBattleTiles(src image.Image) *BattleTiles {
	return &BattleTiles{
		src:   src,
		atlas: ebiten.NewImageFromImage(src),
		tiles: map[int]*ebiten.Image{},
		comp:  map[[2]int]*ebiten.Image{},
	}
}

// Tile 是單純一塊磚。
func (b *BattleTiles) Tile(id int) *ebiten.Image {
	if img, ok := b.tiles[id]; ok {
		return img
	}
	x, y := b.origin(id)
	img := b.atlas.SubImage(image.Rect(x, y,
		x+battle.TileSide, y+battle.TileSide)).(*ebiten.Image)
	b.tiles[id] = img
	return img
}

// WithSprite 把參戰者疊到磚塊上（`sub_1D21D`）。原版先把精靈的四個平面 OR
// 成剪影，再把每一列當成一個 16-bit word 往左右各推一格加寬，取反當遮罩：
// 加寬到的地方畫精靈、其餘留磚塊，精靈四周因此帶一圈黑邊。
func (b *BattleTiles) WithSprite(tile, sprite int) *ebiten.Image {
	key := [2]int{tile, sprite}
	if img, ok := b.comp[key]; ok {
		return img
	}
	out := image.NewRGBA(image.Rect(0, 0, battle.TileSide, battle.TileSide))
	for y := 0; y < battle.TileSide; y++ {
		var silhouette uint16
		for x := 0; x < battle.TileSide; x++ {
			if !isBlack(b.pixel(sprite, x, y)) {
				silhouette |= pixelBit(x)
			}
		}
		wide := silhouette | silhouette<<1 | silhouette>>1
		for x := 0; x < battle.TileSide; x++ {
			c := b.pixel(tile, x, y)
			if wide&pixelBit(x) != 0 {
				c = b.pixel(sprite, x, y)
			}
			out.Set(x, y, c)
		}
	}
	img := ebiten.NewImageFromImage(out)
	b.comp[key] = img
	return img
}

// origin 是一塊磚在 atlas 裡的左上角。
func (b *BattleTiles) origin(id int) (int, int) {
	return (id % battle.TileAtlasColumns) * battle.TileSide,
		(id / battle.TileAtlasColumns) * battle.TileSide
}

func (b *BattleTiles) pixel(id, x, y int) color.Color {
	ox, oy := b.origin(id)
	bounds := b.src.Bounds()
	return b.src.At(bounds.Min.X+ox+x, bounds.Min.Y+oy+y)
}

func isBlack(c color.Color) bool {
	r, g, b, _ := c.RGBA()
	return r == 0 && g == 0 && b == 0
}

// pixelBit 是一個像素在那一列 16-bit word 裡的位元。一列兩個 byte，
// 每個 byte 的最高位是自己那半邊最左邊的像素；word 是小端序，
// 低 byte 放左半邊，所以左半邊落在 bit 7…0、右半邊落在 bit 15…8。
func pixelBit(x int) uint16 {
	if x < 8 {
		return 1 << uint(7-x)
	}
	return 1 << uint(15-(x-8))
}
