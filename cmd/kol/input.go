package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"image"
	"strconv"

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

// keyboardTokens 在文字模式收作業系統字元事件；只有地圖移動使用長按節拍。
func keyboardTokens(typingName, typingSpell, repeatMovement bool) []string {
	var out []string
	for _, k := range keyTokens {
		if typingName && (k.key == ebiten.KeyArrowUp || k.key == ebiten.KeyArrowDown || k.key == ebiten.KeyArrowLeft || k.key == ebiten.KeyArrowRight) {
			// 法術名稱的Left與退格同義（RE-212）；建角文字模式保持既有處理。
			if typingSpell && k.key == ebiten.KeyArrowLeft && inpututil.IsKeyJustPressed(k.key) {
				out = append(out, "BS")
			}
			continue
		}
		if typingName && (k.key == ebiten.KeySpace || !typingSpell && (k.key == ebiten.KeyComma || k.key == ebiten.KeyPeriod)) {
			continue
		}
		pressed := inpututil.IsKeyJustPressed(k.key)
		if repeatMovement && (k.token == "U" || k.token == "D" || k.token == "L" || k.token == "R") {
			pressed = input.RepeatMovement(inpututil.KeyPressDuration(k.key))
		}
		if pressed {
			out = append(out, k.token)
		}
	}
	if typingName {
		for _, r := range ebiten.AppendInputChars(nil) {
			if r >= 32 && r <= 126 && (!typingSpell || r != ',' && r != '.' && r != '<' && r != '>') {
				if r == ' ' {
					out = append(out, "SPACE")
				} else {
					out = append(out, string(r))
				}
			}
		}
		return out
	}
	for i, k := range letterKeys {
		pressed := inpututil.IsKeyJustPressed(k)
		if repeatMovement && (k == ebiten.KeyA || k == ebiten.KeyS || k == ebiten.KeyZ || k == ebiten.KeyX) {
			pressed = input.RepeatMovement(inpututil.KeyPressDuration(k))
		}
		if pressed {
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
func (g *game) mouseTokens(targets []input.Target, fallback string) []string {
	var out []string
	if _, dy := ebiten.Wheel(); dy != 0 && g.wheelAllowed() {
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
		// 視窗縮放由Ebitengine處理；HD呈現的二倍座標換回既有文字畫布。
		x, y := g.cursorPosition()
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
				return insideTargets(in, cjk)
			}
			targets := worldTargets(g.worldSession, cjk)
			// 野外走路的輸入迴圈也收「點肖像＝按數字」（`sub_29004:0x29105`）。
			if ws := g.worldSession; ws != nil && g.play != nil && !ws.Camping && ws.Prompt == nil {
				targets = append(targets, portraitTargets(len(g.play.Party), cjk)...)
			}
			return targets
		}
		if g.flow.Town == nil {
			return nil
		}
		if g.flow.Town.Mode == town.ModeInside {
			in := g.flow.Town.Inside
			targets := insideTargets(in, cjk)
			if g.flow.Town.PortraitClicks() {
				targets = append(targets, portraitTargets(len(g.flow.Town.Party), cjk)...)
			}
			return targets
		}
		icons, selected := g.flow.Town.Outdoor.IconRow()
		targets := iconTargets(len(icons), selected, cjk)
		if g.flow.Town.PortraitClicks() {
			targets = append(targets, portraitTargets(len(g.flow.Town.Party), cjk)...)
		}
		return targets
	}
	return nil
}

// 隊伍肖像列的點擊（RE-167，`sub_1B100:0x1B451`…`0x1B4C0`）：左鍵落在畫面上緣 y < 50、
// x 從 16 起每 48 像素一位（原版滑鼠 x 以 640 寬計，`(x − 0x20) ÷ 0x60`），而且那一位在隊伍裡，
// 就回傳 `0x31 + 位置`，也就是按 1…6。只有 `word_2DF3E == 1` 的輸入迴圈收（城鎮戶外走路、
// 問誰開口、建築內一般圖示列、野外走路）；第一格左緣的半個像素（x = 0x20 整）不算。
const (
	portraitLeft   = 16
	portraitStep   = 48
	portraitBottom = 50
)

// portraitTargets 是隊伍前 count 位肖像的可點區域，點下去送 `1`…`6`。
func portraitTargets(count int, cjk bool) []input.Target {
	if count > town.MaxParty {
		count = town.MaxParty
	}
	out := make([]input.Target, 0, count)
	for i := 0; i < count; i++ {
		r := image.Rect(portraitLeft+i*portraitStep, 0, portraitLeft+(i+1)*portraitStep, portraitBottom)
		if cjk {
			r = image.Rect(r.Min.X*book.CJKScale, r.Min.Y*book.CJKScale, r.Max.X*book.CJKScale, r.Max.Y*book.CJKScale)
		}
		out = append(out, input.Target{Rect: r, Tokens: []string{strconv.Itoa(i + 1)}})
	}
	return out
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
	// 問句已有自己的 YES／NO 列；露營與通知仍只等確認。
	if s.Camping || s.Prompt != nil && s.Prompt.Kind == world.PromptNotice {
		return nil
	}
	return iconTargets(len(s.Screen.Icons), s.Screen.Selected, cjk)
}

// wheelAllowed 只在可捲清單與地圖收滾輪。等待頁不把動量當成確認。
func (g *game) wheelAllowed() bool {
	if g.flow == nil || len(g.notice) > 0 {
		return false
	}
	if g.flow.Scene != flow.SceneTown {
		return g.flow.Scene == flow.SceneMenu || g.flow.Scene == flow.SceneRoster || g.flow.Scene == flow.SceneSelect
	}
	if g.quest != nil && g.quest.state != nil {
		s := g.quest.state
		return s.Stage == battle.StageSpoils || s.Stage == battle.StageField && (s.Plan.Prompt == battle.PromptSpell || s.Plan.Prompt == battle.PromptAffect)
	}
	if g.play == nil {
		return false
	}
	if g.play.ShowingItemPage() {
		p := g.play.Inside.ItemPage
		return !p.Acting && len(p.Message) == 0 && len(p.DropMessage) == 0
	}
	if g.leftTown() && !g.play.ShowingSheetOutside() {
		return g.worldSession != nil && !g.worldSession.Camping && g.worldSession.Prompt == nil
	}
	in := g.play.Inside
	return g.play.Mode == town.ModeMap || g.play.Mode == town.ModeInside && in != nil && (in.List != nil || in.SpellEdit != nil && in.SpellEdit.Typing == nil)
}

func (g *game) walkingInput() bool {
	if g.flow == nil || g.flow.Scene != flow.SceneTown || g.play == nil || len(g.notice) > 0 || g.endingReached() || g.quest != nil && g.quest.state != nil {
		return false
	}
	if g.leftTown() {
		return !g.play.ShowingSheetOutside() && g.worldSession != nil && g.worldSession.Mode == world.ModeMap && !g.worldSession.Camping && g.worldSession.Prompt == nil
	}
	return g.play.Mode == town.ModeMap
}

func movementHeld() bool {
	for _, key := range []ebiten.Key{ebiten.KeyArrowUp, ebiten.KeyArrowDown, ebiten.KeyArrowLeft, ebiten.KeyArrowRight, ebiten.KeyA, ebiten.KeyS, ebiten.KeyZ, ebiten.KeyX} {
		if ebiten.IsKeyPressed(key) {
			return true
		}
	}
	return false
}

func insideTargets(in *town.Inn, cjk bool) []input.Target {
	if in == nil {
		return nil
	}
	out := iconTargets(len(in.Icons), in.Selected, cjk)
	if in.List != nil {
		out = append(out, recruitTargets(in, cjk)...)
	}
	if p := in.SpellEdit; p != nil && p.Typing == nil {
		if p.Menu != nil {
			x, y, w, step := town.SpellMenuGeometry(p, cjk)
			for i := range p.Menu.Rows {
				row := input.Row(x, y, w, step, i, p.Menu.Selected, "U", "D")
				out = append(out, confirmRow(row, in.Selected))
			}
		} else {
			for i, field := range p.Fields {
				scale, h := 1, 8
				if cjk {
					scale, h = book.CJKScale, 18
				}
				w := len([]rune(field.Text)) * town.CellWidth * scale
				if w < 24*scale {
					w = 24 * scale
				}
				row := input.Row(field.Col*town.CellWidth*scale, field.Row*scale, w, h, 0, 0, "U", "D")
				row.Tokens = input.Row(0, 0, 0, 0, i, p.Cursor, "U", "D").Tokens
				out = append(out, confirmRow(row, in.Selected))
			}
		}
	}
	if p := in.ItemPage; p != nil && !p.Acting && len(p.Message) == 0 && len(p.DropMessage) == 0 {
		l := town.ItemPageLayout
		scale := 1
		if cjk {
			scale = book.CJKScale
		}
		for i := range p.Rows {
			row := input.Row(l.ListCol*town.CellWidth*scale, l.ListRow*scale, 48*scale, l.ListStep*scale, i, p.Selected, "U", "D")
			out = append(out, confirmRow(row, in.Selected))
		}
	}
	return out
}

func confirmRow(row input.Target, selected int) input.Target {
	row.Tokens = row.Tokens[:len(row.Tokens)-1]
	for n := 0; n < selected; n++ {
		row.Tokens = append(row.Tokens, "<")
	}
	row.Tokens = append(row.Tokens, "RET")
	return row
}

// battleTargets 是戰鬥那一層。規劃段與戰利品段各有一列圖示，兩列都走 `<`／`>`
// 與 RET（`applyBattleToken`）。
//
// 選目標格那一層是滑鼠座標直接指格（SPEC-010 §3、`sub_24131:0x24332`），見
// fieldCellTargets；同一層的圖示列（YES、U-TURN）與其他層走同一組可點區域（RE-167）。
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
			// 選目標格那一層的圖示列（YES、U-TURN）：原版 `sub_24131` 用 `sub_1B100(6)` 取鍵，
			// 圖示列的點擊分支與其他層共用（RE-167）。圖示區在戰場下方、不與格重疊；
			// 格那一組最後有一個吃掉場外點擊的區域，所以圖示要排在它前面。
			targets := iconTargets(len(s.Plan.Bar), s.Plan.Selected, cjk)
			return append(targets, fieldCellTargets(cjk)...)
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
			l.Row = s.SpellListRow()
			x, y, w, step := l.Col*town.CellWidth, l.Row,
				battle.SpellListBox.Width-(l.Col*town.CellWidth-battle.SpellListBox.X)-4,
				l.LineStep
			if cjk {
				x, y, w, step = x*book.CJKScale, y*book.CJKScale,
					w*book.CJKScale, step*book.CJKScale
			}
			out := make([]input.Target, 0, count)
			for i := 0; i < count; i++ {
				row := input.Row(x, y, w, step, i, s.Plan.ListPick, "U", "D")
				row.Tokens = row.Tokens[:len(row.Tokens)-1]
				for n := 0; n < s.Plan.Selected; n++ {
					row.Tokens = append(row.Tokens, "<")
				}
				row.Tokens = append(row.Tokens, "RET")
				out = append(out, row)
			}
			return append(out, iconTargets(len(s.Plan.Bar), s.Plan.Selected, cjk)...)
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
