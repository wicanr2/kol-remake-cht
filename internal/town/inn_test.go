package town

import "testing"

// 白框在圖示列上兩端都繞回，與 town-icon-move-run1 的收據一致。
func TestMoveIconWraps(t *testing.T) {
	in := &Inn{Icons: InnIcons}
	in.MoveIcon(1)
	if in.Selected != 1 {
		t.Fatalf("往右一格應為 1，得到 %d", in.Selected)
	}
	in.Selected = len(InnIcons) - 1
	in.MoveIcon(1)
	if in.Selected != 0 {
		t.Fatalf("末格往右應繞回 0，得到 %d", in.Selected)
	}
	in.MoveIcon(-1)
	if in.Selected != len(InnIcons)-1 {
		t.Fatalf("首格往左應繞回末格，得到 %d", in.Selected)
	}
}

// 名牌只顯示名字的前五個字元。
func TestPlateName(t *testing.T) {
	if got := (Member{Name: "Theodoric"}).PlateName(); got != "Theod" {
		t.Fatalf("名牌應為 Theod，得到 %q", got)
	}
	if got := (Member{Name: "Jax"}).PlateName(); got != "Jax" {
		t.Fatalf("短名字不應被補齊，得到 %q", got)
	}
}
