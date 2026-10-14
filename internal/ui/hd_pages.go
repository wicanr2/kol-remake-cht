package ui

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/kol-remake-cht/internal/bitmapedit"
	"github.com/wicanr2/kol-remake-cht/internal/book"
	"github.com/wicanr2/kol-remake-cht/internal/font"
)

type HDPageSource struct {
	Archive string
	Index   int
	RGB     string
}
type hdPages struct {
	book         *ebiten.Image
	portraits    map[string]*ebiten.Image
	books        map[string]int
	background   string
	icons        map[string]*ebiten.Image
	sources      []HDPageSource
	validated    bool
	editorImage  *ebiten.Image
	editorPixels bitmapedit.Bitmap
}
type hdNativeRender struct {
	h    *HD
	dst  *ebiten.Image
	book bool
	err  error
}

var activeHDUI *hdNativeRender

func (h *HD) AttachPages(book image.Image, portraits map[string]image.Image, books map[string]int, background string, sources []HDPageSource) {
	p := &hdPages{book: ebiten.NewImageFromImage(book), portraits: map[string]*ebiten.Image{}, books: books, background: background, icons: map[string]*ebiten.Image{}, sources: sources}
	for key, pic := range portraits {
		p.portraits[key] = ebiten.NewImageFromImage(pic)
	}
	h.pages = p
}

// ValidatePageSources 在繪圖環境就緒後，先核對實際PNG圖號及RGB指紋。
func (h *HD) ValidatePageSources(sprites Sprites) error {
	p := h.pages
	if p == nil || p.validated {
		return nil
	}
	for _, source := range p.sources {
		pic, err := sprites.Get(source.Archive, source.Index)
		if err != nil {
			return err
		}
		if h.fingerprint(pic) != source.RGB {
			return fmt.Errorf("HD頁面來源%s第%d圖指紋不符", source.Archive, source.Index)
		}
	}
	for i := 3; i < 71; i++ {
		cmd := hdCommandFor(i-3, false)
		if cmd.key == "" {
			continue
		}
		pic, err := sprites.Get("KOLMPICS", i)
		if err != nil {
			return err
		}
		p.icons[h.fingerprint(pic)] = h.style.images["icon-"+cmd.key+".png"]
	}
	p.validated = true
	return nil
}

// RenderUI 只在同一呈現主執行緒替換繪製，不動文字量測、輸入及保存。
func (h *HD) RenderUI(dst *ebiten.Image, render func() error) error {
	if h.pages == nil || !h.pages.validated {
		return fmt.Errorf("HD頁面素材尚未完整核對")
	}
	n := &hdNativeRender{h: h, dst: dst}
	previous := activeHDUI
	activeHDUI = n
	defer func() { activeHDUI = previous }()
	if err := font.WithPainter(n, render); err != nil {
		return err
	}
	return n.err
}

func (h *HD) pagePortrait(src *ebiten.Image) (*ebiten.Image, bool) {
	if h.pages == nil {
		return nil, false
	}
	pic, ok := h.pages.portraits[h.fingerprint(src)]
	return pic, ok
}

func fillHDUI(dst *ebiten.Image, x, y, w, height int, c [3]uint8) bool {
	n := activeHDUI
	if n == nil || dst != n.dst {
		return false
	}
	confirm := book.CJKConfirm
	if c == [3]uint8{255, 255, 255} && x == confirm.X+1 && y == confirm.Y+1 &&
		w == confirm.Width-2 && height == confirm.Height-2 {
		bg := color.RGBA{20, 17, 22, 255}
		if n.book {
			bg = color.RGBA{238, 221, 185, 255}
		}
		vector.DrawFilledRect(dst, float32(x*2), float32(y*2), float32(w*2), float32(height*2), bg, false)
		vector.StrokeRect(dst, float32(x*2-2), float32(y*2-2), float32(w*2+4), float32(height*2+4), 2, hdGold, false)
		return true
	}
	ink := hdMuted
	switch c {
	case [3]uint8{0, 0, 0}:
		ink = color.RGBA{50, 24, 27, 255}
	case [3]uint8{255, 255, 255}, [3]uint8{170, 85, 0}:
		ink = hdGold
	case [3]uint8{0, 170, 170}:
		ink = color.RGBA{120, 104, 87, 255}
	case [3]uint8{255, 85, 85}:
		ink = color.RGBA{255, 133, 125, 255}
	case [3]uint8{85, 85, 255}:
		ink = color.RGBA{120, 104, 87, 255}
	}
	vector.DrawFilledRect(dst, float32(x*2), float32(y*2), float32(w*2), float32(height*2), ink, false)
	return true
}

func drawHDUI(dst, src *ebiten.Image, x, y, scale int) bool {
	n := activeHDUI
	if n == nil || dst != n.dst {
		return false
	}
	h := n.h
	if n.book && src.Bounds() == image.Rect(160, 0, 320, 200) && x == 480 && y == 0 && scale == 3 {
		return true
	}
	key := h.fingerprint(src)
	if key == h.pages.background {
		dst.DrawImage(h.style.images["background.png"], nil)
		return true
	}
	if page, ok := h.pages.books[key]; ok {
		n.book = true
		if page != 100 {
			h.drawHD(dst, h.pages.book, image.Rect(0, 0, 1920, 1200))
			if page == book.Contents {
				h.style.label(dst, "ORIGIN", image.Rect(180, 540, 846, 660), 72, color.RGBA{95, 63, 37, 255}, true)
				h.style.label(dst, "1989", image.Rect(180, 670, 846, 720), 32, color.RGBA{95, 63, 37, 255}, true)
				vector.StrokeCircle(dst, 453, 696, 13, 2, color.RGBA{95, 63, 37, 255}, true)
				h.style.label(dst, "C", image.Rect(442, 683, 465, 710), 18, color.RGBA{95, 63, 37, 255}, true)
			}
		}
		return true
	}
	painted := src
	if i, ok := h.matching[key]; ok {
		painted = h.portraits[i]
	} else if key == h.style.npcKey {
		painted = h.style.images["innkeeper-cropped.png"]
	} else if key == h.style.crestKey {
		painted = h.style.images["crest.png"]
	}
	r := image.Rect(x*2, y*2, (x+src.Bounds().Dx()*scale)*2, (y+src.Bounds().Dy()*scale)*2)
	if pic, ok := h.pages.portraits[key]; ok {
		h.drawHD(dst, pic, r)
		vector.StrokeRect(dst, float32(r.Min.X), float32(r.Min.Y), float32(r.Dx()), float32(r.Dy()), 2, hdGold, false)
		return true
	}
	if icon, ok := h.pages.icons[key]; ok {
		painted = icon
	}
	if painted != src {
		h.drawHD(dst, painted, r)
		return true
	}
	var op ebiten.DrawImageOptions
	op.GeoM.Scale(float64(scale*2), float64(scale*2))
	op.GeoM.Translate(float64(x*2), float64(y*2))
	op.Filter = ebiten.FilterNearest
	dst.DrawImage(src, &op)
	return true
}

func (n *hdNativeRender) DrawFont(dst *ebiten.Image, f *font.Font, x, y int, text string, foreground color.RGBA, background *color.RGBA) (int, bool) {
	r, g, b := foreground.R, foreground.G, foreground.B
	if dst != n.dst {
		return 0, false
	}
	ink := hdInk
	switch [3]uint8{r, g, b} {
	case [3]uint8{0, 170, 170}:
		ink = hdMuted
	case [3]uint8{170, 85, 0}:
		ink = hdGold
	case [3]uint8{255, 85, 85}, [3]uint8{170, 0, 0}:
		ink = color.RGBA{255, 133, 125, 255}
	}
	if n.book {
		ink = color.RGBA{48, 36, 29, 255}
		switch [3]uint8{r, g, b} {
		case [3]uint8{0, 170, 170}:
			ink = color.RGBA{117, 94, 75, 255}
		case [3]uint8{170, 85, 0}:
			ink = color.RGBA{136, 60, 32, 255}
		case [3]uint8{255, 85, 85}, [3]uint8{170, 0, 0}:
			ink = color.RGBA{169, 36, 31, 255}
		}
	}
	pen, drawn := x*2, 0
	for _, ch := range text {
		advance := f.RuneAdvance(ch) * 2
		glyph, ok := n.h.style.glyphs[string(ch)]
		if !ok {
			n.err = fmt.Errorf("HD介面缺字：%c", ch)
			pen += advance
			continue
		}
		if ch != ' ' {
			cell := image.Rect(pen, y*2, pen+advance, (y+f.Height())*2)
			if !cell.In(dst.Bounds()) {
				n.err = fmt.Errorf("HD介面文字超界：%q at %v", text, cell)
			}
			scaleX := .5
			if glyph.Advance*scaleX > float64(advance) {
				scaleX = float64(advance) / glyph.Advance
			}
			var op ebiten.DrawImageOptions
			op.GeoM.Scale(scaleX, .5)
			op.GeoM.Translate(float64(pen)-10*scaleX, float64(y*2)+32-76*.5)
			op.ColorScale.ScaleWithColor(ink)
			op.Filter = ebiten.FilterLinear
			dst.SubImage(cell).(*ebiten.Image).DrawImage(n.h.style.images["font-atlas.png"].SubImage(hdRect(glyph.X, glyph.Y, 96, 96)).(*ebiten.Image), &op)
		}
		drawn++
		pen += advance
	}
	return drawn, true
}

func (n *hdNativeRender) drawEditor(e *bitmapedit.Editor, blink bool) error {
	dst, h := n.dst, n.h
	dst.DrawImage(h.style.images["background.png"], nil)
	panel := func(x, y, w, height int, active bool) {
		ink := color.RGBA{120, 104, 87, 255}
		if active {
			ink = hdGold
		}
		vector.DrawFilledRect(dst, float32(x), float32(y), float32(w), float32(height), color.RGBA{20, 17, 22, 255}, false)
		vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(height), 2, ink, false)
	}
	panel(84, 12, 984, 936, true)
	panel(1194, 30, 252, 240, false)
	pix := e.Pix
	if !e.ToolsMode && blink {
		pix.Set(e.X, e.Y, pix.At(e.X, e.Y)^15)
	}
	if h.pages.editorImage == nil {
		h.pages.editorImage = ebiten.NewImage(bitmapedit.Width, bitmapedit.Height)
		h.pages.editorImage.WritePixels(pix.Image().Pix)
		h.pages.editorPixels = pix
	}
	if pix != h.pages.editorPixels {
		h.pages.editorImage.WritePixels(pix.Image().Pix)
		h.pages.editorPixels = pix
	}
	for _, p := range []struct{ x, y, scale int }{{96, 24, 24}, {1200, 36, 6}} {
		var op ebiten.DrawImageOptions
		op.GeoM.Scale(float64(p.scale), float64(p.scale))
		op.GeoM.Translate(float64(p.x), float64(p.y))
		op.Filter = ebiten.FilterNearest
		dst.DrawImage(h.pages.editorImage, &op)
	}
	panel(1632, 84, 144, 132, true)
	vector.DrawFilledRect(dst, 1638, 90, 132, 120, bitmapedit.Palette[e.Color], false)
	for i := 0; i < 16; i++ {
		x, y := 1200+(i%4)*144, 372+(i/4)*132
		vector.DrawFilledRect(dst, float32(x), float32(y), 144, 132, bitmapedit.Palette[i], false)
		ink := color.RGBA{120, 104, 87, 255}
		if e.ToolsMode && e.Color == i {
			ink = hdGold
		}
		vector.StrokeRect(dst, float32(x), float32(y), 144, 132, 3, ink, false)
	}
	labels := []string{"返回", "重載", "完成", "填滿", "右移", "上移", "左右翻轉", "上下翻轉"}
	for i, label := range labels {
		for _, ch := range label {
			if _, ok := h.style.glyphs[string(ch)]; !ok {
				return fmt.Errorf("HD編輯器缺字：%c", ch)
			}
		}
		x, y := 192+i*192, 1020
		panel(x, y, 144, 132, e.ToolsMode && e.Tool == i)
		if i == 0 || i == 2 {
			key := "back"
			if i == 2 {
				key = "yes"
			}
			h.drawHD(dst, h.style.images["icon-"+key+".png"], image.Rect(x+35, y+12, x+109, y+86))
		} else {
			c := hdGold
			switch i {
			case 1:
				vector.StrokeCircle(dst, float32(x+72), float32(y+50), 26, 4, c, true)
				vector.StrokeLine(dst, float32(x+45), float32(y+29), float32(x+45), float32(y+49), 4, c, true)
			case 3:
				vector.DrawFilledRect(dst, float32(x+44), float32(y+24), 56, 52, c, false)
			case 4:
				vector.StrokeLine(dst, float32(x+37), float32(y+50), float32(x+105), float32(y+50), 5, c, true)
				vector.StrokeLine(dst, float32(x+105), float32(y+50), float32(x+88), float32(y+33), 5, c, true)
				vector.StrokeLine(dst, float32(x+105), float32(y+50), float32(x+88), float32(y+67), 5, c, true)
			case 5:
				vector.StrokeLine(dst, float32(x+72), float32(y+80), float32(x+72), float32(y+20), 5, c, true)
				vector.StrokeLine(dst, float32(x+72), float32(y+20), float32(x+55), float32(y+37), 5, c, true)
				vector.StrokeLine(dst, float32(x+72), float32(y+20), float32(x+89), float32(y+37), 5, c, true)
			case 6:
				vector.StrokeRect(dst, float32(x+34), float32(y+26), 30, 50, 3, c, true)
				vector.StrokeRect(dst, float32(x+80), float32(y+26), 30, 50, 3, c, true)
				vector.StrokeLine(dst, float32(x+72), float32(y+18), float32(x+72), float32(y+84), 2, c, true)
			case 7:
				vector.StrokeRect(dst, float32(x+46), float32(y+20), 52, 25, 3, c, true)
				vector.StrokeRect(dst, float32(x+46), float32(y+58), 52, 25, 3, c, true)
				vector.StrokeLine(dst, float32(x+34), float32(y+51), float32(x+110), float32(y+51), 2, c, true)
			}
		}
		h.style.label(dst, label, image.Rect(x+6, y+96, x+138, y+128), 26, hdInk, true)
	}
	return nil
}
