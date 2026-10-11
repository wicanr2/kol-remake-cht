//go:build sceneverify

package battle

import (
	"reflect"
	"testing"
)

func TestSceneVerificationPartyOnly(t *testing.T) {
	previous := SceneVerification
	t.Cleanup(func() { SceneVerification = previous })
	SceneVerification.LockHP = true
	SceneVerification.DamageMultiplier = 20
	party := Actor{Side: SideParty, Threshold: 10}
	party.SetParts()
	before := party
	if down, limb := party.Take(100, 0); down || limb || !reflect.DeepEqual(party, before) {
		t.Fatal("鎖血必須在致命傷及肢體失能前阻止變更")
	}
	foe := Actor{Side: SideMonster, Threshold: 10}
	foe.SetParts()
	if down, _ := foe.Take(100, 0); !down {
		t.Fatal("敵人不能取得鎖血")
	}
	if sceneVerificationDamage(&party, 3) != 60 || sceneVerificationDamage(&foe, 3) != 3 {
		t.Fatal("倍率只適用隊伍造成的傷害")
	}
	attacker := Actor{Side: SideParty, Strength: 70, Command: CmdWeapon, Weapon: Weapon{Base: 3}}
	target := Actor{Side: SideMonster, Threshold: 10}
	target.SetParts()
	baseline := target
	SceneVerification.DamageMultiplier = 1
	base := finishDamage(&attacker, &baseline, 0, 0, sceneVerificationRoll{})
	SceneVerification.DamageMultiplier = 20
	if out := finishDamage(&attacker, &target, 0, 0, sceneVerificationRoll{}); base.Damage <= 0 ||
		out.Damage != base.Damage*20 || target.Hurt-out.Damage != baseline.Hurt-base.Damage {
		t.Fatalf("加倍傷害不得改護甲吸收量：damage=%d hurt=%d", out.Damage, target.Hurt)
	}
	SceneVerification.LockHP = false
	SceneVerification.DamageMultiplier = 1
	if down, _ := party.Take(100, 0); !down || sceneVerificationDamage(&party, 3) != 3 {
		t.Fatal("關閉驗收輔助後必須恢復正常傷害")
	}
}

type sceneVerificationRoll struct{}

func (sceneVerificationRoll) Roll(n int) int { return 1 }
