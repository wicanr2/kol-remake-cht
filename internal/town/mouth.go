package town

import "strings"

// MOUTH：玩家在對話框裡打一句話問人（RE-067）。原版是
// `sub_14234(0x500, 0)` 放訊息 18，接著 `sub_1B5C8(0x2E3A, 0x17)` 讀入
// **最多 23 個字元**，整串轉小寫再走關鍵字的字首比對。

// MouthMaxChars 是問句的長度上限（`sub_1B5C8` 的第二個參數 0x17）。
const MouthMaxChars = 23

// mouthState 是打字中的問句。
type mouthState struct {
	// prompt 是訊息 18 的那幾行，一直留在對話框上。
	prompt []string
	buf    []rune
}

// StartAsk 是按下 MOUTH：放問話的敘述，開始收字。**圖示列不收掉**——原版按下
// MOUTH 之後底下那一排與白框原樣留著（收據 `town-mouth-run1`：frame-056 與
// frame-058 的差異只在對話框那 40 列），鍵盤則整段交給 `sub_1B5C8`，所以圖示
// 看得到但按不動。
func (s *Session) StartAsk() {
	if s.Mode != ModeInside || s.Inside == nil {
		return
	}
	s.mouth = &mouthState{prompt: s.message(AskMessage)}
	s.refreshMouth()
}

// Asking 回報現在是不是在打問句。
func (s *Session) Asking() bool { return s.mouth != nil }

// AskInput 是目前打進去的字。
func (s *Session) AskInput() string {
	if s.mouth == nil {
		return ""
	}
	return string(s.mouth.buf)
}

// TypeAsk 收一個字元。超過上限就不再收——原版的讀入函式也是收滿就停。
func (s *Session) TypeAsk(r rune) {
	if s.mouth == nil || len(s.mouth.buf) >= MouthMaxChars {
		return
	}
	s.mouth.buf = append(s.mouth.buf, r)
	s.refreshMouth()
}

// BackspaceAsk 退一格。
func (s *Session) BackspaceAsk() {
	if s.mouth == nil || len(s.mouth.buf) == 0 {
		return
	}
	s.mouth.buf = s.mouth.buf[:len(s.mouth.buf)-1]
	s.refreshMouth()
}

// SubmitAsk 送出問句。空的就當作沒問，收掉輸入框回到圖示列。
func (s *Session) SubmitAsk() bool {
	if s.mouth == nil {
		return false
	}
	question := strings.TrimSpace(string(s.mouth.buf))
	// 讀入函式收工前把筆尖那一格的橫槓蓋掉（RE-114，`sub_1B5C8` 結尾的
	// `sub_1C2DA(' ')`），所以不論問句是否為空，游標都要從對話框消失。
	if s.Inside != nil {
		s.Inside.Dialog = append(append([]string(nil), s.mouth.prompt...), string(s.mouth.buf))
		s.Inside.DialogZh = nil
	}
	s.mouth = nil
	if question == "" {
		s.setInsideIcons(InteriorIcons(s.door.IconFlags, len(s.Party)))
		return false
	}
	return s.Ask(question)
}

// refreshMouth 把敘述與打到一半的字一起放進對話框。筆尖那格的游標與訓練、賭場
// 兩處文字欄共用同一支原版函式（`sub_1B5C8`），所以畫的是同一條橫槓，不是底線
// 字模——見 TextCaret 的說明與 RE-114（版面與原版逐像素對過，
// `plan-town-mouth.json`）。
func (s *Session) refreshMouth() {
	if s.mouth == nil || s.Inside == nil {
		return
	}
	lines := append([]string(nil), s.mouth.prompt...)
	lines = append(lines, string(s.mouth.buf)+TextCaret)
	s.Inside.Dialog = lines
	s.Inside.DialogZh = nil
}
