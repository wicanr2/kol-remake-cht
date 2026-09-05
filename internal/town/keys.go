package town

import "strings"

// 城鎮這一層的鍵位。原本這段散在 `cmd/kol` 與 `internal/flow` 兩處，
// 兩邊各自維護一份 switch——結果是 flow 那一份少了空白鍵，走 flow 進建築之後
// 敘述翻不了頁，看起來像卡住。鍵位屬於城鎮自己的狀態機，放在這裡只有一份。

// walkKeys 是方向鍵與參考卡的斜走鍵。
var walkKeys = map[string][2]int{
	"L": {-1, 0}, "R": {1, 0}, "U": {0, -1}, "D": {0, 1},
	"A": {-1, -1}, "S": {1, -1}, "Z": {-1, 1}, "X": {1, 1},
}

// Key 送進一個操作 token，回傳城鎮這一層認不認得它。
//
// 清單開著的時候上下鍵移動反白而不是走路——原版的招募、商店與訓練清單都吃上下鍵。
func (s *Session) Key(token string) bool {
	// 打問句的時候鍵盤是輸入框：單一字元進緩衝區、`BS` 退格、`RET` 送出。
	// 其餘的鍵在原版的讀入迴圈裡也不做事。
	if s.Asking() {
		switch token {
		case "RET":
			s.SubmitAsk()
			return true
		case "BS":
			s.BackspaceAsk()
			return true
		case "SPACE":
			s.TypeAsk(' ')
			return true
		}
		if r := []rune(token); len(r) == 1 {
			s.TypeAsk(r[0])
			return true
		}
		return false
	}
	// 在城門選了 DOOR 就換到世界地圖。放在這裡而不是呼叫端：`LeftTown` 是這個
	// 狀態機自己翻的旗標，誰送鍵進來都要走同一條路——原本掛在 `cmd/kol` 那一份
	// 包裝上，走 flow 進來的鍵就繞過了它，落到世界地圖的哪一格只是碰巧對。
	defer func() {
		if s.LeftTown && s.World != nil && !s.WorldEntered && s.OnLeaveTown != nil {
			s.WorldEntered = true
			s.OnLeaveTown(s.LeftGate)
		}
	}()
	// 賭場的金額是打字，跟 MOUTH 一樣先攔下來。
	if s.BettingAmount() {
		switch token {
		case "RET":
			s.SubmitBet()
			return true
		case "BS":
			s.BackspaceBet()
			return true
		}
		if r := []rune(token); len(r) == 1 {
			s.TypeBet(r[0])
			return true
		}
		return false
	}
	switch {
	case s.InArena():
		switch token {
		case "U":
			s.MoveArenaList(-1)
			return true
		case "D":
			s.MoveArenaList(1)
			return true
		}
	case s.Recruiting():
		switch token {
		case "U":
			s.MoveList(-1)
			return true
		case "D":
			s.MoveList(1)
			return true
		}
	case s.Shopping():
		switch token {
		case "U":
			s.MoveShop(-1)
			return true
		case "D":
			s.MoveShop(1)
			return true
		}
	case s.Scrolling():
		switch token {
		case "U":
			s.MoveScroll(-1)
			return true
		case "D":
			s.MoveScroll(1)
			return true
		}
	case s.Training():
		switch token {
		case "U":
			s.MoveTrain(-1)
			return true
		case "D":
			s.MoveTrain(1)
			return true
		case "1", "2", "3", "4", "5":
			s.TrainCount(int(token[0] - '0'))
			return true
		}
	}
	if d, ok := walkKeys[token]; ok {
		s.Walk(d[0], d[1])
		return true
	}
	switch token {
	case "<":
		s.MoveCursor(-1)
	case ">":
		s.MoveCursor(1)
	case "RET":
		s.Accept()
	case "SPACE":
		s.Continue()
	case "LISTEN":
		s.Listen()
	default:
		// `ASK:<字>` 是 MOUTH：把玩家打的那句話送進關鍵字比對。原版是在畫面上
		// 打字，remake 的輸入框還沒做，腳本先用這個形式送同一件事。
		if q, ok := strings.CutPrefix(token, "ASK:"); ok {
			return s.Ask(q)
		}
		return false
	}
	return true
}
