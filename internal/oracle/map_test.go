//go:build oracle

package oracle

import (
	"encoding/hex"
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
)

// dosgolem 的 `IDA()` 假設「IDA 線性 0x10000 ＝ 映像的第一個 byte」。那對 rich2
// 成立，對 KOLTOWN 要自己驗——**錯了不會報錯**，只會讀到看起來合理的別的東西。
//
// 判準是拿 IDA 探針記下來的 bytes 去比：對得上才算映射正確。
func TestIDAMappingMatchesProbeBytes(t *testing.T) {
	o := load(t)
	for _, c := range []struct {
		name  string
		ea    uint32
		bytes string
	}{
		// 進入點（`KOLTOWN-unpacked-entry-private.json`）。
		{"entry", 0x2B90C, "b430cd213c02"},
		// `_main` 開頭（RE-132 的探針）。
		{"_main", 0x10000, "558bec81ec02"},
		// `rand()` 的頭兩道：`mov ax, 43FDh` ＋ `mov dx, 3`。
		{"rand", 0x2D538, "b8fd43ba0300"},
		// `srand()`：`push bp` ＋ `mov bp, sp` ＋ `mov ax, [bp+arg_0]`。
		{"srand", 0x2D526, "558bec8b4606"},
	} {
		want, err := hex.DecodeString(c.bytes)
		if err != nil {
			t.Fatalf("%s：測試資料壞了：%v", c.name, err)
		}
		got := o.Bytes(o.IDA(c.ea), len(want))
		if hex.EncodeToString(got) != c.bytes {
			t.Errorf("%s（IDA %#x → 執行期 %s）：讀到 %s，期望 %s",
				c.name, c.ea, o.IDA(c.ea), hex.EncodeToString(got), c.bytes)
		}
	}
	t.Logf("進入點在 %s，DGROUP 0x2DED0 在 %s",
		o.IDA(0x2B90C), o.IDA(0x2DED0))
}

// 走得到 `_main` 嗎——走不到的話上面那個映射就算對，`At()` 也沒有意義。
func TestReachesMain(t *testing.T) {
	o := load(t)
	if err := o.RunUntil(oracle.At(o.IDA(0x10000)), oracle.Budget(50_000_000)); err != nil {
		t.Fatalf("跑不到 _main：%v", err)
	}
	t.Logf("第 %d 道指令走到 _main", o.Steps())
}
