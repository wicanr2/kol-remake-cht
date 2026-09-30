# SPEC-038：套件邊界、固定步進與注入點

狀態：**READY**。這是 remake 自己的工程契約，不含原版行為；守門測試是
`internal/layers/layers_test.go`（`go test ./internal/layers/`），違反時會失敗。存檔格式見
[SPEC-023](023-complete-save.md)（`kol-save-v3`）與 [SPEC-006](006-rest-and-save.md)。

## 分層

| 層 | 套件 | 規則 |
|---|---|---|
| 資料與規則 | `assets` `battle` `bitmapedit` `book` `chardata` `ending` `input` `magic` `portrait` `rng` `roster` `save` `title` `world` | 彼此不依賴，不碰畫面與 Ebitengine |
| 規則（用到一兩個資料套件） | `charcreate`→`chardata`；`i18n`→`magic`；`town`→`magic`、`portrait` | 只能依賴括號裡的套件 |
| 場景串接 | `flow`→`bitmapedit`、`charcreate`、`roster`、`town` | 只看規則，不看畫面 |
| 對拍輔助 | `oracle`→`battle` | 只讀戰鬥規則 |
| 畫面 | `font`、`screen`、`ui` | 只有這三個內部套件與 `cmd/kol` 可以依賴 Ebitengine；`ui` 可讀各規則套件 |
| 最外層 | `cmd/kol` | 可依賴全部；組裝、輸入、存檔時間戳記與隨機 ID 都在這一層 |

依賴方向單向：`cmd/kol` → 畫面／場景串接 → 規則 → 資料。允許的邊寫在測試的 `allowed` 表裡，
新增依賴要改那張表，改表就是一次看得見的決定；新增套件要先登記它在哪一層。

## 固定步進

- 遊戲以 Ebitengine 預設的每秒 60 次 `Update` 推進（`ebiten.DefaultTPS`）。專案沒有呼叫
  `SetTPS`，也沒有讀 `ActualTPS`／`ActualFPS`。
- 規則層完全不看時間：沒有任何規則套件引用 `time`，測試會擋 `time.Now`、`Since`、`After`、
  `Tick`、`Sleep`、`NewTimer`、`NewTicker`。回合、季節、天候節拍與戰鬥回合都以事件計數。
- 動畫與走路節拍只在畫面層與最外層以幀計數（例如 `cmd/kol` 的 `walkTicksPerFrame`）。
  改變幀率不會改變規則結果，只會改變動畫速度。

## 可注入的來源

| 來源 | 注入點 | 備註 |
|---|---|---|
| 亂數 | `title.Rand`、`charcreate.Rand`、`world.Rand`、`battle.Rand` 介面；`town.Session.Rand` 函式 | 預設實作是 `rng.LCG`（原版那一支，RE-132）；測試與玩家路徑以 `-seed`、`-ambush-rolls` 固定 |
| 時鐘 | 規則層沒有；存檔時間戳記在 `cmd/kol/main.go` 直接讀系統時間 | 只寫進存檔的 `SavedAt`，不影響規則；要對拍時比欄位而不比時間戳記 |
| 輸入 | 規則層收字串 token（`Session.Key`、戰鬥與野外的 token 函式），不收裝置事件 | 鍵盤、滑鼠與 `-play-script` 都先轉成 token，所以同一串 token 能重播 |
| 隨機 ID | `cmd/kol/creation_commit.go` 用 `crypto/rand` 產新角色的 ID | 最外層；規則層只看已產生的 ID |

守門測試的例外清單 `clockAndRandom` 只列上面兩個最外層的檔案；列了卻沒有用到會失敗，
避免清單過期。

## 守門測試

`internal/layers/layers_test.go` 有四項：

1. 每個內部套件都登記在 `allowed`，表裡也沒有不存在的套件。
2. 每個檔案的內部依賴都在該套件的允許清單內。
3. 除了 `cmd/kol`、`ui`、`font`、`screen`，沒有套件依賴 Ebitengine。
4. 除了例外清單，沒有檔案使用 `time.Now` 系列、`math/rand` 或 `crypto/rand`。

負對照：暫時在 `town` 放一個依賴 `battle`、Ebitengine 與 `time.Now` 的檔案，在 `battle`
放一個引用 `math/rand` 的檔案，第 2、3、4 項都失敗；檔案移除後通過。

## 不含

- 沒有檢查同一層內的呼叫是否合理，只檢查套件層級的 import。
- `oracle` 套件的測試需要 dosgolem，在 `-tags oracle` 之下才會編譯，不在預設守門範圍。
- 存檔 schema 沒有另立機器可讀的 schema 檔；v3 的欄位與驗證規則寫在 SPEC-023，
  程式的讀寫與拒絕缺欄由 `internal/save` 的測試守住。
