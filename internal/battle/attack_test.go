package battle

import "testing"

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
	// 那一格空著就是空武器——不會退回手上那件。
	if got := a.AttackWeapon(2); got.Kind != 0 || got.Count != 0 {
		t.Fatalf("空格該回空武器，得到 %+v", got)
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
