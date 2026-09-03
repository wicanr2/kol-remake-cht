package battle

// 一場戰鬥的狀態機。這一版只做到「進得去、看得到、逃得掉」：
// 回合結構、移動與命中結算還沒解出來（RE-079），因此不在這裡。

// Stage 是戰鬥停在哪一步。
type Stage int

const (
	// StageIntro 是任務地點的敘述與「Will you partake?」。
	StageIntro Stage = iota
	// StageEncounter 是那一張滿版怪物圖（記錄 0xE9）。
	StageEncounter
	// StageField 是戰場畫面，問「Begin next round?」。
	StageField
	// StageFlee 是回答否之後問「Do you want to flee?」。
	StageFlee
	// StageDone 是這一場結束，該回野外了。
	StageDone
)

// 戰鬥會用到的 KOLDATA 記錄。
const (
	// PartakeRecord 是任務地點的「Will you partake?」。
	PartakeRecord = 0xEA
	// QuestCreaturesRecord 是滿版怪物圖上那一句。
	QuestCreaturesRecord = 0xE9
	// EquipmentRecord 是左上角的裝備欄骨架：`W:@H:@T:@L:@S:@B:@`。
	EquipmentRecord = 0x101
	// NextRoundRecord 是「Begin next round?」。
	NextRoundRecord = 0x108
	// FleeRecord 是「Do you want to flee?」。
	FleeRecord = 0x109
	// FleeLifeRecord ＋ FleeSingleRecord／FleePluralRecord 是逃走那一句。
	FleeLifeRecord   = 0x17B
	FleeSingleRecord = 0x17C
	FleePluralRecord = 0x17D
)

// Member 是隊伍那一側要用到的欄位；battle 不認得 town 的型別，由呼叫端填。
type Member struct {
	Name string
	// Worn 是六個裝備欄的名稱，順序同記錄 0x101：武器、頭、軀幹、腿、盾、腰帶。
	Worn [6]string
}

// State 是一場戰鬥。
type State struct {
	Site   *Site
	Stage  Stage
	Actors []Actor
	Field  *Field
	// Party 是隊伍那一側的顯示資料，索引同 Actors 的前六個位置。
	Party []Member
	// Cursor 是目前輪到誰；這一版固定指向隊伍第一名。
	Cursor int
}

// NewQuest 依任務地點的記錄擺好一場戰鬥。
func NewQuest(site *Site, party []Member, field *Field) *State {
	s := &State{Site: site, Stage: StageIntro, Field: field, Party: party}
	s.Actors = make([]Actor, MaxActors)
	for i := range s.Actors {
		s.Actors[i].Side = SideEmpty
		s.Actors[i].Member = -1
	}
	for i := range party {
		if i >= PartySlots || i >= len(site.PartyX) || i >= len(site.PartyY) {
			break
		}
		s.Actors[i] = Actor{Kind: PartyKind, Side: SideParty,
			X: site.PartyX[i], Y: site.PartyY[i],
			Name: party[i].Name, Member: i}
	}
	for i := 0; i < site.Count && i < len(site.MonsterX) && i < len(site.MonsterY); i++ {
		slot := PartySlots + i
		if slot >= MaxActors {
			break
		}
		s.Actors[slot] = Actor{Kind: site.Monster, Side: SideMonster,
			X: site.MonsterX[i], Y: site.MonsterY[i],
			Name: site.MonsterName, Level: site.Level, Member: -1}
	}
	return s
}

// Answer 是圖示列上的是／否。回傳這一場是不是結束了。
func (s *State) Answer(yes bool) bool {
	switch s.Stage {
	case StageIntro:
		if !yes {
			s.Stage = StageDone
			return true
		}
		s.Stage = StageEncounter
	case StageField:
		// 回合還沒實作：答應開下一回合也只是停在原地，答否才問要不要逃。
		if !yes {
			s.Stage = StageFlee
		}
	case StageFlee:
		if yes {
			s.Stage = StageDone
			return true
		}
		s.Stage = StageField
	}
	return false
}

// Continue 是滿版怪物圖上按下的那一個鍵。
func (s *State) Continue() {
	if s.Stage == StageEncounter {
		s.Stage = StageField
	}
}

// Leader 回傳目前顯示裝備欄的那一名隊員；沒有就回 nil。
func (s *State) Leader() *Member {
	if s.Cursor < 0 || s.Cursor >= len(s.Party) {
		return nil
	}
	return &s.Party[s.Cursor]
}
