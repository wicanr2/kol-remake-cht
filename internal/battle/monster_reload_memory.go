package battle

import "encoding/hex"

// SPEC-022 / RE-269: DS:D0C5 is pool row 6 head byte 3. Startup clears it.
// The last seventh monster's live head record remains across battles in this process.
// This cache does not enter character saves and does not model later aliases.
type MonsterReloadMemory struct {
	actor   *Actor
	source  byte
	hadHead bool
	unknown bool
}

func (m *MonsterReloadMemory) capture(a *Actor, gear MonsterEquipment) {
	if m == nil {
		return
	}
	m.actor, m.unknown = a, true
	if len(gear.Records) != 6 {
		return
	}
	b, err := hex.DecodeString(gear.Records[1])
	if err != nil || len(b) != 16 {
		return
	}
	m.source, m.hadHead, m.unknown = b[3], b[0] != 0, false
}

func (m *MonsterReloadMemory) value() (int, bool) {
	if m == nil || m.unknown {
		return 0, false
	}
	if m.actor != nil && m.hadHead && m.actor.Armour[0] == nil {
		return 0, true
	}
	return int(m.source), true
}

func (s *State) monsterReloadByte(i int) (int, bool) {
	if i == PartySlots {
		return s.firstMonsterReloadByte(i)
	}
	if i != PartySlots+1 || i >= len(s.Actors) || s.Actors[i].Side != SideMonster {
		return 0, false
	}
	return s.MonsterReloadMemory.value()
}
