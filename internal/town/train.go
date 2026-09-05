package town

import (
	"strconv"
	"strings"
)

// AWARD 圖示底下的訓練（RE-076）。旗標 bit7 的建築各教四門武器技能，
// 技能表在建築 payload 的 0x50 起、一筆 12 byte。訓練同時要花金幣（建築的價錢欄）
// 與冒險點數（每級 100 點），技能值一次加 10 × 級數。

// 訓練用得到的 KOLDATA 記錄。
const (
	// ArenaRecord 是位階不相稱時的 `"Go to the Arena!"`。
	ArenaRecord = 0x6B
	// TrainOfferRecord 是 `"If you want to train`。
	TrainOfferRecord = 0x3E
	// WhichSkillRecord 是 `"In which skill?"`。
	WhichSkillRecord = 0x6C
	// LearnedEnoughRecord 是四列都占滿又沒練過這一門時的 `"You've learned enough!"`。
	LearnedEnoughRecord = 0x6D
	// TooWeakRecord／TooGoodRecord 是技能值不在這門課範圍內的兩句。
	TooWeakRecord = 0x6E
	TooGoodRecord = 0x6F
	// TrainDoneRecord 是 `"Let's train!"`。
	TrainDoneRecord = 0x70
	// HowManyRecord ＋ 0xCE/0xCF ＋ 0x40 ＋ 0x41 是「要幾點技能」那一問。
	HowManyRecord   = 0x3F
	SkillWordRecord = 0x40
	PointsRecord    = 0x41
	// SideRecord 是 Offensive／Defensive。
	SideRecord = 0xCE
	// NoPointsRecord 是 `"Sorry, not enough Adventure Points!"`。
	NoPointsRecord = 0xD0
)

// TrainMaxLevel 是一次能練幾級：原版只收 1…5 的單一位數。
const TrainMaxLevel = 5

// TrainPointsPerLevel 是一級要花的冒險點數；技能值一級加 10。
const (
	TrainPointsPerLevel = 100
	TrainSkillPerLevel  = 10
)

// TrainSkill 是訓練建築教的一門技能。
type TrainSkill struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
	// Weapon 是武器識別，與角色武器列的 +8 相同。
	Weapon int `json:"weapon"`
	// Min／Max 是這門課收得起的技能值上下限。
	Min    int    `json:"min"`
	Max    int    `json:"max"`
	Record string `json:"record"`
}

// trainStage 是訓練流程停在哪一步。
type trainStage int

const (
	trainPrice trainStage = iota
	trainSkillList
	trainSide
	trainCount
)

// trainState 是一次訓練的狀態。
type trainState struct {
	stage  trainStage
	member int
	cursor int
	top    int
	skill  TrainSkill
	// row 是要寫進哪一列武器；side 是 0 攻擊、1 防禦；count 是打進去的位數。
	row   int
	side  int
	count int
	next func()
}

// Training 回報現在是不是停在訓練的畫面。
func (s *Session) Training() bool { return s.train != nil }

// StartAward 是按下 AWARD。
func (s *Session) StartAward() {
	if s.Mode != ModeInside || s.Inside == nil || s.door == nil {
		return
	}
	i := s.speakerIndex()
	if i < 0 || s.Party[i].Sheet == nil || len(s.door.Skills) == 0 {
		s.note(iconName(IconAward))
		return
	}
	sh := s.Party[i].Sheet
	if sh.Rank != sh.EarnedRank() {
		s.say(nonEmpty(s.uiLines(ArenaRecord)), nil)
		return
	}
	s.train = &trainState{stage: trainPrice, member: i}
	s.Inside.Dialog = s.priceLines(TrainOfferRecord, s.door.Price)
	s.Inside.DialogZh = nil
	s.setInsideIcons([]int{IconYes, IconNo})
}

// trainAnswer 是訓練流程裡對圖示的回應。
func (s *Session) trainAnswer(id int) {
	t := s.train
	if t == nil {
		return
	}
	switch t.stage {
	case trainPrice:
		if id != IconYes {
			s.endTrain()
			return
		}
		s.payTuition()
	case trainSkillList:
		if id == IconUTurn {
			s.endTrain()
			return
		}
		s.pickSkill()
	case trainSide:
		if id == IconUTurn {
			s.endTrain()
			return
		}
		t.side = t.cursor
		s.askCount()
	case trainCount:
		// 原版那一格是文字欄：打完數字還要按 Return 才算數。
		s.confirmCount()
	}
}

// payTuition 收學費，收完問要練哪一門。
func (s *Session) payTuition() {
	t := s.train
	if s.Party[t.member].Gold < s.door.Price {
		s.trainSay(nonEmpty(s.uiLines(NotEnoughRecord)), s.endTrain)
		return
	}
	s.Party[t.member].Gold -= s.door.Price
	s.syncParty()
	s.trainSay(nonEmpty(s.uiLines(WhichSkillRecord)), s.showSkillList)
}

// showSkillList 把對話框換成四門技能的清單。
func (s *Session) showSkillList() {
	t := s.train
	t.stage, t.cursor, t.top = trainSkillList, 0, 0
	rows := make([]string, 0, len(s.door.Skills))
	for _, sk := range s.door.Skills {
		rows = append(rows, sk.Name)
	}
	s.Inside.Dialog = nil
	s.Inside.DialogZh = nil
	s.Inside.List = &InnList{Rows: rows, Selected: 0}
	s.setInsideIcons([]int{IconYes, IconUTurn})
}

// MoveTrain 是訓練清單上的方向鍵。
func (s *Session) MoveTrain(delta int) {
	t := s.train
	if t == nil || (t.stage != trainSkillList && t.stage != trainSide) {
		return
	}
	limit := len(s.door.Skills)
	if t.stage == trainSide {
		limit = 2
	}
	t.cursor += delta
	if t.cursor < 0 {
		t.cursor = 0
	}
	if t.cursor >= limit {
		t.cursor = limit - 1
	}
	s.Inside.List.Selected = t.cursor
}

// pickSkill 選定一門技能：找出要寫哪一列，再看技能值在不在這門課的範圍內。
func (s *Session) pickSkill() {
	t := s.train
	t.skill = s.door.Skills[t.cursor]
	sh := s.Party[t.member].Sheet
	t.row = -1
	for _, w := range sh.Weapons {
		if w.Weapon == t.skill.Weapon {
			t.row = w.Row
			break
		}
	}
	if t.row < 0 {
		t.row = s.emptyWeaponRow(sh)
	}
	if t.row < 0 {
		s.Inside.List = nil
		s.trainSay(nonEmpty(s.uiLines(LearnedEnoughRecord)), s.endTrain)
		return
	}
	// 武器識別的高三位元決定要不要問攻擊還是防禦。
	if t.skill.Weapon&0xE0 == 0 {
		t.side = 0
		s.askCount()
		return
	}
	t.stage, t.cursor = trainSide, 0
	s.Inside.List = &InnList{Rows: []string{
		s.uiRecord(SideRecord), s.uiRecord(SideRecord + 1)}}
	s.setInsideIcons([]int{IconYes, IconUTurn})
}

// emptyWeaponRow 回傳第一個沒有武器的列；四列都滿了就回 -1。
func (s *Session) emptyWeaponRow(sh *Sheet) int {
	used := map[int]bool{}
	for _, w := range sh.Weapons {
		used[w.Row] = true
	}
	for row := 0; row < WeaponRows; row++ {
		if !used[row] {
			return row
		}
	}
	return -1
}

// askCount 問要練幾點：先看技能值在不在範圍內。
func (s *Session) askCount() {
	t := s.train
	sh := s.Party[t.member].Sheet
	level := s.skillLevel(sh, t.row, t.side)
	switch {
	case t.skill.Min > level:
		s.Inside.List = nil
		s.trainSay(nonEmpty(s.uiLines(TooWeakRecord)), s.endTrain)
		return
	case t.skill.Max < level:
		s.Inside.List = nil
		s.trainSay(nonEmpty(s.uiLines(TooGoodRecord)), s.endTrain)
		return
	}
	t.stage, t.count = trainCount, 0
	s.Inside.List = nil
	// 記錄 0x3F 是 `"How many]"`，結尾的 `]` 就是那個空白；
	// 0x40 是 `" Skill@"`，前面的空白要留、結尾的換行不要。
	lines := []string{strings.TrimRight(s.uiRaw(HowManyRecord)+
		s.uiRecord(SideRecord+t.side)+s.uiRaw(SkillWordRecord), " ")}
	s.trainPrompt(lines)
	// 原版問數字時不動圖示列，還是清單留下來的大拇指與 U-TURN。
	s.setInsideIcons([]int{IconYes, IconUTurn})
}

// trainPrompt 畫「要幾點技能」那一問，末尾補上原版文字欄的游標。
func (s *Session) trainPrompt(lines []string) {
	t := s.train
	tail := nonEmpty(s.uiLines(PointsRecord))
	last := ""
	if len(tail) > 0 {
		last = tail[len(tail)-1]
	}
	if t.count > 0 {
		last += strconv.Itoa(t.count)
	}
	last += TextCaret
	if len(tail) > 0 {
		tail[len(tail)-1] = last
	} else {
		tail = []string{last}
	}
	s.Inside.Dialog = append(lines, tail...)
	s.Inside.DialogZh = nil
}

// TextCaret 是文字欄游標的佔位字元。原版的游標不是字模：`sub_1B5C8` 每次等鍵
// 都呼叫 `sub_1B100(1)`，那支每 5 個 timer tick 把 `sub_1CA6E` 的參數在 0x00／0xFF
// 之間反相，EGA 版 `sub_1D01D` 直接把那個 byte 寫進
// `A000:((列 + 5 + 2 × word_2DF18) × 0x28 + 行)`——也就是筆尖那一格底下一條
// 8 px 寬、1 px 高的橫槓，約每 0.27 秒翻一次（RE-114）。這裡用一個不會出現在
// 文字裡的哨符佔一格，實際的橫槓由繪製端畫（`internal/ui`）。
//
// 訓練、MOUTH 與賭場的金額都走同一支 `sub_1B5C8`，所以三處畫的是同一條橫槓。
const TextCaret = "\x01"

// CaretRow 是橫槓在字模格裡的列號：橫槓貼在字模格的最後一條掃描線，原版兩種
// 字型都一樣（小字 6 條在第 5 條、大字 8 條在第 7 條），remake 的字模格統一
// 高 8，因此是第 7 列。MOUTH 那一格用小字，量到的橫槓在 y = 125、字模在
// 120…124，與這裡算出來的位置逐像素相同（`plan-town-mouth.json`）。
const CaretRow = 7

// TrainCount 是在「要練幾點」那一格打一個數字。原版只收 1…5 的單一位數，
// 打完之後畫面上會回顯，還要按 Return 才算數。
func (s *Session) TrainCount(n int) {
	t := s.train
	if t == nil || t.stage != trainCount || n < 1 || n > TrainMaxLevel {
		return
	}
	t.count = n
	lines := s.Inside.Dialog
	if len(lines) > 0 {
		s.trainPrompt(lines[:1])
	}
}

// confirmCount 是打完數字按下 Return。
func (s *Session) confirmCount() {
	t := s.train
	n := t.count
	if n < 1 || n > TrainMaxLevel {
		return
	}
	sh := s.Party[t.member].Sheet
	cost := n * TrainPointsPerLevel
	if sh.AdventurePoints < cost {
		s.trainSay(nonEmpty(s.uiLines(NoPointsRecord)), s.endTrain)
		return
	}
	sh.AdventurePoints -= cost
	s.learn(sh, t.row, t.side, n*TrainSkillPerLevel)
	s.syncParty()
	s.trainSay(nonEmpty(s.uiLines(TrainDoneRecord)), s.endTrain)
}

// skillLevel 是某一列某一側目前的技能值（畫面值，也就是原始值除以 10）。
func (s *Session) skillLevel(sh *Sheet, row, side int) int {
	for _, w := range sh.Weapons {
		if w.Row != row {
			continue
		}
		if side == 0 {
			return w.OffenceRaw / 10
		}
		return w.DefenceRaw / 10
	}
	return 0
}

// learn 把練到的技能寫回去；那一列本來是空的就先把武器種進去。
func (s *Session) learn(sh *Sheet, row, side, gain int) {
	t := s.train
	for i := range sh.Weapons {
		if sh.Weapons[i].Row != row {
			continue
		}
		s.addSkill(&sh.Weapons[i], side, gain)
		return
	}
	w := WeaponRow{Row: row, Name: t.skill.Name, Weapon: t.skill.Weapon}
	s.addSkill(&w, side, gain)
	sh.Weapons = append(sh.Weapons, w)
}

// addSkill 加上技能值，並把畫面用的除以 10 那一份跟著更新。
func (s *Session) addSkill(w *WeaponRow, side, gain int) {
	if side == 0 {
		w.OffenceRaw += gain * 10
	} else {
		w.DefenceRaw += gain * 10
	}
	w.Offence, w.Defence = w.OffenceRaw/10, w.DefenceRaw/10
}

// trainSay 顯示一段話，讀完之後接著做 next。
func (s *Session) trainSay(lines []string, next func()) {
	s.say(lines, nil)
	if s.train != nil {
		s.train.next = next
	}
}

// endTrain 離開訓練，回到建築內的一般畫面。
func (s *Session) endTrain() {
	s.train = nil
	if s.Inside == nil || s.door == nil {
		return
	}
	s.Inside.List = nil
	s.Inside.Dialog = s.greeting()
	s.Inside.DialogZh = nil
	s.setInsideIcons(InteriorIcons(s.door.IconFlags, len(s.Party)))
}
