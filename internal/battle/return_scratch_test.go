package battle

import "testing"

// RE-192正常來源：Warder一般丟棄留816，Olear之後強制掉落留地面格1。
func TestReturnScratchDropKinds(t *testing.T) {
	s := State{Actors: []Actor{{Member: 0}, {Member: 1}},
		Party: []Member{{HasWeapon: true, WeaponName: "Olear"},
			{HasWeapon: true, WeaponName: "Warder"}}}
	if record, _ := s.DropItem(1); record != DropsRecord || s.ReturnScratch != 816 {
		t.Fatalf("一般丟棄：record=%d，暫存=%d", record, s.ReturnScratch)
	}
	s.forcedDrop(0)
	if s.ReturnScratch != 1 || s.Ground[1].Name != "Olear" {
		t.Fatalf("強制掉落：暫存=%d，地面=%+v", s.ReturnScratch, s.Ground[1])
	}
}

func TestReturnScratchFailedDropPreservesWord(t *testing.T) {
	s := State{Actors: []Actor{{Member: 0}},
		Party: []Member{{HasWeapon: true, WeaponName: "held"}}, ReturnScratch: 720}
	for i := range s.Ground {
		s.Ground[i].Name = "occupied"
	}
	s.DropItem(0)
	s.forcedDrop(0)
	if s.ReturnScratch != 720 || !s.Party[0].HasWeapon {
		t.Fatalf("地面全滿不應覆寫：%+v", s)
	}
	s.Ground[0] = GroundItem{}
	s.Party[0].HasWeapon = false
	s.DropItem(0)
	s.forcedDrop(0)
	if s.ReturnScratch != 720 {
		t.Fatalf("空手不應覆寫：%d", s.ReturnScratch)
	}
}

// RE-190／193原始Warder法術的成本，不以命中或Flinch改變暫存寫入。
func TestReturnScratchCastAndRest(t *testing.T) {
	s := State{Actors: []Actor{{Member: 0, Command: CmdCast}},
		Party: []Member{{Spells: []Spell{{Guards: [4]int{0x74}, Effect: 5, PowerCode: 0x8a}}}}}
	s.spendTime(0)
	if s.ReturnScratch != 720 {
		t.Fatalf("原始法術應留720，得到%d", s.ReturnScratch)
	}
	s.Actors[0].Command = CmdRest
	s.Actors[0].Stamina = 100
	s.EndTurnFatigue(0)
	if s.ReturnScratch != 720 {
		t.Fatalf("無傷休息不呼叫spendTime，得到%d", s.ReturnScratch)
	}
}

type scratchLootRoll struct{}

func (scratchLootRoll) Roll(int) int { return 1 }

func TestReturnScratchLootOverwritesAndDoesNotRepeat(t *testing.T) {
	s := State{Stage: StageDone, Verdict: VerdictLost,
		Actors: []Actor{{Side: SideParty, Member: 0}}, Party: []Member{{}},
		Monster: Monster{LootChance: 100}, ReturnScratch: 720}
	s.LootDownParty(scratchLootRoll{})
	if s.ReturnScratch != 16 {
		t.Fatalf("清槽最後留下16，得到%d", s.ReturnScratch)
	}
	s.ReturnScratch = 840
	s.LootDownParty(scratchLootRoll{})
	if s.ReturnScratch != 840 {
		t.Fatalf("共同出口重入不再搜刮，得到%d", s.ReturnScratch)
	}
}
