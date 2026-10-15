package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/kol-remake-cht/internal/assets"
	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/font"
	"github.com/wicanr2/kol-remake-cht/internal/screen"
	"github.com/wicanr2/kol-remake-cht/internal/ui"
	"os"
	"path/filepath"
	"testing"
)

func TestMonsterUnderlayErrorRestoresSourceCanvasAndObserver(t *testing.T) {
	lib, err := assets.Load("../../assets/original-derived-png/ega")
	if err != nil {
		t.Fatal(err)
	}
	g := &game{scene: scenePlay, canvas: screen.NewCanvasScaled(3), sprites: &spriteCache{lib: lib, m: map[string]*ebiten.Image{}}, hdPortraitPath: "../../assets/original-derived-data/portrait-layout.json"}
	original, err := g.sprites.Get("MONPICS", 0)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err = os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(`{"schema":"kol-ega-production-manifest-test","archives":[]}`), 0644); err != nil {
		t.Fatal(err)
	}
	broken, err := assets.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	g.sprites.lib = broken
	g.quest = &questBattle{state: &battle.State{Stage: battle.StageEncounter, Site: &battle.Site{Monster: 0, Count: 1}}}
	base := g.canvas
	previous := font.DrawObserver
	defer func() { font.DrawObserver = previous }()
	called := false
	font.DrawObserver = func(string, string) { called = true }
	if err := g.renderMonsterUnderlay([]ui.HDMonsterPlacement{{Index: 0, Original: original}}); err == nil {
		t.Fatal("missing underlay source accepted")
	}
	if g.canvas != base || g.sprites.m["MONPICS/0"] != original || font.DrawObserver == nil {
		t.Fatal("error did not restore canvas, source or observer")
	}
	font.DrawObserver("test", "test")
	if !called {
		t.Fatal("wrong observer restored")
	}
}
