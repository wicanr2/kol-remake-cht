# 原版與 remake 角色、站位對齊

日期：2026-09-08。**受控驗證情境的角色、站位、目前角色、鏡頭中心與標記定位通過。**
尚未宣稱地形、完整角色數值、規劃順序、攻擊結算或正常 remake GUI 同狀態通過。

## 比較配置

座標沿用原版參戰記錄：X 是列，Y 是欄。

| 槽位 | 身分 | 原版（X,Y） | remake（X,Y） |
|---|---|---|---|
| 0 | Olear | (33,30) | (33,30) |
| 1 | Theodoric，目前規劃角色 | (31,30) | (31,30) |
| 6 | 敵人，伏擊畫面為 SLEDGE | (31,37) | (31,37) |

敵人數為 1，有效槽位上限為 7（索引 0…6）；隊伍空槽仍留在前六格。
原版有效範圍後的記憶體為零填，不得因 Side byte 為 0 就當成更多隊員。
remake 使用 `NewQuest` 建構同一配置，保留其餘空槽，再將目前角色設為槽 1，
以 `Current`、`Camera` 及 `SpriteAt` 核對實際定位。這是測試情境注入，不是
把正常 remake 遭遇流程改成固定配置；測試也未聲稱驗證整輪規劃順序。

## 來源與固定 seed

原版由 dosgolem 正常選角、招募、離店、出城、露營、次日 SLEDGE 伏擊及 Space
確認進入規劃畫面。完整路徑見 `battle-normal-seed-20260908.md`。
沿用 RE-080／RE-132 與 SPEC-020 的已證實欄位，DS 相對記錄基底 0x6712、
每格 50 byte；Side +1，X +0x1F，Y +0x20。目前角色 DS:9E5A，敵人數 DS:2D9C，
最後有效槽位 DS:1CF0。原始記錄與解碼欄位並列保存，不以名稱取代原始定位。
這些位移的原版輸入與 IDA 位址基準見上述正常路徑報告及既有規格／RE 證據。

共同 seed=1 設定在正常第一名隊員規劃入口、原版畫格擷取後；remake 載入對齊情境
後同樣設定 seed=1。這只是後續規則比較的起點，沒有重排已完成的原版規劃或追查
跨整段遊玩的 RND() 呼叫次數。

## 驗證與產物

- 原版 `TestNormalBookEntry`：105.56 秒通過，來源變動旗標 false。
- remake `TestOriginalBattleActorAlignment`：讀取原版快照，檢查原始記錄／解碼一致、
  合法站位、無重疊、有效槽位與兩名角色身分，再驗正式建構與定位函式。
- `go test ./internal/battle -count=1 -v -timeout 120s` 全部通過（0.009 秒），
  執行時設定 `KOL_ALIGNMENT_INPUT` 為原版快照、`KOL_ALIGNMENT_OUTPUT` 為比較輸出。
- 首次 remake 測試有 seed 的 uint32／int 編譯錯誤；修正測試參數轉型後乾淨重跑，
  原始失敗另存，不列為遊戲缺陷。

產物根目錄：`workplace/battle-alignment-20260908/bounded/`。
`battle-alignment.json` 是原版狀態，`remake-alignment.json` 是 remake 定位結果，
`alignment-receipt.json` 保存兩側結果、測試與腳本雜湊；原版 `manifest.json` 保留
它自行重生的收據。這些原始記錄與衍生狀態只留本機，不加入公開素材包。
前一份缺少有效上限的快照留在上一層，不能單獨作為完整有效角色清單。
檔案擁有權均為 1000:1000，容器已清理；本輪未改遊戲正式程式或重打包。
