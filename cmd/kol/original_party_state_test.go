package main

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/rng"
	"github.com/wicanr2/kol-remake-cht/internal/roster"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

// 同一份原版角色資料先經抽取器，再走正式適配器；預期值取自原版參戰記錄。
func TestOriginalNormalPartyState(t *testing.T) {
	path := os.Getenv("KOL_PARTY_STATE_INPUT")
	if path == "" {
		t.Skip("需要 dosgolem 正常入口角色資料")
	}
	var input struct {
		Source string
		Seed   int
		Party  []roster.Slot
		Monster battle.Monster
		MonsterLevel int `json:"monster_level"`
		Actors []struct {
			Slot   int
			Record string `json:"record_hex"`
			X, Y   int
		}
		Counters []struct{ Slot, Stamina, Fatigue int }
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(body, &input); err != nil {
		t.Fatal(err)
	}
	if input.Source != "dosgolem-normal-player-path" || input.Seed != 1 || len(input.Party) != 2 {
		t.Fatal("來源、種子或隊伍不符")
	}
	site := &battle.Site{Index: -1}
	var party []battle.Member
	for i, slot := range input.Party {
		if slot.Slot != i || slot.Sheet == nil {
			t.Fatal("角色槽位或資料頁缺漏")
		}
		m := town.Member{Name: slot.Name, Wounds: slot.Wounds, Sheet: townSheet(slot.Sheet),
			InHand: townGood(slot.InHand), Worn: townWorn(slot.Equipment)}
		party = append(party, battleMember(m))
		site.PartyX = append(site.PartyX, input.Actors[i].X)
		site.PartyY = append(site.PartyY, input.Actors[i].Y)
	}
	if input.Monster.Name != "SLEDGE" || input.MonsterLevel != 3 {
		t.Fatal("正常遭遇怪物身分或等級不符")
	}
	site.Count, site.Level, site.MonsterName = 1, input.MonsterLevel, input.Monster.Name
	site.Monster = input.Monster.Index
	site.MonsterX, site.MonsterY = []int{input.Actors[6].X}, []int{input.Actors[6].Y}
	s := battle.NewQuest(site, party, input.Monster, nil)
	// SPEC-029：本次 SLEDGE 的機率為零，任何固定種子皆產生能力零。
	// 初始化亂數與規劃入口的共同種子分開；不宣稱重播整段遭遇亂數流。
	if input.Monster.RangedChance != 0 || input.Monster.RangedCode != 0 {
		t.Fatal("原版 SLEDGE 遠程初始化欄位不符")
	}
	s.InitializeMonsterRanged(rng.New(input.Seed))
	s.Rand = rng.New(input.Seed)
	checks := 0
	for _, i := range []int{0, 1, 6} {
		raw, err := hex.DecodeString(input.Actors[i].Record)
		if err != nil || len(raw) != 50 || input.Actors[i].Slot != i {
			t.Fatal("參戰記錄不符")
		}
		a := s.Actors[i]
		for _, field := range []struct {
			Name           string
			Offset, Actual int
		}{
			{"Kind", 0, a.Kind}, {"Side", 1, int(a.Side)}, {"Level", 2, a.Level},
			{"Strength", 8, a.Strength}, {"Quickness", 9, a.Quickness},
			{"Foresight", 10, a.Foresight}, {"Size", 11, a.Size}, {"Intellect", 12, a.Intellect},
			{"Damage", 14, a.Damage}, {"Threshold", 16, a.Threshold},
			{"Head", 17, a.Head}, {"Chest", 19, a.Chest}, {"RightArm", 21, a.RightArm},
			{"LeftArm", 23, a.LeftArm}, {"RightLeg", 25, a.RightLeg}, {"LeftLeg", 27, a.LeftLeg},
			{"X", 31, a.X}, {"Y", 32, a.Y},
		} {
			checks++
			if field.Actual != int(raw[field.Offset]) {
				t.Errorf("%s %s 參戰記錄 +%02X：原版 %d，remake %d", a.Name, field.Name, field.Offset, raw[field.Offset], field.Actual)
			}
		}
		found := false
		for _, c := range input.Counters {
			if c.Slot != i {
				continue
			}
			found = true
			checks += 2
			if a.Stamina != c.Stamina || a.Fatigue != c.Fatigue {
				t.Errorf("%s：原版耐力／疲勞 %d/%d，remake %d/%d", a.Name, c.Stamina, c.Fatigue, a.Stamina, a.Fatigue)
			}
		}
		if !found {
			t.Fatal("缺少耐力／疲勞原版記錄")
		}
	}
	t.Logf("兩名隊員與 SLEDGE 共比較 %d 個初始欄位；未涵蓋完整裝備能力、地形、規劃與攻擊結算", checks)
}
