package main

import (
	"image"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/font"
)

func TestGameFontPreferenceMigrationAndReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "toolbar.json")
	if err := os.WriteFile(path, []byte(`{"enemy_multiplier":3,"stats_floor":"110","skills_floor":"200","gold_floor":"9000"}`), 0600); err != nil {
		t.Fatal(err)
	}
	p, err := loadToolbarPreferences(path)
	if err != nil || p.GameFont != "current" || p.EnemyMultiplier != 3 {
		t.Fatal(p, err)
	}
	p.GameFont = "rounded"
	if err := saveToolbarPreferences(path, p); err != nil {
		t.Fatal(err)
	}
	got, err := loadToolbarPreferences(path)
	if err != nil || got != p {
		t.Fatal(got, err)
	}
	if err := os.WriteFile(path, []byte(`{"game_font":"unknown"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadToolbarPreferences(path); err == nil {
		t.Fatal("未知字型未拒絕")
	}
}

func TestGameFontEmbeddedSelectionAndPanelFit(t *testing.T) {
	f, err := font.Load("../../assets/ui-font", "zh-hant-16")
	if err != nil {
		t.Fatal(err)
	}
	g := &game{cjk: f}
	g.toolbar.prefs = defaultToolbarPreferences()
	if err := g.setGameFont("rounded"); err != nil {
		t.Fatal(err)
	}
	if g.gameFont.normal == nil || g.gameFont.large == nil || g.displayCursor() != 0 {
		t.Fatal("字模或美術選項未初始化")
	}
	g.gameFont.tab = 1
	if g.displayCursor() != 1 {
		t.Fatal("字型游標不同步")
	}
	if err := g.setGameFont("current"); err != nil {
		t.Fatal(err)
	}
	if g.displayCursor() != 0 {
		t.Fatal("無法回到原字型")
	}
	root, _ := fs.Sub(toolbarFonts, "toolbarfont")
	face, err := font.LoadFS(root, "toolbar")
	if err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"畫面設定", "美術", "字型", "遊戲中文字型", "粉圓柔邊　使用中", "現有字型　使用中", "切換後立即預覽，並記住選擇。", "工具列目前已使用粉圓。", "保留原有字距與操作位置。", "左右鍵切分頁，上下鍵選擇。", "此版本沒有可用的HD素材。", "HD無法載入，保留目前畫面。", "字模無法載入，保留目前字型。", "英文模式保留原字模。"} {
		w, h := face.Measure(text)
		if w > 500 || h > 24 {
			t.Fatal("設定文字超界", text, w, h)
		}
		for _, ch := range text {
			if !face.HasRune(ch) {
				t.Fatalf("設定缺字：%c", ch)
			}
		}
	}
	if !desktopDisplay.panel.In(imageForPanel()) {
		t.Fatal("面板超出最小桌面")
	}
}

// 最小桌面為隱藏工具列後960×600；面板不依賴工具列高度。
func imageForPanel() image.Rectangle { return image.Rect(0, 0, 960, 600) }
