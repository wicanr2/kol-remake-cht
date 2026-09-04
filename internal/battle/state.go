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
	// StageSpoils 是打贏之後那張戰利品清單（RE-096）。
	StageSpoils
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
	// Skills 是四列武器的攻防原始值，用來算戰鬥等級（RE-080）。
	Skills []int
	// Foresight／Size／Health 是規劃順序與昏迷門檻要用的三個屬性（角色記錄 0x21、0x1F、0x20）。
	Foresight, Size, Health int
	// Strength／Quickness 是攻擊與防禦要用的兩個屬性；Nutrition 用來折算力量。
	Strength, Quickness, Nutrition int
	// Weapon 是手上那件武器；Attack／Parry 是那一列的攻防技能。
	Weapon        Weapon
	Attack, Parry int
	// ArmourWeight 是三件護甲的重量和；ShieldWeight／ShieldGrade 是盾牌
	// （沒有盾牌時 ShieldGrade 是 NoShield）。
	ArmourWeight, ShieldWeight, ShieldGrade int
	// Wounds 是開場就帶著的傷（角色記錄 0x1B）。
	Wounds int
	// HasWeapon 是角色記錄 +0x30 非零，也就是手上真的有武器
	// （`sub_228B7` 幾乎每一格都先問這個）。
	HasWeapon bool
	// Female 是角色記錄 0x1A，訊息的代名詞要用（RE-086）。
	Female bool
	// WeaponName 是手上那件武器的名字（`|E`）；BeltName 是 +0x190 那一格的
	// 名字（`|I`）。
	WeaponName, BeltName string
	// HasSecondItem 是角色記錄 +0x190 非零；PackCount 是 0x200 起每 0x10 一筆、
	// 十六筆之中非零的筆數。兩者的語意都還沒定（SPEC-010 末段），
	// 這裡只當指令列的條件用。
	HasSecondItem bool
	PackCount     int
	// Pack 是背包六格的名字（角色記錄 0xA0 起，一格 16 byte，RE-075）。
	// 空字串是空格。戰利品優先往這裡放。
	Pack [BackpackSlots]string
}

// State 是一場戰鬥。
type State struct {
	Site   *Site
	Stage  Stage
	Actors []Actor
	Field  *Field
	// Party 是隊伍那一側的顯示資料，索引同 Actors 的前六個位置。
	Party []Member
	// Cursor 是目前輪到誰：Round.Order 走到的那一名的索引。
	Cursor int
	// Round 是回合與出手順序。
	Round Round
	// Plan 是規劃段停在哪一層選單（SPEC-010）。
	Plan Planning
	// Exec 是執行段的出手順序（RE-084）。
	Exec Execution
	// Rand 是排規劃順序要用的亂數來源。
	Rand Rand
	// Monster 是這一場的敵人種類。原版一場戰鬥只放同一種（RE-090），
	// 所以這裡是一份不是一組。
	Monster Monster
	// Verdict 是這一場的結果；打贏時 Gold 是每個人分到的金冠、Adventure 是每個人
	// 分到的冒險點數（RE-091／RE-096）。
	Verdict Verdict
	Gold      int
	Adventure int
	// MonsterLevel 是這一場敵人的等級（`word_3B26A`）：訊息上的「LEVEL n」、
	// 技能折算與金冠的輪數都用它，而參戰者記錄那一格的等級是它 ÷ 5 ＋ 1
	// （上限 5，`0x1EE12`）——**兩個不同的尺度**（RE-097）。
	MonsterLevel int
	// SpoilItems 是 `ITEM.LST` 的物品表，打贏之後擲掉落物要用（RE-096）。
	// 沒載入就不會掉東西。
	SpoilItems []SpoilItem
	// Spoil 是戰利品畫面的游標與清單。
	Spoil SpoilScreen
	// Ground 是戰場上的十六格物品（RE-094）。丟會放進來、撿會拿走，
	// 打贏之後掉落的東西也放這裡（RE-091 的 `sub_289EE(0)`，還沒做）。
	Ground [GroundSlots]GroundItem
}

// NewQuest 依任務地點的記錄擺好一場戰鬥。
func NewQuest(site *Site, party []Member, monster Monster, field *Field) *State {
	s := &State{Site: site, Stage: StageIntro, Field: field, Party: party,
		Monster: monster}
	s.Actors = make([]Actor, MaxActors)
	for i := range s.Actors {
		s.Actors[i].Side = SideEmpty
		s.Actors[i].Member = -1
	}
	for i := range party {
		if i >= PartySlots || i >= len(site.PartyX) || i >= len(site.PartyY) {
			break
		}
		m := party[i]
		level := CombatLevel(m.Skills)
		enc, speed := Encumbrance(level, m.Weapon, m.ArmourWeight, m.ShieldWeight)
		s.Actors[i] = Actor{Kind: PartyKind, Side: SideParty,
			X: site.PartyX[i], Y: site.PartyY[i],
			Name: m.Name, Member: i,
			Level: level, Foresight: m.Foresight, Size: m.Size,
			Strength: Nourished(m.Strength, m.Nutrition), Quickness: m.Quickness,
			Attack: m.Attack, Parry: m.Parry, Weapon: m.Weapon,
			Encumbrance: enc, WeaponSpeed: speed, ShieldGrade: m.ShieldGrade,
			Threshold: Threshold(m.Size, m.Health), Damage: m.Wounds}
		s.Actors[i].SetParts()
	}
	for i := 0; i < site.Count && i < len(site.MonsterX) && i < len(site.MonsterY); i++ {
		slot := PartySlots + i
		if slot >= MaxActors {
			break
		}
		weapon := monster.Weapon.CombatWeapon()
		shieldWeight, shieldGrade := 0, NoShield
		if monster.Shield != nil {
			shieldWeight, shieldGrade = monster.Shield.Encumbrance, monster.Shield.SizeClass
		}
		enc, speed := Encumbrance(site.Level, weapon, monster.ArmourWeight, shieldWeight)
		s.Actors[slot] = Actor{Kind: site.Monster, Side: SideMonster,
			X: site.MonsterX[i], Y: site.MonsterY[i],
			Name: site.MonsterName, Level: MonsterCombatLevel(site.Level), Member: -1,
			Foresight: monster.Foresight, Size: monster.Size,
			Strength: monster.Strength, Quickness: monster.Quickness,
			// 敵人的技能要按等級縮放（`sub_26AC2:0x26AF6`）。
			Attack: monster.AttackSkill * site.Level / 5,
			Parry:  monster.ParrySkill * site.Level / 5,
			// 戰後結算讀的是記錄裡的原值，另外留一份（RE-096）。
			BaseAttack: monster.AttackSkill, BaseParry: monster.ParrySkill,
			GoldRoll: monster.Gold.Raw,
			Weapon: weapon, Encumbrance: enc, WeaponSpeed: speed,
			ShieldGrade: shieldGrade,
			Threshold: Threshold(monster.Size, monster.Health)}
		s.Actors[slot].SetParts()
	}
	s.MonsterLevel = site.Level
	return s
}

// MonsterCombatLevel 把怪物等級換成參戰者記錄那一格的等級（`0x1EE12`）：
// 除以 5 再加一，超過 5 就是 5。規劃順序與敵人 AI 的權重用的是這一個尺度，
// 不是原始等級（RE-097）。
func MonsterCombatLevel(level int) int {
	n := level/5 + 1
	if n > 5 {
		return 5
	}
	return n
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
		// 原版整輪規劃完問「Begin next round?」：U-TURN 是重新規劃這一回合、
		// YES 才進執行段、逃走圖示才問要不要逃。指令與執行段還沒實作，
		// 這一版只留「再排一輪」與「要不要逃」兩條路。
		if yes {
			// 答應之後進執行段：一手一句訊息，由 Step 推進（RE-084）。
			if s.Rand != nil {
				s.BeginExecution(s.Rand)
				s.Exec.Log = nil
				// 圖示列留著不清：原版進執行段之後沒有重畫那一列，
				// 訊息是逐格蓋在它上面的（RE-089）。
				s.Plan.Prompt = PromptExecuting
				break
			}
			s.AdvanceToParty()
			s.BeginTurn()
			break
		}
		s.Stage = StageFlee
	case StageFlee:
		if yes {
			s.Stage = StageDone
			return true
		}
		s.Stage = StageField
	}
	return false
}

// PartyDown 回報隊伍是不是全倒了。
func (s *State) PartyDown() bool {
	for i := 0; i < PartySlots && i < len(s.Actors); i++ {
		a := s.Actors[i]
		if a.Side == SideParty && !a.Down() {
			return false
		}
	}
	return true
}

// Continue 是滿版怪物圖上按下的那一個鍵：進戰場，並排出第一回合的規劃順序。
func (s *State) Continue(rng Rand) {
	if s.Stage != StageEncounter {
		return
	}
	s.Stage = StageField
	s.BeginPlanning(rng)
	s.AdvanceToParty()
	s.BeginTurn()
}

// Leader 回傳目前顯示裝備欄的那一名隊員；沒有就回 nil。
func (s *State) Leader() *Member {
	if s.Cursor < 0 || s.Cursor >= len(s.Party) {
		return nil
	}
	return &s.Party[s.Cursor]
}

// 遭遇那一張滿版怪物圖的版面（`sub_2A975` 與 `sub_2B0FF:0x2B2E4`）。

// ColourFieldEdge 是遭遇那一張外框的顏色：EGA 索引 3。
var ColourFieldEdge = [3]uint8{0, 170, 170}

// EncounterClear 是先清掉的那一塊。
var EncounterClear = struct{ X, Y, Width, Height int }{X: 14, Y: 2, Width: 292, Height: 155}

// EncounterBox 是文字框的黑底（`sub_1B779(3Ch, 0Ah, 0F8h, 26h, 0)`）；
// EncounterFrame 是外面那一圈白框（`sub_1B6DC(3Dh, 9, 0F9h, 27h, 0Fh)`）。
var (
	EncounterBox   = struct{ X, Y, Width, Height int }{X: 60, Y: 10, Width: 189, Height: 29}
	EncounterFrame = struct{ X, Y, Width, Height int }{X: 61, Y: 9, Width: 189, Height: 31}
)

// EncounterText 是文字的起點與行距（記錄 0xE9 自帶 `~V0B~H08`）。
// 這一頁用的是 8×6 的小字型，行距 6。
var EncounterText = struct{ X, Y, Step int }{X: 64, Y: 11, Step: 6}

// MonsterSpots 是怪物畫在哪裡（DGROUP+0x13B0，一筆是 byte 欄與掃描列）。
var MonsterSpots = [12][2]int{{15, 44}, {5, 44}, {25, 44}, {10, 54}, {20, 54},
	{3, 64}, {15, 64}, {27, 64}, {10, 73}, {20, 73}, {3, 82}, {27, 82}}

// MonsterArchive 是怪物圖的 archive；一隻怪物一張圖。
const MonsterArchive = "MONPICS"

// Spots 回傳這一場要畫幾隻怪物、各自畫在哪個像素座標。
func (s *State) Spots() [][2]int {
	n := len(s.Site.MonsterX)
	if s.Site.Count < n {
		n = s.Site.Count
	}
	if n > len(MonsterSpots) {
		n = len(MonsterSpots)
	}
	out := make([][2]int, n)
	for i := 0; i < n; i++ {
		out[i] = [2]int{MonsterSpots[i][0] * 8, MonsterSpots[i][1]}
	}
	return out
}

// 戰場那一頁的版面（RE-079）。

// FieldWindow 是戰場視窗的白框（`sub_1E3A8:0x1E558` 的
// `sub_1BA9D(97h, 7, 128h, 98h, 0Fh)`）；FieldView 是裡面 9×9 格的內容區。
var (
	FieldWindow = struct{ X, Y, Width, Height int }{X: 151, Y: 7, Width: 146, Height: 146}
	FieldView   = struct{ X, Y int }{X: 152, Y: 8}
)

// SheetBox 是左上角的裝備欄框（`0x1F3D5` 的 `sub_1BA9D(10h, 2, 88h, 4Ch, 0Fh)`）；
// BodyBox 是左下角的人形框（`0x1FB20` 的 `sub_1BA9D(8, 52h, 8Eh, 9Dh, 0Fh)`）。
var (
	SheetBox = struct{ X, Y, Width, Height int }{X: 16, Y: 2, Width: 121, Height: 75}
	BodyBox  = struct{ X, Y, Width, Height int }{X: 8, Y: 82, Width: 135, Height: 76}
)

// FieldArchive 是戰場磚圖；TileAtlasColumns 是 atlas 一列幾塊。
const (
	FieldArchive     = "BATTLE"
	TileAtlasColumns = 10
)

// SpriteAt 回傳站在某一格的參戰者要畫哪一塊磚；那一格沒有人就回 −1。
// 原版是逐一比對參戰者的 +0x1F／+0x20（`sub_223C0:0x225BF`…`0x22682`）：
// 隊伍前六名各畫自己的號碼圖，其餘畫共用的敵人圖，倒下的另有兩張。
func (s *State) SpriteAt(x, y int) int {
	for i := range s.Actors {
		a := s.Actors[i]
		if a.Side == SideEmpty || a.X != x || a.Y != y {
			continue
		}
		if i >= PartySlots {
			return SpriteMonster
		}
		return SpritePartyBase + i
	}
	return -1
}

// Camera 回傳戰場視窗的中心格：目前輪到的那一名。
func (s *State) Camera() (int, int) {
	if s.Cursor >= 0 && s.Cursor < len(s.Actors) {
		a := s.Actors[s.Cursor]
		if a.Side != SideEmpty {
			return a.X, a.Y
		}
	}
	return FieldSide / 2, FieldSide / 2
}
