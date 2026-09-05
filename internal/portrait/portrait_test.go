package portrait

import (
	"os"
	"path/filepath"
	"testing"
)

// layoutPath 是抽出來的版面表。它依現行權利邊界留在本機，所以測試在找不到的時候
// 跳過——但只要在，就用真的那一份驗，不另外編一套假數字。
func layoutPath(t *testing.T) string {
	t.Helper()
	p := filepath.Join("..", "..", "assets", "original-derived-data", "portrait-layout.json")
	if _, err := os.Stat(p); err != nil {
		t.Skip("沒有 portrait-layout.json，跳過")
	}
	return p
}

func load(t *testing.T) *Layout {
	t.Helper()
	l, err := Load(layoutPath(t))
	if err != nil {
		t.Fatalf("讀不進來：%v", err)
	}
	return l
}

func TestBuildIsGenderTimesFourPlusRace(t *testing.T) {
	for _, c := range []struct {
		gender, race, want int
	}{{0, 0, 0}, {0, 3, 3}, {1, 0, 4}, {1, 1, 5}} {
		got := Figure{Gender: c.gender, Race: c.race}.Build()
		if got != c.want {
			t.Fatalf("性別 %d 種族 %d 得到 %d，期望 %d", c.gender, c.race, got, c.want)
		}
	}
}

// 沒有裝備的人只有體型與一隻手。
func TestComposeBareFigure(t *testing.T) {
	l := load(t)
	f := Figure{Styles: [LayerCount]int{-1, -1, -1}}
	got := l.Compose(f, 0)
	if len(got) != 2 {
		t.Fatalf("得到 %d 層：%+v", len(got), got)
	}
	if got[0].Archive != ArchiveBody || got[0].Index != 0 {
		t.Fatalf("第一層應該是體型 0：%+v", got[0])
	}
	if got[1].Archive != ArchiveHand || got[1].Index != 0 {
		t.Fatalf("最後一層應該是手 0：%+v", got[1])
	}
}

// 三層護甲由後往前：腿、身、頭。
func TestComposeLayerOrder(t *testing.T) {
	l := load(t)
	f := Figure{Styles: [LayerCount]int{8, 8, 8}}
	got := l.Compose(f, 0)
	want := []string{ArchiveBody, ArchiveLegs, ArchiveTorso, ArchiveHead, ArchiveHand}
	if len(got) != len(want) {
		t.Fatalf("得到 %d 層：%+v", len(got), got)
	}
	for i, w := range want {
		if got[i].Archive != w {
			t.Fatalf("第 %d 層是 %s，期望 %s", i, got[i].Archive, w)
		}
	}
	// 外觀號是 9 × 體型組合 ＋ 那一格的低 nibble。
	for _, p := range got[1:4] {
		if p.Index != 8 {
			t.Fatalf("%s 的組號是 %d，期望 8", p.Archive, p.Index)
		}
	}
}

// 雙手武器沒有手拿盾，但會多畫一隻手。
func TestTwoHandedDropsShieldAddsHand(t *testing.T) {
	l := load(t)
	f := Figure{
		Styles: [LayerCount]int{-1, -1, -1},
		Weapon: &Item{Class: 87, Style: 1, Kind: 1, TwoHand: true},
		Shield: &Item{Style: 1},
	}
	got := l.Compose(f, 0)
	hands := 0
	for _, p := range got {
		if p.Archive == ArchiveShield {
			t.Fatalf("雙手武器不該有盾：%+v", got)
		}
		if p.Archive == ArchiveHand {
			hands++
		}
	}
	if hands != 2 {
		t.Fatalf("雙手武器應該畫兩隻手，得到 %d", hands)
	}
	f.Weapon.TwoHand = false
	got = l.Compose(f, 0)
	hands = 0
	shields := 0
	for _, p := range got {
		switch p.Archive {
		case ArchiveHand:
			hands++
		case ArchiveShield:
			shields++
		}
	}
	if hands != 1 || shields != 1 {
		t.Fatalf("單手武器應該是一隻手加一面盾，得到 %d／%d", hands, shields)
	}
}

// 手上那件不是武器（種類 0 或 2）就不畫那一層，盾也照畫。
func TestNonWeaponInHandIsNotDrawn(t *testing.T) {
	l := load(t)
	for _, kind := range []int{0, 2} {
		f := Figure{
			Styles: [LayerCount]int{-1, -1, -1},
			Weapon: &Item{Class: 87, Style: 1, Kind: kind, TwoHand: true},
			Shield: &Item{Style: 0},
		}
		for _, p := range l.Compose(f, 0) {
			if p.Archive == ArchiveWeapon {
				t.Fatalf("種類 %d 不該畫武器層", kind)
			}
		}
	}
}

// 武器組號＝基底表加樣式兩倍再除以二；種類 5 的樣式 2 與 3 對調。
func TestWeaponIndex(t *testing.T) {
	l := load(t)
	for _, c := range []struct {
		class, style, want int
	}{
		{0, 0, 0},                              // 種類 0、樣式 0 → payload 0
		{4 << 5, 1, (l.WeaponBase[4] + 2) / 2}, // 基底加一組
		{5 << 5, 2, (l.WeaponBase[5] + 2*3) / 2},
		{5 << 5, 3, (l.WeaponBase[5] + 2*2) / 2},
	} {
		got, ok := l.WeaponIndex(Item{Class: c.class, Style: c.style})
		if !ok {
			t.Fatalf("class %d style %d 算不出來", c.class, c.style)
		}
		if got != c.want {
			t.Fatalf("class %d style %d 得到 %d，期望 %d", c.class, c.style, got, c.want)
		}
	}
}

// 隊伍頁層次大的先畫，位移照畫的順序取。
func TestPartyOrder(t *testing.T) {
	l := load(t)
	builds := []int{0, 1, 2, 3, 4, 5}
	order, dx := l.PartyOrder(builds)
	if len(order) != len(builds) || len(dx) != len(builds) {
		t.Fatalf("得到 %d／%d 個", len(order), len(dx))
	}
	prev := len(l.PartyLayer) + 1
	for _, i := range order {
		layer := l.PartyLayer[builds[i]]
		if layer > prev {
			t.Fatalf("層次沒有由大到小：%v", order)
		}
		prev = layer
	}
	for n, d := range dx {
		if d != l.PartyX[n] {
			t.Fatalf("第 %d 個的位移是 %d，期望 %d", n, d, l.PartyX[n])
		}
	}
}

// 位移會加到每一層的 X 上，Y 不受影響。
func TestComposeAppliesOffset(t *testing.T) {
	l := load(t)
	f := Figure{Styles: [LayerCount]int{0, 0, 0}}
	base := l.Compose(f, 0)
	moved := l.Compose(f, 7)
	for i := range base {
		if moved[i].X != base[i].X+7 || moved[i].Y != base[i].Y {
			t.Fatalf("第 %d 層沒有照位移移動：%+v → %+v", i, base[i], moved[i])
		}
	}
}
