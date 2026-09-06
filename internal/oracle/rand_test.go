//go:build oracle

package oracle

import (
	"os"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"

	"github.com/wicanr2/kol-remake-cht/internal/rng"
)

// 原版的位址（KOLTOWN-unpacked.EXE，IDA 線性位址；RE-132）。
const (
	// srandReturn 是 `_main` 呼叫 `srand(BIOS tick)` 之後的下一道指令。
	// 種子要在這一刻覆蓋掉，之後兩邊才在同一個起點。
	srandReturn = 0x10145
	// seedLow／seedHigh 是種子那兩個 word（`word_2F4D0`／`word_2F4D2`）。
	seedLow  = 0x2F4D0
	seedHigh = 0x2F4D2
	// randEnter／randLeave 是 `rand()` 的進入點與 `retf`。
	randEnter = 0x2D538
	randLeave = 0x2D55F
	// rollEnter／rollLeave 是 `rand(n)` 那一層（`sub_1B5A8`）。
	rollEnter = 0x1B5A8
	rollLeave = 0x1B5C7
)

// 原版檔案由玩家自備，只留在本機。沒有就整支跳過——不假裝驗過。
const (
	exePath  = "../../workplace/analysis/koltown/KOLTOWN-unpacked.EXE"
	rootPath = "../../workplace/oracle/game"
)

func load(t *testing.T) *oracle.Oracle {
	t.Helper()
	if _, err := os.Stat(exePath); err != nil {
		t.Skipf("沒有原版執行檔（%s）：%v", exePath, err)
	}
	if _, err := os.Stat(rootPath); err != nil {
		t.Skipf("沒有原版素材目錄（%s）：%v", rootPath, err)
	}
	o, err := oracle.Load(exePath, rootPath)
	if err != nil {
		t.Fatalf("載入原版：%v", err)
	}
	t.Cleanup(o.Close)
	return o
}

// draw 是原版抽的一次亂數。
type draw struct {
	// N 是 `rand(n)` 的引數；0 表示這一次是直接呼叫 `rand()`，沒有經過那一層。
	N int
	// Raw 是 `rand()` 回的 0…0x7FFF；Roll 是 `rand(n)` 回的 1…n。
	Raw, Roll int
}

// collect 讓原版從冷啟動跑到抽夠 want 次為止，回傳抽取序列。
//
// 種子在 `srand` 回來的那一刻被覆蓋成 seed，所以序列與 `rng.New(seed)` 同一個起點。
func collect(t *testing.T, o *oracle.Oracle, seed, want int, budget uint64) []draw {
	t.Helper()
	if err := o.RunUntil(At(o, srandReturn), oracle.Budget(budget)); err != nil {
		t.Fatalf("跑不到 srand 的返回點：%v", err)
	}
	o.WriteU16(o.IDA(seedLow), uint16(seed))
	o.WriteU16(o.IDA(seedHigh), 0)

	var out []draw
	var pending *draw
	o.OnCall(o.IDA(rollEnter), func(p *oracle.Oracle) {
		pending = &draw{N: int(p.Arg(0))}
	})
	o.OnCall(o.IDA(randLeave), func(p *oracle.Oracle) {
		d := draw{}
		if pending != nil {
			d.N = pending.N
		}
		d.Raw = int(p.AX())
		out = append(out, d)
	})
	o.OnCall(o.IDA(rollLeave), func(p *oracle.Oracle) {
		if len(out) > 0 {
			out[len(out)-1].Roll = int(p.AX())
		}
		pending = nil
	})

	err := o.RunUntil(oracle.NewCond("抽夠了", func(*oracle.Oracle) bool {
		return len(out) >= want
	}), oracle.Budget(budget))
	if err != nil && len(out) < want {
		t.Fatalf("只抽到 %d 次（要 %d）：%v", len(out), want, err)
	}
	return out
}

// 原版的 `rand()` 與 remake 的 `rng.LCG` 是同一串數。
//
// 這一條把 RE-132 的三個常數從「反組譯讀出來的」升成「實跑量到的」：判準是原版
// 自己在執行期回的 AX，不是我對指令的解讀。
func TestRandStreamMatchesOriginal(t *testing.T) {
	const want = 200
	for _, seed := range []int{0, 1, 23, 141, 4242, 65535} {
		o := load(t)
		got := collect(t, o, seed, want, 400_000_000)
		mine := rng.New(seed)
		for i, d := range got {
			if raw := mine.Next(); raw != d.Raw {
				t.Fatalf("種子 %d 第 %d 抽：原版 %d，remake %d",
					seed, i, d.Raw, raw)
			}
		}
		t.Logf("種子 %d：%d 抽逐次相同", seed, len(got))
	}
}

// `rand(n)` 那一層也一樣：同一串 n 餵進去，回的 1…n 逐次相同。
func TestRollMatchesOriginal(t *testing.T) {
	const want = 200
	o := load(t)
	got := collect(t, o, 23, want, 400_000_000)
	mine := rng.New(23)
	rolls := 0
	for i, d := range got {
		if d.N == 0 {
			// 沒有經過 `rand(n)` 那一層的直接呼叫：只推進種子。
			if raw := mine.Next(); raw != d.Raw {
				t.Fatalf("第 %d 抽（直接呼叫）：原版 %d，remake %d", i, d.Raw, raw)
			}
			continue
		}
		if roll := mine.Roll(d.N); roll != d.Roll {
			t.Fatalf("第 %d 抽 rand(%d)：原版 %d，remake %d", i, d.N, d.Roll, roll)
		}
		rolls++
	}
	if rolls == 0 {
		t.Fatal("一次都沒有經過 rand(n)——攔截點不對，不是「原版沒有抽」")
	}
	t.Logf("%d 抽裡有 %d 次走 rand(n)，逐次相同", len(got), rolls)
}

// 種子那兩個 word 就是 remake 的 32 位元狀態：抽完之後兩邊的種子相同。
func TestSeedStateMatchesOriginal(t *testing.T) {
	o := load(t)
	got := collect(t, o, 4242, 50, 400_000_000)
	mine := rng.New(4242)
	for range got {
		mine.Next()
	}
	lo := uint32(o.Word(o.IDA(seedLow)))
	hi := uint32(o.Word(o.IDA(seedHigh)))
	if state := hi<<16 | lo; state != mine.State() {
		t.Fatalf("抽 %d 次之後的種子：原版 %#08x，remake %#08x",
			len(got), state, mine.State())
	}
}
