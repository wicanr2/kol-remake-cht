package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"strings"

	"github.com/wicanr2/kol-remake-cht/internal/font"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

// Inn 畫出旅店畫面：紅磚底圖之上疊隊員頭像格、對話對象肖像、標題框、對話框與圖示列。
// 每一塊的座標都由原版畫面機械量得（RE-054）。
func Inn(dst *ebiten.Image, sprites Sprites, small, large *font.Font, in *town.Inn) error {
	bg, err := sprites.Get(town.Background.Archive, town.Background.Index)
	if err != nil {
		return err
	}
	draw(dst, bg, 0, 0)

	// 上格放角色圖像、下格放名字，兩格共用外框：上格外框的下緣就是下格的上緣。
	// 整塊先填黑再畫圖像與名字——角色圖像只有 40 px 寬，左右各留一欄黑底，
	// 名牌沒有字的地方也是黑的，不是讓紅磚透出來。
	if err := drawPartySlots(dst, sprites, small, in.Party); err != nil {
		return err
	}

	// 法術編輯器蓋掉整個畫面，連隊員格都不留；底下那排圖示照畫（RE-128）。
	if in.SpellEdit != nil {
		if err := drawSpellEdit(dst, sprites, large, in.SpellEdit); err != nil {
			return err
		}
		return drawIconBar(dst, sprites, in)
	}

	// MIRROR 的角色資料頁蓋掉肖像、標題框與對話框那一整塊。
	if in.Sheet != nil {
		// 物品與隊伍那兩頁是**換頁**，不是蓋在資料頁上（RE-126／RE-127）：
		// 原版按下去之後整頁重畫，資料頁的屬性與武器框都不見了。
		switch {
		case in.ItemPage != nil || len(in.PartyPage) > 0:
			// 換頁，底下各自畫。
		case in.QuestList != nil || in.SpellList != nil:
			// 任務一覽與法術一覽都先呼叫 `sub_16239`，只重畫抬頭那一層，
			// 再蓋上徽章（RE-111）或兩欄各八格（RE-123）。
			if err := drawSheetHeader(dst, sprites, small, in.Sheet); err != nil {
				return err
			}
			if in.QuestList != nil {
				if err := drawQuestBadges(dst, sprites, in.QuestList); err != nil {
					return err
				}
			}
			if in.SpellList != nil {
				drawSpellCells(dst, small, in.SpellList)
			}
		default:
			if err := drawSheet(dst, sprites, small, in.SheetLabels, in.Sheet); err != nil {
				return err
			}
		}
		// 物品那一頁把整塊換成十七列的清單（RE-127）。
		// 隊伍立繪那一頁蓋掉整塊（RE-126）。
		if len(in.PartyPage) > 0 {
			if err := drawPartyPage(dst, sprites, in.PartyPage); err != nil {
				return err
			}
		}
		if in.ItemPage != nil {
			if err := drawItemPage(dst, sprites, small, large, in.ItemPage, in.Party); err != nil {
				return err
			}
		}
	} else {
		// 競技場沒有肖像可畫（`sub_2B454` 畫的是什麼還沒解），Archive 空著就跳過。
		if in.Portrait.Archive != "" {
			pic, err := sprites.Get(in.Portrait.Archive, in.Portrait.Index)
			if err != nil {
				return err
			}
			w, h := pic.Bounds().Dx(), pic.Bounds().Dy()
			shadow(dst, town.Portrait.X, town.Portrait.Y, w, h)
			draw(dst, pic, town.Portrait.X, town.Portrait.Y)
		}

		drawBox(dst, large, town.TitleBox, in.Title)
		switch {
		case in.List != nil:
			drawInnList(dst, large, in.List)
		default:
			b := town.DialogBox
			if in.DialogRow != 0 {
				b.TextY = in.DialogRow
			}
			drawBox(dst, large, b, in.Dialog)
			// 馬廄那一頁在同一個框裡多一匹馬與牠的名字（RE-120）。
			if in.Steed != nil {
				if err := drawSteed(dst, sprites, large, in.Steed); err != nil {
					return err
				}
			}
			// 商品明細畫在同一個對話框裡，座標由記錄自己帶（RE-075）。
			for _, seg := range in.Detail {
				cell(dst, large, seg.Col, seg.Row, seg.Text)
			}
		}
	}

	return drawIconBar(dst, sprites, in)
}

// drawIconBar 畫底下那排圖示，白框標出選中的那一格。
func drawIconBar(dst *ebiten.Image, sprites Sprites, in *town.Inn) error {
	b := town.IconBar
	for i, index := range in.Icons {
		x := b.X + i*b.Step
		icon, err := sprites.Get("KOLMPICS", index)
		if err != nil {
			return err
		}
		draw(dst, icon, x, b.Y+1)
		c := town.ColourIdle
		if i == in.Selected {
			c = town.ColourSelected
		}
		frame(dst, x, b.Y, b.Width, b.Height, c)
	}
	return nil
}

// drawSpellCells 畫資料頁上的法術一覽：一格一個白框，框裡印名字（RE-123）。
// 每一格都是 `sub_1BA9D` 畫的，那一支收尾一定呼叫 `sub_1B7E9`，所以格子右下
// 自帶四像素的黑色投影（`0x1BAF1`）。
func drawSpellCells(dst *ebiten.Image, f *font.Font, p *town.SpellPage) {
	for _, c := range p.Cells {
		fill(dst, c.FrameX, c.FrameY, c.FrameW, c.FrameH, town.ColourBoxFill)
		frame(dst, c.FrameX, c.FrameY, c.FrameW, c.FrameH, town.ColourBoxEdge)
		shadow(dst, c.FrameX, c.FrameY, c.FrameW, c.FrameH)
		cell(dst, f, c.TextCol, c.TextRow, c.Name)
	}
}

// drawSteed 畫馬廄明細頁那一匹：圖在 (112, 86)，名字在圖的右邊（RE-120）。
func drawSteed(dst *ebiten.Image, sprites Sprites, f *font.Font,
	s *town.SteedPic) error {
	pic, err := sprites.Get("STEEDS", s.Picture)
	if err != nil {
		return err
	}
	draw(dst, pic, town.StablePic.X, town.StablePic.Y)
	cell(dst, f, town.StablePic.NameCol, town.StablePic.NameRow, s.Name)
	return nil
}

// drawInnList 把對話框畫成可選清單：反白那一列是白字、其餘青色，
// 右側第 23 格畫上下兩個紅色箭頭（大字型的碼位 94／95）。原版兩個箭頭一直都在，
// 不隨捲動位置消失（RE-072）。
func drawInnList(dst *ebiten.Image, f *font.Font, list *town.InnList) {
	b := town.DialogBox
	shadow(dst, b.X, b.Y, b.Width, b.Height)
	fill(dst, b.X, b.Y, b.Width, b.Height, town.ColourBoxEdge)
	fill(dst, b.X+1, b.Y+1, b.Width-2, b.Height-2, town.ColourBoxFill)
	for i, line := range list.Rows {
		c := town.ColourListIdle
		if i == list.Selected {
			c = town.ColourText
		}
		f.Draw(dst, b.TextX, b.TextY+i*b.LineStep, line, c[0], c[1], c[2])
	}
	// 捲不動就沒有箭頭；捲得動時往下那一個畫在**最後一列**，不是固定第八列。
	if !list.Scrolls() {
		return
	}
	x := b.TextX + town.ListColumn*f.Advance()
	a := town.ColourListArrow
	f.Draw(dst, x, b.TextY, string(rune(town.ListArrowUp)), a[0], a[1], a[2])
	f.Draw(dst, x, b.TextY+(len(list.Rows)-1)*b.LineStep,
		string(rune(town.ListArrowDown)), a[0], a[1], a[2])
}

// drawBox 畫一個城鎮文字視窗：先投影、再白框黑底，最後逐行畫字。
func drawBox(dst *ebiten.Image, f *font.Font, b town.Box, lines []string) {
	shadow(dst, b.X, b.Y, b.Width, b.Height)
	fill(dst, b.X, b.Y, b.Width, b.Height, town.ColourBoxEdge)
	fill(dst, b.X+1, b.Y+1, b.Width-2, b.Height-2, town.ColourBoxFill)
	for i, line := range lines {
		x := b.TextX
		if b.Centred {
			x += (b.Cells - len([]rune(line))) / 2 * f.Advance()
		}
		y := b.TextY + i*b.LineStep
		text, caret := splitCaret(line)
		f.Draw(dst, x, y, text,
			town.ColourText[0], town.ColourText[1], town.ColourText[2])
		if caret >= 0 {
			fill(dst, x+caret*f.Advance(), y+town.CaretRow, f.Advance(), 1,
				town.ColourText)
		}
	}
}

// splitCaret 把游標哨符抽掉，回傳要畫的字串與游標所在的格號（沒有就是 −1）。
func splitCaret(line string) (string, int) {
	at := strings.Index(line, town.TextCaret)
	if at < 0 {
		return line, -1
	}
	return line[:at] + strings.Repeat(" ", len([]rune(line[at:]))),
		len([]rune(line[:at]))
}

// shadow 畫城鎮視窗共用的右下投影。
func shadow(dst *ebiten.Image, x, y, w, h int) {
	s := town.Shadow
	fill(dst, x+w, y+s.Offset, s.Thickness, h, town.ColourShadow)
	fill(dst, x+s.Offset, y+h, w, s.Thickness, town.ColourShadow)
}

// frame 畫一個 1 px 外框，不動內部。
func frame(dst *ebiten.Image, x, y, w, h int, c [3]uint8) {
	fill(dst, x, y, w, 1, c)
	fill(dst, x, y+h-1, w, 1, c)
	fill(dst, x, y, 1, h, c)
	fill(dst, x+w-1, y, 1, h, c)
}

// drawQuestBadges 把回報過的任務畫在角色資料頁上：一格是 AWDnPICS 的一張
// 24×21 圖，外框比圖大一圈，顏色是原版的 0x0A（`sub_1BA9D` 的最後一個參數）。
func drawQuestBadges(dst *ebiten.Image, sprites Sprites, list *town.QuestList) error {
	for _, b := range list.Badges {
		frame(dst, b.FrameX, b.FrameY, b.FrameW, b.FrameH, town.ColourQuestBadge)
		pic, err := sprites.Get(list.Archive, b.Site)
		if err != nil {
			return err
		}
		draw(dst, pic, b.X, b.Y)
	}
	return nil
}

// InnCJK 是旅店畫面的中文模式。素材與版面按兩倍放大，文字改用中文字格；
// 隊員名字是原版角色名，維持原文。這個模式不對拍。
func InnCJK(dst *ebiten.Image, sprites Sprites, f *font.Font, t Translator,
	in *town.Inn) error {
	l := town.CJKInn
	bg, err := sprites.Get(town.Background.Archive, town.Background.Index)
	if err != nil {
		return err
	}
	drawScaled(dst, bg, 0, 0, l.Scale)
	if err := drawCJKPartySlots(dst, sprites, f, in.Party); err != nil {
		return err
	}

	// 法術編輯器蓋掉整個畫面，連隊員格都不留；底下那排圖示照畫（RE-128，SPEC-035）。
	if in.SpellEdit != nil {
		if err := drawSpellEditCJK(dst, sprites, f, t, in.SpellEdit); err != nil {
			return err
		}
		return drawIconBarCJK(dst, sprites, in)
	}

	// 資料頁蓋掉整個建築畫面，與英文版同一個語意（`Town` 那一支的 `in.Sheet`）。
	// 物品與隊伍那兩頁是**換頁**不是蓋上去（RE-126／RE-127），所以各自畫。
	if in.Sheet != nil {
		if in.ItemPage == nil && len(in.PartyPage) == 0 && in.Sheet.Steed != nil {
			if err := drawSheetSteedCJK(dst, sprites, f, t, in.Sheet.Steed); err != nil {
				return err
			}
		}
		switch {
		case in.ItemPage != nil:
			if err := drawItemPageCJK(dst, sprites, f, t, in.ItemPage); err != nil {
				return err
			}
		case len(in.PartyPage) > 0:
			if err := drawPartyPageCJK(dst, sprites, in.PartyPage); err != nil {
				return err
			}
		case in.QuestList != nil || in.SpellList != nil:
			// 與英文版同一個分界：這兩頁只重畫 `sub_16239` 那一層。
			drawSheetHeaderCJK(dst, f, t, in.Sheet)
			if in.QuestList != nil {
				if err := drawQuestBadgesCJK(dst, sprites, in.QuestList); err != nil {
					return err
				}
			}
			if in.SpellList != nil {
				drawSpellCellsCJK(dst, f, in.SpellList)
			}
		default:
			drawSheetCJK(dst, f, t, in.SheetLabels, in.Sheet)
		}
		return drawIconBarCJK(dst, sprites, in)
	}

	// 資料頁只蓋建築右側；頂端隊員格在兩種語言模式都保持可見。

	// 競技場尚無已證實的店主肖像，與英文介面一致略過空素材。
	if in.Portrait.Archive != "" {
		pic, err := sprites.Get(in.Portrait.Archive, in.Portrait.Index)
		if err != nil {
			return err
		}
		drawScaled(dst, pic, l.PicX, l.PicY, l.Scale)
	}

	fill(dst, l.TitleX, l.TitleY, l.TitleW, l.TitleH, town.ColourBoxEdge)
	fill(dst, l.TitleX+1, l.TitleY+1, l.TitleW-2, l.TitleH-2, town.ColourBoxFill)
	for i, line := range in.Title {
		s := t.T(line)
		w, _ := f.Measure(s)
		f.Draw(dst, l.TitleX+(l.TitleW-w)/2, l.TitleTextY+i*l.TitleStep, s,
			town.ColourText[0], town.ColourText[1], town.ColourText[2])
	}

	fill(dst, l.DialogX, l.DialogY, l.DialogW, l.DialogH, town.ColourBoxEdge)
	fill(dst, l.DialogX+1, l.DialogY+1, l.DialogW-2, l.DialogH-2, town.ColourBoxFill)
	lines := in.DialogZh
	if len(lines) == 0 {
		lines = in.Dialog
	}
	if in.List != nil {
		// 八列沿用原版清單視窗；16 px 行距可完整放入既有 136 px 對話框。
		for i, line := range in.List.Rows {
			if i >= town.ListRows {
				break
			}
			c := town.ColourListIdle
			if i == in.List.Selected {
				c = town.ColourText
			}
			text := []rune(displayName(t, line))
			for len(text) > 0 {
				w, _ := f.Measure(string(text))
				if w <= l.DialogW-32 {
					break
				}
				text = text[:len(text)-1]
			}
			f.Draw(dst, l.DialogTextX, l.DialogTextY+i*16, string(text), c[0], c[1], c[2])
			// 選中那一列的注解接在名字後面（SPEC-034）；用未選取的顏色，放不下就截斷，
			// 不蓋到右側的捲動箭頭。
			if i == in.List.Selected && i < len(in.List.Notes) && in.List.Notes[i] != "" {
				nameW, _ := f.Measure(string(text))
				x := l.DialogTextX + nameW + 16
				note := []rune(in.List.Notes[i])
				for len(note) > 0 {
					w, _ := f.Measure(string(note))
					if x+w <= l.DialogX+l.DialogW-24 {
						break
					}
					note = note[:len(note)-1]
				}
				g := town.ColourListIdle
				f.Draw(dst, x, l.DialogTextY+i*16, string(note), g[0], g[1], g[2])
			}
		}
		if in.List.Scrolls() {
			a := town.ColourListArrow
			x := l.DialogX + l.DialogW - 20
			// 中文字模的 ▲▼（docs/i18n/ui-symbols-zh-Hant.json）；半形 v 在中文畫面像字母。
			f.Draw(dst, x, l.DialogTextY, "▲", a[0], a[1], a[2])
			f.Draw(dst, x, l.DialogTextY+7*16, "▼", a[0], a[1], a[2])
		}
		lines = nil
	}
	if in.Steed != nil {
		pic, err := sprites.Get("STEEDS", in.Steed.Picture)
		if err != nil {
			return err
		}
		drawScaled(dst, pic, town.StablePic.X*l.Scale, town.StablePic.Y*l.Scale, l.Scale)
		cellCJKAt(dst, f, town.StablePic.NameCol*town.CellWidth*l.Scale,
			town.StablePic.NameRow*l.Scale, displayName(t, in.Steed.Name))
	}
	for _, seg := range in.Detail {
		cellCJKAt(dst, f, seg.Col*town.CellWidth*l.Scale, seg.Row*l.Scale,
			displayName(t, t.T(seg.Text)))
	}
	dialogY := l.DialogTextY
	if in.DialogRow != 0 {
		dialogY = in.DialogRow * l.Scale
	}
	for i, line := range lines {
		x, y := l.DialogTextX, dialogY+i*l.DialogStep
		value, caret := splitCaret(line)
		f.Draw(dst, x, y, value,
			town.ColourText[0], town.ColourText[1], town.ColourText[2])
		if caret >= 0 {
			prefix := []rune(line)[:caret]
			width, _ := f.Measure(string(prefix))
			fill(dst, x+width, y+f.Height()-1, f.Advance(), 1, town.ColourText)
		}
	}

	return drawIconBarCJK(dst, sprites, in)
}

func drawCJKPartySlots(dst *ebiten.Image, sprites Sprites, f *font.Font,
	party []town.Member) error {
	l := town.CJKInn
	for i, m := range party {
		x := l.PartyX + i*l.PartyStep
		c := town.ColourIdle
		if m.Speaker {
			c = town.ColourSelected
		}
		h := l.PortraitBoxHeight + l.PlateHeight - 1
		fill(dst, x, l.PartyY, l.PartyWidth, h, town.ColourBoxFill)
		frame(dst, x, l.PartyY, l.PartyWidth, l.PortraitBoxHeight, c)
		frame(dst, x, l.PartyY+l.PortraitBoxHeight-1, l.PartyWidth, l.PlateHeight, c)
		portrait, err := sprites.Get("CHARPICS", m.Portrait)
		if err != nil {
			return err
		}
		drawScaled(dst, portrait, l.PortraitX+i*l.PartyStep, l.PortraitY, l.Scale)
		f.Draw(dst, l.NameX+i*l.PartyStep, l.NameY, m.PlateName(),
			town.ColourText[0], town.ColourText[1], town.ColourText[2])
	}

	return nil
}

// drawIconBarCJK 畫中文模式的圖示列。建築畫面與資料頁共用同一條。
func drawIconBarCJK(dst *ebiten.Image, sprites Sprites, in *town.Inn) error {
	l := town.CJKInn
	for i, index := range in.Icons {
		x := l.IconX + i*l.IconStep
		icon, err := sprites.Get("KOLMPICS", index)
		if err != nil {
			return err
		}
		drawScaled(dst, icon, x, l.IconY+l.Scale, l.Scale)
		c := town.ColourIdle
		if i == in.Selected {
			c = town.ColourSelected
		}
		frame(dst, x, l.IconY, l.IconW, l.IconH, c)
	}
	return nil
}
