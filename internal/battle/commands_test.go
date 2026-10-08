package battle

import "testing"

// 一名手無寸鐵、四肢健全的隊員：指令列只剩徒手、移動與休息。
func TestCommandBarBareHanded(t *testing.T) {
	a := &Actor{Threshold: 100}
	a.SetParts()
	m := &Member{}
	got := CommandBar(a, m, CommandEnv{})
	want := []int{CmdYes, CmdUnarmed, CmdMove, CmdRest}
	assertBar(t, got, want)
}

// 原版 frame-887 的主指令列是 7、0x21、0x29、0x24、0x15、0x27、5
// ——也就是 YES、武器攻擊、徒手、**切換**、**丟**、移動、休息（RE-094 更正了
// 0x24 與 0x15 的語意，值沒變）。
// （RE-080 逐像素對過 KOLMPICS 的 10、36、44、39、24、42、8）。
//
// 這裡的武器數字是**從那一列反推出來的最小條件**（傷害碼 bit 0 為 1、種類高三位元
// 非零的單手武器，四肢健全），不是從原版存檔讀回來的那一名的實際裝備。所以這一支
// 只證明「條件組合能產生原版那一列」，不證明 parity——真正的對拍要走截圖收據。
func TestCommandBarProducesOriginalFrameLayout(t *testing.T) {
	a := &Actor{Threshold: 100}
	a.SetParts()
	m := &Member{HasWeapon: true, Weapon: Weapon{Kind: 0x20, Code: 1}}
	got := CommandBar(a, m, CommandEnv{})
	want := []int{CmdYes, CmdWeapon, CmdUnarmed, CmdSwitch, CmdDrop, CmdMove, CmdRest}
	assertBar(t, got, want)
}

// 兩隻手臂都廢就整段武器類與物品都跳過（`sub_228B7:0x22920`）。
func TestCommandBarBothArmsLost(t *testing.T) {
	a := &Actor{Threshold: 100}
	a.SetParts()
	a.RightArm, a.LeftArm = 0xFF, 0xFF
	m := &Member{HasWeapon: true, Weapon: Weapon{Kind: 0x20, Code: 1}}
	got := CommandBar(a, m, CommandEnv{})
	assertBar(t, got, []int{CmdYes, CmdMove, CmdRest})
}

// 雙手武器少一隻手臂就整組武器攻擊跳過（`0x2296D`），徒手還在。
func TestCommandBarTwoHandNeedsBothArms(t *testing.T) {
	a := &Actor{Threshold: 100}
	a.SetParts()
	a.RightArm = 0xFF
	m := &Member{HasWeapon: true, Weapon: Weapon{Kind: 0x22, Code: 1}}
	got := CommandBar(a, m, CommandEnv{})
	assertBar(t, got, []int{CmdYes, CmdUnarmed, CmdSwitch, CmdDrop, CmdMove, CmdRest})
}

// 種類 6、0x0A、0x0B 直接給射擊，並把 +0x1D 設起來（`0x229B0`、`0x229F5`）。
func TestCommandBarMissileKinds(t *testing.T) {
	for _, kind := range []int{6, 0x0A, 0x0B} {
		a := &Actor{Threshold: 100}
		a.SetParts()
		m := &Member{HasWeapon: true, Weapon: Weapon{Kind: kind, Code: 1}}
		got := CommandBar(a, m, CommandEnv{})
		want := []int{CmdYes, CmdMissile, CmdUnarmed, CmdDrop, CmdMove, CmdRest}
		assertBar(t, got, want)
		if a.Flag1D != 1 {
			t.Fatalf("種類 %#x 推出射擊之後 Flag1D 應該是 1，得到 %d", kind, a.Flag1D)
		}
	}
}

// 種類低段且沒射過就是投擲；射過一次之後同一件武器改成射擊（`0x229CB`）。
func TestCommandBarThrowThenMissile(t *testing.T) {
	a := &Actor{Threshold: 100}
	a.SetParts()
	m := &Member{HasWeapon: true, Weapon: Weapon{Kind: 4, Code: 1}}
	if got := CommandBar(a, m, CommandEnv{}); got[1] != CmdLoad {
		t.Fatalf("第一次應該是投擲，得到 %#x", got[1])
	}
	a.Flag1D = 1
	if got := CommandBar(a, m, CommandEnv{}); got[1] != CmdMissile {
		t.Fatalf("Flag1D 之後應該是射擊，得到 %#x", got[1])
	}
}

// 沒有武器但背包裡有東西才有背包那一格（`0x22A82`）。
func TestCommandBarPackNeedsEmptyHand(t *testing.T) {
	a := &Actor{Threshold: 100}
	a.SetParts()
	got := CommandBar(a, &Member{PackCount: 3}, CommandEnv{})
	assertBar(t, got, []int{CmdYes, CmdUnarmed, CmdCast, CmdMove, CmdRest})
	// 手上有武器就不掃背包，即使背包裡有東西。
	m := &Member{HasWeapon: true, PackCount: 3, Weapon: Weapon{Kind: 0x20, Code: 1}}
	got = CommandBar(a, m, CommandEnv{})
	for _, icon := range got {
		if icon == CmdCast {
			t.Fatal("手上有武器時不該出現背包那一格")
		}
	}
}

// 地上的東西要與自己同格才算（`0x22C0A`）。
func TestCommandBarGroundNeedsSameCell(t *testing.T) {
	a := &Actor{Threshold: 100, X: 12, Y: 34}
	a.SetParts()
	m := &Member{}
	env := CommandEnv{Ground: []Cell{{Row: 12, Col: 35}}}
	for _, icon := range CommandBar(a, m, env) {
		if icon == CmdPickUp {
			t.Fatal("不同格不該出現地上那一格")
		}
	}
	env.Ground = append(env.Ground, Cell{Row: 12, Col: 34})
	got := CommandBar(a, m, env)
	assertBar(t, got, []int{CmdYes, CmdUnarmed, CmdPickUp, CmdMove, CmdRest})
}

// Flag1E 非零時換掉整組近戰與移動（`0x22C5E`）。
func TestCommandBarFlag1E(t *testing.T) {
	a := &Actor{Threshold: 100, Flag1E: 1}
	a.SetParts()
	m := &Member{HasWeapon: true, Weapon: Weapon{Kind: 0x20, Code: 1}}
	got := CommandBar(a, m, CommandEnv{Standable: true})
	assertBar(t, got, []int{CmdYes, CmdSwitch, CmdDrop, CmdLand, CmdMountedMove, CmdRest})
	// 腳下站不了、或有別人同格，就下不了馬，只剩騎著移動那一格。
	got = CommandBar(a, m, CommandEnv{Standable: true, Occupied: true})
	assertBar(t, got, []int{CmdYes, CmdSwitch, CmdDrop, CmdMountedMove, CmdRest})
}

// 兩條腿都廢就沒有移動那一格（`0x22CCC`）。
func TestCommandBarNoLegs(t *testing.T) {
	a := &Actor{Threshold: 100}
	a.SetParts()
	a.RightLeg, a.LeftLeg = 0xFF, 0xFF
	got := CommandBar(a, &Member{}, CommandEnv{})
	assertBar(t, got, []int{CmdYes, CmdUnarmed, CmdRest})
}

// 四列子選單的選項換算（`0x1F727`、`0x1F9AD`、`0x1F8F8`、`0x24B85`）。
func TestFormOptions(t *testing.T) {
	weapon := map[int]int{0x2A: 0, 0x2B: 1, 0x2C: 2, 0x2D: 5, 0x2E: 8}
	for icon, want := range weapon {
		if got := WeaponFormOption(icon); got != want {
			t.Errorf("武器攻擊式 %#x：得到 %d，應該是 %d", icon, got, want)
		}
	}
	unarmed := map[int]int{0x2A: 0, 0x37: 1, 0x38: 2, 0x39: 3, 0x3A: 4}
	for icon, want := range unarmed {
		if got := UnarmedFormOption(icon); got != want {
			t.Errorf("徒手攻擊式 %#x：得到 %d，應該是 %d", icon, got, want)
		}
	}
	guard := map[int]int{0x2A: 0, 0x30: 1, 0x31: 2, 0x32: 3, 0x33: 4, 0x34: 5, 0x36: 6}
	for icon, want := range guard {
		if got := GuardOption(icon); got != want {
			t.Errorf("防禦方式 %#x：得到 %d，應該是 %d", icon, got, want)
		}
	}
	if MoveGuard(IconMoveWalk) != 2 || MoveGuard(IconMoveNext) != 0 {
		t.Error("移動方式的防禦選項應該只有走路那一格是 2")
	}
}

// 移動方式那一列分兩組（`0x1F7EA`…`0x1F8D0`）。
func TestMoveForms(t *testing.T) {
	a := &Actor{}
	assertBar(t, MoveForms(a, 0), []int{CmdCancel, IconMoveWalk, IconMoveNext})
	a.Flag03 = 1
	assertBar(t, MoveForms(a, 0),
		[]int{CmdCancel, IconMoveWalk, IconMoveNext, IconMoveFast})
	// Flag03 為 2 時後三格出現，但腳下是 0x0C 那一塊就不出現（`loc_1F87E`，RE-148）。
	a.Flag03 = 2
	assertBar(t, MoveForms(a, 0),
		[]int{CmdCancel, IconMoveWalk, IconMoveNext, IconMoveFast,
			IconMoveAlt1, IconMoveAlt2, IconMoveAlt3})
	assertBar(t, MoveForms(a, StuckTile),
		[]int{CmdCancel, IconMoveWalk, IconMoveNext, IconMoveFast})
	// Flag1E 非零時前三格消失、後三格無條件出現。
	a.Flag1E = 1
	assertBar(t, MoveForms(a, 0),
		[]int{CmdCancel, IconMoveAlt1, IconMoveAlt2, IconMoveAlt3})
}

// 攻擊式與防禦式的收尾（`0x1F775`、`0x1F9D2`）。
func TestResolveAttack(t *testing.T) {
	// 防禦選項 1 只把攻擊選項歸零；指令不變，因為第二條規則要兩個都是 0。
	if cmd, style, _ := ResolveAttack(CmdWeapon, 5, 1); style != 0 || cmd != CmdWeapon {
		t.Errorf("防禦選項 1 應該只把攻擊選項歸零，得到 %#x／%d", cmd, style)
	}
	if cmd, _, _ := ResolveAttack(CmdWeapon, 0, 0); cmd != CmdRest {
		t.Errorf("兩個選項都是 0 應該改成休息，得到 %#x", cmd)
	}
	if cmd, style, guard := ResolveAttack(CmdWeapon, 2, 3); cmd != CmdWeapon ||
		style != 2 || guard != 3 {
		t.Errorf("其餘情形不該改動，得到 %#x／%d／%d", cmd, style, guard)
	}
}

// 兩張方向表互為反表，中心格對到 NoFacing。
func TestFacingTablesAreInverse(t *testing.T) {
	for facing := 0; facing < 8; facing++ {
		if got := CellFacing(FacingCell(facing)); got != facing {
			t.Errorf("方向 %d 往返得到 %d", facing, got)
		}
	}
	if CellFacing(4) != NoFacing {
		t.Errorf("中心格應該回 NoFacing，得到 %d", CellFacing(4))
	}
}

func assertBar(t *testing.T, got, want []int) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("指令列長度 %d，應該是 %d：%#x vs %#x", len(got), len(want), got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("第 %d 格是 %#x，應該是 %#x：%#x vs %#x", i, got[i], want[i], got, want)
		}
	}
}
