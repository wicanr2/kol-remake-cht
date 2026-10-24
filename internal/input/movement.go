package input

// RepeatMovement 將實體按住時間換成地圖連走節拍。選單不使用它。
func RepeatMovement(duration int) bool {
	return duration == 1 || duration >= 18 && (duration-18)%6 == 0
}

// MovementGate 在等待頁出現後要求放開移動鍵，避免按住方向穿過遭遇。
type MovementGate struct{ blocked bool }

func (g *MovementGate) Ready(walking, held bool) bool {
	if !held {
		g.blocked = false
	} else if !walking {
		g.blocked = true
	}
	return walking && !g.blocked
}
