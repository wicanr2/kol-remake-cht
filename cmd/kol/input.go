package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/wicanr2/kol-remake-cht/internal/book"
	"github.com/wicanr2/kol-remake-cht/internal/flow"
	"github.com/wicanr2/kol-remake-cht/internal/input"
	"github.com/wicanr2/kol-remake-cht/internal/town"
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
		// 走出城門之後畫面換成野外與戰鬥，那兩層的可點區域還沒做，先只吃
		// fallback 與滾輪。
		if g.leftTown() || g.endingReached() || g.flow.Town == nil {
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
