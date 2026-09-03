package town

import (
	"strings"
	"testing"
)

// brettleSkills 是原版城 1 布瑞特堡壘教的四門，欄位抄自建築 payload。
func brettleSkills() []TrainSkill {
	return []TrainSkill{
		{Index: 0, Name: "Longsword", Weapon: 92, Min: 0, Max: 25},
		{Index: 1, Name: "Broadsword", Weapon: 93, Min: 0, Max: 26},
		{Index: 2, Name: "Short Spear", Weapon: 156, Min: 0, Max: 25},
		{Index: 3, Name: "Battle Axe", Weapon: 124, Min: 0, Max: 26},
	}
}

func trainText() *UIText {
	t := adjustText()
	t.Banks[0].Records = append(t.Banks[0].Records,
		uiRecord{Index: ArenaRecord, Text: `"Go to the Arena!"`},
		uiRecord{Index: TrainOfferRecord, Text: `"If you want to train@`},
		uiRecord{Index: WhichSkillRecord, Text: `"In which skill?"@`},
		uiRecord{Index: LearnedEnoughRecord, Text: `"You've learned enough!"`},
		uiRecord{Index: TooWeakRecord, Text: `"You're not that good!"`},
		uiRecord{Index: TooGoodRecord, Text: `"You're too good!"`},
		uiRecord{Index: TrainDoneRecord, Text: `"Let's train!"`},
		uiRecord{Index: HowManyRecord, Text: "How many]"},
		uiRecord{Index: SkillWordRecord, Text: " Skill@"},
		uiRecord{Index: PointsRecord, Text: `~H0DPoints did you want?"`},
		uiRecord{Index: SideRecord, Text: "Offensive"},
		uiRecord{Index: SideRecord + 1, Text: "Defensive"},
		uiRecord{Index: NoPointsRecord, Text: `"Sorry, not enough@~H0DAdventure Points!"`})
	return t
}

// theodoricSheet 是原版存檔第 0 格：只有一列巨劍、攻防各 100。
func theodoricSheet() *Sheet {
	s := theodoric()
	s.Weapons = []WeaponRow{{Row: 0, Name: "Greatsword", Weapon: 87,
		Offence: 10, Defence: 10, OffenceRaw: 100, DefenceRaw: 100}}
	s.AdventurePoints = 674
	s.Rank = 1
	return s
}

func trainSession(t *testing.T, gold int) *Session {
	t.Helper()
	s := restSession(t, [2]int{0x81, 0x00}, 240, 0,
		Member{Name: "Theodoric", Gold: gold, Speaker: true, Sheet: theodoricSheet()})
	s.Buildings.Doors[0].Skills = brettleSkills()
	s.door = &s.Buildings.Doors[0]
	s.UIText = trainText()
	return s
}

// 位階與應得的不同就叫你去競技場。八個攻防值 200 / 200 ＝ 1，記錄裡也是 1，所以相同。
func TestAwardRankMismatch(t *testing.T) {
	s := trainSession(t, 3000)
	s.Party[0].Sheet.Rank = 2
	s.StartAward()
	if s.Training() {
		t.Fatal("位階不符不該進訓練")
	}
	if s.Inside.Dialog[0] != `"Go to the Arena!"` {
		t.Fatalf("畫的是 %q", s.Inside.Dialog)
	}
}

// 學費先收，接著問要練哪一門。
func TestAwardChargesTuition(t *testing.T) {
	s := trainSession(t, 3000)
	s.StartAward()
	if s.Inside.Dialog[1] != "it will cost you 240" {
		t.Fatalf("學費那一行是 %q", s.Inside.Dialog[1])
	}
	yes(s)
	if s.Party[0].Gold != 3000-240 {
		t.Fatalf("金幣是 %d", s.Party[0].Gold)
	}
	if s.Inside.Dialog[0] != `"In which skill?"` {
		t.Fatalf("畫的是 %q", s.Inside.Dialog)
	}
	s.Continue()
	if s.Inside.List == nil || len(s.Inside.List.Rows) != 4 ||
		s.Inside.List.Rows[0] != "Longsword" {
		t.Fatalf("技能清單是 %+v", s.Inside.List)
	}
}

// 練一門沒學過的：占掉空的一列，扣冒險點數，技能值加 10 × 級數。
func TestAwardLearnsNewSkill(t *testing.T) {
	s := trainSession(t, 3000)
	s.StartAward()
	yes(s)      // 付學費
	s.Continue() // 讀完「要練哪一門」
	yes(s)      // 選長劍（武器 92，高三位元非 0 → 要問攻擊還是防禦）
	if s.Inside.List == nil || s.Inside.List.Rows[0] != "Offensive" {
		t.Fatalf("應該問攻擊還是防禦：%+v", s.Inside.List)
	}
	yes(s) // 攻擊
	if s.Inside.Dialog[0] != "How many Offensive Skill" {
		t.Fatalf("問句是 %q", s.Inside.Dialog)
	}
	s.TrainCount(3)
	if got := s.Inside.Dialog[1]; !strings.HasSuffix(got, "3"+TextCaret) {
		t.Fatalf("打了 3 之後畫的是 %q", got)
	}
	yes(s) // 原版的數字欄要按 Return 才算數
	sh := s.Party[0].Sheet
	if sh.AdventurePoints != 674-300 {
		t.Fatalf("冒險點數是 %d", sh.AdventurePoints)
	}
	if len(sh.Weapons) != 2 || sh.Weapons[1].Name != "Longsword" ||
		sh.Weapons[1].OffenceRaw != 300 {
		t.Fatalf("學到的是 %+v", sh.Weapons)
	}
	if s.Inside.Dialog[0] != `"Let's train!"` {
		t.Fatalf("最後一句是 %q", s.Inside.Dialog)
	}
}

// 冒險點數不夠就講記錄 0xD0，技能不變。
func TestAwardNotEnoughPoints(t *testing.T) {
	s := trainSession(t, 3000)
	s.Party[0].Sheet.AdventurePoints = 50
	s.StartAward()
	yes(s)
	s.Continue()
	yes(s)
	yes(s)
	s.TrainCount(1)
	yes(s)
	sh := s.Party[0].Sheet
	if sh.AdventurePoints != 50 || len(sh.Weapons) != 1 {
		t.Fatalf("點數不夠卻改了狀態：%d／%d", sh.AdventurePoints, len(sh.Weapons))
	}
	if s.Inside.Dialog[0] != `"Sorry, not enough` {
		t.Fatalf("畫的是 %q", s.Inside.Dialog)
	}
}

// 技能值超過這門課的上限就講「你太強了」。
func TestAwardTooGood(t *testing.T) {
	s := trainSession(t, 3000)
	// 巨劍那一列改成長劍，攻擊值 300 → 畫面值 30，超過長劍課的上限 25。
	s.Party[0].Sheet.Weapons[0] = WeaponRow{Row: 0, Name: "Longsword", Weapon: 92,
		OffenceRaw: 300, DefenceRaw: 100, Offence: 30, Defence: 10}
	s.Party[0].Sheet.Rank = s.Party[0].Sheet.EarnedRank()
	s.StartAward()
	yes(s)
	s.Continue()
	yes(s)
	yes(s)
	if s.Inside.Dialog[0] != `"You're too good!"` {
		t.Fatalf("畫的是 %q", s.Inside.Dialog)
	}
}

// 四列都占滿又沒練過這一門就講「你學夠了」。
func TestAwardLearnedEnough(t *testing.T) {
	s := trainSession(t, 3000)
	sh := s.Party[0].Sheet
	sh.Weapons = nil
	for row := 0; row < WeaponRows; row++ {
		sh.Weapons = append(sh.Weapons, WeaponRow{Row: row, Weapon: 200 + row})
	}
	sh.Rank = sh.EarnedRank()
	s.StartAward()
	yes(s)
	s.Continue()
	yes(s)
	if s.Inside.Dialog[0] != `"You've learned enough!"` {
		t.Fatalf("畫的是 %q", s.Inside.Dialog)
	}
}
