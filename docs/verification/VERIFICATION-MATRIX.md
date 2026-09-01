# 驗證矩陣

| 子系統 | 單元／真實資料測試 | 原版 oracle | 視覺 oracle | 玩家路徑 | 結果 |
|---|---|---|---|---|---|
| 原版輸入清冊 | ZIP CRC、逐檔 SHA-256、雙容器重生 byte compare | 不適用 | 不適用 | 不適用 | 通過 |
| KOLTOWN 解包研究映像 | header、2,281 relocations、IDA entry／337 functions | startup 前 129,312 bytes、2,281 relocations、五錨點全吻合 | 不適用 | `KOLTOWN.EXE A` 正常啟動至 entry | SPEC-001 READY |
| 啟動／新遊戲 | Go 1.26.7＋Ebitengine v2.9.9 建置通過；資產載入拒非 PNG | 書本封面／目錄頁對原版畫面 | 171 對對拍，游標與原版自繪文字外逐像素相同 | 尚無互動；只有截圖模式 | M4 垂直切片；玩法未接 |
| 角色建立 UI | 八個情境各兩輪，檢查點畫面 SHA-256 交叉比對 | KOL.EXE 正常入口；DOSBox-X 2026.06.02 固定機器設定 | 2,115 張去重畫格 PNG，含 80 張圖像 selector | 目錄頁 → Character options → Create Character → 接受屬性 → 接受圖像 → 保存 | SPEC-005 READY；16 槽全滿、重名、字元集合未涵蓋 |
| 角色刪除 | `DeleteFlow` 與 `DeleteSlot` 單元測試；刪除只寫一個 byte | 兩輪 `CHARDATA` 差分完全相同（slot 1 offset 0x000：0x50→0x00） | 13 對逐畫面對拍，允許區外零像素差 | 角色選項 → Delete Character → 選角色 → 確認視窗回答 y／n | SPEC-005 涵蓋；視窗按 ESC、大寫 Y／N、空名冊未涵蓋 |
| 移動／時間 | 尚無 | 尚無 | 尚無 | 尚無 | 未開始 |
| 地圖轉換 | 尚無 | 尚無 | 尚無 | 尚無 | 未開始 |
| 背包／裝備 | 尚無 | 尚無 | 尚無 | 尚無 | 未開始 |
| 戰鬥／AI／獎勵 | 尚無 | 尚無 | 尚無 | 尚無 | 未開始 |
| 事件／任務 | 尚無 | 尚無 | 尚無 | 尚無 | 未開始 |
| 存讀檔 | 尚無 | 尚無 | 尚無 | 尚無 | 未開始 |
| 圖像／主題 | production 抽取器逐張重讀驗證，752 張 PNG 全部對上 visual probe 雜湊 | IDA consumer＋BIOS mode 0Dh；CHARPICS 80/80、BOOKPICS[0] 對原版執行畫面 | 80 張角色模板在 (208,106) 逐像素相同；封面僅差滑鼠游標 8×11 | 由 picture-cycle 正常玩家路徑取得對照畫面 | SPEC-002 READY；首版 EGA-only，CGA／Hercules 列後續 |
| 在地化／字型 | 兩套原版字型各 93 字模抽出並重生；中文字模表 222 字，文字安全矩形 113 項全過 | HUMAN 五字模逐 byte 命中壓縮 payload | 空白 New Character 頁 64,000 px 只差 31 px | 由正常玩家路徑取得對照畫面 | 原版字型已關閉；繁中採雙模式，書本四頁已中文化 |
| 封包 | 尚無 | 不適用 | 不適用 | 尚無 | 未開始 |

## 截圖 metadata

原版角色建立 UI 的逐畫面 PNG 已由 `tools/dosbox_character_ui_oracle.py` 產生，
2,115 個去重畫格連同收據、跨輪比對與 provenance manifest 存於
`docs/verification/character-ui/`；抽取資產對原版畫面的收據見
`docs/verification/asset-parity/`。權利分類為 private-repo 可保存、不得公開散布。
remake 同狀態截圖存於 `docs/verification/remake-parity/shots/`，對拍計畫與收據同目錄。

## 發行 gate

- [ ] 全庫測試與靜態分析
- [ ] 翻譯／編碼檢查
- [ ] 無除錯捷徑的正常玩家路徑
- [ ] 可寫存檔與 pristine original 不變檢查
- [ ] 原版／remake 同狀態截圖審查
- [ ] 封包啟動 smoke
