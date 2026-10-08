package ui

import (
	"encoding/csv"
	"image"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/kol-remake-cht/internal/font"
	"github.com/wicanr2/kol-remake-cht/internal/i18n"
	"github.com/wicanr2/kol-remake-cht/internal/world"
)

type promptSprites struct{ pic *ebiten.Image }

func (s promptSprites) Get(string, int) (*ebiten.Image, error) { return s.pic, nil }

// 走真正 WorldCJK 的地圖與城鎮外觀出口；拿掉任一提示繪製呼叫都會漏掉觀察字串。
func TestWorldCJKDrawsPromptOnBothMapExits(t *testing.T) {
	f, err := font.Load("../../assets/ui-font", "zh-hant-16")
	if err != nil {
		t.Fatal(err)
	}
	tab, err := i18n.Load("../../docs/i18n/newchar-zh-Hant.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := tab.LoadParagraphs("../../docs/i18n/world-prompts-zh-Hant.tsv"); err != nil {
		t.Fatal(err)
	}
	grid, err := world.LoadGrid(image.NewGray(image.Rect(0, 0, world.GridSide, world.GridSide)))
	if err != nil {
		t.Fatal(err)
	}
	canvas := ebiten.NewImage(960, 600)
	defer canvas.Deallocate()
	pic := ebiten.NewImage(512, 512)
	defer pic.Deallocate()
	old := font.DrawObserver
	defer func() { font.DrawObserver = old }()
	for _, mode := range []world.Mode{world.ModeMap, world.ModeTown} {
		for _, shown := range []bool{false, true} {
			s := &world.Screen{Travel: &world.Travel{Grid: grid}, MarkerOff: true}
			sess := &world.Session{Screen: s, Mode: mode, Town: &world.Town{Index: 0}}
			if shown {
				sess.Prompt = &world.Prompt{Kind: world.PromptNotice, Lines: []string{"You cannot ride away."}}
			}
			seen := false
			font.DrawObserver = func(_, text string) {
				if text == "你無法騎馬離開。" {
					seen = true
				}
			}
			if err := WorldCJK(canvas, promptSprites{pic}, f, tab, nil, s, sess); err != nil {
				t.Fatal(err)
			}
			if seen != shown {
				t.Fatalf("mode=%d prompt=%t observed=%t", mode, shown, seen)
			}
		}
	}
}

func TestWorldPromptCJKRealCatalogFitsAndPreservesOriginal(t *testing.T) {
	f, err := font.Load("../../assets/ui-font", "zh-hant-16")
	if err != nil {
		t.Fatal(err)
	}
	tab, err := i18n.Load("../../docs/i18n/newchar-zh-Hant.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := tab.LoadParagraphs("../../docs/i18n/world-prompts-zh-Hant.tsv"); err != nil {
		t.Fatal(err)
	}
	raw, err := os.Open("../../docs/i18n/world-prompts-zh-Hant.tsv")
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	r := csv.NewReader(raw)
	r.Comma = '\t'
	rows, err := r.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows)-1 != 67 {
		t.Fatalf("catalog 數量 %d", len(rows)-1)
	}
	for _, row := range rows[1:] {
		q := &world.Prompt{Lines: strings.Split(row[1], " "), Ask: "Will you partake?"}
		original := append([]string(nil), q.Lines...)
		lines, width, err := worldPromptCJKLines(f, tab, q)
		if err != nil {
			t.Fatalf("%s：%v", row[0], err)
		}
		body := lines[:len(lines)-2]
		if strings.Join(body, "") != row[2] {
			t.Fatalf("%s：內容丟失 %q", row[0], lines)
		}
		if width != (world.PromptBox.X+world.PromptBox.Width-8-world.PromptText.X)*world.CJK.Scale {
			t.Fatal("安全矩形沒有由原框推導")
		}
		if lines[len(lines)-1] != "要接下嗎？" || lines[len(lines)-2] != "" {
			t.Fatal("問句／間隔缺失")
		}
		if !reflect.DeepEqual(original, q.Lines) || q.Ask != "Will you partake?" {
			t.Fatal("原始提示被回寫")
		}
		for _, line := range lines {
			if w, _ := f.Measure(line); w > width {
				t.Fatal("超寬", row[0], line)
			}
		}
	}
	q := &world.Prompt{Lines: []string{strings.Repeat("長", 1000)}}
	if _, _, err := worldPromptCJKLines(f, tab, q); err == nil {
		t.Fatal("溢出敘述未拒絕")
	}
	q = &world.Prompt{Ask: strings.Repeat("長", 1000)}
	if _, _, err := worldPromptCJKLines(f, tab, q); err == nil {
		t.Fatal("溢出問句未拒絕")
	}
}
