package battle

import (
	"encoding/json"
	"testing"
)

// RE-197：四式徒手共用未修飾分數與第1列部位；全零槽仍擲1d2。
func TestResolveUnarmedOriginalBranches(t *testing.T) {
	for _, style := range []int{1, 2, 3, 4} {
		a := Actor{Command: CmdUnarmed, Strength: 128, Attack: 150, Stamina: 200}
		d := Actor{Guard: 2, Stamina: 200, ShieldGrade: NoShield,
			Threshold: 200, Head: 100, Chest: 100, RightArm: 100, LeftArm: 100}
		r := &expectedAttackRoll{t: t, faces: []int{100, 100, 2, 100, 2}, values: []int{90, 1, 2, 20, 1}}
		out := Resolve(&a, &d, style, r)
		if out.Miss || out.Parried || out.Damage != 2 || out.Location != 1 || out.OverCap != 11 || d.Hurt != 3 || d.Chest != 98 || r.i != 5 {
			t.Fatalf("徒手式%d原版分支不符：%+v，目標%+v，擲骰%d", style, out, d, r.i)
		}
	}
}

func TestResolveUnarmedMissDoesNotRollStyle(t *testing.T) {
	a := Actor{Command: CmdUnarmed, Strength: 128, Attack: 150, Encumbrance: 2000, Stamina: 200}
	d := Actor{Guard: 2, Stamina: 200}
	r := &expectedAttackRoll{t: t, faces: []int{100}, values: []int{90}}
	out := Resolve(&a, &d, 1, r)
	if !out.Miss || out.Margin != 82 || r.i != 1 || d.Parries != 0 {
		t.Fatalf("徒手落空不應修飾或招架：%+v，擲骰%d", out, r.i)
	}
}

func TestUnarmedBrokenArmourRetainsZeroCodeDie(t *testing.T) {
	a := Actor{Command: CmdUnarmed, LoadLevel: 1, ShieldGrade: NoShield, ArmourWeight: 5}
	a.Armour[0] = &Item{Encumbrance: 5}
	a.UnarmedWeapon[3] = Weapon{Base: 4, Count: 2, Sides: 5, Weight: 5}
	a.breakArmour(0)
	if got := a.AttackWeapon(3); got != (Weapon{Count: 1, Sides: 2}) {
		t.Fatalf("破甲後全零槽傷害碼不可漏擲：%+v", got)
	}
}

// seqRand 依序回傳指定的擲骰結果，用完就從頭再來。
type seqRand struct {
	values []int
	i      int
}

func (r *seqRand) Roll(n int) int {
	v := r.values[r.i%len(r.values)]
	r.i++
	if v > n {
		return n
	}
	return v
}

// 傷害碼：高位是骰數、中間三位查骰面表。
func TestDamageCode(t *testing.T) {
	count, sides := DamageCode(0x33)
	if count != 2 || sides != 9 {
		t.Fatalf("0x33 解出 %dd%d", count, sides)
	}
	if c, s := DamageCode(0); c != 1 || s != 2 {
		t.Fatalf("0 解出 %dd%d", c, s)
	}
}

// 負重：係數隨等級遞減，武器只算一半重量。
func TestEncumbrance(t *testing.T) {
	w := Weapon{Weight: 100}
	enc, speed := Encumbrance(1, w, 60, 0)
	// 係數 70：速度 70、重量 50 + 60 = 110 → 77
	if speed != 70 || enc != 77 {
		t.Fatalf("負重 %d 速度 %d", enc, speed)
	}
	if enc, _ := Encumbrance(3, w, 60, 40); enc != 60*150/100 {
		t.Fatalf("三級帶盾的負重 %d", enc)
	}
}

// 飢餓折算力量：0 是全力、255 剩一半。
func TestNourished(t *testing.T) {
	if got := Nourished(100, 0); got != 100 {
		t.Fatalf("不餓時 %d", got)
	}
	if got := Nourished(100, 255); got != 50 {
		t.Fatalf("餓到底時 %d", got)
	}
}

func testAttacker() Actor {
	a := Actor{Side: SideParty, Level: 2, Strength: 120, Quickness: 90,
		Attack: 300, Parry: 200, Threshold: 100, ShieldGrade: NoShield,
		Weapon: Weapon{Kind: 0x9E, Base: 2, Count: 2, Sides: 9, Weight: 75}}
	a.SetParts()
	return a
}

func testDefender() Actor {
	d := Actor{Side: SideMonster, Level: 3, Strength: 130, Quickness: 120,
		Attack: 180, Parry: 36, Threshold: 105, ShieldGrade: NoShield, Guard: 1,
		Weapon: Weapon{Kind: 60, Count: 2, Sides: 5, Weight: 35}}
	d.SetParts()
	return d
}

// 攻擊值與防禦值都照 SPEC-009 的式子。
func TestAttackAndDefenceValue(t *testing.T) {
	a := testAttacker()
	a.Encumbrance = 50
	// 120 + 300 = 420，沒受傷所以不折損，減 0 減 0 再減負重 50 → 370
	if got := a.AttackValue(); got != 370 {
		t.Fatalf("攻擊值 %d", got)
	}
	// 累積傷害只扣線性的那一項（× 10）——原版沒有「傷得越重乘法惩罰」那一條，
	// 乘法項吃的是疲勞（RE-130）。
	a.Damage = 50
	if got := a.AttackValue(); got != 420-50*10-50 {
		t.Fatalf("受傷後的攻擊值 %d", got)
	}
	// 疲勞才會打折：耐力 200、疲勞 100 → 比例 25。
	a.Stamina, a.Fatigue = 200, 100
	if got := a.AttackValue(); got != 420-420*25/200-50*10-50 {
		t.Fatalf("疲勞之後的攻擊值 %d", got)
	}
	d := testDefender()
	d.Encumbrance = 20
	// 36 ÷ 1 + 120 = 156，扣負重 20 → 136
	if got := d.DefenceValue(); got != 136 {
		t.Fatalf("防禦值 %d", got)
	}
	d.Parries = 2 // 招架第二次：技能減半
	if got := d.DefenceValue(); got != 36/2+120-20 {
		t.Fatalf("第二次招架的防禦值 %d", got)
	}
	d.ShieldGrade = 3
	if got := d.DefenceValue(); got != 36/2+120+100*3+200-20 {
		t.Fatalf("帶盾的防禦值 %d", got)
	}
}

// 攻擊分數是骰前值除以十加擲骰，再加選項表；武器沒有那個位元就砍半。
func TestAttackScore(t *testing.T) {
	a := testAttacker()
	a.Encumbrance = 50
	base := a.AttackValue()/10 + 40
	if got := AttackScore(a, 0, 0, 1, 40); got != base+optionModifier[1][0] {
		t.Fatalf("攻擊分數 %d", got)
	}
	// 選項 2 要武器種類含 0x10；0x9E 有那個位元，不砍半。
	if got := AttackScore(a, 2, 0, 3, 40); got != base+optionModifier[3][0] {
		t.Fatalf("選項 2 的攻擊分數 %d", got)
	}
	a.Weapon.Kind = 0x01 // 沒有 0x10 這個位元
	want := (base + optionModifier[3][0]) / 2
	if got := AttackScore(a, 2, 0, 3, 40); got != want {
		t.Fatalf("武器不合的攻擊分數 %d，預期 %d", got, want)
	}
}

// RE-225：原版完整12×7表的末三列，涵蓋合法高攻擊式與武器減半條件。
func TestHighAttackModifiers(t *testing.T) {
	rows := [3][7]int{
		{42, -38, -8, -35, -33, -20, -18},
		{55, -25, 5, -27, 15, -2, -10},
		{59, -21, 9, -18, 24, 7, -41},
	}
	for i, row := range rows {
		style := i + 8
		for guard, modifier := range row {
			a := Actor{Strength: 100, Weapon: Weapon{Kind: 4}}
			want := 110 + modifier
			if got := AttackScore(a, style, guard, style+1, 100); got != want {
				t.Fatalf("攻擊式%d防禦%d=%d，原版應為%d", style, guard, got, want)
			}
			a.Weapon.Kind = 0
			if got := AttackScore(a, style, guard, style+1, 100); got != want/2 {
				t.Fatalf("攻擊式%d武器減半=%d，原版應為%d", style, got, want/2)
			}
		}
	}
	// 修正後為負的分數不再減半。
	a := Actor{Encumbrance: 100, Weapon: Weapon{}}
	if got := AttackScore(a, 8, 3, 9, 1); got != -44 {
		t.Fatalf("負分數=%d，原版應為-44", got)
	}
}

func TestNormalRoundSevenHighAttack(t *testing.T) {
	a := Actor{Strength: 218, Attack: 50, Fatigue: 57, Stamina: 398, Encumbrance: 185, Command: CmdWeapon, Style: 8, Weapon: Weapon{Kind: 183, Base: 2, Count: 3, Sides: 9}, ShieldGrade: NoShield}
	d := Actor{Quickness: 100, Stamina: 230, Encumbrance: 259, Command: CmdUnarmed, Guard: 2, ShieldGrade: NoShield}
	if got := AttackScore(a, 8, 2, 9, 65); got != 64 {
		t.Fatalf("第七回合攻擊=%d，原版局部應為64", got)
	}
	r := &seqRand{values: []int{65, 86}}
	if out := Resolve(&a, &d, 8, r); !out.Parried || out.Damage != 0 || r.i != 2 {
		t.Fatalf("64 < 71應退縮避開，僅兩次骰：%+v，骰%d", out, r.i)
	}
}

// 不招架的人防禦分數是 0。
func TestDefenceScoreNoGuard(t *testing.T) {
	d := testDefender()
	d.Guard = 0
	if got := DefenceScore(d, 50); got != 0 {
		t.Fatalf("不招架卻有 %d", got)
	}
}

// 命中部位查累積機率表：第 0 列的 1…25 是頭、26…70 是胸。
func TestHitLocation(t *testing.T) {
	// 選項 2 用第 0 列。
	if got := HitLocation(2, 25); got != 0 {
		t.Fatalf("25 打到 %d", got)
	}
	if got := HitLocation(2, 26); got != 1 {
		t.Fatalf("26 打到 %d", got)
	}
	if got := HitLocation(2, 100); got != 3 {
		t.Fatalf("100 打到 %d", got)
	}
	// 選項 4 用第 2 列，偏腿。
	if got := HitLocation(4, 100); got != 5 {
		t.Fatalf("第 2 列的 100 打到 %d", got)
	}
}

// 傷害＝基礎 + 骰子 + 力量加成 + 攻擊式加成 + 餘裕加成，並夾在武器上限內。
func TestDamageRoll(t *testing.T) {
	a := testAttacker()
	rng := &seqRand{values: []int{1, 1}}
	// 力量加成 = (120 − 70) ÷ 8 = 6，再乘 (70 − 0) ÷ 100 = 4
	// 基礎 2 + 骰 1 + 1 + 加成 4 + 選項 0 + 餘裕 0 = 8
	if got := DamageRoll(a, 0, 0, rng); got != 8 {
		t.Fatalf("傷害 %d", got)
	}
	// 餘裕 60 → 表的第 3 格 = 4
	rng = &seqRand{values: []int{1, 1}}
	if got := DamageRoll(a, 0, 60, rng); got != 8+marginBonus[3] {
		t.Fatalf("有餘裕的傷害 %d", got)
	}
	// 骰滿也不能超過 2 × 9 + 1 + 2 = 21
	rng = &seqRand{values: []int{9, 9}}
	if got := DamageRoll(a, 0, 300, rng); got != 21 {
		t.Fatalf("上限之外的傷害 %d", got)
	}
}

// 一擊打下去：攻擊分數為負是落空，低於防禦是被擋，其餘就落到部位上。
func TestResolve(t *testing.T) {
	a, d := testAttacker(), testDefender()
	a.Encumbrance, d.Encumbrance = 50, 20
	// 擲骰順序：攻擊、防禦、傷害骰 ×2、部位、空槽護甲骰。
	rng := &seqRand{values: []int{90, 1, 1, 1, 20, 2}}
	out := Resolve(&a, &d, 0, rng)
	if out.Miss || out.Parried {
		t.Fatalf("應該打中：%+v", out)
	}
	if out.Location != 1 {
		t.Fatalf("部位 %d", out.Location)
	}
	if out.Damage != 14 || d.Damage != 14 || d.Hurt != 16 || rng.i != 6 {
		t.Fatalf("傷害沒記上去：%+v", d)
	}
	if d.Parries != 1 {
		t.Fatalf("招架次數 %d", d.Parries)
	}
	// 防禦分數高就擋下來
	a2, d2 := testAttacker(), testDefender()
	a2.Encumbrance = 400 // 攻擊值壓低
	rng = &seqRand{values: []int{1, 99}}
	out = Resolve(&a2, &d2, 0, rng)
	if !out.Parried && !out.Miss {
		t.Fatalf("應該擋下或落空：%+v", out)
	}
}

// SPEC-009 第 3 節及正常原版第 15 回合：先攻擊 rand(100)，再選列 rand(3)。
// 被擋下與負分落空的兩條路都必須保留這個順序。
func TestResolveRandomStyleOrder(t *testing.T) {
	for _, tc := range []struct {
		name          string
		encumbrance   int
		faces, values []int
		miss          bool
	}{
		{"被擋下", 400, []int{100, 3, 100}, []int{40, 2, 99}, false},
		{"負分落空", 2000, []int{100, 3}, []int{40, 2}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, d := testAttacker(), testDefender()
			a.Encumbrance = tc.encumbrance
			r := &expectedAttackRoll{t: t, faces: tc.faces, values: tc.values}
			out := Resolve(&a, &d, 1, r)
			if r.i != len(tc.faces) || out.Miss != tc.miss || out.Parried == tc.miss {
				t.Fatalf("擲骰次數 %d，結果 %+v", r.i, out)
			}
		})
	}
}

type expectedAttackRoll struct {
	t             *testing.T
	faces, values []int
	i             int
}

func (r *expectedAttackRoll) Roll(n int) int {
	r.t.Helper()
	if r.i >= len(r.faces) || n != r.faces[r.i] {
		r.t.Fatalf("第 %d 次擲骰面數 %d，預期序列 %v", r.i+1, n, r.faces)
	}
	v := r.values[r.i]
	r.i++
	return v
}

// 正常 SLEDGE：遭遇等級 3，但負重讀參戰者等級 1（sub_24FD3:0x24FEA）。
// 建場及戰鬥中換裝重算都須相同，攻防技能仍按遭遇等級縮放。
func TestMonsterGearUsesCombatLevel(t *testing.T) {
	site := &Site{Index: -1, Level: 3, Count: 1, MonsterX: []int{31}, MonsterY: []int{37}}
	m := Monster{Weapon: &Item{Encumbrance: 50}, ArmourWeight: 100,
		AttackSkill: 200, ParrySkill: 100}
	s := NewQuest(site, nil, m, &Field{})
	a := &s.Actors[PartySlots]
	if a.LoadLevel != 1 || a.Encumbrance != 87 || a.WeaponSpeed != 35 {
		t.Fatalf("建場：負重等級 %d，負重 %d，速度 %d", a.LoadLevel, a.Encumbrance, a.WeaponSpeed)
	}
	a.Encumbrance, a.WeaponSpeed = 0, 0
	a.ReloadGear()
	if a.Encumbrance != 87 || a.WeaponSpeed != 35 {
		t.Fatalf("重算：負重 %d，速度 %d", a.Encumbrance, a.WeaponSpeed)
	}
	if s.MonsterLevel != 3 || a.Level != 1 || a.Attack != 120 || a.Parry != 60 {
		t.Fatalf("遭遇等級或攻防技能被改：遭遇 %d，戰鬥 %d，攻防 %d／%d", s.MonsterLevel, a.Level, a.Attack, a.Parry)
	}
}

// 打到部位歸零那一肢就廢了；累積傷害到門檻就倒下。
func TestTake(t *testing.T) {
	d := Actor{Threshold: 100}
	d.SetParts()
	if down, limb := d.Take(20, 2); down || !limb {
		t.Fatalf("右臂 20 血吃 20 傷：down=%v limb=%v", down, limb)
	}
	if d.RightArm != 0 {
		t.Fatalf("右臂剩 %d", d.RightArm)
	}
	if down, _ := d.Take(80, 1); !down {
		t.Fatal("累積到門檻應該倒下")
	}
	if !d.Down() {
		t.Fatal("倒下之後 Down 要回真")
	}
}

// 徒手拿的是穿戴品，不是手上那件（RE-131）。
func TestUnarmedUsesWornItem(t *testing.T) {
	// UnarmedSlots 是 {0, 5, 6, 0, 4, 1}：式 1 是 FEET、式 2 是 SHIELD、
	// 式 3 是 HEAD、式 4 是 HANDS。
	if UnarmedSlots[1] != 5 || UnarmedSlots[2] != 6 ||
		UnarmedSlots[3] != 0 || UnarmedSlots[4] != 4 {
		t.Fatalf("徒手四式的槽是 %v", UnarmedSlots)
	}
	a := Actor{Command: CmdWeapon, Weapon: Weapon{Kind: 60, Count: 2, Sides: 5}}
	a.UnarmedWeapon[3] = Weapon{Kind: 4, Count: 1, Sides: 3}
	// 武器攻擊拿手上那件。
	if got := a.AttackWeapon(3); got.Kind != 60 {
		t.Fatalf("武器攻擊拿到 %#x", got.Kind)
	}
	// 徒手拿那一格穿戴品。
	a.Command = CmdUnarmed
	if got := a.AttackWeapon(3); got.Kind != 4 || got.Sides != 3 {
		t.Fatalf("徒手拿到 %+v", got)
	}
	// 全零槽傷害碼0仍是1d2，不退回手上那件（RE-197）。
	if got := a.AttackWeapon(2); got.Kind != 0 || got.Count != 1 || got.Sides != 2 {
		t.Fatalf("全零槽該保留1d2，得到 %+v", got)
	}
	// 出界也不會 panic。
	if got := a.AttackWeapon(99); got.Kind != 0 {
		t.Fatalf("出界該回空武器，得到 %+v", got)
	}
}

// 傷害骰跟著換：同一名角色，徒手與武器擲出來的上限不同。
func TestDamageRollUsesUnarmedWeapon(t *testing.T) {
	a := testAttacker()
	a.Command = CmdUnarmed
	a.UnarmedWeapon[3] = Weapon{Kind: 4, Base: 1, Count: 1, Sides: 3}
	rng := &seqRand{values: []int{3, 3}}
	// 力量加成 = (120 − 70) ÷ 8 = 6，再乘 (70 − 0) ÷ 100 = 4；
	// 基礎 1 ＋ 骰 3 ＋ 加成 4 ＝ 8，上限是 1×3 + 1 + 1 = 5，所以夾成 5。
	if got := DamageRoll(a, 3, 0, rng); got != 5 {
		t.Fatalf("徒手傷害 %d，期望夾在 5", got)
	}
}

// 盾牌只在「兩隻手臂都還在、手上不是雙手武器」時算數。
//
// 原版每次算負重都重新判斷（`sub_24FD3`），實跑量到的三個條件見
// `internal/oracle` 的 `TestProbeShieldConditions`。remake 之前把盾牌等級當成
// 建戰鬥時定下的常數，手臂被打斷或換上雙手武器之後還在加 `100 × 等級 + 200`。
func TestShieldNeedsBothArmsAndOneHandedWeapon(t *testing.T) {
	base := Actor{ShieldGrade: 3, Quickness: 60, Parries: 1,
		RightArm: 10, LeftArm: 10, Weapon: Weapon{Weight: 40}}
	full := base.DefenceValue()

	for _, c := range []struct {
		name string
		a    Actor
	}{
		{"右臂廢了", func() Actor { a := base; a.RightArm = 0; return a }()},
		{"左臂廢了", func() Actor { a := base; a.LeftArm = 0; return a }()},
		{"雙手武器", func() Actor { a := base; a.Weapon.TwoHand = true; return a }()},
	} {
		if got := c.a.DefenceValue(); got >= full {
			t.Errorf("%s：防禦值 %d，四肢健全時 %d——盾牌還在算", c.name, got, full)
		}
		if g := c.a.EffectiveShieldGrade(); g != NoShield {
			t.Errorf("%s：盾牌等級回 %d，期望 NoShield", c.name, g)
		}
	}
	// 射擊那一條也吃同一個判斷。
	noArm := base
	noArm.RightArm = 0
	if noArm.MissileDefenceValue() >= base.MissileDefenceValue() {
		t.Error("被射中時右臂廢了，盾牌還在算")
	}
}

// RE-241：規則終態相同，玩家等鍵前不可預先傷害或消耗。
func TestAttackMessagePhases(t *testing.T) {
	for _, command := range []int{CmdMissile, CmdWeapon, CmdUnarmed} {
		makeState := func() *State {
			s := shootState()
			s.Actors[0].Command, s.Actors[0].Kind, s.Actors[0].Flag1D = command, 5, 1
			s.Actors[0].Style = 2
			s.Actors[0].Attack = 1000
			s.Actors[1].Chest = 1
			if command != CmdMissile {
				s.Actors[1].Y = 11
			}
			return s
		}
		atomic, phased := makeState(), makeState()
		ra, rp := &seqRand{values: []int{90, 1, 2, 50, 1}}, &seqRand{values: []int{90, 1, 2, 50, 1}}
		full, ok := atomic.Act(ra)
		if !ok || full.Outcome == nil || full.Outcome.Miss || full.Outcome.Parried {
			t.Fatalf("沒有涵蓋命中: %+v", full)
		}
		before := phased.Actors[1]
		first, ok := phased.Step(rp)
		if !ok || !first.HitPending || first.DamageOnly {
			t.Fatalf("第一句未停在命中: %+v", first)
		}
		if phased.Actors[1].Damage != before.Damage || phased.Actors[1].Chest != before.Chest || phased.Actors[1].Hurt != before.Hurt || phased.Actors[1].Down() {
			t.Fatal("確認前已扣血或倒地")
		}
		if rp.i >= ra.i || phased.Actors[0].Kind != 5 || phased.Actors[0].Flag1D != 1 {
			t.Fatal("確認前已擲傷害或消耗")
		}
		second, ok := phased.Step(rp)
		if !ok || !second.DamageOnly || second.HitPending {
			t.Fatalf("確認後未完成傷害: %+v", second)
		}
		if *second.Outcome != *full.Outcome || rp.i != ra.i {
			t.Fatalf("傷害或骰序改變: %+v / %+v", second, full)
		}
		for i := range phased.Actors {
			// Actor 含裝備指標；用既有 JSON 表示核對所有可輸出欄位。
			a, _ := json.Marshal(atomic.Actors[i])
			b, _ := json.Marshal(phased.Actors[i])
			if string(a) != string(b) {
				t.Fatalf("指令%d角色%d終態不同", command, i)
			}
		}
		if phased.Exec.Turn != 1 || phased.Exec.pendingDamage != nil {
			t.Fatal("確認後未恢復下一手")
		}
	}
}

func TestAttackMessagePhasesNoExtraWait(t *testing.T) {
	for _, branch := range []string{"miss", "wall", "reload"} {
		s := shootState()
		s.Actors[0].Kind, s.Actors[0].Flag1D = 5, 1
		switch branch {
		case "miss":
			s.Actors[0].Encumbrance = 20000
		case "wall":
			s.Field.Tiles[10*FieldSide+13] = 0x29
		case "reload":
			s.Actors[0].Weapon.Kind, s.Actors[0].Flag1D = 18, 0
		}
		r, ok := s.Step(&seqRand{values: []int{50}})
		if !ok || r.HitPending || r.DamageOnly || s.Exec.pendingDamage != nil {
			t.Fatalf("%s 增加確認頁", branch)
		}
		want := 4
		if branch == "reload" {
			want = 5
		}
		if s.Actors[0].Kind != want {
			t.Fatalf("%s 消耗不同", branch)
		}
	}
}

func TestExecutionCameraFollowsActor(t *testing.T) {
	s := shootState()
	s.Cursor = 1
	s.Actors[0].Attack = 1000
	rng := &seqRand{values: []int{90, 1, 2, 50, 1}}
	first, ok := s.Step(rng)
	if !ok || !first.HitPending || s.Cursor != first.Actor {
		t.Fatalf("命中鏡頭沒有跟出手者: %+v cursor%d", first, s.Cursor)
	}
	x, y := s.Camera()
	if x != s.Actors[0].X || y != s.Actors[0].Y {
		t.Fatal("鏡頭中心不是出手者座標")
	}
	second, ok := s.Step(rng)
	if !ok || !second.DamageOnly || s.Cursor != first.Actor {
		t.Fatal("傷害句鏡頭離開出手者")
	}
	// 額外移動與結算訊息也使用各自的 Actor，沒有固定到規劃的最後一名。
	s = shootState()
	s.Cursor = 1
	s.Actors[0].Command = CmdRest
	r, ok := s.Step(&seqRand{values: []int{50}})
	if !ok || r.Actor != 0 || s.Cursor != 0 {
		t.Fatal("休息鏡頭未跟出手者")
	}
}

func TestSpriteAtOriginalPriority(t *testing.T) {
	s := &State{Actors: make([]Actor, 8), Cursor: 0}
	for i := range s.Actors {
		s.Actors[i] = Actor{Side: SideEmpty}
	}
	s.Ground[0] = GroundItem{Name: "Elf Bow", X: 10, Y: 12}
	if s.SpriteAt(10, 12) != SpriteObject {
		t.Fatal("地上物品未畫")
	}
	s.Actors[0] = Actor{Side: SideParty, X: 10, Y: 12, Head: 10, Chest: 10}
	if s.SpriteAt(10, 12) != SpritePartyBase {
		t.Fatal("隊員未遮物品")
	}
	s.Actors[0].Chest = 0
	if s.SpriteAt(10, 12) != SpritePartyDown {
		t.Fatal("倒地隊員未使用倒地圖示")
	}
	s.Actors[6] = Actor{Side: SideMonster, X: 10, Y: 12, Head: 10, Chest: 10}
	if s.SpriteAt(10, 12) != SpriteMonster {
		t.Fatal("站立未遮倒地")
	}
	s.Actors[6].Head = 0
	if s.SpriteAt(10, 12) != SpriteMonsterDown {
		t.Fatal("倒地敵人未蓋原有倒地")
	}
	s.Actors[0].Head, s.Actors[0].Chest = 10, 10
	s.Actors[6].Head = 10
	if s.SpriteAt(10, 12) != SpritePartyBase {
		t.Fatal("先前站立圖示應保留")
	}
	s.Cursor = 6
	if s.SpriteAt(10, 12) != SpriteMonster {
		t.Fatal("中心槽未取得同格優先")
	}
	s.Actors[7] = Actor{Side: SideMonster, X: 10, Y: 12, Head: 10, Chest: 10}
	if s.SpriteAt(10, 12) != SpriteMonster {
		t.Fatal("後槽覆蓋了中心站立圖示")
	}
	if s.SpriteAt(11, 12) != -1 {
		t.Fatal("空格或不同座標錯畫")
	}
	s.Ground[0].Name = ""
	s.Actors[0].Side = SideEmpty
	s.Actors[6].Side = SideEmpty
	s.Actors[7].Side = SideEmpty
	if s.SpriteAt(10, 12) != -1 {
		t.Fatal("空人物或空物品未略過")
	}
}

func TestMoveDisplayPreservesPreMoveState(t *testing.T) {
	s := shootState()
	s.Actors[0].Command, s.Actors[0].Style = CmdMove, 60
	oldX, oldY := s.Actors[0].X, s.Actors[0].Y
	r, ok := s.Step(&seqRand{values: []int{50}})
	if !ok || !r.Redraw || (s.Actors[0].X == oldX && s.Actors[0].Y == oldY) {
		t.Fatalf("未涵蓋成功移動: %+v", r)
	}
	x, y := s.Camera()
	if x != oldX || y != oldY || s.SpriteAt(oldX, oldY) != SpritePartyBase {
		t.Fatal("第一句未保留移動前画面")
	}
	newX, newY := s.Actors[0].X, s.Actors[0].Y
	if s.SpriteAt(newX, newY) == SpritePartyBase {
		t.Fatal("第一句提前畫新位置")
	}
	s.RefreshView()
	x, y = s.Camera()
	if x != newX || y != newY || s.SpriteAt(newX, newY) != SpritePartyBase {
		t.Fatal("確認後未刷新畫面")
	}
	if s.Actors[0].X != newX || s.Actors[0].Y != newY {
		t.Fatal("呈現修改了遊戲座標")
	}
}

func TestUnconsciousViewUsesTargetWithoutChangingActor(t *testing.T) {
	s := shootState()
	s.Actors[0].Attack = 1000
	s.Actors[1].Chest = 1
	rng := &seqRand{values: []int{90, 1, 2, 50, 1}}
	first, ok := s.Step(rng)
	if !ok || !first.HitPending {
		t.Fatal("未涵蓋命中等鍵")
	}
	second, ok := s.Step(rng)
	if !ok || !second.DamageOnly || !s.Actors[1].Down() {
		t.Fatal("未涵蓋傷害後倒地")
	}
	x, y := s.Camera()
	if s.Cursor != 0 || x != s.Actors[1].X || y != s.Actors[1].Y {
		t.Fatal("倒地鏡頭與出手者混在一起")
	}
	a, ok := s.ViewActor()
	if !ok || a.X != s.Actors[1].X {
		t.Fatal("共用視線沒有跟被擊中者")
	}
	s.Rand = &seqRand{values: []int{50}}
	s.Replan(s.Rand)
	if s.Exec.viewActor != nil || s.Exec.viewBefore != nil {
		t.Fatal("規劃重排未清呈現狀態")
	}
}

// RE-245：獨立原版條件與正常Moraine成本反例。
func TestShieldWeightAvailabilityAndRefresh(t *testing.T) {
	base := Actor{Side: SideParty, LoadLevel: 1, ShieldGrade: 1, ShieldWeight: 40,
		ArmourWeight: 300, RightArm: 21, LeftArm: 21, Threshold: 106,
		Head: 15, Chest: 15, Stamina: 223}
	for _, c := range []struct {
		name  string
		actor Actor
		enc   int
	}{
		{"雙臂", base, 238},
		{"右臂廢", func() Actor { a := base; a.RightArm = 0; return a }(), 210},
		{"左臂廢", func() Actor { a := base; a.LeftArm = 0; return a }(), 210},
		{"雙手武器", func() Actor { a := base; a.Weapon = Weapon{Weight: 40, TwoHand: true}; return a }(), 224},
		{"無盾", func() Actor { a := base; a.ShieldGrade = NoShield; return a }(), 210},
	} {
		a := c.actor
		a.ReloadGear()
		if a.Encumbrance != c.enc || a.ShieldWeight != 40 {
			t.Fatalf("%s：負重%d，期望%d；原始盾重%d", c.name, a.Encumbrance, c.enc, a.ShieldWeight)
		}
	}
	a := base
	a.Damage, a.Hurt, a.Fatigue = 38, 6, 50
	a.ReloadGear()
	a.Take(24, 3)
	if a.Encumbrance != 210 || a.Damage != 62 || a.Hurt != 30 || a.ShieldWeight != 40 {
		t.Fatalf("左臂受傷後依賴未更新：%+v", a)
	}
	s := State{Actors: []Actor{a}}
	if msg := s.EndTurnFatigue(0); msg != 0 || s.Actors[0].Fatigue != 110 {
		t.Fatalf("正常原版Moraine疲勞110：%d，訊息%d", s.Actors[0].Fatigue, msg)
	}
	s.castArea(0, 3, 1, false)
	if s.Actors[0].Encumbrance != 238 || s.Actors[0].ShieldWeight != 40 {
		t.Fatal("手臂恢復後盾牌負重未恢復")
	}
	s.Actors[0].Weapon = Weapon{Weight: 40, TwoHand: true}
	s.Actors[0].ReloadGear()
	if s.Actors[0].Encumbrance != 224 {
		t.Fatal("改雙手武器仍加盾重")
	}
	s.Actors[0].Weapon = Weapon{Weight: 40}
	s.Actors[0].ReloadGear()
	if s.Actors[0].Encumbrance != 252 {
		t.Fatal("改回單手武器未恢復盾重")
	}
}

func TestSeriousDamageWaitsBeforeKnockback(t *testing.T) {
	for _, missile := range []bool{false, true} {
		for _, success := range []bool{false, true} {
			a := Actor{Side: SideMonster, Member: -1, Command: CmdWeapon, Size: 200,
				X: 10, Y: 10, Head: 100, Chest: 100, RightArm: 100, LeftArm: 100, Stamina: 200,
				Weapon: Weapon{Base: 30, Count: 1, Sides: 5}, ShieldGrade: NoShield}
			d := Actor{Side: SideParty, Member: -1, Command: CmdRest, Size: 100,
				X: 10, Y: 11, Head: 100, Chest: 100, RightArm: 100, LeftArm: 100, RightLeg: 100, LeftLeg: 100,
				Threshold: 250, Stamina: 200, ShieldGrade: NoShield}
			if !success {
				a.Size = 100
			}
			if missile {
				a.Side = SideParty
				a.Command = CmdMissile
				a.Kind = 3
				a.Flag1D = 1
			}
			s := State{Actors: []Actor{a, d}}
			s.Exec.pendingDamage = &attackWait{actor: 0, target: 1, command: a.Command, style: 2, missile: missile, reload: missile}
			rng := &seqRand{values: []int{4, 50, 1, 1, 1}}
			first, ok := s.Step(rng)
			if !ok || !first.DamageOnly || first.Outcome == nil || !KnocksBack(*first.Outcome) {
				t.Fatalf("未涵蓋嚴重傷害：%+v", first)
			}
			if rng.i != 3 || s.Actors[1].X != 10 || s.Actors[1].Y != 11 || s.Actors[1].Command != CmdRest {
				t.Fatalf("傷害句確認前已擊退或擲骰：%+v，骰%d", s.Actors[1], rng.i)
			}
			if missile && (s.Actors[0].Kind != 3 || s.Actors[0].Flag1D != 1) {
				t.Fatal("傷害句確認前已扣彈藥或清裝填")
			}
			next, more := s.nextAct(rng, true)
			if success {
				if !more || !next.Plain || len(next.Records) == 0 || rng.i != 5 || s.Actors[1].Command != 0 {
					t.Fatalf("確認後未恢復擊退：%+v，骰%d", next, rng.i)
				}
				if s.Actors[1].X == 10 && s.Actors[1].Y == 11 {
					t.Fatal("未涵蓋成功位移")
				}
			} else if more || rng.i != 4 || s.Actors[1].Command != CmdRest {
				t.Fatalf("不成立擊退新增等鍵或改指令：%+v，骰%d", next, rng.i)
			}
			if missile && (s.Actors[0].Kind != 2 || s.Actors[0].Flag1D != 0 || s.Actors[0].Command != CmdLoad) {
				t.Fatal("擊退判定後未完成射擊收尾")
			}
		}
	}
}

func TestKnockbackIgnoresFallenAndMountedBlockers(t *testing.T) {
	for _, c := range []struct {
		name                 string
		head, chest, mounted int
		blocked              bool
	}{
		{"倒地人物", 16, 0, 0, false},
		{"在場人物", 16, 16, 0, true},
		{"騎乘人物", 16, 16, 1, false},
	} {
		s := State{Actors: []Actor{
			{Side: SideMonster, Member: -1, X: 32, Y: 32, Size: 232, Head: 30, Chest: 30},
			{Side: SideParty, Member: -1, X: 31, Y: 31, Size: 106, Head: 15, Chest: 15, Command: CmdRest},
			{Side: SideParty, Member: -1, X: 31, Y: 30, Head: c.head, Chest: c.chest, Flag1E: c.mounted},
		}}
		rng := &seqRand{values: []int{39, 2}}
		records := s.Knockback(0, 1, rng)
		if rng.i != 2 {
			t.Fatalf("%s：擊退骰數%d", c.name, rng.i)
		}
		if c.blocked {
			if len(records) != 0 || s.Actors[1].X != 31 || s.Actors[1].Y != 31 || s.Actors[1].Command != CmdRest {
				t.Fatalf("%s：徒手阻擋分支不符：%+v", c.name, s.Actors[1])
			}
		} else if len(records) != 2 || s.Actors[1].X != 30 || s.Actors[1].Y != 31 || s.Actors[1].Command != 0 {
			t.Fatalf("%s：檢查格31/30、落點30/31未通過：%+v", c.name, s.Actors[1])
		}
	}
}
