package battle

// 戰場上的丟、撿與切換（RE-094）。這三手不走執行段的分派表——原版在同一支的
// 前一段（`0x20189`…`0x20621`）就處理掉了，分派表只認另外十個指令碼（RE-093）。

// GroundSlots 是戰場上的物品格數（`0x9D20` 十六格，每格 16 byte）。
const GroundSlots = 16

// GroundItem 是戰場上的一格物品。原版把物品記錄放 `0x9D20`、座標分開放兩張
// 平行的表（`0x2BAE` 是 X、`0x2D9E` 是 Y）；remake 併成一個結構，語意相同。
type GroundItem struct {
	// Name 是物品的名字，空字串代表這一格是空的（原版看記錄的第一個 byte）。
	Name string
	X, Y int
}

// DropItem 是丟那一手（指令碼 `0x15`，`0x20189`）：把手上那件放到自己腳下。
//
// 十六格都滿了就什麼都沒做，只印記錄 0x10C。回傳訊息記錄與丟掉的東西的名字——
// **名字要另外帶出來**：原版是先印記錄 0x10D（`drops|B|E!`）再搬東西，`|E` 展開
// 的時候手上還有那件；remake 這一支做完手上已經空了，訊息會變成 `DROPS HIS !`。
func (s *State) DropItem(i int) (int, string) {
	a := &s.Actors[i]
	m := s.member(i)
	if m == nil || !m.HasWeapon {
		return CannotDropRecord, ""
	}
	slot := -1
	for n := range s.Ground {
		if s.Ground[n].Name == "" {
			slot = n
			break
		}
	}
	if slot < 0 {
		return CannotDropRecord, ""
	}
	name := m.WeaponName
	s.Ground[slot] = GroundItem{Name: name, X: a.X, Y: a.Y}
	m.HasWeapon = false
	m.WeaponName = ""
	m.Worn[0] = ""
	a.Weapon = Weapon{}
	return DropsRecord, name
}

// GroundUnder 找第 from 格之後、腳下那一格上的第一件東西（`0x202DE`…`0x20315`）。
//
// **只撿得到腳下那一格**——原版 X 與 Y 都要對上。沒有就回 −1。
func (s *State) GroundUnder(i, from int) int {
	if i < 0 || i >= len(s.Actors) || from < 0 {
		return -1
	}
	a := &s.Actors[i]
	for n := from; n < len(s.Ground); n++ {
		g := s.Ground[n]
		if g.Name != "" && g.X == a.X && g.Y == a.Y {
			return n
		}
	}
	return -1
}

// TakeItem 把地上第 slot 格那件東西撿起來（`0x203CC`…`0x204E4`）。
//
// 回傳訊息記錄與那件東西的名字。手上已經有東西就撿不起來，只印記錄 0x112
// （`0x20417`）——**那一手就到這裡為止**，不會再掃下一格。
func (s *State) TakeItem(i, slot int) (int, string) {
	if slot < 0 || slot >= len(s.Ground) || s.Ground[slot].Name == "" {
		return 0, ""
	}
	m := s.member(i)
	if m != nil && m.HasWeapon {
		return MustDropRecord, m.WeaponName
	}
	g := s.Ground[slot]
	if m != nil {
		m.HasWeapon = true
		m.WeaponName = g.Name
		m.Worn[0] = g.Name
	} else if i >= 0 && i < len(s.Actors) {
		// 敵人撿的那一件放進怪物記錄手上那一格（`word_39A4E = 0x20`，
		// `sub_1E3A8:0x203E0`），同時把 +0x2B 清成 0。手上不再空著，AI 那一條
		// 「沒有武器就去撿」下一回合就不成立了（RE-116）。
		//
		// **武器的數字沒有跟著換**：地上那一格只留名字，要換算成傷害骰得先把
		// ITEM.LST 那一筆接起來——撿到的東西這一版還是照原本的空手數字打。
		s.Actors[i].Unarmed = false
		s.Actors[i].PickedUp = 0
	}
	s.Ground[slot] = GroundItem{}
	return FoundItemRecord, g.Name
}

// QuestItemCall 回報撿起來那一件要不要接記錄 0x113
// （`|G, this is the quest item!  Run for the exit!`，`0x2049B`…`0x204B9`）。
//
// 三個條件：這一場是任務戰鬥（原版看 `word_34968`，remake 看 `Site.Index`）、
// 撿到的正是任務要的那一件（`sub_2886C`），而且**這一場還沒喊過**
// （`word_32382`）。喊過就把旗標立起來，一場只喊一次。
func (s *State) QuestItemCall(name string) bool {
	if s.questCalled || name == "" || name != s.QuestItem {
		return false
	}
	if s.Site == nil || s.Site.Index < 0 {
		return false
	}
	s.questCalled = true
	return true
}

// PickUpItem 是撿那一手的完整動作，給不必問的那一側（敵人）與測試用。
// 回傳訊息記錄、撿到的格號與名字；腳下沒東西就回 (0, −1, "")。
func (s *State) PickUpItem(i int) (int, int, string) {
	slot := s.GroundUnder(i, 0)
	if slot < 0 {
		return 0, -1, ""
	}
	msg, item := s.TakeItem(i, slot)
	if msg == MustDropRecord {
		return msg, -1, item
	}
	return msg, slot, item
}

// SwitchItem 是切換那一手（指令碼 `0x24`／`0x25`／`0x16`，`0x2050E`）：
// 手上那件與腰帶那件對調。原版是把角色記錄 +0x30 起 16 byte 與 +0x190 起 16 byte
// 整段交換。
func (s *State) SwitchItem(i int) int {
	m := s.member(i)
	if m == nil {
		return 0
	}
	m.WeaponName, m.BeltName = m.BeltName, m.WeaponName
	m.HasWeapon, m.HasSecondItem = m.HasSecondItem, m.HasWeapon
	m.Worn[0], m.Worn[5] = m.Worn[5], m.Worn[0]
	if !m.HasWeapon {
		s.Actors[i].Weapon = Weapon{}
	}
	return PutsInBeltRecord
}

// askPickUp 把執行段停在「要不要撿」那一問上（`0x20354`，RE-113）。
// 訊息是記錄 0x111 —— 它自己就帶著「You have found a|H.」與問句兩半，
// 所以名字要跟著出去，`|H` 展開的時候東西還在地上。
func (s *State) askPickUp(i, slot int) ActResult {
	s.PickUp = &PickUpAsk{Actor: i, Slot: slot}
	s.Plan.Prompt = PromptPickUp
	return ActResult{Actor: i, Command: CmdPickUp, Target: -1, Plain: true,
		Message: FoundItemRecord, Item: s.Ground[slot].Name}
}

// AnswerPickUp 是那一問的回答（`0x2036E`…`0x203B5`）。
//
//   - 答是 → 撿起來，接著印撿到那一句（或手上已經有東西的 0x112）。
//   - 答否 → 這一格跳過，**繼續掃下一格**（`0x20501` 回到同一個迴圈），
//     沒有下一格這一手就結束。
//
// 回傳「有沒有話要說」；有的話留給執行段下一次取手時交出去。
func (s *State) AnswerPickUp(yes bool) bool {
	ask := s.PickUp
	if ask == nil {
		return false
	}
	if !yes {
		if next := s.GroundUnder(ask.Actor, ask.Slot+1); next >= 0 {
			ask.Slot = next
			s.resume = &ActResult{Actor: ask.Actor, Command: CmdPickUp,
				Target: -1, Plain: true, Message: FoundItemRecord,
				Item: s.Ground[next].Name}
			return true
		}
		s.PickUp = nil
		s.Plan.Prompt = PromptExecuting
		return false
	}
	i, slot := ask.Actor, ask.Slot
	name := s.Ground[slot].Name
	s.PickUp = nil
	s.Plan.Prompt = PromptExecuting
	msg, item := s.TakeItem(i, slot)
	out := ActResult{Actor: i, Command: CmdPickUp, Target: -1, Plain: true,
		Item: item}
	switch {
	case msg == MustDropRecord:
		// 手上已經有東西：只印 0x112，那一手到這裡為止（`0x20417`）。
		out.Message = msg
	case s.QuestItemCall(name):
		// 撿到的正是任務要的那一件就喊一句（記錄 0x113），一場只喊一次。
		out.Message = QuestItemRecord
	default:
		// **撿成功本身不印任何一句**——問句 0x111 已經把「找到什麼」講完了
		// （`0x204E4` 直接收尾）。
		return false
	}
	s.resume = &out
	return true
}
