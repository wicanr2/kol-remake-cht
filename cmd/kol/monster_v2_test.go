package main

import (
	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"testing"
)

func TestMonsterV2SpeakerUsesActorEquipment(t *testing.T) {
	s := &battle.State{Monster: battle.Monster{Weapon: &battle.Item{Name: "Mother weapon"}, EquipmentVariants: make([]battle.MonsterEquipment, 4)}, Actors: make([]battle.Actor, 7)}
	s.Actors[6] = battle.Actor{Name: "Cliff Troll", Side: battle.SideMonster, Member: -1, HandItem: &battle.Item{Name: "Troll Maul"}}
	g := &game{quest: &questBattle{state: s}}
	if got := g.battleSpeaker(6).Weapon; got != "Troll Maul" {
		t.Fatalf("讀到母表武器：%q", got)
	}
	s.Actors[6].HandItem = nil
	if got := g.battleSpeaker(6).Weapon; got != "" {
		t.Fatalf("空手仍回退母表：%q", got)
	}
}
