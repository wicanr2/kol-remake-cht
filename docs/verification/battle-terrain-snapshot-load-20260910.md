# 正常路徑地形快照接入 remake（2026-09-10）

## 範圍

把 `workplace/battle-terrain-normal-fixed-20260908/battle-terrain.json` 接進
remake。這一份是原版在**正常遭遇**裡實際產生、當時畫在螢幕上的那 4096 格，
與 `TestBattleFieldGenerationOriginal` 不同——那一支拿人造 seed 叫原版的生成
常式，驗的是「同一個 seed 兩邊算出同一張圖」。

**這一份只能被載入，不能被重算。** 地形生成當下的亂數狀態沒有留在快照裡；
`alignment_seed: 1` 是規劃入口對齊之後設下的值，`seed_before_alignment`
（1323149826）是地形已經生成完畢之後的值。兩者都反推不回生成前的狀態。
它的用途是給後續的指令與結算對拍一個共同起點。

不宣稱：完整戰鬥結算、屬性對拍、正常 GUI 畫面比較。

## 輸入

| 項目 | 值 |
|---|---|
| 快照 | `battle-terrain.json`，`kol-battle-terrain-receipt-v1` |
| 來源 | `dosgolem-normal-player-path` |
| 原版輸入 | `KOLTOWN-unpacked.EXE`，SHA-256 `01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2` |
| 索引契約 | `Y*64+X`；`offset_contract` 記「繪製直接使用 ES:SI，SI 為 Y*64+X，無額外基底位移」 |
| 證據 | RE-080；`docs/re/20260908-battle-field-index.md` |

## 索引：這一格寫反不會報錯

原版存的是繪製時直接吃的線性位移 `Y*64 + X`。remake 的正式儲存是
`Tiles[X*64 + Y]`（`Field.At`），而且它的 **X 是畫面上的列、Y 是欄**——
兩邊的 X／Y 語意相同，差的只有哪一軸乘 64。轉換因此與 `battle.Generate`
裡那一行相同：

```go
x, y := i%64, i/64          // i 是原版的 Y*64 + X
f.Tiles[x*64+y] = int(v)    // remake 存 X*64 + Y
```

寫反的話 4096 個 byte 照樣填滿、磚塊編號照樣合法，整張圖只是沿對角線翻過去。
所以判準要能分辨那個差別：

- `TestTerrainSnapshotIndexTransposition` 用**不對稱**的合成地形（`x*2 + y*3`）
  逐格比對。拿 `x+y` 這種對稱的圖來驗，翻過來每一格都一樣，測試照樣全綠而
  什麼都沒證明。
- `TestTerrainSnapshotTransposedIsDetected` 是負對照：把同一張圖翻過來再載入，
  上面那個判準**必須**抓到。沒有這一支，前一支證明不了自己在看什麼。
- `TestTerrainSnapshotRejectsBadPremises`：schema 換版本、索引契約改了、寬高
  不是 64、byte 數不足、`record_hex` 不是十六進位——五種前提對不上都當場失敗，
  不猜也不補預設。

## 這張圖真的是那條規則產生的

磚塊組成（`TestNormalPathTerrainSnapshot`）：

| 磚塊 | 格數 |
|---|---:|
| 底色 `0x09` | 2,848 |
| 粗糙 `0x0D` | 941 |
| 灑落 `0x0A` | 307 |
| 合計 | 4,096 |

SPEC-008 的生成規則是逐格擲 1…4、中 1 填粗糙，期望 4096 ÷ 4 = 1,024 格。
帳面上的 941 少了 83——因為灑落那一段會把一部分粗糙格換成 `TileScatter`
（`Scatter` 只換 Ground／Rough）。把蓋掉的加回來（灑落落點均勻，其中約
「粗糙佔全場的比例」那麼多本來是粗糙格）得 **1,017**，與期望的 1,024 差
**0.24 個標準差**（二項分布 σ ≈ 27.7）。

這一條的價值在於它**與索引轉換無關**：沿對角線翻過來，每一種磚塊的格數一個
都不會變。它回答的是另一個問題——這張圖確實出自那條生成規則，不是別處來的
資料或解錯的 payload。

灑落 307 格也對得上：擲 100 次 1…6 加總期望 350 次，扣掉重複選中同一格之後
剩下的相異格數會比 350 少。

## 程式落點

- `internal/oracle/terrain.go`：`LoadTerrainSnapshot`、`TerrainTileCounts`。
  不帶 `oracle` build tag——它只讀 JSON，不需要原版執行檔。
- `internal/oracle/terrain_test.go`：四支測試。真快照那一支在檔案不存在時
  `t.Skip`（研究輸入不進版控）。

runtime 沒有動：載入器在 `internal/oracle`，正式遊戲路徑不讀 `workplace/`。

## 驗證

容器內 `go test ./cmd/... ./internal/...` 全綠（`cmd/kol` 與 `internal/ui`
需要 X11，另起 Xvfb 跑過）。`gofmt` 對本輪新增的兩個檔案乾淨。

## 還沒做

- 完整屬性對拍：`battle-state-records.json` 的兩名隊員與 SLEDGE 記錄還沒接。
- 指令與結算比較：這一份只是共同起點。
- 正常 GUI：`fight-entry.png` 還沒與 remake 的畫面比對。
