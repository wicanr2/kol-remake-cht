package battle

import "testing"

// 射擊用的戰場：整片底色，兩名面對面隔七格。
func shootState() *State {
	s := &State{Field: NewField(), Actors: make([]Actor, 2)}
	s.Actors[0] = Actor{Side: SideParty, X: 10, Y: 10, Facing: 2,
		Threshold: 100, Quickness: 50, Strength: 90, Level: 3, Attack: 400,
		Weapon: Weapon{Kind: 6, Base: 3, Count: 1, Sides: 5},
		ShieldGrade: NoShield}
	s.Actors[0].SetParts()
	s.Actors[1] = Actor{Side: SideMonster, X: 10, Y: 17, Threshold: 100,
		Quickness: 40, Size: 60, ShieldGrade: NoShield}
	s.Actors[1].SetParts()
	s.Exec.Order = []int{0}
	s.Plan.Memo = map[int]Cell{0: {Row: 10, Col: 17}}
	s.Actors[0].Command, s.Actors[0].Style = CmdMissile, 2
	return s
}

// 一路暢通就打到目標那一格上的人。
func TestShootReachesTarget(t *testing.T) {
	s := shootState()
	shot := s.Shoot(0, 10, 17, 100)
	if shot.Stopped {
		t.Fatalf("整片底色不該擋，停在 (%d, %d) 原因 %d", shot.X, shot.Y, shot.Block)
	}
	if shot.X != 10 || shot.Y != 17 {
		t.Fatalf("應該走到 (10, 17)，得到 (%d, %d)", shot.X, shot.Y)
	}
	if shot.Score != 100 {
		t.Fatalf("沒東西擋就不該扣命中值，得到 %d", shot.Score)
	}
}

// 牆停在牆那一格（`sub_25F8A:0x26099` 的 word_30A74 = 3）。
func TestShootStopsAtWall(t *testing.T) {
	s := shootState()
	// 0x29 是牆，位元圖有設而且不在 0x23…0x2D 的例外裡。
	s.Field.Tiles[10*FieldSide+13] = 0x29
	shot := s.Shoot(0, 10, 17, 100)
	if !shot.Stopped || shot.Block != MissileWall {
		t.Fatalf("應該停在牆上，得到 stopped=%v block=%d", shot.Stopped, shot.Block)
	}
	if shot.X != 10 || shot.Y != 13 {
		t.Fatalf("應該停在 (10, 13)，得到 (%d, %d)", shot.X, shot.Y)
	}
}

// 樹扣 0x23 點命中值：扣得起就穿過去，扣不起就停在樹上。
func TestShootTreeCostsScore(t *testing.T) {
	s := shootState()
	s.Field.Tiles[10*FieldSide+13] = TreeTile
	shot := s.Shoot(0, 10, 17, 100)
	if shot.Stopped {
		t.Fatal("命中值夠就該穿過樹")
	}
	if shot.Score != 100-TreeCost {
		t.Fatalf("穿過樹應該扣 %d，剩 %d", TreeCost, shot.Score)
	}
	// 停下來的原因留在 Block 上——原版的 word_30A74 也不會被清掉。
	if shot.Block != MissileTree {
		t.Fatalf("Block 應該留成 %d，得到 %d", MissileTree, shot.Block)
	}
	shot = s.Shoot(0, 10, 17, 20)
	if !shot.Stopped || shot.Block != MissileTree || shot.Y != 13 {
		t.Fatalf("命中值不夠應該停在樹上，得到 %+v", shot)
	}
}

// 半路站著的人擋掉一半體型的命中值；擋不住就繼續往前。
func TestShootBlockedByActor(t *testing.T) {
	s := shootState()
	s.Actors = append(s.Actors, Actor{Side: SideMonster, X: 10, Y: 13,
		Threshold: 100, Size: 80, ShieldGrade: NoShield})
	s.Actors[2].SetParts()
	shot := s.Shoot(0, 10, 17, 100)
	if shot.Stopped {
		t.Fatal("100 扣掉 40 還有剩，應該穿過去")
	}
	if shot.Score != 60 || shot.Block != MissileActor {
		t.Fatalf("應該扣掉體型的一半，得到 score=%d block=%d", shot.Score, shot.Block)
	}
	shot = s.Shoot(0, 10, 17, 30)
	if !shot.Stopped || shot.Target != 2 {
		t.Fatalf("擋不過去應該停在那個人身上，得到 %+v", shot)
	}
	// 目標格上的人不算擋路（`0x260BA`）。
	shot = s.Shoot(0, 10, 13, 30)
	if shot.Stopped || shot.Target >= 0 {
		t.Fatalf("目標格本身不該算擋路，得到 %+v", shot)
	}
}

// 要裝填的武器第一回合只裝填，射完又清回沒裝填（RE-093）。
func TestShootNeedsReload(t *testing.T) {
	if NeedsReload(6) || NeedsReload(10) || NeedsReload(11) {
		t.Fatal("6／10／11 這三種不用裝填")
	}
	if !NeedsReload(0x12) {
		t.Fatal("其餘種類要裝填")
	}
	s := shootState()
	s.Actors[0].Weapon.Kind = 0x12
	r, _ := s.Act(&fixedRoll{values: []int{50}})
	if r.Message != ReloadRecord || r.Command != 0 {
		t.Fatalf("第一回合應該只裝填，得到 %#x／%#x", r.Message, r.Command)
	}
	if s.Actors[0].Flag1D != 1 {
		t.Fatal("裝填之後 Flag1D 應該是 1")
	}
	// 下一回合真的射出去，射完那一格清回 0。
	s.Actors[0].Command = CmdMissile
	s.Exec.Turn = 0
	r, _ = s.Act(&fixedRoll{values: []int{99, 1, 5, 50}})
	if r.Target != 1 || r.Outcome == nil {
		t.Fatalf("第二回合應該射得出去，得到 target=%d outcome=%v", r.Target, r.Outcome)
	}
	if s.Actors[0].Flag1D != 0 {
		t.Fatal("射完 Flag1D 應該清回 0")
	}
}

// 撞在樹或牆上印兩筆記錄，射到空地印「nothing is there」。
func TestShootMessages(t *testing.T) {
	s := shootState()
	s.Field.Tiles[10*FieldSide+13] = 0x29
	r, _ := s.Act(&fixedRoll{values: []int{50}})
	if len(r.Records) != 2 || r.Records[0] != MissileIntoRecord ||
		r.Records[1] != WallRecord {
		t.Fatalf("撞牆應該印 %#x + %#x，得到 %v", MissileIntoRecord, WallRecord, r.Records)
	}
	// 命中值是 490 ÷ 10 ＋ 擲骰，一棵樹扣 35 還穿得過，第二棵就扣不起了。
	s = shootState()
	s.Field.Tiles[10*FieldSide+13] = TreeTile
	s.Field.Tiles[10*FieldSide+14] = TreeTile
	r, _ = s.Act(&fixedRoll{values: []int{1}})
	if len(r.Records) != 2 || r.Records[1] != TreeRecord {
		t.Fatalf("命中值不夠時撞樹應該印 %#x，得到 %v", TreeRecord, r.Records)
	}
	// 目標格沒人。
	s = shootState()
	s.Actors[1].Y = 20
	r, _ = s.Act(&fixedRoll{values: []int{50}})
	if r.Message != MissileNothingRecord {
		t.Fatalf("射到空地應該印 %#x，得到 %#x", MissileNothingRecord, r.Message)
	}
}

// 被射的人防禦換成快捷的一半，招架與招架次數都不算（`0x271E3`）。
func TestMissileDefence(t *testing.T) {
	d := Actor{Quickness: 80, Parry: 900, Guard: 3, Parries: 1,
		Threshold: 100, ShieldGrade: NoShield}
	d.SetParts()
	if got, want := d.MissileDefenceValue(), 40; got != want {
		t.Fatalf("防禦骰前應該是 %d，得到 %d", want, got)
	}
	// 近戰那一套會把招架算進去，兩者不該相同。
	if d.DefenceValue() == d.MissileDefenceValue() {
		t.Fatal("射擊的防禦不該與近戰相同")
	}
	// 負的夾成 0（`0x27454`）。
	d.Encumbrance = 5000
	if got := MissileDefenceScore(d, 10); got != 0 {
		t.Fatalf("負的防禦應該夾成 0，得到 %d", got)
	}
}

// 打到半路擋住的人時攻防都改成純亂數（`0x27153`）。
func TestResolveMissileStray(t *testing.T) {
	a := Actor{Strength: 90, Attack: 400, Threshold: 100, Level: 3,
		Weapon: Weapon{Base: 3, Count: 1, Sides: 5}, ShieldGrade: NoShield}
	a.SetParts()
	d := Actor{Quickness: 40, Threshold: 100, ShieldGrade: NoShield}
	d.SetParts()
	// 攻 99、防 1：命中。彈道帶進來的 −500 不算數。
	out := ResolveMissile(&a, &d, 2, -500, true, &fixedRoll{values: []int{99, 1, 5, 50}})
	if out.Miss || out.Parried {
		t.Fatalf("攻 99 防 1 應該命中，得到 %+v", out)
	}
	// 同樣的命中值，不是誤傷就整個落空。
	out = ResolveMissile(&a, &d, 2, -500, false, &fixedRoll{values: []int{99}})
	if !out.Miss || out.Margin != 500 {
		t.Fatalf("命中值是負的應該落空 500，得到 %+v", out)
	}
}
