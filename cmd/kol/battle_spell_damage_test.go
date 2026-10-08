package main

import (
	"strings"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

// RE-206：使用正常部位回答入口與實際雙語文字庫，確認真實目標及傷勢。
func TestSpellAreaDamageMessage(t *testing.T) {
	bank, err := town.LoadUIText("../../assets/original-derived-data/ui-text.json")
	if err != nil {
		t.Fatal(err)
	}
	trans, err := loadTranslations("../../docs/i18n/town-zh-Hant.json", "../../docs/i18n/ui-zh-Hant.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		name  string
		tr    *town.Translations
		wants []string
	}{
		{"en", nil, []string{"The spell hits TARGET's chest", "serious", "wound!"}},
		{"zh", trans, []string{"法術", "TARGET", "胸", "傷"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			rng := &areaPromptRand{}
			s := &battle.State{Rand: rng, Actors: []battle.Actor{
				{Name: "CASTER", Side: battle.SideParty, Member: 0, Negative: true},
				{Name: "TARGET", Side: battle.SideParty, Member: 1, Head: 16, Chest: 16, Threshold: 100},
			}, Party: []battle.Member{{Name: "CASTER"}, {Name: "TARGET"}},
				Area: &battle.AreaQuestion{Caster: 0, Target: 1, Power: 11, Negative: true}}
			s.Plan.Prompt = battle.PromptArea
			g := &game{quest: &questBattle{state: s, bank: bank, trans: c.tr}, arenaFighter: -1}
			if err := g.applyBattleToken(s, "2", func(*battle.State) { t.Fatal("回答推進了另一手") }); err != nil {
				t.Fatal(err)
			}
			got := strings.Join(s.Exec.Text, " ")
			for _, want := range c.wants {
				if !strings.Contains(got, want) {
					t.Errorf("缺 %q：%q", want, got)
				}
			}
			if len(rng.calls) != 0 || s.Actors[1].Chest != 5 || s.Actors[1].Damage != 11 || s.Actors[1].Hurt != 11 {
				t.Fatal("訊息修正改動傷害或擲骰")
			}
			t.Log(got)
		})
	}
}

func TestSpellAreaDamageBranches(t *testing.T) {
	bank, err := town.LoadUIText("../../assets/original-derived-data/ui-text.json")
	if err != nil {
		t.Fatal(err)
	}
	trans, err := loadTranslations("../../docs/i18n/town-zh-Hant.json", "../../docs/i18n/ui-zh-Hant.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		name                         string
		power, area, threshold, part int
		monster, shatters            bool
		en, zh                       string
	}{
		{"zero", 0, 1, 100, 30, false, false, "little damage", "輕微"},
		{"party-down", 11, 1, 10, 30, false, false, "collapse", "倒了下去"},
		{"monster-down", 11, 1, 10, 30, true, false, "die", "死去"},
		{"shatter", 11, 1, 10, 30, true, true, "shatter into pieces", "碎成了片片"},
		{"limb", 11, 2, 100, 10, false, false, "limb useless", "肢體"},
		{"head-party", 11, 0, 100, 10, false, false, "unconscious", "昏了過去"},
		{"head-monster", 11, 0, 100, 10, true, false, "killing it instantly", "當場"},
		{"terrible", 16, 1, 100, 30, false, false, "terrible", "可怕"},
		{"grievous", 24, 1, 100, 30, false, false, "grievous", "致命"},
	} {
		for _, tr := range []*town.Translations{nil, trans} {
			t.Run(c.name+map[bool]string{true: "-zh", false: "-en"}[tr != nil], func(t *testing.T) {
				a := battle.Actor{Name: "TARGET", Side: battle.SideParty, Member: 1, Threshold: c.threshold, Head: c.part, Chest: c.part, RightArm: c.part, LeftArm: c.part, RightLeg: c.part, LeftLeg: c.part, Shatters: c.shatters}
				if c.monster {
					a.Side = battle.SideMonster
					a.Member = -1
				}
				s := &battle.State{Actors: []battle.Actor{{Name: "CASTER", Side: battle.SideParty, Member: 0, Negative: true}, a}, Party: []battle.Member{{Name: "CASTER"}, {Name: "TARGET"}}, Area: &battle.AreaQuestion{Caster: 0, Target: 1, Power: c.power, Negative: true}}
				g := &game{quest: &questBattle{state: s, bank: bank, trans: tr}, arenaFighter: -1}
				res, caster, ok := s.AnswerArea(c.area)
				if !ok || res.Target != 1 || res.Damage == nil {
					t.Fatal("遺失实际目標或傷害結果")
				}
				g.battleAreaLine(s, res, caster)
				got := strings.Join(s.Exec.Text, " ")
				want := c.en
				if tr != nil {
					want = c.zh
				}
				if !strings.Contains(got, want) {
					t.Fatalf("期望%q：%q", want, got)
				}
				t.Log(got)
			})
		}
	}
}

// 遠距選格的正式執行入口直接套部位，也必須留下傷害頁。骰序另由獨立任務追蹤。
func TestSpellDirectDamageMessage(t *testing.T) {
	bank, err := town.LoadUIText("../../assets/original-derived-data/ui-text.json")
	if err != nil {
		t.Fatal(err)
	}
	rng := &areaPromptRand{}
	s := &battle.State{Rand: rng, Party: []battle.Member{{Name: "CASTER", Spells: []battle.Spell{{Name: "Kelnalon", Effect: 6, PowerCode: 128 | 10}}}, {Name: "TARGET"}}, Actors: []battle.Actor{
		{Name: "CASTER", Side: battle.SideParty, Member: 0, Negative: true, Command: battle.CmdCast, Head: 30, Chest: 30, Threshold: 100, Stamina: 100, X: 5, Y: 5},
		{Name: "TARGET", Side: battle.SideParty, Member: 1, Head: 30, Chest: 30, Threshold: 100, Stamina: 100, X: 10, Y: 10},
	}}
	s.Plan.Memo = map[int]battle.Cell{0: {Row: 10, Col: 10}}
	s.Plan.Prompt = battle.PromptExecuting
	s.Exec.Order = []int{0}
	g := &game{quest: &questBattle{state: s, bank: bank}, arenaFighter: -1}
	g.battleStep()
	if s.Area != nil || len(s.Exec.Pending) != 1 {
		t.Fatal("直接部位法術未留下結果頁")
	}
	got := strings.Join(s.Exec.Pending[0], " ")
	if !strings.Contains(got, "The spell hits TARGET's head") || !strings.Contains(got, "serious wound") {
		t.Fatal(got)
	}
	if s.Actors[1].Head != 29 {
		t.Fatal("沒有套用實際部位傷害")
	}
	g.battleStep()
	if strings.Join(s.Exec.Text, " ") != got || len(s.Exec.Pending) != 0 {
		t.Fatal("結果頁等待層錯誤")
	}
}

func TestSpellGenericMessagePreserved(t *testing.T) {
	bank, err := town.LoadUIText("../../assets/original-derived-data/ui-text.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, effect := range []int{1, 5, 6} {
		for _, negative := range []bool{false, true} {
			if effect == 6 && negative {
				continue
			}
			s := &battle.State{Actors: []battle.Actor{{Name: "CASTER", Side: battle.SideParty, Member: 0, Negative: negative}, {Name: "TARGET", Side: battle.SideParty, Member: 1, Head: 10, Chest: 10, Threshold: 100}}, Party: []battle.Member{{Name: "CASTER", Spells: []battle.Spell{{Name: "Kelnalon", Effect: effect, PowerCode: 10}}}, {Name: "TARGET"}}}
			res := s.Cast(0, 1, 1, &areaPromptRand{})
			g := &game{quest: &questBattle{state: s, bank: bank}}
			got := strings.Join(g.battleSpellResultLine(s, res, 0), " ")
			if res.Damage != nil || !strings.Contains(got, "The spell has a mildly") {
				t.Fatal(got)
			}
		}
	}
}
