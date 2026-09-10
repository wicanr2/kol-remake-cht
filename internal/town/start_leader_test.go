package town

import (
	"image"
	"testing"
)

func TestStartLeaderAcrossTownsCommitsOnlyValidLodging(t *testing.T) {
	for _, valid := range []bool{false, true} {
		s := recruitSession(t, 1)
		oldOutdoor, oldTown, oldCode := s.Outdoor, s.Outdoor.Town, s.TownCode
		oldName := s.Party[0].Name
		s.Roster = append(s.Roster, Candidate{Name: "TRAVELLER", Town: "BRT", Place: "TROLLSBANE", BuildingID: 14})
		grid, err := LoadGrid(image.NewGray(image.Rect(0, 0, GridSide, GridSide)))
		if err != nil {
			t.Fatal(err)
		}
		s.LoadStartingTown = func(code string) (TownEntry, *TownBuildings, *Grid, error) {
			if code != "BRT" {
				t.Fatalf("錯誤城鎮代碼：%s", code)
			}
			id := 14
			if !valid {
				id = 99
			}
			return TownEntry{Code: code, Town: 2, MaxX: 63, MaxY: 63},
				&TownBuildings{Town: 2, Doors: []Door{{RecordID: id, Name: "Trollsbane", X: 20, Y: 21}}}, grid, nil
		}
		if got := s.StartLeader("TRAVELLER"); got != valid {
			t.Fatalf("有效=%v，結果=%v", valid, got)
		}
		if s.Outdoor != oldOutdoor {
			t.Fatal("切換遺失上層持有的戶外指標")
		}
		if !valid {
			if s.Outdoor.Town != oldTown || s.TownCode != oldCode || s.Party[0].Name != oldName {
				t.Fatal("失敗的跨城選角改動原有狀態")
			}
			continue
		}
		x, y := s.Outdoor.PartyPosition()
		if s.TownCode != "BRT" || s.Outdoor.Town != 2 || s.Mode != ModeInside || x != 20 || y != 21 || s.Party[0].Name != "TRAVELLER" {
			t.Fatalf("跨城選角未完整切換：城鎮=%s 座標=%d,%d", s.TownCode, x, y)
		}
	}
}
