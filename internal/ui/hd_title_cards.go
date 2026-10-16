package ui

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/kol-remake-cht/internal/font"
	"image"
	"image/color"
)

func NativeRendering(dst *ebiten.Image) bool { return activeHDUI != nil && activeHDUI.dst == dst }
func (h *HD) AttachTitleCards(cards []HDPageSource) error {
	if h.motion == nil || len(cards) != 7 {
		return fmt.Errorf("HD字幕卡需要完整七張來源")
	}
	next := map[string]int{}
	sources := append([]HDPageSource{}, h.pages.sources...)
	for i, c := range cards {
		if c.Archive != "BACKPICS" || c.Index != i+7 || len(c.RGB) != 64 {
			return fmt.Errorf("HD字幕卡來源不符")
		}
		next[c.RGB] = c.Index
		h.motion.sourceSizes[c.RGB] = image.Pt(320, 200)
		sources = append(sources, c)
	}
	h.motion.cards = next
	h.pages.sources = sources
	h.pages.validated = false
	return nil
}
func (h *HD) drawNativeTitleCard(dst *ebiten.Image, index int) {
	dst.Fill(color.Black)
	s := h.style
	ink := color.RGBA{192, 224, 229, 255}
	row := func(text string, y, height int, size float64) {
		s.label(dst, text, image.Rect(80, y, 1840, y+height), size, ink, true)
	}
	switch index {
	case 7:
		row("ORIGIN", 180, 260, 200)
		row("Presents...", 740, 180, 100)
	case 8:
		for i, line := range []string{"A", "Dallas Snell", "&", "Alan Gardner", "Production", "Of"} {
			row(line, 70+i*155, 142, 92)
		}
	case 9:
		for i, line := range []string{"A", "Fantasy", "Role Playing", "Epic", "By", "Todd Porter"} {
			row(line, 70+i*155, 142, 90)
		}
	case 10:
		row("KNIGHTS", 30, 330, 230)
		row("LEGEND", 690, 320, 220)
		row("OF", 495, 110, 80)
		vector.DrawFilledRect(dst, 220, 534, 1500, 56, color.RGBA{166, 176, 180, 255}, false)
		vector.StrokeLine(dst, 260, 540, 1720, 540, 3, ink, true)
		vector.StrokeLine(dst, 260, 590, 1720, 590, 3, hdMuted, true)
		vector.DrawFilledRect(dst, 120, 520, 155, 84, color.RGBA{121, 58, 37, 255}, false)
		vector.StrokeLine(dst, 280, 492, 280, 633, 16, hdGold, true)
		vector.StrokeCircle(dst, 128, 562, 30, 5, hdGold, true)
		row("OF", 498, 112, 80)
		row("Copyright 1989 Origin Systems, Inc.", 1090, 56, 36)
	case 11, 13:
		row("Programming:", 150, 160, 100)
		names := []string{"Naser Jannati", "Gary Scott Smith"}
		if index == 13 {
			names[0], names[1] = names[1], names[0]
		}
		for i, line := range names {
			row(line, 430+i*240, 170, 100)
		}
	case 12:
		row("Graphics:", 50, 120, 96)
		for i, line := range []string{"Denis Loubet", "Dan Bourbonnais", "Glen Johnson", "Keith Berdak", "Jeff Dee"} {
			row(line, 220+i*160, 130, 86)
		}
	}
}
func (h *HD) drawNativeCoverTitle(dst *ebiten.Image) {
	for i, line := range []string{"KNIGHTS", "OF", "LEGEND"} {
		h.style.label(dst, line, image.Rect(1144, 240+i*130, 1704, 340+i*130), 68, hdGold, true)
	}
}
func DrawNativeSubtitle(dst *ebiten.Image, f *font.Font, cx, cy int, text string, framed bool) bool {
	n := activeHDUI
	if n == nil || n.dst != dst {
		return false
	}
	for _, r := range text {
		if _, ok := n.h.style.glyphs[string(r)]; !ok {
			n.err = fmt.Errorf("HD字幕缺字：%c", r)
			return true
		}
	}
	w, height := f.Measure(text)
	w *= 4
	height *= 4
	rect := image.Rect(cx-w/2-32, cy-height/2-16, cx+w/2+32, cy+height/2+16)
	if !rect.In(dst.Bounds()) {
		n.err = fmt.Errorf("HD字幕超出畫面：%q", text)
		return true
	}
	vector.DrawFilledRect(dst, float32(rect.Min.X), float32(rect.Min.Y), float32(rect.Dx()), float32(rect.Dy()), color.Black, false)
	if framed {
		vector.StrokeRect(dst, float32(rect.Min.X), float32(rect.Min.Y), float32(rect.Dx()), float32(rect.Dy()), 2, hdGold, false)
	}
	n.h.style.label(dst, text, rect.Inset(16), 64, hdInk, true)
	return true
}
func FillNativeRect(dst *ebiten.Image, x, y, w, h int, c [3]uint8) bool {
	return fillHDUI(dst, x, y, w, h, c)
}

// ClipNativeText explicitly registers a logical text rectangle. Other offscreen
// font buffers keep their normal renderer and are never redirected accidentally.
func ClipNativeText(dst *ebiten.Image, rect image.Rectangle) *ebiten.Image {
	clip := dst.SubImage(rect).(*ebiten.Image)
	n := activeHDUI
	if n != nil && n.dst == dst {
		if n.clips == nil {
			n.clips = map[*ebiten.Image]image.Rectangle{}
		}
		n.clips[clip] = image.Rect(rect.Min.X*2, rect.Min.Y*2, rect.Max.X*2, rect.Max.Y*2)
	}
	return clip
}
