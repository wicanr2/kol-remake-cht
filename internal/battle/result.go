package battle

// 戰鬥結果圖（RE-147；出口分派見 RE-146，COMBPICS 格式見 RE-006）。原版在共同出口
// `0x211AD` 之後依去向呼叫 `sub_288D4(圖號)`：把 COMBPICS 的一張 296×144 圖貼在
// 位元組欄 [1,38)、掃描列 [8,152)，再畫一圈白框（`sub_1B6DC(7, 7, 0x131, 0x98, 0x0F)`）
// 與投影（`sub_1B7E9(7, 7, 0x131, 0x98)`）。
//
//	一般戰鬥打贏   圖 3  `0x2125C`：印完「You have defeated…」之後、等鍵之前
//	一般戰鬥打輸   圖 2  `0x2157B`：印完「…defeated you!」之後、等鍵之前
//	逃走格逃走     圖 0  `0x215AC`：先畫圖，再印「You flee for your…」
//	競技場／訓練戰打贏 圖 4  `0x2161F`：打贏那一句之後、升級那一句之前
//	競技場／訓練戰打輸 圖 1  `0x216B5`：賭注取負之後、結算之前
//
// 五處都是在第一句勝負訊息印出、等第一次鍵之前畫，而且之後的訊息只印在卷軸上，
// 不重畫戰場，所以圖一直留到這一場收尾。戰場邊上的逃走在一般戰鬥落到 `0x21607`，
// 不畫圖（RE-146）。

// ResultArchive 是戰鬥結果圖的來源。
const ResultArchive = "COMBPICS"

// 結果圖的圖號（`sub_288D4` 的參數）。
const (
	ResultFled     = 0
	ResultRankLost = 1
	ResultLost     = 2
	ResultWon      = 3
	ResultRankWon  = 4
	// ResultNone 表示這一場收尾不畫圖。
	ResultNone = -1
)

// ResultPictureAt 是圖貼的位置：位元組欄 1、掃描列 8（`sub_288D4` 開頭設的
// `word_3499E = 1`、`word_349A0 = 8`，右下界是 `word_34A06 = 0x26`、`word_34A0E = 0x98`）。
var ResultPictureAt = struct{ X, Y int }{X: 8, Y: 8}

// ResultFrame 是圖外那一圈白框與投影的範圍：(7,7)–(305,152)，含端點。
var ResultFrame = struct{ X, Y, Width, Height int }{X: 7, Y: 7, Width: 299, Height: 146}

// ResultPicture 回報這一場收尾要畫哪一張結果圖；還沒分勝負、戰場邊上的逃走，
// 或已經進到戰利品畫面時回 ResultNone。
func (s *State) ResultPicture() int {
	if s.Stage != StageDone || s.resultOver {
		return ResultNone
	}
	switch s.Verdict {
	case VerdictWon:
		if s.RankFight {
			return ResultRankWon
		}
		return ResultWon
	case VerdictLost:
		if s.RankFight {
			return ResultRankLost
		}
		return ResultLost
	case VerdictFled:
		if s.FleeLine() {
			return ResultFled
		}
	}
	return ResultNone
}
