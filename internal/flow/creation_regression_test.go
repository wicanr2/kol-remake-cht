package flow

import (
	"github.com/wicanr2/kol-remake-cht/internal/charcreate"
	"github.com/wicanr2/kol-remake-cht/internal/roster"
	"github.com/wicanr2/kol-remake-cht/internal/town"
	"testing"
)

func TestNameAcceptsDirectionLettersAndSpace(t *testing.T) {
	f := testFlow(t)
	f.Scene = SceneNewChar
	f.Create = charcreate.New(nil, nil)
	for _, key := range []string{"L", "U", "N", "A", "SPACE", "D", "R"} {
		f.Key(key)
	}
	if f.Create.Name != "LUNA DR" {
		t.Fatalf("名字字母被當成方向鍵：%q", f.Create.Name)
	}
}

func TestCreationEntryAndLayerCancel(t *testing.T) {
	f := testFlow(t)
	f.Create = charcreate.New(nil, nil)
	for _, key := range []string{"RET", "RET", "5", "D", "RET", "A"} {
		f.Key(key)
	}
	if f.Create.Stage != charcreate.StageName || f.Create.Name != "A" {
		t.Fatalf("建角入口無法輸入：%+v", f.Create)
	}
	f.Key("RET")
	f.Key("ESC")
	if f.Scene != SceneNewChar || f.Create.Stage != charcreate.StageName || f.Create.Name != "" {
		t.Fatalf("種族層取消未回空白名字層：%s %+v", f.Scene, f.Create)
	}
	f.Key("ESC")
	f.Key("RET")
	if f.Scene != SceneNewChar || f.Create.Stage != charcreate.StageName {
		t.Fatalf("重新建角未重設：%s %+v", f.Scene, f.Create)
	}
}

func TestSelectedLeaderReachesInnAndTown(t *testing.T) {
	f := testFlow(t)
	f.Roster.Slots = append(f.Roster.Slots, roster.Slot{Slot: 1, Name: "SECOND", Occupied: true})
	f.Town.Roster = append(f.Town.Roster, town.Candidate{Name: "SECOND", Town: "MIN", Place: "HALFWAY HO", BuildingID: 86, Gold: 37, Sheet: &town.Sheet{Ward: 0x20}})
	for _, key := range []string{"RET", "RET", "6", "D", "RET", "SPACE", "RET"} {
		f.Key(key)
	}
	if f.Scene != SceneTown || len(f.Town.Party) != 1 || f.Town.Party[0].Name != "SECOND" || f.Inn.Party[0].Gold != 37 || f.Town.Party[0].Sheet.Ward != 0x20 {
		t.Fatalf("選角未傳入旅店／城鎮：%s %+v", f.Scene, f.Town.Party)
	}
	if f.Town.Outdoor.Party[0].Name != "SECOND" {
		t.Fatal("戶外畫面仍使用預設隊員")
	}
}
