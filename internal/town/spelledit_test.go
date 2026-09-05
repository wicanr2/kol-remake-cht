package town

import (
	"path/filepath"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/magic"
)

func spellDictionary(t *testing.T) *magic.Dictionary {
	t.Helper()
	d, err := magic.Load(filepath.Join("..", "..", "assets",
		"original-derived-data", "magic.json"))
	if err != nil {
		t.Skip("沒有 magic.json，跳過")
	}
	return d
}

// 原版名冊唯一那一筆法術（第 9 槽的 Warder，記錄
// `bb0654b959993520 74 00 00 00 05 00 8a 53`）。
func vartwelyrmi() Spell {
	return Spell{
		Name: "Vartwelyrmi", Guards: []int{0x74, 0, 0, 0},
		Effect: 5, TurnCode: 0, PowerCode: 0x8A,
	}
}

// 十三個欄位讀出來的意思，逐欄對原版那一筆。
func TestSpellFieldsOfOriginalRecord(t *testing.T) {
	d := spellDictionary(t)
	sp := vartwelyrmi()
	want := []string{
		SpellFieldName:       "Vartwelyrmi",
		SpellFieldClassFirst: "Legendary", 2: "None", 3: "None", 4: "None",
		SpellFieldSubFirst: "Blom", 6: "None", 7: "None", 8: "None",
		SpellFieldStat: "Fatigue",
		// 屬性落在 3…6，持續回合那一欄原版連畫都不畫。
		SpellFieldTargetFlag: "", SpellFieldTurn: "", SpellFieldPower: "1-12",
	}
	for field, text := range want {
		if got := SpellFieldText(d, sp, field); got != text {
			t.Fatalf("欄位 %d 是 %q，應該是 %q", field, got, text)
		}
	}
}

// 造價：四格護符 ＋ 屬性 ＋ 目標旗標 ＋ 強度那一欄；持續那一欄不算。
func TestSpellCostOfOriginalRecord(t *testing.T) {
	d := spellDictionary(t)
	ap, gc := SpellCost(d, vartwelyrmi())
	// 400＋90＋200＋140 與 200＋90＋200＋140。
	if ap != 830 || gc != 630 {
		t.Fatalf("造價是 %d／%d，應該是 830／630", ap, gc)
	}
}

// 屬性換成一個會持續的（1 ＝ Strength）之後，持續那一欄就要算錢也要顯示。
func TestTurnsColumnAppearsWithLastingStat(t *testing.T) {
	d := spellDictionary(t)
	sp := vartwelyrmi()
	base, _ := SpellCost(d, sp)
	SetSpellField(d, &sp, SpellFieldStat, 1)
	if got := SpellFieldText(d, sp, SpellFieldTurn); got != "1-3" {
		t.Fatalf("持續那一欄是 %q，應該是 1-3", got)
	}
	ap, _ := SpellCost(d, sp)
	// 屬性從 Fatigue（90）換成 Strength（50）少 40，持續那一欄（1-3）多 40。
	if ap != base-40+40 {
		t.Fatalf("造價是 %d，應該是 %d", ap, base)
	}
}

// 挑選單第 n 項寫回去之後，讀回來要停在同一項——骰式那兩欄中間隔著
// `MAGIC.0` 的 `0x200` 值表，來回一趟才看得出對不對。
func TestEveryFieldRoundTrips(t *testing.T) {
	d := spellDictionary(t)
	for field := 0; field < SpellFieldCount; field++ {
		options := SpellFieldOptions(d, vartwelyrmi(), field)
		for n := range options {
			sp := vartwelyrmi()
			SetSpellField(d, &sp, field, n)
			if got := SpellFieldValue(d, sp, field); got != n {
				t.Fatalf("欄位 %d 寫第 %d 項，讀回第 %d 項", field, n, got)
			}
		}
	}
}

// 骰式那兩欄只換中間六個位元：bit0 與 bit7 不能被動到。
func TestDiceKeepsOuterBits(t *testing.T) {
	d := spellDictionary(t)
	sp := vartwelyrmi()
	sp.PowerCode = 0x81
	for n := range d.Dice {
		SetSpellField(d, &sp, SpellFieldPower, n)
		if sp.PowerCode&0x81 != 0x81 {
			t.Fatalf("第 %d 項寫成 0x%02X，bit0 與 bit7 被動到了", n, sp.PowerCode)
		}
	}
}

// 持續回合那一欄只看得到前二十四項，強度那一欄看得到四十八項（`0x17E76`）。
func TestTurnColumnHasFewerOptions(t *testing.T) {
	d := spellDictionary(t)
	sp := vartwelyrmi()
	turns := len(SpellFieldOptions(d, sp, SpellFieldTurn))
	power := len(SpellFieldOptions(d, sp, SpellFieldPower))
	if turns != 0x18 || power != 0x30 {
		t.Fatalf("兩欄的選項是 %d／%d，應該是 24／48", turns, power)
	}
}

// 換掉某一格的類，那一格的種要跟著換選單。
func TestSubclassMenuFollowsItsClass(t *testing.T) {
	d := spellDictionary(t)
	sp := vartwelyrmi()
	SetSpellField(d, &sp, SpellFieldClassFirst, 5) // Elemental
	got := SpellFieldOptions(d, sp, SpellFieldSubFirst)
	want := []string{"None", "Golem", "Sylph", "Djinn", "Salamandar"}
	if len(got) != len(want) {
		t.Fatalf("Elemental 底下有 %d 種：%v", len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("第 %d 種是 %q，應該是 %q", i, got[i], want[i])
		}
	}
}
