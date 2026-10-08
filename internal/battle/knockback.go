package battle

// 擊退與打暈（RE-095）。攻擊命中之後 `0x27B77` 起判一次：體型差越大越容易把對手
// 推開一格；推得動就是擊退、推不動就是打暈。成功擊退才掉武器。
// 受阻持械保留完整持物並取消指令（SPEC-008／RE-253、RE-254）。

// 擊退與打暈的訊息記錄。
const (
	StunsRecord     = 0x158 // "The blow stuns|F]"
	KnocksRecord    = 0x159 // "The force of the blow knocks|F]"
	BackwardsComma  = 0x15A // "backwards,"
	FallsRecord     = 0x15B // "and|A|H falls to the ground!"
	BackwardsRecord = 0x15C // "backwards!"
)

// knockbackChoices 是 `[0x123A]`：相對位置 → 三個候選方向，每組三個，
// 由 `rand(3) − 1` 挑一個。索引是原版的
// `(X差 < 0 ? 0 : X差 == 0 ? 4 : 8) + (Y差 < 0 ? 1 : Y差 > 0 ? 2 : 0)`。
//
// 組 3、5、7 的原始 bytes 是 `0 0 0`，保留原值。
var knockbackChoices = [11][3]int{
	{6, 7, 0}, {5, 6, 7}, {4, 5, 6}, {0, 0, 0},
	{7, 0, 1}, {0, 0, 0}, {3, 4, 5}, {0, 0, 0},
	{0, 1, 2}, {1, 2, 3}, {2, 3, 4},
}

// KnockbackIndex 把「目標相對於攻擊者在哪」換成 knockbackChoices 的索引。
//
// RE-219 勘誤：原版 +0x1F 對應 Actor.X，+0x20 對應 Actor.Y。
// 0x27BE9..0x27C30 以 X 決定主索引；0x27C43..0x27C81 以 Y 決定增量。
func KnockbackIndex(attacker, target Actor) int {
	index := 4
	switch {
	case target.X < attacker.X:
		index = 0
	case target.X > attacker.X:
		index = 8
	}
	switch {
	case target.Y < attacker.Y:
		index++
	case target.Y > attacker.Y:
		index += 2
	}
	return index
}

// KnocksBack 回報這一擊要不要判擊退。原版的判定在受傷那一支 `sub_277F4` 裡，只有
// 「causing … wound」那一條、而且傷害到 0x10 以上（印 0x156／0x157 的那兩種）才走到
// `loc_27B56`；傷害為 0、倒下、部位廢掉、傷得輕（0x155）都不判（#9 原版執行段對拍）。
// 近戰與射擊共用這一支。
func KnocksBack(out Outcome) bool {
	return !out.Miss && !out.Parried && out.Damage >= woundSerious && !out.Down && !out.Limb
}

// Knockback 是攻擊命中之後的擊退判定（`0x27B77`）。回傳這一手還要印的訊息記錄
// （依序），沒有擊退也沒有打暈就回 nil。
//
// 機率是**體型差**：`rand(100)` 小於「攻擊者體型 − 目標體型」才成立，
// 所以體型差越大越容易推得動，同體型幾乎推不動。
//
// 推得動（那一格沒人又站得下）就把目標移過去，印 0x159；推不動就印 0x158。
// 成功擊退才呼叫 sub_269BA；受阻持械只取消指令，空手受阻直接返回。
func (s *State) Knockback(attacker, target int, rng Rand) []int {
	records, _ := s.knockback(attacker, target, rng)
	return records
}

// knockback 同時帶出實際掉落名稱，供訊息的 |H 使用（RE-252）。
func (s *State) knockback(attacker, target int, rng Rand) ([]int, string) {
	if attacker < 0 || target < 0 ||
		attacker >= len(s.Actors) || target >= len(s.Actors) {
		return nil, ""
	}
	a, d := s.Actors[attacker], s.Actors[target]
	// 走到這裡一定擲 rand(100)，體型差不是正的也照擲（`0x27BC4`），只是必定不成立。
	margin := a.Size - d.Size
	if rng.Roll(100) >= margin {
		return nil, ""
	}

	choices := knockbackChoices[KnockbackIndex(a, d)]
	facing := choices[rng.Roll(3)-1]
	step := ring0[facing]
	x, y := d.X+step.Row, d.Y+step.Col

	armed := s.Actors[target].Weapon.Kind != 0
	var out []int
	item := ""
	if !s.knockBlocked(target, x, y) {
		// RE-222：成功位移另讀 DGROUP:0x0C1C／0x0C2C，對應 ring0 的 Col／Row。
		// 碰撞檢查仍使用上面的 x／y；原版對 +0x1F／+0x20 做 byte 加法。
		s.Actors[target].X = int(uint8(d.X + step.Col))
		s.Actors[target].Y = int(uint8(d.Y + step.Row))
		out = append(out, KnocksRecord)
		if !armed {
			out = append(out, BackwardsRecord)
		} else {
			out = append(out, BackwardsComma, FallsRecord)
		}
		item = s.forcedDrop(target)
	} else {
		// 推不動：只有手上有東西時才印那兩句（`0x27DD4`）。
		if !armed {
			return nil, ""
		}
		out = append(out, StunsRecord, FallsRecord)
	}

	// RE-253／254：共同尾端清角色記錄 +0x25，受阻分支不搬移物品。
	s.Actors[target].Command = 0
	return out, item
}

// 原版 0x27CE6／0x27D07：遮罩8的人物搜尋，再查地形表。
// 保留既有安全邊界；不套一般移動的騎乘或腳下條件。
func (s *State) knockBlocked(target, x, y int) bool {
	if x < FieldMin || x > FieldMax || y < FieldMin || y > FieldMax {
		return true
	}
	if s.Field != nil && BlocksMovement(s.Field.At(x, y)) {
		return true
	}
	for i, a := range s.Actors {
		if i != target && a.Side != SideEmpty && !a.Down() && a.Flag1E == 0 && a.X == x && a.Y == y {
			return true
		}
	}
	return false
}
