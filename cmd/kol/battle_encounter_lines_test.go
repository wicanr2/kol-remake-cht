package main

import (
	"reflect"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

func TestQuestEncounterDisplayKeepsOriginalTrace(t *testing.T) {
	bank, err := town.LoadUIText("../../assets/original-derived-data/ui-text.json")
	if err != nil {
		t.Fatal(err)
	}
	trans, err := loadTranslations("../../docs/i18n/town-zh-Hant.json", "../../docs/i18n/ui-zh-Hant.json")
	if err != nil {
		t.Fatal(err)
	}
	original := nonEmptyLines(bank.Lines(town.BankUI, battle.QuestCreaturesRecord))
	before := append([]string(nil), original...)
	g := &game{quest: &questBattle{bank: bank, trans: trans, lines: original}}
	want := []string{"你遇上了任務中", "那些可怕的", "生物！"}
	for i := 0; i < 2; i++ {
		if got := g.battleEncounterLines(); !reflect.DeepEqual(got, want) {
			t.Fatalf("繁中遭遇顯示：%v，預期 %v", got, want)
		}
		if !reflect.DeepEqual(g.quest.lines, before) {
			t.Fatalf("繪圖改寫原始軌跡文字：%v", g.quest.lines)
		}
	}
	g.quest.trans = nil
	if got := g.battleEncounterLines(); !reflect.DeepEqual(got, before) {
		t.Fatalf("缺譯回退改變原文：%v", got)
	}
	// 伏擊句帶動態數值與怪物名，不能換成任務地點的固定訊息。
	dynamic := ambushLines(bank, trans, &battle.Ambush{Count: 2, Level: 1}, "大哥布林", false)
	g.quest.trans, g.quest.lines = trans, dynamic
	if got := g.battleEncounterLines(); !reflect.DeepEqual(got, dynamic) {
		t.Fatalf("動態伏擊被改成任務提示：%v", got)
	}
}
