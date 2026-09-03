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

// Actor 是一名參戰者。欄位名對應原版記錄的位移（RE-080）：
// Kind 是 +0x00（隊員填 0x13）、Side 是 +0x01、X／Y 是 +0x1F／+0x20。
type Actor struct {
	Kind  int
	Side  Side
	X, Y  int
	Name  string
	Level int
	// Strength／Quickness／Foresight／Size 是記錄 +0x08…+0x0B：力量已依營養折算過
	// （`sub_2198A:0x219CF`），快捷用在防禦、洞察用在規劃順序、體型用在昏迷門檻。
	Strength  int
	Quickness int
	Foresight int
	Size      int
	// Attack／Parry 是手上武器那一列的攻防技能（`sub_26AC2`）。
	Attack int
	Parry  int
	// Weapon 是手上那件武器的數字；Encumbrance 是負重、ShieldGrade 是盾牌等級
	// （NoShield 表示沒有盾牌）。
	Weapon      Weapon
	Encumbrance int
	ShieldGrade int
	// WeaponSpeed 是 `sub_24FD3` 算出來的 `word_37CF0`：執行段的鍵值要用它
	// 折算指令的時間成本（RE-084）。
	WeaponSpeed int
	// Fatigue 是 +0x2D 的挨打累計；Parries 是這一回合已經招架幾次（+0x0F）。
	Fatigue int
	Parries int
	// Style／Guard／Command 是這一回合選的攻擊選項（+0x23）、防禦選項（+0x24）
	// 與指令碼（+0x25）。三個都由規劃段填：隊員從指令列選、敵人由 AI 填（RE-081）。
	Style   int
	Guard   int
	Command int
	// Facing 是 +0x28 的方向 0…7（上、右上、右、右下、下、左下、左、左上）。
	Facing int
	// Flag1D／Flag1E 是記錄 +0x1D／+0x1E。組指令列與選單都要看它們，
	// 但語意還沒定（SPEC-010 末段）：+0x1D 在推出射擊圖示時會被設成 1，
	// +0x1E 非零時換掉整組近戰與移動選項。
	Flag1D int
	Flag1E int
	// Flag03 是記錄 +0x03，只在移動方式選單裡當門檻用；語意同樣未定。
	Flag03 int
	// Threshold 是 +0x10 的昏迷門檻，Damage 是 +0x0E 的累積傷害。
	Threshold int
	Damage    int
	// Head…LeftLeg 是六個部位剩下的血量，順序同命中部位（+0x11 起每兩個 byte 一組）。
	Head, Chest, RightArm, LeftArm, RightLeg, LeftLeg int
	// Member 是隊伍那一側對應到哪一個隊員；敵人是 −1。
	Member int
}

// SetParts 依昏迷門檻填好六個部位的血量。
func (a *Actor) SetParts() {
	p := Parts(a.Threshold)
	a.Head, a.Chest = p[0], p[1]
	a.RightArm, a.LeftArm = p[2], p[3]
	a.RightLeg, a.LeftLeg = p[4], p[5]
}

// 穿戴欄的部位編號（RE-080）：手上那件武器在部位 9，三件護甲是頭、軀幹、腿，
// 盾牌是部位 6。負重只算這幾件。
const (
	WeaponSlot = 9
	ShieldSlot = 6
)

// ArmourSlots 是負重要算的三件護甲（頭、軀幹、腿）。
var ArmourSlots = [3]int{0, 1, 2}

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
