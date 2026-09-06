//go:build oracle

package oracle

import (
	"testing"

	"github.com/wicanr2/dosgolem/oracle"
)

// 原版載完資料就停在等輸入，一次亂數都還沒抽。這一條找出「要送幾個鍵才會開始抽」
// ——`rand` 的第一個呼叫端是誰，決定了對拍要從哪裡接。
func TestFindFirstDraw(t *testing.T) {
	o := load(t)
	var draws int
	callers := map[string]int{}
	o.OnCall(o.IDA(randEnter), func(p *oracle.Oracle) {
		draws++
		callers[p.Caller().String()]++
	})

	// 先跑到它開始等輸入。
	if err := o.Run(20_000_000); err != nil {
		t.Logf("開機階段停了：%v", err)
	}
	t.Logf("開機跑完 %d 道，抽了 %d 次，停在 %s", o.Steps(), draws, o.IP())

	// 一批一批送鍵，看第幾批開始抽。空白與 Enter 是原版每一頁的推進鍵。
	for round := 0; round < 12 && draws == 0; round++ {
		o.Type(" \r")
		if err := o.Run(20_000_000); err != nil {
			t.Logf("第 %d 批之後停了：%v", round, err)
			break
		}
	}
	t.Logf("送鍵之後總共 %d 道、抽了 %d 次；呼叫端 %v", o.Steps(), draws, callers)
	if draws == 0 {
		t.Skip("送了鍵還是沒抽——要更長的操作腳本才走得到會抽亂數的地方")
	}
}
