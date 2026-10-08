package battle

// CommandPanelIcon 使用原版字母標籤與圖碼；圖片索引另加 IconOffset。
type CommandPanelIcon struct {
	Label      string
	X, Y, Code int
}

func (s *State) CommandPanel() []CommandPanelIcon {
	if _, ok := s.BodyPanel(); !ok {
		return nil
	}
	i := s.Plan.BodyActor - 1
	a := s.Actors[i]
	current := s.Current()
	if current < 0 {
		return nil
	}
	if a.Side == SideMonster && (s.Round.InspectionRoll < 1 || !EnemyIntellectCheck(s.Actors[current].Intellect, a.Level, s.Round.InspectionRoll)) {
		return nil
	}
	command, style, guard := a.Command, a.Style, a.Guard
	if i == current && s.Plan.Prompt != PromptCommand {
		command, style, guard = s.Plan.Command, s.Plan.Style, s.Plan.Guard
	}
	var out []CommandPanelIcon
	add := func(label string, x, y, code int) { out = append(out, CommandPanelIcon{label, x, y, code}) }
	face := a.Facing
	if face < 0 || face > 7 {
		face = 0
	}
	face += 0x47
	block := 0x2A
	if guard != 0 {
		block = guard + 0x2F
		if block == 0x35 {
			block++
		}
	}
	switch command {
	case 0, 0xF, 0x2A:
		return nil
	case 0x21:
		forms := [10]int{0x2B, 0x2C, 0x2C, 0x2C, 0x2D, 0x2D, 0x2D, 0x2E, 0x2E, 0x2E}
		heights := [10]int{0x41, 0x2F, 0x41, 0x42, 0x2F, 0x41, 0x42, 0x2F, 0x41, 0x42}
		if style < 0 || style > 10 {
			return nil
		}
		if style == 0 {
			add("ATK", 16, 99, 0x2A)
		} else {
			add("ATK", 16, 99, forms[style-1])
			add("LOC", 16, 131, heights[style-1])
		}
		add("DEF", 48, 131, block)
		add("DIR", 48, 99, face)
	case 0x27, 0x3F:
		add("OPT", 16, 99, style)
		add("DIR", 48, 99, face)
	case 0x29:
		code := 0x2A
		if style != 0 {
			code = style + 0x36
		}
		add("ATK", 16, 99, code)
		add("DEF", 48, 131, block)
		add("DIR", 48, 99, face)
	default:
		add("OPT", 16, 99, command)
	}
	return out
}
