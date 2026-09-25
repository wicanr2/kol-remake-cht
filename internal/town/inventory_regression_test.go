package town

import (
	"encoding/hex"
	"testing"
)

func TestQuestTurnInDoesNotRestoreConsumedItem(t *testing.T) {
	item := Good{Name: "TOKEN", Record: "1102030405060708090a0b0c0d0e0f10"}
	reward := Good{Name: "REWARD", Price: 31, Record: "2202030405060708090a0b0c0d0e0f10"}
	wants, _ := hex.DecodeString(item.Record)
	gives, _ := hex.DecodeString(reward.Record)
	for _, hand := range []bool{true, false} {
		m := Member{Quests: []int{QuestDone, 0, 0, 0, 0, 0}, Progress: make([]int, 5)}
		if hand {
			m.InHand = &item
		} else {
			m.Backpack = []Good{item}
		}
		m.SyncCarried()
		slot := m.SwapQuestItem(&QuestBuilding{Site: 0, Wants: wants, Gives: gives, GivesItem: &reward})
		if slot < 0 || !ReportQuest(m.Quests, m.Progress, 0, 0) || m.Progress[0] != 1 {
			t.Fatalf("未完成交付：%d", slot)
		}
		m.SyncCarried()
		for _, raw := range m.Carried {
			if sameItem(raw, wants) {
				t.Fatal("交付後原物品復活")
			}
		}
		var got *Good
		if hand {
			got = m.InHand
		} else {
			got = &m.Backpack[0]
		}
		if got == nil || got.Name != reward.Name || got.Price != reward.Price {
			t.Fatalf("獎勵資料遺失：%+v", got)
		}
	}
}
