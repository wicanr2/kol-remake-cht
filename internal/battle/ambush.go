package battle

// 野外的隨機遭遇（RE-078／RE-097）。原版在**時段變換之後**擲一次
// （`sub_29987:0x29A3E`），一天最多四次；走進某一格不會擲。

// AmbushQuest 是「任務怪物那一條」的怪物編號（原版用 `0xFF`）。走到這一條的時候
// 原版不擲數量與等級，改印記錄 0xE6 問一句，答應之後把六名角色的傷清掉。
const AmbushQuest = 0xFF

// AmbushTries 是挑怪物的重試次數（`0x2A5A9` 的 `0x1F`）。試滿都沒有一隻同時符合
// 地形與季節，就當作沒有遭遇。
const AmbushTries = 0x1F

// MonthSeason 是月份 0…11 對應的季節（DGROUP `0x1332`）：0 冬、1 春、2 夏、3 秋。
var MonthSeason = [12]int{0, 0, 1, 1, 1, 2, 2, 2, 3, 3, 3, 0}

// NightWeather 是「夜裡」那一組天候（RE-078）：機率表要改讀 `0x20` 那一半。
const NightWeather = 5

// QuestChance 是平均位階 1…10 走任務怪那一條的機率（DGROUP `0x133E`）。
// 位階 11 以上不走那一條。
var QuestChance = [10]int{21, 21, 21, 16, 16, 16, 11, 11, 11, 11}

// LevelAdjust 是級距 0…4 對平均位階的修正（DGROUP `0x1348`，五個 word）。
var LevelAdjust = [5]int{-2, -1, 0, 1, 2}

// CountCap 是級距 × 隊伍人數（1…6）的數量上限（DGROUP `0x1352`，30 byte）。
// 擲出來的數量要**小於**表上的值才收，所以等級越高遇到的隻數越少、人越多越多。
var CountCap = [5][6]int{
	{2, 4, 6, 8, 10, 12},
	{2, 3, 4, 5, 6, 7},
	{1, 2, 3, 4, 5, 6},
	{1, 1, 2, 2, 3, 3},
	{1, 1, 2, 2, 3, 3},
}

// LevelBandSides／LevelBandDivisor 是級距的擲法（`0x2A79F`）：`rand(0x63) ÷ 0x14`。
const (
	LevelBandSides   = 0x63
	LevelBandDivisor = 0x14
)

// AmbushMaps 是戰場地圖的張數（`0x2A819` 的 `rand(5) − 1`）。任務地點那一條是
// `rand(3) − 1`（`0x1E3E7`），兩者不同。
const AmbushMaps = 5

// 遭遇會用到的 KOLDATA 記錄。
const (
	RideAwayRecord  = 0xE3 // "Attempt to ride away?"
	RodeAwayRecord  = 0xE4 // 脫身
	StoodGroundRec  = 0xE5 // 沒跑掉
	QuestAskRecord  = 0xE6 // 任務怪那一條的問句
	AmbushedRecord  = 0xE7 // "YOU HAVE BEEN AMBUSHED@BY]"
	AmbushLevelRec  = 0xE8 // ", LEVEL]"
	NoMountPace     = 20   // `word_2FA2C` 等於 20 表示有人沒坐騎，跑不掉
	RideAwayPerPace = 5    // `rand(100) > 5 × 移動率` 才脫身
)

// Ambush 是一次遭遇的擲點結果。
type Ambush struct {
	// Monster 是怪物編號 0…31，或 AmbushQuest。
	Monster int
	// Picture 是那一隻要畫哪一張（`(rand(3) − 1) × 2`，所以是 0、2 或 4）。
	Picture int
	// Level／Count 是敵人的等級與隻數；任務怪那一條不擲，兩個都是 0。
	Level, Count int
	// Map 是戰場地圖編號 0…4。
	Map int
}

// TerrainBit／SeasonBit 是遮罩的位元（DGROUP `0x1312`／`0x1316`）：
// 地形 0 是 bit 7、季節 0（冬）是 bit 3。
func TerrainBit(terrain int) int { return 0x80 >> uint(terrain&7) }

// SeasonBit 同上，季節只有四個。
func SeasonBit(season int) int { return 0x08 >> uint(season&3) }

// PartyRank 是隊伍的平均位階（`sub_2A337`）：在籍角色的實得位階加總除以人數再加一。
// 空的格子不算，除數是**隊伍人數**而不是加總過幾個。
func PartyRank(ranks []int, party int) int {
	if party <= 0 {
		return 1
	}
	total := 0
	for _, r := range ranks {
		total += r
	}
	return total/party + 1
}

// RollAmbush 擲出「有沒有遇到、遇到誰」（`0x2A4EE`…`0x2A678`）。
//
// 沒遇到就回 nil。回傳的 Ambush 只填了 Monster 與 Picture——數量與等級要等玩家決定
// 不騎走之後才擲（原版就是這個順序，見 RollForce）。
//
// `terrain` 是地形 0…7、`season` 是季節 0…3（0 是冬）、`night` 是天候第 5 組、
// `rank` 是隊伍平均位階、`party` 是人數。
func (m *Monsters) RollAmbush(terrain, season int, night bool, rank int, rng Rand) *Ambush {
	if rng.Roll(100) >= m.EncounterChance(season, terrain, night)/2 {
		return nil
	}
	monster := -1
	if rank < 11 && rank >= 1 && rng.Roll(100) < QuestChance[rank-1] {
		monster = AmbushQuest
	} else {
		for try := AmbushTries; try >= 0; try-- {
			n := rng.Roll(MonsterCount) - 1
			mon := m.At(n)
			if mon.TerrainMask&TerrainBit(terrain) == 0 {
				continue
			}
			if mon.SeasonMask&SeasonBit(season) == 0 {
				continue
			}
			monster = n
			break
		}
		if monster < 0 {
			return nil
		}
	}
	return &Ambush{Monster: monster, Picture: (rng.Roll(3) - 1) * 2}
}

// RollForce 擲出敵人的等級、隻數與戰場地圖（`0x2A79F`…`0x2A828`）。
//
// 級距先擲一次，等級是平均位階加上級距的修正（不到 1 就當 1）；隻數重擲到**小於**
// 上限為止，畫面上顯示的是它加一。任務怪那一條不走這裡。
func (a *Ambush) RollForce(rank, party int, rng Rand) {
	if a == nil || a.Monster == AmbushQuest {
		return
	}
	band := rng.Roll(LevelBandSides) / LevelBandDivisor
	if band > 4 {
		band = 4
	}
	a.Level = rank + LevelAdjust[band]
	if a.Level < 1 {
		a.Level = 1
	}
	cap := 1
	if party >= 1 && party <= 6 {
		cap = CountCap[band][party-1]
	}
	n := 0
	for {
		n = rng.Roll(12) - 1
		if n < cap {
			break
		}
	}
	a.Count = n + 1
	a.Map = rng.Roll(AmbushMaps) - 1
}

// CanRideAway 是「問不問要不要騎走」（`0x2A67B`）：隊伍最慢的移動率等於
// NoMountPace 就是有人沒坐騎，直接開打。
func CanRideAway(pace int) bool { return pace != NoMountPace }

// RideAway 是騎走的成功判定（`0x2A6B5`）：任務怪那一條一定跑得掉，否則
// `rand(100) > 5 × 隊伍移動率`——移動率越小（坐騎越快）越容易脫身。
func RideAway(monster, pace int, rng Rand) bool {
	if monster == AmbushQuest {
		return true
	}
	return rng.Roll(100) > RideAwayPerPace*pace
}
