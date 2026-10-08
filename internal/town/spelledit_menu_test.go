package town

import (
	"reflect"
	"strconv"
	"testing"
)

// RE-209：sub_1AAE8 在 IDA linear 0x1AB6E 清零選列。
// 非零現值仍顯示在欄位；開啟、上移及取消不得寫回。
func TestSpellEditorMenusStartAtFirstRow(t *testing.T) {
	d := spellDictionary(t)
	for field := SpellFieldClassFirst; field < SpellFieldCount; field++ {
		t.Run(strconv.Itoa(field), func(t *testing.T) {
			s := scrollSession(t, 1000, 109)
			s.Magic = d
			s.Party[0].Order = hostelOrder()
			sp := vartwelyrmi()
			sp.Effect = 1
			SetSpellField(d, &sp, field, 1)
			if SpellFieldValue(d, sp, field) != 1 {
				t.Fatal("測試前提：欄位現值須為第 1 項")
			}
			s.Party[0].Spells = []Spell{sp}
			s.StartScroll()
			s.MoveScroll(1)
			yes(s)
			yes(s)
			s.MoveSpellField(field)
			before := s.scroll.edit.draft
			before.Guards = append([]int(nil), before.Guards...)
			yes(s)
			if s.Inside.SpellEdit.Menu == nil || s.scroll.edit.cursor != 0 || s.scroll.edit.top != 0 {
				t.Errorf("開啟欄位 %d：cursor=%d top=%d", field, s.scroll.edit.cursor, s.scroll.edit.top)
			}
			s.MoveScroll(-1)
			if s.scroll.edit.cursor != 0 {
				t.Errorf("向上須停在第 0 項，得到 %d", s.scroll.edit.cursor)
			}
			no(s)
			if !reflect.DeepEqual(s.scroll.edit.draft, before) || !reflect.DeepEqual(s.Party[0].Spells[0], sp) {
				t.Fatal("取消選單修改了工作副本或角色法術")
			}
		})
	}
}

func TestSpellEditorBodyDownOnceSelectsStrength(t *testing.T) {
	s := scrollSession(t, 1000, 109)
	s.Magic = spellDictionary(t)
	s.Party[0].Order = hostelOrder()
	sp := vartwelyrmi()
	sp.Effect = 6
	s.Party[0].Spells = []Spell{sp}
	s.StartScroll()
	s.MoveScroll(1)
	yes(s)
	yes(s)
	s.MoveSpellField(SpellFieldStat)
	yes(s)
	s.MoveScroll(1)
	yes(s)
	if s.scroll.edit.draft.Effect != 1 || s.Party[0].Spells[0].Effect != 6 {
		t.Fatalf("Down 一次須只把工作副本改為 Strength：draft=%d stored=%d", s.scroll.edit.draft.Effect, s.Party[0].Spells[0].Effect)
	}
}
