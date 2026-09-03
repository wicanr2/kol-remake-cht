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
}

// InnList 是對話框改畫成清單時的內容。Rows 是**視窗內**那幾列。
type InnList struct {
	Rows     []string
	Selected int
}

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
	return &InnList{Rows: rows, Selected: r.Cursor - r.Top}
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
	s.Party = append(s.Party, Member{Name: c.Name, Portrait: c.Portrait,
		Female: c.Female, Wounds: c.Wounds, Gold: c.Gold, Place: c.Place,
		TownCode: c.Town, BuildingID: c.BuildingID, Speaker: true})
	s.Inside.Party = s.Party
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
