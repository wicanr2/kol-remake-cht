package battle

import "testing"

// 額度的式子：耐力 ＋ 耐力÷2 − 疲勞 − 累積傷害×10 ＋ (門檻 − 累積傷害)×2。
func TestTimeAllowance(t *testing.T) {
	a := Actor{Stamina: 200, Threshold: 100}
	if got := a.TimeAllowance(); got != 200+100+200 {
		t.Fatalf("沒受傷的額度 %d，期望 500", got)
	}
	// 傷得重額度就變負：300 − 500 + (100−50)×2 = −100。
	a.Damage = 50
	if got := a.TimeAllowance(); got != -100 {
		t.Fatalf("受傷後的額度 %d，期望 −100", got)
	}
	// 疲勞本身也扣額度，所以會自己加速。
	a.Fatigue = 40
	if got := a.TimeAllowance(); got != -140 {
		t.Fatalf("帶疲勞的額度 %d，期望 −140", got)
	}
}

// 額度為負時疲勞往上加——「受傷 → 疲勞 → 撤退」那條迴路的唯一入口。
func TestFatigueRisesWhenAllowanceIsNegative(t *testing.T) {
	s := &State{Actors: []Actor{
		{Side: SideParty, X: 5, Y: 5, Head: 10, Chest: 10, Threshold: 100,
			Command: CmdRest, Stamina: 200, Damage: 50, Hurt: 1},
	}}
	s.Exec.Order = []int{0}
	if _, ok := s.Act(&fixedRand{values: []int{1}}); !ok {
		t.Fatal("該做出一手")
	}
	// 額度 −100 − 挨傷 5 ＝ −105，疲勞 −= −105÷10 ＝ +10。
	if got := s.Actors[0].Fatigue; got != 10 {
		t.Fatalf("疲勞是 %d，期望 10", got)
	}
}

// 力竭：比例到 100 就當場倒下，三句話依隊員／怪物／碎掉分開。
func TestExhaustionKnocksDown(t *testing.T) {
	for _, c := range []struct {
		name     string
		side     Side
		shatters bool
		record   int
	}{
		{"隊員", SideParty, false, ExhaustedPartyRecord},
		{"怪物", SideMonster, false, ExhaustedMonsterRecord},
		{"碎掉", SideMonster, true, ExhaustedShattersRecord},
	} {
		s := &State{Actors: []Actor{
			{Side: c.side, X: 5, Y: 5, Head: 10, Chest: 10, Threshold: 100,
				Command: CmdRest, Stamina: 100, Fatigue: 400, Damage: 90,
				Shatters: c.shatters},
		}}
		s.Exec.Order = []int{0}
		out, ok := s.Act(&fixedRand{values: []int{1}})
		if !ok {
			t.Fatalf("%s：該做出一手", c.name)
		}
		if out.Exhausted != c.record {
			t.Fatalf("%s：力竭那一句是 %#x，期望 %#x", c.name, out.Exhausted, c.record)
		}
		if !s.Actors[0].Down() {
			t.Fatalf("%s：力竭該當場倒下", c.name)
		}
	}
	// 沒累到門檻就沒有那一句。
	s := &State{Actors: []Actor{
		{Side: SideParty, X: 5, Y: 5, Head: 10, Chest: 10, Threshold: 100,
			Command: CmdRest, Stamina: 200},
	}}
	s.Exec.Order = []int{0}
	if out, _ := s.Act(&fixedRand{values: []int{1}}); out.Exhausted != 0 {
		t.Fatalf("不該力竭，得到 %#x", out.Exhausted)
	}
}

// 施法的額度花費：四格護符挑類別最貴的那一格，加種、效果與骰式。
func TestCastTime(t *testing.T) {
	// 類別 1（200）最貴，種 3；效果 1 是 120；強度碼 0x02 → 骰數 (0x02>>4)&7 ＝ 0
	// 加一 ＝ 1，骰面碼 (0x02>>1)&7 ＝ 1，查表用碼加一 ＝ diceCastTime[2] ＝ 70。
	// 所以 200 + 3×16 + 120 + 1×70 ＝ 438。
	sp := Spell{Effect: 1, PowerCode: 0x02, Guards: [4]int{0x13, 0x00, 0x00, 0x00}}
	if got := CastTime(sp); got != 200+48+120+70 {
		t.Fatalf("施法成本 %d，期望 438", got)
	}
	// 要選目標格（bit 7）就把骰式那一項加倍。
	sp.PowerCode = 0x82
	if got := CastTime(sp); got != 200+48+120+140 {
		t.Fatalf("要選目標格的成本 %d，期望 508", got)
	}
	// 骰面碼 7（99 面）查到的是表尾之後那一格——原版讀出界，值是 0x46。
	sp = Spell{Effect: 1, PowerCode: 0x0E}
	if got := CastTime(sp); got != guardCastTime[0]+effectCastTime[1]+0x46 {
		t.Fatalf("99 面的成本 %d", got)
	}
}

// 花費與夾限：`sub_251B0` 是兩個乘積相加，`sub_2526C` 把比例夾進 1/3…1。
func TestActionCostAndClampEffort(t *testing.T) {
	a := Actor{WeaponSpeed: 20, Encumbrance: 50}
	if got := ActionCost(a, 15, 1); got != 15*20+50 {
		t.Fatalf("花費 %d，期望 350", got)
	}
	// 額度比花費大 → 直接當 1。
	if n, d := ClampEffort(1000, 350); n != 1 || d != 1 {
		t.Fatalf("夾成 %d/%d，期望 1/1", n, d)
	}
	// 花費超過額度三倍 → 1/3。
	if n, d := ClampEffort(100, 350); n != 1 || d != 3 {
		t.Fatalf("夾成 %d/%d，期望 1/3", n, d)
	}
	// 額度是負的也是 1/3。
	if n, d := ClampEffort(-40, 350); n != 1 || d != 3 {
		t.Fatalf("夾成 %d/%d，期望 1/3", n, d)
	}
	// 落在區間內就原樣回。
	if n, d := ClampEffort(200, 350); n != 200 || d != 350 {
		t.Fatalf("夾成 %d/%d，期望 200/350", n, d)
	}
}

// 動作越貴、額度越少，命中部位越偏向索引 0、1、6 那三項。
func TestEnemyLocationFavoursSlotsWhenSlow(t *testing.T) {
	fast := Actor{Level: 3, Stamina: 400, Threshold: 200}
	slow := Actor{Level: 3, Stamina: 20, Threshold: 10,
		WeaponSpeed: 40, Encumbrance: 200}
	hit := func(a Actor) [7]int {
		var n [7]int
		for r := 1; r <= 256; r++ {
			n[EnemyLocation(a, 2, &fixedRand{values: []int{r}})]++
		}
		return n
	}
	f, s := hit(fast), hit(slow)
	favoured := func(n [7]int) int { return n[0] + n[1] + n[6] }
	if favoured(s) <= favoured(f) {
		t.Fatalf("慢的那一邊該更偏向那三項：%v vs %v", s, f)
	}
}
