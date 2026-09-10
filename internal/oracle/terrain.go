package oracle

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/wicanr2/kol-remake-cht/internal/battle"
)

// 正常玩家路徑上那一場戰鬥的地形快照（`kol-battle-terrain-receipt-v1`）。
//
// 這份東西跟 `TestBattleFieldGenerationOriginal` 不一樣：那一支拿人造 seed
// 叫原版的生成常式，驗的是「同一個 seed 兩邊算出同一張圖」；這一份是原版在
// **正常遭遇**裡真的產生出來、當時畫在螢幕上的那一張，用途是給後續的指令與
// 結算對拍一個共同起點。地形生成時的亂數狀態沒有留下來，所以它只能被載入，
// 不能被重算。
//
// # 索引
//
// 快照存的是原版繪製時直接吃的線性位移：`SI = Y*64 + X`（`offset_contract`）。
// remake 的正式儲存是 `Tiles[X*64 + Y]`（`Field.At`），而且它的 **X 是畫面上的
// 列、Y 是欄**。兩邊的 X／Y 語意相同，差的只有哪一軸乘 64——所以轉換就是
// `battle.Generate` 裡那一行同樣的轉置。
//
// 這一格寫反不會報錯：4096 個 byte 照樣填滿，磚塊編號照樣合法，只是整張圖
// 沿對角線翻過來。所以載入器自己不判斷對錯，由 `terrain_test.go` 的正反對照
// 釘住：用一張不對稱的合成地形，翻過來就必須被抓到。

// TerrainSnapshot 是快照檔的內容。只列用得到的欄位，其餘留在檔案裡。
type TerrainSnapshot struct {
	Schema         string   `json:"schema"`
	Source         string   `json:"source"`
	Input          string   `json:"input"`
	InputSHA256    string   `json:"input_sha256"`
	Index          string   `json:"index"`
	OffsetContract string   `json:"offset_contract"`
	Width          int      `json:"width"`
	Height         int      `json:"height"`
	AlignmentSeed  int      `json:"alignment_seed"`
	CurrentActor   int      `json:"current_actor"`
	Confidence     string   `json:"confidence"`
	Evidence       string   `json:"evidence"`
	Limitations    []string `json:"limitations"`
	RecordHex      string   `json:"record_hex"`
}

// TerrainSchema 是這支載入器認得的 schema。換版本要改這裡，不是預設接受。
const TerrainSchema = "kol-battle-terrain-receipt-v1"

// terrainIndex 是快照宣告的索引契約。載入器只認這一種；換了就要重新看轉換。
const terrainIndex = "Y*64+X"

// LoadTerrainSnapshot 讀快照並轉成 remake 的戰場。
//
// 每一項前提都當場檢查：schema、索引契約、寬高、byte 數。缺一項就失敗，
// 不猜、不補預設——載進來的是要拿去當「原版當時長這樣」的證據，前提對不上
// 而繼續算，後面每一格比較都在比別的東西。
func LoadTerrainSnapshot(path string) (*battle.Field, *TerrainSnapshot, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("讀地形快照：%w", err)
	}
	var s TerrainSnapshot
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, nil, fmt.Errorf("解析地形快照：%w", err)
	}
	if s.Schema != TerrainSchema {
		return nil, nil, fmt.Errorf("schema 是 %q，這支載入器只認 %q", s.Schema, TerrainSchema)
	}
	if s.Index != terrainIndex {
		return nil, nil, fmt.Errorf("索引契約是 %q，這支載入器只認 %q——換了就要重看轉置",
			s.Index, terrainIndex)
	}
	if s.Width != battle.FieldSide || s.Height != battle.FieldSide {
		return nil, nil, fmt.Errorf("快照是 %d×%d，remake 的戰場是 %d×%d",
			s.Width, s.Height, battle.FieldSide, battle.FieldSide)
	}
	tiles, err := hex.DecodeString(s.RecordHex)
	if err != nil {
		return nil, nil, fmt.Errorf("record_hex 不是十六進位：%w", err)
	}
	if len(tiles) != battle.FieldSide*battle.FieldSide {
		return nil, nil, fmt.Errorf("record_hex 解出 %d 個 byte，應該是 %d",
			len(tiles), battle.FieldSide*battle.FieldSide)
	}

	f := &battle.Field{Tiles: make([]int, battle.FieldSide*battle.FieldSide)}
	for i, v := range tiles {
		// i 是原版的 `Y*64 + X`；remake 存的是 `X*64 + Y`。
		x, y := i%battle.FieldSide, i/battle.FieldSide
		f.Tiles[x*battle.FieldSide+y] = int(v)
	}
	return f, &s, nil
}

// TerrainTileCounts 數每一種磚塊各幾格。轉置寫反時總數不會變，所以它不是
// 用來抓轉置的——它回答的是另一個問題：這張圖的組成看起來合不合理
// （一般遭遇應該只有底色、粗糙地面與灑上去的那三種）。
func TerrainTileCounts(f *battle.Field) map[int]int {
	counts := map[int]int{}
	for _, v := range f.Tiles {
		counts[v]++
	}
	return counts
}
