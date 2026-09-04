package battle

import "testing"

// 骰式的拆法：先右移一位再取低六位，低三位查骰面表、整段除以 8 得骰數。
func TestSpellDice(t *testing.T) {
	for _, c := range []struct{ code, count, sides int }{
		// 0x02 → v=1 → 1 顆 3 面
		{0x02, 1, 3},
		// 0x10 → v=8 → 2 顆 2 面
		{0x10, 2, 2},
		// 0x8E → v=0x47&0x3F=7 → 1 顆 99 面（bit 7 是「要選目標格」，不影響骰式）
		{0x8E, 1, 99},
		// 0x7E → v=0x3F → 8 顆 99 面
		{0x7E, 8, 99},
	} {
		count, sides := SpellDice(c.code)
		if count != c.count || sides != c.sides {
			t.Fatalf("碼 %#02x 拆成 %d 顆 %d 面，期望 %d 顆 %d 面",
				c.code, count, sides, c.count, c.sides)
		}
	}
	// 與武器傷害碼那一套不同：同一個碼拆出來的骰數與骰面都不一樣。
	wc, ws := DamageCode(0x10)
	sc, ss := SpellDice(0x10)
	if wc == sc && ws == ss {
		t.Fatal("法術與武器的拆法應該不同（差一個位移）")
	}
}

// bit 7 表示這一手要先選一個目標格。
func TestSpellNeedsTarget(t *testing.T) {
	if (Spell{PowerCode: 0x0E}).NeedsTarget() {
		t.Fatal("bit 7 沒設不該要選目標格")
	}
	if !(Spell{PowerCode: 0x8E}).NeedsTarget() {
		t.Fatal("bit 7 設了就要選目標格")
	}
}

// 擋招：倒下的一定擋不住；高 nibble 要相同，那一格 ≥ 0x50 時低 nibble 還要夠大。
func TestSpellBlocked(t *testing.T) {
	sp := Spell{Guards: [4]int{0x31, 0x55, 0x00, 0x00}}
	if SpellBlocked(sp, 0xFF, true) {
		t.Fatal("倒下的目標一定擋不住")
	}
	// 高 nibble 3 對上 0x31，低 nibble 不看（0x30 < 0x50）。
	if SpellBlocked(sp, 0x3F, false) {
		t.Fatal("高 nibble 相同就該穿得過去")
	}
	// 高 nibble 5 對上 0x55，但目標的低 nibble 是 7 > 5，擋下來。
	if !SpellBlocked(sp, 0x57, false) {
		t.Fatal("那一格 ≥ 0x50 時低 nibble 不夠大就該擋下來")
	}
	// 低 nibble 剛好相等就過得去。
	if SpellBlocked(sp, 0x55, false) {
		t.Fatal("低 nibble 相等該穿得過去")
	}
	// 高 nibble 一格都對不上。
	if !SpellBlocked(sp, 0x71, false) {
		t.Fatal("高 nibble 都不同就該擋下來")
	}
}

// 強度形容詞：強度除以 8，上限 3。
func TestStrengthWord(t *testing.T) {
	for _, c := range []struct{ power, want int }{
		{0, MildlyRecord}, {7, MildlyRecord}, {8, MildlyRecord + 1},
		{16, MildlyRecord + 2}, {24, MildlyRecord + 3}, {99, MildlyRecord + 3},
	} {
		if got := StrengthWord(c.power); got != c.want {
			t.Fatalf("強度 %d 用記錄 %#x，期望 %#x", c.power, got, c.want)
		}
	}
	if DirectionWord(true) != NegativeRecord || DirectionWord(false) != PositiveRecord {
		t.Fatal("正負那一格對錯了")
	}
}

func castState(sp Spell) *State {
	s := &State{Party: []Member{{Name: "POONTANG", Spells: []Spell{sp}}}}
	s.Actors = []Actor{
		{Side: SideParty, Member: 0, Head: 10, Chest: 10, Threshold: 100},
		{Side: SideMonster, Member: -1, Head: 10, Chest: 10, Threshold: 100,
			Strength: 100, Quickness: 100, Foresight: 100, Intellect: 100},
	}
	s.Actors[1].SetParts()
	return s
}

// 目標格上沒有人就什麼都不做。
func TestCastNothingThere(t *testing.T) {
	s := castState(Spell{Effect: 1, TurnCode: 0x02, PowerCode: 0x02})
	if got := s.Cast(0, 9, -1, maxRand{}); got.Verdict != SpellNothingThere {
		t.Fatalf("超出範圍的目標該回 SpellNothingThere，得到 %+v", got)
	}
	// 沒有法術可施也一樣。
	empty := &State{Party: []Member{{Name: "X"}}, Actors: []Actor{
		{Side: SideParty, Member: 0}, {Side: SideMonster, Member: -1}}}
	if got := empty.Cast(0, 1, -1, maxRand{}); got.Verdict != SpellNothingThere {
		t.Fatalf("沒有法術該回 SpellNothingThere，得到 %+v", got)
	}
}

// 擋下來就只印 0x16A，不擲骰也不改任何東西。
func TestCastBlocked(t *testing.T) {
	s := castState(Spell{Effect: 1, Guards: [4]int{0x31, 0, 0, 0},
		TurnCode: 0x02, PowerCode: 0x02})
	s.Actors[1].Ward = 0x71 // 高 nibble 對不上
	before := s.Actors[1].Strength
	got := s.Cast(0, 1, -1, maxRand{})
	if got.Verdict != SpellNoEffect {
		t.Fatalf("該被擋下來，得到 %+v", got)
	}
	if s.Actors[1].Strength != before {
		t.Fatal("擋下來不該改屬性")
	}
}

// 效果 1：力量。負面往下扣、正面往上加並掛倒數。
func TestCastStatEffect(t *testing.T) {
	// 持續 1 顆 3 面、強度 1 顆 3 面；maxRand 兩擲都是滿分。
	sp := Spell{Effect: 1, TurnCode: 0x02, PowerCode: 0x02}
	s := castState(sp)
	s.Actors[0].Negative = true
	got := s.Cast(0, 1, -1, maxRand{})
	if got.Verdict != SpellHits || got.Turns != 3 || got.Power != 3 {
		t.Fatalf("結果是 %+v，期望 3 回合 3 強度", got)
	}
	if s.Actors[1].Strength != 97 {
		t.Fatalf("力量變成 %d，期望 97", s.Actors[1].Strength)
	}
	if s.Actors[1].SpellTurns != 4 {
		t.Fatalf("倒數是 %d，期望 4（持續 ＋ 1）", s.Actors[1].SpellTurns)
	}

	// 正面：往上加，上限 255。
	s = castState(sp)
	s.Actors[1].Strength = 254
	if got := s.Cast(0, 1, -1, maxRand{}); got.Verdict != SpellHits {
		t.Fatalf("該打中，得到 %+v", got)
	}
	if s.Actors[1].Strength != 255 {
		t.Fatalf("力量變成 %d，期望夾在 255", s.Actors[1].Strength)
	}
}

// 效果 4 動的是智力，效果 7／8 那兩格語意未解，什麼都不該動。
func TestCastStatFields(t *testing.T) {
	s := castState(Spell{Effect: 4, TurnCode: 0x02, PowerCode: 0x02})
	s.Cast(0, 1, -1, maxRand{})
	if s.Actors[1].Intellect != 103 {
		t.Fatalf("智力變成 %d，期望 103", s.Actors[1].Intellect)
	}
	for _, effect := range []int{7, 8} {
		s := castState(Spell{Effect: effect, TurnCode: 0x02, PowerCode: 0x02})
		before := s.Actors[1]
		if got := s.Cast(0, 1, -1, maxRand{}); got.Verdict != SpellHits {
			t.Fatalf("效果 %d 該算打中，得到 %+v", effect, got)
		}
		a := s.Actors[1]
		if a.Strength != before.Strength || a.Quickness != before.Quickness ||
			a.Foresight != before.Foresight || a.Intellect != before.Intellect {
			t.Fatalf("效果 %d 的欄位還沒解，不該動任何屬性", effect)
		}
	}
}

// 效果 6：指定部位。負面走傷害那一套，正面補到那個部位的上限為止。
func TestCastAreaEffect(t *testing.T) {
	sp := Spell{Effect: SpellEffectArea, TurnCode: 0x02, PowerCode: 0x02}
	s := castState(sp)
	s.Actors[0].Negative = true
	got := s.Cast(0, 1, 2, maxRand{}) // 右臂
	if got.Verdict != SpellHits || got.Area != 2 {
		t.Fatalf("結果是 %+v，期望打中右臂", got)
	}
	caps := Parts(s.Actors[1].Threshold)
	if s.Actors[1].RightArm != caps[2]-3 {
		t.Fatalf("右臂剩 %d，期望 %d", s.Actors[1].RightArm, caps[2]-3)
	}

	// 正面：補回上限，而且不掛倒數（部位那一條不是持續效果）。
	s = castState(sp)
	s.Actors[1].RightArm = 0
	s.Actors[1].SpellTurns = 9
	s.Cast(0, 1, 2, maxRand{})
	if s.Actors[1].RightArm != 3 {
		t.Fatalf("右臂補成 %d，期望 3", s.Actors[1].RightArm)
	}
	if s.Actors[1].SpellTurns != 0 {
		t.Fatalf("倒數是 %d，部位那一條該把它設回 0", s.Actors[1].SpellTurns)
	}
	// 補不會超過上限。
	s = castState(Spell{Effect: SpellEffectArea, TurnCode: 0x02, PowerCode: 0x7E})
	s.Cast(0, 1, 2, maxRand{})
	if s.Actors[1].RightArm != caps[2] {
		t.Fatalf("右臂補成 %d，期望夾在上限 %d", s.Actors[1].RightArm, caps[2])
	}
}

// 效果 5 改的那個全域還沒解，這一版什麼都不做但仍算打中。
func TestCastGlobalEffectDoesNothingYet(t *testing.T) {
	s := castState(Spell{Effect: SpellEffectGlobal, TurnCode: 0x02, PowerCode: 0x02})
	before := s.Actors[1]
	got := s.Cast(0, 1, -1, maxRand{})
	if got.Verdict != SpellHits {
		t.Fatalf("該算打中，得到 %+v", got)
	}
	if s.Actors[1] != before {
		t.Fatal("效果 5 的語意還沒解，不該改任何東西")
	}
}

// 倒數到 0 就把屬性整批重算回原值，並回報「這一回合失效」。
func TestWearsOffRestoresStats(t *testing.T) {
	s := castState(Spell{Effect: 1, TurnCode: 0x02, PowerCode: 0x02})
	s.Monster = Monster{Strength: 100, Quickness: 100, Foresight: 100, Intellect: 100}
	s.Actors[0].Negative = true
	s.Cast(0, 1, -1, maxRand{})
	if s.Actors[1].Strength != 97 || s.Actors[1].SpellTurns != 4 {
		t.Fatalf("施法之後是 %d 力量、%d 回合", s.Actors[1].Strength, s.Actors[1].SpellTurns)
	}
	// 前三回合只是倒數。
	for n := 0; n < 3; n++ {
		if s.WearsOff(1) {
			t.Fatalf("第 %d 回合不該失效", n+1)
		}
		if s.Actors[1].Strength != 97 {
			t.Fatal("還沒失效不該改屬性")
		}
	}
	// 第四回合失效，力量回到怪物記錄的原值。
	if !s.WearsOff(1) {
		t.Fatal("第四回合該失效")
	}
	if s.Actors[1].Strength != 100 {
		t.Fatalf("失效之後力量是 %d，期望重算回 100", s.Actors[1].Strength)
	}
	// 失效之後再叫不會有事。
	if s.WearsOff(1) {
		t.Fatal("已經失效的不該再回報一次")
	}
}

// 隊員那一側重算會把營養折算算進去。
func TestWearsOffUsesNourishedStrength(t *testing.T) {
	s := castState(Spell{Effect: 1, TurnCode: 0x02, PowerCode: 0x02})
	s.Party[0].Strength, s.Party[0].Nutrition = 200, 255
	s.Actors[0].SpellTurns = 1
	s.Actors[0].Strength = 999
	if !s.WearsOff(0) {
		t.Fatal("該失效")
	}
	if got, want := s.Actors[0].Strength, Nourished(200, 255); got != want {
		t.Fatalf("力量重算成 %d，期望 %d", got, want)
	}
}

// 回合開始時先跑倒數，過期的記進 WoreOff 讓呼叫端印記錄 0x173。
func TestBeginPlanningRunsTheSpellCountdown(t *testing.T) {
	s := castState(Spell{Effect: 1, TurnCode: 0x02, PowerCode: 0x02})
	s.Monster = Monster{Strength: 100, Quickness: 100, Foresight: 100, Intellect: 100}
	s.Actors[1].SpellTurns = 1
	s.Actors[1].Strength = 1
	s.BeginPlanning(maxRand{})
	if len(s.WoreOff) != 1 || s.WoreOff[0] != 1 {
		t.Fatalf("失效清單是 %v，期望 [1]", s.WoreOff)
	}
	if s.Actors[1].Strength != 100 {
		t.Fatalf("屬性沒有重算，力量是 %d", s.Actors[1].Strength)
	}
	// 下一回合沒有人失效。
	s.BeginPlanning(maxRand{})
	if len(s.WoreOff) != 0 {
		t.Fatalf("不該再有失效，得到 %v", s.WoreOff)
	}
}

// 規劃段選施法：先出法術清單，選完問正負，再去選目標格（RE-100）。
func TestPlanningSpellFlow(t *testing.T) {
	sp := Spell{Name: "Vartwelyrmi", Effect: 5, TurnCode: 0, PowerCode: 0x8A}
	s := &State{Party: []Member{{Name: "WARDER", Spells: []Spell{sp}, PackCount: 1}}}
	s.Actors = []Actor{
		{Side: SideParty, Member: 0, Head: 10, Chest: 10, Threshold: 100, X: 30, Y: 30},
		{Side: SideMonster, Member: -1, Head: 10, Chest: 10, Threshold: 100, X: 31, Y: 30},
	}
	s.Field = NewField()
	s.Exec.Order, s.Exec.Turn = []int{0}, 0
	s.Round.Order, s.Round.Turn = []int{0}, 0
	s.Plan.Bar = []int{CmdYes, CmdCast}
	s.Plan.Selected = 1

	if s.Choose() {
		t.Fatal("選施法不該結束這一輪")
	}
	if s.Plan.Prompt != PromptSpell {
		t.Fatalf("該進法術清單，得到 %d", s.Plan.Prompt)
	}
	if len(s.Plan.List) != 1 || s.Plan.List[0] != "Vartwelyrmi" {
		t.Fatalf("清單是 %v", s.Plan.List)
	}

	// 選中第 0 個法術，接著問正負。
	s.Choose()
	if s.Plan.Prompt != PromptAffect {
		t.Fatalf("該進正負那一張，得到 %d", s.Plan.Prompt)
	}
	if s.Actors[0].SpellPick != 0 {
		t.Fatalf("選中的法術是 %d", s.Actors[0].SpellPick)
	}
	if len(AffectRecords()) != 2 || AffectRecords()[0] != PositiveRecord {
		t.Fatalf("正負那兩項的記錄是 %v", AffectRecords())
	}

	// 游標移到第二項（Negative）再確認，然後去選目標格。
	s.MoveSelection(1)
	if s.Plan.ListPick != 1 {
		t.Fatalf("游標是 %d，期望 1", s.Plan.ListPick)
	}
	s.Choose()
	if !s.Actors[0].Negative {
		t.Fatal("選了第二項該是負面")
	}
	if s.Plan.Prompt != PromptTarget {
		t.Fatalf("該進選目標格，得到 %d", s.Plan.Prompt)
	}
	// 這個法術的 bit 7 有設，用的是遠距游標。
	if !sp.NeedsTarget() {
		t.Fatal("這一個法術該要選目標格")
	}
}
