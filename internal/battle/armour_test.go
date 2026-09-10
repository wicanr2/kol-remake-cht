package battle

import (
	"encoding/hex"
	"testing"
)

func testArmour(bonus, condition, flags byte) *Item {
	raw := make([]byte, 16)
	raw[0], raw[8], raw[10], raw[11], raw[15] = 1, bonus, 4, condition, flags
	return &Item{Name: "TEST ARMOUR", Record: hex.EncodeToString(raw), Encumbrance: 4, Condition: int(condition), Percent: 40}
}

func TestArmourBothAttackPaths(t *testing.T) {
	for _, missile := range []bool{false, true} {
		a := Actor{Attack: 10000, Strength: 70, Stamina: 100, Weapon: Weapon{Base: 5, Count: 1, Sides: 2}}
		d := Actor{Side: SideParty, Threshold: 100, Head: 100, Chest: 100, RightArm: 100, LeftArm: 100, RightLeg: 100, LeftLeg: 100}
		for i := range d.Armour {
			d.Armour[i] = testArmour(0x70, 255, 0x20)
		}
		rng := &seqRand{values: []int{1}}
		var out Outcome
		if missile {
			out = ResolveMissile(&a, &d, 0, 10000, false, rng)
		} else {
			out = Resolve(&a, &d, 0, rng)
		}
		if out.Miss || out.Parried || out.Damage != 0 || d.Damage != 0 || d.Hurt == 0 {
			t.Fatalf("射擊=%v：護甲應擋完傷勢並保留護甲前受創值，out=%+v damage=%d hurt=%d", missile, out, d.Damage, d.Hurt)
		}
		for _, item := range d.Armour {
			if item.Condition != 255 || item.Percent != 40 {
				t.Fatal("完全阻擋不應扣耐久或修改盾牌百分比")
			}
		}
	}
}

func TestBattleArmourOwnershipAndBreakage(t *testing.T) {
	item := testArmour(0, 32, 0x20)
	party := []Member{{Armour: [3]*Item{item}, ArmourWeight: 4}}
	monster := Monster{Worn: []*Item{item}, ArmourWeight: 4}
	site := &Site{Index: -1, Count: 2, PartyX: []int{0}, PartyY: []int{0}, MonsterX: []int{1, 2}, MonsterY: []int{1, 2}}
	s := NewQuest(site, party, monster, nil)
	for _, index := range []int{0, PartySlots, PartySlots + 1} {
		if s.Actors[index].Armour[0] == item {
			t.Fatal("參戰者與來源共用可變護甲")
		}
	}
	a := &s.Actors[PartySlots]
	if got := a.AbsorbArmour(5, 0, &seqRand{values: []int{1}}); got != 4 || a.Armour[0] != nil || a.ArmourWeight != 0 {
		t.Fatalf("護甲破損未清除或負重未更新：傷害=%d actor=%+v", got, a)
	}
	if item.Condition != 32 || s.Actors[0].Armour[0].Condition != 32 || s.Actors[PartySlots+1].Armour[0].Condition != 32 {
		t.Fatal("破損污染來源、隊員或另一名敵人")
	}
}

func TestArmourRejectsMissingRecord(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("缺少原始記錄時不應猜補耐久旗標")
		}
	}()
	a := Actor{Armour: [3]*Item{{Name: "INCOMPLETE"}}}
	a.AbsorbArmour(5, 0, &seqRand{values: []int{1}})
}
