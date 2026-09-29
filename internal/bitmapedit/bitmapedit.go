// Package bitmapedit 是書本 EDIT PICTURE／EDIT SHIELD 的點陣圖編輯器（SPEC-036，KOLBOOK
// `sub_1293E`）。這裡只有狀態與規則，畫面與按鍵對應在 flow／ui。
//
// 一張圖是 40 × 38 個 4 位元像素，一列 20 byte，高 nibble 在左，共 760 byte，與原版
// CHARPICS／SHLDPICS 的一筆相同。
package bitmapedit

import (
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
)

// 圖的尺寸。
const (
	Width  = 40
	Height = 38
	Bytes  = Width * Height / 2
)

// Bitmap 是一張圖。
type Bitmap [Bytes]byte

// At 回傳 (x, y) 的顏色索引。
func (b *Bitmap) At(x, y int) int {
	v := b[y*Width/2+x/2]
	if x%2 == 0 {
		return int(v >> 4)
	}
	return int(v & 0x0F)
}

// Set 把 (x, y) 設成顏色 c（`sub_14ED3` 的寫入那一條）。
func (b *Bitmap) Set(x, y, c int) {
	i := y*Width/2 + x/2
	if x%2 == 0 {
		b[i] = b[i]&0x0F | byte(c&0x0F)<<4
	} else {
		b[i] = b[i]&0xF0 | byte(c&0x0F)
	}
}

// 原版的按鍵碼：一般字元是 ASCII，延伸鍵是掃描碼左移 8 位（`sub_146CE` 讀兩次 getch）。
const (
	KeyEnter  = 0x0D
	KeyEsc    = 0x1B
	KeyUp     = 0x4800
	KeyDown   = 0x5000
	KeyLeft   = 0x4B00
	KeyRight  = 0x4D00
	KeyHome   = 0x4700
	KeyPgUp   = 0x4900
	KeyEnd    = 0x4F00
	KeyPgDn   = 0x5100
	KeyCenter = 0x4C00 // 數字鍵盤 5
)

// 工具列八格（`0x12E71` 的分派表）。
const (
	ToolExit    = 0 // 與 2 同一條路
	ToolReload  = 1
	ToolExit2   = 2
	ToolFill    = 3
	ToolShiftR  = 4
	ToolShiftUp = 5
	ToolMirrorH = 6
	ToolMirrorV = 7
	ToolCount   = 8
)

// Session 是跨越多次進出編輯器仍留著的全域：`word_194F4`（滑鼠補送的鍵）與
// `word_194F6`／`word_194F8`（上一次滑鼠點到的畫布格）。原版沒有在進場時清掉。
type Session struct {
	pending      int
	lastX, lastY int
}

// Editor 是一次編輯。
type Editor struct {
	// Pix 是工作緩衝，Orig 是進場時那一筆（工具 1 從它重載，存檔前不動）。
	Pix, Orig Bitmap
	// Modified 是 `var_12`：有沒有改過；離開時決定要不要問存檔。
	Modified bool
	// ToolsMode 是 `word_1A36C`：真為工具列與色盤模式，假為畫布模式。
	ToolsMode bool
	// Tool 是 `word_19B4A`（白框所在的工具格），Color 是 `word_19B54`。
	Tool, Color int
	// X、Y 是畫布游標（`word_26904`／`word_26906`）。
	X, Y int
	// Pen 是 `var_10`：落筆之後每讀一次鍵都先在游標處畫一點。
	Pen bool
	// Asking 是停在 BOOKDATA 0x3A「SAVE CHANGES@YES/NO」上。
	Asking bool
	// Done 是離開了；Save 是離開時要不要寫回。
	Done, Save bool

	prevTool int // var_16
	session  *Session
}

// New 照 `sub_1293E` 開頭擺好一次編輯：工具列模式、目前顏色 3、白框在第 0 格、游標 (20, 20)。
func New(pic Bitmap, s *Session) *Editor {
	if s == nil {
		s = &Session{}
	}
	return &Editor{Pix: pic, Orig: pic, ToolsMode: true, Color: 3, X: 20, Y: 20, session: s}
}

// Key 是讀到一個鍵（`sub_146CE` 回傳之後的那一輪主迴圈）。
func (e *Editor) Key(k int) {
	if e.Done {
		return
	}
	if e.Asking {
		e.answer(k)
		return
	}
	e.step(k, k == KeyEsc)
}

// step 是主迴圈一輪。toggled 是 `word_19290`：ESC 或滑鼠命中三塊區域時立起，編輯器
// 看到就把模式翻到另一種。
func (e *Editor) step(k int, toggled bool) {
	if e.ToolsMode {
		e.toolsStep(k, toggled)
	} else {
		e.canvasStep(k, toggled)
	}
}

// toolsStep 是 `0x12C2D`…`0x13023`。
func (e *Editor) toolsStep(k int, toggled bool) {
	switch k {
	case KeyEnd, ',', '<':
		e.Tool--
	case KeyPgDn, '.', '>':
		e.Tool++
	}
	if e.Tool != e.prevTool {
		if e.Tool > 7 {
			e.Tool = 0
		}
		if e.Tool < 0 {
			e.Tool = 7
		}
		e.prevTool = e.Tool
	}
	if k == KeyEnter {
		if e.runTool() {
			return
		}
	}
	if toggled {
		e.ToolsMode = false
		return
	}
	switch k {
	case KeyUp:
		e.Color -= 4
	case KeyLeft:
		e.Color--
	case KeyRight:
		e.Color++
	case KeyDown:
		e.Color += 4
	default:
		return
	}
	if e.Color > 15 {
		e.Color -= 16
	} else if e.Color < 0 {
		e.Color += 16
	}
}

// runTool 執行白框所在的工具；回報是不是離開了（或停在存檔問句上）。
func (e *Editor) runTool() bool {
	switch e.Tool {
	case ToolExit, ToolExit2:
		if !e.Modified {
			e.Done = true
		} else {
			e.Asking = true
		}
		return true
	case ToolReload:
		e.Modified = false
		e.Pix = e.Orig
	case ToolFill:
		for x := 0; x < Width; x++ {
			for y := 0; y < Height; y++ {
				e.Pix.Set(x, y, e.Color)
			}
		}
		e.Modified = true
	case ToolShiftR:
		ShiftRight(&e.Pix)
		e.Modified = true
	case ToolShiftUp:
		ShiftUp(&e.Pix)
		e.Modified = true
	case ToolMirrorH:
		MirrorH(&e.Pix)
		e.Modified = true
	case ToolMirrorV:
		MirrorV(&e.Pix)
		e.Modified = true
	}
	return false
}

// answer 是存檔問句（`0x12CF3`）：只認 y／Y／n／N，其餘鍵重讀。
func (e *Editor) answer(k int) {
	switch k {
	case 'y', 'Y':
		e.Asking, e.Done, e.Save = false, true, true
	case 'n', 'N':
		e.Asking, e.Done = false, true
	}
}

// canvasStep 是 `0x13026`…`0x13342`。
func (e *Editor) canvasStep(k int, toggled bool) {
	if toggled {
		e.ToolsMode = true
	}
	if e.Pen {
		// 落筆時每一輪先畫目前那一格；這一條不標記已修改（`0x1311D`）。
		e.Pix.Set(e.X, e.Y, e.Color)
	}
	switch k {
	case 'x', 'X', KeyCenter:
		e.Pen = false
		e.Pix.Set(e.X, e.Y, e.Color)
		e.Modified = true
	case 'u', 'U':
		e.Pen = false
	case 'd', 'D':
		e.Pen = true
		e.Pix.Set(e.X, e.Y, e.Color)
		e.Modified = true
	case 'c', 'C':
		e.Color = e.Pix.At(e.X, e.Y)
	}
	switch k {
	case KeyUp:
		e.Y--
	case KeyLeft:
		e.X--
	case KeyRight:
		e.X++
	case KeyDown:
		e.Y++
	case KeyHome:
		e.X--
		e.Y--
	case KeyPgUp:
		e.X++
		e.Y--
	case KeyEnd:
		e.X--
		e.Y++
	case KeyPgDn:
		e.X++
		e.Y++
	}
	if e.X > Width-1 {
		e.X = 0
	}
	if e.X < 0 {
		e.X = Width - 1
	}
	if e.Y > Height-1 {
		e.Y = 0
	}
	if e.Y < 0 {
		e.Y = Height - 1
	}
}

// 滑鼠命中的三塊區域（`sub_146CE` 的 0x0E 那一段；座標已除以 2，是 320 × 200 的畫面座標）。
const (
	toolbarTop, toolbarBottom = 0xA9, 0xC0 // y 開區間
	toolbarLeft               = 0x40 / 2   // x 開區間下界
	paletteTop, paletteBottom = 0x3D, 0x95
	paletteLeft, paletteRight = 0x190 / 2, 0x258 / 2
	canvasLeft, canvasTop     = 0x20 / 2, 4
	canvasRight, canvasBottom = 0x160 / 2, 0x9C
)

// Click 是左鍵落在 (x, y)（320 × 200 座標）。原版在畫布一帶按住不放會一直讀，
// 其他地方要等放開；按住拖曳時呼叫端每移到一個新位置就再呼叫一次。
//
// 命中的區域與目前模式不同時，`sub_146CE` 先把要送的鍵存進 `word_194F4`、立起
// `word_19290` 回傳 1；主迴圈翻模式之後下一次讀鍵就拿到存著的那一個。
func (e *Editor) Click(x, y int) {
	if e.Done || e.Asking {
		return
	}
	vx := x * 2
	switch {
	case y > toolbarTop && y < toolbarBottom && vx > 0x40:
		i := (vx/2 - 0x20) / 0x20
		if i >= ToolCount {
			return
		}
		v := 1
		if i == e.Tool {
			v = KeyEnter
		} else {
			e.Tool = i
		}
		e.deliver(v, e.ToolsMode)
	case y > paletteTop && y < paletteBottom && vx > 0x190 && vx < 0x258:
		e.Color = (y-0x3E)/0x16*4 + (vx/2-0xC8)/0x18
		e.deliver(1, e.ToolsMode)
	case vx > 0x20 && y > 4 && vx < 0x160 && y < 0x9C:
		cx, cy := (vx/2-0x10)/4, (y-4)/4
		if cx == e.session.lastX && cy == e.session.lastY {
			return
		}
		e.X, e.Y = cx, cy
		e.session.lastX, e.session.lastY = cx, cy
		e.deliver('x', !e.ToolsMode)
	}
}

// deliver 把滑鼠換成的鍵送進主迴圈。direct 為真表示目前模式就是那一塊所屬的模式。
func (e *Editor) deliver(v int, direct bool) {
	if direct {
		e.step(v, false)
		return
	}
	e.step(1, true)
	if !e.Done && !e.Asking {
		e.step(v, false)
	}
}

// ShiftRight 是 `sub_11740`：每一列往右移一格，最右一格繞到最左。
func ShiftRight(b *Bitmap) {
	for y := 0; y < Height; y++ {
		last := b.At(Width-1, y)
		for x := Width - 1; x > 0; x-- {
			b.Set(x, y, b.At(x-1, y))
		}
		b.Set(0, y, last)
	}
}

// ShiftUp 是 `sub_1177B`：整張往上移一列，第 0 列繞到最底。
func ShiftUp(b *Bitmap) {
	var top [Width / 2]byte
	copy(top[:], b[:Width/2])
	copy(b[:], b[Width/2:])
	copy(b[Bytes-Width/2:], top[:])
}

// MirrorH 是 `sub_117AF`：每一列 byte 反序、nibble 互換，也就是左右鏡射。
func MirrorH(b *Bitmap) {
	for y := 0; y < Height; y++ {
		row := b[y*Width/2 : (y+1)*Width/2]
		for i, j := 0, len(row)-1; i < j; i, j = i+1, j-1 {
			row[i], row[j] = row[j], row[i]
		}
		for i := range row {
			row[i] = row[i]<<4 | row[i]>>4
		}
	}
}

// MirrorV 是 `sub_117EB`：第 i 列與第 37 − i 列互換，換 19 次。
func MirrorV(b *Bitmap) {
	for i := 0; i < Height/2; i++ {
		a := b[i*Width/2 : (i+1)*Width/2]
		c := b[(Height-1-i)*Width/2 : (Height-i)*Width/2]
		for k := range a {
			a[k], c[k] = c[k], a[k]
		}
	}
}

// Palette 是 BIOS 模式 0Dh 的 EGA 16 色，索引就是原版的 4 位元值（與抽取器相同）。
var Palette = [16]color.RGBA{
	{0x00, 0x00, 0x00, 0xFF}, {0x00, 0x00, 0xAA, 0xFF},
	{0x00, 0xAA, 0x00, 0xFF}, {0x00, 0xAA, 0xAA, 0xFF},
	{0xAA, 0x00, 0x00, 0xFF}, {0xAA, 0x00, 0xAA, 0xFF},
	{0xAA, 0x55, 0x00, 0xFF}, {0xAA, 0xAA, 0xAA, 0xFF},
	{0x55, 0x55, 0x55, 0xFF}, {0x55, 0x55, 0xFF, 0xFF},
	{0x55, 0xFF, 0x55, 0xFF}, {0x55, 0xFF, 0xFF, 0xFF},
	{0xFF, 0x55, 0x55, 0xFF}, {0xFF, 0x55, 0xFF, 0xFF},
	{0xFF, 0xFF, 0x55, 0xFF}, {0xFF, 0xFF, 0xFF, 0xFF},
}

// Image 把一張圖畫成 40 × 38 的 RGBA（不透明）。
func (b *Bitmap) Image() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, Width, Height))
	for y := 0; y < Height; y++ {
		for x := 0; x < Width; x++ {
			img.SetRGBA(x, y, Palette[b.At(x, y)])
		}
	}
	return img
}

// FromImage 把一張 40 × 38、只用 EGA 16 色的圖轉回索引（從模板 PNG 起編輯時用）。
// 尺寸不符或出現調色盤以外的顏色就失敗，不猜最接近的色。
func FromImage(img image.Image) (Bitmap, error) {
	var b Bitmap
	r := img.Bounds()
	if r.Dx() != Width || r.Dy() != Height {
		return b, fmt.Errorf("圖像尺寸 %d×%d，應為 %d×%d", r.Dx(), r.Dy(), Width, Height)
	}
	for y := 0; y < Height; y++ {
		for x := 0; x < Width; x++ {
			c := color.RGBAModel.Convert(img.At(r.Min.X+x, r.Min.Y+y)).(color.RGBA)
			i := -1
			for k, p := range Palette {
				if p.R == c.R && p.G == c.G && p.B == c.B {
					i = k
					break
				}
			}
			if i < 0 {
				return b, fmt.Errorf("(%d,%d) 的顏色 %v 不是 EGA 16 色", x, y, c)
			}
			b.Set(x, y, i)
		}
	}
	return b, nil
}

// Hex 是保存用的十六進位字串（存檔 v3 的 picture／shield）。
func (b *Bitmap) Hex() string { return hex.EncodeToString(b[:]) }

// ParseHex 讀回 Hex 的結果；長度不是 760 byte 就失敗。
func ParseHex(s string) (Bitmap, error) {
	var b Bitmap
	raw, err := hex.DecodeString(s)
	if err != nil {
		return b, err
	}
	if len(raw) != Bytes {
		return b, fmt.Errorf("自訂圖像 %d byte，應為 %d", len(raw), Bytes)
	}
	copy(b[:], raw)
	return b, nil
}
