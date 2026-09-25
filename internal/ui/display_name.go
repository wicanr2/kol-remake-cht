package ui

// displayName 只在中文繪製端使用。測試用的簡易 Translator 可以只實作 T；
// 正式字串表另提供 DisplayName，避免改動遊戲狀態裡的原始名稱。
func displayName(t Translator, en string) string {
	if names, ok := t.(interface{ DisplayName(string) string }); ok {
		return names.DisplayName(en)
	}
	return en
}
