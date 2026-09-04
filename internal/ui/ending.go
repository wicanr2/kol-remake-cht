package ui

import (
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/kol-remake-cht/internal/ending"
	"github.com/wicanr2/kol-remake-cht/internal/font"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

// Ending 畫結局的一段（RE-106）：先清畫面，再把 ENDPICS 那一張畫進矩形，
// 最後把 ENDTXT 那兩三筆敘事依它們自己帶的 `~Vxx`／`~Hxx` 座標寫上去。
//
// 原版每一段之前呼叫 `sub_2B845`，依 `word_2DF1E` 選清畫面的方式（`0x2B84E` 的
// 三條路對到三種顯示模式）。remake 只有一種模式，所以直接填黑。
//
// segments 由呼叫端展開——文字庫與譯文表走同一套 `town.UISegments`，
// 這一層不知道那幾筆是英文還是中文。
func Ending(dst *ebiten.Image, sprites Sprites, f *font.Font,
	s *ending.State, segments []town.Segment) error {
	panel := s.Current()
	if panel == nil {
		return nil
	}
	b := dst.Bounds()
	fill(dst, 0, 0, b.Dx(), b.Dy(), town.ColourBoxFill)
	pic, err := sprites.Get(ending.Archive, panel.Picture)
	if err != nil {
		return err
	}
	draw(dst, pic, panel.Box.X(), panel.Box.Y())
	for _, seg := range segments {
		// 原版的敘事是大寫畫上去的，與訊息區同一種處理（RE-074）。
		f.Draw(dst, seg.Col*8, seg.Row, strings.ToUpper(seg.Text),
			EndingInk[0], EndingInk[1], EndingInk[2])
	}
	return nil
}

// EndingInk 是結局敘事的字色。原版整段用同一個亮色寫在黑底上
// （`sub_1A8FD` 沒有換色碼，字色沿用 `word_2DF1C`，`0x2B4FB` 設成 0xFF）。
var EndingInk = [3]uint8{0xFF, 0xFF, 0xFF}
