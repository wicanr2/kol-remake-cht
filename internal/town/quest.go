package town

// 任務物品的交付與回報（RE-107／RE-109）。
//
// 玩家把任務要的東西帶進對的建築，走進去就自動交付：那件東西被換成建築給的另一件
// （有的話），任務旗標從「完成」變成「回報過」，角色的「已回報數」加一。回報滿
// 二十三個才接得到最終任務，打完它就播結局（RE-106）。
//
// 這一份只做規則；接進建築內那一層要先有角色身上那十六格的 typed data。

// 任務旗標的四個狀態。一個 byte 塞四個 2-bit 欄位，除數見 QuestDivisors。
const (
	// QuestNone 是沒接；QuestTaken 是接了（`sub_2B0FF:0x2B108` 要這個值才讓
	// 玩家踏上任務地點）；QuestDone 是打完了（`sub_28F5D`，RE-101）；
	// QuestReported 是交付過了（`sub_11A93:0x11CC1`，RE-107）。
	QuestNone = iota
	QuestTaken
	QuestDone
	QuestReported
)

// QuestDivisors 是任務編號低兩位對應的除數（DGROUP `0x007C`）。
var QuestDivisors = [4]int{1, 4, 16, 64}

// QuestFlagBytes 是一組旗標有幾個 byte（24 個地點 ÷ 4）。
const QuestFlagBytes = 6

// FinalQuestReports 是最終任務的門檻：已回報數要到這個值才接得到
// （`sub_14564:0x145E7`），而打完它時那一格正好是這個值，結局就是這樣觸發的。
const FinalQuestReports = 0x17

// 身上要掃哪幾格（`sub_11A93:0x11B85`…`0x11CF7`）。一格十六 byte，從 0x30 起、
// **不到 0x1A0 為止**（`0x11CF2` 的 `jge` 是先加再比，所以最後一格是 0x190），
// 中間 0x40…0x9F 那一段跳過——那裡是四列武器技能，不是物品。
const (
	CarriedFirst  = 0x30
	CarriedLast   = 0x190
	CarriedStride = 0x10
	CarriedSkipLo = 0x40
	CarriedSkipHi = 0x9F
)

// CarriedOffsets 是實際會被掃到的那幾格在角色記錄裡的位移，由小到大。
func CarriedOffsets() []int {
	var out []int
	for off := CarriedFirst; off <= CarriedLast; off += CarriedStride {
		if off >= CarriedSkipLo && off <= CarriedSkipHi {
			continue
		}
		out = append(out, off)
	}
	return out
}

// QuestBuilding 是一棟建築的任務欄位（抽取器的 quest_* 那幾格）。
// Wants 為空表示這棟不收任務物品。
type QuestBuilding struct {
	// Site 是任務地點編號 0…23；Final 為真表示這是最終任務。
	Site  int
	Final bool
	// Wants 是要收的那件物品十六 byte；Gives 是換給你的那件，空的話不給。
	Wants []byte
	Gives []byte
}

// QuestFlag 取一名角色第 site 個任務的旗標。
func QuestFlag(flags []int, site int) int {
	v, div, ok := questField(flags, site)
	if !ok {
		return QuestNone
	}
	return (v / div) & 3
}

// SetQuestFlag 把第 site 個任務的旗標換成 state，回傳有沒有改到。
func SetQuestFlag(flags []int, site, state int) bool {
	v, div, ok := questField(flags, site)
	if !ok || state < 0 || state > 3 {
		return false
	}
	flags[site/4] = (v &^ (3 * div)) | (state * div)
	return true
}

func questField(flags []int, site int) (value, divisor int, ok bool) {
	if site < 0 || site >= QuestFlagBytes*4 || site/4 >= len(flags) {
		return 0, 0, false
	}
	return flags[site/4], QuestDivisors[site&3], true
}

// TurnIn 是走進一棟收任務物品的建築時發生的事（`sub_11A93:0x11B5A` 起）。
//
// carried 是一名角色身上那幾格的內容，索引與 CarriedOffsets 對齊；flags 是他的
// 任務旗標；progress 是「已回報數」那幾格（角色記錄 0x1DD 起），group 是
// `word_2DF70`。
//
// 找到那件東西就換成建築給的那件（Gives 為空就清空那一格），旗標改成
// QuestReported，已回報數加一，回傳換掉的是第幾格。**旗標不是「完成」也照樣換**
// ——原版比的只有物品，旗標那一段是分開跑的（`0x11C9C` 另外判 == 2 才記）。
func TurnIn(carried [][]byte, flags, progress []int, group int,
	b QuestBuilding) (slot int, reported bool) {
	slot = -1
	if len(b.Wants) != CarriedStride {
		return -1, false
	}
	for i, item := range carried {
		if len(item) != CarriedStride || !sameItem(item, b.Wants) {
			continue
		}
		give := b.Gives
		if len(give) != CarriedStride {
			give = make([]byte, CarriedStride)
		}
		copy(item, give)
		slot = i
		break
	}
	if slot < 0 {
		return -1, false
	}
	// 回報那一段只收「已完成」的（`0x11C9C`）。
	if QuestFlag(flags, b.Site) == QuestDone {
		SetQuestFlag(flags, b.Site, QuestReported)
		if group >= 0 && group < len(progress) {
			progress[group]++
			reported = true
		}
	}
	return slot, reported
}

// CanTakeFinal 回報「已回報數」夠不夠接最終任務（`sub_14564:0x145E7`）。
func CanTakeFinal(progress []int, group int) bool {
	if group < 0 || group >= len(progress) {
		return false
	}
	return progress[group] >= FinalQuestReports
}

func sameItem(a, b []byte) bool {
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
