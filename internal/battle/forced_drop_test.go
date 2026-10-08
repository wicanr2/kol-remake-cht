package battle

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
)

func TestForcedDropOriginalCases(t *testing.T) {
	path := os.Getenv("KOL_DROP_CASES")
	if path == "" {
		t.Skip("需要原版掉落收據")
	}
	// 原版欄名使用底線，保留實際來源而不重算預期。
	var source struct {
		Schema string           `json:"schema"`
		Cases  []map[string]any `json:"cases"`
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	if source.Schema != "kol-forced-drop-source-v1" || len(source.Cases) != 12 {
		t.Fatal("來源身份錯誤")
	}
	for _, c := range source.Cases {
		decode := func(key string) []byte {
			b, err := hex.DecodeString(c[key].(string))
			if err != nil {
				t.Fatal(err)
			}
			return b
		}
		hand, ground := decode("hand_before"), decode("ground_before")
		s := &State{Actors: make([]Actor, 7), Party: []Member{{}}}
		actor := int(c["actor"].(float64))
		armed := int(c["armed"].(float64)) != 0
		item := &Item{Name: "物品", Record: hex.EncodeToString(hand)}
		s.Actors[actor] = Actor{Side: SideParty, Member: 0, X: 30, Y: 33, Weapon: Weapon{Kind: 1}}
		if actor == 0 {
			s.Party[0] = Member{HasWeapon: armed, WeaponName: "物品", HandItem: item}
		} else {
			s.Actors[actor].Side = SideMonster
			s.Actors[actor].Member = -1
			s.Actors[actor].HandItem = item
			s.Actors[actor].Unarmed = !armed
		}
		for k := range s.Ground {
			name := ""
			if ground[k*16] != 0 {
				name = "地上物品"
			}
			s.Ground[k] = GroundItem{Name: name, X: 50 + k, Y: 10 + k, Item: &Item{Record: hex.EncodeToString(ground[k*16 : k*16+16])}}
		}
		s.forcedDrop(actor)
		actualHand := s.Actors[actor].HandItem
		if actor == 0 {
			actualHand = s.Party[0].HandItem
		}
		gotHand := make([]byte, 16)
		if actualHand != nil {
			gotHand, _ = hex.DecodeString(actualHand.Record)
		}
		if hex.EncodeToString(gotHand) != c["hand_after"].(string) {
			t.Fatal("手持與原版不同")
		}
		xs, ys := decode("xs_after"), decode("ys_after")
		for k, g := range s.Ground {
			if g.Item == nil || g.Item.Record != c["ground_after"].(string)[k*32:(k+1)*32] || g.X != int(ys[k]) || g.Y != int(xs[k]) {
				t.Fatalf("地上%d與原版不同", k)
			}
		}
	}
}

func TestExhaustionDropsCompleteHeldItem(t *testing.T) {
	item := &Item{Name: "弓", Record: "0102030405060708090a0b0c0d0e0f10"}
	s := &State{Party: []Member{{HasWeapon: true, WeaponName: item.Name, HandItem: item}}, Actors: []Actor{{Side: SideParty, Member: 0, Head: 1, Chest: 1, Command: CmdRest, Fatigue: 400, Stamina: 1, X: 30, Y: 33}}}
	if s.EndTurnFatigue(0) != ExhaustedPartyRecord || s.Party[0].HandItem != nil || s.Ground[0].Item != item {
		t.Fatal("力竭未搬完整手持")
	}
}
