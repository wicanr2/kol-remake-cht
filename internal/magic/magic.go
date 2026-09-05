// Package magic 讀 `MAGIC.0` 抽出來的法術參數字典（RE-128）。
//
// 法術編輯器（SCROLL 選單第 1 列）十三個欄位的選項全部來自這一份：法術對哪一類
// 怪物有效（CLASS）、對其中哪一種有效（SUBCLASS，依 CLASS 分組）、影響哪個屬性
// （STAT）、傷害與持續的骰式，以及每一格要花多少冒險點數與金冠。
package magic

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Entries 是一段位址表有幾筆——原版每一段都是十六筆，高 nibble 因此是 0…15。
const Entries = 16

// Dictionary 是整份字典。
type Dictionary struct {
	Schema string `json:"schema"`
	// Class 是九類目標：None 之外都是怪物的族群。
	Class []string `json:"class"`
	// Subclass 依 Class 分組，一組最多十六筆；沒有東西的那幾組是空的。
	Subclass [][]string `json:"subclass"`
	// Stat 是九個屬性；Dice 是四十八個骰式，字串本身就是 `低-高`。
	Stat []string `json:"stat"`
	Dice []string `json:"dice"`
	// Cost 的索引是那一格的整個 byte（高 nibble ＝ Class、低 nibble ＝ Subclass）。
	// 這是唯一兩種資源不同額的一張表：對 Golem 是 100 金冠 ＋ 50 冒險點數。
	Cost struct {
		Adventure []int `json:"adventure"`
		Gold      []int `json:"gold"`
	} `json:"cost"`
	// StatCost 是 `+0x0C` 那一格自己的兩種成本，索引是那一格的值。
	StatCost struct {
		Adventure []int `json:"adventure"`
		Gold      []int `json:"gold"`
	} `json:"stat_cost"`
	// DiceCode 是骰式選單第 n 項真正寫進記錄的值（`0x18109`）。挑第 n 項時原版
	// 只換記錄那一格的中間六個位元，所以這裡存的就是那六個位元的位置。
	DiceCode []int `json:"dice_code"`
	// DiceCost 是骰式的成本：骰數與骰面分開查、兩邊相加（`sub_18296`）。
	DiceCost struct {
		CountAdventure []int `json:"count_adventure"`
		CountGold      []int `json:"count_gold"`
		SidesAdventure []int `json:"sides_adventure"`
		SidesGold      []int `json:"sides_gold"`
	} `json:"dice_cost"`
	// TargetFlagCostStep 是「要不要先選目標」那一欄的級距：兩種選擇分別是
	// 一倍與兩倍，冒險點數與金冠同額（`0x1864B`）。
	TargetFlagCostStep int `json:"target_flag_cost_step"`
}

// TurnDiceOptions 是持續回合那一欄的選項數（`0x17E76`）：只到骰面 1-6 那一組。
// 強度那一欄看得到全部四十八項（`0x17E8E`）。
const TurnDiceOptions = 0x18

// Load 讀一份字典。欄位少一塊就直接失敗——法術的選單少一項不會有錯誤訊息，
// 只會靜靜地少一個選擇。
func Load(path string) (*Dictionary, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var d Dictionary
	if err := json.Unmarshal(raw, &d); err != nil {
		return nil, fmt.Errorf("%s：%w", path, err)
	}
	if !strings.HasPrefix(d.Schema, "kol-magic-dictionary-") {
		return nil, fmt.Errorf("%s：schema 是 %q", path, d.Schema)
	}
	if err := d.check(); err != nil {
		return nil, fmt.Errorf("%s：%w", path, err)
	}
	return &d, nil
}

func (d *Dictionary) check() error {
	if len(d.Class) == 0 || d.Class[0] != "None" {
		return fmt.Errorf("Class 是 %v", d.Class)
	}
	if len(d.Stat) == 0 || d.Stat[0] != "None" {
		return fmt.Errorf("Stat 是 %v", d.Stat)
	}
	if len(d.Subclass) != Entries {
		return fmt.Errorf("Subclass 有 %d 組，應該是 %d", len(d.Subclass), Entries)
	}
	if len(d.Dice) == 0 {
		return fmt.Errorf("沒有骰式")
	}
	for _, c := range [][]int{d.Cost.Adventure, d.Cost.Gold} {
		if len(c) != 0x100 {
			return fmt.Errorf("成本表有 %d 筆，應該是 256", len(c))
		}
	}
	if len(d.DiceCode) != len(d.Dice) {
		return fmt.Errorf("骰式有 %d 個名字卻有 %d 個值", len(d.Dice), len(d.DiceCode))
	}
	if d.TargetFlagCostStep <= 0 {
		return fmt.Errorf("目標旗標的級距是 %d", d.TargetFlagCostStep)
	}
	return nil
}

// DiceValue 是骰式選單第 n 項要寫進記錄的值。
func (d *Dictionary) DiceValue(n int) int {
	if n < 0 || n >= len(d.DiceCode) {
		return 0
	}
	return d.DiceCode[n]
}

// DiceIndex 是「記錄裡這一格現在停在選單第幾項」。原版畫面不做這個反查——
// 它直接把值拆成骰數與骰面印出來（`sub_18296`）——但 remake 的選單要有游標。
// 找不到就回 0；那一格的中間六個位元本來就只有這四十八種組合。
func (d *Dictionary) DiceIndex(value int) int {
	for i, code := range d.DiceCode {
		if code == value&0x7E {
			return i
		}
	}
	return 0
}

// DiceCostOf 是一個骰式那一格的兩種成本（`sub_18296`）：骰數與骰面各查一張表
// 再相加，兩者都是單一 byte。
func (d *Dictionary) DiceCostOf(value int) (adventure, gold int) {
	sides, count := (value&0x0E)>>1, (value&0x70)>>4
	return costAt(d.DiceCost.CountAdventure, count) + costAt(d.DiceCost.SidesAdventure, sides),
		costAt(d.DiceCost.CountGold, count) + costAt(d.DiceCost.SidesGold, sides)
}

// TargetFlagCost 是「要不要先選目標」那一欄的成本：`100 × 值 + 100`，兩種資源同額。
func (d *Dictionary) TargetFlagCost(on bool) (adventure, gold int) {
	n := d.TargetFlagCostStep
	if on {
		n *= 2
	}
	return n, n
}

func costAt(table []int, n int) int {
	if n < 0 || n >= len(table) {
		return 0
	}
	return table[n]
}

// ClassName 是第 n 類的名字；超出範圍回空字串，畫面上就是那一格空著。
func (d *Dictionary) ClassName(n int) string { return at(d.Class, n) }

// StatName 是第 n 個屬性的名字。
func (d *Dictionary) StatName(n int) string { return at(d.Stat, n) }

// DiceName 是第 n 個骰式的名字。
func (d *Dictionary) DiceName(n int) string { return at(d.Dice, n) }

// SubclassName 是「第 class 類的第 n 種」。
func (d *Dictionary) SubclassName(class, n int) string {
	if class < 0 || class >= len(d.Subclass) {
		return ""
	}
	return at(d.Subclass[class], n)
}

// SubclassGroup 是第 class 類底下的選項。
func (d *Dictionary) SubclassGroup(class int) []string {
	if class < 0 || class >= len(d.Subclass) {
		return nil
	}
	return d.Subclass[class]
}

// TargetCost 是「讓法術對這一格的目標有效」要花的兩種資源。索引就是那一格的整個
// byte——高 nibble 是 Class、低 nibble 是 Subclass（RE-128）。
func (d *Dictionary) TargetCost(value int) (adventure, gold int) {
	if value < 0 || value >= len(d.Cost.Adventure) {
		return 0, 0
	}
	return d.Cost.Adventure[value], d.Cost.Gold[value]
}

// StatCostOf 是 `+0x0C` 那一格的兩種成本。
func (d *Dictionary) StatCostOf(value int) (adventure, gold int) {
	if value < 0 || value >= len(d.StatCost.Adventure) {
		return 0, 0
	}
	return d.StatCost.Adventure[value], d.StatCost.Gold[value]
}

// Target 把一個目標 byte 拆成「哪一類、哪一種」。
func Target(value int) (class, sub int) { return value >> 4, value & 0x0F }

func at(rows []string, n int) string {
	if n < 0 || n >= len(rows) {
		return ""
	}
	return rows[n]
}
