//go:build sceneverify

package battle

// SceneVerification 只存在於以 sceneverify 標籤建置的驗收程式。
// 它不改 typed data、武器或保存格式；用途與限制見 SPEC-040。
var SceneVerification = struct {
	LockHP           bool
	DamageMultiplier int
}{DamageMultiplier: 1}

func sceneVerificationDamage(a *Actor, damage int) int {
	if a.Side == SideParty && SceneVerification.DamageMultiplier > 1 {
		return damage * SceneVerification.DamageMultiplier
	}
	return damage
}

func sceneVerificationLocked(a *Actor) bool {
	return SceneVerification.LockHP && a.Side == SideParty
}
