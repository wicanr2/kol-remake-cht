package town

import (
	"path/filepath"
	"strings"
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

// 造價：四格護符 ＋ 屬性 ＋ 射程 ＋ 強度那一欄；持續那一欄不算。四格那張表是
// 唯一兩種資源不同額的——Legendary／Blom 是 400 金冠但只要 200 冒險點數。
func TestSpellCostOfOriginalRecord(t *testing.T) {
	d := spellDictionary(t)
	ap, gc := SpellCost(d, vartwelyrmi())
	// 金冠 400＋90＋200＋140，冒險點數 200＋90＋200＋140。
	if gc != 830 || ap != 630 {
		t.Fatalf("造價是 %d 冒險點數／%d 金冠，應該是 630／830", ap, gc)
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

// 走完整條路：SCROLL 選單第 1 列 → 挑一筆 → 改一欄 → 離開 → 答應存檔 →
// 扣冒險點數與金冠、寫回那一格。
func TestSpellEditorFullPath(t *testing.T) {
	d := spellDictionary(t)
	s := scrollSession(t, 1000, 109)
	s.Magic = d
	s.Party[0].Order = hostelOrder()
	s.Party[0].Spells = []Spell{vartwelyrmi()}
	s.Party[0].Sheet.AdventurePoints = 5000
	gold := s.Party[0].Gold

	s.StartScroll()
	s.MoveScroll(1)
	yes(s) // 第 1 列：改法術
	yes(s) // 挑第一筆
	if !s.SpellEditing() {
		t.Fatalf("沒進編輯頁：%+v", s.Inside.List)
	}
	page := s.Inside.SpellEdit
	if page == nil || page.Picture != SpellEditPicture {
		t.Fatalf("編輯頁是 %+v", page)
	}
	if page.Gold != 830 || page.Adventure != 630 {
		t.Fatalf("造價畫成 %d 冒險點數／%d 金冠", page.Adventure, page.Gold)
	}
	if page.Fields[SpellFieldName].Text != "Vartwelyrmi" ||
		page.Fields[SpellFieldStat].Text != "Fatigue" {
		t.Fatalf("欄位畫成 %+v", page.Fields)
	}
	// 游標在第 0 欄，只有那一欄是亮的。
	if page.Fields[0].Color != SpellCursorColor ||
		page.Fields[1].Color != SpellFieldColor {
		t.Fatalf("游標的顏色是 %d／%d", page.Fields[0].Color, page.Fields[1].Color)
	}

	// 走到 STAT 那一欄，換成 Strength。
	for i := 0; i < SpellFieldStat; i++ {
		s.MoveScroll(1)
	}
	yes(s) // 開選單
	if s.Inside.SpellEdit.Menu == nil {
		t.Fatalf("選單沒開")
	}
	for s.scroll.edit.cursor > 1 {
		s.MoveScroll(-1)
	}
	for s.scroll.edit.cursor < 1 {
		s.MoveScroll(1)
	}
	yes(s) // 選定
	if got := s.Inside.SpellEdit.Fields[SpellFieldStat].Text; got != "Strength" {
		t.Fatalf("屬性換成 %q", got)
	}
	want := s.Inside.SpellEdit.Adventure
	wantGold := s.Inside.SpellEdit.Gold

	// 離開 → 答應存檔。
	no(s)
	if s.Inside.SpellEdit != nil || len(s.Inside.Dialog) == 0 {
		t.Fatalf("沒問存不存：%+v", s.Inside.Dialog)
	}
	yes(s)
	if s.Scrolling() {
		t.Fatalf("存完應該收掉 SCROLL")
	}
	if got := s.Party[0].Spells[0].Effect; got != 1 {
		t.Fatalf("寫回去的屬性是 %d", got)
	}
	if got := s.Party[0].Sheet.AdventurePoints; got != 5000-want {
		t.Fatalf("冒險點數剩 %d，應該是 %d", got, 5000-want)
	}
	if got := s.Party[0].Gold; got != gold-wantGold {
		t.Fatalf("金冠剩 %d，應該是 %d", got, gold-wantGold)
	}
}

// 冒險點數不夠就不寫回，而且原版在這裡印的是金冠那一句（`0x1818A`）。
func TestSpellEditorRefusesWithoutPoints(t *testing.T) {
	d := spellDictionary(t)
	s := scrollSession(t, 1000, 109)
	s.Magic = d
	s.Party[0].Order = hostelOrder()
	s.Party[0].Spells = []Spell{vartwelyrmi()}
	s.Party[0].Sheet.AdventurePoints = 10

	s.StartScroll()
	s.MoveScroll(1)
	yes(s)
	yes(s)
	no(s)
	yes(s)
	if s.Party[0].Sheet.AdventurePoints != 10 {
		t.Fatalf("點數不夠卻扣了：%d", s.Party[0].Sheet.AdventurePoints)
	}
	if len(s.Inside.Dialog) == 0 ||
		!strings.Contains(s.Inside.Dialog[0], "not enough") {
		t.Fatalf("講的是 %v", s.Inside.Dialog)
	}
}

// 名稱那一欄是文字輸入，不是選單；收滿 12 個字就不再收。
func TestSpellNameTyping(t *testing.T) {
	d := spellDictionary(t)
	s := scrollSession(t, 1000, 109)
	s.Magic = d
	s.Party[0].Order = hostelOrder()
	s.Party[0].Spells = []Spell{vartwelyrmi()}
	s.StartScroll()
	s.MoveScroll(1)
	yes(s)
	yes(s)
	yes(s) // 第 0 欄：開輸入
	if s.Inside.SpellEdit.Typing == nil {
		t.Fatalf("沒進輸入")
	}
	s.BackspaceSpellName()
	for _, r := range "xyzabc" {
		s.TypeSpellName(r)
	}
	if got := *s.Inside.SpellEdit.Typing; got != "xyzabc" {
		t.Fatalf("打成 %q", got)
	}
	yes(s)
	if got := s.Inside.SpellEdit.Fields[SpellFieldName].Text; got != "Xyzabc" {
		t.Fatalf("名字變成 %q", got)
	}
}

// 教派只准法術對它那幾類生效：Blue Gem 是矮人與凱爾登，選別的就講一句話、
// 那一欄不動（`0x17FB1`，RE-128）。
func TestOrderLimitsTheClassColumn(t *testing.T) {
	d := spellDictionary(t)
	s := scrollSession(t, 1000, 109)
	s.Magic = d
	order := hostelOrder()
	order.Allow = []int{3, 4} // Dwarven／Kelden
	s.Party[0].Order = order
	s.Buildings.Doors[0].Order = order
	s.door = &s.Buildings.Doors[0]
	s.Party[0].Spells = []Spell{vartwelyrmi()}

	s.StartScroll()
	s.MoveScroll(1)
	yes(s)
	yes(s)
	s.MoveScroll(1) // 第 1 欄：第一格護符的類
	yes(s)          // 開選單
	for s.scroll.edit.cursor > 3 {
		s.MoveScroll(-1)
	}
	for s.scroll.edit.cursor < 3 {
		s.MoveScroll(1)
	}
	yes(s) // 選 Dwarven：准
	if got := s.Inside.SpellEdit.Fields[SpellFieldClassFirst].Text; got != "Dwarven" {
		t.Fatalf("准的那一類沒寫進去：%q", got)
	}
	yes(s)
	for s.scroll.edit.cursor < 8 {
		s.MoveScroll(1)
	}
	yes(s) // 選 Undead：不准
	if len(s.Inside.Dialog) == 0 {
		t.Fatalf("不准的那一類應該講一句話：%v", s.Inside.Dialog)
	}
	s.Continue()
	if got := s.Inside.SpellEdit.Fields[SpellFieldClassFirst].Text; got != "Dwarven" {
		t.Fatalf("不准的那一類卻寫進去了：%q", got)
	}
}
