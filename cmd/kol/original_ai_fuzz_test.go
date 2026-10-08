package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/rng"
)

// 敵人 AI 的擾動對拍（#11，RE-168）：原版那一側從一場正常戰鬥第一次 AI 進場的整機快照出發，
// 每一組隨機改寫隊員與敵人的參戰者記錄、怪物記錄、耐力／疲勞與亂數種子，呼叫原版自己的
// `sub_2374C`，記下每次 `rand(n)` 的 n、結束後的亂數狀態與三個結果全域。這一側拿同一份
// 原始位元組建狀態、跑 `PlanEnemy`，比擲骰序列、亂數狀態與寫回的指令、攻擊選項、防禦選項、
// 方向、撤退旗標與撿起旗標。
//
// 沒有 `KOL_AIFUZZ_INPUT` 就跳過。收據由 `tools/oracle/book_startup_test.go` 的
// `KOL_BOOK_AIFUZZ=<組數>` 產生，見 docs/verification/enemy-ai-fuzz-20260930.md。
func TestOriginalEnemyAIFuzz(t *testing.T) {
	path := os.Getenv("KOL_AIFUZZ_INPUT")
	if path == "" {
		t.Skip("需要 dosgolem 正常路徑的 AI 擾動收據")
	}
	var input struct {
		Schema      string
		Actor       int
		MonsterBase string `json:"monster_base"`
		ScratchHex  string `json:"scratch_hex"`
		TerrainHex  string `json:"terrain_hex"`
		Cases       []struct {
			Actors   []string
			Counters []struct{ Slot, Stamina, Fatigue int }
			Seed     uint32
			Monster  string
			Stale    int `json:"stale_style"`
			StaleG   int `json:"stale_guard"`
			Error    string
			Rolls    []int
			After    uint32 `json:"seed_after"`
			Command  int
			Style    int
			Guard    int
			Record   string `json:"record_after"`
			Dbg      map[string]int
		}
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(body, &input); err != nil {
		t.Fatal(err)
	}
	if input.Schema != "kol-enemy-ai-fuzz-v1" || len(input.Cases) == 0 {
		t.Fatalf("收據 schema %q、%d 組", input.Schema, len(input.Cases))
	}
	scratch, err := hex.DecodeString(input.ScratchHex)
	if err != nil || len(scratch) < 0x2A {
		t.Fatal("收據沒有共用緩衝初值")
	}
	tail := [2]int{int(scratch[0x28]), int(scratch[0x29])}

	base := originalPlanningState(t)
	// 兩邊要是同一張戰場：收據存的是 Y*64+X，remake 的 `Field.At(x, y)` 是 X*64+Y 的轉置。
	terrain, err := hex.DecodeString(input.TerrainHex)
	if err != nil || len(terrain) != 4096 {
		t.Fatal("收據的戰場不是 4096 byte")
	}
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			if int(terrain[y*64+x]) != base.Field.At(x, y) {
				t.Fatalf("戰場 (%d,%d)：原版 %d，remake 快照 %d——不是同一場", x, y, terrain[y*64+x], base.Field.At(x, y))
			}
		}
	}

	monster, err := hex.DecodeString(input.MonsterBase)
	if err != nil || len(monster) != 0x80 {
		t.Fatal("收據沒有怪物記錄")
	}
	_ = monster
	ran, failedOriginal := 0, 0
	commands := map[string]int{}
	var mismatches []string
	var dump []map[string]any
	for n, c := range input.Cases {
		if c.Error != "" {
			failedOriginal++
			continue
		}
		s := *base
		s.Actors = append([]battle.Actor(nil), base.Actors...)
		s.StyleMaskTail = tail
		s.SetEnemyAIGlobals(c.Stale, c.StaleG)
		raw := func(slot int) []byte {
			b, err := hex.DecodeString(c.Actors[slot])
			if err != nil || len(b) != 50 {
				t.Fatalf("第 %d 組槽 %d 的記錄不是 50 byte", n, slot)
			}
			return b
		}
		for _, slot := range []int{0, 1, input.Actor} {
			applyFuzzRecord(&s.Actors[slot], raw(slot), slot == input.Actor)
		}
		mon, _ := hex.DecodeString(c.Monster)
		a := &s.Actors[input.Actor]
		a.FleeAt, a.FleeRecover = int(mon[0x0D]), int(mon[0x0E])
		a.Unarmed = mon[0x20] == 0
		if a.Unarmed {
			// 探針把手上記錄的第一個 byte 清成 0 時，原版 sub_24FD3
			// 不計那件武器的重量；兩側的裝備也必須同步清空。
			a.Weapon = battle.Weapon{}
		}
		a.ReloadGear()
		for _, k := range c.Counters {
			if k.Slot < len(s.Actors) && s.Actors[k.Slot].Side != battle.SideEmpty {
				s.Actors[k.Slot].Stamina, s.Actors[k.Slot].Fatigue = k.Stamina, k.Fatigue
			}
		}
		var order []int
		for i := range s.Actors {
			if i != input.Actor && s.Actors[i].Side != battle.SideEmpty && raw(i)[0x22] != 0 && raw(i)[1] != 2 {
				order = append(order, i)
			}
		}
		s.Round.Turn = len(order)
		s.Round.Order = append(order, input.Actor)

		rec := &recordingRoll{lcg: rng.NewState(c.Seed)}
		d := s.PlanEnemy(input.Actor, rec)
		var bad []string
		if fmt.Sprint(rec.ns) != fmt.Sprint(c.Rolls) {
			bad = append(bad, fmt.Sprintf("擲骰 原版 %v remake %v", c.Rolls, rec.ns))
		}
		if got := rec.lcg.State(); got != c.After {
			bad = append(bad, fmt.Sprintf("之後的亂數狀態 原版 %#x remake %#x", c.After, got))
		}
		for _, f := range []struct {
			name         string
			want, actual int
		}{
			{"指令碼", c.Command, a.Command}, {"攻擊選項", c.Style, a.Style}, {"防禦選項", c.Guard, a.Guard},
		} {
			if f.want != f.actual {
				bad = append(bad, fmt.Sprintf("%s 原版 %#x remake %#x", f.name, f.want, f.actual))
			}
		}
		rout, _ := hex.DecodeString(c.Record)
		if len(rout) == 50 {
			if int(rout[0x28]) != a.Facing {
				bad = append(bad, fmt.Sprintf("方向 原版 %d remake %d", rout[0x28], a.Facing))
			}
			if (rout[0x29] != 0) != a.Retreating {
				bad = append(bad, fmt.Sprintf("撤退旗標 原版 %d remake %t", rout[0x29], a.Retreating))
			}
			if int(rout[0x2B]) != a.PickedUp {
				bad = append(bad, fmt.Sprintf("撿起旗標 原版 %d remake %d", rout[0x2B], a.PickedUp))
			}
		}
		ran++
		commands[fmt.Sprintf("%#x", c.Command)]++
		if len(bad) > 0 {
			mismatches = append(mismatches, fmt.Sprintf("第 %d 組（決策 %+v）：%v", n, d, bad))
			dump = append(dump, map[string]any{"case": n, "bad": bad, "decision": fmt.Sprintf("%+v", d),
				"actors": c.Actors[:2], "enemy": c.Actors[input.Actor], "monster": c.Monster, "counters": c.Counters[input.Actor],
				"rolls_original": c.Rolls, "rolls_remake": rec.ns, "seed": c.Seed,
				"dbg_original": c.Dbg, "dbg_remake": remakeMoveDebug(&s, input.Actor, d.Ring)})
		}
	}
	if out := os.Getenv("KOL_AIFUZZ_DUMP"); out != "" {
		b, _ := json.MarshalIndent(dump, "", " ")
		if err := os.WriteFile(out, b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	keys := make([]string, 0, len(commands))
	for k := range commands {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		t.Logf("指令碼 %s：%d 組", k, commands[k])
	}
	t.Logf("擾動 %d 組：比較 %d 組，原版呼叫失敗 %d 組，不一致 %d 組", len(input.Cases), ran, failedOriginal, len(mismatches))
	if failedOriginal != 0 {
		t.Errorf("有 %d 組原版呼叫失敗；不能略過後宣稱整批對拍通過", failedOriginal)
	}
	for i, m := range mismatches {
		if i >= 15 {
			t.Logf("……另有 %d 組不一致", len(mismatches)-i)
			break
		}
		t.Error(m)
	}
}

// applyFuzzRecord 把一筆 50 byte 的參戰者記錄套到 remake 的參戰者上。
//
// 部位耐久的兩邊表示法不同：原版活著是 1…0x7F、被打廢存 0xFF（`sub_277F4:0x278DE`），
// remake 活著是正數、被打廢是 0（`Actor.Take`）。所以 0x80 以上一律換成 0，
// 頭或胸是 0xFF 就是倒下。原版那一格是 0 的記錄打不出來，收據不擾動到那裡。
// 敵人另外帶種類（+0x00）、等級（+0x02）與智力（+0x0C）。
func applyFuzzRecord(a *battle.Actor, raw []byte, enemy bool) {
	hp := func(off int) int {
		if raw[off] >= 0x80 {
			return 0
		}
		return int(raw[off])
	}
	a.X, a.Y = int(raw[0x1F]), int(raw[0x20])
	a.Damage, a.Parries = int(raw[0x0E]), int(raw[0x0F])
	a.Head, a.Chest = hp(0x11), hp(0x13)
	a.RightArm, a.LeftArm = hp(0x15), hp(0x17)
	a.RightLeg, a.LeftLeg = hp(0x19), hp(0x1B)
	a.Flag1D, a.Flag1E = int(raw[0x1D]), int(raw[0x1E])
	a.Style, a.Guard, a.Command = int(raw[0x23]), int(raw[0x24]), int(raw[0x25])
	a.Facing = int(raw[0x28])
	a.Retreating = raw[0x29] != 0
	a.LastHurt = int(raw[0x2A])
	a.PickedUp = int(raw[0x2B])
	a.Hurt = int(raw[0x2D])
	if enemy {
		a.Kind, a.Level, a.Intellect = int(raw[0x00]), int(raw[0x02]), int(raw[0x0C])
		// RE-225／RE-237：負重等級讀同一個參戰者 +0x02。
		a.LoadLevel = a.Level
	}
}

// remakeMoveDebug 列出這一側算移動三式權重時的中間值，好和收據裡原版的 `dbg` 對照。
func remakeMoveDebug(s *battle.State, i, ring int) map[string]int {
	a := s.Actors[i]
	num, den := battle.ClampEffort(a.TimeAllowance(), battle.ActionCost(a, 0, battle.MoveFormEffortTime))
	w := s.AIWeights().Move
	return map[string]int{"base0": w[0], "base1": w[1], "base2": w[2], "num": num, "den": den, "ring": ring,
		"fat": a.FatigueRatio(), "allow": a.TimeAllowance(), "cost": battle.ActionCost(a, 0, battle.MoveFormEffortTime)}
}
