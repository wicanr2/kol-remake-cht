package main

import (
	"strings"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/rng"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

func TestSpellReactionDamagePageThenKnockback(t *testing.T) {
	bank, err := town.LoadUIText("../../assets/original-derived-data/ui-text.json")
	if err != nil {
		t.Fatal(err)
	}
	trans, err := loadTranslations("../../docs/i18n/town-zh-Hant.json", "../../docs/i18n/ui-zh-Hant.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, tr := range []*town.Translations{nil, trans} {
		for _, direct := range []bool{false, true} {
			t.Run(map[bool]string{true: "zh", false: "en"}[tr != nil]+map[bool]string{true: "-direct", false: "-manual"}[direct], func(t *testing.T) {
				item := &battle.Item{Name: "Fixture", Record: "b9a1f2a027028644105528dc139c9004"}
				random := rng.NewState(3884216597)
				if direct {
					random = rng.New(1)
				}
				s := &battle.State{Rand: random, Party: []battle.Member{
					{Name: "CASTER", Spells: []battle.Spell{{Name: "Kelnalon", Effect: 6, PowerCode: 156}}},
					{Name: "TARGET", HasWeapon: true, WeaponName: item.Name, HandItem: item},
				}, Actors: []battle.Actor{
					{Name: "CASTER", Side: battle.SideParty, Member: 0, Negative: true, Command: battle.CmdCast, Size: 255, X: 10, Y: 10, Head: 16, Chest: 16, Threshold: 109},
					{Name: "TARGET", Side: battle.SideParty, Member: 1, Command: battle.CmdRest, Size: 1, X: 10, Y: 11, Head: 16, Chest: 16, RightArm: 21, LeftArm: 21, RightLeg: 32, LeftLeg: 32, Threshold: 109, Weapon: battle.Weapon{Kind: 19}},
				}}
				s.Plan.Prompt = battle.PromptExecuting
				s.Plan.Memo = map[int]battle.Cell{0: {Row: 10, Col: 11}}
				s.Exec.Order = []int{0, 1}
				g := &game{quest: &questBattle{state: s, bank: bank, trans: tr}, arenaFighter: -1}
				if direct {
					g.battleStep()
					if len(s.Exec.Pending) != 1 {
						t.Fatal("直接施法遺失傷害頁")
					}
					g.battleStep()
				} else {
					s.Exec.Turn = 1
					s.Area = &battle.AreaQuestion{Caster: 0, Target: 1, Power: 27, Negative: true}
					s.Plan.Prompt = battle.PromptArea
					if err := g.applyBattleToken(s, "5", func(*battle.State) { t.Fatal("部位回答推進下一手") }); err != nil {
						t.Fatal(err)
					}
				}
				damagePage := strings.Join(s.Exec.Text, " ")
				if !strings.Contains(damagePage, "TARGET") || random.State() != 3884216597 || s.Actors[1].RightLeg != 5 || s.Actors[1].X != 10 || s.Actors[1].Y != 11 || s.Party[1].HandItem != item {
					t.Fatalf("傷害頁前已擊退或掉落：%s", damagePage)
				}
				g.battleStep()
				reactionPage := strings.Join(s.Exec.Text, " ")
				// RE-222：索引6第二候選方向4，先檢查11/11，成功後更新10/12。
				if !strings.Contains(reactionPage, "TARGET") || !strings.Contains(reactionPage, item.Name) || strings.Contains(reactionPage, "CASTER") || random.State() != 1030492215 || s.Actors[1].X != 10 || s.Actors[1].Y != 12 || s.Party[1].HandItem != nil || s.Ground[0].Item != item || s.Exec.Turn != 1 {
					t.Fatalf("擊退頁順序／名稱／狀態不符：%s", reactionPage)
				}
				g.battleStep()
				if s.Exec.Turn != 2 {
					t.Fatal("反應重複或未推進下一名角色")
				}
				t.Log(damagePage, reactionPage)
			})
		}
	}
}
