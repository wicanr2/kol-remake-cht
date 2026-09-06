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
	Kind int
	Side Side
	X, Y int
	Name string
	// Level 是參戰者記錄 +0x02 的等級：隊員是四列武器攻防值加總 ÷ 1000 ＋ 1，
	// 敵人是怪物等級 ÷ 5 ＋ 1（上限 5，`0x1EE12`）。規劃鍵值與敵人 AI 的權重都用
	// 這一格。**怪物本身的等級在 `State.MonsterLevel`**，那是另一個尺度（RE-097）。
	Level int
	// Strength／Quickness／Foresight／Size 是記錄 +0x08…+0x0B：力量已依營養折算過
	// （`sub_2198A:0x219CF`），快捷用在防禦、洞察用在規劃順序、體型用在昏迷門檻。
	Strength  int
	Quickness int
	Foresight int
	Size      int
	// Attack／Parry 是手上武器那一列的攻防技能（`sub_26AC2`），敵人已折算過等級。
	Attack int
	Parry  int
	// BaseAttack／BaseParry 是怪物記錄 +0x17／+0x19 的原值，GoldRoll 是 +0x0F 的
	// 骰式。三個都只有戰後結算用得到，而結算讀的是記錄本身、不是折算後的技能
	// （RE-096）。隊員那一側留 0。
	BaseAttack, BaseParry, GoldRoll int
	// Weapon 是手上那件武器的數字；Encumbrance 是負重、ShieldGrade 是盾牌等級
	// （NoShield 表示沒有盾牌）。
	Weapon      Weapon
	Encumbrance int
	ShieldGrade int
	// WeaponRows 是四列武器技能（隊員專用）。換武器要重選那一列，
	// 所以整份帶著（`ReloadSkills`）。怪物那一側的技能在本體記錄 `+0x17`，
	// 與手上那件無關。
	WeaponRows []WeaponSkillRow
	// LoadLevel、ArmourWeight、ShieldWeight 是重算負重與武器速度要用的三個輸入
	// （`ReloadGear`）。**LoadLevel 不是 `Level`**：隊員這一格是技能推出來的戰鬥
	// 等級，怪物是遭遇等級，而 `Level` 那一格對怪物另外除以 5 再加一。
	LoadLevel    int
	ArmourWeight int
	ShieldWeight int
	// UnarmedWeapon 是徒手各式拿的那一格穿戴品（`UnarmedSlots`）。原版徒手攻擊
	// 的傷害骰、種類與重量都從那一格來，不是手上那件（`sub_26AC2:0x26BDA`）。
	UnarmedWeapon [6]Weapon
	// Unarmed 是「手上那一格空著」：怪物記錄 +0x20 那筆 16 byte 的物品記錄第一個
	// byte 是 0（`0x23BC2`，RE-116）。敵人 AI 拿它決定要不要先去撿一件（RE-116）。
	// 隊員那一側由 `Member.HasWeapon` 表示，這一格只給敵人用。
	Unarmed bool
	// PickedUp 是參戰者記錄 +0x2B：撿到手上那一格之後由執行段清成 0
	// （`sub_1E3A8:0x203E0`）。原版沒有任何地方把它設成非零，所以「沒有武器就
	// 改成移動」那一條分支走不到——照樣實作，理由記在 RE-116。
	PickedUp int
	// WeaponSpeed 是 `sub_24FD3` 算出來的 `word_37CF0`：執行段的鍵值要用它
	// 折算指令的時間成本（RE-084）。
	WeaponSpeed int
	// Hurt 是 +0x2D：**從上一次輪到自己之後**挨的傷（`sub_26D76:0x2762E` 累加），
	// 攻擊分數會扣掉它的 20 倍。Parries 是已經招架幾次（+0x0F），愈多擋得愈差。
	// 兩個都在一手執行完的收尾清掉（`sub_1E3A8:0x20E64`）。
	//
	// **它不是疲勞**——疲勞是另一格（`Fatigue`），兩者無關（RE-130）。
	Hurt    int
	Parries int
	// Fatigue 是 `[0x322E + 2i]`：`MAGIC.0` 屬性表第 5 項的疲勞。開打歸零，
	// 只有法術效果 5 會加，休息減三成。Stamina 是 `[0x1C50 + 2i]`＝力量 ＋ 健康，
	// 折算成比例時當分母（`sub_23683`，RE-129／RE-130）。
	//
	// **和累積傷害是兩個獨立的累計**：傷害落到身上那一支寫的是 `+0x0E`
	// （`Damage`），全庫沒有任何地方把傷害加進疲勞。
	Fatigue int
	Stamina int
	// Shatters 是怪物記錄 +0x10 的 bit 7：力竭倒下時印「碎成一地」而不是
	// 「失血而死」（RE-131）。隊員那一側用不到。
	Shatters bool
	// LastHurt 是 +0x2A：上一手的 Fatigue，由那個收尾搬過來。敵人 AI 拿它判
	// 「已經在撤退、而且這一手還在挨打」（RE-081）。
	LastHurt int
	// Retreating 是 +0x29：撤退狀態。傷勢達到 FleeAt 又沒有人貼著就設起來，
	// 降到 FleeRecover 以下就清掉。
	Retreating bool
	// Stuck 是 `word_34A08`：這一次規劃裡「八個方向都走不了」的次數（RE-117）。
	// 非零時貼近打那一段不再抽攻擊式，固定用第 1 式。`PlanEnemy` 開頭歸零，
	// 所以它不跨回合。
	Stuck int
	// FleeAt／FleeRecover 是怪物記錄的兩個撤退門檻（+0x0D／+0x0E）。隊員那一側
	// 留 0——原版的撤退判斷只跑在敵人身上。
	FleeAt, FleeRecover int
	// Intellect 是記錄 +0x0C。目前只有法術的效果 4 會動它（RE-100）。
	Intellect int
	// SpellPick 是選中的法術（+0x26，0…15）；Negative 是這一手要傷害還是治療
	// （+0x27，RE-100）；SpellTurns 是身上的法術還剩幾回合（+0x30），
	// 倒數到 0 就把屬性重算回原值並印記錄 0x173。
	SpellPick  int
	Negative   bool
	SpellTurns int
	// Ward 是擋法術的那一格：敵人取怪物記錄 +0x1D、隊員取角色記錄 +0x28
	// （`sub_22229` 決定看哪一份，RE-100）。
	Ward int
	// SkillPenaltyOff／SkillPenaltyDef 是記錄 +0x2E／+0x2F：法術效果 7、8
	// （`Off. Skill`／`Def. Skill`）寫的那兩格。攻防技能每次要用都由 `sub_26AC2`
	// 從原始記錄重算一遍，最後 `sub_26CD5` 把這兩格的**十倍**扣掉、夾在 0 以上，
	// 所以它們是罰值而不是技能本身（RE-128）。法術過期時把屬性搬回來的那兩支
	// （`sub_2198A`／`sub_2215A`）只寫 +0x07…+0x0C，碰不到這裡——**罰值不會隨
	// 法術過期還回來**，一整場都算數。
	SkillPenaltyOff, SkillPenaltyDef int
	// Style／Guard／Command 是這一回合選的攻擊選項（+0x23）、防禦選項（+0x24）
	// 與指令碼（+0x25）。三個都由規劃段填：隊員從指令列選、敵人由 AI 填（RE-081）。
	Style   int
	Guard   int
	Command int
	// Facing 是 +0x28 的方向 0…7（上、右上、右、右下、下、左下、左、左上）。
	Facing int
	// Flag1D 是記錄 +0x1D：**已上弦**（RE-093）。裝填那一手做完設 1、射出去清 0，
	// 主指令列推出射擊圖示時也設 1；敵人不必先裝填（RE-104）。
	// Flag1E 是 +0x1E：**騎在坐騎上**（RE-093）。移動方式選到第 3 式以後設起來
	// （`sub_26384`），非零時整組近戰與移動選項換掉，視線也改用騎乘那一份。
	Flag1D int
	Flag1E int
	// Flag03 是記錄 +0x03，只在移動方式選單裡當門檻用。它與 +0x04…+0x06 是
	// `sub_21FCD` 依裝備賦予的四格，各自是什麼還沒解（RE-081）。
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
	// QuestItem 是這個任務要的那一件東西（記錄 +0x29 起十六個 byte，RE-113）。
	// 帶著它逃出戰場就算完成，不必把怪物打完（RE-103）。
	QuestItem QuestItem `json:"quest_item"`
}

// QuestItem 是任務要的那一件東西，以及它擺在戰場哪一格。
//
// `Spot` 來自記錄的 `+0x27`（列）與 `+0x28`（欄）：`sub_2B0FF:0x2B1F0` 在任務物品
// 那十六個 byte 的第 0 格非零時，把它們抄進 `byte_30C6E`／`byte_30A7E`——那兩個是
// **戰場物品第 0 件的列與欄**，各自是一張陣列的第 0 格（`0x2D9E + i`／`0x2BAE + i`）。
// 讀的地方是 `sub_223C0:0x226C5` 起那一段：掃全部物品，欄與列都對上行動者站的
// 那一格才撿得到（RE-128 那一輪解出，更正 RE-113 的「被誰讀還沒解」）。
type QuestItem struct {
	// Item 帶著名字、種類與傷害骰——撿起來的人要拿它去打，只留名字的話那幾個
	// 數字全都沒有。記錄的版面與商店的商品相同（RE-113）。
	Item
	Spot Spot `json:"spot"`
}

// Spot 是戰場上的一格。戰場是 64 × 64，二十四個地點的值全部落在範圍內。
type Spot struct {
	X int `json:"x"`
	Y int `json:"y"`
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
