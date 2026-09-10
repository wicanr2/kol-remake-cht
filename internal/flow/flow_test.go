package flow

import (
	"image"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/roster"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

func testFlow(t *testing.T) *Flow {
	t.Helper()
	slots := &roster.Roster{}
	slots.Layout.SlotCount = 16
	slots.Slots = []roster.Slot{{Slot: 0, Name: "Theodoric", Occupied: true}}

	out := &town.Outdoor{Town: 7}
	out.Camera.X, out.Camera.Y = 57, 8
	out.Bounds.MaxX, out.Bounds.MaxY = town.GridSide-1, town.GridSide-1
	grid, err := town.LoadGrid(image.NewGray(image.Rect(0, 0, town.GridSide, town.GridSide)))
	if err != nil {
		t.Fatalf("LoadGrid：%v", err)
	}
	out.SetGrid(grid)
	buildings := &town.TownBuildings{Town: 7, Doors: []town.Door{
		{Index: 5, X: 57, Y: 8, Picture: 13, Portrait: 12, Name: "Halfway House", RecordID: 86, IconFlags: [2]int{9, 0}},
	}}
	session := town.NewSession(out, buildings, nil, nil, nil)
	session.TownCode = "MIN"
	session.Roster = []town.Candidate{{Name: "Theodoric", Town: "MIN", Place: "HALFWAY HO", BuildingID: 86}}

	inn := &town.Inn{Icons: []int{town.DoorIcon, 20, 7, 6, 8, 9}, Selected: 5}
	return New(slots, nil, roster.NewDeleteFlow(slots), inn, session)
}

// 從標題一路走到戶外地圖：每一步的畫面都要照原版手冊的順序。
func TestFullChainReachesTown(t *testing.T) {
	f := testFlow(t)
	steps := []struct {
		token string
		want  Scene
	}{
		{"SPACE", SceneCover},
		{"SPACE", SceneContents},
		{"6", SceneSelect},
		{"RET", SceneTown},
		{"SPACE", SceneTown},
		{"RET", SceneTown},
	}
	for i, step := range steps {
		f.Key(step.token)
		if f.Scene != step.want {
			t.Fatalf("第 %d 步（%s）之後在 %s，應為 %s", i+1, step.token, f.Scene, step.want)
		}
	}
	if f.Town.Mode != town.ModeMap {
		t.Fatalf("進城鎮應該是地圖模式，取到 %v", f.Town.Mode)
	}
}

// 開場旅店必須接到正式招募流程，不能停在只支援 DOOR 的樣本。
func TestStartingInnRecruitsThroughSession(t *testing.T) {
	f := testFlow(t)
	f.Town.Roster = append(f.Town.Roster, town.Candidate{Name: "SECOND", Town: "MIN", Place: "HALFWAY HO", BuildingID: 86})
	f.Key("SPACE")
	f.Key("SPACE")
	f.Key("6")
	f.Key("RET")
	f.Key("SPACE")
	// DOOR 向左繞到 COMPANION，確認搜尋後選第一名候選人。
	for _, key := range []string{"<", "RET", "RET", "RET"} {
		f.Key(key)
	}
	if f.Scene != SceneTown || f.Town.Mode != town.ModeInside || len(f.Town.Party) != 2 || f.Town.Party[1].Name != "SECOND" {
		t.Fatalf("招募未接通：%s %+v", f.Scene, f.Town.Party)
	}
	if len(f.Unimplemented) != 0 || len(f.Town.Outdoor.Party) != 2 {
		t.Fatalf("招募後狀態不同步：%v", f.Unimplemented)
	}
}

func TestStartingInnRejectsWrongLocationWithoutChangingParty(t *testing.T) {
	f := testFlow(t)
	f.Town.Roster[0].Town = "BRT"
	f.Town.Party = []town.Member{{Name: "EXISTING"}}
	for _, key := range []string{"SPACE", "SPACE", "6", "RET"} {
		f.Key(key)
	}
	if f.Scene != SceneSelect || f.Town.Party[0].Name != "EXISTING" || f.Town.Inside != nil {
		t.Fatal("所在地不一致仍改變隊伍或進入旅店")
	}
}

// 角色選項頁的反白條會繞回；EDIT PICTURE／EDIT SHIELD 不轉場但要記下來。
func TestMenuCursorWrapsAndRecordsGaps(t *testing.T) {
	f := testFlow(t)
	f.Key("SPACE")
	f.Key("SPACE")
	f.Key("5")
	if f.Scene != SceneMenu || f.MenuCursor != 0 {
		t.Fatalf("目錄頁 5 之後在 %s／%d", f.Scene, f.MenuCursor)
	}
	f.Key("U")
	if f.MenuCursor != MenuCount-1 {
		t.Fatalf("反白條往上應該繞到最後一項，取到 %d", f.MenuCursor)
	}
	f.Key("D")
	f.Key("D")
	f.Key("D")
	f.Key("D") // 0 → 1 → 2 → 3（EDIT PICTURE）
	if f.MenuCursor != MenuEditPicture {
		t.Fatalf("反白條在 %d", f.MenuCursor)
	}
	if f.Key("RET") {
		t.Fatal("EDIT PICTURE 還沒實作，不該轉場")
	}
	if len(f.Unimplemented) == 0 {
		t.Fatal("未實作的功能要記下來")
	}
	f.Key("ESC")
	if f.Scene != SceneContents {
		t.Fatalf("ESC 應該回目錄頁，取到 %s", f.Scene)
	}
}

// 名冊是空的時候，目錄頁 6 走不下去——這一條原版另有路徑，先記成未實作。
func TestSelectNeedsCharacters(t *testing.T) {
	f := testFlow(t)
	f.Roster.Slots = nil
	f.Key("SPACE")
	f.Key("SPACE")
	if f.Key("6") {
		t.Fatal("名冊空的時候不該進角色選擇清單")
	}
	if f.Scene != SceneContents {
		t.Fatalf("應該留在目錄頁，取到 %s", f.Scene)
	}
}
