package main

import (
	"image"
	"reflect"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/input"
	"github.com/wicanr2/kol-remake-cht/internal/screen"
	"github.com/wicanr2/kol-remake-cht/internal/ui"
)

// 用實際命中區驗證呈現座標縮放後仍選到相同指令，含邊界與畫布外。
func TestHDInputKeepsTargets(t *testing.T) {
	g := game{hd: &ui.HD{}}
	targets := append(portraitTargets(6, true), input.Target{Rect: image.Rect(96, 510, 168, 576), Tokens: []string{"RET"}})
	for _, point := range [][2]int{{42, 3}, {170, 122}, {171, 122}, {96, 510}, {167, 575}, {168, 576}, {-1, 3}, {959, 599}} {
		want := input.Hit(targets, point[0], point[1], "")
		for dx := 0; dx < 2; dx++ {
			for dy := 0; dy < 2; dy++ {
				x, y := g.logicalPoint(point[0]*2+dx, point[1]*2+dy)
				if got := input.Hit(targets, x, y, ""); !reflect.DeepEqual(got, want) {
					t.Fatalf("%v 偏移%d,%d：%v / %v", point, dx, dy, got, want)
				}
			}
		}
	}
	plain := game{}
	if x, y := plain.logicalPoint(42, 3); x != 42 || y != 3 {
		t.Fatal("EGA命中區不得改動")
	}
}

func TestGraphicsSwitchIsAtomicAndReusesPack(t *testing.T) {
	g := game{canvas: screen.NewCanvasScaled(3)}
	if err := g.enableHD("ega", "../../assets/hd/first-slice", "zh-Hant"); err != nil {
		t.Fatal(err)
	}
	base := g.outputCanvas()
	if err := g.setGraphics("hd"); err != nil {
		t.Fatal(err)
	}
	hd, canvas := g.hd, g.outputCanvas()
	if canvas.Image().Bounds() != image.Rect(0, 0, 1920, 1200) {
		t.Fatal("HD畫布尺寸錯誤")
	}
	if err := g.setGraphics("invalid"); err == nil || g.hd != hd || g.outputCanvas() != canvas {
		t.Fatal("錯誤選項破壞既有呈現")
	}
	if err := g.setGraphics("ega"); err != nil || g.hd != nil || g.outputCanvas() != base {
		t.Fatal("回到EGA未沿用原畫布")
	}
	// 已驗過的完整素材包常駐；磁碟路徑不再可用也不重新讀半套素材。
	g.hdDir = t.TempDir()
	if err := g.setGraphics("hd"); err != nil || g.hd != hd || g.outputCanvas() != canvas {
		t.Fatal("切換重複載入素材或重建畫布")
	}
	broken := game{canvas: base}
	if err := broken.enableHD("ega", t.TempDir(), "zh-Hant"); err != nil {
		t.Fatal(err)
	}
	if err := broken.setGraphics("hd"); err == nil || broken.hd != nil || broken.outputCanvas() != base {
		t.Fatal("缺失素材包不得改變EGA")
	}
}
