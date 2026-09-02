package town

import (
	"image"
	"image/color"
	"testing"
)

// 造一個 16×160 的 atlas：只有第 unit 個單元的第 8 列第 3 欄有一個白點。
func testMarkers(t *testing.T, unit, px, py int) *Markers {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, MarkerSide, MarkerSide*MarkerUnits))
	img.Set(px, unit*MarkerSide+py, color.RGBA{255, 255, 255, 255})
	m, err := LoadMarkers(img)
	if err != nil {
		t.Fatalf("LoadMarkers：%v", err)
	}
	return m
}

// 單一亮點會膨脹成三個像素：自己是原色，左右兩側是黑色外框。
func TestOverlayDilatesOnePixelEachSide(t *testing.T) {
	m := testMarkers(t, 0, 3, 8)
	got := m.MarkerOverlay(0)
	for x := 0; x < MarkerSide; x++ {
		c := got[8*MarkerSide+x]
		switch x {
		case 3:
			if c != (color.RGBA{255, 255, 255, 255}) {
				t.Fatalf("第 %d 欄應為亮點本身，取到 %v", x, c)
			}
		case 2, 4:
			if c != black {
				t.Fatalf("第 %d 欄應為黑色外框，取到 %v", x, c)
			}
		default:
			if c.A != 0 {
				t.Fatalf("第 %d 欄應為透明，取到 %v", x, c)
			}
		}
	}
}

// 膨脹不跨越 8 像素的位元組邊界——少了這一條，原版畫面會多出一個像素對不上。
func TestOverlayDilationStopsAtByteBoundary(t *testing.T) {
	m := testMarkers(t, 0, 8, 4)
	got := m.MarkerOverlay(0)
	if got[4*MarkerSide+7].A != 0 {
		t.Fatal("第 7 欄不該被第 8 欄的亮點膨脹到")
	}
	if got[4*MarkerSide+9] != black {
		t.Fatal("第 9 欄應該是黑色外框")
	}

	m = testMarkers(t, 0, 7, 4)
	got = m.MarkerOverlay(0)
	if got[4*MarkerSide+8].A != 0 {
		t.Fatal("第 8 欄不該被第 7 欄的亮點膨脹到")
	}
	if got[4*MarkerSide+6] != black {
		t.Fatal("第 6 欄應該是黑色外框")
	}
}

func TestLoadMarkersRejectsWrongSize(t *testing.T) {
	if _, err := LoadMarkers(image.NewRGBA(image.Rect(0, 0, 16, 16))); err == nil {
		t.Fatal("尺寸不符時應該回錯")
	}
}
