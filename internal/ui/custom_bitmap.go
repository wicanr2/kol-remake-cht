package ui

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/kol-remake-cht/internal/bitmapedit"
	"github.com/wicanr2/kol-remake-cht/internal/portrait"
)

// customImages 快取自訂圖像（SPEC-036）：鍵是存檔裡那一串十六進位。
var customImages = map[string]*ebiten.Image{}

// customImage 把一張自訂圖像轉成可畫的影像。
func customImage(h string) (*ebiten.Image, error) {
	if img, ok := customImages[h]; ok {
		return img, nil
	}
	b, err := bitmapedit.ParseHex(h)
	if err != nil {
		return nil, err
	}
	img := ebiten.NewImageFromImage(b.Image())
	customImages[h] = img
	return img, nil
}

// memberPortrait 是一名角色的頭像：改過（picture 非 nil）就畫那一張，否則畫
// CHARPICS 的模板。原版城鎮隊伍格、資料頁抬頭與戰鬥資訊框畫的都是每名角色自己那
// 760 byte（KOLTOWN DS 0x3254 的隊伍副本），沒改過時它與模板逐像素相同。
func memberPortrait(sprites Sprites, portrait int, picture *string) (*ebiten.Image, error) {
	if picture != nil {
		return customImage(*picture)
	}
	return sprites.Get("CHARPICS", portrait)
}

// pieceImage 是立繪的一片：自訂圖（個人盾牌改過）優先，否則讀圖庫。
func pieceImage(sprites Sprites, p portrait.Piece) (*ebiten.Image, error) {
	if p.Custom != "" {
		return customImage(p.Custom)
	}
	return sprites.Get(p.Archive, p.Index)
}
