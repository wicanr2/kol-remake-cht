package ui

import (
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/town"
)

// 戶外與野外的隊員格一律是藍框，但不能順手改掉角色狀態裡的發言者。
func TestIdlePartyKeepsSpeakerInState(t *testing.T) {
	party := []town.Member{{Name: "A", Speaker: true}, {Name: "B"}}
	out := idleParty(party)
	for i, m := range out {
		if m.Speaker {
			t.Fatalf("第 %d 位仍畫成發言者", i)
		}
	}
	if !party[0].Speaker {
		t.Fatal("idleParty 改到了原本的隊伍")
	}
}
