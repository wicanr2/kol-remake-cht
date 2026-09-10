package oracle_test

import (
	"encoding/hex"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/kol-remake-cht/internal/battle"
	"github.com/wicanr2/kol-remake-cht/internal/oracle"
)

// 正常玩家路徑的地形快照。gitignore 的研究輸入，沒有就跳過那幾支。
const snapshotPath = "../../workplace/battle-terrain-normal-fixed-20260908/battle-terrain.json"

// synth 造一張快照，內容按原版的 `Y*64+X` 排。
func synth(t *testing.T, tile func(x, y int) byte, tweak func(m map[string]any)) string {
	t.Helper()
	buf := make([]byte, battle.FieldSide*battle.FieldSide)
	for y := 0; y < battle.FieldSide; y++ {
		for x := 0; x < battle.FieldSide; x++ {
			buf[y*battle.FieldSide+x] = tile(x, y)
		}
	}
	m := map[string]any{
		"schema": oracle.TerrainSchema, "index": "Y*64+X",
		"width": battle.FieldSide, "height": battle.FieldSide,
		"record_hex": hex.EncodeToString(buf),
	}
	if tweak != nil {
		tweak(m)
	}
	raw, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(t.TempDir(), "battle-terrain.json")
	if err := os.WriteFile(p, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

// 轉置正著寫，`At(x, y)` 就回得到原版 `Y*64+X` 那一格。
//
// **地形一定要不對稱**：拿 `tile = x+y` 這種對稱的圖來驗，翻過來每一格都一樣，
// 測試照樣全綠而什麼都沒證明。這裡用 `x*2 + y*3`，交換 x／y 就換一個值。
func TestTerrainSnapshotIndexTransposition(t *testing.T) {
	tile := func(x, y int) byte { return byte(x*2 + y*3) }
	f, s, err := oracle.LoadTerrainSnapshot(synth(t, tile, nil))
	if err != nil {
		t.Fatal(err)
	}
	if s.Index != "Y*64+X" {
		t.Fatalf("索引契約沒讀出來：%q", s.Index)
	}
	wrong := 0
	for y := 0; y < battle.FieldSide; y++ {
		for x := 0; x < battle.FieldSide; x++ {
			if got, want := f.At(x, y), int(tile(x, y)); got != want {
				if wrong++; wrong <= 3 {
					t.Errorf("At(%d, %d) = %d，快照那一格是 %d", x, y, got, want)
				}
			}
		}
	}
	if wrong > 0 {
		t.Fatalf("%d 格對不上", wrong)
	}
}

// 負對照：把快照沿對角線翻過來再載入，一定要被上面那個判準抓到。
//
// 沒有這一支，`TestTerrainSnapshotIndexTransposition` 證明不了自己在看什麼——
// 轉置寫反時 4096 個 byte 照樣填滿、磚塊編號照樣合法，唯一的差別就是這裡
// 要驗出來的那個差別。
func TestTerrainSnapshotTransposedIsDetected(t *testing.T) {
	tile := func(x, y int) byte { return byte(x*2 + y*3) }
	// 造快照時把 x／y 對調＝資料被翻過來了。
	f, _, err := oracle.LoadTerrainSnapshot(synth(t, func(x, y int) byte { return tile(y, x) }, nil))
	if err != nil {
		t.Fatal(err)
	}
	for y := 0; y < battle.FieldSide; y++ {
		for x := 0; x < battle.FieldSide; x++ {
			if f.At(x, y) != int(tile(x, y)) {
				return // 抓到了，正確
			}
		}
	}
	t.Fatal("翻過來的地形竟然每一格都對得上——判準沒有在看轉置")
}

// 前提對不上就要當場失敗，不猜也不補預設。
func TestTerrainSnapshotRejectsBadPremises(t *testing.T) {
	tile := func(x, y int) byte { return byte(x + y) }
	for _, c := range []struct {
		name  string
		tweak func(m map[string]any)
	}{
		{"schema 換版本", func(m map[string]any) { m["schema"] = "kol-battle-terrain-receipt-v2" }},
		{"索引契約改了", func(m map[string]any) { m["index"] = "X*64+Y" }},
		{"寬高不是 64", func(m map[string]any) { m["width"] = 32 }},
		{"byte 數不足", func(m map[string]any) { m["record_hex"] = "0d09" }},
		{"record_hex 不是十六進位", func(m map[string]any) { m["record_hex"] = "在座的各位都是" }},
	} {
		t.Run(c.name, func(t *testing.T) {
			if _, _, err := oracle.LoadTerrainSnapshot(synth(t, tile, c.tweak)); err == nil {
				t.Fatal("前提對不上卻載進來了")
			}
		})
	}
}

// 真的那一份：正常遭遇的地形只該有底色、粗糙地面與灑上去的三種磚塊。
func TestNormalPathTerrainSnapshot(t *testing.T) {
	if _, err := os.Stat(snapshotPath); err != nil {
		t.Skipf("沒有正常路徑地形快照（研究輸入，不進版控）：%v", err)
	}
	f, s, err := oracle.LoadTerrainSnapshot(snapshotPath)
	if err != nil {
		t.Fatal(err)
	}
	if s.Source != "dosgolem-normal-player-path" {
		t.Fatalf("來源是 %q，這一支要驗的是正常玩家路徑那一份", s.Source)
	}
	counts := oracle.TerrainTileCounts(f)
	total := 0
	for tile, n := range counts {
		switch tile {
		case battle.TileGround, battle.TileRough, battle.TileScatter:
			total += n
		default:
			t.Errorf("出現不該在一般遭遇裡的磚塊 %#x，%d 格", tile, n)
		}
	}
	if total != battle.FieldSide*battle.FieldSide {
		t.Fatalf("認得的磚塊只有 %d 格，整張圖是 %d 格", total, battle.FieldSide*battle.FieldSide)
	}
	for _, tile := range []int{battle.TileGround, battle.TileRough, battle.TileScatter} {
		if counts[tile] == 0 {
			t.Errorf("磚塊 %#x 一格都沒有——一般遭遇三種都該出現", tile)
		}
	}
	t.Logf("正常路徑地形：底色 %d、粗糙 %d、灑落 %d（seed=%d、current_actor=%d）",
		counts[battle.TileGround], counts[battle.TileRough], counts[battle.TileScatter],
		s.AlignmentSeed, s.CurrentActor)

	// 組成要對得上 SPEC-008 的生成規則。這一條的價值在於它**跟索引轉換無關**：
	// 沿對角線翻過來，每一種磚塊的格數一個都不會變，所以它驗的是另一件事——
	// 這張圖真的是那條規則產生的，不是別處來的資料或解錯的 payload。
	//
	// 規則是逐格擲 1…4、中 1 填粗糙，期望 4096÷4 = 1024 格。但灑落那一段會把
	// 一部分粗糙格換成 TileScatter（`Scatter` 只換 Ground／Rough），所以要先把
	// 蓋掉的加回來：灑落的落點是均勻的，其中約「粗糙佔全場的比例」那麼多本來
	// 是粗糙格。
	side := float64(battle.FieldSide)
	cells := side * side
	scatter := float64(counts[battle.TileScatter])
	rough := float64(counts[battle.TileRough])
	// 灑落之前粗糙格佔的比例，用還原後的數量迭代一次就夠準。
	share := (rough + scatter*0.25) / cells
	restored := rough + scatter*share
	const want = 1024.0                  // 4096 × 1/4
	sd := math.Sqrt(cells * 0.25 * 0.75) // 二項分布的標準差，約 27.7
	if off := math.Abs(restored-want) / sd; off > 3 {
		t.Errorf("還原後的粗糙格 %.0f，期望 %.0f，差 %.1f 個標準差——這張圖不像是"+
			"逐格擲 1…4 產生的", restored, want, off)
	} else {
		t.Logf("還原粗糙格 %.0f（期望 %.0f，差 %.2f 個標準差）：組成與 SPEC-008 的"+
			"生成規則一致", restored, want, off)
	}
}
