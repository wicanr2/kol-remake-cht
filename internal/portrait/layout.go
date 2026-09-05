package portrait

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// StylesPerBuild 是一個體型組合有幾套外觀（原版的 `9 × 體型組合`）。
const StylesPerBuild = 9

// SwappedWeaponKind 是那個樣式 2 與 3 對調的武器種類（原版 `xor 1` 只對它做）。
const SwappedWeaponKind = 5

// Cell 是一層的貼圖矩形。X 與 W 的單位是 byte 欄，Y 與 H 是像素。
type Cell struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

// BodyCell 是體型那一層：原版直接給兩個角落，不是位置加尺寸。
type BodyCell struct {
	X  int `json:"x"`
	Y  int `json:"y"`
	X2 int `json:"x2"`
	Y2 int `json:"y2"`
}

// Layout 是整份版面表，由 `tools/extract_portrait_layout.py` 從 KOLTOWN 的資料段抽出。
type Layout struct {
	Schema string `json:"schema"`
	// Body 一個體型組合一格；Layers 是 6 × 9 套外觀，每套三層。
	Body   []BodyCell `json:"body"`
	Layers [][3]Cell  `json:"layers"`
	// Weapons 與 Shields 的位置按體型分，尺寸不分（抽取器已經驗過這件事）。
	Weapons [][]Cell `json:"weapons"`
	Shields [][]Cell `json:"shields"`
	Hands   []Cell   `json:"hands"`
	// WeaponBase 是武器種類 → WEAPPICS 的起始 payload 索引。
	WeaponBase []int `json:"weapon_base"`
	// PartyLayer 是體型組合在隊伍頁的層次（6 最遠、1 最近）；
	// PartyX 是第 n 個畫的人往右挪幾個 byte 欄。
	PartyLayer []int `json:"party_layer"`
	PartyX     []int `json:"party_x"`
	// ItemSlots 是物品那一頁的清單項 → 角色記錄位移（RE-127）：穿戴欄十格、
	// 手上那件、背包六格，順序就是清單上的順序。
	ItemSlots []int `json:"item_slots"`
	// SlotMarks 是物品那一頁左上角那個小框畫在第幾列：一個種族十七筆，
	// 指向那個部位在立繪上的高度（`sub_16D0B`，RE-127）。
	SlotMarks [][]int `json:"slot_marks"`
}

// ItemSlotCount 是物品那一頁的清單有幾項。
const ItemSlotCount = 17

// Load 讀一份版面表。欄位少一塊就直接失敗——立繪少一層不會有錯誤訊息，
// 只會靜靜地畫錯，所以寧可開不起來。
func Load(path string) (*Layout, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var l Layout
	if err := json.Unmarshal(raw, &l); err != nil {
		return nil, fmt.Errorf("%s：%w", path, err)
	}
	if !strings.HasPrefix(l.Schema, "kol-portrait-layout-") {
		return nil, fmt.Errorf("%s：schema 是 %q", path, l.Schema)
	}
	if err := l.check(); err != nil {
		return nil, fmt.Errorf("%s：%w", path, err)
	}
	return &l, nil
}

// check 驗筆數之間的關係。這些數字互相牽制，對不起來就是抽錯了。
func (l *Layout) check() error {
	builds := len(l.Body)
	if builds == 0 {
		return fmt.Errorf("沒有體型")
	}
	if len(l.Layers) != builds*StylesPerBuild {
		return fmt.Errorf("外觀有 %d 套，應該是 %d 個體型 × %d",
			len(l.Layers), builds, StylesPerBuild)
	}
	for _, name := range []struct {
		what string
		rows [][]Cell
	}{{"武器", l.Weapons}, {"盾", l.Shields}} {
		if len(name.rows) != builds {
			return fmt.Errorf("%s的位置表有 %d 組，應該是 %d", name.what, len(name.rows), builds)
		}
		for i, row := range name.rows {
			if len(row) != len(name.rows[0]) {
				return fmt.Errorf("%s第 %d 組有 %d 筆，第 0 組是 %d",
					name.what, i, len(row), len(name.rows[0]))
			}
		}
	}
	if len(l.Hands) != builds*2 {
		return fmt.Errorf("手有 %d 筆，應該是 %d", len(l.Hands), builds*2)
	}
	if len(l.PartyLayer) != builds || len(l.PartyX) != builds {
		return fmt.Errorf("隊伍頁那兩張表是 %d／%d 筆，應該都是 %d",
			len(l.PartyLayer), len(l.PartyX), builds)
	}
	if len(l.ItemSlots) != ItemSlotCount {
		return fmt.Errorf("物品頁的清單有 %d 項，應該是 %d",
			len(l.ItemSlots), ItemSlotCount)
	}
	if len(l.SlotMarks) == 0 {
		return fmt.Errorf("沒有小框那張表")
	}
	for r, row := range l.SlotMarks {
		if len(row) != ItemSlotCount {
			return fmt.Errorf("小框第 %d 個種族有 %d 筆，應該是 %d",
				r, len(row), ItemSlotCount)
		}
	}
	seen := make(map[int]bool, len(l.ItemSlots))
	for _, off := range l.ItemSlots {
		if seen[off] {
			return fmt.Errorf("物品頁的清單有重複的位移 0x%x", off)
		}
		seen[off] = true
	}
	return nil
}
