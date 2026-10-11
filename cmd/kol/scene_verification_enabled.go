//go:build sceneverify

package main

import (
	"flag"
	"fmt"

	"github.com/wicanr2/kol-remake-cht/internal/battle"
)

var verifyScenes = flag.Bool("verify-scenes", false, "啟用 SPEC-040 場景驗收，僅供腳本擷取")
var verifyLockHP = flag.Bool("verify-lock-hp", false, "驗收隊伍免受部位與累積傷害")
var verifyAttack = flag.Int("verify-attack", 1, "驗收隊伍命中後的近戰／射擊傷害倍率，1…100")
var verifyWorldStart = flag.String("verify-world-start", "", "驗收世界落點 X,Y，透過現行到達回呼進入場景")

func validateSceneVerification() error {
	if !*verifyScenes {
		if *verifyLockHP || *verifyAttack != 1 || *verifyWorldStart != "" {
			return fmt.Errorf("驗收輔助參數必須同時指定 -verify-scenes")
		}
		return nil
	}
	value := func(name string) string { return flag.Lookup(name).Value.String() }
	if (value("scene") != "play" && value("scene") != "flow") ||
		(value("play-script") == "" && value("flow-script") == "") ||
		value("trace") == "" || (value("shot") == "" && value("shot-dir") == "") ||
		value("save") != "" || *verifyAttack < 1 || *verifyAttack > 100 {
		return fmt.Errorf("場景驗收需要 play／flow 腳本、trace 與截圖，禁止寫入存檔；傷害倍率須為 1…100")
	}
	battle.SceneVerification.LockHP = *verifyLockHP
	battle.SceneVerification.DamageMultiplier = *verifyAttack
	return nil
}

func sceneVerificationSnapshot() any {
	if !*verifyScenes {
		return nil
	}
	return map[string]any{"scope": "controlled-scene-ui", "lock_hp": *verifyLockHP,
		"party_damage_multiplier": *verifyAttack, "world_start": *verifyWorldStart,
		"original_parity_claimed": false}
}

func prepareSceneVerification(g *game) error {
	if !*verifyScenes || *verifyWorldStart == "" {
		return nil
	}
	x, y, err := parsePair(*verifyWorldStart)
	if err != nil || x < 0 || x > 63 || y < 0 || y > 63 ||
		g.scene != scenePlay || g.play == nil || g.worldSession == nil ||
		g.worldSession.OnArrive == nil {
		return fmt.Errorf("驗收世界落點需要 play 場景與 0…63 的 X,Y")
	}
	g.play.LeftTown = true
	g.worldState.Travel.X, g.worldState.Travel.Y = x, y
	g.worldState.Travel.SubX, g.worldState.Travel.SubY = 4, 4
	g.worldState.SyncTerrain()
	g.worldSession.Sync()
	if prompt := g.worldSession.OnArrive(x, y); prompt != nil {
		g.worldSession.SetPrompt(prompt)
	}
	return nil
}
