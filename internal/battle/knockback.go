package battle

// 擊退與打暈（RE-095）。攻擊命中之後 `0x27B77` 起判一次：體型差越大越容易把對手
// 推開一格；推得動就是擊退、推不動就是打暈。兩種都會讓對手掉武器，
// 並把它這一回合還沒做的那一手取消掉。

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
// `(欄差 < 0 ? 0 : 欄差 == 0 ? 4 : 8) + (列差 < 0 ? 1 : 列差 > 0 ? 2 : 0)`。
//
// 用不到的組（3、7 與同格的 4）原版就是 `0 0 0`，照抄不動。
var knockbackChoices = [11][3]int{
	{6, 7, 0}, {5, 6, 7}, {4, 5, 6}, {0, 0, 0},
	{7, 0, 1}, {0, 0, 0}, {3, 4, 5}, {0, 0, 0},
	{0, 1, 2}, {1, 2, 3}, {2, 3, 4},
}

// KnockbackIndex 把「目標相對於攻擊者在哪」換成 knockbackChoices 的索引。
//
// 原版用 +0x1F（欄）與 +0x20（列）；remake 的軸是對調的（RE-095），
// 所以欄差是 `Y` 的差、列差是 `X` 的差。
func KnockbackIndex(attacker, target Actor) int {
	index := 4
	switch {
	case target.Y < attacker.Y:
		index = 0
	case target.Y > attacker.Y:
		index = 8
	}
	switch {
	case target.X < attacker.X:
		index++
	case target.X > attacker.X:
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
// 兩種都會讓對手掉武器（`sub_269BA`）並把它這一回合的指令取消。
func (s *State) Knockback(attacker, target int, rng Rand) []int {
	if attacker < 0 || target < 0 ||
		attacker >= len(s.Actors) || target >= len(s.Actors) {
		return nil
	}
	a, d := s.Actors[attacker], s.Actors[target]
	// 走到這裡一定擲 rand(100)，體型差不是正的也照擲（`0x27BC4`），只是必定不成立。
	margin := a.Size - d.Size
	if rng.Roll(100) >= margin {
		return nil
	}

	choices := knockbackChoices[KnockbackIndex(a, d)]
	facing := choices[rng.Roll(3)-1]
	step := ring0[facing]
	x, y := d.X+step.Row, d.Y+step.Col

	armed := s.Actors[target].Weapon.Kind != 0
	var out []int
	if s.actorAt(x, y) < 0 && !s.moveBlocked(target, x, y) {
		// 推得動：目標往那一格退一步。
		s.Actors[target].X, s.Actors[target].Y = x, y
		out = append(out, KnocksRecord)
		if !armed {
			out = append(out, BackwardsRecord)
		} else {
			out = append(out, BackwardsComma, FallsRecord)
		}
	} else {
		// 推不動：只有手上有東西時才印那兩句（`0x27DD4`）。
		if !armed {
			return nil
		}
		out = append(out, StunsRecord, FallsRecord)
	}

	// 掉武器，並取消這一回合還沒做的那一手。
	if m := s.member(target); m != nil {
		m.HasWeapon = false
		m.WeaponName = ""
		m.Weapon = Weapon{}
		m.Worn[0] = ""
	}
	s.Actors[target].Weapon = Weapon{}
	s.Actors[target].Command = 0
	return out
}
