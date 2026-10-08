package world

// ReturnCost 讀取原版 DS:uint16(0x12B8+index) 的已證實部分（SPEC-011 RE-195）。
// 未解出的別名回報 false，不能用初態快照冒充目前的動態記憶體。
func (c *Clock) ReturnCost(index uint16, movement int, random Rand) (int, bool) {
	var value int
	switch {
	case index == 0xffff:
		value = 0
	case index < 8:
		value = [...]int{10, 10, 10, 20, 6, 14, 30, 8}[index]
	case index < 20:
		value = monthDays[index-8]
	case index >= 840 && index < 844:
		state, ok := random.(interface{ State() uint32 })
		if !ok {
			return 0, false
		}
		value = int(byte(state.State() >> (8 * (index - 840))))
	default:
		return 0, false
	}
	if movement == 0 {
		movement = DefaultMovement
	}
	return 2 * (value + movement), true
}

// ResumeAfterBattle 只覆寫仍在等待的步行索引；換日及無待處理步行時沒有索引。
// 呼叫端必須先完成搜刮及清除實際隨機戰鬥狀態（SPEC-011 RE-185／187／195）。
func (s *Session) ResumeAfterBattle(index uint16) {
	if s.pendingStepIndex != nil {
		*s.pendingStepIndex = index
	}
	s.ResumeTimePart()
}

// ResumeAfterQuest 只接續這一場任務的到達操作。呼叫端依實際狀態指標辨識
// 來源，先關閉戰鬥；不接續別場遭遇的時間（SPEC-011 RE-214）。
func (s *Session) ResumeAfterQuest(index uint16) {
	if s.pendingArrivalIndex != nil {
		*s.pendingArrivalIndex = index
	}
	s.ResumeArrival()
}
