package main

import (
	"fmt"
	"image"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/kol-remake-cht/internal/assets"
	"github.com/wicanr2/kol-remake-cht/internal/ending"
	"github.com/wicanr2/kol-remake-cht/internal/flow"
	"github.com/wicanr2/kol-remake-cht/internal/screen"
	"github.com/wicanr2/kol-remake-cht/internal/town"
	"github.com/wicanr2/kol-remake-cht/internal/ui"
)

func (g *game) enableHD(mode, dir, lang string) error {
	g.hdDir, g.hdLanguage = dir, lang
	return g.setGraphics(mode)
}

// setGraphics 完整載入成功後才換呈現；只交換畫布，不重建遊戲或觸碰保存。
func (g *game) setGraphics(mode string) error {
	if mode == "ega" {
		g.hd, g.presentation = nil, nil
		return nil
	}
	if mode != "hd" || g.hdLanguage != "zh-Hant" {
		return fmt.Errorf("graphics須為ega或hd；HD首批只支援zh-Hant")
	}
	if g.hdCache == nil {
		lib, err := assets.LoadHD(g.hdDir)
		if err != nil {
			return fmt.Errorf("HD素材載入失敗：%w", err)
		}
		controls, err := assets.LoadHDInterface(filepath.Join(g.hdDir, "interface"))
		if err != nil {
			return fmt.Errorf("HD介面載入失敗：%w", err)
		}
		events, err := assets.LoadHDEnding(filepath.Join(g.hdDir, "events"))
		if err != nil {
			return fmt.Errorf("HD結局載入失敗：%w", err)
		}
		pages, err := assets.LoadHDPages(filepath.Join(g.hdDir, "pages"))
		if err != nil {
			return fmt.Errorf("HD頁面載入失敗：%w", err)
		}
		scroll, err := assets.LoadHDScroll(filepath.Join(g.hdDir, "scroll"))
		if err != nil {
			return fmt.Errorf("HD法術卷軸載入失敗：%w", err)
		}
		templates, err := assets.LoadHDTemplates(filepath.Join(g.hdDir, "templates"))
		if err != nil {
			return fmt.Errorf("HD角色模板載入失敗：%w", err)
		}
		actions, err := assets.LoadHDActions(filepath.Join(g.hdDir, "actions"))
		if err != nil {
			return fmt.Errorf("HD動作圖示載入失敗：%w", err)
		}
		if g.sprites == nil {
			return fmt.Errorf("HD動作圖示缺少來源PNG庫")
		}
		if err := actions.ValidateSources(g.sprites.lib); err != nil {
			return fmt.Errorf("HD動作圖示來源驗證失敗：%w", err)
		}
		equipment, binding, err := g.loadHDEquipment()
		if err != nil {
			return fmt.Errorf("HD裝備載入失敗：%w", err)
		}
		npcMounts, err := assets.LoadHDNPCMounts(filepath.Join(g.hdDir, "npc-mounts"))
		if err != nil {
			return fmt.Errorf("HD NPC與坐騎載入失敗：%w", err)
		}
		if err := npcMounts.ValidateSources(g.sprites.lib); err != nil {
			return fmt.Errorf("HD NPC與坐騎來源驗證失敗：%w", err)
		}
		hd := ui.NewHD(lib.Facade, lib.Party, lib.Manifest.FacadeRGB, lib.Manifest.PortraitRGB)
		glyphs := make(map[string]ui.HDGlyph, len(controls.Font.Glyphs))
		for ch, g := range controls.Font.Glyphs {
			glyphs[ch] = ui.HDGlyph{X: g.X, Y: g.Y, Advance: g.Advance}
		}
		hd.AttachInterface(controls.Images, glyphs, controls.NPCKey, controls.CrestKey, g.strings)
		hd.AttachEnding(events.Images, events.SourceRGB)
		bookKeys := make(map[string]int, len(pages.Manifest.Books))
		for _, source := range pages.Manifest.Books {
			bookKeys[source.SourceRGB] = source.Index
		}
		var pageSources []ui.HDPageSource
		for _, source := range pages.Sources() {
			pageSources = append(pageSources, ui.HDPageSource{Archive: source.Archive, Index: source.Index, RGB: source.SourceRGB})
		}
		hd.AttachPages(pages.Book, pages.Portraits, bookKeys, pages.Manifest.Background.SourceRGB, pageSources)
		source := scroll.Manifest.Source
		hd.AttachScroll(scroll.Image, ui.HDPageSource{Archive: source.Archive, Index: source.Index, RGB: source.SourceRGB})
		var preserved []ui.HDPageSource
		for _, p := range templates.Manifest.Preserved {
			preserved = append(preserved, ui.HDPageSource{Archive: p.Archive, Index: p.Index, RGB: p.SourceRGB})
		}
		var portraitFrames []ui.HDTemplatePortrait
		for _, p := range templates.Manifest.Portraits {
			portraitFrames = append(portraitFrames, ui.HDTemplatePortrait{HDPageSource: ui.HDPageSource{Archive: p.Archive, Index: p.Index, RGB: p.SourceRGB}, File: p.File, Rect: image.Rect(p.Rect[0], p.Rect[1], p.Rect[2], p.Rect[3])})
		}
		if err := hd.AttachTemplates(templates.Images, preserved, portraitFrames); err != nil {
			return err
		}
		var actionSources []ui.HDActionSource
		for _, p := range actions.Manifest.Sources {
			actionSources = append(actionSources, ui.HDActionSource{HDPageSource: ui.HDPageSource{Archive: p.Archive, Index: p.Index, RGB: p.SourceRGB}, Rect: image.Rect(p.Rect[0], p.Rect[1], p.Rect[2], p.Rect[3])})
		}
		if err := hd.AttachActions(actions.Image, actionSources); err != nil {
			return err
		}
		if err := hd.AttachEquipment(equipment.Images, equipment.Manifest, binding); err != nil {
			return err
		}
		var npcPreserved, npcSources []ui.HDPageSource
		for _, p := range npcMounts.Manifest.Preserved {
			npcPreserved = append(npcPreserved, ui.HDPageSource{Archive: p.Archive, Index: p.Index, RGB: p.SourceRGB})
		}
		for _, s := range npcMounts.Sources() {
			npcSources = append(npcSources, ui.HDPageSource{Archive: s.Archive, Index: s.Index, RGB: s.SourceRGB})
		}
		if err := hd.AttachNPCMounts(npcMounts.Portraits, npcMounts.Mounts, npcPreserved, npcSources); err != nil {
			return err
		}
		canvas := screen.NewCanvasScaled(6)
		g.hdCache, g.hdCanvas = hd, canvas
	}
	g.hd, g.presentation = g.hdCache, g.hdCanvas
	return nil
}

func (g *game) outputCanvas() *screen.Canvas {
	if g.presentation != nil {
		return g.presentation
	}
	return g.canvas
}

func (g *game) logicalPoint(x, y int) (int, int) {
	if g.hd != nil {
		return x >> 1, y >> 1
	}
	return x, y
}

func (g *game) cursorPosition() (int, int) {
	x, y := ebiten.CursorPosition()
	return g.logicalPoint(x, y)
}

func (g *game) presentHD() error {
	if g.hd == nil {
		return nil
	}
	g.presentation.Clear()
	g.presentation.DrawScaled(g.canvas.Image(), 0, 0, 2)
	dst := g.presentation.Image()
	if err := g.hd.ValidatePageSources(g.sprites); err != nil {
		return err
	}
	if g.nativeUIPage() {
		return g.renderNativeUI()
	}
	if g.endingReached() || g.scene == sceneEnding {
		var segments []town.Segment
		if g.ending != nil && g.ending.Current() != nil {
			segments = town.JoinedSegments(g.endBank, g.endTrans, ending.Bank, g.ending.Current().Records)
		}
		return g.hd.Ending(dst, g.sprites, g.ending, segments)
	}
	play := func(s *town.Session) error {
		if s == nil {
			return nil
		}
		if s.InArena() || s.ShowingSheetOutside() {
			return g.hd.Inn(dst, g.sprites, s.Inside)
		}
		if g.quest != nil && g.quest.state != nil {
			return g.hd.Battle(dst, g.sprites, g.quest.state)
		}
		if s.LeftTown {
			return nil
		}
		if s.Mode == town.ModeInside {
			return g.hd.Inn(dst, g.sprites, s.Inside)
		}
		return g.hd.Outdoor(dst, g.sprites, s.Outdoor)
	}
	switch g.scene {
	case sceneInn, sceneBuilding:
		return g.hd.Inn(dst, g.sprites, g.inn)
	case sceneOutdoor:
		return g.hd.Outdoor(dst, g.sprites, g.outdoor)
	case scenePlay:
		return play(g.play)
	case sceneFlow:
		if g.flow == nil {
			return nil
		}
		switch g.flow.Scene {
		case flow.SceneInn:
			return g.hd.Inn(dst, g.sprites, g.flow.Inn)
		case flow.SceneTown:
			if g.leftTown() {
				return play(g.play)
			}
			return play(g.flow.Town)
		}
	}
	return nil
}

// 提示層沿原版只畫到螢幕，不覆蓋批次截圖所用的場景畫布。
func (g *game) drawPresentationNotice(dst *ebiten.Image) {
	if g.hd == nil {
		g.drawNotice(dst)
		return
	}
	if g.noticeCanvas == nil {
		g.noticeCanvas = screen.NewCanvasScaled(3)
	}
	g.drawNotice(g.noticeCanvas.Image())
	var op ebiten.DrawImageOptions
	op.GeoM.Scale(2, 2)
	op.Filter = ebiten.FilterNearest
	dst.DrawImage(g.noticeCanvas.Image(), &op)
}
