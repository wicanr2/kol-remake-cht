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

// 節拍照 `sub_11390`（RE-164）：進野外先畫一格（雲層先轉一格），之後每兩個 tick 一格；
// 雨雪畫格索引換組不歸零，雲層換組從 0 起算；動畫不碰世界時間。
func TestWeatherTickBoundary(t *testing.T) {
	s := &Screen{Group: 0}
	step := func(ticks float64, want int, why string) {
		t.Helper()
		s.AdvanceWeather(ticks)
		if s.Phase != want {
			t.Fatalf("%s：相位 %d，期望 %d", why, s.Phase, want)
		}
	}
	step(0.5, 1, "進野外那一次先轉一格就畫")
	step(1, 1, "一個 tick 還在等")
	step(1, 2, "第二個 tick 等完立刻畫下一格")
	s.Group = 3
	step(0.5, 0, "換成雨：顯示下一張（索引從 0）")
	step(1, 0, "等完畫第 0 張")
	step(2, 1, "兩個 tick 後第 1 張")
	s.Group = 4
	step(2, 2, "換成雪：索引不歸零，接著畫第 2 張")
	step(2, 0, "三張繞回")
	s.Group = 0
	step(2, 1, "換回雲層：重新載入，第一次畫轉一格")
	if s.Clock != nil || s.TimeIndex != 0 {
		t.Fatal("動畫改動世界時間")
	}
}

// WeatherRedraw 是地圖重畫：計數清零、立刻畫一格。
func TestWeatherRedrawDrawsAtOnce(t *testing.T) {
	s := &Screen{Group: 3}
	s.AdvanceWeather(0.5)
	if s.Phase != 0 {
		t.Fatalf("第一格 %d", s.Phase)
	}
	s.WeatherRedraw()
	if s.Phase != 1 {
		t.Fatalf("重畫之後應該立刻換到第 1 張，得到 %d", s.Phase)
	}
}
