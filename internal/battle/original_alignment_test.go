package battle

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/rng"
)

// 使用 dosgolem 正常入口收據建立受控比較情境；不宣稱完整規則或正常 remake GUI 已對拍。
func TestOriginalBattleActorAlignment(t *testing.T) {
	path := os.Getenv("KOL_ALIGNMENT_INPUT")
	if path == "" {
		t.Skip("需指定 dosgolem 正常戰鬥快照")
	}
	type row struct {
		Slot   int    `json:"slot"`
		Name   string `json:"name"`
		Record string `json:"record_hex"`
		Side   int    `json:"side"`
		X      int    `json:"x"`
		Y      int    `json:"y"`
	}
	var in struct {
		Schema  string `json:"schema"`
		Source  string `json:"source"`
		Current int    `json:"current_actor"`
		Count   int    `json:"enemy_count"`
		Limit   int    `json:"active_slot_limit"`
		Seed    uint32 `json:"alignment_seed"`
		Actors  []row  `json:"actors"`
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(body, &in); err != nil {
		t.Fatal(err)
	}
	if in.Schema != "kol-battle-alignment-v1" || in.Source != "dosgolem-normal-player-path" || len(in.Actors) != MaxActors || in.Count < 1 || in.Count > 12 || in.Limit != PartySlots+in.Count {
		t.Fatal("快照來源或參戰範圍不符")
	}
	if in.Current < 0 || in.Current >= PartySlots || in.Seed != 1 {
		t.Fatal("目前角色或共同 seed 不符")
	}
	site := &Site{Index: -1, Count: in.Count, PartyX: make([]int, PartySlots), PartyY: make([]int, PartySlots)}
	var party []Member
	seen := map[[2]int]bool{}
	for i, r := range in.Actors {
		raw, err := hex.DecodeString(r.Record)
		if err != nil || len(raw) != 50 || r.Slot != i || int(raw[1]) != r.Side || int(raw[0x1f]) != r.X || int(raw[0x20]) != r.Y {
			t.Fatalf("槽位 %d 與原始記錄不符", i)
		}
		if i >= in.Limit || r.Side == int(SideEmpty) {
			continue
		}
		if r.X < 0 || r.X >= FieldSide || r.Y < 0 || r.Y >= FieldSide {
			t.Fatalf("槽位 %d 座標超界", i)
		}
		pos := [2]int{r.X, r.Y}
		if seen[pos] {
			t.Fatal("有效角色站位重疊")
		}
		seen[pos] = true
		if i < PartySlots {
			if r.Side != int(SideParty) || i != len(party) || r.Name == "" {
				t.Fatal("本情境隊員槽位不連續或身分缺漏")
			}
			party = append(party, Member{Name: r.Name})
			site.PartyX[i], site.PartyY[i] = r.X, r.Y
		} else {
			if r.Side != int(SideMonster) {
				t.Fatal("有效敵人槽位的陣營不符")
			}
			site.MonsterX = append(site.MonsterX, r.X)
			site.MonsterY = append(site.MonsterY, r.Y)
		}
	}
	if len(party) != 2 || party[0].Name != "Olear" || party[1].Name != "Theodoric" {
		t.Fatal("不是已驗證的正常 Olear／Theodoric 隊伍")
	}
	if len(site.MonsterX) != in.Count {
		t.Fatal("有效敵人數不符")
	}
	s := NewQuest(site, party, Monster{}, nil)
	s.Stage = StageField
	s.Cursor = in.Current
	s.Round = Round{Order: []int{in.Current}}
	r := rng.New(int(in.Seed))
	s.Rand = r
	if r.State() != in.Seed {
		t.Fatal("remake 未設定共同 seed")
	}
	var out []row
	for i, a := range s.Actors {
		want := in.Actors[i]
		if i >= in.Limit || want.Side == int(SideEmpty) {
			if a.Side != SideEmpty {
				t.Fatalf("未使用槽位 %d 被誤建為角色", i)
			}
			continue
		}
		if a.Side != Side(want.Side) || a.X != want.X || a.Y != want.Y {
			t.Fatalf("槽位 %d 的陣營或站位未對齊", i)
		}
		if i < PartySlots && a.Name != want.Name {
			t.Fatalf("槽位 %d 身分錯誤", i)
		}
		sprite := SpriteMonster
		if i < PartySlots {
			sprite = SpritePartyBase + i
		}
		if s.SpriteAt(want.X, want.Y) != sprite {
			t.Fatalf("槽位 %d 的地圖標記錯誤", i)
		}
		out = append(out, row{Slot: i, Name: a.Name, Side: int(a.Side), X: a.X, Y: a.Y})
	}
	x, y := s.Camera()
	current := in.Actors[in.Current]
	if s.Current() != in.Current || x != current.X || y != current.Y {
		t.Fatal("目前角色與鏡頭中心未對齊")
	}
	result := map[string]any{"current_actor": s.Current(), "current_name": current.Name, "camera": []int{x, y}, "seed": r.State(), "actors": out, "scope": "角色、站位、目前角色、鏡頭中心及地圖標記；未驗地形、完整屬性、規劃順序或結算"}
	if output := os.Getenv("KOL_ALIGNMENT_OUTPUT"); output != "" {
		b, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(output, append(b, '\n'), 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("目前角色 %s，鏡頭 (%d,%d)，有效參戰者 %d，seed=%d", current.Name, x, y, len(out), r.State())
}
