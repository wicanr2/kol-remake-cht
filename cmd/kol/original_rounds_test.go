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
