package main

import (
	"encoding/hex"
	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/roster"
	"github.com/wicanr2/kol-remake-cht/internal/town"
	"testing"
)

type armourOneRand struct{}

func (armourOneRand) Roll(int) int { return 1 }

func TestBattleSpaceTogglesWithoutConfirming(t *testing.T) {
	s := &battle.State{Stage: battle.StageField, Field: battle.NewField(),
		Actors: []battle.Actor{{Side: battle.SideParty, Member: 0, Threshold: 100, Head: 15, Chest: 15}},
		Party:  []battle.Member{{Name: "PLAYER"}}, Round: battle.Round{Order: []int{0}},
		Plan: battle.Planning{Prompt: battle.PromptCommand, BodyActor: 1, Bar: []int{battle.CmdRest}}}
	g := &game{}
	if err := g.applyBattleToken(s, "SPACE", func(*battle.State) { t.Fatal("切換不應推進執行段") }); err != nil {
		t.Fatal(err)
	}
	if !s.Plan.EquipmentPanel || s.Plan.Prompt != battle.PromptCommand || s.Actors[0].Command != 0 {
		t.Fatal("空白鍵錯誤確認了指令")
	}
	if err := g.applyBattleToken(s, "RET", func(*battle.State) {}); err != nil {
		t.Fatal(err)
	}
	if s.Actors[0].Command != battle.CmdRest {
		t.Fatal("Return 應仍能確認指令")
	}
}

func TestBattleAdapterCarriesPortraitAndShieldName(t *testing.T) {
	m := town.Member{Name: "PLAYER", Portrait: 23, Sheet: &town.Sheet{}}
	m.Worn[battle.ShieldSlot] = &town.Good{Name: "Shield"}
	b := battleMember(m)
	if b.Portrait != 23 || b.Worn[4] != "Shield" {
		t.Fatalf("肖像／盾牌未傳入：%+v", b)
	}
}

func TestArmourBattleExitWritesRecordAndCarried(t *testing.T) {
	for _, verdict := range []battle.Verdict{battle.VerdictWon, battle.VerdictFled} {
		m := town.Member{Name: "ARMOUR", Sheet: &town.Sheet{}}
		for pos, condition := range []byte{32, 200} {
			raw := make([]byte, 16)
			raw[0], raw[10], raw[11], raw[15] = 1, 4, condition, 0x20
			m.Worn[pos] = &town.Good{Name: "ARMOUR", Record: hex.EncodeToString(raw), Condition: int(condition), Encumbrance: 4, Percent: 40}
		}
		bm := battleMember(m)
		s := battle.NewQuest(&battle.Site{Index: -1, PartyX: []int{0}, PartyY: []int{0}}, []battle.Member{bm}, battle.Monster{}, nil)
		s.Verdict = verdict
		s.Actors[0].AbsorbArmour(5, 0, armourOneRand{})
		s.Actors[0].AbsorbArmour(3, 1, armourOneRand{})
		g := &game{play: &town.Session{Party: []town.Member{m}}}
		g.writeBattleParty(s)
		got := g.play.Party[0]
		if got.Worn[0] != nil || got.Worn[1] == nil || got.Worn[1].Condition != 136 || got.Worn[1].Percent != 40 {
			t.Fatalf("結局=%v：破損或耐久未回寫：%+v", verdict, got.Worn)
		}
		if len(got.Carried) != 17 || got.Carried[7][0] != 0 || got.Carried[8][11] != 136 {
			t.Fatalf("結局=%v：Carried 未同步", verdict)
		}
		if m.Worn[0].Condition != 32 || m.Worn[1].Condition != 200 {
			t.Fatal("戰鬥修改污染輸入隊員資料")
		}
	}
}

func TestBattleAdapterPreservesWardAndBelt(t *testing.T) {
	for _, ward := range []int{0x10, 0x20, 0x30, 0x40} {
		sheet := townSheet(&roster.Sheet{Stats: make([]int, town.StatCount), Ward: ward})
		m := town.Member{Name: "CHECK", Sheet: sheet}
		m.Worn[9] = &town.Good{Name: "BELT SWORD", Class: 0x21, Encumbrance: 17, Record: "00000000000000000000000021"}
		got := battleMember(m)
		if got.Ward != ward || !got.HasSecondItem || got.BeltName != "BELT SWORD" || got.BeltWeapon.Kind != 0x21 || got.BeltWeapon.Weight != 17 {
			t.Fatalf("戰鬥轉接遺失角色資料：%+v", got)
		}
	}
}

func TestBattleExitPreservesInventoryAndWounds(t *testing.T) {
	weapon := town.Good{Name: "SWORD", Class: 0x21, Price: 78, Condition: 99, Record: "0102030405060708090a0b0c0d0e0f10"}
	loot := town.Good{Name: "TOKEN", Price: 13, Record: "1102030405060708090a0b0c0d0e0f10"}
	for _, verdict := range []battle.Verdict{battle.VerdictWon, battle.VerdictFled} {
		member := town.Member{Name: "CHECK", Sheet: &town.Sheet{}, InHand: &weapon}
		bm := battleMember(member)
		s := &battle.State{Stage: battle.StageDone, Verdict: verdict, Party: []battle.Member{bm}, Actors: []battle.Actor{{Member: 0, Damage: 23, Weapon: bm.Weapon}}}
		s.Party[0].Pack[0], s.Party[0].PackItems[0] = loot.Name, battleItem(&loot)
		s.Exec.Text = []string{"已顯示結算"}
		g := &game{play: &town.Session{Party: []town.Member{member}}, quest: &questBattle{state: s}, arenaFighter: -1}
		g.battleStep()
		m := g.play.Party[0]
		if g.quest.state != nil || m.Wounds != 11 || len(m.Backpack) != 1 || m.Backpack[0].Record != loot.Record || m.InHand.Price != weapon.Price || m.InHand.Condition != weapon.Condition || len(m.Carried) != 17 || m.Carried[1][0] != 0x11 {
			t.Fatalf("戰後未保留完整結果：%+v", m)
		}
	}
}
