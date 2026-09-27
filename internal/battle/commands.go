package battle

// 戰場上玩家那一手（SPEC-010／RE-082）。指令列是一個會被清空重建的堆疊，
// `sub_228B7` 逐項看裝備、傷勢與腳下有什麼決定要不要放；選中之後依圖號走到
// 對應的子選單，把攻擊選項與防禦選項算出來，最後與指令碼一起寫回參戰者記錄。

// 指令列上的命令碼。名稱只寫已證實的用途；語意未定的沿用圖號
// （SPEC-010 末段列了哪幾個還沒對上 KOLDATA 的動作敘述）。
const (
	// CmdYes 是大姆指，與城鎮共用同一個圖。
	CmdYes = 0x07
	// CmdCancel 是 U-TURN。
	CmdCancel = 0x0F
	// CmdCast 是施法（`0x208ED` 對它呼叫 `sub_2818D`，RE-093）。原本從指令列的
	// 圖號推成「背包清單」，推錯了。
	CmdCast = 0x0D
	// 丟、撿與切換（RE-094）。三個原本都從指令列的圖號推，推錯了：0x15 是丟、
	// 0x26 是撿、0x24／0x25／0x16 是把手上的武器與腰帶那件對調。
	CmdDrop = 0x15
	// CmdSwitch 是切換；0x25 與 0x16 走同一條路。
	CmdSwitch     = 0x24
	CmdSwitchAlt1 = 0x16
	CmdSwitchAlt2 = 0x25
	CmdPickUp     = 0x26
	// CmdLoad 是**還沒上弦的射擊武器**那一格（RE-105）。它的出現條件正是
	// 「種類不是 6／0x0A／0x0B 而且 +0x1D 為零」，選下去在執行段 `0x20799`
	// 一定變成裝填，這一回合就沒了；下一回合 +0x1D 非零，同一格換成 CmdMissile。
	// **不是投擲**——`sub_26D76` 裡沒有一處比 0x22，它根本走不到那裡。
	CmdLoad = 0x22
	// CmdMissile 是射擊。
	CmdMissile = 0x23
	// CmdWeapon 是武器攻擊；CmdUnarmed 是徒手。
	CmdWeapon  = 0x21
	CmdUnarmed = 0x29
	// CmdMove 是移動。另外兩格只在 Flag1E 非零（騎著）時出現：
	// CmdLand 是**下馬**——執行段把 +0x1E 清掉並印記錄 0x11B「lands!」
	// （`0x2070f`…`0x20774`，RE-124）；CmdMountedMove 是騎著移動，執行段與
	// 一般移動同一支。
	CmdMove        = 0x27
	CmdLand        = 0x28
	CmdMountedMove = 0x3F
	// CmdRest 是休息，永遠是最後一格。
	CmdRest = 0x05
)

// 子選單的圖號。
const (
	// 武器攻擊式（`0x1F6BA`）。
	IconWeaponFormBase = 0x2A
	IconWeaponFormLast = 0x2E
	// 徒手攻擊式（`0x1F940`）：0x2A 落到選項 0，0x37…0x3A 是四種動作。
	IconUnarmedFormBase = 0x37
	IconUnarmedFormLast = 0x3A
	// 移動方式（`0x1F7D9`）。
	IconMoveWalk = 0x3C
	IconMoveNext = 0x3D
	IconMoveFast = 0x3E
	IconMoveAlt1 = 0x3F
	IconMoveAlt2 = 0x40
	IconMoveAlt3 = 0x35
	// 防禦方式（`sub_24AF8`）。
	IconGuardBase = 0x30
	IconGuardLast = 0x34
	IconGuardHigh = 0x36
	// 體型那一列（`sub_24A02`）。
	IconShoveWeak = 0x2F
	IconShove     = 0x41
	IconShoveHard = 0x42
)

// PartLost 是部位被打到廢掉的門檻：現值超過 0x7F 就算沒了。
const PartLost = 0x80

// IconOffset 是命令碼換成 KOLMPICS 圖號要加的數。原版 frame-887 的七格
// （命令碼 7、0x21、0x29、0x24、0x15、0x27、5）逐像素命中 KOLMPICS 的
// 10、36、44、39、24、42、8（RE-080）。
const IconOffset = 3

// weaponFormOption 是 `[0x0B42]`：武器攻擊式的圖號減 0x2A 之後查這張表。
var weaponFormOption = [5]int{0, 1, 2, 5, 8}

// facingCell 是 `[0x0B76]`：方向 0…7 對到 3×3 的格編號。
var facingCell = [9]int{1, 2, 5, 8, 7, 6, 3, 0, 4}

// cellFacing 是 `[0x0B6C]`：3×3 的格編號回推方向，中心格對到 8（無效）。
var cellFacing = [9]int{7, 0, 1, 6, 8, 2, 5, 4, 3}

// NoFacing 是 cellFacing 的中心格：不是任何一個方向。
const NoFacing = 8

// Cell 是戰場上的一格；Row 是畫面的列、Col 是畫面的欄。
//
// **與原版的軸是對調的**：原版參戰者記錄的 +0x1F 是欄、+0x20 是列（RE-095）。
// remake 從抽取到繪製一致地轉置，所以畫面正確；讀原版的位移表時要記得換過來。
type Cell struct{ Row, Col int }

// CommandEnv 是組指令列要用到、但不在 Actor 或 Member 身上的三件事。
type CommandEnv struct {
	// Ground 是地上東西的座標（`[0x2BAE]`／`[0x2D9E]`，十六筆）。
	Ground []Cell
	// Standable 是腳下那一格能不能站（`sub_24E2B(0, 8)`——環 0 的位移陣列
	// 除了八個方向之外還多一項 (0, 0)，索引 8 就是自己站的那一格）。
	Standable bool
	// Occupied 是有沒有別人也站在這一格（`sub_227C0`）。
	Occupied bool
}

// ArmsLost 數這一名廢掉幾隻手臂（`sub_228B7:0x228ED`…`0x2291B`）。
func (a *Actor) ArmsLost() int {
	n := 0
	if a.RightArm > PartLost {
		n++
	}
	if a.LeftArm > PartLost {
		n++
	}
	return n
}

// LegsUsable 回報還有沒有腿能走（`0x22CCC`…`0x22CED`）。
func (a *Actor) LegsUsable() bool {
	return a.RightLeg < PartLost || a.LeftLeg < PartLost
}

// CommandBar 依原版的順序組出這一名的指令列（`sub_228B7`）。
// 第一格固定是 CmdYes、最後一格固定是 CmdRest；中間各格的條件見 SPEC-010。
// 這一支只決定「哪幾格出現」，不決定選中之後做什麼。
func CommandBar(a *Actor, m *Member, env CommandEnv) []int {
	bar := []int{CmdYes}
	// 兩隻手臂都廢就整段武器類與丟撿都跳過，直接到移動段（`0x22920`）。
	if a.ArmsLost() != 2 {
		bar = appendWeaponCommands(bar, a, m)
		bar = appendItemCommands(bar, a, m, env)
	}
	bar = appendMoveCommands(bar, a, env)
	return append(bar, CmdRest)
}

// appendWeaponCommands 是 `0x2292D`…`0x22A6C`：投擲、射擊、武器與徒手四格。
func appendWeaponCommands(bar []int, a *Actor, m *Member) []int {
	if m.HasWeapon && m.Weapon.Code&1 == 1 {
		kind := m.Weapon.Kind
		// 雙手武器少一隻手臂就整組跳過（`0x2296D`）。
		twoHandBlocked := kind&2 != 0 && a.ArmsLost() != 0
		if !twoHandBlocked {
			if kind&0xE0 == 0 {
				// 種類高三位元為零這一段還要行動者的 +0x00 不是 0xFF（`0x22996`）。
				if a.Kind != 0xFF {
					if kind == 6 || kind == 0x0A || kind == 0x0B || a.Flag1D != 0 {
						bar = append(bar, CmdMissile)
						a.Flag1D = 1
					} else {
						bar = append(bar, CmdLoad)
					}
				}
			} else if a.Flag1E == 0 {
				bar = append(bar, CmdWeapon)
			}
		}
	}
	if a.Flag1E == 0 {
		bar = append(bar, CmdUnarmed)
	}
	return bar
}

// appendItemCommands 是 `0x22A6C`…`0x22C5B`：背包、四格物品動作與地上那一格。
func appendItemCommands(bar []int, a *Actor, m *Member, env CommandEnv) []int {
	// 沒有武器時才看背包裡有沒有東西（`0x22A82`）。
	if !m.HasWeapon {
		if m.PackCount > 0 {
			bar = append(bar, CmdCast)
		}
	}
	// 有武器、或沒武器但 +0x190 非零，才進這一段（`0x22ABD`）。
	if m.HasWeapon || m.HasSecondItem {
		// 有武器但傷害碼 bit 0 為零就跳過中間三格（`0x22AF6`）。
		if !(m.HasWeapon && m.Weapon.Code&1 == 0) {
			switch {
			case m.HasWeapon && m.HasSecondItem:
				if m.Weapon.Kind&0xE0 != 0 {
					bar = append(bar, CmdSwitchAlt2)
				}
			case m.HasSecondItem:
				bar = append(bar, CmdSwitchAlt1)
			case m.Weapon.Kind&0xE0 != 0:
				bar = append(bar, CmdSwitch)
			}
		}
		if m.HasWeapon {
			bar = append(bar, CmdDrop)
		}
	}
	if a.Flag1E == 0 {
		for _, g := range env.Ground {
			if g.Row == a.X && g.Col == a.Y {
				bar = append(bar, CmdPickUp)
				break
			}
		}
	}
	return bar
}

// appendMoveCommands 是 `0x22C5E`…`0x22CFE`：移動那一段分成兩條路。
func appendMoveCommands(bar []int, a *Actor, env CommandEnv) []int {
	if a.Flag1E != 0 {
		if env.Standable && !env.Occupied {
			bar = append(bar, CmdLand)
		}
		return append(bar, CmdMountedMove)
	}
	if a.LegsUsable() {
		bar = append(bar, CmdMove)
	}
	return bar
}

// WeaponForms 是武器攻擊式那一列（`0x1F6BA`）。
func WeaponForms() []int {
	forms := []int{CmdCancel}
	for icon := IconWeaponFormBase; icon <= IconWeaponFormLast; icon++ {
		forms = append(forms, icon)
	}
	return forms
}

// WeaponFormOption 把武器攻擊式的圖號換成攻擊選項（`0x1F727`）。
func WeaponFormOption(icon int) int {
	i := icon - IconWeaponFormBase
	if i < 0 || i >= len(weaponFormOption) {
		return 0
	}
	return weaponFormOption[i]
}

// UnarmedForms 是徒手攻擊式那一列（`0x1F940`）。
func UnarmedForms() []int {
	forms := []int{CmdCancel, IconWeaponFormBase}
	for icon := IconUnarmedFormBase; icon <= IconUnarmedFormLast; icon++ {
		forms = append(forms, icon)
	}
	return forms
}

// UnarmedFormOption 把徒手攻擊式的圖號換成攻擊選項（`0x1F9AD`）：
// 圖號減 0x36，負數夾成 0，所以 0x2A 是 0、0x37…0x3A 是 1…4。
func UnarmedFormOption(icon int) int {
	if v := icon - 0x36; v > 0 {
		return v
	}
	return 0
}

// MoveForms 是移動方式那一列（`0x1F7D9`）。groundTile 是腳下那一塊磚。
func MoveForms(a *Actor, groundTile int) []int {
	forms := []int{CmdCancel}
	if a.Flag1E == 0 {
		forms = append(forms, IconMoveWalk, IconMoveNext)
		if a.Flag03 != 0 {
			forms = append(forms, IconMoveFast)
		}
	}
	if a.Flag03 == 2 || a.Flag1E != 0 {
		// `loc_1F87E`：腳下是 0x0C 而且 +0x1E 為零就跳過；不是 0x0C 或已經非零
		// 才推（RE-148 更正早先「0x0C 才解鎖」的讀法）。
		if groundTile != StuckTile || a.Flag1E != 0 {
			forms = append(forms, IconMoveAlt1, IconMoveAlt2, IconMoveAlt3)
		}
	}
	return forms
}


// MoveGuard 是選了移動方式之後的防禦選項（`0x1F8F8`）：
// 走路那一格給 2、其餘 0。敵人 AI 的輪盤也是同一條規則（RE-081）。
func MoveGuard(icon int) int {
	if icon == IconMoveWalk {
		return 2
	}
	return 0
}

// GuardForms 是防禦方式那一列（`sub_24AF8`）。
func GuardForms() []int {
	forms := []int{CmdCancel, IconWeaponFormBase}
	for icon := IconGuardBase; icon <= IconGuardLast; icon++ {
		forms = append(forms, icon)
	}
	return append(forms, IconGuardHigh)
}

// GuardOption 把防禦方式的圖號換成防禦選項（`0x24B85`）：圖號減 0x2F，
// 負數夾成 0，等於 7 再減 1——0x35 不在列上，所以 0x36 補回 6。
func GuardOption(icon int) int {
	v := icon - 0x2F
	if v < 0 {
		v = 0
	}
	if v == 7 {
		v--
	}
	return v
}

// ResolveAttack 是武器與徒手兩列共用的收尾（`0x1F775`／`0x1F9D2`）：
// 防禦選項為 1 時攻擊選項歸零；兩個都是 0 時整個指令改成休息。
func ResolveAttack(command, style, guard int) (int, int, int) {
	if guard == 1 {
		style = 0
	}
	if style == 0 && guard == 0 {
		command = CmdRest
	}
	return command, style, guard
}

// FacingCell 回傳面向某個方向時，3×3 裡是哪一格（`[0x0B76]`）。
func FacingCell(facing int) int {
	if facing < 0 || facing >= len(facingCell) {
		return facingCell[len(facingCell)-1]
	}
	return facingCell[facing]
}

// CellFacing 把 3×3 的格編號回推成方向（`[0x0B6C]`）；中心格回 NoFacing。
func CellFacing(cell int) int {
	if cell < 0 || cell >= len(cellFacing) {
		return NoFacing
	}
	return cellFacing[cell]
}

// CommandEnv 填出組指令列要用的三件事（`sub_228B7` 讀的那幾個外部狀態）。
//
// **這一支原本沒有人呼叫**，`Plan.Env` 一直是零值，所以撿那一格從來沒出現過
// ——地上明明有東西也選不到（RE-113 實跑才看得出來）。
func (s *State) CommandEnv(i int) CommandEnv {
	if i < 0 || i >= len(s.Actors) {
		return CommandEnv{}
	}
	a := s.Actors[i]
	out := CommandEnv{Standable: true}
	for _, g := range s.Ground {
		if g.Name != "" {
			out.Ground = append(out.Ground, Cell{Row: g.X, Col: g.Y})
		}
	}
	if s.Field != nil {
		out.Standable = !BlocksMovement(s.Field.At(a.X, a.Y))
	}
	// 腳下那一格有沒有別人（`sub_227C0`）：下馬那一格要的是「沒有別人擋著」。
	for n := range s.Actors {
		if n == i {
			continue
		}
		o := s.Actors[n]
		if o.Side != SideEmpty && !o.Down() && o.X == a.X && o.Y == a.Y {
			out.Occupied = true
			break
		}
	}
	return out
}

// 種族（角色記錄 +0x24）。
const (
	RaceHuman   = 0
	RaceElven   = 1
	RaceDwarven = 2
	RaceKelden  = 3
)

// MoveAbility 是 `sub_21D94` 依種族寫進參戰者記錄 `+0x03` 的值（RE-148）：
// Kelden 2、Dwarven 0、其餘 1。移動方式那一列 1 以上多一格 0x3E，等於 2 而且
// 腳下是 0x0C 那一塊時再多 0x3F、0x40、0x35（`MoveForms`）。
func MoveAbility(race int) int {
	switch race {
	case RaceKelden:
		return 2
	case RaceDwarven:
		return 0
	}
	return 1
}
