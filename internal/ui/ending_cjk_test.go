package ui

import (
	"image"
	"strings"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/book"
	"github.com/wicanr2/kol-remake-cht/internal/ending"
	"github.com/wicanr2/kol-remake-cht/internal/font"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

// 檢查真實七段譯文與圖框，防止縮放修復後仍因段落變短把文字寫在圖片上。
func TestCJKEndingTextOutsidePictures(t *testing.T) {
	f, err := font.Load("../../assets/ui-font", "zh-hant-16")
	if err != nil {
		t.Fatal(err)
	}
	bank, err := town.LoadUIText("../../assets/original-derived-data/ui-text.json")
	if err != nil {
		t.Fatal(err)
	}
	tr, err := town.LoadTranslations("../../docs/i18n/town-zh-Hant.json")
	if err != nil {
		t.Fatal(err)
	}
	tr.UI, err = town.LoadUITranslations("../../docs/i18n/ui-zh-Hant.json")
	if err != nil {
		t.Fatal(err)
	}
	scale := book.CJKScale
	canvas := image.Rect(0, 0, 320*scale, 200*scale)
	for i, panel := range ending.Panels {
		pic := image.Rect(panel.Box.X()*scale, panel.Box.Y()*scale,
			(panel.Box.X()+panel.Box.Width())*scale, (panel.Box.Y()+panel.Box.Height())*scale)
		var previous []image.Rectangle
		for _, seg := range town.JoinedSegments(bank, tr, ending.Bank, panel.Records) {
			if strings.TrimSpace(seg.Text) == "" {
				continue
			}
			w, h := f.Measure(seg.Text)
			x, y := seg.Col*8*scale, seg.Row*scale
			r := image.Rect(x, y, x+w, y+h)
			if !r.In(canvas) || r.Overlaps(pic) {
				t.Errorf("段 %d 文字 %q，文字框 %v，圖框 %v", i, seg.Text, r, pic)
			}
			for _, p := range previous {
				if r.Overlaps(p) {
					t.Errorf("段 %d 文字互疊：%v / %v", i, r, p)
				}
			}
			previous = append(previous, r)
		}
	}
}
