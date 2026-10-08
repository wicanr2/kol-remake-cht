package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/book"
	"github.com/wicanr2/kol-remake-cht/internal/font"
	"github.com/wicanr2/kol-remake-cht/internal/title"
	"github.com/wicanr2/kol-remake-cht/internal/town"
	"github.com/wicanr2/kol-remake-cht/internal/ui"
	"image"
)

func (g *game) advanceBattleMarker() {
	if g.quest == nil || g.quest.state == nil || g.quest.tiles == nil {
		return
	}
	if _, shown := g.quest.state.TargetMarker(); shown {
		g.quest.tiles.AdvanceTargetMarker(title.TickHz / 60)
	}
}

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
			return ui.BattleEncounterCJK(dst, g.sprites, g.cjk, s, g.battleEncounterLines())
		case s.Stage == battle.StageSpoils:
			return ui.BattleSpoilsCJK(dst, g.sprites, g.quest.tiles, g.cjk, g.strings,
				s, town.UISegments(g.quest.bank, g.quest.trans, town.BankUI, s.SpoilTitle()),
				&g.battleBody)
		case s.Plan.Prompt == battle.PromptSpell || s.Plan.Prompt == battle.PromptAffect:
			return ui.BattleSpellListCJK(dst, g.sprites, g.quest.tiles, g.cjk, g.strings,
				s, town.UISegments(g.quest.bank, g.quest.trans, town.BankUI, s.SpellListTitle()),
				g.battleSpellRows(), &g.battleBody)
		default:
			if err := ui.BattleFieldCJK(dst, g.sprites, g.quest.tiles, g.cjk, g.strings,
				s, &g.battleBody); err != nil {
				return err
			}
			return g.drawNextRoundQuestion(dst, g.cjk, book.CJKScale)
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
	g.battleSheet.NameFont = g.large
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
	if err := g.drawNextRoundQuestion(dst, g.large, 1); err != nil {
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

// battleEncounterLines 只在繁中繪圖時取任務遭遇記錄的譯文。
// quest.lines 保留原始文字供軌跡核對；伏擊的動態人數與怪物名已由 ambushLines 處理。
func (g *game) battleEncounterLines() []string {
	lines := g.quest.lines
	if g.quest.bank == nil || !sameLines(lines, nonEmptyLines(
		g.quest.bank.Lines(town.BankUI, battle.QuestCreaturesRecord))) {
		return lines
	}
	segments := town.UISegments(g.quest.bank, g.quest.trans, town.BankUI,
		battle.QuestCreaturesRecord)
	out := make([]string, 0, len(segments))
	for _, segment := range segments {
		out = append(out, segment.Text)
	}
	return nonEmptyLines(out)
}

// drawNextRoundQuestion 補原版 0x1FB30 的左上問句，保持文字庫中的大小寫與座標。
// 只在待答確認頁畫；原版 0x1FB9B 在選擇後清框，執行段不保留這句。
func (g *game) drawNextRoundQuestion(dst *ebiten.Image, f *font.Font, scale int) error {
	if g.quest == nil || g.quest.state == nil {
		return nil
	}
	s := g.quest.state
	if s.Stage != battle.StageField || s.Plan.Prompt != battle.PromptNextRound {
		return nil
	}
	segments := town.UISegments(g.quest.bank, g.quest.trans, town.BankUI, battle.NextRoundRecord)
	if len(segments) == 0 {
		return fmt.Errorf("下一回合問句缺少 KOLDATA 記錄 0x108")
	}
	b := battle.SheetBox
	safe := image.Rect((b.X+1)*scale, (b.Y+1)*scale, (b.X+b.Width-1)*scale, (b.Y+b.Height-1)*scale)
	clip := dst.SubImage(safe).(*ebiten.Image)
	for _, seg := range segments {
		x, y := seg.Col*town.CellWidth*scale, seg.Row*scale
		w, h := f.Measure(seg.Text)
		if !image.Rect(x, y, x+w, y+h).In(safe) {
			return fmt.Errorf("下一回合問句超出左上文字安全區")
		}
		ink := town.ColourText
		if drawn := f.Draw(clip, x, y, seg.Text, ink[0], ink[1], ink[2]); drawn != len([]rune(seg.Text)) {
			return fmt.Errorf("下一回合問句含字型未涵蓋的字元")
		}
	}
	return nil
}
