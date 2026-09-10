package world

import "testing"

func TestTownByCodeRequiresUniqueEntry(t *testing.T) {
	d := Data{towns: []Town{{Code: "MIN", Town: 7, EntryPoint: 3}, {Code: "BRT", Town: 2}}}
	if got := d.TownByCode("MIN"); got == nil || got.EntryPoint != 3 {
		t.Fatal("未保留城鎮入口")
	}
	if d.TownByCode("UNKNOWN") != nil {
		t.Fatal("未知城鎮不能猜入口")
	}
	d.towns = append(d.towns, Town{Code: "MIN", Town: 1})
	if d.TownByCode("MIN") != nil {
		t.Fatal("重複代碼不能猜入口")
	}
}
