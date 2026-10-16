package ui

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/kol-remake-cht/internal/font"
	"github.com/wicanr2/kol-remake-cht/internal/town"
	"github.com/wicanr2/kol-remake-cht/internal/world"
)

// World 畫出野外地圖畫面（RE-068）。與城鎮共用紅磚底圖、隊員格、日期框、
// 地圖視窗與圖示列；不同的是左側換成天候圖與「時段／地形」標籤，
// 而且隊伍是一個 2×2 像素的點，畫在視窗第 3 列第 5 欄那一格內、
// 偏移「次格座標 × 2」像素。
func World(dst *ebiten.Image, sprites Sprites, small, large *font.Font,
	party []town.Member, s *world.Screen) error {
	return WorldSession(dst, sprites, small, large, party, s, nil)
}

// WorldSession 同上，但多帶一個狀態機——站在城鎮入口上時地圖視窗換成外觀圖。
func WorldSession(dst *ebiten.Image, sprites Sprites, small, large *font.Font,
	party []town.Member, s *world.Screen, sess *world.Session) error {
	bg, err := sprites.Get(town.Background.Archive, town.Background.Index)
	if err != nil {
		return err
	}
	draw(dst, bg, 0, 0)

	if err := drawPartySlots(dst, sprites, small, idleParty(party)); err != nil {
		return err
	}
	drawBox(dst, small, town.DateBox, s.Date)

	if err := drawWorldMap(dst, sprites, s, sess); err != nil {
		return err
	}
	if sess != nil && sess.Prompt != nil {
		drawPrompt(dst, large, sess.Prompt)
	}
	drawWeather(dst, s)
	drawWorldLabel(dst, small, s, camping(sess))

	b := town.IconBar
	for i, index := range s.Icons {
		x := b.X + i*b.Step
		icon, err := sprites.Get("KOLMPICS", index)
		if err != nil {
			return err
		}
		draw(dst, icon, x, b.Y+1)
		frame(dst, x, b.Y, b.Width, b.Height, iconFrameColour(s, i))
	}
	return nil
}

// drawPrompt 畫蓋在地圖上的那一頁（RE-079／RE-112）：框是 `sub_2A437` 的
// (96,55)–(296,152)，敘述從 (104,65) 起、行距 8，最後空一列再印問句
// （記錄 0xE1 與 0xE3 開頭都自帶一個 `@`）。任務地點是八列，遭遇是四列。
func drawPrompt(dst *ebiten.Image, f *font.Font, q *world.Prompt) {
	b := world.PromptBox
	shadow(dst, b.X, b.Y, b.Width, b.Height)
	fill(dst, b.X, b.Y, b.Width, b.Height, town.ColourBoxEdge)
	fill(dst, b.X+1, b.Y+1, b.Width-2, b.Height-2, town.ColourBoxFill)
	t := world.PromptText
	for i, line := range q.Lines {
		f.Draw(dst, t.X, t.Y+i*t.Step, line,
			town.ColourText[0], town.ColourText[1], town.ColourText[2])
	}
	if q.Ask != "" {
		f.Draw(dst, t.X, t.Y+(len(q.Lines)+1)*t.Step, q.Ask,
			town.ColourText[0], town.ColourText[1], town.ColourText[2])
	}
}

// camping 回傳是不是停在換日那一張畫面。
func camping(sess *world.Session) bool { return sess != nil && sess.Camping }

// drawWorldMap 畫地圖視窗：投影、白框，12×6 格磚塊，最後蓋上隊伍那個點。
func drawWorldMap(dst *ebiten.Image, sprites Sprites, s *world.Screen,
	sess *world.Session) error {
	win := world.MapWindow
	shadow(dst, win.X, win.Y, win.Width, win.Height)
	fill(dst, win.X, win.Y, win.Width, win.Height, town.ColourBoxEdge)

	// 站在城鎮入口上時整個視窗換成該地點的 192×96 外觀圖，不畫磚塊也不畫那個點。
	// 換日那一張也是同一個位置，畫的是腳下地形那一張（`sub_297F2:0x298D6`）。
	if sess != nil {
		index := sess.TownPicture()
		if sess.Camping {
			index = s.TerrainCode
		}
		if index >= 0 {
			pic, err := sprites.Get(world.TownPictureArchive, index)
			if err != nil {
				return err
			}
			draw(dst, pic, world.Viewport.X, world.Viewport.Y)
			return nil
		}
	}

	atlas, err := sprites.Get("TOWN0", town.TileAtlas.Index)
	if err != nil {
		return err
	}
	v := world.Viewport
	a := town.TileAtlas
	for row, line := range s.View() {
		for col, tile := range line {
			sx := (tile % a.Columns) * a.Side
			sy := (tile / a.Columns) * a.Side
			src := atlas.SubImage(image.Rect(sx, sy, sx+a.Side, sy+a.Side)).(*ebiten.Image)
			draw(dst, src, v.X+col*v.Tile, v.Y+row*v.Tile)
		}
	}

	if !s.MarkerOff {
		mx, my := s.Travel.MarkerPixel(v.Tile)
		for y := 0; y < world.MarkerSide; y++ {
			for x := 0; x < world.MarkerSide; x++ {
				px, py := v.X+mx+x, v.Y+my+y
				r, g, b, a := dst.At(px, py).RGBA()
				dst.Set(px, py, world.InvertEGA(color.RGBA{
					uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}))
			}
		}
	}
	return nil
}

// drawWeather 畫左側的天候圖：白框加上合成後的 64×62 像素。
func drawWeather(dst *ebiten.Image, s *world.Screen) {
	box := world.WeatherBox
	shadow(dst, box.X, box.Y, box.Width, box.Height)
	fill(dst, box.X, box.Y, box.Width, box.Height, town.ColourBoxEdge)
	if s.Weather == nil {
		return
	}
	pixels := s.Weather.CompositeGroup(s.Group, s.Phase)
	if len(pixels) != world.WeatherWidth*world.WeatherHeight {
		return
	}
	p := world.WeatherPicture
	for y := 0; y < world.WeatherHeight; y++ {
		for x := 0; x < world.WeatherWidth; x++ {
			dst.Set(p.X+x, p.Y+y, pixels[y*world.WeatherWidth+x])
		}
	}
}

// drawWorldLabel 畫「時段／地形」的標籤框。兩行的起點是原版的固定欄位。
func drawWorldLabel(dst *ebiten.Image, f *font.Font, s *world.Screen, camp bool) {
	box := world.LabelBox
	shadow(dst, box.X, box.Y, box.Width, box.Height)
	fill(dst, box.X, box.Y, box.Width, box.Height, town.ColourBoxEdge)
	fill(dst, box.X+1, box.Y+1, box.Width-2, box.Height-2, town.ColourBoxFill)
	lines := []string{s.TimeLine(), s.TerrainLine()}
	// 換日那一張把第一行換成 CAMPING!（`sub_297F2:0x298A9`）。
	if camp {
		lines[0] = s.CampText
	}
	for i, line := range lines {
		pos := world.LabelLines[i]
		f.Draw(dst, pos.X, pos.Y, line,
			town.ColourText[0], town.ColourText[1], town.ColourText[2])
	}
}

// WorldCJK 畫出中文模式的野外畫面。版面是 remake 自訂的（`world.CJK`），
// 素材以最近鄰放大三倍，日期與標籤改用中文字模並在框內置中。這個模式不對拍。
func WorldCJK(dst *ebiten.Image, sprites Sprites, f *font.Font, t Translator,
	party []town.Member, s *world.Screen, sess *world.Session) error {
	l := world.CJK
	bg, err := sprites.Get(town.Background.Archive, town.Background.Index)
	if err != nil {
		return err
	}
	drawScaled(dst, bg, 0, 0, l.Scale)

	for i, m := range idleParty(party) {
		x := l.PartyX + i*l.PartyStep
		c := town.ColourIdle
		if m.Speaker {
			c = town.ColourSelected
		}
		h := l.PortraitBoxHeight + l.PlateHeight - 1
		fill(dst, x, l.PartyY, l.PartyWidth, h, town.ColourBoxFill)
		frame(dst, x, l.PartyY, l.PartyWidth, l.PortraitBoxHeight, c)
		frame(dst, x, l.PartyY+l.PortraitBoxHeight-1, l.PartyWidth, l.PlateHeight, c)
		portrait, err := memberPortrait(sprites, m.Portrait, m.Picture)
		if err != nil {
			return err
		}
		drawScaled(dst, portrait, l.PortraitX+i*l.PartyStep, l.PortraitY, l.Scale)
		f.Draw(dst, l.NameX+i*l.PartyStep, l.NameY, m.PlateName(),
			town.ColourText[0], town.ColourText[1], town.ColourText[2])
	}

	fill(dst, l.DateX, l.DateY, l.DateW, l.DateH, town.ColourBoxEdge)
	fill(dst, l.DateX+1, l.DateY+1, l.DateW-2, l.DateH-2, town.ColourBoxFill)
	dateLines := s.DateZh
	if s.Clock != nil {
		dateLines = s.Clock.DateLinesZh()
	} else if len(dateLines) == 0 {
		dateLines = s.Date
	}
	for i, line := range dateLines {
		f.Draw(dst, l.DateTextX, l.DateTextY+i*l.DateStep, t.T(line),
			town.ColourText[0], town.ColourText[1], town.ColourText[2])
	}

	fill(dst, l.WeatherX, l.WeatherY, l.WeatherW, l.WeatherH, town.ColourBoxEdge)
	if s.Weather != nil {
		pixels := s.Weather.CompositeGroup(s.Group, s.Phase)
		if len(pixels) != world.WeatherWidth*world.WeatherHeight {
			return fmt.Errorf("無效天候組 %d", s.Group)
		}
		weatherScale := 1
		if activeHDUI != nil && activeHDUI.dst == dst {
			weatherScale = 2
		}
		for y := 0; y < world.WeatherHeight; y++ {
			for x := 0; x < world.WeatherWidth; x++ {
				c := pixels[y*world.WeatherWidth+x]
				for dy := 0; dy < l.Scale*weatherScale; dy++ {
					for dx := 0; dx < l.Scale*weatherScale; dx++ {
						dst.Set((l.WeatherPicX+x*l.Scale)*weatherScale+dx, (l.WeatherPicY+y*l.Scale)*weatherScale+dy, c)
					}
				}
			}
		}
	}

	fill(dst, l.LabelX, l.LabelY, l.LabelW, l.LabelH, town.ColourBoxEdge)
	fill(dst, l.LabelX+1, l.LabelY+1, l.LabelW-2, l.LabelH-2, town.ColourBoxFill)
	for i, line := range []string{s.TimeLine(), s.TerrainLine()} {
		text := strings.TrimSpace(line)
		w, _ := f.Measure(text)
		x := l.LabelX + (l.LabelW-w)/2
		f.Draw(dst, x, l.LabelTextY+i*l.LabelStep, text,
			town.ColourText[0], town.ColourText[1], town.ColourText[2])
	}

	fill(dst, l.MapX, l.MapY, l.MapW, l.MapH, town.ColourBoxEdge)
	if sess != nil {
		if index := sess.TownPicture(); index >= 0 {
			pic, err := sprites.Get(world.TownPictureArchive, index)
			if err != nil {
				return err
			}
			drawScaled(dst, pic, l.ViewX, l.ViewY, l.Scale)
			if err := drawWorldPromptCJK(dst, f, t, sess); err != nil {
				return err
			}
			return drawWorldIconsCJK(dst, sprites, s)
		}
	}
	if drawn, err := drawHDWorldMap(dst, sprites, s); drawn || err != nil {
		if err != nil {
			return err
		}
		if err := drawWorldPromptCJK(dst, f, t, sess); err != nil {
			return err
		}
		return drawWorldIconsCJK(dst, sprites, s)
	}
	atlas, err := sprites.Get("TOWN0", town.TileAtlas.Index)
	if err != nil {
		return err
	}
	a := town.TileAtlas
	for row, line := range s.View() {
		for col, tile := range line {
			sx := (tile % a.Columns) * a.Side
			sy := (tile / a.Columns) * a.Side
			src := atlas.SubImage(image.Rect(sx, sy, sx+a.Side, sy+a.Side)).(*ebiten.Image)
			drawScaled(dst, src, l.ViewX+col*l.Tile, l.ViewY+row*l.Tile, l.Scale)
		}
	}
	if !s.MarkerOff {
		mx, my := s.Travel.MarkerPixel(l.Tile / l.Scale)
		markerScale := 1
		if activeHDUI != nil && activeHDUI.dst == dst {
			markerScale = 2
		}
		x0, y0 := (l.ViewX+mx*l.Scale)*markerScale, (l.ViewY+my*l.Scale)*markerScale
		for y := 0; y < l.MarkerSide*markerScale; y++ {
			for x := 0; x < l.MarkerSide*markerScale; x++ {
				px, py := x0+x, y0+y
				r, g, b, a := dst.At(px, py).RGBA()
				dst.Set(px, py, world.InvertEGA(color.RGBA{
					uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}))
			}
		}
	}
	if err := drawWorldPromptCJK(dst, f, t, sess); err != nil {
		return err
	}
	return drawWorldIconsCJK(dst, sprites, s)
}

// worldPromptParagraph 只把原始斷行接回顯示段落，不修改提示或資料。
func worldPromptParagraph(t Translator, lines []string) string {
	en := strings.Join(strings.Fields(strings.Join(lines, " ")), " ")
	if p, ok := t.(interface{ Paragraph(string) string }); ok {
		return p.Paragraph(en)
	}
	if t != nil {
		return t.T(en)
	}
	return en
}

// worldPromptCJKLines 使用原提示框的內縮與倍率，依實際字模重新換行。
// 正式 catalog 必須完整放入；拒絕溢出，不依賴畫布裁切吞掉敘述或問句。
func worldPromptCJKLines(f *font.Font, t Translator, q *world.Prompt) ([]string, int, error) {
	sc := world.CJK.Scale
	b, p := world.PromptBox, world.PromptText
	width := (b.X + b.Width - 8 - p.X) * sc
	body := rewrapCJK(f, []string{worldPromptParagraph(t, q.Lines)}, plainTranslator{}, width)
	if len(body) > 8 {
		return nil, 0, fmt.Errorf("世界提示敘述超出八列")
	}
	lines := append([]string(nil), body...)
	if q.Ask != "" {
		lines = append(lines, "")
		lines = append(lines, rewrapCJK(f, []string{worldPromptParagraph(t, []string{q.Ask})}, plainTranslator{}, width)...)
	}
	_, height := f.Measure("M")
	bottom := (b.Y + b.Height - 8) * sc
	for i, line := range lines {
		w, _ := f.Measure(line)
		if w > width || p.Y*sc+i*p.Step*sc+height > bottom {
			return nil, 0, fmt.Errorf("世界提示文字超出安全矩形")
		}
	}
	return lines, width, nil
}

type plainTranslator struct{}

func (plainTranslator) T(text string) string { return text }

func drawWorldPromptCJK(dst *ebiten.Image, f *font.Font, t Translator, sess *world.Session) error {
	if sess == nil || sess.Prompt == nil {
		return nil
	}
	lines, _, err := worldPromptCJKLines(f, t, sess.Prompt)
	if err != nil {
		return err
	}
	sc := world.CJK.Scale
	b, p := world.PromptBox, world.PromptText
	drawBattleBoxCJK(dst, b.X, b.Y, b.Width, b.Height, sc)
	for i, line := range lines {
		f.Draw(dst, p.X*sc, (p.Y+i*p.Step)*sc, line,
			town.ColourText[0], town.ColourText[1], town.ColourText[2])
	}
	return nil
}

// iconFrameColour 是野外圖示列的框色：白框是目前這一格，被選過又離開的是青色，
// 其餘是亮藍（RE-070）。
func iconFrameColour(s *world.Screen, i int) [3]uint8 {
	switch {
	case i == s.Selected:
		return town.ColourSelected
	case i < len(s.Vacated) && s.Vacated[i]:
		return world.ColourVacated
	}
	return town.ColourIdle
}

// drawWorldIconsCJK 是中文模式的圖示列，內容與英文模式相同，只是放大三倍。
func drawWorldIconsCJK(dst *ebiten.Image, sprites Sprites, s *world.Screen) error {
	l := world.CJK
	for i, index := range s.Icons {
		x := l.IconX + i*l.IconStep
		icon, err := sprites.Get("KOLMPICS", index)
		if err != nil {
			return err
		}
		drawScaled(dst, icon, x, l.IconY+l.Scale, l.Scale)
		frame(dst, x, l.IconY, l.IconW, l.IconH, iconFrameColour(s, i))
	}
	return nil
}
