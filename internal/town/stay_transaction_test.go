package town

import (
	"errors"
	"testing"
)

func TestFaithfulStayCommitsRetiredRosterAndRollsBackOnlyFailingMember(t *testing.T) {
	s := innSession(t, 10, Member{Name: "A", Gold: 100, Speaker: true}, Member{Name: "B", Gold: 200})
	s.FaithfulInnStay = true
	s.Roster = []Candidate{CandidateFromMember(s.Party[0]), CandidateFromMember(s.Party[1])}
	calls := 0
	s.SaveGame = func() error {
		calls++
		if calls == 1 {
			if len(s.Party) != 1 || s.Party[0].Name != "B" || s.Roster[0].Gold != 90 || s.Roster[0].Full.Gold != 90 {
				t.Fatal("寫檔看到的不是第一人的完整成交狀態")
			}
			return nil
		}
		if len(s.Party) != 0 || s.Roster[0].Gold != 90 || s.Roster[1].Gold != 190 {
			t.Fatal("第二次保存遺失第一人或尚未離隊")
		}
		return errors.New("測試寫入失敗")
	}
	s.StartRest()
	yes(s)
	yes(s)
	yes(s)
	yes(s)
	yes(s)
	if calls != 2 || len(s.Party) != 1 || s.Party[0].Name != "B" || s.Party[0].Gold != 200 || s.Roster[0].Gold != 90 || s.Roster[1].Gold != 200 {
		t.Fatalf("逐人回滾錯誤：%+v %+v", s.Party, s.Roster)
	}
}

func TestRetiredCharacterKeepsOrderTokenAndSteedWhenRecruited(t *testing.T) {
	m := Member{Name: "A", ID: "a", Gold: 42, Order: &Order{Name: "ORDER", Allow: []int{1}}, Token: &Good{Name: "TOKEN"}, Steed: &Good{Name: "STEED"}, Sheet: &Sheet{Ward: 14}}
	c := CandidateFromMember(m)
	m.Order.Allow[0] = 9
	got := c.Member()
	got.Order.Allow[0] = 7
	if c.Full.Order.Allow[0] != 1 || got.ID != "a" || got.Token.Name != "TOKEN" || got.Steed.Name != "STEED" || got.Sheet.Ward != 14 {
		t.Fatal("退隊再招募遺失角色資料或共用可變欄位")
	}
}
