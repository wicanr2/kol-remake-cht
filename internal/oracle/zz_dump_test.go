//go:build oracle

package oracle

import "testing"

// 臨時探針：倒出敵人 AI 用到的幾張小表（RE-168）。
func TestZZDumpAITables(t *testing.T) {
	o := prepare(t)
	dump := func(name string, off uint16, n int) {
		t.Logf("%s DS:%#x = % x", name, off, o.Bytes(o.DS(off), n))
	}
	dump("方向表 0x10D8", 0x10D8, 12)
	dump("環格數 0xFC0", 0xFC0, 8)
	dump("步行位元圖 0xFC8", 0xFC8, 16)
	dump("位元遮罩 0x1312", 0x1312, 8)
}
