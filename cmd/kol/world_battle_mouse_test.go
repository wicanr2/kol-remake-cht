package main

import (
	"reflect"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/town"
	"github.com/wicanr2/kol-remake-cht/internal/world"
)

// 野外的圖示列：點第 n 格等於把白框搬過去再確認，與城鎮同一套 token。
func TestWorldIconsClickMovesFrameAndConfirms(t *testing.T) {
	for _, cjk := range []bool{false, true} {
		s := &world.Session{Screen: &world.Screen{Icons: []int{1, 2, 3, 4}, Selected: 2}}
		targets := worldTargets(s, cjk)
		if len(targets) != 4 {
			t.Fatalf("中文=%v 可點格數 %d，圖示有 4 格", cjk, len(targets))
		}
		for i, want := range [][]string{
			{"<", "<", "RET"}, {"<", "RET"}, {"RET"}, {">", "RET"},
		} {
			if !reflect.DeepEqual(targets[i].Tokens, want) {
				t.Fatalf("中文=%v 第 %d 格：%v，要 %v", cjk, i, targets[i].Tokens, want)
			}
			if i > 0 && targets[i-1].Rect.Overlaps(targets[i].Rect) {
				t.Fatalf("中文=%v 第 %d 格與前一格重疊", cjk, i)
			}
		}
		if !cjk && targets[0].Rect.Min.X != town.IconBar.X {
			t.Fatalf("英文版第 0 格起點 %d，圖示列是 %d", targets[0].Rect.Min.X, town.IconBar.X)
		}
	}
}

// 露營與問句那幾頁蓋在圖示列上，那時候不該還能點著搬白框。
func TestWorldIconsSilentWhileCampingOrAsking(t *testing.T) {
	base := func() *world.Session {
		return &world.Session{Screen: &world.Screen{Icons: []int{1, 2}, Selected: 0}}
	}
	if got := worldTargets(base(), false); len(got) != 2 {
		t.Fatalf("正常狀態該有 2 格，得到 %d——這個測試分不出擋掉的效果", len(got))
	}
	camping := base()
	camping.Camping = true
	if got := worldTargets(camping, false); got != nil {
		t.Fatalf("露營時不該有可點區域，得到 %d 個", len(got))
	}
	asking := base()
	asking.Prompt = &world.Prompt{}
	if got := worldTargets(asking, false); got != nil {
		t.Fatalf("問句時不該有可點區域，得到 %d 個", len(got))
	}
}

// 戰鬥規劃段的指令列：點第 n 格等於把白框搬過去再確認。
func TestBattleCommandBarClick(t *testing.T) {
	s := &battle.State{Stage: battle.StageField}
	s.Plan.Prompt = battle.PromptCommand
	s.Plan.Bar = []int{battle.CmdYes, battle.CmdUnarmed, battle.CmdMove, battle.CmdRest}
	s.Plan.Selected = 1
	targets := battleTargets(s, false)
	if len(targets) != 4 {
		t.Fatalf("可點格數 %d，指令列有 4 格", len(targets))
	}
	for i, want := range [][]string{{"<", "RET"}, {"RET"}, {">", "RET"}, {">", ">", "RET"}} {
		if !reflect.DeepEqual(targets[i].Tokens, want) {
			t.Fatalf("第 %d 格：%v，要 %v", i, targets[i].Tokens, want)
		}
	}
	// 選目標格那一層收的是座標指格，不是搬白框——不列進來，免得點下去變成另一回事。
	s.Plan.Prompt = battle.PromptTarget
	if got := battleTargets(s, false); got != nil {
		t.Fatalf("選目標格那一層不該列可點區域，得到 %d 個", len(got))
	}
}

// 戰利品畫面：兩張清單各自可點，圖示列另外兩格。
func TestBattleSpoilsClick(t *testing.T) {
	s := &battle.State{Stage: battle.StageSpoils}
	s.Ground[0] = battle.GroundItem{Name: "DAGGER"}
	s.Ground[3] = battle.GroundItem{Name: "SHIELD"}
	if !s.BeginSpoils() {
		t.Fatal("兩件東西在地上，戰利品畫面該開得起來")
	}
	lines := len(s.SpoilLines())
	if lines != 2 {
		t.Fatalf("清單 %d 列，地上有兩件", lines)
	}
	targets := battleTargets(s, false)
	if len(targets) != lines+len(battle.SpoilIcons) {
		t.Fatalf("可點區域 %d 個，要 %d 列加 %d 格圖示",
			len(targets), lines, len(battle.SpoilIcons))
	}
	// 游標在第 0 列、白框在第 0 格，點第 1 列就是往下一步再確認。
	if want := []string{"D", "RET"}; !reflect.DeepEqual(targets[1].Tokens, want) {
		t.Fatalf("第 1 列：%v，要 %v", targets[1].Tokens, want)
	}
	// 白框停在取消那一格時，點某一列仍要把白框搬回第一格才確認。
	s.Spoil.Icon = battle.SpoilIconUTurn
	moved := battleTargets(s, false)
	if want := []string{"<", "RET"}; !reflect.DeepEqual(moved[0].Tokens, want) {
		t.Fatalf("白框在取消格時第 0 列：%v，要 %v", moved[0].Tokens, want)
	}
}
