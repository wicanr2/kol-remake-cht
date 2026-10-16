package world

// BankNamer 從 UI 文字庫取標籤文字。原版直接畫 KOLDATA 記錄 `0xD5 + 時段`
// 與 `0xDA + 地形碼`，而且**保留記錄自帶的空白**——那是用來置中的。
type BankNamer struct {
	Padded func(index int) string
	Zh     func(index int) string
}

// TimeText 回傳時段那一行。
func (b BankNamer) TimeText(index int) string { return b.text(TimeRecordBase + index) }

// TerrainText 回傳地形那一行。
func (b BankNamer) TerrainText(code int) string { return b.text(TerrainRecordBase + code) }

// CampingText uses the same UI-bank translation as the time and terrain labels.
func (b BankNamer) CampingText() string { return b.text(CampRecord) }

func (b BankNamer) text(record int) string {
	if b.Zh != nil {
		if zh := b.Zh(record); zh != "" {
			return zh
		}
	}
	if b.Padded == nil {
		return ""
	}
	return b.Padded(record)
}
