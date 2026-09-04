package town

// 任務一覽（RE-111）：角色資料頁上按第 19 個圖示（`0x161E8` 比的 0x13），
// 畫面在資料頁上蓋一層任務徽章——**只列回報過的那些**。
// 一組（`word_2DF70` 的那五格）一頁，按一個鍵換下一組。

// QuestBadgeArchive 是徽章的圖檔前綴。原版的檔名樣板是 `awd%d%s`
// （DGROUP `0x17C4`），`%d` 是組別加一、`%s` 由顯示模式選 `pics`／`pcga`／`herc`，
// 所以 EGA 下第 0 組是 `AWD1PICS`。二十四張各 24×21 的 4-bit 索引圖（RE-057）。
const QuestBadgeArchive = "AWD%dPICS"

// 徽章的版面（`sub_168F8`）。左欄與上列都是**字格**，乘八才是像素；
// 框比圖大一圈：左上各退 3 欄／2 列，右下到 (左欄+3)×8 與 上列+0x15。
const (
	BadgeColStart = 1
	BadgeColStep  = 5
	BadgeRowStart = 0x3C
	BadgeRowStep  = 0x23
	// BadgePerRow 是一列放幾個：編號 % 8 決定欄、編號 / 8 決定列。
	BadgePerRow = 8
	// BadgeWidth／BadgeHeight 是圖本身的大小。
	BadgeWidth, BadgeHeight = 24, 21
	// 框相對圖的四個邊。
	BadgeFrameLeft, BadgeFrameTop = 3, 2
	BadgeFrameRight               = 3 * 8
	BadgeFrameBottom              = 0x15
)

// QuestBadge 是一格：哪一個任務、畫在哪裡。座標已經換算成像素。
type QuestBadge struct {
	// Site 是任務編號，也是圖號。
	Site int
	// X／Y 是圖的左上角；FrameX／FrameY／FrameW／FrameH 是外框。
	X, Y                       int
	FrameX, FrameY             int
	FrameW, FrameH             int
}

// QuestList 是一頁任務一覽。
type QuestList struct {
	// Group 是第幾組（0…4）；Archive 是這一組的徽章圖檔。
	Group   int
	Archive string
	Badges  []QuestBadge
}

// BadgeAt 依任務編號算出那一格的座標（`sub_167D9:0x1687C`／`sub_168F8`）。
func BadgeAt(site int) QuestBadge {
	col := BadgeColStart + BadgeColStep*(site%BadgePerRow)
	row := BadgeRowStart + BadgeRowStep*(site/BadgePerRow)
	b := QuestBadge{Site: site, X: col * 8, Y: row}
	b.FrameX = col*8 - BadgeFrameLeft
	b.FrameY = row - BadgeFrameTop
	b.FrameW = (col+3)*8 - b.FrameX + 1
	b.FrameH = row + BadgeFrameBottom - b.FrameY + 1
	return b
}

// QuestBadges 是一名角色在某一組裡「回報過」的那些任務（`0x1686E` 只收欄位 3）。
func QuestBadges(flags []int, group int) []QuestBadge {
	var out []QuestBadge
	for b := 0; b < QuestFlagBytes; b++ {
		i := group*QuestFlagBytes + b
		if i >= len(flags) {
			break
		}
		value := flags[i]
		for f := 0; f < 4; f++ {
			if value&3 == QuestReported {
				out = append(out, BadgeAt(4*b+f))
			}
			value >>= 2
		}
	}
	return out
}

// QuestGroups 是有東西可看的那幾組。原版逐組先看 `[0x1DD + 組]`，
// 是零就整組跳過（`0x1680A`）——那一格是已回報的任務數。
func QuestGroups(progress []int) []int {
	var out []int
	for i, n := range progress {
		if n != 0 {
			out = append(out, i)
		}
	}
	return out
}

// questListState 是翻到第幾組。
type questListState struct {
	groups []int
	at     int
}

// OpenQuestList 是資料頁上按下第 19 個圖示。一件都沒回報過就什麼都不做——
// 原版逐組檢查，五組都是零就直接走完迴圈回資料頁。
func (s *Session) OpenQuestList() {
	if s.Inside == nil || !s.ShowingSheet() {
		return
	}
	i := s.speakerIndex()
	if i < 0 {
		return
	}
	groups := QuestGroups(s.Party[i].Progress)
	if len(groups) == 0 {
		return
	}
	s.questList = &questListState{groups: groups}
	s.showQuestList()
}

// NextQuestPage 換下一組；翻完就回資料頁。
func (s *Session) NextQuestPage() {
	if s.questList == nil {
		return
	}
	s.questList.at++
	if s.questList.at >= len(s.questList.groups) {
		s.closeQuestList()
		return
	}
	s.showQuestList()
}

// ShowingQuestList 回報現在是不是停在任務一覽。
func (s *Session) ShowingQuestList() bool {
	return s.Inside != nil && s.Inside.QuestList != nil
}

func (s *Session) showQuestList() {
	i := s.speakerIndex()
	if i < 0 || s.questList == nil {
		return
	}
	group := s.questList.groups[s.questList.at]
	s.Inside.QuestList = &QuestList{Group: group,
		Archive: badgeArchive(group),
		Badges:  QuestBadges(s.Party[i].Quests, group)}
	// 原版一組一頁，圖示列縮成 SPACEBAR 等一個鍵（`0x168D4` 的 `sub_144C3`）。
	s.setInsideIcons([]int{IconSpacebar})
}

func (s *Session) closeQuestList() {
	s.questList = nil
	if s.Inside == nil {
		return
	}
	s.Inside.QuestList = nil
	i := s.speakerIndex()
	if i >= 0 {
		s.setInsideIcons(sheetIcons(s.Party[i]))
	}
}

// badgeArchive 把組別換成圖檔名：第 0 組是 AWD1PICS。
func badgeArchive(group int) string {
	return "AWD" + string(rune('1'+group)) + "PICS"
}
