package town

import "testing"

// 二十四個任務排成八欄三列：編號 % 8 決定欄、編號 / 8 決定列（RE-111）。
func TestBadgeLayout(t *testing.T) {
	first := BadgeAt(0)
	if first.X != BadgeColStart*8 || first.Y != BadgeRowStart {
		t.Fatalf("第 0 格畫在 (%d,%d)", first.X, first.Y)
	}
	// 同一列往右一格是五個字格。
	if BadgeAt(1).X-first.X != BadgeColStep*8 {
		t.Fatalf("欄距是 %d", BadgeAt(1).X-first.X)
	}
	// 第 8 個換到下一列，欄位回到最左。
	row2 := BadgeAt(8)
	if row2.X != first.X || row2.Y-first.Y != BadgeRowStep {
		t.Fatalf("第 8 格在 (%d,%d)", row2.X, row2.Y)
	}
	// 框比圖大一圈。
	if first.FrameX != first.X-BadgeFrameLeft || first.FrameY != first.Y-BadgeFrameTop {
		t.Fatalf("框的左上角是 (%d,%d)", first.FrameX, first.FrameY)
	}
	if first.FrameW <= BadgeWidth || first.FrameH <= BadgeHeight {
		t.Fatalf("框 %d×%d 沒有比圖大", first.FrameW, first.FrameH)
	}
}

// 只列回報過的（欄位 3）。
func TestQuestBadgesOnlyReported(t *testing.T) {
	flags := make([]int, QuestFlagBytes*5)
	SetQuestFlag(flags, 0, QuestTaken)
	SetQuestFlag(flags, 3, QuestDone)
	SetQuestFlag(flags, 5, QuestReported)
	SetQuestFlag(flags, 17, QuestReported)
	got := QuestBadges(flags, 0)
	if len(got) != 2 {
		t.Fatalf("應該只有兩格，得到 %d", len(got))
	}
	if got[0].Site != 5 || got[1].Site != 17 {
		t.Fatalf("列出來的是 %d 與 %d", got[0].Site, got[1].Site)
	}
}

// 一件都沒回報的那一組整組跳過。
func TestQuestGroupsSkipsEmpty(t *testing.T) {
	got := QuestGroups([]int{0, 3, 0, 1, 0})
	if len(got) != 2 || got[0] != 1 || got[1] != 3 {
		t.Fatalf("有東西的組是 %v", got)
	}
}

// 第 0 組的徽章圖檔是 AWD1PICS。
func TestBadgeArchive(t *testing.T) {
	if got := badgeArchive(0); got != "AWD1PICS" {
		t.Fatalf("第 0 組是 %q", got)
	}
	if got := badgeArchive(4); got != "AWD5PICS" {
		t.Fatalf("第 4 組是 %q", got)
	}
}
