package ui

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/kol-remake-cht/internal/font"
	"github.com/wicanr2/kol-remake-cht/internal/town"
	"github.com/wicanr2/kol-remake-cht/internal/world"
)

// World 畫出野外地圖畫面（RE-068）。與城鎮共用紅磚底圖、隊員格、日期框、
// 地圖視窗與圖示列；不同的是左側換成天候圖與「時段／地形」標籤，
// 而且隊伍是一個 2×2 像素的點，畫在視窗第 3 列第 5 欄那一格內、
// 偏移「次格座標 × 2」像素。
func World(dst *ebiten.Image, sprites Sprites, small *font.Font,
	party []town.Member, s *world.Screen) error {
	return WorldSession(dst, sprites, small, party, s, nil)
}

// WorldSession 同上，但多帶一個狀態機——站在城鎮入口上時地圖視窗換成外觀圖。
func WorldSession(dst *ebiten.Image, sprites Sprites, small *font.Font,
	party []town.Member, s *world.Screen, sess *world.Session) error {
	bg, err := sprites.Get(town.Background.Archive, town.Background.Index)
	if err != nil {
		return err
	}
	draw(dst, bg, 0, 0)

	if err := drawPartySlots(dst, sprites, small, party); err != nil {
		return err
	}
	drawBox(dst, small, town.DateBox, s.Date)

	if err := drawWorldMap(dst, sprites, s, sess); err != nil {
		return err
	}
	drawWeather(dst, s)
	drawWorldLabel(dst, small, s)

	b := town.IconBar
	for i, index := range s.Icons {
		x := b.X + i*b.Step
		icon, err := sprites.Get("KOLMPICS", index)
		if err != nil {
			return err
		}
		draw(dst, icon, x, b.Y+1)
		c := town.ColourIdle
		if i == s.Selected {
			c = town.ColourSelected
		}
		frame(dst, x, b.Y, b.Width, b.Height, c)
	}
	return nil
}

// drawWorldMap 畫地圖視窗：投影、白框，12×6 格磚塊，最後蓋上隊伍那個點。
func drawWorldMap(dst *ebiten.Image, sprites Sprites, s *world.Screen,
	sess *world.Session) error {
	win := world.MapWindow
	shadow(dst, win.X, win.Y, win.Width, win.Height)
	fill(dst, win.X, win.Y, win.Width, win.Height, town.ColourBoxEdge)

	// 站在城鎮入口上時整個視窗換成該地點的 192×96 外觀圖，不畫磚塊也不畫那個點。
	if sess != nil {
		if index := sess.TownPicture(); index >= 0 {
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
	pixels := s.Weather.Composite(s.Phase)
	p := world.WeatherPicture
	for y := 0; y < world.WeatherHeight; y++ {
		for x := 0; x < world.WeatherWidth; x++ {
			dst.Set(p.X+x, p.Y+y, pixels[y*world.WeatherWidth+x])
		}
	}
}

// drawWorldLabel 畫「時段／地形」的標籤框。兩行的起點是原版的固定欄位。
func drawWorldLabel(dst *ebiten.Image, f *font.Font, s *world.Screen) {
	box := world.LabelBox
	shadow(dst, box.X, box.Y, box.Width, box.Height)
	fill(dst, box.X, box.Y, box.Width, box.Height, town.ColourBoxEdge)
	fill(dst, box.X+1, box.Y+1, box.Width-2, box.Height-2, town.ColourBoxFill)
	lines := []string{s.TimeLine(), s.TerrainLine()}
	for i, line := range lines {
		pos := world.LabelLines[i]
		f.Draw(dst, pos.X, pos.Y, line,
			town.ColourText[0], town.ColourText[1], town.ColourText[2])
	}
}
