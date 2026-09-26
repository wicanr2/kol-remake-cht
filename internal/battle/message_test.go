package battle

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// 用 KOLDATA 原文當樣本：代名詞、名字與武器名都要填對。
func testLookup(record int) string {
	switch record {
	case 0x10E:
		return "him"
	case 0x10F:
		return "her"
	case 0x110:
		return "his"
	case TheRecord:
		return "The]"
	}
	return ""
}

func testContext() MessageContext {
	return MessageContext{
		Actor:  Speaker{Name: "Poontang", Weapon: "Long Spear", Belt: "Dagger"},
		Target: Speaker{Name: "BINDERAK", Monster: true, Female: true},
		Lookup: testLookup,
	}
}

// `]` 換成空白（`sub_1A8FD:0x1A990`）。
func TestSubstituteBracket(t *testing.T) {
	c := testContext()
	if got := c.Substitute("The]"); got != "The " {
		t.Fatalf("得到 %q，應該是 %q", got, "The ")
	}
	if got := c.Substitute(", and]"); got != ", and " {
		t.Fatalf("得到 %q", got)
	}
}

// 代名詞：所有格是 0x110 − 性別、受格是 0x10E + 性別（`sub_265F3`）。
func TestSubstitutePronouns(t *testing.T) {
	c := testContext()
	// 行動者是男的：所有格 his、受格 him。
	if got := c.Substitute("kicks|B leg"); got != "kicks his leg" {
		t.Errorf("得到 %q", got)
	}
	if got := c.Substitute("making|D unconscious!"); got != "making him unconscious!" {
		t.Errorf("得到 %q", got)
	}
	// 目標是女的：兩種格都是 her。
	if got := c.Substitute("|A sword"); got != " her sword" {
		t.Errorf("得到 %q", got)
	}
	if got := c.Substitute("making|C unconscious!"); got != "making her unconscious!" {
		t.Errorf("得到 %q", got)
	}
}

// `|F` 是目標的名字，敵人前面要補 `The `（`sub_265F3:0x2673A`）。
func TestSubstituteTargetName(t *testing.T) {
	c := testContext()
	if got := c.Substitute(" hits|F's]"); got != " hits The BINDERAK's " {
		t.Fatalf("得到 %q", got)
	}
	// 隊員直接印名字，不補 The。
	c.Target = Speaker{Name: "Theodoric"}
	if got := c.Substitute(" hits|F's]"); got != " hits Theodoric's " {
		t.Fatalf("得到 %q", got)
	}
}

// `|E` 是行動者手上的武器、`|I` 是 +0x190 那一格。
func TestSubstituteEquipment(t *testing.T) {
	c := testContext()
	if got := c.Substitute("whips|B|E through the air."); got !=
		"whips his Long Spear through the air." {
		t.Fatalf("得到 %q", got)
	}
	if got := c.Substitute("puts|B|I in|B belt"); got != "puts his Dagger in his belt" {
		t.Fatalf("得到 %q", got)
	}
}

// `|N` 印數字、`|H` 與 `|S` 各有自己的來源；認不得的碼原樣留著。
func TestSubstituteOthers(t *testing.T) {
	c := testContext()
	c.Number = 42
	c.Item = "Shield"
	c.Scratch = "Helm"
	if got := c.Substitute("|N 點"); got != " 42 點" {
		t.Errorf("得到 %q", got)
	}
	if got := c.Substitute("a|H."); got != "a Shield." {
		t.Errorf("得到 %q", got)
	}
	if got := c.Substitute("|S"); got != " Helm" {
		t.Errorf("得到 %q", got)
	}
	// `|G` 是行動者的名字，與 `|F` 同一套規則（`sub_2687F`）。
	if got := c.Substitute("|G, run!"); got != " Poontang, run!" {
		t.Errorf("得到 %q", got)
	}
	// 認不得的碼原樣留著，不默默吞掉。
	if got := c.Substitute("|Z!"); got != " |Z!" {
		t.Errorf("認不得的碼應該原樣留著，得到 %q", got)
	}
}

// 一手的整句：行動者的名字加一個空白，再接記錄的內容（`sub_2685A`）。
func TestMessageLine(t *testing.T) {
	c := testContext()
	if got := c.Line("is resting!"); got != "Poontang is resting!" {
		t.Fatalf("得到 %q", got)
	}
	// 敵人動的時候名字前面要補 The。
	c.Actor = Speaker{Name: "BINDERAK", Monster: true}
	if got := c.Line("moves!"); got != "The BINDERAK moves!" {
		t.Fatalf("得到 %q", got)
	}
}

// 沒有 Lookup 時代名詞退回內建的英文，規則層自己也測得動。
func TestSubstituteWithoutLookup(t *testing.T) {
	c := MessageContext{Actor: Speaker{Name: "A"}, Target: Speaker{Female: true}}
	if got := c.Substitute("|B/|C"); got != " his/ her" {
		t.Fatalf("得到 %q", got)
	}
}

// 斷行：從行寬處往回找空白，跳過斷點上的空白（`sub_1C1BF`）。
func TestWrapMessage(t *testing.T) {
	// 短句不斷行。
	if got := WrapMessage("Poontang is resting!"); len(got) != 1 ||
		got[0] != "Poontang is resting!" {
		t.Fatalf("短句不該斷行，得到 %q", got)
	}
	// 第一行的上限是 0x21 − 3 ＝ 30 個字元，斷在最後一個空白。
	line := "The BINDERAK blow hits Poontang's right arm doing little damage!"
	got := WrapMessage(line)
	if len(got) < 2 {
		t.Fatalf("這一句應該斷成多行，得到 %q", got)
	}
	if len(got[0]) > 30 {
		t.Fatalf("第一行 %d 個字元，超過 30：%q", len(got[0]), got[0])
	}
	// 每一行都不該以空白開頭或結尾——斷點上的空白會被跳過。
	for i, l := range got {
		if l != "" && (l[0] == ' ' || l[len(l)-1] == ' ') {
			t.Errorf("第 %d 行的兩端不該有空白：%q", i, l)
		}
	}
	// 接回去要等於原句（斷點的空白換成換行）。
	joined := ""
	for i, l := range got {
		if i > 0 {
			joined += " "
		}
		joined += l
	}
	if joined != line {
		t.Fatalf("接回去得到 %q", joined)
	}
}

// 整段沒有空白時硬斷，不讀到界外。
func TestWrapMessageNoSpace(t *testing.T) {
	long := ""
	for i := 0; i < 80; i++ {
		long += "x"
	}
	got := WrapMessage(long)
	if len(got) < 2 {
		t.Fatalf("八十個字元應該斷成多行，得到 %d 行", len(got))
	}
	total := 0
	for _, l := range got {
		total += len(l)
	}
	if total > len(long) {
		t.Fatalf("斷出來的字元數 %d 多過原句 %d", total, len(long))
	}
}

// 第一句：動作加結果（RE-088）。
func TestAttackLine(t *testing.T) {
	// 武器攻擊、命中。
	got := AttackLine(CmdWeapon, 2, Outcome{}, false)
	want := []int{WeaponActionBase + 2, WeaponAtRecord, CannotParryRecord}
	assertRecords(t, got, want)
	// 徒手、落空很多。
	got = AttackLine(CmdUnarmed, 1, Outcome{Miss: true, Margin: 0x20}, false)
	want = []int{UnarmedActionBase + 1, AtTargetRecord, MissWideRecord}
	assertRecords(t, got, want)
	// 落空的三段門檻。
	for _, c := range []struct {
		margin, record int
	}{{0x1F, MissWideRecord}, {0x0B, MissRecord}, {0x05, MissNarrowRecord}} {
		got = AttackLine(CmdWeapon, 1, Outcome{Miss: true, Margin: c.margin}, false)
		if got[len(got)-1] != c.record {
			t.Errorf("落空 %d 應該印 %#x，得到 %#x", c.margin, c.record, got[len(got)-1])
		}
	}
	// 被擋：近戰印 who ＋ 招架句，射擊印另一句。
	got = AttackLine(CmdWeapon, 1, Outcome{Parried: true, Message: 3}, false)
	assertRecords(t, got, []int{WeaponActionBase + 1, WeaponAtRecord,
		WhoRecord, ParryBase + 3})
	got = AttackLine(CmdMissile, 1, Outcome{Parried: true, Message: 3}, false)
	assertRecords(t, got, []int{MissileRecord, AtTargetRecord, MissileMissRecord})
	// 射擊的武器種類大於 0x7F 換成丟石頭。
	got = AttackLine(CmdMissile, 1, Outcome{}, true)
	if got[0] != BoulderRecord {
		t.Errorf("重武器應該印 %#x，得到 %#x", BoulderRecord, got[0])
	}
}

// 第二句：傷害（`sub_277F4`）。
func TestDamageLine(t *testing.T) {
	head := []int{TheBlowRecord, BlowKindBase, HitsRecord, LocationBase + 2}
	// 傷害為 0。
	assertRecords(t, DamageLine(Outcome{Location: 2}, 0, false, false),
		append(append([]int{}, head...), LittleRecord))
	// 三段輕重。
	for _, c := range []struct {
		damage, record int
	}{{1, SeriousRecord}, {0x0F, SeriousRecord}, {0x10, TerribleRecord},
		{0x17, TerribleRecord}, {0x18, GrievousRecord}} {
		got := DamageLine(Outcome{Damage: c.damage, Location: 2}, 0, false, false)
		if got[len(got)-2] != c.record {
			t.Errorf("傷害 %d 應該印 %#x，得到 %#x", c.damage, c.record, got[len(got)-2])
		}
		if got[len(got)-1] != WoundRecord {
			t.Errorf("結尾應該是 %#x", WoundRecord)
		}
	}
	// 手腳廢了。
	got := DamageLine(Outcome{Damage: 5, Location: 2, Limb: true}, 0, false, false)
	assertRecords(t, got, append(append([]int{}, head...), LimbUselessRecord))
	// 頭廢了：敵人是 killing it、隊員是 unconscious。
	headHit := []int{TheBlowRecord, BlowKindBase, HitsRecord, LocationBase}
	got = DamageLine(Outcome{Damage: 5, Location: 0, Limb: true}, 0, true, false)
	assertRecords(t, got, append(append([]int{}, headHit...), KilledRecord))
	got = DamageLine(Outcome{Damage: 5, Location: 0, Limb: true}, 0, false, false)
	assertRecords(t, got, append(append([]int{}, headHit...), UnconsciousRecord))
	// 倒下（`0x2796F`）：隊員 collapse、會碎的怪物 shatter、其餘怪物 die。
	down := Outcome{Damage: 5, Location: 2, Down: true}
	assertRecords(t, DamageLine(down, 0, false, false), append(append([]int{}, head...),
		MakingRecord, CollapseRecord, BloodLossRecord))
	assertRecords(t, DamageLine(down, 0, true, true), append(append([]int{}, head...),
		MakingRecord, ShatterRecord))
	assertRecords(t, DamageLine(down, 0, true, false), append(append([]int{}, head...),
		MakingRecord, DieRecord, BloodLossRecord))
	// shatters 只看怪物；隊員就算帶了也照 collapse。
	assertRecords(t, DamageLine(down, 0, false, true), append(append([]int{}, head...),
		MakingRecord, CollapseRecord, BloodLossRecord))
	// 法術用 0x142。
	got = DamageLine(Outcome{Location: 2}, 1, false, false)
	if got[1] != BlowKindBase+1 {
		t.Errorf("法術應該印 %#x，得到 %#x", BlowKindBase+1, got[1])
	}
}

func assertRecords(t *testing.T, got, want []int) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("記錄數 %d，應該是 %d：%#x vs %#x", len(got), len(want), got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("第 %d 筆是 %#x，應該是 %#x：%#x vs %#x", i, got[i], want[i], got, want)
		}
	}
}

// 中文譯表把 `The]` 刻意留空：文字庫回空字串時敵人名字前不補任何冠詞，
// 不能退回內建的英文 `The `。
func TestMonsterNameWithoutArticleWhenRecordBlank(t *testing.T) {
	c := testContext()
	c.Lookup = func(record int) string {
		if record == TheRecord {
			return ""
		}
		return testLookup(record)
	}
	c.Actor = Speaker{Name: "巨魔", Monster: true}
	if got := c.Line("移動了！"); got != "巨魔 移動了！" {
		t.Fatalf("得到 %q", got)
	}
}

// 中文譯文沒有空白時硬斷，不能切在多位元組字元中間。
func TestWrapMessageKeepsRunesWhole(t *testing.T) {
	text := strings.Repeat("這道法術產生了明顯負面的效果", 4)
	for _, line := range WrapMessage(text) {
		if !utf8.ValidString(line) {
			t.Fatalf("斷行切壞了字元：%q", line)
		}
	}
}
