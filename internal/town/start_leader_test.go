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
		s.LoadStartingTown = func(code string) ([]TownEntry, *TownBuildings, *Grid, error) {
			if code != "BRT" {
				t.Fatalf("錯誤城鎮代碼：%s", code)
			}
			id := 14
			if !valid {
				id = 99
			}
			return []TownEntry{{Code: code, Town: 2, MaxX: 63, MaxY: 63}},
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

// 同一個代碼、同一座城，旅店卻在另一個入城區：讀檔的位置放不下門口時，
// 要依代碼挑一個夾限放得下門口的入城點，不能停在名冊不動。
func TestStartLeaderSameCodeOtherEntryRegion(t *testing.T) {
	for _, reachable := range []bool{false, true} {
		s := recruitSession(t, 1)
		s.Roster = append(s.Roster, Candidate{Name: "WESTER", Town: "MIN", Place: "FARHOUSE", BuildingID: 50})
		s.Outdoor.Bounds.MinX, s.Outdoor.Bounds.MaxX = 0, 0
		s.Outdoor.Bounds.MinY, s.Outdoor.Bounds.MaxY = 0, 0
		oldName, oldEntry := s.Party[0].Name, s.Outdoor.EntryPoint
		grid, err := LoadGrid(image.NewGray(image.Rect(0, 0, GridSide, GridSide)))
		if err != nil {
			t.Fatal(err)
		}
		far := TownEntry{Code: "MIN", Town: 7, MinX: 0, MaxX: 0, MinY: 0, MaxY: 0, EntryPoint: 3}
		near := TownEntry{Code: "MIN", Town: 7, MinX: 30, MaxX: 50, MinY: 20, MaxY: 40, EntryPoint: 7}
		s.LoadStartingTown = func(code string) ([]TownEntry, *TownBuildings, *Grid, error) {
			entries := []TownEntry{far}
			if reachable {
				entries = append(entries, near)
			}
			return entries, &TownBuildings{Town: 7,
				Doors: []Door{{RecordID: 50, Name: "Farhouse", X: 40, Y: 30}}}, grid, nil
		}
		if got := s.StartLeader("WESTER"); got != reachable {
			t.Fatalf("放得下=%v，結果=%v", reachable, got)
		}
		if !reachable {
			if s.Party[0].Name != oldName || s.Outdoor.EntryPoint != oldEntry {
				t.Fatal("失敗的選角改動原有狀態")
			}
			continue
		}
		x, y := s.Outdoor.PartyPosition()
		if x != 40 || y != 30 || s.Outdoor.EntryPoint != 7 || s.Mode != ModeInside || s.Party[0].Name != "WESTER" {
			t.Fatalf("選角沒有換到放得下門口的入城點：座標=%d,%d 入城點=%d", x, y, s.Outdoor.EntryPoint)
		}
	}
}
