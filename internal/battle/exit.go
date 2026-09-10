package battle

// PostBattleWounds 保留 SPEC-027 原版共同出口的 16-bit 無號除法與 byte 寫回。
// 平常是舊傷勢加本場新增傷勢的一半；邊界行為亦由 dosgolem 驗證。
func PostBattleWounds(original, current int) int {
	old := uint8(original)
	delta := uint16(uint8(current)) - uint16(old)
	return int(old + uint8(delta/2))
}
