package main

import (
	"fmt"
	"image"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/world"
)

func TestActualQuestExitResumesOnlyItsArrival(t *testing.T) {
	for _, verdict := range []battle.Verdict{battle.VerdictWon, battle.VerdictLost, battle.VerdictFled} {
		for _, origin := range []string{"quest", "other", "stale"} {
			t.Run(fmt.Sprintf("%s/%d", origin, verdict), func(t *testing.T) {
				state := &battle.State{Stage: battle.StageDone, Verdict: verdict, ReturnScratch: 16, Rand: &worldExitRand{}}
				quest := &questBattle{awarded: true, spoilsOpened: true}
				clock := &world.Clock{Ticks: 840, Part: 3, Month: 3}
				grid, err := world.LoadGrid(image.NewGray(image.Rect(0, 0, world.GridSide, world.GridSide)))
				if err != nil {
					t.Fatal(err)
				}
				data := &world.Data{Terrain: make([]int, 256), Road: make([]int, 256)}
				travel := &world.Travel{Data: data, Grid: grid, X: 49, Y: 23, SubX: 7, SubY: 3, Step: 1}
				random := &worldExitRand{}
				session := &world.Session{Screen: &world.Screen{Clock: clock, Travel: travel, Movement: 20}, Rand: random}
				session.OnArrive = func(x, y int) *world.Prompt { return &world.Prompt{Kind: world.PromptQuest} }
				session.OnTimePartBusy = func() bool { return quest.state != nil }
				session.OnPromptAnswer = func(bool) { quest.state = state }
				if !session.Walk(1, 0) {
					t.Fatal("到達未移動")
				}
				session.Accept()
				if origin == "quest" {
					quest.worldQuest = state
				} else if origin == "stale" {
					quest.worldQuest = &battle.State{}
				}
				g := &game{quest: quest, worldSession: session, arenaFighter: -1}
				g.battleStep()
				if quest.state != nil || travel.X != 50 || travel.Y != 23 || session.Rand != random {
					t.Fatal("共同出口未保存世界位置及既有亂數來源")
				}
				want := 840
				if origin == "quest" {
					want = 740
					if travel.SubX != 4 || travel.SubY != 4 || quest.worldQuest != nil || len(random.calls) != 1 {
						t.Fatal("實際任務未設次格4、4及完成一次天候")
					}
				} else if travel.SubX != 0 || travel.SubY != 3 || len(random.calls) != 0 {
					t.Fatal("其他或過期來源接續了這一場到達")
				}
				g.battleStep()
				if clock.Ticks != want {
					t.Fatalf("返回時刻錯誤或被重複扣：got=%d want=%d", clock.Ticks, want)
				}
			})
		}
	}
}
