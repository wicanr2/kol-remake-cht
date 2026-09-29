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
	"strconv"
	"strings"

	"github.com/wicanr2/kol-remake-cht/internal/bitmapedit"
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
	// SceneEditPick 是 EDIT PICTURE／EDIT SHIELD 的選角清單（與 SceneSelect 同一張畫面，
	// RE-052）；SceneEditor 是點陣圖編輯器本身（SPEC-036）。
	SceneEditPick Scene = "editpick"
	SceneEditor   Scene = "editor"
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

	// EditShield 為真表示這一次是 EDIT SHIELD；EditCursor 是選角清單的反白列。
	EditShield bool
	EditCursor int
	// Editor 是編輯器的狀態；EditSession 是跨越多次進出仍留著的滑鼠全域。
	Editor      *bitmapedit.Editor
	EditSession bitmapedit.Session
	// LoadBitmap 取某一名角色目前的圖（改過的那一張，沒改過就是模板或預設盾牌）；
	// SaveBitmap 在答 YES 時把改好的圖寫回角色並存檔。兩者由呼叫端接到存檔。
	LoadBitmap func(slot roster.Slot, shield bool) (bitmapedit.Bitmap, error)
	SaveBitmap func(slot roster.Slot, shield bool, b bitmapedit.Bitmap) error
	// EditError 是最近一次載入或存檔失敗的原因。
	EditError error
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
	case SceneEditPick:
		return f.editPick(token)
	case SceneEditor:
		return f.editorKey(token)
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
		case MenuEditPicture, MenuEditShield:
			// 與 DELETE CHARACTER 同一張選角清單，反白從第 0 列起（RE-052）。
			if len(f.Roster.Occupied()) == 0 {
				f.note("EDIT PICTURE／EDIT SHIELD：名冊是空的（原版印 BOOKDATA 0x39）")
				return false
			}
			f.EditShield = f.MenuCursor == MenuEditShield
			f.EditCursor = 0
			f.Scene = SceneEditPick
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

// editPick 是編輯器的選角清單：上下移動不繞圈，Return 進編輯器，ESC 回角色選項且
// 反白回到第 0 項（RE-052）。
func (f *Flow) editPick(token string) bool {
	slots := f.Roster.Occupied()
	switch token {
	case "ESC":
		f.Scene = SceneMenu
		f.MenuCursor = 0
	case "U":
		if f.EditCursor > 0 {
			f.EditCursor--
		}
	case "D":
		if f.EditCursor < len(slots)-1 {
			f.EditCursor++
		}
	case "RET":
		if f.EditCursor >= len(slots) || f.LoadBitmap == nil {
			return false
		}
		b, err := f.LoadBitmap(slots[f.EditCursor], f.EditShield)
		if err != nil {
			f.EditError = err
			return false
		}
		f.Editor = bitmapedit.New(b, &f.EditSession)
		f.Scene = SceneEditor
	default:
		return false
	}
	return true
}

// editorTokens 是編輯器的按鍵 token。方向鍵沿用 `U D L R`，所以字母鍵在這個畫面送
// **小寫**（原版大小寫都認）；`<`／`>` 與逗號、句號同義；數字鍵盤 5 是 `5`。
var editorTokens = map[string]int{
	"U": bitmapedit.KeyUp, "D": bitmapedit.KeyDown, "L": bitmapedit.KeyLeft, "R": bitmapedit.KeyRight,
	"HOME": bitmapedit.KeyHome, "PGUP": bitmapedit.KeyPgUp, "END": bitmapedit.KeyEnd, "PGDN": bitmapedit.KeyPgDn,
	"5": bitmapedit.KeyCenter, "RET": bitmapedit.KeyEnter, "ESC": bitmapedit.KeyEsc,
	"<": '<', ">": '>', ",": ',', ".": '.',
	"d": 'd', "u": 'u', "x": 'x', "c": 'c', "y": 'y', "n": 'n',
	"Y": 'Y', "N": 'N',
}

// editorKey 把一個 token 交給編輯器。`CLICK:x,y` 是滑鼠左鍵落在 320 × 200 的 (x, y)，
// 按住拖曳時每到一個新位置送一次。離開之後答 YES 就寫回，接著回到選角清單
// （`_main` 在 `sub_1293E` 之後跳回清單那一段）。
func (f *Flow) editorKey(token string) bool {
	e := f.Editor
	if e == nil {
		f.Scene = SceneEditPick
		return true
	}
	if rest, ok := strings.CutPrefix(token, "CLICK:"); ok {
		xs, ys, found := strings.Cut(rest, ",")
		x, errX := strconv.Atoi(xs)
		y, errY := strconv.Atoi(ys)
		if !found || errX != nil || errY != nil {
			return false
		}
		e.Click(x, y)
	} else if k, ok := editorTokens[token]; ok {
		e.Key(k)
	} else {
		return false
	}
	if e.Done {
		if e.Save && f.SaveBitmap != nil {
			slots := f.Roster.Occupied()
			if f.EditCursor < len(slots) {
				f.EditError = f.SaveBitmap(slots[f.EditCursor], f.EditShield, e.Pix)
			}
		}
		f.Editor = nil
		f.Scene = SceneEditPick
	}
	return true
}
