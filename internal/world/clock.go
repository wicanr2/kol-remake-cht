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

// hungerPerDay 是換日時每名隊員的營養條要加多少（DGROUP+0x12D6）。
// 數字越大條子越紅：一月四月最餓、六月與九月不餓，與說明書「晚春與初秋最豐饒」相符。
var hungerPerDay = [12]int{4, 4, 3, 2, 1, 0, 1, 1, 0, 1, 2, 3}

// weatherChance 是每月「天氣轉壞」的門檻（DGROUP+0x131A）。
var weatherChance = [12]int{40, 40, 35, 30, 25, 25, 20, 15, 25, 30, 35, 40}

// weatherBounds 把上下限包在一個 byte 裡（DGROUP+0x1326）：
// 低三位元是上限、除以 8 之後的低三位元是下限；大於 0x7F 的月份會多一段（下雪）。
var weatherBounds = [12]int{139, 139, 131, 3, 3, 2, 2, 1, 2, 10, 139, 139}

// terrainStep 是每種地形走一步的基本成本（DGROUP+0x12B8，索引是地形碼）：
// 沙漠、河、沼澤、丘陵、道路、森林、岸、平原。道路最省、岸最貴，與說明書
// 「公路最快、丘陵最難走、平原最好走」一致。
var terrainStep = [8]int{10, 10, 10, 20, 6, 14, 30, 8}

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

// StepCost 回傳走一步要扣的時刻：`2 × (地形成本 + 隊伍移動率)`
// （`sub_297F2:0x29800`…`0x2981F`）。移動率取隊伍裡最大的那個，0 當成 20。
//
// 這個式子由原版實測釘住：`world-clock-run2` 在平原上左右來回，第 15 次按鍵
// 跨過 DAWN 的門檻（4200 − 15 × 56 = 3360 < 3367，第 14 次的 3416 還沒跨）。
// 平原成本 8、隊伍移動率 20，只有 56 對得上。
func (c *Clock) StepCost(terrain, movement int) int {
	if movement == 0 {
		movement = DefaultMovement
	}
	step := 0
	if terrain >= 0 && terrain < len(terrainStep) {
		step = terrainStep[terrain]
	}
	return 2 * (step + movement)
}

// Advance 走一步。回傳是不是換日了——換日要跳出 CAMPING! 畫面。
//
// **日期不在這裡推。** 原版的順序是：時刻歸位、補營養、停在 CAMPING! 等鍵，
// 收掉之後才推日期並重畫日期框（`sub_297F2:0x2993B`），所以紮營那一張上面
// 印的還是前一天。推日期是 `NextDay`。
func (c *Clock) Advance(terrain, movement int) bool {
	c.Ticks -= c.StepCost(terrain, movement)
	if c.Ticks >= 0 {
		c.Part = c.partFor(c.Ticks)
		return false
	}
	c.Ticks = ClockStart
	c.Part = c.partFor(c.Ticks)
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
