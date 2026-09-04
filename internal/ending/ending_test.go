package ending

import "testing"

// 七段的矩形換算出來要與 RE-017 量到的七組尺寸一模一樣。那七組是從
// LZW 解碼的像素數反推的，與這裡的四個全域是兩條獨立的證據。
func TestPanelBoxesMatchPictureSizes(t *testing.T) {
	want := [][2]int{
		{168, 122}, {192, 111}, {296, 93}, {176, 97},
		{168, 123}, {192, 102}, {184, 98},
	}
	if len(Panels) != len(want) {
		t.Fatalf("應該有 %d 段，得到 %d", len(want), len(Panels))
	}
	for i, p := range Panels {
		if p.Picture != i {
			t.Errorf("第 %d 段的圖號是 %d", i, p.Picture)
		}
		if got := [2]int{p.Box.Width(), p.Box.Height()}; got != want[i] {
			t.Errorf("第 %d 段的尺寸是 %v，期望 %v", i, got, want[i])
		}
	}
}

// 十七筆記錄依序用完，不重複也不跳號（ENDTXT 有十八筆，最後一筆沒有呼叫點）。
func TestPanelRecordsAreConsecutive(t *testing.T) {
	next := 0
	for i, p := range Panels {
		for _, r := range p.Records {
			if r != next {
				t.Fatalf("第 %d 段的記錄是 %d，期望 %d", i, r, next)
			}
			next++
		}
	}
	if next != 17 {
		t.Fatalf("用掉 %d 筆記錄，期望 17", next)
	}
}

// 觸發是「那一格等於 0x17」，別的值都不算。
func TestTriggered(t *testing.T) {
	if !Triggered([]int{TriggerValue, 0, 0, 0, 0}, 0) {
		t.Fatal("等於 0x17 應該觸發")
	}
	for _, v := range []int{0, 1, 0x16, 0x18, 0xFF} {
		if Triggered([]int{v, 0, 0, 0, 0}, 0) {
			t.Fatalf("值 %#x 不該觸發", v)
		}
	}
	if Triggered([]int{TriggerValue}, 5) {
		t.Fatal("群組超出範圍不該觸發")
	}
	if !AnyTriggered([][]int{{0}, {TriggerValue}}, 0) {
		t.Fatal("隊上有人達成就算")
	}
	if AnyTriggered(nil, 0) {
		t.Fatal("空隊伍不該觸發")
	}
}

// 七段依序播完。
func TestStateAdvance(t *testing.T) {
	var s State
	for i := 0; i < len(Panels); i++ {
		p := s.Current()
		if p == nil || p.Picture != i {
			t.Fatalf("第 %d 段拿到 %v", i, p)
		}
		s.Advance()
	}
	if !s.Done() || s.Current() != nil {
		t.Fatal("七段之後應該結束")
	}
}
