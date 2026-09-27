package battle

import "testing"

// `sub_21D94` 依種族寫 `+0x03`：Kelden 2、Dwarven 0、其餘 1（RE-148）。
func TestMoveAbility(t *testing.T) {
	for race, want := range map[int]int{RaceHuman: 1, RaceElven: 1, RaceDwarven: 0, RaceKelden: 2} {
		if got := MoveAbility(race); got != want {
			t.Errorf("種族 %d：得到 %d，應為 %d", race, got, want)
		}
	}
}

// 腿廢掉清 `+0x03` 與 `+0x25`；手臂廢掉只清 `+0x25`（`sub_277F4:loc_27A5B`）。
func TestTakeClearsCommandAndMoveAbility(t *testing.T) {
	for _, c := range []struct {
		location    int
		wantMove    int
		wantCommand int
	}{
		{location: 2, wantMove: 2, wantCommand: 0},
		{location: 4, wantMove: 0, wantCommand: 0},
	} {
		a := Actor{Threshold: 200, Flag03: 2, Command: CmdMove}
		a.SetParts()
		if _, limb := a.Take(100, c.location); !limb {
			t.Fatalf("部位 %d 應該廢掉", c.location)
		}
		if a.Flag03 != c.wantMove || a.Command != c.wantCommand {
			t.Errorf("部位 %d：+0x03=%d、指令 %#x，應為 %d、%#x",
				c.location, a.Flag03, a.Command, c.wantMove, c.wantCommand)
		}
	}
	// 頭與胸廢掉是昏迷，不走這一條。
	a := Actor{Threshold: 200, Flag03: 2, Command: CmdMove}
	a.SetParts()
	a.Take(100, 0)
	if a.Command != CmdMove || a.Flag03 != 2 {
		t.Errorf("頭部不該清指令或移動能力，得到 %#x／%d", a.Command, a.Flag03)
	}
}

// 治好隊員的腿再跑一次 `sub_21D94`，移動能力照種族回來。
func TestHealLegRestoresMoveAbility(t *testing.T) {
	s := fleeState()
	a := &s.Actors[0]
	a.Race, a.Flag03 = RaceKelden, 0
	a.RightLeg = 0
	s.castArea(0, 4, 5, false)
	if a.Flag03 != 2 {
		t.Fatalf("治好 Kelden 的腿之後 +0x03 應為 2，得到 %d", a.Flag03)
	}
}
