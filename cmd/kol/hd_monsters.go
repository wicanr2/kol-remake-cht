package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/kol-remake-cht/internal/flow"
	"github.com/wicanr2/kol-remake-cht/internal/font"
	"github.com/wicanr2/kol-remake-cht/internal/screen"
	"github.com/wicanr2/kol-remake-cht/internal/ui"
	"image"
)

// presentMonsters 重畫同一底層時只隱藏本次怪物，結束前還原來源。
func (g *game) presentMonsters() (bool, error) {
	if g.quest == nil || g.quest.state == nil {
		return false, nil
	}
	if g.endingReached() || g.scene != scenePlay && !(g.scene == sceneFlow && g.flow != nil && g.flow.Scene == flow.SceneTown) || g.play != nil && g.play.InArena() {
		return false, nil
	}
	placements, err := g.hd.MonsterPlacements(g.quest.state, g.sprites)
	if err != nil || len(placements) == 0 {
		return false, err
	}
	if err := g.renderMonsterUnderlay(placements); err != nil {
		return false, err
	}
	g.presentation.Clear()
	g.presentation.DrawScaled(g.hdMonsterUnderlay.Image(), 0, 0, 2)
	g.hd.DrawMonsters(g.presentation.Image(), placements)
	return true, nil
}

func (g *game) renderMonsterUnderlay(placements []ui.HDMonsterPlacement) error {
	if g.hdMonsterBlanks == nil {
		g.hdMonsterBlanks = map[image.Point]*ebiten.Image{}
	}
	old := map[string]*ebiten.Image{}
	for _, p := range placements {
		key := fmt.Sprintf("MONPICS/%d", p.Index)
		if _, ok := old[key]; ok {
			continue
		}
		old[key] = p.Original
		size := p.Original.Bounds().Size()
		blank := g.hdMonsterBlanks[size]
		if blank == nil {
			blank = ebiten.NewImage(size.X, size.Y)
			g.hdMonsterBlanks[size] = blank
		}
		g.sprites.m[key] = blank
	}
	observer := font.DrawObserver
	font.DrawObserver = nil
	if g.hdMonsterUnderlay == nil {
		g.hdMonsterUnderlay = screen.NewCanvasScaled(3)
	}
	originalCanvas := g.canvas
	g.canvas = g.hdMonsterUnderlay
	err := func() error {
		defer func() {
			for key, pic := range old {
				g.sprites.m[key] = pic
			}
			font.DrawObserver = observer
			g.canvas = originalCanvas
		}()
		return g.render()
	}()
	if err != nil {
		return err
	}
	return nil
}
