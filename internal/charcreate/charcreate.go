// Package charcreate 實作 SPEC-005 的角色建立狀態機。
//
// 狀態轉移、取消目的地、名字欄容量與滿格自動前進、圖像 selector 的 80 張循環，
// 都出自 docs/verification/character-ui/ 的正常玩家路徑收據；屬性擲骰公式出自
// RE-035／RE-036。亂數可注入，畫面才能重播。
package charcreate

import (
	"github.com/wicanr2/kol-remake-cht/internal/chardata"
)

// Stage 是建立流程的層。
type Stage int

const (
	StageName Stage = iota
	StageRace
	StageSecondary
	StageProfession
	StageStats
	StagePicture
	StageDone      // 已接受圖像，等待寫入
	StageAbandoned // 於名字層取消，流程結束
)

// NameCapacity 是名字欄的可見字元數；輸入第 16 個字元會直接前進到種族層。
const NameCapacity = 16

// PortraitCount 是角色圖像 selector 的循環長度。
const PortraitCount = 80

// Rand 是可注入的亂數來源；n 為骰面數，回傳 1..n。
type Rand interface {
	Roll(n int) int
}

// State 是一次角色建立的完整狀態。
type State struct {
	data *chardata.Data
	rng  Rand

	Stage      Stage
	Name       string
	Race       int
	Secondary  int
	Profession int // 目前種族／secondary 群組內的索引
	Cursor     int // 目前這一層清單的反白位置
	Portrait   int

	Primary [7]int // 內部值，未套顯示量尺
	Derived [3]int
	Rolled  bool

	// summaryVisible 反映原版行為：剛進入 New Character 頁時，左頁摘要就以
	// 預設選項（種族 0／secondary 0／職業 0）顯示；由種族層取消回到名字層之後，
	// 整份摘要會清空。兩種畫面都有原版收據。
	summaryVisible bool
}

func New(data *chardata.Data, rng Rand) *State {
	return &State{data: data, rng: rng, Stage: StageName, summaryVisible: true}
}

// Professions 是目前種族／secondary 對應的職業清單。
func (s *State) Professions() []chardata.Profession {
	return s.data.Professions(s.Race, s.Secondary)
}

// ListLength 是目前這一層右頁清單的項目數；名字層沒有清單。
func (s *State) ListLength() int {
	switch s.Stage {
	case StageRace:
		return len(s.data.Races)
	case StageSecondary:
		return len(s.data.Secondary)
	case StageProfession:
		return len(s.Professions())
	case StageStats:
		return 2
	}
	return 0
}

// ListItems 是目前這一層右頁清單的文字。
func (s *State) ListItems() []string {
	switch s.Stage {
	case StageRace:
		out := make([]string, 0, len(s.data.Races))
		for _, r := range s.data.Races {
			out = append(out, r.Name)
		}
		return out
	case StageSecondary:
		out := make([]string, 0, len(s.data.Secondary))
		for _, r := range s.data.Secondary {
			out = append(out, r.Name)
		}
		return out
	case StageProfession:
		ps := s.Professions()
		out := make([]string, 0, len(ps))
		for _, p := range ps {
			out = append(out, p.Name)
		}
		return out
	case StageStats:
		return []string{"REROLL STATS", "ACCEPT STATS"}
	}
	return nil
}

// TypeRune 在名字層輸入一個字元。填滿 16 格會直接前進到種族層，這是原版行為。
func (s *State) TypeRune(r rune) {
	if s.Stage != StageName || len([]rune(s.Name)) >= NameCapacity {
		return
	}
	s.Name += string(r)
	if len([]rune(s.Name)) == NameCapacity {
		s.enterRace()
	}
}

// Backspace 刪除名字最後一個字元。
func (s *State) Backspace() {
	if s.Stage != StageName || s.Name == "" {
		return
	}
	r := []rune(s.Name)
	s.Name = string(r[:len(r)-1])
}

// MoveCursor 在清單內上下移動；不繞圈，與原版反白條一致。
func (s *State) MoveCursor(delta int) {
	n := s.ListLength()
	if n == 0 {
		return
	}
	s.Cursor += delta
	if s.Cursor < 0 {
		s.Cursor = 0
	}
	if s.Cursor >= n {
		s.Cursor = n - 1
	}
}

// MovePortrait 在圖像 selector 上移動，於 0..79 循環。
func (s *State) MovePortrait(delta int) {
	if s.Stage != StagePicture {
		return
	}
	s.Portrait = ((s.Portrait+delta)%PortraitCount + PortraitCount) % PortraitCount
}

func (s *State) enterRace() {
	s.Stage = StageRace
	s.Cursor = 0
}

// Accept 是 Return：依目前層前進。空名字不前進，與原版一致。
func (s *State) Accept() {
	switch s.Stage {
	case StageName:
		if s.Name == "" {
			return
		}
		s.enterRace()
	case StageRace:
		s.Race = s.data.Races[s.Cursor].Value
		if s.data.HasSecondary(s.Race) {
			s.Stage = StageSecondary
		} else {
			s.Secondary = 0
			s.Stage = StageProfession
		}
		s.Cursor = 0
	case StageSecondary:
		s.Secondary = s.data.Secondary[s.Cursor].Value
		s.Stage = StageProfession
		s.Cursor = 0
	case StageProfession:
		s.Profession = s.Cursor
		s.Roll()
		s.Stage = StageStats
		s.Cursor = 0
	case StageStats:
		if s.Cursor == 0 {
			s.Roll()
			return
		}
		s.Stage = StagePicture
		s.Cursor = 0
	case StagePicture:
		s.Stage = StageDone
	}
}

// Cancel 是 ESC：回到上一層。回到名字層時名字與摘要一併清空，與原版一致。
func (s *State) Cancel() {
	switch s.Stage {
	case StageName:
		s.Stage = StageAbandoned
	case StageRace:
		s.reset()
	case StageSecondary:
		s.Stage = StageRace
		s.Cursor = s.Race
	case StageProfession:
		if s.data.HasSecondary(s.Race) {
			s.Stage = StageSecondary
			s.Cursor = s.Secondary
		} else {
			s.Stage = StageRace
			s.Cursor = s.Race
		}
	case StageStats:
		s.Stage = StageProfession
		s.Cursor = s.Profession
	case StagePicture:
		s.Stage = StageStats
		s.Cursor = 0
	}
}

func (s *State) reset() {
	s.Stage = StageName
	s.Name = ""
	s.Race, s.Secondary, s.Profession, s.Cursor, s.Portrait = 0, 0, 0, 0, 0
	s.Primary = [7]int{}
	s.Derived = [3]int{}
	s.Rolled = false
	s.summaryVisible = false
}

// ProfessionRecordIndex 是目前選定職業在 39 筆記錄裡的索引；寫入槽位時用的是這個值，
// 不是群組內的序號。
func (s *State) ProfessionRecordIndex() int {
	ps := s.Professions()
	if s.Profession >= len(ps) {
		return 0
	}
	return ps[s.Profession].RecordIndex
}

// Roll 依職業記錄重擲全部主要屬性並重算衍生值。
func (s *State) Roll() {
	ps := s.Professions()
	if s.Profession >= len(ps) {
		return
	}
	p := ps[s.Profession]
	for i := 0; i < 7; i++ {
		code := p.DiceCodes[i]
		sides := s.data.DiceSides[code>>4]
		count := code & 0x0F
		total := p.InternalBases[i]
		for r := 0; r < count; r++ {
			total += s.rng.Roll(sides)
		}
		s.Primary[i] = total
	}
	// BALANCE、ENDURANCE、BODY PTS. 依序為力量與智力、力量與健康、體型與健康的
	// 內部無號整數平均，向下取整。
	s.Derived[0] = (s.Primary[0] + s.Primary[6]) / 2
	s.Derived[1] = (s.Primary[0] + s.Primary[3]) / 2
	s.Derived[2] = (s.Primary[2] + s.Primary[3]) / 2
	s.Rolled = true
}

// DisplayValue 把內部值換成畫面值：floor(內部值 × 100 / 150)。
func (s *State) DisplayValue(internal int) int {
	return internal * s.data.DisplayScale.Numerator / s.data.DisplayScale.Denominator
}

// RaceName 回傳要顯示的種族名稱。尚未選定時，若摘要可見則顯示目前預設值。
func (s *State) RaceName() string {
	if s.Stage <= StageRace && !s.summaryVisible {
		return ""
	}
	return s.data.Races[s.Race].Name
}

// SecondaryName 回傳已選 secondary 名稱；尚未選定或該種族沒有 secondary 時回空字串。
func (s *State) SecondaryName() string {
	if !s.data.HasSecondary(s.Race) {
		return ""
	}
	if s.Stage <= StageSecondary && !s.summaryVisible {
		return ""
	}
	return s.data.Secondary[s.Secondary].Name
}

// ProfessionName 回傳已選職業名稱。
func (s *State) ProfessionName() string {
	if s.Stage == StageProfession {
		return ""
	}
	if s.Stage < StageProfession && !s.summaryVisible {
		return ""
	}
	ps := s.Professions()
	if s.Profession >= len(ps) {
		return ""
	}
	return ps[s.Profession].Name
}

// PreviewProfessionName 是職業層反白中的名稱；原版會即時顯示在左頁。
func (s *State) PreviewProfessionName() string {
	if s.Stage != StageProfession {
		return ""
	}
	ps := s.Professions()
	if s.Cursor >= len(ps) {
		return ""
	}
	return ps[s.Cursor].Name
}
