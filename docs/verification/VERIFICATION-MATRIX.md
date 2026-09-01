# 驗證矩陣

| 子系統 | 單元／真實資料測試 | 原版 oracle | 視覺 oracle | 玩家路徑 | 結果 |
|---|---|---|---|---|---|
| 原版輸入清冊 | ZIP CRC、逐檔 SHA-256、雙容器重生 byte compare | 不適用 | 不適用 | 不適用 | 通過 |
| KOLTOWN 解包研究映像 | header、2,281 relocations、IDA entry／337 functions | startup 前 129,312 bytes、2,281 relocations、五錨點全吻合 | 不適用 | `KOLTOWN.EXE A` 正常啟動至 entry | SPEC-001 READY |
| 啟動／新遊戲 | 尚無 | 尚無 | 尚無 | 尚無 | 未開始 |
| 角色建立 UI | 四個情境各兩輪，檢查點畫面 SHA-256 交叉比對 | KOL.EXE 正常入口；DOSBox-X 2026.06.02 固定機器設定 | 名字欄 16 字元、各層 ESC、REROLL 逐畫面 PNG（僅本機） | 目錄頁 → Character options → Create Character 全鍵盤操作 | SPEC-005 DRAFT；接受屬性／圖像／保存邊界未涵蓋 |
| 移動／時間 | 尚無 | 尚無 | 尚無 | 尚無 | 未開始 |
| 地圖轉換 | 尚無 | 尚無 | 尚無 | 尚無 | 未開始 |
| 背包／裝備 | 尚無 | 尚無 | 尚無 | 尚無 | 未開始 |
| 戰鬥／AI／獎勵 | 尚無 | 尚無 | 尚無 | 尚無 | 未開始 |
| 事件／任務 | 尚無 | 尚無 | 尚無 | 尚無 | 未開始 |
| 存讀檔 | 尚無 | 尚無 | 尚無 | 尚無 | 未開始 |
| 圖像／主題 | 30 邏輯 archive／1,022 payload LZW 探針；COMBPICS 與 TWN index 24 二階 RLE | IDA consumer＋BIOS mode 0Dh；TWN 八張 continuation PNG；MAKEMODE 28 組 catalog | 兩張魔旅士同場景 JPEG 正配 MAE < 錯配一半；TWN 接觸表目視 | 尚無 | SPEC-002 DRAFT；首版範圍已定為 EGA，CGA／Hercules 列後續 |
| 在地化／字型 | 尚無 | 尚無 | 尚無 | 尚無 | 未開始 |
| 封包 | 尚無 | 不適用 | 不適用 | 尚無 | 未開始 |

## 截圖 metadata

原版角色建立 UI 的逐畫面 PNG 已由 `tools/dosbox_character_ui_oracle.py` 產生，收據見
`docs/verification/character-ui/`：尺寸、PNG SHA-256、時間軸與檢查點都已版控，畫面
PNG 本身只留本機。原版執行畫面截圖與 `assets/original-derived-png/` 是不同素材類別，
2026-09-01 的 private repo PNG 特例沒有涵蓋，入庫前需另行權利決策。remake 同狀態截圖
尚無。

## 發行 gate

- [ ] 全庫測試與靜態分析
- [ ] 翻譯／編碼檢查
- [ ] 無除錯捷徑的正常玩家路徑
- [ ] 可寫存檔與 pristine original 不變檢查
- [ ] 原版／remake 同狀態截圖審查
- [ ] 封包啟動 smoke
