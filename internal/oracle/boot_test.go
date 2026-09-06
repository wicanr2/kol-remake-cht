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

	// 在等什麼？輪詢的埠就是答案：`3DA` 是 CRT 狀態（等垂直回掃）、
	// `60`／`64` 是鍵盤、`40`–`43` 是計時器、`388` 是 OPL。
	t.Logf("讀過的 I/O 埠（次數多的在前）：%v", o.PortReads())

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
