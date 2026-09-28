// Package flow 把已經各自對拍過的畫面串成一條可玩的路徑。
//
// 它只描述「按了什麼鍵、畫面走到哪」，不碰繪製也不碰輸入裝置，因此可以單獨測試、
// 可以用腳本重播。每個畫面自己的狀態機（charcreate.State、roster.DeleteFlow、
// town.Session）維持原樣，flow 只負責在它們之間轉場。
//
// 轉場依原版手冊與參考卡：標題序列按任意鍵離開 → 書本封面 → 目錄頁按項目編號 →
// 角色選項用反白條加 Return → 目錄頁 6 走角色選擇清單 → 旅店 → DOOR → 戶外地圖。
package flow

import (
	"github.com/wicanr2/kol-remake-cht/internal/charcreate"
	"github.com/wicanr2/kol-remake-cht/internal/roster"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

// Scene 是目前的畫面。名稱與 cmd/kol 的場景名相同，繪製端直接照著分派。
type Scene string

const (
	SceneTitle    Scene = "title"
	SceneCover    Scene = "cover"
	SceneContents Scene = "contents"
	SceneMenu     Scene = "menu"
	SceneRoster   Scene = "roster"
	SceneNewChar  Scene = "newchar"
	SceneDelete   Scene = "delete"
	SceneHarbor   Scene = "harbor"
	SceneSelect   Scene = "select"
	SceneInn      Scene = "inn"
	SceneTown     Scene = "town"
)

// 角色選項頁的項目順序，與 `book.CharacterOptionsItems` 相同。
const (
	MenuList = iota
	MenuCreate
	MenuDelete
	MenuEditPicture
	MenuEditShield
	MenuHarbor
	MenuCount
)

// Flow 是一次遊玩。各畫面的狀態由呼叫端建好後交進來。
type Flow struct {
	CommitCreate func(*charcreate.State) error
	Scene        Scene

	Roster *roster.Roster
	Create *charcreate.State
	Remove *roster.DeleteFlow
	Inn    *town.Inn
	Town   *town.Session

	// MenuCursor 是角色選項頁的反白條，SelectCursor 是角色選擇清單的游標。
	MenuCursor   int
	SelectCursor int

	// Unimplemented 記下玩家選到但還沒實作的功能，讓測試與收據看得見。
	Unimplemented []string
}

// New 從標題序列開始。
func New(r *roster.Roster, create *charcreate.State, remove *roster.DeleteFlow,
	inn *town.Inn, session *town.Session) *Flow {
	return &Flow{Scene: SceneTitle, Roster: r, Create: create, Remove: remove,
		Inn: inn, Town: session}
}

// Key 送進一個操作 token，回傳畫面是否可能改變。token 對應原版鍵位：
// 方向鍵 `U D L R`、斜走 `A S Z X`、`<` `>`、`RET`、`ESC`、`SPACE`、
// 數字鍵 `1`…`6`，以及輸入名字用的單一字母。
func (f *Flow) Key(token string) bool {
	switch f.Scene {
	case SceneTitle:
		f.Scene = SceneCover
	case SceneCover:
		f.Scene = SceneContents
	case SceneContents:
		return f.contents(token)
	case SceneMenu:
		return f.menu(token)
	case SceneRoster, SceneHarbor:
		f.Scene = SceneMenu
	case SceneNewChar:
		return f.newchar(token)
	case SceneDelete:
		return f.removeChar(token)
	case SceneSelect:
		return f.selectLeader(token)
	case SceneInn:
		return f.innKey(token)
	case SceneTown:
		return f.townKey(token)
	}
	return true
}

func (f *Flow) contents(token string) bool {
	switch token {
	case "5":
		f.Scene = SceneMenu
		f.MenuCursor = 0
	case "6":
		if len(f.Roster.Occupied()) == 0 {
			f.note("目錄頁 6：名冊是空的，原版會走另一條路徑")
			return false
		}
		f.Scene = SceneSelect
		f.SelectCursor = 0
	default:
		f.note("目錄頁項目 " + token)
		return false
	}
	return true
}

func (f *Flow) menu(token string) bool {
	switch token {
	case "U":
		f.MenuCursor = (f.MenuCursor - 1 + MenuCount) % MenuCount
	case "D":
		f.MenuCursor = (f.MenuCursor + 1) % MenuCount
	case "ESC":
		f.Scene = SceneContents
	case "RET":
		switch f.MenuCursor {
		case MenuList:
			f.Scene = SceneRoster
		case MenuCreate:
			if f.Roster.Full() {
				f.note("名冊已滿，最多十六名角色")
				return false
			}
			f.Create.Restart()
			f.Scene = SceneNewChar
		case MenuDelete:
			f.Scene = SceneDelete
		case MenuHarbor:
			f.Scene = SceneHarbor
		default:
			// EDIT PICTURE／EDIT SHIELD 的點陣圖編輯器還沒實作：原版行為見 SPEC-036，是否實作待使用者決定。
			f.note("角色選項 " + []string{"", "", "", "EDIT PICTURE", "EDIT SHIELD"}[f.MenuCursor])
			return false
		}
	default:
		return false
	}
	return true
}

func (f *Flow) newchar(token string) bool {
	if f.Create.Stage == charcreate.StageName {
		if len(token) == 1 {
			f.Create.TypeRune(rune(token[0]))
			return true
		}
		if token == "SPACE" {
			f.Create.TypeRune(' ')
			return true
		}
	}
	if f.Create.Stage == charcreate.StageDone && token != "RET" {
		return false
	}
	switch token {
	case "ESC":
		f.Create.Cancel()
		if f.Create.Stage == charcreate.StageAbandoned {
			f.Scene = SceneMenu
		}
		return true
	case "RET":
		f.Create.Accept()
	case "U":
		f.Create.MoveCursor(-1)
		f.Create.MovePortrait(-1)
	case "D":
		f.Create.MoveCursor(1)
		f.Create.MovePortrait(1)
	case "BS":
		f.Create.Backspace()
	default:
		if len(token) == 1 {
			f.Create.TypeRune(rune(token[0]))
		}
	}
	if f.Create.Stage == charcreate.StageDone && f.CommitCreate != nil {
		if err := f.CommitCreate(f.Create); err != nil {
			f.note("角色保存失敗：" + err.Error())
			return false
		}
		f.Scene = SceneMenu
	}
	return true
}

func (f *Flow) removeChar(token string) bool {
	switch token {
	case "ESC":
		f.Scene = SceneMenu
	case "RET":
		f.Remove.Accept()
	case "U":
		f.Remove.MoveCursor(-1)
	case "D":
		f.Remove.MoveCursor(1)
	case "y", "Y":
		f.Remove.Answer(true)
	case "n", "N":
		f.Remove.Answer(false)
	default:
		return false
	}
	return true
}

func (f *Flow) selectLeader(token string) bool {
	slots := f.Roster.Occupied()
	switch token {
	case "ESC":
		f.Scene = SceneContents
	case "U":
		if f.SelectCursor > 0 {
			f.SelectCursor--
		}
	case "D":
		if f.SelectCursor < len(slots)-1 {
			f.SelectCursor++
		}
	case "RET":
		if len(slots) == 0 || f.SelectCursor >= len(slots) ||
			!f.Town.StartLeader(slots[f.SelectCursor].Name) {
			f.note("選角：角色所在地尚未載入或資料不一致")
			return false
		}
		f.Inn = f.Town.Inside
		f.Scene = SceneTown
	default:
		return false
	}
	return true
}

func (f *Flow) innKey(token string) bool {
	f.Scene = SceneTown
	return f.townKey(token)
}

// townKey 把 token 交給城鎮那一層。鍵位由 `town.Session.Key` 統一定義——
// flow 只負責轉發，不再自己維護一份 switch。
func (f *Flow) townKey(token string) bool {
	changed := f.Town.Key(token)
	if f.Town.ConsumeQuit() {
		f.Scene = SceneContents
		return true
	}
	return changed
}

func (f *Flow) note(what string) {
	for _, seen := range f.Unimplemented {
		if seen == what {
			return
		}
	}
	f.Unimplemented = append(f.Unimplemented, what)
}
