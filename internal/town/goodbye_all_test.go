package town

import (
	"reflect"
	"testing"
)

// 全城逐店的告別（SPEC-024，#10）：用正式資料把每一棟建築走一次「兩人隊伍按 DOOR」，
// 畫面要是那一棟的訊息 17、圖示列只剩 SPACEBAR；按一次回到地圖。英文讀原文，
// 中文模式讀譯文，兩種都要每一棟有字。
//
// 缺文字的只有 TOWN7 的門 21：原版資料本身沒有那一棟的文字 payload（RE-061），
// remake 直接離開，另列在差異審查裡。
func TestGoodbyeEveryBuilding(t *testing.T) {
	buildings, err := LoadBuildings("../../assets/original-derived-data/town-buildings.json")
	if err != nil {
		t.Skipf("需要正式建築資料：%v", err)
	}
	text, err := LoadText("../../assets/original-derived-data/town-text.json")
	if err != nil {
		t.Skipf("需要正式城鎮文字：%v", err)
	}
	trans, err := LoadTranslations("../../docs/i18n/town-zh-Hant.json")
	if err != nil {
		t.Skipf("需要城鎮譯文：%v", err)
	}
	party := []Member{{Name: "Theodoric", Speaker: true}, {Name: "Olear"}}
	checked, missing := 0, 0
	for _, tb := range buildings.Towns {
		for di := range tb.Doors {
			door := &tb.Doors[di]
			for _, zh := range []bool{false, true} {
				s := &Session{Outdoor: &Outdoor{Town: tb.Town}, Text: text,
					Party: append([]Member(nil), party...), Trans: &Translations{}}
				if zh {
					s.Trans = trans
				}
				bt := text.Building(tb.Town, door.Index)
				s.Inside = EnterBuilding("TWN0PICS", door, bt, s.Party)
				s.Mode, s.door = ModeInside, door
				s.leaveDoor()
				if door.TextMissing {
					if s.Mode != ModeMap || s.leaving {
						t.Errorf("城 %d 門 %d 缺文字，應該直接離開", tb.Town, door.Index)
					}
					missing++
					continue
				}
				want := bt.Message(17).Display
				if zh {
					want = trans.Message(tb.Town, door.Index, 17)
					if len(want) == 0 {
						t.Errorf("城 %d 門 %d 的告別沒有譯文", tb.Town, door.Index)
						continue
					}
				}
				if !s.leaving || !reflect.DeepEqual(s.Inside.Dialog, want) {
					t.Errorf("城 %d 門 %d（中文=%v）：告別是 %q，應為 %q", tb.Town, door.Index, zh, s.Inside.Dialog, want)
					continue
				}
				if len(s.Inside.Icons) != 1 || s.Inside.Icons[0] != IconSprite(IconSpacebar) {
					t.Errorf("城 %d 門 %d：圖示列應只剩 SPACEBAR，得到 %v", tb.Town, door.Index, s.Inside.Icons)
				}
				s.Continue()
				if s.Mode != ModeMap || s.Inside != nil || s.leaving {
					t.Errorf("城 %d 門 %d：按一次應回到地圖", tb.Town, door.Index)
				}
				checked++
			}
		}
	}
	if checked != 2*101 || missing != 2 {
		t.Fatalf("檢查了 %d 次告別、%d 次缺文字，應為 202 與 2", checked, missing)
	}
}
