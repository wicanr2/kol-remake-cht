# 驗證矩陣

| 子系統 | 單元／真實資料測試 | 原版 oracle | 視覺 oracle | 玩家路徑 | 結果 |
|---|---|---|---|---|---|
| 原版輸入清冊 | ZIP CRC、逐檔 SHA-256、雙容器重生 byte compare | 不適用 | 不適用 | 不適用 | 通過 |
| KOLTOWN 解包研究映像 | header、2,281 relocations、IDA entry／337 functions | startup 前 129,312 bytes、2,281 relocations、五錨點全吻合 | 不適用 | `KOLTOWN.EXE A` 正常啟動至 entry | SPEC-001 READY |
| 啟動／新遊戲 | Go 1.26.7＋Ebitengine v2.9.9 建置通過；資產載入拒非 PNG | 書本封面／目錄頁對原版畫面 | 180 對對拍，游標與原版自繪文字外逐像素相同 | 尚無互動；只有截圖模式 | M4 垂直切片；玩法未接 |
| 角色建立 UI | 22 個情境、32 輪，核心情境的檢查點畫面 SHA-256 交叉比對 | KOL.EXE 正常入口；DOSBox-X 2026.06.02 固定機器設定 | 3,117 張去重畫格 PNG，含 80 張圖像 selector | 目錄頁 → Character options → Create Character → 接受屬性 → 接受圖像 → 保存 | SPEC-005 READY；16 槽全滿、重名、字元集合未涵蓋 |
| 書本其他功能 | 尚無 | 四個功能的角色選擇清單是同一個 PNG；HARBOR MASTER 為三行訊息頁 | HARBOR MASTER 1 對對拍通過 | 角色選項 → Harbor Master → 選角色 → 訊息頁 → 任意鍵 | 點陣圖編輯器未實作：滑鼠驅動，鍵盤 oracle 走不完 |
| 角色刪除 | `DeleteFlow` 與 `DeleteSlot` 單元測試；刪除只寫一個 byte | 兩輪 `CHARDATA` 差分完全相同（slot 1 offset 0x000：0x50→0x00） | 13 對逐畫面對拍，允許區外零像素差 | 角色選項 → Delete Character → 選角色 → 確認視窗回答 y／n | SPEC-005 涵蓋；視窗按 ESC、大寫 Y／N、空名冊未涵蓋 |
| 組隊上路 | `town.Inn` 的圖示繞回與名牌單元測試 | 進入鏈、圖示列鍵位與招募流程已量測（RE-053）；招募不寫磁碟 | 旅店畫面 8 對逐畫面對拍，允許區外零像素差 | 目錄頁 6 → 選領隊 → 旅店 → COMPANION → 選角色 | 旅店畫面已對拍；戶外地圖與其餘圖示未開始 |
| 移動／時間 | `Grid`／`Outdoor`／`Markers`／`Session` 單元測試；鏡頭跟隨直接重播 oracle 序列 | 94 張移動 oracle 中 22 張戶外畫格的鏡頭全部唯一解；門口狀態由 town-enter-building 畫格證實 | 22 對逐畫面對拍，允許區外零像素差 | **14 步可重播玩家路徑**：走動 → 門口 → U-TURN → 進建築 → 出來 → 續走 | 地圖、標記、門口與進門都已對拍；離開城鎮與建築內多輪對話未實作 |
| 建築內部 | `EnterBuilding` 組合單元測試 | 原版畫格 072 的八行逐格讀回，與 5-bit 解碼結果逐字相同 | 1 對逐畫面對拍，允許區外零像素差 | 門口 → DOOR → 進門敘述 | 畫面完全由 typed data 組出；後續對話的觸發條件未解 |
| 地圖轉換 | 尚無 | 尚無 | 尚無 | 尚無 | 未開始 |
| 背包／裝備 | 尚無 | 尚無 | 尚無 | 尚無 | 未開始 |
| 戰鬥／AI／獎勵 | 尚無 | 尚無 | 尚無 | 尚無 | 未開始 |
| 事件／任務 | 尚無 | 尚無 | 尚無 | 尚無 | 未開始 |
| 存讀檔 | 尚無 | 尚無 | 尚無 | 尚無 | 未開始 |
| 圖像／主題 | production 抽取器逐張重讀驗證，769 張 PNG 全部對上 visual probe 雜湊 | IDA consumer＋BIOS mode 0Dh；CHARPICS 80/80、BOOKPICS[0] 對原版執行畫面 | 80 張角色模板在 (208,106) 逐像素相同；封面僅差滑鼠游標 8×11 | 由 picture-cycle 正常玩家路徑取得對照畫面 | SPEC-002 READY；首版 EGA-only，CGA／Hercules 列後續 |
| 在地化／字型 | 兩套原版字型各 93 字模抽出並重生；中文字模表 344 字，文字安全矩形 122 項全過 | HUMAN 五字模逐 byte 命中壓縮 payload | 空白 New Character 頁 64,000 px 只差 31 px | 由正常玩家路徑取得對照畫面 | 原版字型已關閉；繁中採雙模式，書本五頁與旅店畫面已中文化 |
| 封包 | 尚無 | 不適用 | 不適用 | 尚無 | 未開始 |

## 截圖 metadata

原版角色建立 UI 的逐畫面 PNG 已由 `tools/dosbox_character_ui_oracle.py` 產生，
3,117 個去重畫格連同收據、跨輪比對與 provenance manifest 存於
`docs/verification/character-ui/`；抽取資產對原版畫面的收據見
`docs/verification/asset-parity/`。權利分類為 private-repo 可保存、不得公開散布。
remake 同狀態截圖存於 `docs/verification/remake-parity/shots/`，對拍計畫與收據同目錄。

## 正常玩家路徑

`player-path/` 有十二份可重播收據，全部要求每一步有可見效果，且指定的檢查點與
已通過對拍的截圖逐 byte 相同：

| 路徑 | 步數 | 內容 |
|---|---|---|
| `town` | 14 | 城鎮內走動、門口、進建築，全部經過 `town.Session` |
| `town-edge` | 14 | 走到地圖邊界、林地圖的城門提示、U-TURN 回地圖 |
| `town-talk` | 10 | 建築內的 SPACEBAR 圖示列、LISTEN 的傳聞與 MOUTH 的問話 |
| `town-leave` | 13 | 在城門選 DOOR 離開城鎮，落到世界地圖 |
| `town-round-trip` | 20 | 城鎮 → 野外走到路上 → 退回城鎮入口 → 選 DOOR 進城 |
| `town-recruit` | 15 | 進旅店、把白框移到 COMPANION、回答是否搜尋硬碟、在候選人清單上下移動並加入第一位 |
| `town-rest` | 11 | 進旅店按 REST 住一晚：付掉 56 枚金幣、記下留宿的建築、寫出存檔並逐欄驗證 |
| `town-mirror` | 10 | 進旅店按 MIRROR 打開角色資料頁，移動白框再選 U-TURN 回建築內 |
| `town-load` | 7 | 從 `town-rest` 寫出的存檔開場，資料頁上除了金幣那一欄與沒讀檔時逐像素相同 |
| `town-shop` | 14 | 走進甲冑鋪按 TRADE、移動反白、看明細、買下第一件，扣錢之後直接穿上 |
| `world-road` | 34 | 出城往北走上道路，連按七次 ROAD 自動循路，再按 SIGNPOST 換另一條 |
| `full` | 26 | 從標題序列走到城鎮：標題 → 書本封面 → 目錄頁 → 角色選項（走過 CREATE、DELETE、HARBOR MASTER）→ 角色選擇清單 → 旅店 → DOOR → 戶外地圖 → 門口 → 進建築 → 出來 → 續走，由 `internal/flow` 串起來 |

尚未實作的功能（EDIT PICTURE／EDIT SHIELD、旅店其餘三個圖示、資料頁上的三個圖示）會被記進
`flow.Unimplemented`，不會默默吞掉。

## 發行 gate

- [ ] 全庫測試與靜態分析
- [ ] 翻譯／編碼檢查
- [~] 無除錯捷徑的正常玩家路徑（標題→書本→旅店→城鎮已有 26 步收據；戰鬥與存讀檔未接）
- [~] 可寫存檔與 pristine original 不變檢查（存檔的迴圈已閉合：`town-rest` 寫出
      並逐欄驗證，`town-load` 從那一份開場、畫面只差金幣那一欄。原版 `CHARDATA`
      一律唯讀，匯入是離線工具。戰鬥與任務進度尚未進存檔）
- [ ] 原版／remake 同狀態截圖審查
- [ ] 封包啟動 smoke
