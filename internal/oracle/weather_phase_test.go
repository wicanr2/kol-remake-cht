//go:build oracle

package oracle

import (
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/kol-remake-cht/internal/world"
)

// 天候節拍的逐次比較（#20，RE-164）：同一串「呼叫、換組、地圖重畫」送進原版 `sub_11390`
// 野外那一半與 remake 的 `Screen.WeatherCall`／`WeatherRedraw`，每一次比「等了一個 tick
// 還是畫了一格」、畫的是雨雪第幾張或雲層轉到第幾格。換組時原版 `sub_2A0C6` 會重新載入
// 雲層（旋轉歸零），這裡只寫分派組 `word_2DF5A`，雲層旋轉次數由探針自己從換組那一刻起算。
func TestWeatherPhaseOriginal(t *testing.T) {
	o := prepare(t)
	for _, ea := range []uint32{0x2DF60, 0x2DF2C, 0x2DF48, 0x2DF28, 0x2DF5C} {
		o.WriteU16(o.IDA(ea), 0)
	}
	o.WriteU16(o.IDA(0x2DF44), 1)
	o.WriteU16(o.IDA(0x34A1A), 0x9000)
	var draws []uint16
	turns, waits := 0, 0
	o.Stub(o.IDA(0x19E9D), func(p *oracle.Oracle) uint32 { draws = append(draws, p.Arg(0)); return 0 })
	o.Stub(o.IDA(0x2A18E), func(*oracle.Oracle) uint32 { turns++; return 0 })
	o.Stub(o.IDA(0x1A8DF), func(*oracle.Oracle) uint32 { waits++; return 0 })

	type op struct {
		group  int
		calls  int
		redraw bool
	}
	plan := []op{{3, 10, true}, {4, 7, false}, {0, 8, false}, {3, 5, true}, {3, 6, false}, {2, 9, false}, {4, 8, true}}
	s := &world.Screen{}
	n := 0
	for _, step := range plan {
		if step.group != int(o.Word(o.IDA(0x2DF5A))) || n == 0 {
			turns = 0
		}
		o.WriteU16(o.IDA(0x2DF5A), uint16(step.group))
		s.Group = step.group
		for c := 0; c < step.calls; c++ {
			redraw := step.redraw && c == 0
			draws, waits = nil, 0
			if redraw {
				o.WriteU16(o.IDA(0x2DF28), 0)
			}
			if _, err := o.CallBudget(100_000, o.IDA(0x11390)); err != nil {
				t.Fatal(err)
			}
			var waited bool
			if redraw {
				s.WeatherRedraw()
			} else {
				waited = s.WeatherCall()
			}
			if waited != (waits == 1) || waited == (len(draws) == 1) {
				t.Fatalf("第 %d 次（組 %d）：原版 等待 %d 繪製 %d；remake 等待=%t", n, step.group, waits, len(draws), waited)
			}
			if !waited {
				want := turns
				if step.group == 3 || step.group == 4 {
					want = int(draws[0]-0x9000) / 0x7C
				}
				if s.Phase != want {
					t.Fatalf("第 %d 次（組 %d）：原版畫第 %d 格，remake 相位 %d", n, step.group, want, s.Phase)
				}
			}
			n++
		}
	}
	t.Logf("天候節拍：%d 次呼叫逐次一致（含換組與地圖重畫）", n)
}
