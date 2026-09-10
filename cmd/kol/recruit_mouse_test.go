package main

import (
	"github.com/wicanr2/kol-remake-cht/internal/town"
	"reflect"
	"testing"
)

func TestRecruitMouseSelectsClickedRowAndConfirms(t *testing.T) {
	for _, cjk := range []bool{false, true} {
		in := &town.Inn{Selected: 1, List: &town.InnList{Rows: []string{"A", "B", "C"}, Selected: 1, Total: 11}}
		targets := recruitTargets(in, cjk)
		if len(targets) != 3 {
			t.Fatal("清單可點列數不符")
		}
		for i, want := range [][]string{{"U", "<", "RET"}, {"<", "RET"}, {"D", "<", "RET"}} {
			if !reflect.DeepEqual(targets[i].Tokens, want) {
				t.Fatalf("中文=%v 第%d列：%v", cjk, i, targets[i].Tokens)
			}
			if i > 0 && targets[i-1].Rect.Overlaps(targets[i].Rect) {
				t.Fatal("列範圍重疊")
			}
		}
		if cjk && targets[0].Rect.Min.Y != town.CJKInn.DialogTextY {
			t.Fatal("可點列未對齊繁中畫面")
		}
	}
}
