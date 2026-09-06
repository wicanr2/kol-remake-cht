//go:build oracle

package oracle

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"

	"github.com/wicanr2/kol-remake-cht/internal/rng"
)

// 原版的位址（KOLTOWN-unpacked.EXE，IDA 線性位址；RE-132）。
const (
	// mainEnter 是 `_main` 的進入點。C 的啟動碼到這一刻已經把 DS 設成 DGROUP，
	// 全域讀得到——這是「直接呼叫原版常式」最早的安全點，兩百多道指令就到。
	mainEnter = 0x10000
	// seedLow／seedHigh 是種子那兩個 word（`word_2F4D0`／`word_2F4D2`）。
	seedLow  = 0x2F4D0
	seedHigh = 0x2F4D2
	// randEnter 是 `rand()`；rollEnter 是包在外面的 `rand(n)`（`sub_1B5A8`）。
	randEnter = 0x2D538
	rollEnter = 0x1B5A8
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

// ready 把原版跑到 `_main` 的進入點，再把種子壓成 seed。
//
// # 為什麼不等遊戲自己抽
//
// 等遊戲走到會抽亂數的地方，得先解出開機、選單、建隊、進戰鬥每一步的按鍵序列，
// 而**中間任何一步錯了，看到的都是「一次都沒抽」**——與「原版根本不抽」長得
// 一模一樣。實測跑滿 2.6 億道指令、送 12 批鍵仍然是 0 次。
//
// 直接呼叫走的是同一段機器碼、同一份全域種子，所以驗到的就是原版的實作。
// 代價是**這裡不驗呼叫時機**：誰在什麼時候抽、抽幾次，仍然只能靠實際跑遊戲確認。
func ready(t *testing.T, o *oracle.Oracle, seed int) {
	t.Helper()
	if err := o.RunUntil(oracle.At(o.IDA(mainEnter)), oracle.Budget(2_000_000)); err != nil {
		t.Fatalf("跑不到 _main：%v", err)
	}
	o.WriteU16(o.IDA(seedLow), uint16(seed))
	o.WriteU16(o.IDA(seedHigh), uint16(seed>>16))
}

// seedOf 讀原版現在的 32 位元種子。
func seedOf(o *oracle.Oracle) uint32 {
	return uint32(o.Word(o.IDA(seedHigh)))<<16 | uint32(o.Word(o.IDA(seedLow)))
}

// 正對照：`_main` 進入時種子是 MS C runtime 的初值 1。
//
// 這一條先於底下每一條。種子讀到 1，就同時證明了三件事：`IDA()` 的映射對、
// `_main` 這一刻的 DS 已經是 DGROUP、`word_2F4D0` 真的是種子。**沒有它的話，
// 底下對不上時分不出是 remake 錯還是我讀錯了位址。**
func TestSeedStartsAtOne(t *testing.T) {
	o := load(t)
	if err := o.RunUntil(oracle.At(o.IDA(mainEnter)), oracle.Budget(2_000_000)); err != nil {
		t.Fatalf("跑不到 _main：%v", err)
	}
	if got := seedOf(o); got != 1 {
		t.Fatalf("`_main` 進入時的種子是 %#08x，期望 MS C 的初值 1"+
			"（位址或 DS 不對，不是 remake 的問題）", got)
	}
	t.Logf("第 %d 道指令到 _main，種子 = 1", o.Steps())
}

// 原版的 `rand()` 與 remake 的 `rng.LCG` 是同一串數。
//
// 這一條把 RE-132 的三個常數從「反組譯讀出來的」升成「實跑量到的」：判準是原版
// 自己在執行期回的 AX，不是我對指令的解讀。
func TestRandStreamMatchesOriginal(t *testing.T) {
	const want = 500
	for _, seed := range []int{0, 1, 23, 141, 4242, 65535} {
		o := load(t)
		ready(t, o, seed)
		mine := rng.New(seed)
		for i := 0; i < want; i++ {
			got, err := o.Call(o.IDA(randEnter))
			if err != nil {
				t.Fatalf("種子 %d 第 %d 抽：%v", seed, i, err)
			}
			if raw := mine.Next(); raw != int(uint16(got)) {
				t.Fatalf("種子 %d 第 %d 抽：原版 %d，remake %d",
					seed, i, uint16(got), raw)
			}
		}
		if state := seedOf(o); state != mine.State() {
			t.Fatalf("種子 %d 抽 %d 次之後：原版 %#08x，remake %#08x",
				seed, want, state, mine.State())
		}
		t.Logf("種子 %d：%d 抽逐次相同，收尾狀態 %#08x", seed, want, mine.State())
	}
}

// `rand(n)` 那一層也一樣：同一個 n 餵進去，回的 1…n 逐次相同。
//
// 這一層是遊戲真正在用的入口（擲骰、命中、AI 選招都經過它），所以它比
// 裸 `rand()` 更接近玩家看得到的行為。
func TestRollMatchesOriginal(t *testing.T) {
	o := load(t)
	ready(t, o, 23)
	mine := rng.New(23)
	for _, n := range []int{2, 3, 4, 6, 8, 10, 12, 20, 100, 1000} {
		for i := 0; i < 40; i++ {
			got, err := o.Call(o.IDA(rollEnter), uint16(n))
			if err != nil {
				t.Fatalf("rand(%d) 第 %d 次：%v", n, i, err)
			}
			roll := int(int16(uint16(got)))
			if want := mine.Roll(n); roll != want {
				t.Fatalf("rand(%d) 第 %d 次：原版 %d，remake %d", n, i, roll, want)
			}
			if roll < 1 || roll > n {
				t.Fatalf("rand(%d) 回了 %d，超出 1…%d", n, roll, n)
			}
		}
		t.Logf("rand(%d)：40 次逐次相同", n)
	}
	if state := seedOf(o); state != mine.State() {
		t.Fatalf("收尾種子：原版 %#08x，remake %#08x", state, mine.State())
	}
}

// `rand(1)` 只有一個面：原版與 remake 都必須回 1。
func TestRollOfOne(t *testing.T) {
	o := load(t)
	ready(t, o, 7)
	mine := rng.New(7)
	for i := 0; i < 10; i++ {
		got, err := o.Call(o.IDA(rollEnter), 1)
		if err != nil {
			t.Fatalf("rand(1) 第 %d 次：%v", i, err)
		}
		if roll, want := int(int16(uint16(got))), mine.Roll(1); roll != 1 || want != 1 {
			t.Fatalf("rand(1) 第 %d 次：原版 %d，remake %d", i, roll, want)
		}
	}
	if state := seedOf(o); state != mine.State() {
		t.Fatalf("收尾種子：原版 %#08x，remake %#08x", state, mine.State())
	}
}

// `rand(0)` 在原版是**除以零**，整個行程帶著 MS C 的 R6003 離開。
//
// remake 的 `Roll(0)` 回 0 不當機，所以這一格是**故意的不一致**：原版沒有這道
// 保護，代表原版永遠不會拿 0 去呼叫它。反過來說，remake 只要有一條路徑真的傳了
// 0 進去，那條路徑在原版就走不到——**回 0 是在掩蓋上游的錯，不是在對齊原版**。
//
// 這一條把那件事釘住：原版的行為是崩潰，不是某個「安全的預設值」。
func TestRollOfZeroCrashesOriginal(t *testing.T) {
	o := load(t)
	ready(t, o, 7)
	_, err := o.Call(o.IDA(rollEnter), 0)
	if err == nil {
		t.Fatal("原版 rand(0) 沒有當掉——那表示它有除零保護，remake 要跟上")
	}
	var exit *oracle.ExitError
	if !errors.As(err, &exit) {
		t.Fatalf("原版 rand(0) 停下來的方式不是離開行程：%v", err)
	}
	if !strings.Contains(exit.Console, "divide by 0") {
		t.Fatalf("離開了，但不是除零：%q", exit.Console)
	}
	t.Logf("原版 rand(0)：回傳碼 %d，%q", exit.Code, strings.TrimSpace(exit.Console))
}
