package main

import (
	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"strings"

	"github.com/wicanr2/kol-remake-cht/internal/flow"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

func (g *game) renderNativeUI() error {
	original := g.canvas
	g.canvas = g.presentation
	defer func() { g.canvas = original }()
	return g.hd.RenderUI(g.presentation.Image(), g.render)
}

// nativeUIPage 只選本批已驗的完整重畫頁。其他分支依各自HD契約續作。
func (g *game) nativeUIPage() bool {
	if g.endingReached() {
		return false
	}
	playPage := func(s *town.Session) bool {
		if s == nil {
			return false
		}
		if g.quest != nil && g.quest.state != nil {
			return g.quest.state.ResultPicture() != battle.ResultNone
		}
		if !s.InArena() && !s.ShowingSheetOutside() && s.LeftTown && g.worldState != nil {
			return true
		}
		return (s.InArena() || s.ShowingSheetOutside() || s.Mode == town.ModeInside) && nativeInnPage(s.Inside)
	}
	switch g.scene {
	case sceneContents, sceneMenu, sceneRoster, sceneDelete, sceneHarbor, sceneScriptCJK, sceneWorld:
		return true
	case sceneInn, sceneBuilding:
		return nativeInnPage(g.inn)
	case scenePlay:
		return playPage(g.play)
	case sceneFlow:
		if g.flow == nil {
			return false
		}
		switch g.flow.Scene {
		case flow.SceneContents, flow.SceneMenu, flow.SceneRoster, flow.SceneNewChar, flow.SceneDelete,
			flow.SceneHarbor, flow.SceneSelect, flow.SceneEditPick, flow.SceneEditor:
			return true
		case flow.SceneInn:
			return nativeInnPage(g.flow.Inn)
		case flow.SceneTown:
			if g.leftTown() {
				return playPage(g.play)
			}
			return playPage(g.flow.Town)
		}
	}
	return false
}

func nativeInnPage(in *town.Inn) bool {
	if in == nil {
		return false
	}
	if in.SpellEdit != nil {
		return true
	}
	if in.ItemPage != nil || len(in.PartyPage) > 0 {
		return true
	}
	if in.Sheet != nil || in.List != nil || in.Steed != nil || len(in.Detail) > 0 || in.DialogRow != 0 {
		return true
	}
	lines := in.DialogZh
	if len(lines) == 0 {
		lines = in.Dialog
	}
	if len(in.Title) > 2 || len(lines) > 9 {
		// 多行及其他標題版面仍待獨立驗收，沿原呈現。
		return false
	}
	for _, group := range [][]string{in.Dialog, in.DialogZh} {
		for _, line := range group {
			if strings.Contains(line, town.TextCaret) {
				return true
			}
		}
	}
	return false
}
