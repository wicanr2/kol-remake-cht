package main

import (
	"encoding/hex"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/roster"
	"github.com/wicanr2/kol-remake-cht/internal/save"
	"github.com/wicanr2/kol-remake-cht/internal/town"
	"github.com/wicanr2/kol-remake-cht/internal/world"
)

// 從既有結算狀態測正式出口及檔案往返；不冒稱正常戰鬥或原版對拍。
func TestBattleSettlementCompleteSaveReload(t *testing.T) {
	for _, tc := range []struct {
		name    string
		verdict battle.Verdict
		arena   bool
	}{
		{"勝利", battle.VerdictWon, false},
		{"撤退", battle.VerdictFled, false},
		{"敗北", battle.VerdictLost, false},
		{"競技場勝利", battle.VerdictWon, true},
		{"競技場敗北", battle.VerdictLost, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			party := make([]town.Member, 3)
			r := &roster.Roster{Slots: make([]roster.Slot, 16)}
			slots := []int{7, 2, 15}
			for i, name := range []string{"ALPHA", "BRAVO", "CHARLIE"} {
				m := &party[i]
				*m = town.Member{ID: name, Name: name, Gold: 100, Wounds: 5, Speaker: i == 1,
					Quests: make([]int, 6), Progress: make([]int, 5), Sheet: &town.Sheet{Rank: 2, AdventurePoints: 50}}
				for j := range m.Sheet.Stats {
					m.Sheet.Stats[j] = 60
				}
				for j := range m.Worn {
					raw := make([]byte, 16)
					raw[0], raw[10], raw[11], raw[15] = byte(j+1), 4, 200, 0x20
					m.Worn[j] = &town.Good{Name: "GEAR", Record: hex.EncodeToString(raw), Condition: 200, Encumbrance: 4}
				}
				m.InHand = &town.Good{Name: "SWORD", Class: 0x21, Record: "01000000000000000000002121000000"}
				if i == 1 {
					// +0x0F 帶 0x20、耐久 200：超上限 7 就扣光（RE-155）。
					m.InHand = &town.Good{Name: "INGOT", Class: 0x21, Condition: 200, Record: "010000000000000000000ac821000020"}
				}
				m.SyncCarried()
				r.Slots[slots[i]] = roster.Slot{Slot: slots[i], ID: name, Name: name, Occupied: true}
			}
			members := make([]battle.Member, len(party))
			for i, m := range party {
				members[i] = battleMember(m)
			}
			s := battle.NewQuest(&battle.Site{Index: -1, PartyX: []int{1, 2, 3}, PartyY: []int{1, 1, 1}}, members, battle.Monster{}, battle.NewField())
			s.Stage, s.Verdict = battle.StageDone, tc.verdict
			s.Exec.Text = []string{"結算訊息已顯示"}
			s.Gold, s.Adventure = 11, 7
			s.RankFight = tc.arena
			for i := range party {
				if tc.arena && i != 1 {
					s.Actors[i].Side = battle.SideEmpty
				}
			}
			s.Actors[1].Damage = 24
			rng := newRand(1)
			s.Actors[1].AbsorbArmour(255, 0, rng)
			s.Actors[1].AbsorbArmour(3, 1, rng)
			if s.Actors[1].Armour[0] != nil || s.Actors[1].Armour[1] == nil || s.Actors[1].Armour[1].Condition >= 200 {
				t.Fatal("測試前置狀態未形成破損及耐久降低")
			}
			s.Actors[1].Command = battle.CmdWeapon
			s.WearOverCap(1, 7)
			if s.Party[1].HandItem != nil || s.Actors[1].Command != 0 {
				t.Fatal("測試前置狀態未形成手上那件磨光")
			}
			// 本測試驗資料跨邊界保留，不重測 SPEC-017 的骰式公式。
			wantCondition := s.Actors[1].Armour[1].Condition
			wantStep := (&town.Good{Condition: wantCondition}).ConditionStep()
			g := &game{play: &town.Session{Party: party, LeftTown: true}, quest: &questBattle{state: s}, arenaFighter: -1}
			if tc.arena {
				g.arenaFighter, g.arenaBets = 1, []int{17, 0, -9}
			}
			for n := 0; n < 4 && g.quest.state != nil; n++ {
				g.battleStep()
			}
			if g.quest.state != nil {
				t.Fatal("結算未返回世界")
			}
			for i, m := range g.play.Party {
				wantGold, wantPoints, wantRank := 100, 50, 2
				if tc.arena {
					bet := []int{17, 0, -9}[i]
					if tc.verdict == battle.VerdictLost {
						bet = -bet
					}
					wantGold += bet
					if i == 1 && tc.verdict == battle.VerdictWon {
						wantRank++
					}
				} else if tc.verdict == battle.VerdictWon {
					wantGold += 11
					wantPoints += 7
				}
				if m.Gold != wantGold || m.Sheet.AdventurePoints != wantPoints || m.Sheet.Rank != wantRank {
					t.Fatalf("%s 獎勵／升階／賭金錯誤：%d/%d/%d", m.Name, m.Gold, m.Sheet.AdventurePoints, m.Sheet.Rank)
				}
				if i == 1 && (m.Wounds != 14 || m.Worn[0] != nil || m.Worn[1] == nil || m.Worn[1].Condition != wantCondition) {
					t.Fatal("傷勢或護甲未回寫")
				}
				// 物品頁的狀況名稱只看 Condition 分檔（`goodDetail`）；磨光的手上那件整格空掉。
				if i == 1 && (m.InHand != nil || m.Worn[1].ConditionStep() != wantStep) {
					t.Fatalf("破損或狀況檔次未回寫：手上 %v、檔次 %d／%d", m.InHand, m.Worn[1].ConditionStep(), wantStep)
				}
				if i != 1 && m.Wounds != 5 {
					t.Fatal("未受傷隊員或觀眾傷勢被改寫")
				}
				for j := 3; j < 9; j++ {
					if m.Worn[j] == nil || m.Worn[j].Condition != 200 {
						t.Fatal("未受損配件遺失")
					}
				}
			}
			w := &world.Screen{Travel: &world.Travel{}, Clock: world.NewClock()}
			out, err := buildCompleteSave(g.play, w, r, false)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(t.TempDir(), "battle-v2.json")
			if err = save.Write(path, out); err != nil {
				t.Fatal(err)
			}
			loaded, err := save.Read(path)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(loaded.Characters.Party, slots) || loaded.Characters.Speaker != 1 {
				t.Fatal("隊伍順序、槽位或發言者改變")
			}
			for i, got := range savedParty(loaded) {
				want, err := savedCharacter(g.play.Party[i])
				if err != nil {
					t.Fatal(err)
				}
				actual, err := savedCharacter(got)
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(actual, want) {
					t.Fatalf("%s 戰後完整資料重載不符", got.Name)
				}
			}
		})
	}
}
