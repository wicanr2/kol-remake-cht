//go:build oracle

package oracle

import (
	"fmt"

	"github.com/wicanr2/dosgolem/oracle"
)

// At 是「CS:IP 走到這個 IDA 線性位址」，比的是**線性位址**。
//
// dosgolem 自己的 `oracle.At` 比的是 `(Seg, Off)` 結構相等，而 `IDA()` 會把位址
// 正規化成 `offset 0…15` 的形式；KOL 是 far call 的大模式，每一支常式跑起來的 CS
// 是它自己那一段，兩者只有在剛好對齊時才相等。同一個位址交給 `OnCall` 卻是用線性
// 比的——所以 hook 掛得上、`At` 卻永遠不成立，而且**不會報錯**，只會跑滿上限。
//
// 這裡統一成線性，兩邊的語意才一致。
func At(o *oracle.Oracle, ea uint32) oracle.Cond {
	want := o.IDA(ea).Linear()
	return oracle.NewCond(fmt.Sprintf("走到 IDA %#x", ea), func(p *oracle.Oracle) bool {
		return p.IP().Linear() == want
	})
}
