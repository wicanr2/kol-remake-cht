package main

import (
	"image"
	"reflect"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/flow"
	"github.com/wicanr2/kol-remake-cht/internal/input"
	"github.com/wicanr2/kol-remake-cht/internal/town"
	"github.com/wicanr2/kol-remake-cht/internal/world"
)

func TestInsideListClicksWorkBeyondRecruiting(t *testing.T) {
	for _, cjk := range []bool{false, true} {
		in := &town.Inn{Icons: []int{town.IconSprite(town.IconYes), town.IconSprite(town.IconUTurn)}, Selected: 1, List: &town.InnList{Rows: []string{"A", "B", "C"}, Selected: 0, Total: 12}}
		g := &game{flow: &flow.Flow{Scene: flow.SceneTown, Town: &town.Session{Mode: town.ModeInside, Inside: in}}}
		if cjk { // 實際入口用 strings 決定版面；純幾何另驗兩套。
			g.flow = nil
		}
		targets := insideTargets(in, cjk)
		rows := recruitTargets(in, cjk)
		p := rows[2].Rect.Min.Add(image.Pt(1, 1))
		got := input.Hit(targets, p.X, p.Y, "")
		if !reflect.DeepEqual(got, []string{"D", "D", "<", "RET"}) {
			t.Fatalf("cjk=%v %v", cjk, got)
		}
		if !cjk && len(g.flowTargets()) != len(targets) {
			t.Fatal("正式分派未接非招募清單")
		}
	}
}

func TestSpellMenuClickUsesDrawGeometry(t *testing.T) {
	for _, cjk := range []bool{false, true} {
		p := &town.SpellEditPage{Menu: &town.InnList{Rows: []string{"A", "B"}, Selected: 0}, Fields: []town.SpellFieldCell{{Col: 7}}}
		in := &town.Inn{Selected: 1, Icons: []int{10, 18}, SpellEdit: p}
		x, y, _, step := town.SpellMenuGeometry(p, cjk)
		got := input.Hit(insideTargets(in, cjk), x+1, y+step+1, "")
		if !reflect.DeepEqual(got, []string{"D", "<", "RET"}) {
			t.Fatalf("cjk=%v %v", cjk, got)
		}
	}
}

func TestWorldPromptArrowsSelectAndWheelCannotAcknowledge(t *testing.T) {
	s := &world.Session{Screen: &world.Screen{}}
	s.SetPrompt(&world.Prompt{Kind: world.PromptRideAway})
	for _, key := range []string{"R", "U", "D", "L", ">", "<"} {
		before := s.Cursor
		if err := applyWorldToken(s, key); err != nil {
			t.Fatal(err)
		}
		if s.Cursor == before {
			t.Fatalf("%s 沒有切換選項", key)
		}
	}
	g := &game{flow: &flow.Flow{Scene: flow.SceneTown}, play: &town.Session{LeftTown: true}, worldSession: s}
	if g.wheelAllowed() {
		t.Fatal("滾輪可穿過遭遇")
	}
	if len(worldTargets(s, false)) != len(s.Screen.Icons) {
		t.Fatal("遭遇問句圖示不可點")
	}
	s.Prompt = &world.Prompt{Kind: world.PromptNotice}
	if g.wheelAllowed() {
		t.Fatal("滾輪可確認通知")
	}
	s.Prompt = nil
	s.Camping = true
	if g.wheelAllowed() {
		t.Fatal("滾輪可跳過紮營")
	}
	g.quest = &questBattle{state: &battle.State{Stage: battle.StageEncounter}}
	if g.wheelAllowed() {
		t.Fatal("滾輪可跳過戰鬥介紹")
	}
	g.quest.state.Stage = battle.StageField
	g.quest.state.Plan.Prompt = battle.PromptExecuting
	if g.wheelAllowed() {
		t.Fatal("滾輪可推進戰鬥")
	}
	s.Prompt = nil
	s.Camping = false
	s.Mode = world.ModeTown
	s.Screen.Icons = []int{world.IconDoor + 3, world.IconUTurn + 3}
	s.Cursor = 0
	applyWorldToken(s, "R")
	if s.Cursor != 1 {
		t.Fatal("入城選單方向鍵無效")
	}
}
