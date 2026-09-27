package battle

import "testing"

// 穿戴部位 3／5／7／8 的種類 byte 大於 0x7F 才算數，最低兩位元決定設哪一格（RE-154）。
// 五件實際會過關的：Flying Cloak 0x83、Courage Coat 0x82、Shade Ring 0x81、
// Speed Boots 0x80、Shoes 0xE0。
func TestApplyCharms(t *testing.T) {
	for _, c := range []struct {
		name               string
		race, class        int
		f03, f04, f05, f06 int
	}{
		{"Flying Cloak 讓 Human 會飛", RaceHuman, 0x83, 2, 0, 0, 0},
		{"Courage Coat 免疫怯戰", RaceHuman, 0x82, 1, 0, 0, 1},
		{"Shade Ring", RaceHuman, 0x81, 1, 0, 1, 0},
		{"Speed Boots 補上 Dwarven 的衝刺", RaceDwarven, 0x80, 1, 1, 0, 0},
		{"Shoes 不動 Kelden 的 2", RaceKelden, 0xE0, 2, 1, 0, 0},
		{"種類 byte 不大於 0x7F 不算", RaceDwarven, 0x03, 0, 0, 0, 0},
	} {
		a := Actor{Flag03: MoveAbility(c.race)}
		a.ApplyCharms([4]int{c.class})
		if a.Flag03 != c.f03 || a.Flag04 != c.f04 || a.Flag05 != c.f05 || a.Flag06 != c.f06 {
			t.Errorf("%s：+0x03…+0x06 = %d %d %d %d，應為 %d %d %d %d", c.name,
				a.Flag03, a.Flag04, a.Flag05, a.Flag06, c.f03, c.f04, c.f05, c.f06)
		}
	}
}

// 第二輪多走的步數（`0x20927`…`0x20A48`），加上第一輪那一步就是說明書的格數。
func TestExtraSteps(t *testing.T) {
	for _, c := range []struct {
		style, flag04, want int
	}{
		{IconMoveWalk, 0, 0}, {IconMoveWalk, 1, 0},
		{IconMoveNext, 0, 0}, {IconMoveNext, 1, 1},
		{IconMoveFast, 0, 1}, {IconMoveFast, 1, 3},
		{IconMoveAlt1, 1, 0}, {IconMoveAlt2, 0, 1}, {IconMoveAlt3, 0, 2},
	} {
		a := Actor{Side: SideParty, Head: 1, Chest: 1, Command: CmdMove, Style: c.style, Flag04: c.flag04}
		if got := ExtraSteps(a); got != c.want {
			t.Errorf("方式 %#x、+0x04=%d：多走 %d，應為 %d", c.style, c.flag04, got, c.want)
		}
	}
	a := Actor{Side: SideParty, Head: 1, Chest: 1, Command: CmdRest, Style: IconMoveFast}
	if ExtraSteps(a) != 0 {
		t.Error("不是移動的指令不該多走")
	}
}

// 聰明的敵人看不到戴 Shade Ring 的隊員；只有碰到 +0x05 才需要擲智力骰。
func TestHiddenFromSmartEnemy(t *testing.T) {
	s := &State{Actors: []Actor{{Side: SideParty}, {Side: SideMonster, Intellect: 90, Level: 5}}}
	rng := &fixedRoll{values: []int{1}}
	if s.hiddenFrom(1, 0, rng) || rng.i != 0 {
		t.Fatal("沒戴 Shade Ring 不該隱形，也不該擲骰")
	}
	s.Actors[0].Flag05 = 1
	if !s.hiddenFrom(1, 0, rng) {
		t.Fatal("聰明的敵人看不到戴 Shade Ring 的隊員")
	}
	s.aiRoll = 0
	s.Actors[1].Intellect = 1
	if s.hiddenFrom(1, 0, &fixedRoll{values: []int{100}}) {
		t.Fatal("智力擲骰沒過的敵人照樣看得到")
	}
}

// 主迴圈跑完才輪到多走那一輪，順序同出手順序（`0x20927`）：穿 Speed Boots 衝刺
// 一共走四格，沒穿的走兩格。
func TestExtraStepsAfterMainPass(t *testing.T) {
	s := &State{Actors: make([]Actor, 2)}
	s.Actors[0] = Actor{Side: SideParty, X: 20, Y: 10, Facing: 2,
		Threshold: 100, Command: CmdMove, Style: IconMoveFast, Flag04: 1}
	s.Actors[1] = Actor{Side: SideParty, X: 30, Y: 10, Facing: 2,
		Threshold: 100, Command: CmdMove, Style: IconMoveFast}
	for i := range s.Actors {
		s.Actors[i].SetParts()
	}
	s.Exec.Order = []int{0, 1}
	var who []int
	for {
		r, ok := s.Act(&fixedRoll{values: []int{50}})
		if !ok {
			break
		}
		who = append(who, r.Actor)
	}
	want := []int{0, 1, 0, 0, 0, 1}
	if len(who) != len(want) {
		t.Fatalf("出手順序 %v，應為 %v", who, want)
	}
	for k := range want {
		if who[k] != want[k] {
			t.Fatalf("出手順序 %v，應為 %v", who, want)
		}
	}
	if s.Actors[0].Y != 14 || s.Actors[1].Y != 12 {
		t.Fatalf("應該走到欄 14 與 12，得到 %d 與 %d", s.Actors[0].Y, s.Actors[1].Y)
	}
}
