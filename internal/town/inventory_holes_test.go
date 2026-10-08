package town

import (
	"strings"
	"testing"
)

func TestBackpackHolesKeepItemSlotsAndCarried(t *testing.T) {
	a := Good{Name: "A", Record: strings.Repeat("01", 16)}
	b := Good{Name: "B", Record: strings.Repeat("02", 16)}
	m := Member{Backpack: []Good{a, b, a, b, a, b}}
	got := m.removeItem(WornSlots + 1 + 2)
	if got == nil || got.Name != "A" || m.itemAt(WornSlots+1+2) != nil || m.Backpack[3].Name != "B" || len(m.Backpack) != 6 {
		t.Fatal("清除中間槽移動了後方物品")
	}
	if m.emptyPackSlot() != 2 {
		t.Fatal("六格slice中的空位被當成背包滿")
	}
	m.SyncCarried()
	if m.Carried[3][0] != 0 || m.Carried[4][0] != 2 {
		t.Fatal("原始Carried槽未保留")
	}
	m.putItem(WornSlots+1+2, &b)
	if m.emptyPackSlot() != -1 || m.Backpack[2].Name != "B" {
		t.Fatal("空格回填錯誤")
	}
	m.clearPack(5)
	if len(m.Backpack) != 5 || m.Backpack[4].Name != "A" {
		t.Fatal("尾部空格省略錯誤")
	}
}
