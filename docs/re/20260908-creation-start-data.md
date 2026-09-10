# 建角起始資料接線審查

日期：2026-09-08。用途：將 SPEC-005 已有 writer 契約接到 remake typed data，
不新增戰鬥規則或深入追查 RND()。

## 來源與定位

- 固定輸入 `Knights-of-legend.zip`，SHA-256：
  `2b455e3c13ec7ecb727c918eefbe261fe741a460285bdd0e638a6aeb6ec49f46`。
- 封裝內 `Knights-of-legend/REGION0.DAT`，48 bytes，SHA-256：
  `21d1291a41792688268af02bbeabcf56a5402c614d3c0d86da2adbe55a9e6ca4`。
  本輪以 Python 3.11 在 `kol-package:20260907` 容器唯讀查證，只有這一筆 REGION 檔案。
- 既有 IDA Pro 9.4 證據：`docs/re/ida-fingerprints/KOLBOOK-character-writer-summary.json`。
  分析輸入 `KOLBOOK.EXE`，SHA-256：
  `2cada02856d3cf5c3bec948ef3736be6f4bd47e3abce94ddf835c1678e0e5832`。
  位址空間為該 MZ 的 IDA 線性位址；本輪未改寫 IDA 資料庫或符號。

## 最小充分對照

| 附加語意 | 原始定位 | 等級與出處 |
|---|---|---|
| 出生建築 ID 14 | REGION 檔案偏移 0x01、0x02 → 角色記錄相對偏移 0x2E、0x2F；IDA 0x1249C、0x124B1 | 已證實：既有 writer copy map，加 RE-073 建築 ID 欄位 |
| 出生旅館 TROLLSBANE | REGION 檔案偏移 0x03 起十字元 → 角色記錄相對偏移 0x40；IDA 0x124D1 | 已證實：固定來源內容、writer copy map、既有名冊所在地契約 |
| 出生城鎮 BRT | REGION 檔案偏移 0x0D 起三字元 → 角色記錄相對偏移 0x90；IDA 0x12501 | 已證實：固定來源內容、writer copy map、SPEC-016 城鎮代碼 |
| 起始金冠及武器技能 | IDA 職業 metadata 0x19348、word table 0x1940C；摘要內 39 筆 starting contracts | 已證實：SPEC-005／RE-050；不從職業名稱猜預設值 |
| 身高與體重 | `extract_character_roster.character_sheet` 的既有 RE-074 公式 | 已證實：沿用現有抽取器常數，依已接受的體型／種族／性別計算 |

REGION 其他複製欄位的未知語意仍保留未知，本輪沒有將全部欄位升格成已證實。
本表同時保留檔案偏移、角色記錄相對偏移及 IDA 線性位址的區別。

## 可重生輸出

容器內執行 `python3 tools/extract_creation_start.py`，產生
`assets/original-derived-data/creation-start.json`，包含來源雜湊、39 筆起始資料、
出生地與體型公式常數。runtime 只讀該 typed JSON，不解析 REGION、BOOKDATA 或 CHARDATA。
衍生資料含原始物品記錄，僅留本機，不適用 private PNG 特例。

原版曾以同槽寫入角色與肖像；remake 在同一 JSON 原子交易內保存完整角色與已驗證
PNG 模板索引。這是 SPEC-005 的保存設計要求，不宣稱原版具有跨檔案原子回滾。
