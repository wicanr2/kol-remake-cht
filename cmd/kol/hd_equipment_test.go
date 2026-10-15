package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/kol-remake-cht/internal/assets"
	"github.com/wicanr2/kol-remake-cht/internal/screen"
)

func TestHDEquipmentWrongTypedSourceKeepsEGACanvas(t *testing.T) {
	lib, err := assets.Load("../../assets/original-derived-png/ega")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "layout.json")
	if err = os.WriteFile(path, []byte("{}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	g := game{canvas: screen.NewCanvasScaled(3), sprites: &spriteCache{lib: lib, m: map[string]*ebiten.Image{}}, hdPortraitPath: path}
	if err = g.enableHD("ega", "../../assets/hd/first-slice", "zh-Hant"); err != nil {
		t.Fatal(err)
	}
	before := g.outputCanvas()
	err = g.setGraphics("hd")
	if err == nil || !strings.Contains(err.Error(), "角色版面來源雜湊不符") {
		t.Fatalf("wrong typed source accepted or failed for another reason: %v", err)
	}
	if g.hd != nil || g.hdCache != nil || g.hdCanvas != nil || g.presentation != nil || g.outputCanvas() != before {
		t.Fatal("failed HD enablement changed EGA or cached a partial pack")
	}
}
