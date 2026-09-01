package book

import (
	"encoding/json"
	"fmt"
	"image"
	"os"
)

// OpenAnimation 是翻書動畫的參數表，出自原版 `KOLBOOK.EXE` 的 `sub_114F5`（RE-048）。
//
// 每一步先把畫面上某一段 byte 欄整體左移一欄，再把來源畫面的一欄複製過來；欄的單位是
// EGA 的 byte column（8 像素），而且只動固定的那幾列。這裡不含任何原版像素，只有座標。
type OpenAnimation struct {
	Schema   string `json:"schema"`
	Geometry struct {
		FirstRow     int `json:"first_row"`
		RowCount     int `json:"row_count"`
		ColumnPixels int `json:"column_pixels"`
	} `json:"geometry"`
	DelayTicks int        `json:"delay_ticks_per_step"`
	Steps      []OpenStep `json:"steps"`
}

// OpenStep 是動畫的一步。
type OpenStep struct {
	Stage        string `json:"stage"`
	ScrollColumn int    `json:"scroll_column"`
	ScrollSpan   int    `json:"scroll_span"`
	FillColumn   int    `json:"fill_column"`
	SourceColumn int    `json:"source_column"`
}

// LoadOpenAnimation 讀取動畫參數表並做基本一致性檢查。
func LoadOpenAnimation(path string) (*OpenAnimation, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("讀取翻書動畫失敗：%w", err)
	}
	var a OpenAnimation
	if err := json.Unmarshal(raw, &a); err != nil {
		return nil, fmt.Errorf("解析翻書動畫失敗：%w", err)
	}
	if a.Schema != "kol-book-open-animation-v1" {
		return nil, fmt.Errorf("非預期的翻書動畫 schema：%s", a.Schema)
	}
	if len(a.Steps) == 0 || a.Geometry.ColumnPixels <= 0 || a.DelayTicks <= 0 {
		return nil, fmt.Errorf("翻書動畫內容不完整")
	}
	return &a, nil
}

// OpenPlayer 在 CPU 端的像素緩衝上重播動畫。原版動的是顯示記憶體，這裡動的是同一份
// 畫面的 RGBA 複本，之後再整張上傳，行為等價而且可以逐步取樣。
type OpenPlayer struct {
	anim   *OpenAnimation
	screen *image.RGBA
	source *image.RGBA
	index  int
}

// NewOpenPlayer 以封面為起始畫面、目錄頁為來源建立播放器。
func NewOpenPlayer(anim *OpenAnimation, cover, source *image.RGBA) (*OpenPlayer, error) {
	if cover.Bounds() != source.Bounds() {
		return nil, fmt.Errorf("起始畫面與來源畫面尺寸不同")
	}
	screen := image.NewRGBA(cover.Bounds())
	copy(screen.Pix, cover.Pix)
	return &OpenPlayer{anim: anim, screen: screen, source: source}, nil
}

// Image 是目前的畫面。
func (p *OpenPlayer) Image() *image.RGBA { return p.screen }

// Index 是已完成的步數。
func (p *OpenPlayer) Index() int { return p.index }

// Done 表示動畫已經播完。
func (p *OpenPlayer) Done() bool { return p.index >= len(p.anim.Steps) }

// Step 前進一步；已播完時不做事並回傳 false。
func (p *OpenPlayer) Step() bool {
	if p.Done() {
		return false
	}
	s := p.anim.Steps[p.index]
	unit := p.anim.Geometry.ColumnPixels * 4
	stride := p.screen.Stride
	first, count := p.anim.Geometry.FirstRow, p.anim.Geometry.RowCount
	for row := first; row < first+count; row++ {
		base := row * stride
		if s.ScrollSpan > 0 {
			dst := base + s.ScrollColumn*unit
			src := dst + unit
			copy(p.screen.Pix[dst:dst+s.ScrollSpan*unit],
				p.screen.Pix[src:src+s.ScrollSpan*unit])
		}
		dst := base + s.FillColumn*unit
		src := base + s.SourceColumn*unit
		copy(p.screen.Pix[dst:dst+unit], p.source.Pix[src:src+unit])
	}
	p.index++
	return true
}
