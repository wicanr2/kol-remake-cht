package town

import "testing"

func spellNameKeySession(t *testing.T) *Session {
	t.Helper()
	s := scrollSession(t, 2000, 109)
	s.Magic = spellDictionary(t)
	s.Party[0].Sheet.AdventurePoints = 3000
	s.Party[0].Order = hostelOrder()
	s.Party[0].Spells = []Spell{vartwelyrmi()}
	s.StartScroll()
	s.Key("D")
	s.Key("RET")
	s.Key("RET")
	s.Key("RET")
	if s.Inside.SpellEdit == nil || s.Inside.SpellEdit.Typing == nil {
		t.Fatal("正常按鍵未進名稱輸入")
	}
	return s
}

func TestSpellNameKeysWriteDraftAndChargeOnRecord(t *testing.T) {
	s := spellNameKeySession(t)
	for _, key := range []string{"BS", "BS", "BS", "BS", "BS", "BS", "BS", "BS", "BS", "BS", "BS", "U", "D", "L", "R", "A", "S", "Z", "X", "SPACE", "N", "O"} {
		if !s.Key(key) {
			t.Fatalf("名稱輸入未接收 %q", key)
		}
	}
	if got := *s.Inside.SpellEdit.Typing; got != "UDLRASZX NO" {
		t.Fatalf("typing=%q", got)
	}
	if s.scroll.edit.field != SpellFieldName || s.Party[0].Spells[0].Name != "Vartwelyrmi" {
		t.Fatal("文字輸入改了欄位或未確認的角色法術")
	}
	s.Key("W")
	if s.SpellNameTyping() {
		t.Fatal("第12字未自動確認")
	}
	if s.scroll.edit.draft.Name != "Udlraszx Now" || s.Party[0].Gold != 2000 {
		t.Fatal("確認名稱未寫副本或提前扣款")
	}
	s.Key(">")
	s.Key("RET")
	s.Key("RET")
	if p := s.Party[0]; p.Spells[0].Name != "Udlraszx Now" || p.Gold != 1170 || p.Sheet.AdventurePoints != 2370 {
		t.Fatalf("存錄後名稱／金冠／點數=%q/%d/%d", p.Spells[0].Name, p.Gold, p.Sheet.AdventurePoints)
	}
}

func TestSpellNameKeysCancelKeepsCharacterAndResources(t *testing.T) {
	s := spellNameKeySession(t)
	s.Key("BS")
	s.Key("N")
	if got := *s.Inside.SpellEdit.Typing; got != "N" {
		t.Fatalf("正常修改未生效=%q", got)
	}
	s.Key(">")
	s.Key("RET")
	if s.scroll.edit.draft.Name != "Vartwelyrmi" || s.Inside.SpellEdit.Typing != nil {
		t.Fatal("取消名稱未保留副本")
	}
	s.Key(">")
	s.Key("RET")
	s.Key(">")
	s.Key("RET")
	if p := s.Party[0]; p.Spells[0].Name != "Vartwelyrmi" || p.Gold != 2000 || p.Sheet.AdventurePoints != 3000 {
		t.Fatal("拒絕保存改了角色或資源")
	}
}
