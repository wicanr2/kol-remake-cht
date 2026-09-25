package ui

import (
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/roster"
)

type namesStub map[string]string

func (n namesStub) T(en string) string { return en }

func (n namesStub) DisplayName(en string) string {
	if zh, ok := n[en]; ok {
		return zh
	}
	return en
}

func TestLocationCJKTranslatesPlaceAndTownSeparately(t *testing.T) {
	names := namesStub{"HALFWAY HO": "半途之家", "MIN": "霍布"}
	slot := roster.Slot{Place: "HALFWAY HO", Town: "MIN"}
	if got := locationCJK(names, slot); got != "半途之家 霍布" {
		t.Fatalf("locationCJK = %q", got)
	}
	// 原始欄位不得被改寫：招募比對與存檔仍用英文鍵。
	if slot.Location() != "HALFWAY HO MIN" {
		t.Fatalf("Location 被改動：%q", slot.Location())
	}
}

func TestMovementCJKAddsRate(t *testing.T) {
	for in, want := range map[string]string{"U": "U（20）", "E": "E（4）", "": ""} {
		if got := movementCJK(in); got != want {
			t.Errorf("movementCJK(%q) = %q，應為 %q", in, got, want)
		}
	}
}
