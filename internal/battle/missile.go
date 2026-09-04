package battle

// 射擊那一手的彈道（RE-102）。指令碼 0x23 在 `sub_26D76:0x26EAD` 先把行動者的
// 位置與規劃段記住的目標格擺好，再 `sub_25D7A(0)` 拉一條線；線上每一格由
// `sub_25F8A(0, 步數)` 判一次，停下來的原因記在 `word_30A74`。

// 彈道停下來的原因，值就是 `word_30A74`。
const (
	// MissileClear 是一路走到目標格。
	MissileClear = 0
	// MissileTree 是撞在樹上（記錄 0x121 + 0x122）。
	MissileTree = 1
	// MissileActor 是被線上的人擋掉（`0x2610C`）。
	MissileActor = 2
	// MissileWall 是撞在牆或碎石上（記錄 0x121 + 0x123）。
	MissileWall = 3
)

const (
	// TreeTile 是樹（`sub_25F8A:0x25FD4`）。穿過去要花 TreeCost 點命中值，
	// 花不起就停在那一格。
	TreeTile = 0x0A
	TreeCost = 0x23
)

// MissileShot 是一發射出去的結果。
type MissileShot struct {
	// X／Y 是停下來的那一格；沒射出去（起點就是終點）時是行動者自己那一格。
	X, Y int
	// Block 是停下來的原因，見上面四個常數。
	Block int
	// Score 是扣掉沿路擋住的東西之後剩下的命中值。
	Score int
	// Target 是被擋下來時擋在前面的那一名；沒擋到人是 -1。
	Target int
	// Stopped 是這一發沒走到目標格就停了（`sub_25D7A` 回 1）。**Block 不能拿來
	// 判這件事**：穿得過去的樹與人照樣把 Block 留在 `word_30A74` 上，
	// 原版後面那一段 `word_30A74 == 2` 的分支就是靠這個留下來的值。
	Stopped bool
}

// Shoot 把一發從第 i 名射向 (tx, ty)，score 是這一擊的命中值。
//
// 每一格依序判三件事（`sub_25F8A`）：是樹就扣 TreeCost、擋路的磚就停、
// 都不是就看那一格站不站人。站著人的話扣掉他體型的一半再往前，扣到負就停在
// 他身上——**除了目標格本身**，走到那裡就算送到了。
//
// 原版扣成負之後會回頭再判同一格（`0x2612B` 跳回 `0x260B1`），而 `sub_24C35`
// 每次都找回同一個人，要靠 16-bit 溢位才繞得出來。這裡扣一次就停，
// 結果與原版繞完之後相同。
func (s *State) Shoot(i, tx, ty, score int) MissileShot {
	shot := MissileShot{Block: MissileClear, Score: score, Target: -1}
	if i < 0 || i >= len(s.Actors) {
		return shot
	}
	a := s.Actors[i]
	shot.X, shot.Y = a.X, a.Y
	foot := s.Field.At(a.X, a.Y)
	if a.Flag1E != 0 {
		foot = MountedSight
	}
	walkLine(a.X, a.Y, tx, ty, func(k, r, c int) bool {
		shot.X, shot.Y = r, c
		tile := s.Field.At(r, c)
		switch {
		case tile == TreeTile:
			shot.Block = MissileTree
			shot.Score -= TreeCost
			if shot.Score < 0 {
				shot.Stopped = true
				return false
			}
		case Blocks(tile) && !seesOver(tile, foot, k):
			shot.Block = MissileWall
			shot.Stopped = true
			return false
		}
		// 走到目標格就結束，那一格站的人是要打的人，不算擋路（`0x260BA`）。
		if r == tx && c == ty {
			return true
		}
		j := s.actorAt(r, c)
		if j < 0 || j == i || s.Actors[j].Down() {
			return true
		}
		shot.Block = MissileActor
		shot.Score -= s.Actors[j].Size / 2
		if shot.Score < 0 {
			shot.Target = j
			shot.Stopped = true
			return false
		}
		return true
	})
	return shot
}
