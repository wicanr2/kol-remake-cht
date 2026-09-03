package battle

import "strings"

// 訊息的代換（RE-086）。KOLDATA 的記錄裡有兩種記號：`]` 換成一個空白，
// `|X` 交給代換表換成名字、代名詞或裝備名。`~Vxx`／`~Hxx`／`~Cxx` 是座標與
// 顏色，屬於繪製那一層，這裡原樣留著讓呼叫端處理。

// 代名詞用的 KOLDATA 記錄（`sub_265F3:0x26630`／`0x26682`）。
const (
	// PossessiveBase − 性別 是所有格：男 his、女 her。
	PossessiveBase = 0x110
	// ObjectiveBase + 性別 是受格：男 him、女 her。
	ObjectiveBase = 0x10E
	// TheRecord 是敵人名字前面那個 `The]`。
	TheRecord = 0x140
)

// 訊息區的版面（`sub_1C1BF:0x1C1CE`…`0x1C25C`）。X 是 byte 欄，一欄 8 像素
// （`sub_1DE84:0x1DE9F` 的 `shl ax, 3`）；Y 是掃描列，行高 6 是小字型的高度。
var MessageArea = struct{ Col, Y, LineHeight int }{Col: 3, Y: 0xA8, LineHeight: 6}

// MessageWidth 是斷行用的字元數上限（`0x21`）。第一行要扣掉起始欄，所以只有
// `MessageWidth − MessageArea.Col` 個字元；第二行起原版直接加 `0x21`，兩者
// 不對稱——照抄原版的行為（RE-087）。
const MessageWidth = 0x21

// WrapMessage 依原版的規則把一整句斷行（`sub_1C1BF`）：從行寬處往回找空白當
// 斷點，跳過斷點上的空白，下一行再從那裡起算。整行找不到空白時原版會往回
// 掃出緩衝區，這裡改成硬斷，不讓它讀到界外。
func WrapMessage(text string) []string {
	var lines []string
	start := 0
	limit := MessageWidth - MessageArea.Col
	for limit < len(text) {
		end := limit
		for end > start && text[end] != ' ' {
			end--
		}
		if end == start {
			// 這一段裡沒有空白：原版會繼續往回掃，這裡硬斷在行寬處。
			end = limit
		}
		lines = append(lines, text[start:end])
		start = end + 1
		for start < len(text) && text[start] == ' ' {
			start++
		}
		limit = start + MessageWidth
	}
	return append(lines, text[start:])
}

// Speaker 是代換要用到的一名參戰者。
type Speaker struct {
	// Name 是名字；Monster 為真時原版會在名字前面補 `The `。
	Name    string
	Monster bool
	// Female 對應參戰者記錄 +0x07（隊員取角色記錄 0x1A）。
	Female bool
	// Weapon 是手上那件武器的名字（`|E`）；Belt 是角色記錄 +0x190 那一格的
	// 名字（`|I`），語意見 RE-086。
	Weapon string
	Belt   string
}

// Pronoun 回傳這一名的所有格或受格要用哪一筆 KOLDATA 記錄。
func (s Speaker) Pronoun(possessive bool) int {
	sex := 0
	if s.Female {
		sex = 1
	}
	if possessive {
		return PossessiveBase - sex
	}
	return ObjectiveBase + sex
}

// MessageContext 是展開一則訊息要用到的東西。Lookup 依 KOLDATA 記錄號取文字；
// 沒有的話代名詞會退回內建的英文。
type MessageContext struct {
	Actor  Speaker
	Target Speaker
	Lookup func(record int) string
	// Item 是 `|H` 要填的名字；Scratch 是 `|S` 那一塊暫存區的名字。
	Item    string
	Scratch string
	// Number 是 `|N` 要印的數字。
	Number int
}

// Substitute 把一則訊息裡的 `]` 與 `|X` 展開（`sub_1A8FD:0x1A961`、
// `sub_265F3`）。認不得的代換碼原樣留著，不默默吞掉。
func (c MessageContext) Substitute(text string) string {
	var b strings.Builder
	for i := 0; i < len(text); i++ {
		ch := text[i]
		if ch == ']' {
			b.WriteByte(' ')
			continue
		}
		if ch != '|' || i+1 >= len(text) {
			b.WriteByte(ch)
			continue
		}
		i++
		// 代換之前先補一個空白（`sub_265F3:0x265FC` 的 `sub_1C2DA(' ')`）。
		// 原版由 `word_345E0` 控制，那個旗標的語意還沒定，但 KOLDATA 裡每一句
		// 帶代換碼的訊息都需要這個空白才讀得通（例如 `kicks|B leg`）。
		b.WriteByte(' ')
		b.WriteString(c.expand(text[i]))
	}
	return b.String()
}

// expand 是 `sub_265F3` 的分派表。
func (c MessageContext) expand(code byte) string {
	switch code {
	case 'A':
		return c.pronoun(c.Target, true)
	case 'B':
		return c.pronoun(c.Actor, true)
	case 'C':
		return c.pronoun(c.Target, false)
	case 'D':
		return c.pronoun(c.Actor, false)
	case 'E':
		return c.Actor.Weapon
	case 'F':
		return c.name(c.Target)
	case 'G':
		return c.name(c.Actor)
	case 'H':
		return c.Item
	case 'I':
		return c.Actor.Belt
	case 'N':
		return itoa(c.Number)
	case 'S':
		return c.Scratch
	}
	// 其餘的碼原版也不認，原樣留著不默默吞掉。
	return "|" + string(code)
}

// pronoun 取代名詞。沒給 Lookup 就退回內建的英文，讓規則層自己也能測。
func (c MessageContext) pronoun(s Speaker, possessive bool) string {
	record := s.Pronoun(possessive)
	if c.Lookup != nil {
		if t := c.Lookup(record); t != "" {
			return t
		}
	}
	switch record {
	case 0x10E:
		return "him"
	case 0x10F:
		return "her"
	case 0x110:
		return "his"
	}
	return ""
}

// name 是 `|F`：敵人的名字前面要補 KOLDATA 0x140（`The]`），隊員直接印。
func (c MessageContext) name(s Speaker) string {
	if !s.Monster {
		return s.Name
	}
	prefix := "The "
	if c.Lookup != nil {
		if t := c.Lookup(TheRecord); t != "" {
			prefix = strings.ReplaceAll(t, "]", " ")
			// 記錄本身以 `]` 收尾（`The]`），有些取法會把它 trim 掉；
			// 名字前面一定要有一個空白。
			if !strings.HasSuffix(prefix, " ") {
				prefix += " "
			}
		}
	}
	return prefix + s.Name
}

// Line 把一手的訊息組成畫面上那一整句（`sub_2685A` ＋ 記錄）：
// 先印行動者的名字，補一個空白，再接記錄展開之後的內容（RE-086）。
func (c MessageContext) Line(text string) string {
	return c.name(c.Actor) + " " + c.Substitute(text)
}

// itoa 是給 `|N` 用的十進位轉換；不用 strconv 是為了讓這一支不帶額外相依。
func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
