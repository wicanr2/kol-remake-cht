package main

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/kol-remake-cht/internal/assets"
	"github.com/wicanr2/kol-remake-cht/internal/flow"
	"github.com/wicanr2/kol-remake-cht/internal/screen"
	"github.com/wicanr2/kol-remake-cht/internal/town"
	"github.com/wicanr2/kol-remake-cht/internal/ui"
)

func (g *game) enableHD(mode, dir, lang string) error {
	if mode == "ega" {
		return nil
	}
	if mode != "hd" || lang != "zh-Hant" {
		return fmt.Errorf("graphics須為ega或hd；HD首批只支援zh-Hant")
	}
	lib, err := assets.LoadHD(dir)
	if err != nil {
		return fmt.Errorf("HD素材載入失敗：%w", err)
	}
	g.hd = ui.NewHD(lib.Facade, lib.Party, lib.Manifest.FacadeRGB, lib.Manifest.PortraitRGB)
	g.presentation = screen.NewCanvasScaled(6)
	return nil
}

func (g *game) outputCanvas() *screen.Canvas {
	if g.presentation != nil {
		return g.presentation
	}
	return g.canvas
}

func (g *game) logicalPoint(x, y int) (int, int) {
	if g.hd != nil {
		return x >> 1, y >> 1
	}
	return x, y
}

func (g *game) cursorPosition() (int, int) {
	x, y := ebiten.CursorPosition()
	return g.logicalPoint(x, y)
}

func (g *game) presentHD() error {
	if g.hd == nil {
		return nil
	}
	g.presentation.Clear()
	g.presentation.DrawScaled(g.canvas.Image(), 0, 0, 2)
	dst := g.presentation.Image()
	if g.endingReached() {
		return nil
	}
	play := func(s *town.Session) error {
		if s == nil {
			return nil
		}
		if s.InArena() || s.ShowingSheetOutside() {
			return g.hd.Inn(dst, g.sprites, s.Inside)
		}
		if g.quest != nil && g.quest.state != nil {
			return g.hd.Battle(dst, g.sprites, g.quest.state)
		}
		if s.LeftTown {
			return nil
		}
		if s.Mode == town.ModeInside {
			return g.hd.Inn(dst, g.sprites, s.Inside)
		}
		return g.hd.Outdoor(dst, g.sprites, s.Outdoor)
	}
	switch g.scene {
	case sceneInn, sceneBuilding:
		return g.hd.Inn(dst, g.sprites, g.inn)
	case sceneOutdoor:
		return g.hd.Outdoor(dst, g.sprites, g.outdoor)
	case scenePlay:
		return play(g.play)
	case sceneFlow:
		if g.flow == nil {
			return nil
		}
		switch g.flow.Scene {
		case flow.SceneInn:
			return g.hd.Inn(dst, g.sprites, g.flow.Inn)
		case flow.SceneTown:
			if g.leftTown() {
				return play(g.play)
			}
			return play(g.flow.Town)
		}
	}
	return nil
}

// 提示層沿原版只畫到螢幕，不覆蓋批次截圖所用的場景畫布。
func (g *game) drawPresentationNotice(dst *ebiten.Image) {
	if g.hd == nil {
		g.drawNotice(dst)
		return
	}
	if g.noticeCanvas == nil {
		g.noticeCanvas = screen.NewCanvasScaled(3)
	}
	g.drawNotice(g.noticeCanvas.Image())
	var op ebiten.DrawImageOptions
	op.GeoM.Scale(2, 2)
	op.Filter = ebiten.FilterNearest
	dst.DrawImage(g.noticeCanvas.Image(), &op)
}
