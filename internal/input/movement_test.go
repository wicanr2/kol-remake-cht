package input

import "testing"

func TestMovementRepeatStopsAtModalUntilRelease(t *testing.T) {
	var gate MovementGate
	steps := 0
	for tick := 1; tick <= 30; tick++ {
		if gate.Ready(true, true) && RepeatMovement(tick) {
			steps++
		}
	}
	if steps != 4 {
		t.Fatalf("連走 %d 次", steps)
	}
	if gate.Ready(false, true) || gate.Ready(true, true) {
		t.Fatal("按住方向穿過等待頁")
	}
	gate.Ready(true, false)
	if !gate.Ready(true, true) || !RepeatMovement(1) {
		t.Fatal("放開再按不能移動")
	}
}
