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
	Cost struct {
		Adventure []int `json:"adventure"`
		Gold      []int `json:"gold"`
	} `json:"cost"`
	// StatCost 是 `+0x0C` 那一格自己的兩種成本，索引是那一格的值。
	StatCost struct {
		Adventure []int `json:"adventure"`
		Gold      []int `json:"gold"`
	} `json:"stat_cost"`
}

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
	return nil
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
