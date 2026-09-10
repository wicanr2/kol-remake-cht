package main

import (
	"errors"
	"github.com/wicanr2/kol-remake-cht/internal/charcreate"
	"github.com/wicanr2/kol-remake-cht/internal/chardata"
	"github.com/wicanr2/kol-remake-cht/internal/flow"
	"github.com/wicanr2/kol-remake-cht/internal/roster"
	"github.com/wicanr2/kol-remake-cht/internal/save"
	"github.com/wicanr2/kol-remake-cht/internal/town"
	"github.com/wicanr2/kol-remake-cht/internal/world"
	"path/filepath"
	"reflect"
	"testing"
)

func TestNormalCreationCommitFullSaveReloadAndRetry(t *testing.T) {
	data, err := chardata.Load("../../assets/original-derived-data/character-creation.json")
	if err != nil {
		t.Fatal(err)
	}
	start, err := loadCreationStart("../../assets/original-derived-data/creation-start.json")
	if err != nil {
		t.Fatal(err)
	}
	r := &roster.Roster{Slots: make([]roster.Slot, 16)}
	r.Layout.SlotCount = 16
	s := &town.Session{Outdoor: &town.Outdoor{}}
	w := &world.Screen{Travel: &world.Travel{}, Clock: world.NewClock()}
	path := filepath.Join(t.TempDir(), "complete.json")
	persist := func() error {
		out, err := buildCompleteSave(s, w, r, false)
		if err != nil {
			return err
		}
		return save.Write(path, out)
	}
	s.SaveGame = persist
	f := flow.New(r, charcreate.New(data, newRand(1)), roster.NewDeleteFlow(r), nil, s)
	f.CommitCreate = func(c *charcreate.State) error { return commitCreated(c, start, r, s) }
	for _, key := range []string{"RET", "RET", "5", "D", "RET", "A", "RET", "RET", "RET", "RET", "D", "RET", "RET"} {
		f.Key(key)
	}
	if f.Scene != flow.SceneMenu || len(r.Occupied()) != 1 {
		t.Fatalf("建角沒有提交：%s %+v", f.Scene, f.Create)
	}
	f.Key("RET")
	f.Key("B")
	f.Key("RET")
	f.Key("RET")
	f.Key("RET")
	f.Key("RET")
	f.Key("D")
	f.Key("RET")
	s.SaveGame = func() error { return errors.New("測試磁碟失敗") }
	f.Key("RET")
	if len(r.Occupied()) != 1 || f.Create.Stage != charcreate.StageDone {
		t.Fatal("失敗建角占用名冊或丟失草稿")
	}
	s.SaveGame = persist
	f.Key("RET")
	if len(r.Occupied()) != 2 {
		t.Fatal("重試未成功")
	}
	s.SelectLeader("A")
	s.Party[0].InHand = nil
	s.Party[0].Backpack = nil
	s.Party[0].Sheet.Weapons[0].OffenceRaw = 0
	s.Party[0].SyncCarried()
	w.Clock = &world.Clock{Ticks: 800, Part: 4, Day: 27, Month: 1, Year: 2, Weather: 4}
	if err = persist(); err != nil {
		t.Fatal(err)
	}
	loaded, err := save.Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Schema != save.CompleteSchema || len(loaded.Characters.Party) != 1 || loaded.Characters.Characters[0].ID == loaded.Characters.Characters[1].ID {
		t.Fatal("完整名冊或身分錯誤")
	}
	got := savedParty(loaded)
	if !reflect.DeepEqual(got, s.Party) {
		t.Fatalf("完整角色往返不符：\n%+v\n%+v", got[0], s.Party[0])
	}
	rr := &roster.Roster{}
	ss := &town.Session{}
	if err = restoreRoster(loaded, rr, ss); err != nil {
		t.Fatal(err)
	}
	if len(rr.Occupied()) != 2 || !ss.SelectLeader("B") || ss.Party[0].Place != "TROLLSBANE" || ss.Party[0].TownCode != "BRT" {
		t.Fatal("重啟名冊無法選擇新角色")
	}
	c := loaded.Clock
	next := world.Clock{Ticks: c.Ticks, Part: c.Part, Day: c.Day, Month: c.Month, Year: c.Year, Weather: c.Weather}
	next.NextDay()
	if next.Month != 2 || next.Day != 0 {
		t.Fatal("保存時間跨月不正確")
	}
}
