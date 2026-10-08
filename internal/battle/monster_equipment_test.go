package battle

import (
	"encoding/hex"
	"encoding/json"
	"github.com/wicanr2/kol-remake-cht/internal/rng"
	"os"
	"testing"
)

func TestEquipmentV2NormalOriginal(t *testing.T) {
	path := os.Getenv("KOL_EQUIPMENT_CASE")
	if path == "" {
		t.Skip("由隔離原版收據驗收命令提供KOL_EQUIPMENT_CASE")
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var native struct {
		Gear []struct {
			Actor   int
			Phase   string
			Variant int
			Record  string `json:"record_hex"`
		} `json:"gear_copies"`
		Placement []struct {
			Label  string
			Seed   uint32
			Actors []string
		}
	}
	if err = json.Unmarshal(body, &native); err != nil {
		t.Fatal(err)
	}
	mons, err := LoadMonsters("../../assets/original-derived-data/monsters.json")
	if err != nil {
		t.Fatal(err)
	}
	mon := mons.At(9)
	s := NewQuest(&Site{Index: -1, Level: 1, Count: 3, MonsterName: mon.Name, MonsterX: []int{35, 37, 36}, MonsterY: []int{32, 29, 34}}, nil, mon, nil)
	random := rng.NewState(native.Placement[1].Seed)
	s.initializeMonsterEquipment(random)
	for j := 0; j < 3; j++ {
		a := s.Actors[j+6]
		chosen, copied := native.Gear[j*2], native.Gear[j*2+1]
		if a.EquipmentVariant != chosen.Variant {
			t.Fatalf("槽%d 裝備組%d 原版%d", j+6, a.EquipmentVariant, chosen.Variant)
		}
		raw, err := hex.DecodeString(copied.Record)
		if err != nil || len(raw) != 128 {
			t.Fatal("原版記錄錯誤")
		}
		its := []*Item{a.HandItem, a.Armour[0], a.Armour[1], a.Armour[2], a.MonsterShield, a.MonsterRanged}
		for k, it := range its {
			got := make([]byte, 16)
			if it != nil {
				got, _ = hex.DecodeString(it.Record)
			}
			if hex.EncodeToString(got) != hex.EncodeToString(raw[32+16*k:48+16*k]) {
				t.Fatalf("槽%d 裝備%d 不符", j+6, k)
			}
		}
		if a.Weapon.Kind != int(raw[45]) || a.Weapon.Weight != int(raw[42]) {
			t.Fatal("武器適配不符")
		}
	}
	seed := random.State()
	s.initializeMonsterEquipment(random)
	if random.State() != seed {
		t.Fatal("重複初始化擲骰")
	}
	s.InitializeMonsterRanged(random)
	if random.State() != native.Placement[2].Seed {
		t.Fatalf("能力初始化種子%d 原版%d", random.State(), native.Placement[2].Seed)
	}
	for i := 6; i < 9; i++ {
		raw, _ := hex.DecodeString(native.Placement[2].Actors[i])
		if s.Actors[i].Kind != int(raw[0]) {
			t.Fatalf("能力槽%d 得%d 原版%d", i, s.Actors[i].Kind, raw[0])
		}
	}
	// 可變耐久不可共用母表或其他敵人。
	a := s.Actors[6].Armour[0]
	old := mon.EquipmentVariants[2].Items[1].Condition
	a.Condition--
	if mon.EquipmentVariants[2].Items[1].Condition != old {
		t.Fatal("耐久污染母表")
	}
	t.Log("正常原版三隻怪物，18件完整裝備與能力14/8/0相符；一次性初始化通過")
}
