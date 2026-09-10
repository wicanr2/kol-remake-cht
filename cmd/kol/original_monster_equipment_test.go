package main

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/battle"
)

// 怪物的裝備能力：原版戰鬥期記錄 ↔ monsters.json ↔ remake 的參戰者。
//
// 先前只比了六個屬性、武器記錄與遠程那兩格（`TestOriginalNormalPartyState` 的 60 欄
// 涵蓋的部分）。收據當時記著「怪物記錄目前只擷取 0x80 byte，不能當成全部穿戴欄已
// 完整保存」——**那句話有一半不成立**：0x80 就是完整的一筆（`0xCD92 + (index-6)*0x80`，
// RE-080／SPEC-017，`TestMonsterTemplateCopyMatchesOriginal` 也用 `bodySize = 128`），
// 護甲三件、盾牌、技能、旗標、逃走門檻與金幣全都在裡面。真正裝不下的只有母表
// `worn` 的第 5、6 槽（`+0x80`／`+0x90`）——而那兩格原版在戰鬥中根本不讀
// （`docs/re/20260910-monster-unarmed-slots.md`）。
//
// 所以這一支不需要重新擷取，只是把已經在手上的 128 個 byte 拿去比。三方比對：
//
//  1. 戰鬥期記錄的 bytes ↔ `monsters.json`（母表抽取器讀對了沒）
//  2. `monsters.json` ↔ remake 的 `Actor`（適配器接對了沒）
func TestOriginalNormalMonsterEquipment(t *testing.T) {
	statePath := os.Getenv("KOL_PARTY_STATE_INPUT")
	recordsPath := os.Getenv("KOL_STATE_RECORDS_INPUT")
	if statePath == "" || recordsPath == "" {
		t.Skip("需要 dosgolem 正常入口的 fixture 與原始記錄")
	}

	var input struct {
		Source       string
		Seed         int
		Monster      battle.Monster
		MonsterLevel int `json:"monster_level"`
		Actors       []struct {
			Slot int
			X, Y int
		}
	}
	body, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(body, &input); err != nil {
		t.Fatal(err)
	}

	var records struct {
		Source   string
		Schema   string
		Monsters []struct {
			Slot   int
			Record string `json:"record_hex"`
		}
	}
	body, err = os.ReadFile(recordsPath)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(body, &records); err != nil {
		t.Fatal(err)
	}
	if records.Source != input.Source {
		t.Fatalf("兩份資料來源不同：%q vs %q", records.Source, input.Source)
	}
	if len(records.Monsters) != 1 || records.Monsters[0].Slot != 6 {
		t.Fatal("這一場的怪物不是單一隻、也不在第 6 槽")
	}
	rec, err := hex.DecodeString(records.Monsters[0].Record)
	if err != nil {
		t.Fatal(err)
	}
	// 步幅就是這麼大。少一個 byte 都不能照下面的位移讀。
	if len(rec) != 128 {
		t.Fatalf("怪物本體記錄是 %d byte，原版步幅是 128", len(rec))
	}

	m := input.Monster
	be := func(off int) int { return int(rec[off])<<8 | int(rec[off+1]) }
	checks := 0
	cmp := func(name string, got, want int) {
		t.Helper()
		checks++
		if got != want {
			t.Errorf("%s：原版記錄 %d，monsters.json %d", name, want, got)
		}
	}

	// 一、母表抽取器讀對了沒。位移取自 tools/extract_monsters.py 的常數。
	cmp("旗標 +0x10", m.Flags, int(rec[0x10]))
	cmp("撤退門檻 +0x0D", m.FleeAt, int(rec[0x0D]))
	cmp("撤退恢復 +0x0E", m.FleeRecover, int(rec[0x0E]))
	cmp("金冠骰式 +0x0F", m.Gold.Raw, int(rec[0x0F]))
	cmp("攻擊技能 +0x17", m.AttackSkill, be(0x17))
	cmp("招架技能 +0x19", m.ParrySkill, be(0x19))

	// 護甲重量是三件的和，空的那一件不算（`extract_monsters.py` 的 ARMOUR_OFFSETS）。
	armour := 0
	worn := 0
	for _, off := range []int{0x30, 0x40, 0x50} {
		empty := true
		for _, b := range rec[off : off+16] {
			if b != 0 {
				empty = false
				break
			}
		}
		if !empty {
			armour += int(rec[off+10])
		}
	}
	cmp("護甲重量（三件之和）", m.ArmourWeight, armour)

	// 盾牌在 +0x60；整格全零就是沒有。
	shieldEmpty := true
	for _, b := range rec[0x60:0x70] {
		if b != 0 {
			shieldEmpty = false
			break
		}
	}
	checks++
	if shieldEmpty != (m.Shield == nil) {
		t.Errorf("盾牌 +0x60：原版記錄%s，monsters.json%s",
			map[bool]string{true: "是空的", false: "有東西"}[shieldEmpty],
			map[bool]string{true: "是 nil", false: "有一件"}[m.Shield == nil])
	}

	// 穿戴欄前五槽（+0x30…+0x7F）逐格比「有沒有東西」。第 5、6 槽不在這 128 byte 裡，
	// 而原版戰鬥中也不讀它們，所以這裡只驗裝得下的那五格。
	//
	// **`monsters.json` 的兩個欄位對「空」的約定不一樣**：`shield` 用 `item()`，
	// 全零回 `null`；`worn` 要保持七格對位，所以全零那幾格解出來是一個**每個欄位
	// 都是零的物件**而不是 `null`。拿 `!= nil` 判斷有沒有東西會把它們算成有——
	// 第一版就是這樣，SLEDGE 的第 3、4 槽（原版是空的）被判成不符。
	for slot := 0; slot < 5; slot++ {
		off := 0x30 + slot*0x10
		empty := true
		for _, b := range rec[off : off+16] {
			if b != 0 {
				empty = false
				break
			}
		}
		checks++
		has := slot < len(m.Worn) && m.Worn[slot] != nil && !blankRecord(m.Worn[slot].Record)
		if empty == has {
			t.Errorf("穿戴欄第 %d 槽（+%#04x）：原版記錄%s，monsters.json%s", slot, off,
				map[bool]string{true: "是空的", false: "有東西"}[empty],
				map[bool]string{true: "有一件", false: "是 nil"}[has])
		}
		if has {
			worn++
		}
	}

	// 二、適配器把這些接進參戰者了沒。
	site := &battle.Site{Index: -1, Count: 1, Level: input.MonsterLevel,
		MonsterName: m.Name, Monster: m.Index,
		MonsterX: []int{input.Actors[6].X}, MonsterY: []int{input.Actors[6].Y}}
	s := battle.NewQuest(site, nil, m, nil)
	a := s.Actors[6]
	cmp("Actor.ArmourWeight", a.ArmourWeight, armour)
	cmp("Actor.BaseAttack", a.BaseAttack, be(0x17))
	cmp("Actor.BaseParry", a.BaseParry, be(0x19))
	cmp("Actor.FleeAt", a.FleeAt, int(rec[0x0D]))
	cmp("Actor.FleeRecover", a.FleeRecover, int(rec[0x0E]))
	cmp("Actor.GoldRoll", a.GoldRoll, int(rec[0x0F]))
	checks++
	if (a.ShieldGrade == battle.NoShield) != shieldEmpty {
		t.Errorf("Actor.ShieldGrade：原版記錄的盾牌格%s，remake %s",
			map[bool]string{true: "是空的", false: "有東西"}[shieldEmpty],
			map[bool]string{true: "沒有盾牌", false: "有盾牌"}[a.ShieldGrade == battle.NoShield])
	}

	t.Logf("%s：本體記錄 128 byte 比對 %d 項通過（穿戴欄前五槽有 %d 件；"+
		"第 5、6 槽不在這 128 byte 裡，原版戰鬥中也不讀）", m.Name, checks, worn)
}

// blankRecord 判斷一份 16-byte 物品記錄是不是整片零。
//
// `monsters.json` 的 `worn` 為了保持七格對位，空的那幾格解出來是「每個欄位都是零
// 的物件」而不是 `null`，所以判斷有沒有東西要看記錄本身。
func blankRecord(record string) bool {
	raw, err := hex.DecodeString(record)
	if err != nil {
		return false
	}
	for _, b := range raw {
		if b != 0 {
			return false
		}
	}
	return true
}
