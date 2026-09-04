package battle

import "testing"

// ambushMonsters 造一份最小的怪物表：兩隻，遮罩不同，好驗篩選。
func ambushMonsters() *Monsters {
	m := &Monsters{Chance: []SeasonChance{
		{Season: "WINTER", Day: []int{30, 65, 30, 45, 25, 45, 55, 35},
			Night: []int{45, 50, 60, 25, 15, 25, 30, 35}},
		{Season: "SPRING", Day: []int{25, 60, 35, 50, 30, 50, 50, 40},
			Night: []int{35, 45, 70, 30, 20, 30, 25, 40}},
	}}
	m.List = make([]Monster, MonsterCount)
	for i := range m.List {
		// 預設全部都不符合，測試再挑幾隻放行。
		m.List[i] = Monster{Index: i, Name: "NOBODY"}
	}
	return m
}

// 機率要除以 2：春天白天的道路是 30，也就是 15%。
func TestEncounterChanceIsHalved(t *testing.T) {
	m := ambushMonsters()
	if got := m.EncounterChance(1, 4, false); got != 30 {
		t.Fatalf("春天白天道路的原值是 %d，期望 30", got)
	}
	// 擲到 15 就不遇到（`rand(100) >= 30/2`）。
	if a := m.RollAmbush(4, 1, false, 3, &fixedRand{values: []int{15}}); a != nil {
		t.Fatalf("擲 15 不該遇到，得到 %+v", a)
	}
	// 擲 14 就過門檻，99 避開任務怪那一條——沒有一隻怪符合遮罩，所以仍然回 nil，
	// 但那是另一個原因。
	if a := m.RollAmbush(4, 1, false, 3, &fixedRand{values: []int{14, 99, 1}}); a != nil {
		t.Fatalf("沒有一隻怪符合遮罩時該回 nil，得到 %+v", a)
	}
}

// 遮罩要同時符合地形與季節才收。
func TestRollAmbushFiltersByMasks(t *testing.T) {
	m := ambushMonsters()
	// 第 7 隻只在地形 4（道路）出沒，季節只有春（bit 3 >> 1 = 4）。
	m.List[7] = Monster{Index: 7, Name: "RUFFIAN",
		TerrainMask: TerrainBit(4), SeasonMask: SeasonBit(1)}
	// 第 9 隻地形對、季節不對，不該被收。
	m.List[9] = Monster{Index: 9, Name: "ORC",
		TerrainMask: TerrainBit(4), SeasonMask: SeasonBit(0)}

	// 門檻過（1）、不走任務怪（99）、先抽到第 9 隻（10）、再抽到第 7 隻（8）、圖擲 2。
	rng := &fixedRand{values: []int{1, 99, 10, 8, 2}}
	a := m.RollAmbush(4, 1, false, 3, rng)
	if a == nil {
		t.Fatal("該遇到第 7 隻")
	}
	if a.Monster != 7 {
		t.Fatalf("遇到第 %d 隻，期望 7", a.Monster)
	}
	// 圖是 (rand(3) − 1) × 2。
	if a.Picture != 2 {
		t.Fatalf("圖號 %d，期望 2", a.Picture)
	}
	// 還沒擲數量與等級。
	if a.Level != 0 || a.Count != 0 {
		t.Fatalf("這一步不該有等級與隻數，得到 %+v", a)
	}
}

// 平均位階夠低時會走任務怪那一條，不擲怪物編號。
func TestRollAmbushTakesTheQuestBranch(t *testing.T) {
	m := ambushMonsters()
	// 門檻過（1）、任務怪那一擲 20 < QuestChance[2] = 21、圖擲 1。
	a := m.RollAmbush(4, 1, false, 3, &fixedRand{values: []int{1, 20, 1}})
	if a == nil || a.Monster != AmbushQuest {
		t.Fatalf("該走任務怪那一條，得到 %+v", a)
	}
	// 位階 11 以上不走那一條。
	if a := m.RollAmbush(4, 1, false, 11, &fixedRand{values: []int{1, 1}}); a != nil &&
		a.Monster == AmbushQuest {
		t.Fatal("位階 11 不該走任務怪那一條")
	}
}

// 等級是平均位階加級距修正，隻數要小於上限。
func TestRollForce(t *testing.T) {
	a := &Ambush{Monster: 7}
	// 級距擲 1 → 1/20 = 0 → 修正 −2；位階 5 → 等級 3。
	// 隻數先擲 3（>= CountCap[0][0] = 2，重擲）、再擲 1 → 0，顯示 1。地圖擲 4 → 3。
	a.RollForce(5, 1, &fixedRand{values: []int{1, 3, 1, 4}})
	if a.Level != 3 {
		t.Fatalf("等級 %d，期望 3", a.Level)
	}
	if a.Count != 1 {
		t.Fatalf("隻數 %d，期望 1", a.Count)
	}
	if a.Map != 3 {
		t.Fatalf("地圖 %d，期望 3", a.Map)
	}
	// 等級不會低於 1。
	b := &Ambush{Monster: 7}
	b.RollForce(1, 1, &fixedRand{values: []int{1, 1, 1}})
	if b.Level != 1 {
		t.Fatalf("等級 %d，期望 1", b.Level)
	}
	// 任務怪那一條不擲。
	q := &Ambush{Monster: AmbushQuest}
	q.RollForce(5, 4, &fixedRand{values: []int{1}})
	if q.Level != 0 || q.Count != 0 {
		t.Fatalf("任務怪那一條不該擲數量與等級，得到 %+v", q)
	}
}

// 人多會遇到比較多隻：同一個級距，六人的上限是一人的六倍。
func TestCountCapGrowsWithParty(t *testing.T) {
	if CountCap[0][0] != 2 || CountCap[0][5] != 12 {
		t.Fatalf("級距 0 的上限是 %v", CountCap[0])
	}
	// 等級越高遇到的隻數越少。
	if CountCap[4][5] >= CountCap[0][5] {
		t.Fatal("級距 4 的上限該比級距 0 小")
	}
}

// 隊伍平均位階：加總除以人數再加一。
func TestPartyRank(t *testing.T) {
	if got := PartyRank([]int{3, 5, 4}, 3); got != 5 {
		t.Fatalf("平均位階 %d，期望 5", got)
	}
	// 空的隊伍不會除以零。
	if got := PartyRank(nil, 0); got != 1 {
		t.Fatalf("空隊伍該回 1，得到 %d", got)
	}
}

// 騎走：有人沒坐騎就問都不問；任務怪一定跑得掉；否則看移動率。
func TestRideAway(t *testing.T) {
	if CanRideAway(NoMountPace) {
		t.Fatal("移動率 20 表示有人沒坐騎，不該問")
	}
	if !CanRideAway(12) {
		t.Fatal("有坐騎就該問")
	}
	if !RideAway(AmbushQuest, 12, &fixedRand{values: []int{1}}) {
		t.Fatal("任務怪那一條一定跑得掉")
	}
	// 移動率 12 → 門檻 60；擲 61 脫身、擲 60 沒跑掉。
	if !RideAway(7, 12, &fixedRand{values: []int{61}}) {
		t.Fatal("擲 61 該脫身")
	}
	if RideAway(7, 12, &fixedRand{values: []int{60}}) {
		t.Fatal("擲 60 不該脫身")
	}
}
