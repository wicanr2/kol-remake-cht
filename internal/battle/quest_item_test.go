package battle

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestQuestItemIdentityControlsFleeAndPickup(t *testing.T) {
	s := fleeState()
	s.Actors[1].X = FieldMin
	// 同名但狀況 byte 不同，不可完成任務或印任務提示。
	s.Party[1].HandItem.Record = "01000000000000000000000100000000"
	if s.FleeOffer().Carrying != 0 || s.FleeOffer().Ask || s.QuestItemCall(s.Party[1].HandItem) {
		t.Fatal("同名不同記錄被當成任務物品")
	}
	// 呈現名稱不同不影響身分；前次失敗不得消耗一次性提示。
	s.Party[1].HandItem.Record = s.Site.QuestItem.Record
	s.Party[1].WeaponName = "顯示名稱"
	if s.FleeOffer().Carrying != 1 || !s.QuestItemCall(s.Party[1].HandItem) || s.QuestItemCall(s.Party[1].HandItem) {
		t.Fatal("完整物品身分或一次性提示錯誤")
	}
	for _, bad := range []string{"", "01", "zz000000000000000000000000000000"} {
		if EqualQuestItemRecords(bad, bad) {
			t.Fatal("缺少或損壞記錄不能視為相等")
		}
	}
}

// 使用實際 24 個任務資料，驗證資料至規則的垂直鏈；站位受控，不冒稱 GUI。
func TestAllQuestItemsPickupAndFlee(t *testing.T) {
	sites, err := LoadSites("../../assets/original-derived-data/quest-sites.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(sites.List) != 24 {
		t.Fatalf("任務數量=%d", len(sites.List))
	}
	for _, source := range sites.List {
		t.Run(fmt.Sprintf("%02d", source.Index), func(t *testing.T) {
			site := source
			party := []Member{
				{Name: "CARRIER", Health: 60, Strength: 60, Size: 20, Quests: make([]int, 6)},
				{Name: "WATCHER", Health: 60, Strength: 60, Size: 20, Quests: make([]int, 6)},
			}
			for i := range party {
				for j := range party[i].Quests {
					party[i].Quests[j] = 0x55
				}
			}
			s := NewQuest(&site, party, Monster{}, NewField())
			item := s.Ground[0]
			if item.Item == nil || !EqualQuestItemRecords(item.Item.Record, source.QuestItem.Record) {
				t.Fatal("任務原始記錄未進入地面物品")
			}
			s.Actors[0].X, s.Actors[0].Y = item.X, item.Y
			s.Actors[1].X, s.Actors[1].Y = 30, 30
			s.PickUp = &PickUpAsk{Actor: 0, Slot: 0}
			if !s.AnswerPickUp(true) || s.Party[0].HandItem == nil || s.Ground[0].Item != nil {
				t.Fatal("正式拾取問答未將物品搬至手持")
			}
			if s.resume == nil || s.resume.Message != QuestItemRecord {
				t.Fatal("未產生任務物品提示")
			}
			s.Actors[0].X = FieldMin
			if !s.FleeOffer().Ask || s.FleeOffer().Carrying != 1 {
				t.Fatal("完整任務物品未提供逃走")
			}
			raw, err := hex.DecodeString(s.Party[0].HandItem.Record)
			if err != nil {
				t.Fatal(err)
			}
			raw[11] ^= 1
			s.Party[0].HandItem.Record = hex.EncodeToString(raw)
			if s.FleeOffer().Ask || s.FleeOffer().Carrying != 0 {
				t.Fatal("同名不同狀況物品誤判")
			}
			s.Party[0].HandItem.Record = source.QuestItem.Record
			if s.Flee() != 2 || s.Verdict != VerdictFled {
				t.Fatal("攜物逃走未標記全隊")
			}
			for _, m := range s.Party {
				for i, v := range m.Quests {
					if i == source.Index/4 {
						shift := 2 * (source.Index % 4)
						if (v>>shift)&3 != 2 || (v & ^(3<<shift)) != (0x55 & ^(3<<shift)) {
							t.Fatal("任務旗標或同 byte 其他任務錯誤")
						}
					} else if v != 0x55 {
						t.Fatal("不相關任務旗標被改寫")
					}
				}
			}
			if site.QuestItem.Record != source.QuestItem.Record {
				t.Fatal("物品操作污染任務母表")
			}
		})
	}
}
