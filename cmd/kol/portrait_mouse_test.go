package main

import (
	"reflect"
	"strings"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/book"
	"github.com/wicanr2/kol-remake-cht/internal/input"
)

// 點隊伍肖像等於按數字（RE-167）：x 從 16 起每 48 像素一位、y < 50、那一位要在隊伍裡。
func TestPortraitClickZones(t *testing.T) {
	for _, cjk := range []bool{false, true} {
		scale := 1
		if cjk {
			scale = book.CJKScale
		}
		targets := portraitTargets(3, cjk)
		if len(targets) != 3 {
			t.Fatalf("三人隊伍該有 3 個區域，得到 %d", len(targets))
		}
		for _, c := range []struct {
			x, y int
			want []string
		}{
			{16, 0, []string{"1"}}, {63, 49, []string{"1"}},
			{64, 0, []string{"2"}}, {111, 25, []string{"2"}},
			{112, 10, []string{"3"}}, {159, 10, []string{"3"}},
			{160, 10, nil}, // 第四位不在隊伍裡
			{15, 10, nil},  // 左緣之外
			{30, 50, nil},  // y 必須小於 50
		} {
			got := input.Hit(targets, c.x*scale, c.y*scale, "")
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("中文=%v (%d,%d)：%v，應為 %v", cjk, c.x, c.y, got, c.want)
			}
		}
	}
	if got := portraitTargets(9, false); len(got) != 6 {
		t.Fatalf("隊伍最多六人，得到 %d 個區域", len(got))
	}
}

// 選目標格那一層：圖示列（YES、U-TURN）與戰場格各自可點，場外的點擊仍被吃掉。
func TestTargetLayerIconBarClicks(t *testing.T) {
	for _, cjk := range []bool{false, true} {
		scale := 1
		if cjk {
			scale = book.CJKScale
		}
		s := &battle.State{Stage: battle.StageField}
		s.Plan.Prompt = battle.PromptTarget
		s.Plan.Bar = []int{battle.CmdYes, battle.CmdCancel}
		s.Plan.Selected = 0
		targets := battleTargets(s, cjk)
		hit := func(x, y int) []string { return input.Hit(targets, x*scale, y*scale, "RET") }
		// 圖示列在畫面下緣：第 0 格確認，第 1 格（U-TURN）把白框移過去再確認。
		if got := hit(40, 175); !reflect.DeepEqual(got, []string{"RET"}) {
			t.Errorf("中文=%v 第 0 格：%v", cjk, got)
		}
		if got := hit(72, 175); !reflect.DeepEqual(got, []string{">", "RET"}) {
			t.Errorf("中文=%v 第 1 格：%v", cjk, got)
		}
		// 戰場中間是格，token 是 CELL:列,欄。
		if got := hit(200, 100); len(got) != 1 || !strings.HasPrefix(got[0], "CELL:") {
			t.Errorf("中文=%v 戰場格：%v", cjk, got)
		}
		// 戰場外、圖示列外：吃掉，不落到 fallback 的 RET。
		if got := hit(5, 5); len(got) != 0 {
			t.Errorf("中文=%v 場外：%v", cjk, got)
		}
	}
}
