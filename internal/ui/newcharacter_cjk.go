package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/kol-remake-cht/internal/book"
	"github.com/wicanr2/kol-remake-cht/internal/charcreate"
	"github.com/wicanr2/kol-remake-cht/internal/font"
	"github.com/wicanr2/kol-remake-cht/internal/i18n"
)

// Translator 提供 UI 字串的譯文。
type Translator interface {
	T(en string) string
}

// NewCharacterCJK 以中文模式畫出 New Character 頁。
//
// 這個模式**不對拍**：原版沒有中文畫面，版面是 remake 自訂的。畫布放大到原版的兩倍，
// 書本美術以最近鄰放大，文字改用 16 px 的中文字模格。英文嚴格對拍模式仍走
// NewCharacter，兩者共用同一個狀態機。
func NewCharacterCJK(dst *ebiten.Image, sprites Sprites, f *font.Font,
	t Translator, st *charcreate.State) error {
	s := book.CJKScale
	l := book.CJKLayout
	bg, err := sprites.Get("BOOKPICS", book.OpenSpread)
	if err != nil {
		return err
	}
	drawScaled(dst, bg, 0, 0, s)

	text(dst, f, l.HeadingX, l.HeadingY, t.T("New character"), book.ColourValue)

	active := func(i int) bool {
		switch st.Stage {
		case charcreate.StageName:
			return i == book.RowName
		case charcreate.StageRace:
			return i == book.RowRace
		case charcreate.StageSecondary:
			return i == book.RowSecondary
		case charcreate.StageProfession:
			return i == book.RowClass
		case charcreate.StageStats:
			return i >= book.RowFirstAttribute
		}
		return false
	}
	rowY := func(i int) int {
		if i < book.RowFirstAttribute {
			return l.FieldY + i*l.FieldStep
		}
		return l.AttrY + (i-book.RowFirstAttribute)*l.AttrStep
	}

	values := make([]string, len(book.NewCharacterLabels))
	values[book.RowName] = st.Name
	values[book.RowRace] = t.T(st.RaceName())
	values[book.RowSecondary] = t.T(st.SecondaryName())
	if n := st.ProfessionName(); n != "" {
		values[book.RowClass] = t.T(n)
	}
	preview := ""
	if n := st.PreviewProfessionName(); n != "" {
		preview = t.T(n)
	}
	if st.Rolled {
		all := make([]int, 0, 10)
		for _, v := range st.Primary {
			all = append(all, st.DisplayValue(v))
		}
		for _, v := range st.Derived {
			all = append(all, st.DisplayValue(v))
		}
		for i, v := range all {
			values[book.RowFirstAttribute+i] = fmt.Sprintf("%d", v)
		}
	}

	for i, label := range book.NewCharacterLabels {
		c := book.ColourInactive
		if active(i) {
			c = book.ColourActive
		}
		y := rowY(i)
		text(dst, f, l.FieldX, y, t.T(label.Text), c)
		text(dst, f, l.ColonX, y, ":", c)
		if values[i] != "" {
			text(dst, f, l.ValueX, y, values[i], book.ColourValue)
		} else if i == book.RowClass && preview != "" {
			text(dst, f, l.ValueX, y, preview, book.ColourPreview)
		}
	}

	if st.Stage == charcreate.StageName {
		width, _ := f.Measure(st.Name)
		fill(dst, l.ValueX+width, rowY(book.RowName)+f.Height()-2, book.CJKCell/2, 2,
			book.ColourValue)
	}

	if items := st.ListItems(); len(items) > 0 {
		text(dst, f, l.RightX, l.RightHeadingY, t.T("Select one"), book.ColourValue)
		for i, item := range items {
			c := book.ColourInactive
			if i == st.Cursor {
				c = book.ColourValue
			}
			text(dst, f, l.RightX, l.RightListY+i*l.RightStep, t.T(item), c)
		}
	}

	if st.Stage == charcreate.StagePicture {
		text(dst, f, l.RightX, l.RightHeadingY, t.T("Select one"), book.ColourValue)
		for i, line := range book.PictureInstructions {
			text(dst, f, l.RightX, l.InstructionsY+i*l.RightStep, t.T(line),
				book.ColourInactive)
		}
		text(dst, f, l.ActionsX, l.ActionsY,
			t.T("CANCEL")+"　"+t.T("ACCEPT"), book.ColourActive)

		fill(dst, l.PortraitFrameX, l.PortraitFrameY, l.PortraitFrameW, l.PortraitFrameH,
			book.ColourInactive)
		fill(dst, l.PortraitFrameX+2, l.PortraitFrameY+2, l.PortraitFrameW-4,
			l.PortraitDivider-l.PortraitFrameY-2, book.ColourValue)
		fill(dst, l.PortraitFrameX+2, l.PortraitDivider+2, l.PortraitFrameW-4,
			l.PortraitFrameY+l.PortraitFrameH-l.PortraitDivider-4, book.ColourValue)
		portrait, err := sprites.Get("CHARPICS", st.Portrait)
		if err != nil {
			return err
		}
		drawScaled(dst, portrait, l.PortraitX, l.PortraitY, s)
		f.Draw(dst, l.PortraitNameX, l.PortraitNameY, st.Name,
			book.PageInk.R, book.PageInk.G, book.PageInk.B)
	}
	return nil
}

// TranslatorFromTable 讓 *i18n.Table 直接當 Translator 使用。
func TranslatorFromTable(t *i18n.Table) Translator { return t }

func drawScaled(dst, src *ebiten.Image, x, y, scale int) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(scale), float64(scale))
	op.GeoM.Translate(float64(x), float64(y))
	op.Filter = ebiten.FilterNearest
	dst.DrawImage(src, op)
}

var _ = color.RGBA{}
