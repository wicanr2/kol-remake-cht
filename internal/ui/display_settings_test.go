package ui

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/kol-remake-cht/internal/font"
	"github.com/wicanr2/kol-remake-cht/internal/i18n"
)

func TestDisplaySettingsChineseFitsIncludingFailure(t *testing.T) {
	f, err := font.Load("../../assets/ui-font", "zh-hant-16")
	if err != nil {
		t.Fatal(err)
	}
	tr, err := i18n.Load("../../docs/i18n/newchar-zh-Hant.json")
	if err != nil {
		t.Fatal(err)
	}
	keys := []string{"Display F10", "Display settings", "Changes preview immediately", "EGA Classic", "Original pixel artwork", "HD Painted", "Inn and six characters", "In use", "Other artwork uses the classic style", "Back to game  Esc", "HD images are unavailable.", "Current display retained. Try again later."}
	for _, key := range keys {
		if tr.T(key) == key {
			t.Errorf("選單缺翻譯：%s", key)
		}
	}
	dst := ebiten.NewImage(960, 600)
	for _, open := range []bool{false, true} {
		for _, hd := range []bool{false, true} {
			for _, problem := range []bool{false, true} {
				for cursor := 0; cursor < 3; cursor++ {
					dst.Clear()
					if err := DrawDisplaySettings(dst, f, tr, DisplaySettings{Open: open, HD: hd, Problem: problem, Cursor: cursor}); err != nil {
						t.Fatal(err)
					}
				}
			}
		}
	}
}
