package font

import (
	"errors"
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

type testPainter struct{ id int }

func (*testPainter) DrawFont(*ebiten.Image, *Font, int, int, string, color.RGBA, *color.RGBA) (int, bool) {
	return 0, false
}

func TestPainterScopeRestoresAfterNestedErrorAndPanic(t *testing.T) {
	before := painter
	defer func() { painter = before }()
	outer, inner := &testPainter{1}, &testPainter{2}
	want := errors.New("render failed")
	err := WithPainter(outer, func() error {
		if painter != outer {
			t.Fatal("outer painter missing")
		}
		if err := WithPainter(inner, func() error {
			if painter != inner {
				t.Fatal("inner painter missing")
			}
			return want
		}); err != want || painter != outer {
			t.Fatal("nested error leaked painter")
		}
		func() {
			defer func() {
				if recover() != "interrupted" || painter != outer {
					t.Fatal("panic leaked painter")
				}
			}()
			_ = WithPainter(inner, func() error { panic("interrupted") })
		}()
		return want
	})
	if err != want || painter != before {
		t.Fatal("render error leaked painter into EGA")
	}
}
