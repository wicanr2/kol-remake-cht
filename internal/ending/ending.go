// Package ending 是結局的資料與觸發判準（RE-106）。
//
// 結局接在**任務戰鬥打完之後**：`sub_2B0FF` 回來時看 `word_34968` 等不等於
// 0xFFFF（`sub_28F5D` 標到任務完成才會是那個值，RE-101），等於才逐一檢查六名
// 角色，其中一名的 `[0x1DD + 群組]` 等於 0x17 就播（`0x2B36B`…`0x2B3A9`）。
//
// 這一份只放規則與版面，畫面在 `internal/ui`。
package ending

// Archive 是七張結局圖所在的 archive（RE-017）。
const Archive = "ENDPICS"

// Bank 是結局敘事所在的文字庫。十八筆記錄裡用掉十七筆，最後一筆沒有呼叫點。
const Bank = "ENDTXT"

const (
	// TriggerOffset 是角色記錄裡決定結局的那一格（`0x2B392` 的 `bx + 1DDh`）。
	// 它在任務旗標區（0x1E2）前面五個 byte，群組偏移與旗標共用 `word_2DF70`。
	TriggerOffset = 0x1DD
	// TriggerValue 是那一格要等於的值（`0x2B3A1`）。**為什麼是 0x17 還沒解**：
	// 任務地點編號是 0…23，「完成到最後一個」是合理的猜測但沒有對到寫入點。
	TriggerValue = 0x17
)

// Rect 是一張圖畫在哪裡。四個值是原版的四個全域：`word_3499E` 左欄、
// `word_349A0` 上列、`word_34A06` 右欄、`word_34A0E` 下列。**欄乘八得像素**，
// 右欄減左欄乘八正好是 RE-017 量到的圖寬、下列減上列正好是圖高。
type Rect struct {
	Left, Top, Right, Bottom int
}

// X／Y 是左上角的像素座標；Width／Height 是像素尺寸。
func (r Rect) X() int      { return r.Left * 8 }
func (r Rect) Y() int      { return r.Top }
func (r Rect) Width() int  { return (r.Right - r.Left) * 8 }
func (r Rect) Height() int { return r.Bottom - r.Top }

// Panel 是結局的一段：一張圖、幾筆敘事，按一次鍵推到下一段。
type Panel struct {
	// Picture 是 ENDPICS 的圖號。
	Picture int
	// Records 是這一段要印的 ENDTXT 記錄，依序接在一起。
	Records []int
	// Box 是圖畫在哪裡。
	Box Rect
}

// Panels 是七段的順序（`sub_2B4A7` 逐段抄出來）。文字自帶 `~Vxx`／`~Hxx`
// 定位，擺在圖的哪一側由文字自己決定，程式不另外算。
var Panels = []Panel{
	{Picture: 0, Records: []int{0, 1}, Box: Rect{1, 13, 22, 135}},
	{Picture: 1, Records: []int{2, 3}, Box: Rect{15, 41, 39, 152}},
	{Picture: 2, Records: []int{4, 5}, Box: Rect{1, 50, 38, 143}},
	{Picture: 3, Records: []int{6, 7, 8}, Box: Rect{1, 50, 23, 147}},
	{Picture: 4, Records: []int{9, 10}, Box: Rect{17, 12, 38, 135}},
	{Picture: 5, Records: []int{11, 12, 13}, Box: Rect{0, 46, 24, 148}},
	{Picture: 6, Records: []int{14, 15, 16}, Box: Rect{16, 45, 39, 143}},
}

// Triggered 回報一名角色的 progress 那幾格有沒有達成結局條件。
// group 是 `word_2DF70`，remake 目前固定 0（同任務旗標，RE-101）。
func Triggered(progress []int, group int) bool {
	if group < 0 || group >= len(progress) {
		return false
	}
	return progress[group] == TriggerValue
}

// AnyTriggered 回報一隊裡有沒有人達成。原版走的是**角色記錄**那六格，
// 空著的（`+0x00` 為零）跳過；remake 這一側傳進來的本來就只有在籍的人。
func AnyTriggered(party [][]int, group int) bool {
	for _, p := range party {
		if Triggered(p, group) {
			return true
		}
	}
	return false
}

// State 是結局播到第幾段。
type State struct {
	Panel int
}

// Done 回報七段都播完了。
func (s *State) Done() bool { return s.Panel >= len(Panels) }

// Current 是目前那一段；播完回 nil。
func (s *State) Current() *Panel {
	if s.Done() {
		return nil
	}
	return &Panels[s.Panel]
}

// Advance 推到下一段（原版每一段之後 `sub_1B100(0)` 等一個鍵）。
func (s *State) Advance() { s.Panel++ }
