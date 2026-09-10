package town

import (
	"reflect"
	"testing"
)

func TestGoodbyeOnlyForMultipleMembers(t *testing.T) {
	for _, n := range []int{1, 2, 6} {
		for _, confirm := range []string{"SPACE", "RET"} {
			s := recruitSession(t, n)
			s.Trans = nil
			s.Text = &Text{Towns: []townText{{Town: s.Outdoor.Town, Buildings: []BuildingText{{Door: s.door.Index, Messages: []Message{{Index: 17, Display: []string{"Farewell, travellers!"}}}}}}}}
			x, y := s.Outdoor.PartyPosition()
			before := append([]Member(nil), s.Party...)
			s.setInsideIcons([]int{IconDoor})
			s.Key("RET")
			if n == 1 {
				if s.Mode != ModeMap {
					t.Fatal("單人離店不應等待告別")
				}
				continue
			}
			if !s.leaving || s.Mode != ModeInside || !reflect.DeepEqual(s.Inside.Dialog, []string{"Farewell, travellers!"}) {
				t.Fatal("多人沒有顯示該建築的告別")
			}
			for _, key := range []string{"R", "LISTEN", "ASK:test", "ESC"} {
				s.Key(key)
			}
			if !s.leaving || !reflect.DeepEqual(s.Party, before) {
				t.Fatal("等待期間改動角色或提前離店")
			}
			s.Key(confirm)
			xx, yy := s.Outdoor.PartyPosition()
			if s.Mode != ModeMap || s.Inside != nil || s.leaving || xx != x || yy != y {
				t.Fatal("確認應只返回原門格")
			}
		}
	}
}

func TestDoorwayCancellationHasNoGoodbye(t *testing.T) {
	s := recruitSession(t, 2)
	s.Mode = ModeDoor
	s.leaveDoor()
	if s.Mode != ModeMap || s.leaving {
		t.Fatal("門口返回不應出現建築告別")
	}
}
