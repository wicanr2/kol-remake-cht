package battle

import "testing"

func itemState() *State {
	s := &State{Party: []Member{{
		Name: "POONTANG", HasWeapon: true, WeaponName: "LONG SPEAR",
		BeltName: "DAGGER", HasSecondItem: true,
	}}}
	s.Party[0].Worn[0] = "LONG SPEAR"
	s.Party[0].Worn[5] = "DAGGER"
	s.Actors = []Actor{{Side: SideParty, Member: 0, X: 12, Y: 34,
		Head: 10, Chest: 10, Threshold: 100}}
	return s
}

// pickUpState 是一名空手的隊員站在一件東西上，指令是撿。
func pickUpState() *State {
	s := &State{Party: []Member{{Name: "POONTANG"}},
		Plan: Planning{Prompt: PromptExecuting}}
	s.Actors = []Actor{{Side: SideParty, Member: 0, X: 30, Y: 30,
		Head: 10, Chest: 10, Threshold: 100, Command: CmdPickUp}}
	s.Ground[0] = GroundItem{Name: "DAGGER", X: 30, Y: 30}
	return s
}

// 丟：東西掉在自己腳下那一格，手上變空。
func TestDropItemLandsUnderfoot(t *testing.T) {
	s := itemState()
	got, name := s.DropItem(0)
	if got != DropsRecord {
		t.Fatalf("訊息是 %#x，期望 %#x", got, DropsRecord)
	}
	if name != "LONG SPEAR" {
		t.Fatalf("丟掉的名字是 %q——訊息要用它展開 |E", name)
	}
	if s.Ground[0].Name != "LONG SPEAR" {
		t.Fatalf("第 0 格是 %q", s.Ground[0].Name)
	}
	if s.Ground[0].X != 12 || s.Ground[0].Y != 34 {
		t.Fatalf("掉在 (%d,%d)，應該在 (12,34)", s.Ground[0].X, s.Ground[0].Y)
	}
	if s.Party[0].HasWeapon || s.Party[0].WeaponName != "" {
		t.Fatal("丟完手上應該是空的")
	}
}

// 十六格都滿了就丟不下去，只印 0x10C。
func TestDropItemWhenGroundIsFull(t *testing.T) {
	s := itemState()
	for i := range s.Ground {
		s.Ground[i] = GroundItem{Name: "ROCK", X: 1, Y: 1}
	}
	if got, _ := s.DropItem(0); got != CannotDropRecord {
		t.Fatalf("訊息是 %#x，期望 %#x", got, CannotDropRecord)
	}
	if !s.Party[0].HasWeapon {
		t.Fatal("丟不下去的時候武器應該還在手上")
	}
}

// 撿：只撿得到腳下那一格，不同格的碰不到。
func TestPickUpOnlyReachesTheSameCell(t *testing.T) {
	s := itemState()
	s.Party[0].HasWeapon = false
	s.Party[0].WeaponName = ""
	s.Ground[0] = GroundItem{Name: "MACE", X: 12, Y: 35}   // 隔壁
	s.Ground[1] = GroundItem{Name: "SHIELD", X: 12, Y: 34} // 腳下

	msg, slot, name := s.PickUpItem(0)
	if name != "SHIELD" {
		t.Fatalf("撿到的名字是 %q——訊息要用它展開 |H", name)
	}
	if msg != FoundItemRecord {
		t.Fatalf("訊息是 %#x，期望 %#x", msg, FoundItemRecord)
	}
	if slot != 1 {
		t.Fatalf("撿到第 %d 格，應該是第 1 格", slot)
	}
	if s.Party[0].WeaponName != "SHIELD" {
		t.Fatalf("手上是 %q", s.Party[0].WeaponName)
	}
	if s.Ground[1].Name != "" {
		t.Fatal("撿走之後那一格應該空了")
	}
	if s.Ground[0].Name != "MACE" {
		t.Fatal("隔壁那一格不該被動到")
	}
}

// 手上已經有東西就撿不起來，印 0x112。
func TestPickUpNeedsAnEmptyHand(t *testing.T) {
	s := itemState()
	s.Ground[0] = GroundItem{Name: "MACE", X: 12, Y: 34}
	msg, slot, _ := s.PickUpItem(0)
	if msg != MustDropRecord || slot != -1 {
		t.Fatalf("手上有東西時得到 %#x／%d", msg, slot)
	}
	if s.Ground[0].Name != "MACE" {
		t.Fatal("撿不起來的東西應該還在地上")
	}
}

// 腳下沒東西就什麼都不印。
func TestPickUpNothingThere(t *testing.T) {
	s := itemState()
	if msg, slot, _ := s.PickUpItem(0); msg != 0 || slot != -1 {
		t.Fatalf("腳下沒東西時得到 %#x／%d", msg, slot)
	}
}

// 切換：手上與腰帶對調。
func TestSwitchItemSwapsHandAndBelt(t *testing.T) {
	s := itemState()
	if got := s.SwitchItem(0); got != PutsInBeltRecord {
		t.Fatalf("訊息是 %#x，期望 %#x", got, PutsInBeltRecord)
	}
	m := s.Party[0]
	if m.WeaponName != "DAGGER" || m.BeltName != "LONG SPEAR" {
		t.Fatalf("交換後手上 %q、腰帶 %q", m.WeaponName, m.BeltName)
	}
	if m.Worn[0] != "DAGGER" || m.Worn[5] != "LONG SPEAR" {
		t.Fatalf("裝備欄沒跟著換：%q／%q", m.Worn[0], m.Worn[5])
	}
	// 換回去要回到原狀。
	s.SwitchItem(0)
	if s.Party[0].WeaponName != "LONG SPEAR" {
		t.Fatal("換兩次應該回到原狀")
	}
}

// 三手都走執行段的 act，不再被標成還沒實作。
func TestActHandlesTheThreeItemCommands(t *testing.T) {
	for _, cmd := range []int{CmdDrop, CmdPickUp, CmdSwitch, CmdSwitchAlt1, CmdSwitchAlt2} {
		s := itemState()
		s.Actors[0].Command = cmd
		r := s.act(0, &fixedRand{values: []int{1}})
		if r.Unimplemented {
			t.Fatalf("%#x 不該被標成還沒實作", cmd)
		}
	}
}

// 撿那一手對隊員會先問（記錄 0x111），答是才撿走（RE-113）。
func TestPickUpAsksPartyMember(t *testing.T) {
	s := pickUpState()
	s.Exec.Order, s.Exec.Turn = []int{0}, 0
	r, ok := s.Step(maxRand{})
	if !ok {
		t.Fatal("該有一手")
	}
	if r.Message != FoundItemRecord || r.Item != "DAGGER" {
		t.Fatalf("問句該是 0x111 ＋ 東西的名字，得到 %#x／%q", r.Message, r.Item)
	}
	if s.Plan.Prompt != PromptPickUp || s.PickUp == nil {
		t.Fatalf("該停在撿那一問，Prompt %d PickUp %+v", s.Plan.Prompt, s.PickUp)
	}
	if s.Ground[0].Name == "" {
		t.Fatal("還沒答應之前東西還在地上")
	}
	// **撿成功本身不印任何一句**：問句已經把「找到什麼」講完了（`0x204E4`）。
	if s.AnswerPickUp(true) {
		t.Fatal("撿成功不該再多印一句")
	}
	if s.Ground[0].Name != "" {
		t.Fatal("答應了東西該離開地面")
	}
	if s.Plan.Prompt != PromptExecuting || s.PickUp != nil {
		t.Fatal("那一問該收掉")
	}
	if m := s.Member(0); m == nil || m.WeaponName != "DAGGER" {
		t.Fatalf("手上該是 DAGGER，得到 %+v", m)
	}
}

// 答否就跳過這一格；腳下沒有別的東西，這一手就結束。
func TestPickUpDeclineEndsTheHand(t *testing.T) {
	s := pickUpState()
	s.Exec.Order, s.Exec.Turn = []int{0}, 0
	if _, ok := s.Step(maxRand{}); !ok {
		t.Fatal("該有一手")
	}
	if s.AnswerPickUp(false) {
		t.Fatal("腳下只有一件，答否之後不該再有話說")
	}
	if s.PickUp != nil || s.Plan.Prompt == PromptPickUp {
		t.Fatal("那一問該收掉")
	}
	if s.Ground[0].Name != "DAGGER" {
		t.Fatal("答否東西該留在地上")
	}
}

// 腳下有兩件時答否會問下一件（`0x20501` 回到同一個迴圈）。
func TestPickUpDeclineAsksNextItem(t *testing.T) {
	s := pickUpState()
	s.Ground[1] = GroundItem{Name: "SHIELD", X: 30, Y: 30}
	s.Exec.Order, s.Exec.Turn = []int{0}, 0
	if _, ok := s.Step(maxRand{}); !ok {
		t.Fatal("該有一手")
	}
	if !s.AnswerPickUp(false) {
		t.Fatal("還有一件，該再問一次")
	}
	r, ok := s.Step(maxRand{})
	if !ok || r.Item != "SHIELD" {
		t.Fatalf("第二問該是 SHIELD，得到 %q（ok=%v）", r.Item, ok)
	}
	if s.PickUp == nil || s.Ground[1].Name == "" {
		t.Fatal("第二問還沒答，東西還在地上")
	}
}

// 敵人不問，直接撿（`0x20332` 比參戰者記錄 +0x01）。
func TestPickUpEnemyDoesNotAsk(t *testing.T) {
	s := pickUpState()
	s.Actors[0].Side = SideMonster
	s.Party = nil
	s.Exec.Order, s.Exec.Turn = []int{0}, 0
	r, ok := s.Step(maxRand{})
	if !ok || r.Message != FoundItemRecord {
		t.Fatalf("敵人該直接撿，得到 %#x（ok=%v）", r.Message, ok)
	}
	if s.Plan.Prompt == PromptPickUp || s.PickUp != nil {
		t.Fatal("敵人不該被問")
	}
	if s.Ground[0].Name != "" {
		t.Fatal("東西該離開地面")
	}
}

// 撿到任務要的那一件會多喊一句（記錄 0x113），一場只喊一次。
func TestPickUpCallsQuestItemOnce(t *testing.T) {
	s := pickUpState()
	s.Site = &Site{Index: 3}
	s.QuestItem = "DAGGER"
	s.Exec.Order, s.Exec.Turn = []int{0}, 0
	if _, ok := s.Step(maxRand{}); !ok {
		t.Fatal("該有一手")
	}
	if !s.AnswerPickUp(true) {
		t.Fatal("撿到任務物品該多喊一句")
	}
	r, _ := s.Step(maxRand{})
	if r.Message != QuestItemRecord || !r.Plain {
		t.Fatalf("該是記錄 0x113 且不加名字，得到 %#x／Plain=%v", r.Message, r.Plain)
	}
	// 再撿一次同一件：旗標已經立起來，不再喊。
	if m := s.Member(0); m != nil {
		m.HasWeapon = false
		m.WeaponName = ""
	}
	s.Ground[0] = GroundItem{Name: "DAGGER", X: 30, Y: 30}
	s.Exec.Order, s.Exec.Turn = []int{0}, 0
	if _, ok := s.Step(maxRand{}); !ok {
		t.Fatal("該再有一手")
	}
	if s.AnswerPickUp(true) {
		t.Fatal("同一場只喊一次")
	}
}

// 撿起來的那一件要帶著自己的傷害骰，不是沿用原本手上那件（RE-116 的補完）。
func TestPickUpCarriesTheWeaponNumbers(t *testing.T) {
	broadsword := Weapon{Kind: 0x58, Base: 2, Count: 1, Sides: 10, Weight: 60}
	s := &State{Actors: []Actor{
		{Side: SideMonster, X: 4, Y: 4, Head: 10, Chest: 10, Threshold: 100,
			Member: -1, Unarmed: true},
	}}
	s.Ground[0] = GroundItem{Name: "BROADSWORD", X: 4, Y: 4, Weapon: broadsword}
	if _, name := s.TakeItem(0, 0); name != "BROADSWORD" {
		t.Fatalf("撿到的是 %q", name)
	}
	if s.Actors[0].Weapon != broadsword {
		t.Fatalf("武器數字沒換過來：%+v", s.Actors[0].Weapon)
	}
}

// 丟出去的那一件也要把數字留在地上，撿回來才拿得回同一把。
func TestDropThenPickUpKeepsTheSameWeapon(t *testing.T) {
	dagger := Weapon{Kind: 0x58, Base: 1, Count: 1, Sides: 3, Weight: 10}
	s := &State{
		Party: []Member{{Name: "POONTANG", HasWeapon: true, WeaponName: "DAGGER"}},
		Actors: []Actor{
			{Side: SideParty, X: 4, Y: 4, Head: 10, Chest: 10, Threshold: 100,
				Member: 0, Weapon: dagger},
		}}
	if record, name := s.DropItem(0); record != DropsRecord || name != "DAGGER" {
		t.Fatalf("丟的結果是 %#x／%q", record, name)
	}
	if s.Ground[0].Weapon != dagger {
		t.Fatalf("地上那一格沒留住數字：%+v", s.Ground[0].Weapon)
	}
	if s.Actors[0].Weapon != (Weapon{}) {
		t.Fatal("丟掉之後手上不該還有數字")
	}
	if _, name := s.TakeItem(0, 0); name != "DAGGER" {
		t.Fatalf("撿回來的是 %q", name)
	}
	if s.Actors[0].Weapon != dagger {
		t.Fatalf("撿回來的數字不對：%+v", s.Actors[0].Weapon)
	}
}

// 任務地點開打時那一件東西就擺在戰場上，座標來自記錄的 +0x27／+0x28（RE-113／128）。
func TestQuestItemStartsOnTheGround(t *testing.T) {
	site := &Site{Index: 3, Count: 0}
	site.QuestItem.Name = "Truth Sword"
	site.QuestItem.Spot = Spot{X: 6, Y: 38}
	site.QuestItem.Item.Class = 87
	site.QuestItem.Item.Encumbrance = 75
	s := NewQuest(site, nil, Monster{}, &Field{})
	if s == nil {
		t.Fatal("擺不出戰場")
	}
	g := s.Ground[0]
	if g.Name != "TRUTH SWORD" || g.X != 6 || g.Y != 38 {
		t.Fatalf("第 0 格是 %+v", g)
	}
	// 撿起來要拿它去打，所以武器數字要跟著。
	if g.Weapon.Weight != 75 || g.Weapon.Kind != 87 {
		t.Fatalf("武器數字是 %+v", g.Weapon)
	}
	// 沒有任務物品的場地不放（`site.Index` 是 −1 那一種）。
	empty := &Site{Index: -1}
	if e := NewQuest(empty, nil, Monster{}, &Field{}); e.Ground[0].Name != "" {
		t.Fatalf("不該放：%+v", e.Ground[0])
	}
}

// 撿起一件武器之後，負重與武器速度要跟著變。
//
// 原版每次算一手的花費都重跑 `sub_24FD3`（`sub_251B0` 開頭就呼叫它，
// `internal/oracle` 的 `TestActionCostMatchesOriginal` 實跑量到），所以戰鬥中
// 換武器兩個數字會立刻反映。remake 之前只換 `Weapon`，兩格停在開場算的那一次
// ——花費 → 疲勞 → 撤退判斷整條鏈都跟著偏。
func TestPickUpReloadsGear(t *testing.T) {
	s := pickUpState()
	a := &s.Actors[0]
	a.LoadLevel, a.ArmourWeight, a.ShieldWeight = 3, 20, 10
	a.ReloadGear()
	bareEnc, bareSpeed := a.Encumbrance, a.WeaponSpeed

	s.Ground[0].Weapon = Weapon{Weight: 60}
	if rec, slot, _ := s.PickUpItem(0); rec == MustDropRecord || slot < 0 {
		t.Fatalf("撿不起來：記錄 %d、格 %d", rec, slot)
	}
	if a.WeaponSpeed <= bareSpeed {
		t.Errorf("撿了重武器之後速度是 %d，撿之前 %d——沒有重算",
			a.WeaponSpeed, bareSpeed)
	}
	if a.Encumbrance <= bareEnc {
		t.Errorf("撿了重武器之後負重是 %d，撿之前 %d——沒有重算",
			a.Encumbrance, bareEnc)
	}
	wantEnc, wantSpeed := Encumbrance(3, Weapon{Weight: 60}, 20, 10)
	if a.Encumbrance != wantEnc || a.WeaponSpeed != wantSpeed {
		t.Errorf("撿完是 (負重 %d, 速度 %d)，直接算是 (%d, %d)",
			a.Encumbrance, a.WeaponSpeed, wantEnc, wantSpeed)
	}
}

// 丟掉手上那件之後同樣要回到徒手的數字。
func TestDropReloadsGear(t *testing.T) {
	s := itemState()
	a := &s.Actors[0]
	a.LoadLevel, a.ArmourWeight, a.ShieldWeight = 3, 20, 10
	a.Weapon = Weapon{Weight: 60}
	a.ReloadGear()
	armedEnc, armedSpeed := a.Encumbrance, a.WeaponSpeed

	if rec, _ := s.DropItem(0); rec == CannotDropRecord {
		t.Fatal("丟不掉")
	}
	if a.WeaponSpeed >= armedSpeed || a.Encumbrance >= armedEnc {
		t.Errorf("丟掉之後還是 (負重 %d, 速度 %d)，丟之前 (%d, %d)",
			a.Encumbrance, a.WeaponSpeed, armedEnc, armedSpeed)
	}
	wantEnc, wantSpeed := Encumbrance(3, Weapon{}, 20, 10)
	if a.Encumbrance != wantEnc || a.WeaponSpeed != wantSpeed {
		t.Errorf("丟完是 (負重 %d, 速度 %d)，直接算是 (%d, %d)",
			a.Encumbrance, a.WeaponSpeed, wantEnc, wantSpeed)
	}
}

// 撿起不同種類的武器之後，攻防技能要換成那一列。
//
// 原版每次要用技能都重查（`sub_26AC2`）：拿手上武器的種類去角色記錄 `+0x52` 起
// 四筆裡找，**找不到就是 0**（實跑量到，`internal/oracle` 的
// `TestPartySkillRowFollowsWeaponClass`）。remake 之前在建戰鬥時選一次就不再換。
func TestPickUpReloadsWeaponSkills(t *testing.T) {
	s := pickUpState()
	a := &s.Actors[0]
	a.WeaponRows = []WeaponSkillRow{
		{Class: 1, Offence: 111, Defence: 211},
		{Class: 2, Offence: 122, Defence: 222},
	}
	a.Weapon = Weapon{Kind: 1}
	a.ReloadSkills()
	if a.Attack != 111 || a.Parry != 211 {
		t.Fatalf("種類 1：技能 (%d, %d)，期望 (111, 211)", a.Attack, a.Parry)
	}

	s.Ground[0].Weapon = Weapon{Kind: 2, Weight: 20}
	if rec, slot, _ := s.PickUpItem(0); rec == MustDropRecord || slot < 0 {
		t.Fatalf("撿不起來：記錄 %d、格 %d", rec, slot)
	}
	if a.Attack != 122 || a.Parry != 222 {
		t.Errorf("撿了種類 2 的武器之後技能是 (%d, %d)，期望 (122, 222)",
			a.Attack, a.Parry)
	}

	// 種類不在四列裡：原版算 0，remake 要跟上。
	a.Weapon = Weapon{Kind: 0x40}
	a.ReloadSkills()
	if a.Attack != 0 || a.Parry != 0 {
		t.Errorf("種類不在四列裡時技能是 (%d, %d)，期望 (0, 0)", a.Attack, a.Parry)
	}

	// 怪物那一側不吃這一條——技能在本體記錄，與手上那件無關。
	m := Actor{Side: SideMonster, Attack: 99, Parry: 88,
		WeaponRows: a.WeaponRows, Weapon: Weapon{Kind: 1}}
	m.ReloadSkills()
	if m.Attack != 99 || m.Parry != 88 {
		t.Errorf("怪物的技能被換掉了：(%d, %d)，期望 (99, 88)", m.Attack, m.Parry)
	}
}
