package ui

import (
	"crypto/sha256"
	"encoding/hex"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

// HD 在最後1920×1200畫布呈現核准圖片與介面，不改動原畫布、字型與規則狀態。
type HD struct {
	facade    *ebiten.Image
	portraits [6]*ebiten.Image
	matching  map[string]int
	facadeKey string
	keys      map[*ebiten.Image]string
	style     *hdInterface
	samples   map[hdSampleKey]*ebiten.Image
}

func NewHD(facade, party image.Image, facadeKey string, portraitKeys []string) *HD {
	h := &HD{facade: ebiten.NewImageFromImage(facade), matching: map[string]int{},
		facadeKey: facadeKey, keys: map[*ebiten.Image]string{}}
	atlas := ebiten.NewImageFromImage(party)
	for i, key := range portraitKeys {
		x, y := i%3*512, i/3*512
		h.portraits[i] = atlas.SubImage(image.Rect(x, y, x+512, y+512)).(*ebiten.Image)
		h.matching[key] = i
	}
	return h
}

func (h *HD) fingerprint(img *ebiten.Image) string {
	if key, ok := h.keys[img]; ok {
		return key
	}
	b := img.Bounds()
	pixels := make([]byte, b.Dx()*b.Dy()*4)
	img.ReadPixels(pixels)
	rgb := make([]byte, 0, b.Dx()*b.Dy()*3)
	for i := 0; i < len(pixels); i += 4 {
		rgb = append(rgb, pixels[i:i+3]...)
	}
	sum := sha256.Sum256(rgb)
	key := hex.EncodeToString(sum[:])
	h.keys[img] = key
	return key
}

func (h *HD) drawHD(dst, src *ebiten.Image, rect image.Rectangle) {
	sampled := h.sampledHD(src, rect.Dx(), rect.Dy())
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(float64(rect.Min.X), float64(rect.Min.Y))
	op.Filter = ebiten.FilterNearest
	dst.DrawImage(sampled, &op)
}

func (h *HD) party(dst *ebiten.Image, sprites Sprites, members []town.Member) error {
	l := town.CJKInn
	for i, m := range members {
		pic, err := memberPortrait(sprites, m.Portrait, m.Picture)
		if err != nil {
			return err
		}
		index, ok := h.matching[h.fingerprint(pic)]
		if !ok {
			continue
		}
		x, y := (l.PartyX+i*l.PartyStep)*2+4, l.PartyY*2+4
		h.drawHD(dst, h.portraits[index], image.Rect(x, y, x+(l.PartyWidth-1)*2-8, y+l.PortraitBoxHeight*2-8))
	}
	return nil
}

func (h *HD) Inn(dst *ebiten.Image, sprites Sprites, in *town.Inn) error {
	if in == nil || in.SpellEdit != nil {
		return nil
	}
	if h.style != nil && in.Sheet == nil && in.List == nil && in.Steed == nil && len(in.Detail) == 0 && in.DialogRow == 0 {
		return h.innHD(dst, sprites, in)
	}
	return h.party(dst, sprites, sheetVisibleParty(in))
}

func (h *HD) Outdoor(dst *ebiten.Image, sprites Sprites, out *town.Outdoor) error {
	if out == nil {
		return nil
	}
	if h.style != nil {
		return h.outdoorHD(dst, sprites, out)
	}
	if err := h.party(dst, sprites, out.Party); err != nil {
		return err
	}
	if door := out.AtDoor(); door != nil && !out.AtEdge {
		pic, err := sprites.Get(out.CrestArchive(), door.Picture)
		if err != nil {
			return err
		}
		if pic.Bounds().Dx() == 192 && pic.Bounds().Dy() == 96 && h.fingerprint(pic) == h.facadeKey {
			l := town.CJKOutdoor
			h.drawHD(dst, h.facade, image.Rect(l.ViewX*2, l.ViewY*2, (l.ViewX+192*l.Scale)*2, (l.ViewY+96*l.Scale)*2))
		}
	}
	return nil
}

func (h *HD) Battle(dst *ebiten.Image, sprites Sprites, s *battle.State) error {
	if s == nil {
		return nil
	}
	d, shown := s.EquipmentPanel()
	if !shown || d.Monster || d.Equipment {
		return nil
	}
	if h.style != nil && s.Stage == battle.StageField && s.Plan.Prompt == battle.PromptCommand && len(s.Plan.Bar) > 0 {
		return h.battleHD(dst, sprites, s)
	}
	pic, err := memberPortrait(sprites, d.Portrait, d.Picture)
	if err != nil {
		return err
	}
	i, ok := h.matching[h.fingerprint(pic)]
	if !ok {
		return nil
	}
	// 姓名底緣為78。清除舊立繪時保留完整文字，不沿原型的70像素切線。
	dst.SubImage(image.Rect(100, 80, 816, 454)).(*ebiten.Image).Fill(color.Black)
	h.drawHD(dst, h.portraits[i], image.Rect(260, 80, 628, 448))
	return nil
}
