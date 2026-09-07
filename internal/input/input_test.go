package input

import (
	"reflect"
	"testing"
)

// 點選單的某一列要送出「把游標移過去再 Return」。游標往哪個方向移、移幾次，
// 是滑鼠能不能對上鍵盤狀態機的關鍵。
func TestRowTokens(t *testing.T) {
	for _, tc := range []struct {
		name          string
		index, cursor int
		want          []string
	}{
		{"往下三列", 3, 0, []string{"D", "D", "D", "RET"}},
		{"往上兩列", 1, 3, []string{"U", "U", "RET"}},
		{"點在游標上", 2, 2, []string{"RET"}},
	} {
		got := Row(24, 52, 100, 6, tc.index, tc.cursor, "U", "D").Tokens
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s：得到 %v，期望 %v", tc.name, got, tc.want)
		}
	}
}

// 第 i 列的矩形要落在 y+i*step 起算的那一格，否則點下去會選到隔壁。
func TestRowRect(t *testing.T) {
	r := Row(24, 52, 100, 6, 3, 0, "U", "D").Rect
	if r.Min.X != 24 || r.Min.Y != 52+3*6 || r.Max.X != 124 || r.Max.Y != 52+4*6 {
		t.Fatalf("第 3 列的矩形是 %v", r)
	}
}

// 圖示列是橫的，白框走 `<` `>`。
func TestCellTokens(t *testing.T) {
	got := Cell(8, 176, 24, 20, 26, 4, 1).Tokens
	want := []string{">", ">", ">", "RET"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("得到 %v，期望 %v", got, want)
	}
}

// 命中判斷用的是畫布座標——Ebitengine 的 CursorPosition 已經換算過縮放。
func TestHit(t *testing.T) {
	targets := []Target{Row(24, 52, 100, 6, 2, 0, "U", "D")}
	// 52+2*6=64 起的那一列，(30, 65) 落在裡面。
	got := Hit(targets, 30, 65, "")
	want := []string{"D", "D", "RET"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("得到 %v，期望 %v", got, want)
	}
}

// 點在可點區域以外走 fallback——「按任意鍵繼續」那些畫面靠它前進。
func TestHitFallback(t *testing.T) {
	targets := []Target{Row(24, 52, 100, 6, 0, 0, "U", "D")}
	if got := Hit(targets, 300, 190, "RET"); !reflect.DeepEqual(got, []string{"RET"}) {
		t.Fatalf("落空時得到 %v", got)
	}
	if got := Hit(targets, 300, 190, ""); got != nil {
		t.Fatalf("沒有 fallback 時應該不送 token，卻得到 %v", got)
	}
}
