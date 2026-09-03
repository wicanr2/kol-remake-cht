package town

import "testing"

// theodoric 是原版存檔第 0 格的角色，欄位直接抄自 CHARDATA。
func theodoric() *Sheet {
	return &Sheet{
		Stats:           [StatCount]int{130, 100, 116, 102, 103, 94, 109},
		Race:            3,
		Profession:      36,
		Rank:            1,
		AdventurePoints: 674,
		MovementRate:    0,
		Nutrition:       2,
		HeightFeet:      7, HeightInches: 4, Weight: 296,
		Weapons: []WeaponRow{{Name: "GREATSWORD", Offence: 10, Defence: 10}},
	}
}

// 八個畫面值是內部值 × 100 / 150；平衡與耐力是兩個平均。
func TestSheetDisplayedStats(t *testing.T) {
	got := theodoric().Displayed()
	want := [8]int{86, 66, 68, 68, 62, 72, 79, 77}
	if got != want {
		t.Fatalf("畫面值是 %v，原版畫面上是 %v", got, want)
	}
}

// 身高體重的字串格式與原版相同。
func TestSheetHeightAndWeight(t *testing.T) {
	s := theodoric()
	if got := s.Height(); got != `7' 4"` {
		t.Fatalf("身高是 %q", got)
	}
	if got := s.WeightText(); got != "296 LBS" {
		t.Fatalf("體重是 %q", got)
	}
}

// 移動率 0 當成 20，印出來是 `U`。
func TestSheetMovementLetter(t *testing.T) {
	s := theodoric()
	if got := s.MovementLetter(); got != "U" {
		t.Fatalf("移動率是 %q", got)
	}
	s.MovementRate = 1
	if got := s.MovementLetter(); got != "B" {
		t.Fatalf("移動率 1 應該是 B，得到 %q", got)
	}
}

// 長條格數是值右移三位，非 0 但不足一格算一格。
func TestBarCells(t *testing.T) {
	cases := map[int]int{0: 0, 1: 1, 2: 1, 7: 1, 8: 1, 16: 2, 20: 2, 24: 3}
	for v, want := range cases {
		if got := BarCells(v); got != want {
			t.Fatalf("值 %d 應該是 %d 格，得到 %d", v, want, got)
		}
	}
}

// 職業名稱的記錄：矮人與凱爾登、女性精靈用低六號的那一筆。
func TestProfessionRecord(t *testing.T) {
	cases := []struct {
		profession, race, gender, want int
	}{
		{36, 3, 0, 36}, // 凱爾登男性：36 + 6 − 6
		{36, 0, 0, 42}, // 人類男性：不減
		{36, 1, 0, 42}, // 男性精靈：不減
		{36, 1, 1, 36}, // 女性精靈：減
		{36, 2, 0, 36}, // 矮人：減
		{1, 3, 0, 7},   // 記錄號 ≤ 0x0B 一律原樣
	}
	for _, c := range cases {
		if got := professionRecord(c.profession, c.race, c.gender); got != c.want {
			t.Fatalf("職業 %d／種族 %d／性別 %d 應該是記錄 %d，得到 %d",
				c.profession, c.race, c.gender, c.want, got)
		}
	}
}

// 標籤記錄的座標語法：`~V` 換列、`~H` 設左緣、`@` 回到左緣並往下一行。
func TestParseSegments(t *testing.T) {
	got := ParseSegments("~V3A~H02STRENGTH :@~V42QUICKNESS:@HEALTH   :")
	want := []Segment{
		{Col: 2, Row: 0x3A, Text: "STRENGTH :"},
		{Col: 2, Row: 0x42, Text: "QUICKNESS:"},
		{Col: 2, Row: 0x42 + LineStep, Text: "HEALTH   :"},
	}
	if len(got) != len(want) {
		t.Fatalf("拆出 %d 段：%+v", len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("第 %d 段是 %+v，應為 %+v", i+1, got[i], want[i])
		}
	}
}

// 按下 MIRROR 換成資料頁，選 U-TURN 回到建築內。
func TestMirrorOpensAndCloses(t *testing.T) {
	s := innSession(t, 56, Member{Name: "Theodoric", Gold: 1009,
		Portrait: 47, Speaker: true, Sheet: theodoric()})
	s.Buildings.Doors[0].IconFlags = [2]int{0x09, 0x00}
	s.OpenSheet()
	if !s.ShowingSheet() {
		t.Fatal("按下 MIRROR 應該換成角色資料頁")
	}
	page := s.Inside.Sheet
	if page.Name != "THEODORIC" || page.Gold != 1009 || page.Adventure != 674 {
		t.Fatalf("資料頁是 %+v", page)
	}
	if page.NutritionCells != 1 || page.DamageCells != 0 {
		t.Fatalf("長條格數是 %d／%d", page.DamageCells, page.NutritionCells)
	}
	icons, _ := s.Icons()
	if len(icons) != 3 || icons[0] != IconSprite(IconUTurn) {
		t.Fatalf("資料頁的圖示列是 %v", icons)
	}
	s.Inside.Selected = 0
	s.Accept()
	if s.ShowingSheet() {
		t.Fatal("選 U-TURN 應該回到建築內")
	}
}

// 中文模式還沒有資料頁的版面，選 MIRROR 只記一筆，不開出排錯的畫面。
func TestMirrorSkippedInCJK(t *testing.T) {
	s := innSession(t, 56, Member{Name: "Theodoric", Speaker: true, Sheet: theodoric()})
	s.CJK = true
	s.OpenSheet()
	if s.ShowingSheet() {
		t.Fatal("中文模式不該開資料頁")
	}
	if len(s.Unimplemented) != 1 {
		t.Fatalf("未實作清單是 %v", s.Unimplemented)
	}
}
