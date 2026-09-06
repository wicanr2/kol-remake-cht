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
