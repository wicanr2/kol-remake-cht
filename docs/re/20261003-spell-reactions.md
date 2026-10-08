# RE-207：部位法術骰序與傷害反應

日期：2026-10-03。實作前證據審查通過，現為局部有限 CONFORMED。整款重製仍未完成。

入口為 [DRAFT](../../workplace/spell-reactions-20261003/spec-draft.md)、
[READY](../../workplace/spell-reactions-20261003/spec-ready.md)、
[證據審查](../../workplace/spell-reactions-20261003/evidence-review.json)與
[原版局部收據](../../workplace/spell-reactions-20261003/original-spells-private.json)。
研究工作區為 `workplace/spell-reactions-20261003/`，索引自本頁。

固定輸入：`KOLTOWN-unpacked.EXE`，SHA-256
`01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2`。
IDA Pro 9.4 資料庫 SHA-256
`5c32c977bd420c3a51e5c068ece78f830c4f252445cf138623fbaba64fe551c8`。
位址為 IDA 線性；記憶體夾具寫入標為 DS 偏移，DGROUP 基底 `0x2DED0`。
沿 RE-206 匯出核對 1148 道指令的原始 bytes 與 MZ relocation，不以散文註記為證據。

已證實：`0x283C2` 持續骰、`0x28430` 強度骰、`0x28473` 部位骰依序執行。
成功部位法術清除目標倒數；空格或護符擋下不擲部位。
負面部位以類別 1 呼叫共享傷害。倒地及頭胸失效取消指令並呼叫 `sub_269BA`。
仍站立、部位有效且傷害至少 16，才在 `0x27B6F` 等玩家繼續後判擊退。

dosgolem 固定版本 `968ca774705853276ae55c8d1163c369481ac71f`，八個案例均於執行前固定 seed 1，沒有重擲。
重傷案例傷害 27，等待前 seed `3884216597`，擊退後 `1030492215`。
有武器的倒地與擊退均搬移完整 16-byte 物品。
這是明示初始狀態注入、顯示及等待 stub 的局部常式驗證，未冒充正常玩家路徑。

前批 [RE-206](20261003-spell-area-messages.md) 的 3946 份附件保持；
336 份共用輸入先保存至本工作區的 `parent-inputs/`。
正式接線通過八個原版同狀態案例與四個雙語適配器案例。
雙語各631正常鍵／632狀態及PNG、角色欄位與新v3保存保持RE-206結果；
正式重載各16圖、34物品槽及14持久欄位相符。隔離裝備原型各305狀態／圖保持。
原版局部治療案例使用來源角色的109門檻與16／21／32部位上限，避免夾具的比較條件不一致。
原版正常購買高傷害法術、未決圖示版面與怪物四套資料格式維持開放。

本輪驗收入口為 [正式收據](../verification/spell-reactions-20261003/receipt.json)與
[唯讀重驗](../verification/spell-reactions-20261003/verify.py)。腳本需要專案 `/src:ro` 及本工作區 `/out:ro` 掛載。
