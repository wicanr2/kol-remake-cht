package ui

import (
	"image"
	"testing"
)

func TestFitHDPortraitPreservesWideWeaponAndTallStaff(t *testing.T) {
	target := image.Rect(260, 80, 628, 448)
	cases := []struct{ source, want image.Rectangle }{
		{image.Rect(1, 1, 705, 547), image.Rect(260, 121, 628, 406)},
		{image.Rect(10, 10, 610, 810), image.Rect(306, 80, 582, 448)},
		{image.Rect(0, 0, 600, 600), target},
	}
	for _, tc := range cases {
		got := fitHDPortrait(tc.source, target)
		if got != tc.want || !got.In(target) {
			t.Errorf("source=%v got=%v want=%v", tc.source, got, tc.want)
		}
	}
}
