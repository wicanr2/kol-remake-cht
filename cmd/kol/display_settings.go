package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"os"

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
		g.displayAwaitFrame = true
	}
}

// 開啟設定時先攔截輸入與自動動畫，返回的那一幀也不送給遊戲。
func (g *game) updateDisplaySettings() bool {
	if g.strings == nil {
		return false
	}
	toggle := inpututil.IsKeyJustPressed(ebiten.KeyF10)
	x, y := g.cursorPosition()
	clicked := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	if toggle || clicked && image.Pt(x, y).In(ui.DisplayControls.Button) {
		g.displaySettings.Open = !g.displaySettings.Open
		g.displaySettings.Cursor = 0
		if g.hd != nil {
			g.displaySettings.Cursor = 1
		}
		g.displaySettings.Problem = false
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
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		g.displaySettings.Cursor = (g.displaySettings.Cursor + 1) % 3
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		g.displaySettings.Cursor = (g.displaySettings.Cursor + 2) % 3
	}
	choice := -1
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter) {
		choice = g.displaySettings.Cursor
	}
	if clicked {
		for i, rect := range []image.Rectangle{ui.DisplayControls.Modes[0], ui.DisplayControls.Modes[1], ui.DisplayControls.Back} {
			if image.Pt(x, y).In(rect) {
				choice = i
			}
		}
	}
	if choice == 2 {
		g.displaySettings.Open = false
		g.displayEvent("back")
	} else if choice >= 0 {
		mode := [2]string{"ega", "hd"}[choice]
		g.displaySettings.Cursor = choice
		err := g.setGraphics(mode)
		g.displaySettings.Problem = err != nil
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		g.displayEvent(mode)
	}
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
	if g.hd != nil {
		g.hd.DrawDisplaySettings(g.displayCanvas.Image(), g.displaySettings)
	} else if err := ui.DrawDisplaySettings(g.displayCanvas.Image(), g.cjk, g.strings, g.displaySettings); err != nil {
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
