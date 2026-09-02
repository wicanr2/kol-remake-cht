package world

// 野外的狀態機（RE-068）。走到城鎮入口那一格會跳出跟城門一樣的提示：
// 地圖視窗換成該地點的外觀圖、圖示列剩 DOOR 與 U-TURN；選 DOOR 就進城。
// 這一段對應 `sub_29004:0x294A8` 的比對迴圈與 `sub_29BBB`。

// Mode 是野外畫面目前的模式。
type Mode int

const (
	// ModeMap 是在野外走動。
	ModeMap Mode = iota
	// ModeTown 是站在城鎮入口上：換成外觀圖，圖示列剩 DOOR 與 U-TURN。
	ModeTown
)

// 野外會用到的圖示編號（畫面上畫的是 KOLMPICS[編號 + 3]）。
const (
	IconDoor     = 2
	IconUTurn    = 15
	IconRoad     = 23
	IconSignpost = 24
)

// TownPictureBase 是城鎮外觀圖的偏移：`sub_29BBB` 畫的是 `TWN0PICS[記錄索引 + 8]`。
const TownPictureBase = 8

// TownPictureArchive 是野外用的圖庫。原版組檔名時 `word_2DF2C` 是 0，
// 所以 `twn%d%s` 展開成 `TWN0PICS`。
const TownPictureArchive = "TWN0PICS"

// Session 是一次野外旅行。它只描述狀態轉移，不碰繪製也不碰輸入裝置。
type Session struct {
	Screen *Screen
	// OnEnterTown 在玩家於城鎮入口選 DOOR 時被呼叫，參數是那一筆記錄。
	OnEnterTown func(*Town)

	Mode   Mode
	Cursor int
	// Direction 是目前循的道路方向（0…7，順時針由北算起），`NoDirection` 是沒有；
	// Excluded 是 SIGNPOST 排除掉的那個方向。原版是 `word_2FB5E` 與 `word_2FBBA`。
	Direction int
	Excluded  int
	// Town 是腳下那個城鎮入口；不在入口上就是 nil。
	Town *Town
}

// NewSession 建立一次野外旅行。起始狀態是「剛進野外」：地圖模式、空的圖示列。
func NewSession(s *Screen) *Session {
	sess := &Session{Screen: s, Direction: NoDirection, Excluded: NoDirection}
	sess.Reset()
	return sess
}

// Reset 是剛進野外的狀態。原版 `sub_29004` 的入口把次格座標放回 (3, 3)、
// 每次移動一格（`word_322F6 = 1`），而且一個圖示都不推——比對城鎮入口的迴圈
// 要等第一次移動之後才跑，所以剛從城門走出來、人還站在城鎮那一格上時
// 不會立刻彈出進城提示。
func (s *Session) Reset() {
	s.Mode = ModeMap
	s.Town = nil
	s.Direction, s.Excluded = NoDirection, NoDirection
	s.setIcons(nil)
	s.Screen.SyncTerrain()
}

// Sync 依腳下那一格決定模式與圖示列。原版的比對迴圈也是接在移動後面。
func (s *Session) Sync() {
	s.Town = s.Screen.Travel.TownHere()
	if s.Town != nil {
		s.Mode = ModeTown
		s.setIcons([]int{IconDoor, IconUTurn})
		return
	}
	s.Mode = ModeMap
	s.setIcons(s.mapIcons(false))
}

// mapIcons 是在野外走動時的圖示列。空曠處是空的；站在路上會多出 ROAD 與 SIGNPOST，
// 而剛從入口提示退回來時會保留一格 DOOR（`sub_29BBB:0x29C63`）。
func (s *Session) mapIcons(keepDoor bool) []int {
	var icons []int
	if keepDoor {
		icons = append(icons, IconDoor)
	}
	if s.Screen.Travel.OnRoad() {
		icons = append(icons, IconRoad, IconSignpost)
	}
	return icons
}

func (s *Session) setIcons(ids []int) {
	icons := make([]int, len(ids))
	for i, id := range ids {
		icons[i] = id + 3
	}
	s.Screen.Icons = icons
	s.Screen.Vacated = make([]bool, len(ids))
	s.Screen.Selected = 0
	s.Cursor = 0
}

// leaveSlot 把白框目前那一格記成「離開過」，之後它的框色是青色而不是亮藍。
func (s *Session) leaveSlot() {
	if s.Cursor >= 0 && s.Cursor < len(s.Screen.Vacated) {
		s.Screen.Vacated[s.Cursor] = true
	}
}

// Walk 走一步。站在入口提示裡時方向鍵沒有作用。
//
// 比對城鎮入口的迴圈只在**世界座標真的變了**的時候跑（`sub_29004:0x293B7`
// 那個分支把整段跳過去）。格內的次格移動不重跑，否則站在入口那一格上時
// 每走一次次格都會再彈一次提示。
func (s *Session) Walk(dx, dy int) bool {
	if s.Mode != ModeMap {
		return false
	}
	t := s.Screen.Travel
	x, y := t.X, t.Y
	moved := t.Walk(dx, dy)
	if t.X != x || t.Y != y {
		s.Sync()
		// 走到城鎮入口的那一步不更新地形標籤：原版此時跳進 `sub_29BBB`，
		// 更新標籤的 `sub_297F2` 整段被略過，標籤因此留在上一格的地形。
		if s.Mode != ModeTown {
			s.Screen.SyncTerrain()
		}
	}
	return moved
}

// MoveCursor 移動圖示列的白框，兩端繞回。
func (s *Session) MoveCursor(delta int) {
	n := len(s.Screen.Icons)
	if n == 0 {
		return
	}
	s.leaveSlot()
	s.Cursor = ((s.Cursor+delta)%n + n) % n
	s.Screen.Selected = s.Cursor
}

// Accept 是 Return，按下白框那一格的圖示。
//
//   - DOOR：進城。入口提示裡與 U-TURN 之後保留的那一格都走這條
//     （`sub_29004:0x29799` 對 `sub_29BBB` 的 DOOR 分支是同一支 `sub_29CA3`）。
//   - U-TURN：回地圖，圖示列只剩一格 DOOR。
//   - ROAD／SIGNPOST：沿道路走一步。
func (s *Session) Accept() {
	if len(s.Screen.Icons) == 0 || s.Cursor >= len(s.Screen.Icons) {
		return
	}
	// 按下之後白框一律回到第 0 格，原本那一格改畫青色（`sub_29004:0x29765`）。
	id := s.Screen.Icons[s.Cursor] - 3
	if s.Mode == ModeMap {
		s.leaveSlot()
		s.Cursor, s.Screen.Selected = 0, 0
	}
	switch id {
	case IconDoor:
		if s.OnEnterTown != nil && s.Town != nil {
			s.OnEnterTown(s.Town)
		}
	case IconUTurn:
		if s.Mode == ModeTown {
			s.Mode = ModeMap
			s.setIcons(s.mapIcons(true))
		}
	case IconRoad:
		s.FollowRoad(false)
	case IconSignpost:
		s.FollowRoad(true)
	}
}

// TownPicture 回傳入口提示要畫的那張外觀圖；不在入口上就回 −1。
func (s *Session) TownPicture() int {
	if s.Mode != ModeTown || s.Town == nil {
		return -1
	}
	return s.Town.Index + TownPictureBase
}
