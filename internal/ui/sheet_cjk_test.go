package ui

import (
	"strings"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/book"
	"github.com/wicanr2/kol-remake-cht/internal/font"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

// 中文字模格高：版面的每一列都要放得下它（SPEC-030）。
const cjkSheetGlyphHeight = 18

// 三倍畫布下原版版面照抄就放得下，這是 `book.CJKScale` 取 3 的理由。
// 這一支同時是那個理由的反面對照：兩倍放不下。
func TestCJKScaleFitsTheGlyphHeight(t *testing.T) {
	l := town.SheetLayout
	blocks := []struct {
		name              string
		box               town.Rect
		first, step, rows int
	}{
		{"抬頭", l.Header, l.NameRow, 6, 6},
		{"屬性", l.Stats, town.SheetStatFirstRow, town.LineStep, 8},
		{"武器", l.Weapons, l.WeaponFirstRow, l.WeaponRowStep, l.WeaponRows},
		{"金幣", l.Gold, l.GoldRow, 8, 4},
	}
	fits := func(scale int) []string {
		var bad []string
		for _, b := range blocks {
			h := (b.box.Y1 - b.box.Y0 + 1) * scale
			bottom := (b.first-b.box.Y0+(b.rows-1)*b.step)*scale + cjkSheetGlyphHeight
			if bottom > h {
				bad = append(bad, b.name)
			}
		}
		return bad
	}
	if bad := fits(book.CJKScale); len(bad) > 0 {
		t.Errorf("倍率 %d 放不下：%v", book.CJKScale, bad)
	}
	// 反面對照：倍率不夠時這個判準要能開口，否則它證明不了什麼。
	if bad := fits(2); len(bad) == 0 {
		t.Error("兩倍畫布也放得下的話，三倍就不是必要的——這個測試分不出差別")
	}
}

// 原版最小行距乘上倍率要不小於字高，這是「三倍」這個數字的來源。
func TestCJKScaleCoversTheSmallestLineStep(t *testing.T) {
	l := town.SheetLayout
	smallest := l.WeaponRowStep // 資料頁最小的行距
	if got := smallest * book.CJKScale; got < cjkSheetGlyphHeight {
		t.Fatalf("原版最小行距 %d × %d = %d，放不下 %d px 的字",
			smallest, book.CJKScale, got, cjkSheetGlyphHeight)
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
	if short := "力量"; clip(f, short, safe) != short {
		t.Fatalf("放得下的字串被截了：%q", clip(f, short, safe))
	}
}

// 屬性標籤那一欄放得下最長的中文標籤。
func TestCJKSheetStatLabelsFit(t *testing.T) {
	f, err := font.Load("../../assets/ui-font", "zh-hant-16")
	if err != nil {
		t.Skipf("沒有中文字模表：%v", err)
	}
	safe := (town.SheetLayout.StatValueCol - 2) * town.CellWidth * book.CJKScale
	for _, s := range []string{"力量：", "敏捷：", "健康：", "先見：",
		"魅力：", "智力：", "平衡：", "耐力："} {
		if w, _ := f.Measure(s); w > safe {
			t.Errorf("%q 要 %d px，屬性標籤欄只有 %d", s, w, safe)
		}
	}
}
