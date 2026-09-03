// Package battle 是戰鬥這一段的狀態（RE-079）。
//
// 戰場是 64×64 格的磚塊圖，視窗一次看 9×9 格、一格 16×16。參戰者最多 18 名：
// 前六個位置是隊伍、其餘是敵人，站位由任務地點的記錄給（一般遭遇另有五組陣型表，
// 但隨機遭遇還沒實作，見 SPEC-007）。
package battle

import (
	"encoding/json"
	"fmt"
	"os"
)

// FieldSide 是戰場的邊長（格）。緩衝區 4,096 byte ＝ 64 × 64。
const FieldSide = 64

// ViewCols／ViewRows 是戰場視窗一次看幾格；TileSide 是一格的像素邊長。
// 視窗內容區 144×144，剛好九列九欄（`sub_223C0:0x224E0` 起的兩層迴圈
// 都以 9 為界，畫面位移是 `0x280 × 列 + 0x153 + 2 × 欄`，0x153 即 (152, 8)）。
const (
	ViewCols = 9
	ViewRows = 9
	TileSide = 16
)

// ViewCentre 是鏡頭那一格在視窗裡的列／欄：兩軸都減四（`0x224F6` 的 `+0FFFCh`）。
var ViewCentre = struct{ Row, Col int }{Row: 4, Col: 4}

// FieldCursor 是目前輪到的那一名外面那一圈白框
// （`sub_223C0:0x2279E` 的 `sub_1B6DC(0D8h, 48h, 0E5h, 57h, 0Fh)`）。
var FieldCursor = struct{ X, Y, Width, Height int }{X: 216, Y: 72, Width: 14, Height: 16}

// 參戰者畫成哪一塊磚（`sub_223C0:0x2264D`…`0x22682`，畫的時候是
// 「精靈編號 − 1」）。隊伍前六名各有自己的號碼圖，敵人共用一張。
const (
	// SpritePartyBase 是隊伍第一名的磚塊編號，之後每名加一。
	SpritePartyBase = 0
	// SpriteMonster 是還站著的敵人；SpriteMonsterDown 是倒下的敵人。
	SpriteMonster     = 6
	SpriteMonsterDown = 7
	// SpritePartyDown 是倒下的隊員。
	SpritePartyDown = 8
	// SpriteObject 是地上的東西（原版畫的是一把劍）。
	SpriteObject = 38
)

// MaxActors 是參戰者上限；PartySlots 是其中留給隊伍的位置數
// （`sub_1E3A8` 清的是 18 × 50 byte 的陣列，隊伍固定占前六個）。
const (
	MaxActors  = 18
	PartySlots = 6
)

// Side 是參戰者的陣營，對應記錄位移 +0x01。
type Side int

const (
	SideParty Side = iota
	SideMonster
	SideEmpty
)

// Actor 是一名參戰者。欄位名對應原版記錄的位移：
// Kind 是 +0x00（隊員填 0x13）、Side 是 +0x01、X／Y 是 +0x1F／+0x20。
type Actor struct {
	Kind  int
	Side  Side
	X, Y  int
	Name  string
	Level int
	// Member 是隊伍那一側對應到哪一個隊員；敵人是 −1。
	Member int
}

// PartyKind 是隊員在記錄 +0x00 填的值（`sub_1E3A8:0x1E78D`）。
const PartyKind = 0x13

// Site 是一個任務地點，欄位與 `tools/extract_quest_sites.py` 的輸出同名。
type Site struct {
	Index       int      `json:"index"`
	X           int      `json:"x"`
	Y           int      `json:"y"`
	Monster     int      `json:"monster"`
	MonsterName string   `json:"monster_name"`
	Level       int      `json:"level"`
	Count       int      `json:"count"`
	QuestByte   int      `json:"quest_byte"`
	QuestDiv    int      `json:"quest_divisor"`
	MonsterX    []int    `json:"monster_x"`
	MonsterY    []int    `json:"monster_y"`
	PartyX      []int    `json:"party_x"`
	PartyY      []int    `json:"party_y"`
	Narrative   []string `json:"narrative"`
}

// Sites 是抽取器產生的任務地點表。
type Sites struct {
	Schema string `json:"schema"`
	List   []Site `json:"sites"`
}

// LoadSites 讀取任務地點的數值表。
func LoadSites(path string) (*Sites, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("讀取任務地點資料：%w", err)
	}
	var s Sites
	if err := json.Unmarshal(body, &s); err != nil {
		return nil, fmt.Errorf("解析任務地點資料：%w", err)
	}
	if s.Schema != "kol-quest-site-data-v1" {
		return nil, fmt.Errorf("任務地點資料 schema 不符：%q", s.Schema)
	}
	return &s, nil
}

// At 回傳世界座標上的任務地點；沒有就回 nil。
func (s *Sites) At(x, y int) *Site {
	for i := range s.List {
		if s.List[i].X == x && s.List[i].Y == y {
			return &s.List[i]
		}
	}
	return nil
}

// Taken 回報這一份任務旗標有沒有帶著這個地點的任務。
// 原版是「`[0x1E2 + i / 4]` 除以 `(1, 4, 16, 64)[i & 3]` 之後低兩位是 1」
// （`sub_2B0FF:0x2B127`…`0x2B153`）。旗標是角色記錄 0x1E2 起的六個 byte。
func (s *Site) Taken(quests []int) bool {
	index := s.QuestByte - QuestFieldBase
	if index < 0 || index >= len(quests) || s.QuestDiv == 0 {
		return false
	}
	return (quests[index]/s.QuestDiv)&3 == 1
}

// QuestFieldBase 是任務旗標在角色記錄裡的起點。
const QuestFieldBase = 0x1E2
