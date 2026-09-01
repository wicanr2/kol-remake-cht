// Package book 定義《武士傳說》書本介面的畫面組成。
//
// 這裡只放已由原版同狀態畫面證實的組成事實：哪張素材是哪個畫面的底圖，以及素材被
// 畫在哪個座標。未經證實的位置不得寫進來。
package book

// BOOKPICS 的四張 320×200 底圖。圖號與尺寸出自
// docs/re/kolbook-bookpics-ega-visual-probe.json。
const (
	// Cover 是啟動後的書本封面；已對原版畫面驗證，除滑鼠游標外逐像素相同。
	Cover = 0
	// Contents 是目錄頁；整頁文字都烘在素材裡，已對原版畫面驗證。
	Contents = 1
	// OpenSpread 是翻開的空白書頁，New Character 等頁面以它為底圖。
	OpenSpread = 2
	// Table 是書本所在的紅色木桌背景。
	Table = 3
)

// PortraitOrigin 是角色圖像在 New Character 頁面的 blit 座標。
// 由 tools/locate_asset_in_frame.py 在原版畫面搜出，差異為 0。
var PortraitOrigin = struct{ X, Y int }{X: 208, Y: 106}

// LeftPageText 是原版在 New Character 頁面上繪製文字的範圍。
// 由 BOOKPICS[2] 與原版畫面的差異外接矩形量得，用來把「素材／版面」與「文字」
// 分成兩層對拍：素材層必須逐像素相同，文字層在繁中版本必然不同。
var LeftPageText = struct{ X, Y, Width, Height int }{X: 24, Y: 16, Width: 146, Height: 183}
