package ui

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/kol-remake-cht/internal/battle"
)

const bodyX, bodyY, bodyWidth, bodyHeight = 72, 85, 56, 70

// BattleBody 只快取三張不可變底圖與一張顯示結果；不隨角色／傷勢累積圖片。
type BattleBody struct {
	source [3]*ebiten.Image
	base   [3][]byte
	image  *ebiten.Image
	pixels []byte
	key    battle.BodyPanelData
	valid  bool
}

func (b *BattleBody) Draw(dst *ebiten.Image, sprites Sprites, s *battle.State) error {
	data, show := s.BodyPanel()
	if !show {
		return nil
	}
	rects, err := battle.BodyDamageRects(data.Kind, data.Current, data.Maximum)
	if err != nil {
		return err
	}
	boundary, err := battle.BodyFatigueBoundary(data.Stamina, data.Fatigue)
	if err != nil {
		return err
	}
	src, err := sprites.Get("MISCPICS", 3+data.Kind)
	if err != nil {
		return err
	}
	if src == nil || src.Bounds().Dx() != bodyWidth || src.Bounds().Dy() != bodyHeight {
		return fmt.Errorf("MISCPICS %d 人形尺寸須為 %d×%d", 3+data.Kind, bodyWidth, bodyHeight)
	}
	if b.source[data.Kind] != src {
		b.source[data.Kind] = src
		b.base[data.Kind] = make([]byte, bodyWidth*bodyHeight*4)
		src.ReadPixels(b.base[data.Kind])
		b.valid = false
	}
	if b.image == nil {
		b.image = ebiten.NewImage(bodyWidth, bodyHeight)
		b.pixels = make([]byte, bodyWidth*bodyHeight*4)
	}
	if !b.valid || b.key != data {
		copy(b.pixels, b.base[data.Kind])
		shadeBody(b.pixels, rects)
		b.image.WritePixels(b.pixels)
		b.key, b.valid = data, true
	}
	draw(dst, b.image, bodyX, bodyY)
	// 原版兩個矩形包含端點；後畫的藍色蓋住共同分界列。
	fill(dst, 132, 84, 7, boundary-84+1, [3]uint8{255, 85, 85})
	fill(dst, 132, boundary, 7, 157-boundary, [3]uint8{0, 0, 170})
	return nil
}

func shadeBody(pixels []byte, rects [6]battle.BodyRect) {
	for _, r := range rects {
		for y := r.Top; y < r.Bottom; y++ {
			for x := r.Left; x < r.Right; x++ {
				if x < bodyX || x >= bodyX+bodyWidth || y < bodyY || y >= bodyY+bodyHeight {
					continue
				}
				p := ((y-bodyY)*bodyWidth + x - bodyX) * 4
				v := pixels[p]
				if pixels[p+1] == v && pixels[p+2] == v && (v == 85 || v == 170 || v == 255) {
					pixels[p], pixels[p+1], pixels[p+2] = 255, 85, 85
				}
			}
		}
	}
}
