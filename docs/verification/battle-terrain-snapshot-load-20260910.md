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

## 接進受控情境

`TestOriginalNormalPartyState` 建戰鬥時給 `NewQuest` 的地形是 `nil`——60 個初始
欄位都對得上，但那一場打在一片空白上。`TestOriginalNormalBattleTerrain`
（`cmd/kol/original_battle_terrain_test.go`）補上這一塊，並且順便驗到一件更根本的事。

**兩份快照是同一場。** `battle-state-records-20260908/` 與
`battle-terrain-normal-fixed-20260908/` 的 `battle-state-records.json` 與
`battle-alignment.json` 逐欄比對**完全相同**——dosgolem 兩次獨立執行到達同一個
狀態，決定性成立，後者（含地形）因此可以當統一的共同起點。第三份
`battle-terrain-normal-20260908/` 沒有 `battle-terrain.json`：那一次匯出沿用暫存
指標，被資料檢查擋下來了。

三項判準：

1. **地形整張傳進去**：4096 格逐格比對 `s.Field.At(x, y)` 與快照。
2. **誰在場上**：在場的正好是 slot 0、1、6，其餘 15 個槽 remake 都判為
   `SideEmpty`。這一項要小心——**不能拿原版記錄的 `+1`（陣營）當「有沒有人」用**：
   沒用到的隊員槽 7…17 在原版記憶體裡整條是零，陣營因此讀成 0＝隊伍。那是殘留值
   不是語意。第一版判準就是這樣把 14 個槽當成參戰者，數字看起來還很合理。
3. **站位與地形同一座標系**：站位取自 `battle-alignment.json` 的參戰記錄
   `+31`／`+32`，地形取自 `battle-terrain.json`——**兩份各自獨立的原版資料**。
   原版繪製時的索引是 `SI = Y*64 + X`，所以「`Field.At(a.X, a.Y)` 等於
   `record_hex[a.Y*64 + a.X]`」一旦成立，就同時釘住了地形的轉置方向、`Actor.X／Y`
   的語意，以及兩者用的是同一個座標系。

三名在場參戰者都不在對角線上（Olear 33/30、Theodoric 31/30、SLEDGE 31/37），
所以這個判準真的分得出轉置——測試會在全部落在 `x == y` 時直接失敗，不讓一組
碰巧無效的座標默默通過。

## 負對照：實際把轉置寫反跑一次

把 `terrain.go` 的 `f.Tiles[x*64+y]` 改成 `f.Tiles[y*64+x]` 之後重跑：

| 判準 | 反應 |
|---|---|
| `TestOriginalNormalBattleTerrain` | FAIL，**1,818／4,096 格對不上** |
| `TestTerrainSnapshotIndexTransposition` | FAIL |
| `TestTerrainSnapshotTransposedIsDetected` | FAIL |

那個 1,818 值得看一眼：翻過來只有 **44%** 的格子報錯，其餘 56% 碰巧一樣——因為
底色佔了七成，隨便兩格拿來比都很可能相同。**這就是為什麼合成地形那一支要用
不對稱的 `x*2 + y*3`**：拿真實地形驗轉置，一半以上的格子幫不上忙；要是再挑到
對稱的判準，測試會全綠而什麼都沒證明。

改回來之後 `go test ./cmd/... ./internal/...` 十六個套件全綠。

## 還沒做

- 完整裝備能力：`battle-state-records.json` 的怪物記錄目前只擷取 0x80 byte，
  不能當成全部穿戴欄已完整保存。
- 規劃、指令與攻擊結算比較：這一份只是共同起點。
- 正常 GUI：`fight-entry.png` 還沒與 remake 的畫面比對。
