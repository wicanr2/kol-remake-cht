package ui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/kol-remake-cht/internal/font"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

// Outdoor 畫出城鎮的戶外畫面：紅磚底圖之上疊隊員頭像格、日期框、城鎮徽章、
// 12×6 格的地圖視窗與圖示列。地圖每一格的磚塊編號來自 RE-059 的格子陣列。
func Outdoor(dst *ebiten.Image, sprites Sprites, small *font.Font, out *town.Outdoor) error {
	bg, err := sprites.Get(town.Background.Archive, town.Background.Index)
	if err != nil {
		return err
	}
	draw(dst, bg, 0, 0)

	if err := drawPartySlots(dst, sprites, small, out.Party); err != nil {
		return err
	}

	drawBox(dst, small, town.DateBox, out.Date)

	// 徽章直接畫上去，沒有投影——原版在它右方與下方留的是紅磚底圖，量過。
	crest, err := sprites.Get(out.CrestArchive(), town.Crest.Index)
	if err != nil {
		return err
	}
	draw(dst, crest, town.Crest.X, town.Crest.Y)

	if err := drawMapWindow(dst, sprites, out); err != nil {
		return err
	}

	b := town.IconBar
	for i, index := range out.Icons {
		x := b.X + i*b.Step
		icon, err := sprites.Get("KOLMPICS", index)
		if err != nil {
			return err
		}
		draw(dst, icon, x, b.Y+1)
		c := town.ColourIdle
		if i == out.Selected {
			c = town.ColourSelected
		}
		frame(dst, x, b.Y, b.Width, b.Height, c)
	}
	return nil
}

// drawPartySlots 是旅店與戶外共用的隊員格：上格圖像、下格名字，整塊先填黑再畫。
func drawPartySlots(dst *ebiten.Image, sprites Sprites, small *font.Font,
	party []town.Member) error {
	for i, m := range party {
		s := town.PartySlot
		x := s.X + i*s.Step
		c := town.ColourIdle
		if m.Speaker {
			c = town.ColourSelected
		}
		h := s.PortraitHeight + s.PlateHeight - 1
		shadow(dst, x, s.Y, s.Width, h)
		fill(dst, x, s.Y, s.Width, h, town.ColourBoxFill)
		frame(dst, x, s.Y, s.Width, s.PortraitHeight, c)
		frame(dst, x, s.Y+s.PortraitHeight-1, s.Width, s.PlateHeight, c)
		portrait, err := sprites.Get("CHARPICS", m.Portrait)
		if err != nil {
			return err
		}
		draw(dst, portrait, s.PortraitX+i*s.Step, s.PortraitY)
		small.Draw(dst, s.NameX+i*s.Step, s.NameY, m.PlateName(),
			town.ColourText[0], town.ColourText[1], town.ColourText[2])
	}
	return nil
}

// drawMapWindow 畫地圖視窗：投影、白框，然後把 12×6 格磚塊由 atlas 貼進內容區。
func drawMapWindow(dst *ebiten.Image, sprites Sprites, out *town.Outdoor) error {
	win := town.MapWindow
	shadow(dst, win.X, win.Y, win.Width, win.Height)
	fill(dst, win.X, win.Y, win.Width, win.Height, town.ColourBoxEdge)

	atlas, err := sprites.Get(out.Archive(), town.TileAtlas.Index)
	if err != nil {
		return err
	}
	v := town.Viewport
	a := town.TileAtlas
	for row, line := range out.View() {
		for col, tile := range line {
			sx := (tile % a.Columns) * a.Side
			sy := (tile / a.Columns) * a.Side
			src := atlas.SubImage(image.Rect(sx, sy, sx+a.Side, sy+a.Side)).(*ebiten.Image)
			draw(dst, src, v.X+col*v.Tile, v.Y+row*v.Tile)
		}
	}
	drawPartyMarker(dst, out)
	return nil
}

// drawPartyMarker 把隊伍標記蓋在它固定佔的那一格上。覆蓋層已經含膨脹出來的黑色外框，
// 透明像素代表該處保留底下的磚塊。
func drawPartyMarker(dst *ebiten.Image, out *town.Outdoor) {
	overlay := out.MarkerOverlay()
	if overlay == nil {
		return
	}
	row, col := out.MarkerCell()
	v := town.Viewport
	x0 := v.X + col*v.Tile
	y0 := v.Y + row*v.Tile
	for y := 0; y < town.MarkerSide; y++ {
		for x := 0; x < town.MarkerSide; x++ {
			c := overlay[y*town.MarkerSide+x]
			if c.A == 0 {
				continue
			}
			dst.Set(x0+x, y0+y, c)
		}
	}
}

// OutdoorCJK 畫出中文模式的戶外畫面。版面是 remake 自訂的（`town.CJKOutdoor`），
// 素材以最近鄰放大兩倍，日期改用中文字模。這個模式不對拍。
func OutdoorCJK(dst *ebiten.Image, sprites Sprites, f *font.Font, t Translator,
	out *town.Outdoor) error {
	l := town.CJKOutdoor
	bg, err := sprites.Get(town.Background.Archive, town.Background.Index)
	if err != nil {
		return err
	}
	drawScaled(dst, bg, 0, 0, l.Scale)

	for i, m := range out.Party {
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

	fill(dst, l.DateX, l.DateY, l.DateW, l.DateH, town.ColourBoxEdge)
	fill(dst, l.DateX+1, l.DateY+1, l.DateW-2, l.DateH-2, town.ColourBoxFill)
	lines := out.DateZh
	if len(lines) == 0 {
		lines = out.Date
	}
	for i, line := range lines {
		f.Draw(dst, l.DateTextX, l.DateTextY+i*l.DateStep, t.T(line),
			town.ColourText[0], town.ColourText[1], town.ColourText[2])
	}

	crest, err := sprites.Get(out.CrestArchive(), town.Crest.Index)
	if err != nil {
		return err
	}
	drawScaled(dst, crest, l.CrestX, l.CrestY, l.Scale)

	fill(dst, l.MapX, l.MapY, l.MapW, l.MapH, town.ColourBoxEdge)
	atlas, err := sprites.Get(out.Archive(), town.TileAtlas.Index)
	if err != nil {
		return err
	}
	a := town.TileAtlas
	for row, line := range out.View() {
		for col, tile := range line {
			sx := (tile % a.Columns) * a.Side
			sy := (tile / a.Columns) * a.Side
			src := atlas.SubImage(image.Rect(sx, sy, sx+a.Side, sy+a.Side)).(*ebiten.Image)
			drawScaled(dst, src, l.ViewX+col*l.Tile, l.ViewY+row*l.Tile, l.Scale)
		}
	}
	if overlay := out.MarkerOverlay(); overlay != nil {
		row, col := out.MarkerCell()
		x0 := l.ViewX + col*l.Tile
		y0 := l.ViewY + row*l.Tile
		for y := 0; y < town.MarkerSide; y++ {
			for x := 0; x < town.MarkerSide; x++ {
				c := overlay[y*town.MarkerSide+x]
				if c.A == 0 {
					continue
				}
				for dy := 0; dy < l.Scale; dy++ {
					for dx := 0; dx < l.Scale; dx++ {
						dst.Set(x0+x*l.Scale+dx, y0+y*l.Scale+dy, c)
					}
				}
			}
		}
	}

	for i, index := range out.Icons {
		x := l.IconX + i*l.IconStep
		icon, err := sprites.Get("KOLMPICS", index)
		if err != nil {
			return err
		}
		drawScaled(dst, icon, x, l.IconY+l.Scale, l.Scale)
		c := town.ColourIdle
		if i == out.Selected {
			c = town.ColourSelected
		}
		frame(dst, x, l.IconY, l.IconW, l.IconH, c)
	}
	return nil
}

