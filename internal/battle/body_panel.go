package battle

import "fmt"

// BodyRect 是原版人形傷勢遮罩的半開像素矩形（SPEC-018）。
type BodyRect struct{ Left, Top, Right, Bottom int }

var bodyRects = [3][6]BodyRect{
	{{96, 86, 105, 96}, {91, 96, 110, 116}, {78, 97, 91, 119}, {110, 97, 123, 119}, {80, 120, 99, 155}, {101, 120, 121, 155}},
	{{97, 87, 104, 97}, {92, 97, 108, 118}, {81, 98, 92, 119}, {107, 98, 120, 119}, {83, 119, 99, 155}, {102, 119, 118, 155}},
	{{89, 87, 112, 98}, {89, 98, 112, 119}, {75, 98, 89, 121}, {112, 98, 126, 121}, {81, 121, 97, 155}, {104, 121, 120, 155}},
}

// BodyDamageRects 保留原版的兩次百分比截斷，不把傷勢改成一般生命條。
func BodyDamageRects(kind int, current, maximum [6]int) ([6]BodyRect, error) {
	if kind < 0 || kind >= len(bodyRects) {
		return [6]BodyRect{}, fmt.Errorf("人形種類越界：%d", kind)
	}
	out := bodyRects[kind]
	for i := range out {
		if maximum[i] <= 0 || maximum[i] > 255 || current[i] < 0 || current[i] > 255 {
			return [6]BodyRect{}, fmt.Errorf("人形部位 %d 的當前／最大值無效：%d/%d", i, current[i], maximum[i])
		}
		v := current[i]
		if v > 127 {
			v = 0
		}
		loss := 100 - v*100/maximum[i]
		out[i].Bottom = out[i].Top + (out[i].Bottom-out[i].Top)*loss/100
	}
	return out, nil
}

func BodyFatigueBoundary(stamina, fatigue int) (int, error) {
	if stamina <= 0 || stamina > 32767 || fatigue < 0 || fatigue > 65535 {
		return 0, fmt.Errorf("人形疲勞條的體力／疲勞無效：%d/%d", stamina, fatigue)
	}
	y := 84 + (100-(int(int16(2*stamina-fatigue))/2)*100/stamina)*72/100
	if y < 84 {
		y = 84
	}
	if y > 155 {
		y = 155
	}
	return y, nil
}

type BodyPanelData struct {
	Kind             int
	Current, Maximum [6]int
	Stamina, Fatigue int
}

// BodyPanel 讀取被檢視者，不改動行動者。零值 Planning 不會憑空顯示角色零。
func (s *State) BodyPanel() (BodyPanelData, bool) {
	i := s.Plan.BodyActor - 1
	if s.Stage != StageField || s.Plan.Prompt >= PromptNextRound || i < 0 || i >= len(s.Actors) || s.Actors[i].Side == SideEmpty {
		return BodyPanelData{}, false
	}
	a := &s.Actors[i]
	kind := 2
	if a.Side == SideParty {
		m := s.member(i)
		if m == nil {
			return BodyPanelData{}, false
		}
		kind = 0
		if m.Female {
			kind = 1
		}
	}
	return BodyPanelData{Kind: kind, Current: [6]int{a.Head, a.Chest, a.RightArm, a.LeftArm, a.RightLeg, a.LeftLeg}, Maximum: Parts(a.Threshold), Stamina: a.Stamina, Fatigue: a.Fatigue}, true
}

func (s *State) inspectBodyTarget() {
	s.Plan.BodyActor = 0
	c := s.Plan.Cursor
	i := s.Current()
	if i < 0 || s.Field == nil || c.Row < 0 || c.Row >= ViewRows || c.Col < 0 || c.Col >= ViewCols {
		return
	}
	a := s.Actors[i]
	visible := s.Field.Visible(a.X, a.Y, a.Flag1E != 0)
	if !visible[c.Row][c.Col] && (c.Row != ViewCentre.Row || c.Col != ViewCentre.Col) {
		return
	}
	x, y := clampCell(a.X+c.Row-ViewCentre.Row), clampCell(a.Y+c.Col-ViewCentre.Col)
	for j, target := range s.Actors {
		if target.Side != SideEmpty && target.X == x && target.Y == y {
			s.Plan.BodyActor = j + 1
			return
		}
	}
}
