package ui

import (
	"strings"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/font"
)

type plainT struct{}

func (plainT) T(en string) string { return en }

// 依原版 byte 數斷好的中文要接回整句並依像素寬度重排：行數變少、每行不超寬，
// 英文名字不被切開，內容一字不漏。
func TestRewrapCJKUsesPixelWidth(t *testing.T) {
	f, err := font.Load("../../assets/ui-font", "zh-hant-16")
	if err != nil {
		t.Fatal(err)
	}
	in := []string{"這道法術產生了明顯", "負面的效果。Warder", "Knight 撿起了長矛。"}
	out := rewrapCJK(f, in, plainT{}, 400)
	if len(out) >= len(in) {
		t.Fatalf("行數沒有變少：%q", out)
	}
	for _, line := range out {
		if w, _ := f.Measure(line); w > 400 {
			t.Fatalf("超寬 %d：%q", w, line)
		}
	}
	joined := strings.Join(out, "")
	if strings.ReplaceAll(joined, " ", "") != strings.ReplaceAll(strings.Join(in, ""), " ", "") {
		t.Fatalf("內容有遺漏：%q", out)
	}
	if !strings.Contains(strings.Join(out, "\n"), "Warder Knight") && !strings.Contains(strings.Join(out, "\n"), "Warder\nKnight") {
		t.Fatalf("英文名字之間的空白沒有補回：%q", out)
	}
}
