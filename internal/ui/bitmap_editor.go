package ui

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/kol-remake-cht/internal/bitmapedit"
	"github.com/wicanr2/kol-remake-cht/internal/book"
	"github.com/wicanr2/kol-remake-cht/internal/font"
)

// 點陣圖編輯器的畫面（SPEC-036，KOLBOOK `sub_1293E`）。座標是原版 320 × 200。
const (
	editorBackground = 3    // BOOKPICS 第 3 張：整頁紅磚（`sub_1137B(3)`）
	editorToolbar    = 0x0D // BOOKPICS 第 13 張：八格工具列長條（`sub_1137B(0xD)`）
	editorToolbarX   = 32
	editorToolbarY   = 170
	editorCanvasX    = 16 // 放大圖左上角；每個像素 4 × 4（`sub_153EA(seg, 2, 4)`）
	editorCanvasY    = 4
	editorPreviewX   = 200 // 預覽 1:1（`sub_15335(seg, 25, 6)`）
	editorPreviewY   = 6
	editorPaletteX   = 200 // 色盤 4 × 4，一格 24 × 22
	editorPaletteY   = 62
)

// editorColour 是 EGA 索引對應的顏色。
func editorColour(i int) [3]uint8 {
	c := bitmapedit.Palette[i&0x0F]
	return [3]uint8{c.R, c.G, c.B}
}

// edBox 畫含端點的矩形 (x0, y0)-(x1, y1) 的 1 px 外框（`sub_14D11`）。
func edBox(dst *ebiten.Image, x0, y0, x1, y1, c int) {
	frame(dst, x0, y0, x1-x0+1, y1-y0+1, editorColour(c))
}

// solid 畫含端點的實心矩形（`sub_14DA2`）。
func solid(dst *ebiten.Image, x0, y0, x1, y1, c int) {
	fill(dst, x0, y0, x1-x0+1, y1-y0+1, editorColour(c))
}

// drop 畫黑色投影：右下偏移 4、厚 4（`sub_14E0C`）。
func drop(dst *ebiten.Image, x0, y0, x1, y1 int) {
	shadow(dst, x0, y0, x1-x0+1, y1-y0+1)
}

// BitmapEditor 畫編輯器。blink 是畫布模式游標閃爍的相位：原版在沒有滑鼠驅動時讓游標
// 那一格（與預覽的對應點）在原色與「原色 xor 0xF」之間交替；有滑鼠驅動時完全不畫
// 游標、兩種模式的選中框也不變。remake 照沒有滑鼠驅動的那一條畫（RE-161），
// 鍵盤操作才看得到游標。
func BitmapEditor(dst *ebiten.Image, sprites Sprites, small *font.Font, e *bitmapedit.Editor, blink bool) error {
	if err := drawBitmapEditor(dst, sprites, e, blink); err != nil {
		return err
	}
	if e.Asking {
		drawConfirmWindow(dst, small, []book.Label{
			{"SAVE CHANGES", 104, 64, 170, 0, 0},
			{"YES/NO", 128, 70, 170, 0, 0},
		})
	}
	return nil
}

// BitmapEditorCJK 是中文模式：編輯區照原版版面放大三倍，只有存檔問句改成中文視窗。
func BitmapEditorCJK(dst *ebiten.Image, sprites Sprites, f *font.Font, t Translator,
	e *bitmapedit.Editor, blink bool) error {
	canvas := ebiten.NewImage(320, 200)
	if err := drawBitmapEditor(canvas, sprites, e, blink); err != nil {
		return err
	}
	drawScaled(dst, canvas, 0, 0, book.CJKScale)
	if e.Asking {
		w := book.CJKConfirm
		fill(dst, w.X+w.Width, w.Y+w.ShadowOffset, w.ShadowThickness, w.Height, book.ColourValue)
		fill(dst, w.X+w.ShadowOffset, w.Y+w.Height, w.Width, w.ShadowThickness, book.ColourValue)
		fill(dst, w.X, w.Y, w.Width, w.Height, book.ColourValue)
		fill(dst, w.X+1, w.Y+1, w.Width-2, w.Height-2,
			[3]uint8{book.PageInk.R, book.PageInk.G, book.PageInk.B})
		p := book.ColourPrompt
		f.Draw(dst, w.Line1X, w.Line1Y, t.T("SAVE CHANGES"), p[0], p[1], p[2])
		f.Draw(dst, w.Line2X, w.Line2Y, t.T("YES/NO"), p[0], p[1], p[2])
	}
	return nil
}

func drawBitmapEditor(dst *ebiten.Image, sprites Sprites, e *bitmapedit.Editor, blink bool) error {
	bg, err := sprites.Get("BOOKPICS", editorBackground)
	if err != nil {
		return err
	}
	draw(dst, bg, 0, 0)
	bar, err := sprites.Get("BOOKPICS", editorToolbar)
	if err != nil {
		return err
	}
	draw(dst, bar, editorToolbarX, editorToolbarY)

	// 放大圖：兩層色 3 外框，外層有投影；每個像素 4 × 4 實心，沒有格線。
	drop(dst, 14, 2, 177, 157)
	edBox(dst, 14, 2, 177, 157, 3)
	edBox(dst, 15, 3, 176, 156, 3)
	// 預覽：外框 (199,5)-(240,44) 色 3，有投影。
	drop(dst, 199, 5, 240, 44)
	edBox(dst, 199, 5, 240, 44, 3)
	cursorOn := !e.ToolsMode && blink
	for y := 0; y < bitmapedit.Height; y++ {
		for x := 0; x < bitmapedit.Width; x++ {
			c := e.Pix.At(x, y)
			if cursorOn && x == e.X && y == e.Y {
				c ^= 0x0F
			}
			solid(dst, editorCanvasX+4*x, editorCanvasY+4*y, editorCanvasX+4*x+3, editorCanvasY+4*y+3, c)
			solid(dst, editorPreviewX+x, editorPreviewY+y, editorPreviewX+x, editorPreviewY+y, c)
		}
	}

	// 目前顏色：白框 (272,14)-(295,35)，內部 (273,15)-(294,34) 填目前顏色，有投影。
	drop(dst, 272, 14, 295, 35)
	edBox(dst, 272, 14, 295, 35, 15)
	solid(dst, 273, 15, 294, 34, e.Color)

	// 色盤：第 i 格填索引 i、色 3 框；整塊有投影。選中的那一格白框。
	drop(dst, 200, 62, 295, 149)
	for i := 0; i < 16; i++ {
		c, r := i%4, i/4
		x0, y0 := editorPaletteX+24*c, editorPaletteY+22*r
		solid(dst, x0, y0, x0+23, y0+21, i)
		edBox(dst, x0, y0, x0+23, y0+21, 3)
	}
	// 沒有滑鼠驅動時，進畫布模式把兩個白框都改回色 3（`0x12EF2`）。
	if e.ToolsMode {
		c, r := e.Color%4, e.Color/4
		x0, y0 := editorPaletteX+24*c, editorPaletteY+22*r
		edBox(dst, x0, y0, x0+23, y0+21, 15)
		x := editorToolbarX + 32*e.Tool
		edBox(dst, x, editorToolbarY, x+23, editorToolbarY+21, 15)
	}
	return nil
}
