package battle

import (
	"encoding/json"
	"fmt"
	"os"
)

// 戰後的戰利品（RE-096）。原版在報完冒險點數與金冠之後呼叫 `sub_289EE(0)`：
// 前半擲骰決定掉什麼、後半把 `0x9D20` 那十六格排成清單讓玩家一件一件領走。

// SpoilChance 是掉東西的門檻（`0x28A00`）：`rand(100)` 小於它才掉。亂數回 1…100，
// 所以實際上是十九分之一百。
const SpoilChance = 20

// ItemTableCount 是 `ITEM.LST` 的筆數（976 byte ÷ 16）。第 0 筆是全零的空位，
// 抽中它等於什麼都沒掉——原版照樣把那 16 個零複製進槽裡，而開頭為零的槽在清單
// 那一關會被當成空的跳過。
const ItemTableCount = 61

// 戰利品畫面的訊息記錄（KOLDATA）。
const (
	ItemsFoundRecord = 0x181 // "~H02~V09Items Found:"
	CharactersRecord = 0x182 // "~H02~V09Characters:"
	// EmptyHandsRecord 是背包塞不下、改用手上那一格時印的那一句。
	EmptyHandsRecord = 0x183 // "You empty your hands and pick up the item."
)

// BackpackSlots 是背包格數：角色記錄 0xA0 起、一格 16 byte、共六格（RE-075）。
// 與城鎮那一側的 `town.BackpackSlots` 是同一個東西。
const BackpackSlots = 6

// SpoilItem 是 `ITEM.LST` 裡的一筆。名字空的就是那個空位。
type SpoilItem struct {
	Index int  `json:"index"`
	Empty bool `json:"empty"`
	// Item 帶著名字與裝備欄位，與怪物身上那幾格同一個型別——掉在地上的東西被
	// 撿起來要換算成傷害骰就靠它（`CombatWeapon`）。`ITEM.LST` 的 JSON 是平的，
	// 所以這裡內嵌而不是另開一層。
	Item
}

// DropSpoils 擲出這一場掉了什麼，放進地面那十六格，回傳新占用的格號。
//
// 原版的順序（`0x28A00`…`0x28B18`）：先擲一次 `rand(100)`，不到門檻整段跳過；
// 過了就擲 `rand(敵人數)` 決定**掉幾件**——亂數回 1…n，所以過了門檻至少掉一件。
// 每一件掃十六格找第一個空的，複製整筆記錄進去，再把 +0x0B（狀況）覆寫成
// `rand(255)`；十六格滿了那一件就丟掉。
//
// **掉落物沒有座標。**原版一個字都沒寫進那兩張座標表——它們不落在地上，領取是靠
// 後面那張清單，不是走過去撿。
func (s *State) DropSpoils(items []SpoilItem, rng Rand) []int {
	if len(items) == 0 || rng.Roll(100) >= SpoilChance {
		return nil
	}
	enemies := s.EnemyCount()
	if enemies <= 0 {
		return nil
	}
	var got []int
	for k := rng.Roll(enemies); k > 0; k-- {
		slot := s.emptyGroundSlot()
		if slot < 0 {
			break
		}
		it := items[(rng.Roll(ItemTableCount)-1)%len(items)]
		// 狀況那一格照樣擲，即使抽中的是空位——原版沒有為此分支。
		rng.Roll(255)
		if it.Empty || it.Name == "" {
			continue
		}
		s.Ground[slot] = GroundItem{Name: it.Name, Weapon: it.CombatWeapon()}
		got = append(got, slot)
	}
	return got
}

func (s *State) emptyGroundSlot() int {
	for n := range s.Ground {
		if s.Ground[n].Name == "" {
			return n
		}
	}
	return -1
}

// SpoilSlots 是戰利品清單：地面那十六格裡還有東西的格號，由小到大。
//
// 戰鬥中丟在地上沒撿回來的東西也在裡面——`0x9D20` 打完之後就是戰利品池，原版沒有
// 把兩者分開。清單空的時候整個畫面跳過。
func (s *State) SpoilSlots() []int {
	var out []int
	for n := range s.Ground {
		if s.Ground[n].Name != "" {
			out = append(out, n)
		}
	}
	return out
}

// SpoilTakers 是能收下戰利品的隊員（參戰者索引），由小到大。
//
// 判準是 `sub_22866`：頭或軀幹的血量歸零就不算數——倒下的人領不到東西。空格子
// 同樣跳過。
func (s *State) SpoilTakers() []int {
	var out []int
	for i := 0; i < PartySlots && i < len(s.Actors); i++ {
		if s.Actors[i].Side != SideParty || s.Actors[i].Down() {
			continue
		}
		out = append(out, i)
	}
	return out
}

// TakeSpoil 把一格戰利品交給一位隊員，回傳那件東西的名字與要印的訊息記錄
// （沒有話要說就是 0）。
//
// 原版的順序（`0x28E5A`…`0x28F4C`）：先找背包六格的第一個空格；六格都滿了就改用
// 手上那一格。放法是**十六個 byte 交換**，不是搬移——所以背包滿的時候，手上原本
// 那件會被換到地上那一格去，而且只有那時才印記錄 0x183。
//
// 領完之後原版跳回清單開頭（`0x28F4F jmp loc_28B1E`），所以一場可以領很多件。
func (s *State) TakeSpoil(slot, actor int) (string, int) {
	if slot < 0 || slot >= len(s.Ground) || s.Ground[slot].Name == "" {
		return "", 0
	}
	m := s.member(actor)
	if m == nil {
		return "", 0
	}
	g := s.Ground[slot]
	name := g.Name
	for n := range m.Pack {
		if m.Pack[n] == "" {
			m.Pack[n] = name
			s.Ground[slot] = GroundItem{}
			return name, 0
		}
	}
	// 背包滿了：與手上那一格交換。手上空著也照樣換，換過去的是空的。
	record := 0
	if m.HasWeapon {
		record = EmptyHandsRecord
	}
	if a := actor; a >= 0 && a < len(s.Actors) {
		s.Ground[slot] = GroundItem{Name: m.WeaponName, Weapon: s.Actors[a].Weapon}
		s.Actors[a].Weapon = g.Weapon
	} else {
		s.Ground[slot] = GroundItem{Name: m.WeaponName}
	}
	m.HasWeapon, m.WeaponName, m.Worn[0] = true, name, name
	return name, record
}

// SpoilTable 是 `ITEM.LST` 抽出來的物品表（`tools/extract_items.py`）。
type SpoilTable struct {
	Schema string      `json:"schema"`
	Items  []SpoilItem `json:"items"`
}

// LoadSpoilItems 讀取掉落物品表。
func LoadSpoilItems(path string) ([]SpoilItem, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("讀取掉落物品表：%w", err)
	}
	var t SpoilTable
	if err := json.Unmarshal(body, &t); err != nil {
		return nil, fmt.Errorf("解析掉落物品表：%w", err)
	}
	if t.Schema != "kol-battle-items-v1" {
		return nil, fmt.Errorf("掉落物品表 schema 不符：%q", t.Schema)
	}
	if len(t.Items) != ItemTableCount {
		return nil, fmt.Errorf("掉落物品表有 %d 筆，期望 %d 筆",
			len(t.Items), ItemTableCount)
	}
	return t.Items, nil
}

// SpoilBox 是戰利品清單的框（`0x28B75` 的 `sub_1BA9D(4, 7, 80h, 98h, 0Fh)`）。
// 那一支收的是左上與右下兩點，寬高各要加一——與 SheetBox／BodyBox 同一套算法。
var SpoilBox = struct{ X, Y, Width, Height int }{X: 4, Y: 7, Width: 125, Height: 146}

// SpoilList 是清單第一行的位置與行距，來自 `sub_1AAE8(1, 1, 13h, n, n, 2E3Ah, 0Dh)`
// 的第二與第三個參數（RE-115）：行 1（x = 8）、列 0x13。行距是那一支自己算的
// `6 + 2 × word_2DF18`，而戰利品畫面進來先把 `word_2DF18` 設成 1（大字），所以是 8。
//
// 第七個參數的 13 欄只用來擺捲動箭頭，而這張清單的可見列數等於總筆數（最多十六件，
// 十六列 × 8 放得進框裡），所以**永遠不捲動、不畫箭頭**，寬度在這裡沒有作用。
//
// **還沒對拍**：原版收據裡沒有戰利品畫面（要打贏一場而且擲中兩成的掉落），
// 這一組是靜態讀出來的。
var SpoilList = struct{ Col, Row, LineStep int }{Col: 1, Row: 0x13, LineStep: 8}

// SpoilPhase 是戰利品畫面停在哪一關。
type SpoilPhase int

const (
	// SpoilPickItem 是「Items Found:」那張清單。
	SpoilPickItem SpoilPhase = iota
	// SpoilPickTaker 是「Characters:」那張清單。
	SpoilPickTaker
)

// SpoilIcons 是清單底下那條圖示列：原版在開清單之前先 `sub_1449C()` 把整條清空，
// 再 `sub_11703(7)`、`sub_11703(0x0F)` 加兩格，`sub_11722(1)` 畫出來並把白框放在
// 第 0 格（`0x28C21`／`0x28DCC`，RE-115）。編號與 `town.IconYes`／`town.IconUTurn`
// 相同——商店的「要不要買」用的也是這兩格。battle 不認得 town 的型別，所以這裡
// 自己寫一份。
var SpoilIcons = []int{7, 15}

// SpoilIconUTurn 是 SpoilIcons 裡取消那一格的位置。
const SpoilIconUTurn = 1

// SpoilScreen 是戰利品畫面的狀態。
//
// 原版把兩張清單都交給同一支選單（`sub_1AAE8`），回傳值的低 byte 是絕對列號、
// 高 byte 是白框那一格的圖示編號；圖示 0x0F（U-TURN）與 ESC 都是取消（RE-115）。
// remake 把游標、白框與兩張清單留在這裡。
type SpoilScreen struct {
	Phase SpoilPhase
	// Slots 是物品清單（地面格號）；Takers 是隊員清單（參戰者索引）。
	Slots  []int
	Takers []int
	// Cursor 是目前這張清單的游標。
	Cursor int
	// Icon 是圖示列白框停在第幾格，索引進 SpoilIcons。
	Icon int
	// Slot 是在第一關選好的那一格。
	Slot int
	// Say 是領完之後要說的那一句（只有背包塞不下時才有，記錄 0x183）。
	Say int
}

// BeginSpoils 進戰利品畫面，回傳還有沒有東西可領。
//
// 清單是空的就沒有這個畫面（`0x28B72`）。領完一件之後原版跳回這裡重新掃一次
// （`0x28F4F jmp loc_28B1E`），所以同一支也用來「回到清單」。
func (s *State) BeginSpoils() bool {
	slots := s.SpoilSlots()
	if len(slots) == 0 {
		return false
	}
	s.Spoil = SpoilScreen{Phase: SpoilPickItem, Slots: slots, Slot: -1}
	return true
}

// SpoilMove 移動游標，超出兩端就停住。
func (s *State) SpoilMove(d int) {
	n := len(s.spoilList())
	if n == 0 {
		return
	}
	s.Spoil.Cursor += d
	if s.Spoil.Cursor < 0 {
		s.Spoil.Cursor = 0
	}
	if s.Spoil.Cursor >= n {
		s.Spoil.Cursor = n - 1
	}
}

// SpoilIconMove 移圖示列的白框。原版是 `,`／`<` 與 `.`／`>`，兩端繞圈
// （`sub_1AAE8:0x1AF59`，RE-115）。
func (s *State) SpoilIconMove(d int) {
	n := len(SpoilIcons)
	s.Spoil.Icon = ((s.Spoil.Icon+d)%n + n) % n
}

// SpoilAccept 按下 Return：白框停在 U-TURN 就是取消，否則領走反白那一列。
// 回傳與 SpoilConfirm 相同的三個值。
func (s *State) SpoilAccept() (name string, record int, done bool) {
	if s.Spoil.Icon == SpoilIconUTurn {
		return "", 0, s.SpoilCancel()
	}
	return s.SpoilConfirm()
}

func (s *State) spoilList() []int {
	if s.Spoil.Phase == SpoilPickTaker {
		return s.Spoil.Takers
	}
	return s.Spoil.Slots
}

// SpoilCancel 取消目前這一關，回傳整個畫面是不是結束了。
//
// 物品那一關取消就結束（`0x28C9C`）；隊員那一關取消只是退回物品清單
// （`0x28E47` 同樣跳回 `0x28B1E`）。
func (s *State) SpoilCancel() bool {
	if s.Spoil.Phase == SpoilPickTaker {
		return !s.BeginSpoils()
	}
	return true
}

// SpoilConfirm 按下確認。回傳領到的東西、要印的訊息記錄，以及整個畫面是不是結束了。
//
// 在物品那一關按確認會挑出能收下的隊員：一個都沒有就結束（`0x28D11`）、
// 只有一個就直接給他（`0x28D1D`）、多於一個才進第二張清單。
func (s *State) SpoilConfirm() (name string, record int, done bool) {
	list := s.spoilList()
	if s.Spoil.Cursor < 0 || s.Spoil.Cursor >= len(list) {
		return "", 0, true
	}
	if s.Spoil.Phase == SpoilPickTaker {
		name, record = s.TakeSpoil(s.Spoil.Slot, list[s.Spoil.Cursor])
		done = !s.BeginSpoils()
		s.Spoil.Say = record
		return name, record, done
	}
	s.Spoil.Slot = list[s.Spoil.Cursor]
	takers := s.SpoilTakers()
	switch len(takers) {
	case 0:
		return "", 0, true
	case 1:
		name, record = s.TakeSpoil(s.Spoil.Slot, takers[0])
		done = !s.BeginSpoils()
		s.Spoil.Say = record
		return name, record, done
	}
	s.Spoil.Phase, s.Spoil.Takers, s.Spoil.Cursor = SpoilPickTaker, takers, 0
	return "", 0, false
}

// SpoilLines 是目前這張清單要顯示的文字，順序同游標。
func (s *State) SpoilLines() []string {
	out := make([]string, 0, len(s.spoilList()))
	for _, i := range s.spoilList() {
		if s.Spoil.Phase == SpoilPickTaker {
			out = append(out, s.actorName(i))
			continue
		}
		out = append(out, s.Ground[i].Name)
	}
	return out
}

// SpoilTitle 是目前這一關的標題記錄。
func (s *State) SpoilTitle() int {
	if s.Spoil.Phase == SpoilPickTaker {
		return CharactersRecord
	}
	return ItemsFoundRecord
}

func (s *State) actorName(i int) string {
	if m := s.member(i); m != nil {
		return m.Name
	}
	if i >= 0 && i < len(s.Actors) {
		return s.Actors[i].Name
	}
	return ""
}
