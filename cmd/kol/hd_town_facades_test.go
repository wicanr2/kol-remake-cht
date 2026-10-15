package main

import (
	"encoding/json"
	"github.com/wicanr2/kol-remake-cht/internal/assets"
	"github.com/wicanr2/kol-remake-cht/internal/screen"
	"github.com/wicanr2/kol-remake-cht/internal/ui"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBadTownFacadesKeepCurrentPresentation(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", "..", "assets"))
	if err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(root, "hd", "first-slice")
	dir := t.TempDir()
	entries, err := os.ReadDir(source)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.Name() == "town-facades" {
			continue
		}
		if err := os.Symlink(filepath.Join(source, entry.Name()), filepath.Join(dir, entry.Name())); err != nil {
			t.Fatal(err)
		}
	}
	bad := filepath.Join(dir, "town-facades")
	if err := os.Mkdir(bad, 0700); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(source, "town-facades", "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest assets.HDTownFacadesManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	manifest.Count--
	raw, err = json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bad, "manifest.json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	lib, err := assets.Load(filepath.Join(root, "original-derived-png", "ega"))
	if err != nil {
		t.Fatal(err)
	}
	oldHD := &ui.HD{}
	oldCanvas := screen.NewCanvasScaled(6)
	g := &game{hdDir: dir, hdLanguage: "zh-Hant", hdPortraitPath: filepath.Join(root, "original-derived-data", "portrait-layout.json"), hd: oldHD, presentation: oldCanvas, sprites: &spriteCache{lib: lib}}
	err = g.setGraphics("hd")
	if err == nil || !strings.Contains(err.Error(), "城鎮門面") {
		t.Fatalf("wrong failure: %v", err)
	}
	if g.hd != oldHD || g.presentation != oldCanvas || g.hdCache != nil || g.hdCanvas != nil {
		t.Fatal("failed group changed presentation")
	}
}
