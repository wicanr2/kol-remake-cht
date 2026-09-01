// Package font 載入由 tools/extract_fonts.py 抽出的原版點陣字型。
//
// 字模表是 RGBA PNG：墨色為不透明白、其餘透明，因此繪製時以 ColorScale 上色即可
// 得到原版的任一 EGA 顏色。字距與字模格由 manifest 提供，不在程式裡硬編。
package font

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

type manifest struct {
	Schema string  `json:"schema"`
	Fonts  []entry `json:"fonts"`
}

type entry struct {
	Name           string `json:"name"`
	File           string `json:"file"`
	Advance        int    `json:"advance"`
	FirstCodepoint int    `json:"first_codepoint"`
	GlyphCount     int    `json:"glyph_count"`
	Cell           struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"cell"`
	Sheet struct {
		Columns int `json:"columns"`
		Width   int `json:"width"`
		Height  int `json:"height"`
	} `json:"sheet"`
}

// Font 是一套已載入的原版點陣字型。
type Font struct {
	meta   entry
	sheet  *ebiten.Image
	glyphs map[rune]*ebiten.Image
}

// Load 讀取字型目錄下的指定字型。
func Load(root, name string) (*Font, error) {
	raw, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		return nil, fmt.Errorf("讀取字型 manifest 失敗：%w", err)
	}
	var m manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("解析字型 manifest 失敗：%w", err)
	}
	if !strings.HasPrefix(m.Schema, "kol-font-manifest-") {
		return nil, fmt.Errorf("非預期的字型 manifest schema：%s", m.Schema)
	}
	var meta entry
	found := false
	for _, e := range m.Fonts {
		if e.Name == name {
			meta, found = e, true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("字型 manifest 沒有 %s", name)
	}
	path := filepath.Join(root, meta.File)
	if strings.ToLower(filepath.Ext(path)) != ".png" {
		return nil, fmt.Errorf("拒絕載入非 PNG 字模表：%s", meta.File)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	src, format, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	if format != "png" {
		return nil, fmt.Errorf("%s 不是 PNG", meta.File)
	}
	b := src.Bounds()
	if b.Dx() != meta.Sheet.Width || b.Dy() != meta.Sheet.Height {
		return nil, fmt.Errorf("%s 尺寸與 manifest 不符", meta.File)
	}
	return &Font{meta: meta, sheet: ebiten.NewImageFromImage(src),
		glyphs: map[rune]*ebiten.Image{}}, nil
}

// Advance 是每個字元前進的像素數。
func (f *Font) Advance() int { return f.meta.Advance }

// Height 是字模格高度。
func (f *Font) Height() int { return f.meta.Cell.Height }

// Measure 回傳字串佔用的文字框尺寸。
func (f *Font) Measure(text string) (int, int) {
	return f.meta.Advance * len([]rune(text)), f.meta.Cell.Height
}

func (f *Font) glyph(r rune) (*ebiten.Image, bool) {
	if img, ok := f.glyphs[r]; ok {
		return img, img != nil
	}
	index := int(r) - f.meta.FirstCodepoint
	if index < 0 || index >= f.meta.GlyphCount {
		f.glyphs[r] = nil
		return nil, false
	}
	cw, ch := f.meta.Cell.Width, f.meta.Cell.Height
	x := (index % f.meta.Sheet.Columns) * cw
	y := (index / f.meta.Sheet.Columns) * ch
	sub := f.sheet.SubImage(image.Rect(x, y, x+cw, y+ch)).(*ebiten.Image)
	f.glyphs[r] = sub
	return sub, true
}

// DrawOpaque 逐字模格先填 bg 再畫墨色。原版就是這樣畫的：書頁底色有黃色斑點，
// 被字模格覆蓋的地方會清成白底。空白字元不清底——原版在字串內的空白處保留書頁
// 斑點，這一點由原版畫面逐像素比對確認。
func (f *Font) DrawOpaque(dst *ebiten.Image, x, y int, text string,
	r, g, b uint8, bgR, bgG, bgB uint8) int {
	cw, ch := f.meta.Cell.Width, f.meta.Cell.Height
	cell := ebiten.NewImage(cw, ch)
	cell.Fill(color.RGBA{bgR, bgG, bgB, 0xFF})
	drawn := 0
	for i, r0 := range []rune(text) {
		if r0 == ' ' {
			continue
		}
		if _, ok := f.glyph(r0); !ok {
			continue
		}
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x+i*f.meta.Advance), float64(y))
		op.Filter = ebiten.FilterNearest
		dst.DrawImage(cell, op)
		drawn++
	}
	f.Draw(dst, x, y, text, r, g, b)
	_ = ch
	return drawn
}

// Draw 以指定顏色把字串畫到目標影像；字模表是白色，靠 ColorScale 上色。
// 回傳實際畫出的字元數，未涵蓋的 codepoint 會被跳過並照樣前進，避免版面位移。
func (f *Font) Draw(dst *ebiten.Image, x, y int, text string, r, g, b uint8) int {
	drawn := 0
	for i, ch := range []rune(text) {
		img, ok := f.glyph(ch)
		if !ok {
			continue
		}
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x+i*f.meta.Advance), float64(y))
		op.Filter = ebiten.FilterNearest
		op.ColorScale.Scale(float32(r)/255, float32(g)/255, float32(b)/255, 1)
		dst.DrawImage(img, op)
		drawn++
	}
	return drawn
}
