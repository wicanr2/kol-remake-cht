//go:build oracle

package oracle

import (
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
