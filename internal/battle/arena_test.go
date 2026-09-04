package battle

import "testing"

func arenaParty() []Member {
	return []Member{{Name: "POONTANG"}, {Name: "MORAINE"}, {Name: "FIGHTING BULL"},
		{Name: "KENNETH"}, {Name: "SHARAKAN"}, {Name: "OLEAR"}}
}

// 競技場那一場：只有上場那一名參戰，站固定格；那一隻敵人也站固定格。
func TestNewArenaPlacesOnlyTheFighter(t *testing.T) {
	mon := Monster{Index: 3, Name: "HILL GIANT", Size: 100, Health: 100,
		AttackSkill: 200, ParrySkill: 100}
	s := NewArena(0, 3, 4, mon, arenaParty(), Generate(&fixedRand{values: []int{1, 2}}))
	if s == nil {
		t.Fatal("該開得起來")
	}
	if s.Stage != StageEncounter {
		t.Fatalf("階段是 %d，期望 StageEncounter", s.Stage)
	}
	if a := s.Actors[0]; a.Side != SideParty || a.X != ArenaFighterX || a.Y != ArenaFighterY {
		t.Fatalf("上場那一名站在 (%d,%d) 陣營 %d，期望 (%d,%d) SideParty",
			a.X, a.Y, a.Side, ArenaFighterX, ArenaFighterY)
	}
	// 其餘五名排到觀眾席，記錄 `+0x01` 是 2——這一版的 SideEmpty。
	for n := 1; n < 6; n++ {
		a := s.Actors[n]
		if a.Side != SideEmpty {
			t.Fatalf("第 %d 名的陣營是 %d，期望 SideEmpty（不參戰）", n, a.Side)
		}
		if a.X != ArenaWatcherX[n-1] || a.Y != ArenaWatcherY[n-1] {
			t.Fatalf("第 %d 名站在 (%d,%d)，期望觀眾席第 %d 格 (%d,%d)",
				n, a.X, a.Y, n-1, ArenaWatcherX[n-1], ArenaWatcherY[n-1])
		}
	}
	if a := s.Actors[PartySlots]; a.Side != SideMonster ||
		a.X != ArenaEnemyX || a.Y != ArenaEnemyY {
		t.Fatalf("敵人站在 (%d,%d) 陣營 %d，期望 (%d,%d) SideMonster",
			a.X, a.Y, a.Side, ArenaEnemyX, ArenaEnemyY)
	}
	if s.Actors[PartySlots+1].Side != SideEmpty {
		t.Fatal("只有一隻敵人")
	}
	if s.MonsterLevel != 4 {
		t.Fatalf("對手等級 %d，期望 4", s.MonsterLevel)
	}
}

// 觀眾序號只有排到觀眾席的人才加一：上場那一名不佔號碼（`0x1E72C` 的 `word_39A4E`）。
func TestNewArenaWatcherIndexSkipsTheFighter(t *testing.T) {
	mon := Monster{Index: 3, Name: "HILL GIANT", Size: 100, Health: 100}
	s := NewArena(2, 3, 4, mon, arenaParty(), Generate(&fixedRand{values: []int{1, 2}}))
	if s == nil {
		t.Fatal("該開得起來")
	}
	want := []int{0, 1, -1, 2, 3, 4} // 每一名用觀眾席的第幾格，−1 是上場那一名
	for n, k := range want {
		a := s.Actors[n]
		if k < 0 {
			if a.Side != SideParty || a.X != ArenaFighterX || a.Y != ArenaFighterY {
				t.Fatalf("第 %d 名該是上場那一名，站在 (%d,%d)", n, a.X, a.Y)
			}
			continue
		}
		if a.X != ArenaWatcherX[k] || a.Y != ArenaWatcherY[k] {
			t.Fatalf("第 %d 名站在 (%d,%d)，期望觀眾席第 %d 格 (%d,%d)",
				n, a.X, a.Y, k, ArenaWatcherX[k], ArenaWatcherY[k])
		}
	}
}

// 上場那一名不在隊伍裡就開不起來。
func TestNewArenaRejectsBadFighter(t *testing.T) {
	mon := Monster{Index: 3, Name: "HILL GIANT"}
	for _, n := range []int{-1, 6} {
		if s := NewArena(n, 3, 4, mon, arenaParty(),
			Generate(&fixedRand{values: []int{1, 2}})); s != nil {
			t.Fatalf("fighter=%d 該回 nil", n)
		}
	}
}

// 隊伍只有一人（原版的單人隊伍跳過賭場直接打）：沒有人排到觀眾席。
func TestNewArenaSoloParty(t *testing.T) {
	mon := Monster{Index: 3, Name: "HILL GIANT", Size: 100, Health: 100}
	s := NewArena(0, 3, 1, mon, []Member{{Name: "POONTANG"}},
		Generate(&fixedRand{values: []int{1, 2}}))
	if s == nil {
		t.Fatal("該開得起來")
	}
	if a := s.Actors[0]; a.Side != SideParty || a.X != ArenaFighterX {
		t.Fatalf("那一名站在 (%d,%d) 陣營 %d", a.X, a.Y, a.Side)
	}
	for n := 1; n < PartySlots; n++ {
		if s.Actors[n].Side != SideEmpty || s.Actors[n].Name != "" {
			t.Fatalf("第 %d 格該是空的，得到 %+v", n, s.Actors[n])
		}
	}
}
