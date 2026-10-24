package main

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/font"
	"github.com/wicanr2/kol-remake-cht/internal/roster"
	"github.com/wicanr2/kol-remake-cht/internal/save"
	"github.com/wicanr2/kol-remake-cht/internal/screen"
	"github.com/wicanr2/kol-remake-cht/internal/town"
	"github.com/wicanr2/kol-remake-cht/internal/world"
)

func TestToolbarTextFitsButtons(t *testing.T) {
	root, e := fs.Sub(toolbarFonts, "toolbarfont")
	if e != nil {
		t.Fatal(e)
	}
	f, e := font.LoadFS(root, "toolbar")
	if e != nil {
		t.Fatal(e)
	}
	for _, set := range []struct {
		labels []string
		rects  [][4]int
	}{
		{[]string{"原貌", "HD畫面", "輔助 F8", "配樂 F9", "畫面 F10", "暫停", "快捷鍵", "收起 F12"}, nil},
		{[]string{"玩家不疲勞：關", "敵疲勞：6倍", "免怯戰：關", "角色強化（永久）", "戰鬥輔助歸零"}, nil},
	} {
		for i, label := range set.labels {
			r := toolbarTop[i]
			if len(set.labels) == 5 {
				r = toolbarQuick[i]
			}
			w, h := f.Measure(label)
			if w > r.Dx()-12 || h > r.Dy()-8 {
				t.Fatalf("文字超框：%s %dx%d", label, w, h)
			}
			dst := screen.NewCanvasScaled(3).Image()
			if f.Draw(dst, 0, 0, label, 255, 255, 255) != len([]rune(label)) {
				t.Fatal("工具列缺字", label)
			}
		}
	}
}

func TestToolbarPreferencesColdStartAndCorruption(t *testing.T) {
	path := filepath.Join(t.TempDir(), "prefs", "toolbar.json")
	p, e := loadToolbarPreferences(path)
	if e != nil || p != defaultToolbarPreferences() {
		t.Fatal(p, e)
	}
	p.Stats = "110"
	p.EnemyMultiplier = 3
	if e = saveToolbarPreferences(path, p); e != nil {
		t.Fatal(e)
	}
	got, e := loadToolbarPreferences(path)
	if e != nil || got != p {
		t.Fatal(got, e)
	}
	var cold toolbarState
	cold.prefs = got
	if cold.assist.NoFatigue || cold.assist.FearImmune || cold.assist.Multiplier() != 1 {
		t.Fatal("重開自動啟用輔助")
	}
	for _, raw := range []string{"{", `{"enemy_multiplier":99}`, `{"enemy_multiplier":6,"stats_floor":"999"}`} {
		os.WriteFile(path, []byte(raw), 0600)
		got, e = loadToolbarPreferences(path)
		if e == nil || got != defaultToolbarPreferences() {
			t.Fatal("壞偏好沒有回預設", raw, got, e)
		}
	}
}

func TestToolbarActionsReachBattleAndHiddenControlsReturn(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	s := &battle.State{Actors: []battle.Actor{{Side: battle.SideParty, Head: 10, Chest: 10, Fatigue: 300}}}
	g := &game{quest: &questBattle{state: s}}
	g.toolbar.prefs = defaultToolbarPreferences()
	g.toolbarAction(8)
	g.toolbarAction(9)
	g.toolbarAction(10)
	if s.Actors[0].Fatigue != 0 || !s.Assistance.FearImmune || s.Assistance.Multiplier() != 6 {
		t.Fatal("工具列沒有接到正式戰鬥")
	}
	g.toolbarAction(12)
	if s.Assistance.NoFatigue || s.Assistance.FearImmune || s.Assistance.Multiplier() != 1 {
		t.Fatal("歸零沒有返回原規則")
	}
	g.toolbar.quick, g.toolbar.character = true, true
	g.toolbarAction(7)
	if !g.toolbar.hidden || g.toolbar.quick || g.toolbar.character {
		t.Fatal("收合後留下不可見的設定頁")
	}
	g.toolbarAction(2)
	if g.toolbar.hidden || !g.toolbar.quick {
		t.Fatal("F8無法找回設定")
	}
}

func TestToolbarViewportMapping(t *testing.T) {
	for _, scale := range []int{1, 3, 6} {
		g := &game{interactive: true, canvas: screen.NewCanvasScaled(scale)}
		for _, quick := range []bool{false, true} {
			for _, hidden := range []bool{false, true} {
				g.toolbar.quick, g.toolbar.hidden = quick, hidden
				w, h := g.desktopSize()
				native := g.canvas.Image().Bounds().Dx()
				if w < 960 || h != 600*w/960+g.toolbarHeader()*w/960 {
					t.Fatal("桌面尺寸錯誤")
				}
				x, y := g.gamePoint(w/2, g.toolbarHeader()*w/960+100*w/960)
				if x != native/2 || y != 100*native/960 {
					t.Fatal("遊戲滑鼠座標錯誤", scale, x, y)
				}
				if !hidden {
					x, y = g.gamePoint(50, 2)
					if x != -1000 || y != -1000 {
						t.Fatal("工具列穿透遊戲")
					}
				}
			}
		}
	}
}

func TestAssistanceFullSaveAndBattleWriteGuard(t *testing.T) {
	m := town.Member{ID: "PLAYER", Name: "PLAYER", Gold: 100, Sheet: &town.Sheet{Weapons: []town.WeaponRow{{Offence: 12, Defence: 13}}}}
	m.SyncCarried()
	s := &town.Session{Party: []town.Member{m}, Outdoor: &town.Outdoor{}}
	r := &roster.Roster{Slots: make([]roster.Slot, 16)}
	r.Slots[0] = roster.Slot{Slot: 0, ID: m.ID, Name: m.Name, Occupied: true}
	w := &world.Screen{Travel: &world.Travel{}, Clock: world.NewClock()}
	g := &game{play: s, quest: &questBattle{state: &battle.State{}}}
	g.strengthenFromToolbar("gold", "9000")
	if s.Party[0].Gold != 100 || g.toolbar.message != "戰鬥結束後才能永久修改角色" {
		t.Fatal("戰鬥中改寫角色")
	}
	g.quest = nil
	for _, x := range []struct {
		k string
		n int
	}{{"stats", 110}, {"skills", 200}, {"gold", 9000}} {
		if _, e := s.Strengthen(false, x.k, x.n); e != nil {
			t.Fatal(e)
		}
	}
	out, e := buildCompleteSave(s, w, r, false)
	if e != nil {
		t.Fatal(e)
	}
	path := filepath.Join(t.TempDir(), "save.dat")
	if e = save.Write(path, out); e != nil {
		t.Fatal(e)
	}
	loaded, e := save.Read(path)
	if e != nil {
		t.Fatal(e)
	}
	party := savedParty(loaded)
	if !reflect.DeepEqual(party[0].Sheet, s.Party[0].Sheet) || party[0].Gold != 9000 {
		t.Fatal("永久強化沒有完整保存")
	}
	ss, rr := &town.Session{}, &roster.Roster{}
	if e = restoreRoster(loaded, rr, ss); e != nil {
		t.Fatal(e)
	}
	if ss.Roster[0].Member().Gold != 9000 || ss.Roster[0].Member().Sheet.Stats[0] != 165 {
		t.Fatal("名冊重載丟失強化")
	}
	raw, _ := json.Marshal(loaded)
	var fields map[string]any
	json.Unmarshal(raw, &fields)
	if loaded.Schema != save.CompleteSchema || fields["assistance"] != nil {
		t.Fatal("擴張了遊戲存檔格式")
	}
}
