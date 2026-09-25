package ui

import (
	"strings"

	"github.com/wicanr2/kol-remake-cht/internal/roster"
)

// displayName 只在中文繪製端使用。測試用的簡易 Translator 可以只實作 T；
// 正式字串表另提供 DisplayName，避免改動遊戲狀態裡的原始名稱。
func displayName(t Translator, en string) string {
	if names, ok := t.(interface{ DisplayName(string) string }); ok {
		return names.DisplayName(en)
	}
	return en
}

// spellGloss 是法術名的中文注解（SPEC-034）；正式字串表才有，測試替身回空字串。
func spellGloss(t Translator, name string) string {
	if g, ok := t.(interface{ SpellGloss(string) string }); ok {
		return g.SpellGloss(name)
	}
	return ""
}

// locationCJK 是名冊右欄的中文所在地：所在地鍵（建築名前 10 個大寫字元）與
// 三字城鎮代碼分別查顯示譯名（SPEC-033）。原始欄位仍用於招募比對與存檔。
func locationCJK(t Translator, s roster.Slot) string {
	return displayName(t, strings.TrimSpace(s.Place)) + " " + displayName(t, strings.TrimSpace(s.Town))
}
