package ui

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/kol-remake-cht/internal/font"
)

// DisplayControls 共用960×600文字畫布的可見矩形，輸入端沿用同一組邊界。
var DisplayControls = struct {
	Button, Panel image.Rectangle
	Modes         [2]image.Rectangle
	Back          image.Rectangle
}{
	Button: image.Rect(852, 576, 948, 600),
	Panel:  image.Rect(572, 176, 944, 532),
	Modes:  [2]image.Rectangle{image.Rect(592, 248, 924, 322), image.Rect(592, 332, 924, 406)},
	Back:   image.Rect(708, 476, 924, 512),
}

type DisplaySettings struct {
	Open    bool
	HD      bool
	Cursor  int
	Problem bool
}

// DrawDisplaySettings 是互動螢幕的獨立透明層，不寫進場景截圖或遊戲資料。
func DrawDisplaySettings(dst *ebiten.Image, f *font.Font, t Translator, s DisplaySettings) error {
	gold := color.RGBA{214, 183, 112, 255}
	ink := color.RGBA{239, 233, 218, 255}
	muted := color.RGBA{170, 180, 188, 255}
	fill := func(r image.Rectangle, c color.RGBA) { dst.SubImage(r).(*ebiten.Image).Fill(c) }
	box := func(r image.Rectangle, edge, inside color.RGBA) {
		fill(r, edge)
		fill(r.Inset(1), inside)
	}
	label := func(r image.Rectangle, text string, c color.RGBA, centered bool) error {
		w, h := f.Measure(text)
		if w > r.Dx() || h > r.Dy() {
			return fmt.Errorf("畫面設定文字超界：%s", text)
		}
		x := r.Min.X
		if centered {
			x += (r.Dx() - w) / 2
		}
		f.Draw(dst, x, r.Min.Y+(r.Dy()-h)/2, text, c.R, c.G, c.B)
		return nil
	}
	if s.Open {
		fill(image.Rect(0, 0, 960, 600), color.RGBA{0, 0, 0, 104})
		box(DisplayControls.Panel, gold, color.RGBA{18, 26, 34, 255})
		for _, l := range []struct {
			r image.Rectangle
			k string
			c color.RGBA
		}{
			{image.Rect(592, 192, 924, 218), "Display settings", ink},
			{image.Rect(592, 220, 924, 242), "Changes preview immediately", muted},
		} {
			if err := label(l.r, t.T(l.k), l.c, false); err != nil {
				return err
			}
		}
		active := 0
		if s.HD {
			active = 1
		}
		for i, r := range DisplayControls.Modes {
			edge := color.RGBA{67, 80, 88, 255}
			if i == s.Cursor {
				edge = ink
			}
			if i == active {
				edge = gold
			}
			box(r, edge, color.RGBA{28, 39, 47, 255})
			titles := [2]string{"EGA Classic", "HD Painted"}
			details := [2]string{"Original pixel artwork", "Inn and six characters"}
			if err := label(image.Rect(r.Min.X+16, r.Min.Y+10, r.Max.X-76, r.Min.Y+32), t.T(titles[i]), ink, false); err != nil {
				return err
			}
			if err := label(image.Rect(r.Min.X+16, r.Min.Y+39, r.Max.X-16, r.Max.Y-10), t.T(details[i]), muted, false); err != nil {
				return err
			}
			if i == active {
				if err := label(image.Rect(r.Max.X-70, r.Min.Y+10, r.Max.X-12, r.Min.Y+32), t.T("In use"), gold, true); err != nil {
					return err
				}
			}
		}
		message, message2 := "Other artwork uses the classic style", ""
		c := muted
		if s.Problem {
			message, message2 = "HD images are unavailable.", "Current display retained. Try again later."
			c = color.RGBA{255, 166, 140, 255}
		}
		for i, k := range []string{message, message2} {
			if k != "" {
				if err := label(image.Rect(592, 418+i*24, 924, 440+i*24), t.T(k), c, false); err != nil {
					return err
				}
			}
		}
		edge := gold
		if s.Cursor == 2 {
			edge = ink
		}
		box(DisplayControls.Back, edge, color.RGBA{37, 49, 57, 255})
		if err := label(DisplayControls.Back.Inset(4), t.T("Back to game  Esc"), ink, true); err != nil {
			return err
		}
	}
	box(DisplayControls.Button, gold, color.RGBA{18, 26, 34, 255})
	return label(DisplayControls.Button.Inset(2), t.T("Display F10"), ink, true)
}
