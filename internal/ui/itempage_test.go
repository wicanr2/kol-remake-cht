package ui

import (
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/town"
)

func TestItemPageInteractionBaseFollowsOriginalBranches(t *testing.T) {
	cases := []struct {
		name string
		page *town.ItemPage
		want bool
	}{
		{name: "complete list", page: &town.ItemPage{}, want: false},
		{name: "selected detail", page: &town.ItemPage{Detail: []town.Segment{}}, want: true},
		{name: "selected action", page: &town.ItemPage{Acting: true}, want: true},
		{name: "empty slot message", page: &town.ItemPage{Message: []string{"YOU HAVE NOTHING"}}, want: true},
		{name: "post-action note redraw", page: &town.ItemPage{Note: []string{"IT IS DROPPED!"}}, want: false},
		{name: "giving note remains interaction", page: &town.ItemPage{Acting: true, Note: []string{"GIVE TO WHOM?"}, Giving: true}, want: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := itemPageUsesInteractionBase(tc.page); got != tc.want {
				t.Fatalf("interaction base = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestItemPageDetailShadowGeometryHasFourPixelBands(t *testing.T) {
	l := town.ItemPageLayout.Detail
	rightX := l.X0 + l.Width()
	rightY := l.Y0 + town.Shadow.Offset
	rightW := town.Shadow.Thickness
	rightH := l.Height()
	if rightX != l.X1+1 || rightY != 62 || rightW != 4 || rightH != 68 {
		t.Fatalf("右側投影幾何 = (%d,%d) %dx%d，want (258,62) 4x68",
			rightX, rightY, rightW, rightH)
	}

	bottomX := l.X0 + town.Shadow.Offset
	bottomY := l.Y0 + l.Height()
	bottomW := l.Width()
	bottomH := town.Shadow.Thickness
	if bottomX != 66 || bottomY != l.Y1+1 || bottomW != 196 || bottomH != 4 {
		t.Fatalf("下側投影幾何 = (%d,%d) %dx%d，want (66,126) 196x4",
			bottomX, bottomY, bottomW, bottomH)
	}
}
