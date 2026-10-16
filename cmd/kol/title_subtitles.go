package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"os"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/kol-remake-cht/internal/font"
	"github.com/wicanr2/kol-remake-cht/internal/ui"
)

// titleSubtitles 是開場字幕卡與書本封面的中文字幕（docs/i18n/title-zh-Hant.json）。
// 原版美術不動，中文模式另外畫一條黑底字幕；人名與版權行保留原文。
type titleSubtitles struct {
	Schema string `json:"schema"`
	Backs  map[string]struct {
		Zh        string `json:"zh"`
		Placement string `json:"placement"`
	} `json:"backs"`
	Cover struct {
		Zh string `json:"zh"`
	} `json:"cover"`
}

func loadTitleSubtitles(path string) (*titleSubtitles, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("讀取開場字幕失敗：%w", err)
	}
	var t titleSubtitles
	if err := json.Unmarshal(body, &t); err != nil {
		return nil, fmt.Errorf("解析開場字幕失敗：%w", err)
	}
	if t.Schema != "kol-title-subtitles-v1" || len(t.Backs) == 0 || t.Cover.Zh == "" {
		return nil, fmt.Errorf("開場字幕欄位不符")
	}
	return &t, nil
}

// 字幕的配色：與標題美術的淡青字同色，底是黑色，外框用白色。
var (
	subtitleInk   = [3]uint8{0x55, 0xFF, 0xFF}
	subtitleFrame = [3]uint8{0xFF, 0xFF, 0xFF}
)

// subtitleScale 是字幕的放大倍率：中文字模以最近鄰放大兩倍，與原版美術的粗像素
// 風格一致，也不會被巨大的美術字比下去。
const subtitleScale = 2

// plate 在 (cx, cy) 置中畫一塊黑底字幕板；framed 時外加白框。
func plate(dst *ebiten.Image, f *font.Font, cx, cy int, text string, framed bool) {
	if ui.DrawNativeSubtitle(dst, f, cx, cy, text, framed) {
		return
	}
	w, h := f.Measure(text)
	if w == 0 {
		return
	}
	glyphs := ebiten.NewImage(w, h)
	f.Draw(glyphs, 0, 0, text, subtitleInk[0], subtitleInk[1], subtitleInk[2])
	sw, sh := w*subtitleScale, h*subtitleScale
	pad := 8 * subtitleScale
	x, y := cx-sw/2-pad, cy-sh/2-4*subtitleScale
	bw, bh := sw+2*pad, sh+8*subtitleScale
	if framed {
		fillRect(dst, x-subtitleScale, y-subtitleScale, bw+2*subtitleScale, bh+2*subtitleScale, subtitleFrame)
	}
	fillRect(dst, x, y, bw, bh, [3]uint8{0, 0, 0})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(subtitleScale, subtitleScale)
	op.GeoM.Translate(float64(cx-sw/2), float64(cy-sh/2))
	op.Filter = ebiten.FilterNearest
	dst.DrawImage(glyphs, op)
}

// drawTitleSubtitle 畫開場字幕卡的中文字幕。一般字幕在畫面下緣；標誌那一張
// 畫成劍身右段上的銘牌，避開英文標誌與版權行。
func drawTitleSubtitle(dst *ebiten.Image, f *font.Font, t *titleSubtitles, back int) {
	if t == nil || f == nil {
		return
	}
	s, ok := t.Backs[strconv.Itoa(back)]
	if !ok || s.Zh == "" {
		return
	}
	b := dst.Bounds()
	if s.Placement == "logo" {
		plate(dst, f, b.Dx()*790/960, b.Dy()*278/600, s.Zh, true)
		return
	}
	bottom := b.Dy() - 30
	if ui.NativeRendering(dst) {
		bottom = b.Dy() - 60
	}
	plate(dst, f, b.Dx()/2, bottom, s.Zh, false)
}

// drawCoverSubtitle 在書本封面書名板正下方畫中文副標。
func drawCoverSubtitle(dst *ebiten.Image, f *font.Font, t *titleSubtitles) {
	if t == nil || f == nil {
		return
	}
	b := dst.Bounds()
	plate(dst, f, b.Dx()*712/960, b.Dy()*372/600, t.Cover.Zh, true)
}

func fillRect(dst *ebiten.Image, x, y, w, h int, c [3]uint8) {
	if ui.FillNativeRect(dst, x, y, w, h, c) {
		return
	}
	img := ebiten.NewImage(w, h)
	img.Fill(color.RGBA{c[0], c[1], c[2], 0xFF})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	dst.DrawImage(img, op)
}
