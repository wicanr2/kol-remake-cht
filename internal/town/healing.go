package town

// 野外遭遇裡的「治療」那一條（怪物編號 0xFF，`0x2A719`…`0x2A794`，RE-112；
// 敘述的來源由 RE-128 解出）。低位階的隊伍才擲得到：路上遇見誦經的僧侶或一個
// 披著斗篷替傷者療傷的人，答應之後全隊的傷勢與飢餓一起清零。

// HealParty 是答應治療之後做的事：每一格有人的角色，傷勢（角色記錄 `+0x1B`）與
// 飢餓（`+0x9F`）都設成零。回傳被治好的人數。
//
// 原版是逐格掃六格、看那一格的第一個 byte（名字的第一個字元）非零才動
// （`0x2A761`…`0x2A78A`），所以空位不算。remake 的隊伍本來就沒有空位。
func HealParty(party []Member) int {
	healed := 0
	for i := range party {
		if party[i].Name == "" {
			continue
		}
		party[i].Wounds = 0
		if sh := party[i].Sheet; sh != nil {
			copied := *sh
			copied.Nutrition = 0
			party[i].Sheet = &copied
		}
		healed++
	}
	return healed
}
