package ui

import (
	"crypto/sha256"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"image"
	"image/color"
)

// The atlas is selected from the actual decoded80-cell image, never a flag.
func motionBattleAtlas(tiles *BattleTiles) (int, error) {
	if tiles == nil || tiles.src == nil || tiles.src.Bounds().Size() != image.Pt(160, 128) {
		return 0, fmt.Errorf("HD戰場來源須為實際80格圖集")
	}
	var rgb []byte
	b := tiles.src.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bb, a := tiles.src.At(x, y).RGBA()
			if a != 65535 {
				return 0, fmt.Errorf("HD戰場來源須不透明")
			}
			rgb = append(rgb, uint8(r>>8), uint8(g>>8), uint8(bb>>8))
		}
	}
	sum := fmt.Sprintf("%x", sha256.Sum256(rgb))
	switch sum {
	case motionBattle08RGB:
		return 8, nil
	case motionBattle12RGB:
		return 12, nil
	}
	return 0, fmt.Errorf("HD戰場實際圖集指紋不符")
}
func drawHDBattleField(dst *ebiten.Image, tiles *BattleTiles, s *battle.State, scale int) (bool, error) {
	n := activeHDUI
	if n == nil || n.dst != dst || n.h.motion == nil {
		return false, nil
	}
	h := n.h
	atlas, e := motionBattleAtlas(tiles)
	if e != nil {
		return true, e
	}
	cx, cy := s.Camera()
	mounted := false
	if a, ok := s.ViewActor(); ok {
		mounted = a.Flag1E != 0
	}
	sight := s.Field.Visible(cx, cy, mounted)
	v := battle.FieldView
	side := battle.TileSide * scale * 2
	for row := 0; row < battle.ViewRows; row++ {
		for col := 0; col < battle.ViewCols; col++ {
			x, y := clampField(cx-battle.ViewCentre.Row+row), clampField(cy-battle.ViewCentre.Col+col)
			r := image.Rect((v.X+col*battle.TileSide)*scale*2, (v.Y+row*battle.TileSide)*scale*2, (v.X+(col+1)*battle.TileSide)*scale*2, (v.Y+(row+1)*battle.TileSide)*scale*2)
			if !sight[row][col] && (row != battle.ViewCentre.Row || col != battle.ViewCentre.Col) {
				vector.DrawFilledRect(dst, float32(r.Min.X), float32(r.Min.Y), float32(side), float32(side), color.RGBA{0, 0, 0, 255}, false)
				continue
			}
			tile := s.Field.At(x, y)
			if tile < 9 {
				dst.SubImage(r).(*ebiten.Image).Fill(color.Black)
			}
			digit := -1
			if tile >= 0 && tile <= 5 {
				digit = tile
			}
			pic, e := h.motionBattleTile(tile, atlas)
			if e != nil {
				return true, e
			}
			h.drawHD(dst, pic, r)
			sprite := s.SpriteAt(x, y)
			if sprite >= 0 {
				key := sprite
				if sprite >= 0 && sprite <= 5 {
					key = 0
				}
				paint := h.motion.images[fmt.Sprintf("battle-sprite-%d", key)]
				if paint == nil {
					return true, fmt.Errorf("HD戰場精靈%d沒有綁定", sprite)
				}
				h.drawHD(dst, paint, r)
				digit = -1
				if sprite <= 5 {
					digit = sprite
				}

			}
			if digit >= 0 {
				h.style.label(dst, fmt.Sprint(digit+1), image.Rect(r.Min.X+side*7/16, r.Min.Y, r.Min.X+side*12/16, r.Min.Y+side*6/16), float64(side)*.28, hdInk, true)
			}

		}
	}
	return true, nil
}
func (h *HD) motionBattleTile(id, atlas int) (*ebiten.Image, error) {
	if h.motion.tiles == nil {
		h.motion.tiles = map[[2]int]*ebiten.Image{}
	}
	key := [2]int{id, atlas}
	if p := h.motion.tiles[key]; p != nil {
		return p, nil
	}
	p, e := h.motionBattleTileUncached(id, atlas)
	if e == nil {
		h.motion.tiles[key] = p
	}
	return p, e
}
func (h *HD) motionBattleTileUncached(id, atlas int) (*ebiten.Image, error) {
	if id >= 0 && id <= 8 {
		key := id
		if id <= 5 {
			key = 0
		}
		return h.motion.images[fmt.Sprintf("battle-sprite-%d", key)], nil
	}
	if id == battle.TileUnseen {
		pic := h.motion.images["native-unseen"]
		if pic == nil {
			pic = ebiten.NewImage(128, 128)
			pic.Fill(color.Black)
			h.motion.images["native-unseen"] = pic
		}
		return pic, nil
	}
	if id == 9 {
		return h.motion.images["battle-ground-9"], nil
	}
	if id < 10 || id >= 80 {
		return nil, fmt.Errorf("HD戰場地形索引%d越界", id)
	}
	if atlas == 12 {
		ids := []int{23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 73}
		for i, tile := range ids {
			if tile == id {
				sheet := h.motion.images["battle-alternate-0"]
				x, y := i%5*128, i/5*128
				if i >= 10 {
					sheet = h.motion.images["battle-alternate-1"]
					x, y = (i-10)*128, 0
				}
				return h.paintSubImage(sheet, image.Rect(x, y, x+128, y+128)), nil
			}
		}
	}
	start := id / 10 * 10
	sheet := h.motion.images[fmt.Sprintf("battle-terrain-%d", start)]
	i := id - start
	x, y := i%5*128, i/5*128
	return h.paintSubImage(sheet, image.Rect(x, y, x+128, y+128)), nil
}
func drawHDTargetRing(dst *ebiten.Image, tiles *BattleTiles, source *ebiten.Image, row, col, scale int) bool {
	n := activeHDUI
	if n == nil || n.dst != dst || n.h.motion == nil {
		return false
	}
	phase := tiles.ringPhase
	minX, minY, maxX, maxY := 16, 16, -1, -1
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			r, g, b, _ := source.At(x, (5+phase)*16+y).RGBA()
			if r|g|b != 0 {
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}
	if maxX < minX {
		return true
	}
	v := battle.FieldView
	k := float32(scale * 2)
	cx := float32(v.X+col*16)*k + float32(minX+maxX+1)*k/2
	cy := float32(v.Y+row*16)*k + float32(minY+maxY+1)*k/2
	radius := float32(maxX-minX+1) * k / 2
	vector.StrokeCircle(dst, cx, cy, radius, k*.65, hdGold, true)
	return true
}
