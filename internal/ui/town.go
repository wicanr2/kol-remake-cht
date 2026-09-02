package ui

import (
	"github.com/hajimehoshi/ebiten/v2"

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

	pic, err := sprites.Get(in.Portrait.Archive, in.Portrait.Index)
	if err != nil {
		return err
	}
	w, h := pic.Bounds().Dx(), pic.Bounds().Dy()
	shadow(dst, town.Portrait.X, town.Portrait.Y, w, h)
	draw(dst, pic, town.Portrait.X, town.Portrait.Y)

	drawBox(dst, large, town.TitleBox, in.Title)
	drawBox(dst, large, town.DialogBox, in.Dialog)

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
		f.Draw(dst, x, b.TextY+i*b.LineStep, line,
			town.ColourText[0], town.ColourText[1], town.ColourText[2])
	}
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

	for i, m := range in.Party {
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

	pic, err := sprites.Get(in.Portrait.Archive, in.Portrait.Index)
	if err != nil {
		return err
	}
	drawScaled(dst, pic, l.PicX, l.PicY, l.Scale)

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
	for i, line := range lines {
		f.Draw(dst, l.DialogTextX, l.DialogTextY+i*l.DialogStep, line,
			town.ColourText[0], town.ColourText[1], town.ColourText[2])
	}

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
