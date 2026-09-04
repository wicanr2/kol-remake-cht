package battle

import "testing"

// maxRand 每一次都擲出最大值，讓獎勵的算式可以逐項驗。
type maxRand struct{}

func (maxRand) Roll(n int) int { return n }

// 兩邊都還站著就繼續打。
func TestVerdictContinue(t *testing.T) {
	s := &State{Actors: []Actor{
		{Side: SideParty, Head: 10, Chest: 10},
		{Side: SideMonster, Head: 10, Chest: 10},
		{Side: SideEmpty},
	}}
	if got := s.RoundVerdict(); got != VerdictContinue {
		t.Fatalf("兩邊都站著應該繼續打，得到 %d", got)
	}
}

// 敵人全倒是打贏；隊員倒了幾個不影響。
func TestVerdictWon(t *testing.T) {
	s := &State{Actors: []Actor{
		{Side: SideParty, Head: 10, Chest: 10},
		{Side: SideParty, Head: 0, Chest: 10},
		{Side: SideMonster, Head: 0, Chest: 10},
		{Side: SideMonster, Head: 10, Chest: 0},
	}}
	if got := s.RoundVerdict(); got != VerdictWon {
		t.Fatalf("敵人全倒應該是打贏，得到 %d", got)
	}
}

// 隊伍全倒是打輸。原版的四個判斷依序來，打輸排在打贏前面，
// 所以兩邊同時全倒時算打輸。
func TestVerdictLostWinsTheTie(t *testing.T) {
	s := &State{Actors: []Actor{
		{Side: SideParty, Head: 0, Chest: 10},
		{Side: SideMonster, Head: 0, Chest: 10},
	}}
	if got := s.RoundVerdict(); got != VerdictLost {
		t.Fatalf("兩邊同時全倒應該算打輸，得到 %d", got)
	}
}

// 金冠：擲「骰數 ＋ 1」次、重複 waves 輪，總額除以隊伍人數。
func TestGoldReward(t *testing.T) {
	// 骰數 2 → 一輪三次；三輪 ＝ 九次；每次最大 99 ＝ 891；四個人分 ＝ 222。
	if got := GoldReward(2, 99, 3, 4, maxRand{}); got != 891/4 {
		t.Fatalf("每人金冠 %d，期望 %d", got, 891/4)
	}
	// 每人不到兩枚就一枚都不給。
	if got := GoldReward(0, 2, 1, 4, maxRand{}); got != 0 {
		t.Fatalf("每人不到兩枚應該歸零，得到 %d", got)
	}
	// 沒有人分不會除以零。
	if got := GoldReward(2, 99, 3, 0, maxRand{}); got != 0 {
		t.Fatalf("隊伍人數 0 應該回 0，得到 %d", got)
	}
}

// 骰式的拆法：高 nibble 的低三位是骰數、低三位查骰面表（RE-091）。
func TestGoldRoll(t *testing.T) {
	for _, c := range []struct{ raw, rolls, sides int }{
		{0x08, 0, 2},  // SKELETON
		{0x37, 3, 99}, // 高 nibble 3、低三位 7
		{0xF1, 7, 3},  // 高 nibble 只取低三位
	} {
		rolls, sides := GoldRoll(c.raw)
		if rolls != c.rolls || sides != c.sides {
			t.Fatalf("+0x0F = %#02x 拆成 %d 顆 %d 面，期望 %d 顆 %d 面",
				c.raw, rolls, sides, c.rolls, c.sides)
		}
	}
}

// 冒險點數：每一個參戰敵人的攻防技能加總，先除 2 再除隊伍人數。用的是記錄裡的
// 原值，而且不看敵人死了沒（RE-096）。
func TestAdventureReward(t *testing.T) {
	s := &State{Actors: []Actor{
		{Side: SideParty}, {Side: SideParty},
		// 死掉的這一隻照樣算進去。
		{Side: SideMonster, BaseAttack: 230, BaseParry: 100, Attack: 46, Parry: 20},
		{Side: SideMonster, BaseAttack: 170, BaseParry: 60},
	}}
	// (330 + 230) / 2 / 2 = 140。
	if got := s.AdventureReward(); got != 140 {
		t.Fatalf("每人冒險點數 %d，期望 140", got)
	}
	// 折算過等級的 Attack／Parry 不該被用到。
	if got := (&State{Actors: []Actor{{Side: SideParty},
		{Side: SideMonster, Attack: 46, Parry: 20}}}).AdventureReward(); got != 0 {
		t.Fatalf("沒有原值時應該是 0，得到 %d", got)
	}
	// 沒有人分不會除以零。
	if got := (&State{Actors: []Actor{
		{Side: SideMonster, BaseAttack: 10}}}).AdventureReward(); got != 0 {
		t.Fatalf("隊伍人數 0 應該回 0，得到 %d", got)
	}
}

// 金冠的骰式只讀場上第一隻敵人的，後面那幾隻不影響金額（RE-096）。
func TestGoldAwardUsesFirstEnemyOnly(t *testing.T) {
	s := &State{Actors: []Actor{
		{Side: SideParty},
		{Side: SideMonster, GoldRoll: 0x17}, // 1 顆 99 面
		{Side: SideMonster, GoldRoll: 0x70}, // 這一隻的骰式不該被讀到
	}}
	// 一輪、兩次、每次 99 → 198，一個人分。
	if got := s.GoldAward(maxRand{}); got != 198 {
		t.Fatalf("每人金冠 %d，期望 198", got)
	}
}

// 英文複數：名字以 S 結尾補 ES，否則補 S；只有一隻不補。
func TestPluralSuffix(t *testing.T) {
	for _, c := range []struct {
		name  string
		count int
		want  string
	}{
		{"BINDERAK", 12, "S"},
		{"MUCK THINGS", 2, "ES"},
		{"RUFFIAN", 1, ""},
		{"", 5, ""},
	} {
		if got := PluralSuffix(c.name, c.count); got != c.want {
			t.Fatalf("%q × %d 的複數是 %q，期望 %q", c.name, c.count, got, c.want)
		}
	}
}

// 打輸那一句的記錄：單數用 ` has`、複數用 `s have`。
func TestDefeatRecords(t *testing.T) {
	before, after := DefeatRecords(1)
	if len(before) != 1 || before[0] != TheRecord {
		t.Fatalf("名字前面應該只有 The，得到 %v", before)
	}
	if after[0] != SingularHasRecord {
		t.Fatalf("一隻應該用 has，得到 %v", after)
	}
	if _, after = DefeatRecords(3); after[0] != PluralHaveRecord {
		t.Fatalf("多隻應該用 have，得到 %v", after)
	}
}

// 獎勵那幾句：單人不印 `Each player`，兩個 `|N` 拆在前後兩段。
func TestRewardRecords(t *testing.T) {
	before, after := RewardRecords(1)
	if len(before) != 1 || before[0] != AdventureRecord {
		t.Fatalf("單人前半只該有冒險點數那一句，得到 %v", before)
	}
	if len(after) != 1 || after[0] != GoldRecord {
		t.Fatalf("後半該是金冠那一句，得到 %v", after)
	}
	before, after = RewardRecords(4)
	if len(before) != 2 || before[0] != EachPlayerRecord || before[1] != AdventureRecord {
		t.Fatalf("多人要先印 Each player 再印冒險點數，得到 %v", before)
	}
	if len(after) != 1 || after[0] != GoldRecord {
		t.Fatalf("後半該是金冠那一句，得到 %v", after)
	}
}

// Step 走完執行段之後要自己收尾：敵人全倒就結束這一場並算好金冠，
// 沒分勝負才排下一回合。這一支測的是收尾本身，攻擊結算另有測試。
func TestStepFinishesTheBattle(t *testing.T) {
	newState := func(enemyHead int) *State {
		s := &State{Actors: []Actor{
			{Side: SideParty, Head: 10, Chest: 10, Threshold: 100},
			// GoldRoll 0x27 拆成 2 顆 99 面；攻防技能餵冒險點數。
			{Side: SideMonster, Head: enemyHead, Chest: 10, Threshold: 100,
				GoldRoll: 0x27, BaseAttack: 230, BaseParry: 100},
		}}
		// 執行段已經走完：Turn 停在 Order 的尾端。
		s.Exec.Order, s.Exec.Turn = []int{0, 1}, 2
		return s
	}

	// 敵人倒了 → 打贏、結束、金冠算出來。
	won := newState(0)
	if _, ok := won.Step(maxRand{}); ok {
		t.Fatal("執行段已經走完，Step 不該回報還有下一手")
	}
	if won.Stage != StageDone {
		t.Fatalf("打贏之後 Stage 是 %d，期望 StageDone", won.Stage)
	}
	if won.Verdict != VerdictWon {
		t.Fatalf("Verdict 是 %d，期望 VerdictWon", won.Verdict)
	}
	// 一輪、擲三次滿分 99、一個人分 → 297。
	if won.Gold != 297 {
		t.Fatalf("每人金冠 %d，期望 297", won.Gold)
	}
	// (230 + 100) / 2 / 1 = 165。
	if won.Adventure != 165 {
		t.Fatalf("每人冒險點數 %d，期望 165", won.Adventure)
	}

	// 敵人還站著 → 排下一回合，不結束。
	going := newState(10)
	if _, ok := going.Step(maxRand{}); ok {
		t.Fatal("執行段已經走完，Step 不該回報還有下一手")
	}
	if going.Stage == StageDone {
		t.Fatal("兩邊都還站著不該結束")
	}
	if going.Verdict != VerdictContinue {
		t.Fatalf("Verdict 是 %d，期望 VerdictContinue", going.Verdict)
	}
	if going.Gold != 0 {
		t.Fatalf("沒打贏不該有金冠，得到 %d", going.Gold)
	}
}

// 打輸那一句的複數和打贏不一樣：記錄 0x178（`s have`）本身帶 s，名字結尾是 S
// 時只補 E，否則什麼都不補。兩句用同一套會印出 `BINDERAKSS HAVE`。
func TestDefeatSuffixDiffersFromVictory(t *testing.T) {
	for _, c := range []struct {
		name       string
		count      int
		won, lost  string
	}{
		{"BINDERAK", 12, "S", ""},
		{"MUCK THINGS", 2, "ES", "E"},
		{"RUFFIAN", 1, "", ""},
	} {
		if got := PluralSuffix(c.name, c.count); got != c.won {
			t.Fatalf("%q 打贏的後綴是 %q，期望 %q", c.name, got, c.won)
		}
		if got := DefeatSuffix(c.name, c.count); got != c.lost {
			t.Fatalf("%q 打輸的後綴是 %q，期望 %q", c.name, got, c.lost)
		}
	}
}
