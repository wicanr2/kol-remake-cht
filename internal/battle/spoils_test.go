package battle

import "testing"

func spoilState() *State {
	s := &State{Party: []Member{
		{Name: "POONTANG"},
		{Name: "DDDUHH", HasWeapon: true, WeaponName: "MACE"},
	}}
	s.Party[1].Worn[0] = "MACE"
	s.Actors = []Actor{
		{Side: SideParty, Member: 0, Head: 10, Chest: 10, Threshold: 100},
		{Side: SideParty, Member: 1, Head: 10, Chest: 10, Threshold: 100},
		{Side: SideMonster, Head: 0, Chest: 10, Threshold: 100},
		{Side: SideMonster, Head: 0, Chest: 10, Threshold: 100},
	}
	return s
}

var spoilTable = []SpoilItem{
	{Index: 0, Name: "", Empty: true},
	{Index: 1, Name: "Dagger"},
	{Index: 2, Name: "Shortsword"},
}

// 第一擲不到門檻就整段跳過，一件都不掉。
func TestDropSpoilsMissesTheThreshold(t *testing.T) {
	s := spoilState()
	if got := s.DropSpoils(spoilTable, &fixedRand{values: []int{SpoilChance}}); got != nil {
		t.Fatalf("沒過門檻不該掉東西，得到 %v", got)
	}
	if len(s.SpoilSlots()) != 0 {
		t.Fatal("地面應該還是空的")
	}
}

// 過了門檻至少掉一件：亂數回 1…n，所以件數不會是 0。
func TestDropSpoilsTakesEmptySlots(t *testing.T) {
	s := spoilState()
	// 門檻過、掉兩件、第一件抽第 2 筆、第二件抽第 3 筆，中間各夾一次狀況骰。
	rng := &fixedRand{values: []int{1, 2, 2, 255, 3, 255}}
	got := s.DropSpoils(spoilTable, rng)
	if len(got) != 2 || got[0] != 0 || got[1] != 1 {
		t.Fatalf("占用的格號是 %v，期望 [0 1]", got)
	}
	if s.Ground[0].Name != "Dagger" || s.Ground[1].Name != "Shortsword" {
		t.Fatalf("兩格是 %q 與 %q", s.Ground[0].Name, s.Ground[1].Name)
	}
	// 掉落物沒有座標，原版一個字都沒寫進那兩張表。
	if s.Ground[0].X != 0 || s.Ground[0].Y != 0 {
		t.Fatalf("掉落物不該有座標，得到 (%d,%d)", s.Ground[0].X, s.Ground[0].Y)
	}
}

// 抽中第 0 筆（全零的空位）等於什麼都沒掉，但狀況那一骰照擲。
func TestDropSpoilsEmptyEntryYieldsNothing(t *testing.T) {
	s := spoilState()
	rng := &fixedRand{values: []int{1, 1, 1, 255}}
	if got := s.DropSpoils(spoilTable, rng); len(got) != 0 {
		t.Fatalf("抽中空位不該占格，得到 %v", got)
	}
	if rng.i != 4 {
		t.Fatalf("擲了 %d 次，期望 4 次（門檻、件數、品項、狀況）", rng.i)
	}
}

// 清單把戰鬥中丟在地上的東西一起算進去——打完之後那十六格就是戰利品池。
func TestSpoilSlotsIncludeItemsDroppedInBattle(t *testing.T) {
	s := spoilState()
	s.Ground[3] = GroundItem{Name: "LONG SPEAR", X: 5, Y: 6}
	s.Ground[7] = GroundItem{Name: "Dagger"}
	got := s.SpoilSlots()
	if len(got) != 2 || got[0] != 3 || got[1] != 7 {
		t.Fatalf("清單是 %v，期望 [3 7]", got)
	}
}

// 倒下的隊員領不到東西（`sub_22866`：頭或軀幹歸零）。
func TestSpoilTakersSkipTheFallen(t *testing.T) {
	s := spoilState()
	s.Actors[0].Chest = 0
	got := s.SpoilTakers()
	if len(got) != 1 || got[0] != 1 {
		t.Fatalf("候選是 %v，期望 [1]", got)
	}
}

// 領取：先放背包的第一個空格，那一格清掉，沒有話要說。
func TestTakeSpoilGoesToTheBackpack(t *testing.T) {
	s := spoilState()
	s.Party[0].Pack[0] = "Rope"
	s.Ground[0] = GroundItem{Name: "Dagger"}
	name, record := s.TakeSpoil(0, 0)
	if name != "Dagger" || record != 0 {
		t.Fatalf("領到 %q 訊息 %#x，期望 Dagger 與 0", name, record)
	}
	if s.Party[0].Pack[1] != "Dagger" {
		t.Fatalf("背包第 1 格是 %q，期望 Dagger", s.Party[0].Pack[1])
	}
	if s.Ground[0].Name != "" {
		t.Fatal("領走之後那一格要清掉")
	}
}

// 背包滿了就與手上那件交換：手上原本那件掉到地上那一格，並印記錄 0x183。
func TestTakeSpoilSwapsWithTheHandWhenFull(t *testing.T) {
	s := spoilState()
	for i := range s.Party[1].Pack {
		s.Party[1].Pack[i] = "Rope"
	}
	s.Ground[0] = GroundItem{Name: "Dagger"}
	name, record := s.TakeSpoil(0, 1)
	if name != "Dagger" || record != EmptyHandsRecord {
		t.Fatalf("領到 %q 訊息 %#x，期望 Dagger 與 %#x",
			name, record, EmptyHandsRecord)
	}
	if s.Party[1].WeaponName != "Dagger" || !s.Party[1].HasWeapon {
		t.Fatalf("手上那件是 %q，期望 Dagger", s.Party[1].WeaponName)
	}
	if s.Ground[0].Name != "MACE" {
		t.Fatalf("換下來的那一格是 %q，期望 MACE", s.Ground[0].Name)
	}
}

// 背包滿而手上空著：照樣交換，只是換過去的是空的，也就不必印 0x183。
func TestTakeSpoilIntoAnEmptyHand(t *testing.T) {
	s := spoilState()
	for i := range s.Party[0].Pack {
		s.Party[0].Pack[i] = "Rope"
	}
	s.Ground[0] = GroundItem{Name: "Dagger"}
	name, record := s.TakeSpoil(0, 0)
	if name != "Dagger" || record != 0 {
		t.Fatalf("領到 %q 訊息 %#x，期望 Dagger 與 0", name, record)
	}
	if s.Ground[0].Name != "" {
		t.Fatalf("換過去的該是空的，得到 %q", s.Ground[0].Name)
	}
}

// 清單空的就沒有這個畫面。
func TestBeginSpoilsWithNothingToTake(t *testing.T) {
	if spoilState().BeginSpoils() {
		t.Fatal("地面是空的，不該進戰利品畫面")
	}
}

// 只有一位隊員收得下就直接給他，不出第二張清單；領完回到第一張。
func TestSpoilConfirmWithASingleTaker(t *testing.T) {
	s := spoilState()
	s.Actors[1].Chest = 0 // 倒下的那一位不算候選
	s.Ground[0] = GroundItem{Name: "Dagger"}
	s.Ground[4] = GroundItem{Name: "Mace"}
	if !s.BeginSpoils() {
		t.Fatal("有兩件東西，應該進得了戰利品畫面")
	}
	name, record, done := s.SpoilConfirm()
	if name != "Dagger" || record != 0 || done {
		t.Fatalf("領到 %q 訊息 %#x 結束 %v，期望 Dagger、0、false", name, record, done)
	}
	if s.Party[0].Pack[0] != "Dagger" {
		t.Fatalf("背包第 0 格是 %q", s.Party[0].Pack[0])
	}
	// 回到第一張清單，剩下的那一件還在。
	if s.Spoil.Phase != SpoilPickItem {
		t.Fatal("領完應該回到物品清單")
	}
	if got := s.SpoilLines(); len(got) != 1 || got[0] != "Mace" {
		t.Fatalf("剩下的清單是 %v，期望 [Mace]", got)
	}
	// 最後一件領完就結束。
	if _, _, done := s.SpoilConfirm(); !done {
		t.Fatal("領完最後一件應該結束")
	}
}

// 兩位以上收得下就出第二張清單；那一關取消只是退回第一張，不結束。
func TestSpoilTakerListAndCancel(t *testing.T) {
	s := spoilState()
	s.Ground[2] = GroundItem{Name: "Dagger"}
	if !s.BeginSpoils() {
		t.Fatal("應該進得了戰利品畫面")
	}
	if _, _, done := s.SpoilConfirm(); done {
		t.Fatal("兩位候選時不該直接結束")
	}
	if s.Spoil.Phase != SpoilPickTaker {
		t.Fatal("應該進到隊員清單")
	}
	if got := s.SpoilLines(); len(got) != 2 || got[0] != "POONTANG" || got[1] != "DDDUHH" {
		t.Fatalf("隊員清單是 %v", got)
	}
	if s.SpoilTitle() != CharactersRecord {
		t.Fatalf("標題是 %#x，期望 %#x", s.SpoilTitle(), CharactersRecord)
	}
	if s.SpoilCancel() {
		t.Fatal("隊員那一關取消只該退回物品清單")
	}
	if s.Spoil.Phase != SpoilPickItem {
		t.Fatal("取消之後應該回到物品清單")
	}
	// 東西還在。
	if got := s.SpoilLines(); len(got) != 1 || got[0] != "Dagger" {
		t.Fatalf("物品清單是 %v，期望 [Dagger]", got)
	}
	// 物品那一關取消才是結束。
	if !s.SpoilCancel() {
		t.Fatal("物品那一關取消應該結束")
	}
}

// 選第二位隊員：游標移動之後領給他。
func TestSpoilMoveAndTake(t *testing.T) {
	s := spoilState()
	s.Ground[0] = GroundItem{Name: "Dagger"}
	s.BeginSpoils()
	s.SpoilConfirm()
	s.SpoilMove(1)
	name, _, done := s.SpoilConfirm()
	if name != "Dagger" || !done {
		t.Fatalf("領到 %q 結束 %v，期望 Dagger、true", name, done)
	}
	if s.Party[1].Pack[0] != "Dagger" {
		t.Fatalf("第二位的背包是 %q", s.Party[1].Pack[0])
	}
}

// 游標不會越過兩端。
func TestSpoilMoveClamps(t *testing.T) {
	s := spoilState()
	s.Ground[0] = GroundItem{Name: "Dagger"}
	s.Ground[1] = GroundItem{Name: "Mace"}
	s.BeginSpoils()
	s.SpoilMove(-5)
	if s.Spoil.Cursor != 0 {
		t.Fatalf("往上越界後游標是 %d，期望 0", s.Spoil.Cursor)
	}
	s.SpoilMove(9)
	if s.Spoil.Cursor != 1 {
		t.Fatalf("往下越界後游標是 %d，期望 1", s.Spoil.Cursor)
	}
}
