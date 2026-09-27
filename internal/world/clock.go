package world

import "fmt"

// 野外的時刻、日曆與天候（RE-077）。整段在 `sub_297F2` 與 `sub_29987` 裡：
// 每走一步扣時刻，扣到負的就換日；換日要補營養、推日期；時段變了才重算天候。

// ClockStart 是時刻的起點與換日之後的重設值（`sub_297F2:0x2982D`，
// DGROUP+0x70 的初值也是這個數）。時刻是倒數的：4200 走到 0 是一天。
const ClockStart = 0x1068

// DayParts 是一天分成幾段。
const DayParts = 5

// dayPartFloor 是每一段的下限（DGROUP+0x12CC 的五個 word）。
// 時段是「時刻落在哪一段」：第一個滿足 `時刻 >= 下限` 的段。
var dayPartFloor = [DayParts]int{3367, 2527, 1687, 847, 0}

// monthDays 是每月天數（DGROUP+0x12C0）。原版沒有閏年。
var monthDays = [12]int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}

// monthName 是日期框裡的月名（DGROUP+0x63A，一筆 4 byte 的 C 字串）。
var monthName = [12]string{"Jan", "Feb", "Mar", "Apr", "May", "Jun",
	"Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

var chineseNumeral = [10]string{"", "一", "二", "三", "四", "五", "六", "七", "八", "九"}

// DateLinesZh 是中文畫面的日期標籤；不改原版時鐘或英文對拍的 DateLines。
func (c *Clock) DateLinesZh() []string {
	return []string{chineseDay(c.Month+1) + "月" + chineseDay(c.Day+1) + "日",
		fmt.Sprintf("西元 %d", YearBase+c.Year)}
}

func chineseDay(n int) string {
	if n < 10 {
		return chineseNumeral[n]
	}
	if n < 20 {
		return "十" + chineseNumeral[n%10]
	}
	return chineseNumeral[n/10] + "十" + chineseNumeral[n%10]
}

// hungerPerDay 是換日時每名隊員的營養條要加多少（DGROUP+0x12D6）。
// 數字越大條子越紅：一月四月最餓、六月與九月不餓，與說明書「晚春與初秋最豐饒」相符。
var hungerPerDay = [12]int{4, 4, 3, 2, 1, 0, 1, 1, 0, 1, 2, 3}

// weatherChance 是每月「天氣轉壞」的門檻（DGROUP+0x131A）。
var weatherChance = [12]int{40, 40, 35, 30, 25, 25, 20, 15, 25, 30, 35, 40}

// weatherBounds 把上下限包在一個 byte 裡（DGROUP+0x1326）：
// 低三位元是上限、除以 8 之後的低三位元是下限；大於 0x7F 的月份會多一段（下雪）。
var weatherBounds = [12]int{139, 139, 131, 3, 3, 2, 2, 1, 2, 10, 139, 139}

// stepCost 是 `sub_297F2` 查的那三個 byte（DGROUP+0x12B7 起：`00 0A 0A`）。
// DGROUP+0x12B8 本來是八種地形的成本表（`10 10 10 20 6 14 30 8`），但原版拿來
// 當索引的是 `word_39A54`——這一步的 Y 位移（−1、0、+1），不是地形碼，所以往北
// 讀到表前面那個 0、其餘方向讀到沙漠與河的 10（RE-151，原版記憶體實測）。
var stepCost = [3]int{0, 10, 10}

// DefaultMovement 是移動率 0 時代入的值（`sub_29004:0x290C2`）。
const DefaultMovement = 20

// CampWeather 是換日那一張畫面用的天候組；ClearWeather 是收掉之後設回去的那一組。
const (
	CampWeather  = 5
	ClearWeather = 1
)

// CampRecord 是換日那一張畫面的 KOLDATA 記錄：`"CAMPING!"`。
const CampRecord = 0xE2

// Clock 是野外的時刻與日曆。
type Clock struct {
	// Ticks 是這一天剩下的時刻，從 ClockStart 倒數。
	Ticks int
	// Part 是時段（0 DAWN…4 TWILIGHT）。
	Part int
	// Day 是當月第幾天，從 0 起算；畫面上印 Day+1。
	Day int
	// Month 是月份 0…11。
	Month int
	// Year 是從 1100 起算的年份差；畫面上印 1100+Year。
	Year int
	// Weather 是天候組別 0…4。
	Weather int
}

// YearBase 是畫面上那個年份的基準（`sub_17095:0x1712C` 的 0x44C）。
const YearBase = 1100

// NewClock 是新遊戲的起點：西元 1100 年四月一日，天亮，時刻滿格。
func NewClock() *Clock {
	return &Clock{Ticks: ClockStart, Part: 0, Day: 0, Month: 3, Year: 0, Weather: 1}
}

// StepCost 回傳走一步要扣的時刻：`2 × ([0x12B8 + dy] + 隊伍移動率)`
// （`sub_297F2` 開頭）。dy 是這一步的 Y 位移，往北 −1、往南 +1、東西 0；
// 移動率取隊伍裡最大的那個，0 當成 20。所以移動率 20 時往北 40、其餘 60，
// 與腳下地形無關（RE-151：dosgolem 在平原上四個方向各走八步量到）。
// 循路那一步也走同一支，dy 是模擬出來那一步的方向。
func (c *Clock) StepCost(dy, movement int) int {
	if movement == 0 {
		movement = DefaultMovement
	}
	i := 1
	switch {
	case dy < 0:
		i = 0
	case dy > 0:
		i = 2
	}
	return 2 * (stepCost[i] + movement)
}

// Advance 走一步。回傳是不是換日了——換日要跳出 CAMPING! 畫面。
//
// **日期不在這裡推。** 原版的順序是：時刻歸位、補營養、停在 CAMPING! 等鍵，
// 收掉之後才推日期並重畫日期框（`sub_297F2:0x2993B`），所以紮營那一張上面
// 印的還是前一天。推日期是 `NextDay`。
func (c *Clock) Advance(dy, movement int) bool {
	return c.Spend(c.StepCost(dy, movement))
}

// Spend 扣掉指定的時刻，回傳這一步是不是把日子推過去了。**時段不在這裡更新**：
// 原版 `sub_297F2` 先呼叫 `sub_29987` 依扣之前的時刻判時段，再扣（RE-151），
// 所以跨過門檻的那一步畫面上還是舊時段，下一步才換。時段由 `UpdatePart` 更新。
func (c *Clock) Spend(cost int) bool {
	c.Ticks -= cost
	if c.Ticks >= 0 {
		return false
	}
	c.Ticks = ClockStart
	return true
}

// UpdatePart 是 `sub_29987` 判時段那一段：依目前的時刻重算，回傳是不是換了。
// 換了原版接著重畫標籤、擲野外遭遇與天候（呼叫端處理）。
func (c *Clock) UpdatePart() bool {
	p := c.partFor(c.Ticks)
	if p == c.Part {
		return false
	}
	c.Part = p
	return true
}

// partFor 是時刻對應的時段。
func (c *Clock) partFor(ticks int) int {
	for i := 0; i < DayParts; i++ {
		if ticks >= dayPartFloor[i] {
			return i
		}
	}
	return DayParts - 1
}

// NextDay 把日期往前推一天（`sub_297F2:0x2993B`）。收掉紮營那一張之後才做。
func (c *Clock) NextDay() {
	c.Day++
	if c.Day != monthDays[c.Month] {
		return
	}
	c.Day = 0
	c.Month++
	if c.Month != 12 {
		return
	}
	c.Month = 0
	c.Year++
}

// Hunger 是換日時每名隊員的營養條要加的量。
func (c *Clock) Hunger() int { return hungerPerDay[c.Month] }

// NextWeather 依這個月的機率表走一步隨機漫步（`sub_29987:0x29A83`…`0x29B0C`）。
// roll 是 `sub_1B5A8(0x80)` 的結果減一，也就是 0…127。
func (c *Clock) NextWeather(roll int) int {
	w, bounds := c.Weather, weatherBounds[c.Month]
	switch {
	case roll < weatherChance[c.Month]:
		if w < bounds&7 {
			w++
		}
	case roll < 100:
		if w > (bounds/8)&7 {
			w--
		}
	}
	// 大於 0x7F 的月份會把第 3 段再往上推一段——冬天才有的那一張。
	if bounds > 0x7F && w == 3 {
		w++
	}
	return w
}

// DateLines 是日期框裡的兩行（`sub_17095`）：月名畫在 byte 欄 3、日期在欄 7，
// `AD` 在欄 2、年份在欄 6。框的文字起點是欄 2，所以第一行前面補一格空白。
func (c *Clock) DateLines() []string {
	return []string{fmt.Sprintf(" %s %d", monthName[c.Month], c.Day+1),
		fmt.Sprintf("AD  %d", YearBase+c.Year)}
}
