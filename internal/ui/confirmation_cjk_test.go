package ui

import (
	"image"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/book"
	"github.com/wicanr2/kol-remake-cht/internal/font"
	"github.com/wicanr2/kol-remake-cht/internal/i18n"
)

func TestChineseConfirmationTextStaysInsideDialog(t *testing.T) {
	f, err := font.Load("../../assets/ui-font", "zh-hant-16")
	if err != nil {
		t.Fatal(err)
	}
	tr, err := i18n.Load("../../docs/i18n/newchar-zh-Hant.json")
	if err != nil {
		t.Fatal(err)
	}
	l := book.CJKConfirm
	box := image.Rect(l.X, l.Y, l.X+l.Width, l.Y+l.Height)
	width, height := f.Measure(tr.T("YES/NO"))
	second := image.Rect(l.Line2X, l.Line2Y, l.Line2X+width, l.Line2Y+height)
	if !second.In(box) {
		t.Fatalf("是／否在確認框外：%v，框%v", second, box)
	}
	for _, key := range []string{"ARE YOU SURE?", "SAVE CHANGES"} {
		text := tr.T(key)
		if text == key {
			t.Fatalf("確認問句未翻譯：%s", key)
		}
		width, height := f.Measure(text)
		first := image.Rect(l.Line1X, l.Line1Y, l.Line1X+width, l.Line1Y+height)
		if !first.In(box) || first.Overlaps(second) {
			t.Fatalf("%s超界或重疊：%v／%v，框%v", text, first, second, box)
		}
	}
}
