package battle

import (
	"reflect"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/rng"
)

// RE-219 正常第五回合：原版人物 +1F/+20 為 31/30，擊退至 30/29。
// 使用原版兩次骰的入口種子；此局部回歸不取代正常八回合 GUI。
func TestKnockbackNormalOriginalRoundFive(t *testing.T) {
	s := &State{Field: NewField(), Actors: []Actor{
		{Side: SideMonster, Member: -1, X: 32, Y: 30, Size: 232, Head: 30, Chest: 30, Threshold: 206},
		{Side: SideParty, Member: -1, X: 31, Y: 30, Size: 116, Head: 16, Chest: 16, Threshold: 109},
	}}
	random := rng.NewState(1535194068)
	got := s.Knockback(0, 1, random)
	if s.Actors[1].X != 30 || s.Actors[1].Y != 29 || random.State() != 3413147886 {
		t.Fatalf("原版擊退座標30/29、seed3413147886；得到%d/%d、seed%d",
			s.Actors[1].X, s.Actors[1].Y, random.State())
	}
	if len(got) != 2 || got[0] != KnocksRecord || got[1] != BackwardsRecord {
		t.Fatalf("原版空手擊退訊息不符：%v", got)
	}
}

// RE-222 正常第七回合：方向6先檢查30/28，成功後原版更新為29/29。
func TestKnockbackNormalOriginalRoundSeven(t *testing.T) {
	s := &State{Field: NewField(), Actors: []Actor{
		{Side: SideMonster, Member: -1, X: 31, Y: 29, Size: 232, Head: 30, Chest: 30, Threshold: 206},
		{Side: SideParty, Member: -1, X: 30, Y: 29, Size: 116, Head: 16, Chest: 16, Threshold: 109},
	}}
	random := rng.NewState(1854365306)
	got := s.Knockback(0, 1, random)
	if s.Actors[1].X != 29 || s.Actors[1].Y != 29 || random.State() != 2304995524 {
		t.Fatalf("原版擊退座標29/29、seed2304995524；得到%d/%d、seed%d",
			s.Actors[1].X, s.Actors[1].Y, random.State())
	}
	if len(got) != 2 || got[0] != KnocksRecord || got[1] != BackwardsRecord {
		t.Fatalf("原版空手擊退訊息不符：%v", got)
	}
}

func knockState(attackerSize, targetSize int) *State {
	s := &State{Field: &Field{Tiles: make([]int, FieldSide*FieldSide)},
		Party: []Member{{Name: "POONTANG", HasWeapon: true, WeaponName: "LONG SPEAR"}},
		Actors: []Actor{
			{Side: SideMonster, X: 10, Y: 10, Size: attackerSize, Member: -1,
				Head: 10, Chest: 10, Threshold: 100},
			{Side: SideParty, Member: 0, X: 10, Y: 11, Size: targetSize,
				Head: 10, Chest: 10, Threshold: 100, Weapon: Weapon{Kind: 0x20},
				Command: CmdWeapon},
		}}
	return s
}

// 體型差是機率：差多少就是百分之多少。同體型或更小推不動。
func TestKnockbackNeedsSizeAdvantage(t *testing.T) {
	s := knockState(100, 100)
	if got := s.Knockback(0, 1, &fixedRand{values: []int{1}}); got != nil {
		t.Fatalf("同體型不該推得動，得到 %v", got)
	}
	s = knockState(100, 150)
	if got := s.Knockback(0, 1, &fixedRand{values: []int{1}}); got != nil {
		t.Fatalf("比對手小不該推得動，得到 %v", got)
	}
	// 體型差 40：rand(100) 擲出 41 以上就不成立。
	s = knockState(140, 100)
	if got := s.Knockback(0, 1, &fixedRand{values: []int{41}}); got != nil {
		t.Fatalf("擲出 41 不該推得動（差 40），得到 %v", got)
	}
}

// 推得動：目標退一格，印 0x159，而且武器掉了、這一回合的手取消。
func TestKnockbackPushesTargetAway(t *testing.T) {
	s := knockState(140, 100)
	got := s.Knockback(0, 1, &fixedRand{values: []int{1}})
	if len(got) == 0 || got[0] != KnocksRecord {
		t.Fatalf("第一句應該是 %#x，得到 %v", KnocksRecord, got)
	}
	d := s.Actors[1]
	if d.X == 10 && d.Y == 11 {
		t.Fatal("推得動的話目標應該換一格")
	}
	// 目標在攻擊者的 Y 正向，所以會被往 Y 正向推開，不會被推向攻擊者。
	if d.Y < 11 {
		t.Fatalf("被推向攻擊者了：(%d,%d)", d.X, d.Y)
	}
	if s.Actors[1].Weapon.Kind != 0 || s.Party[0].HasWeapon {
		t.Fatal("被擊退的人應該掉了武器")
	}
	if s.Actors[1].Command != 0 {
		t.Fatal("被擊退的人這一回合的手應該取消")
	}
}

// 推不動（那一格有人）就變成打暈，印 0x158。
func TestKnockbackBecomesStunWhenBlocked(t *testing.T) {
	s := knockState(140, 100)
	// 把退路堵住：目標會被推向 Y 正向。
	for y := 10; y <= 12; y++ {
		for x := 9; x <= 11; x++ {
			if x == 10 && y == 10 || x == 10 && y == 11 {
				continue
			}
			s.Actors = append(s.Actors, Actor{Side: SideMonster, Member: -1,
				X: x, Y: y, Head: 10, Chest: 10, Threshold: 100})
		}
	}
	got := s.Knockback(0, 1, &fixedRand{values: []int{1}})
	if len(got) != 2 || got[0] != StunsRecord || got[1] != FallsRecord {
		t.Fatalf("堵住了應該印 %#x＋%#x，得到 %v", StunsRecord, FallsRecord, got)
	}
	if s.Actors[1].X != 10 || s.Actors[1].Y != 11 {
		t.Fatal("推不動就不該移動")
	}
}

// 手上空空的人被推不動時什麼都不印（`0x27DD4`）。
func TestKnockbackSilentWhenUnarmedAndBlocked(t *testing.T) {
	s := knockState(140, 100)
	s.Actors[1].Weapon = Weapon{}
	for y := 10; y <= 12; y++ {
		for x := 9; x <= 11; x++ {
			if x == 10 && y == 10 || x == 10 && y == 11 {
				continue
			}
			s.Actors = append(s.Actors, Actor{Side: SideMonster, Member: -1,
				X: x, Y: y, Head: 10, Chest: 10, Threshold: 100})
		}
	}
	if got := s.Knockback(0, 1, &fixedRand{values: []int{1}}); got != nil {
		t.Fatalf("空手又推不動應該什麼都不印，得到 %v", got)
	}
}

// RE-219：+1F 對應 X 主索引，+20 對應 Y 增量。
func TestKnockbackIndex(t *testing.T) {
	at := Actor{X: 10, Y: 10}
	for _, c := range []struct {
		name string
		t    Actor
		want int
	}{
		{"正右（欄大）", Actor{X: 10, Y: 11}, 6},
		{"正左（欄小）", Actor{X: 10, Y: 9}, 5},
		{"正上（列小）", Actor{X: 9, Y: 10}, 0},
		{"正下（列大）", Actor{X: 11, Y: 10}, 8},
		{"右下", Actor{X: 11, Y: 11}, 10},
	} {
		if got := KnockbackIndex(at, c.t); got != c.want {
			t.Fatalf("%s 的索引是 %d，期望 %d", c.name, got, c.want)
		}
	}
}

// SPEC-008 2026-10-07：RE-253／254 的受阻分支不呼叫 sub_269BA。
// 正常 seed2 Olear 的完整物品記錄只作不透明資料，語意由正常路徑另驗。
func TestBlockedStunKeepsCompleteHeldItem(t *testing.T) {
	for _, monster := range []bool{false, true} {
		for _, terrain := range []bool{false, true} {
			label := "隊員／人物阻擋"
			if monster {
				label = "怪物／人物阻擋"
			}
			if terrain {
				label += "／地形阻擋"
			}
			t.Run(label, func(t *testing.T) {
				s := knockState(255, 1)
				item := &Item{Name: "Elf Bow", Record: "72b0c013ee000000225828ff190be200"}
				s.ReturnScratch = 6
				if monster {
					s.Party = nil
					s.Actors[1].Side, s.Actors[1].Member = SideMonster, -1
					s.Actors[1].HandItem = item
				} else {
					s.Party[0].HandItem, s.Party[0].WeaponName = item, item.Name
				}
				for y := 10; y <= 12; y++ {
					for x := 9; x <= 11; x++ {
						if x == 10 && (y == 10 || y == 11) {
							continue
						}
						if terrain {
							s.Field.Tiles[x*FieldSide+y] = 0x0B
						} else {
							s.Actors = append(s.Actors, Actor{Side: SideMonster, Member: -1,
							X: x, Y: y, Head: 10, Chest: 10, Threshold: 100})
						}
					}
				}
				ground := s.Ground
				weapon := s.Actors[1].Weapon
				random, expected := rng.NewState(2), rng.NewState(2)
				expected.Roll(100)
				expected.Roll(3)
				records, name := s.knockback(0, 1, random)
				if !reflect.DeepEqual(records, []int{StunsRecord, FallsRecord}) || name != "" {
					t.Fatalf("打暈訊息或未核H名稱改變：%v %q", records, name)
				}
				actor := s.Actors[1]
				if actor.X != 10 || actor.Y != 11 || actor.Command != 0 || actor.Weapon != weapon {
					t.Fatalf("受阻位置、指令或武器改變：%+v", actor)
				}
				if monster && (actor.HandItem != item || actor.Unarmed) {
					t.Fatal("受阻怪物的完整持物被移走")
				}
				if !monster && (s.Party[0].HandItem != item || !s.Party[0].HasWeapon || s.Party[0].WeaponName != item.Name) {
					t.Fatal("受阻隊員的完整持物被移走")
				}
				if !reflect.DeepEqual(s.Ground, ground) || s.ReturnScratch != 6 || random.State() != expected.State() {
					t.Fatal("受阻打暈改變地上槽、返回暫存或擊退骰序")
				}
			})
		}
	}
}
