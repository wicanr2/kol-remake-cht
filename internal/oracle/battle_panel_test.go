//go:build oracle

package oracle

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
	"github.com/wicanr2/kol-remake-cht/internal/battle"
)

// 觀察原版左下框的繪製參數；以 arg_0=0 排除命令圖示，不作正常 GUI 證據。
func TestBattleStatusBarOriginal(t *testing.T) {
	o := prepare(t)
	o.StubValue(o.IDA(0x1BA9D), 0)
	for _, c := range []struct{ slot, sex, stamina, fatigue, y int }{
		{0, 0, 100, 0, 84}, {0, 0, 100, 1, 84},
		{0, 0, 100, 2, 84}, {0, 1, 100, 3, 85},
		{0, 1, 100, 25, 93}, {0, 0, 100, 50, 102},
		{0, 0, 100, 99, 120}, {0, 0, 100, 100, 120},
		{6, 0, 100, 199, 155}, {6, 0, 100, 200, 155},
		{6, 0, 100, 201, 155},
		{0, 0, 7, 1, 94},
	} {
		o.WriteU16(o.DS(0x9E5A), uint16(c.slot))
		o.WriteU16(o.DS(uint16(0x1C50+2*c.slot)), uint16(c.stamina))
		o.WriteU16(o.DS(uint16(0x322E+2*c.slot)), uint16(c.fatigue))
		if c.slot < 6 {
			o.WriteU8(o.DS(uint16(0xBB88+0x300*c.slot+0x1A)), byte(c.sex))
		}
		o.WriteU16(o.IDA(0x2DF1E), 0)
		var rects [][5]uint16
		o.Stub(o.IDA(0x1B779), func(p *oracle.Oracle) uint32 {
			rects = append(rects, [5]uint16{p.Arg(0), p.Arg(1), p.Arg(2), p.Arg(3), p.Arg(4)})
			return 0
		})
		var bodies []uint16
		o.Stub(o.IDA(0x27E20), func(p *oracle.Oracle) uint32 {
			bodies = append(bodies, p.Arg(0))
			return 0
		})
		if _, err := o.Call(o.IDA(0x2324F), 0); err != nil {
			t.Fatal(err)
		}
		wantBody := uint16(0x47 + c.sex)
		if c.slot >= 6 {
			wantBody = 0x49
		}
		want := [][5]uint16{{132, 84, 138, uint16(c.y), 12}, {132, uint16(c.y), 138, 156, 1}}
		boundary, err := battle.BodyFatigueBoundary(c.stamina, c.fatigue)
		if err != nil || boundary != c.y {
			t.Fatalf("remake 疲勞分界=%d 預期=%d 錯誤=%v", boundary, c.y, err)
		}
		if len(rects) != 2 || rects[0] != want[0] || rects[1] != want[1] || len(bodies) != 1 || bodies[0] != wantBody {
			t.Fatalf("slot=%d fatigue=%d：矩形=%v 預期=%v；人形=%v 預期=%x", c.slot, c.fatigue, rects, want, bodies, wantBody)
		}
	}
	t.Log("十二組疲勞直條邊界及男女／敵方人形參數通過；保留先除二再換百分比的整數截斷")
}

// 真正執行 sub_27E20；僅替代底圖與像素介面，以觀察遮罩座標、順序及顏色。
// 表格由 IDA linear 0x2F12C 的 72 bytes 獨立匯出，不由 remake 產生。
func TestBattleBodyOverlayOriginal(t *testing.T) {
	rects := [3][6][4]uint16{
		{{96, 86, 105, 96}, {91, 96, 110, 116}, {78, 97, 91, 119}, {110, 97, 123, 119}, {80, 120, 99, 155}, {101, 120, 121, 155}},
		{{97, 87, 104, 97}, {92, 97, 108, 118}, {81, 98, 92, 119}, {107, 98, 120, 119}, {83, 119, 99, 155}, {102, 119, 118, 155}},
		{{89, 87, 112, 98}, {89, 98, 112, 119}, {75, 98, 89, 121}, {112, 98, 126, 121}, {81, 121, 97, 155}, {104, 121, 120, 155}},
	}
	o := prepare(t)
	for kind, parts := range rects {
		for part, r := range parts {
			for j, v := range r {
				if got := o.Byte(o.IDA(uint32(0x2F12C + kind*24 + part*4 + j))); got != byte(v) {
					t.Fatalf("原版矩形資料不同：kind=%d part=%d byte=%d got=%d want=%d", kind, part, j, got, v)
				}
			}
			for _, hp := range [][2]byte{{100, 100}, {99, 100}, {50, 100}, {1, 100}, {0, 100}, {255, 100}, {128, 100}, {1, 7}} {
				t.Run(fmt.Sprintf("人形%d/部位%d/%d_%d", kind, part, hp[0], hp[1]), func(t *testing.T) {
					slot := 2
					if kind == 2 {
						slot = 6
					}
					o.WriteU16(o.DS(0x9E5A), uint16(slot))
					for limb := 0; limb < 6; limb++ {
						base := uint16(0x6712 + slot*50 + 0x11 + limb*2)
						o.WriteU8(o.DS(base), 100)
						o.WriteU8(o.DS(base+1), 100)
					}
					base := uint16(0x6712 + slot*50 + 0x11 + part*2)
					o.WriteU8(o.DS(base), hp[0])
					o.WriteU8(o.DS(base+1), hp[1])
					var bodies []uint16
					var reads, wantReads [][2]uint16
					var writes, wantWrites [][3]uint16
					o.Stub(o.IDA(0x1103F), func(p *oracle.Oracle) uint32 {
						bodies = append(bodies, p.Arg(0))
						return 0
					})
					o.Stub(o.IDA(0x1C94E), func(p *oracle.Oracle) uint32 {
						x, y := p.Arg(0), p.Arg(1)
						reads = append(reads, [2]uint16{x, y})
						return uint32((x + 3*y) % 16)
					})
					o.Stub(o.IDA(0x1C9AB), func(p *oracle.Oracle) uint32 {
						writes = append(writes, [3]uint16{p.Arg(0), p.Arg(1), p.Arg(2)})
						return 0
					})
					current, maximum := [6]int{100, 100, 100, 100, 100, 100}, [6]int{100, 100, 100, 100, 100, 100}
					current[part], maximum[part] = int(hp[0]), int(hp[1])
					masks, err := battle.BodyDamageRects(kind, current, maximum)
					if err != nil {
						t.Fatal(err)
					}
					for _, mask := range masks {
						for y := uint16(mask.Top); y < uint16(mask.Bottom); y++ {
							for x := uint16(mask.Left); x < uint16(mask.Right); x++ {
								wantReads = append(wantReads, [2]uint16{x, y})
								c := (x + 3*y) % 16
								if c == 7 || c == 8 || c == 15 {
									wantWrites = append(wantWrites, [3]uint16{x, y, 12})
								}
							}
						}
					}
					if _, err := o.Call(o.IDA(0x27E20), uint16(0x47+kind)); err != nil {
						t.Fatal(err)
					}
					if !reflect.DeepEqual(bodies, []uint16{uint16(0x47 + kind)}) || !reflect.DeepEqual(reads, wantReads) || !reflect.DeepEqual(writes, wantWrites) {
						t.Fatalf("人形=%v；讀取=%d/%d；寫入=%d/%d，座標、順序或顏色不符", bodies, len(reads), len(wantReads), len(writes), len(wantWrites))
					}
				})
			}
		}
	}
}

func TestBattleBodySpriteOriginal(t *testing.T) {
	o := prepare(t)
	const cacheSegment = 0x7777
	o.WriteU16(o.IDA(0x2FBB8), cacheSegment)
	for kind := 0; kind < 3; kind++ {
		calls := 0
		o.Stub(o.IDA(0x11262), func(p *oracle.Oracle) uint32 {
			calls++
			if p.Arg(0) != uint16(3+kind) || p.Arg(1) != cacheSegment {
				t.Fatalf("人形 %d 路由錯誤：%d %x", kind, p.Arg(0), p.Arg(1))
			}
			got := [4]uint16{p.Word(p.IDA(0x3499E)), p.Word(p.IDA(0x349A0)), p.Word(p.IDA(0x34A06)), p.Word(p.IDA(0x34A0E))}
			if got != [4]uint16{9, 85, 16, 155} {
				t.Fatalf("人形目的區域：%v", got)
			}
			return 0
		})
		if _, err := o.Call(o.IDA(0x1103F), uint16(0x47+kind)); err != nil {
			t.Fatal(err)
		}
		if calls != 1 {
			t.Fatalf("人形 %d 繪製次數 %d", kind, calls)
		}
	}
	t.Logf("原版圖片格式字串=%q；模式副檔字串表=%q", o.Bytes(o.DS(0x1678), 10), o.Bytes(o.DS(0x670), 25))
}
