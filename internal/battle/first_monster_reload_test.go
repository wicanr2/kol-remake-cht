package battle

import (
	"encoding/hex"
	"testing"
)

func TestRE267FirstSourceByte(t *testing.T) {
	for _, raw := range []byte{0, 6, 10, 11, 33, 255} {
		s := &State{Actors: make([]Actor, MaxActors), Monster: Monster{EquipmentVariants: make([]MonsterEquipment, 4)}}
		s.Actors[PartySlots] = Actor{Side: SideMonster, EquipmentVariant: 2, Kind: 2, Command: CmdMissile, Flag1D: 1}
		record := make([]byte, 16)
		record[3] = raw
		s.Monster.EquipmentVariants[2].Records = make([]string, 6)
		s.Monster.EquipmentVariants[2].Records[1] = hex.EncodeToString(record)
		// Nil typed item and zero first byte must not discard nonzero source padding.
		value, known := s.firstMonsterReloadByte(PartySlots)
		if !known || value != int(raw) {
			t.Fatalf("raw=%d value=%d known=%t", raw, value, known)
		}
		s.consumeShotAmmunition(PartySlots)
		wantCommand, wantFlag := CmdLoad, 0
		if raw == 6 || raw == 10 || raw == 11 {
			wantCommand, wantFlag = CmdMissile, 1
		}
		a := s.Actors[PartySlots]
		if a.Command != wantCommand || a.Flag1D != wantFlag || a.Kind != 1 {
			t.Fatalf("raw=%d actor=%+v", raw, a)
		}
		if _, known = s.firstMonsterReloadByte(PartySlots + 1); known {
			t.Fatal("unmodeled second alias guessed")
		}
		s.Actors[PartySlots].Kind = 0
		s.consumeShotAmmunition(PartySlots)
		if a = s.Actors[PartySlots]; a.Kind != 255 || a.Command != 0 {
			t.Fatalf("underflow did not override load: %+v", a)
		}
	}
}

func TestRE267PendingShotConsumesOnce(t *testing.T) {
	s := &State{Actors: make([]Actor, MaxActors), Monster: Monster{EquipmentVariants: make([]MonsterEquipment, 4)}}
	a := &s.Actors[PartySlots]
	*a = Actor{Side: SideMonster, Kind: 3, EquipmentVariant: 0, Flag1D: 1, Command: CmdMissile}
	records := make([]string, 6)
	records[1] = "00000021000000000000000000000000"
	s.Monster.EquipmentVariants[0].Records = records
	s.finishAttackAmmo(attackWait{actor: PartySlots, missile: true, reload: true})
	if a.Kind != 2 || a.Command != CmdLoad || a.Flag1D != 0 {
		t.Fatalf("pending-shot tail: %+v", a)
	}
}
