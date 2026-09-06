//go:build oracle

package oracle

import (
	"crypto/sha256"
	"fmt"
	"encoding/hex"
	"image"
	"image/color"
	"image/png"
	"os"
	"sort"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
)

// 探測：原版開機之後停在哪、在等什麼。
//
// 這是 dosgolem 第四層（把原版推進戰鬥）的第一步。判準不是猜，是原版自己留下的
// 路標：開過哪些檔、缺哪些檔、視訊模式切到哪、主控台說了什麼、有沒有用到沒實作
// 的服務。**「跑得動」與「跑得動但走錯路」的差別就在這幾項。**
func TestProbeBootState(t *testing.T) {
	o := load(t)
	var rands int
	o.OnCall(o.IDA(randEnter), func(*oracle.Oracle) { rands++ })

	if err := o.Run(40_000_000); err != nil {
		t.Logf("停了：%v", err)
	}
	t.Logf("跑了 %d 道指令，停在 %s", o.Steps(), o.IP())
	t.Logf("視訊模式 %#02x（03h ＝ 還在文字模式）", o.VideoMode())
	t.Logf("開過的檔：%v", o.Opened())
	t.Logf("開不起來的檔：%v", o.Missing())
	t.Logf("抽了 %d 次亂數", rands)
	if c := o.Console(); c != "" {
		t.Logf("主控台：%q", c)
	}
	if u := o.Unimplemented(); len(u) > 0 {
		t.Logf("沒實作的服務：%v", u)
	}
	nz := 0
	for _, v := range o.Indexed() {
		if v != 0 {
			nz++
		}
	}
	t.Logf("畫面非零像素 %d/64000", nz)

	// 畫面上有什麼？**`Indexed()` 讀的是圖形記憶體**，程式還在文字模式時那裡
	// 當然是全黑——而「全黑」看起來就像「什麼都沒畫」。先看模式再選讀哪一份。
	if o.VideoMode() == 0x03 {
		for i, line := range o.TextScreen() {
			if line != "" {
				t.Logf("文字畫面第 %2d 列：%q", i, line)
			}
		}
	}

	// 在等誰？滑鼠輪詢次數是最直接的答案。
	polls, pressed := o.MouseActivity()
	t.Logf("滑鼠輪詢 %d 次（其中回報按著 %d 次）；`int 33h` 各功能號：%v",
		polls, pressed, o.MouseCalls())

	// 在等什麼？輪詢的埠就是答案：`3DA` 是 CRT 狀態（等垂直回掃）、
	// `60`／`64` 是鍵盤、`40`–`43` 是計時器、`388` 是 OPL。
	t.Logf("讀過的 I/O 埠（次數多的在前）：%v", o.PortReads())
	// 寫過的埠是「設了什麼」：EGA／VGA 直接設模式走 3C0–3CF 與 3D4／3D5，
	// 完全不經過 `int 10h`——那時 `VideoMode()` 會一直回 03h 而畫面早就是圖形了。
	t.Logf("寫過的 I/O 埠：%v", o.PortWrites())

	// 停在哪一支？把執行期位址換回 IDA，再把那裡的 bytes 印出來——
	// 「在等輸入」與「跑飛了」在只看位址時分不出來。
	t.Logf("停止點 IDA %#x，bytes %x", o.ToIDA(o.IP()), o.Bytes(o.IP(), 16))

	// 是不是在原地打轉：再跑幾段，看落點集中在哪。
	seen := map[uint32]int{}
	for i := 0; i < 200; i++ {
		if err := o.Run(2_000); err != nil {
			t.Logf("第 %d 次取樣停了：%v", i, err)
			break
		}
		seen[o.ToIDA(o.IP())]++
	}
	t.Logf("兩百次取樣落在 %d 個不同位址", len(seen))
	for ea, n := range seen {
		if n >= 5 {
			t.Logf("  IDA %#x：%d 次", ea, n)
		}
	}
}

// 探測：送鍵推得動嗎。
//
// 開機之後原版停在 IDA 0x1B126…0x1B566 那一段裡打轉，**沒有讀任何 I/O 埠**
// ——不是在等硬體。DOS 的鍵盤走 `int 16h`／`int 21h AH=3Fh`，兩條都不碰埠，
// 所以「在等鍵盤」是最可能的解釋。這一條逐個鍵試，看哪一個讓它離開那一段。
//
// 判準是三個路標一起看：視訊模式（切到圖形模式沒）、畫面非零像素（畫東西沒）、
// 停止位置（離開那一段沒）。**只看其中一個會誤判**——畫面有東西不代表推進了，
// 位置變了也可能只是換一個等待迴圈。
func TestProbeKeysAdvanceBoot(t *testing.T) {
	o := load(t)
	if err := o.Run(20_000_000); err != nil {
		t.Logf("開機停了：%v", err)
	}
	report := func(label string) (mode uint8, nz int, ea uint32) {
		nz = 0
		for _, v := range o.Indexed() {
			if v != 0 {
				nz++
			}
		}
		mode, ea = o.VideoMode(), o.ToIDA(o.IP())
		t.Logf("%s：模式 %#02x、非零像素 %d、IDA %#x、步數 %d",
			label, mode, nz, ea, o.Steps())
		return
	}
	report("開機後")

	for _, key := range []struct {
		name string
		send string
	}{
		{"空白", " "}, {"Enter", "\r"}, {"ESC", "\x1b"},
		{"Y", "Y"}, {"N", "N"}, {"1", "1"}, {"A", "A"},
	} {
		o.Type(key.send)
		if err := o.Run(10_000_000); err != nil {
			t.Logf("送「%s」之後停了：%v", key.name, err)
			break
		}
		mode, nz, _ := report("送「" + key.name + "」")
		if mode != 0x03 || nz > 0 {
			t.Logf("→ 推進了：模式 %#02x、畫面 %d 個非零像素", mode, nz)
			return
		}
	}
	t.Log("七個鍵都推不動——要換別的辦法問它在等什麼")
}

// 探測：開機時是誰在等滑鼠。
//
// `sub_1B100`（IDA 0x1B100）是輪詢 `int 33h` 的那一支，全庫有 49 個呼叫端。
// 程式停在它裡面，而畫面是空的、一個 I/O 埠都沒寫過——所以它在**畫面初始化之前**
// 就進了等滑鼠的迴圈。`Caller()` 直接回答是哪一個呼叫端。
func TestProbeWhoWaitsForMouse(t *testing.T) {
	o := load(t)
	// 正對照先行：`0x1B12E`（那一道 `int 33h`）一定會被執行，取樣就落在它上面。
	// 它不觸發就表示位址換算有問題，而不是「函式沒被呼叫」。
	hits := map[uint32]int{}
	callers := map[uint32]int{}
	probes := []uint32{0x1B100, 0x1B101, 0x1B103, 0x1B107, 0x1B109,
		0x1B10F, 0x1B119, 0x1B11E, 0x1B121, 0x1B126, 0x1B12E, 0x1B566}
	for _, ea := range probes {
		ea := ea
		o.OnCall(o.IDA(ea), func(p *oracle.Oracle) {
			hits[ea]++
			if ea == 0x1B100 {
				callers[p.ToIDA(p.Caller())]++
			}
		})
	}
	if err := o.Run(40_000_000); err != nil {
		t.Logf("停了：%v", err)
	}
	for _, ea := range probes {
		t.Logf("IDA %#x 觸發 %d 次", ea, hits[ea])
	}
	if hits[0x1B12E] == 0 {
		t.Fatal("那一道 `int 33h` 都沒觸發——位址換算有問題，不是函式沒被呼叫")
	}
	if len(callers) == 0 {
		t.Log("程式沒有從 `0x1B100` 進來過——入口不是 IDA 標的函式起點，" +
			"或者它是被跳進去的")
		return
	}
	for ea, n := range callers {
		t.Logf("呼叫端 IDA %#x：%d 次", ea, n)
	}
}

// 探測：程式是在哪一刻飛掉的。
//
// `sub_1B100` 的 prologue（0x1B100…0x1B11E）**一次都沒執行**，而迴圈體
// （0x1B121…0x1B566）跑了 65 萬次——程式是**跳進迴圈中間**的，不是呼叫進來的。
// 加上 `CS = 0`（正常的 far call 目標段不會是 0），這是跑飛的樣子：線性位址剛好
// 落在程式碼區，於是「看起來像在執行正常的程式碼」。
//
// 這一條找出 `CS` 第一次變成 0 的那一刻——那是跳轉的落點，往回追才有起點。
func TestProbeWhereItDerails(t *testing.T) {
	o := load(t)
	err := o.RunUntil(oracle.NewCond("CS 變成 0", func(p *oracle.Oracle) bool {
		return p.IP().Seg == 0
	}), oracle.Budget(40_000_000))
	if err != nil {
		t.Skipf("四千萬道指令內 CS 沒有變成 0：%v", err)
	}
	t.Logf("第 %d 道指令時 CS 變成 0，停在 %s（IDA %#x）",
		o.Steps(), o.IP(), o.ToIDA(o.IP()))
	t.Logf("落點的 bytes：%x", o.Bytes(o.IP(), 16))
	t.Logf("此時開過的檔：%v", o.Opened())
	t.Logf("堆疊頂端四個 word：%04x %04x %04x %04x",
		o.StackWord(0), o.StackWord(1), o.StackWord(2), o.StackWord(3))

	// **哪一道指令把 CS 變成 0**才是根因。堆疊頂端不可靠——`jmp far` 不推返回
	// 位址，而 `retf` 到垃圾時頂端已經被彈掉了。重跑一次，停在前一道。
	derail := o.Steps()
	o2 := load(t)
	if err := o2.Run(derail - 1); err != nil {
		t.Fatalf("重跑到前一道：%v", err)
	}
	t.Logf("最後一道正常指令在 %s（IDA %#x），bytes %x",
		o2.IP(), o2.ToIDA(o2.IP()), o2.Bytes(o2.IP(), 8))
	// 那一道是 `mov es, dx`——它改不了 CS。所以換 CS 的是**中斷**：
	// 看向量表就知道被送進哪一個空向量。
	for _, v := range []int{0x08, 0x09, 0x1B, 0x1C, 0x21, 0x33} {
		b := o2.Bytes(oracle.Addr{Seg: 0, Off: uint16(v * 4)}, 4)
		t.Logf("int %02Xh 的向量：%02x%02x:%02x%02x",
			v, b[3], b[2], b[1], b[0])
	}
	if err := o2.Run(1); err != nil {
		t.Logf("再走一道：%v", err)
	}
	t.Logf("走完之後在 %s（IDA %#x）", o2.IP(), o2.ToIDA(o2.IP()))
}

// 探測：誰把 `int 08h` 的向量清成 0。
//
// dosgolem 開機時把 256 個向量全部指到 StubSeg（0x0080）的 `iret`，而程式飛掉那一刻
// `int 08h` 與 `int 09h` 都是 `0000:0000`——**是程式自己改的**，而且不是走
// `int 21h AH=25h`（那一支會把 DS:DX 寫進去，兩支都實作正確），所以它直接寫記憶體。
//
// 向量一旦是 0，下一次計時器中斷就把 CS:IP 換成 `0000:0000`，程式從那裡開始
// 「執行」全 0 的記憶體（`add [bx+si], al` 一路解得下去），最後落進程式碼區的某個
// 迴圈——**看起來像在跑正常的程式**。這是整條線最難看穿的一段。
func TestProbeWhoClearsTimerVector(t *testing.T) {
	const stubSeg = 0x0080
	o := load(t)
	err := o.RunUntil(oracle.NewCond("int 08h 的向量被改", func(p *oracle.Oracle) bool {
		return p.Word(oracle.Addr{Seg: 0, Off: 0x08*4 + 2}) != stubSeg
	}), oracle.Budget(40_000_000))
	if err != nil {
		t.Skipf("四千萬道指令內沒被改：%v", err)
	}
	seg := o.Word(oracle.Addr{Seg: 0, Off: 0x08*4 + 2})
	off := o.Word(oracle.Addr{Seg: 0, Off: 0x08 * 4})
	t.Logf("第 %d 道指令之後 int 08h 變成 %04X:%04X", o.Steps(), seg, off)
	t.Logf("改完之後停在 %s（IDA %#x），接下來的 bytes：%x",
		o.IP(), o.ToIDA(o.IP()), o.Bytes(o.IP(), 12))
	t.Logf("此時開過的檔：%v", o.Opened())
	// 段暫存器才是答案：搬移或清空迴圈寫錯地方時，錯的通常是 `ES`／`DS`，
	// 而那從指令本身看不出來。
	t.Logf("暫存器：%s", o.Regs())
}

// 探測：`ES` 是在哪裡被設成 0 的。
//
// 覆寫向量表的那個迴圈用 `ES:BX` 寫，而當下 `ES = 0000`。程式初始的 `ES` 是 PSP
// 段（非 0），所以**第一次變成 0 的那一刻**就是可疑點。原版在真 DOS 上跑得起來，
// 所以那個 0 多半是某個服務回的——量出位置才知道是哪一支。
func TestProbeWhereESGoesZero(t *testing.T) {
	o := load(t)
	for i := 0; i < 6; i++ {
		err := o.RunUntil(oracle.NewCond("ES 變成 0", func(p *oracle.Oracle) bool {
			return p.Regs().ES == 0
		}), oracle.Budget(40_000_000))
		if err != nil {
			t.Logf("第 %d 次：四千萬道指令內 ES 沒有變成 0：%v", i, err)
			return
		}
		t.Logf("第 %d 次：步數 %d，在 %s（IDA %#x），%s",
			i, o.Steps(), o.IP(), o.ToIDA(o.IP()), o.Regs())
		t.Logf("    接下來的 bytes：%x", o.Bytes(o.IP(), 12))
		// 走過這一段再找下一次，否則條件立刻又成立。
		if err := o.Run(200); err != nil {
			t.Logf("往前走停了：%v", err)
			return
		}
	}
}

// 探測：覆寫向量表的那個迴圈長什麼樣。
//
// `ES = 0` 本身是正常的——`_main` 在第 306 道用 `es:[0449]` 讀 BIOS 資料區的視訊
// 模式，之後就沒再改。問題是 `0x1DC59` 那個迴圈**用了沒設好的 `ES`**。
// 前面那幾十個 byte 會說它是從哪裡拿到段的。
func TestProbeDerailLoopBytes(t *testing.T) {
	o := load(t)
	for _, c := range []struct {
		name string
		ea   uint32
		n    int
	}{
		{"覆寫迴圈往前 0x40", 0x1DC20, 0x50},
		{"讀 far 指標那一段", 0x1A1C0, 0x30},
	} {
		t.Logf("%s（IDA %#x）：%x", c.name, c.ea, o.Bytes(o.IDA(c.ea), c.n))
	}
}

// 探測：誰設了 `ds:1C84`——覆寫向量表那個迴圈的目標段。
//
// 那個迴圈是 `mov es, [ds:1C84]` ／ `rep movsw`：目標段從一個全域來，而那個全域
// 是 0。**誰把它設成 0（或者從來沒設過）才是根因。**
func TestProbeWhoSetsDestSegment(t *testing.T) {
	o := prepare(t) // 要先對好 DGROUP，`WatchWrites` 收的是 DGROUP 偏移
	log := o.WatchWrites(0x1C84, 0x1C86)
	before := o.Word(o.DS(0x1C84))
	t.Logf("`_main` 進入時 ds:1C84 ＝ %04X（IDA %#x）", before, dgroupBase+0x1C84)

	if err := o.Run(2_000_000); err != nil {
		t.Logf("停了：%v", err)
	}
	t.Logf("跑了 %d 道之後 ds:1C84 ＝ %04X，寫入 %d 次",
		o.Steps(), o.Word(o.DS(0x1C84)), len(*log))
	for i, w := range *log {
		if i >= 20 {
			t.Logf("…還有 %d 筆", len(*log)-i)
			break
		}
		t.Logf("  第 %d 道：ds:%04X %02X→%02X，寫的人在 %s（IDA %#x）",
			w.Step, w.Off, w.Old, w.New, w.IP, o.ToIDA(w.IP))
	}
}

// 探測：BIOS 資料區的視訊模式那一格。
//
// `word_2FB54` 是**視訊記憶體的段**，`_main` 只有兩個寫入點：彩色寫 `0B800h`、
// 單色寫 `0B000h`。執行期量到它從頭到尾是 0——兩條分支都沒走到。
//
// 而 `_main` 在第 306 道就 `mov al, es:[0449]`（`ES = 0`），讀的是 BIOS 資料區的
// `0040:0049`：**目前視訊模式**。那一格若是 0，顯示卡偵測就判不出彩色或單色。
func TestProbeBDAVideoMode(t *testing.T) {
	o := load(t)
	bda := func() uint8 { return o.Bytes(oracle.Addr{Seg: 0x40, Off: 0x49}, 1)[0] }
	t.Logf("冷啟動時 0040:0049 ＝ %#02x（`VideoMode()` 回 %#02x）", bda(), o.VideoMode())
	if err := o.Run(2_000_000); err != nil {
		t.Logf("停了：%v", err)
	}
	t.Logf("跑了 %d 道之後 0040:0049 ＝ %#02x（`VideoMode()` 回 %#02x）",
		o.Steps(), bda(), o.VideoMode())
	t.Logf("視訊記憶體段 `word_2FB54` ＝ %04X（期望 B800 或 B000）",
		o.Word(o.IDA(0x2FB54)))
}

// 探測：`_main` 開頭的顯示卡偵測走了哪一條。
func TestProbeVideoDetect(t *testing.T) {
	o := load(t)
	hits := map[uint32]int{}
	for _, ea := range []uint32{0x10036, 0x1004D, 0x100BB, 0x10410} {
		ea := ea
		o.OnCall(o.IDA(ea), func(*oracle.Oracle) { hits[ea]++ })
	}
	if err := o.Run(2_000_000); err != nil {
		t.Logf("停了：%v", err)
	}
	for _, ea := range []uint32{0x10036, 0x1004D, 0x100BB, 0x10410} {
		t.Logf("IDA %#x 觸發 %d 次", ea, hits[ea])
	}
	t.Logf("`_main` 0x10036 起 0x60 個 byte：%x", o.Bytes(o.IDA(0x10036), 0x60))
	t.Logf("`_main` 0x10096 起 0x40 個 byte：%x", o.Bytes(o.IDA(0x10096), 0x40))
}

// `_main` 只認三種啟動模式：4（CGA 320×200）、9（Tandy）、7（單色）。
//
// dosgolem 冷啟動把 BDA 的 `0040:0049` 填成 3（彩色 80×25 文字，DOS 的預設），
// 三條分支都不中，於是視訊記憶體段 `word_2FB54` 從沒被設，後面每一支繪製常式
// 都拿 `ES = 0` 去寫——中斷向量表就是這樣被 `rep movsw` 蓋掉的。
//
// **KOLTOWN 不是從 DOS 提示字元直接跑的**：它是主程式設好圖形模式之後 EXEC 起來
// 的子程式。補上那個前置狀態，程式就走得下去了。
//
// 模式 4 是 KOL 的那一個（CGA 320×200，視訊記憶體在 B800）。
func TestBootWithGraphicsModeSet(t *testing.T) {
	o := load(t)
	o.WriteU8(oracle.Addr{Seg: 0x40, Off: 0x49}, 4)
	if err := o.Run(40_000_000); err != nil {
		t.Logf("停了：%v", err)
	}
	seg := o.Word(o.IDA(0x2FB54))
	if seg != 0xB800 {
		t.Fatalf("視訊記憶體段是 %04X，期望 B800", seg)
	}
	// 向量表沒被蓋掉——dosgolem 開機時把它們全指到 StubSeg（0x0080）。
	if v := o.Word(oracle.Addr{Seg: 0, Off: 0x22}); v != 0x0080 {
		t.Fatalf("int 08h 的向量變成 %04X——向量表又被寫壞了", v)
	}
	polls, pressed := o.MouseActivity()
	t.Logf("跑了 %d 道，停在 %s（IDA %#x），開了 %d 個檔",
		o.Steps(), o.IP(), o.ToIDA(o.IP()), len(o.Opened()))
	t.Logf("滑鼠輪詢 %d 次（按著 %d 次）——**這一次是真的在等玩家點**", polls, pressed)
	if polls < 1000 {
		t.Fatalf("滑鼠只輪詢 %d 次——它沒有停在等輸入的地方", polls)
	}
}

// 另外兩種啟動模式走得到多遠：探索用，不下斷言。
//
// 模式 9（Tandy）很特別——它**沒有**設視訊段，卻開了最多檔案。那條分支值得再看，
// 但不是現在要走的路：KOL 的圖形是 CGA 320×200。
func TestProbeOtherBootModes(t *testing.T) {
	for _, mode := range []uint8{9, 7} {
		o := load(t)
		o.WriteU8(oracle.Addr{Seg: 0x40, Off: 0x49}, mode)
		if err := o.Run(40_000_000); err != nil {
			t.Logf("模式 %d 停了：%v", mode, err)
		}
		polls, _ := o.MouseActivity()
		t.Logf("模式 %d：視訊段 %04X、停在 IDA %#x、開了 %d 個檔、滑鼠輪詢 %d",
			mode, o.Word(o.IDA(0x2FB54)), o.ToIDA(o.IP()), len(o.Opened()), polls)
	}
}

// 畫面到底畫了沒——CGA 模式 4 的視訊記憶體在 `B8000`，掃描線交錯。
//
// `Indexed()` 讀的是 mode 13h 的 `A0000`，所以 CGA 程式在它底下永遠是全黑。
// 這一條用 `CGA4()` 問真正的那一塊：畫面有東西，才談得上「該點哪裡」。
func TestBootDrawsSomething(t *testing.T) {
	o := load(t)
	o.WriteU8(oracle.Addr{Seg: 0x40, Off: 0x49}, 4)
	if err := o.Run(40_000_000); err != nil {
		t.Logf("停了：%v", err)
	}
	shot := o.CGA4()
	hist := map[uint8]int{}
	for _, v := range shot {
		hist[v]++
	}
	t.Logf("CGA 畫面色號分布：%v（共 %d 個像素）", hist, len(shot))
	if hist[0] == len(shot) {
		t.Fatal("整張畫面都是色號 0——`B8000` 那一塊什麼都沒有")
	}
	// 粗略看一下輪廓：每 8 列取一列，每 4 欄取一格。
	for y := 0; y < 200; y += 8 {
		line := make([]byte, 0, 80)
		for x := 0; x < 320; x += 4 {
			line = append(line, " .:#"[shot[y*320+x]])
		}
		t.Logf("%3d |%s|", y, line)
	}
}

// 掃描：哪一種啟動模式讓原版走得最遠。
//
// `_main` 認的模式清單是從反組譯讀出來的（4／9／7），但那只是我解碼到的部分。
// 與其猜，不如把 `0040:0049` 的每一個值都試一遍——判準有四個，**一起看才準**：
// 視訊段被設了沒、向量表完好沒、開了幾個檔、畫面畫了幾個像素。
//
// 只看其中一個會誤判：模式 4 設了視訊段卻只開 4 個檔，模式 9 開了 19 個檔卻沒設
// 視訊段。**走得遠**與**走得對**是兩件事。
func TestProbeAllBootModes(t *testing.T) {
	for mode := 0; mode <= 0x13; mode++ {
		o := load(t)
		o.WriteU8(oracle.Addr{Seg: 0x40, Off: 0x49}, uint8(mode))
		if err := o.Run(20_000_000); err != nil {
			t.Logf("模式 %02X 停了：%v", mode, err)
		}
		nz := 0
		for _, v := range o.CGA4() {
			if v != 0 {
				nz++
			}
		}
		vec := o.Word(oracle.Addr{Seg: 0, Off: 0x22})
		t.Logf("模式 %02X：視訊段 %04X、向量 int08 %04X%s、開 %2d 個檔、"+
			"CGA 非零 %5d、停在 IDA %#x",
			mode, o.Word(o.IDA(0x2FB54)), vec,
			map[bool]string{true: "", false: "（壞了）"}[vec == 0x0080],
			len(o.Opened()), nz, o.ToIDA(o.IP()))
	}
}

// 模式 09 與 0D 是唯二走得完載入的（19 個檔），其餘五個檔就 `AH=4Ch` 離開。
// 這一條分辨它們：09 是 Tandy（視訊記憶體 B800），0D 是 EGA 320×200（A0000）。
//
// **KOL 是 EGA 遊戲**（`assets/original-derived-png/ega/`），所以 0D 才是那一條。
// 判準是兩塊記憶體一起看——只讀一塊會把「畫在另一塊」誤判成「沒畫」。
func TestProbeTandyVsEGA(t *testing.T) {
	for _, mode := range []uint8{9, 0x0D} {
		o := load(t)
		o.WriteU8(oracle.Addr{Seg: 0x40, Off: 0x49}, mode)
		if err := o.Run(40_000_000); err != nil {
			t.Logf("模式 %02X 停了：%v", mode, err)
		}
		count := func(b []uint8) int {
			n := 0
			for _, v := range b {
				if v != 0 {
					n++
				}
			}
			return n
		}
		polls, pressed := o.MouseActivity()
		t.Logf("模式 %02X：B8000 非零 %d、A0000 非零 %d、滑鼠輪詢 %d（按著 %d）、"+
			"停在 IDA %#x、開了 %v",
			mode, count(o.CGA4()), count(o.Indexed()), polls, pressed,
			o.ToIDA(o.IP()), o.Opened())
	}
}

// 原版載入完成之後畫的是什麼。
//
// Tandy 模式 09h：四個 bit 一個像素、一列 160 bytes、**四段交錯**。解錯段數的話
// 畫面會變成四張交疊的梳子——看起來像雜訊，不像「解錯了」，所以下面同時印出
// 色號分布與輪廓：**分布集中在少數幾個色號**才是真的畫面，均勻散開就是解錯了。
func TestTownScreenDecodes(t *testing.T) {
	o := load(t)
	o.WriteU8(oracle.Addr{Seg: 0x40, Off: 0x49}, 9)
	if err := o.Run(40_000_000); err != nil {
		t.Logf("停了：%v", err)
	}
	shot := o.Tandy16()
	hist := map[uint8]int{}
	for _, v := range shot {
		hist[v]++
	}
	t.Logf("色號分布：%v", hist)
	if len(hist) == 1 {
		t.Fatal("整張畫面同一個色號——什麼都沒畫")
	}
	for y := 0; y < 200; y += 5 {
		line := make([]byte, 0, 80)
		for x := 0; x < 320; x += 4 {
			line = append(line, " .:-=+*#%@$&XW#M"[shot[y*320+x]])
		}
		t.Logf("%3d |%s|", y, line)
	}
}

// cgaPalette 是 CGA／EGA／Tandy 共用的十六色盤。
var cgaPalette = [16][3]uint8{
	{0x00, 0x00, 0x00}, {0x00, 0x00, 0xAA}, {0x00, 0xAA, 0x00}, {0x00, 0xAA, 0xAA},
	{0xAA, 0x00, 0x00}, {0xAA, 0x00, 0xAA}, {0xAA, 0x55, 0x00}, {0xAA, 0xAA, 0xAA},
	{0x55, 0x55, 0x55}, {0x55, 0x55, 0xFF}, {0x55, 0xFF, 0x55}, {0x55, 0xFF, 0xFF},
	{0xFF, 0x55, 0x55}, {0xFF, 0x55, 0xFF}, {0xFF, 0xFF, 0x55}, {0xFF, 0xFF, 0xFF},
}

// 把原版的畫面存成 PNG 給人眼看。
//
// **ASCII 輪廓與色號分布都證明不了解碼參數對**——段數記錯時讀的是同一批 bytes，
// 統計幾乎不變，變的是像素的位置。只有人眼看得出「這是一張圖」還是「四張交疊的
// 梳子」。輸出留在 `workplace/`（本機，不進版控）。
func TestDumpTownScreenPNG(t *testing.T) {
	out := os.Getenv("KOL_ORACLE_SHOT")
	if out == "" {
		t.Skip("設 KOL_ORACLE_SHOT=<輸出路徑> 才存圖")
	}
	o := load(t)
	o.WriteU8(oracle.Addr{Seg: 0x40, Off: 0x49}, 9)
	// 畫面在前兩億道指令內還會長：24,480 → 49,198 個非零像素之後才穩定。
	if err := o.Run(200_000_000); err != nil {
		t.Logf("停了：%v", err)
	}
	shot := o.Tandy16()
	img := image.NewRGBA(image.Rect(0, 0, 320, 200))
	for y := 0; y < 200; y++ {
		for x := 0; x < 320; x++ {
			c := cgaPalette[shot[y*320+x]&0x0F]
			img.Set(x, y, color.RGBA{c[0], c[1], c[2], 0xFF})
		}
	}
	f, err := os.Create(out)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	t.Logf("寫出 %s", out)
}

// 探測：那條鏈是不是繞成了環。
//
// 迴圈的終止條件是「讀到 -1」。指標有效、終止值也在，卻走不完——最可能的解釋是
// **鏈結繞回去了**。判準是取樣 `[bp-14]`（目前走到的節點）：值週期性重複就是環。
func TestProbeChainIsCircular(t *testing.T) {
	o := load(t)
	o.WriteU8(oracle.Addr{Seg: 0x40, Off: 0x49}, 9)
	if err := o.Run(40_000_000); err != nil {
		t.Logf("停了：%v", err)
	}
	seen := map[uint16]int{}
	var order []uint16
	for i := 0; i < 400; i++ {
		r := o.Regs()
		node := o.Word(oracle.Addr{Seg: r.SS, Off: r.BP - 0x14})
		idx := o.Word(oracle.Addr{Seg: r.SS, Off: r.BP - 0x1C})
		if seen[node] == 0 && len(order) < 24 {
			order = append(order, node)
		}
		seen[node]++
		if i == 0 || i == 399 {
			t.Logf("取樣 %d：節點 %04X、索引 %04X", i, node, idx)
		}
		if err := o.Run(5_000); err != nil {
			t.Logf("往前走停了：%v", err)
			break
		}
	}
	t.Logf("四百次取樣看到 %d 個不同的節點值", len(seen))
	t.Logf("前幾個：%04X", order)
	if len(seen) < 400 {
		t.Log("節點值重複——這條鏈繞成了環，永遠走不到 -1")
	}
}

// 探測：第六次解壓之前，程式對 `TOWN7.0` 做了什麼 seek 與 read。
//
// 讀進來的資料**逐 byte 等於檔案的位置 0**（比對過），也就是 26 個 entry 的
// offset 表，不是壓縮資料。`seek` 的實作看過是對的，所以問參數本身：
// 程式到底 seek 到哪、讀了多少。
func TestProbeTownFileOps(t *testing.T) {
	o := load(t)
	o.WriteU8(oracle.Addr{Seg: 0x40, Off: 0x49}, 9)
	if err := o.Run(2_500_000); err != nil {
		t.Logf("停了：%v", err)
	}
	ops := o.FileOps()
	t.Logf("共 %d 次檔案操作", len(ops))
	from := 0
	if len(ops) > 24 {
		from = len(ops) - 24
		t.Logf("（前面 %d 次略過）", from)
	}
	for _, op := range ops[from:] {
		switch op.Kind {
		case "seek":
			t.Logf("  第 %8d 道  seek %-12s handle %d  whence %d  位移 %d → %d",
				op.Step, op.Name, op.Handle, op.Whence, op.Offset, op.Got)
		default:
			t.Logf("  第 %8d 道  read %-12s handle %d  從 %d 要 %d 得 %d → %05X",
				op.Step, op.Name, op.Handle, op.Offset, op.Want, op.Got, op.Into)
		}
	}
}

// 探測：從主程式 `KOL.EXE` 開始跑。
//
// 遊戲目錄裡有 `KOL.EXE`（11,971 byte 的未壓縮 MS C 程式）、**`MAKEMODE.EXE`**
// （名字就是「設模式」）、`TANDYPIN.EXE`、`TITLE.EXE`、`KOLBOOK.EXE`、`KOLTOWN.EXE`。
//
// `KOL.EXE` 是主程式：設好視訊模式，再用 `int 21h AH=4Bh AL=03h`（載入 overlay）
// 把 `TITLE.EXE`／`KOLBOOK.EXE`／`KOLTOWN.EXE` 搬進自己配好的記憶體、far call
// 進入點。**不是 EXEC 子程式**——直接載 KOLTOWN 少的就是主程式設好的那些前置。
func TestProbeMainExecutable(t *testing.T) {
	const main = rootPath + "/KOL.EXE"
	if _, err := os.Stat(main); err != nil {
		t.Skipf("沒有 %s：%v", main, err)
	}
	o, err := oracle.Load(main, rootPath)
	if err != nil {
		t.Fatalf("載入主程式：%v", err)
	}
	t.Cleanup(o.Close)

	if err := o.Run(40_000_000); err != nil {
		t.Logf("停了：%v", err)
	}
	t.Logf("跑了 %d 道，停在 %s", o.Steps(), o.IP())
	t.Logf("視訊模式 %#02x、BDA 0040:0049 ＝ %#02x",
		o.VideoMode(), o.Bytes(oracle.Addr{Seg: 0x40, Off: 0x49}, 1)[0])
	t.Logf("開過的檔：%v", o.Opened())
	t.Logf("開不起來的：%v", o.Missing())
	if c := o.Console(); c != "" {
		t.Logf("主控台：%q", c)
	}
	if u := o.Unimplemented(); len(u) > 0 {
		t.Logf("沒實作的服務：%v", u)
	}
	for _, op := range o.FileOps() {
		t.Logf("  %s %s 從 %d 要 %d 得 %d", op.Kind, op.Name, op.Offset, op.Want, op.Got)
	}
}

// chainRoot 是啟動鏈用的遊戲目錄：與 `rootPath` 同一批檔（硬連結），只有還跑不動
// 的那幾支換成「立刻結束」的樁。
//
// **目的是讓 KOLTOWN 拿到正確的前置狀態**，不是重現標題動畫。`KOL.EXE` 會把視訊
// 模式設成 0Dh 再一支一支 EXEC 下去；中間哪一支跑飛，後面就都拿不到狀態。
// 樁掉跑不動的那幾支，鏈才走得到 KOLTOWN。
const chainRoot = "../../workplace/oracle/game-chain"

// 探測：樁掉 TITLE 之後，啟動鏈走到哪。
func TestProbeBootChain(t *testing.T) {
	main := chainRoot + "/KOL.EXE"
	if _, err := os.Stat(main); err != nil {
		t.Skipf("沒有 %s（跑 `tools/` 那支建目錄的腳本）：%v", main, err)
	}
	o, err := oracle.Load(main, chainRoot)
	if err != nil {
		t.Fatalf("載入主程式：%v", err)
	}
	t.Cleanup(o.Close)
	if err := o.Run(60_000_000); err != nil {
		t.Logf("停了：%v", err)
	}
	t.Logf("跑了 %d 道，停在 %s；視訊模式 %#02x",
		o.Steps(), o.IP(), o.VideoMode())
	t.Logf("開過的檔：%v", o.Opened())
	t.Logf("開不起來的：%v", o.Missing())
	nz := 0
	for _, v := range o.Indexed() {
		if v != 0 {
			nz++
		}
	}
	t.Logf("A0000 非零 %d、B8000 非零 %d", nz, func() int {
		n := 0
		for _, v := range o.Tandy16() {
			if v != 0 {
				n++
			}
		}
		return n
	}())
	if c := o.Console(); c != "" {
		t.Logf("主控台：%q", c)
	}

	// 卡在 KOL.EXE 自己的碼裡而沒有繼續 EXEC——多半在等按鍵。逐批送，
	// 看 `Opened` 有沒有長出下一支執行檔。
	for round := 0; round < 8; round++ {
		before := len(o.Opened())
		o.Type(" \r")
		if err := o.Run(20_000_000); err != nil {
			t.Logf("第 %d 批之後停了：%v", round, err)
			break
		}
		if len(o.Opened()) != before {
			t.Logf("第 %d 批之後開過的檔變成：%v", round, o.Opened())
		}
	}
	t.Logf("送完鍵：停在 %s、視訊模式 %#02x、開過 %v",
		o.IP(), o.VideoMode(), o.Opened())
}

// 探測：記憶體配置器把哪些段配出去。
//
// 卡住那一段是「把所有記憶體配光」的慣用法（`48h` 要 1 段 →`4Ah` 撐到最大 →
// 把上一塊的段寫進這塊的開頭串成鏈）。它寫壞了自己的碼，表示 `48h` 回的段
// **落在映像裡**。
func TestProbeMemOps(t *testing.T) {
	main := chainRoot + "/KOL.EXE"
	if _, err := os.Stat(main); err != nil {
		t.Skipf("沒有 %s：%v", main, err)
	}
	o, err := oracle.Load(main, chainRoot)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(o.Close)
	if err := o.Run(40_000_000); err != nil {
		t.Logf("停了：%v", err)
	}
	ops := o.MemOps()
	t.Logf("共 %d 筆記憶體服務", len(ops))
	for i, op := range ops {
		if i >= 24 {
			t.Logf("…還有 %d 筆", len(ops)-i)
			break
		}
		t.Logf("  第 %d 道 %-8s 段=%04X 要=%04X 得=%04X 可用=%04X free=%04X 失敗=%v",
			op.Step, op.Kind, op.Block, op.Want, op.Got, op.Avail, op.FreeSeg, op.Failed)
	}
}

// 探測：片頭在播，接下來呢。overlay 通了之後 `TITLE.EXE` 會一路讀 `BACKPICS.0`，
// 所以看「開過哪些檔」的變化就知道階段有沒有推進。
func TestProbeTitleProgress(t *testing.T) {
	main := chainRoot + "/KOL.EXE"
	if _, err := os.Stat(main); err != nil {
		t.Skipf("沒有 %s：%v", main, err)
	}
	o, err := oracle.Load(main, chainRoot)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(o.Close)
	seen := 0
	for round := 0; round < 8; round++ {
		err := o.Run(50_000_000)
		files := o.Opened()
		nonzero := 0
		for _, v := range o.Indexed() {
			if v != 0 {
				nonzero++
			}
		}
		t.Logf("第 %d 輪：停在 %s、模式 %02Xh、A0000 非零 %d、開過 %d 個檔",
			round, o.IP(), o.VideoMode(), nonzero, len(files))
		if len(files) > seen {
			t.Logf("    新開：%v", files[seen:])
			seen = len(files)
		}
		if err != nil {
			t.Logf("    停了：%v", err)
			break
		}
	}
}

// 探測：城鎮那條線的記憶體配置。前綴表落在 `2EFF:0000`，而 `CHARDATA` 也讀進
// `2EFF0`——同一塊被用了兩次。誰配的、配了幾次。
func TestProbeTownMemOps(t *testing.T) {
	o := load(t)
	o.WriteU8(oracle.Addr{Seg: 0x40, Off: 0x49}, 9)
	if err := o.Run(3_000_000); err != nil {
		t.Logf("停了：%v", err)
	}
	ops := o.MemOps()
	t.Logf("共 %d 筆記憶體服務", len(ops))
	for i, op := range ops {
		if i >= 30 {
			t.Logf("…還有 %d 筆", len(ops)-i)
			break
		}
		t.Logf("  第 %d 道 %-8s 段=%04X 要=%04X 得=%04X 可用=%04X free=%04X 失敗=%v",
			op.Step, op.Kind, op.Block, op.Want, op.Got, op.Avail, op.FreeSeg, op.Failed)
	}
}

// 探測：`TOWN7.0` 前後完整的檔案軌跡（含開、關與失敗的呼叫）。
func TestProbeTownFullTrace(t *testing.T) {
	o := load(t)
	o.WriteU8(oracle.Addr{Seg: 0x40, Off: 0x49}, 9)
	if err := o.Run(3_000_000); err != nil {
		t.Logf("停了：%v", err)
	}
	ops := o.FileOps()
	t.Logf("共 %d 次檔案操作", len(ops))
	for _, op := range ops {
		mark := ""
		if op.Failed {
			mark = "  ← 失敗"
		}
		t.Logf("  第 %8d 道 %-5s %-12s handle %2d whence %d 位移 %6d 要 %5d 得 %5d%s",
			op.Step, op.Kind, op.Name, op.Handle, op.Whence, op.Offset, op.Want, op.Got, mark)
	}
}

// 探測：城鎮載完之後怎麼推進。原版停在 `sub_1B100` 輪詢 `int 33h`，
// 畫面是建築物內的對話框。逐個鍵與幾個點擊位置試，看哪一個讓畫面動。
func TestProbeTownAdvance(t *testing.T) {
	o := load(t)
	o.WriteU8(oracle.Addr{Seg: 0x40, Off: 0x49}, 9)
	if err := o.Run(200_000_000); err != nil {
		t.Logf("停了：%v", err)
	}
	hash := func() string {
		sum := sha256.Sum256(o.Tandy16())
		return hex.EncodeToString(sum[:8])
	}
	base := hash()
	t.Logf("基準畫面 %s、停在 IDA %#x、開了 %d 個檔", base, o.ToIDA(o.IP()), len(o.Opened()))

	for _, k := range []struct{ name, send string }{
		{"Enter", "\r"}, {"Esc", "\x1b"}, {"空白", " "},
		{"L", "L"}, {"T", "T"}, {"X", "X"}, {"Q", "Q"},
	} {
		o.ClearInput()
		o.Type(k.send)
		if err := o.Run(20_000_000); err != nil {
			t.Logf("  鍵 %-4s：停了 %v", k.name, err)
			break
		}
		if h := hash(); h != base {
			t.Logf("  鍵 %-4s：畫面變了 %s → %s、開了 %d 個檔", k.name, base, h, len(o.Opened()))
			base = h
		} else {
			t.Logf("  鍵 %-4s：沒反應", k.name)
		}
	}

	for _, p := range []struct {
		name string
		x, y int
	}{
		{"文字框", 200, 110}, {"頭像", 40, 100}, {"隊員圖示", 20, 20},
		{"下方空白", 160, 190}, {"畫面中央", 160, 100},
	} {
		if err := o.Click(p.x, p.y, oracle.NoCursorWait(),
			oracle.Frame((*oracle.Oracle).Tandy16)); err != nil {
			t.Logf("  點 %-8s (%3d,%3d)：%v", p.name, p.x, p.y, err)
			continue
		}
		h := hash()
		t.Logf("  點 %-8s (%3d,%3d)：畫面 %s%s、開了 %d 個檔", p.name, p.x, p.y, h,
			map[bool]string{true: "（變了）", false: "（沒變）"}[h != base], len(o.Opened()))
		base = h
	}
}

// savePNG 把目前的 Tandy 畫面存成 PNG（探測用）。
func savePNG(t *testing.T, o *oracle.Oracle, path string) {
	t.Helper()
	shot := o.Tandy16()
	img := image.NewRGBA(image.Rect(0, 0, 320, 200))
	for y := 0; y < 200; y++ {
		for x := 0; x < 320; x++ {
			c := cgaPalette[shot[y*320+x]&0x0F]
			img.Set(x, y, color.RGBA{c[0], c[1], c[2], 0xFF})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	t.Logf("寫出 %s", path)
}

// 探測：在城鎮畫面上點幾個位置，各存一張圖，人眼判斷哪一個推得動。
func TestDumpTownClicks(t *testing.T) {
	dir := os.Getenv("KOL_ORACLE_DUMP")
	if dir == "" {
		t.Skip("沒設 KOL_ORACLE_DUMP")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	o := load(t)
	o.WriteU8(oracle.Addr{Seg: 0x40, Off: 0x49}, 9)
	if err := o.Run(200_000_000); err != nil {
		t.Logf("停了：%v", err)
	}
	savePNG(t, o, dir+"/00-base.png")
	for i, p := range []struct {
		name string
		x, y int
	}{
		{"文字框", 200, 110}, {"頭像", 40, 100}, {"隊員圖示", 20, 20},
		{"下方空白", 160, 190}, {"右下角", 300, 190}, {"左下手形", 20, 190},
	} {
		err := o.Click(p.x, p.y, oracle.NoCursorWait(), oracle.Frame((*oracle.Oracle).Tandy16))
		savePNG(t, o, fmt.Sprintf("%s/%02d-%s.png", dir, i+1, p.name))
		t.Logf("  點 %s (%d,%d)：%v、IDA %#x", p.name, p.x, p.y, err, o.ToIDA(o.IP()))
	}
}

// 探測：建築物內那六個圖示各做什麼。點一個、存一張圖、再點文字框回到對話。
func TestDumpTownIcons(t *testing.T) {
	dir := os.Getenv("KOL_ORACLE_DUMP")
	if dir == "" {
		t.Skip("沒設 KOL_ORACLE_DUMP")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	o := load(t)
	o.WriteU8(oracle.Addr{Seg: 0x40, Off: 0x49}, 9)
	if err := o.Run(200_000_000); err != nil {
		t.Logf("停了：%v", err)
	}
	click := func(x, y int) error {
		return o.Click(x, y, oracle.NoCursorWait(), oracle.Frame((*oracle.Oracle).Tandy16))
	}
	// 先點文字框讓指令列長出來。
	if err := click(200, 110); err != nil {
		t.Fatalf("點文字框：%v", err)
	}
	savePNG(t, o, dir+"/icons-00-bar.png")
	for i := 0; i < 6; i++ {
		x := 47 + i*29
		err := click(x, 186)
		savePNG(t, o, fmt.Sprintf("%s/icons-%02d-x%d.png", dir, i+1, x))
		t.Logf("  點圖示 %d（x=%d）：%v、IDA %#x、開了 %d 個檔",
			i, x, err, o.ToIDA(o.IP()), len(o.Opened()))
	}
}

// townMap 把原版推到城鎮地圖：等載入完成 → 點文字框叫出指令列 → 點最右邊那個
// 圖示（離開建築）。
func townMap(t *testing.T) *oracle.Oracle {
	t.Helper()
	o := load(t)
	o.WriteU8(oracle.Addr{Seg: 0x40, Off: 0x49}, 9)
	if err := o.Run(200_000_000); err != nil {
		t.Fatalf("載入城鎮：%v", err)
	}
	click := func(x, y int) {
		t.Helper()
		if err := o.Click(x, y, oracle.NoCursorWait(),
			oracle.Frame((*oracle.Oracle).Tandy16)); err != nil {
			t.Fatalf("點 (%d,%d)：%v", x, y, err)
		}
	}
	click(200, 110) // 文字框：叫出指令列
	click(47, 186)  // 最左邊那個圖示（U-TURN）：離開建築
	return o
}

// 探測：城鎮地圖上怎麼移動。角色在地圖中央偏上，逐一試鍵盤與鄰格點擊。
func TestDumpTownMove(t *testing.T) {
	dir := os.Getenv("KOL_ORACLE_DUMP")
	if dir == "" {
		t.Skip("沒設 KOL_ORACLE_DUMP")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	o := townMap(t)
	savePNG(t, o, dir+"/move-00-map.png")
	hash := func() string {
		sum := sha256.Sum256(o.Tandy16())
		return hex.EncodeToString(sum[:8])
	}
	base := hash()
	t.Logf("地圖基準 %s", base)

	// 鍵盤：方向鍵是「00 ＋ 掃描碼」，數字鍵盤是 ASCII。
	for _, k := range []struct{ name, send string }{
		{"↑掃描", "\x00\x48"}, {"↓掃描", "\x00\x50"},
		{"←掃描", "\x00\x4B"}, {"→掃描", "\x00\x4D"},
		{"8", "8"}, {"2", "2"}, {"4", "4"}, {"6", "6"},
	} {
		o.ClearInput()
		o.Type(k.send)
		if err := o.Run(10_000_000); err != nil {
			t.Logf("  鍵 %s：停了 %v", k.name, err)
			break
		}
		h := hash()
		t.Logf("  鍵 %-6s：%s%s", k.name, h,
			map[bool]string{true: "（變了）", false: ""}[h != base])
		base = h
	}

	// 滑鼠：角色在 (195,95) 附近，試四個鄰格與遠一點的位置。
	for i, p := range []struct {
		name string
		x, y int
	}{
		{"右一格", 211, 95}, {"左一格", 179, 95},
		{"上一格", 195, 79}, {"下一格", 195, 111},
		{"右邊界", 295, 95}, {"下邊界", 195, 150},
	} {
		err := o.Click(p.x, p.y, oracle.NoCursorWait(), oracle.Frame((*oracle.Oracle).Tandy16))
		savePNG(t, o, fmt.Sprintf("%s/move-%02d-%s.png", dir, i+1, p.name))
		h := hash()
		t.Logf("  點 %-6s (%3d,%3d)：%v %s", p.name, p.x, p.y, err, h)
		base = h
	}
}

// 探測：**原版每走一步抽幾次骰**。
//
// 這是 `rand_test.go` 那句「這裡不驗呼叫時機」留下的洞：對拍證實的是 `rand()`
// 與 `rand(n)` 的實作逐次相同，誰在什麼時候抽、抽幾次一直沒驗過。收據基準
// 要重生就得先知道這件事——remake 每步抽的次數若與原版不同，重生出來的基準
// 還是要再做一次。
//
// ⚠ **先做正對照。**「一次都沒抽」與「攔截器根本沒掛上」長得一模一樣
// （`~/diagnosis-notes/docs/02-query-returned-empty/`）。
func TestProbeRandPerStep(t *testing.T) {
	o := townMap(t)
	type draw struct {
		caller uint32
		n      uint16
		roll   bool
	}
	var draws []draw
	o.OnCall(o.IDA(rollEnter), func(p *oracle.Oracle) {
		draws = append(draws, draw{caller: p.ToIDA(p.Caller()), n: p.Arg(0), roll: true})
	})
	o.OnCall(o.IDA(randEnter), func(p *oracle.Oracle) {
		draws = append(draws, draw{caller: p.ToIDA(p.Caller())})
	})

	// 正對照：直接呼叫一次，攔截器沒掛上的話這裡就會是 0。
	if _, err := o.Call(o.IDA(rollEnter), 6); err != nil {
		t.Fatalf("正對照呼叫 rand(6)：%v", err)
	}
	if len(draws) == 0 {
		t.Fatal("正對照：直接呼叫 rand(6) 也沒被攔到——攔截器沒掛上，後面的 0 不算數")
	}
	t.Logf("正對照過了：直接呼叫記到 %d 次", len(draws))

	report := func(label string, from int) {
		t.Logf("%s 抽了 %d 次：", label, len(draws)-from)
		counts := map[string]int{}
		for _, d := range draws[from:] {
			kind := "rand()"
			if d.roll {
				kind = fmt.Sprintf("rand(%d)", d.n)
			}
			counts[fmt.Sprintf("%s ← IDA %#x", kind, d.caller)]++
		}
		keys := make([]string, 0, len(counts))
		for k := range counts {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			t.Logf("    %-34s ×%d", k, counts[k])
		}
	}

	// 往同一個方向走到出城。城鎮不擲骰，野外才有遭遇判定——換場景會讀新的檔，
	// 那是最可靠的路標。
	seen := len(o.Opened())
	for i := 1; i <= 40; i++ {
		before := len(draws)
		err := o.Click(295, 95, oracle.NoCursorWait(),
			oracle.Frame((*oracle.Oracle).Tandy16))
		files := ""
		if len(o.Opened()) > seen {
			files = fmt.Sprintf("　新開 %v", o.Opened()[seen:])
			seen = len(o.Opened())
		}
		if err != nil {
			t.Logf("第 %2d 步：畫面沒變（撞牆或不能走）%s", i, files)
			continue
		}
		if n := len(draws) - before; n > 0 || files != "" {
			report(fmt.Sprintf("第 %2d 步%s", i, files), before)
		}
	}
	t.Logf("走完 40 步共抽 %d 次；開了 %d 個檔", len(draws), len(o.Opened()))
	if dir := os.Getenv("KOL_ORACLE_DUMP"); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err == nil {
			savePNG(t, o, dir+"/walk-40.png")
		}
	}
}
