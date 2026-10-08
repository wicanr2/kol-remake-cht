package main

import (
	"github.com/wicanr2/kol-remake-cht/internal/input"
	"reflect"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/book"
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
	// 選目標格那一層是座標指格（SPEC-010 §3）：9×9 格各送 CELL:列,欄，戰場外的點擊吃掉。
	s.Plan.Prompt = battle.PromptTarget
	for _, tc := range []struct {
		cjk  bool
		x, y int
		want []string
	}{
		{false, 153, 9, []string{"CELL:0,0"}},
		{false, 152, 9, []string{}}, // X 0x130 本身不算（嚴格大於）
		{false, 153, 8, []string{}}, // Y 8 本身不算
		{false, 295, 151, []string{"CELL:8,8"}},
		{false, 296, 100, []string{}}, // X 0x250 不算
		{false, 10, 190, []string{}},  // 圖示列：這一層只收鍵盤
		{true, 153*3 + 1, 9*3 + 1, []string{"CELL:0,0"}},
		{true, (152 + 16*4 + 8) * 3, (8 + 16*4 + 8) * 3, []string{"CELL:4,4"}},
	} {
		got := input.Hit(battleTargets(s, tc.cjk), tc.x, tc.y, "RET")
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("cjk=%v (%d,%d) 得到 %q，要 %q", tc.cjk, tc.x, tc.y, got, tc.want)
		}
	}
}

// 戰利品畫面：兩張清單各自可點，圖示列另外兩格。
func TestBattleSpoilsClick(t *testing.T) {
	for _, cjk := range []bool{false, true} {
		s := &battle.State{Stage: battle.StageSpoils}
		s.Ground[0] = battle.GroundItem{Name: "DAGGER"}
		s.Ground[3] = battle.GroundItem{Name: "SHIELD"}
		if !s.BeginSpoils() {
			t.Fatal("兩件東西在地上，戰利品畫面該開得起來")
		}
		lines := len(s.SpoilLines())
		if lines != 2 {
			t.Fatalf("中文=%v 清單 %d 列，地上有兩件", cjk, lines)
		}
		targets := battleTargets(s, cjk)
		if len(targets) != lines+len(battle.SpoilIcons) {
			t.Fatalf("中文=%v 可點區域 %d 個，要 %d 列加 %d 格圖示",
				cjk, len(targets), lines, len(battle.SpoilIcons))
		}
		for i := 1; i < lines; i++ {
			if targets[i-1].Rect.Overlaps(targets[i].Rect) {
				t.Fatalf("中文=%v 第 %d 列與前一列重疊", cjk, i)
			}
		}
		// 游標在第 0 列、白框在第 0 格，點第 1 列就是往下一步再確認。
		if want := []string{"D", "RET"}; !reflect.DeepEqual(targets[1].Tokens, want) {
			t.Fatalf("中文=%v 第 1 列：%v，要 %v", cjk, targets[1].Tokens, want)
		}
		if cjk {
			wantX := battle.SpoilList.Col * town.CellWidth * book.CJKScale
			wantY := battle.SpoilList.Row * book.CJKScale
			if targets[0].Rect.Min.X != wantX || targets[0].Rect.Min.Y != wantY {
				t.Fatalf("中文戰利品第 0 列起點 %v，要 (%d,%d)",
					targets[0].Rect, wantX, wantY)
			}
			if targets[0].Rect.Max.X > 960 || targets[0].Rect.Max.Y > 600 {
				t.Fatalf("中文戰利品命中區超出畫布：%v", targets[0].Rect)
			}
		}
		// 白框停在取消那一格時，點某一列仍要把白框搬回第一格才確認。
		s.Spoil.Icon = battle.SpoilIconUTurn
		moved := battleTargets(s, cjk)
		if want := []string{"<", "RET"}; !reflect.DeepEqual(moved[0].Tokens, want) {
			t.Fatalf("中文=%v 白框在取消格時第 0 列：%v，要 %v", cjk, moved[0].Tokens, want)
		}
	}
}

// 法術與正負效果清單必須和鍵盤共用上下移動 token；中文模式只改變畫布座標。
func TestBattleSpellListClick(t *testing.T) {
	for _, tc := range []struct {
		name   string
		prompt battle.Prompt
		rows   []string
		pick   int
		row    int
	}{
		{name: "法術", prompt: battle.PromptSpell, rows: []string{"A", "B", "C"}, pick: 1, row: 17},
		{name: "正負效果", prompt: battle.PromptAffect, rows: []string{"Positive", "Negative"}, pick: 0, row: 25},
	} {
		for _, cjk := range []bool{false, true} {
			s := &battle.State{Stage: battle.StageField}
			s.Plan.Prompt = tc.prompt
			s.Plan.List = tc.rows
			s.Plan.ListPick = tc.pick
			targets := battleTargets(s, cjk)
			if len(targets) != len(tc.rows) {
				t.Fatalf("%s 中文=%v 可點列數 %d，要 %d", tc.name, cjk, len(targets), len(tc.rows))
			}
			for i := 1; i < len(targets); i++ {
				if targets[i-1].Rect.Overlaps(targets[i].Rect) {
					t.Fatalf("%s 中文=%v 第 %d 列與前一列重疊", tc.name, cjk, i)
				}
			}
			if !reflect.DeepEqual(targets[0].Tokens, []string{"U", "RET"}) && tc.pick == 1 {
				t.Fatalf("%s 中文=%v 第 0 列 token：%v", tc.name, cjk, targets[0].Tokens)
			}
			if tc.pick == 0 && !reflect.DeepEqual(targets[1].Tokens, []string{"D", "RET"}) {
				t.Fatalf("%s 中文=%v 第 1 列 token：%v", tc.name, cjk, targets[1].Tokens)
			}
			if cjk {
				wantX := battle.SpellListRows.Col * town.CellWidth * book.CJKScale
				wantY := tc.row * book.CJKScale
				if targets[0].Rect.Min.X != wantX || targets[0].Rect.Min.Y != wantY {
					t.Fatalf("%s 中文第 0 列起點 %v，要 (%d,%d)", tc.name, targets[0].Rect, wantX, wantY)
				}
				if targets[len(targets)-1].Rect.Max.X > 960 || targets[len(targets)-1].Rect.Max.Y > 600 {
					t.Fatalf("%s 中文命中區超出畫布：%v", tc.name, targets[len(targets)-1].Rect)
				}
			}
		}
	}
}
