# 最小玩家鏈：PNG → typed data → 互動 → UI → 存讀檔（#30，2026-09-30）

## 結果

這一條最初的用意是先挑一條窄的走通。現有的玩家路徑收據已經把每一段都走過，範圍比「最小」寬，
所以不另建新路徑，直接以下列收據認定完成。完成範圍只到表列的路徑，不宣稱可以從新遊戲一路玩到結局
（見 [玩家路徑 README](player-path/README.md) 的「不宣稱」）。

## 各段的證據

| 段 | 證據 | 怎麼確認 |
|---|---|---|
| PNG | `assets/original-derived-png/{ega,bookdata,fonts}` 與 manifest；`internal/assets` 的測試 | 2026-09-30 掃 `cmd/` 與 `internal/` 的非測試碼：沒有 `archive/zip` 的引用，也沒有原版封裝檔名，runtime 只讀 PNG 與 JSON |
| typed data | `assets/original-derived-data/*.json`（城鎮建築、文字、關鍵字、UI 文字、世界地圖、怪物、法術等）與各自的 loader | 抽取器產出，loader 對缺欄與越界失敗；`full-chain` 的進門、住宿、戰鬥都吃這些資料 |
| 互動 | `full-chain`：城鎮戶外 → 出城 → 野外伏擊 → 戰鬥 → 回城 → 進門 → 住宿。`quest-chain`：標題 → 書本 → 招募 → 接任務 → 戰鬥撿物 → 交付 → 住宿 | `tools/check_player_paths.py` 重跑後狀態軌跡與收據一致 |
| UI | 兩條路徑的檢查點畫格與里程碑（隊伍列、圖示列、對話、戰鬥階段） | 里程碑驗機制，截圖只作示意 |
| 存讀檔 | `full-chain` 與 `quest-chain` 在住宿時寫出 `kol-save-v3`，收據逐欄核對（金幣、傷勢、所在地、任務旗標）；`full-chain-reload` 與 `quest-chain-reload` 讀回，隊伍列與住宿那一步逐字相同 | 兩條重載路徑各自的里程碑 |

## 不宣稱

- 沒有單獨一條「只走這五段」的最小路徑；證據是既有較長路徑的交集。
- 標題選角、書本與建角走的是名冊角色，不是新遊戲從零開始的結局鏈。
- 平台與發行封包另由 #37、#42 追蹤。
