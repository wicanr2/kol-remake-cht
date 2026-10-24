package town

import (
	"github.com/wicanr2/kol-remake-cht/internal/magic"
	"github.com/wicanr2/kol-remake-cht/internal/portrait"
	"strconv"
	"strings"
)

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
	// DisplayName 只在中文模式設定：把代入對話句的資料原名（例如教派名）換成
	// 顯示譯名（SPEC-033）。nil 時照原名；遊戲狀態與存檔永遠存原名。
	DisplayName func(string) string
	// SpellGloss 只在中文模式設定：法術名的音節注解（SPEC-034），法術名本身不譯。
	SpellGloss func(string) string
	Outdoor    *Outdoor
	Buildings  *TownBuildings
	Text       *Text
	Topics     *Topics
	UIText     *UIText
	Trans      *Translations
	Party      []Member
	// Roster 是全部角色檔的清單，招募同伴時用來挑出這間建築裡的人。
	Roster []Candidate
	// RosterListTail 是原版名冊清單留在共用緩衝 `byte_37CF8` +0x28／+0x29 的兩格：
	// 清單第 8、9 名的槽號，清單不夠長就保留上一次的值；遊戲啟動時是 0。戰鬥裡敵人
	// 抽攻擊式會讀到它（RE-159）。只存在記憶體裡，不進存檔。
	RosterListTail [2]int
	// TownCode 是這座城鎮在世界地圖上的三字母代碼（BRT、MIN…），
	// 與角色檔所在地欄的第二段比對。空字串代表不比對城鎮。
	TownCode string
	// Unimplemented 記下玩家選到但還沒實作的圖示，讓收據看得見。
	Unimplemented []string
	// TurnedIn 是這一次進門交出去的任務物品（RE-109）；沒有就是 nil。
	TurnedIn []QuestTurnIn
	// TookQuest 是這一次接下的任務（RE-110）；沒接就是 nil。
	TookQuest *TakenQuest
	// CJK 為真時是中文模式。資料頁的中文版面走 SPEC-030 的那一組座標
	// （`CJKSheet`），與英文版共用同一份 `SheetPage`——欄位與計算相同，只是
	// 中文字模格高 18 而原版行距只有 6…8，等比放大會溢出，所以垂直重排過。
	CJK bool

	// Rand 是 LISTEN 挑話題槽用的亂數來源。原版是 `rand(6)`；這裡可以注入，
	// 玩家路徑與測試才能重播。回傳 0…n-1。
	Rand func(n int) int

	Mode   Mode
	Cursor int  // 門口圖示列的白框位置
	Inside *Inn // ModeInside 時的畫面狀態
	// Portraits 是立繪的合成版面（RE-126）。沒有它就畫不出資料頁那兩頁的人。
	Portraits *portrait.Layout
	// Magic 是法術的參數字典（RE-128）。沒有它法術編輯器的每一欄都是空的，
	// 造價也算不出來，所以 SCROLL 選單的第二列不會出現。
	Magic *magic.Dictionary

	// told 記的是這一次進門裡哪些話題槽已經講過。原版是進門時清零的六個旗標。
	told    [TopicSlots]bool
	greeted bool
	// greetWait 是讀完被拒或被趕那一句之後要做的事；askingWho 是停在「誰要開口」（RE-166）。
	greetWait greetWait
	askingWho bool
	// 任務物品先在第一個 SPACE 後交換，訊息 6／7 讀完才替全隊回報（RE-139）。
	turnInChecked        bool
	turnInAwaitingReport bool
	// recruit 是招募流程的狀態；nil 代表不在招募中。
	recruit *Recruit
	// rest 是 REST 流程的狀態；nil 代表不在住宿／治療中（SPEC-006）。
	rest *restState
	// shop 是商店流程的狀態；nil 代表不在買賣中（RE-075）。
	shop *shopState
	// train 是訓練流程的狀態；nil 代表不在訓練中（RE-076）。
	train *trainState
	// scroll 是 SCROLL 那一格的流程（入會與買法術，RE-121）。
	scroll *scrollState
	// quest 是接任務流程的狀態；nil 代表不在接任務中（RE-110）。
	quest *questOffer
	// mouth 是 MOUTH 打字中的問句；nil 代表沒在打字（RE-067）。
	mouth *mouthState
	// questList 是任務一覽翻到第幾組；nil 代表沒開（RE-111）。
	questList *questListState
	// arena 是競技場的對話；nil 代表不在競技場（RE-108）。
	arena *arenaState
	// pendingFight 是談完之後等著上層開打的那一場。
	pendingFight *ArenaFight
	// quit 記下玩家在「你要離開嗎」那一問答了 YES。
	quit bool

	// FaithfulInnStay 開啟時客棧住宿完全照原版：問「存到硬碟嗎」、角色離隊、
	// 隊伍空了問是否離開遊戲。預設關閉——與原版對拍時才打開（SPEC-006）。
	FaithfulInnStay bool
	// SaveGame 在客棧住宿的那一刻被呼叫，把整份存檔寫出去。回傳錯誤代表存檔
	// 失敗，此時不扣錢、不改所在地、也不離隊。沒有掛就當作存檔成功。
	SaveGame func() error
	// SaveError 記下最後一次存檔失敗的原因。
	SaveError error
	// asking 為真時停在「是否搜尋硬碟」那一問。
	asking bool
	// pending 是「按空白鍵才會出現」的下一段文字。原版每段話之間都會停在
	// `sub_144C3`——圖示列縮成一格 SPACEBAR，等玩家讀完。
	pending []string
	// leaving 等待 SPEC-024 多人隊伍的建築告別確認，不屬於存檔資料。
	leaving bool
	topic   *BuildingTopics
	door    *Door
	// sheetHost 非 nil 表示資料頁是在城鎮戶外或野外按數字鍵打開的（RE-150）。
	sheetHost *sheetHost

	// OnLeaveTown 在玩家於城門選 DOOR 時被呼叫一次，參數是隊伍在世界地圖上的座標。
	// 上層用它換到野外畫面；沒有掛就只是把事情記下來。
	OnLeaveTown func(Point)
	// LoadStartingTown 依角色所在地載入已驗證的城鎮資料，不自行改變遊玩狀態。
	// 同一個代碼可能有多個入城點（TOWN7 十個都是 MIN），全部依記錄順序回傳。
	LoadStartingTown func(code string) ([]TownEntry, *TownBuildings, *Grid, error)
	// World 只是給上層判斷「有沒有野外畫面可以換過去」的旗標。
	World        any
	WorldEntered bool

	// LeftTown 記的是玩家在地圖邊界選了 DOOR。原版此時把鏡頭夾限放回 0…63、
	// 讀出城門在世界地圖上的座標，再載入 TOWN0 的野外地圖。**這一層不自己走出去**：
	// 只把「離開了城鎮」與離開的門記下來，野外那一段由 `internal/world` 接手，
	// 交給上層決定要走到哪裡。
	LeftTown bool
	LeftGate Point
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

// SetBarForMode 依目前的位置重建圖示列。讀檔之後畫面要與走到那一格時相同。
func (s *Session) SetBarForMode() { s.syncMode() }

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
	// 物品那一頁把上下鍵吃掉當成移動清單（`sub_196A9` 的 0x4800／0x5000）。
	if s.ShowingItemPage() {
		if dy != 0 {
			s.MoveItemSelection(dy)
		}
		return true
	}
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
		// 建築內的圖示列用同一組鍵，白框在自己的 Inside.Selected 上繞回。
		if s.Inside == nil {
			return
		}
		n := len(s.Inside.Icons)
		if n == 0 {
			return
		}
		s.Inside.Selected = ((s.Inside.Selected+delta)%n + n) % n
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
		s.LeftGate = s.worldGate()
		s.LeftTown = true
		s.leaveDoor()
	case ModeDoor:
		if s.Cursor == doorIconUTurn {
			s.leaveDoor()
			return
		}
		s.enter()
	case ModeInside:
		s.acceptInside()
	}
}

// acceptInside 是建築內按下 Return：執行白框那一格的圖示。
func (s *Session) acceptInside() {
	if s.Inside == nil || len(s.Inside.Icons) == 0 {
		return
	}
	id := s.Inside.Icons[s.Inside.Selected] - 3
	switch {
	case s.arena != nil:
		switch id {
		case IconYes:
			switch s.arena.stage {
			case arenaChoosing:
				s.ArenaChoose()
			case arenaBetAsk:
				s.pendingFight = s.BetAnswer(true)
			case arenaBetWho:
				s.BetChoose()
			case arenaBetSide:
				s.BetSide(true)
			default:
				s.pendingFight = s.ArenaAnswer(true)
			}
		case IconNo:
			switch s.arena.stage {
			case arenaBetAsk:
				s.pendingFight = s.BetAnswer(false)
			case arenaBetSide:
				s.BetSide(false)
			default:
				s.pendingFight = s.ArenaAnswer(false)
			}
		case IconUTurn:
			if s.arena.stage == arenaBetWho {
				s.askBet()
				return
			}
			s.LeaveArena()
		case IconSpacebar:
			// 賭場那幾句讀完要回去問「要不要下注」。
			if s.arena.stage == arenaBetAsk && s.pending == nil {
				s.askBet()
				return
			}
			s.Continue()
		}
	case s.ShowingQuestList():
		// 一組一頁，任何鍵翻下一組（原版是 `sub_144C3` 的 SPACEBAR）。
		s.NextQuestPage()
	case s.ShowingSpellList():
		// 那一頁只剩 U-TURN，按下去回資料頁（`sub_166CC:0x167bb`）。
		s.closeSpellList()
	case s.ShowingPartyPage():
		// 那一頁只等一個鍵（`sub_18D61` 收尾的 `sub_144C3`）。
		s.closePartyPage()
	case s.ShowingItemPage():
		p := s.Inside.ItemPage
		switch {
		case p.GiftConfirm:
			if id == IconSpacebar {
				s.confirmGift()
			}
		case p.DropConfirm:
			s.answerItemDrop(id)
		case len(p.DropMessage) > 0:
			s.afterItemAction()
		case len(p.Message) > 0:
			// 空格確認層唯一的SPACEBAR圖示返回同一選中列（SPEC-031）。
			s.closeItemAction()
		case p.Giving:
			// 等玩家按數字選人；U-TURN 取消。
			if id == IconUTurn {
				s.closeItemAction()
			}
		case p.Acting:
			s.ItemAction(id)
		case id == IconItemUse:
			s.AcceptItemSlot()
		case id == IconUTurn:
			s.closeItemPage()
		}
	case s.ShowingSheet():
		switch id {
		case IconUTurn:
			s.closeSheet()
		case IconSheetItems:
			s.OpenItemPage()
		case IconCompanion:
			s.OpenPartyPage()
		case IconSheetNotes:
			s.OpenQuestList()
		case IconSheetSpells:
			s.OpenSpellList()
		default:
			s.note(sheetIconName(id))
		}
	case s.train != nil:
		if id == IconSpacebar {
			s.Continue()
			return
		}
		s.trainAnswer(id)
	case s.shop != nil:
		if id == IconSpacebar {
			s.Continue()
			return
		}
		s.shopAnswer(id)
	case s.scroll != nil:
		if id == IconSpacebar {
			s.Continue()
			return
		}
		s.scrollAnswer(id)
	case s.rest != nil:
		switch id {
		case IconYes:
			s.restAnswer(true)
		case IconNo:
			s.restAnswer(false)
		case IconSpacebar:
			s.Continue()
		}
	case s.asking:
		// 兩個答案都走同一條路：這份原版輸入只有一份角色檔，搜不搜硬碟結果相同。
		s.openList()
	case s.quest != nil:
		switch id {
		case IconYes:
			s.questAnswer(true)
		case IconNo:
			s.questAnswer(false)
		case IconSpacebar:
			s.nextQuestPage()
		}
	case s.recruit != nil:
		if id == IconUTurn || id == IconNo {
			s.cancelRecruit()
			return
		}
		s.joinSelected()
	case id == IconDoor:
		s.leaveDoor()
	case id == IconSpacebar:
		s.Continue()
	case id == IconListen:
		s.Listen()
	case id == IconMouth:
		s.StartAsk()
	case id == IconCompanion:
		s.StartRecruit()
	case id == IconRest:
		s.StartRest()
	case id == IconMirror:
		s.OpenSheet()
	case id == IconTrade:
		s.StartShop()
	case id == IconSell:
		s.StartSell()
	case id == IconTrain:
		s.StartTrain()
	case id == IconScroll:
		s.StartScroll()
	case id == IconAward:
		s.StartAward()
	default:
		s.note(iconName(id))
	}
}

// note 記下玩家選到但還沒實作的圖示。
func (s *Session) note(what string) {
	for _, seen := range s.Unimplemented {
		if seen == what {
			return
		}
	}
	s.Unimplemented = append(s.Unimplemented, what)
}

// iconName 是圖示的英文名稱，只用在 Unimplemented 的紀錄裡。
func iconName(id int) string {
	switch id {
	case IconMouth:
		return "MOUTH"
	case IconRest:
		return "REST/SAVE/HEAL"
	case IconTrade:
		return "BUY"
	case IconTrain:
		return "TRAIN"
	case IconSell:
		return "SELL"
	case IconAward:
		return "AWARD"
	case IconScroll:
		return "SCROLL"
	case IconMirror:
		return "MIRROR"
	}
	return "圖示 " + strconv.Itoa(id)
}

// worldGate 回傳隊伍離開城鎮之後落在世界地圖的哪一格。原版讀的是這座城鎮
// payload 3 位移 0x27 + 2k 的那一組座標（k ＝ `word_2DF62`），而那組座標與世界地圖
// 城鎮記錄的 (x, y) 逐一相同——16 個入城點全部對上（RE-065／RE-066）。
func (s *Session) worldGate() Point {
	if s.Buildings != nil && s.Outdoor.EntryPoint < len(s.Buildings.EntryPoints) {
		return s.Buildings.EntryPoints[s.Outdoor.EntryPoint]
	}
	x, y := s.Outdoor.PartyPosition()
	return Point{X: x, Y: y}
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
	s.door = door
	s.told = [TopicSlots]bool{} // 原版進門時把六個「講過了」旗標清零
	s.greeted = false
	s.greetWait, s.askingWho = greetNone, false
	s.pending = nil
	s.topic = nil
	if s.Topics != nil {
		s.topic = s.Topics.Building(town, door.Index)
	}
	s.quest = nil
	s.mouth = nil
	s.questList = nil
	s.TookQuest = nil
	s.TurnedIn = nil
	s.turnInChecked = false
	s.turnInAwaitingReport = false
}

// swapFirstQuestItem 依隊員與物品格順序只交換第一件吻合物品（RE-139）。
// 原版先讓玩家讀完進門敘述，回報旗標還要等訊息 6／7 的下一個 SPACE。
func (s *Session) swapFirstQuestItem(door *Door) bool {
	q := door.Quest()
	if q == nil {
		return false
	}
	for i := range s.Party {
		slot := s.Party[i].SwapQuestItem(q)
		if slot < 0 {
			continue
		}
		s.TurnedIn = []QuestTurnIn{{
			Member: i, Slot: slot, Site: q.Site,
			Gave: q.Wants != nil, Name: door.QuestWants, Got: door.QuestGives}}
		return true
	}
	return false
}

// reportTurnedInQuest 把同一件任務已完成的所有隊員都標成已回報，
// 不只更新剛才持有任務物品的角色（`sub_11A93:0x11C63..0x11CD0`）。
func (s *Session) reportTurnedInQuest() {
	if len(s.TurnedIn) == 0 {
		return
	}
	for i := range s.Party {
		if ReportQuest(s.Party[i].Quests, s.Party[i].Progress, 0, s.TurnedIn[0].Site) {
			s.TurnedIn[0].Reported = true
		}
	}
}

// Door 是現在進的這一棟；不在建築內就是 nil。
func (s *Session) Door() *Door { return s.door }

// QuestTurnIn 是一次交付的結果。
type QuestTurnIn struct {
	// Member 是隊伍裡的第幾名；Slot 是身上的第幾格（索引與 CarriedOffsets 對齊）。
	Member, Slot int
	// Site 是這一棟收的任務編號；Reported 為真表示全隊至少一人的完成旗標
	// 在交付台詞後改成「回報過」。
	Site     int
	Reported bool
	// Gave 恆為真（有交出東西才會有這一筆）；Name 是交出去的那件、Got 是換回來的。
	Gave bool
	Name string
	Got  string
}

// Continue 是建築內按 SPACEBAR。有等著的下一段就先顯示它；沒有了才把圖示列
// 換回這棟建築真正的那一列。
func (s *Session) Continue() {
	// 競技場不在建築裡，沒有 door——它的那幾句一樣要按鍵推進，所以 door 的檢查
	// 放在它後面。
	if s.Mode != ModeInside || s.Inside == nil {
		return
	}
	if len(s.Inside.Icons) != 1 || s.Inside.Icons[0] != IconSprite(IconSpacebar) {
		return
	}
	if s.leaving {
		s.finishLeaveDoor()
		return
	}
	if s.continueGreeting() {
		return
	}
	// 隊伍立繪頁等待的是任一按鍵（RE-128 的 sub_144C3），空白鍵要和
	// Return 一樣回資料頁；只恢復圖示列會留下上一頁的立繪。
	if s.ShowingPartyPage() {
		s.closePartyPage()
		return
	}
	// 賭場那幾句讀完就回去問「要不要下注」（RE-108）。
	if s.arena != nil && s.arena.stage == arenaBetAsk {
		s.askBet()
		return
	}
	if s.door == nil {
		return
	}
	if s.pending != nil {
		s.Inside.Dialog = s.pending
		s.Inside.DialogZh = nil
		s.pending = nil
		return
	}
	// 接任務的敘述有兩頁，讀完才問「Will you do this?」（RE-110）。
	if s.quest != nil {
		s.nextQuestPage()
		return
	}
	// REST 的流程裡有幾段話讀完之後還要繼續（免費客棧、錢不夠、治好了）。
	if s.rest != nil && s.rest.next != nil {
		step := s.rest.next
		s.rest.next = nil
		step()
		return
	}
	// 商店同理：錢不夠與背包滿了都要讀完才回清單。
	if s.shop != nil && s.shop.next != nil {
		step := s.shop.next
		s.shop.next = nil
		step()
		return
	}
	// SCROLL 同理：錢不夠、背包滿了、法術學滿了與入會成功都要讀完才接下一步。
	if s.scroll != nil && s.scroll.next != nil {
		step := s.scroll.next
		s.scroll.next = nil
		step()
		return
	}
	// 訓練也有幾段話讀完才接下一步。
	if s.train != nil && s.train.next != nil {
		step := s.train.next
		s.train.next = nil
		step()
		return
	}
	if s.turnInAwaitingReport {
		s.reportTurnedInQuest()
		s.turnInAwaitingReport = false
	} else if !s.turnInChecked {
		s.turnInChecked = true
		if s.swapFirstQuestItem(s.door) {
			s.Inside.Dialog = append(s.message(6), s.message(7)...)
			s.Inside.DialogZh = nil
			s.setInsideIcons([]int{IconSpacebar})
			s.turnInAwaitingReport = true
			return
		}
	}
	if !s.greeted {
		// 進門那一刻說話的是建築裡的人，所以隊員外框是藍的；讀完敘述換玩家開口，
		// 發言者的白框就回到隊員身上（原版畫格 071 與 072 的差別就在這裡）。
		// 多人隊伍要先選誰開口，被討厭的人與全隊被趕見 greet.go（RE-166）。
		s.startGreeting()
		return
	}
	s.setInsideIcons(InteriorIcons(s.door.IconFlags, len(s.Party)))
}

// restoreSpeaker 把白框還給隊伍裡的發言者。沒有指定過就是隊長。
func (s *Session) restoreSpeaker() {
	if len(s.Inside.Party) == 0 {
		return
	}
	speaker := 0
	for i, m := range s.Party {
		if m.Speaker {
			speaker = i
			break
		}
	}
	for i := range s.Inside.Party {
		s.Inside.Party[i].Speaker = i == speaker
	}
}

// greeting 是選定發言者之後店主說的那一段：招呼語再加上
// KOLDATA 記錄 45 的 `"What can I do for you?"`。
// 隊伍不只一人時，訊息 16 是「誰要開口」那一問，不在這裡（greet.go，RE-166）。
func (s *Session) greeting() []string {
	speaker := Member{}
	if i := s.speakerIndex(); i >= 0 {
		speaker = s.Party[i]
	}
	// 發言者身上有傷時，招呼語整段換成 KOLDATA 的兩筆記錄；能治療的建築
	// （旗標 byte 0 的 bit6）換成另一筆（`sub_11A93:0x11F47`，RE-072）。
	if speaker.Wounds != 0 {
		var lines []string
		if s.door != nil && s.door.IsAbbey() {
			lines = s.uiLines(WoundedHealRecord)
		} else {
			lines = append(s.uiLines(WoundedRecord), s.uiLines(WoundedRecord+1)...)
		}
		lines = nonEmpty(lines)
		if ask := s.uiRecord(GreetingRecord); ask != "" {
			lines = append(lines, ask)
		}
		return lines
	}
	// 沒有傷、建築欣賞他的種族或職業：訊息 15（`sub_11A93:0x11FA4`）。
	index := GreetMessage(speaker)
	if s.rule().Likes(speaker) {
		index = FavoredMessage
	}
	lines := s.message(index)
	if ask := s.uiRecord(GreetingRecord); ask != "" {
		lines = append(lines, ask)
	}
	return lines
}

// nonEmpty 去掉空行。原版的記錄常以 `@` 結尾，拆完最後會多一個空字串。
func nonEmpty(lines []string) []string {
	out := lines[:0]
	for _, l := range lines {
		if l != "" {
			out = append(out, l)
		}
	}
	return out
}

// uiLines 取 UI 文字庫的一筆記錄並拆成幾行；中文模式優先用譯文。
//
// 譯文與原文用**同一套語法**（`@` 換行、`~Vxx`／`~Hxx` 定位），所以走同一套處理。
// 中文比英文短，譯文常把原版的兩行併成一行，`@` 因此可以比原文少（
// `tools/verify_ui_translation.py` 的判準）。
func (s *Session) uiLines(index int) []string {
	if zh, ok := s.Trans.UIRecordOK(index); ok {
		return splitUILines(zh)
	}
	if s.UIText == nil {
		return nil
	}
	return s.UIText.Lines(BankUI, index)
}

// uiRecord 取 UI 文字庫的一筆記錄，中文模式優先用譯文。版面標記照樣剝掉。
func (s *Session) uiRecord(index int) string {
	if zh, ok := s.Trans.UIRecordOK(index); ok {
		return strip(zh)
	}
	if s.UIText == nil {
		return ""
	}
	return s.UIText.Record(BankUI, index)
}

// uiRaw 取一筆記錄但**保留前後空白**。原版有些記錄靠自己帶的空白接在別段後面，
// 例如訓練問句裡的 `" Skill"`（記錄 0x40）。
func (s *Session) uiRaw(index int) string {
	if zh, ok := s.Trans.UIRecordOK(index); ok {
		return stripKeepSpaces(zh)
	}
	if s.UIText == nil {
		return ""
	}
	return s.UIText.RecordPadded(BankUI, index)
}

// heard 組出 MOUTH 沒對上時的那一句，逐段都走譯文優先。
func (s *Session) heard(who string, category int) string {
	parts := []string{s.uiRecord(HeardPrefix), who,
		s.uiRecord(HeardMiddle), s.uiRecord(HeardObject + category)}
	var out []string
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, " ")
}

// say 顯示一段話並把圖示列縮成一格 SPACEBAR；next 是按空白鍵之後才出現的下一段。
func (s *Session) say(now, next []string) {
	s.Inside.Dialog = now
	s.Inside.DialogZh = nil
	s.pending = next
	s.setInsideIcons([]int{IconSpacebar})
}

// setInsideIcons 把圖示編號換成畫面上的 KOLMPICS 圖號，並把白框收回第一格。
func (s *Session) setInsideIcons(ids []int) {
	icons := make([]int, len(ids))
	for i, id := range ids {
		icons[i] = IconSprite(id)
	}
	s.Inside.Icons = icons
	s.Inside.Selected = 0
	s.Cursor = 0
}

// message 取一則建築訊息的畫面文字。中文模式優先用譯文，沒有譯文就退回原文，
// 畫面因此不會出現空白。
func (s *Session) message(index int) []string {
	if s.door == nil {
		return nil
	}
	if zh := s.Trans.Message(s.Outdoor.Town, s.door.Index, index); len(zh) > 0 {
		return zh
	}
	if s.Text == nil {
		return nil
	}
	b := s.Text.Building(s.Outdoor.Town, s.door.Index)
	if b == nil {
		return nil
	}
	m := b.Message(index)
	if m == nil {
		return nil
	}
	return m.Display
}

// Listen 是 LISTEN：隨機挑一個起點，再**往回**逐格找還沒講過的**沒有關鍵字**的槽
// （`sub_11A93:0x125FD` 是 `var_6--`，繞回 5，回到起點就停）。找到就先講開場白
// 再講內容；六格都講完就講「都跟你說完了」。
func (s *Session) Listen() bool {
	if s.Mode != ModeInside || s.topic == nil || s.Inside == nil {
		return false
	}
	pick := s.Rand
	if pick == nil {
		pick = func(int) int { return 0 }
	}
	start := pick(TopicSlots)
	n := start
	for {
		slot := s.topic.Slot(n)
		// 話題碼為 0 的槽是「不必開口問也會說」的傳聞，LISTEN 只講這些；
		// 有話題碼的槽要玩家用 MOUTH 問到那個關鍵字才會講
		// （`sub_11A93:0x1258C` 是 `jz` 繼續、非零才重試）。
		if slot != nil && slot.Topic == 0 && !s.told[n] {
			s.told[n] = true
			lines := s.message(AskMessage)
			if lead := s.uiRecord(ListenLeadInBase + slot.Category); lead != "" {
				lines = append(lines, lead)
			}
			s.say(lines, s.message(TopicMessageBase+n))
			return true
		}
		n = (n - 1 + TopicSlots) % TopicSlots
		if n == start {
			break
		}
	}
	if ask := s.uiRecord(ListenExhaustedAsk); ask != "" {
		s.say(append(s.message(AskMessage), ask),
			[]string{s.uiRecord(ListenExhaustedReply)})
	}
	return false
}

// Ask 是 MOUTH：玩家打一句話進來，比對關鍵字字首，命中就講對應的話題槽。
// 沒對上但表裡有這個話題，就指出鎮上誰知道；整張表都沒中就講「我不知道」。
func (s *Session) Ask(question string) bool {
	if s.Mode != ModeInside || s.Topics == nil {
		return false
	}
	town := s.Outdoor.Town
	key, ok := s.Topics.Lookup(town, question)
	if !ok {
		s.say(s.message(UnknownMessage), nil)
		return true
	}
	if s.topic != nil {
		if slot := s.topic.SlotFor(key.Topic); slot >= 0 {
			s.told[slot] = true
			// 第 5 槽在收任務物品的建築裡是接任務那一條（RE-110）；
			// 接不了就照常放這一槽的回應。
			if slot == QuestSlot && s.offerQuest() {
				return true
			}
			s.say(s.message(TopicMessageBase+slot), nil)
			return true
		}
	}
	if who := s.Topics.AskWho(town, key.Topic); who != "" && s.UIText != nil {
		s.say([]string{s.heard(who, key.Category)}, nil)
		return true
	}
	s.say(s.message(UnknownMessage), nil)
	return true
}

// leaveDoor 從門口或建築內回到地圖。原版的掃門迴圈接在移動後面，
// 所以走出來時人還站在門的格子上，畫面卻是地圖——門口狀態要一併清掉。
func (s *Session) leaveDoor() {
	if s.leaving {
		return
	}
	// 原版 word_2DF2A 是隊員數：一人跳過告別；多人讀建築訊息 17。
	// 門口與城門取消不走建築告別。
	if s.Mode == ModeInside && s.Inside != nil && len(s.Party) > 1 {
		if lines := s.message(17); len(lines) > 0 {
			s.leaving = true
			// 原版清掉目前發言者再重畫隊伍；只改畫面快照，不改角色保存的發言者。
			s.Inside.Party = append([]Member(nil), s.Inside.Party...)
			for i := range s.Inside.Party {
				s.Inside.Party[i].Speaker = false
			}
			s.say(lines, nil)
			return
		}
	}
	s.finishLeaveDoor()
}

func (s *Session) finishLeaveDoor() {
	s.leaving = false
	s.pending = nil
	s.Inside = nil
	s.Outdoor.AtDoorCell = false
	s.Outdoor.AtEdge = false
	s.Mode = ModeMap
	s.setBar(MapIcons)
}

// Icons 回傳目前該畫的圖示列與白框位置。
func (s *Session) Icons() ([]int, int) {
	if s.Mode == ModeInside {
		if s.Inside == nil {
			return nil, 0
		}
		return s.Inside.Icons, s.Inside.Selected
	}
	return s.Outdoor.Icons, s.Outdoor.Selected
}
