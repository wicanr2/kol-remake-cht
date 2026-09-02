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
)

// DoorIcons 是門口狀態的兩個圖示，順序與原版 `sub_11703(2)`、`sub_11703(0Fh)` 相同。
var DoorIcons = []int{DoorIcon, UTurnIcon}

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
}

// NewSession 建立一次城鎮遊玩。
func NewSession(out *Outdoor, buildings *TownBuildings, text *Text,
	trans *Translations, party []Member) *Session {
	out.SetBuildings(buildings)
	s := &Session{Outdoor: out, Buildings: buildings, Text: text,
		Trans: trans, Party: party}
	s.syncMode()
	return s
}

// syncMode 依隊伍位置決定模式。只在移動之後呼叫——原版的掃門迴圈也是接在移動後面。
func (s *Session) syncMode() {
	if s.Outdoor.AtDoorCell {
		s.Mode = ModeDoor
		s.setCursor(doorIconDoor)
		return
	}
	s.Mode = ModeMap
}

// Walk 依方向鍵移動隊伍。只有在戶外地圖模式才有作用。
func (s *Session) Walk(dx, dy int) bool {
	if s.Mode != ModeMap {
		return false
	}
	if !s.Outdoor.Walk(dx, dy) {
		return false
	}
	s.syncMode()
	return true
}

// MoveCursor 依原版鍵位移動門口圖示列的白框：`>` 往右、`<` 往左，兩端繞回。
func (s *Session) MoveCursor(delta int) {
	if s.Mode != ModeDoor {
		return
	}
	n := len(DoorIcons)
	s.setCursor(((s.Cursor+delta)%n + n) % n)
}

// setCursor 同時更新繪製端看得到的位置。
func (s *Session) setCursor(v int) {
	s.Cursor = v
	s.Outdoor.DoorCursor = v
}

// Accept 是 Return：門口選 DOOR 就進去、選 U-TURN 就回地圖；建築內就出來。
func (s *Session) Accept() {
	switch s.Mode {
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
	s.Mode = ModeMap
	s.setCursor(doorIconDoor)
}

// Icons 回傳目前該畫的圖示列與白框位置。
func (s *Session) Icons() ([]int, int) {
	switch s.Mode {
	case ModeDoor:
		return DoorIcons, s.Cursor
	case ModeInside:
		return []int{BuildingIcon}, 0
	default:
		return s.Outdoor.Icons, s.Outdoor.Selected
	}
}
