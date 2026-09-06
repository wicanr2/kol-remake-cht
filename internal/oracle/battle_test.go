//go:build oracle

package oracle

import (
	"testing"

	"github.com/wicanr2/dosgolem/oracle"

	"github.com/wicanr2/kol-remake-cht/internal/battle"
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
