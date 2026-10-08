package world

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/rng"
)

func TestReturnCostOriginalSlices(t *testing.T) {
	path := os.Getenv("KOL_RETURN_COST_CASES")
	if path == "" {
		t.Skip("需要封存的原版讀取切片")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var input struct {
		Rows []struct{ Index, Cost, Seed, FinalSeed, Calls int }
	}
	if err := json.Unmarshal(raw, &input); err != nil {
		t.Fatal(err)
	}
	if len(input.Rows) != 10 {
		t.Fatal("原版切片數量改變")
	}
	for _, row := range input.Rows {
		r := rng.New(row.Seed)
		cost, ok := NewClock().ReturnCost(uint16(row.Index), 20, r)
		if !ok || cost != row.Cost || r.State() != uint32(row.FinalSeed) || row.Calls != 0 {
			t.Fatalf("index=%d seed=%d：成本=%d、可讀=%v、種子=%d", row.Index, row.Seed, cost, ok, r.State())
		}
	}
}

func TestReturnCostReadsAfterWeatherAndClearsPendingIndex(t *testing.T) {
	s, _ := delayedTimeSession(t)
	busy := true
	s.OnTimePartBusy = func() bool { return busy }
	r := rng.New(1)
	s.Rand = r
	if !s.Walk(0, -1) || s.Screen.Clock.Ticks != 3360 || r.State() != 1 {
		t.Fatal("正常步行沒有等待遭遇")
	}
	// 待處理的值是返回840，不是北向FFFF；真正讀值必須晚於天候骰。
	busy = false
	s.ResumeAfterBattle(840)
	wantRand := rng.New(1)
	wantRand.Roll(128)
	want := 3360 - 2*(20+int(byte(wantRand.State())))
	if s.Screen.Clock.Ticks != want || r.State() != wantRand.State() || s.UnresolvedReturnIndex != nil {
		t.Fatalf("天候後讀取錯誤：ticks=%d seed=%d", s.Screen.Clock.Ticks, r.State())
	}
	s.ResumeAfterBattle(16)
	if s.Screen.Clock.Ticks != want || r.State() != wantRand.State() {
		t.Fatal("過期返回重複扣時刻或擲天候")
	}
	s.Walk(0, -1)
	if s.Screen.Clock.Ticks != want-40 {
		t.Fatal("下一步沿用舊戰鬥暫存")
	}
}

func TestUnresolvedReturnCostRemainsExplicit(t *testing.T) {
	for _, index := range []uint16{20, 150, 586, 720, 1400, 0xfffe} {
		if _, ok := NewClock().ReturnCost(index, 20, rng.New(1)); ok {
			t.Fatalf("未證實的動態byte被當成已解：%d", index)
		}
	}
	if _, ok := NewClock().ReturnCost(840, 20, stepRand{}); ok {
		t.Fatal("沒有State的亂數來源被猜補")
	}
	s, _ := delayedTimeSession(t)
	busy := true
	s.OnTimePartBusy = func() bool { return busy }
	s.Walk(0, -1)
	busy = false
	s.ResumeAfterBattle(720)
	if s.UnresolvedReturnIndex == nil || *s.UnresolvedReturnIndex != 720 || s.Screen.Clock.Ticks != 3320 {
		t.Fatal("未解分支沒有保留診斷及既有近似")
	}
	s.Walk(0, -1)
	if s.UnresolvedReturnIndex != nil {
		t.Fatal("舊診斷污染下一步")
	}
}
