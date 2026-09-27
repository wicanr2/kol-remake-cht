package battle

import "testing"

// 收尾畫哪一張結果圖由去向與「是不是競技場／訓練戰」決定（`sub_288D4` 的五個
// 呼叫點，RE-147）；戰場邊上的逃走不畫。
func TestResultPicture(t *testing.T) {
	for _, c := range []struct {
		name    string
		verdict Verdict
		rank    bool
		fromBar bool
		want    int
	}{
		{"一般戰鬥打贏", VerdictWon, false, false, ResultWon},
		{"一般戰鬥打輸", VerdictLost, false, false, ResultLost},
		{"競技場打贏", VerdictWon, true, false, ResultRankWon},
		{"競技場打輸", VerdictLost, true, false, ResultRankLost},
		{"逃走格逃走", VerdictFled, false, true, ResultFled},
		{"戰場邊上逃走", VerdictFled, false, false, ResultNone},
	} {
		s := fleeState()
		s.Stage, s.Verdict, s.RankFight, s.fledFromBar = StageDone, c.verdict, c.rank, c.fromBar
		if got := s.ResultPicture(); got != c.want {
			t.Errorf("%s：得到圖 %d，應為 %d", c.name, got, c.want)
		}
	}
}

// 還沒分勝負不畫；進了戰利品畫面之後也不再畫。
func TestResultPictureOnlyWhileClosing(t *testing.T) {
	s := fleeState()
	s.Verdict = VerdictWon
	if s.ResultPicture() != ResultNone {
		t.Fatal("還在戰場上不該畫結果圖")
	}
	s.Stage = StageDone
	if s.ResultPicture() != ResultWon {
		t.Fatal("打贏收尾應該畫圖 3")
	}
	s.resultOver = true
	if s.ResultPicture() != ResultNone {
		t.Fatal("戰利品畫面之後不該再畫")
	}
}
