package battle

import "testing"

func knockState(attackerSize, targetSize int) *State {
	s := &State{Field: &Field{Tiles: make([]int, FieldSide*FieldSide)},
		Party: []Member{{Name: "POONTANG", HasWeapon: true, WeaponName: "LONG SPEAR"}},
		Actors: []Actor{
			{Side: SideMonster, X: 10, Y: 10, Size: attackerSize, Member: -1,
				Head: 10, Chest: 10, Threshold: 100},
			{Side: SideParty, Member: 0, X: 10, Y: 11, Size: targetSize,
				Head: 10, Chest: 10, Threshold: 100, Weapon: Weapon{Kind: 0x20},
				Command: CmdWeapon},
		}}
	return s
}

// 體型差是機率：差多少就是百分之多少。同體型或更小推不動。
func TestKnockbackNeedsSizeAdvantage(t *testing.T) {
	s := knockState(100, 100)
	if got := s.Knockback(0, 1, &fixedRand{values: []int{1}}); got != nil {
		t.Fatalf("同體型不該推得動，得到 %v", got)
	}
	s = knockState(100, 150)
	if got := s.Knockback(0, 1, &fixedRand{values: []int{1}}); got != nil {
		t.Fatalf("比對手小不該推得動，得到 %v", got)
	}
	// 體型差 40：rand(100) 擲出 41 以上就不成立。
	s = knockState(140, 100)
	if got := s.Knockback(0, 1, &fixedRand{values: []int{41}}); got != nil {
		t.Fatalf("擲出 41 不該推得動（差 40），得到 %v", got)
	}
}

// 推得動：目標退一格，印 0x159，而且武器掉了、這一回合的手取消。
func TestKnockbackPushesTargetAway(t *testing.T) {
	s := knockState(140, 100)
	got := s.Knockback(0, 1, &fixedRand{values: []int{1}})
	if len(got) == 0 || got[0] != KnocksRecord {
		t.Fatalf("第一句應該是 %#x，得到 %v", KnocksRecord, got)
	}
	d := s.Actors[1]
	if d.X == 10 && d.Y == 11 {
		t.Fatal("推得動的話目標應該換一格")
	}
	// 目標在攻擊者的 Y 正向，所以會被往 Y 正向推開，不會被推向攻擊者。
	if d.Y < 11 {
		t.Fatalf("被推向攻擊者了：(%d,%d)", d.X, d.Y)
	}
	if s.Actors[1].Weapon.Kind != 0 || s.Party[0].HasWeapon {
		t.Fatal("被擊退的人應該掉了武器")
	}
	if s.Actors[1].Command != 0 {
		t.Fatal("被擊退的人這一回合的手應該取消")
	}
}

// 推不動（那一格有人）就變成打暈，印 0x158。
func TestKnockbackBecomesStunWhenBlocked(t *testing.T) {
	s := knockState(140, 100)
	// 把退路堵住：目標會被推向 Y 正向。
	for y := 10; y <= 12; y++ {
		for x := 9; x <= 11; x++ {
			if x == 10 && y == 10 || x == 10 && y == 11 {
				continue
			}
			s.Actors = append(s.Actors, Actor{Side: SideMonster, Member: -1,
				X: x, Y: y, Head: 10, Chest: 10, Threshold: 100})
		}
	}
	got := s.Knockback(0, 1, &fixedRand{values: []int{1}})
	if len(got) != 2 || got[0] != StunsRecord || got[1] != FallsRecord {
		t.Fatalf("堵住了應該印 %#x＋%#x，得到 %v", StunsRecord, FallsRecord, got)
	}
	if s.Actors[1].X != 10 || s.Actors[1].Y != 11 {
		t.Fatal("推不動就不該移動")
	}
}

// 手上空空的人被推不動時什麼都不印（`0x27DD4`）。
func TestKnockbackSilentWhenUnarmedAndBlocked(t *testing.T) {
	s := knockState(140, 100)
	s.Actors[1].Weapon = Weapon{}
	for y := 10; y <= 12; y++ {
		for x := 9; x <= 11; x++ {
			if x == 10 && y == 10 || x == 10 && y == 11 {
				continue
			}
			s.Actors = append(s.Actors, Actor{Side: SideMonster, Member: -1,
				X: x, Y: y, Head: 10, Chest: 10, Threshold: 100})
		}
	}
	if got := s.Knockback(0, 1, &fixedRand{values: []int{1}}); got != nil {
		t.Fatalf("空手又推不動應該什麼都不印，得到 %v", got)
	}
}

// 相對位置換成候選組的索引：原版用欄與列，remake 的軸是對調的。
func TestKnockbackIndex(t *testing.T) {
	at := Actor{X: 10, Y: 10}
	for _, c := range []struct {
		name string
		t    Actor
		want int
	}{
		{"正右（欄大）", Actor{X: 10, Y: 11}, 8},
		{"正左（欄小）", Actor{X: 10, Y: 9}, 0},
		{"正上（列小）", Actor{X: 9, Y: 10}, 5},
		{"正下（列大）", Actor{X: 11, Y: 10}, 6},
		{"右下", Actor{X: 11, Y: 11}, 10},
	} {
		if got := KnockbackIndex(at, c.t); got != c.want {
			t.Fatalf("%s 的索引是 %d，期望 %d", c.name, got, c.want)
		}
	}
}
