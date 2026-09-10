package main

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/oracle"
	"github.com/wicanr2/kol-remake-cht/internal/roster"
	"github.com/wicanr2/kol-remake-cht/internal/town"
)

// 把原版那一場遭遇的地形接進受控情境，並用它交叉驗證座標語意。
//
// `TestOriginalNormalPartyState` 建戰鬥時給 `NewQuest` 的地形是 `nil`——60 個
// 初始欄位都對得上，但那一場打在一片空白上。這一支補的就是那一塊。
//
// # 為什麼這裡才驗得到座標語意
//
// 站位（`battle-alignment.json` 的參戰記錄 `+31`／`+32`）與地形
// （`battle-terrain.json` 的 4096 格）是**兩份各自獨立的原版資料**。原版繪製時
// 拿站位去索引地形的方式是 `SI = Y*64 + X`，所以「remake 的 `Field.At(a.X, a.Y)`
// 等於快照的 `record_hex[a.Y*64 + a.X]`」這件事一旦成立，就同時釘住了三件事：
// 地形的轉置方向、`Actor.X／Y` 的語意，以及兩者用的是同一個座標系。
//
// 轉置寫反的話，`At(x, y)` 會回到 `record_hex[x*64 + y]`——只有 x == y 的格子
// 碰巧一樣。這一場的參戰者裡 Olear 在 (33, 30)、Theodoric 在 (31, 30)，都不在
// 對角線上，所以寫反一定會被抓到。
func TestOriginalNormalBattleTerrain(t *testing.T) {
	statePath := os.Getenv("KOL_PARTY_STATE_INPUT")
	terrainPath := os.Getenv("KOL_TERRAIN_INPUT")
	if statePath == "" || terrainPath == "" {
		t.Skip("需要 dosgolem 正常入口的角色資料與地形快照")
	}

	var input struct {
		Source       string
		Seed         int
		Party        []roster.Slot
		Monster      battle.Monster
		MonsterLevel int `json:"monster_level"`
		Actors       []struct {
			Slot   int
			Record string `json:"record_hex"`
			Side   int
			X, Y   int
		}
	}
	body, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(body, &input); err != nil {
		t.Fatal(err)
	}
	field, snapshot, err := oracle.LoadTerrainSnapshot(terrainPath)
	if err != nil {
		t.Fatal(err)
	}
	// 兩份資料要出自同一場遭遇，否則底下每一格比較都在比別的東西。
	if input.Source != snapshot.Source || input.Seed != snapshot.AlignmentSeed {
		t.Fatalf("角色資料是 %q／seed %d，地形快照是 %q／seed %d——不是同一場",
			input.Source, input.Seed, snapshot.Source, snapshot.AlignmentSeed)
	}

	site := &battle.Site{Index: -1}
	var party []battle.Member
	for i, slot := range input.Party {
		m := town.Member{Name: slot.Name, Wounds: slot.Wounds, Sheet: townSheet(slot.Sheet),
			InHand: townGood(slot.InHand), Worn: townWorn(slot.Equipment)}
		party = append(party, battleMember(m))
		site.PartyX = append(site.PartyX, input.Actors[i].X)
		site.PartyY = append(site.PartyY, input.Actors[i].Y)
	}
	site.Count, site.Level, site.MonsterName = 1, input.MonsterLevel, input.Monster.Name
	site.Monster = input.Monster.Index
	site.MonsterX, site.MonsterY = []int{input.Actors[6].X}, []int{input.Actors[6].Y}

	s := battle.NewQuest(site, party, input.Monster, field)
	if s.Field == nil {
		t.Fatal("戰鬥狀態沒有地形")
	}

	// 一、地形整張傳進去，沒有被生成或改寫。
	raw, err := hex.DecodeString(snapshot.RecordHex)
	if err != nil {
		t.Fatal(err)
	}
	wrong := 0
	for i, v := range raw {
		x, y := i%battle.FieldSide, i/battle.FieldSide
		if got := s.Field.At(x, y); got != int(v) {
			if wrong++; wrong <= 3 {
				t.Errorf("(%d, %d)：原版 %#x，remake %#x", x, y, v, got)
			}
		}
	}
	if wrong > 0 {
		t.Fatalf("%d／%d 格對不上", wrong, len(raw))
	}

	// 二、誰在場上。**不能拿原版記錄的 `+1`（陣營）當「有沒有人」用**：
	// 沒用到的隊員槽 7…17 在原版記憶體裡整條是零，於是陣營也讀成 0＝隊伍。
	// 那是殘留值不是語意。真正在場的是 fixture 的兩名隊員與 slot 6 的怪物。
	inPlay := map[int]bool{0: true, 1: true, 6: true}
	for i, a := range s.Actors {
		if want, got := inPlay[i], a.Side != battle.SideEmpty; want != got {
			t.Errorf("槽 %d：remake 認為%s在場上，原版那一場%s", i,
				map[bool]string{true: "", false: "不"}[got],
				map[bool]string{true: "有", false: "沒有"}[want])
		}
	}

	// 三、站位與地形是同一個座標系。座標取自另一份原版資料。
	offDiagonal, checked := 0, 0
	for _, a := range input.Actors {
		if !inPlay[a.Slot] {
			continue
		}
		rec, err := hex.DecodeString(a.Record)
		if err != nil || len(rec) != 50 {
			t.Fatalf("槽 %d 的參戰記錄不是 50 byte", a.Slot)
		}
		// 站位取自記錄本身，不是取自 fixture 整理過的欄位。
		x, y := int(rec[31]), int(rec[32])
		if x != a.X || a.Y != y {
			t.Fatalf("槽 %d 的站位 fixture 說 (%d, %d)，原版記錄是 (%d, %d)", a.Slot, a.X, a.Y, x, y)
		}
		if x >= battle.FieldSide || y >= battle.FieldSide {
			t.Fatalf("槽 %d 站在戰場外 (%d, %d)", a.Slot, x, y)
		}
		checked++
		if x != y {
			offDiagonal++
		}
		want := int(raw[y*battle.FieldSide+x])
		if got := s.Field.At(x, y); got != want {
			t.Errorf("槽 %d（%s）站在 (%d, %d)：原版那一格是 %#x，remake 讀到 %#x",
				a.Slot, a.Record[:8], x, y, want, got)
		}
		switch want {
		case battle.TileGround, battle.TileRough, battle.TileScatter:
		default:
			t.Errorf("槽 %d 站在 %#x 這種磚塊上——一般遭遇只該有三種", a.Slot, want)
		}
	}
	if checked == 0 {
		t.Fatal("一個參戰者都沒檢查到")
	}
	// 正對照：判準要真的分得出轉置。全部站在對角線上就什麼都沒驗到。
	if offDiagonal == 0 {
		t.Fatal("每個參戰者都站在 x == y 上，這個判準分不出轉置寫沒寫反")
	}
	if checked != len(inPlay) {
		t.Fatalf("在場的有 %d 名，只檢查到 %d 名", len(inPlay), checked)
	}
	t.Logf("地形 %d 格與 %d 名在場參戰者的站位同一座標系（其中 %d 名不在對角線上）；"+
		"另確認未使用的 15 個槽 remake 都判為空", len(raw), checked, offDiagonal)
}
