package main

import (
	"encoding/json"
	"os"
	"testing"
)

// planningCellsReceipt 是原版逐格走規劃段的摘要（#84）。每一組從規劃入口送一串鍵，
// dosgolem 記下每一鍵之後的圖示列格數、白框位置，以及最後寫進槽 1 的四個值。
// 摘要由 `tools/summarize_planning_probes.py` 從 `battle-planning.json` 產生。
const planningCellsReceipt = "../../docs/verification/planning-cells-20260927/planning-cells.json"

type planningCellsCase struct {
	Name  string `json:"name"`
	Steps []struct {
		Key    string `json:"key"`
		Icons  int    `json:"icon_slots"`
		Framed int    `json:"framed_slot"`
	} `json:"steps"`
	Final struct {
		Command int `json:"command_code"`
		Attack  int `json:"attack_option"`
		Defence int `json:"defence_option"`
		Facing  int `json:"facing"`
	} `json:"final"`
}

// 規劃段逐格：主指令列上每一個指令、徒手攻擊式與防禦方式的每一格、移動方式的
// 每一格，照原版同一串鍵在 remake 重播。比的是機制——每一鍵之後的圖示列格數與
// 白框位置，以及寫回的指令碼、攻擊選項、防禦選項與方向——不比畫面。
//
// 鍵的對應：`.`／`,` 是 `MoveSelection(±1)`，Return 是 `Choose`（SPEC-010 §3）。
func TestOriginalNormalPlanningCells(t *testing.T) {
	body, err := os.ReadFile(planningCellsReceipt)
	if err != nil {
		t.Fatal(err)
	}
	var receipt struct {
		Cases []planningCellsCase `json:"cases"`
	}
	if err = json.Unmarshal(body, &receipt); err != nil {
		t.Fatal(err)
	}
	if len(receipt.Cases) == 0 {
		t.Fatal("收據裡沒有任何一組")
	}
	for _, c := range receipt.Cases {
		t.Run(c.Name, func(t *testing.T) {
			s := originalPlanningState(t)
			const actor = 1
			for i, st := range c.Steps {
				switch st.Key {
				case ".":
					s.MoveSelection(1)
				case ",":
					s.MoveSelection(-1)
				case "Return":
					s.Choose()
				default:
					t.Fatalf("第 %d 鍵 %q 不在對應表裡", i, st.Key)
				}
				icons := len(s.Plan.Bar)
				// 決定完這一名之後原版換到下一名的指令列；remake 這裡只排了一名，
				// 圖示列會清空。寫回之後的那一步只比四個值。
				if s.Actors[actor].Command != 0 && i == len(c.Steps)-1 {
					break
				}
				if icons != st.Icons || s.Plan.Selected != st.Framed {
					t.Errorf("第 %d 鍵 %q：remake %d 格、白框第 %d 格；原版 %d 格、第 %d 格（%#x）",
						i, st.Key, icons, s.Plan.Selected, st.Icons, st.Framed, s.Plan.Bar)
				}
			}
			a := s.Actors[actor]
			for _, v := range []struct {
				name      string
				got, want int
			}{
				{"指令碼 +0x25", a.Command, c.Final.Command},
				{"攻擊選項 +0x23", a.Style, c.Final.Attack},
				{"防禦選項 +0x24", a.Guard, c.Final.Defence},
				{"方向 +0x28", a.Facing, c.Final.Facing},
			} {
				if v.got != v.want {
					t.Errorf("%s：remake %#x，原版 %#x", v.name, v.got, v.want)
				}
			}
		})
	}
}
