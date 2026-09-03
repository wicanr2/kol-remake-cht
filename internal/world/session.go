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

	// Rand 是天候隨機漫步用的亂數來源；nil 就不換天候。
	Rand Rand
	// OnNewDay 在換日時被呼叫，參數是每名隊員的營養條要加多少。
	OnNewDay func(hunger int)
	// Camping 是換日那一張 CAMPING! 畫面：原版停在那裡等一個鍵
	// （`sub_297F2:0x2989D`…`0x298E2`）。
	Camping bool

	// Quest 是站上任務地點時蓋在地圖上的提示；沒有就是 nil。
	Quest *QuestPrompt
	// OnArrive 在世界座標真的變了之後被呼叫，回傳要顯示的任務提示；
	// 沒有東西發生就回 nil。原版的比對迴圈也在同一個位置（`sub_29004:0x29450`）。
	OnArrive func(x, y int) *QuestPrompt
	// OnQuestAnswer 是玩家對「Will you partake?」的回答。
	OnQuestAnswer func(yes bool)
}

// Rand 是可注入的亂數來源；n 為骰面數，回傳 1…n。與角色建立用的是同一個介面。
type Rand interface {
	Roll(n int) int
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
		// 走到城鎮入口的那一步不更新地形標籤，時刻也不推進：原版此時跳進
		// `sub_29BBB`，整支 `sub_297F2` 被略過（RE-070）。
		if s.Mode == ModeTown {
			return moved
		}
		s.Screen.SyncTerrain()
		if s.OnArrive != nil {
			if p := s.OnArrive(t.X, t.Y); p != nil {
				s.SetQuest(p)
			}
		}
	}
	if moved {
		s.advanceTime()
	}
	return moved
}

// advanceTime 是走一步之後的時刻推進（`sub_297F2`）。格內的次格移動也算一步：
// 原版「世界座標沒變」那一條路（`0x295C1`）同樣呼叫這一支。**被地形擋住的那一步
// 不算**：原版走 `0x2958F` 把座標復原，之後 `jmp 0x295DD` 跳過整支 `sub_297F2`。
func (s *Session) advanceTime() {
	c := s.Screen.Clock
	if c == nil {
		return
	}
	part := c.Part
	if c.Advance(s.Screen.TerrainCode, s.Screen.Movement) {
		s.Camping = true
		// 換日那一張把天候切到第 5 組（`sub_297F2:0x2989D`）。
		s.Screen.Group = CampWeather
		if s.OnNewDay != nil {
			s.OnNewDay(c.Hunger())
		}
	}
	if c.Part != part {
		s.rollWeather()
	}
	s.Screen.TimeIndex = c.Part
	s.Screen.Date = c.DateLines()
}

// rollWeather 是時段變換時的天候隨機漫步。原版每次都擲，擲出同一組就不重畫。
func (s *Session) rollWeather() {
	c := s.Screen.Clock
	if s.Rand == nil {
		return
	}
	next := c.NextWeather(s.Rand.Roll(0x80) - 1)
	if next == c.Weather {
		return
	}
	c.Weather = next
	s.Screen.Group = next
}

// EndCamping 是 CAMPING! 畫面上按下任何一鍵。
func (s *Session) EndCamping() {
	if !s.Camping {
		return
	}
	s.Camping = false
	c := s.Screen.Clock
	// 原版收掉那一張之後把天候設回第 1 組，再推日期、重畫日期框
	// （`sub_297F2:0x298FC`…`0x2997C`）。
	c.Weather, s.Screen.Group = ClearWeather, ClearWeather
	c.NextDay()
	s.Screen.Date = c.DateLines()
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
	// 任務提示蓋著地圖時，圖示列只有是／否，答完才回地圖。
	if s.Quest != nil {
		yes := s.Screen.Icons[s.Cursor]-3 == IconYes
		if s.OnQuestAnswer != nil {
			s.OnQuestAnswer(yes)
		}
		if s.Quest != nil {
			s.ClearQuest()
		}
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

// 任務地點的提示（RE-079）。走進 `payload3[0x323 + i]`／`[0x33B + i]` 那一格，
// 而且隊上有人帶著這個任務時，原版會把地圖視窗蓋掉，畫一個 (96,55)–(296,152) 的框
// （`sub_2A437`），印八列敘述，再問記錄 0xEA「Will you partake?」。

// QuestBox 是那個框；QuestText 是敘述第一列的左上角，行距 8。
var (
	QuestBox  = Box{X: 96, Y: 55, Width: 201, Height: 98}
	QuestText = struct{ X, Y, Step int }{X: 104, Y: 65, Step: 8}
)

// 是／否的圖示編號，與城鎮共用。
const (
	IconYes = 7
	IconNo  = 8
)

// QuestPrompt 是站上任務地點時蓋在地圖上的那一頁。
type QuestPrompt struct {
	// Lines 是八列敘述；Ask 是最後那一句問話。
	Lines []string
	Ask   string
}

// SetQuest 把畫面切成任務地點的提示，圖示列換成是／否。
func (s *Session) SetQuest(p *QuestPrompt) {
	s.Quest = p
	s.setIcons([]int{IconYes, IconNo})
	s.Cursor, s.Screen.Selected = 0, 0
}

// ClearQuest 收掉提示，圖示列回到地圖的那一組。
func (s *Session) ClearQuest() {
	s.Quest = nil
	s.setIcons(s.mapIcons(false))
	s.Cursor, s.Screen.Selected = 0, 0
}
