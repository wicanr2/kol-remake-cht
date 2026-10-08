package main

import (
	"fmt"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/world"
)

// 走正式共同出口，驗隨機來源、過期標記及再次按鍵；位置預期來自RE-184正常原版返回。
func TestBattleExitWorldPosition(t *testing.T) {
	for _, verdict := range []battle.Verdict{battle.VerdictWon, battle.VerdictLost, battle.VerdictFled} {
		for _, origin := range []string{"ambush", "other", "stale"} {
			t.Run(fmt.Sprintf("%s/%d", origin, verdict), func(t *testing.T) {
				state := &battle.State{Stage: battle.StageDone, Verdict: verdict}
				quest := &questBattle{state: state, awarded: true, spoilsOpened: true}
				if origin == "ambush" {
					quest.worldAmbush = state
				} else if origin == "stale" {
					quest.worldAmbush = &battle.State{}
				}
				travel := &world.Travel{X: 49, Y: 23, SubX: 2, SubY: 3, Step: 1}
				screen := &world.Screen{Travel: travel}
				g := &game{quest: quest, worldSession: &world.Session{Screen: screen}, arenaFighter: -1}
				g.battleStep()
				if quest.state != nil {
					t.Fatal("正式共同出口未返回世界")
				}
				wantX, wantY := 2, 3
				if origin == "ambush" {
					wantX, wantY = 4, 4
					if quest.worldAmbush != nil {
						t.Fatal("隨機戰鬥一次性標記未清除")
					}
				}
				if travel.X != 49 || travel.Y != 23 || travel.Step != 1 || travel.SubX != wantX || travel.SubY != wantY {
					t.Fatalf("共同出口位置錯誤：%+v", travel)
				}
				travel.SubX++
				g.battleStep()
				if travel.SubX != wantX+1 {
					t.Fatal("重按鍵重設了返回後位置")
				}
			})
		}
	}
}
