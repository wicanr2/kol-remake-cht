package town

// Session 是城鎮這一段的可玩狀態機：戶外走動 → 站到門口 → 進建築 → 出來。
// 它只描述狀態轉移，不碰繪製也不碰輸入裝置，因此可以單獨測試與重播。

// Mode 是城鎮畫面目前的模式。
type Mode int

const (
	// ModeMap 是戶外地圖：方向鍵移動隊伍。
	ModeMap Mode = iota
	// ModeDoor 是站在門口：地圖視窗換成建築門面圖，圖示列剩 DOOR 與 U-TURN。
	ModeDoor
	// ModeInside 是進了建築：旅店那張版面，顯示進門敘述。
	ModeInside
	// ModeGate 是走到地圖邊界：地圖視窗換成林地圖，圖示列同樣是 DOOR 與 U-TURN。
	// 原版走的是同一支 `sub_1198E`，差別只在 `word_2FBC2` 是 0FFFFh（RE-065）。
	ModeGate
)

// EdgePicture 是走到地圖邊界時畫在地圖視窗的那張圖：MISCPICS 的圖 0，192×96，
// 與門面圖同一個矩形。原版是 `sub_1198E` 裡的 `sub_1103F(22h)`（RE-065）。
var EdgePicture = struct {
	Archive string
	Index   int
}{Archive: "MISCPICS", Index: 0}

// DoorIcons 是門口狀態的兩個圖示，順序與原版 `sub_11703(2)`、`sub_11703(0Fh)` 相同。
var DoorIcons = []int{DoorIcon, UTurnIcon}

// MapIcons 是回到地圖時剩下的圖示列。原版的 U-TURN 分支是
// `sub_1449C()`（清空圖示列與白框、重畫地圖視窗）接一個 `sub_11703(2)`，
// 從建築物出來也是先清空再補一個 DOOR，所以兩種情形都只剩這一格（RE-065）。
var MapIcons = []int{DoorIcon}

// 門口圖示列的位置。
const (
	doorIconDoor = iota
	doorIconUTurn
)

// Session 是一次城鎮遊玩。
type Session struct {
	Outdoor   *Outdoor
	Buildings *TownBuildings
	Text      *Text
	Trans     *Translations
	Party     []Member

	Mode   Mode
	Cursor int  // 門口圖示列的白框位置
	Inside *Inn // ModeInside 時的畫面狀態

	// LeftTown 記的是玩家在地圖邊界選了 DOOR。原版此時把鏡頭夾限放回 0…63、
	// 讀出城門在世界地圖上的座標，再載入 TOWN0 的野外地圖；野外地圖本身還沒實作，
	// 所以這裡只把「離開了城鎮」與離開的門記下來，交給上層決定要走到哪裡。
	LeftTown  bool
	LeftGate  Point
}

// NewSession 建立一次城鎮遊玩。
func NewSession(out *Outdoor, buildings *TownBuildings, text *Text,
	trans *Translations, party []Member) *Session {
	out.SetBuildings(buildings)
	s := &Session{Outdoor: out, Buildings: buildings, Text: text,
		Trans: trans, Party: party}
	// 起始狀態的圖示列由樣本自己帶進來（例如剛從建築物出來時的那一格 DOOR），
	// 不在這裡重建——原版清空圖示列的地方在移動之後，不是進畫面的時候。
	s.setCursor(out.Selected)
	switch {
	case out.AtEdge:
		s.Mode = ModeGate
	case out.AtDoorCell:
		s.Mode = ModeDoor
	default:
		s.Mode = ModeMap
	}
	return s
}

// syncMode 依隊伍位置決定模式。只在**吃了方向鍵之後**呼叫——原版的掃門迴圈也是
// 接在移動後面，而且撞到地形而沒走成的那一步同樣會跑掃門迴圈。
//
// 圖示列的生命週期也在這裡：走到門口或地圖邊界時重建成 DOOR ＋ U-TURN；
// 走到沒有門的格子上時，原版在掃門迴圈落空之後把圖示列清掉並重畫地圖視窗
// （`sub_10000:0x10B0B`），所以圖示列不會一直跟著隊伍走。
func (s *Session) syncMode() {
	switch {
	case s.Outdoor.AtEdge:
		s.Mode = ModeGate
		s.setBar(DoorIcons)
	case s.Outdoor.AtDoorCell:
		s.Mode = ModeDoor
		s.setBar(DoorIcons)
	default:
		s.Mode = ModeMap
		s.setBar(nil)
	}
}

// setBar 重建圖示列並把白框收回第一格，對應原版的「清空 word_2DF3C 之後再逐個
// `sub_11703`」。列與白框都寫回 Outdoor，繪製端只讀那一份。
func (s *Session) setBar(icons []int) {
	s.Outdoor.Icons = icons
	s.setCursor(0)
}

// Walk 依方向鍵移動隊伍。只有在戶外地圖模式才有作用。
func (s *Session) Walk(dx, dy int) bool {
	if s.Mode != ModeMap {
		return false
	}
	moved := s.Outdoor.Walk(dx, dy)
	s.syncMode()
	return moved
}

// MoveCursor 依原版鍵位移動圖示列的白框：`>` 往右、`<` 往左，兩端繞回。
// 原版在地圖迴圈與門口迴圈都吃這兩個鍵，條件都是圖示列非空（`word_2DF3C != 0`）。
func (s *Session) MoveCursor(delta int) {
	if s.Mode == ModeInside {
		return
	}
	n := len(s.Outdoor.Icons)
	if n == 0 {
		return
	}
	s.setCursor(((s.Cursor+delta)%n + n) % n)
}

// setCursor 同時更新繪製端看得到的位置。
func (s *Session) setCursor(v int) {
	s.Cursor = v
	s.Outdoor.Selected = v
	s.Outdoor.DoorCursor = v
}

// Accept 是 Return：門口選 DOOR 就進去、選 U-TURN 就回地圖；建築內就出來。
func (s *Session) Accept() {
	switch s.Mode {
	case ModeGate:
		if s.Cursor == doorIconUTurn {
			s.leaveDoor()
			return
		}
		x, y := s.Outdoor.PartyPosition()
		s.LeftGate = Point{X: x, Y: y}
		s.LeftTown = true
		s.leaveDoor()
	case ModeDoor:
		if s.Cursor == doorIconUTurn {
			s.leaveDoor()
			return
		}
		s.enter()
	case ModeInside:
		s.leaveDoor()
	}
}

// enter 進入建築：畫面由門的資料、肖像與 5-bit 文字組出來。
func (s *Session) enter() {
	door := s.Outdoor.AtDoor()
	if door == nil {
		return
	}
	town := s.Outdoor.Town
	archive := s.Outdoor.CrestArchive()
	var text *BuildingText
	if s.Text != nil {
		text = s.Text.Building(town, door.Index)
	}
	in := EnterBuilding(archive, door, text, s.Party)
	in.Apply(s.Trans.Door(town, door.Index))
	s.Inside = in
	s.Mode = ModeInside
}

// leaveDoor 從門口或建築內回到地圖。原版的掃門迴圈接在移動後面，
// 所以走出來時人還站在門的格子上，畫面卻是地圖——門口狀態要一併清掉。
func (s *Session) leaveDoor() {
	s.Inside = nil
	s.Outdoor.AtDoorCell = false
	s.Outdoor.AtEdge = false
	s.Mode = ModeMap
	s.setBar(MapIcons)
}

// Icons 回傳目前該畫的圖示列與白框位置。
func (s *Session) Icons() ([]int, int) {
	if s.Mode == ModeInside {
		return []int{BuildingIcon}, 0
	}
	return s.Outdoor.Icons, s.Outdoor.Selected
}
