package main

import (
	"embed"
	"fmt"
	"image"
	"image/color"
	"io/fs"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/kol-remake-cht/internal/font"
)

//go:embed roundedfont/*.png roundedfont/*.json
var roundedFonts embed.FS

type gameFontSettings struct {
	normal, large *font.Font
	tab           int
	problem       string
}

// 與定版原型共用960寬桌面座標，包含上方工具列；不使用遊戲座標。
var desktopDisplay = struct {
	panel   image.Rectangle
	tabs    [2]image.Rectangle
	choices [2]image.Rectangle
	back    image.Rectangle
}{
	image.Rect(400, 182, 942, 600),
	[2]image.Rectangle{image.Rect(420, 240, 650, 278), image.Rect(664, 240, 922, 278)},
	[2]image.Rectangle{image.Rect(420, 330, 922, 374), image.Rect(420, 386, 922, 430)},
	image.Rect(718, 550, 922, 586),
}

func (g *game) setGameFont(mode string) error {
	if mode != "current" && mode != "rounded" {
		return fmt.Errorf("無效的遊戲字型")
	}
	if g.cjk != nil {
		if mode == "rounded" {
			if g.gameFont.normal == nil {
				root, err := fs.Sub(roundedFonts, "roundedfont")
				if err != nil {
					return err
				}
				normal, err := font.LoadFS(root, "rounded-16")
				if err != nil {
					return err
				}
				large, err := font.LoadFS(root, "rounded-32")
				if err != nil {
					return err
				}
				g.gameFont.normal, g.gameFont.large = normal, large
			}
			if err := g.cjk.SetAlternatives(g.gameFont.normal, g.gameFont.large); err != nil {
				return err
			}
		} else if err := g.cjk.SetAlternatives(nil, nil); err != nil {
			return err
		}
	}
	g.toolbar.prefs.GameFont = mode
	if g.logInput {
		fmt.Printf("game-font mode=%s\n", mode)
	}
	g.drawn = false
	return nil
}

func (g *game) displayCursor() int {
	if g.gameFont.tab == 1 {
		if g.toolbar.prefs.GameFont == "rounded" {
			return 1
		}
		return 0
	}
	if g.hd != nil {
		return 1
	}
	return 0
}

func (g *game) toggleDisplayPanel() {
	g.displaySettings.Open = !g.displaySettings.Open
	g.displaySettings.Cursor = g.displayCursor()
	g.displaySettings.Problem = false
	g.gameFont.problem = ""
	g.toolbar.release = true
}

func (g *game) switchDisplayTab(tab int) {
	g.gameFont.tab = tab
	g.displaySettings.Cursor = g.displayCursor()
	g.gameFont.problem = ""
	g.displayEvent(fmt.Sprintf("tab-%d", tab))
}

func (g *game) updateDesktopDisplay(point image.Point, clicked bool) {
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) || inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
		g.switchDisplayTab(1 - g.gameFont.tab)
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
		for i, r := range desktopDisplay.tabs {
			if point.In(r) {
				g.switchDisplayTab(i)
				return
			}
		}
		for i, r := range desktopDisplay.choices {
			if point.In(r) {
				choice = i
			}
		}
		if point.In(desktopDisplay.back) {
			choice = 2
		}
	}
	if choice == 2 {
		g.displaySettings.Open = false
		g.displayEvent("back")
		return
	}
	if choice < 0 {
		return
	}
	g.displaySettings.Cursor = choice
	if g.gameFont.tab == 1 {
		mode := [2]string{"current", "rounded"}[choice]
		if err := g.setGameFont(mode); err != nil {
			g.gameFont.problem = "字模無法載入，保留目前字型。"
		} else {
			g.toolbar.message = ""
			g.rememberToolbar()
			g.gameFont.problem = g.toolbar.message
		}
		g.displayEvent("font-" + mode)
		return
	}
	if choice == 1 && (g.strings == nil || !g.hdInstalled()) {
		g.displaySettings.Problem = true
		g.displayEvent("hd-unavailable")
		return
	}
	mode := [2]string{"ega", "hd"}[choice]
	err := g.setGraphics(mode)
	g.displaySettings.Problem = err != nil
	g.displayEvent(mode)
}

// drawDesktopDisplay 使用工具列原有粉圓，避免改字型時設定頁自己換寬。
func (g *game) drawDesktopDisplay(dst *ebiten.Image) {
	if !g.displaySettings.Open {
		return
	}
	ink := color.RGBA{230, 235, 245, 255}
	gold := color.RGBA{181, 161, 106, 255}
	toolbarRect(dst, desktopDisplay.panel, color.RGBA{17, 29, 39, 255})
	vector.StrokeRect(dst, 400, 182, 542, 418, 2, color.RGBA{158, 148, 120, 255}, false)
	title := ebiten.NewImage(160, 24)
	g.toolbar.face.Draw(title, 0, 0, "畫面設定", 230, 235, 245)
	var titleOp ebiten.DrawImageOptions
	titleOp.GeoM.Scale(1.2, 1.2)
	titleOp.GeoM.Translate(422, 200)
	titleOp.Filter = ebiten.FilterLinear
	dst.DrawImage(title, &titleOp)
	label := func(r image.Rectangle, text string, selected, focus bool, pad int) {
		bg := color.RGBA{42, 52, 70, 255}
		if selected {
			bg = color.RGBA{40, 90, 115, 255}
		}
		toolbarRect(dst, r, bg)
		if focus || selected {
			c := gold
			if focus {
				c = ink
			}
			vector.StrokeRect(dst, float32(r.Min.X), float32(r.Min.Y), float32(r.Dx()), float32(r.Dy()), 1, c, false)
		}
		g.toolbarText(dst, r.Min.X+pad, r.Min.Y+(r.Dy()-20)/2, text)
	}
	for i, text := range []string{"美術", "字型"} {
		label(desktopDisplay.tabs[i], text, i == g.gameFont.tab, false, 16)
	}
	heading, labels := "遊戲中文字型", [2]string{"現有字型", "粉圓柔邊"}
	if g.gameFont.tab == 0 {
		heading, labels = "遊戲美術", [2]string{"EGA 經典", "HD 手繪"}
	}
	g.toolbarText(dst, 422, 296, heading)
	for i, text := range labels {
		active := i == g.displayCursor()
		if active {
			text += "　使用中"
		}
		label(desktopDisplay.choices[i], text, active, i == g.displaySettings.Cursor, 18)
	}
	lines := [3]string{"切換後立即預覽，並記住選擇。", "工具列目前已使用粉圓。", "保留原有字距與操作位置。"}
	if g.gameFont.tab == 0 {
		lines = [3]string{"切換後立即預覽。", "保留隊伍與操作位置。", "左右鍵切分頁，上下鍵選擇。"}
		if g.strings == nil || !g.hdInstalled() {
			lines[0] = "此版本沒有可用的HD素材。"
		}
		if g.displaySettings.Problem {
			lines[0] = "HD無法載入，保留目前畫面。"
		}
	} else if g.strings == nil {
		lines[1] = "英文模式保留原字模。"
	}
	if g.gameFont.problem != "" {
		lines[0] = toolbarMessage(g.gameFont.problem)
	}
	for i, y := range []int{454, 488, 520} {
		g.toolbarText(dst, 422, y, displayPanelMessage(g.toolbar.face, lines[i]))
	}
	label(desktopDisplay.back, "返回遊戲 Esc", false, g.displaySettings.Cursor == 2, 22)
}

func displayPanelMessage(face *font.Font, text string) string {
	if width, _ := face.Measure(text); width <= 500 {
		return text
	}
	runes := []rune(text)
	for len(runes) > 0 {
		runes = runes[:len(runes)-1]
		result := string(runes) + "…"
		if width, _ := face.Measure(result); width <= 500 {
			return result
		}
	}
	return "…"
}
