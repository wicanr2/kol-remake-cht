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

// 驗證已結算狀態接正式共同出口、完整v2檔案往返；正常GUI另產收據。
func TestBattleLootSaveKeepsPocketHoles(t *testing.T) {
	m := town.Member{ID: "PLAYER", Name: "PLAYER", Gold: 1009, Sheet: &town.Sheet{}}
	for k := 0; k < 6; k++ {
		raw := make([]byte, 16)
		raw[0] = byte(k + 1)
		m.Backpack = append(m.Backpack, town.Good{Name: "ITEM", Record: hex.EncodeToString(raw)})
	}
	m.SyncCarried()
	s := battle.NewQuest(&battle.Site{Index: -1, PartyX: []int{0}, PartyY: []int{0}}, []battle.Member{battleMember(m)}, battle.Monster{LootChance: 20}, battle.NewField())
	s.Stage, s.Verdict = battle.StageDone, battle.VerdictLost
	s.Actors[0].Head = 0
	s.Actors[0].Damage = 48
	s.Exec.Text = []string{"已顯示敗北"}
	// 元件測試顯式控制骰值：只清背包第三格與錢包，其餘槽保留。
	s.Rand = &lootHoleRand{}
	g := &game{play: &town.Session{Party: []town.Member{m}, LeftTown: true}, quest: &questBattle{state: s}, arenaFighter: -1}
	g.battleStep()
	got := g.play.Party[0]
	if g.quest.state != nil || got.Wounds != 24 || got.Gold != 0 || len(got.Backpack) != 6 || got.Backpack[2].Name != "" || got.Backpack[3].Record != m.Backpack[3].Record || got.Carried[3][0] != 0 || got.Carried[4][0] != 4 {
		t.Fatal("搜刮出口未保留固定背包槽")
	}
	r := &roster.Roster{Slots: make([]roster.Slot, 16)}
	r.Slots[0] = roster.Slot{Slot: 0, ID: m.ID, Name: m.Name, Occupied: true}
	w := &world.Screen{Travel: &world.Travel{}, Clock: world.NewClock()}
	complete, err := buildCompleteSave(g.play, w, r, false)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "loot-v2.json")
	if err := save.Write(path, complete); err != nil {
		t.Fatal(err)
	}
	loaded, err := save.Read(path)
	if err != nil {
		t.Fatal(err)
	}
	party := savedParty(loaded)
	if len(party) != 1 || party[0].Gold != 0 || party[0].Wounds != 24 || !reflect.DeepEqual(party[0].Backpack, got.Backpack) || !reflect.DeepEqual(party[0].Carried, got.Carried) {
		t.Fatal("搜刮後資料重載不符")
	}
}

type lootHoleRand struct{ calls int }

func (r *lootHoleRand) Roll(int) int {
	r.calls++
	if r.calls == 3 || r.calls == 16 {
		return 1
	}
	return 100
}
