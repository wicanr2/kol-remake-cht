package ui

import (
	"strings"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/font"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

// 中文字模格高：版面的每一列都要放得下它（SPEC-030）。
const cjkSheetGlyphHeight = 18

// 五個框要在畫布內、彼此不重疊。中文版面是重排出來的，這一支擋的是「算錯一個數字」。
func TestCJKSheetBoxesFitTheCanvas(t *testing.T) {
	const canvasW, canvasH = 640, 400
	l := town.CJKSheet
	boxes := []struct {
		name string
		r    town.Rect
	}{
		{"抬頭", l.Header}, {"屬性", l.Stats}, {"武器", l.Weapons},
		{"金幣", l.Gold}, {"長條", l.Bars},
	}
	for _, b := range boxes {
		if b.r.X0 < 0 || b.r.Y0 < 0 || b.r.X1 >= canvasW || b.r.Y1 >= canvasH {
			t.Errorf("%s框 %v 超出 %d×%d 畫布", b.name, b.r, canvasW, canvasH)
		}
	}
	for i := range boxes {
		for j := i + 1; j < len(boxes); j++ {
			a, b := boxes[i].r, boxes[j].r
			if a.X0 <= b.X1 && b.X0 <= a.X1 && a.Y0 <= b.Y1 && b.Y0 <= a.Y1 {
				t.Errorf("%s框與%s框重疊：%v／%v", boxes[i].name, boxes[j].name, a, b)
			}
		}
	}
}

// 每一個區塊的最後一列加上字高要留在框內——等比放大放不下正是這一頁重排的理由，
// 所以這一條是版面的核心判準。
func TestCJKSheetRowsStayInsideTheirBoxes(t *testing.T) {
	l := town.CJKSheet
	last := func(first, step, rows int) int { return first + (rows-1)*step + cjkSheetGlyphHeight }
	for _, c := range []struct {
		name   string
		bottom int
		box    town.Rect
	}{
		{"抬頭四列", l.BodyRow + cjkSheetGlyphHeight, l.Header},
		{"屬性八列", last(l.StatFirstRow, l.LineStep, 8), l.Stats},
		{"武器四列", last(l.WeaponFirstRow, l.LineStep, l.WeaponRows), l.Weapons},
		{"金幣四列", l.EquipRow + cjkSheetGlyphHeight, l.Gold},
	} {
		if c.bottom > c.box.Y1 {
			t.Errorf("%s排到 %d，框底只有 %d", c.name, c.bottom, c.box.Y1)
		}
	}
	// 反面對照：等比放大（英文座標 ×2）真的放不下，這個測試分得出來。
	eng := town.SheetLayout
	if doubled := (eng.Header.Y1 - eng.Header.Y0 + 1) * 2; doubled >= 6*l.LineStep {
		t.Fatalf("英文抬頭框 ×2 是 %d，六列要 %d——這個測試的前提不成立了",
			doubled, 6*l.LineStep)
	}
}

// clip 把超寬的字串截進安全寬度，而且不是無論如何都截。
func TestClipRespectsSafeWidth(t *testing.T) {
	f, err := font.Load("../../assets/ui-font", "zh-hant-16")
	if err != nil {
		t.Skipf("沒有中文字模表：%v", err)
	}
	safe := 8 * f.Advance()
	long := strings.Repeat("長", 20)
	got := clip(f, long, safe)
	if w, _ := f.Measure(got); w > safe {
		t.Fatalf("截過還是 %d px，安全寬度是 %d", w, safe)
	}
	if got == long {
		t.Fatal("超寬的字串沒有被截")
	}
	// 放得下的不動。
	short := "力量"
	if clip(f, short, safe) != short {
		t.Fatalf("放得下的字串被截了：%q", clip(f, short, safe))
	}
}

// 抬頭那三欄各佔三分之一，最長的種族／職業名要放得進去。
func TestCJKSheetHeaderColumnsHoldLongestNames(t *testing.T) {
	f, err := font.Load("../../assets/ui-font", "zh-hant-16")
	if err != nil {
		t.Skipf("沒有中文字模表：%v", err)
	}
	l := town.CJKSheet
	third := (l.Header.Width() - 2*f.Advance()) / 3
	for _, s := range []string{"半身人", "劍術大師", "巨魔剋星", "見習騎士"} {
		if w, _ := f.Measure(s); w > third {
			t.Errorf("%q 要 %d px，抬頭一欄只有 %d", s, w, third)
		}
	}
}
