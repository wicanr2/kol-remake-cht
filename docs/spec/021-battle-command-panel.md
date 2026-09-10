# SPEC-021：戰鬥左下命令面板

日期：2026-09-08。狀態：READY（先建 DRAFT 後完成下列證據審查）。

依 `docs/re/20260908-command-panel.md` 的原版位址、固定輸入與 dosgolem 測試，
顯示左下命令、招式、部位、防禦與方向；不改指令結果，不消耗亂數。
原版輸入 KOLTOWN-unpacked.EXE 的 SHA-256 為
`01f6f9a95c4adf8e9da59fa83bae57aa282e5576767b41bfc2f23302aeb9d1d2`，
工具 IDA Pro 9.4 線性位址；動態工具為 dosgolem 現用 fd746405 分支加本機修正。

顯示範圍沿用 SPEC-018 被檢視人物與規劃階段。當前行動者讀 Planning 的
Command／Style／Guard；其他人物讀 Actor。方向皆讀被檢視者 Facing。
敵人只有行動者 Intellect × 敵人 Level ÷3 嚴格大於 SPEC-020 保存門檻才顯示命令；
隊員免檢定。門檻未初始化時不得假定檢定通過；上方裝備與人形不受此閘門影響。

命令 0、0x0F、0x2A 不畫；0x21 畫 ATK／LOC／DEF／DIR，招式零不畫 LOC；
0x27、0x3F 畫 OPT（圖號為招式）及 DIR；0x29 畫 ATK／DEF／DIR；其餘畫 OPT。
圖號換算、十項攻擊／部位表與防禦跳號依原版參數測試，不猜補越界招式。
標籤於 (16,91)、(16,123)、(48,123)、(48,91)，三字母長度使用現有原版字型；
圖示位於同 x、標籤 y+8，以 KOLMPICS 圖碼加三取得，必須為 24×20。
保留當前原版字母標籤；完整繁中戰鬥排版另有既定待驗收項目，不在此猜改。

驗收：原版已通過的 3,168 組參數範圍、敵人閘門及文字／方向路由作依據；
實作需驗來源選擇、檢定邊界、無亂數副作用、缺檔／尺寸拒絕及正常玩家操作。
未驗正常戰鬥與原版同狀態，不升 CONFORMED。

審查：CommandPanel、CommandInspection、CommandIntellect、CommandText、
CommandDirection 的既有原版測試已分別核對參數、可見性、字元及素材路由；
SPEC-020 已接一次性門檻。此範圍不另開規則或忠實度取捨，授權實作上述唯讀面板。
