package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/town"
	"github.com/wicanr2/kol-remake-cht/internal/ui"
)

// renderBattle 共用同一個縮放出口，避免部分子頁仍只佔繁中畫布四分之一。
func (g *game) renderBattle() error {
	dst := g.canvas.Image()
	w, h := dst.Bounds().Dx(), dst.Bounds().Dy()
	// 中文模式的遭遇、戰場、戰利品與法術清單都直接畫在放大後的畫布上，文字才會
	// 使用中文字模；英文模式仍走下面的 320×200 renderer。
	if g.strings != nil && g.quest != nil && g.quest.state != nil {
		s := g.quest.state
		switch {
		case s.Stage == battle.StageEncounter:
			return ui.BattleEncounterCJK(dst, g.sprites, g.cjk, s, g.quest.lines)
		case s.Stage == battle.StageSpoils:
			return ui.BattleSpoilsCJK(dst, g.sprites, g.quest.tiles, g.cjk, g.strings,
				s, town.UISegments(g.quest.bank, g.quest.trans, town.BankUI, s.SpoilTitle()),
				&g.battleBody)
		case s.Plan.Prompt == battle.PromptSpell || s.Plan.Prompt == battle.PromptAffect:
			return ui.BattleSpellListCJK(dst, g.sprites, g.quest.tiles, g.cjk, g.strings,
				s, town.UISegments(g.quest.bank, g.quest.trans, town.BankUI, s.SpellListTitle()),
				g.battleSpellRows(), &g.battleBody)
		default:
			return ui.BattleFieldCJK(dst, g.sprites, g.quest.tiles, g.cjk, g.strings,
				s, &g.battleBody)
		}
	}
	if w != 320 || h != 200 {
		if g.battleCanvas == nil {
			g.battleCanvas = ebiten.NewImage(320, 200)
		}
		dst = g.battleCanvas
		dst.Clear()
	}
	s := g.quest.state
	var err error
	switch {
	case s.Stage == battle.StageEncounter:
		err = ui.BattleEncounter(dst, g.sprites, g.small, s, g.quest.lines)
	case s.Plan.Prompt == battle.PromptSpell || s.Plan.Prompt == battle.PromptAffect:
		err = ui.BattleSpellList(dst, g.sprites, g.quest.tiles, g.small, s,
			town.UISegments(g.quest.bank, g.quest.trans, town.BankUI, s.SpellListTitle()), g.battleSpellRows(), &g.battleBody, &g.battleSheet)
	case s.Stage == battle.StageSpoils:
		err = ui.BattleSpoils(dst, g.sprites, g.quest.tiles, g.small, s,
			town.UISegments(g.quest.bank, g.quest.trans, town.BankUI, s.SpoilTitle()), &g.battleBody, &g.battleSheet)
	default:
		err = ui.BattleField(dst, g.sprites, g.quest.tiles, g.small, s, &g.battleBody, &g.battleSheet)
	}
	if err != nil {
		return err
	}
	if dst != g.canvas.Image() {
		var options ebiten.DrawImageOptions
		options.GeoM.Scale(float64(w)/320, float64(h)/200)
		options.Filter = ebiten.FilterNearest
		g.canvas.Image().DrawImage(dst, &options)
	}
	return nil
}
