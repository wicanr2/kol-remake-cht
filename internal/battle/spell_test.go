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

// 擋招：倒下的目標一律沒有效果（RE-163）；高 nibble 要相同，那一格 ≥ 0x50 時低 nibble
// 還要夠大。
func TestSpellBlocked(t *testing.T) {
	sp := Spell{Guards: [4]int{0x31, 0x55, 0x00, 0x00}}
	if !SpellBlocked(sp, 0x31, true) {
		t.Fatal("倒下的目標，護符對得上也沒有效果")
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

// 效果 4 動的是智力；效果 7／8 動的是攻防技能的罰值，四個屬性都不該動。
func TestCastStatFields(t *testing.T) {
	s := castState(Spell{Effect: 4, TurnCode: 0x02, PowerCode: 0x02})
	s.Cast(0, 1, -1, maxRand{})
	if s.Actors[1].Intellect != 103 {
		t.Fatalf("智力變成 %d，期望 103", s.Actors[1].Intellect)
	}
	for _, c := range []struct {
		effect int
		field  func(Actor) int
	}{
		{7, func(a Actor) int { return a.SkillPenaltyOff }},
		{8, func(a Actor) int { return a.SkillPenaltyDef }},
	} {
		s := castState(Spell{Effect: c.effect, TurnCode: 0x02, PowerCode: 0x02})
		before := s.Actors[1]
		if got := s.Cast(0, 1, -1, maxRand{}); got.Verdict != SpellHits {
			t.Fatalf("效果 %d 該算打中，得到 %+v", c.effect, got)
		}
		a := s.Actors[1]
		if a.Strength != before.Strength || a.Quickness != before.Quickness ||
			a.Foresight != before.Foresight || a.Intellect != before.Intellect {
			t.Fatalf("效果 %d 動的是罰值，不該碰四個屬性", c.effect)
		}
		if got := c.field(a); got != 3 {
			t.Fatalf("效果 %d 的罰值是 %d，期望 3", c.effect, got)
		}
	}
}

// 罰值折算成技能：一點抵十，扣到 0 為止，而且過期不還。
func TestSkillPenalty(t *testing.T) {
	a := Actor{Attack: 230, Parry: 100}
	if a.OffenceSkill() != 230 || a.DefenceSkill() != 100 {
		t.Fatalf("沒有罰值時該是原值，得到 %d／%d",
			a.OffenceSkill(), a.DefenceSkill())
	}
	a.SkillPenaltyOff, a.SkillPenaltyDef = 3, 3
	if a.OffenceSkill() != 200 {
		t.Fatalf("攻擊技能是 %d，期望 230 − 30", a.OffenceSkill())
	}
	if a.DefenceSkill() != 70 {
		t.Fatalf("防禦技能是 %d，期望 100 − 30", a.DefenceSkill())
	}
	// 扣過頭夾在 0，不會變成負的。
	a.SkillPenaltyDef = 99
	if a.DefenceSkill() != 0 {
		t.Fatalf("防禦技能是 %d，期望夾在 0", a.DefenceSkill())
	}
	// 攻擊分數跟著少：`sub_26D76:0x26E1C` 加的是折算後的那一個。
	b := Actor{Strength: 100, Attack: 230, Threshold: 100}
	b.SetParts()
	c := b
	c.SkillPenaltyOff = 3
	if b.AttackValue()-c.AttackValue() != 30 {
		t.Fatalf("攻擊值只差 %d，期望 30", b.AttackValue()-c.AttackValue())
	}
}

// 原版沒有為罰值那兩格把正負反過來：正面的法術反而把對方的技能扣掉
// （`0x2878E` 對六個欄位一視同仁），負面的法術是把罰值減回來。
func TestCastSkillPenaltyDirection(t *testing.T) {
	// 正面：罰值往上加 → 技能被扣。
	s := castState(Spell{Effect: 7, TurnCode: 0x02, PowerCode: 0x02})
	s.Actors[1].Attack = 230
	s.Cast(0, 1, -1, maxRand{})
	if s.Actors[1].OffenceSkill() != 200 {
		t.Fatalf("正面之後攻擊技能是 %d，期望 200", s.Actors[1].OffenceSkill())
	}
	// 負面：往下減，罰值本來是 0 就什麼都不會發生。
	s = castState(Spell{Effect: 7, TurnCode: 0x02, PowerCode: 0x02})
	s.Actors[0].Negative = true
	s.Actors[1].Attack = 230
	s.Cast(0, 1, -1, maxRand{})
	if s.Actors[1].SkillPenaltyOff != 0 {
		t.Fatalf("罰值是 %d，期望仍然夾在 0", s.Actors[1].SkillPenaltyOff)
	}
	// 罰值不隨法術過期還回來：`sub_2198A`／`sub_2215A` 只寫 +0x07…+0x0C。
	s = castState(Spell{Effect: 7, TurnCode: 0x02, PowerCode: 0x02})
	s.Actors[1].Attack = 230
	s.Cast(0, 1, -1, maxRand{})
	for i := 0; i < 8 && s.Actors[1].SpellTurns > 0; i++ {
		s.WearsOff(1)
	}
	if s.Actors[1].SkillPenaltyOff != 3 {
		t.Fatalf("過期之後罰值變成 %d，期望還是 3", s.Actors[1].SkillPenaltyOff)
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

// 效果 5 動的是疲勞，而且正負是反過來的：負面加、正面減。
func TestCastFatigueEffect(t *testing.T) {
	// 強度 1 顆 3 面，maxRand 擲滿 → 3；加的是兩倍。
	s := castState(Spell{Effect: SpellEffectFatigue, TurnCode: 0x02, PowerCode: 0x02})
	s.Actors[0].Negative = true
	s.Actors[1].Stamina = 200
	got := s.Cast(0, 1, -1, maxRand{})
	if got.Verdict != SpellHits {
		t.Fatalf("該算打中，得到 %+v", got)
	}
	if s.Actors[1].Fatigue != 6 {
		t.Fatalf("疲勞是 %d，期望 6", s.Actors[1].Fatigue)
	}
	// 比例 = 50 × 6 ÷ 200 = 1。
	if r := s.Actors[1].FatigueRatio(); r != 1 {
		t.Fatalf("疲勞比例是 %d，期望 1", r)
	}
	// 這一條不掛倒數——它走自己的分支，不碰 +0x30。
	if s.Actors[1].SpellTurns != 0 {
		t.Fatalf("效果 5 不該掛倒數，得到 %d", s.Actors[1].SpellTurns)
	}

	// 正面往下減，而且原版的下限是死碼，所以減得到負數。
	s = castState(Spell{Effect: SpellEffectFatigue, TurnCode: 0x02, PowerCode: 0x02})
	s.Actors[1].Stamina = 200
	s.Cast(0, 1, -1, maxRand{})
	if s.Actors[1].Fatigue != -6 {
		t.Fatalf("疲勞是 %d，期望 −6（原版下限那一段是死碼）", s.Actors[1].Fatigue)
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

// 休息一手把疲勞減成七成，而且這一手挨過打就不算（`0x20AE0`）。
func TestRestRecoversFatigue(t *testing.T) {
	s := &State{Actors: []Actor{
		{Side: SideParty, X: 5, Y: 5, Head: 10, Chest: 10, Threshold: 100,
			Command: CmdRest, Stamina: 200, Fatigue: 100},
	}}
	s.Exec.Order = []int{0}
	drainAct(s, &fixedRand{values: []int{1}})
	if got := s.Actors[0].Fatigue; got != 70 {
		t.Fatalf("疲勞是 %d，期望 70", got)
	}
	// 這一手挨過打就走另一條：算時間額度再扣。
	// 額度 ＝ 200 ＋ 100 − 100 − 0 ＋ (100 − 0) × 2 ＝ 400，挨的傷扣 3 × 5 ＝ 15，
	// 休息的指令成本是 0，所以剩 385；疲勞 −= 38。
	s = &State{Actors: []Actor{
		{Side: SideParty, X: 5, Y: 5, Head: 10, Chest: 10, Threshold: 100,
			Command: CmdRest, Stamina: 200, Fatigue: 100, Hurt: 3},
	}}
	s.Exec.Order = []int{0}
	drainAct(s, &fixedRand{values: []int{1}})
	if got := s.Actors[0].Fatigue; got != 62 {
		t.Fatalf("挨過打該走額度那一條，疲勞是 %d，期望 62", got)
	}
}

// 效果 6 分兩條：要選目標格的擲隨機部位，不用選的停下來問玩家。
func TestCastAreaAsksWhenNoTargetCell(t *testing.T) {
	// PowerCode 0x02（bit 7 沒設）→ 不用選目標格 → 要問。
	s := castState(Spell{Effect: SpellEffectArea, TurnCode: 0x02, PowerCode: 0x02})
	s.Actors[0].Negative = true
	s.Actors[0].Command = CmdCast
	s.Actors[0].Facing = 0
	// ring0[0] 是 {−1, 0}：面前那一格在自己上方一列。
	s.Actors[0].X, s.Actors[0].Y = 5, 5
	s.Actors[1].X, s.Actors[1].Y = 4, 5
	s.Exec.Order = []int{0}
	out, ok := s.Act(maxRand{})
	if !ok {
		t.Fatal("該做出一手")
	}
	if out.Cast == nil || !out.Cast.NeedsArea {
		t.Fatalf("該停下來問部位，得到 %+v", out.Cast)
	}
	if s.Plan.Prompt != PromptArea {
		t.Fatalf("提示層是 %d，期望 PromptArea", s.Plan.Prompt)
	}
	// 骰已經擲完：強度是 3（1 顆 3 面擲滿）。
	if out.Cast.Power != 3 {
		t.Fatalf("強度 %d，期望 3", out.Cast.Power)
	}
	// 這時候還沒有人受傷。
	before := s.Actors[1].RightArm
	// 超出範圍的鍵不算數，也不離開那一層。
	if _, _, ok := s.AnswerArea(9); ok || s.Plan.Prompt != PromptArea {
		t.Fatal("超出範圍的部位不該收")
	}
	res, caster, ok := s.AnswerArea(2) // 右臂
	if !ok || caster != 0 || res.Area != 2 {
		t.Fatalf("回答之後是 %+v／施法者 %d", res, caster)
	}
	if s.Plan.Prompt != PromptArea && s.Actors[1].RightArm != before-3 {
		t.Fatalf("右臂剩 %d，期望 %d", s.Actors[1].RightArm, before-3)
	}
}

// 要選目標格的那一種不問，直接擲。
func TestCastAreaRollsWhenTargetCellIsPicked(t *testing.T) {
	s := castState(Spell{Effect: SpellEffectArea, TurnCode: 0x02, PowerCode: 0x8E})
	got := s.Cast(0, 1, 3, maxRand{})
	if got.NeedsArea {
		t.Fatal("要選目標格的法術不該問部位")
	}
	if got.Area != 3 {
		t.Fatalf("部位 %d，期望 3", got.Area)
	}
}

// 問句那三筆記錄：正面是 heal、負面是 hurt。
func TestAreaPromptRecords(t *testing.T) {
	if got := AreaPrompt(true); got[0] != SelectAreaRecord ||
		got[1] != HurtRecord || got[2] != AreaListRecord {
		t.Fatalf("負面那三筆是 %v", got)
	}
	if got := AreaPrompt(false); got[1] != HealRecord {
		t.Fatalf("正面該是 heal，得到 %#x", got[1])
	}
}
