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

// PickUpItem 是撿那一手（指令碼 `0x26`，`0x202CC`）：撿自己腳下那一格的東西。
//
// **只撿得到腳下那一格**——原版 X 與 Y 都要對上。回傳訊息記錄與撿到的格號；
// 沒東西可撿就回 (0, -1)。
//
// **與原版的一個差異**：原版對隊員會先印記錄 0x111 問「要不要撿」再等鍵，
// 敵人才直接撿。remake 這一版一律直接撿——那個問句要在執行段中間插一次互動，
// 目前的執行段還沒有那個機制。
func (s *State) PickUpItem(i int) (int, int, string) {
	a := &s.Actors[i]
	m := s.member(i)
	for n := range s.Ground {
		g := s.Ground[n]
		if g.Name == "" || g.X != a.X || g.Y != a.Y {
			continue
		}
		if m != nil && m.HasWeapon {
			// 手上已經有東西，原版要先丟或收進腰帶（記錄 0x112）。
			return MustDropRecord, -1, m.WeaponName
		}
		if m != nil {
			m.HasWeapon = true
			m.WeaponName = g.Name
			m.Worn[0] = g.Name
		}
		s.Ground[n] = GroundItem{}
		return FoundItemRecord, n, g.Name
	}
	return 0, -1, ""
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
