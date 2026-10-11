//go:build !sceneverify

package battle

func sceneVerificationDamage(a *Actor, damage int) int { return damage }
func sceneVerificationLocked(a *Actor) bool            { return false }
