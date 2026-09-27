package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/oracle"
	"github.com/wicanr2/kol-remake-cht/internal/roster"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

// 規劃段的同狀態比較：同一場遭遇裡，remake 為 Theodoric 排出的主指令列與
// 三個決策值要與原版一致。
//
// # 原版那一側量到什麼
//
// `tools/oracle/book_startup_test.go` 的 `KOL_BOOK_PLAN` 模式，序列
// `.,Return,Return,.,.,Return,.,.,.,Return`，收據在
// `docs/verification/normal-path-planning-entry-20260911.md`：
//
//	規劃入口   圖示列 5 格，白框在第 0 格，行動者是槽 1（Theodoric）
//	`.`        白框到第 1 格
//	Return     圖示列 2 格 → 選目標格（SPEC-010 mode 1）
//	Return     圖示列 6 格 → 攻擊式那一列
//	`.` ×2     白框到第 2 格
//	Return     圖示列 8 格 → 防禦方式那一列
//	`.` ×3     白框到第 3 格
//	Return     寫回槽 1：+0x23=1、+0x24=2、+0x25=0x29
//
// 8 格那一列正是 SPEC-010 §4 的防禦方式（`0x0F、0x2A、0x30…0x34、0x36`）；
// 指令碼 `0x29` 是徒手，於是 6 格那一列是徒手攻擊式（`0x0F、0x2A、0x37…0x3A`）。
//
// # 為什麼比的是「格數與圖號」而不是按鍵
//
// 原版收的是掃描碼，remake 收的是 token，兩邊的輸入層不可能逐鍵對齊。可比的是
// **同一個位置代表哪一格**：白框停在第 n 格時圖號是多少，以及那個圖號換算出來的
// 三個值。按鍵只是把白框搬到第 n 格的手段。
func TestOriginalNormalPlanning(t *testing.T) {
	s := originalPlanningState(t)
	const actor = 1

	// 一、主指令列：格數與第 1 格的圖號。
	if got := len(s.Plan.Bar); got != 5 {
		t.Errorf("主指令列 %d 格，原版是 5 格：%#x", got, s.Plan.Bar)
	}
	if len(s.Plan.Bar) > 1 {
		if got := s.Plan.Bar[1]; got != battle.CmdUnarmed {
			t.Errorf("主指令列第 1 格是 %#x，原版是 %#x（徒手）", got, battle.CmdUnarmed)
		}
	}

	// 二、走原版那一條：第 1 格 → 選目標格 → 攻擊式第 2 格 → 防禦式第 3 格。
	step := func(what string, slot, wantBar int) {
		t.Helper()
		if wantBar > 0 && len(s.Plan.Bar) != wantBar {
			t.Fatalf("%s：圖示列 %d 格，原版是 %d 格：%#x", what, len(s.Plan.Bar), wantBar, s.Plan.Bar)
		}
		s.Plan.Selected = slot
		s.Choose()
	}
	step("主指令列", 1, 5)
	if s.Plan.Prompt != battle.PromptTarget {
		t.Fatalf("選徒手之後應該進選目標格，得到 %d", s.Plan.Prompt)
	}
	s.Choose() // 確定面前那一格，原版那一步的圖示列是 2 格
	step("攻擊式", 2, 6)
	step("防禦式", 3, 8)

	// 三、三個決策值。
	a := s.Actors[actor]
	for _, c := range []struct {
		name      string
		got, want int
	}{
		{"指令碼 +0x25", a.Command, 0x29},
		{"攻擊選項 +0x23", a.Style, 1},
		{"防禦選項 +0x24", a.Guard, 2},
	} {
		if c.got != c.want {
			t.Errorf("%s：remake %#x，原版 %#x", c.name, c.got, c.want)
		}
	}
}

// originalPlanningState 用 dosgolem 正常入口的角色資料與地形快照建同一場遭遇，
// 並把規劃推到原版那一場第一個輪到的槽 1（Theodoric）。缺輸入就跳過。
func originalPlanningState(t *testing.T) *battle.State {
	t.Helper()
	statePath := os.Getenv("KOL_PARTY_STATE_INPUT")
	terrainPath := os.Getenv("KOL_TERRAIN_INPUT")
	if statePath == "" || terrainPath == "" {
		t.Skip("需要 dosgolem 正常入口的角色資料與地形快照")
	}
	var input struct {
		Source       string
		Seed         int
		Party        []roster.Slot
		Monster      battle.Monster
		MonsterLevel int `json:"monster_level"`
		Actors       []struct {
			Slot int
			X, Y int
		}
	}
	body, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(body, &input); err != nil {
		t.Fatal(err)
	}
	field, snapshot, err := oracle.LoadTerrainSnapshot(terrainPath)
	if err != nil {
		t.Fatal(err)
	}
	if input.Source != snapshot.Source || input.Seed != snapshot.AlignmentSeed {
		t.Fatalf("角色資料是 %q／seed %d，地形快照是 %q／seed %d——不是同一場",
			input.Source, input.Seed, snapshot.Source, snapshot.AlignmentSeed)
	}

	site := &battle.Site{Index: -1}
	var party []battle.Member
	for i, slot := range input.Party {
		m := town.Member{Name: slot.Name, Wounds: slot.Wounds, Sheet: townSheet(slot.Sheet),
			InHand: townGood(slot.InHand), Worn: townWorn(slot.Equipment)}
		party = append(party, battleMember(m))
		site.PartyX = append(site.PartyX, input.Actors[i].X)
		site.PartyY = append(site.PartyY, input.Actors[i].Y)
	}
	site.Count, site.Level, site.MonsterName = 1, input.MonsterLevel, input.Monster.Name
	site.Monster = input.Monster.Index
	site.MonsterX, site.MonsterY = []int{input.Actors[6].X}, []int{input.Actors[6].Y}
	s := battle.NewQuest(site, party, input.Monster, field)

	// 原版那一場規劃是從 Theodoric（槽 1）開始的。
	const actor = 1
	if s.Party[actor].Name != "Theodoric" {
		t.Fatalf("槽 %d 不是 Theodoric，是 %q", actor, s.Party[actor].Name)
	}
	s.Stage = battle.StageField
	s.Round.Order = []int{actor}
	s.Round.Turn = 0
	s.Cursor = actor
	s.BeginTurn()
	return s
}
