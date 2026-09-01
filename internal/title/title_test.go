package title

import "testing"

type fixedRand int

func (f fixedRand) Choose(n int) int { return int(f) % n }

func TestExpandAndPlay(t *testing.T) {
	seq, err := Load("../../docs/re/title-sequence.json")
	if err != nil {
		t.Fatalf("載入序列失敗：%v", err)
	}
	frames, err := seq.Expand(fixedRand(0))
	if err != nil {
		t.Fatalf("展開序列失敗：%v", err)
	}
	if len(frames) != 92 {
		t.Fatalf("展開後畫面數為 %d，預期 92", len(frames))
	}
	if frames[0].Back != 6 {
		t.Fatalf("第一格應該是閃電的 BACKPICS[6]，實際 %d", frames[0].Back)
	}

	// 播放器必須在停留 tick 用完之後才換格，且走完一輪回到第一格。
	p := NewPlayer(frames, ThunderSeconds*TickHz)
	if p.Current().Index != 0 {
		t.Fatalf("播放器起點不是第一格")
	}
	// 逐格推進正好用完停留時間，走完一輪應回到起點；中途不得跳格。
	for i := range frames {
		if got := p.Current().Index; got != i {
			t.Fatalf("第 %d 步停在第 %d 格", i, got)
		}
		if p.Pending() < MinDwellTicks {
			t.Fatalf("第 %d 格的停留時間小於最小值", i)
		}
		p.Advance(p.Pending())
	}
	if p.Current().Index != 0 {
		t.Fatalf("走完一輪之後沒有回到起點，停在第 %d 格", p.Current().Index)
	}
}

func TestRectsMatchArchives(t *testing.T) {
	seq, err := Load("../../docs/re/title-sequence.json")
	if err != nil {
		t.Fatalf("載入序列失敗：%v", err)
	}
	want := map[string][4]int{
		"SET0PICS": {112, 53, 88, 70},
		"SET1PICS": {96, 53, 104, 84},
		"SET2PICS": {48, 53, 160, 125},
		"FACEPICS": {104, 80, 112, 80},
	}
	for name, w := range want {
		r, ok := seq.Rect(name)
		if !ok {
			t.Fatalf("缺少 %s 的疊圖矩形", name)
		}
		if r.X != w[0] || r.Y != w[1] || r.Width != w[2] || r.Height != w[3] {
			t.Fatalf("%s 的矩形是 (%d,%d,%dx%d)，預期 (%d,%d,%dx%d)",
				name, r.X, r.Y, r.Width, r.Height, w[0], w[1], w[2], w[3])
		}
	}
}
