package main

import (
	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/rng"
	"github.com/wicanr2/kol-remake-cht/internal/town"
	"strings"
	"testing"
)

func TestSpellLifecycleExpiryPageBeforeExit(t *testing.T) {
	bank, err := town.LoadUIText("../../assets/original-derived-data/ui-text.json")
	if err != nil {
		t.Fatal(err)
	}
	trans, err := loadTranslations("../../docs/i18n/town-zh-Hant.json", "../../docs/i18n/ui-zh-Hant.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, tr := range []*town.Translations{nil, trans} {
		for _, edge := range []bool{false, true} {
			t.Run(map[bool]string{true: "zh", false: "en"}[tr != nil]+map[bool]string{true: "-flee", false: "-victory"}[edge], func(t *testing.T) {
				s := &battle.State{Rand: rng.New(1), RankFight: true, Party: []battle.Member{{Name: "TARGET", Strength: 100}}, Actors: make([]battle.Actor, 7)}
				for i := range s.Actors {
					s.Actors[i].Side = battle.SideEmpty
				}
				s.Actors[0] = battle.Actor{Name: "TARGET", Side: battle.SideParty, Member: 0, Strength: 150, SpellTurns: 1, Head: 16, Chest: 16, Stamina: 200, X: 10, Y: 10}
				s.Actors[6] = battle.Actor{Side: battle.SideMonster, Member: -1, Head: 0, Chest: 16, X: 12, Y: 10}
				if edge {
					s.Actors[0].X = battle.FieldMin
					s.Actors[6].Head = 16
				}
				s.Plan.Prompt = battle.PromptExecuting
				g := &game{quest: &questBattle{state: s, bank: bank, trans: tr}, arenaFighter: -1}
				g.battleStep()
				page := strings.Join(s.Exec.Text, " ")
				word := "spell wears off"
				if tr != nil {
					word = "法術"
				}
				if !strings.Contains(page, "TARGET") || !strings.Contains(page, word) || len(s.Exec.Pending) != 0 || s.Actors[0].SpellTurns != 0 || s.Stage == battle.StageDone || s.Plan.Prompt == battle.PromptFlee {
					t.Fatalf("失效頁或等待順序不符：%s", page)
				}
				t.Log(page)
			})
		}
	}
}
