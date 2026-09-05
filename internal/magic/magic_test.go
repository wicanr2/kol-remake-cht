package magic

import (
	"os"
	"path/filepath"
	"testing"
)

func load(t *testing.T) *Dictionary {
	t.Helper()
	p := filepath.Join("..", "..", "assets", "original-derived-data", "magic.json")
	if _, err := os.Stat(p); err != nil {
		t.Skip("沒有 magic.json，跳過")
	}
	d, err := Load(p)
	if err != nil {
		t.Fatalf("讀不進來：%v", err)
	}
	return d
}

func TestClassAndStat(t *testing.T) {
	d := load(t)
	for _, c := range []struct {
		n    int
		want string
	}{{0, "None"}, {5, "Elemental"}, {8, "Undead"}} {
		if got := d.ClassName(c.n); got != c.want {
			t.Fatalf("第 %d 類是 %q，期望 %q", c.n, got, c.want)
		}
	}
	if got := d.StatName(5); got != "Fatigue" {
		t.Fatalf("第 5 個屬性是 %q", got)
	}
	// 超出範圍是空字串，不是 panic。
	if d.ClassName(-1) != "" || d.ClassName(99) != "" {
		t.Fatal("超出範圍應該回空字串")
	}
}

// SUBCLASS 依 CLASS 分組：Undead 那一組是四隻不死系。
func TestSubclassGroups(t *testing.T) {
	d := load(t)
	if got := d.SubclassName(8, 3); got != "Zombie" {
		t.Fatalf("第 8 類第 3 種是 %q", got)
	}
	if got := d.SubclassName(6, 7); got != "Hill Giant" {
		t.Fatalf("第 6 類第 7 種是 %q", got)
	}
	// 人類那幾組是清一色的 All。
	for _, name := range d.SubclassGroup(1) {
		if name != "All" {
			t.Fatalf("第 1 類有 %q", name)
		}
	}
}

// 成本的索引是整個 byte：高 nibble 是類、低 nibble 是種。
func TestTargetCost(t *testing.T) {
	d := load(t)
	class, sub := Target(0x51)
	if class != 5 || sub != 1 {
		t.Fatalf("0x51 拆成 %d／%d", class, sub)
	}
	// 這是唯一兩種資源不同額的一張表：金冠是冒險點數的兩倍。
	ap, gc := d.TargetCost(0x51) // Elemental／Golem
	if ap != 50 || gc != 100 {
		t.Fatalf("Golem 的成本是 %d 冒險點數／%d 金冠", ap, gc)
	}
	ap, gc = d.TargetCost(0x54) // Elemental／Salamandar
	if ap != 200 || gc != 400 {
		t.Fatalf("Salamandar 的成本是 %d 冒險點數／%d 金冠", ap, gc)
	}
	// 沒有目標就不用錢。
	if ap, gc := d.TargetCost(0); ap != 0 || gc != 0 {
		t.Fatalf("空的那一格要 %d／%d", ap, gc)
	}
}

// 骰式字串本身就是「低-高」。
func TestDice(t *testing.T) {
	d := load(t)
	if got := d.DiceName(0); got != "1-3" {
		t.Fatalf("第 0 個骰式是 %q", got)
	}
	if len(d.Dice) != 48 {
		t.Fatalf("骰式有 %d 個", len(d.Dice))
	}
}
