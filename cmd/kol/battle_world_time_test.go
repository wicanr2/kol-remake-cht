package main

import (
	"fmt"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/world"
)

type worldExitRand struct{ calls []int }

func (r *worldExitRand) Roll(n int) int { r.calls = append(r.calls, n); return 1 }

func TestBattleExitResumesOnlyItsWorldTime(t *testing.T) {
	for _, verdict := range []battle.Verdict{battle.VerdictWon, battle.VerdictLost, battle.VerdictFled} {
		for _, origin := range []string{"ambush", "other", "stale"} {
			t.Run(fmt.Sprintf("%s/%d", origin, verdict), func(t *testing.T) {
				battleRand, oldWorldRand := &worldExitRand{}, &worldExitRand{}
				state := &battle.State{Stage: battle.StageDone, Verdict: verdict, Rand: battleRand}
				quest := &questBattle{state: state, awarded: true, spoilsOpened: true}
				if origin == "ambush" {
					quest.worldAmbush = state
				} else if origin == "stale" {
					quest.worldAmbush = &battle.State{}
				}
				clock := &world.Clock{Ticks: world.ClockStart, Part: world.DayParts - 1, Month: 3}
				session := &world.Session{Screen: &world.Screen{Clock: clock, Travel: &world.Travel{}}, Rand: oldWorldRand, Camping: true}
				session.OnTimePartBusy = func() bool { return quest.state != nil }
				session.EndCamping()
				if clock.Day != 0 || len(oldWorldRand.calls) != 0 {
					t.Fatal("戰鬥仍在等待卻已推日期")
				}
				g := &game{quest: quest, worldSession: session, arenaFighter: -1}
				g.battleStep()
				if origin == "ambush" {
					if clock.Day != 1 || session.Rand != battleRand || len(battleRand.calls) != 1 || battleRand.calls[0] != 128 {
						t.Fatal("實際隨機出口未沿戰鬥亂數接續天候")
					}
				} else if clock.Day != 0 || session.Rand != oldWorldRand || len(battleRand.calls) != 0 {
					t.Fatal("其他來源或過期標記接續了世界時間")
				}
				g.battleStep()
				if len(oldWorldRand.calls) != 0 || (origin == "ambush" && len(battleRand.calls) != 1) {
					t.Fatal("重按鍵重複擲天候")
				}
			})
		}
	}
}

func TestBattleExitReadsActualScratchOnlyForItsPendingWalk(t *testing.T) {
	for _, verdict := range []battle.Verdict{battle.VerdictWon, battle.VerdictLost, battle.VerdictFled} {
		for _, origin := range []string{"ambush", "other", "stale"} {
			t.Run(fmt.Sprintf("%s/%d", origin, verdict), func(t *testing.T) {
				// 所有結算故意留同一個已證實索引，驗證不依勝負重建成本。
				random := &worldExitRand{}
				state := &battle.State{Stage: battle.StageDone, Verdict: verdict, Rand: random, ReturnScratch: 16}
				quest := &questBattle{state: state, awarded: true, spoilsOpened: true}
				if origin == "ambush" {
					quest.worldAmbush = state
				} else if origin == "stale" {
					quest.worldAmbush = &battle.State{}
				}
				clock := &world.Clock{Ticks: 3360, Month: 3}
				travel := &world.Travel{X: 49, Y: 23, SubX: 3, SubY: 3, Step: 1}
				session := &world.Session{Screen: &world.Screen{Clock: clock, Travel: travel, Movement: 20}, Rand: &worldExitRand{}}
				session.OnTimePartBusy = func() bool { return quest.state != nil }
				if !session.Walk(0, -1) || clock.Ticks != 3360 {
					t.Fatal("步行未等待實際戰鬥")
				}
				g := &game{quest: quest, worldSession: session, arenaFighter: -1}
				g.battleStep()
				want := 3360
				if origin == "ambush" {
					want = 3260
				}
				if clock.Ticks != want {
					t.Fatalf("共同出口成本錯誤：ticks=%d want=%d", clock.Ticks, want)
				}
				g.battleStep()
				if clock.Ticks != want {
					t.Fatal("共同出口再次扣時刻")
				}
			})
		}
	}
}
