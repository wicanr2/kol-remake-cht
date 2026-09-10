//go:build oracle

package oracle

import (
	"testing"

	"github.com/wicanr2/dosgolem/oracle"

	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/world"
)

// 原版戰鬥規則的位址（KOLTOWN-unpacked.EXE，IDA 線性；RE-129／RE-130）。
const (
	// dgroupBase 是 DGROUP 的 IDA 線性基底。筆記裡的 `ds:XXXX` 都以它為準。
	dgroupBase = 0x2DED0
	// fatigueRatioFn 是 `sub_23683(i)`：把疲勞折算成比例寫進 `word_2FB4A`。
	fatigueRatioFn = 0x23683
	// fatigueRatioOut 是它寫結果的那一格（`word_2FB4A`）。
	fatigueRatioOut = 0x2FB4A
	// actorFatigue／actorStamina 是參戰者 i 的疲勞與「力量 ＋ 健康」（`ds:` 偏移）。
	actorFatigue = 0x322E
	actorStamina = 0x1C50
)

// prepare 把原版跑到 `_main` 並把 DGROUP 基底對好。
//
// # 為什麼這樣就夠
//
// 要驗的是**規則**：同一組輸入，原版與 remake 算出同一個數。輸入直接寫進原版的
// 記憶體，規則直接呼叫原版的常式——不必等遊戲自己走到那一步，也不必讓兩邊的亂數
// 序列一樣（種子、抽取順序、抽幾次都不同，**強求同步是在解錯的題目**）。
// 需要骰值的常式用 `Stub` 把它釘成常數，剩下的差異就只剩規則本身。
func prepare(t *testing.T) *oracle.Oracle {
	t.Helper()
	o := load(t)
	if err := o.RunUntil(oracle.At(o.IDA(mainEnter)), oracle.Budget(2_000_000)); err != nil {
		t.Fatalf("跑不到 _main：%v", err)
	}
	// 正對照：筆記說 DGROUP 在 IDA 線性 0x2DED0，程式自己的 DS 必須指到同一處。
	// 對不上就表示底下每一個 `ds:` 位址都在讀別的東西——**而那不會報錯**。
	o.SyncDGroupFromDS()
	if got, want := uint32(o.DGroupSeg())*16, o.IDA(dgroupBase).Linear(); got != want {
		t.Fatalf("程式的 DS 在線性 %05X，筆記說 DGROUP 是 IDA %#x（線性 %05X）",
			got, dgroupBase, want)
	}
	return o
}

// 疲勞比例：原版與 remake 算出同一個數（RE-130）。
//
// 這一條驗的是本輪修掉的那個對錯——`sub_23683` 算的是
// `50 × 疲勞 ÷ (力量 ＋ 健康)`，remake 原本寫成 `50 × 累積傷害 ÷ 昏迷門檻`。
// 規劃鍵值、執行鍵值、攻防值、力量加成與敵人 AI 的撤退門檻五處都吃它，所以
// 這一格錯了，整場戰鬥的數字都跟著錯。
func TestFatigueRatioMatchesOriginal(t *testing.T) {
	o := prepare(t)
	const slot = 0
	for _, c := range []struct{ fatigue, stamina int }{
		{0, 30}, {1, 30}, {15, 30}, {29, 30}, {30, 30}, {60, 30},
		{7, 12}, {100, 40}, {3, 200}, {199, 200}, {1, 1}, {255, 255},
	} {
		o.WriteU16(o.DS(actorFatigue+2*slot), uint16(c.fatigue))
		o.WriteU16(o.DS(actorStamina+2*slot), uint16(c.stamina))
		o.WriteU16(o.IDA(fatigueRatioOut), 0xDEAD) // 沒被寫就看得出來

		if _, err := o.Call(o.IDA(fatigueRatioFn), uint16(slot)); err != nil {
			t.Fatalf("疲勞 %d／體力 %d：%v", c.fatigue, c.stamina, err)
		}
		got := int(o.Word(o.IDA(fatigueRatioOut)))
		if got == 0xDEAD {
			t.Fatalf("疲勞 %d／體力 %d：`word_2FB4A` 沒被寫——"+
				"呼叫的可能不是這一支，或它把結果放在別處", c.fatigue, c.stamina)
		}
		a := battle.Actor{Fatigue: c.fatigue, Stamina: c.stamina}
		if want := a.FatigueRatio(); got != want {
			t.Errorf("疲勞 %d／體力 %d：原版 %d，remake %d",
				c.fatigue, c.stamina, got, want)
		}
	}
}

// 攻防技能的罰值（RE-129）。
const (
	// skillPenaltyFn 是 `sub_26CD5(i)`：把兩個罰值折進當前技能。
	skillPenaltyFn = 0x26CD5
	// offSkill／defSkill 是它改的那兩格（`word_34A12`／`word_32380`）。
	offSkill = 0x34A12
	defSkill = 0x32380
	// actorBase／actorStride 是參戰者陣列（`ds:6712`，18 筆 × 50 byte）。
	actorBase, actorStride = 0x6712, 50
	// penaltyOff／penaltyDef 是記錄裡的兩格罰值。
	penaltyOff, penaltyDef = 0x2E, 0x2F
)

// 罰值一點抵十、扣到 0 為止——原版與 remake 算出同一個數。
//
// 這一支是**每次取用時才折算**的：`sub_26AC2` 每要用一次技能就重算一遍，
// 所以記錄裡的技能本身不變。remake 的 `OffenceSkill()`／`DefenceSkill()` 是
// 同一個意思，這一條把兩邊釘在一起。
func TestSkillPenaltyMatchesOriginal(t *testing.T) {
	o := prepare(t)
	const slot = 0
	rec := uint16(actorBase + actorStride*slot)
	for _, c := range []struct{ skill, pen int }{
		{0, 0}, {100, 0}, {100, 1}, {100, 9}, {100, 10}, {100, 11},
		{55, 5}, {55, 6}, {7, 1}, {250, 25}, {250, 24}, {3, 200},
	} {
		o.WriteU16(o.IDA(offSkill), uint16(c.skill))
		o.WriteU16(o.IDA(defSkill), uint16(c.skill))
		o.WriteU8(o.DS(rec+penaltyOff), uint8(c.pen))
		o.WriteU8(o.DS(rec+penaltyDef), uint8(c.pen))

		if _, err := o.Call(o.IDA(skillPenaltyFn), uint16(slot)); err != nil {
			t.Fatalf("技能 %d／罰值 %d：%v", c.skill, c.pen, err)
		}
		a := battle.Actor{
			Attack: c.skill, Parry: c.skill,
			SkillPenaltyOff: c.pen, SkillPenaltyDef: c.pen,
		}
		if got, want := int(o.Word(o.IDA(offSkill))), a.OffenceSkill(); got != want {
			t.Errorf("攻擊技能 %d 罰 %d：原版 %d，remake %d", c.skill, c.pen, got, want)
		}
		if got, want := int(o.Word(o.IDA(defSkill))), a.DefenceSkill(); got != want {
			t.Errorf("防禦技能 %d 罰 %d：原版 %d，remake %d", c.skill, c.pen, got, want)
		}
	}
}

// 額度與花費的夾擠（RE-131）。
const (
	// clampEffortFn 是 `sub_2526C()`：把「額度 ÷ 花費」夾進 1/3…1。不吃參數。
	clampEffortFn = 0x2526C
	// effortAllow／effortCost 是它改的那兩格（`word_30D06`／`word_2F940`）。
	effortAllow = 0x30D06
	effortCost  = 0x2F940
)

// 夾擠的四條分支：原版與 remake 走同一條。
//
// 這一支決定「動作越貴、額度越少，命中部位越偏向那三項」的倍率，所以它錯了
// 不會當機，只會讓命中部位的分布悄悄偏掉——**正是最難從畫面看出來的那種錯**。
func TestClampEffortMatchesOriginal(t *testing.T) {
	o := prepare(t)
	for _, c := range []struct{ allow, cost int }{
		{100, 50},  // 額度大於花費
		{50, 50},   // 相等
		{50, 100},  // 花費是額度的兩倍
		{50, 150},  // 剛好三倍
		{50, 151},  // 超過三倍
		{50, 1000}, // 遠超過
		{0, 10},    // 額度用完
		{0, 0},     // 兩個都零
		{1, 3}, {1, 4}, {7, 20}, {7, 21},
	} {
		o.WriteU16(o.IDA(effortAllow), uint16(c.allow))
		o.WriteU16(o.IDA(effortCost), uint16(c.cost))

		if _, err := o.Call(o.IDA(clampEffortFn)); err != nil {
			t.Fatalf("額度 %d／花費 %d：%v", c.allow, c.cost, err)
		}
		gotAllow := int(int16(o.Word(o.IDA(effortAllow))))
		gotCost := int(int16(o.Word(o.IDA(effortCost))))
		wantAllow, wantCost := battle.ClampEffort(c.allow, c.cost)
		if gotAllow != wantAllow || gotCost != wantCost {
			t.Errorf("額度 %d／花費 %d：原版 (%d, %d)，remake (%d, %d)",
				c.allow, c.cost, gotAllow, gotCost, wantAllow, wantCost)
		}
	}
}

// 釘死骰值：stub 掉 `rand()`，上層的 `rand(n)` 就變成決定性的。
//
// # 這一條在證明什麼
//
// 前面三支規則不擲骰，所以「同一組輸入」就夠了。會擲骰的規則不行——兩邊的亂數
// 序列**本來就不會一樣**（種子、抽取順序、抽幾次都不同），強求同步是在解錯的
// 題目。做法是把骰值釘成常數，剩下的差異就只剩規則本身。
//
// 判準是 stub 要在**原版自己的呼叫鏈**上生效，不是只有被直接呼叫的那一層：
// `rand(n)` 內部呼叫 `rand()`，stub 掉後者，前者必須回 `釘死的值 % n + 1`。
// 這一條不成立的話，任何「上層規則 → 擲骰」的對拍都做不了。
func TestStubbedDiceReachInnerCalls(t *testing.T) {
	o := prepare(t)
	for _, fixed := range []uint32{0, 1, 7, 1000, 0x7FFE} {
		o.StubValue(o.IDA(randEnter), fixed)
		for _, n := range []int{2, 3, 6, 20, 100} {
			got, err := o.Call(o.IDA(rollEnter), uint16(n))
			if err != nil {
				t.Fatalf("釘死 %d、rand(%d)：%v", fixed, n, err)
			}
			roll := int(int16(uint16(got)))
			if want := int(fixed)%n + 1; roll != want {
				t.Fatalf("釘死 %d：原版 rand(%d) 回 %d，期望 %d"+
					"（stub 沒有到達內層呼叫）", fixed, n, roll, want)
			}
		}
	}
	// 取消之後種子要重新前進——不然後面的測試會拿到凍住的骰。
	o.Stub(o.IDA(randEnter), nil)
	before := seedOf(o)
	if _, err := o.Call(o.IDA(rollEnter), 6); err != nil {
		t.Fatal(err)
	}
	if seedOf(o) == before {
		t.Fatal("取消 stub 之後種子沒有前進——常式還被攔著")
	}
}

// 一手的花費（RE-131）。
const (
	// actionCostFn 是 `sub_251B0(a, b)`：`word_2F940 = a × 武器速度 + b × 負重`。
	actionCostFn = 0x251B0
	// costOut 是它寫的那一格（`word_2F940`，與夾擠那支的 effortCost 同一格）。
	costOut = effortCost
	// weaponSpeed／encumbrance 是它讀的兩格（`word_37CF0`／`word_30C88`）。
	weaponSpeed = 0x37CF0
	encumbrance = 0x30C88
	// plannedFlag 是參戰者記錄的 `+0x22`（這一回合規劃過了，RE-081）。
	// 非零而且這一手徒手時，速度那一項改從穿戴品取——這一條測基本路徑，
	// 所以**十八筆全部清零**，不必知道它讀的是哪一個行動者。
	plannedFlag = 0x22
	actorSlots  = 18
	// weightFn 是 `sub_24FD3()`：從穿戴品重新算出速度與負重那兩格。
	// 要餵指定的速度與負重就得把它擋掉——不然它會用空白的記錄把兩格算成 0。
	weightFn = 0x24FD3
)

// 花費 ＝ 時間項 × 武器速度 ＋ 重量項 × 負重：原版與 remake 算出同一個數。
//
// 這是疲勞鏈的入口——每一手把「額度減花費」除以十扣進疲勞，而疲勞驅動整套撤退
// 判斷。算錯不會當機，只會讓戰鬥的長度悄悄偏掉。
func TestActionCostMatchesOriginal(t *testing.T) {
	o := prepare(t)
	for i := 0; i < actorSlots; i++ {
		o.WriteU8(o.DS(uint16(actorBase+actorStride*i+plannedFlag)), 0)
	}
	// `sub_251B0` 自己會先叫 `sub_24FD3` 從穿戴品重算速度與負重，把寫進去的值
	// 蓋掉——沒有裝備的空白記錄算出來是 0，於是每一組都回 0。**擋掉它**，
	// 這一條才問得到「給定速度與負重時，花費是怎麼算的」。
	//
	// 這正是 stub 的第二個用途：不只釘死骰值，也用來**切斷上游的依賴**，
	// 讓一支常式變回純函式。
	o.StubValue(o.IDA(weightFn), 0)
	for _, c := range []struct{ speed, enc, timeIdx, weightIdx int }{
		{10, 5, 2, 1}, {10, 5, 0, 0}, {10, 5, 15, 1}, {10, 5, 6, 2},
		{1, 1, 1, 1}, {30, 0, 4, 2}, {0, 30, 4, 2}, {7, 13, 3, 6},
		{100, 100, 15, 2}, {12, 9, 2, 3},
	} {
		o.WriteU16(o.IDA(weaponSpeed), uint16(c.speed))
		o.WriteU16(o.IDA(encumbrance), uint16(c.enc))
		o.WriteU16(o.IDA(costOut), 0xDEAD)

		if _, err := o.Call(o.IDA(actionCostFn), uint16(c.timeIdx), uint16(c.weightIdx)); err != nil {
			t.Fatalf("速度 %d／負重 %d、(%d, %d)：%v",
				c.speed, c.enc, c.timeIdx, c.weightIdx, err)
		}
		got := int(int16(o.Word(o.IDA(costOut))))
		if got == 0xDEAD-0x10000 || got == 0xDEAD {
			t.Fatalf("速度 %d／負重 %d、(%d, %d)：`word_2F940` 沒被寫",
				c.speed, c.enc, c.timeIdx, c.weightIdx)
		}
		a := battle.Actor{WeaponSpeed: c.speed, Encumbrance: c.enc}
		if want := battle.ActionCost(a, c.timeIdx, c.weightIdx); got != want {
			t.Errorf("速度 %d／負重 %d、(%d, %d)：原版 %d，remake %d",
				c.speed, c.enc, c.timeIdx, c.weightIdx, got, want)
		}
	}
}

// 誰是敵人（`sub_22229(i)`）。
const sideFn = 0x22229

// remake 假設「參戰者 0…5 是隊員、6…17 是敵人」（`PartySlots = 6`）。這一條先問
// 原版判準在哪：十八筆空白記錄全部回 0，所以**不是**靠索引分的。
//
// 定位法是逐格探測：把記錄的第 k 個 byte 設成非零，看回傳值變不變。50 次呼叫，
// 比開 IDA 便宜得多，而且答案是原版自己在執行期給的。
func TestSideFlagIsARecordField(t *testing.T) {
	o := prepare(t)
	const slot = 0
	rec := uint16(actorBase + actorStride*slot)

	base, err := o.Call(o.IDA(sideFn), slot)
	if err != nil {
		t.Fatal(err)
	}
	if base != 0 {
		t.Fatalf("空白記錄回 %d，期望 0", base)
	}
	var hits []int
	for k := 0; k < actorStride; k++ {
		o.WriteU8(o.DS(rec+uint16(k)), 1)
		v, err := o.Call(o.IDA(sideFn), slot)
		if err != nil {
			t.Fatalf("第 %d 格：%v", k, err)
		}
		if v != 0 {
			hits = append(hits, k)
		}
		o.WriteU8(o.DS(rec+uint16(k)), 0)
	}
	t.Logf("把記錄的第 k 個 byte 設 1 之後 `sub_22229` 回非零的 k：%v", hits)
	if len(hits) == 0 {
		t.Skip("五十格都不影響——判準在記錄之外（另一個全域或索引以外的東西）")
	}
}

// 敵人的攻防技能（`sub_26AC2`，RE-080／RE-129）。
const (
	// skillFn 是 `sub_26AC2(i)`：算出「這一刻」的攻防技能放進兩個全域。
	skillFn = 0x26AC2
	// bodyPtr 是 `word_39A48`：指向本體記錄的 near 指標（`ds:` 偏移）。
	bodyPtr = 0x39A48
	// monsterBase／monsterStride 是怪物記錄陣列（`ds:CD92`，一筆 128 byte）。
	monsterBase, monsterStride = 0xCD92, 128
	// monsterSkill 是怪物記錄裡攻防技能那兩組大端 16-bit。
	monsterSkill = 0x17
	// encounterLevel 是 `word_3B26A`：這一場遭遇的等級。
	encounterLevel = 0x3B26A
	// sideFlag 是參戰者記錄的 `+0x01`：非零＝敵人（實跑逐格探出來的，
	// `TestSideFlagIsARecordField`）。
	sideFlag = 0x01
)

// 敵人的技能要按等級縮放：`技能 × 等級 ÷ 5`。
//
// remake 在建戰鬥時就折算好（`state.go` 的 `monster.AttackSkill * site.Level / 5`），
// 這一條把那個折算釘在原版上。**整數除法的截斷位置是重點**——先乘後除與先除後乘
// 差得出來，而差出來的那一兩點會一路走進命中判定。
func TestMonsterSkillScalesByLevel(t *testing.T) {
	o := prepare(t)
	const slot = 6 // 第一個敵人格
	rec := uint16(actorBase + actorStride*slot)
	body := uint16(monsterBase + monsterStride*(slot-6))

	o.WriteU8(o.DS(rec+sideFlag), 1)     // 這一格是敵人
	o.WriteU8(o.DS(rec+plannedFlag), 0)  // 不要走徒手那條分支
	o.WriteU8(o.DS(rec+penaltyOff), 0)   // 罰值另有測試
	o.WriteU8(o.DS(rec+penaltyDef), 0)
	o.WriteU16(o.IDA(bodyPtr), body)

	for _, c := range []struct{ atk, par, level int }{
		{100, 80, 5}, {100, 80, 1}, {100, 80, 3}, {100, 80, 10},
		{1234, 999, 5}, {1234, 999, 2}, {7, 3, 4}, {0, 0, 5},
		{65000, 60000, 1}, {33, 44, 7},
	} {
		// 攻防技能各一組大端 16-bit。
		o.WriteU8(o.DS(body+monsterSkill), uint8(c.atk>>8))
		o.WriteU8(o.DS(body+monsterSkill+1), uint8(c.atk))
		o.WriteU8(o.DS(body+monsterSkill+2), uint8(c.par>>8))
		o.WriteU8(o.DS(body+monsterSkill+3), uint8(c.par))
		o.WriteU16(o.IDA(encounterLevel), uint16(c.level))
		o.WriteU16(o.IDA(offSkill), 0xDEAD)
		o.WriteU16(o.IDA(defSkill), 0xDEAD)

		if _, err := o.Call(o.IDA(skillFn), uint16(slot)); err != nil {
			t.Fatalf("攻 %d／防 %d／等級 %d：%v", c.atk, c.par, c.level, err)
		}
		gotAtk := int(o.Word(o.IDA(offSkill)))
		gotPar := int(o.Word(o.IDA(defSkill)))
		if gotAtk == 0xDEAD || gotPar == 0xDEAD {
			t.Fatalf("攻 %d／防 %d／等級 %d：兩格沒被寫——走的可能不是敵人那條分支",
				c.atk, c.par, c.level)
		}
		if want := c.atk * c.level / 5; gotAtk != want {
			t.Errorf("攻擊技能 %d × 等級 %d ÷ 5：原版 %d，remake 算 %d",
				c.atk, c.level, gotAtk, want)
		}
		if want := c.par * c.level / 5; gotPar != want {
			t.Errorf("防禦技能 %d × 等級 %d ÷ 5：原版 %d，remake 算 %d",
				c.par, c.level, gotPar, want)
		}
	}
}

// 溢位：原版那一步是 16 位元乘法，remake 的 `int` 不會環繞。
//
// **實際資料到不了那裡**——`monsters.json` 的攻擊技能最大 300，乘上等級上限 10
// 是 3000，離 65535 很遠。這一條把那個邊界量出來當紀錄，而不是留著不知道：
// 對得上就是兩邊在整個 16 位元範圍內都一致，對不上就寫明差異從哪一點開始。
func TestMonsterSkillOverflowBoundary(t *testing.T) {
	o := prepare(t)
	const slot = 6
	rec := uint16(actorBase + actorStride*slot)
	body := uint16(monsterBase + monsterStride*(slot-6))
	o.WriteU8(o.DS(rec+sideFlag), 1)
	o.WriteU8(o.DS(rec+plannedFlag), 0)
	o.WriteU8(o.DS(rec+penaltyOff), 0)
	o.WriteU8(o.DS(rec+penaltyDef), 0)
	o.WriteU16(o.IDA(bodyPtr), body)

	for _, c := range []struct{ skill, level int }{
		{300, 10},   // 實際資料的上限
		{6553, 10},  // 剛好不溢位
		{6554, 10},  // 剛好溢位
		{20000, 10}, // 遠超過
	} {
		o.WriteU8(o.DS(body+monsterSkill), uint8(c.skill>>8))
		o.WriteU8(o.DS(body+monsterSkill+1), uint8(c.skill))
		o.WriteU16(o.IDA(encounterLevel), uint16(c.level))
		if _, err := o.Call(o.IDA(skillFn), uint16(slot)); err != nil {
			t.Fatalf("技能 %d × 等級 %d：%v", c.skill, c.level, err)
		}
		got := int(o.Word(o.IDA(offSkill)))
		plain := c.skill * c.level / 5
		if got != plain {
			t.Logf("技能 %d × 等級 %d ÷ 5：原版 %d，直接算 %d"+
				"（16 位元環繞；實際資料最大 300×10＝3000，到不了這裡）",
				c.skill, c.level, got, plain)
			continue
		}
		t.Logf("技能 %d × 等級 %d ÷ 5：兩邊都是 %d", c.skill, c.level, got)
	}
}

// 負重與武器速度本身（`sub_24FD3`，RE-080）。
const (
	// weightOut／speedOut 是它寫的兩格（`word_30C88`／`word_37CF0`）。
	weightOut = encumbrance
	speedOut  = weaponSpeed
	// shieldGradeOut 是盾牌等級（`word_34A0C`）。
	shieldGradeOut = 0x34A0C
	// 怪物記錄裡的物品格：武器 +0x20、三件護甲 +0x30／+0x40／+0x50、盾牌 +0x60，
	// 一格 16 byte，重量在格內 +0x0A（`tools/extract_monsters.py`）。
	itemWeight  = 0x0A
	monWeapon   = 0x20
	monShield   = 0x60
	actorLevel  = 0x02 // 參戰者記錄的等級
)

var monArmour = [3]uint16{0x30, 0x40, 0x50}

// 探測：`sub_24FD3` 實際讀的是哪一筆記錄的哪幾格。
//
// 它**不吃參數**（RESEARCH-LOG 記的就是 `sub_24FD3()`），用的是「當前行動者」的
// 全域——所以照怪物那一側設好資料再傳 slot 進去，原版一律回 0。
//
// 逐格探測：把候選記錄的第 k 個 byte 設成 0xFF，看武器速度變不變。這個手法在
// 敵我旗標那一次一次就問到了答案（五十次呼叫定位到 `+0x01`），比讀反組譯快。
func TestProbeEncumbranceInputs(t *testing.T) {
	o := prepare(t)
	for _, region := range []struct {
		name string
		base uint16
		size uint16
	}{
		{"參戰者 0 的記錄", uint16(actorBase), uint16(actorStride)},
		{"角色記錄 0（ds:BB88 起 0x180）", 0xBB88, 0x180},
		{"角色記錄 0 的穿戴欄（+0x100 起 0xC0）", 0xBB88 + 0x100, 0xC0},
	} {
		var hits []string
		for k := uint16(0); k < region.size; k++ {
			old := o.Byte(o.DS(region.base + k))
			o.WriteU8(o.DS(region.base+k), 0xFF)
			o.WriteU16(o.IDA(speedOut), 0)
			o.WriteU16(o.IDA(weightOut), 0)
			if _, err := o.Call(o.IDA(weightFn)); err != nil {
				t.Fatalf("%s 第 %#x 格：%v", region.name, k, err)
			}
			sp := o.Word(o.IDA(speedOut))
			en := o.Word(o.IDA(weightOut))
			if sp != 0 || en != 0 {
				hits = append(hits, fmtHit(k, sp, en))
			}
			o.WriteU8(o.DS(region.base+k), old)
		}
		t.Logf("%s：%d 格影響輸出 %v", region.name, len(hits), hits)
	}
}

func fmtHit(k, sp, en uint16) string {
	return "+0x" + hex4(k) + "(速度" + dec(sp) + "/負重" + dec(en) + ")"
}

func hex4(v uint16) string {
	const d = "0123456789ABCDEF"
	return string([]byte{d[(v>>8)&0xF], d[(v>>4)&0xF], d[v&0xF]})
}

func dec(v uint16) string {
	if v == 0 {
		return "0"
	}
	var b []byte
	for v > 0 {
		b = append([]byte{byte('0' + v%10)}, b...)
		v /= 10
	}
	return string(b)
}

// 負重 ＝ 係數 × (武器重量÷2 ＋ 三件護甲 ＋ 盾牌) ÷ 100，速度 ＝ 武器重量 × 係數 ÷ 100。
//
// 係數是 `70 − (等級−1) × 5`。這一支是上一輪抓到缺口的那一支——原版每次算一手的
// 花費都重跑它，所以戰鬥中換武器兩個數字會立刻變。這一條把公式本身釘在原版上。
//
// 輸入的位置是**逐格探出來的**（`TestProbeEncumbranceInputs`）：它不吃參數，用的是
// 當前行動者（參戰者 0）與它的本體記錄 `ds:BB88`；手上武器在 `+0x30`（重量 `+0x3A`），
// 三件護甲的重量在 `+0x10A`／`+0x11A`／`+0x12A`，等級在參戰者記錄 `+0x02`。
func TestEncumbranceMatchesOriginal(t *testing.T) {
	o := prepare(t)
	const (
		slot     = 0
		charBase = 0xBB88 // 角色記錄 0
		charWeapon = 0x3A // 手上武器的重量
		charShield = 0x160
	)
	charArmour := [3]uint16{0x10A, 0x11A, 0x12A}
	rec := uint16(actorBase + actorStride*slot)
	o.WriteU8(o.DS(rec+sideFlag), 0) // 隊員
	// 沒有盾牌：那一格的第一個 byte 是 0（同 `extract_monsters.py` 的判準）。
	for k := uint16(0); k < 16; k++ {
		o.WriteU8(o.DS(charBase+charShield+k), 0)
	}

	for _, c := range []struct {
		level, weapon int
		armour        [3]int
	}{
		{1, 40, [3]int{10, 20, 10}},
		{1, 0, [3]int{0, 0, 0}},
		{3, 40, [3]int{10, 20, 10}},
		{5, 100, [3]int{30, 40, 30}},
		{10, 12, [3]int{1, 2, 3}},
		{2, 255, [3]int{0, 0, 0}},
		{7, 33, [3]int{5, 5, 5}},
		{1, 100, [3]int{90, 90, 87}}, // 實際資料的上限：武器 100、護甲加總 267
	} {
		o.WriteU8(o.DS(rec+actorLevel), uint8(c.level))
		o.WriteU8(o.DS(charBase+charWeapon), uint8(c.weapon))
		for k, w := range c.armour {
			o.WriteU8(o.DS(charBase+charArmour[k]), uint8(w))
		}
		o.WriteU16(o.IDA(weightOut), 0xDEAD)
		o.WriteU16(o.IDA(speedOut), 0xDEAD)

		if _, err := o.Call(o.IDA(weightFn)); err != nil {
			t.Fatalf("等級 %d／武器 %d：%v", c.level, c.weapon, err)
		}
		gotEnc := int(int16(o.Word(o.IDA(weightOut))))
		gotSpeed := int(int16(o.Word(o.IDA(speedOut))))
		if uint16(gotEnc) == 0xDEAD || uint16(gotSpeed) == 0xDEAD {
			t.Fatalf("等級 %d／武器 %d：兩格沒被寫", c.level, c.weapon)
		}
		armour := c.armour[0] + c.armour[1] + c.armour[2]
		wantEnc, wantSpeed := battle.Encumbrance(
			c.level, battle.Weapon{Weight: c.weapon}, armour, 0)
		if gotEnc != wantEnc || gotSpeed != wantSpeed {
			t.Errorf("等級 %d／武器 %d／護甲 %v：原版 (負重 %d, 速度 %d)，remake (%d, %d)",
				c.level, c.weapon, c.armour, gotEnc, gotSpeed, wantEnc, wantSpeed)
		}
	}
}

// 溢位邊界：原版那一步是**有號** 16 位元乘法。
//
// `係數 × 總重` 超過 32767 就變負數，負重跟著變負。等級 1 的係數是 70，所以門檻是
// 總重 468。**實際資料到不了**——怪物護甲重量最大 267、武器負重最大 100（折半算 50），
// 加起來 317。但這個邊界**比技能那一格近得多**（那一格要 6554 × 10），所以記下來。
func TestEncumbranceOverflowBoundary(t *testing.T) {
	o := prepare(t)
	const (
		slot       = 0
		charBase   = 0xBB88
		charWeapon = 0x3A
		charShield = 0x160
	)
	charArmour := [3]uint16{0x10A, 0x11A, 0x12A}
	rec := uint16(actorBase + actorStride*slot)
	o.WriteU8(o.DS(rec+sideFlag), 0)
	o.WriteU8(o.DS(rec+actorLevel), 1) // 係數 70
	for k := uint16(0); k < 16; k++ {
		o.WriteU8(o.DS(charBase+charShield+k), 0)
	}
	o.WriteU8(o.DS(charBase+charWeapon), 0)

	for _, armour := range [][3]int{
		{90, 90, 87},    // 實際資料的上限：總重 267
		{156, 156, 156}, // 總重 468，剛好不溢位
		{157, 156, 156}, // 總重 469，剛好溢位
		{255, 255, 255}, // 遠超過
	} {
		for k, w := range armour {
			o.WriteU8(o.DS(charBase+charArmour[k]), uint8(w))
		}
		if _, err := o.Call(o.IDA(weightFn)); err != nil {
			t.Fatalf("護甲 %v：%v", armour, err)
		}
		got := int(int16(o.Word(o.IDA(weightOut))))
		total := armour[0] + armour[1] + armour[2]
		want, _ := battle.Encumbrance(1, battle.Weapon{}, total, 0)
		if got != want {
			t.Logf("護甲加總 %d：原版 %d，直接算 %d（有號 16 位元溢位）",
				total, got, want)
			continue
		}
		t.Logf("護甲加總 %d：兩邊都是 %d", total, got)
	}
}

// 探測：什麼情況下盾牌不算數。
//
// `sub_24FD3` 的盾牌那一段有條件：**盾牌存在、兩隻手臂都還在、而且武器不是雙手**
// （RE-080）。remake 的 `ShieldGrade` 是建戰鬥時就定下的常數，**沒有**這些條件
// ——所以手臂被打斷之後 remake 的盾牌還在算，原版可能不算。
//
// 先確認原版真的讀得到盾牌，再逐格找出「哪一格廢掉就沒有盾牌」。
func TestProbeShieldConditions(t *testing.T) {
	o := prepare(t)
	const (
		slot       = 0
		charBase   = 0xBB88
		charWeapon = 0x30
		charShield = 0x160
		itemKind   = 0x0E // 種類 byte，盾牌等級取它的低四位
	)
	rec := uint16(actorBase + actorStride*slot)
	reset := func() {
		o.WriteU8(o.DS(rec+sideFlag), 0)
		o.WriteU8(o.DS(rec+actorLevel), 1)
		// 部位欄位 `+0x11 + 2×部位` 全部設成健康（小於 0x80）。
		for k := uint16(0x11); k < 0x1D; k++ {
			o.WriteU8(o.DS(rec+k), 1)
		}
		// 盾牌：第一個 byte 非 0 表示有這件東西，等級取 `+0x0E` 的低四位。
		o.WriteU8(o.DS(charBase+charShield), 1)
		o.WriteU8(o.DS(charBase+charShield+itemKind), 3)
		o.WriteU8(o.DS(charBase+charShield+itemWeight), 20)
		// 手上一件普通武器。
		o.WriteU8(o.DS(charBase+charWeapon), 1)
		o.WriteU8(o.DS(charBase+charWeapon+itemWeight), 40)
	}
	reset()
	o.WriteU16(o.IDA(shieldGradeOut), 0xDEAD)
	if _, err := o.Call(o.IDA(weightFn)); err != nil {
		t.Fatal(err)
	}
	base := o.Word(o.IDA(shieldGradeOut))
	t.Logf("有盾牌、四肢健全時 `word_34A0C` ＝ %04X", base)
	if base == 0xDEAD {
		t.Fatal("那一格沒被寫——盾牌那一段沒走到")
	}
	if base == 0 {
		t.Skip("盾牌等級是 0——這一組輸入還沒讓它算出東西，條件要再找")
	}

	var breaks []string
	for k := uint16(0); k < uint16(actorStride); k++ {
		reset()
		o.WriteU8(o.DS(rec+k), 0xFF)
		o.WriteU16(o.IDA(shieldGradeOut), 0)
		if _, err := o.Call(o.IDA(weightFn)); err != nil {
			t.Fatalf("第 %#x 格：%v", k, err)
		}
		if o.Word(o.IDA(shieldGradeOut)) != base {
			breaks = append(breaks, "+0x"+hex4(k))
		}
	}
	t.Logf("參戰者記錄裡把哪一格設成 0xFF 會讓盾牌等級變掉：%v", breaks)

	// 另一個條件是「武器不是雙手」。逐格問武器格（16 byte）：哪一格改了會讓
	// 盾牌不算——那一格就是「雙手」的旗標所在。
	var twoHanded []string
	for k := uint16(0); k < 16; k++ {
		reset()
		o.WriteU8(o.DS(charBase+charWeapon+k), 0xFF)
		o.WriteU16(o.IDA(shieldGradeOut), 0)
		if _, err := o.Call(o.IDA(weightFn)); err != nil {
			t.Fatalf("武器格第 %#x 格：%v", k, err)
		}
		if o.Word(o.IDA(shieldGradeOut)) != base {
			twoHanded = append(twoHanded, "+0x"+hex4(k))
		}
	}
	t.Logf("武器格裡把哪一格設成 0xFF 會讓盾牌等級變掉：%v", twoHanded)
}

// 隊員的攻防技能是**依手上武器的種類**選一列（`sub_26AC2`，RE-080）。
//
// 角色記錄 `+0x52` 起四筆、一筆 13 byte：每筆開頭是類別，`+0x01` 起兩組大端
// 16-bit 是攻防。比對的是手上武器的種類 byte（`+0x3D`，也就是武器格 `+0x30`
// 的 `+0x0D`）。
//
// **remake 在建戰鬥時選一次就不再換**（`cmd/kol` 的 `battleMember`），所以撿起
// 不同種類的武器之後技能還是舊那一列。這一條先問原版：選列的規則是什麼、
// 種類不在四列裡時算多少。
func TestPartySkillRowFollowsWeaponClass(t *testing.T) {
	o := prepare(t)
	const (
		slot      = 0
		charBase  = 0xBB88
		charKind  = 0x3D // 手上武器的種類 byte
		rowBase   = 0x52
		rowStride = 13
	)
	rec := uint16(actorBase + actorStride*slot)
	o.WriteU8(o.DS(rec+sideFlag), 0) // 隊員
	o.WriteU8(o.DS(rec+plannedFlag), 0)
	o.WriteU8(o.DS(rec+penaltyOff), 0)
	o.WriteU8(o.DS(rec+penaltyDef), 0)
	o.WriteU16(o.IDA(bodyPtr), charBase)

	// 四列：類別 1／2／4／8，攻防各給不一樣的值好認。
	rows := []struct{ class, off, def int }{
		{1, 111, 211}, {2, 122, 222}, {4, 144, 244}, {8, 188, 288},
	}
	for k, r := range rows {
		at := uint16(charBase + rowBase + rowStride*k)
		o.WriteU8(o.DS(at), uint8(r.class))
		o.WriteU8(o.DS(at+1), uint8(r.off>>8))
		o.WriteU8(o.DS(at+2), uint8(r.off))
		o.WriteU8(o.DS(at+3), uint8(r.def>>8))
		o.WriteU8(o.DS(at+4), uint8(r.def))
	}

	for _, c := range []struct {
		kind     string
		class    int
		off, def int
	}{
		{"第 0 列", 1, 111, 211},
		{"第 1 列", 2, 122, 222},
		{"第 2 列", 4, 144, 244},
		{"第 3 列", 8, 188, 288},
	} {
		o.WriteU8(o.DS(charBase+charKind), uint8(c.class))
		o.WriteU16(o.IDA(offSkill), 0xDEAD)
		o.WriteU16(o.IDA(defSkill), 0xDEAD)
		if _, err := o.Call(o.IDA(skillFn), slot); err != nil {
			t.Fatalf("%s：%v", c.kind, err)
		}
		gotOff := int(o.Word(o.IDA(offSkill)))
		gotDef := int(o.Word(o.IDA(defSkill)))
		if gotOff != c.off || gotDef != c.def {
			t.Errorf("種類 %d（%s）：原版 (攻 %d, 防 %d)，那一列是 (%d, %d)",
				c.class, c.kind, gotOff, gotDef, c.off, c.def)
		}
	}

	// 種類不在四列裡：問原版算多少，不預設答案。
	o.WriteU8(o.DS(charBase+charKind), 0x40)
	o.WriteU16(o.IDA(offSkill), 0xDEAD)
	o.WriteU16(o.IDA(defSkill), 0xDEAD)
	if _, err := o.Call(o.IDA(skillFn), slot); err != nil {
		t.Fatal(err)
	}
	t.Logf("種類不在四列裡時：攻 %d、防 %d",
		o.Word(o.IDA(offSkill)), o.Word(o.IDA(defSkill)))
}

// 探測：`sub_22084`（扣裝備耐久）的參數與效果。
//
// RESEARCH-LOG 只記了「傷害超過上限就扣裝備耐久，扣光了那件裝備整個清成 0」，
// 沒記參數與扣多少。讀它的機器碼比再開一次 IDA 便宜——同一個 `.i64` 連續開關
// 會把資料庫弄壞，這一輪已經開過兩次。
func TestProbeDurabilityRoutine(t *testing.T) {
	o := prepare(t)
	t.Logf("`sub_22084` 起 0x60 個 byte：%x", o.Bytes(o.IDA(0x22084), 0x60))

	// 解出來的骨架：arg0 是物品格的位移（相對 `[ds:BB78]` 指的記錄），
	// 扣量是 `(記錄[+0x0F] & 0x20) × arg1`，耐久在 `+0x0B`，扣到負數整格清 0。
	// 用實驗確認比繼續解碼可靠。
	const (
		charBase = 0xBB88
		recPtr   = 0xBB78 // 指向當前記錄的指標
		slotOff  = 0x30   // 手上武器那一格
		durab    = 0x0B
		flags    = 0x0F
	)
	o.WriteU16(o.DS(recPtr), charBase)
	for _, c := range []struct {
		durability, flag, arg int
	}{
		{200, 0x20, 1}, // bit5 設：扣 32
		{200, 0x20, 2}, // 扣 64
		{200, 0x00, 5}, // bit5 沒設：不扣
		{40, 0x20, 1},  // 扣到負數
		{32, 0x20, 1},  // 剛好扣到 0
		{20, 0x20, 1},  // 扣成負的——整格應該被清掉
		{200, 0x20, 7}, // 一次扣爆
	} {
		o.WriteU8(o.DS(charBase+slotOff), 1) // 這一格有東西
		o.WriteU8(o.DS(charBase+slotOff+durab), uint8(c.durability))
		o.WriteU8(o.DS(charBase+slotOff+flags), uint8(c.flag))
		if _, err := o.Call(o.IDA(0x22084), slotOff, uint16(c.arg)); err != nil {
			t.Fatalf("耐久 %d／旗標 %#02x／arg %d：%v",
				c.durability, c.flag, c.arg, err)
		}
		t.Logf("耐久 %3d、旗標 %#02x、arg %d → 耐久 %3d、第一個 byte %d",
			c.durability, c.flag, c.arg,
			o.Byte(o.DS(charBase+slotOff+durab)),
			o.Byte(o.DS(charBase+slotOff)))
	}
}

// recordingRand 依序吐出給定的骰值，並記下每一次的 `n`。
//
// 對拍的是**擲點的參數序列**：誰在什麼時候抽、抽的上限是多少。骰值本身由呼叫端
// 決定，好把兩邊逼進同一條分支。
type recordingRand struct {
	values []int
	ns     []int
	i      int
}

func (r *recordingRand) Roll(n int) int {
	r.ns = append(r.ns, n)
	v := 0
	if r.i < len(r.values) {
		v = r.values[r.i]
	}
	r.i++
	return v
}

// ambushFixture 造一份怪物表：`hit` 指定哪一隻符合當下的地形與季節，
// 傳 -1 表示一隻都不符合。
func ambushFixture(hit int) *battle.Monsters {
	m := &battle.Monsters{Chance: []battle.SeasonChance{{
		Season: "WINTER",
		Day:    []int{60, 60, 60, 60, 60, 60, 60, 60},
		Night:  []int{60, 60, 60, 60, 60, 60, 60, 60},
	}}}
	m.List = make([]battle.Monster, battle.MonsterCount)
	for i := range m.List {
		m.List[i] = battle.Monster{Index: i, Name: "NOBODY"}
	}
	if hit >= 0 && hit < battle.MonsterCount {
		m.List[hit].TerrainMask = battle.TerrainBit(0)
		m.List[hit].SeasonMask = battle.SeasonBit(0)
	}
	return m
}

// 野外遭遇：**設好狀態之後兩邊算出同一個結果**（RE-078／RE-098）。
//
// # 對拍的是機制，不是亂數
//
// 骰值一律釘死，只是為了讓亂數不介入——比的是「同樣的輸入算出同樣的怪物、
// 等級、隻數與站位組」。隊伍平均位階也釘死（`sub_2A337`），兩邊因此吃同一個輸入。
//
// # 原版那一側的三個前置
//
//  1. **資料要先載入。** `_main` 剛進來時 `payload3` 還是空的，機率算出來一律 0，
//     `rand(100) >= 0` 恆真——不論骰值釘成什麼都回「沒有遭遇」，看起來像原版
//     根本沒有那條分支。先跑到城鎮地圖（`townMap`）表就在記憶體裡了。
//  2. **`word_2FA2C = 20`**（隊裡有人沒坐騎）跳過「Attempt to ride away?」那一問。
//     原版自己就有這條路，不是繞過它的邏輯。
//  3. **`Call` 等不到返回是正常的**：算完之後原版接著畫戰場、停在 `sub_1B100`
//     等滑鼠。結果那時已經寫進全域了。
//
// 結果在四個全域（RE-078 的「開打前設好的其他全域」）：
//
//	word_3B262  挑中的怪物編號（治療的遭遇是 0xFF）
//	word_3B26A  等級
//	word_30C6C  敵人數（**已經加一**）
//	word_37CF6  站位組 0…4
func TestAmbushOutcomeMatchesOriginal(t *testing.T) {
	o := townMap(t)
	o.SyncDGroupFromDS()
	const (
		chanceEA = 0x2a554 // rand(100) 遭遇機率
		healerEA = 0x2a584 // rand(100) 治療者門檻
		rankEA   = 0x2A337 // 隊伍平均位階
		paceAddr = 0x2FA2C // 隊伍最慢移動率，20 ＝ 有人沒坐騎
		noMount  = 20

		monsterAddr = 0x3B262
		levelAddr   = 0x3B26A
		countAddr   = 0x30C6C
		groupAddr   = 0x37CF6
	)
	o.WriteU16(o.IDA(paceAddr), noMount)

	for _, c := range []struct {
		name string
		rank int    // 隊伍平均位階（釘死 sub_2A337）
		pick uint32 // rand(0x20)：怪物編號 ＋ 1
		band uint32 // rand(0x63)：級距的原值
		grp  uint32 // rand(5)：站位組 ＋ 1
	}{
		{"位階 3、編號 0、級距 0、組 0", 3, 1, 1, 1},
		{"位階 3、編號 0、級距 1、組 2", 3, 1, 0x14, 3},
		{"位階 5、編號 0、級距 2、組 4", 5, 1, 0x28, 5},
		{"位階 1、編號 0、級距 0、組 1", 1, 1, 1, 2},
		{"位階 8、編號 0、級距 4、組 3", 8, 1, 0x62, 4},
	} {
		t.Run(c.name, func(t *testing.T) {
			// 哨兵：跑完沒被寫就看得出來，不會把上一輪的殘留當結果。
			for _, a := range []uint32{monsterAddr, levelAddr, countAddr, groupAddr} {
				o.WriteU16(o.IDA(a), 0xDEAD)
			}
			o.StubValue(o.IDA(rankEA), uint32(c.rank))
			o.Stub(o.IDA(rollEnter), func(p *oracle.Oracle) uint32 {
				switch ea, n := p.ToIDA(p.Caller()), p.Arg(0); {
				case ea == chanceEA:
					return 0 // 一定遭遇
				case ea == healerEA:
					return 99 // 不走治療者
				case n == 0x20:
					return c.pick
				case n == 0x63:
					return c.band
				case n == 12:
					return 1 // 隻數：0 一定小於上限，不重擲
				case n == 5:
					return c.grp
				}
				return 1
			})
			if _, err := o.CallBudget(50_000_000, o.IDA(ambushEnter)); err != nil {
				t.Logf("沒有返回（算完之後接著進戰鬥畫面）")
			}
			o.Stub(o.IDA(rollEnter), nil)
			o.Stub(o.IDA(rankEA), nil)

			got := struct{ monster, level, count, group int }{
				int(o.Word(o.IDA(monsterAddr))), int(o.Word(o.IDA(levelAddr))),
				int(o.Word(o.IDA(countAddr))), int(o.Word(o.IDA(groupAddr))),
			}
			if got.monster == 0xDEAD {
				t.Fatalf("四個全域一個都沒被寫（%+v）——原版沒走進遭遇那一條", got)
			}

			// remake：同樣的輸入。`ambushFixture(0)` 讓編號 0 符合當下的地形與
			// 季節——原版在這個狀態下也是編號 0 符合、編號 1 不符合。
			r := &recordingRand{values: []int{
				0,               // 機率
				99,              // 治療者門檻
				int(c.pick),     // 挑怪物
				1,               // 場景
				int(c.band),     // 級距
				1,               // 隻數
				int(c.grp),      // 站位組
			}}
			a := ambushFixture(0).RollAmbush(0, 0, false, c.rank, r)
			if a == nil {
				t.Fatal("remake 這一組沒有遭遇，走的不是同一條路")
			}
			a.RollForce(c.rank, 1, r)

			for _, f := range []struct {
				name           string
				original, mine int
			}{
				{"怪物編號", got.monster, a.Monster},
				{"等級", got.level, a.Level},
				{"敵人數", got.count, a.Count},
				{"站位組", got.group, a.Group},
			} {
				if f.original != f.mine {
					t.Errorf("%s：原版 %d，remake %d", f.name, f.original, f.mine)
				}
			}
			t.Logf("怪物 %d、等級 %d、%d 隻、站位組 %d", got.monster, got.level,
				got.count, got.group)
		})
	}
}

// 天候的隨機漫步：設好狀態之後兩邊算出同一組（RE-077）。
//
// 天候在 `word_2DF5A`（IDA `0x2DF5A`，DGROUP+0x8A），只有**時段變換**時才擲
// （`sub_29987:0x29A83`）。時段是「第一個滿足 `時刻 >= 下限`」的那一段，
// 目前時段記在 `word_2DF42`——把它設成與時刻算出來的不一樣，變換就成立。
//
//	roll = rand(0x80) − 1                       ; 0…127
//	上限 = 表[月] & 7、下限 = 表[月] / 8 & 7      ; DGROUP+0x1326
//	roll < 表2[月]（DGROUP+0x131A）: 沒到上限就 +1
//	否則 roll < 100:                沒到下限就 −1
//	表[月] > 0x7F 且天候 == 3:       再 +1        ; 冬天才有的第 4 組
//
// RE-077 原本把這一段標成 `強推論`（式子逐指令讀出，沒有實跑）。
func TestWeatherMatchesOriginal(t *testing.T) {
	o := townMap(t)
	o.SyncDGroupFromDS()
	const (
		stepEnter   = 0x29987
		paceAddr    = 0x2FA2C
		noMount     = 20
		clockAddr   = 0x2DF40
		partAddr    = 0x2DF42
		monthAddr   = 0x2FBB4
		weatherAddr = 0x2DF5A
	)
	o.WriteU16(o.IDA(paceAddr), noMount)

	for _, c := range []struct {
		name    string
		month   int    // 0 是一月
		weather int    // 進去的天候
		roll    uint32 // rand(0x80) 的回傳，程式自己減一
	}{
		{"四月、晴、小骰：往壞一格", 3, 1, 1},
		{"四月、已到上限：不動", 3, 3, 1},
		{"四月、中骰：往好一格", 3, 2, 50},
		{"四月、已到下限：不動", 3, 0, 50},
		{"四月、大骰：完全不動", 3, 2, 120},
		{"一月、小骰推到第 3 段：再多一段（下雪）", 0, 2, 1},
		{"一月、已到下限 1：不動", 0, 1, 50},
		{"八月、上限 1、小骰", 7, 0, 1},
		{"八月、門檻 15，骰 20 往好", 7, 1, 20},
	} {
		t.Run(c.name, func(t *testing.T) {
			o.WriteU16(o.IDA(monthAddr), uint16(c.month))
			o.WriteU16(o.IDA(weatherAddr), uint16(c.weather))
			o.WriteU16(o.IDA(clockAddr), 2000)
			o.WriteU16(o.IDA(partAddr), 0) // 與時刻算出來的 2 不同 → 一定變換
			o.Stub(o.IDA(rollEnter), func(p *oracle.Oracle) uint32 {
				if p.Arg(0) == 0x80 {
					return c.roll
				}
				return 99 // 其餘一律大值：不遭遇
			})
			if _, err := o.CallBudget(50_000_000, o.IDA(stepEnter)); err != nil {
				t.Logf("沒有返回（算完之後接著畫畫面）")
			}
			o.Stub(o.IDA(rollEnter), nil)

			got := int(o.Word(o.IDA(weatherAddr)))
			clock := world.Clock{Month: c.month, Weather: c.weather}
			want := clock.NextWeather(int(c.roll) - 1)
			if got != want {
				t.Fatalf("天候 %d 擲 %d：原版走到 %d，remake 走到 %d",
					c.weather, c.roll, got, want)
			}
			t.Logf("%d → %d", c.weather, got)
		})
	}
}

// 攻擊結算（`sub_26D76`）的入口與出口。
//
// # 入口
//
// 開頭的碼（從記憶體讀出來解的）：
//
//	b8 32 00      mov ax, 50            ; actorStride
//	f7 2e 5a 9e   imul word [ds:9E5A]   ; ★ 當前行動者索引
//	b9 12 67      mov cx, 0x6712        ; actorBase
//	8a 47 28      mov al, [bx+0x28]
//	83 7e 06 00   cmp word [bp+6], 0    ; ★ arg0：0 走攻擊那一條
//
// 當前行動者索引在 `ds:9E5A`（`× 50 + 0x6712` 就是參戰者記錄）。**`arg0 == 0`
// 才走攻擊**——`arg0 != 0` 只跑 11.7 萬道就返回，一支子常式都沒呼叫。
//
// 最小狀態：兩名參戰者的「力量 ＋ 健康」（`ds:1C50 + 2i`）不能是 0，
// 否則 `sub_23683` 直接 `R6003 integer divide by 0` 讓原版離開；再加等級
// （`+0x02`）、敵我旗標（`+0x01`）與這一回合的招架次數（`+0x0F`，防禦值的除數）。
//
// 跟遭遇那一支一樣，算完之後接著畫訊息，所以 `Call` 等不到返回——**那不是失敗**。
//
// # 出口
//
// 單跑一次的差分有七百多個 word，絕大多數是畫面緩衝（`ds:1AAA` 起那一大片）。
// **兩組只差一個輸入的執行，差異就只剩結果**：基礎攻擊 70 → 120（加 50）之後
// 只有三個 word 不同，其中 `ds:671A` 是輸入本身（參戰者記錄 `+0x08`），
// 另外兩個各加 5——正是 RE-133 那條「攻擊值 ÷ 10」。
func TestProbeAttackResolveResult(t *testing.T) {
	const (
		resolveEA   = 0x26D76
		currentSlot = 0x9E5A
		dgBase      = 0x2DED0
		span        = 0x8000
	)
	snapshot := func(t *testing.T, attack uint8) []byte {
		t.Helper()
		o := townMap(t)
		o.SyncDGroupFromDS()
		o.WriteU16(o.DS(currentSlot), 0)
		for i := uint16(0); i < 2; i++ {
			o.WriteU16(o.DS(uint16(actorStamina)+2*i), 30)
			rec := uint16(actorBase) + uint16(actorStride)*i
			o.WriteU16(o.DS(rec+0x02), 1)          // 等級
			o.WriteU8(o.DS(rec+0x01), uint8(i))    // 敵我旗標
			o.WriteU8(o.DS(rec+0x08), 70)          // 基礎攻擊
			o.WriteU8(o.DS(rec+0x09), 40)          // 基礎防禦
			o.WriteU8(o.DS(rec+0x0F), 1)           // 這一回合招架過幾次
		}
		// 只有攻擊者的基礎攻擊不同。
		o.WriteU8(o.DS(uint16(actorBase)+0x08), attack)
		o.Stub(o.IDA(rollEnter), func(p *oracle.Oracle) uint32 { return 50 })
		_, _ = o.CallBudget(20_000_000, o.IDA(resolveEA), 0)
		o.Stub(o.IDA(rollEnter), nil)
		return append([]byte(nil), o.Bytes(o.IDA(dgBase), span)...)
	}
	lo := snapshot(t, 70)
	hi := snapshot(t, 120)

	changed := 0
	for i := 0; i+1 < span; i += 2 {
		a := uint16(lo[i]) | uint16(lo[i+1])<<8
		b := uint16(hi[i]) | uint16(hi[i+1])<<8
		if a == b {
			continue
		}
		changed++
		if changed <= 40 {
			t.Logf("  ds:%04X（IDA %#x）：攻擊 70 → %04X、攻擊 120 → %04X",
				i, dgBase+i, a, b)
		}
	}
	t.Logf("兩組共 %d 個 word 不同", changed)
}

// 探測：那兩個結果各是什麼，以及**對手還沒進到算式裡**。
//
// `ds:1C76` 只跟攻擊者的 `+0x08` 走（加 50 → 加 5，也就是 ÷10），
// `ds:44AC` 只跟攻擊者的 `+0x09` 走（加 50 → 加 50，不除）。兩個都是
// **攻擊者這一側**的數字。
//
// 對手那一側目前一格都沒進來：改防禦者的 `+0x08`／`+0x09` 沒有反應，
// 把整筆參戰者 1 或 2 的記錄塗成 0xFF 也一樣——**目標不在參戰者記錄裡**。
// `+0x23`／`+0x25` 都沒反應，`+0x24` 會讓兩個值一起變（`ds:44AC` 成了有號的
// −219），但那比較像是它換了查表的選項，不是指定目標。
//
// 下一步要開一支 IDA 探針把 `sub_26D76` 一次問完：它讀哪些全域、指令碼怎麼分派。
// 實跑到這裡已經問不出來了。
func TestProbeAttackScoreMeaning(t *testing.T) {
	const (
		resolveEA   = 0x26D76
		currentSlot = 0x9E5A
		targetField = 0x24
		scoreA      = 0x2FB46 // ds:1C76
		scoreB      = 0x3237C // ds:44AC
	)
	run := func(t *testing.T, atkBase, atkGuard, defBase, defGuard uint8) (int16, int16) {
		t.Helper()
		o := townMap(t)
		o.SyncDGroupFromDS()
		o.WriteU16(o.DS(currentSlot), 0)
		for i := uint16(0); i < 2; i++ {
			o.WriteU16(o.DS(uint16(actorStamina)+2*i), 30)
			rec := uint16(actorBase) + uint16(actorStride)*i
			o.WriteU16(o.DS(rec+0x02), 1)
			o.WriteU8(o.DS(rec+0x01), uint8(i))
			o.WriteU8(o.DS(rec+0x0F), 1)
		}
		atk := uint16(actorBase)
		def := uint16(actorBase) + uint16(actorStride)
		o.WriteU8(o.DS(atk+0x08), atkBase)
		o.WriteU8(o.DS(atk+0x09), atkGuard)
		o.WriteU8(o.DS(def+0x08), defBase)
		o.WriteU8(o.DS(def+0x09), defGuard)
		o.WriteU8(o.DS(atk+targetField), 1) // 目標 ＝ 參戰者 1
		o.Stub(o.IDA(rollEnter), func(p *oracle.Oracle) uint32 { return 50 })
		_, _ = o.CallBudget(20_000_000, o.IDA(resolveEA), 0)
		o.Stub(o.IDA(rollEnter), nil)
		return int16(o.Word(o.IDA(scoreA))), int16(o.Word(o.IDA(scoreB)))
	}
	for _, c := range []struct {
		name                               string
		atkBase, atkGuard, defBase, defGuard uint8
	}{
		{"基準", 70, 40, 70, 40},
		{"攻擊者 +0x08 加 50", 120, 40, 70, 40},
		{"攻擊者 +0x09 加 50", 70, 90, 70, 40},
		{"防禦者 +0x08 加 50", 70, 40, 120, 40},
		{"防禦者 +0x09 加 50", 70, 40, 70, 90},
	} {
		a, b := run(t, c.atkBase, c.atkGuard, c.defBase, c.defGuard)
		t.Logf("%-20s → ds:1C76=%d、ds:44AC=%d", c.name, a, b)
	}
}

// 攻擊分數的骰前算式：設好狀態之後兩邊算出同一個數（RE-133 的 `sub_26D76`）。
//
// 原版那一段（`0x26E08`…`0x26EA0`，反組譯逐條讀出來的）：
//
//	sub_26AC2(當前行動者)                          ; 武器攻防技能 → word_34A12
//	sub_23683(當前行動者)                          ; 疲勞比例 → word_2FB4A
//	word_3237C = 記錄[+0x08] + word_34A12          ; 力量 ＋ 武器攻擊技能
//	var_10 = 記錄[+0x2D] × 20                       ; 這一輪挨的傷
//	var_12 = 記錄[+0x0E] × 10                       ; 累積傷害
//	word_3237C −= word_3237C × word_2FB4A ÷ 200 + var_12 + var_10
//	sub_24FD3()                                     ; 負重 → word_30C88
//	word_3237C −= word_30C88
//	word_3237C = word_3237C ÷ 10 + rand(100)        ; ★ 攻擊分數
//
// **三支子常式都要釘住**（`Stub` 讓它們不執行），技能、疲勞比例與負重就保持
// 我們寫進去的值——這樣驗的是 `sub_26D76` 自己那段算式，不是那三支（它們在
// RE-133 已經各自對過）。
//
// ⚠ 漏掉 `sub_26AC2` 的症狀很好認：`word_34A12` 被重算成 0，於是**每一組的差
// 都剛好等於「技能 ÷ 10」**——看起來像 remake 多加了一項，實際上是原版那一側
// 的技能被吃掉了。
func TestAttackScoreMatchesOriginal(t *testing.T) {
	const (
		resolveEA   = 0x26D76
		currentSlot = 0x9E5A // word_37D2A
		attackOut   = 0x3237C // word_3237C：攻擊值 → 攻擊分數
		fatigueOut  = 0x2FB4A // word_2FB4A：疲勞比例
		encumbOut   = 0x30C88 // word_30C88：負重
		skillOut    = 0x34A12 // word_34A12：武器攻擊技能
		fatigueFn   = 0x23683
		encumbFn    = 0x24FD3
		skillFn     = 0x26AC2
		roll        = 50
	)
	for _, c := range []struct {
		name                                     string
		strength, skill, ratio, damage, hurt, enc int
	}{
		{"乾淨的一擊", 70, 30, 0, 0, 0, 0},
		{"帶疲勞比例", 70, 30, 40, 0, 0, 0},
		{"帶累積傷害", 70, 30, 0, 3, 0, 0},
		{"帶這一輪的傷", 70, 30, 0, 0, 2, 0},
		{"帶負重", 70, 30, 0, 0, 0, 25},
		{"四項一起", 90, 45, 60, 4, 3, 30},
		{"高技能", 120, 200, 20, 1, 1, 10},
	} {
		t.Run(c.name, func(t *testing.T) {
			o := townMap(t)
			o.SyncDGroupFromDS()
			o.WriteU16(o.DS(currentSlot), 0)
			rec := uint16(actorBase)
			o.WriteU8(o.DS(rec+0x08), uint8(c.strength))
			o.WriteU8(o.DS(rec+0x0E), uint8(c.damage))
			o.WriteU8(o.DS(rec+0x2D), uint8(c.hurt))
			o.WriteU8(o.DS(rec+0x01), 0) // 隊員
			o.WriteU16(o.DS(rec+0x02), 1)
			o.WriteU8(o.DS(rec+0x0F), 1)
			o.WriteU16(o.IDA(skillOut), uint16(c.skill))
			o.WriteU16(o.IDA(fatigueOut), uint16(c.ratio))
			o.WriteU16(o.IDA(encumbOut), uint16(c.enc))
			// 釘住那兩支，讓疲勞比例與負重保持上面寫進去的值。
			o.StubValue(o.IDA(fatigueFn), 0)
			o.StubValue(o.IDA(encumbFn), 0)
			o.StubValue(o.IDA(skillFn), 0)
			o.Stub(o.IDA(rollEnter), func(p *oracle.Oracle) uint32 { return roll })
			_, _ = o.CallBudget(20_000_000, o.IDA(resolveEA), 0)
			o.Stub(o.IDA(rollEnter), nil)
			o.Stub(o.IDA(fatigueFn), nil)
			o.Stub(o.IDA(encumbFn), nil)
			o.Stub(o.IDA(skillFn), nil)
			got := int(int16(o.Word(o.IDA(attackOut))))

			// remake：同一組輸入。`SkillPenaltyOff` 設 0，`OffenceSkill()` 就是
			// `Attack` 本身；`Fatigue`／`Stamina` 湊出同一個疲勞比例。
			a := battle.Actor{
				Strength: c.strength, Attack: c.skill,
				Damage: c.damage, Hurt: c.hurt, Encumbrance: c.enc,
				Fatigue: c.ratio, Stamina: 50, // 50 × ratio ÷ 50 ＝ ratio
			}
			if a.FatigueRatio() != c.ratio {
				t.Fatalf("remake 的疲勞比例是 %d，要湊成 %d", a.FatigueRatio(), c.ratio)
			}
			want := a.AttackValue()/10 + roll
			if got != want {
				t.Fatalf("原版 %d，remake %d（力量 %d 技能 %d 比例 %d 傷害 %d 這輪 %d 負重 %d）",
					got, want, c.strength, c.skill, c.ratio, c.damage, c.hurt, c.enc)
			}
			t.Logf("兩邊都是 %d", got)
		})
	}
}

// 防禦分數的骰前算式：設好狀態之後兩邊算出同一個數（`sub_26D76:0x27204`…）。
//
// **防禦者索引在 `word_3237E`（`ds:44AE`）**——先前改防禦者的記錄毫無反應，
// 就是因為那一格是 0，「對手」指到攻擊者自己。原版那一段：
//
//	記錄[+0x24] == 0 → word_2F942 = 0，整段跳過   ; 防禦者沒有擺出防禦
//	記錄[+0x0F] += 1                               ; ★ 招架次數先加一
//	sub_26AC2(防禦者)                              ; 武器防禦技能 → word_32380
//	word_3237C = word_32380 ÷ 記錄[+0x0F] + 記錄[+0x09]
//	sub_23683(防禦者)                              ; 疲勞比例 → word_2FB4A
//	var_A = word_3237C × word_2FB4A ÷ 200
//	word_37D2A ← word_3237E                        ; 暫時把「當前」換成防禦者
//
// ⚠ **招架次數是先加再除。** remake 的 `DefenceValue` 用的是傳進來的 `Parries`
// （小於 1 就當 1），所以原版的「這一次也算進去」要由呼叫端反映——這一支順便
// 把兩種語意都試出來。
func TestDefenceScoreMatchesOriginal(t *testing.T) {
	const (
		resolveEA   = 0x26D76
		currentSlot = 0x9E5A // word_37D2A：攻擊者
		defenderEA  = 0x3237E // word_3237E：防禦者索引
		// word_3237C 是中間值（防禦值）；**防禦分數是 `word_2F942`**
		// （`0x2737F`），下一行就是命中判定：`word_2FB46 >= word_2F942`。
		scoreOut    = 0x2F942
		fatigueOut  = 0x2FB4A
		encumbOut   = 0x30C88
		defSkillOut = 0x32380 // word_32380：武器防禦技能
		shieldOut   = 0x34A0C // word_34A0C：盾牌等級
		fatigueFn   = 0x23683
		encumbFn    = 0x24FD3
		skillFn     = 0x26AC2
		roll        = 50
	)
	for _, c := range []struct {
		name                                       string
		skill, quick, parries, ratio, damage, enc  int
	}{
		{"沒擋過", 60, 40, 0, 0, 0, 0},
		{"擋過一次", 60, 40, 1, 0, 0, 0},
		{"擋過三次", 60, 40, 3, 0, 0, 0},
		{"帶疲勞比例", 60, 40, 0, 40, 0, 0},
		{"帶累積傷害", 60, 40, 0, 0, 3, 0},
		{"帶負重", 60, 40, 0, 0, 0, 25},
		{"四項一起", 120, 55, 2, 60, 4, 30},
	} {
		t.Run(c.name, func(t *testing.T) {
			o := townMap(t)
			o.SyncDGroupFromDS()
			o.WriteU16(o.DS(currentSlot), 0) // 攻擊者 ＝ 參戰者 0
			o.WriteU16(o.IDA(defenderEA), 1) // 防禦者 ＝ 參戰者 1
			atk := uint16(actorBase)
			def := uint16(actorBase) + uint16(actorStride)
			o.WriteU8(o.DS(atk+0x01), 0)
			o.WriteU16(o.DS(atk+0x02), 1)
			o.WriteU8(o.DS(def+0x01), 1)
			o.WriteU16(o.DS(def+0x02), 1)
			o.WriteU8(o.DS(def+0x24), 1) // 非 0，否則整段跳過
			o.WriteU8(o.DS(def+0x09), uint8(c.quick))
			o.WriteU8(o.DS(def+0x0E), uint8(c.damage))
			o.WriteU8(o.DS(def+0x0F), uint8(c.parries))
			o.WriteU16(o.IDA(defSkillOut), uint16(c.skill))
			o.WriteU16(o.IDA(fatigueOut), uint16(c.ratio))
			o.WriteU16(o.IDA(encumbOut), uint16(c.enc))
			// ⚠ 沒有盾牌是 **`>= 0x80`**（`0x272E6` 的 `cmp … 80h / jl`）。
			// 寫 0xFFFF 反而會被當成有號的 −1 而落進「有盾牌」那一條，
			// 加上 `100 × (−1) + 200`。
			o.WriteU16(o.IDA(shieldOut), 0x80)
			o.StubValue(o.IDA(fatigueFn), 0)
			o.StubValue(o.IDA(encumbFn), 0)
			o.StubValue(o.IDA(skillFn), 0)
			o.Stub(o.IDA(rollEnter), func(p *oracle.Oracle) uint32 { return roll })
			_, _ = o.CallBudget(20_000_000, o.IDA(resolveEA), 0)
			o.Stub(o.IDA(rollEnter), nil)
			for _, fn := range []uint32{fatigueFn, encumbFn, skillFn} {
				o.Stub(o.IDA(fn), nil)
			}
			got := int(int16(o.Word(o.IDA(scoreOut))))
			t.Logf("原版 %d（招架次數進去是 %d，出來是 %d）",
				got, c.parries, o.Byte(o.DS(def+0x0F)))

			// remake：原版先把招架次數加一再除，所以傳進去的要是「加一之後」。
			a := battle.Actor{
				Parry: c.skill, Quickness: c.quick, Parries: c.parries + 1,
				Damage: c.damage, Encumbrance: c.enc,
				Fatigue: c.ratio, Stamina: 50,
				ShieldGrade: battle.NoShield,
			}
			want := a.DefenceValue()/10 + roll
			if got != want {
				t.Fatalf("原版 %d，remake %d（技能 %d 敏捷 %d 招架 %d 比例 %d 傷害 %d 負重 %d）",
					got, want, c.skill, c.quick, c.parries, c.ratio, c.damage, c.enc)
			}
			t.Logf("兩邊都是 %d", got)
		})
	}
}

// 傷害鏈：設好狀態之後兩邊算出同一個數（`sub_26D76:0x274D6`…`0x27645`）。
//
// 輸入全是全域：`word_30A76` 傷害碼（`÷32+1` 是骰數、`(÷4)&7` 挑骰面）、
// `sub_26D3C` 的回傳是基礎傷害、`word_31122` 指令碼。落點是**防禦者記錄的
// `+0x2D`**（remake 的 `Actor.Hurt`），而且是 `add`，所以先清 0。
//
// ⚠ **命中之後原版先印訊息再 `sub_1B100` 等玩家點滑鼠**，傷害在那之後才算。
// 不釘住那一支，擲骰序列只有攻防那兩次——看起來像「命中了但傷害是 0」。
func TestDamageMatchesOriginal(t *testing.T) {
	const (
		resolveEA   = 0x26D76
		currentSlot = 0x9E5A
		defenderEA  = 0x3237E
		codeEA      = 0x31122
		dmgCodeEA   = 0x30A76
		baseFn      = 0x26D3C
		fatigueOut  = 0x2FB4A
		encumbOut   = 0x30C88
		shieldOut   = 0x34A0C
		atkScoreEA  = 0x2FB46
		defScoreEA  = 0x2F942
		fatigueFn   = 0x23683
		encumbFn    = 0x24FD3
		skillFn     = 0x26AC2
		durabFn     = 0x22084
		waitFn      = 0x1B100
		hurtOff     = 0x2D
	)
	for _, c := range []struct {
		name    string
		dmgCode int
		base    int
		roll    uint32
	}{
		{"1 骰 2 面、基礎 0", 0, 0, 1},
		{"1 骰 2 面、基礎 5", 0, 5, 1},
		{"2 骰 2 面", 32, 0, 1},
		{"1 骰 3 面", 4, 0, 1},
		{"1 骰 5 面、骰值 3", 8, 2, 3},
		{"3 骰 7 面、基礎 4", 64 + 12, 4, 2},
	} {
		t.Run(c.name, func(t *testing.T) {
			o := townMap(t)
			o.SyncDGroupFromDS()
			o.WriteU16(o.DS(currentSlot), 0)
			o.WriteU16(o.IDA(defenderEA), 1)
			atk := uint16(actorBase)
			def := uint16(actorBase) + uint16(actorStride)
			o.WriteU8(o.DS(atk+0x01), 0)
			o.WriteU16(o.DS(atk+0x02), 1)
			o.WriteU8(o.DS(atk+0x08), 70) // 力量 70 → 加成 0
			o.WriteU8(o.DS(def+0x01), 1)
			o.WriteU16(o.DS(def+0x02), 1)
			o.WriteU8(o.DS(def+0x24), 1)
			o.WriteU8(o.DS(def+0x09), 10)
			o.WriteU8(o.DS(def+0x0F), 0)
			o.WriteU8(o.DS(def+hurtOff), 0)
			o.WriteU16(o.IDA(codeEA), 0) // 指令碼 0：沒有 styleDamage 那一項
			o.WriteU16(o.IDA(dmgCodeEA), uint16(c.dmgCode))
			o.WriteU16(o.IDA(fatigueOut), 0)
			o.WriteU16(o.IDA(encumbOut), 0)
			o.WriteU16(o.IDA(shieldOut), 0x80)
			for _, fn := range []uint32{fatigueFn, encumbFn, skillFn, durabFn, waitFn} {
				o.StubValue(o.IDA(fn), 0)
			}
			o.StubValue(o.IDA(baseFn), uint32(c.base))
			o.Stub(o.IDA(rollEnter), func(p *oracle.Oracle) uint32 { return c.roll })
			_, _ = o.CallBudget(20_000_000, o.IDA(resolveEA), 0)
			for _, fn := range []uint32{fatigueFn, encumbFn, skillFn, durabFn, baseFn, waitFn, uint32(rollEnter)} {
				o.Stub(o.IDA(fn), nil)
			}
			margin := int(int16(o.Word(o.IDA(atkScoreEA)))) - int(int16(o.Word(o.IDA(defScoreEA))))
			got := int(o.Byte(o.DS(def + hurtOff)))

			count, sides := battle.DamageCode(c.dmgCode)
			a := battle.Actor{
				Strength: 70, Fatigue: 0, Stamina: 50,
				Weapon: battle.Weapon{Base: c.base, Count: count, Sides: sides},
			}
			r := &recordingRand{values: []int{
				int(c.roll), int(c.roll), int(c.roll), int(c.roll), int(c.roll),
			}}
			want := battle.DamageRoll(a, 0, margin, r)
			if got != want {
				t.Fatalf("原版 %d，remake %d（傷害碼 %d ＝ %d 骰 %d 面、基礎 %d、骰值 %d、餘裕 %d）",
					got, want, c.dmgCode, count, sides, c.base, c.roll, margin)
			}
			t.Logf("兩邊都是 %d（%d 骰 %d 面、基礎 %d、餘裕 %d）",
				got, count, sides, c.base, margin)
		})
	}
}

// 把一次傷害記到身上：設好狀態之後兩邊算出同一份記錄（`sub_277F4`）。
//
// 攻擊常式在 IDA 0x277E6 呼叫此函式，傳入傷害、部位及第三參數零；
// 舊註解誤以為攻擊常式沒有呼叫，訂正見 docs/re/20260908-armour-control-flow.md。
// 此處只驗扣傷結果，不代替前段護甲與耐久的整合驗證。
//
// 六格部位在防禦者記錄的 `+0x11`／`+0x13`／`+0x15`／`+0x17`／`+0x19`／`+0x1B`
// （頭、胸、右臂、左臂、右腿、左腿），累積傷害在 `+0x0E`、昏迷門檻在 `+0x10`。
//
// ⚠ `Call` 等不到返回（後面接著印訊息），**那不是失敗**——記錄在那之前已經寫好了。
func TestTakeDamageMatchesOriginal(t *testing.T) {
	const (
		takeEA     = 0x277F4
		defenderEA = 0x3237E
		damageOff  = 0x0E
		thresOff   = 0x10
	)
	parts := []uint16{0x11, 0x13, 0x15, 0x17, 0x19, 0x1B}
	for _, c := range []struct {
		name                       string
		damage, location, already  int
		threshold                  int
		health                     int
	}{
		{"打頭 5", 5, 0, 0, 100, 50},
		{"打胸 12", 12, 1, 0, 100, 50},
		{"打右腿 7", 7, 4, 0, 100, 50},
		{"那一肢剛好扣光", 50, 2, 0, 100, 50},
		{"那一肢扣過頭", 80, 3, 0, 100, 50},
		{"累積傷害到門檻", 30, 5, 75, 100, 50},
		{"累積傷害夾在 0x7F", 100, 0, 100, 250, 50},
	} {
		t.Run(c.name, func(t *testing.T) {
			o := townMap(t)
			o.SyncDGroupFromDS()
			o.WriteU16(o.IDA(defenderEA), 1)
			def := uint16(actorBase) + uint16(actorStride)
			o.WriteU8(o.DS(def+damageOff), uint8(c.already))
			o.WriteU8(o.DS(def+thresOff), uint8(c.threshold))
			for _, off := range parts {
				o.WriteU8(o.DS(def+off), uint8(c.health))
			}
			_, _ = o.CallBudget(5_000_000, o.IDA(takeEA), uint16(c.damage), uint16(c.location))
			var got [6]int
			for i, off := range parts {
				got[i] = int(o.Byte(o.DS(def + off)))
			}
			gotDamage := int(o.Byte(o.DS(def + damageOff)))

			a := battle.Actor{
				Damage: c.already, Threshold: c.threshold,
				Head: c.health, Chest: c.health,
				RightArm: c.health, LeftArm: c.health,
				RightLeg: c.health, LeftLeg: c.health,
			}
			a.Take(c.damage, c.location)
			want := [6]int{a.Head, a.Chest, a.RightArm, a.LeftArm, a.RightLeg, a.LeftLeg}
			// ⚠ **廢掉的那一肢，原版寫 −1（記錄裡是 0xFF），remake 寫 0。**
			// 兩邊所有判準都是「`<= 0` 就算廢」，所以行為一致；但記錄裡的值不同，
			// 存檔或畫面若讀那一格就會差。這裡比語意，差異登記在 `WORKLIST.md`。
			down := func(v int) bool { return v == 0xFF || v <= 0 }
			for i := range got {
				if down(got[i]) != down(want[i]) {
					t.Fatalf("第 %d 格：原版 %d，remake %d（部位 %v／%v）",
						i, got[i], want[i], got, want)
				}
				if !down(got[i]) && got[i] != want[i] {
					t.Fatalf("第 %d 格：原版 %d，remake %d", i, got[i], want[i])
				}
			}
			if gotDamage != a.Damage {
				t.Fatalf("累積傷害：原版 %d，remake %d", gotDamage, a.Damage)
			}
			t.Logf("兩邊都是 部位 %v（remake %v）、累積傷害 %d", got, want, gotDamage)
		})
	}
}

// 探測：怪物記錄 `+0x10` 的 bit0 是什麼。
//
// bit7 是「碎成一地」那一句（RE-131），bit0 只有 CYCLOPS／CLIFF TROLL／SLEDGE／
// MIST GRUB 四隻有，語意一直沒解。全庫掃「讀某個暫存器 `+10h`」只有七處，
// 其中 `0x22172` 在 `sub_2215A` 裡把整個 byte 餵給
// `sub_22376(值, 0, 1)`——依 RE-118 已解的公式 `(值 >> 0) & 1`，**那就是 bit0**。
//
// `sub_2215A` 在圖示列那一份 dump 裡（75 道指令，呼叫端 `0x20F30`／`0x22132`），
// 但那份只有摘要沒有反組譯。所以改用實跑：設好 `word_39A48` 指的記錄，

// 怪物本體記錄搬進參戰者記錄：`sub_2215A`。
//
// 這是 MIST GRUB 那一位的落點。全庫掃「讀某個暫存器 `+10h`」只有七處，其中
// `0x22172` 把整個旗標 byte 餵給 `sub_22376(值, 0, 1)`；`sub_22376` 的原文是
// `(值 >> (高位元 + 1 − 位元數)) & ~(0xFFFF << 位元數)`，代入就是 `值 & 1`。
//
// 前置的 `sub_22296` 把記錄指標算出來：`word_39A48 = 0xCD92 + ((索引 − 6) << 7)`，
// 也就是**怪物本體記錄一筆 128 bytes、從參戰者索引 6 起算**（0..5 是隊伍）。
// 這一點很重要：外面先寫好 `word_39A48` 是沒用的，進門第一件事就把它重算掉。
func TestMonsterTemplateCopyMatchesOriginal(t *testing.T) {
	const (
		initFn   = 0x2215A
		bodyBase = 0xCD92 // 怪物本體記錄表
		bodySize = 128
		flagOut  = 0x3AC58 // word_3AC58：這一場怪物的 bit0
		partBase = 0x6712  // 參戰者記錄表
		partSize = 50
		slot     = 6 // 第一個怪物槽
	)
	body := uint16(bodyBase + (slot-6)*bodySize)
	part := uint16(partBase + slot*partSize)

	for _, tc := range []struct {
		name string
		flag uint8
		want uint16
	}{
		{"沒有那一位", 0x00, 0},
		{"MIST GRUB 那一位", 0x01, 1},
		{"碎成一地那一位", 0x80, 0},
		{"兩位都有", 0x81, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o := townMap(t)
			o.SyncDGroupFromDS()
			// 記錄裡放可辨識的值，才看得出哪一格搬到哪一格。
			o.WriteU8(o.DS(body+0x10), tc.flag)
			fields := [5]struct {
				src uint16
				val uint8
			}{{0x11, 0x21}, {0x12, 0x22}, {0x13, 0x23}, {0x14, 0x24}, {0x16, 0x26}}
			for _, f := range fields {
				o.WriteU8(o.DS(body+f.src), f.val)
			}
			o.WriteU8(o.DS(body+0x15), 0xEE) // 這一格原版沒搬，拿來當反證
			for off := uint16(0x08); off <= 0x0C; off++ {
				o.WriteU8(o.DS(part+off), 0)
			}

			if _, err := o.CallBudget(400_000, o.IDA(initFn), slot); err != nil {
				t.Logf("呼叫沒等到返回（結果已經寫進全域）：%v", err)
			}
			if got := o.Word(o.IDA(flagOut)); got != tc.want {
				t.Errorf("word_3AC58：原版 %d，規則 %d（旗標 %#02x）", got, tc.want, tc.flag)
			}
			for k, f := range fields {
				dst := uint16(0x08 + k)
				if got := o.Byte(o.DS(part + dst)); got != f.val {
					t.Errorf("參戰者 +%#x：原版 %#02x，應為記錄 +%#x 的 %#02x", dst, got, f.src, f.val)
				}
			}
		})
	}
}

// 雲層動畫 `sub_2A18E`：呼叫一次，雲層點陣就整列右捲一個像素。
//
// 呼叫端只有 `0x11593`（在 `sub_11390` 裡）：天候 `word_2DF5A` 是 3 或 4 時走另一條
// （三張圖輪播，幀號在 `word_2DF5C`），其餘天候才捲雲，捲完把 `word_34A1A + 0x117`
// 段貼上畫面。
//
// 版面全部相對於 `word_34A1A` 這個段：
//
//	base:0x0000  地形底圖
//	base:0x07C0  雲層
//	base:0x0F80  遮罩（31 列 × 16 bytes，四個區塊共用）
//	base:0x1170  合成結果（= base + 0x117 段的 0 offset）
//
// EGA（`word_2DF1E == 0x0D`）雲層與遮罩都是 1bpp，一起當成 155 列 × 16 bytes 右旋
// 1 位元；Tandy（`== 0x09`）雲層是 4bpp，31 列 × 64 bytes 右旋 1 個 nibble（＝1 像素），
// 遮罩另外 31 列 × 16 bytes 右旋 1 位元。
func TestCloudScrollMatchesOriginal(t *testing.T) {
	const (
		cloudFn = 0x2A18E
		videoEA = 0x2DF1E
		bufEA   = 0x34A1A
	)
	// rotRow 把一列右旋 bits 個位元（列首是最高位）。
	rotRow := func(row []byte, bits uint) []byte {
		out := make([]byte, len(row))
		n := len(row)
		for i := range row {
			prev := row[(i-1+n)%n]
			out[i] = row[i]>>bits | prev<<(8-bits)
		}
		return out
	}
	fill := func(o *oracle.Oracle, seg, off uint16, n int) []byte {
		src := make([]byte, n)
		x := uint32(0x1234)
		for i := range src {
			x = x*1103515245 + 12345
			src[i] = uint8(x >> 16)
			o.WriteU8(oracle.Addr{Seg: seg, Off: off + uint16(i)}, src[i])
		}
		return src
	}
	check := func(t *testing.T, o *oracle.Oracle, seg, off uint16, src []byte, rowLen int, bits uint, label string) {
		t.Helper()
		bad := 0
		for r := 0; r*rowLen < len(src); r++ {
			want := rotRow(src[r*rowLen:(r+1)*rowLen], bits)
			got := o.Bytes(oracle.Addr{Seg: seg, Off: off + uint16(r*rowLen)}, rowLen)
			for i := range want {
				if want[i] != got[i] {
					if bad < 4 {
						t.Errorf("%s 第 %d 列第 %d 格：原版 %#02x，右旋 %d 位元應為 %#02x",
							label, r, i, got[i], bits, want[i])
					}
					bad++
					break
				}
			}
		}
		if bad > 4 {
			t.Errorf("%s：共 %d 列不符", label, bad)
		}
	}

	for _, tc := range []struct {
		name  string
		mode  uint16
		cloud struct {
			rowLen, rows int
			bits         uint
		}
	}{
		{name: "EGA", mode: 0x0D, cloud: struct {
			rowLen, rows int
			bits         uint
		}{16, 155, 1}},
		{name: "Tandy", mode: 0x09, cloud: struct {
			rowLen, rows int
			bits         uint
		}{64, 31, 4}},
		{name: "CGA", mode: 0x04, cloud: struct {
			rowLen, rows int
			bits         uint
		}{32, 31, 2}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o := townMap(t)
			o.SyncDGroupFromDS()
			buf := o.Word(o.IDA(bufEA))
			if buf == 0 {
				t.Skipf("word_34A1A 還沒配置（值 %#04x），這個狀態下量不到雲層", buf)
			}
			o.WriteU16(o.IDA(videoEA), tc.mode)

			n := tc.cloud.rowLen * tc.cloud.rows
			src := fill(o, buf, 0x7C0, n)
			var mask []byte
			if tc.name != "EGA" {
				// EGA 的遮罩接在雲層後面，同一個迴圈就掃完了；另外兩條要單獨填。
				mask = fill(o, buf, 0xF80, 31*16)
			}
			if _, err := o.CallBudget(2_000_000, o.IDA(cloudFn)); err != nil {
				t.Logf("呼叫沒等到返回：%v", err)
			}
			check(t, o, buf, 0x7C0, src, tc.cloud.rowLen, tc.cloud.bits, "雲層")
			if mask != nil {
				check(t, o, buf, 0xF80, mask, 16, 1, "遮罩")
			}
		})
	}
}
