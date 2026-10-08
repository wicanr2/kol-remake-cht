package main

import (
	"reflect"
	"strings"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/book"
	"github.com/wicanr2/kol-remake-cht/internal/input"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

type areaPromptRand struct{ calls []int }

func (r *areaPromptRand) Roll(n int) int { r.calls = append(r.calls, n); return 1 }

// 從實際施法入口驗證可見問句、無效鍵、傷害與回答後的訊息生命週期。
func TestAreaPromptVisibleBeforeItAcceptsNumbers(t *testing.T) {
	bank, err := town.LoadUIText("../../assets/original-derived-data/ui-text.json")
	if err != nil {
		t.Fatal(err)
	}
	rng := &areaPromptRand{}
	s := &battle.State{Party: []battle.Member{{Name: "CASTER", Spells: []battle.Spell{{Name: "Kelnalon", Effect: 6, TurnCode: 0, PowerCode: 10}}}}, Rand: rng}
	s.Actors = []battle.Actor{
		{Name: "CASTER", Side: battle.SideParty, Member: 0, Head: 10, Chest: 10, Threshold: 100, Stamina: 100, X: 5, Y: 5, Command: battle.CmdCast, Facing: 0, Negative: true},
		{Name: "TARGET", Side: battle.SideMonster, Member: -1, Head: 10, Chest: 10, Threshold: 100, Stamina: 100, X: 4, Y: 5},
	}
	s.Exec.Order = []int{0}
	s.Plan.Prompt = battle.PromptExecuting
	g := &game{quest: &questBattle{state: s, bank: bank}, arenaFighter: -1}
	g.battleStep()
	if s.Plan.Prompt != battle.PromptArea || s.Area == nil {
		t.Fatalf("未到問句：%+v", s.Plan)
	}
	text := strings.Join(s.Exec.Text, " ")
	for _, want := range []string{"Select an area to", "hurt", "1 head", "2 chest", "6 left leg"} {
		if !strings.Contains(text, want) {
			t.Fatalf("可接受輸入前缺提示 %q：%q", want, text)
		}
	}
	if len(s.Exec.Pending) != 0 {
		t.Fatal("問句仍藏在 Return 佇列")
	}
	calls := append([]int(nil), rng.calls...)
	for _, key := range []string{"RET", "0", "7"} {
		if err := g.applyBattleToken(s, key, func(*battle.State) { t.Fatal("無效鍵推進執行") }); err != nil {
			t.Fatal(err)
		}
		if s.Plan.Prompt != battle.PromptArea || s.Actors[1].Chest != 10 || strings.Join(s.Exec.Text, " ") != text {
			t.Fatal("無效鍵改動待答狀態")
		}
	}
	if err := g.applyBattleToken(s, "2", func(*battle.State) { t.Fatal("回答時多執行一手") }); err != nil {
		t.Fatal(err)
	}
	if s.Actors[1].Chest != 9 || s.Plan.Prompt != battle.PromptExecuting || !reflect.DeepEqual(calls, rng.calls) {
		t.Fatal("回答未只套用已擲完的胸部傷害")
	}
	if len(s.Exec.Pending) != 0 || strings.Contains(strings.Join(s.Exec.Text, " "), "Select an area") {
		t.Fatal("回答後仍殘留問句")
	}
}

func spellListGame(t *testing.T) (*game, *battle.State) {
	t.Helper()
	s := &battle.State{Stage: battle.StageField, Field: battle.NewField(),
		Actors: []battle.Actor{{Side: battle.SideParty, Member: 0, Threshold: 100, Head: 15, Chest: 15}},
		Party:  []battle.Member{{Name: "PLAYER", PackCount: 1, Spells: []battle.Spell{{Name: "Kelnalon", Effect: 6, PowerCode: 10}}}},
		Round:  battle.Round{Order: []int{0}},
		Plan:   battle.Planning{Prompt: battle.PromptCommand, Bar: []int{battle.CmdYes, battle.CmdCast}, Selected: 1}}
	g := &game{quest: &questBattle{state: s}, arenaFighter: -1}
	spellListKeys(t, g, s, "RET", "RET")
	if s.Plan.Prompt != battle.PromptAffect {
		t.Fatal("沒有進入效果清單")
	}
	return g, s
}

func spellListKeys(t *testing.T, g *game, s *battle.State, keys ...string) {
	t.Helper()
	for _, key := range keys {
		if err := g.applyBattleToken(s, key, func(*battle.State) { t.Fatal("清單操作不能執行一手") }); err != nil {
			t.Fatal(err)
		}
	}
}

func TestSpellListMouseChoosesNegative(t *testing.T) {
	for _, cjk := range []bool{false, true} {
		for _, selected := range []int{0, 1} {
			g, s := spellListGame(t)
			s.Plan.Selected = selected
			l := battle.SpellListRows
			l.Row = s.SpellListRow()
			scale := 1
			if cjk {
				scale = book.CJKScale
			}
			keys := input.Hit(battleTargets(s, cjk), (l.Col*town.CellWidth+1)*scale, (l.Row+l.LineStep+1)*scale, "")
			spellListKeys(t, g, s, keys...)
			if !s.Actors[0].Negative || s.Plan.Prompt != battle.PromptTarget {
				t.Fatalf("中文=%v 圖示=%d：滑鼠負面列未選中，keys=%v prompt=%v negative=%v", cjk, selected, keys, s.Plan.Prompt, s.Actors[0].Negative)
			}
		}
	}
}

func TestSpellListCancelReturnsToPreviousLayer(t *testing.T) {
	g, s := spellListGame(t)
	s.Actors[0].Command, s.Actors[0].Negative = battle.CmdMove, true
	spellListKeys(t, g, s, ">", "RET")
	if s.Plan.Prompt != battle.PromptSpell || !s.Actors[0].Negative {
		t.Fatal("效果清單取消未返回法術清單，或改寫先前效果方向")
	}
	spellListKeys(t, g, s, ">", "RET")
	if s.Plan.Prompt != battle.PromptCommand || s.Actors[0].Command != battle.CmdMove {
		t.Fatal("法術清單取消未返回主列，或改寫先前指令")
	}
}

func TestSpellListKeyboardSeparatesRowsAndIcons(t *testing.T) {
	g, s := spellListGame(t)
	spellListKeys(t, g, s, "D", "D")
	if s.Plan.ListPick != 1 || s.Plan.Selected != 0 {
		t.Fatal("下界未夾限，或方向鍵改了圖示")
	}
	spellListKeys(t, g, s, "U", "U", "R", "L")
	if s.Plan.ListPick != 0 {
		t.Fatal("上界或左右鍵改了文字選項")
	}
	spellListKeys(t, g, s, ">", "<")
	if s.Plan.ListPick != 0 || s.Plan.Selected != 0 {
		t.Fatal("圖示鍵錯改文字選項")
	}
	spellListKeys(t, g, s, "D", "RET")
	if !s.Actors[0].Negative || s.Plan.Prompt != battle.PromptTarget {
		t.Fatal("方向鍵未確認負面")
	}
}

func TestSpellListMouseCancelsBothLayers(t *testing.T) {
	for _, cjk := range []bool{false, true} {
		g, s := spellListGame(t)
		scale := 1
		if cjk {
			scale = book.CJKScale
		}
		b := town.IconBar
		for _, want := range []battle.Prompt{battle.PromptSpell, battle.PromptCommand} {
			keys := input.Hit(battleTargets(s, cjk), (b.X+b.Step+1)*scale, (b.Y+1)*scale, "")
			if len(keys) == 0 {
				t.Fatal("取消圖示沒有點擊區域")
			}
			spellListKeys(t, g, s, keys...)
			if s.Plan.Prompt != want {
				t.Fatalf("中文=%v：取消到%d，期望%d", cjk, s.Plan.Prompt, want)
			}
		}
	}
}
