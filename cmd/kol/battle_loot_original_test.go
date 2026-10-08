package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/rng"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

// 直接比較固定種子的真實原版出口收據，不以重製公式生成預期值。
func TestBattleLootOriginalCases(t *testing.T) {
	path := os.Getenv("KOL_LOOT_CASES")
	if path == "" {
		t.Skip("需要本機dosgolem原版搜刮案例收據")
	}
	var input struct {
		Schema string `json:"schema"`
		Cases  []struct {
			Threshold, Side, Down, Protect, Skip, Calls int
			Seed, FinalSeed                             uint32
			Before, After                               []string
		} `json:"cases"`
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(body, &input); err != nil {
		t.Fatal(err)
	}
	if input.Schema != "kol-loot-source-cases-v1" || len(input.Cases) != 144 {
		t.Fatal("原版案例身份或數量錯誤")
	}
	if extraPath := os.Getenv("KOL_LOOT_RESIDUAL_CASES"); extraPath != "" {
		extra := input
		extra.Cases = nil
		body, err := os.ReadFile(extraPath)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(body, &extra); err != nil {
			t.Fatal(err)
		}
		if extra.Schema != input.Schema || len(extra.Cases) != 144 {
			t.Fatal("空槽殘留原版案例身份錯誤")
		}
		input.Cases = append(input.Cases, extra.Cases...)
	}
	for n, c := range input.Cases {
		t.Run(fmt.Sprintf("原版%d", n), func(t *testing.T) {
			s := &battle.State{Stage: battle.StageDone, Verdict: battle.VerdictLost,
				Monster: battle.Monster{LootChance: c.Threshold}, Site: &battle.Site{Index: -1},
				Party: make([]battle.Member, 6), Actors: make([]battle.Actor, 6)}
			if c.Skip != 0 {
				s.Verdict = battle.VerdictWon
			}
			play := &town.Session{Party: make([]town.Member, 6)}
			for i := range s.Party {
				s.Actors[i] = battle.Actor{Side: battle.SideEmpty, Member: i, Head: 1, Chest: 1}
				raw, err := hex.DecodeString(c.Before[i])
				if err != nil || len(raw) != 768 {
					t.Fatal("原版角色記錄損壞")
				}
				item := func(off int) *battle.Item {
					if raw[off] == 0 {
						return nil
					}
					return &battle.Item{Name: fmt.Sprintf("槽%x", off), Record: hex.EncodeToString(raw[off : off+16])}
				}
				for k := range s.Party[i].PackItems {
					s.Party[i].PackItems[k] = item(0xA0 + 16*k)
					if it := s.Party[i].PackItems[k]; it != nil {
						s.Party[i].Pack[k] = it.Name
					}
				}
				for k := range s.Party[i].WornItems {
					s.Party[i].WornItems[k] = item(0x100 + 16*k)
				}
				for k := range s.Actors[i].Armour {
					s.Actors[i].Armour[k] = s.Party[i].WornItems[k]
				}
				s.Party[i].BeltItem = s.Party[i].WornItems[9]
				s.Party[i].HandItem = item(0x30)
				s.Party[i].Wounds = int(raw[0x1B])
				m := &play.Party[i]
				m.Wounds, m.Gold = int(raw[0x1B]), int(raw[0x2A])<<16|int(raw[0x2B])<<8|int(raw[0x2C])
				m.InHand = battleGood(s.Party[i].HandItem)
				m.Backpack = make([]town.Good, 6)
				for k, it := range s.Party[i].PackItems {
					if good := battleGood(it); good != nil {
						m.Backpack[k] = *good
					}
				}
				for k, it := range s.Party[i].WornItems {
					m.Worn[k] = battleGood(it)
				}
				m.Carried = append(m.Carried, append([]byte(nil), raw[0x30:0x40]...))
				for off := 0xA0; off < 0x1A0; off += 16 {
					m.Carried = append(m.Carried, append([]byte(nil), raw[off:off+16]...))
				}
			}
			s.Actors[3].Side = battle.Side(c.Side)
			s.Actors[3].Damage = 23
			if c.Down != 0 {
				s.Actors[3].Head = 0
			}
			if c.Protect != 0 {
				s.Site.Index = 0
				s.Site.QuestItem.Record = c.Before[3][0x100*2 : 0x110*2]
			}
			r := &lootRecordingRand{LCG: rng.NewState(c.Seed)}
			s.LootDownParty(r)
			g := &game{play: play}
			g.writeBattleParty(s)
			for i, m := range s.Party {
				original, err := hex.DecodeString(c.After[i])
				if err != nil {
					t.Fatal(err)
				}
				if play.Party[i].Wounds != int(original[0x1B]) || play.Party[i].Gold != int(original[0x2A])<<16|int(original[0x2B])<<8|int(original[0x2C]) {
					t.Fatalf("角色%d保存傷勢／錢包與原版不同", i)
				}
				if hex.EncodeToString(play.Party[i].Carried[0]) != c.After[i][0x30*2:0x40*2] {
					t.Fatal("搜刮修改了手持")
				}
				for slot := 0; slot < 16; slot++ {
					got := play.Party[i].Carried[1+slot]
					off := 0xA0 + 16*slot
					if hex.EncodeToString(got) != c.After[i][off*2:(off+16)*2] {
						t.Fatalf("角色%d 槽%x與原版不同", i, off)
					}
				}
				wantGoldLost := c.After[i][0x2A*2:0x2D*2] == "000000"
				if m.GoldLooted != wantGoldLost {
					t.Fatalf("角色%d錢包搜刮與原版不同", i)
				}
			}
			if r.Calls != c.Calls || r.LCG.State() != c.FinalSeed {
				t.Fatalf("骰數／種子與原版不同：%d／%d", r.Calls, r.LCG.State())
			}
			s.LootDownParty(r)
			if r.Calls != c.Calls || r.LCG.State() != c.FinalSeed {
				t.Fatal("出口重入再次搜刮")
			}
		})
	}
}

type lootRecordingRand struct {
	LCG   *rng.LCG
	Calls int
}

func (r *lootRecordingRand) Roll(n int) int { r.Calls++; return r.LCG.Roll(n) }
