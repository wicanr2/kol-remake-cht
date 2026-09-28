package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/rng"
)

// 執行段的同狀態比較（#9，攻擊結算）：原版每一回合進執行段那一刻的參戰者記錄與
// 32 位元亂數狀態，交給 remake 從同一刻跑完執行段，比三樣東西：
//
//   - 擲骰：每一次 `rand(n)` 的 n，順序要一樣（原版由 `sub_1B5A8` 的 OnCall 記下）；
//   - 回合結束時的參戰者欄位：位置、累積傷害、六個部位、挨的傷、招架次數與疲勞；
//   - 回合結束時的亂數狀態。
//
// 原版那一側是 `tools/oracle/book_startup_test.go` 的 `KOL_BOOK_EXECUTE` 模式，收據
// `battle-rounds.json`。規劃段不在這裡比：每一回合都從原版的決策與站位接手。
func TestOriginalNormalRoundsExecution(t *testing.T) {
	path := os.Getenv("KOL_ROUNDS_INPUT")
	if path == "" {
		t.Skip("需要 dosgolem 正常路徑的多回合執行段收據")
	}
	var input struct {
		Schema     string
		Boundaries []struct {
			Round int
			At    string
			State struct {
				Seed     uint32
				Actors   []string
				Counters []struct{ Slot, Stamina, Fatigue int }
			}
		}
		Rolls []struct {
			N      int
			Caller string `json:"caller_ida"`
			Phase  string
			Round  int
		}
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(body, &input); err != nil {
		t.Fatal(err)
	}
	if input.Schema != "kol-battle-rounds-v1" {
		t.Fatalf("收據 schema 不符：%q", input.Schema)
	}
	slots := []int{0, 1, 6}
	compared := 0
	for bi, b := range input.Boundaries {
		if !strings.HasPrefix(b.At, "執行段入口") || bi+1 >= len(input.Boundaries) {
			continue
		}
		end := input.Boundaries[bi+1]
		t.Run(fmt.Sprintf("第%d回合", b.Round), func(t *testing.T) {
			s := originalPlanningState(t)
			for _, i := range slots {
				applyOriginalRecord(t, &s.Actors[i], b.State.Actors[i])
			}
			for _, c := range b.State.Counters {
				if c.Slot < len(s.Actors) && s.Actors[c.Slot].Side != battle.SideEmpty {
					s.Actors[c.Slot].Stamina, s.Actors[c.Slot].Fatigue = c.Stamina, c.Fatigue
				}
			}
			rec := &recordingRoll{lcg: rng.NewState(b.State.Seed)}
			s.Rand = rec
			s.BeginExecution(rec)
			for n := 0; n < 100; n++ {
				if _, ok := s.Act(rec); !ok {
					break
				}
			}

			var want []int
			var callers []string
			for _, r := range input.Rolls {
				if r.Round == b.Round && r.Phase == "exec" {
					want = append(want, r.N)
					callers = append(callers, r.Caller)
				}
			}
			if fmt.Sprint(rec.ns) != fmt.Sprint(want) {
				t.Errorf("擲骰序列不同：\n原版  %v\n      呼叫點 %v\nremake %v", want, callers, rec.ns)
			}
			if got := rec.lcg.State(); got != end.State.Seed {
				t.Errorf("回合結束的亂數狀態：原版 %#x，remake %#x", end.State.Seed, got)
			}
			for _, i := range slots {
				raw, _ := hex.DecodeString(end.State.Actors[i])
				a := s.Actors[i]
				for _, f := range []struct {
					name        string
					off, actual int
				}{
					{"X", 0x1F, a.X}, {"Y", 0x20, a.Y}, {"累積傷害", 0x0E, a.Damage},
					{"頭", 0x11, a.Head}, {"胸", 0x13, a.Chest}, {"右臂", 0x15, a.RightArm},
					{"左臂", 0x17, a.LeftArm}, {"右腿", 0x19, a.RightLeg}, {"左腿", 0x1B, a.LeftLeg},
					{"挨的傷 +0x2D", 0x2D, a.Hurt}, {"招架次數 +0x0F", 0x0F, a.Parries},
					{"方向 +0x28", 0x28, a.Facing},
				} {
					if f.actual != int(raw[f.off]) {
						t.Errorf("槽 %d %s：原版 %d，remake %d", i, f.name, raw[f.off], f.actual)
					}
				}
				for _, c := range end.State.Counters {
					if c.Slot == i && c.Fatigue != a.Fatigue {
						t.Errorf("槽 %d 疲勞：原版 %d，remake %d", i, c.Fatigue, a.Fatigue)
					}
				}
			}
		})
		compared++
	}
	if compared == 0 {
		t.Fatal("收據裡沒有執行段")
	}
}

// applyOriginalRecord 把原版參戰者記錄裡執行段會讀的欄位放進 remake 的參戰者。
// 位移見 `battle.Actor` 的欄位說明；角色與怪物本體的靜態欄位由 NewQuest 建好，不覆蓋。
func applyOriginalRecord(t *testing.T, a *battle.Actor, record string) {
	t.Helper()
	raw, err := hex.DecodeString(record)
	if err != nil || len(raw) != 50 {
		t.Fatalf("參戰者記錄不是 50 byte：%q", record)
	}
	if int(raw[1]) != int(a.Side) {
		t.Fatalf("陣營不符：原版 %d，remake %d", raw[1], a.Side)
	}
	a.X, a.Y = int(raw[0x1F]), int(raw[0x20])
	a.Damage = int(raw[0x0E])
	a.Parries = int(raw[0x0F])
	a.Head, a.Chest = int(raw[0x11]), int(raw[0x13])
	a.RightArm, a.LeftArm = int(raw[0x15]), int(raw[0x17])
	a.RightLeg, a.LeftLeg = int(raw[0x19]), int(raw[0x1B])
	a.Flag1D, a.Flag1E = int(raw[0x1D]), int(raw[0x1E])
	a.Style, a.Guard, a.Command = int(raw[0x23]), int(raw[0x24]), int(raw[0x25])
	a.Facing = int(raw[0x28])
	a.Retreating = raw[0x29] != 0
	a.LastHurt = int(raw[0x2A])
	a.PickedUp = int(raw[0x2B])
	a.Hurt = int(raw[0x2D])
}

// recordingRoll 是原版那一支亂數，另外記下每一次 Roll(n) 的 n。
type recordingRoll struct {
	lcg *rng.LCG
	ns  []int
}

func (r *recordingRoll) Roll(n int) int {
	r.ns = append(r.ns, n)
	return r.lcg.Roll(n)
}

// 敵人規劃的同狀態比較（#11）：原版每一回合 AI（`sub_2374C`）進場那一刻的參戰者記錄、
// 規劃過了沒有（`+0x22`）與亂數狀態交給 remake 的 `PlanEnemy`，比 AI 這一段的擲骰、
// 之後的亂數狀態，以及寫回的指令、攻擊選項、防禦選項與方向（取同一回合執行段入口的記錄）。
func TestOriginalNormalEnemyPlanning(t *testing.T) {
	path := os.Getenv("KOL_ROUNDS_INPUT")
	if path == "" {
		t.Skip("需要 dosgolem 正常路徑的多回合收據")
	}
	var input struct {
		Boundaries []struct {
			Round int
			At    string
			State struct{ Actors []string }
		}
		EnemyPlans []struct {
			Round, Actor int
			State        struct {
				Seed     uint32
				Actors   []string
				Counters []struct{ Slot, Stamina, Fatigue int }
			}
		} `json:"enemy_plans"`
		// ScratchStart 是探測開始時共用緩衝 `byte_37CF8` 的前 0x30 byte：攻擊式遮罩第 8、9 筆
		// 讀它的 +0x28／+0x29（RE-159）。
		ScratchStart string `json:"scratch_start_hex"`
		Rolls        []struct {
			N      int
			Caller string `json:"caller_ida"`
			Before uint32 `json:"seed_before"`
			Phase  string
			Round  int
		}
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(body, &input); err != nil {
		t.Fatal(err)
	}
	if len(input.EnemyPlans) == 0 {
		t.Fatal("收據裡沒有敵人規劃（要用加了 AI 入口鉤子的探針重跑）")
	}
	scratch, err := hex.DecodeString(input.ScratchStart)
	if err != nil || len(scratch) < 0x2A {
		t.Fatal("收據裡沒有共用緩衝的初值（要用加了寫入監看的探針重跑）")
	}
	tail := [2]int{int(scratch[0x28]), int(scratch[0x29])}
	// 規劃段裡不屬於 AI 的三個呼叫點：規劃鍵值、檢視擲骰。
	notAI := map[string]bool{"0x1f01c": true, "0x1f07f": true}
	for _, p := range input.EnemyPlans {
		t.Run(fmt.Sprintf("第%d回合", p.Round), func(t *testing.T) {
			s := originalPlanningState(t)
			s.StyleMaskTail = &tail
			for _, i := range []int{0, 1, 6} {
				applyOriginalRecord(t, &s.Actors[i], p.State.Actors[i])
			}
			for _, c := range p.State.Counters {
				if c.Slot < len(s.Actors) && s.Actors[c.Slot].Side != battle.SideEmpty {
					s.Actors[c.Slot].Stamina, s.Actors[c.Slot].Fatigue = c.Stamina, c.Fatigue
				}
			}
			// 規劃過了沒有：`+0x22` 立起來的排在前面，AI 那一名接著。
			var order []int
			for i, rec := range p.State.Actors {
				raw, _ := hex.DecodeString(rec)
				if i != p.Actor && len(raw) == 50 && raw[0x22] != 0 && raw[1] != 2 {
					order = append(order, i)
				}
			}
			turn := len(order)
			order = append(order, p.Actor)
			s.Round.Order, s.Round.Turn = order, turn

			// 原版 AI 那一段：從進場時的亂數起，連續、不屬於規劃鍵值與檢視擲骰的那幾次。
			var want []int
			var after uint32
			start := -1
			for k, r := range input.Rolls {
				if r.Round == p.Round && r.Phase == "plan" && r.Before == p.State.Seed && !notAI[r.Caller] {
					start = k
					break
				}
			}
			if start < 0 {
				t.Fatalf("找不到 AI 進場那一刻的擲骰（seed %#x）", p.State.Seed)
			}
			for k := start; k < len(input.Rolls); k++ {
				r := input.Rolls[k]
				if r.Round != p.Round || r.Phase != "plan" || notAI[r.Caller] {
					after = r.Before
					break
				}
				want = append(want, r.N)
			}
			rec := &recordingRoll{lcg: rng.NewState(p.State.Seed)}
			d := s.PlanEnemy(p.Actor, rec)
			if fmt.Sprint(rec.ns) != fmt.Sprint(want) {
				t.Errorf("AI 擲骰不同：原版 %v，remake %v", want, rec.ns)
			}
			if got := rec.lcg.State(); after != 0 && got != after {
				t.Errorf("AI 之後的亂數狀態：原版 %#x，remake %#x", after, got)
			}
			var exec string
			for _, b := range input.Boundaries {
				if b.Round == p.Round && strings.HasPrefix(b.At, "執行段入口") {
					exec = b.State.Actors[p.Actor]
				}
			}
			raw, _ := hex.DecodeString(exec)
			a := s.Actors[p.Actor]
			for _, f := range []struct {
				name        string
				off, actual int
			}{
				{"指令碼 +0x25", 0x25, a.Command}, {"攻擊選項 +0x23", 0x23, a.Style},
				{"防禦選項 +0x24", 0x24, a.Guard}, {"方向 +0x28", 0x28, a.Facing},
			} {
				if len(raw) != 50 || f.actual != int(raw[f.off]) {
					t.Errorf("%s：原版 %#x，remake %#x（決策 %+v）", f.name, raw[f.off], f.actual, d)
				}
			}
		})
	}
}
