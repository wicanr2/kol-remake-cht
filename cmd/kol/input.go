package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"image"

	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/book"
	"github.com/wicanr2/kol-remake-cht/internal/flow"
	"github.com/wicanr2/kol-remake-cht/internal/input"
	"github.com/wicanr2/kol-remake-cht/internal/town"
	"github.com/wicanr2/kol-remake-cht/internal/world"
)

// 這一份是輸入裝置那一半：把鍵盤與滑鼠讀成 token。可點區域怎麼算在
// `internal/input`，那邊不碰 Ebitengine 所以在無頭容器裡測得動。
//
// 鍵位照原版參考卡。掃描順序固定，同一幀按下多個鍵時產生的序列才可重現。
var keyTokens = []struct {
	key   ebiten.Key
	token string
}{
	{ebiten.KeyArrowUp, "U"}, {ebiten.KeyArrowDown, "D"},
	{ebiten.KeyArrowLeft, "L"}, {ebiten.KeyArrowRight, "R"},
	{ebiten.KeyEnter, "RET"}, {ebiten.KeyNumpadEnter, "RET"},
	{ebiten.KeyEscape, "ESC"},
	{ebiten.KeyBackspace, "BS"},
	{ebiten.KeySpace, "SPACE"},
	// 圖示列的白框。參考卡寫的是 `<` `>`，鍵盤上不按 Shift 就是逗號與句號。
	{ebiten.KeyComma, "<"}, {ebiten.KeyPeriod, ">"},
}

// letterKeys 是 A…Z。它們既是參考卡的斜走鍵也是名字與問句的字母，一律送大寫
// ——收據腳本用的就是大寫，打字的畫面由狀態機自己先攔截。
var letterKeys = [26]ebiten.Key{
	ebiten.KeyA, ebiten.KeyB, ebiten.KeyC, ebiten.KeyD, ebiten.KeyE,
	ebiten.KeyF, ebiten.KeyG, ebiten.KeyH, ebiten.KeyI, ebiten.KeyJ,
	ebiten.KeyK, ebiten.KeyL, ebiten.KeyM, ebiten.KeyN, ebiten.KeyO,
	ebiten.KeyP, ebiten.KeyQ, ebiten.KeyR, ebiten.KeyS, ebiten.KeyT,
	ebiten.KeyU, ebiten.KeyV, ebiten.KeyW, ebiten.KeyX, ebiten.KeyY,
	ebiten.KeyZ,
}

// digitKeys 是 0…9，主鍵盤與數字鍵盤各一組。訓練人數、賭場金額與目錄頁的項目
// 都吃數字。
var digitKeys = [10][2]ebiten.Key{
	{ebiten.KeyDigit0, ebiten.KeyNumpad0}, {ebiten.KeyDigit1, ebiten.KeyNumpad1},
	{ebiten.KeyDigit2, ebiten.KeyNumpad2}, {ebiten.KeyDigit3, ebiten.KeyNumpad3},
	{ebiten.KeyDigit4, ebiten.KeyNumpad4}, {ebiten.KeyDigit5, ebiten.KeyNumpad5},
	{ebiten.KeyDigit6, ebiten.KeyNumpad6}, {ebiten.KeyDigit7, ebiten.KeyNumpad7},
	{ebiten.KeyDigit8, ebiten.KeyNumpad8}, {ebiten.KeyDigit9, ebiten.KeyNumpad9},
}

// keyboardTokens 回傳這一幀按下的鍵。只認「剛按下」，不認持續按住——原版的
// 讀入迴圈也是一次一個鍵，按住不會連走。
func keyboardTokens(typingName ...bool) []string {
	var out []string
	for _, k := range keyTokens {
		if len(typingName) > 0 && typingName[0] && (k.key == ebiten.KeyArrowUp || k.key == ebiten.KeyArrowDown || k.key == ebiten.KeyArrowLeft || k.key == ebiten.KeyArrowRight) {
			continue
		}
		if inpututil.IsKeyJustPressed(k.key) {
			out = append(out, k.token)
		}
	}
	for i, k := range letterKeys {
		if inpututil.IsKeyJustPressed(k) {
			out = append(out, string(rune('A'+i)))
		}
	}
	for i, pair := range digitKeys {
		if inpututil.IsKeyJustPressed(pair[0]) || inpututil.IsKeyJustPressed(pair[1]) {
			out = append(out, string(rune('0'+i)))
		}
	}
	return out
}

// mouseTokens 讀這一幀的滑鼠。左鍵點中可點區域就送那一格的 token，點在別處
// 送 fallback；右鍵是 ESC，滾輪是上下。全部翻成既有的鍵盤 token，所以滑鼠與
// 鍵盤走的是同一條路徑。
func mouseTokens(targets []input.Target, fallback string) []string {
	var out []string
	if _, dy := ebiten.Wheel(); dy != 0 {
		if dy > 0 {
			out = append(out, "U")
		} else {
			out = append(out, "D")
		}
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		out = append(out, "ESC")
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		// CursorPosition 回的是 logical 座標，已經考慮視窗縮放，所以直接拿來
		// 比對版面常數。
		x, y := ebiten.CursorPosition()
		out = append(out, input.Hit(targets, x, y, fallback)...)
	}
	return out
}

// flowTargets 是目前畫面上可以點的地方。座標取自各畫面既有的版面常數，
// 沒有另外量一份——版面改了這裡就跟著改。
//
// 還沒列進來的畫面靠 fallback（點一下等於 Return）與滾輪操作，不會沒有反應。
func (g *game) flowTargets() []input.Target {
	if g.flow == nil {
		return nil
	}
	cjk := g.strings != nil
	switch g.flow.Scene {
	case flow.SceneMenu:
		x, y, step, w := book.CharacterOptionsList.X, book.CharacterOptionsList.Y,
			book.CharacterOptionsList.Step, 120
		if cjk {
			l := book.CJKLayout
			x, y, step, w = l.MenuX, l.MenuY, l.MenuStep, 240
		}
		out := make([]input.Target, 0, flow.MenuCount)
		for i := 0; i < flow.MenuCount; i++ {
			out = append(out, input.Row(x, y, w, step, i, g.flow.MenuCursor, "U", "D"))
		}
		return out
	case flow.SceneInn:
		if g.flow.Inn == nil {
			return nil
		}
		return iconTargets(len(g.flow.Inn.Icons), g.flow.Inn.Selected, cjk)
	case flow.SceneTown:
		// 走出城門之後畫面換成野外與戰鬥。結局那幾頁只吃 fallback。
		if g.endingReached() {
			return nil
		}
		if g.quest != nil && g.quest.state != nil {
			return battleTargets(g.quest.state, cjk)
		}
		if g.leftTown() {
			if g.play != nil && g.play.ShowingSheetOutside() {
				in := g.play.Inside
				return iconTargets(len(in.Icons), in.Selected, cjk)
			}
			return worldTargets(g.worldSession, cjk)
		}
		if g.flow.Town == nil {
			return nil
		}
		if g.flow.Town.Mode == town.ModeInside {
			in := g.flow.Town.Inside
			targets := iconTargets(len(in.Icons), in.Selected, cjk)
			if g.flow.Town.Recruiting() && in.List != nil {
				targets = append(targets, recruitTargets(in, cjk)...)
			}
			return targets
		}
		icons, selected := g.flow.Town.Outdoor.IconRow()
		return iconTargets(len(icons), selected, cjk)
	}
	return nil
}

// worldTargets 是野外那一層的圖示列。白框走 `<`／`>`、確認是 RET，與城鎮同一套，
// 所以直接沿用 `iconTargets`——圖示列的位置本來就是同一組版面常數。
//
// 野外的移動是方向鍵，不在這裡：點地圖要另外決定「點某一格等於往哪走幾步」，
// 那是一條有語意的規則，不是版面換算，沒有原版依據之前不補。
func worldTargets(s *world.Session, cjk bool) []input.Target {
	if s == nil || s.Screen == nil || len(s.Screen.Icons) == 0 {
		return nil
	}
	// 露營與問句那幾頁蓋在圖示列上，點下去不該還去搬白框。
	if s.Camping || s.Prompt != nil {
		return nil
	}
	return iconTargets(len(s.Screen.Icons), s.Screen.Selected, cjk)
}

// battleTargets 是戰鬥那一層。規劃段與戰利品段各有一列圖示，兩列都走 `<`／`>`
// 與 RET（`applyBattleToken`）。
//
// 選目標格那一層是滑鼠座標直接指格（SPEC-010 §3、`sub_24131:0x24332`），見
// fieldCellTargets；那一層的圖示列原版怎麼接滑鼠還沒證據，只收鍵盤。
func battleTargets(s *battle.State, cjk bool) []input.Target {
	if s == nil {
		return nil
	}
	switch s.Stage {
	case battle.StageSpoils:
		// 清單與圖示列都可點。清單是「Items Found:」或「Characters:」那一張，
		// 座標與 `ui.BattleSpoils` 畫的是同一組常數。
		var out []input.Target
		l := battle.SpoilList
		x, y, w, step := l.Col*town.CellWidth, l.Row,
			battle.SpoilBox.Width-(l.Col*town.CellWidth-battle.SpoilBox.X)-4, l.LineStep
		if cjk {
			x, y, w, step = x*book.CJKScale, y*book.CJKScale,
				w*book.CJKScale, step*book.CJKScale
		}
		for i := range s.SpoilLines() {
			row := input.Row(x, y, w, step, i, s.Spoil.Cursor, "U", "D")
			// 白框可能停在 U-TURN，點某一列仍應確認那一列。
			row.Tokens = row.Tokens[:len(row.Tokens)-1]
			for n := 0; n < s.Spoil.Icon; n++ {
				row.Tokens = append(row.Tokens, "<")
			}
			row.Tokens = append(row.Tokens, "RET")
			out = append(out, row)
		}
		return append(out, iconTargets(len(battle.SpoilIcons), s.Spoil.Icon, cjk)...)
	case battle.StageField:
		if s.Plan.Prompt == battle.PromptTarget {
			return fieldCellTargets(cjk)
		}
		if s.Plan.Prompt == battle.PromptSpell || s.Plan.Prompt == battle.PromptAffect {
			count := len(s.Plan.List)
			if count == 0 && s.Plan.Prompt == battle.PromptAffect {
				count = len(battle.AffectRecords())
			}
			if count > battle.SpellListRows.Visible {
				count = battle.SpellListRows.Visible
			}
			l := battle.SpellListRows
			x, y, w, step := l.Col*town.CellWidth, l.Row,
				battle.SpellListBox.Width-(l.Col*town.CellWidth-battle.SpellListBox.X)-4,
				l.LineStep
			if cjk {
				x, y, w, step = x*book.CJKScale, y*book.CJKScale,
					w*book.CJKScale, step*book.CJKScale
			}
			out := make([]input.Target, 0, count)
			for i := 0; i < count; i++ {
				out = append(out, input.Row(x, y, w, step, i,
					s.Plan.ListPick, "U", "D"))
			}
			return out
		}
		if n := len(s.Plan.Bar); n > 0 {
			return iconTargets(n, s.Plan.Selected, cjk)
		}
	}
	return nil
}

// recruitTargets 的每列都送出正常移動及確認，包含清單已捲動的情形。
func recruitTargets(in *town.Inn, cjk bool) []input.Target {
	if in == nil || in.List == nil {
		return nil
	}
	b := town.DialogBox
	x, y, w, step := b.TextX, b.TextY, b.Width-12, b.LineStep
	if cjk {
		l := town.CJKInn
		x, y, w, step = l.DialogTextX, l.DialogTextY, l.DialogW-32, 16
	}
	var targets []input.Target
	for i := range in.List.Rows {
		if i >= town.ListRows {
			break
		}
		row := input.Row(x, y, w, step, i, in.List.Selected, "U", "D")
		// 游標可能停在取消圖示，點列仍應確認所點的角色。
		row.Tokens = row.Tokens[:len(row.Tokens)-1]
		for n := 0; n < in.Selected; n++ {
			row.Tokens = append(row.Tokens, "<")
		}
		row.Tokens = append(row.Tokens, "RET")
		targets = append(targets, row)
	}
	return targets
}

// iconTargets 是畫面底部那一列圖示的可點區域。旅店、建築內與戶外共用同一組
// 版面常數。
// fieldCellTargets 是選目標格時戰場上的 9×9 格。原版滑鼠 X 以 640 寬計，
// `(X − 0x130) ÷ 0x20`、`(Y − 8) ÷ 0x10` 換成格，且座標必須**嚴格**落在
// X 0x130…0x250、Y 8…0x98 之間（`0x2433C`…`0x24362`）。換成 320×200 畫布就是
// x 153…295、y 9…151：最左一欄與最上一列的第一個像素不算。點下去送 `CELL:列,欄`，
// 由 applyBattleToken 換成相對位移交給 TargetCursor.Move（mouse 為真）；
// 點在游標本身那一格位移為零，原版整段跳過（`0x24411`）。
func fieldCellTargets(cjk bool) []input.Target {
	v := battle.FieldView
	valid := image.Rect(v.X+1, v.Y+1, v.X+battle.ViewCols*battle.TileSide, v.Y+battle.ViewRows*battle.TileSide)
	out := make([]input.Target, 0, battle.ViewRows*battle.ViewCols)
	for r := 0; r < battle.ViewRows; r++ {
		for c := 0; c < battle.ViewCols; c++ {
			cell := image.Rect(v.X+c*battle.TileSide, v.Y+r*battle.TileSide,
				v.X+(c+1)*battle.TileSide, v.Y+(r+1)*battle.TileSide).Intersect(valid)
			if cjk {
				cell = image.Rect(cell.Min.X*book.CJKScale, cell.Min.Y*book.CJKScale,
					cell.Max.X*book.CJKScale, cell.Max.Y*book.CJKScale)
			}
			out = append(out, input.Target{Rect: cell, Tokens: []string{fmt.Sprintf("CELL:%d,%d", r, c)}})
		}
	}
	// 落在戰場外的點擊原版整段略過（位移為零），不能掉進 fallback 的 RET 而確定目標。
	return append(out, input.Target{Rect: image.Rect(-1<<20, -1<<20, 1<<20, 1<<20), Tokens: []string{}})
}

func iconTargets(count, cursor int, cjk bool) []input.Target {
	b := town.IconBar
	x, y, step, w, h := b.X, b.Y, b.Step, b.Width, b.Height
	if cjk {
		l := town.CJKInn
		x, y, step, w, h = l.IconX, l.IconY, l.IconStep, l.IconW, l.IconH
	}
	out := make([]input.Target, 0, count)
	for i := 0; i < count; i++ {
		out = append(out, input.Cell(x, y, w, h, step, i, cursor))
	}
	return out
}
