package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/battle"
)

// 戰鬥左側兩框的同狀態比較（#17，SPEC-018／019）：原版正常戰鬥打到有人受傷之後，在下一回合
// 規劃段對槽 0、1、6 呼叫左下框（疲勞直條、人形）、傷勢遮罩與左上框的裝備模式，收據記下
// 交給繪製層的參數與原版自己解出的裝備名字。remake 從同一刻的參戰者記錄建狀態，用正式的
// `BodyPanel`／`EquipmentPanel` 算同樣的東西比對。
func TestOriginalNormalBattlePanels(t *testing.T) {
	path := os.Getenv("KOL_ROUNDS_INPUT")
	if path == "" {
		t.Skip("需要 dosgolem 正常路徑的多回合收據")
	}
	var input struct {
		EndAt      string `json:"end_at"`
		Boundaries []struct {
			Round int
			At    string
			State struct {
				Actors   []string
				Counters []struct{ Slot, Stamina, Fatigue int }
			}
		}
		Panels []struct {
			Slot    int
			Broken  int
			Body    []int
			Rects   [][5]int `json:"fatigue_rects"`
			Mask    [][2]int `json:"mask_reads"`
			Records []struct {
				Offset int
				Text   string
				Y      int
			} `json:"equipment"`
		}
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(body, &input); err != nil {
		t.Fatal(err)
	}
	if len(input.Panels) == 0 {
		t.Skip("收據沒有面板擷取（舊版收據）")
	}
	last := input.Boundaries[len(input.Boundaries)-1]
	if last.At != input.EndAt {
		t.Fatalf("最後一個邊界是 %q，收據停在 %q", last.At, input.EndAt)
	}
	s := originalPlanningState(t)
	for _, p := range input.Panels {
		applyOriginalRecord(t, &s.Actors[p.Slot], last.State.Actors[p.Slot])
	}
	for _, c := range last.State.Counters {
		if c.Slot < len(s.Actors) && s.Actors[c.Slot].Side != battle.SideEmpty {
			s.Actors[c.Slot].Stamina, s.Actors[c.Slot].Fatigue = c.Stamina, c.Fatigue
		}
	}
	damaged := 0
	for _, p := range input.Panels {
		if p.Broken != 0 {
			continue
		}
		t.Run(fmt.Sprintf("槽%d", p.Slot), func(t *testing.T) {
			s.Stage, s.Plan.Prompt = battle.StageField, battle.PromptCommand
			s.Plan.BodyActor, s.Plan.EquipmentPanel = p.Slot+1, true
			d, ok := s.BodyPanel()
			if !ok {
				t.Fatal("remake 沒有畫左下框")
			}
			raw, _ := hex.DecodeString(last.State.Actors[p.Slot])
			for k := 0; k < 6; k++ {
				if d.Maximum[k] != int(raw[0x12+2*k]) || d.Current[k] != int(raw[0x11+2*k]) {
					t.Fatalf("部位 %d：原版 %d/%d，remake %d/%d", k, raw[0x11+2*k], raw[0x12+2*k], d.Current[k], d.Maximum[k])
				}
				if d.Current[k] < d.Maximum[k] {
					damaged++
				}
			}
			if want := []int{0x47 + d.Kind}; !reflect.DeepEqual(p.Body, want) {
				t.Fatalf("人形：原版 %#x，remake %#x", p.Body, want)
			}
			y, err := battle.BodyFatigueBoundary(d.Stamina, d.Fatigue)
			if err != nil {
				t.Fatal(err)
			}
			if want := [][5]int{{132, 84, 138, y, 12}, {132, y, 138, 156, 1}}; !reflect.DeepEqual(p.Rects, want) {
				t.Fatalf("疲勞直條：原版 %v，remake %v", p.Rects, want)
			}
			masks, err := battle.BodyDamageRects(d.Kind, d.Current, d.Maximum)
			if err != nil {
				t.Fatal(err)
			}
			var reads [][2]int
			for _, m := range masks {
				for y := m.Top; y < m.Bottom; y++ {
					for x := m.Left; x < m.Right; x++ {
						reads = append(reads, [2]int{x, y})
					}
				}
			}
			if len(reads) == 0 {
				reads = nil
			}
			if !reflect.DeepEqual(p.Mask, reads) {
				t.Fatalf("傷勢遮罩：原版讀 %d 點，remake %d 點", len(p.Mask), len(reads))
			}
			e, ok := s.EquipmentPanel()
			if !ok {
				t.Fatal("remake 沒有左上框")
			}
			seen := map[int]bool{}
			for _, r := range p.Records {
				row := (r.Y - 24) / 6
				if row < 0 || row >= e.Count || (r.Y-24)%6 != 0 {
					t.Fatalf("原版裝備列畫在 y=%d，不在 %d 列之內", r.Y, e.Count)
				}
				seen[row] = true
				if !strings.EqualFold(strings.TrimSpace(r.Text), e.Rows[row]) {
					t.Fatalf("第 %d 列（記錄 +%#x）：原版 %q，remake %q", row, r.Offset, r.Text, e.Rows[row])
				}
			}
			for row := 0; row < e.Count; row++ {
				if e.Rows[row] != "" && !seen[row] {
					t.Fatalf("第 %d 列 remake 有 %q，原版沒畫", row, e.Rows[row])
				}
			}
			t.Logf("人形 %#x、部位 %v/%v、疲勞分界 %d、遮罩 %d 點、裝備 %q", p.Body[0], d.Current, d.Maximum, y, len(reads), e.Rows)
		})
	}
	if damaged == 0 {
		t.Fatal("三名都沒有受傷，這一刻驗不到傷勢遮罩")
	}
	// 受控破損：同一刻把那一格改成可損、耐久 31，remake 用護甲吸收那一支打到壞
	// （`AbsorbArmour` → 耐久扣成負 → 整格清空），比壞掉之後的裝備框。
	broken := 0
	for _, p := range input.Panels {
		if p.Broken == 0 {
			continue
		}
		broken++
		pos := (p.Broken - 0x100) / 0x10
		a := &s.Actors[p.Slot]
		item := a.Armour[pos]
		if item == nil {
			t.Fatalf("槽 %d 第 %d 格沒有護甲可以打壞", p.Slot, pos)
		}
		raw, _ := hex.DecodeString(item.Record)
		raw[11], raw[15] = 31, raw[15]|0x20
		item.Record = hex.EncodeToString(raw)
		a.AbsorbArmour(200, 1, alwaysOne{})
		if a.Armour[pos] != nil {
			t.Fatal("remake 打完那一格還在")
		}
		s.Plan.BodyActor, s.Plan.EquipmentPanel = p.Slot+1, true
		e, ok := s.EquipmentPanel()
		if !ok {
			t.Fatal("remake 沒有左上框")
		}
		var got, want []string
		for _, r := range p.Records {
			want = append(want, fmt.Sprintf("%d:%s", (r.Y-24)/6, strings.ToUpper(strings.TrimSpace(r.Text))))
		}
		for row := 0; row < e.Count; row++ {
			if e.Rows[row] != "" {
				got = append(got, fmt.Sprintf("%d:%s", row, strings.ToUpper(e.Rows[row])))
			}
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("破損之後的裝備框：原版 %q，remake %q", want, got)
		}
		t.Logf("槽 %d 記錄 +%#x 破損之後：%q", p.Slot, p.Broken, got)
	}
	if broken == 0 {
		t.Fatal("收據沒有受控破損那一筆")
	}
}

// alwaysOne 讓護甲吸收的每一顆骰都擲 1。
type alwaysOne struct{}

func (alwaysOne) Roll(int) int { return 1 }
