package town

// 建築內部的圖示列與對話（RE-067）。圖示列由建築 payload 位移 0x3C／0x3D 的兩個
// 旗標 byte 決定，`sub_11A93:0x120A4` 起逐位元推進去；畫的是 `KOLMPICS[id + 3]`。

// 建築內部會用到的圖示編號。畫面上畫的是 KOLMPICS[編號 + 3]。
const (
	IconSpacebar  = 1
	IconDoor      = 2
	IconMouth     = 3
	IconListen    = 4
	IconRest      = 5
	IconCompanion = 6
	IconTrade     = 9
	IconTrain     = 10
	IconHeal      = 11
	IconAward     = 12
	IconScroll    = 13
	IconMirror    = 17
	// IconUTurn 是取消：與門口狀態的 U-TURN 同一個圖示（KOLMPICS 18）。
	IconUTurn = 15
)

// IconSprite 是圖示編號對應的 KOLMPICS 圖號。
func IconSprite(id int) int { return id + 3 }

// InteriorIcons 依兩個旗標 byte 組出建築內部的圖示列，順序與原版的推進順序相同。
// 條件逐條對應 `sub_11A93:0x120B8`…`0x121F2`。
func InteriorIcons(flags [2]int, party int) []int {
	lo, hi := flags[0], flags[1]
	icons := []int{IconDoor, IconMirror}
	if hi&0x04 != 0 || lo&0x01 != 0 {
		icons = append(icons, IconListen, IconMouth)
	}
	if lo&0x02 != 0 || lo&0x20 != 0 || lo&0x10 != 0 {
		icons = append(icons, IconTrade)
	}
	if lo&0x04 != 0 {
		icons = append(icons, IconHeal)
	}
	if lo&0x08 != 0 {
		icons = append(icons, IconRest)
		// `word_2DF2A` 是隊伍人數；滿員就不給 COMPANION。
		if party != MaxParty {
			icons = append(icons, IconCompanion)
		}
	}
	if lo&0x40 != 0 {
		icons = append(icons, IconRest)
	}
	if lo&0x80 != 0 {
		icons = append(icons, IconAward)
	}
	if hi&0x0A != 0 {
		icons = append(icons, IconTrain)
	}
	if hi&0x01 != 0 {
		icons = append(icons, IconScroll)
	}
	return icons
}

// MaxParty 是隊伍人數上限。說明書：一次最多帶六名冒險者。
const MaxParty = 6
