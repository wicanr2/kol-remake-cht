package battle

import (
	"encoding/json"
	"github.com/wicanr2/kol-remake-cht/internal/rng"
	"os"
	"testing"
)

func TestMonsterRangedWeaponSource(t *testing.T) {
	hand := Weapon{Kind: 87, Base: 2, Code: 83, Count: 3, Sides: 9, Weight: 120, TwoHand: true}
	rock := &Item{Class: 2, Encumbrance: 90, Hands: 1, Record: "713eac216400000000005aff5102e000"}
	a := Actor{Side: SideMonster, Kind: 14, LoadLevel: 1, Weapon: hand, MonsterRanged: rock, ArmourWeight: 201, ShieldGrade: NoShield}
	a.ReloadGear()
	if a.Encumbrance != 172 || a.WeaponSpeed != 63 {
		t.Fatalf("遠程資格負重／速度=%d/%d，原版172/63", a.Encumbrance, a.WeaponSpeed)
	}
	for _, command := range []int{CmdWeapon, CmdUnarmed, CmdMissile} {
		a.Command = command
		w := a.AttackWeapon(1)
		if command == CmdMissile {
			if w.Code != 81 || w.Base != 0 || w.Count != 3 || w.Sides != 9 || w.Weight != 90 {
				t.Fatalf("射擊未取Boulder：%+v", w)
			}
		} else if w != hand {
			t.Fatalf("近戰／徒手改取遠程：%+v", w)
		}
	}
	a.Kind = 0
	a.ReloadGear()
	if a.Encumbrance != 182 || a.WeaponSpeed != 84 {
		t.Fatal("零資格沒有取手持負重")
	}
	// 指令與資格是不同的取槽條件。
	if a.AttackWeapon(0).Code != 81 {
		t.Fatal("射擊傷害錯依資格選槽")
	}
	a.MonsterRanged = nil
	a.Kind = 1
	a.ReloadGear()
	if a.Encumbrance != 140 || a.WeaponSpeed != 0 {
		t.Fatal("空遠程槽回退手持負重")
	}
	if w := a.AttackWeapon(0); w.Count != 1 || w.Sides != 2 || w.Base != 0 || w.Weight != 0 {
		t.Fatalf("空遠程碼0不是1d2：%+v", w)
	}
	a.Side = SideParty
	a.ReloadGear()
	if a.AttackWeapon(0) != hand || a.WeaponSpeed != 84 {
		t.Fatal("隊員射擊／負重改變")
	}
	// 雙手條件依負重所選武器，而非近戰那件。
	a.Side = SideMonster
	a.MonsterRanged = &Item{Class: 2, Hands: 0}
	a.ShieldGrade = 3
	a.RightArm = 10
	a.LeftArm = 10
	if a.EffectiveShieldGrade() != 3 {
		t.Fatal("盾牌錯用雙手近戰武器")
	}
	a.MonsterRanged.Hands = 1
	if a.EffectiveShieldGrade() != NoShield {
		t.Fatal("雙手遠程武器仍使用盾牌")
	}
}

func TestMonsterRangedInitializedLoadNormalOriginal(t *testing.T) {
	path := os.Getenv("KOL_RANGED_TEST_CASE")
	if path == "" {
		t.Skip("正常原版收據由隔離驗收命令提供")
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var native struct {
		Placement []struct{ Seed uint32 }
		Reads     []struct {
			Phase              string
			Actor, Encumbrance int
			Speed              int `json:"weapon_speed"`
		} `json:"ranged_reads"`
	}
	if err = json.Unmarshal(body, &native); err != nil {
		t.Fatal(err)
	}
	mons, err := LoadMonsters("../../assets/original-derived-data/monsters.json")
	if err != nil {
		t.Fatal(err)
	}
	mon := mons.At(9)
	s := NewQuest(&Site{Index: -1, Level: 1, Count: 3, MonsterName: mon.Name, MonsterX: []int{35, 37, 36}, MonsterY: []int{32, 29, 34}}, nil, mon, nil)
	random := rng.NewState(native.Placement[1].Seed)
	s.initializeMonsterEquipment(random)
	s.InitializeMonsterRanged(random)
	seen := map[int]bool{}
	for _, w := range native.Reads {
		if w.Phase != "weight-read" || w.Actor < 6 || w.Actor > 8 || seen[w.Actor] {
			continue
		}
		seen[w.Actor] = true
		a := s.Actors[w.Actor]
		if a.Encumbrance != w.Encumbrance || a.WeaponSpeed != w.Speed {
			t.Fatalf("槽%d=%d/%d 原版%d/%d", w.Actor, a.Encumbrance, a.WeaponSpeed, w.Encumbrance, w.Speed)
		}
	}
	if len(seen) != 3 {
		t.Fatal("缺少正常三名怪物負重觀察")
	}
	before := random.State()
	s.InitializeMonsterRanged(random)
	if random.State() != before {
		t.Fatal("重擲資格")
	}
}
