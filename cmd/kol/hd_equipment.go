package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/wicanr2/kol-remake-cht/internal/assets"
	"github.com/wicanr2/kol-remake-cht/internal/hdart"
	"github.com/wicanr2/kol-remake-cht/internal/portrait"
)

func (g *game) loadHDEquipment() (*assets.HDEquipment, hdart.Binding, error) {
	var binding hdart.Binding
	pack, err := assets.LoadHDEquipment(filepath.Join(g.hdDir, "equipment"))
	if err != nil {
		return nil, binding, err
	}
	if g.sprites == nil {
		return nil, binding, fmt.Errorf("HD裝備缺少來源PNG庫")
	}
	if err = pack.ValidateSources(g.sprites.lib); err != nil {
		return nil, binding, err
	}
	raw, err := os.ReadFile(g.hdPortraitPath)
	if err != nil {
		return nil, binding, fmt.Errorf("HD裝備角色版面載入失敗：%w", err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != pack.Manifest.LayoutSHA256 {
		return nil, binding, fmt.Errorf("HD裝備角色版面來源雜湊不符")
	}
	layout, err := portrait.Load(g.hdPortraitPath)
	if err != nil {
		return nil, binding, err
	}
	if len(layout.Body) != 6 || len(layout.Hands) != 12 {
		return nil, binding, fmt.Errorf("HD裝備六型與雙手版面數量不符")
	}
	for i, body := range layout.Body {
		binding.BaseOrigins[i] = body.X
	}
	for i, hand := range layout.Hands {
		b := pack.Manifest.HandPaint[strconv.Itoa(i)]
		binding.SourceHands[i] = [4]int{hand.X*8 + b[0], hand.Y + b[1], hand.X*8 + b[2], hand.Y + b[3]}
	}
	for _, s := range pack.Manifest.Sources {
		if s.Archive == "BOOKDATA" && s.Index == 4 {
			binding.DefaultShieldRGB = s.RGB
		}
	}
	return pack, binding, nil
}
