package main

import (
	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/town"
	"reflect"
	"testing"
)

// 從競技場的正式確認入口開兩場，檢查載圖、scatter 與底圖生命週期。
func TestArenaEntryUsesFixedMapWithoutMutatingTemplate(t *testing.T) {
	img, err := loadBattlePNG("../../assets/original-derived-png/ega/BATTLE/battle-map-24.png")
	if err != nil {
		t.Fatal(err)
	}
	template, err := battle.LoadField(img)
	if err != nil {
		t.Fatal(err)
	}
	before := append([]int(nil), template.Tiles...)
	s := &town.Session{Outdoor: &town.Outdoor{}, UIText: &town.UIText{}, Trans: &town.Translations{},
		Party: []town.Member{{Name: "TEST", Sheet: &town.Sheet{Weapons: []town.WeaponRow{{OffenceRaw: 300, DefenceRaw: 300}}}}}}
	g := &game{play: s, quest: &questBattle{}, arena: &arenaDeps{mapField: template, monsters: &battle.Monsters{},
		bank: &town.UIText{}, trans: &town.Translations{}, rand: newRand(1), field: newRand(1)}}
	for round := 0; round < 2; round++ {
		if !s.EnterArena(&town.Inn{}) {
			t.Fatal("競技場入口失敗")
		}
		s.Key("RET")
		if s.PendingFight() == nil {
			t.Fatal("正式確認沒有排入戰鬥")
		}
		g.startArenaFight()
		if g.quest.state == nil || !g.quest.state.RankFight {
			t.Fatal("未建立競技場戰鬥")
		}
		for i, v := range before {
			got := g.quest.state.Field.Tiles[i]
			if v != battle.TileGround && v != battle.TileRough && got != v {
				t.Fatalf("固定建築格 %d 被改成 %d", i, got)
			}
		}
		if !reflect.DeepEqual(before, template.Tiles) {
			t.Fatal("scatter 污染共用底圖")
		}
		g.quest.state.Field.Tiles[0] = 0
	}
}
