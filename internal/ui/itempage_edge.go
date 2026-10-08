package ui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

// RE-178：只保留一欄GPU暫存，沒有每幀讀回或持久遊戲資料。
// Ebitengine的繪製由同一goroutine依序提交；來源複製先於框線。
var itemEdgeScratch *ebiten.Image
var itemEdgeShader *ebiten.Shader

var itemEdgePalette = []float32{
	0, 0, 0, 0, 0, 170.0 / 255, 0, 170.0 / 255, 0, 0, 170.0 / 255, 170.0 / 255,
	170.0 / 255, 0, 0, 170.0 / 255, 0, 170.0 / 255, 170.0 / 255, 85.0 / 255, 0, 170.0 / 255, 170.0 / 255, 170.0 / 255,
	85.0 / 255, 85.0 / 255, 85.0 / 255, 85.0 / 255, 85.0 / 255, 1, 85.0 / 255, 1, 85.0 / 255, 85.0 / 255, 1, 1,
	1, 85.0 / 255, 85.0 / 255, 1, 85.0 / 255, 1, 1, 1, 85.0 / 255, 1, 1, 1,
}

const itemEdgeShaderSource = `//kage:unit pixels
package main

var Palette [16]vec3
var PlaneRows [4]vec2
var FirstRow float

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	c := imageSrc0At(srcPos)
	index := 0.0
	for i := 0; i < 16; i++ {
		d := c.rgb - Palette[i]
		if dot(d, d) < 0.000001 { index = float(i) }
	}
	planes := vec4(index, floor(index / 2), floor(index / 4), floor(index / 8))
	bits := planes - 2 * floor(planes / 2)
	y := FirstRow + srcPos.y - imageSrc0Origin().y
	if y >= PlaneRows[0].x && y < PlaneRows[0].y { bits.x = 1 }
	if y >= PlaneRows[1].x && y < PlaneRows[1].y { bits.y = 1 }
	if y >= PlaneRows[2].x && y < PlaneRows[2].y { bits.z = 0 }
	if y >= PlaneRows[3].x && y < PlaneRows[3].y { bits.w = 0 }
	modified := int(dot(bits, vec4(1, 2, 4, 8)))
	return vec4(Palette[modified], 1)
}
`

func itemEdgeSource(dst *ebiten.Image, p *town.ItemPage) (*ebiten.Image, error) {
	shown := false
	for y := range p.RestoredEdge {
		shown = shown || itemEdgeVisible(p, y)
	}
	if !shown {
		return nil, nil
	}
	if itemEdgeShader == nil {
		shader, err := ebiten.NewShader([]byte(itemEdgeShaderSource))
		if err != nil {
			return nil, err
		}
		itemEdgeShader = shader
	}
	if itemEdgeScratch == nil {
		itemEdgeScratch = ebiten.NewImage(1, len(p.RestoredEdge))
	}
	x := town.ItemPageLayout.Frame.X0
	itemEdgeScratch.Clear()
	itemEdgeScratch.DrawImage(dst.SubImage(image.Rect(x, 0, x+1, len(p.RestoredEdge))).(*ebiten.Image), nil)
	return itemEdgeScratch, nil
}

func drawRestoredItemEdge(dst, source *ebiten.Image, p *town.ItemPage) {
	for first := 0; first < len(p.RestoredEdge); {
		if !itemEdgeVisible(p, first) {
			first++
			continue
		}
		last := first + 1
		for last < len(p.RestoredEdge) && itemEdgeVisible(p, last) {
			last++
		}
		op := &ebiten.DrawRectShaderOptions{}
		op.GeoM.Translate(float64(town.ItemPageLayout.Frame.X0), float64(first))
		op.Images[0] = source.SubImage(image.Rect(0, first, 1, last)).(*ebiten.Image)
		op.Uniforms = map[string]any{
			"Palette": itemEdgePalette,
			// 每段156*40+780h=8160，平面間距8000，故依序偏0／4／8／12列。
			"PlaneRows": []float32{0, 156, 4, 160, 8, 164, 12, 168},
			"FirstRow":  float32(first),
		}
		dst.DrawRectShader(1, last-first, itemEdgeShader, op)
		first = last
	}
}

// 完整清單後的sub_1103F(2)重畫圖示底板（RE-178返回位址0x144BA），
// 覆蓋快取；目前完整背景已含同一底板。
// 以正式矩形保留這次覆蓋，不能讓快取的第四平面再改到其前四列。
func itemEdgeVisible(p *town.ItemPage, y int) bool {
	b := town.ItemPageLayout.ControlBack
	x := town.ItemPageLayout.Frame.X0
	covered := x >= b.X0 && x <= b.X1 && y >= b.Y0 && y <= b.Y1
	return p.RestoredEdge[y] && !covered
}
