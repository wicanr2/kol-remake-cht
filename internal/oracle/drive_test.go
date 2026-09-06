//go:build oracle

package oracle

import (
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
)

// 原版在開機到等輸入這一段**一次亂數都不抽**。
//
// 這是 `rand_test.go` 為什麼改成直接呼叫的理由，記在這裡當事實：載入資料、
// 畫標題、進到等鍵盤那一刻，`rand` 與 `rand(n)` 的呼叫次數都是 0。所以
// 「掛上 hook 抽到 0 次」在這一段是正確結果，不是 hook 沒掛上——後者由
// `TestOnCallFires` 用 `_main` 與 `srand` 排除。
//
// 想知道**遊戲什麼時候才抽**，要能驅動到戰鬥或世界地圖；那需要一份完整的
// 操作腳本，不在這一條的範圍內。
func TestBootDrawsNothing(t *testing.T) {
	o := load(t)
	var rands, rolls int
	o.OnCall(o.IDA(randEnter), func(*oracle.Oracle) { rands++ })
	o.OnCall(o.IDA(rollEnter), func(*oracle.Oracle) { rolls++ })

	if err := o.Run(20_000_000); err != nil {
		t.Logf("開機階段停了：%v", err)
	}
	t.Logf("開機跑完 %d 道，停在 %s；rand %d 次、rand(n) %d 次",
		o.Steps(), o.IP(), rands, rolls)
	if rands != 0 || rolls != 0 {
		t.Fatalf("開機階段抽了亂數（rand %d、rand(n) %d）——"+
			"那表示對拍可以從自然流程接，這一條的前提要重寫", rands, rolls)
	}
}
