// Package input 把 Ebitengine 的鍵盤與滑鼠事件翻成遊戲的操作 token。
//
// 這裡只放與輸入裝置無關的那一半：可點區域怎麼算、點中之後送哪幾個 token。
// 讀鍵盤與讀滑鼠的那一半在 `cmd/kol/input.go`，因為 Ebitengine 的套件初始化需要
// 開得起視窗，混在一起這裡就無法在無頭容器內測試。
//
// 遊戲的每一層狀態機（`flow.Flow`、`town.Session`、`battle.State`）都只認 token，
// 不知道有鍵盤也不知道有滑鼠。對拍收據餵的是同一套 token，所以互動遊玩走的是
// 與二十六份收據完全相同的那條路徑，而不是另一份平行的輸入處理。
//
// 滑鼠同樣翻成既有 token——點選單第三列產生的是 `D D RET`，不是新的 API。
// 這讓「現代化滑鼠」不必碰任何已經對拍過的狀態機。
package input

import "image"

// Target 是畫面上一個可以點的區域。點中就送出 Tokens——那是既有的鍵盤 token，
// 所以呼叫端不必為滑鼠準備第二套狀態轉移。
//
// Rect 用的是畫布座標（英文模式 320×200、中文模式 640×400）。Ebitengine 的
// `CursorPosition` 回傳的已經是這個座標空間（logical position，已考慮視窗縮放），
// 所以兩邊直接比對，不需要再換算倍率。
type Target struct {
	Rect   image.Rectangle
	Tokens []string
}

// Row 是清單類畫面常見的那種可點區域：從 (x, y) 起、每列高 step 的第 i 列，
// 點下去等於把游標移過去再按 Return。up／down 是那個畫面的上下 token。
//
// 產生的是「移動 n 次再 RET」而不是「直接設游標」，因為游標是各狀態機自己的
// 欄位，只有它知道邊界與連動（例如建角頁的游標與立繪是一起動的）。
func Row(x, y, w, step, index, cursor int, up, down string) Target {
	tokens := make([]string, 0, 4)
	for i := cursor; i > index; i-- {
		tokens = append(tokens, up)
	}
	for i := cursor; i < index; i++ {
		tokens = append(tokens, down)
	}
	return Target{
		Rect:   image.Rect(x, y+index*step, x+w, y+index*step+step),
		Tokens: append(tokens, "RET"),
	}
}

// Cell 是圖示列那種橫向排列的可點區域，語意與 Row 相同，只是白框走的是 `<` `>`。
func Cell(x, y, w, h, step, index, cursor int) Target {
	tokens := make([]string, 0, 8)
	for i := cursor; i > index; i-- {
		tokens = append(tokens, "<")
	}
	for i := cursor; i < index; i++ {
		tokens = append(tokens, ">")
	}
	return Target{
		Rect:   image.Rect(x+index*step, y, x+index*step+w, y+h),
		Tokens: append(tokens, "RET"),
	}
}

// Hit 是左鍵那一下的命中判斷，與輸入裝置無關所以可以單獨測試。
// cursorX、cursorY 已經是畫布座標，直接比對。
func Hit(targets []Target, cursorX, cursorY int, fallback string) []string {
	p := image.Pt(cursorX, cursorY)
	for _, t := range targets {
		if p.In(t.Rect) {
			return t.Tokens
		}
	}
	if fallback != "" {
		return []string{fallback}
	}
	return nil
}
