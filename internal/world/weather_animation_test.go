package world

import (
	"image"
	"image/color"
	"reflect"
	"testing"
)

func TestWeatherAnimationGeometryAndGroups(t *testing.T) {
	decode := func(index int) (image.Image, error) {
		img := image.NewRGBA(image.Rect(0, 0, 64, 62))
		for y := 0; y < 62; y++ {
			for x := 0; x < 64; x++ {
				c := color.RGBA{uint8(index + 1), 0, 0, 255}
				if index%3 == 2 && index/3 != 3 && index/3 != 4 {
					c = color.RGBA{0, 0, 0, 255}
				}
				img.SetRGBA(x, y, c)
			}
		}
		if index == 1 {
			img.SetRGBA(63, 1, color.RGBA{200, 0, 0, 255})
		}
		return img, nil
	}
	w, err := LoadWeatherGroups(decode)
	if err != nil {
		t.Fatal(err)
	}
	got := w.CompositeGroup(0, 1)
	if got[0].R != 200 {
		t.Fatal("128 像素列尾沒有繞回")
	}
	if got[31*64].R != 1 {
		t.Fatal("下半底圖被雲層覆蓋")
	}
	if !reflect.DeepEqual(got, w.CompositeGroup(0, 129)) {
		t.Fatal("128 像素週期錯誤")
	}
	for _, group := range []int{3, 4} {
		for phase := 0; phase < 6; phase++ {
			pixels := w.CompositeGroup(group, phase)
			if pixels[0].R != uint8(group*3+phase%3+1) || pixels[len(pixels)-1] != pixels[0] {
				t.Fatal("雨雪沒有切換完整畫格")
			}
		}
	}
	if reflect.DeepEqual(w.CompositeGroup(0, 0), w.CompositeGroup(1, 0)) {
		t.Fatal("組別切換未換圖")
	}
}

func TestWeatherTickBoundary(t *testing.T) {
	s := &Screen{Group: 0}
	s.AdvanceWeather(1)
	if s.Phase != 0 {
		t.Fatal("動畫太早")
	}
	s.AdvanceWeather(1)
	if s.Phase != 1 {
		t.Fatal("兩個 tick 未更新")
	}
	s.Group = 3
	s.AdvanceWeather(1)
	if s.Phase != 0 {
		t.Fatal("換組未重置")
	}
	s.AdvanceWeather(5)
	if s.Phase != 0 {
		t.Fatal("雨雪三畫格未繞回")
	}
	if s.Clock != nil || s.TimeIndex != 0 {
		t.Fatal("動畫改動世界時間")
	}
}
