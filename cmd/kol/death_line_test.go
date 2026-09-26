package main

import (
	"strings"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

// 攻擊致死那一句的三個分支（`sub_277F4:0x2796F`，#52）用真實的 KOLDATA 文字庫與
// 繁中譯表展開：隊員 collapse、SKELETON（怪物記錄 +0x10 bit 7）shatter、其餘怪物 die。
func TestAttackDeathLinesExpandFromBank(t *testing.T) {
	bank, err := town.LoadUIText("../../assets/original-derived-data/ui-text.json")
	if err != nil {
		t.Fatal(err)
	}
	trans, err := loadTranslations("../../docs/i18n/town-zh-Hant.json", "../../docs/i18n/ui-zh-Hant.json")
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name              string
		target            battle.Speaker
		monster, shatters bool
		en, zh            string
	}{
		{"隊員", battle.Speaker{Name: "Moraine", Female: true}, false, false,
			"collapse from blood loss!", "倒了下去"},
		{"SKELETON", battle.Speaker{Name: "SKELETON", Monster: true}, true, true,
			"shatter into pieces!", "碎成了片片"},
		{"ZOMBIE", battle.Speaker{Name: "ZOMBIE", Monster: true}, true, false,
			"die from blood loss!", "死去"},
	}
	out := battle.Outcome{Damage: 5, Location: 2, Down: true}
	for _, c := range cases {
		records := battle.DamageLine(out, 0, c.monster, c.shatters)
		for _, lang := range []struct {
			tr   *town.Translations
			want string
		}{{nil, c.en}, {trans, c.zh}} {
			var text strings.Builder
			for _, r := range records {
				text.WriteString(town.UIRaw(bank, lang.tr, town.BankUI, r))
			}
			tr := lang.tr
			ctx := battle.MessageContext{Target: c.target, Lookup: func(r int) string {
				return town.UIRaw(bank, tr, town.BankUI, r)
			}}
			line := ctx.Substitute(text.String())
			if !strings.Contains(line, lang.want) {
				t.Errorf("%s：展開成 %q，應含 %q", c.name, line, lang.want)
			}
			t.Logf("%s：%s", c.name, line)
		}
	}
}
