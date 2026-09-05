// Package portrait 把角色立繪疊成一串「貼哪一張、貼在哪」（RE-126）。
//
// 原版的 `sub_18F3A` 依序畫七層：體型、腿、身、頭、盾、武器、手。每一層畫在哪、
// 多大，來自 KOLTOWN 資料段那一串版面表（`assets/original-derived-data/
// portrait-layout.json`）。這個套件只算座標，不碰圖檔——載入與貼圖是 UI 那一層的事。
//
// 座標沿用原版：X 與寬的單位是 **byte 欄**（EGA 平面模式一個 byte 橫跨八個像素），
// Y 與高才是像素。
package portrait

// 層的編號與原版的迴圈變數相同：0 頭、1 身、2 腿，畫的順序是 2 → 1 → 0。
const (
	LayerHead  = 0
	LayerTorso = 1
	LayerLegs  = 2
	LayerCount = 3
)

// 圖庫名。原版是把這幾個字接上模式後綴（EGA 是 `pics`）組出檔名。
const (
	ArchiveBody   = "PERPICS"
	ArchiveHead   = "HEADPICS"
	ArchiveTorso  = "TORPICS"
	ArchiveLegs   = "LEGSPICS"
	ArchiveShield = "SHLDPICS"
	ArchiveWeapon = "WEAPPICS"
	ArchiveHand   = "HANDPICS"
)

// layerArchive 是三層各自的圖庫，索引與 LayerHead…LayerLegs 對齊。
var layerArchive = [LayerCount]string{ArchiveHead, ArchiveTorso, ArchiveLegs}

// Item 是立繪要看的那幾個欄位。Class 是物品記錄 `+0x0D` 整個 byte、
// Style 是 `+0x0E` 的低 nibble，兩者都直接來自 typed data。
type Item struct {
	Class int
	Style int
	// Kind 是 `+0x0C` 的最低兩位元：0 Item、1 Wield、2 Wear、3 Both。
	Kind int
	// TwoHand 是 `+0x0D` 的 bit1，原版拿它決定要不要畫盾與第二隻手。
	TwoHand bool
}

// Figure 是一名角色的外觀。Styles 那三格是穿戴欄 0…2 的外觀號，
// 負數代表那個部位空著（原版看那一格的第 0 個 byte 是不是 0）。
type Figure struct {
	Gender int
	Race   int
	Styles [LayerCount]int
	// Weapon 是手上那件（角色記錄 `+0x30`）；Shield 是穿戴欄部位 6（`+0x160`）。
	Weapon *Item
	Shield *Item
	// Belt 是腰帶那一格（`+0x190`，物品頁上叫 BELT）。原版會為它多畫一層，
	// 但組出來的檔名在原版資料裡不存在，所以這裡只記著、不畫（RE-126）。
	Belt *Item
}

// Build 是體型組合：性別 × 4 ＋ 種族。原版資料下只有 0…5——女性只有種族 0 與 1。
func (f Figure) Build() int { return f.Gender*4 + f.Race }

// held 回報手上那件算不算「拿著的武器」。原版連比三次：那一格要非空，
// 而且種類的最低兩位元不能是 0（純物品）或 2（穿戴品）。
func (f Figure) held() bool {
	if f.Weapon == nil {
		return false
	}
	switch f.Weapon.Kind & 3 {
	case 0, 2:
		return false
	}
	return true
}

// Piece 是要貼的一張圖。X 與 W 的單位是 byte 欄，Y 與 H 是像素。
type Piece struct {
	Archive string `json:"archive"`
	Index   int    `json:"index"`
	X       int    `json:"x"`
	Y       int    `json:"y"`
	W       int    `json:"w"`
	H       int    `json:"h"`
}

// Compose 把一名角色疊成一串要貼的圖，順序就是畫的順序。dx 是整個人往右挪幾個
// byte 欄（原版的 `arg_0`）；隊伍那一頁靠它把六個人錯開。
func (l *Layout) Compose(f Figure, dx int) []Piece {
	v := f.Build()
	if l == nil || v < 0 || v >= len(l.Body) {
		return nil
	}
	body := l.Body[v]
	out := []Piece{{Archive: ArchiveBody, Index: v,
		X: body.X + dx, Y: body.Y, W: body.X2 - body.X, H: body.Y2 - body.Y}}

	// 腿、身、頭：那一格穿戴欄空著就整層不畫。
	for layer := LayerLegs; layer >= LayerHead; layer-- {
		style := f.Styles[layer]
		if style < 0 {
			continue
		}
		s := v*StylesPerBuild + style
		if s < 0 || s >= len(l.Layers) {
			continue
		}
		c := l.Layers[s][layer]
		out = append(out, Piece{Archive: layerArchive[layer], Index: s,
			X: c.X + dx, Y: c.Y, W: c.W, H: c.H})
	}

	held := f.held()
	// 盾：雙手武器就沒有手拿盾。
	if f.Shield != nil && !(held && f.Weapon.TwoHand) {
		if t := f.Shield.Style; t >= 0 && v < len(l.Shields) && t < len(l.Shields[v]) {
			c := l.Shields[v][t]
			out = append(out, Piece{Archive: ArchiveShield, Index: t,
				X: c.X + dx, Y: c.Y, W: c.W, H: c.H})
		}
	}
	if held {
		if i, ok := l.WeaponIndex(*f.Weapon); ok && v < len(l.Weapons) {
			c := l.Weapons[v][i]
			out = append(out, Piece{Archive: ArchiveWeapon, Index: i,
				X: c.X + dx, Y: c.Y, W: c.W, H: c.H})
		}
	}
	// 手：一律畫一隻，雙手武器再畫第二隻。
	for k := 0; k < 2; k++ {
		if k == 1 && !(held && f.Weapon.TwoHand) {
			break
		}
		h := v*2 + k
		if h >= len(l.Hands) {
			break
		}
		c := l.Hands[h]
		out = append(out, Piece{Archive: ArchiveHand, Index: h,
			X: c.X + dx, Y: c.Y, W: c.W, H: c.H})
	}
	return out
}

// WeaponIndex 把手上那件換成 WEAPPICS 的組號（RE-126）。原版算的是 payload 索引
// ——基底表加上樣式的兩倍——組號要再除以二。種類 5 的樣式 2 與 3 對調。
func (l *Layout) WeaponIndex(w Item) (int, bool) {
	kind := w.Class >> 5
	if kind < 0 || kind >= len(l.WeaponBase) {
		return 0, false
	}
	style := w.Style
	if kind == SwappedWeaponKind && (style == 2 || style == 3) {
		style ^= 1
	}
	payload := l.WeaponBase[kind] + 2*style
	index := payload / 2
	if len(l.Weapons) == 0 || index < 0 || index >= len(l.Weapons[0]) {
		return 0, false
	}
	return index, true
}

// PartyOrder 是隊伍那一頁的畫圖順序（`sub_18D61`）：層次大的先畫（站得遠），
// 同一層次的照名冊順序。回傳的是 builds 的索引，以及每個人要挪幾個 byte 欄。
func (l *Layout) PartyOrder(builds []int) (order []int, dx []int) {
	if l == nil {
		return nil, nil
	}
	for layer := len(l.PartyLayer); layer >= 1; layer-- {
		for i, v := range builds {
			if v < 0 || v >= len(l.PartyLayer) || l.PartyLayer[v] != layer {
				continue
			}
			n := len(order)
			order = append(order, i)
			if n < len(l.PartyX) {
				dx = append(dx, l.PartyX[n])
			} else {
				dx = append(dx, 0)
			}
		}
	}
	return order, dx
}
