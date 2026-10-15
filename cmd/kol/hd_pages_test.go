package main

import (
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/portrait"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

func TestNativeInnPageLimits(t *testing.T) {
	long := make([]string, 10)
	for _, tc := range []struct {
		name string
		in   *town.Inn
		want bool
	}{
		{"missing", nil, false},
		{"approved standard inn", &town.Inn{DialogZh: []string{"歡迎"}}, false},
		{"unvisited long dialog", &town.Inn{DialogZh: long}, false},
		{"unvisited title", &town.Inn{Title: []string{"一", "二", "三"}}, false},
		{"input field", &town.Inn{DialogZh: []string{town.TextCaret}}, true},
		{"unvisited long input", &town.Inn{Dialog: []string{town.TextCaret}, DialogZh: long}, false},
		{"character sheet", &town.Inn{Sheet: &town.SheetPage{}}, true},
		{"service list", &town.Inn{List: &town.InnList{}}, true},
		{"spell editor", &town.Inn{SpellEdit: &town.SpellEditPage{}}, true},
		{"complete item page", &town.Inn{ItemPage: &town.ItemPage{}}, true},
		{"complete party page", &town.Inn{PartyPage: []portrait.Piece{{Archive: "PERPICS", Index: 0}}}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := nativeInnPage(tc.in); got != tc.want {
				t.Fatalf("nativeInnPage = %v, want %v", got, tc.want)
			}
		})
	}
}
