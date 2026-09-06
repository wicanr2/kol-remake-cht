//go:build oracle

package oracle

import (
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
)

// 正對照：`OnCall` 真的會觸發嗎。
//
// 亂數那一支掛上去抽到 0 次的時候，「原版沒有抽」與「hook 根本沒掛上」長得一模一樣
// （`~/diagnosis-notes/docs/02-query-returned-empty`）。`_main` 已知在第 263 道被走到，
// 拿它當正對照：這一條不過就表示問題在 hook，不在原版。
func TestOnCallFires(t *testing.T) {
	o := load(t)
	var mainHits, srandHits, randHits, rollHits int
	o.OnCall(o.IDA(0x10000), func(*oracle.Oracle) { mainHits++ })
	o.OnCall(o.IDA(0x2D526), func(*oracle.Oracle) { srandHits++ })
	o.OnCall(o.IDA(randEnter), func(*oracle.Oracle) { randHits++ })
	o.OnCall(o.IDA(rollEnter), func(*oracle.Oracle) { rollHits++ })

	if err := o.Run(80_000_000); err != nil {
		t.Logf("跑到一半停了：%v", err)
	}
	t.Logf("_main %d 次、srand %d 次、rand %d 次、rand(n) %d 次；跑了 %d 道，停在 %s",
		mainHits, srandHits, randHits, rollHits, o.Steps(), o.IP())
	t.Logf("開過的檔：%v", o.Opened())
	if c := o.Console(); c != "" {
		t.Logf("主控台：%q", c)
	}
	if mainHits == 0 {
		t.Fatal("`_main` 一次都沒觸發——問題在 hook，不在原版")
	}
}
