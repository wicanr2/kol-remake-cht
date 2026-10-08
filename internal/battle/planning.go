package battle

// 規劃段的選單（SPEC-010）。輪到一名隊員時，`sub_228B7` 組出主指令列，玩家選中
// 之後依圖號走到選目標格與子選單，最後把指令碼、攻擊選項與防禦選項寫回那一名。
// 整輪決定完才問記錄 0x108「Begin next round?」。
//
// 沒接的兩塊：背包（0x0D）要兩層清單，體型那一列（`sub_24A02`）要判斷面前那一格
// 有沒有符合遮罩 0x18 的人，而遮罩用到語意未定的 +0x05。兩者按下去都記進
// Unimplemented，不默默吞掉。

// Prompt 是規劃段停在哪一層。
type Prompt int

const (
	// PromptCommand 是主指令列。
	PromptCommand Prompt = iota
	// PromptTarget 是選目標格。
	PromptTarget
	// PromptWeaponForm／PromptUnarmedForm／PromptMoveForm／PromptGuardForm 是四列子選單。
	PromptWeaponForm
	PromptUnarmedForm
	PromptMoveForm
	PromptGuardForm
	// PromptSpell 是「Which Spell?」那張法術清單；PromptAffect 是接著問的
	// 「Select Affect」（Positive／Negative）。兩張都是**文字清單**，不是圖示列
	// （`0x1F3C1`、`0x1F572`，RE-100）。
	PromptSpell
	PromptAffect
	// PromptNextRound 是整輪決定完的「Begin next round?」。
	PromptNextRound
	// PromptFlee 是執行段走完、還沒分勝負又符合條件時問的
	//「Do you want to flee?」（RE-103）。
	PromptFlee
	// PromptExecuting 是執行段：一手一句訊息，按鍵推進。
	PromptExecuting
	// PromptPickUp 是撿那一手在執行段中間問的「Do you want to pick it up?」
	// （記錄 0x111 的後半，`0x20354`，RE-113）。只有隊員會被問，敵人直接撿。
	// **這一問只認 y／n**，與逃走那一問同一支讀鍵（`sub_1B100(5)`）。
	PromptPickUp
	// PromptArea 是效果 6 的法術在執行段中間問的「Select an area to hurt／heal」
	// （記錄 0x184 ＋ 0x186／0x185 ＋ 0x16C，`0x2845E`，RE-100）。**只認 1…6**。
	// 要選目標格的法術不問，直接擲 `rand(6) − 1`。
	PromptArea
)

// Planning 是規劃段那一名走到哪裡。
type Planning struct {
	Prompt Prompt
	// BodyActor 為左下框被檢視的 actor 索引加一；零為空框，與行動者分開。
	BodyActor int
	// EquipmentPanel 為左上裝備模式；新隊員規劃重設，取消子選單保留。
	EquipmentPanel bool
	// Bar 是目前那一列的圖號；Selected 是白框停在第幾格。
	Bar      []int
	Selected int
	// Cursor 是選目標格的游標；四列子選單也沿用其最後位置。
	Cursor TargetCursor
	// Command／Style／Guard 是這一名已經決定的三個值。
	Command, Style, Guard int
	// after 是選完目標格之後要去哪一層。
	after Prompt
	// Env 是組指令列要用到的環境，由呼叫端每一名填一次。
	Env CommandEnv
	// GroundTile 是腳下那一塊磚，移動方式那一列要用。
	GroundTile int
	// Memo 是每一名上次記住的射擊目標（`[0x1A74 + 2i]`／`[0x1B36 + 2i]`）。
	Memo map[int]Cell
	// List 是文字清單那兩張（PromptSpell／PromptAffect）要顯示的內容；
	// ListPick 是游標。法術那一張放名字、正負那一張放兩筆記錄展開的文字。
	List     []string
	ListPick int
	// Unimplemented 記下按了但還沒實作的圖示，照 town.Session 的做法不默默吞掉。
	Unimplemented []int
	// Summary 是主指令列重畫時標在地圖上的那一格（視窗座標，`0x1F16A`…`0x1F38D`）；
	// SummaryShown 為假就不標。
	Summary      Cell
	SummaryShown bool
}

// BeginTurn 為現在輪到的那一名組出主指令列。不是隊員就什麼都不做。
func (s *State) BeginTurn() {
	i := s.Current()
	if i < 0 || s.Actors[i].Side != SideParty {
		return
	}
	m := s.member(i)
	if m == nil {
		return
	}
	// sub_249AE:0x249CB（RE-171-source）：主列重建先重載目前人物指令。
	s.LastCommand = s.Actors[i].Command
	s.Plan.Prompt = PromptCommand
	s.Plan.BodyActor = i + 1
	s.Plan.Env = s.CommandEnv(i)
	s.Plan.Bar = CommandBar(&s.Actors[i], m, s.Plan.Env)
	s.Plan.Selected = 0
	s.Plan.Command, s.Plan.Style, s.Plan.Guard = 0, 0, 0
	s.trackMain(i)
	s.Plan.Summary, s.Plan.SummaryShown = s.planSummary(i)
}

// planSummary 是 `0x1F16A`…`0x1F38D`：主指令列重畫前，依 `word_31122`（上一個寫回的
// 指令碼，`State.LastCommand`）決定要不要在地圖上標出這一名的目標格。
//
//	武器、徒手、施法、移動、騎乘移動，以及已上弦（`+0x1D`）的射擊 → 標
//	施法而且法術不用選目標格（`+0x0E` < 0x80）→ 改走面向那一格
//	射擊與要選目標格的施法 → 記住的目標格（`[0x1A74]`／`[0x1B36]`），出了 0…8 不標
//	其餘                      → 面向那一格（`[0x0B76]`）
//
// RE-171 訂正：有人條件只控制人物暫存重建，空格仍畫圓環。
// 全域指令已由 BeginTurn 對應 sub_249AE 從目前人物 +0x25 重載，不跨人物沿用。
func (s *State) planSummary(i int) (Cell, bool) {
	a := &s.Actors[i]
	cmd := s.LastCommand
	switch cmd {
	case CmdWeapon, CmdUnarmed, CmdCast, CmdMove, CmdMountedMove:
	case CmdMissile:
		if a.Flag1D == 0 {
			return Cell{}, false
		}
	default:
		return Cell{}, false
	}
	if cmd == CmdCast {
		m := s.member(i)
		if m == nil || a.SpellPick < 0 || a.SpellPick >= len(m.Spells) || !m.Spells[a.SpellPick].NeedsTarget() {
			cmd = 0
		}
	}
	var c Cell
	if cmd == CmdMissile || cmd == CmdCast {
		memo := s.memo(i)
		c = Cell{Row: memo.Row + ViewCentre.Row - a.X, Col: memo.Col + ViewCentre.Col - a.Y}
		if c.Row < 0 || c.Col < 0 || c.Row > 8 || c.Col > 8 {
			return Cell{}, false
		}
	} else {
		if a.Facing < 0 || a.Facing > 7 {
			return Cell{}, false
		}
		n := FacingCell(a.Facing)
		c = Cell{Row: n/3 + 3, Col: n%3 + 3}
	}
	return c, true
}

// TargetMarker 是 RE-171 已驗證的等待層圓環位置；子選單沿用最後目標格。
func (s *State) TargetMarker() (Cell, bool) {
	var c Cell
	switch s.Plan.Prompt {
	case PromptTarget, PromptWeaponForm, PromptUnarmedForm, PromptMoveForm, PromptGuardForm:
		c = Cell{Row: s.Plan.Cursor.Row, Col: s.Plan.Cursor.Col}
	case PromptCommand:
		if !s.Plan.SummaryShown {
			return Cell{}, false
		}
		c = s.Plan.Summary
	default:
		return Cell{}, false
	}
	return c, c.Row >= 0 && c.Row < ViewRows && c.Col >= 0 && c.Col < ViewCols
}

// MoveSelection 移動圖示列的白框：末格繞回，與城鎮同一條規則。
// 選目標格時 `,`／`.` 移的是同一個白框（`sub_24131:0x24771`）。
func (s *State) MoveSelection(step int) {
	n := len(s.Plan.Bar)
	if n == 0 {
		return
	}
	s.Plan.Selected = ((s.Plan.Selected+step)%n + n) % n
	if s.Plan.Prompt == PromptTarget {
		s.Plan.Cursor.Icon = s.Plan.Selected
	}
}

// MoveCursor 處理文字清單的上下鍵及目標格方向鍵；mouse 為真表示滑鼠指格。
func (s *State) MoveCursor(dRow, dCol int, mouse bool) {
	if s.Plan.Prompt == PromptSpell || s.Plan.Prompt == PromptAffect {
		// RE-205：文字選項與 YES／U-TURN 白框分開，首末項夾限。
		if mouse || dCol != 0 || len(s.Plan.List) == 0 {
			return
		}
		s.Plan.ListPick += dRow
		if s.Plan.ListPick < 0 {
			s.Plan.ListPick = 0
		}
		if s.Plan.ListPick >= len(s.Plan.List) {
			s.Plan.ListPick = len(s.Plan.List) - 1
		}
		return
	}
	if s.Plan.Prompt != PromptTarget {
		return
	}
	s.Plan.Cursor.Move(dRow, dCol, mouse)
	if dRow != 0 || dCol != 0 {
		s.inspectBodyTarget()
	}
}

// Choose 是按下 Return。回報這一輪的規劃是不是結束了。
func (s *State) Choose() bool {
	i := s.Current()
	if i < 0 || s.Actors[i].Side != SideParty {
		return false
	}
	switch s.Plan.Prompt {
	case PromptCommand:
		return s.chooseCommand(i)
	case PromptTarget:
		return s.chooseTarget(i)
	case PromptWeaponForm, PromptUnarmedForm:
		return s.chooseAttackForm(i)
	case PromptMoveForm:
		return s.chooseMoveForm(i)
	case PromptGuardForm:
		return s.chooseGuardForm(i)
	case PromptSpell:
		return s.chooseSpell(i)
	case PromptAffect:
		return s.chooseAffect(i)
	}
	return false
}

// chooseCommand 是主指令列的分派表（`0x1FA02`）。
func (s *State) chooseCommand(i int) bool {
	icon := s.icon()
	switch icon {
	case CmdYes:
		// 大姆指是確認（`0x1F39E`：`sub_228B7` 回 7 跳 `0x1FAC4`）：立起「規劃過了」、
		// 換下一名。記錄裡的指令碼與兩個選項原樣留著，所以新回合直接按 YES 就是
		// 重複上一回合那一手（RE-158）。
		return s.confirm()
	case CmdWeapon:
		s.Plan.Command = icon
		s.beginTarget(i, CursorAdjacent, PromptWeaponForm)
	case CmdUnarmed:
		s.Plan.Command = icon
		s.beginTarget(i, CursorAdjacent, PromptUnarmedForm)
	case CmdMove, CmdMountedMove:
		s.Plan.Command = icon
		s.beginTarget(i, CursorAdjacent, PromptMoveForm)
	case CmdCast:
		// 施法：先選法術，再選正負，最後才選目標格（`0x1F3C1`，RE-100）。
		s.Plan.Command = icon
		s.beginSpellList(i)
	case CmdMissile:
		s.Plan.Command = icon
		s.Plan.Guard = 0
		s.beginTarget(i, CursorRanged, PromptCommand)
	case CmdLand:
		// `0x1F914`：指令碼直接寫死，選項 0 與 2。
		return s.commit(i, CmdLand, 0, 2)
	case CmdRest, CmdDrop, CmdSwitchAlt1, CmdSwitch, CmdSwitchAlt2, CmdPickUp:
		// `0x1F3B2`：兩個選項都設 0，直接收尾。真正的動作在執行段。
		return s.commit(i, icon, 0, 0)
	case CmdLoad:
		// 分派表（`0x1FA02`）沒有列 0x22，所以它落 default 直接走收尾
		// `0x1FA6D`，指令碼寫的是 `sub_228B7:0x22D44` 存下來的圖號（RE-105）。
		// 兩個選項原版沿用上一次的殘留值，這裡取 0——裝填那一手不看它們。
		return s.commit(i, icon, 0, 0)
	default:
		// `CommandBar` 產生得出的圖號上面都有 case，正常指令列走不到這裡
		// （`TestPlanningEveryBarIconHandled`）。原版的 default 走收尾 `0x1FA6D`、
		// 把圖號寫進 +0x25；remake 只有指令列被改壞時才會到這裡，記下來不假裝做過。
		s.Plan.Unimplemented = append(s.Plan.Unimplemented, icon)
	}
	return false
}

// beginTarget 進選目標格那一層。
func (s *State) beginTarget(i, mode int, after Prompt) {
	if mode == CursorRanged {
		s.Plan.Cursor = NewRangedCursor(&s.Actors[i], s.memo(i))
	} else {
		s.Plan.Cursor = NewAdjacentCursor(&s.Actors[i])
	}
	s.Plan.after = after
	s.Plan.Prompt = PromptTarget
	s.Plan.Bar = []int{CmdYes, CmdCancel}
	s.Plan.Selected = 0
	s.Plan.Cursor.Icon = 0
	// sub_24131:0x2417A／0x24196：新進目標層先清空兩框。
	s.Plan.BodyActor = 0
	s.inspectBodyTarget()
}

// chooseTarget 是選目標格按下 Return（`0x248F5`）。
func (s *State) chooseTarget(i int) bool {
	cell, ok := s.Plan.Cursor.Confirm(&s.Actors[i])
	if !ok {
		// 白框停在 U-TURN 上：退回主指令列。
		s.BeginTurn()
		return false
	}
	if s.Plan.Cursor.Mode == CursorRanged {
		if s.Plan.Memo == nil {
			s.Plan.Memo = map[int]Cell{}
		}
		s.Plan.Memo[i] = cell
	}
	switch s.Plan.after {
	case PromptWeaponForm:
		s.Plan.Prompt = PromptWeaponForm
		s.Plan.Bar = WeaponForms()
	case PromptUnarmedForm:
		s.Plan.Prompt = PromptUnarmedForm
		s.Plan.Bar = UnarmedForms()
	case PromptMoveForm:
		s.Plan.Prompt = PromptMoveForm
		s.Plan.Bar = MoveForms(&s.Actors[i], s.Plan.GroundTile)
	default:
		// 射擊選完目標就收尾（`0x1F7BF`）。
		return s.commit(i, s.Plan.Command, s.Plan.Style, s.Plan.Guard)
	}
	s.Plan.Selected = 0
	return false
}

// chooseAttackForm 是武器與徒手兩列（`0x1F727`／`0x1F9AD`）：選完換防禦式。
func (s *State) chooseAttackForm(i int) bool {
	icon := s.icon()
	if icon == CmdCancel {
		s.beginTarget(i, CursorAdjacent, s.Plan.after)
		return false
	}
	if s.Plan.Prompt == PromptWeaponForm {
		s.Plan.Style = WeaponFormOption(icon)
		// 攻擊選項為 1 時原版不問防禦式，直接給 2 收尾（`0x1F738`）。
		if s.Plan.Style == 1 {
			return s.commit(i, s.Plan.Command, s.Plan.Style, 2)
		}
	} else {
		s.Plan.Style = UnarmedFormOption(icon)
	}
	s.Plan.Prompt = PromptGuardForm
	s.Plan.Bar = GuardForms()
	s.Plan.Selected = 0
	return false
}

// chooseMoveForm 是移動方式那一列（`0x1F8E8`）。
func (s *State) chooseMoveForm(i int) bool {
	icon := s.icon()
	if icon == CmdCancel {
		s.BeginTurn()
		return false
	}
	// 攻擊選項欄放的是圖號本身，防禦選項只有走路那一格是 2。
	return s.commit(i, s.Plan.Command, icon, MoveGuard(icon))
}

// chooseGuardForm 是防禦方式那一列（`sub_24AF8`）。
func (s *State) chooseGuardForm(i int) bool {
	icon := s.icon()
	if icon == CmdCancel {
		// 取消退回攻擊式那一列（`0x1F772`／`0x1F9CF`）。
		if s.Plan.Command == CmdWeapon {
			s.Plan.Prompt = PromptWeaponForm
			s.Plan.Bar = WeaponForms()
		} else {
			s.Plan.Prompt = PromptUnarmedForm
			s.Plan.Bar = UnarmedForms()
		}
		s.Plan.Selected = 0
		return false
	}
	command, style, guard := ResolveAttack(s.Plan.Command, s.Plan.Style, GuardOption(icon))
	return s.commit(i, command, style, guard)
}

// commit 把三個值寫回那一名，往下一名走，並回報這一輪是不是決定完了
// （`0x1FA6D`…`0x1FAE7`）。
//
// 寫回之後原版跳 `0x1F16A` 重畫**同一名**的指令列（白框在 Return 時已歸零，停在 YES），
// 還可以重選；按 YES 才換人（`confirm`）。回傳值恆為假。
func (s *State) commit(i, command, style, guard int) bool {
	s.Actors[i].Command = command
	s.Actors[i].Style = style
	s.Actors[i].Guard = guard
	s.LastCommand = command
	s.BeginTurn()
	return false
}

// confirm 是指令列按下 YES：這一名規劃完，換下一名；整輪決定完回報真。
func (s *State) confirm() bool {
	done := s.EndTurn()
	if !done {
		s.AdvanceToParty()
		if s.Current() >= 0 {
			s.BeginTurn()
			return false
		}
		done = true
	}
	// 整輪決定完印記錄 0x108，圖示列換成 YES 與 U-TURN，競技場與訓練戰以外再多
	// 一格逃走（`NextRoundBar`，RE-143）。這一列在執行段不會被重畫，訊息就蓋在它上面。
	s.Plan.Prompt = PromptNextRound
	s.Plan.Bar = s.NextRoundBar()
	s.Plan.Selected = 0
	return true
}

// icon 是白框停著的那一格圖號。
func (s *State) icon() int {
	if s.Plan.Selected < 0 || s.Plan.Selected >= len(s.Plan.Bar) {
		return CmdCancel
	}
	return s.Plan.Bar[s.Plan.Selected]
}

// member 回傳某一名參戰者對應的隊員資料；不是隊員就回 nil。
func (s *State) member(i int) *Member {
	if i < 0 || i >= len(s.Actors) {
		return nil
	}
	m := s.Actors[i].Member
	if m < 0 || m >= len(s.Party) {
		return nil
	}
	return &s.Party[m]
}

// memo 是某一名上次記住的射擊目標；沒有就回自己那一格。
func (s *State) memo(i int) Cell {
	if c, ok := s.Plan.Memo[i]; ok {
		return c
	}
	return Cell{Row: s.Actors[i].X, Col: s.Actors[i].Y}
}
