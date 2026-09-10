package battle

// 競技場與訓練戰那一場的站位（RE-098 的第三條分支）。
//
// `sub_1E3A8` 擺人的時候看 `word_345DE`——它是「這一場要升階的那一名」的角色記錄
// 位移，一般戰鬥是 `0xFFFF`。它 `>= 0` 時兩個迴圈都不讀站位表：上場那一名與唯一
// 那隻敵人各站一格固定的，其餘隊員的記錄 `+0x01` 設成 2（不參戰）並排到觀眾席。
//
// 因此競技場開打前擲的 `word_37CF6 = rand(5) − 1`（陣型）在這一場**用不到**——
// 陣型表只有隨機遭遇那一條會讀。那一次擲骰仍然要擲，否則之後的亂數序列會偏掉。

// 上場那一名與那一隻敵人的固定站位（`0x1E79E` 的 `0x14`／`0x1F` 與 `0x1E830` 的
// `0x16`／`0x1F`）。X 是原版寫進參戰者記錄 `+0x1F` 的那一格（欄）、Y 是 `+0x20`
// （列），與其他站位表同一套（RE-095）。
const (
	ArenaFighterX = 20
	ArenaFighterY = 31
	ArenaEnemyX   = 22
	ArenaEnemyY   = 31
)

// ArenaWatcherX／ArenaWatcherY 是觀眾席（DGROUP `0x12B2` 欄／`0x12AC` 列）。
// 隊伍最多六人、上場一名，所以第六格永遠排不到人，原版那一格是 0。
var (
	ArenaWatcherX = [PartySlots]int{16, 16, 47, 47, 47, 0}
	ArenaWatcherY = [PartySlots]int{25, 35, 36, 30, 31, 0}
)

// NewArena 擺好競技場那一場：一隻對手、只有 fighter 那一名參戰。
//
// `kind` 是對手的種類編號、`level` 是對手等級（上場那一名的實得位階加一）。
// 地形由呼叫端依 SPEC-025 載入 TOWN0 payload 85，再套用共同樹叢散布。
func NewArena(fighter, kind, level int, monster Monster, party []Member,
	field *Field) *State {
	if fighter < 0 || fighter >= len(party) {
		return nil
	}
	site := &Site{Index: -1, Monster: kind, MonsterName: monster.Name,
		Level: level, Count: 1,
		MonsterX: []int{ArenaEnemyX}, MonsterY: []int{ArenaEnemyY}}
	// 觀眾序號只有排到觀眾席的人才加一（`0x1E72C` 的 `word_39A4E`），
	// 上場那一名不佔號碼。
	watcher := 0
	for n := range party {
		if n >= PartySlots {
			break
		}
		if n == fighter {
			site.PartyX = append(site.PartyX, ArenaFighterX)
			site.PartyY = append(site.PartyY, ArenaFighterY)
			continue
		}
		site.PartyX = append(site.PartyX, ArenaWatcherX[watcher])
		site.PartyY = append(site.PartyY, ArenaWatcherY[watcher])
		watcher++
	}
	s := NewQuest(site, party, monster, field)
	if s == nil {
		return nil
	}
	// 這一場打完要替 fighter 升階，而且戰後獎勵整段不發（RE-096 的閘門）。
	s.RankFight = true
	for n := range party {
		if n >= PartySlots || n == fighter {
			continue
		}
		s.Actors[n].Side = SideEmpty
	}
	s.Stage = StageEncounter
	return s
}
