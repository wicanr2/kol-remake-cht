package battle

import "testing"

func TestShotConsumesAmmunition(t *testing.T) {
	for _, kind := range []int{6, 18} {
		for _, wall := range []bool{false, true} {
			s := shootState()
			s.Plan.Memo = map[int]Cell{0: {Row: 10, Col: 15}}
			a := &s.Actors[0]
			a.Kind, a.Weapon.Kind, a.Command = 1, kind, CmdMissile
			if wall {
				s.Field.Tiles[10*FieldSide+13] = 0x29
			}
			if NeedsReload(kind) {
				s.shoot(0, &fixedRoll{values: []int{50}})
				if a.Kind != 1 || a.Flag1D != 1 {
					t.Fatal("裝填不可扣彈藥")
				}
			}
			for shot := 0; shot < 2; shot++ {
				a.Command, a.Flag1D = CmdMissile, 1
				result := s.shoot(0, &fixedRoll{values: []int{50}})
				if wall && len(result.Records) == 0 {
					t.Fatal("未覆蓋撞牆路徑")
				}
				if !wall && result.Message != MissileNothingRecord {
					t.Fatal("未覆蓋射空路徑")
				}
				want := 0
				if shot == 1 {
					want = 255
				}
				if a.Kind != want {
					t.Fatalf("第 %d 發後彈藥原值 %d，預期 %d", shot+1, a.Kind, want)
				}
				if shot == 1 && a.Command != 0 {
					t.Fatal("耗盡後未清除指令")
				}
				if shot == 0 && NeedsReload(kind) && (a.Flag1D != 0 || a.Command != CmdLoad) {
					t.Fatal("射後未返回裝填")
				}
			}
		}
	}
	monster := Actor{Side: SideMonster, Kind: 19, Command: CmdMissile}
	monster.ConsumeAmmunition()
	if monster.Kind != 18 {
		t.Fatal("怪物射後未依RE-217扣除計數")
	}
}
