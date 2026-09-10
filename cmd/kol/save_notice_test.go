package main

import (
	"encoding/json"
	"os"
	"testing"
	"unicode"
)

func TestSaveNoticesHaveChineseGlyphs(t *testing.T) {
	raw, err := os.ReadFile("../../assets/ui-font/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Fonts []struct {
			Name   string `json:"name"`
			Glyphs []struct {
				Codepoint int `json:"codepoint"`
			} `json:"glyphs"`
		} `json:"fonts"`
	}
	if err = json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	available := map[rune]bool{}
	for _, f := range manifest.Fonts {
		if f.Name == "zh-hant-16" {
			for _, g := range f.Glyphs {
				available[rune(g.Codepoint)] = true
			}
		}
	}
	for kind, n := range saveNotices {
		for _, s := range n.zh {
			for _, r := range s {
				if !unicode.IsSpace(r) && !available[r] {
					t.Errorf("%s 提示缺字：%c", kind, r)
				}
			}
		}
	}
}
