package town

import "strings"

// COMPANION 的招募流程（RE-072）。原版按下 COMPANION 之後先問
// `Look for characters on hard disk?`（圖示列換成大拇指朝上與朝下），
// 回答之後在對話框列出**這間旅店裡不在隊伍中的角色**：白字是反白那一列、
// 其餘是青色，右側第 23 格畫上下箭頭（大字型的碼位 94／95，紅色）。
// 上下方向鍵移動反白，清單超過八列就捲動；兩端都夾住，不繞回。

// ListRows 是對話框放得下的列數：框內 68 px、行距 8。
const ListRows = 8

// ListColumn 是捲動箭頭所在的字格欄。
const ListColumn = 23

// ListArrowUp／ListArrowDown 是捲動箭頭的碼位。原版大字型在 94／95 放的是
// 上下箭頭，不是 `^` 與 `_`（字型 manifest 已註明）。
const (
	ListArrowUp   = 94
	ListArrowDown = 95
)

// 招募畫面的圖示編號。畫面上畫的是 KOLMPICS[編號 + 3]。
const (
	IconYes = 7
	IconNo  = 8
)

// Candidate 是一名可招募的角色。Place 與 Town 是角色檔裡的所在地欄位。
type Candidate struct {
	// Slot 是角色檔裡的槽號（0…15）。原版名冊清單記的是它，戰鬥 AI 會讀到（RE-159）。
	Slot     int
	ID       string
	Full     *Member
	Name     string
	Portrait int
	Place    string
	Town     string
	Female   bool
	Wounds   int
	// Gold 是角色自己的錢包；招募進來之後住宿與治療花的就是這一份（RE-073）。
	Gold int
	// BuildingID 是角色檔記下的留宿建築編號。
	BuildingID int
	// Sheet 是角色資料頁的欄位。
	Sheet *Sheet
	// Quests 是角色記錄 0x1E2 起的六個任務旗標 byte（RE-079）。
	Quests []int
	// Worn 是穿戴欄十格，索引同部位（第十格是腰帶）；空的那一格是 nil（RE-080）。
	Worn [WornSlots]*Good
	// InHand 是手上那件（角色記錄 0x30），不屬於穿戴欄（RE-127）。
	InHand   *Good
	Backpack []Good
	// Spells 是角色記錄 0x200 起那十六筆法術裡非空的幾筆（RE-100）。
	Spells []Spell
	// Progress 是「已回報的任務數」那五格；Carried 是交任務物品要逐 byte 比的
	// 那十七格（RE-107／RE-109）。
	Progress []int
	Carried  [][]byte
}

// SelectLeader 以名冊中的角色開始隊伍；不沿用畫面樣本中的預設隊員。
// 名稱定位沿用 SPEC-006 現行契約，穩定身分的保存契約另行修訂。
func (s *Session) SelectLeader(name string) bool {
	for _, c := range s.Roster {
		if c.Name != name {
			continue
		}
		m := c.member()
		s.Party = []Member{m}
		s.syncParty()
		return true
	}
	return false
}

// StartLeader 還原角色所在旅店，再交給正常城鎮流程（SPEC-016）。
// 跨城時由上層提供唯讀載入器；所在地無法對應時保持原狀。
func (s *Session) StartLeader(name string) bool {
	if s.Outdoor == nil || s.Buildings == nil {
		return false
	}
	for _, c := range s.Roster {
		if c.Name != name || c.Town == "" {
			continue
		}
		if c.Town == s.TownCode && s.startAtLodging(c) {
			return true
		}
		// 所在地在別座城，或同一座城的另一個入城區（讀檔的位置不一定在旅店那一區）：
		// 依所在地的代碼載入，挑第一個鏡頭夾限放得下旅店門口的入城點。原版怎麼挑
		// 入城點沒有讀過；這個順序是 remake 的選法，只影響之後走路時的夾限範圍。
		if s.LoadStartingTown == nil {
			return false
		}
		entries, buildings, grid, err := s.LoadStartingTown(c.Town)
		if err != nil || buildings == nil || grid == nil {
			return false
		}
		for _, entry := range entries {
			if entry.Code != c.Town {
				continue
			}
			// 先在副本驗證建築與位置，失敗時保留玩家原本的隊伍及城鎮。
			trial, outdoor := *s, *s.Outdoor
			trial.Outdoor = &outdoor
			trial.EnterFromWorld(entry, buildings, grid)
			if !trial.startAtLodging(c) {
				continue
			}
			*s.Outdoor = outdoor
			trial.Outdoor = s.Outdoor
			*s = trial
			return true
		}
		return false
	}
	return false
}

// startAtLodging 把隊伍放到角色所在旅店的門口並進門；目前的夾限放不下門口就不動。
func (s *Session) startAtLodging(c Candidate) bool {
	for i := range s.Buildings.Doors {
		door := &s.Buildings.Doors[i]
		if door.RecordID != c.BuildingID || placeKey(door.Name) != c.Place {
			continue
		}
		position := *s.Outdoor
		position.PlaceParty(door.X, door.Y)
		if x, y := position.PartyPosition(); x != door.X || y != door.Y {
			return false
		}
		s.Party = []Member{c.member()}
		s.Outdoor.PlaceParty(door.X, door.Y)
		s.Outdoor.AtEdge = false
		s.Outdoor.AtDoorCell = true
		s.syncParty()
		s.enter()
		return true
	}
	return false
}

// member 建立遊玩中的獨立副本；修改隊員不能暗中改寫尚未保存的名冊。
func (c Candidate) member() Member {
	if c.Full != nil {
		m := CloneMember(*c.Full)
		m.Speaker = true
		return m
	}
	m := Member{Name: c.Name, Portrait: c.Portrait, Female: c.Female,
		Wounds: c.Wounds, Gold: c.Gold, Place: c.Place, TownCode: c.Town,
		BuildingID: c.BuildingID, Worn: c.Worn, InHand: c.InHand,
		Backpack: append([]Good(nil), c.Backpack...),
		Spells:   append([]Spell(nil), c.Spells...),
		Quests:   append([]int(nil), c.Quests...), Progress: append([]int(nil), c.Progress...),
		Speaker: true}
	if c.Sheet != nil {
		sheet := *c.Sheet
		sheet.Weapons = append([]WeaponRow(nil), c.Sheet.Weapons...)
		m.Sheet = &sheet
	}
	for _, record := range c.Carried {
		m.Carried = append(m.Carried, append([]byte(nil), record...))
	}
	if c.InHand != nil {
		item := *c.InHand
		m.InHand = &item
	}
	for i, worn := range c.Worn {
		if worn != nil {
			item := *worn
			m.Worn[i] = &item
		}
	}
	for i := range m.Spells {
		m.Spells[i].Guards = append([]int(nil), c.Spells[i].Guards...)
	}
	m.ID = c.ID
	return m
}

// InnList 是對話框改畫成清單時的內容。Rows 是**視窗內**那幾列。
type InnList struct {
	Rows []string
	// Notes 與 Rows 平行，只給中文畫面用：選中那一列後面的注解（SPEC-034 的
	// 法術名音節解析）。英文畫面不畫，nil 表示沒有注解。
	Notes    []string
	Selected int
	// Total 是清單全部有幾筆。原版只有「視窗裝不下」時才畫捲動箭頭
	// （`sub_1AAE8:0x1AB0D` 比的是列數與筆數），0 代表與 Rows 一樣多。
	Total int
}

// Scrolls 回報這份清單捲不捲得動——捲不動就沒有箭頭。
func (l *InnList) Scrolls() bool { return l.Total > len(l.Rows) }

// Recruit 是招募的狀態：候選人、反白位置與視窗第一列。
type Recruit struct {
	Candidates []Candidate
	Cursor     int
	Top        int
}

// View 回傳視窗內那幾列與反白在視窗內的位置。
func (r *Recruit) View() *InnList {
	end := r.Top + ListRows
	if end > len(r.Candidates) {
		end = len(r.Candidates)
	}
	rows := make([]string, 0, end-r.Top)
	for _, c := range r.Candidates[r.Top:end] {
		rows = append(rows, c.Name)
	}
	return &InnList{Rows: rows, Selected: r.Cursor - r.Top,
		Total: len(r.Candidates)}
}

// Move 移動反白。兩端夾住不繞回，視窗跟著捲到反白看得見為止。
func (r *Recruit) Move(delta int) {
	if len(r.Candidates) == 0 {
		return
	}
	r.Cursor += delta
	if r.Cursor < 0 {
		r.Cursor = 0
	}
	if r.Cursor >= len(r.Candidates) {
		r.Cursor = len(r.Candidates) - 1
	}
	if r.Cursor < r.Top {
		r.Top = r.Cursor
	}
	if r.Cursor >= r.Top+ListRows {
		r.Top = r.Cursor - ListRows + 1
	}
}

// placeKey 是角色檔所在地欄的比對值：建築名稱轉大寫後取前 10 個字元。
// 半途之家的 `Halfway House` → `HALFWAY HO`，與 `CHARDATA` 的欄位逐字相同。
func placeKey(name string) string {
	upper := strings.ToUpper(name)
	if len(upper) > PlaceBytes {
		upper = upper[:PlaceBytes]
	}
	return upper
}

// PlaceBytes 是角色檔所在地欄的長度。
const PlaceBytes = 10

// candidatesHere 回傳這間建築裡不在隊伍中的角色。
func (s *Session) candidatesHere() []Candidate {
	if s.door == nil {
		return nil
	}
	key := placeKey(s.door.Name)
	inParty := map[string]bool{}
	for _, m := range s.Party {
		inParty[m.Name] = true
	}
	var out []Candidate
	for _, c := range s.Roster {
		if c.Place != key || (s.TownCode != "" && c.Town != s.TownCode) {
			continue
		}
		if inParty[c.Name] {
			continue
		}
		out = append(out, c)
	}
	return out
}

// recordRosterList 照原版 `sub_147DD:0x149C4` 把清單的槽號依序寫到共用緩衝 +0x20 起：
// 只有第 8、9 名落在 +0x28／+0x29，清單不夠長就不動（RE-159）。
func (s *Session) recordRosterList(list []Candidate) {
	for i := 8; i < 10 && i < len(list); i++ {
		s.RosterListTail[i-8] = list[i].Slot
	}
}

// StartRecruit 是按下 COMPANION：先問要不要搜尋硬碟，圖示列換成大拇指朝上與朝下。
func (s *Session) StartRecruit() {
	if s.Mode != ModeInside || s.Inside == nil {
		return
	}
	s.asking = true
	s.recruit = nil
	s.Inside.List = nil
	s.Inside.Dialog = s.uiLines(AskDiskRecord)
	s.Inside.DialogZh = nil
	s.setInsideIcons([]int{IconYes, IconNo})
}

// openList 進到候選人清單：對話框改畫清單，圖示列換成大拇指朝上與 U-TURN。
func (s *Session) openList() {
	s.asking = false
	s.recruit = &Recruit{Candidates: s.candidatesHere()}
	s.recordRosterList(s.recruit.Candidates)
	s.Inside.Dialog = nil
	s.Inside.DialogZh = nil
	s.Inside.List = s.recruit.View()
	s.setInsideIcons([]int{IconYes, IconUTurn})
}

// MoveList 是清單上的方向鍵。不在清單裡就什麼都不做。
func (s *Session) MoveList(delta int) {
	if s.recruit == nil {
		return
	}
	s.recruit.Move(delta)
	s.Inside.List = s.recruit.View()
}

// joinSelected 把反白那一位加進隊伍，成為新的發言者，畫面回到建築內的一般狀態。
func (s *Session) joinSelected() {
	r := s.recruit
	if r == nil || len(r.Candidates) == 0 || len(s.Party) >= MaxParty {
		s.cancelRecruit()
		return
	}
	c := r.Candidates[r.Cursor]
	for i := range s.Party {
		s.Party[i].Speaker = false
	}
	s.Party = append(s.Party, c.member())
	s.syncParty()
	s.recruit, s.asking = nil, false
	s.Inside.List = nil
	s.greeted = true
	s.Inside.Dialog = s.greeting()
	s.Inside.DialogZh = nil
	s.setInsideIcons(InteriorIcons(s.door.IconFlags, len(s.Party)))
}

// cancelRecruit 放棄招募，回到建築內的一般圖示列。
func (s *Session) cancelRecruit() {
	s.recruit, s.asking = nil, false
	if s.Inside == nil {
		return
	}
	s.Inside.List = nil
	s.Inside.Dialog = s.greeting()
	s.Inside.DialogZh = nil
	s.setInsideIcons(InteriorIcons(s.door.IconFlags, len(s.Party)))
}

// Recruiting 回報現在是不是停在招募的兩個畫面之一。
func (s *Session) Recruiting() bool { return s.asking || s.recruit != nil }
