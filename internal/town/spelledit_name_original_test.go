package town

import "testing"

func TestSpellNameOriginalStartsEmptyAndWritesN(t *testing.T) {
	s := spellNameKeySession(t)
	if got := *s.Inside.SpellEdit.Typing; got != "" {
		t.Fatalf("原版起點應為空輸入，得到%q", got)
	}
	s.Key("BS")
	s.Key("N")
	s.Key("RET")
	if got := s.scroll.edit.draft.Name; got != "N" {
		t.Fatalf("原版正常RET／BS／N／RET應寫N，得到%q", got)
	}
}

func TestSpellNameOriginalPackedDisplayCase(t *testing.T) {
	s := spellNameKeySession(t)
	// 隔離寫回規則；正常來源為RE-212 native-name2，12字上限另驗。
	typed := "ABc EFGHIJM"
	s.scroll.edit.typing = &typed
	s.Key("RET")
	if got := s.scroll.edit.draft.Name; got != "Abc Efghijm" {
		t.Fatalf("原版5-bit名稱顯示應為Abc Efghijm，得到%q", got)
	}
}

func TestSpellNameEmptyConfirmPreservesDraft(t *testing.T) {
	s := spellNameKeySession(t)
	s.Key("RET")
	if s.scroll.edit.draft.Name != "Vartwelyrmi" || s.Party[0].Gold != 2000 {
		t.Fatal("空輸入確認改了原法術或資源")
	}
}

func TestSpellNameOriginalTwelfthCharacterCommits(t *testing.T) {
	s := spellNameKeySession(t)
	for _, key := range []string{"A", "B", "c", "SPACE", "D", "E", "F", "G", "H", "I", "J"} {
		s.Key(key)
	}
	if !s.SpellNameTyping() || *s.Inside.SpellEdit.Typing != "ABc DEFGHIJ" {
		t.Fatal("前11字未保持輸入狀態")
	}
	s.Key("K")
	if s.SpellNameTyping() || s.scroll.edit.draft.Name != "Abc Defghijk" {
		t.Fatal("第12字未依原版確認名稱")
	}
	if s.Party[0].Spells[0].Name != "Vartwelyrmi" || s.Party[0].Gold != 2000 {
		t.Fatal("自動確認提前寫角色或扣款")
	}
}
