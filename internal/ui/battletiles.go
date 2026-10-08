package ui

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/kol-remake-cht/internal/battle"
)

// BattleTiles 是戰場磚圖：一列 10 塊、一塊 16×16，由
// `tools/extract_battle_tiles.py` 從 `KOLPICS` 的 payload 8 產生。
// 疊過精靈的磚塊會留著重複用——一場戰鬥只有幾種組合。
type BattleTiles struct {
	src                 image.Image
	atlas               *ebiten.Image
	tiles               map[int]*ebiten.Image
	comp                map[[2]int]*ebiten.Image
	rings               [5]*ebiten.Image
	ringPhase, ringStep int
	ringWait            float64
}

// AdvanceTargetMarker 只推進呈現相位；每三個 BIOS tick 一格，0…4 往返。
// SPEC-010：現代等待近似，與戰鬥規則、亂數及保存分離。
func (b *BattleTiles) AdvanceTargetMarker(ticks float64) {
	if b.ringStep == 0 {
		b.ringStep = 1
	}
	b.ringWait += ticks
	for b.ringWait >= 3 {
		b.ringWait -= 3
		b.ringPhase += b.ringStep
		if b.ringPhase == 4 {
			b.ringStep = -1
		} else if b.ringPhase == 0 {
			b.ringStep = 1
		}
	}
}

// TargetRing 只讀已抽取的十格 PNG；單元5…9是戰鬥圓環（RE-171）。
func (b *BattleTiles) TargetRing(src image.Image) (*ebiten.Image, error) {
	if src.Bounds().Dx() != 16 || src.Bounds().Dy() != 160 {
		return nil, fmt.Errorf("戰鬥圓環圖集尺寸 %v，應為16×160", src.Bounds())
	}
	if img := b.rings[b.ringPhase]; img != nil {
		return img, nil
	}
	out := image.NewRGBA(image.Rect(0, 0, 16, 16))
	ox, oy := src.Bounds().Min.X, src.Bounds().Min.Y+(5+b.ringPhase)*16
	for y := 0; y < 16; y++ {
		var mask uint16
		for x := 0; x < 16; x++ {
			if !isBlack(src.At(ox+x, oy+y)) {
				mask |= pixelBit(x)
			}
		}
		wide := mask | mask<<1 | mask>>1
		for x := 0; x < 16; x++ {
			if wide&pixelBit(x) != 0 {
				out.Set(x, y, src.At(ox+x, oy+y))
			}
		}
	}
	b.rings[b.ringPhase] = ebiten.NewImageFromImage(out)
	return b.rings[b.ringPhase], nil
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
