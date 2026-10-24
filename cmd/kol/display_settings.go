package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/wicanr2/kol-remake-cht/internal/screen"
	"github.com/wicanr2/kol-remake-cht/internal/ui"
)

func (g *game) displayEvent(action string) {
	if g.logInput {
		// snapshot 平時壓縮重複列；這裡取完整資料並恢復記錄游標。
		party, stage, walls, detail := g.lastParty, g.lastStage, g.wallsFor, g.traceDetail
		g.lastParty, g.lastStage, g.wallsFor, g.traceDetail = nil, -1, nil, true
		data, _ := json.Marshal(g.snapshot())
		g.lastParty, g.lastStage, g.wallsFor, g.traceDetail = party, stage, walls, detail
		fmt.Printf("display action=%s hd=%t state=%x\n", action, g.hd != nil, sha256.Sum256(data))
		g.musicDiagnostics()
		g.displayAwaitFrame = true
		g.toolbar.awaitFrame = true
	}
}

// 開啟設定時先攔截輸入與自動動畫，返回的那一幀也不送給遊戲。
func (g *game) updateDisplaySettings() bool {
	toggle := inpututil.IsKeyJustPressed(ebiten.KeyF10)
	x, y := g.cursorPosition()
	clicked := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	if toggle || clicked && image.Pt(x, y).In(ui.DisplayControls.Button) {
		g.toggleDisplayPanel()
		g.displayEvent("toggle")
		return true
	}
	if !g.displaySettings.Open {
		return false
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		g.displaySettings.Open = false
		g.displayEvent("back")
		return true
	}
	x, y = ebiten.CursorPosition()
	w, _ := g.desktopSize()
	g.updateDesktopDisplay(image.Pt(x*960/w, y*960/w), clicked)
	return true
}

func (g *game) drawDisplaySettings(dst *ebiten.Image) error {
	if !g.interactive || g.strings == nil {
		return nil
	}
	if g.displayCanvas == nil {
		g.displayCanvas = screen.NewCanvasScaled(6)
	}
	g.displayCanvas.Clear()
	g.displaySettings.HD = g.hd != nil
	g.displaySettings.HDUnavailable = !g.hdInstalled()
	// 桌面面板統一由工具列呈現；遊戲層只保留下緣的既有入口。
	state := g.displaySettings
	state.Open = false
	if g.hd != nil {
		g.hd.DrawDisplaySettings(g.displayCanvas.Image(), state)
		g.hd.DrawMusicControl(g.displayCanvas.Image(), g.musicEnabled)
	} else if err := ui.DrawDisplaySettings(g.displayCanvas.Image(), g.cjk, g.strings, state); err != nil {
		return err
	} else if err := ui.DrawMusicControl(g.displayCanvas.Image(), g.cjk, g.musicEnabled); err != nil {
		return err
	}
	var op ebiten.DrawImageOptions
	op.Filter = ebiten.FilterNearest
	dst.DrawImage(g.displayCanvas.Image(), &op)
	if g.displayAwaitFrame {
		fmt.Printf("display-frame hd=%t open=%t cursor=%d\n", g.hd != nil, g.displaySettings.Open, g.displaySettings.Cursor)
		g.displayAwaitFrame = false
	}
	return nil
}

func (g *game) hdInstalled() bool {
	if available, checked := g.hdInstallCache[g.hdDir]; checked {
		return available
	}
	if g.hdInstallCache == nil {
		g.hdInstallCache = map[string]bool{}
	}
	for _, dir := range []string{"", "interface", "pages", "events", "scroll", "templates", "actions", "equipment", "npc-mounts", "monsters", "town-facades", "town-crests", "town-maps", "scene-events", "motion"} {
		info, err := os.Stat(filepath.Join(g.hdDir, dir, "manifest.json"))
		if err != nil || !info.Mode().IsRegular() {
			g.hdInstallCache[g.hdDir] = false
			return false
		}
	}
	g.hdInstallCache[g.hdDir] = true
	return true
}
