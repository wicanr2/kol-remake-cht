package ui

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
)

type HDActionSource struct {
	HDPageSource
	Rect image.Rectangle
}
type hdActions struct {
	icons   map[string]*ebiten.Image
	sources []HDPageSource
}

// AttachActions 接受已通過資產驗證的完整圖集，先驗完再替換子組。
func (h *HD) AttachActions(atlas image.Image, sources []HDActionSource) error {
	if atlas == nil || atlas.Bounds() != image.Rect(0, 0, 1296, 960) || len(sources) != 68 || h.pages == nil {
		return fmt.Errorf("HD動作圖示需要完整圖集與頁面")
	}
	ids := []int{}
	for _, span := range [][2]int{{3, 27}, {31, 31}, {36, 69}, {74, 81}} {
		for i := span[0]; i <= span[1]; i++ {
			ids = append(ids, i)
		}
	}
	seen := map[string]bool{}
	for i, s := range sources {
		x, y := i%9*144, i/9*120
		if s.Archive != "KOLMPICS" || s.Index != ids[i] || len(s.RGB) != 64 || seen[s.RGB] || s.Rect != image.Rect(x, y, x+144, y+120) {
			return fmt.Errorf("HD動作來源或矩形不符：第%d格", i)
		}
		seen[s.RGB] = true
	}
	pic := ebiten.NewImageFromImage(atlas)
	a := &hdActions{icons: map[string]*ebiten.Image{}}
	for _, s := range sources {
		a.icons[s.RGB] = pic.SubImage(s.Rect).(*ebiten.Image)
		a.sources = append(a.sources, s.HDPageSource)
	}
	h.actions = a
	h.pages.validated = false
	return nil
}

func (h *HD) validateActions(sprites Sprites) error {
	if h.actions == nil {
		return nil
	}
	for _, s := range h.actions.sources {
		pic, err := sprites.Get(s.Archive, s.Index)
		if err != nil {
			return err
		}
		if pic.Bounds().Dx() != 24 || pic.Bounds().Dy() != 20 || h.fingerprint(pic) != s.RGB {
			return fmt.Errorf("HD動作第%d圖實際sprite不符", s.Index)
		}
	}
	return nil
}
