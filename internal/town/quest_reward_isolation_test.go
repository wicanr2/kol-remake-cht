package town

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestQuestRewardDoesNotAliasBuildingOrOtherMember(t *testing.T) {
	const wantRecord = "01000000000000000000000000000000"
	const giveRecord = "02000000000000000000006400000000"
	wants, _ := hex.DecodeString(wantRecord)
	gives, _ := hex.DecodeString(giveRecord)
	for _, slot := range []int{0, 1, 7, 16} {
		t.Run(fmt.Sprint(slot), func(t *testing.T) {
			b := &QuestBuilding{Site: 3, Wants: wants, Gives: gives,
				GivesItem: &Good{Name: "REWARD", Record: giveRecord, Condition: 100}}
			makeMember := func() Member {
				m := Member{Quests: make([]int, 6), Progress: make([]int, 5)}
				SetQuestFlag(m.Quests, 3, QuestDone)
				item := &Good{Name: "QUEST", Record: wantRecord}
				switch slot {
				case 0:
					m.InHand = item
				case 1:
					m.Backpack = []Good{*item}
				default:
					m.Worn[slot-7] = item
				}
				m.SyncCarried()
				if got := m.SwapQuestItem(b); got != slot || !ReportQuest(m.Quests, m.Progress, 0, b.Site) || m.Progress[0] != 1 {
					t.Fatalf("交付結果=%d", got)
				}
				return m
			}
			first, second := makeMember(), makeMember()
			get := func(m *Member) *Good {
				if slot == 0 {
					return m.InHand
				}
				if slot == 1 {
					return &m.Backpack[0]
				}
				return m.Worn[slot-7]
			}
			a, other := get(&first), get(&second)
			if a == other || a == b.GivesItem || other == b.GivesItem {
				t.Fatal("獎勵物品共用指標")
			}
			a.Condition = 7
			a.Record = "02000000000000000000000700000000"
			first.SyncCarried()
			if other.Condition != 100 || other.Record != giveRecord || b.GivesItem.Condition != 100 || b.GivesItem.Record != giveRecord {
				t.Fatal("修改污染其他角色或建築母表")
			}
			if first.Carried[slot][11] != 7 || second.Carried[slot][11] != 100 || b.Gives[11] != 100 {
				t.Fatal("原始記錄隔離錯誤")
			}
		})
	}
}
