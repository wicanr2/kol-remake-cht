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
		{Index: 5, X: 57, Y: 8, Picture: 13, Portrait: 12, Name: "Halfway House"},
	}}
	session := town.NewSession(out, buildings, nil, nil, nil)

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
		{"RET", SceneInn},
		{">", SceneInn},
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

// 旅店的圖示列只有 DOOR 會轉場；其餘記到 Unimplemented，不會默默吞掉。
func TestInnOnlyDoorTransitions(t *testing.T) {
	f := testFlow(t)
	f.Key("SPACE")
	f.Key("SPACE")
	f.Key("6")
	f.Key("RET")
	if f.Inn.Selected != 5 {
		t.Fatalf("旅店白框在 %d", f.Inn.Selected)
	}
	if f.Key("RET") {
		t.Fatal("COMPANION 還沒實作，不該轉場")
	}
	if len(f.Unimplemented) != 1 {
		t.Fatalf("未實作清單是 %v", f.Unimplemented)
	}
	if f.Scene != SceneInn {
		t.Fatalf("留在旅店，取到 %s", f.Scene)
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
