// Package ui 依原版量得的版面繪製書本畫面。
//
// 這一層只負責把狀態畫成像素，不含規則；所有座標與顏色都來自 internal/book 的
// 版面表，那些數字全部由原版畫面機械量得。
package ui

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/kol-remake-cht/internal/book"
	"github.com/wicanr2/kol-remake-cht/internal/charcreate"
	"github.com/wicanr2/kol-remake-cht/internal/font"
)

// Sprites 提供繪製所需的素材。
type Sprites interface {
	Get(archive string, index int) (*ebiten.Image, error)
}

// NewCharacter 畫出 New Character 頁的目前狀態。
func NewCharacter(dst *ebiten.Image, sprites Sprites, small, large *font.Font,
	st *charcreate.State) error {
	bg, err := sprites.Get("BOOKPICS", book.OpenSpread)
	if err != nil {
		return err
	}
	draw(dst, bg, 0, 0)
	heading, err := sprites.Get("BOOKPICS", book.NewCharacterHeading.ImageIndex)
	if err != nil {
		return err
	}
	draw(dst, heading, book.NewCharacterHeading.X, book.NewCharacterHeading.Y)

	// 作用中的欄位標籤畫成棕色。名字／種族／性別／職業層各自只有一列作用中；
	// 擲屬性層是十列屬性同時作用；圖像層沒有作用中的列。三種情形都有原版收據。
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

	for i, l := range book.NewCharacterLabels {
		c := book.ColourInactive
		if active(i) {
			c = book.ColourActive
		}
		text(dst, small, l.X, l.Y, l.Text, c)
	}
	for i, sep := range book.NewCharacterSeparators {
		c := book.ColourInactive
		if active(i) {
			c = book.ColourActive
		}
		text(dst, small, sep.X, sep.Y, sep.Text, c)
	}

	if st.Name != "" {
		text(dst, small, book.NameValue.X, book.NameValue.Y, st.Name, book.ColourValue)
	}
	if st.Stage == charcreate.StageName {
		cursor := ebiten.NewImage(book.NameCursor.Width, book.NameCursor.Height)
		cursor.Fill(color.RGBA{book.ColourValue[0], book.ColourValue[1],
			book.ColourValue[2], 0xFF})
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(
			float64(book.NameCursor.X+len([]rune(st.Name))*small.Advance()),
			float64(book.NameCursor.Y))
		dst.DrawImage(cursor, op)
	}
	if n := st.RaceName(); n != "" {
		text(dst, small, book.RaceValue.X, book.RaceValue.Y, n, book.ColourValue)
	}
	if n := st.SecondaryName(); n != "" {
		text(dst, small, book.SecondaryValue.X, book.SecondaryValue.Y, n, book.ColourValue)
	}
	if n := st.ProfessionName(); n != "" {
		text(dst, small, book.ClassValue.X, book.ClassValue.Y, n, book.ColourValue)
	} else if n := st.PreviewProfessionName(); n != "" {
		text(dst, small, book.ClassValue.X, book.ClassValue.Y, n, book.ColourPreview)
	}

	if st.Rolled {
		values := make([]int, 0, 10)
		for _, v := range st.Primary {
			values = append(values, st.DisplayValue(v))
		}
		for _, v := range st.Derived {
			values = append(values, st.DisplayValue(v))
		}
		for i, v := range values {
			s := fmt.Sprintf("%*d", book.AttributeValue.Cells, v)
			text(dst, small, book.AttributeValue.X,
				book.AttributeValue.Y+i*book.AttributeValue.Step, s, book.ColourValue)
		}
	}

	if items := st.ListItems(); len(items) > 0 {
		sel, err := sprites.Get("BOOKPICS", book.SelectHeading.ImageIndex)
		if err != nil {
			return err
		}
		draw(dst, sel, book.SelectHeading.X, book.SelectHeading.Y)
		for i, item := range items {
			c := book.ColourInactive
			if i == st.Cursor {
				c = book.ColourValue
			}
			text(dst, small, book.RightList.X, book.RightList.Y+i*book.RightList.Step,
				item, c)
		}
	}

	if st.Stage == charcreate.StagePicture {
		sel, err := sprites.Get("BOOKPICS", book.SelectHeading.ImageIndex)
		if err != nil {
			return err
		}
		draw(dst, sel, book.SelectHeading.X, book.SelectHeading.Y)
		for i, line := range book.PictureInstructions {
			text(dst, small, book.RightList.X, book.RightList.Y+i*book.RightList.Step,
				line, book.ColourInactive)
		}
		f := book.PortraitFrame
		fill(dst, f.X, f.Y, f.Width, f.Height, book.ColourInactive)
		fill(dst, f.X+1, f.Y+1, f.Width-2, f.Divider-f.Y-1, book.ColourValue)
		fill(dst, f.X+1, f.Divider+1, f.Width-2, f.Y+f.Height-f.Divider-2,
			book.ColourValue)

		portrait, err := sprites.Get("CHARPICS", st.Portrait)
		if err != nil {
			return err
		}
		draw(dst, portrait, book.PortraitOrigin.X, book.PortraitOrigin.Y)
		small.Draw(dst, book.PortraitName.X, book.PortraitName.Y, st.Name,
			book.PageInk.R, book.PageInk.G, book.PageInk.B)

		act := book.PictureActions
		text(dst, small, act.X, act.Y, act.Text, [3]uint8{act.R, act.G, act.B})

		a := book.PortraitArrows
		text(dst, large, a.X, a.Up, string(a.UpRune), book.ColourValue)
		text(dst, large, a.X, a.Down, string(a.DownRune), book.ColourValue)
	}
	return nil
}

func fill(dst *ebiten.Image, x, y, w, h int, c [3]uint8) {
	if w <= 0 || h <= 0 {
		// 非正尺寸沿用既有流程，包括拒絕行為與負尺寸的正規化。
		rect := ebiten.NewImage(w, h)
		rect.Fill(color.RGBA{c[0], c[1], c[2], 0xFF})
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x), float64(y))
		dst.DrawImage(rect, op)
		return
	}
	// 直接填同一畫布的子矩形，避免每幀建立暫存紋理；邊界與不透明色保持相同。
	dst.SubImage(image.Rect(x, y, x+w, y+h)).(*ebiten.Image).Fill(
		color.RGBA{c[0], c[1], c[2], 0xFF})
}

func draw(dst, src *ebiten.Image, x, y int) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	op.Filter = ebiten.FilterNearest
	dst.DrawImage(src, op)
}

func text(dst *ebiten.Image, f *font.Font, x, y int, s string, c [3]uint8) {
	f.DrawOpaque(dst, x, y, s, c[0], c[1], c[2],
		book.PageInk.R, book.PageInk.G, book.PageInk.B)
}
